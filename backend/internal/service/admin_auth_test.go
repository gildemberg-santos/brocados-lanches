package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
	"github.com/brocados-lanches/api/internal/platform/password"
	"github.com/brocados-lanches/api/internal/platform/token"
)

type fakeAdminAuthRepository struct {
	users           map[string]domain.AdminUser
	usersByEmail    map[string]domain.AdminUser
	sessions        map[string]domain.AdminSession
	count           int64
	countErr        error
	createErr       error
	findErr         error
	recordErr       error
	sessionErr      error
	deleteErr       error
	touchErr        error
	lastSession     domain.AdminSession
	lastUserID      string
	lastLoginAt     time.Time
	touchedHash     string
	touchedExpiry   time.Time
	deletedHash     string
	recordCallCount int
}

func newFakeAdminAuthRepository() *fakeAdminAuthRepository {
	return &fakeAdminAuthRepository{
		users:        make(map[string]domain.AdminUser),
		usersByEmail: make(map[string]domain.AdminUser),
		sessions:     make(map[string]domain.AdminSession),
	}
}

func (f *fakeAdminAuthRepository) CountUsers(context.Context) (int64, error) {
	return f.count, f.countErr
}

func (f *fakeAdminAuthRepository) CreateUser(_ context.Context, user domain.AdminUser) error {
	if f.createErr != nil {
		return f.createErr
	}
	id := "665f1c2a9b1e4a0012a3b4f1"
	user.ID = id
	f.users[id] = user
	f.usersByEmail[user.Email] = user
	f.count++
	return nil
}

func (f *fakeAdminAuthRepository) FindUserByEmail(_ context.Context, email string) (domain.AdminUser, bool, error) {
	if f.findErr != nil {
		return domain.AdminUser{}, false, f.findErr
	}
	user, found := f.usersByEmail[email]
	return user, found, nil
}

func (f *fakeAdminAuthRepository) FindUserByID(_ context.Context, id string) (domain.AdminUser, error) {
	if user, found := f.users[id]; found {
		return user, nil
	}
	return domain.AdminUser{}, domain.ErrNotFound
}

func (f *fakeAdminAuthRepository) RecordLogin(_ context.Context, userID string, at time.Time, session domain.AdminSession) error {
	f.recordCallCount++
	f.lastUserID = userID
	f.lastLoginAt = at
	f.lastSession = session
	if f.recordErr != nil {
		return f.recordErr
	}
	f.sessions[session.TokenHash] = session
	return nil
}

func (f *fakeAdminAuthRepository) FindSessionByTokenHash(_ context.Context, hash string) (domain.AdminSession, error) {
	if f.sessionErr != nil {
		return domain.AdminSession{}, f.sessionErr
	}
	session, found := f.sessions[hash]
	if !found {
		return domain.AdminSession{}, domain.ErrNotFound
	}
	return session, nil
}

func (f *fakeAdminAuthRepository) DeleteSessionByTokenHash(_ context.Context, hash string) error {
	f.deletedHash = hash
	if f.deleteErr != nil {
		return f.deleteErr
	}
	delete(f.sessions, hash)
	return nil
}

func (f *fakeAdminAuthRepository) TouchSession(_ context.Context, hash string, expiresAt, _ time.Time) error {
	f.touchedHash = hash
	f.touchedExpiry = expiresAt
	if f.touchErr != nil {
		return f.touchErr
	}
	session, found := f.sessions[hash]
	if !found {
		return domain.ErrNotFound
	}
	session.ExpiresAt = expiresAt
	f.sessions[hash] = session
	return nil
}

func newTestAuthService(t *testing.T, repo AdminAuthRepository) (*AdminAuthService, *time.Time) {
	t.Helper()
	service, err := NewAdminAuthService(repo, AdminSessionIdleDuration, AdminSessionAbsoluteDuration)
	if err != nil {
		t.Fatalf("NewAdminAuthService() error = %v", err)
	}
	now := time.Date(2026, 11, 22, 12, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return now }
	service.newToken = func() (string, error) { return "token-opaco", nil }
	return service, &now
}

func seedAdmin(t *testing.T, repo *fakeAdminAuthRepository, active bool) domain.AdminUser {
	t.Helper()
	hash, err := password.Hash("senha-forte-com-12-caracteres")
	if err != nil {
		t.Fatalf("password.Hash() error = %v", err)
	}
	user := domain.AdminUser{
		ID:           "665f1c2a9b1e4a0012a3b4f1",
		Name:         "Operador",
		Email:        "operador@example.com",
		PasswordHash: hash,
		Role:         domain.AdminRoleOperator,
		Active:       active,
	}
	repo.users[user.ID] = user
	repo.usersByEmail[user.Email] = user
	repo.count = 1
	return user
}

