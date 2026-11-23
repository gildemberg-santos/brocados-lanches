import { describe, expect, it } from 'vitest'

import { ApiError } from '../../services/http'
import {
  adminAuthErrorMessage,
  adminRoleLabel,
  canAccessAdminNav,
  primaryAdminNav,
  visibleAdminNav,
} from './auth'

describe('regras de autenticação administrativa', () => {
  it('troca 401 por mensagem genérica sem revelar conta', () => {
    const error = new ApiError(401, 'UNAUTHORIZED', 'credenciais inválidas')

    expect(adminAuthErrorMessage(error)).toBe('E-mail ou senha incorretos.')
  })

  it('explica limite de tentativas, permissão e validação', () => {
    expect(adminAuthErrorMessage(new ApiError(429, 'RATE_LIMITED', 'muitas'))).toContain(
      'Muitas tentativas',
    )
    expect(adminAuthErrorMessage(new ApiError(403, 'FORBIDDEN', 'negado'))).toBe(
      'Você não tem permissão para acessar esta área.',
    )
    expect(adminAuthErrorMessage(new ApiError(400, 'VALIDATION_ERROR', 'campos inválidos'))).toBe(
      'Revise os dados informados e tente novamente.',
    )
  })

  it('devolve mensagem de rede para erro sem ApiError', () => {
    expect(adminAuthErrorMessage(new Error('network'))).toContain('Não foi possível entrar')
  })

  it('esconde itens de owner do operador', () => {
    const operatorItems = visibleAdminNav('operator').map((item) => item.name)

    expect(operatorItems).toContain('admin-products')
    expect(operatorItems).not.toContain('admin-users')
    expect(operatorItems).not.toContain('admin-audit')
    expect(visibleAdminNav('owner').map((item) => item.name)).toContain('admin-users')
    expect(
      canAccessAdminNav({ label: 'Usuários', name: 'admin-users', roles: ['owner'] }, null),
    ).toBe(false)
  })

  it('mantém apenas os atalhos primários no mobile', () => {
    expect(primaryAdminNav('operator').map((item) => item.name)).toEqual([
      'admin-dashboard',
      'admin-orders',
      'admin-products',
      'admin-account',
    ])
  })

  it('rotula os papéis em português', () => {
    expect(adminRoleLabel('owner')).toBe('Administrador')
    expect(adminRoleLabel('operator')).toBe('Operador')
    expect(adminRoleLabel(null)).toBe('Sem sessão')
  })
})
