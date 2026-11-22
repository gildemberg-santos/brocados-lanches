import { describe, expect, it } from 'vitest'

import { formatOrderDateTime, formatOrderTime } from './format'

describe('formatadores de pedido', () => {
  it('converte UTC para o horário de Fortaleza', () => {
    expect(formatOrderDateTime('2026-11-22T02:30:00Z')).toContain('23:30')
    expect(formatOrderDateTime('2026-11-22T02:30:00Z')).toContain('21/11/2026')
    expect(formatOrderTime('2026-11-22T02:30:00Z')).toContain('23:30')
  })

  it('não inventa data para valor inválido', () => {
    expect(formatOrderDateTime('ontem')).toBe('—')
    expect(formatOrderTime('')).toBe('—')
  })
})
