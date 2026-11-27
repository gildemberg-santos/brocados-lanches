import { describe, expect, it } from 'vitest'

import { WEBHOOK_EVENT_OPTIONS, normalizeWebhookEvents } from './webhookEvents'

describe('eventos selecionáveis do webhook', () => {
  it('lista todos os eventos mapeados na ordem do contrato', () => {
    expect(WEBHOOK_EVENT_OPTIONS.map((option) => option.key)).toEqual([
      'order.created',
      'order.confirmed',
      'order.preparing',
      'order.ready',
      'order.out_for_delivery',
      'order.picked_up',
      'order.delivered',
      'order.cancelled',
    ])
    expect(WEBHOOK_EVENT_OPTIONS.every((option) => option.label.length > 0)).toBe(true)
  })

  it('mantém apenas chaves conhecidas, sem duplicatas', () => {
    expect(
      normalizeWebhookEvents([
        'order.cancelled',
        'order.created',
        'order.cancelled',
        'order.futuro',
      ]),
    ).toEqual(['order.created', 'order.cancelled'])
  })

  it('respeita seleção vazia e configuração ausente', () => {
    expect(normalizeWebhookEvents([])).toEqual([])
    expect(normalizeWebhookEvents(undefined)).toHaveLength(8)
  })
})
