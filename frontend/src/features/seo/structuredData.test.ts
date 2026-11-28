import { describe, expect, it } from 'vitest'

import { makeProduct } from '../../test/fixtures/catalog'
import type { PublicStoreStatus } from '../../types/store'
import {
  buildMenuJsonLd,
  buildProductJsonLd,
  buildRestaurantJsonLd,
  centsToSchemaPrice,
  serializeJsonLd,
} from './structuredData'

const status: PublicStoreStatus = {
  name: 'Brocados Lanches',
  status: 'open',
  isOpenManual: true,
  timezone: 'America/Fortaleza',
  deliveryTime: '30 a 40 min',
  whatsapp: '+5585999999999',
  instagram: '',
  address: 'Rua das Palmeiras, 123',
  openingHours: [
    {
      day: 'monday',
      closed: false,
      intervals: [{ open: '11:00', close: '14:00' }],
    },
    {
      day: 'sunday',
      closed: true,
      intervals: [],
    },
  ],
  pix: { active: false, key: '', merchantName: '', merchantCity: '' },
  nextOpeningAt: null,
  updatedAt: '2026-11-25T12:00:00Z',
}

describe('dados estruturados de SEO', () => {
  it('converte centavos sem ponto flutuante', () => {
    expect(centsToSchemaPrice(2300)).toBe('23.00')
    expect(centsToSchemaPrice(5490)).toBe('54.90')
    expect(centsToSchemaPrice(5)).toBe('0.05')
  })

  it('gera Restaurant apenas com dados da loja', () => {
    const data = buildRestaurantJsonLd(status)

    expect(data['@type']).toBe('Restaurant')
    expect(data.name).toBe('Brocados Lanches')
    expect(data.telephone).toBe('+5585999999999')
    expect(data.address).toEqual({
      '@type': 'PostalAddress',
      streetAddress: 'Rua das Palmeiras, 123',
    })
    expect(data.openingHoursSpecification).toEqual([
      {
        '@type': 'OpeningHoursSpecification',
        dayOfWeek: 'https://schema.org/Monday',
        opens: '11:00',
        closes: '14:00',
      },
    ])
  })

  it('omite campos vazios em vez de inventar dados', () => {
    const data = buildRestaurantJsonLd({ ...status, whatsapp: '', address: '', openingHours: [] })

    expect(data.telephone).toBeUndefined()
    expect(data.address).toBeUndefined()
    expect(data.openingHoursSpecification).toBeUndefined()
  })

  it('gera Menu com preços congelados em reais', () => {
    const product = makeProduct({
      id: 'product-1',
      name: 'Brocado Cheddar',
      slug: 'brocado-cheddar',
      priceCents: 2350,
      description: 'Pão, carne e cheddar',
      imageUrl: '/products/brocado.jpg',
    })

    const data = buildMenuJsonLd([
      {
        id: 'category-1',
        name: 'Hambúrgueres',
        slug: 'hamburgueres',
        sortOrder: 1,
        products: [product],
      },
    ])

    expect(data?.['@type']).toBe('Menu')
    expect(data?.hasMenuSection).toMatchObject([
      {
        '@type': 'MenuSection',
        name: 'Hambúrgueres',
        hasMenuItem: [
          {
            '@type': 'MenuItem',
            name: 'Brocado Cheddar',
            offers: { price: '23.50', priceCurrency: 'BRL' },
          },
        ],
      },
    ])
  })

  it('não gera Menu vazio', () => {
    expect(buildMenuJsonLd([])).toBeNull()
    expect(
      buildMenuJsonLd([{ id: 'c', name: 'Vazia', slug: 'vazia', sortOrder: 1, products: [] }]),
    ).toBeNull()
  })

  it('gera Product para a rota de detalhe', () => {
    const product = makeProduct({ id: 'p', name: 'Brocado', slug: 'brocado', priceCents: 1990 })
    const data = buildProductJsonLd(product)

    expect(data['@type']).toBe('Product')
    expect(data.offers).toMatchObject({ price: '19.90', priceCurrency: 'BRL' })
  })

  it('serializa vários objetos em @graph', () => {
    const serialized = serializeJsonLd([
      buildRestaurantJsonLd(status),
      buildProductJsonLd(makeProduct()),
    ])
    const parsed = JSON.parse(serialized) as { '@graph': unknown[] }

    expect(parsed['@graph']).toHaveLength(2)
    expect(serializeJsonLd([])).toBe('')
  })
})
