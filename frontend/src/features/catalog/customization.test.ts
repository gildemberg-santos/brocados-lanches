import { describe, expect, it } from 'vitest'

import { makeAddon, makeOption, makeOptionGroup, makeProduct } from '../../test/fixtures/catalog'
import {
  calculateConfiguredPriceCents,
  createOptionSelections,
  isOptionSelectionValid,
  minimumSelections,
  validateOptionGroup,
  validateOptionGroups,
} from './customization'

describe('product customization', () => {
  it('creates an empty selection for every option group', () => {
    const group = makeOptionGroup({ id: 'doneness' })

    expect(createOptionSelections([group])).toEqual({ doneness: [] })
  })

  it('uses the required flag as a minimum of one selection', () => {
    expect(minimumSelections(makeOptionGroup({ required: true, minSelections: 0 }))).toBe(1)
    expect(minimumSelections(makeOptionGroup({ required: false, minSelections: 0 }))).toBe(0)
    expect(minimumSelections(makeOptionGroup({ required: false, minSelections: 2 }))).toBe(2)
  })

  it('rejects a required group without a selection', () => {
    expect(validateOptionGroup(makeOptionGroup(), [])).toBe('Escolha uma opção.')
    expect(isOptionSelectionValid([makeOptionGroup()], { 'group-1': [] })).toBe(false)
  })

  it('accepts a valid single selection', () => {
    const group = makeOptionGroup()

    expect(validateOptionGroup(group, ['option-1'])).toBe('')
    expect(isOptionSelectionValid([group], { 'group-1': ['option-1'] })).toBe(true)
  })

  it('accepts an optional group without a selection', () => {
    const group = makeOptionGroup({ required: false, minSelections: 0, maxSelections: 2 })

    expect(validateOptionGroup(group, [])).toBe('')
  })

  it('enforces the maximum number of selections', () => {
    const group = makeOptionGroup({
      required: false,
      minSelections: 0,
      maxSelections: 2,
      options: [makeOption({ id: 'a' }), makeOption({ id: 'b' }), makeOption({ id: 'c' })],
    })

    expect(validateOptionGroup(group, ['a', 'b', 'c'])).toBe('Escolha no máximo 2 opções.')
  })

  it('blocks a group that has no available options', () => {
    const group = makeOptionGroup({ required: false, options: [] })

    expect(validateOptionGroup(group, [])).toBe('Este grupo está sem opções disponíveis.')
  })

  it('returns a message per invalid group', () => {
    const first = makeOptionGroup({ id: 'first' })
    const second = makeOptionGroup({ id: 'second' })
    const errors = validateOptionGroups([first, second], { first: ['option-1'], second: [] })

    expect(errors).toEqual({ second: 'Escolha uma opção.' })
  })

  it('adds selected option and addon prices to the base price', () => {
    const product = makeProduct({ priceCents: 2300 })
    const group = makeOptionGroup({
      required: false,
      minSelections: 0,
      maxSelections: 2,
      options: [
        makeOption({ id: 'bacon', priceCents: 500 }),
        makeOption({ id: 'cheddar', priceCents: 400 }),
      ],
    })
    const addons = [makeAddon({ id: 'extra-bacon', priceCents: 700 })]

    expect(
      calculateConfiguredPriceCents(product, [group], { 'group-1': ['bacon'] }, addons, [
        'extra-bacon',
      ]),
    ).toBe(3500)
  })

  it('ignores unknown, duplicated and unavailable selections in the price', () => {
    const product = makeProduct({ priceCents: 1000 })
    const group = makeOptionGroup({
      required: false,
      minSelections: 0,
      maxSelections: 2,
      options: [
        makeOption({ id: 'valid', priceCents: 100 }),
        makeOption({ id: 'unavailable', priceCents: 900, available: false }),
      ],
    })
    const addons = [makeAddon({ id: 'inactive', priceCents: 700, available: false })]

    expect(
      calculateConfiguredPriceCents(
        product,
        [group],
        { 'group-1': ['valid', 'valid', 'unavailable', 'unknown'] },
        addons,
        ['inactive'],
      ),
    ).toBe(1100)
  })
})
