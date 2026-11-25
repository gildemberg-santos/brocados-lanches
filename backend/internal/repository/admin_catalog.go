package repository

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/brocados-lanches/api/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// As listagens do painel incluem inativos e aceitam paginação, filtros e
// ordenação por campos allowlist. Nenhuma query recebe um campo cru do cliente.

func (r *CatalogRepository) AdminListCategories(ctx context.Context, query domain.AdminListQuery) ([]domain.Category, int64, error) {
	var documents []categoryDocument
	total, err := adminList(ctx, r.categories, query, []string{"name", "slug"}, nil,
		[]string{"name", "slug", "sortOrder", "createdAt", "updatedAt"}, &documents)
	if err != nil {
		return nil, 0, fmt.Errorf("admin list categories: %w", err)
	}
	return convertDocuments(documents, categoryDocument.toDomain), total, nil
}

func (r *CatalogRepository) AdminListProducts(ctx context.Context, query domain.AdminListQuery) ([]domain.Product, int64, error) {
	var extra bson.D
	if query.CategoryID != "" {
		id, err := bson.ObjectIDFromHex(query.CategoryID)
		if err != nil {
			return nil, 0, domain.ErrNotFound
		}
		extra = bson.D{{Key: "categoryId", Value: id}}
	}

	var documents []productDocument
	total, err := adminList(ctx, r.products, query, []string{"name", "slug", "description"}, extra,
		[]string{"name", "slug", "sortOrder", "priceCents", "createdAt", "updatedAt"}, &documents)
	if err != nil {
		return nil, 0, fmt.Errorf("admin list products: %w", err)
	}
	return convertDocuments(documents, productDocument.toDomain), total, nil
}

func (r *CatalogRepository) AdminListAddons(ctx context.Context, query domain.AdminListQuery) ([]domain.Addon, int64, error) {
	var documents []addonDocument
	total, err := adminList(ctx, r.addons, query, []string{"name", "slug"}, nil,
		[]string{"name", "slug", "sortOrder", "priceCents", "createdAt", "updatedAt"}, &documents)
	if err != nil {
		return nil, 0, fmt.Errorf("admin list addons: %w", err)
	}
	return convertDocuments(documents, addonDocument.toDomain), total, nil
}

func (r *CatalogRepository) AdminListDeliveryZones(ctx context.Context, query domain.AdminListQuery) ([]domain.DeliveryZone, int64, error) {
	var documents []deliveryZoneDocument
	total, err := adminList(ctx, r.deliveryZones, query, []string{"name", "slug"}, nil,
		[]string{"name", "slug", "sortOrder", "feeCents", "createdAt", "updatedAt"}, &documents)
	if err != nil {
		return nil, 0, fmt.Errorf("admin list delivery zones: %w", err)
	}
	return convertDocuments(documents, deliveryZoneDocument.toDomain), total, nil
}

func (r *CatalogRepository) AdminCreateCategory(ctx context.Context, category domain.Category, audit domain.AuditLog) error {
	document, err := newCategoryDocument(category)
	if err != nil {
		return err
	}
	return r.writeWithAudit(ctx, r.categories, document, audit, false)
}

func (r *CatalogRepository) AdminUpdateCategory(ctx context.Context, category domain.Category, audit domain.AuditLog) error {
	document, err := newCategoryDocument(category)
	if err != nil {
		return err
	}
	return r.writeWithAudit(ctx, r.categories, document, audit, true)
}

func (r *CatalogRepository) AdminCreateProduct(ctx context.Context, product domain.Product, audit domain.AuditLog) error {
	document, err := newProductDocument(product)
	if err != nil {
		return err
	}
	return r.writeWithAudit(ctx, r.products, document, audit, false)
}

func (r *CatalogRepository) AdminUpdateProduct(ctx context.Context, product domain.Product, audit domain.AuditLog) error {
	document, err := newProductDocument(product)
	if err != nil {
		return err
	}
	return r.writeWithAudit(ctx, r.products, document, audit, true)
}

func (r *CatalogRepository) AdminCreateAddon(ctx context.Context, addon domain.Addon, audit domain.AuditLog) error {
	document, err := newAddonDocument(addon)
	if err != nil {
		return err
	}
	return r.writeWithAudit(ctx, r.addons, document, audit, false)
}

func (r *CatalogRepository) AdminUpdateAddon(ctx context.Context, addon domain.Addon, audit domain.AuditLog) error {
	document, err := newAddonDocument(addon)
	if err != nil {
		return err
	}
	return r.writeWithAudit(ctx, r.addons, document, audit, true)
}

func (r *CatalogRepository) AdminCreateDeliveryZone(ctx context.Context, zone domain.DeliveryZone, audit domain.AuditLog) error {
	document, err := newDeliveryZoneDocument(zone)
	if err != nil {
		return err
	}
	return r.writeWithAudit(ctx, r.deliveryZones, document, audit, false)
}

func (r *CatalogRepository) AdminUpdateDeliveryZone(ctx context.Context, zone domain.DeliveryZone, audit domain.AuditLog) error {
	document, err := newDeliveryZoneDocument(zone)
	if err != nil {
		return err
	}
	return r.writeWithAudit(ctx, r.deliveryZones, document, audit, true)
}

