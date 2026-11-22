package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/brocados-lanches/api/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Nomes das coleções de catálogo em docs/05-modelo-de-dados.md.
const (
	CollectionCategories    = "categories"
	CollectionProducts      = "products"
	CollectionAddons        = "addons"
	CollectionDeliveryZones = "delivery_zones"
)

func activeCatalogFilter() bson.D {
	return bson.D{{Key: "active", Value: true}}
}

func catalogSort() bson.D {
	return bson.D{{Key: "sortOrder", Value: 1}, {Key: "name", Value: 1}}
}

// CatalogRepository lê as coleções de catálogo. As escritas entram nas histórias
// do admin e devem manter o mesmo mapeamento de documentos.
type CatalogRepository struct {
	categories    *mongo.Collection
	products      *mongo.Collection
	addons        *mongo.Collection
	deliveryZones *mongo.Collection
	storeSettings *mongo.Collection
}

// NewCatalogRepository vincula o repositório a um banco.
func NewCatalogRepository(database *mongo.Database) (*CatalogRepository, error) {
	if database == nil {
		return nil, errors.New("mongo database is required")
	}

	return &CatalogRepository{
		categories:    database.Collection(CollectionCategories),
		products:      database.Collection(CollectionProducts),
		addons:        database.Collection(CollectionAddons),
		deliveryZones: database.Collection(CollectionDeliveryZones),
		storeSettings: database.Collection(CollectionStoreSettings),
	}, nil
}

// ListActiveCategories devolve categorias ativas ordenadas para exibição.
func (r *CatalogRepository) ListActiveCategories(ctx context.Context) ([]domain.Category, error) {
	cursor, err := r.categories.Find(ctx, activeCatalogFilter(), options.Find().SetSort(catalogSort()))
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	defer cursor.Close(ctx)

	var documents []categoryDocument
	if err := cursor.All(ctx, &documents); err != nil {
		return nil, fmt.Errorf("decode categories: %w", err)
	}

	categories := make([]domain.Category, len(documents))
	for index, document := range documents {
		categories[index] = document.toDomain()
	}
	return categories, nil
}

// ListActiveProducts devolve produtos ativos, inclusive indisponíveis, para que
// a loja consiga renderizar o estado "esgotado".
func (r *CatalogRepository) ListActiveProducts(ctx context.Context) ([]domain.Product, error) {
	cursor, err := r.products.Find(ctx, activeCatalogFilter(), options.Find().SetSort(catalogSort()))
	if err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}
	defer cursor.Close(ctx)

	var documents []productDocument
	if err := cursor.All(ctx, &documents); err != nil {
		return nil, fmt.Errorf("decode products: %w", err)
	}

	products := make([]domain.Product, len(documents))
	for index, document := range documents {
		products[index] = document.toDomain()
	}
	return products, nil
}

// ListActiveAddons devolve adicionais ativos ordenados para exibição.
func (r *CatalogRepository) ListActiveAddons(ctx context.Context) ([]domain.Addon, error) {
	cursor, err := r.addons.Find(ctx, activeCatalogFilter(), options.Find().SetSort(catalogSort()))
	if err != nil {
		return nil, fmt.Errorf("list addons: %w", err)
	}
	defer cursor.Close(ctx)

	var documents []addonDocument
	if err := cursor.All(ctx, &documents); err != nil {
		return nil, fmt.Errorf("decode addons: %w", err)
	}

	addons := make([]domain.Addon, len(documents))
	for index, document := range documents {
		addons[index] = document.toDomain()
	}
	return addons, nil
}

// ListActiveDeliveryZones devolve zonas ativas ordenadas para exibição.
func (r *CatalogRepository) ListActiveDeliveryZones(ctx context.Context) ([]domain.DeliveryZone, error) {
	cursor, err := r.deliveryZones.Find(ctx, activeCatalogFilter(), options.Find().SetSort(catalogSort()))
	if err != nil {
		return nil, fmt.Errorf("list delivery zones: %w", err)
	}
	defer cursor.Close(ctx)

	var documents []deliveryZoneDocument
	if err := cursor.All(ctx, &documents); err != nil {
		return nil, fmt.Errorf("decode delivery zones: %w", err)
	}

	zones := make([]domain.DeliveryZone, len(documents))
	for index, document := range documents {
		zones[index] = document.toDomain()
	}
	return zones, nil
}

// GetProduct devolve um produto pelo identificador opaco, independentemente de
// estar ativo. Um identificador malformado ou desconhecido vira
// domain.ErrNotFound para não revelar existência; a visibilidade é regra do service.
func (r *CatalogRepository) GetProduct(ctx context.Context, id string) (domain.Product, error) {
	var document productDocument
	if err := findOneByID(ctx, r.products, id, "product", &document); err != nil {
		return domain.Product{}, err
	}
	return document.toDomain(), nil
}

// GetProductBySlug devolve um produto pelo slug público.
func (r *CatalogRepository) GetProductBySlug(ctx context.Context, slug string) (domain.Product, error) {
	var document productDocument
	if err := r.products.FindOne(ctx, bson.D{{Key: "slug", Value: slug}}).Decode(&document); err != nil {
		return domain.Product{}, translateFindError("product", err)
	}
	return document.toDomain(), nil
}

// GetCategory devolve uma categoria pelo identificador opaco.
func (r *CatalogRepository) GetCategory(ctx context.Context, id string) (domain.Category, error) {
	var document categoryDocument
	if err := findOneByID(ctx, r.categories, id, "category", &document); err != nil {
		return domain.Category{}, err
	}
	return document.toDomain(), nil
}

// GetAddon devolve um adicional pelo identificador opaco.
func (r *CatalogRepository) GetAddon(ctx context.Context, id string) (domain.Addon, error) {
	var document addonDocument
	if err := findOneByID(ctx, r.addons, id, "addon", &document); err != nil {
		return domain.Addon{}, err
	}
	return document.toDomain(), nil
}

// GetDeliveryZone devolve uma zona de entrega pelo identificador opaco.
func (r *CatalogRepository) GetDeliveryZone(ctx context.Context, id string) (domain.DeliveryZone, error) {
	var document deliveryZoneDocument
	if err := findOneByID(ctx, r.deliveryZones, id, "delivery zone", &document); err != nil {
		return domain.DeliveryZone{}, err
	}
	return document.toDomain(), nil
}

func findOneByID(ctx context.Context, collection *mongo.Collection, id, entity string, target any) error {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("%w: %s", domain.ErrNotFound, entity)
	}

	if err := collection.FindOne(ctx, bson.D{{Key: "_id", Value: objectID}}).Decode(target); err != nil {
		return translateFindError(entity, err)
	}
	return nil
}

func translateFindError(entity string, err error) error {
	if errors.Is(err, mongo.ErrNoDocuments) {
		return fmt.Errorf("%w: %s", domain.ErrNotFound, entity)
	}
	return fmt.Errorf("find %s: %w", entity, err)
}
