import { describe, expect, it } from 'vitest'

import { adminCatalogErrorMessage } from './errors'
import { ApiError } from '../../services/http'

describe('adminCatalogErrorMessage', () => {
  it('exibe mensagem de sessão expirada', () => {
    expect(adminCatalogErrorMessage(new ApiError(401, 'UNAUTHORIZED', 'x'), 'load')).toContain(
      'sessão expirou',
    )
  })

  it('exibe mensagem de permissão negada', () => {
    expect(adminCatalogErrorMessage(new ApiError(403, 'FORBIDDEN', 'x'), 'save')).toContain(
      'permissão',
    )
  })

  it('exibe mensagem de conflito de slug', () => {
    expect(adminCatalogErrorMessage(new ApiError(409, 'CONFLICT', 'x'), 'save')).toContain('slug')
  })

  it('usa o detalhe de validação quando disponível', () => {
    const error = new ApiError(422, 'VALIDATION_ERROR', 'Dados inválidos.', [
      { field: 'priceCents', code: 'OUT_OF_RANGE', message: 'Preço inválido.' },
    ])
    expect(adminCatalogErrorMessage(error, 'save')).toBe('Preço inválido.')
  })

  it('esconde detalhes internos em erros inesperados', () => {
    expect(
      adminCatalogErrorMessage(new ApiError(500, 'INTERNAL_ERROR', 'stack trace'), 'save'),
    ).toBe('Não foi possível salvar. Tente novamente.')
    expect(adminCatalogErrorMessage(new Error('boom'), 'load')).toContain('carregar')
  })
})
