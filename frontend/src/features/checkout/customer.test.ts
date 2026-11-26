import { describe, expect, it } from 'vitest'

import { initialCheckoutForm } from './checkout'
import type { CheckoutForm } from './checkout'
import { applyCustomerProfile, canLookupCustomer } from './customer'

function formWithIdentity(): CheckoutForm {
  return {
    ...initialCheckoutForm,
    customerName: 'Cliente Digitado',
    phone: '(85) 97777-7777',
    note: 'Sem cebola',
    changeFor: '50,00',
  }
}

describe('perfil de conveniência no checkout', () => {
  it('aceita apenas telefone brasileiro completo para a consulta', () => {
    expect(canLookupCustomer('85977777777')).toBe(true)
    expect(canLookupCustomer('(85) 97777-7777')).toBe(true)
    expect(canLookupCustomer('859777')).toBe(false)
    expect(canLookupCustomer('+341999999999')).toBe(false)
    expect(canLookupCustomer('')).toBe(false)
  })

  it('preenche dados do perfil sem apagar identidade, observação ou troco', () => {
    const result = applyCustomerProfile(
      formWithIdentity(),
      {
        name: 'Maria Silva',
        phone: '(85) 99999-9999',
        fulfillment: 'delivery',
        zone: { id: 'zona-1', name: 'Centro' },
        address: {
          street: 'Rua das Flores',
          number: '450',
          complement: 'Apto 2',
          district: 'Tapera',
          city: 'Aquiraz',
          state: 'CE',
          zipCode: '61700000',
        },
        paymentMethod: 'pix',
      },
      ['zona-1'],
    )

    expect(result).toMatchObject({
      customerName: 'Cliente Digitado',
      phone: '(85) 97777-7777',
      fulfillment: 'delivery',
      zoneId: 'zona-1',
      street: 'Rua das Flores',
      number: '450',
      complement: 'Apto 2',
      district: 'Tapera',
      city: 'Aquiraz',
      state: 'CE',
      zipCode: '61700000',
      paymentMethod: 'pix',
      note: 'Sem cebola',
      changeFor: '50,00',
    })
  })

  it('preenche o nome somente quando o cliente ainda não o informou', () => {
    const result = applyCustomerProfile(
      { ...initialCheckoutForm, phone: '(85) 97777-7777' },
      { name: 'Maria Silva' },
    )

    expect(result.customerName).toBe('Maria Silva')
    expect(result.phone).toBe('(85) 97777-7777')
  })

  it('descarta zona inativa e mantém o carrinho cotado pelo servidor', () => {
    const result = applyCustomerProfile(
      formWithIdentity(),
      { fulfillment: 'delivery', zone: { id: 'zona-antiga', name: 'Antiga' } },
      ['zona-1'],
    )

    expect(result.zoneId).toBe('')
  })
})