func TestNewAdminAuthServiceRejectsInvalidInput(t *testing.T) {
	repo := newFakeAdminAuthRepository()
	if _, err := NewAdminAuthService(nil, time.Hour, 2*time.Hour); err == nil {
		t.Error("nil repository: error = nil, want error")
	}
	if _, err := NewAdminAuthService(repo, 0, 2*time.Hour); err == nil {
		t.Error("zero idle: error = nil, want error")
	}
	if _, err := NewAdminAuthService(repo, time.Hour, time.Hour); err == nil {
		t.Error("absolute equal to idle: error = nil, want error")
	}
}

func TestAdminAuthLogin(t *testing.T) {
	repo := newFakeAdminAuthRepository()
	seedAdmin(t, repo, true)
	service, now := newTestAuthService(t, repo)

	user, rawToken, err := service.Login(context.Background(), "  OPERADOR@example.com ", "senha-forte-com-12-caracteres")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if rawToken != "token-opaco" {
		t.Errorf("token = %q, want token-opaco", rawToken)
	}
	if user.PasswordHash != "" || user.Role != domain.AdminRoleOperator {
		t.Errorf("user = %+v, want safe public copy", user)
	}
	if !repo.lastLoginAt.Equal(*now) {
		t.Errorf("persisted lastLoginAt = %v, want %v", repo.lastLoginAt, *now)
	}
	if repo.recordCallCount != 1 || repo.lastUserID != user.ID {
		t.Errorf("RecordLogin calls = %d, userID = %q", repo.recordCallCount, repo.lastUserID)
	}
	session := repo.lastSession
	if session.TokenHash != token.Hash("token-opaco") {
		t.Error("session stored the raw token instead of its hash")
	}
	if !session.ExpiresAt.Equal(now.Add(AdminSessionIdleDuration)) {
		t.Errorf("session.ExpiresAt = %v", session.ExpiresAt)
	}
	if !session.AbsoluteExpiresAt.Equal(now.Add(AdminSessionAbsoluteDuration)) {
		t.Errorf("session.AbsoluteExpiresAt = %v", session.AbsoluteExpiresAt)
	}
}

func TestAdminAuthLoginRejectsBadCredentialsGenerically(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		password string
		active   bool
		found    bool
	}{
		{name: "unknown email", email: "ningu@example.com", password: "senha-forte-com-12-caracteres", found: false},
		{name: "wrong password", email: "operador@example.com", password: "outra-senha-forte-12", found: true, active: true},
		{name: "inactive user", email: "operador@example.com", password: "senha-forte-com-12-caracteres", found: true, active: false},
		{name: "invalid email", email: "invalido", password: "senha-forte-com-12-caracteres", found: true, active: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := newFakeAdminAuthRepository()
			if test.found {
				seedAdmin(t, repo, test.active)
			}
			service, _ := newTestAuthService(t, repo)

			_, _, err := service.Login(context.Background(), test.email, test.password)
			if !errors.Is(err, domain.ErrUnauthorized) {
				t.Fatalf("Login() error = %v, want ErrUnauthorized", err)
			}
			if repo.recordCallCount != 0 {
				t.Error("session must not be recorded on failed login")
			}
		})
	}
}

func TestAdminAuthLoginPropagatesRepositoryError(t *testing.T) {
	repo := newFakeAdminAuthRepository()
	repo.findErr = errors.New("mongo indisponível")
	service, _ := newTestAuthService(t, repo)

	_, _, err := service.Login(context.Background(), "operador@example.com", "senha-forte-com-12-caracteres")
	if err == nil || errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("Login() error = %v, want wrapped repository error", err)
	}
}

func TestAdminAuthAuthenticate(t *testing.T) {
	repo := newFakeAdminAuthRepository()
	seedAdmin(t, repo, true)
	service, now := newTestAuthService(t, repo)
	if _, _, err := service.Login(context.Background(), "operador@example.com", "senha-forte-com-12-caracteres"); err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	user, err := service.Authenticate(context.Background(), "token-opaco")
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
	if user.PasswordHash != "" {
		t.Error("Authenticate() leaked password hash")
	}
	if repo.touchedHash != token.Hash("token-opaco") {
		t.Error("Authenticate() did not touch the session by hash")
	}
	if !repo.touchedExpiry.Equal(now.Add(AdminSessionIdleDuration)) {
		t.Errorf("touched expiry = %v", repo.touchedExpiry)
	}
}

