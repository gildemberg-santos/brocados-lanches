import { describe, expect, it } from 'vitest'

import { sanitizePath, sanitizeRecord, sanitizeSentryEvent, sanitizeText } from './sentryScrub'

describe('sentryScrub', () => {
  it('remove dados pessoais e segredos de textos', () => {
    const result = sanitizeText(
      'pedido para Maria dono@teste.local +5585999999999 CEP 61700-000 ' +
        'https://wa.me/5585999999999?text=oi Authorization: Bearer segredo Idempotency-Key: abc123',
    )

    for (const forbidden of [
      'dono@teste.local',
      '+5585999999999',
      '61700-000',
      'wa.me',
      'Bearer segredo',
      'abc123',
    ]) {
      expect(result).not.toContain(forbidden)
    }
    expect(result).toContain('pedido para Maria')
  })

  it('reduz caminhos a baixa cardinalidade', () => {
    expect(sanitizePath('/api/v1/orders/507f1f77bcf86cd799439011?phone=5585999999999')).toBe(
      '/api/v1/orders/:id',
    )
    expect(sanitizePath('/checkout#telefone=5585999999999')).toBe('/checkout')
    expect(sanitizePath('/api/v1/orders/history')).toBe('/api/v1/orders/history')
  })

  it('descarta chaves sensíveis em objetos aninhados', () => {
    const result = sanitizeRecord({
      method: 'POST',
      status: 500,
      customer: { phone: '+5585999999999', name: 'Maria' },
      headers: { Authorization: 'Bearer segredo' },
      details: [{ zipCode: '61700-000', field: 'phone' }],
    })

    expect(result).toEqual({
      method: 'POST',
      status: 500,
      customer: { name: 'Maria' },
      details: [{ field: 'phone' }],
    })
  })

  it('sanitiza o evento antes do envio', () => {
    const event = {
      user: { email: 'dono@teste.local' },
      request: { url: '/checkout?phone=5585999999999', headers: { Cookie: 'abc' } },
      message: 'erro em +5585999999999',
      transaction: '/api/v1/orders/507f1f77bcf86cd799439011?phone=1',
      tags: { route: 'checkout', phone: '+5585999999999' },
      extra: { body: '{"phone":"+5585999999999"}', status: 500 },
      modules: { app: 'v1', token: 'segredo' },
      fingerprint: ['{{ default }}', '+5585999999999'],
      contexts: { runtime: { version: 'v1', authorization: 'Bearer segredo' } },
      breadcrumbs: [
        { message: 'wa.me/5585999999999', data: { phone: '5585999999999', status: 500 } },
      ],
      exception: {
        values: [
          {
            value: 'erro para dono@teste.local',
            stacktrace: {
              frames: [
                {
                  filename: 'https://cdn.exemplo.com/app.js?phone=5585999999999',
                  function: 'enviar +5585999999999',
                  vars: { phone: '+5585999999999', status: 500 },
                },
              ],
            },
          },
        ],
      },
    }

    const result = sanitizeSentryEvent(event)
    const rendered = JSON.stringify(result)

    expect(result).toBe(event)
    expect(result.user).toBeUndefined()
    expect(result.request).toBeUndefined()
    for (const forbidden of [
      'dono@teste.local',
      '+5585999999999',
      '5585999999999',
      'wa.me',
      'Bearer segredo',
      'segredo',
      '507f1f77bcf86cd799439011',
    ]) {
      expect(rendered).not.toContain(forbidden)
    }
    expect(rendered).toContain('checkout')
    expect(rendered).toContain('v1')
    expect(rendered).toContain('500')
    expect(rendered).toContain('app.js')
  })
})
