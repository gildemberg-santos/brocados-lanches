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

// Coleções de autenticação administrativa em docs/05-modelo-de-dados.md.
const (
	CollectionAdminUsers = "admin_users"
	CollectionSessions   = "sessions"
)

type adminUserDocument struct {
	ID           bson.ObjectID `bson:"_id,omitempty"`
	Name         string        `bson:"name"`
	Email        string        `bson:"email"`
	PasswordHash string        `bson:"passwordHash"`
	Role         string        `bson:"role"`
	Active       bool          `bson:"active"`
	LastLoginAt  *time.Time    `bson:"lastLoginAt,omitempty"`
	CreatedAt    time.Time     `bson:"createdAt"`
	UpdatedAt    time.Time     `bson:"updatedAt"`
}

type sessionDocument struct {
	ID                bson.ObjectID `bson:"_id,omitempty"`
	TokenHash         string        `bson:"tokenHash"`
	UserID            bson.ObjectID `bson:"userId"`
	CreatedAt         time.Time     `bson:"createdAt"`
	ExpiresAt         time.Time     `bson:"expiresAt"`
	AbsoluteExpiresAt time.Time     `bson:"absoluteExpiresAt"`
	LastSeenAt        time.Time     `bson:"lastSeenAt"`
	UserAgentHash     string        `bson:"userAgentHash,omitempty"`
}

// AdminRepository grava usuários e sessões do painel. Senha em claro e token
// opaco nunca chegam a esta camada.
type AdminRepository struct {
	users    *mongo.Collection
	sessions *mongo.Collection
}

func NewAdminRepository(database *mongo.Database) (*AdminRepository, error) {
	if database == nil {
		return nil, errors.New("mongo database is required")
	}
	return &AdminRepository{
		users:    database.Collection(CollectionAdminUsers),
		sessions: database.Collection(CollectionSessions),
	}, nil
}

// EnsureAdminIndexes cria os índices de usuários e sessões de forma idempotente.
func EnsureAdminIndexes(ctx context.Context, database *mongo.Database) error {
	if database == nil {
		return errors.New("mongo database is required")
	}

	userIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "email", Value: 1}}, Options: options.Index().SetName("admin_users_email_unique").SetUnique(true)},
		{Keys: bson.D{{Key: "active", Value: 1}, {Key: "role", Value: 1}}, Options: options.Index().SetName("admin_users_active_role")},
	}
	if _, err := database.Collection(CollectionAdminUsers).Indexes().CreateMany(ctx, userIndexes); err != nil {
		return fmt.Errorf("ensure indexes on %s: %w", CollectionAdminUsers, err)
	}

	sessionIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "tokenHash", Value: 1}}, Options: options.Index().SetName("sessions_token_unique").SetUnique(true)},
		{Keys: bson.D{{Key: "expiresAt", Value: 1}}, Options: options.Index().SetName("sessions_expires_ttl").SetExpireAfterSeconds(0)},
		{Keys: bson.D{{Key: "userId", Value: 1}}, Options: options.Index().SetName("sessions_user")},
	}
	if _, err := database.Collection(CollectionSessions).Indexes().CreateMany(ctx, sessionIndexes); err != nil {
		return fmt.Errorf("ensure indexes on %s: %w", CollectionSessions, err)
	}
	return nil
}

// CountUsers informa quantos usuários admin existem. Usado apenas no bootstrap
// inicial: depois do primeiro usuário, a criação passa pelo painel (US-066).
func (r *AdminRepository) CountUsers(ctx context.Context) (int64, error) {
	count, err := r.users.CountDocuments(ctx, bson.D{})
	if err != nil {
		return 0, fmt.Errorf("count admin users: %w", err)
	}
	return count, nil
}

// CreateUser grava um usuário admin. E-mail duplicado vira conflito de domínio.
func (r *AdminRepository) CreateUser(ctx context.Context, user domain.AdminUser) error {
	document := adminUserDocument{
		Name:         user.Name,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		Role:         user.Role,
		Active:       user.Active,
		CreatedAt:    user.CreatedAt.UTC(),
		UpdatedAt:    user.UpdatedAt.UTC(),
	}
	if _, err := r.users.InsertOne(ctx, document); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return domain.ErrConflict
		}
		return fmt.Errorf("insert admin user: %w", err)
	}
	return nil
}

// FindUserByEmail localiza um usuário pelo e-mail normalizado.
func (r *AdminRepository) FindUserByEmail(ctx context.Context, email string) (domain.AdminUser, bool, error) {
	var document adminUserDocument
	err := r.users.FindOne(ctx, bson.D{{Key: "email", Value: email}}).Decode(&document)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return domain.AdminUser{}, false, nil
	}
	if err != nil {
		return domain.AdminUser{}, false, fmt.Errorf("find admin user by email: %w", err)
	}
	return document.toDomain(), true, nil
}

// FindUserByID localiza um usuário pelo ID hexadecimal.
func (r *AdminRepository) FindUserByID(ctx context.Context, id string) (domain.AdminUser, error) {
	objectID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return domain.AdminUser{}, domain.ErrNotFound
	}
	var document adminUserDocument
	if err := r.users.FindOne(ctx, bson.D{{Key: "_id", Value: objectID}}).Decode(&document); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.AdminUser{}, domain.ErrNotFound
		}
		return domain.AdminUser{}, fmt.Errorf("find admin user by id: %w", err)
	}
	return document.toDomain(), nil
}

