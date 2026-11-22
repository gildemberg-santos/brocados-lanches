package domain

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

// MaxOrderCancelReasonLength limita o texto livre do cancelamento. O motivo é
// operacional e nunca deve conter dados do cliente.
const MaxOrderCancelReasonLength = 280

// OrderStatusChangeRequest é a entrada da transição de status. O cliente envia
// apenas o destino e, no cancelamento, o motivo.
type OrderStatusChangeRequest struct {
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
}

// Validate confere o destino e a obrigatoriedade do motivo de cancelamento.
func (r OrderStatusChangeRequest) Validate() error {
	if !IsKnownOrderStatus(r.Status) {
		return &ValidationError{Field: "status", Code: CodeInvalid, Message: "must be a supported status"}
	}
	if r.Status != OrderStatusCancelled {
		return nil
	}
	if strings.TrimSpace(r.Reason) == "" {
		return &ValidationError{Field: "reason", Code: CodeRequired, Message: "is required when cancelling"}
	}
	if utf8.RuneCountInString(r.Reason) > MaxOrderCancelReasonLength {
		return &ValidationError{
			Field:   "reason",
			Code:    CodeTooLong,
			Message: fmt.Sprintf("must have at most %d characters", MaxOrderCancelReasonLength),
		}
	}
	return nil
}

// IsKnownOrderStatus informa se o valor pertence ao enum de status da V1.
func IsKnownOrderStatus(status string) bool {
	switch status {
	case OrderStatusPending, OrderStatusConfirmed, OrderStatusPreparing,
		OrderStatusReady, OrderStatusOutForDelivery, OrderStatusPickedUp,
		OrderStatusDelivered, OrderStatusCancelled:
		return true
	default:
		return false
	}
}

// OrderStatusTransitionError sinaliza uma transição fora da matriz da V1.
// Implementa Unwrap para que o HTTP mapeie o erro para 409 CONFLICT.
type OrderStatusTransitionError struct {
	From string
	To   string
}

func (e *OrderStatusTransitionError) Error() string {
	return fmt.Sprintf("invalid order status transition from %q to %q", e.From, e.To)
}

// Unwrap permite errors.Is(err, ErrConflict).
func (e *OrderStatusTransitionError) Unwrap() error { return ErrConflict }

// ValidateOrderStatusTransition aplica a matriz permitida. O desfecho de
// `ready` depende da forma de entrega: entrega vai para out_for_delivery e
// retirada vai para picked_up. Estados finais não transitam.
func ValidateOrderStatusTransition(from, to string, fulfillment Fulfillment) error {
	var allowed bool
	switch from {
	case OrderStatusPending:
		allowed = to == OrderStatusConfirmed || to == OrderStatusCancelled
	case OrderStatusConfirmed:
		allowed = to == OrderStatusPreparing || to == OrderStatusCancelled
	case OrderStatusPreparing:
		allowed = to == OrderStatusReady || to == OrderStatusCancelled
	case OrderStatusReady:
		allowed = (fulfillment == FulfillmentDelivery && to == OrderStatusOutForDelivery) ||
			(fulfillment == FulfillmentPickup && to == OrderStatusPickedUp)
	case OrderStatusOutForDelivery:
		allowed = to == OrderStatusDelivered
	default:
		allowed = false
	}
	if !allowed {
		return &OrderStatusTransitionError{From: from, To: to}
	}
	return nil
}

// ApplyOrderStatusChange valida a transição e devolve uma cópia do pedido com
// o novo status, os timestamps de negócio e o histórico atualizado. O pedido
// recebido não é alterado.
func ApplyOrderStatusChange(current Order, request OrderStatusChangeRequest, actorID string, now time.Time) (Order, error) {
	if err := request.Validate(); err != nil {
		return Order{}, err
	}
	if err := ValidateOrderStatusTransition(current.Status, request.Status, current.Fulfillment); err != nil {
		return Order{}, err
	}

	at := now.UTC()
	updated := current
	updated.Status = request.Status
	updated.UpdatedAt = at
	updated.StatusHistory = append(append([]OrderStatusChange(nil), current.StatusHistory...), OrderStatusChange{
		Status:  request.Status,
		At:      at,
		ActorID: actorID,
	})

	switch request.Status {
	case OrderStatusConfirmed:
		updated.ConfirmedAt = &at
	case OrderStatusCancelled:
		updated.CancelledAt = &at
		updated.CancelReason = strings.TrimSpace(request.Reason)
	}
	return updated, nil
}
