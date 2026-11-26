// Comando de migração versionada aplica os dados de referência do projeto
// (contato, horários, catálogo e zonas) sem apagar documentos. Diferente de
// `make seed`, pode gravar em produção, mas só com confirmação explícita.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/brocados-lanches/api/internal/config"
	"github.com/brocados-lanches/api/internal/repository"
	"github.com/brocados-lanches/api/internal/seed"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const (
	version      = "20261125-dados-referencia"
	confirmation = "APPLY_VALIDATED_REFERENCE_DATA_20261125"
	markerKey    = "migration"
	defaultFile  = "../design-system/seed.json"
)

type options struct {
	file    string
	apply   bool
	confirm string
}

type deactivated struct {
	Categories    int64
	Products      int64
	DeliveryZones int64
}

func main() {
	if err := run(os.Args[1:], os.Getenv); err != nil {
		slog.Error("migration failed", "error", err)
		os.Exit(1)
	}
}

func run(args []string, getenv func(string) string) error {
	opts, err := parseOptions(args)
	if err != nil {
		return err
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	data, err := seed.Load(opts.file)
	if err != nil {
		return err
	}
	plan, err := seed.Build(data)
	if err != nil {
		return err
	}

	if !opts.apply {
		logger.Info("migration plan valid",
			"version", version,
			"file", opts.file,
			"categories", len(plan.Categories),
			"products", len(plan.Products),
			"deliveryZones", len(plan.Zones),
			"storeOpen", plan.Store.IsOpen,
			"openingHours", len(plan.Store.OpeningHours),
		)
		return nil
	}

	cfg, err := config.Load(getenv)
	if err != nil {
		return err
	}
	if err := validateTarget(cfg, opts.confirm); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	client, err := repository.Connect(ctx, cfg.MongoURI)
	if err != nil {
		return err
	}
	defer func() {
		shutdownContext, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelShutdown()

		if err := client.Close(shutdownContext); err != nil {
			logger.Error("mongo disconnect failed", "error", err)
		}
	}()

	database := client.Database(cfg.MongoDB)
	if database == nil {
		return fmt.Errorf("mongo database %q is unavailable", cfg.MongoDB)
	}

	// O kill switch é decisão da operação, não conteúdo de referência. Em
	// produção o valor atual é preservado; o seed apenas o usa em dev/teste.
	killSwitchPreserved := false
	if cfg.Env == "production" {
		current, err := loadKillSwitch(ctx, database)
		if err != nil {
			return err
		}
		plan.Store.IsOpen = current
		killSwitchPreserved = true
	}

	result, err := seed.ApplyWithOptions(ctx, database, plan, seed.ApplyOptions{
		MarkerKey: markerKey,
		Marker:    migrationMarker(),
	})
	if err != nil {
		return err
	}

	stale, err := deactivateStale(ctx, database, plan)
	if err != nil {
		return err
	}

	logger.Info("migration applied",
		"version", version,
		"database", result.Database,
		"killSwitchPreserved", killSwitchPreserved,
		"storeSettings", result.StoreSettings,
		"categories", result.Categories,
		"products", result.Products,
		"deliveryZones", result.DeliveryZones,
		"deactivatedCategories", stale.Categories,
		"deactivatedProducts", stale.Products,
		"deactivatedDeliveryZones", stale.DeliveryZones,
	)
	return nil
}

func validateTarget(cfg config.Config, confirm string) error {
	if confirm != confirmation {
		return fmt.Errorf("confirmação %q é obrigatória para aplicar a migração %s", confirmation, version)
	}

	switch cfg.Env {
	case "production":
		if cfg.MongoDB != "brocados" {
			return fmt.Errorf("APP_ENV=production exige MONGO_DB=brocados, recebido %q", cfg.MongoDB)
		}
	case "development":
		if !strings.HasSuffix(cfg.MongoDB, "_dev") {
			return fmt.Errorf("APP_ENV=development exige banco com sufixo _dev, recebido %q", cfg.MongoDB)
		}
	case "test":
		if !strings.HasSuffix(cfg.MongoDB, "_test") {
			return fmt.Errorf("APP_ENV=test exige banco com sufixo _test, recebido %q", cfg.MongoDB)
		}
	default:
		return fmt.Errorf("APP_ENV inválido: %q", cfg.Env)
	}
	return nil
}

func migrationMarker() bson.D {
	return bson.D{
		{Key: "version", Value: version},
		{Key: "source", Value: defaultFile},
	}
}

func loadKillSwitch(ctx context.Context, database *mongo.Database) (bool, error) {
	if database == nil {
		return false, fmt.Errorf("mongo database is required")
	}

	var document struct {
		IsOpen *bool `bson:"isOpen"`
	}
	err := database.Collection(seed.CollectionStoreSettings).
		FindOne(ctx, bson.D{{Key: "_id", Value: "main"}}).
		Decode(&document)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("load store kill switch: %w", err)
	}
	return killSwitchValue(document.IsOpen), nil
}

