package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
	"github.com/brocados-lanches/api/internal/service"
	"github.com/go-chi/chi/v5"
)

// AdminUsersService é a dependência dos endpoints de usuários e auditoria.
type AdminUsersService interface {
	List(ctx context.Context, query domain.AdminListQuery) (service.AdminPage[domain.AdminUser], error)
	Create(ctx context.Context, actorID string, user domain.AdminUser, password string) (domain.AdminUser, error)
	Update(ctx context.Context, actorID, id string, patch domain.AdminUserPatch) (domain.AdminUser, error)
	AuditList(ctx context.Context, query domain.AuditListQuery) (service.AdminPage[domain.AuditLog], error)
}

type adminUserResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Email       string  `json:"email"`
	Role        string  `json:"role"`
	Active      bool    `json:"active"`
	LastLoginAt *string `json:"lastLoginAt"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
}

type adminAuditResponse struct {
	ID        string         `json:"id"`
	ActorID   string         `json:"actorId"`
	Action    string         `json:"action"`
	Entity    string         `json:"entity"`
	EntityID  string         `json:"entityId"`
	Before    map[string]any `json:"before"`
	After     map[string]any `json:"after"`
	CreatedAt string         `json:"createdAt"`
}

type adminUserRequest struct {
	Name     *string `json:"name"`
	Email    *string `json:"email"`
	Password *string `json:"password"`
	Role     *string `json:"role"`
	Active   *bool   `json:"active"`
}

func handleAdminListUsers(users AdminUsersService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if users == nil {
			writeError(w, r, logger, errors.New("admin users service is not configured"))
			return
		}
		query, err := parseAdminListQuery(r)
		if err != nil {
			writeError(w, r, logger, err)
			return
		}
		page, err := users.List(r.Context(), query)
		if err != nil {
			writeError(w, r, logger, err)
			return
		}
		items := make([]adminUserResponse, 0, len(page.Items))
		for _, user := range page.Items {
			items = append(items, newAdminUserResponse(user))
		}
		writeJSON(w, http.StatusOK, adminListEnvelope[adminUserResponse]{
			Data: items, Meta: newAdminMeta(page.Page, page.PageSize, page.Total, page.TotalPages),
		})
	}
}

func handleAdminCreateUser(users AdminUsersService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if users == nil {
			writeError(w, r, logger, errors.New("admin users service is not configured"))
			return
		}
		var request adminUserRequest
		if err := decodeAdminRequest(w, r, &request); err != nil {
			writeError(w, r, logger, err)
			return
		}
		if request.Name == nil || request.Email == nil || request.Password == nil || request.Role == nil {
			writeError(w, r, logger, &domain.ValidationError{Field: "body", Code: domain.CodeRequired, Message: "name, email, password and role are required"})
			return
		}
		actor, ok := adminFromContext(r.Context())
		if !ok {
			writeError(w, r, logger, domain.ErrUnauthorized)
			return
		}
		user, err := users.Create(r.Context(), actor.ID, domain.AdminUser{
			Name: *request.Name, Email: *request.Email, Role: *request.Role,
		}, *request.Password)
		if err != nil {
			writeError(w, r, logger, err)
			return
		}
		writeJSON(w, http.StatusCreated, adminItemEnvelope[adminUserResponse]{Data: newAdminUserResponse(user)})
	}
}

func handleAdminUpdateUser(users AdminUsersService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if users == nil {
			writeError(w, r, logger, errors.New("admin users service is not configured"))
			return
		}
		var request adminUserRequest
		if err := decodeAdminRequest(w, r, &request); err != nil {
			writeError(w, r, logger, err)
			return
		}
		actor, ok := adminFromContext(r.Context())
		if !ok {
			writeError(w, r, logger, domain.ErrUnauthorized)
			return
		}
		user, err := users.Update(r.Context(), actor.ID, chi.URLParam(r, "id"), domain.AdminUserPatch{
			Name: request.Name, Role: request.Role, Active: request.Active,
		})
		if err != nil {
			writeError(w, r, logger, err)
			return
		}
		writeJSON(w, http.StatusOK, adminItemEnvelope[adminUserResponse]{Data: newAdminUserResponse(user)})
	}
}

func handleAdminListAuditLogs(users AdminUsersService, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if users == nil {
			writeError(w, r, logger, errors.New("admin users service is not configured"))
			return
		}
		query, err := parseAdminAuditQuery(r)
		if err != nil {
			writeError(w, r, logger, err)
			return
		}
		page, err := users.AuditList(r.Context(), query)
		if err != nil {
			writeError(w, r, logger, err)
			return
		}
		items := make([]adminAuditResponse, 0, len(page.Items))
		for _, log := range page.Items {
			items = append(items, adminAuditResponse{
				ID: log.ID, ActorID: log.ActorID, Action: log.Action, Entity: log.Entity,
				EntityID: log.EntityID, Before: log.Before, After: log.After,
				CreatedAt: formatAdminTime(log.CreatedAt),
			})
		}
		writeJSON(w, http.StatusOK, adminListEnvelope[adminAuditResponse]{
			Data: items, Meta: newAdminMeta(page.Page, page.PageSize, page.Total, page.TotalPages),
		})
	}
}

func parseAdminAuditQuery(r *http.Request) (domain.AuditListQuery, error) {
	base, err := parseAdminListQuery(r)
	if err != nil {
		return domain.AuditListQuery{}, err
	}
	query := domain.AuditListQuery{
		AdminListQuery: base,
		ActorID:        strings.TrimSpace(r.URL.Query().Get("actorId")),
		Action:         strings.TrimSpace(r.URL.Query().Get("action")),
		Entity:         strings.TrimSpace(r.URL.Query().Get("entity")),
	}
	for _, field := range []struct {
		name string
		dest **time.Time
	}{{"from", &query.From}, {"to", &query.To}} {
		raw := r.URL.Query().Get(field.name)
		if raw == "" {
			continue
		}
		parsed, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return query, &domain.ValidationError{Field: field.name, Code: domain.CodeInvalid, Message: "must be RFC 3339"}
		}
		*field.dest = &parsed
	}
	return query, nil
}

func newAdminUserResponse(user domain.AdminUser) adminUserResponse {
	var lastLoginAt *string
	if user.LastLoginAt != nil {
		formatted := formatAdminTime(*user.LastLoginAt)
		lastLoginAt = &formatted
	}
	return adminUserResponse{
		ID: user.ID, Name: user.Name, Email: user.Email, Role: user.Role, Active: user.Active,
		LastLoginAt: lastLoginAt, CreatedAt: formatAdminTime(user.CreatedAt), UpdatedAt: formatAdminTime(user.UpdatedAt),
	}
}
