// Comando seed aplica os dados de desenvolvimento de design-system/seed.json
// no banco de desenvolvimento ou teste. Nunca rode contra produção.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/brocados-lanches/api/internal/config"
	"github.com/brocados-lanches/api/internal/repository"
	"github.com/brocados-lanches/api/internal/seed"
)

type options struct {
	file   string
	reset  bool
	dryRun bool
}

func main() {
	if err := run(os.Args[1:], os.Getenv); err != nil {
		slog.Error("seed failed", "error", err)
		os.Exit(1)
	}
}

func run(args []string, getenv func(string) string) error {
	flags, err := parseOptions(args)
	if err != nil {
		return err
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	data, err := seed.Load(flags.file)
	if err != nil {
		return err
	}
	plan, err := seed.Build(data)
	if err != nil {
		return err
	}
	plan = seed.ForEnvironment(plan, getenv("APP_ENV"))

	if flags.dryRun {
		logger.Info("seed plan valid",
			"file", flags.file,
			"categories", len(plan.Categories),
			"products", len(plan.Products),
			"deliveryZones", len(plan.Zones),
		)
		return nil
	}

	cfg, err := config.Load(getenv)
	if err != nil {
		return err
	}
	if err := seed.ValidateTarget(cfg.MongoDB, cfg.Env); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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

	result, err := seed.Apply(ctx, database, plan, flags.reset)
	if err != nil {
		return err
	}

	logger.Info("seed applied",
		"database", result.Database,
		"reset", result.Reset,
		"storeSettings", result.StoreSettings,
		"categories", result.Categories,
		"products", result.Products,
		"deliveryZones", result.DeliveryZones,
	)

	return nil
}

func parseOptions(args []string) (options, error) {
	flags := flag.NewFlagSet("seed", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	file := flags.String("file", "../design-system/seed.json", "caminho do JSON de seed")
	reset := flags.Bool("reset", false, "esvazia as coleções semeadas antes de aplicar")
	dryRun := flags.Bool("dry-run", false, "valida e monta o plano sem acessar o MongoDB")

	if err := flags.Parse(args); err != nil {
		return options{}, err
	}

	return options{file: *file, reset: *reset, dryRun: *dryRun}, nil
}
