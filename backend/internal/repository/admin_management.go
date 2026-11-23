package repository

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var adminUserSortFields = map[string]string{
	"name": "name", "email": "email", "role": "role",
	"active": "active", "createdAt": "createdAt", "updatedAt": "updatedAt",
}

// AdminCreateUser grava um usuário criado pelo painel e sua auditoria.
func (r *AdminRepository) AdminCreateUser(ctx context.Context, user domain.AdminUser, audit domain.AuditLog) error {
	if err := r.CreateUser(ctx, user); err != nil {
		return err
	}
	return r.insertAuditLog(ctx, audit)
}

// ListAdminUsers devolve uma página de usuários do painel.
func (r *AdminRepository) ListAdminUsers(ctx context.Context, query domain.AdminListQuery) ([]domain.AdminUser, int64, error) {
	filter := bson.D{}
	if query.Search != "" {
		pattern := regexp.QuoteMeta(query.Search)
		filter = append(filter, bson.E{Key: "$or", Value: bson.A{
			bson.D{{Key: "name", Value: bson.D{{Key: "$regex", Value: pattern}, {Key: "$options", Value: "i"}}}},
			bson.D{{Key: "email", Value: bson.D{{Key: "$regex", Value: pattern}, {Key: "$options", Value: "i"}}}},
		}})
	}
	if query.Active != nil {
		filter = append(filter, bson.E{Key: "active", Value: *query.Active})
	}

	sortField := adminUserSortFields[query.Sort]
	if sortField == "" {
		sortField = "name"
	}
	direction := 1
	if query.Order == "desc" {
		direction = -1
	}

	opts := options.Find().
		SetSort(bson.D{{Key: sortField, Value: direction}, {Key: "_id", Value: 1}}).
		SetSkip(int64((query.Page - 1) * query.PageSize)).
		SetLimit(int64(query.PageSize))

	cursor, err := r.users.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("list admin users: %w", err)
	}
	defer cursor.Close(ctx)

	var documents []adminUserDocument
	if err := cursor.All(ctx, &documents); err != nil {
		return nil, 0, fmt.Errorf("decode admin users: %w", err)
	}
	users := make([]domain.AdminUser, len(documents))
	for index, document := range documents {
		users[index] = document.toDomain().Public()
	}

	total, err := r.users.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("count admin users: %w", err)
	}
	return users, total, nil
}

// CountActiveOwners informa quantos owners ativos existem.
func (r *AdminRepository) CountActiveOwners(ctx context.Context) (int64, error) {
	count, err := r.users.CountDocuments(ctx, bson.D{
		{Key: "role", Value: domain.AdminRoleOwner},
		{Key: "active", Value: true},
	})
	if err != nil {
		return 0, fmt.Errorf("count active owners: %w", err)
	}
	return count, nil
}

// UpdateAdminUser aplica um patch com compare-and-set na versão conhecida pelo
// chamador. Mudanças concorrentes retornam domain.ErrConflict.
func (r *AdminRepository) UpdateAdminUser(ctx context.Context, user domain.AdminUser, expectedUpdatedAt time.Time, audit domain.AuditLog) error {
	objectID, err := bson.ObjectIDFromHex(user.ID)
	if err != nil {
		return domain.ErrNotFound
	}
	set := bson.D{
		{Key: "name", Value: user.Name},
		{Key: "role", Value: user.Role},
		{Key: "active", Value: user.Active},
		{Key: "updatedAt", Value: user.UpdatedAt.UTC()},
	}
	result, err := r.users.UpdateOne(ctx,
		bson.D{{Key: "_id", Value: objectID}, {Key: "updatedAt", Value: expectedUpdatedAt.UTC()}},
		bson.D{{Key: "$set", Value: set}},
	)
	if err != nil {
		return fmt.Errorf("update admin user: %w", err)
	}
	if result.MatchedCount == 0 {
		return domain.ErrConflict
	}
	return r.insertAuditLog(ctx, audit)
}

// insertAuditLog permite que serviços de usuário registrem mutações na mesma
// coleção usada pelo catálogo.
func (r *AdminRepository) insertAuditLog(ctx context.Context, audit domain.AuditLog) error {
	actorID, err := bson.ObjectIDFromHex(audit.ActorID)
	if err != nil {
		return fmt.Errorf("invalid audit actor id: %w", err)
	}
	document := auditLogDocument{
		ID:        bson.NewObjectID(),
		ActorID:   actorID,
		Action:    audit.Action,
		Entity:    audit.Entity,
		EntityID:  audit.EntityID,
		Before:    audit.Before,
		After:     audit.After,
		CreatedAt: audit.CreatedAt.UTC(),
	}
	if _, err := r.auditLogs.InsertOne(ctx, document); err != nil {
		return fmt.Errorf("insert audit log: %w", err)
	}
	return nil
}

// ListAuditLogs devolve uma página de auditoria com filtros opcionais.
func (r *AdminRepository) ListAuditLogs(ctx context.Context, query domain.AuditListQuery) ([]domain.AuditLog, int64, error) {
	filter := bson.D{}
	if query.ActorID != "" {
		objectID, err := bson.ObjectIDFromHex(query.ActorID)
		if err != nil {
			return nil, 0, domain.ErrNotFound
		}
		filter = append(filter, bson.E{Key: "actorId", Value: objectID})
	}
	if query.Action != "" {
		filter = append(filter, bson.E{Key: "action", Value: query.Action})
	}
	if query.Entity != "" {
		filter = append(filter, bson.E{Key: "entity", Value: query.Entity})
	}
	if query.From != nil || query.To != nil {
		rangeFilter := bson.D{}
		if query.From != nil {
			rangeFilter = append(rangeFilter, bson.E{Key: "$gte", Value: query.From.UTC()})
		}
		if query.To != nil {
			rangeFilter = append(rangeFilter, bson.E{Key: "$lte", Value: query.To.UTC()})
		}
		filter = append(filter, bson.E{Key: "createdAt", Value: rangeFilter})
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}, {Key: "_id", Value: 1}}).
		SetSkip(int64((query.Page - 1) * query.PageSize)).
		SetLimit(int64(query.PageSize))

	cursor, err := r.auditLogs.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("list audit logs: %w", err)
	}
	defer cursor.Close(ctx)

	var documents []auditLogDocument
	if err := cursor.All(ctx, &documents); err != nil {
		return nil, 0, fmt.Errorf("decode audit logs: %w", err)
	}
	logs := make([]domain.AuditLog, len(documents))
	for index, document := range documents {
		logs[index] = domain.AuditLog{
			ID:        document.ID.Hex(),
			ActorID:   document.ActorID.Hex(),
			Action:    document.Action,
			Entity:    document.Entity,
			EntityID:  document.EntityID,
			Before:    document.Before,
			After:     document.After,
			CreatedAt: document.CreatedAt,
		}
	}

	total, err := r.auditLogs.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("count audit logs: %w", err)
	}
	return logs, total, nil
}
