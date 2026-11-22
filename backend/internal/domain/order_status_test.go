package domain

import (
	"errors"
	"testing"
	"time"
)

func TestValidateOrderStatusTransition(t *testing.T) {
	tests := []struct {
		name        string
		from        string
		to          string
		fulfillment Fulfillment
		wantAllowed bool
	}{
		{"pending para confirmed", OrderStatusPending, OrderStatusConfirmed, FulfillmentDelivery, true},
		{"pending para cancelled", OrderStatusPending, OrderStatusCancelled, FulfillmentDelivery, true},
		{"pending para preparing", OrderStatusPending, OrderStatusPreparing, FulfillmentDelivery, false},
		{"confirmed para preparing", OrderStatusConfirmed, OrderStatusPreparing, FulfillmentDelivery, true},
		{"confirmed para cancelled", OrderStatusConfirmed, OrderStatusCancelled, FulfillmentDelivery, true},
		{"confirmed para ready", OrderStatusConfirmed, OrderStatusReady, FulfillmentDelivery, false},
		{"preparing para ready", OrderStatusPreparing, OrderStatusReady, FulfillmentDelivery, true},
		{"preparing para cancelled", OrderStatusPreparing, OrderStatusCancelled, FulfillmentPickup, true},
		{"ready entrega para out_for_delivery", OrderStatusReady, OrderStatusOutForDelivery, FulfillmentDelivery, true},
		{"ready entrega para picked_up", OrderStatusReady, OrderStatusPickedUp, FulfillmentDelivery, false},
		{"ready retirada para picked_up", OrderStatusReady, OrderStatusPickedUp, FulfillmentPickup, true},
		{"ready retirada para out_for_delivery", OrderStatusReady, OrderStatusOutForDelivery, FulfillmentPickup, false},
		{"ready para cancelled", OrderStatusReady, OrderStatusCancelled, FulfillmentDelivery, false},
		{"out_for_delivery para delivered", OrderStatusOutForDelivery, OrderStatusDelivered, FulfillmentDelivery, true},
		{"out_for_delivery para cancelled", OrderStatusOutForDelivery, OrderStatusCancelled, FulfillmentDelivery, false},
		{"delivered é final", OrderStatusDelivered, OrderStatusConfirmed, FulfillmentDelivery, false},
		{"picked_up é final", OrderStatusPickedUp, OrderStatusDelivered, FulfillmentPickup, false},
		{"cancelled é final", OrderStatusCancelled, OrderStatusConfirmed, FulfillmentDelivery, false},
		{"mesmo status", OrderStatusPending, OrderStatusPending, FulfillmentDelivery, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateOrderStatusTransition(test.from, test.to, test.fulfillment)
			if test.wantAllowed {
				if err != nil {
					t.Fatalf("transition error = %v, want allowed", err)
				}
				return
			}
			if err == nil {
				t.Fatal("transition error = nil, want conflict")
			}
			if !errors.Is(err, ErrConflict) {
				t.Errorf("error = %v, want ErrConflict", err)
			}
			var transitionError *OrderStatusTransitionError
			if !errors.As(err, &transitionError) {
				t.Errorf("error = %v, want OrderStatusTransitionError", err)
			}
		})
	}
}

func TestOrderStatusChangeRequestValidate(t *testing.T) {
	tests := []struct {
		name      string
		request   OrderStatusChangeRequest
		wantField string
	}{
		{"status válido", OrderStatusChangeRequest{Status: OrderStatusConfirmed}, ""},
		{"cancelamento com motivo", OrderStatusChangeRequest{Status: OrderStatusCancelled, Reason: " cliente desistiu "}, ""},
		{"status vazio", OrderStatusChangeRequest{}, "status"},
		{"status desconhecido", OrderStatusChangeRequest{Status: "cozendo"}, "status"},
		{"cancelamento sem motivo", OrderStatusChangeRequest{Status: OrderStatusCancelled}, "reason"},
		{"cancelamento com motivo em branco", OrderStatusChangeRequest{Status: OrderStatusCancelled, Reason: "   "}, "reason"},
		{"motivo longo", OrderStatusChangeRequest{Status: OrderStatusCancelled, Reason: longReason()}, "reason"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.request.Validate()
			if test.wantField == "" {
				if err != nil {
					t.Fatalf("Validate() error = %v, want nil", err)
				}
				return
			}
			var validationError *ValidationError
			if !errors.As(err, &validationError) || validationError.Field != test.wantField {
				t.Fatalf("Validate() error = %v, want field %q", err, test.wantField)
			}
		})
	}
}