// RecordLogin atualiza lastLoginAt, remove sessões antigas do usuário e grava a
// nova em uma única operação quando o MongoDB suporta transação.
func (r *AdminRepository) RecordLogin(ctx context.Context, userID string, at time.Time, session domain.AdminSession) error {
	objectID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid admin user id: %w", err)
	}
	document, err := newSessionDocument(session, objectID)
	if err != nil {
		return err
	}

	sessionHandle, err := r.users.Database().Client().StartSession()
	if err != nil {
		return fmt.Errorf("start login session: %w", err)
	}
	defer sessionHandle.EndSession(ctx)

	transactionStarted := false
	writesStarted := false
	transactionErr := sessionHandle.StartTransaction()
	if transactionErr == nil {
		transactionStarted = true
		transactionErr = mongo.WithSession(ctx, sessionHandle, func(sessionCtx context.Context) error {
			err := r.recordLoginWrites(sessionCtx, objectID, at, document)
			writesStarted = err == nil
			if err != nil {
				return err
			}
			return sessionHandle.CommitTransaction(sessionCtx)
		})
	}
	if transactionErr != nil {
		if transactionStarted {
			_ = sessionHandle.AbortTransaction(ctx)
		}
		if writesStarted || !isTransactionUnsupported(transactionErr) {
			return fmt.Errorf("record admin login in transaction: %w", transactionErr)
		}
		return r.recordLoginWrites(ctx, objectID, at, document)
	}
	return nil
}

func (r *AdminRepository) recordLoginWrites(ctx context.Context, userID bson.ObjectID, at time.Time, session sessionDocument) error {
	if _, err := r.users.UpdateOne(ctx,
		bson.D{{Key: "_id", Value: userID}},
		bson.D{{Key: "$set", Value: bson.D{{Key: "lastLoginAt", Value: at.UTC()}, {Key: "updatedAt", Value: at.UTC()}}}},
	); err != nil {
		return fmt.Errorf("update admin last login: %w", err)
	}
	if _, err := r.sessions.DeleteMany(ctx, bson.D{{Key: "userId", Value: userID}}); err != nil {
		return fmt.Errorf("rotate admin sessions: %w", err)
	}
	if _, err := r.sessions.InsertOne(ctx, session); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return domain.ErrConflict
		}
		return fmt.Errorf("insert admin session: %w", err)
	}
	return nil
}

// FindSessionByTokenHash localiza a sessão pelo hash do token do cookie.
func (r *AdminRepository) FindSessionByTokenHash(ctx context.Context, tokenHash string) (domain.AdminSession, error) {
	var document sessionDocument
	if err := r.sessions.FindOne(ctx, bson.D{{Key: "tokenHash", Value: tokenHash}}).Decode(&document); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.AdminSession{}, domain.ErrNotFound
		}
		return domain.AdminSession{}, fmt.Errorf("find admin session: %w", err)
	}
	return document.toDomain(), nil
}

// DeleteSessionByTokenHash remove a sessão atual no logout.
func (r *AdminRepository) DeleteSessionByTokenHash(ctx context.Context, tokenHash string) error {
	if _, err := r.sessions.DeleteOne(ctx, bson.D{{Key: "tokenHash", Value: tokenHash}}); err != nil {
		return fmt.Errorf("delete admin session: %w", err)
	}
	return nil
}

// TouchSession renova a expiração ociosa sem ultrapassar o limite absoluto.
func (r *AdminRepository) TouchSession(ctx context.Context, tokenHash string, expiresAt, lastSeenAt time.Time) error {
	result, err := r.sessions.UpdateOne(ctx,
		bson.D{{Key: "tokenHash", Value: tokenHash}, {Key: "expiresAt", Value: bson.D{{Key: "$gt", Value: lastSeenAt}}}},
		bson.D{{Key: "$set", Value: bson.D{{Key: "expiresAt", Value: expiresAt.UTC()}, {Key: "lastSeenAt", Value: lastSeenAt.UTC()}}}},
	)
	if err != nil {
		return fmt.Errorf("touch admin session: %w", err)
	}
	if result.MatchedCount == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func newSessionDocument(session domain.AdminSession, userID bson.ObjectID) (sessionDocument, error) {
	if session.TokenHash == "" {
		return sessionDocument{}, errors.New("session token hash is required")
	}
	return sessionDocument{
		TokenHash:         session.TokenHash,
		UserID:            userID,
		CreatedAt:         session.CreatedAt.UTC(),
		ExpiresAt:         session.ExpiresAt.UTC(),
		AbsoluteExpiresAt: session.AbsoluteExpiresAt.UTC(),
		LastSeenAt:        session.LastSeenAt.UTC(),
		UserAgentHash:     session.UserAgentHash,
	}, nil
}

func (d adminUserDocument) toDomain() domain.AdminUser {
	return domain.AdminUser{
		ID:           d.ID.Hex(),
		Name:         d.Name,
		Email:        d.Email,
		PasswordHash: d.PasswordHash,
		Role:         d.Role,
		Active:       d.Active,
		LastLoginAt:  d.LastLoginAt,
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
	}
}

func (d sessionDocument) toDomain() domain.AdminSession {
	return domain.AdminSession{
		TokenHash:         d.TokenHash,
		UserID:            d.UserID.Hex(),
		CreatedAt:         d.CreatedAt,
		ExpiresAt:         d.ExpiresAt,
		AbsoluteExpiresAt: d.AbsoluteExpiresAt,
		LastSeenAt:        d.LastSeenAt,
		UserAgentHash:     d.UserAgentHash,
	}
}