// writeWithAudit grava a entidade e a auditoria na mesma transação quando o
// MongoDB suporta; em standalone (desenvolvimento) aplica as duas escritas em
// sequência, como os demais repositórios do projeto.
func (r *CatalogRepository) writeWithAudit(ctx context.Context, collection *mongo.Collection, document any, audit domain.AuditLog, replace bool) error {
	write := func(writeCtx context.Context) error {
		if replace {
			result, err := collection.ReplaceOne(writeCtx, bson.D{{Key: "_id", Value: documentID(document)}}, document)
			if err != nil {
				if mongo.IsDuplicateKeyError(err) {
					return domain.ErrConflict
				}
				return fmt.Errorf("replace catalog document: %w", err)
			}
			if result.MatchedCount == 0 {
				return domain.ErrNotFound
			}
			return nil
		}

		if _, err := collection.InsertOne(writeCtx, document); err != nil {
			if mongo.IsDuplicateKeyError(err) {
				return domain.ErrConflict
			}
			return fmt.Errorf("insert catalog document: %w", err)
		}
		return nil
	}
	writeAudit := func(writeCtx context.Context) error {
		return r.insertAuditLog(writeCtx, audit)
	}

	session, err := collection.Database().Client().StartSession()
	if err != nil {
		return fmt.Errorf("start catalog session: %w", err)
	}
	defer session.EndSession(ctx)

	transactionStarted := false
	writesStarted := false
	transactionErr := session.StartTransaction()
	if transactionErr == nil {
		transactionStarted = true
		transactionErr = mongo.WithSession(ctx, session, func(sessionCtx context.Context) error {
			if err := write(sessionCtx); err != nil {
				return err
			}
			writesStarted = true
			if err := writeAudit(sessionCtx); err != nil {
				return err
			}
			return session.CommitTransaction(sessionCtx)
		})
	}
	if transactionErr != nil {
		if transactionStarted {
			_ = session.AbortTransaction(ctx)
		}
		if writesStarted || !isTransactionUnsupported(transactionErr) {
			return fmt.Errorf("catalog mutation in transaction: %w", transactionErr)
		}
		if err := write(ctx); err != nil {
			return err
		}
		return writeAudit(ctx)
	}
	return nil
}

func documentID(document any) bson.ObjectID {
	switch value := document.(type) {
	case categoryDocument:
		return value.ID
	case productDocument:
		return value.ID
	case addonDocument:
		return value.ID
	case deliveryZoneDocument:
		return value.ID
	default:
		return bson.NilObjectID
	}
}

func adminList[T any](
	ctx context.Context,
	collection *mongo.Collection,
	query domain.AdminListQuery,
	searchFields []string,
	extra bson.D,
	allowedSort []string,
	target *[]T,
) (total int64, err error) {
	filter := bson.D{}
	if len(extra) > 0 {
		filter = append(filter, extra...)
	}
	if query.Active != nil {
		filter = append(filter, bson.E{Key: "active", Value: *query.Active})
	}
	if search := strings.TrimSpace(query.Search); search != "" {
		pattern := regexp.QuoteMeta(search)
		conditions := bson.A{}
		for _, field := range searchFields {
			conditions = append(conditions, bson.D{{Key: field, Value: bson.D{{Key: "$regex", Value: pattern}, {Key: "$options", Value: "i"}}}})
		}
		filter = append(filter, bson.E{Key: "$or", Value: conditions})
	}

	total, err = collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("count documents: %w", err)
	}

	sortDocument, err := adminSortDocument(query, allowedSort)
	if err != nil {
		return 0, err
	}
	findOptions := options.Find().
		SetSort(sortDocument).
		SetSkip(int64((query.Page - 1) * query.PageSize)).
		SetLimit(int64(query.PageSize))
	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		return 0, fmt.Errorf("find documents: %w", err)
	}
	defer func() {
		err = errors.Join(err, cursor.Close(ctx))
	}()

	if err := cursor.All(ctx, target); err != nil {
		return 0, fmt.Errorf("decode documents: %w", err)
	}
	return total, nil
}

func adminSortDocument(query domain.AdminListQuery, allowedSort []string) (bson.D, error) {
	field := query.Sort
	if field == "" {
		field = "sortOrder"
	}
	permitted := false
	for _, candidate := range allowedSort {
		if candidate == field {
			permitted = true
			break
		}
	}
	if !permitted {
		return nil, &domain.ValidationError{Field: "sort", Code: domain.CodeInvalid, Message: "unsupported sort field"}
	}

	direction := 1
	if query.Order == "desc" {
		direction = -1
	}
	return bson.D{
		{Key: field, Value: direction},
		{Key: "_id", Value: 1},
	}, nil
}

func convertDocuments[Document any, Entity any](documents []Document, convert func(Document) Entity) []Entity {
	result := make([]Entity, 0, len(documents))
	for _, document := range documents {
		result = append(result, convert(document))
	}
	return result
}
