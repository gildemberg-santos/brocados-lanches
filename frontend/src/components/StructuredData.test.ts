import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { defineComponent, h } from 'vue'
import { createMemoryHistory, createRouter } from 'vue-router'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { fetchMenu } from '../services/menu'
import { fetchStoreStatus } from '../services/store'
import { makeCategory, makeMenu, makeProduct } from '../test/fixtures/catalog'
import type { PublicStoreStatus } from '../types/store'
import StructuredData from './StructuredData.vue'

vi.mock('../services/menu', () => ({ fetchMenu: vi.fn() }))
vi.mock('../services/store', () => ({ fetchStoreStatus: vi.fn() }))

const mockedFetchMenu = vi.mocked(fetchMenu)
const mockedFetchStoreStatus = vi.mocked(fetchStoreStatus)

const StubView = defineComponent({ name: 'StubView', render: () => h('div') })

const status: PublicStoreStatus = {
  name: 'Brocados Lanches',
  status: 'open',
  isOpenManual: true,
  timezone: 'America/Fortaleza',
  deliveryTime: '30 min',
  whatsapp: '+5585999999999',
  instagram: '',
  address: 'Rua das Palmeiras, 123',
  openingHours: [],
  nextOpeningAt: null,
  updatedAt: '2026-11-25T12:00:00Z',
}

const menu = makeMenu({
  categories: [
    makeCategory({
      id: 'category-1',
      name: 'Hambúrgueres',
      slug: 'hamburgueres',
      products: [
        makeProduct({
          id: 'product-1',
          name: 'Brocado Cheddar',
          slug: 'brocado-cheddar',
          priceCents: 2300,
        }),
      ],
    }),
  ],
})

async function mountAt(path: string) {
  const pinia = createPinia()
  setActivePinia(pinia)

  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', name: 'home', component: StubView },
      { path: '/cardapio', name: 'menu', component: StubView },
      { path: '/produto/:slug', name: 'product', component: StubView },
      { path: '/checkout', name: 'checkout', component: StubView },
    ],
  })

  await router.push(path)
  await router.isReady()

  const wrapper = mount(StructuredData, { global: { plugins: [pinia, router] } })
  await flushPromises()

  return { wrapper, router }
}

function jsonLd(): unknown {
  const content = document.getElementById('brocados-structured-data')?.textContent ?? ''
  return content.length > 0 ? JSON.parse(content) : null
}

describe('StructuredData', () => {
  beforeEach(() => {
    document.getElementById('brocados-structured-data')?.remove()
    mockedFetchStoreStatus.mockResolvedValue(status)
    mockedFetchMenu.mockResolvedValue(menu)
  })

  it('publica Restaurant na home', async () => {
    await mountAt('/')

    expect(jsonLd()).toMatchObject({
      '@type': 'Restaurant',
      name: 'Brocados Lanches',
    })
  })

  it('publica Restaurant e Menu no cardápio', async () => {
    await mountAt('/cardapio')

    expect(jsonLd()).toMatchObject({
      '@graph': [{ '@type': 'Restaurant' }, { '@type': 'Menu' }],
    })
  })

  it('publica Product no detalhe', async () => {
    await mountAt('/produto/brocado-cheddar')

    expect(jsonLd()).toMatchObject({
      '@type': 'Product',
      name: 'Brocado Cheddar',
    })
  })

  it('não publica dados em rotas sem valor de indexação', async () => {
    await mountAt('/checkout')

    expect(document.getElementById('brocados-structured-data')).toBeNull()
  })

  it('remove o bloco quando a loja não pode ser carregada', async () => {
    mockedFetchStoreStatus.mockRejectedValue(new Error('falha'))
    await mountAt('/')

    expect(document.getElementById('brocados-structured-data')).toBeNull()
  })
})
