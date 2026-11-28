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
  isQuoteOutdatedError,
  isValidBrazilianPhone,
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

  it('keeps the registered phone stable when typing, masking and reading it back from the database', () => {
    const storedPhone = '+5585991365507'

    expect(formatPhone(storedPhone)).toBe('(85) 99136-5507')

    for (const typed of ['(85) 991365507', '85991365507', '+55 85 99136-5507']) {
      const masked = formatPhone(typed)
      expect(masked).toBe('(85) 99136-5507')
      expect(normalizePhone(masked)).toBe(storedPhone)
      expect(isValidBrazilianPhone(typed)).toBe(true)
    }
  })

  it('rejects phone numbers that cannot reach a Brazilian WhatsApp account', () => {
    expect(isValidBrazilianPhone('(85) 99999-9999')).toBe(true)
    expect(isValidBrazilianPhone('+558532321234')).toBe(true)
    expect(isValidBrazilianPhone('+5555859913655')).toBe(false)
    expect(isValidBrazilianPhone('+55 85 89999-9999')).toBe(false)
    expect(isValidBrazilianPhone('+55 00 99999-9999')).toBe(false)
    expect(isValidBrazilianPhone('123')).toBe(false)
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

    const request = buildCreateOrderRequest([item], form, 6200)

    expect(request.customer).toEqual({ name: 'Maria Silva', phone: '+5585999999999' })
    expect(request.address?.city).toBe('Aquiraz')
    expect(request.payment).toEqual({ method: 'cash', changeForCents: 10000 })
    expect(request.expectedTotalCents).toBe(6200)
    expect(request.note).toBe('Deixar na portaria')
  })

  it('builds a request even when the profile omitted optional address fields', () => {
    const form = {
      ...initialCheckoutForm,
      zoneId: 'zona-1',
      customerName: 'Maria Silva',
      phone: '(85) 991365507',
      street: 'Rua Manuel Alves da Cunha',
      number: '60',
      complement: undefined as unknown as string,
      district: 'Tabajarra',
      zipCode: '61700000',
    }

    const request = buildCreateOrderRequest([item], form, 6200)

    expect(request.address?.complement).toBe('')
  })

  it('omits address and fee-related fields on pickup', () => {
    const form = { ...initialCheckoutForm, fulfillment: 'pickup' as const }

    const request = buildCreateOrderRequest([item], form, 6200)

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

  it('blocks checkout when the phone has a duplicated country code', () => {
    const form = {
      ...initialCheckoutForm,
      zoneId: 'zona-1',
      customerName: 'Maria Silva',
      phone: '+5555859913655',
      street: 'Rua das Flores',
      number: '450',
      district: 'Tapera',
      zipCode: '61700000',
      changeFor: '100,00',
    }

    expect(validateCheckoutForm(form, 6200).phone).toBeDefined()
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
    expect(
      checkoutApiMessage(
        new ApiError(422, 'VALIDATION_ERROR', 'pix is currently unavailable', [
          { field: 'payment.method', code: 'INVALID' },
        ]),
      ),
    ).toContain('Pix não está disponível')
  })

  it('identifies only the outdated quote conflict', () => {
    const outdated = new ApiError(409, 'CONFLICT', 'total mudou', [
      { field: 'expectedTotalCents', code: 'OUTDATED' },
    ])

    expect(isQuoteOutdatedError(outdated)).toBe(true)
    expect(isQuoteOutdatedError(new ApiError(409, 'CONFLICT', 'replay'))).toBe(false)
    expect(isQuoteOutdatedError(new Error('falha'))).toBe(false)
  })

  it('maps order statuses to Portuguese labels', () => {
    expect(orderStatusLabel('pending')).toBe('Aguardando confirmação')
    expect(orderStatusLabel('out_for_delivery')).toBe('Saiu para entrega')
    expect(orderStatusLabel('novo')).toBe('novo')
  })
})
