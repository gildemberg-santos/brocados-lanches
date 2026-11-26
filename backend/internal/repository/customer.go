package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// CollectionCustomerProfiles guarda o perfil de conveniência reconhecido pelo
// telefone normalizado. Não há leitura pública desta coleção.
const CollectionCustomerProfiles = "customer_profiles"

// CustomerProfileRepository implementa a leitura e escrita do perfil de conveniência.
type CustomerProfileRepository struct {
	profiles *mongo.Collection
}

type customerProfileDocument struct {
	ID                bson.ObjectID         `bson:"_id,omitempty"`
	PhoneKey          string                `bson:"phoneKey"`
	Phone             string                `bson:"phone,omitempty"`
	Name              string                `bson:"name,omitempty"`
	Address           *addressDocument      `bson:"address,omitempty"`
	Fulfillment       string                `bson:"fulfillment,omitempty"`
	Zone              *zoneSnapshotDocument `bson:"zone,omitempty"`
	PaymentMethod     string                `bson:"paymentMethod,omitempty"`
	OrderCount        int                   `bson:"orderCount"`
	LastOrderPublicID string                `bson:"lastOrderPublicId,omitempty"`
	LastOrderAt       time.Time             `bson:"lastOrderAt,omitempty"`
	OptedOut          bool                  `bson:"optedOut"`
	CreatedAt         time.Time             `bson:"createdAt"`
	UpdatedAt         time.Time             `bson:"updatedAt"`
	ExpiresAt         time.Time             `bson:"expiresAt"`
}

// NewCustomerProfileRepository valida a conexão e devolve o repositório de perfis.
func NewCustomerProfileRepository(database *mongo.Database) (*CustomerProfileRepository, error) {
	if database == nil {
		return nil, errors.New("mongo database is required")
	}
	return &CustomerProfileRepository{profiles: database.Collection(CollectionCustomerProfiles)}, nil
}

// EnsureCustomerProfileIndexes cria os índices de perfil de forma idempotente.
// O TTL remove o tombstone de opt-out junto com os dados de conveniência.
func EnsureCustomerProfileIndexes(ctx context.Context, database *mongo.Database) error {
	if database == nil {
		return errors.New("mongo database is required")
	}

	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "phoneKey", Value: 1}},
			Options: options.Index().SetName("customer_profiles_phone_key_unique").SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "expiresAt", Value: 1}},
			Options: options.Index().SetName("customer_profiles_expires_at_ttl").SetExpireAfterSeconds(0),
		},
		{
			Keys:    bson.D{{Key: "lastOrderAt", Value: -1}},
			Options: options.Index().SetName("customer_profiles_last_order"),
		},
	}
	if _, err := database.Collection(CollectionCustomerProfiles).Indexes().CreateMany(ctx, indexes); err != nil {
		return fmt.Errorf("ensure indexes on %s: %w", CollectionCustomerProfiles, err)
	}
	return nil
}

// Upsert atualiza o perfil reconhecido pelo telefone. Um pedido com opt-out
// registrado não reativa o perfil, mesmo que uma chamada posterior peça para
// salvar. A falha não deve impedir o pedido, mas é devolvida ao chamador.
func (r *CustomerProfileRepository) Upsert(ctx context.Context, profile domain.CustomerProfile) error {
	if err := profile.Validate(); err != nil {
		return err
	}

	// O filtro com optedOut:false impede que uma janela entre a leitura e a
	// escrita reative um perfil desativado pelo cliente.
	filter := bson.D{
		{Key: "phoneKey", Value: profile.PhoneKey},
		{Key: "optedOut", Value: false},
	}
	existing, found, err := r.FindByPhoneKey(ctx, profile.PhoneKey)
	if err != nil {
		return err
	}
	if found {
		if existing.OptedOut {
			return nil
		}
		objectID, parseErr := bson.ObjectIDFromHex(existing.ID)
		if parseErr != nil {
			return fmt.Errorf("invalid customer profile id: %w", parseErr)
		}
		filter = bson.D{
			{Key: "_id", Value: objectID},
			{Key: "optedOut", Value: false},
		}
	}

	set := bson.D{
		{Key: "phone", Value: profile.Phone},
		{Key: "name", Value: profile.Name},
		{Key: "fulfillment", Value: string(profile.Fulfillment)},
		{Key: "paymentMethod", Value: profile.PaymentMethod},
		{Key: "lastOrderPublicId", Value: profile.LastOrderPublicID},
		{Key: "lastOrderAt", Value: profile.LastOrderAt.UTC()},
		{Key: "optedOut", Value: false},
		{Key: "updatedAt", Value: profile.UpdatedAt.UTC()},
		{Key: "expiresAt", Value: profile.ExpiresAt.UTC()},
	}
	if profile.Address != nil {
		set = append(set, bson.E{Key: "address", Value: addressDocument{
			Street:     profile.Address.Street,
			Number:     profile.Address.Number,
			Complement: profile.Address.Complement,
			District:   profile.Address.District,
			City:       profile.Address.City,
			State:      profile.Address.State,
			ZipCode:    profile.Address.ZipCode,
			Reference:  profile.Address.Reference,
		}})
	}
	if profile.Zone != nil {
		zoneID, parseErr := bson.ObjectIDFromHex(profile.Zone.ID)
		if parseErr != nil {
			return fmt.Errorf("invalid customer profile zone id: %w", parseErr)
		}
		set = append(set, bson.E{Key: "zone", Value: zoneSnapshotDocument{
			ID:       zoneID,
			Name:     profile.Zone.Name,
			FeeCents: profile.Zone.FeeCents,
		}})
	}

	update := bson.D{
		{Key: "$set", Value: set},
		{Key: "$inc", Value: bson.D{{Key: "orderCount", Value: 1}}},
		{Key: "$setOnInsert", Value: bson.D{{Key: "createdAt", Value: profile.CreatedAt.UTC()}}},
	}
	var document customerProfileDocument
	err = r.profiles.FindOneAndUpdate(ctx, filter, update, options.FindOneAndUpdate().SetUpsert(true)).Decode(&document)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil
	}
	if mongo.IsDuplicateKeyError(err) {
		// Outra requisição criou o perfil primeiro; o dado de conveniência
		// não justifica repetir a escrita nem expor o conflito ao cliente.
		return nil
	}
	if err != nil {
		return fmt.Errorf("upsert customer profile: %w", err)
	}
	return nil
}

