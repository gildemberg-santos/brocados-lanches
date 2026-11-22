import { beforeEach, describe, expect, it } from 'vitest'

import { readHandoffUrl, saveHandoffUrl } from './handoff'

describe('handoff storage', () => {
  beforeEach(() => {
    sessionStorage.clear()
  })

  it('stores and reads the link per order', () => {
    saveHandoffUrl('public-1', 'https://wa.me/5585999999999?text=pedido')

    expect(readHandoffUrl('public-1')).toBe('https://wa.me/5585999999999?text=pedido')
    expect(readHandoffUrl('public-2')).toBe('')
  })

  it('ignores empty values', () => {
    saveHandoffUrl('', 'https://wa.me/5585999999999')
    saveHandoffUrl('public-1', '')

    expect(readHandoffUrl('')).toBe('')
    expect(readHandoffUrl('public-1')).toBe('')
  })
})
