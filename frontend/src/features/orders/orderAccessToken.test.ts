import { beforeEach, describe, expect, it, vi } from 'vitest'

import {
  ORDER_ACCESS_TOKEN_KEY,
  clearOrderAccessToken,
  hasOrderAccessToken,
  isValidOrderAccessToken,
  loadOrderAccessToken,
  saveOrderAccessToken,
} from './orderAccessToken'

const VALID_TOKEN = 'a'.repeat(43)

describe('token de acesso aos pedidos', () => {
  beforeEach(() => {
    localStorage.clear()
    vi.restoreAllMocks()
  })

  it('valida apenas tokens opacos de 43 caracteres base64url', () => {
    expect(isValidOrderAccessToken(VALID_TOKEN)).toBe(true)
    expect(isValidOrderAccessToken('A'.repeat(43))).toBe(true)
    expect(isValidOrderAccessToken('a'.repeat(42))).toBe(false)
    expect(isValidOrderAccessToken('a'.repeat(44))).toBe(false)
    expect(isValidOrderAccessToken('*'.repeat(43))).toBe(false)
    expect(isValidOrderAccessToken('')).toBe(false)
  })

  it('guarda, lê e apaga o token com versão', () => {
    expect(hasOrderAccessToken()).toBe(false)

    saveOrderAccessToken(`  ${VALID_TOKEN}  `)

    expect(loadOrderAccessToken()).toBe(VALID_TOKEN)
    expect(hasOrderAccessToken()).toBe(true)
    const stored = localStorage.getItem(ORDER_ACCESS_TOKEN_KEY) ?? ''
    expect(stored).toContain('"version":1')
    expect(stored).not.toContain('"phone"')

    clearOrderAccessToken()

    expect(loadOrderAccessToken()).toBeNull()
    expect(hasOrderAccessToken()).toBe(false)
  })

  it('não grava token inválido', () => {
    saveOrderAccessToken('curto')

    expect(localStorage.getItem(ORDER_ACCESS_TOKEN_KEY)).toBeNull()
    expect(loadOrderAccessToken()).toBeNull()
  })

  it('descarta dados corrompidos, de versão antiga ou com token inválido', () => {
    localStorage.setItem(ORDER_ACCESS_TOKEN_KEY, '{')
    expect(loadOrderAccessToken()).toBeNull()

    localStorage.setItem(ORDER_ACCESS_TOKEN_KEY, JSON.stringify({ version: 0, token: VALID_TOKEN }))
    expect(loadOrderAccessToken()).toBeNull()

    localStorage.setItem(ORDER_ACCESS_TOKEN_KEY, JSON.stringify({ version: 1, token: 'invalido' }))
    expect(loadOrderAccessToken()).toBeNull()

    localStorage.setItem(ORDER_ACCESS_TOKEN_KEY, JSON.stringify({ version: 1, token: 42 }))
    expect(loadOrderAccessToken()).toBeNull()
  })

  it('não quebra se o navegador bloquear o armazenamento', () => {
    vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
      throw new Error('storage blocked')
    })
    vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
      throw new Error('storage blocked')
    })
    vi.spyOn(Storage.prototype, 'removeItem').mockImplementation(() => {
      throw new Error('storage blocked')
    })

    expect(() => saveOrderAccessToken(VALID_TOKEN)).not.toThrow()
    expect(loadOrderAccessToken()).toBeNull()
    expect(hasOrderAccessToken()).toBe(false)
    expect(() => clearOrderAccessToken()).not.toThrow()
  })
})
