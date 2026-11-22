import { ApiError } from '../../services/http'
import type { AdminOrderStatusRequest } from '../../types/order'

export const ORDER_STATUSES = [
  'pending',
  'confirmed',
  'preparing',
  'ready',
  'out_for_delivery',
  'picked_up',
  'delivered',
  'cancelled',
] as const

export type OrderStatus = (typeof ORDER_STATUSES)[number]
export type OrderFulfillment = 'delivery' | 'pickup'

export interface OrderStatusChange {
  status: OrderStatus
  at: string
  actorId?: string
}

export const ORDER_STATUS_LABELS: Record<OrderStatus, string> = {
  pending: 'Pendente',
  confirmed: 'Confirmado',
  preparing: 'Em preparo',
  ready: 'Pronto',
  out_for_delivery: 'Saiu para entrega',
  picked_up: 'Retirado',
  delivered: 'Entregue',
  cancelled: 'Cancelado',
}

export const ORDER_ACTION_LABELS: Record<OrderStatus, string> = {
  pending: 'Voltar para pendente',
  confirmed: 'Confirmar pedido',
  preparing: 'Iniciar preparo',
  ready: 'Marcar como pronto',
  out_for_delivery: 'Marcar como saído para entrega',
  picked_up: 'Marcar como retirado',
  delivered: 'Marcar como entregue',
  cancelled: 'Cancelar pedido',
}

export const FULFILLMENT_LABELS: Record<OrderFulfillment, string> = {
  delivery: 'Entrega',
  pickup: 'Retirada',
}

const TRANSITIONS: Record<OrderStatus, Partial<Record<OrderFulfillment, OrderStatus[]>>> = {
  pending: { delivery: ['confirmed', 'cancelled'], pickup: ['confirmed', 'cancelled'] },
  confirmed: { delivery: ['preparing', 'cancelled'], pickup: ['preparing', 'cancelled'] },
  preparing: { delivery: ['ready', 'cancelled'], pickup: ['ready', 'cancelled'] },
  ready: { delivery: ['out_for_delivery'], pickup: ['picked_up'] },
  out_for_delivery: { delivery: ['delivered'] },
  picked_up: {},
  delivered: {},
  cancelled: {},
}

export function isOrderStatus(value: string): value is OrderStatus {
  return (ORDER_STATUSES as readonly string[]).includes(value)
}

export function isOrderFulfillment(value: string): value is OrderFulfillment {
  return value === 'delivery' || value === 'pickup'
}

/** Ações disponíveis para o status atual e a forma de entrega do pedido. */
export function availableTransitions(status: string, fulfillment: string): OrderStatus[] {
  if (!isOrderStatus(status) || !isOrderFulfillment(fulfillment)) {
    return []
  }
  return TRANSITIONS[status][fulfillment] ?? []
}

export function isFinalStatus(status: string): boolean {
  return isOrderStatus(status) && availableTransitions(status, 'delivery').length === 0
}

export function isCancellationAllowed(status: string): boolean {
  return availableTransitions(status, 'delivery').includes('cancelled')
}

/** Erro de API transformado em mensagem operacional, sem expor detalhes internos. */
export function adminOrderErrorMessage(error: unknown, action: 'load' | 'status'): string {
  const fallback =
    action === 'load'
      ? 'Não foi possível carregar os pedidos. Tente novamente.'
      : 'Não foi possível atualizar o pedido. Tente novamente.'

  if (!(error instanceof ApiError)) {
    return fallback
  }
  if (error.status === 401) {
    return 'Sua sessão expirou. Entre novamente no painel.'
  }
  if (error.status === 403) {
    return 'Você não tem permissão para esta ação.'
  }
  if (error.status === 409) {
    return 'O pedido mudou de status enquanto você operava. Atualize a tela e tente novamente.'
  }
  if (error.code === 'VALIDATION_ERROR') {
    return action === 'status'
      ? 'Revise a ação escolhida e tente novamente.'
      : 'Revise os filtros informados e tente novamente.'
  }
  // Erros internos nunca são exibidos ao operador.
  return fallback
}

export type StatusChangeValidation =
  { valid: true; request: AdminOrderStatusRequest } | { valid: false; error: string }

/** Valida a intenção antes de chamar a API, que continua sendo autoritativa. */
export function validateStatusChange(status: string, reason: string): StatusChangeValidation {
  if (!isOrderStatus(status)) {
    return { valid: false, error: 'Escolha uma ação válida para este pedido.' }
  }

  const trimmedReason = reason.trim()
  if (status === 'cancelled' && !trimmedReason) {
    return { valid: false, error: 'Informe o motivo do cancelamento.' }
  }
  if (trimmedReason.length > 280) {
    return { valid: false, error: 'O motivo do cancelamento deve ter até 280 caracteres.' }
  }

  return {
    valid: true,
    request: { status, ...(status === 'cancelled' ? { reason: trimmedReason } : {}) },
  }
}
