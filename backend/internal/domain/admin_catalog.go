package domain

import "time"

// AdminListQuery descreve paginação, ordenação e filtros das listagens do painel.
// A ordenação é sempre conferida contra uma allowlist no service; o repositório
// nunca recebe um nome de campo vindo diretamente do cliente.
type AdminListQuery struct {
	Page       int
	PageSize   int
	Sort       string
	Order      string
	Active     *bool
	Search     string
	CategoryID string
}

// AuditLog registra uma mutação administrativa. Os snapshots guardam apenas
// campos de configuração/catálogo: nunca dados pessoais de clientes.
type AuditLog struct {
	ID        string
	ActorID   string
	Action    string
	Entity    string
	EntityID  string
	Before    map[string]any
	After     map[string]any
	CreatedAt time.Time
}

// Ações estáveis usadas no campo action de audit_logs.
const (
	AuditActionCreate = "create"
	AuditActionUpdate = "update"
)

// Entidades auditadas pelo painel.
const (
	AuditEntityCategory     = "categories"
	AuditEntityProduct      = "products"
	AuditEntityAddon        = "addons"
	AuditEntityDeliveryZone = "delivery_zones"
	AuditEntityStore        = "store_settings"
	AuditEntityWebhook      = "webhook_settings"
	AuditEntityOrder        = "orders"
	AuditEntityAdminUser    = "admin_users"
)

// AdminUserPatch contém apenas os campos permissíveis na edição de um usuário.
// Senha nunca é alterada por aqui: a V1 usa senha temporária na criação.
type AdminUserPatch struct {
	Name   *string
	Role   *string
	Active *bool
}

// Empty informa se o PATCH não contém nenhum campo.
func (p AdminUserPatch) Empty() bool {
	return p.Name == nil && p.Role == nil && p.Active == nil
}

// AuditListQuery descreve os filtros da consulta de auditoria.
type AuditListQuery struct {
	AdminListQuery
	ActorID string
	Action  string
	Entity  string
	From    *time.Time
	To      *time.Time
}

// CategoryPatch descreve uma atualização parcial de categoria.
type CategoryPatch struct {
	Name        *string
	Slug        *string
	Description *string
	SortOrder   *int
	Active      *bool
}

// Empty informa se nenhum campo foi enviado.
func (p CategoryPatch) Empty() bool {
	return p.Name == nil && p.Slug == nil && p.Description == nil && p.SortOrder == nil && p.Active == nil
}

// Apply devolve a categoria com os campos enviados substituídos.
func (p CategoryPatch) Apply(current Category) Category {
	applyString(p.Name, &current.Name)
	applyString(p.Slug, &current.Slug)
	applyString(p.Description, &current.Description)
	applyInt(p.SortOrder, &current.SortOrder)
	applyBool(p.Active, &current.Active)
	return current
}

// ProductPatch descreve uma atualização parcial de produto.
type ProductPatch struct {
	CategoryID   *string
	Name         *string
	Slug         *string
	Description  *string
	PriceCents   *int64
	ImageURL     *string
	ImageAlt     *string
	Allergens    *[]string
	Available    *bool
	Featured     *bool
	Active       *bool
	SortOrder    *int
	OptionGroups *[]OptionGroup
	AddonIDs     *[]string
}

// Empty informa se nenhum campo foi enviado.
func (p ProductPatch) Empty() bool {
	return p.CategoryID == nil && p.Name == nil && p.Slug == nil && p.Description == nil &&
		p.PriceCents == nil && p.ImageURL == nil && p.ImageAlt == nil && p.Allergens == nil &&
		p.Available == nil && p.Featured == nil && p.Active == nil && p.SortOrder == nil &&
		p.OptionGroups == nil && p.AddonIDs == nil
}

