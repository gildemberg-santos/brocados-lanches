package repository

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestNewUploadRepositoryRequiresDatabase(t *testing.T) {
	if _, err := NewUploadRepository(nil); err == nil {
		t.Fatal("NewUploadRepository(nil) error = nil, want error")
	}
}

func TestEnsureUploadIndexesRequiresDatabase(t *testing.T) {
	if err := EnsureUploadIndexes(context.Background(), nil); err == nil {
		t.Fatal("EnsureUploadIndexes(nil) error = nil, want error")
	}
}

func TestUploadRepositoryCreateGetAndAudit(t *testing.T) {
	ctx := testContext(t)
	database := openCatalogTestDatabase(t)
	if err := EnsureAuditIndexes(ctx, database); err != nil {
		t.Fatalf("EnsureAuditIndexes() error = %v", err)
	}
	if err := EnsureUploadIndexes(ctx, database); err != nil {
		t.Fatalf("EnsureUploadIndexes() error = %v", err)
	}
	repository, err := NewUploadRepository(database)
	if err != nil {
		t.Fatalf("NewUploadRepository() error = %v", err)
	}

	id := bson.NewObjectID().Hex()
	data := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n', 0, 0}
	upload := domain.Upload{
		ID: id, ContentType: domain.UploadContentTypePNG, SizeBytes: int64(len(data)),
		Data: data, CreatedAt: testTime(),
	}
	audit := domain.AuditLog{
		ActorID: adminTestActorID, Action: domain.AuditActionCreate, Entity: domain.AuditEntityUpload,
		EntityID: id, After: map[string]any{"contentType": upload.ContentType, "sizeBytes": upload.SizeBytes},
		CreatedAt: testTime(),
	}
	if err := repository.CreateUpload(ctx, upload, audit); err != nil {
		t.Fatalf("CreateUpload() error = %v", err)
	}

	stored, err := repository.GetUpload(ctx, id)
	if err != nil {
		t.Fatalf("GetUpload() error = %v", err)
	}
	if stored.ContentType != upload.ContentType || stored.SizeBytes != upload.SizeBytes ||
		!bytes.Equal(stored.Data, upload.Data) {
		t.Errorf("GetUpload() = %+v, want %+v", stored, upload)
	}
	if stored.CreatedAt.Location() != time.UTC {
		t.Errorf("CreatedAt location = %v, want UTC", stored.CreatedAt.Location())
	}

	count, err := database.Collection(CollectionAuditLogs).CountDocuments(ctx, bson.D{
		{Key: "entity", Value: domain.AuditEntityUpload}, {Key: "entityId", Value: id},
	})
	if err != nil {
		t.Fatalf("CountDocuments() error = %v", err)
	}
	if count != 1 {
		t.Errorf("audit logs = %d, want 1", count)
	}

	if _, err := repository.GetUpload(ctx, bson.NewObjectID().Hex()); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("GetUpload() missing error = %v, want domain.ErrNotFound", err)
	}
	if _, err := repository.GetUpload(ctx, "invalido"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("GetUpload() malformed id error = %v, want domain.ErrNotFound", err)
	}
}