func killSwitchValue(current *bool) bool {
	if current == nil {
		return true
	}
	return *current
}

// deactivateStale desativa documentos que não fazem parte do plano. A migração
// não apaga histórico: expansão e contrair acontecem em duas etapas seguras.
func deactivateStale(ctx context.Context, database *mongo.Database, plan seed.Plan) (deactivated, error) {
	if database == nil {
		return deactivated{}, fmt.Errorf("mongo database is required")
	}

	now := time.Now().UTC()
	result := deactivated{}

	counts, err := deactivateCollection(ctx, database, seed.CollectionCategories, categorySlugs(plan), now, nil)
	if err != nil {
		return deactivated{}, err
	}
	result.Categories = counts

	counts, err = deactivateCollection(ctx, database, seed.CollectionProducts, productSlugs(plan), now, bson.D{
		{Key: "available", Value: false},
	})
	if err != nil {
		return deactivated{}, err
	}
	result.Products = counts

	counts, err = deactivateCollection(ctx, database, seed.CollectionDeliveryZones, zoneSlugs(plan), now, nil)
	if err != nil {
		return deactivated{}, err
	}
	result.DeliveryZones = counts

	return result, nil
}

func deactivateCollection(
	ctx context.Context,
	database *mongo.Database,
	collection string,
	slugs []string,
	now time.Time,
	extra bson.D,
) (int64, error) {
	if len(slugs) == 0 {
		return 0, fmt.Errorf("migration plan has no slugs for %s", collection)
	}

	filter := bson.D{
		{Key: "slug", Value: bson.D{{Key: "$nin", Value: slugs}}},
		{Key: "active", Value: true},
	}
	set := bson.D{
		{Key: "active", Value: false},
		{Key: "updatedAt", Value: now},
		{Key: markerKey, Value: migrationMarker()},
	}
	set = append(set, extra...)

	update, err := database.Collection(collection).UpdateMany(ctx, filter, bson.D{{Key: "$set", Value: set}})
	if err != nil {
		return 0, fmt.Errorf("deactivate stale %s: %w", collection, err)
	}
	return update.ModifiedCount, nil
}

func categorySlugs(plan seed.Plan) []string {
	slugs := make([]string, 0, len(plan.Categories))
	for _, category := range plan.Categories {
		slugs = append(slugs, category.Slug)
	}
	return slugs
}

func productSlugs(plan seed.Plan) []string {
	slugs := make([]string, 0, len(plan.Products))
	for _, product := range plan.Products {
		slugs = append(slugs, product.Slug)
	}
	return slugs
}

func zoneSlugs(plan seed.Plan) []string {
	slugs := make([]string, 0, len(plan.Zones))
	for _, zone := range plan.Zones {
		slugs = append(slugs, zone.Slug)
	}
	return slugs
}

func parseOptions(args []string) (options, error) {
	flags := flag.NewFlagSet("migrate", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	file := flags.String("file", defaultFile, "caminho do JSON de referência")
	apply := flags.Bool("apply", false, "grava no MongoDB; sem esta flag o comando apenas valida o plano")
	confirm := flags.String("confirm", "", "confirmação explícita exigida para gravar")

	if err := flags.Parse(args); err != nil {
		return options{}, err
	}

	return options{file: *file, apply: *apply, confirm: *confirm}, nil
}