func TestApplyOrderStatusChange(t *testing.T) {
	now := time.Date(2026, 11, 22, 18, 30, 0, 0, time.FixedZone("teste", -3*3600))
	current := Order{
		ID:          "665f1c2a9b1e4a0012a3b4f1",
		Status:      OrderStatusPending,
		Fulfillment: FulfillmentDelivery,
		CreatedAt:   now.Add(-time.Hour),
		UpdatedAt:   now.Add(-time.Hour),
		StatusHistory: []OrderStatusChange{{
			Status: OrderStatusPending, At: now.Add(-time.Hour),
		}},
	}

	confirmed, err := ApplyOrderStatusChange(current, OrderStatusChangeRequest{Status: OrderStatusConfirmed}, "665f1c2a9b1e4a0012a3b4f9", now)
	if err != nil {
		t.Fatalf("ApplyOrderStatusChange() error = %v", err)
	}
	if confirmed.Status != OrderStatusConfirmed || confirmed.ConfirmedAt == nil || !confirmed.ConfirmedAt.Equal(now.UTC()) {
		t.Fatalf("confirmed order = %+v", confirmed)
	}
	if len(confirmed.StatusHistory) != 2 || confirmed.StatusHistory[1].ActorID != "665f1c2a9b1e4a0012a3b4f9" {
		t.Errorf("status history = %+v", confirmed.StatusHistory)
	}
	if !confirmed.UpdatedAt.Equal(now.UTC()) || len(current.StatusHistory) != 1 || current.Status != OrderStatusPending {
		t.Error("ApplyOrderStatusChange mutated the original order")
	}

	cancelled, err := ApplyOrderStatusChange(current, OrderStatusChangeRequest{
		Status: OrderStatusCancelled, Reason: "  cliente desistiu  ",
	}, "665f1c2a9b1e4a0012a3b4f9", now)
	if err != nil {
		t.Fatalf("cancel error = %v", err)
	}
	if cancelled.CancelledAt == nil || !cancelled.CancelledAt.Equal(now.UTC()) || cancelled.CancelReason != "cliente desistiu" {
		t.Errorf("cancelled order = %+v", cancelled)
	}

	if _, err := ApplyOrderStatusChange(current, OrderStatusChangeRequest{Status: OrderStatusPreparing}, "actor", now); !errors.Is(err, ErrConflict) {
		t.Errorf("invalid transition error = %v, want ErrConflict", err)
	}
	if _, err := ApplyOrderStatusChange(current, OrderStatusChangeRequest{Status: OrderStatusCancelled}, "actor", now); err == nil {
		t.Error("missing reason error = nil, want validation error")
	}
}

func TestAdminOrderListItemMasksCustomerData(t *testing.T) {
	item := NewAdminOrderListItem(Order{
		ID: "665f1c2a9b1e4a0012a3b4f1", Number: 42, PublicID: "opaco",
		Status: OrderStatusPending, Fulfillment: FulfillmentDelivery,
		Customer:   OrderCustomer{Name: "Maria Silva", Phone: "+5585992358708"},
		Items:      []OrderItem{{}, {}},
		TotalCents: 5490, SubtotalCents: 2300, DeliveryFeeCents: 500,
		CreatedAt: testOrderTime(), UpdatedAt: testOrderTime(),
	})
	if item.CustomerName != "M*** S***" {
		t.Errorf("customer name = %q, want masked", item.CustomerName)
	}
	if item.Phone != "****8708" {
		t.Errorf("phone = %q, want masked", item.Phone)
	}
	if item.ItemCount != 2 || item.TotalCents != 5490 || item.Number != 42 {
		t.Errorf("item = %+v", item)
	}
}

func TestMaskFunctions(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{"telefone E.164", "+5585992358708", "****8708"},
		{"telefone formatado", "(85) 99235-8708", "****8708"},
		{"telefone curto", "123", "***"},
		{"vazio", "", ""},
	}
	for _, test := range tests {
		if got := MaskPhone(test.value); got != test.want {
			t.Errorf("MaskPhone(%q) = %q, want %q", test.value, got, test.want)
		}
	}

	nameTests := []struct {
		value string
		want  string
	}{
		{"Maria Silva", "M*** S***"},
		{"  João   Pereira  ", "J*** P***"},
		{"", ""},
	}
	for _, test := range nameTests {
		if got := MaskName(test.value); got != test.want {
			t.Errorf("MaskName(%q) = %q, want %q", test.value, got, test.want)
		}
	}
}

func longReason() string {
	result := make([]byte, MaxOrderCancelReasonLength+1)
	for index := range result {
		result[index] = 'a'
	}
	return string(result)
}

func testOrderTime() time.Time {
	return time.Date(2026, 11, 22, 12, 0, 0, 0, time.UTC)
}
