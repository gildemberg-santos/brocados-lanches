import { describe, expect, it } from 'vitest'

import { ApiError } from '../../services/http'

import {
  ORDER_STATUSES,
  adminOrderErrorMessage,
  availableTransitions,
  isCancellationAllowed,
  isFinalStatus,
  isOrderFulfillment,
  isOrderStatus,
  validateStatusChange,
} from './orderStatus'

describe('availableTransitions', () => {
  it.each([
    ['pending', 'delivery', ['confirmed', 'cancelled']],
    ['pending', 'pickup', ['confirmed', 'cancelled']],
    ['confirmed', 'delivery', ['preparing', 'cancelled']],
    ['preparing', 'pickup', ['ready', 'cancelled']],
    ['ready', 'delivery', ['out_for_delivery']],
    ['ready', 'pickup', ['picked_up']],
    ['out_for_delivery', 'delivery', ['delivered']],
    ['picked_up', 'pickup', []],
    ['delivered', 'delivery', []],
    ['cancelled', 'delivery', []],
  ])('de %s (%s) para %j', (status, fulfillment, expected) => {
    expect(availableTransitions(status, fulfillment)).toEqual(expected)
  })

  it('não oferece retirada em pedido de entrega e vice-versa', () => {
    expect(availableTransitions('ready', 'delivery')).not.toContain('picked_up')
    expect(availableTransitions('ready', 'pickup')).not.toContain('out_for_delivery')
  })

  it('não oferece ações para valores desconhecidos', () => {
    expect(availableTransitions('cozendo', 'delivery')).toEqual([])
    expect(availableTransitions('pending', 'teleporte')).toEqual([])
  })
})

describe('isFinalStatus', () => {
  it.each([...ORDER_STATUSES])('classifica %s', (status) => {
    const expected = status === 'delivered' || status === 'picked_up' || status === 'cancelled'
    expect(isFinalStatus(status)).toBe(expected)
  })
})

describe('isCancellationAllowed', () => {
  it('permite cancelar apenas nos três primeiros estados', () => {
    expect(isCancellationAllowed('pending')).toBe(true)
    expect(isCancellationAllowed('confirmed')).toBe(true)
    expect(isCancellationAllowed('preparing')).toBe(true)
    expect(isCancellationAllowed('ready')).toBe(false)
    expect(isCancellationAllowed('delivered')).toBe(false)
  })
})

describe('adminOrderErrorMessage', () => {
  it('explica conflito de status sem revelar detalhes internos', () => {
    const error = new ApiError(409, 'CONFLICT', 'raw')
    expect(adminOrderErrorMessage(error, 'status')).toContain('mudou de status')
  })

  it('usam mensagens específicas para sessão e permissão', () => {
    expect(adminOrderErrorMessage(new ApiError(401, 'UNAUTHORIZED', 'raw'), 'load')).toContain(
      'sessão expirou',
    )
    expect(adminOrderErrorMessage(new ApiError(403, 'FORBIDDEN', 'raw'), 'status')).toContain(
      'permissão',
    )
  })

  it('tem fallback para erro desconhecido e não expõe detalhe interno', () => {
    expect(adminOrderErrorMessage(new Error('boom'), 'load')).toBe(
      'Não foi possível carregar os pedidos. Tente novamente.',
    )
    expect(
      adminOrderErrorMessage(new ApiError(500, 'INTERNAL_ERROR', 'mongo: boom'), 'status'),
    ).toBe('Não foi possível atualizar o pedido. Tente novamente.')
  })
})

describe('validateStatusChange', () => {
  it('aceita transições sem motivo', () => {
    expect(validateStatusChange('confirmed', '')).toEqual({
      valid: true,
      request: { status: 'confirmed' },
    })
  })

  it('exige e normaliza o motivo no cancelamento', () => {
    expect(validateStatusChange('cancelled', '  Cliente desistiu  ')).toEqual({
      valid: true,
      request: { status: 'cancelled', reason: 'Cliente desistiu' },
    })
    expect(validateStatusChange('cancelled', '   ')).toMatchObject({
      valid: false,
      error: 'Informe o motivo do cancelamento.',
    })
  })

  it('limita o motivo a 280 caracteres', () => {
    expect(validateStatusChange('cancelled', 'a'.repeat(281))).toMatchObject({ valid: false })
    expect(validateStatusChange('cancelled', 'a'.repeat(280))).toMatchObject({ valid: true })
  })

  it('rejeita status desconhecido', () => {
    expect(validateStatusChange('cozendo', '')).toMatchObject({ valid: false })
  })
})

describe('guards de tipo', () => {
  it('reconhece valores conhecidos e rejeita os demais', () => {
    expect(isOrderStatus('pending')).toBe(true)
    expect(isOrderStatus('entregue')).toBe(false)
    expect(isOrderFulfillment('pickup')).toBe(true)
    expect(isOrderFulfillment('retirada')).toBe(false)
  })
})
