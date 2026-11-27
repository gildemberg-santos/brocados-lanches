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

// CollectionAuditLogs registra mutações administrativas (docs/05-modelo-de-dados.md).
const CollectionAuditLogs = "audit_logs"

type auditLogDocument struct {
	ID        bson.ObjectID  `bson:"_id,omitempty"`
	ActorID   bson.ObjectID  `bson:"actorId"`
	Action    string         `bson:"action"`
	Entity    string         `bson:"entity"`
	EntityID  string         `bson:"entityId"`
	Before    map[string]any `bson:"before,omitempty"`
	After     map[string]any `bson:"after,omitempty"`
	CreatedAt time.Time      `bson:"createdAt"`
}

// EnsureAuditIndexes cria os índices de auditoria de forma idempotente.
func EnsureAuditIndexes(ctx context.Context, database *mongo.Database) error {
	if database == nil {
		return errors.New("mongo database is required")
	}

	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "createdAt", Value: -1}},
			Options: options.Index().SetName("audit_logs_created_at"),
		},
		{
			Keys:    bson.D{{Key: "actorId", Value: 1}, {Key: "createdAt", Value: -1}},
			Options: options.Index().SetName("audit_logs_actor_created_at"),
		},
		{
			Keys:    bson.D{{Key: "entity", Value: 1}, {Key: "entityId", Value: 1}, {Key: "createdAt", Value: -1}},
			Options: options.Index().SetName("audit_logs_entity_created_at"),
		},
	}
	if _, err := database.Collection(CollectionAuditLogs).Indexes().CreateMany(ctx, indexes); err != nil {
		return fmt.Errorf("create audit log indexes: %w", err)
	}
	return nil
}

func (r *CatalogRepository) insertAuditLog(ctx context.Context, audit domain.AuditLog) error {
	return writeAuditLog(ctx, r.auditLogs, audit)
}

// writeAuditLog é compartilhado pelos repositórios que gravam auditoria.
func writeAuditLog(ctx context.Context, collection *mongo.Collection, audit domain.AuditLog) error {
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
	if _, err := collection.InsertOne(ctx, document); err != nil {
		return fmt.Errorf("insert audit log: %w", err)
	}
	return nil
}
