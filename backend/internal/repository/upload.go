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

// CollectionUploads guarda as imagens enviadas pelo painel (docs/05-modelo-de-dados.md).
const CollectionUploads = "uploads"

type uploadDocument struct {
	ID          bson.ObjectID `bson:"_id,omitempty"`
	ContentType string        `bson:"contentType"`
	SizeBytes   int64         `bson:"sizeBytes"`
	Data        []byte        `bson:"data"`
	CreatedAt   time.Time     `bson:"createdAt"`
}

// EnsureUploadIndexes cria o índice de uploads de forma idempotente.
func EnsureUploadIndexes(ctx context.Context, database *mongo.Database) error {
	if database == nil {
		return errors.New("mongo database is required")
	}

	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "createdAt", Value: -1}},
			Options: options.Index().SetName("uploads_created_at"),
		},
	}
	if _, err := database.Collection(CollectionUploads).Indexes().CreateMany(ctx, indexes); err != nil {
		return fmt.Errorf("create upload indexes: %w", err)
	}
	return nil
}

// UploadRepository grava e lê imagens no MongoDB usando o mesmo banco do catálogo.
type UploadRepository struct {
	uploads   *mongo.Collection
	auditLogs *mongo.Collection
}

// NewUploadRepository vincula o repositório de imagens a um banco.
func NewUploadRepository(database *mongo.Database) (*UploadRepository, error) {
	if database == nil {
		return nil, errors.New("mongo database is required")
	}
	return &UploadRepository{
		uploads:   database.Collection(CollectionUploads),
		auditLogs: database.Collection(CollectionAuditLogs),
	}, nil
}

// CreateUpload grava a imagem e a auditoria na mesma transação.
func (r *UploadRepository) CreateUpload(ctx context.Context, upload domain.Upload, audit domain.AuditLog) error {
	id, err := bson.ObjectIDFromHex(upload.ID)
	if err != nil {
		return fmt.Errorf("invalid upload id: %w", err)
	}
	document := uploadDocument{
		ID:          id,
		ContentType: upload.ContentType,
		SizeBytes:   upload.SizeBytes,
		Data:        upload.Data,
		CreatedAt:   upload.CreatedAt.UTC(),
	}
	write := func(writeCtx context.Context) error {
		if _, err := r.uploads.InsertOne(writeCtx, document); err != nil {
			return fmt.Errorf("insert upload: %w", err)
		}
		return nil
	}

	session, err := r.uploads.Database().Client().StartSession()
	if err != nil {
		return fmt.Errorf("start upload session: %w", err)
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
			if err := writeAuditLog(sessionCtx, r.auditLogs, audit); err != nil {
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
			return fmt.Errorf("upload mutation in transaction: %w", transactionErr)
		}
		if err := write(ctx); err != nil {
			return err
		}
		return writeAuditLog(ctx, r.auditLogs, audit)
	}
	return nil
}

// GetUpload lê a imagem por ObjectID e devolve ErrNotFound quando não existe.
func (r *UploadRepository) GetUpload(ctx context.Context, id string) (domain.Upload, error) {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return domain.Upload{}, domain.ErrNotFound
	}

	var document uploadDocument
	if err := r.uploads.FindOne(ctx, bson.D{{Key: "_id", Value: objectID}}).Decode(&document); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.Upload{}, domain.ErrNotFound
		}
		return domain.Upload{}, fmt.Errorf("find upload: %w", err)
	}
	return domain.Upload{
		ID:          document.ID.Hex(),
		ContentType: document.ContentType,
		SizeBytes:   document.SizeBytes,
		Data:        document.Data,
		CreatedAt:   document.CreatedAt.UTC(),
	}, nil
}
