package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
)

type fakeAdminUsersRepository struct {
	users      map[string]domain.AdminUser
	audits     []domain.AuditLog
	owners     int64
	created    domain.AdminUser
	updated    domain.AdminUser
	listQuery  domain.AdminListQuery
	auditQuery domain.AuditListQuery
	listItems  []domain.AdminUser
	auditItems []domain.AuditLog
	createErr  error
	updateErr  error
	listErr    error
	auditErr   error
	ownersErr  error
	hashErr    error
}

func newFakeAdminUsersRepository() *fakeAdminUsersRepository {
	return &fakeAdminUsersRepository{users: make(map[string]domain.AdminUser)}
}

func (f *fakeAdminUsersRepository) CountUsers(context.Context) (int64, error) {
	return int64(len(f.users)), nil
}

func (f *fakeAdminUsersRepository) AdminCreateUser(_ context.Context, user domain.AdminUser, audit domain.AuditLog) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.created = user
	f.users[user.ID] = user
	f.audits = append(f.audits, audit)
	return nil
}

func (f *fakeAdminUsersRepository) FindUserByEmail(_ context.Context, email string) (domain.AdminUser, bool, error) {
	for _, user := range f.users {
		if user.Email == email {
			return user, true, nil
		}
	}
	return domain.AdminUser{}, false, nil
}

func (f *fakeAdminUsersRepository) FindUserByID(_ context.Context, id string) (domain.AdminUser, error) {
	user, ok := f.users[id]
	if !ok {
		return domain.AdminUser{}, domain.ErrNotFound
	}
	return user, nil
}

func (f *fakeAdminUsersRepository) ListAdminUsers(_ context.Context, query domain.AdminListQuery) ([]domain.AdminUser, int64, error) {
	f.listQuery = query
	return f.listItems, int64(len(f.listItems)), f.listErr
}

func (f *fakeAdminUsersRepository) CountActiveOwners(context.Context) (int64, error) {
	return f.owners, f.ownersErr
}

func (f *fakeAdminUsersRepository) UpdateAdminUser(_ context.Context, user domain.AdminUser, _ time.Time, audit domain.AuditLog) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	f.updated = user
	f.audits = append(f.audits, audit)
	return nil
}

func (f *fakeAdminUsersRepository) ListAuditLogs(_ context.Context, query domain.AuditListQuery) ([]domain.AuditLog, int64, error) {
	f.auditQuery = query
	return f.auditItems, int64(len(f.auditItems)), f.auditErr
}

func newTestAdminUsersService(repo AdminUsersRepository) *AdminUsersService {
	service, err := NewAdminUsersService(repo)
	if err != nil {
		panic(err)
	}
	service.now = func() time.Time { return time.Date(2026, 11, 23, 12, 0, 0, 0, time.UTC) }
	service.hashPassword = func(string) (string, error) { return "hash", nil }
	service.newID = func() (string, error) { return "665f1c2a9b1e4a0012a3b4c9", nil }
	return service
}

func TestNewAdminUsersServiceRequiresRepository(t *testing.T) {
	if _, err := NewAdminUsersService(nil); err == nil {
		t.Fatal("NewAdminUsersService(nil) error = nil, want error")
	}
}

func TestAdminUsersCreate(t *testing.T) {
	repo := newFakeAdminUsersRepository()
	service := newTestAdminUsersService(repo)

	created, err := service.Create(context.Background(), "665f1c2a9b1e4a0012a3b4f9", domain.AdminUser{
		Name: "  Maria  ", Email: " MARIA@Example.COM ", Role: domain.AdminRoleOperator,
	}, "senha-forte-123")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.Name != "Maria" || created.Email != "maria@example.com" {
		t.Errorf("created = %+v", created)
	}
	if created.PasswordHash != "" {
		t.Error("Create() returned passwordHash")
	}
	if repo.created.PasswordHash != "hash" || repo.created.Active != true {
		t.Errorf("stored user = %+v", repo.created)
	}
	if len(repo.audits) != 1 || repo.audits[0].Entity != domain.AuditEntityAdminUser {
		t.Errorf("audits = %+v", repo.audits)
	}
	if _, ok := repo.audits[0].After["password"]; ok {
		t.Error("audit contains password")
	}
}

func TestAdminUsersCreateValidation(t *testing.T) {
	tests := []struct {
		name     string
		user     domain.AdminUser
		password string
	}{
		{"nome vazio", domain.AdminUser{Name: "", Email: "a@b.com", Role: domain.AdminRoleOperator}, "senha-forte-123"},
		{"e-mail inválido", domain.AdminUser{Name: "Ana", Email: "invalido", Role: domain.AdminRoleOperator}, "senha-forte-123"},
		{"papel inválido", domain.AdminUser{Name: "Ana", Email: "ana@b.com", Role: "admin"}, "senha-forte-123"},
		{"senha fraca", domain.AdminUser{Name: "Ana", Email: "ana@b.com", Role: domain.AdminRoleOperator}, "123"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := newTestAdminUsersService(newFakeAdminUsersRepository())
			if _, err := service.Create(context.Background(), "actor", test.user, test.password); err == nil {
				t.Error("Create() error = nil, want validation error")
			}
		})
	}
}