// Apply devolve o produto com os campos enviados substituídos.
func (p ProductPatch) Apply(current Product) Product {
	applyString(p.CategoryID, &current.CategoryID)
	applyString(p.Name, &current.Name)
	applyString(p.Slug, &current.Slug)
	applyString(p.Description, &current.Description)
	applyInt64(p.PriceCents, &current.PriceCents)
	applyString(p.ImageURL, &current.ImageURL)
	applyString(p.ImageAlt, &current.ImageAlt)
	applyStrings(p.Allergens, &current.Allergens)
	applyBool(p.Available, &current.Available)
	applyBool(p.Featured, &current.Featured)
	applyBool(p.Active, &current.Active)
	applyInt(p.SortOrder, &current.SortOrder)
	applyOptionGroups(p.OptionGroups, &current.OptionGroups)
	applyStrings(p.AddonIDs, &current.AddonIDs)
	return current
}

// AddonPatch descreve uma atualização parcial de adicional.
type AddonPatch struct {
	Name       *string
	Slug       *string
	PriceCents *int64
	SortOrder  *int
	Active     *bool
}

// Empty informa se nenhum campo foi enviado.
func (p AddonPatch) Empty() bool {
	return p.Name == nil && p.Slug == nil && p.PriceCents == nil && p.SortOrder == nil && p.Active == nil
}

// Apply devolve o adicional com os campos enviados substituídos.
func (p AddonPatch) Apply(current Addon) Addon {
	applyString(p.Name, &current.Name)
	applyString(p.Slug, &current.Slug)
	applyInt64(p.PriceCents, &current.PriceCents)
	applyInt(p.SortOrder, &current.SortOrder)
	applyBool(p.Active, &current.Active)
	return current
}

// DeliveryZonePatch descreve uma atualização parcial de zona de entrega.
type DeliveryZonePatch struct {
	Name      *string
	Slug      *string
	FeeCents  *int64
	SortOrder *int
	Active    *bool
}

// Empty informa se nenhum campo foi enviado.
func (p DeliveryZonePatch) Empty() bool {
	return p.Name == nil && p.Slug == nil && p.FeeCents == nil && p.SortOrder == nil && p.Active == nil
}

// Apply devolve a zona com os campos enviados substituídos.
func (p DeliveryZonePatch) Apply(current DeliveryZone) DeliveryZone {
	applyString(p.Name, &current.Name)
	applyString(p.Slug, &current.Slug)
	applyInt64(p.FeeCents, &current.FeeCents)
	applyInt(p.SortOrder, &current.SortOrder)
	applyBool(p.Active, &current.Active)
	return current
}

// StoreSettingsPatch descreve uma atualização parcial da configuração da loja.
type StoreSettingsPatch struct {
	Name         *string
	IsOpen       *bool
	OpeningHours *[]OpeningHours
	DeliveryTime *string
	WhatsApp     *string
	Instagram    *string
	Address      *string
}

// Empty informa se nenhum campo foi enviado.
func (p StoreSettingsPatch) Empty() bool {
	return p.Name == nil && p.IsOpen == nil && p.OpeningHours == nil &&
		p.DeliveryTime == nil && p.WhatsApp == nil && p.Instagram == nil && p.Address == nil
}

// Apply devolve a configuração com os campos enviados substituídos.
func (p StoreSettingsPatch) Apply(current StoreSettings) StoreSettings {
	applyString(p.Name, &current.Name)
	applyBool(p.IsOpen, &current.IsOpen)
	if p.OpeningHours != nil {
		current.OpeningHours = append([]OpeningHours(nil), (*p.OpeningHours)...)
	}
	applyString(p.DeliveryTime, &current.DeliveryTime)
	applyString(p.WhatsApp, &current.WhatsApp)
	applyString(p.Instagram, &current.Instagram)
	applyString(p.Address, &current.Address)
	return current
}

func applyString(value *string, target *string) {
	if value != nil {
		*target = *value
	}
}

func applyBool(value *bool, target *bool) {
	if value != nil {
		*target = *value
	}
}

func applyInt(value *int, target *int) {
	if value != nil {
		*target = *value
	}
}

func applyInt64(value *int64, target *int64) {
	if value != nil {
		*target = *value
	}
}

func applyStrings(value *[]string, target *[]string) {
	if value != nil {
		*target = append([]string(nil), (*value)...)
	}
}

func applyOptionGroups(value *[]OptionGroup, target *[]OptionGroup) {
	if value != nil {
		*target = append([]OptionGroup(nil), (*value)...)
	}
}
