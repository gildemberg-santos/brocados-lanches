/** Chaves estáveis dos eventos que podem ser habilitados no painel. */
export const WEBHOOK_EVENT_KEYS = [
  'order.created',
  'order.confirmed',
  'order.preparing',
  'order.ready',
  'order.out_for_delivery',
  'order.picked_up',
  'order.delivered',
  'order.cancelled',
] as const

export type WebhookEventKey = (typeof WEBHOOK_EVENT_KEYS)[number]

/** Configuração da integração de webhook de notificações de pedidos. */
export interface AdminWebhookSettings {
  url: string
  active: boolean
  events: WebhookEventKey[]
  createdAt: string
  updatedAt: string
}

export interface AdminWebhookInput {
  url: string
  active: boolean
  events: WebhookEventKey[]
}