func TestAdminUsersUpdateProtectsLastOwner(t *testing.T) {
	owner := domain.AdminUser{
		ID: "owner-id", Name: "Dono", Email: "dono@example.com", Role: domain.AdminRoleOwner,
		Active: true, UpdatedAt: time.Date(2026, 11, 20, 12, 0, 0, 0, time.UTC),
	}
	repo := newFakeAdminUsersRepository()
	repo.users[owner.ID] = owner
	repo.owners = 1
	service := newTestAdminUsersService(repo)

	inactive := false
	_, err := service.Update(context.Background(), "actor", owner.ID, domain.AdminUserPatch{Active: &inactive})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("Update() error = %v, want ErrConflict", err)
	}

	repo.owners = 2
	if _, err := service.Update(context.Background(), "actor", owner.ID, domain.AdminUserPatch{Active: &inactive}); err != nil {
		t.Fatalf("Update() error = %v, want nil", err)
	}
	if repo.updated.Active {
		t.Error("Update() did not deactivate owner with another active owner")
	}
}

func TestAdminUsersUpdateAuditsSnapshot(t *testing.T) {
	user := domain.AdminUser{
		ID: "user-id", Name: "Ana", Email: "ana@example.com", Role: domain.AdminRoleOperator,
		Active: true, UpdatedAt: time.Date(2026, 11, 20, 12, 0, 0, 0, time.UTC),
	}
	repo := newFakeAdminUsersRepository()
	repo.users[user.ID] = user
	service := newTestAdminUsersService(repo)

	name := "Ana Maria"
	if _, err := service.Update(context.Background(), "actor", user.ID, domain.AdminUserPatch{Name: &name}); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if len(repo.audits) != 1 || repo.audits[0].Before["name"] != "Ana" || repo.audits[0].After["name"] != "Ana Maria" {
		t.Fatalf("audits = %+v", repo.audits)
	}
}

func TestAdminUsersListAndAuditList(t *testing.T) {
	repo := newFakeAdminUsersRepository()
	repo.listItems = []domain.AdminUser{{ID: "u1"}}
	repo.auditItems = []domain.AuditLog{{ID: "a1"}}
	service := newTestAdminUsersService(repo)

	page, err := service.List(context.Background(), domain.AdminListQuery{Page: 2, PageSize: 10, Sort: "email", Order: "asc"})
	if err != nil || len(page.Items) != 1 || repo.listQuery.Page != 2 {
		t.Fatalf("List() = %+v, %v", page, err)
	}

	defaultPage, err := service.List(context.Background(), domain.AdminListQuery{})
	if err != nil {
		t.Fatalf("List(default) error = %v", err)
	}
	if defaultPage.Page != 1 || defaultPage.PageSize != 20 || repo.listQuery.Sort != "name" || repo.listQuery.Order != "asc" {
		t.Errorf("default list = %+v, query = %+v", defaultPage, repo.listQuery)
	}

	auditPage, err := service.AuditList(context.Background(), domain.AuditListQuery{AdminListQuery: domain.AdminListQuery{Page: 1, PageSize: 5}})
	if err != nil || len(auditPage.Items) != 1 {
		t.Fatalf("AuditList() = %+v, %v", auditPage, err)
	}

	from := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	_, err = service.AuditList(context.Background(), domain.AuditListQuery{
		AdminListQuery: domain.AdminListQuery{Page: 1, PageSize: 5},
		From:           &from, Entity: "products", ActorID: "actor",
	})
	if err != nil {
		t.Fatalf("AuditList(filters) error = %v", err)
	}
	if repo.auditQuery.From == nil || repo.auditQuery.Entity != "products" || repo.auditQuery.ActorID != "actor" {
		t.Errorf("auditQuery = %+v", repo.auditQuery)
	}
}

func TestAdminUsersWrapsRepositoryErrors(t *testing.T) {
	repo := newFakeAdminUsersRepository()
	repo.listErr = errors.New("mongo down")
	repo.auditErr = errors.New("mongo down")
	repo.createErr = errors.New("mongo down")
	service := newTestAdminUsersService(repo)

	if _, err := service.List(context.Background(), domain.AdminListQuery{}); err == nil {
		t.Error("List() error = nil, want wrapped error")
	}
	if _, err := service.AuditList(context.Background(), domain.AuditListQuery{}); err == nil {
		t.Error("AuditList() error = nil, want wrapped error")
	}
	if _, err := service.Create(context.Background(), "actor", domain.AdminUser{
		Name: "Ana", Email: "ana@example.com", Role: domain.AdminRoleOperator,
	}, "senha-forte-123"); err == nil {
		t.Error("Create() error = nil, want repository error")
	}
}
