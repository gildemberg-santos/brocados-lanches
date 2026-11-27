import { WEBHOOK_EVENT_KEYS, type WebhookEventKey } from '../../types/adminWebhook'

export interface WebhookEventOption {
  key: WebhookEventKey
  label: string
  description: string
}

/** Opções do painel, na mesma ordem usada pelo domínio para validação. */
export const WEBHOOK_EVENT_OPTIONS: readonly WebhookEventOption[] = [
  { key: 'order.created', label: 'Pedido criado', description: 'Pedido gravado no servidor.' },
  { key: 'order.confirmed', label: 'Pedido confirmado', description: 'Confirmação da loja.' },
  { key: 'order.preparing', label: 'Em preparo', description: 'Preparo iniciado.' },
  { key: 'order.ready', label: 'Pronto', description: 'Pedido pronto para entrega ou retirada.' },
  { key: 'order.out_for_delivery', label: 'Saiu para entrega', description: 'Pedido a caminho.' },
  { key: 'order.picked_up', label: 'Retirado', description: 'Pedido retirado na loja.' },
  { key: 'order.delivered', label: 'Entregue', description: 'Entrega concluída.' },
  { key: 'order.cancelled', label: 'Cancelado', description: 'Pedido cancelado.' },
]

/** Remove valores desconhecidos e duplicados, preservando a ordem do catálogo. */
export function normalizeWebhookEvents(value: unknown): WebhookEventKey[] {
  if (!Array.isArray(value)) {
    return [...WEBHOOK_EVENT_KEYS]
  }
  const selected = new Set(value.filter((item): item is string => typeof item === 'string'))
  return WEBHOOK_EVENT_KEYS.filter((key) => selected.has(key))
}