func TestAdminAuthAuthenticateRejectsInvalidSessions(t *testing.T) {
	repo := newFakeAdminAuthRepository()
	seedAdmin(t, repo, true)
	service, now := newTestAuthService(t, repo)
	if _, _, err := service.Login(context.Background(), "operador@example.com", "senha-forte-com-12-caracteres"); err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	if _, err := service.Authenticate(context.Background(), ""); !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("empty token error = %v, want ErrUnauthorized", err)
	}
	if _, err := service.Authenticate(context.Background(), "desconhecido"); !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("unknown token error = %v, want ErrUnauthorized", err)
	}

	current := *now
	*now = current.Add(AdminSessionIdleDuration)
	if _, err := service.Authenticate(context.Background(), "token-opaco"); !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("expired session error = %v, want ErrUnauthorized", err)
	}
	*now = current

	session := repo.sessions[token.Hash("token-opaco")]
	session.AbsoluteExpiresAt = current.Add(time.Minute)
	repo.sessions[token.Hash("token-opaco")] = session
	*now = current.Add(2 * time.Minute)
	if _, err := service.Authenticate(context.Background(), "token-opaco"); !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("absolute expiry error = %v, want ErrUnauthorized", err)
	}
}

func TestAdminAuthAuthenticateRejectsInactiveUser(t *testing.T) {
	repo := newFakeAdminAuthRepository()
	seedAdmin(t, repo, true)
	service, _ := newTestAuthService(t, repo)
	if _, _, err := service.Login(context.Background(), "operador@example.com", "senha-forte-com-12-caracteres"); err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	user := repo.users["665f1c2a9b1e4a0012a3b4f1"]
	user.Active = false
	repo.users[user.ID] = user

	if _, err := service.Authenticate(context.Background(), "token-opaco"); !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("Authenticate() error = %v, want ErrUnauthorized", err)
	}
}

func TestAdminAuthLogoutIsIdempotent(t *testing.T) {
	repo := newFakeAdminAuthRepository()
	service, _ := newTestAuthService(t, repo)

	if err := service.Logout(context.Background(), ""); err != nil {
		t.Errorf("Logout(empty) error = %v", err)
	}
	if err := service.Logout(context.Background(), "token-opaco"); err != nil {
		t.Errorf("Logout(unknown) error = %v", err)
	}
	if repo.deletedHash != token.Hash("token-opaco") {
		t.Errorf("deleted hash = %q", repo.deletedHash)
	}
	repo.deleteErr = domain.ErrNotFound
	if err := service.Logout(context.Background(), "token-opaco"); err != nil {
		t.Errorf("Logout(not found) error = %v", err)
	}
}

func TestAdminAuthBootstrapOwner(t *testing.T) {
	repo := newFakeAdminAuthRepository()
	service, now := newTestAuthService(t, repo)

	created, err := service.BootstrapOwner(context.Background(), "  Dono  ", " DONO@EXAMPLE.COM ", "senha-forte-com-12-caracteres")
	if err != nil {
		t.Fatalf("BootstrapOwner() error = %v", err)
	}
	if !created {
		t.Error("BootstrapOwner() = false, want true")
	}
	user := repo.users["665f1c2a9b1e4a0012a3b4f1"]
	if user.Role != domain.AdminRoleOwner || !user.Active || user.Email != "dono@example.com" || user.Name != "Dono" {
		t.Errorf("bootstrap user = %+v", user)
	}
	if !password.Verify(user.PasswordHash, "senha-forte-com-12-caracteres") {
		t.Error("bootstrap password hash does not verify")
	}
	if !user.CreatedAt.Equal(*now) {
		t.Errorf("createdAt = %v", user.CreatedAt)
	}

	created, err = service.BootstrapOwner(context.Background(), "Outro", "outro@example.com", "senha-forte-com-12-caracteres")
	if err != nil {
		t.Fatalf("second BootstrapOwner() error = %v", err)
	}
	if created {
		t.Error("second BootstrapOwner() = true, want false")
	}
	if repo.count != 1 {
		t.Errorf("user count = %d, want 1", repo.count)
	}
}

func TestAdminAuthBootstrapOwnerValidatesInput(t *testing.T) {
	repo := newFakeAdminAuthRepository()
	service, _ := newTestAuthService(t, repo)

	tests := []struct {
		name     string
		userName string
		email    string
		password string
	}{
		{name: "empty name", userName: "", email: "dono@example.com", password: "senha-forte-com-12-caracteres"},
		{name: "invalid email", userName: "Dono", email: "invalido", password: "senha-forte-com-12-caracteres"},
		{name: "short password", userName: "Dono", email: "dono@example.com", password: "curta"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := service.BootstrapOwner(context.Background(), test.userName, test.email, test.password); err == nil {
				t.Error("BootstrapOwner() error = nil, want validation error")
			}
		})
	}
	if repo.count != 0 {
		t.Error("invalid bootstrap must not create users")
	}
}
