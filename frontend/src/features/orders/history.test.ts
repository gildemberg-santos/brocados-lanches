import { describe, expect, it } from 'vitest'

import { ApiError } from '../../services/http'
import { EMPTY_ORDER_HISTORY_META, orderHistoryErrorMessage } from './history'

describe('histórico de pedidos', () => {
  it('começa com uma página vazia e segura', () => {
    expect(EMPTY_ORDER_HISTORY_META).toEqual({
      page: 1,
      pageSize: 20,
      total: 0,
      totalPages: 0,
    })
  })

  it('traduz erros sem expor detalhes internos', () => {
    expect(orderHistoryErrorMessage(new ApiError(429, 'RATE_LIMITED', 'raw'))).toContain(
      'Muitas consultas',
    )
    expect(orderHistoryErrorMessage(new ApiError(400, 'VALIDATION_ERROR', 'raw token'))).toContain(
      'Faça um novo pedido',
    )
    expect(orderHistoryErrorMessage(new ApiError(500, 'INTERNAL_ERROR', 'mongo: boom'))).toBe(
      'Não foi possível carregar seus pedidos. Tente novamente.',
    )
    expect(orderHistoryErrorMessage(new Error('boom'))).not.toContain('boom')
  })
})
