package domain

import (
	"testing"
	"time"
)

func TestNormalizePhoneKey(t *testing.T) {
	tests := []struct {
		name  string
		phone string
		want  string
	}{
		{"E.164", "+5585999999999", "5585999999999"},
		{"formatado com DDD", "(85) 99999-9999", "5585999999999"},
		{"dez dígitos", "8533334444", "558533334444"},
		{"vazio", "   ", ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := NormalizePhoneKey(test.phone); got != test.want {
				t.Errorf("NormalizePhoneKey(%q) = %q, want %q", test.phone, got, test.want)
			}
		})
	}
}

func TestNormalizePhoneKeyMatchesStoredProfileKey(t *testing.T) {
	const storedKey = "5585991365507"
	for _, phone := range []string{
		"(85) 991365507",
		"85991365507",
		"+5585991365507",
		"+55 85 99136-5507",
		"(85) 99136-5507",
	} {
		if got := NormalizePhoneKey(phone); got != storedKey {
			t.Errorf("NormalizePhoneKey(%q) = %q, want %q", phone, got, storedKey)
		}
		if !IsValidBrazilianPhone(phone) {
			t.Errorf("IsValidBrazilianPhone(%q) = false, want true", phone)
		}
	}
}

func TestNewCustomerProfileFromOrderKeepsOnlyConvenienceData(t *testing.T) {
	now := time.Date(2026, 11, 26, 12, 0, 0, 0, time.UTC)
	changeForCents := int64(10000)
	order := Order{
		PublicID:    "publico-opaco",
		Fulfillment: FulfillmentDelivery,
		Customer:    OrderCustomer{Name: "Maria Silva", Phone: "+5585999999999"},
		Address: &OrderAddress{
			Street: "Rua das Flores", Number: "450", District: "Tapera",
			City: "Aquiraz", State: "CE", ZipCode: "61700000",
		},
		Zone:      &OrderZone{ID: "665f1c2a9b1e4a0012a3b4f2", Name: "Centro", FeeCents: 800},
		Payment:   OrderPayment{Method: PaymentMethodCash, ChangeForCents: &changeForCents},
		Note:      "sem cebola",
		CreatedAt: now,
	}

	profile, err := NewCustomerProfileFromOrder(order, now)
	if err != nil {
		t.Fatalf("NewCustomerProfileFromOrder() error = %v", err)
	}
	if profile.PhoneKey != "5585999999999" || profile.Name != "Maria Silva" || profile.Phone != "+5585999999999" {
		t.Errorf("identity = %+v", profile)
	}
	if profile.Address == nil || profile.Address.Street != "Rua das Flores" || profile.Zone == nil || profile.Zone.Name != "Centro" {
		t.Errorf("address/zone = %+v", profile)
	}
	if profile.Fulfillment != FulfillmentDelivery || profile.PaymentMethod != PaymentMethodCash {
		t.Errorf("preferences = %+v", profile)
	}
	if profile.LastOrderPublicID != "publico-opaco" || profile.OrderCount != 1 || !profile.LastOrderAt.Equal(now) {
		t.Errorf("last order = %+v", profile)
	}
	if profile.ExpiresAt.Year() != 2028 || profile.ExpiresAt.Month() != time.November {
		t.Errorf("expiresAt = %v", profile.ExpiresAt)
	}
}

func TestNewCustomerProfileFromOrderRejectsInvalidPhone(t *testing.T) {
	if _, err := NewCustomerProfileFromOrder(Order{}, time.Now()); err == nil {
		t.Fatal("NewCustomerProfileFromOrder() error = nil, want invalid phone")
	}
}

func TestCustomerProfileValidate(t *testing.T) {
	if err := (CustomerProfile{PhoneKey: "5585999999999"}).Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if err := (CustomerProfile{}).Validate(); err == nil {
		t.Error("Validate() error = nil, want required phoneKey")
	}
	if err := (CustomerProfile{PhoneKey: "+55 85 99999-9999"}).Validate(); err == nil {
		t.Error("Validate() error = nil, want normalized phoneKey")
	}
}
