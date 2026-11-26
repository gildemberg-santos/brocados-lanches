import { beforeEach, describe, expect, it, vi } from 'vitest'

import { initialCheckoutForm } from './checkout'
import type { CheckoutForm } from './checkout'
import {
  CHECKOUT_PREFERENCES_KEY,
  buildCheckoutPreferences,
  clearCheckoutPreferences,
  hasCheckoutPreferences,
  loadCheckoutPreferences,
  saveCheckoutPreferences,
} from './preferences'

function filledForm(): CheckoutForm {
  return {
    ...initialCheckoutForm,
    fulfillment: 'delivery',
    zoneId: 'zona-1',
    customerName: 'Maria Silva',
    phone: '(85) 99999-9999',
    street: 'Rua das Flores',
    number: '450',
    complement: 'Apto 2',
    district: 'Tapera',
    city: 'Aquiraz',
    state: 'ce',
    zipCode: '61700000',
    paymentMethod: 'pix',
    changeFor: '50,00',
    note: 'Sem cebola',
  }
}

describe('preferências do checkout', () => {
  beforeEach(() => {
    localStorage.clear()
    vi.restoreAllMocks()
  })

  it('guarda entrega, contato, endereço e pagamento, mas não observação nem troco', () => {
    saveCheckoutPreferences(filledForm())

    expect(loadCheckoutPreferences()).toEqual({
      fulfillment: 'delivery',
      zoneId: 'zona-1',
      customerName: 'Maria Silva',
      phone: '(85) 99999-9999',
      street: 'Rua das Flores',
      number: '450',
      complement: 'Apto 2',
      district: 'Tapera',
      city: 'Aquiraz',
      state: 'CE',
      zipCode: '61700000',
      paymentMethod: 'pix',
    })

    const stored = localStorage.getItem(CHECKOUT_PREFERENCES_KEY) ?? ''
    expect(stored).not.toContain('Sem cebola')
    expect(stored).not.toContain('50,00')
  })

  it('normaliza e limita os valores antes de gravar', () => {
    const form = filledForm()
    form.customerName = `  ${'M'.repeat(300)}  `
    form.zoneId = 'zona inválida'
    form.state = '1X'
    form.fulfillment = 'pickup'
    form.paymentMethod = 'cash'

    const preferences = buildCheckoutPreferences(form)

    expect(preferences.customerName).toHaveLength(160)
    expect(preferences.zoneId).toBe('')
    expect(preferences.state).toBe('')
    expect(preferences.fulfillment).toBe('pickup')
    expect(preferences.paymentMethod).toBe('cash')
  })

  it('ignora dados corrompidos, de versão antiga ou com tipos inválidos', () => {
    localStorage.setItem(CHECKOUT_PREFERENCES_KEY, '{')
    expect(loadCheckoutPreferences()).toBeNull()

    localStorage.setItem(
      CHECKOUT_PREFERENCES_KEY,
      JSON.stringify({ version: 0, preferences: buildCheckoutPreferences(filledForm()) }),
    )
    expect(loadCheckoutPreferences()).toBeNull()

    localStorage.setItem(
      CHECKOUT_PREFERENCES_KEY,
      JSON.stringify({ version: 1, preferences: { customerName: 42, paymentMethod: 'card' } }),
    )
    expect(loadCheckoutPreferences()).toEqual({
      fulfillment: 'delivery',
      zoneId: '',
      customerName: '',
      phone: '',
      street: '',
      number: '',
      complement: '',
      district: '',
      city: '',
      state: '',
      zipCode: '',
      paymentMethod: 'card',
    })
  })

  it('informa e apaga as preferências salvas', () => {
    expect(hasCheckoutPreferences()).toBe(false)
    saveCheckoutPreferences(filledForm())
    expect(hasCheckoutPreferences()).toBe(true)

    clearCheckoutPreferences()

    expect(hasCheckoutPreferences()).toBe(false)
    expect(localStorage.getItem(CHECKOUT_PREFERENCES_KEY)).toBeNull()
  })

  it('não quebra o pedido se o navegador bloquear o armazenamento', () => {
    vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
      throw new Error('storage blocked')
    })
    vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
      throw new Error('storage blocked')
    })
    vi.spyOn(Storage.prototype, 'removeItem').mockImplementation(() => {
      throw new Error('storage blocked')
    })

    expect(() => saveCheckoutPreferences(filledForm())).not.toThrow()
    expect(loadCheckoutPreferences()).toBeNull()
    expect(hasCheckoutPreferences()).toBe(false)
    expect(() => clearCheckoutPreferences()).not.toThrow()
  })
})