// OptOut transforma o perfil em um tombstone mínimo: mantém apenas a chave e
// prazos para impedir a recriação automática por pedidos posteriores.
func (r *CustomerProfileRepository) OptOut(ctx context.Context, phoneKey string, at time.Time) error {
	phoneKey = domain.NormalizePhoneKey(phoneKey)
	if phoneKey == "" {
		return &domain.ValidationError{Field: "phoneKey", Code: domain.CodeRequired, Message: "is required"}
	}
	if at.IsZero() {
		at = time.Now().UTC()
	}
	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "optedOut", Value: true},
			{Key: "updatedAt", Value: at.UTC()},
			{Key: "expiresAt", Value: at.UTC().AddDate(0, domain.CustomerProfileRetentionMonths, 0)},
		}},
		{Key: "$unset", Value: bson.D{
			{Key: "phone", Value: ""},
			{Key: "name", Value: ""},
			{Key: "address", Value: ""},
			{Key: "fulfillment", Value: ""},
			{Key: "zone", Value: ""},
			{Key: "paymentMethod", Value: ""},
			{Key: "orderCount", Value: ""},
			{Key: "lastOrderPublicId", Value: ""},
			{Key: "lastOrderAt", Value: ""},
		}},
		{Key: "$setOnInsert", Value: bson.D{{Key: "createdAt", Value: at.UTC()}}},
	}
	_, err := r.profiles.UpdateOne(ctx, bson.D{{Key: "phoneKey", Value: phoneKey}}, update, options.UpdateOne().SetUpsert(true))
	if err != nil {
		return fmt.Errorf("opt out customer profile: %w", err)
	}
	return nil
}

// FindByPhoneKey localiza o perfil sem devolver dados em log ou em erro.
func (r *CustomerProfileRepository) FindByPhoneKey(ctx context.Context, phoneKey string) (domain.CustomerProfile, bool, error) {
	phoneKey = domain.NormalizePhoneKey(phoneKey)
	var document customerProfileDocument
	err := r.profiles.FindOne(ctx, bson.D{{Key: "phoneKey", Value: phoneKey}}).Decode(&document)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return domain.CustomerProfile{}, false, nil
	}
	if err != nil {
		return domain.CustomerProfile{}, false, fmt.Errorf("find customer profile: %w", err)
	}
	return document.toDomain(), true, nil
}

func (document customerProfileDocument) toDomain() domain.CustomerProfile {
	profile := domain.CustomerProfile{
		ID:                document.ID.Hex(),
		PhoneKey:          document.PhoneKey,
		Phone:             document.Phone,
		Name:              document.Name,
		Fulfillment:       domain.Fulfillment(document.Fulfillment),
		PaymentMethod:     document.PaymentMethod,
		OrderCount:        document.OrderCount,
		LastOrderPublicID: document.LastOrderPublicID,
		LastOrderAt:       document.LastOrderAt,
		OptedOut:          document.OptedOut,
		CreatedAt:         document.CreatedAt,
		UpdatedAt:         document.UpdatedAt,
		ExpiresAt:         document.ExpiresAt,
	}
	if document.Address != nil {
		profile.Address = &domain.OrderAddress{
			Street:     document.Address.Street,
			Number:     document.Address.Number,
			Complement: document.Address.Complement,
			District:   document.Address.District,
			City:       document.Address.City,
			State:      document.Address.State,
			ZipCode:    document.Address.ZipCode,
			Reference:  document.Address.Reference,
		}
	}
	if document.Zone != nil {
		profile.Zone = &domain.OrderZone{
			ID:       document.Zone.ID.Hex(),
			Name:     document.Zone.Name,
			FeeCents: document.Zone.FeeCents,
		}
	}
	return profile
}
