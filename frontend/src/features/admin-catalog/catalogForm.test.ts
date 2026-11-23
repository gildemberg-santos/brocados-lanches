import { describe, expect, it } from 'vitest'

import {
  allergensFromText,
  allergensToText,
  emptyOpeningHours,
  firstFieldError,
  isValidObjectId,
  isValidSlug,
  isValidTime,
  moneyToCents,
  newLocalId,
  slugify,
  validateProductInput,
  validateSimpleInput,
  validateStoreInput,
  WEEKDAYS,
} from './catalogForm'
import type { AdminProductInput, AdminStoreInput } from '../../types/adminCatalog'

const VALID_PRODUCT: AdminProductInput = {
  categoryId: 'a'.repeat(24),
  name: 'Hambúrguer',
  slug: 'hamburger',
  description: 'Descrição',
  priceCents: 2500,
  imageUrl: '',
  imageAlt: '',
  allergens: [],
  available: true,
  featured: false,
  active: true,
  sortOrder: 0,
  optionGroups: [],
  addonIds: [],
}

const VALID_STORE: AdminStoreInput = {
  name: 'Brocados Lanches',
  isOpen: true,
  openingHours: [],
  deliveryTime: '35-50 min',
  whatsapp: '+5585999999999',
  instagram: '@brocados',
  address: 'Av. Principal, 1200',
}

describe('catalogForm', () => {
  it('gera slugs sem acentos, caixa ou símbolos', () => {
    expect(slugify('Hambúrguer da Casa!')).toBe('hamburguer-da-casa')
    expect(slugify('  Duplo   queijo  ')).toBe('duplo-queijo')
  })

  it('valida slug e identificador do Mongo', () => {
    expect(isValidSlug('duplo-cheddar')).toBe(true)
    expect(isValidSlug('Duplo-Cheddar')).toBe(false)
    expect(isValidSlug('duplo--cheddar')).toBe(false)
    expect(isValidObjectId('a'.repeat(24))).toBe(true)
    expect(isValidObjectId('abc')).toBe(false)
  })

  it('converte texto de dinheiro em centavos sem parseFloat', () => {
    expect(moneyToCents('24,90')).toBe(2490)
    expect(moneyToCents('25')).toBe(2500)
    expect(moneyToCents('')).toBeNull()
  })

  it('converte alergênicos de e para texto', () => {
    expect(allergensToText(['glúten', 'leite'])).toBe('glúten, leite')
    expect(allergensFromText(' glúten , , leite ')).toEqual(['glúten', 'leite'])
  })

  it('cria sete dias fechados e identificadores locais', () => {
    const hours = emptyOpeningHours()
    expect(hours.map((hour) => hour.day)).toEqual(WEEKDAYS)
    expect(hours.every((hour) => hour.closed && hour.intervals.length === 0)).toBe(true)
    expect(newLocalId('op')).toMatch(/^op-/)
  })

  it('valida horários no formato HH:MM', () => {
    expect(isValidTime('18:30')).toBe(true)
    expect(isValidTime('24:00')).toBe(false)
  })

  it('aponta erros nos campos obrigatórios de categorias, adicionais e zonas', () => {
    expect(
      validateSimpleInput('category', {
        name: '',
        slug: 'Invalido',
        description: '',
        sortOrder: 0,
        active: true,
      }),
    ).toMatchObject({ name: expect.any(String), slug: expect.any(String) })

    expect(
      validateSimpleInput('addon', {
        name: 'Cheddar',
        slug: 'cheddar',
        priceCents: -1,
        sortOrder: 0,
        active: true,
      }),
    ).toHaveProperty('priceCents')

    expect(
      validateSimpleInput('zone', {
        name: 'Centro',
        slug: 'centro',
        feeCents: 500,
        sortOrder: -1,
        active: true,
      }),
    ).toHaveProperty('sortOrder')
  })

  it('valida produto, grupos de opções e vínculos', () => {
    expect(validateProductInput(VALID_PRODUCT)).toEqual({})

    const errors = validateProductInput({
      ...VALID_PRODUCT,
      categoryId: 'curto',
      imageUrl: '/imagem.webp',
      addonIds: ['curto'],
      optionGroups: [
        {
          id: 'grupo-1',
          name: '',
          required: true,
          minSelections: 2,
          maxSelections: 1,
          options: [{ id: 'op-1', name: '', priceCents: -1, available: true, sortOrder: 0 }],
        },
      ],
    })

    expect(errors).toHaveProperty('categoryId')
    expect(errors).toHaveProperty('imageAlt')
    expect(errors).toHaveProperty('addonIds')
    expect(errors).toHaveProperty('optionGroups.0.name')
    expect(errors).toHaveProperty('optionGroups.0.maxSelections')
    expect(errors).toHaveProperty('optionGroups.0.options.0.priceCents')
  })

  it('valida contato e horários da loja', () => {
    expect(validateStoreInput(VALID_STORE)).toEqual({})

    const errors = validateStoreInput({
      ...VALID_STORE,
      whatsapp: '(85) 99999-9999',
      openingHours: [
        { day: 'sunday', closed: false, intervals: [] },
        {
          day: 'monday',
          closed: false,
          intervals: [
            { open: '11:00', close: '15:00' },
            { open: '14:00', close: '18:00' },
          ],
        },
      ],
    })

    expect(errors).toHaveProperty('whatsapp')
    expect(errors).toHaveProperty('openingHours.0')
    expect(errors).toHaveProperty('openingHours.1')
  })

  it('devolve a primeira mensagem de erro', () => {
    expect(firstFieldError({ name: 'Nome obrigatório', slug: 'Slug inválido' })).toBe(
      'Nome obrigatório',
    )
    expect(firstFieldError({})).toBe('')
  })
})
