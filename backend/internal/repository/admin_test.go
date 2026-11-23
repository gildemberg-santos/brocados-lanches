package repository

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const adminTestDatabase = "brocados_admin_test"

func TestNewAdminRepositoryRequiresDatabase(t *testing.T) {
	if _, err := NewAdminRepository(nil); err == nil {
		t.Fatal("NewAdminRepository(nil) error = nil, want error")
	}
}

func TestEnsureAdminIndexesRequiresDatabase(t *testing.T) {
	if err := EnsureAdminIndexes(context.Background(), nil); err == nil {
		t.Fatal("EnsureAdminIndexes(nil) error = nil, want error")
	}
}

func TestAdminRepositoryUserAndSessionLifecycle(t *testing.T) {
	ctx := testContext(t)
	database := openAdminTestDatabase(t)
	repository, err := NewAdminRepository(database)
	if err != nil {
		t.Fatalf("NewAdminRepository() error = %v", err)
	}

	now := testTime()
	user := domain.AdminUser{
		Name:         "Dono",
		Email:        "dono@example.com",
		PasswordHash: "hash-argon2id",
		Role:         domain.AdminRoleOwner,
		Active:       true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := repository.CreateUser(ctx, user); err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	duplicated := user
	duplicated.Name = "Outro"
	if err := repository.CreateUser(ctx, duplicated); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("duplicate e-mail error = %v, want ErrConflict", err)
	}

	found, ok, err := repository.FindUserByEmail(ctx, "dono@example.com")
	if err != nil || !ok {
		t.Fatalf("FindUserByEmail() found = %v, error = %v", ok, err)
	}
	if found.Name != "Dono" || found.Role != domain.AdminRoleOwner || !found.Active {
		t.Errorf("user = %+v", found)
	}
	if found.ID == "" {
		t.Error("user ID is empty")
	}
	if _, ok, err := repository.FindUserByEmail(ctx, "ningu@example.com"); err != nil || ok {
		t.Errorf("missing user found = %v, error = %v", ok, err)
	}

	byID, err := repository.FindUserByID(ctx, found.ID)
	if err != nil || byID.Email != user.Email {
		t.Fatalf("FindUserByID() user = %+v, error = %v", byID, err)
	}
	if _, err := repository.FindUserByID(ctx, "id-invalido"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("FindUserByID(invalid) error = %v, want ErrNotFound", err)
	}

	count, err := repository.CountUsers(ctx)
	if err != nil || count != 1 {
		t.Fatalf("CountUsers() = %d, error = %v, want 1", count, err)
	}

	session := domain.AdminSession{
		TokenHash:         "hash-do-token-um",
		UserID:            found.ID,
		CreatedAt:         now,
		ExpiresAt:         now.Add(8 * time.Hour),
		AbsoluteExpiresAt: now.Add(7 * 24 * time.Hour),
		LastSeenAt:        now,
	}
	if err := repository.RecordLogin(ctx, found.ID, now, session); err != nil {
		t.Fatalf("RecordLogin() error = %v", err)
	}
	stored, err := repository.FindSessionByTokenHash(ctx, session.TokenHash)
	if err != nil {
		t.Fatalf("FindSessionByTokenHash() error = %v", err)
	}
	if stored.UserID != found.ID || !stored.ExpiresAt.Equal(session.ExpiresAt) {
		t.Errorf("session = %+v", stored)
	}

	updatedUser, err := repository.FindUserByID(ctx, found.ID)
	if err != nil {
		t.Fatalf("FindUserByID() error = %v", err)
	}
	if updatedUser.LastLoginAt == nil || !updatedUser.LastLoginAt.Equal(now) {
		t.Errorf("lastLoginAt = %v, want %v", updatedUser.LastLoginAt, now)
	}

	rotated := session
	rotated.TokenHash = "hash-do-token-dois"
	rotated.CreatedAt = now.Add(time.Minute)
	rotated.ExpiresAt = rotated.CreatedAt.Add(8 * time.Hour)
	rotated.AbsoluteExpiresAt = rotated.CreatedAt.Add(7 * 24 * time.Hour)
	rotated.LastSeenAt = rotated.CreatedAt
	if err := repository.RecordLogin(ctx, found.ID, rotated.CreatedAt, rotated); err != nil {
		t.Fatalf("second RecordLogin() error = %v", err)
	}
	if _, err := repository.FindSessionByTokenHash(ctx, session.TokenHash); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("old session error = %v, want ErrNotFound", err)
	}

	newExpiry := rotated.ExpiresAt.Add(time.Hour)
	if err := repository.TouchSession(ctx, rotated.TokenHash, newExpiry, rotated.CreatedAt); err != nil {
		t.Fatalf("TouchSession() error = %v", err)
	}
	touched, err := repository.FindSessionByTokenHash(ctx, rotated.TokenHash)
	if err != nil || !touched.ExpiresAt.Equal(newExpiry) {
		t.Errorf("touched session = %+v, error = %v", touched, err)
	}
	if err := repository.TouchSession(ctx, "hash-inexistente", newExpiry, rotated.CreatedAt); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("TouchSession(missing) error = %v, want ErrNotFound", err)
	}

	if err := repository.DeleteSessionByTokenHash(ctx, rotated.TokenHash); err != nil {
		t.Fatalf("DeleteSessionByTokenHash() error = %v", err)
	}
	if _, err := repository.FindSessionByTokenHash(ctx, rotated.TokenHash); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("deleted session error = %v, want ErrNotFound", err)
	}
}

func openAdminTestDatabase(t *testing.T) *mongo.Database {
	t.Helper()

	uri := os.Getenv("MONGO_TEST_URI")
	if uri == "" {
		t.Skip("MONGO_TEST_URI not set; admin repository integration tests skipped")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := Connect(ctx, uri)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}

	database := client.Database(adminTestDatabase)
	if err := database.Drop(ctx); err != nil {
		t.Fatalf("Drop() error = %v", err)
	}

	t.Cleanup(func() {
		cleanupCtx, cancelCleanup := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelCleanup()

		if err := database.Drop(cleanupCtx); err != nil {
			t.Errorf("cleanup Drop() error = %v", err)
		}
		if err := client.Close(cleanupCtx); err != nil {
			t.Errorf("cleanup Close() error = %v", err)
		}
	})

	if err := EnsureAdminIndexes(ctx, database); err != nil {
		t.Fatalf("EnsureAdminIndexes() error = %v", err)
	}
	return database
}
