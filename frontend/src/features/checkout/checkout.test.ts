import { describe, expect, it } from 'vitest'

import type { CartItem } from '../../types/cart'
import { ApiError } from '../../services/http'
import {
  buildCreateOrderRequest,
  buildQuoteRequest,
  checkoutApiMessage,
  createIdempotencyKey,
  formatPhone,
  initialCheckoutForm,
  normalizePhone,
  orderStatusLabel,
  parseCentsInput,
  validateCheckoutForm,
} from './checkout'

const item: CartItem = {
  key: 'produto|option-1||',
  productId: 'produto-1',
  optionIds: ['option-1'],
  addonIds: ['addon-1'],
  quantity: 2,
  note: 'sem cebola',
}

describe('checkout', () => {
  it('builds a quote request with IDs and choices only', () => {
    const form = { ...initialCheckoutForm, zoneId: 'zona-1' }

    expect(buildQuoteRequest([item], form)).toEqual({
      items: [
        {
          productId: 'produto-1',
          quantity: 2,
          optionIds: ['option-1'],
          addonIds: ['addon-1'],
          note: 'sem cebola',
        },
      ],
      fulfillment: 'delivery',
      zoneId: 'zona-1',
    })
  })

  it('converts money text to integer cents without floats', () => {
    expect(parseCentsInput('100,00')).toBe(10000)
    expect(parseCentsInput('100.00')).toBe(10000)
    expect(parseCentsInput('100')).toBe(10000)
    expect(parseCentsInput('')).toBeNull()
    expect(parseCentsInput('abc')).toBeNull()
  })

  it('normalizes and formats Brazilian phone numbers', () => {
    expect(normalizePhone('(85) 99999-9999')).toBe('+5585999999999')
    expect(normalizePhone('5585999999999')).toBe('+5585999999999')
    expect(formatPhone('85999999999')).toBe('(85) 99999-9999')
    expect(formatPhone('+5585999999999')).toBe('(85) 99999-9999')
    expect(formatPhone('55 85 99999-9999')).toBe('(85) 99999-9999')
    expect(normalizePhone(formatPhone('+5585999999999'))).toBe('+5585999999999')
  })

  it('builds a create request with customer, address and payment', () => {
    const form = {
      ...initialCheckoutForm,
      zoneId: 'zona-1',
      customerName: 'Maria Silva',
      phone: '(85) 99999-9999',
      street: 'Rua das Flores',
      number: '450',
      district: 'Tapera',
      zipCode: '61700000',
      changeFor: '100,00',
      note: 'Deixar na portaria',
    }

    const request = buildCreateOrderRequest([item], form)

    expect(request.customer).toEqual({ name: 'Maria Silva', phone: '+5585999999999' })
    expect(request.address?.city).toBe('Aquiraz')
    expect(request.payment).toEqual({ method: 'cash', changeForCents: 10000 })
    expect(request.note).toBe('Deixar na portaria')
  })

  it('omits address and fee-related fields on pickup', () => {
    const form = { ...initialCheckoutForm, fulfillment: 'pickup' as const }

    const request = buildCreateOrderRequest([item], form)

    expect(request.zoneId).toBeUndefined()
    expect(request.address).toBeUndefined()
  })

  it('validates required fields and change for cash', () => {
    const errors = validateCheckoutForm(initialCheckoutForm, 6200)

    expect(errors.customerName).toBeDefined()
    expect(errors.phone).toBeDefined()
    expect(errors.zoneId).toBeDefined()
    expect(errors.changeFor).toBeDefined()

    const complete = {
      ...initialCheckoutForm,
      zoneId: 'zona-1',
      customerName: 'Maria Silva',
      phone: '(85) 99999-9999',
      street: 'Rua das Flores',
      number: '450',
      district: 'Tapera',
      zipCode: '61700000',
      changeFor: '100,00',
    }
    expect(validateCheckoutForm(complete, 6200)).toEqual({})
  })

  it('validates fields for pickup without address', () => {
    const complete = {
      ...initialCheckoutForm,
      fulfillment: 'pickup' as const,
      customerName: 'Maria Silva',
      phone: '(85) 99999-9999',
      paymentMethod: 'pix' as const,
    }

    expect(validateCheckoutForm(complete, 6200)).toEqual({})
  })

  it('generates unique idempotency keys', () => {
    expect(createIdempotencyKey()).not.toBe(createIdempotencyKey())
    expect(createIdempotencyKey().length).toBeGreaterThanOrEqual(16)
  })

  it('maps API codes to friendly messages', () => {
    expect(checkoutApiMessage(new ApiError(409, 'STORE_CLOSED', 'closed'))).toContain('fechada')
    expect(checkoutApiMessage(new ApiError(404, 'NOT_FOUND', 'dados de terceiros'))).toBe(
      'Pedido não encontrado ou código de acompanhamento inválido.',
    )
    expect(checkoutApiMessage(new Error('unknown'))).toContain('Seus dados foram mantidos')
  })

  it('maps order statuses to Portuguese labels', () => {
    expect(orderStatusLabel('pending')).toBe('Aguardando confirmação')
    expect(orderStatusLabel('out_for_delivery')).toBe('Saiu para entrega')
    expect(orderStatusLabel('novo')).toBe('novo')
  })
})
