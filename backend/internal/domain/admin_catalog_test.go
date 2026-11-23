package domain

import (
	"errors"
	"testing"
)

func validStoreSettings() StoreSettings {
	return StoreSettings{
		Name:         "Brocados Lanches",
		IsOpen:       true,
		DeliveryTime: "30-90 min",
		WhatsApp:     "+5585992358708",
		Instagram:    "@brocados",
		Address:      "Rua das Palmeiras, 123, Aquiraz-CE",
		OpeningHours: []OpeningHours{{
			Day:       "sunday",
			Intervals: []OpeningInterval{{Open: "18:00", Close: "23:30"}},
		}},
	}
}

func TestStoreSettingsValidate(t *testing.T) {
	tests := []struct {
		name      string
		mutate    func(*StoreSettings)
		wantField string
	}{
		{name: "válido", mutate: func(*StoreSettings) {}},
		{name: "sem nome", mutate: func(s *StoreSettings) { s.Name = "" }, wantField: "name"},
		{name: "sem prazo", mutate: func(s *StoreSettings) { s.DeliveryTime = "" }, wantField: "deliveryTime"},
		{name: "whatsapp vazio", mutate: func(s *StoreSettings) { s.WhatsApp = "" }, wantField: "whatsapp"},
		{name: "whatsapp sem DDI", mutate: func(s *StoreSettings) { s.WhatsApp = "85992358708" }, wantField: "whatsapp"},
		{name: "whatsapp com máscara", mutate: func(s *StoreSettings) { s.WhatsApp = "(85) 99235-8708" }, wantField: "whatsapp"},
		{name: "sem endereço", mutate: func(s *StoreSettings) { s.Address = "" }, wantField: "address"},
		{name: "dia inválido", mutate: func(s *StoreSettings) { s.OpeningHours[0].Day = "domingo" }, wantField: "openingHours[0].day"},
		{name: "dia duplicado", mutate: func(s *StoreSettings) {
			s.OpeningHours = append(s.OpeningHours, OpeningHours{Day: "sunday", Closed: true})
		}, wantField: "openingHours[1].day"},
		{name: "dia aberto sem intervalo", mutate: func(s *StoreSettings) {
			s.OpeningHours[0].Intervals = nil
		}, wantField: "openingHours[0].intervals"},
		{name: "horário malformado", mutate: func(s *StoreSettings) {
			s.OpeningHours[0].Intervals[0].Open = "18h"
		}, wantField: "openingHours[0].intervals[0].open"},
		{name: "intervalo zerado", mutate: func(s *StoreSettings) {
			s.OpeningHours[0].Intervals[0].Close = "18:00"
		}, wantField: "openingHours[0].intervals[0].close"},
		{name: "intervalos sobrepostos", mutate: func(s *StoreSettings) {
			s.OpeningHours[0].Intervals = append(s.OpeningHours[0].Intervals, OpeningInterval{Open: "19:00", Close: "23:45"})
		}, wantField: "openingHours[0].intervals[1]"},
		{name: "mais de sete dias", mutate: func(s *StoreSettings) {
			s.OpeningHours = []OpeningHours{
				{Day: "sunday", Closed: true}, {Day: "monday", Closed: true},
				{Day: "tuesday", Closed: true}, {Day: "wednesday", Closed: true},
				{Day: "thursday", Closed: true}, {Day: "friday", Closed: true},
				{Day: "saturday", Closed: true}, {Day: "monday", Closed: true},
			}
		}, wantField: "openingHours"},
		{name: "fechado ignora intervalos", mutate: func(s *StoreSettings) {
			s.OpeningHours[0].Closed = true
			s.OpeningHours[0].Intervals = nil
		}},
		{name: "virada de meia-noite", mutate: func(s *StoreSettings) {
			s.OpeningHours[0].Intervals[0] = OpeningInterval{Open: "23:00", Close: "02:00"}
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			settings := validStoreSettings()
			test.mutate(&settings)
			err := settings.Validate()
			if test.wantField == "" {
				if err != nil {
					t.Fatalf("Validate() error = %v, want nil", err)
				}
				return
			}
			var validationError *ValidationError
			if !errors.As(err, &validationError) {
				t.Fatalf("Validate() error = %v, want ValidationError", err)
			}
			if validationError.Field != test.wantField {
				t.Errorf("field = %q, want %q", validationError.Field, test.wantField)
			}
		})
	}
}

func TestAdminPatchesApplyOnlyProvidedFields(t *testing.T) {
	name := "Nova categoria"
	active := false
	category := CategoryPatch{Name: &name, Active: &active}.Apply(Category{ID: "1", Name: "Antiga", Active: true, SortOrder: 3})
	if category.Name != name || category.Active || category.ID != "1" || category.SortOrder != 3 {
		t.Errorf("category patch = %+v", category)
	}
	if !(CategoryPatch{}).Empty() {
		t.Error("empty category patch reported fields")
	}

	price := int64(2500)
	product := ProductPatch{PriceCents: &price}.Apply(Product{ID: "2", Name: "Hambúrguer", PriceCents: 2300, Available: true})
	if product.PriceCents != price || product.Name != "Hambúrguer" || !product.Available {
		t.Errorf("product patch = %+v", product)
	}
	if !(ProductPatch{}).Empty() {
		t.Error("empty product patch reported fields")
	}

	slug := "nova-zona"
	addon := AddonPatch{Slug: &slug}.Apply(Addon{ID: "3", Slug: "antiga", Active: true})
	if addon.Slug != slug || !addon.Active {
		t.Errorf("addon patch = %+v", addon)
	}
	if !(AddonPatch{}).Empty() {
		t.Error("empty addon patch reported fields")
	}

	fee := int64(700)
	zone := DeliveryZonePatch{FeeCents: &fee}.Apply(DeliveryZone{ID: "4", FeeCents: 500, Active: true})
	if zone.FeeCents != fee || !zone.Active {
		t.Errorf("zone patch = %+v", zone)
	}
	if !(DeliveryZonePatch{}).Empty() {
		t.Error("empty zone patch reported fields")
	}

	isOpen := false
	settings := StoreSettingsPatch{IsOpen: &isOpen}.Apply(validStoreSettings())
	if settings.IsOpen || settings.Name != "Brocados Lanches" {
		t.Errorf("store patch = %+v", settings)
	}
	if !(StoreSettingsPatch{}).Empty() {
		t.Error("empty store patch reported fields")
	}
}
