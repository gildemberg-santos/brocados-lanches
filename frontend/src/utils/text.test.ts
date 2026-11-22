import { describe, expect, it } from 'vitest'

import { normalizeText, splitSearchTerms } from './text'

describe('text utils', () => {
  it('normalizes accents, case and surrounding whitespace', () => {
    expect(normalizeText('  Brocadão Cheddar  ')).toBe('brocadao cheddar')
  })

  it('splits a search into normalized terms', () => {
    expect(splitSearchTerms('  PÃO   brioche ')).toEqual(['pao', 'brioche'])
  })

  it('returns no terms for a blank search', () => {
    expect(splitSearchTerms('   ')).toEqual([])
  })
})
