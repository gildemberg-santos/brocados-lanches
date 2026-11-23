package repository

import (
	"errors"
	"testing"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
)

func TestAdminRepositoryUserManagementIntegration(t *testing.T) {
	ctx := testContext(t)
	database := openAdminTestDatabase(t)
	repository, err := NewAdminRepository(database)
	if err != nil {
		t.Fatalf("NewAdminRepository() error = %v", err)
	}

	now := testTime()
	owner := domain.AdminUser{
		ID: "665f1c2a9b1e4a0012a3b4f9", Name: "Dono", Email: "dono@example.com",
		PasswordHash: "hash", Role: domain.AdminRoleOwner, Active: true,
		CreatedAt: now, UpdatedAt: now,
	}
	operator := domain.AdminUser{
		ID: "665f1c2a9b1e4a0012a3b4fa", Name: "Ana", Email: "ana@example.com",
		PasswordHash: "hash", Role: domain.AdminRoleOperator, Active: true,
		CreatedAt: now, UpdatedAt: now,
	}
	audit := domain.AuditLog{
		ActorID: owner.ID, Action: domain.AuditActionCreate, Entity: domain.AuditEntityAdminUser,
		EntityID: operator.ID, After: map[string]any{"name": "Ana"}, CreatedAt: now,
	}
	if err := repository.AdminCreateUser(ctx, operator, audit); err != nil {
		t.Fatalf("AdminCreateUser() error = %v", err)
	}
	if err := repository.CreateUser(ctx, owner); err != nil {
		t.Fatalf("CreateUser(owner) error = %v", err)
	}

	stored, err := repository.FindUserByID(ctx, operator.ID)
	if err != nil {
		t.Fatalf("FindUserByID() error = %v", err)
	}
	if stored.ID != operator.ID {
		t.Fatalf("stored ID = %q, want %q", stored.ID, operator.ID)
	}

	owners, err := repository.CountActiveOwners(ctx)
	if err != nil || owners != 1 {
		t.Fatalf("CountActiveOwners() = %d, error = %v", owners, err)
	}

	page, total, err := repository.ListAdminUsers(ctx, domain.AdminListQuery{
		Page: 1, PageSize: 10, Sort: "name", Order: "asc",
	})
	if err != nil || total != 2 || len(page) != 2 {
		t.Fatalf("ListAdminUsers() = %d users, total = %d, error = %v", len(page), total, err)
	}
	if page[0].Name != "Ana" {
		t.Errorf("first user = %q, want Ana", page[0].Name)
	}

	filtered, total, err := repository.ListAdminUsers(ctx, domain.AdminListQuery{
		Page: 1, PageSize: 10, Sort: "name", Order: "asc", Search: "dono",
	})
	if err != nil || total != 1 || filtered[0].Email != "dono@example.com" {
		t.Fatalf("ListAdminUsers(search) = %+v, total = %d, error = %v", filtered, total, err)
	}

	stored.Name = "Ana Maria"
	stored.UpdatedAt = now.Add(time.Minute)
	updateAudit := domain.AuditLog{
		ActorID: owner.ID, Action: domain.AuditActionUpdate, Entity: domain.AuditEntityAdminUser,
		EntityID: stored.ID, Before: map[string]any{"name": "Ana"},
		After: map[string]any{"name": "Ana Maria"}, CreatedAt: stored.UpdatedAt,
	}
	if err := repository.UpdateAdminUser(ctx, stored, now, updateAudit); err != nil {
		t.Fatalf("UpdateAdminUser() error = %v", err)
	}
	if err := repository.UpdateAdminUser(ctx, stored, now, updateAudit); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("stale UpdateAdminUser() error = %v, want ErrConflict", err)
	}

	logs, total, err := repository.ListAuditLogs(ctx, domain.AuditListQuery{
		AdminListQuery: domain.AdminListQuery{Page: 1, PageSize: 10},
		Entity:         domain.AuditEntityAdminUser, ActorID: owner.ID,
	})
	if err != nil || total != 2 || len(logs) != 2 {
		t.Fatalf("ListAuditLogs() = %d logs, total = %d, error = %v", len(logs), total, err)
	}
	if logs[0].ID == "" || logs[0].After["name"] != "Ana Maria" {
		t.Errorf("latest log = %+v", logs[0])
	}
}
