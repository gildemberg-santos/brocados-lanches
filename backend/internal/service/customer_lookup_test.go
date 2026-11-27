package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/brocados-lanches/api/internal/domain"
)

type fakeCustomerLookupStore struct {
	profile domain.CustomerProfile
	found   bool
	err     error
	keys    []string
}

func (f *fakeCustomerLookupStore) Upsert(context.Context, domain.CustomerProfile) error {
	return nil
}

func (f *fakeCustomerLookupStore) OptOut(context.Context, string, time.Time) error {
	return nil
}

func (f *fakeCustomerLookupStore) FindByPhoneKey(_ context.Context, phoneKey string) (domain.CustomerProfile, bool, error) {
	f.keys = append(f.keys, phoneKey)
	return f.profile, f.found, f.err
}

func TestCustomerServiceLookup(t *testing.T) {
	store := &fakeCustomerLookupStore{
		found: true,
		profile: domain.CustomerProfile{
			PhoneKey:      "5585999999999",
			Name:          "Maria Silva",
			Fulfillment:   domain.FulfillmentDelivery,
			PaymentMethod: "pix",
		},
	}
	service, err := NewCustomerService(store)
	if err != nil {
		t.Fatalf("NewCustomerService() error = %v", err)
	}

	profile, err := service.Lookup(context.Background(), "(85) 99999-9999")
	if err != nil {
		t.Fatalf("Lookup() error = %v", err)
	}
	if profile == nil || profile.Name != "Maria Silva" {
		t.Fatalf("Lookup() profile = %+v", profile)
	}
	if len(store.keys) != 1 || store.keys[0] != "5585999999999" {
		t.Errorf("normalizados = %v", store.keys)
	}
}

func TestCustomerServiceLookupNotFoundAndOptOut(t *testing.T) {
	cases := []struct {
		name  string
		found bool
		opted bool
	}{
		{name: "sem perfil", found: false},
		{name: "opt-out", found: true, opted: true},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			store := &fakeCustomerLookupStore{
				found:   testCase.found,
				profile: domain.CustomerProfile{OptedOut: testCase.opted},
			}
			service, err := NewCustomerService(store)
			if err != nil {
				t.Fatalf("NewCustomerService() error = %v", err)
			}

			profile, err := service.Lookup(context.Background(), "+5585999999999")
			if err != nil {
				t.Fatalf("Lookup() error = %v", err)
			}
			if profile != nil {
				t.Errorf("Lookup() profile = %+v, want nil", profile)
			}
		})
	}
}

func TestCustomerServiceLookupValidation(t *testing.T) {
	service, err := NewCustomerService(&fakeCustomerLookupStore{})
	if err != nil {
		t.Fatalf("NewCustomerService() error = %v", err)
	}

	for _, phone := range []string{"", "123", "859999", "+551199999", "+5555859913655", strings.Repeat("9", domain.MaxPhoneLength+1)} {
		_, err := service.Lookup(context.Background(), phone)
		if !errors.Is(err, domain.ErrValidation) {
			t.Errorf("Lookup(%q) error = %v, want validation", phone, err)
		}
	}
}

func TestCustomerServiceLookupRepositoryError(t *testing.T) {
	want := errors.New("mongo indisponível")
	service, err := NewCustomerService(&fakeCustomerLookupStore{found: true, err: want})
	if err != nil {
		t.Fatalf("NewCustomerService() error = %v", err)
	}

	_, err = service.Lookup(context.Background(), "+5585999999999")
	if !errors.Is(err, want) {
		t.Errorf("Lookup() error = %v, want %v", err, want)
	}
}

func TestNewCustomerServiceRequiresStore(t *testing.T) {
	if _, err := NewCustomerService(nil); err == nil {
		t.Error("NewCustomerService(nil) error = nil")
	}
}
