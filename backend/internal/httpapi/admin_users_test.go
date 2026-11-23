package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
	"github.com/brocados-lanches/api/internal/service"
)

type fakeAdminUsersHandlerService struct {
	users       []domain.AdminUser
	audits      []domain.AuditLog
	created     domain.AdminUser
	createdPass string
	updated     domain.AdminUser
	patch       domain.AdminUserPatch
	actorID     string
	auditQuery  domain.AuditListQuery
}

func (f *fakeAdminUsersHandlerService) List(context.Context, domain.AdminListQuery) (service.AdminPage[domain.AdminUser], error) {
	return service.AdminPage[domain.AdminUser]{Items: f.users, Page: 1, PageSize: 20, Total: int64(len(f.users))}, nil
}

func (f *fakeAdminUsersHandlerService) Create(_ context.Context, actorID string, user domain.AdminUser, password string) (domain.AdminUser, error) {
	f.actorID = actorID
	f.created = user
	f.createdPass = password
	return user, nil
}

func (f *fakeAdminUsersHandlerService) Update(_ context.Context, actorID, _ string, patch domain.AdminUserPatch) (domain.AdminUser, error) {
	f.actorID = actorID
	f.patch = patch
	f.updated = domain.AdminUser{ID: "user-id", Name: "Ana Maria", Email: "ana@example.com", Role: domain.AdminRoleOperator, Active: true}
	return f.updated, nil
}

func (f *fakeAdminUsersHandlerService) AuditList(_ context.Context, query domain.AuditListQuery) (service.AdminPage[domain.AuditLog], error) {
	f.auditQuery = query
	return service.AdminPage[domain.AuditLog]{Items: f.audits, Page: 1, PageSize: 20, Total: int64(len(f.audits))}, nil
}

func newAdminUsersTestRouter(auth AdminAuthService, users AdminUsersService) http.Handler {
	return NewRouter(Dependencies{
		AdminAuth:   auth,
		AdminCookie: AdminCookieConfig{Name: "brocados_admin"},
		AdminUsers:  users,
		Logger:      slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
}

func newAdminOwnerAuth() *fakeAdminAuthService {
	return &fakeAdminAuthService{user: domain.AdminUser{
		ID: testActorIDForHTTP, Name: "Dono", Email: "dono@example.com", Role: domain.AdminRoleOwner,
	}}
}

func TestAdminUsersRoutesRequireOwner(t *testing.T) {
	paths := []string{"/api/v1/admin/users", "/api/v1/admin/audit-logs"}
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			router := newAdminUsersTestRouter(newAdminOperatorAuth(), &fakeAdminUsersHandlerService{})

			response := httptest.NewRecorder()
			router.ServeHTTP(response, newAdminRequest(http.MethodGet, path, ""))
			if response.Code != http.StatusForbidden {
				t.Errorf("operator status = %d, want 403", response.Code)
			}

			anonymous := httptest.NewRecorder()
			request := newAdminRequest(http.MethodGet, path, "")
			request.Header.Del("Cookie")
			router.ServeHTTP(anonymous, request)
			if anonymous.Code != http.StatusUnauthorized {
				t.Errorf("anonymous status = %d, want 401", anonymous.Code)
			}
		})
	}
}

func TestAdminListUsers(t *testing.T) {
	users := &fakeAdminUsersHandlerService{users: []domain.AdminUser{{
		ID: "user-id", Name: "Ana", Email: "ana@example.com", Role: domain.AdminRoleOperator, Active: true,
		PasswordHash: "secret", CreatedAt: time.Date(2026, 11, 20, 12, 0, 0, 0, time.UTC),
	}}}
	router := newAdminUsersTestRouter(newAdminOwnerAuth(), users)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, newAdminRequest(http.MethodGet, "/api/v1/admin/users?page=1&pageSize=20", ""))

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", response.Code, response.Body.String())
	}
	var envelope adminListEnvelope[adminUserResponse]
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(envelope.Data) != 1 || envelope.Data[0].Email != "ana@example.com" {
		t.Fatalf("data = %+v", envelope.Data)
	}
	if envelope.Meta.Total != 1 || envelope.Meta.Page != 1 {
		t.Errorf("meta = %+v", envelope.Meta)
	}
}

func TestAdminCreateUser(t *testing.T) {
	users := &fakeAdminUsersHandlerService{}
	router := newAdminUsersTestRouter(newAdminOwnerAuth(), users)

	body := `{"name":"Ana","email":"ana@example.com","password":"senha-forte-123","role":"operator"}`
	response := httptest.NewRecorder()
	router.ServeHTTP(response, newAdminRequest(http.MethodPost, "/api/v1/admin/users", body))

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body = %s", response.Code, response.Body.String())
	}
	if users.actorID != testActorIDForHTTP || users.createdPass != "senha-forte-123" {
		t.Errorf("created = %+v, password = %q, actor = %q", users.created, users.createdPass, users.actorID)
	}
}

func TestAdminCreateUserRequiresFields(t *testing.T) {
	router := newAdminUsersTestRouter(newAdminOwnerAuth(), &fakeAdminUsersHandlerService{})

	response := httptest.NewRecorder()
	router.ServeHTTP(response, newAdminRequest(http.MethodPost, "/api/v1/admin/users", `{"name":"Ana"}`))

	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", response.Code)
	}
}

func TestAdminUpdateUser(t *testing.T) {
	users := &fakeAdminUsersHandlerService{}
	router := newAdminUsersTestRouter(newAdminOwnerAuth(), users)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, newAdminRequest(http.MethodPatch, "/api/v1/admin/users/user-id", `{"name":"Ana Maria"}`))

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", response.Code, response.Body.String())
	}
	if users.patch.Name == nil || *users.patch.Name != "Ana Maria" || users.actorID != testActorIDForHTTP {
		t.Errorf("patch = %+v, actor = %q", users.patch, users.actorID)
	}
}

func TestAdminListAuditLogs(t *testing.T) {
	users := &fakeAdminUsersHandlerService{audits: []domain.AuditLog{{
		ID: "audit-id", ActorID: testActorIDForHTTP, Action: domain.AuditActionUpdate,
		Entity: domain.AuditEntityProduct, EntityID: "product-id",
		CreatedAt: time.Date(2026, 11, 22, 12, 0, 0, 0, time.UTC),
	}}}
	router := newAdminUsersTestRouter(newAdminOwnerAuth(), users)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, newAdminRequest(http.MethodGet,
		"/api/v1/admin/audit-logs?entity=products&from=2026-11-01T00:00:00Z", ""))

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", response.Code, response.Body.String())
	}
	var envelope adminListEnvelope[adminAuditResponse]
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(envelope.Data) != 1 || envelope.Data[0].Entity != "products" {
		t.Fatalf("data = %+v", envelope.Data)
	}
	if users.auditQuery.Entity != "products" || users.auditQuery.From == nil {
		t.Errorf("auditQuery = %+v", users.auditQuery)
	}
}

func TestAdminListAuditLogsRejectsInvalidDate(t *testing.T) {
	router := newAdminUsersTestRouter(newAdminOwnerAuth(), &fakeAdminUsersHandlerService{})

	response := httptest.NewRecorder()
	router.ServeHTTP(response, newAdminRequest(http.MethodGet, "/api/v1/admin/audit-logs?from=ontem", ""))

	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", response.Code)
	}
}
