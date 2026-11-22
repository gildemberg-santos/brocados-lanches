import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { fetchMenu } from '../services/menu'
import { makeCategory, makeMenu, makeProduct } from '../test/fixtures/catalog'
import { useMenuStore } from './menu'

vi.mock('../services/menu', () => ({
  fetchMenu: vi.fn(),
}))

const mockedFetchMenu = vi.mocked(fetchMenu)

describe('menu store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('loads the menu and exposes category metadata on products', async () => {
    mockedFetchMenu.mockResolvedValue(makeMenu())

    const store = useMenuStore()
    await store.loadMenu()

    expect(store.state).toBe('success')
    expect(store.categories).toHaveLength(1)
    expect(store.products[0]).toMatchObject({
      name: 'Brocado Cheddar',
      categoryId: 'category-1',
      categoryName: 'Hambúrgueres',
      categorySlug: 'hamburgueres',
    })
  })

  it('only loads once when the menu is already available', async () => {
    mockedFetchMenu.mockResolvedValue(makeMenu())

    const store = useMenuStore()
    await store.loadMenu()
    await store.loadMenu()
    await store.ensureLoaded()

    expect(mockedFetchMenu).toHaveBeenCalledTimes(1)
  })

  it('forces a reload when requested', async () => {
    mockedFetchMenu.mockResolvedValue(makeMenu())

    const store = useMenuStore()
    await store.loadMenu()
    await store.loadMenu(true)

    expect(mockedFetchMenu).toHaveBeenCalledTimes(2)
  })

  it('stores a friendly error when the request fails', async () => {
    mockedFetchMenu.mockRejectedValue(new Error('network error'))

    const store = useMenuStore()
    await store.loadMenu()

    expect(store.state).toBe('error')
    expect(store.error).toBe('Não foi possível carregar o cardápio. Tente novamente.')
    expect(store.products).toEqual([])
  })

  it('retries after an error through ensureLoaded', async () => {
    mockedFetchMenu
      .mockRejectedValueOnce(new Error('network error'))
      .mockResolvedValueOnce(makeMenu())

    const store = useMenuStore()
    await store.ensureLoaded()
    await store.ensureLoaded()

    expect(mockedFetchMenu).toHaveBeenCalledTimes(2)
    expect(store.state).toBe('success')
  })

  it('filters featured products to available items', async () => {
    mockedFetchMenu.mockResolvedValue(
      makeMenu({
        categories: [
          makeCategory({
            products: [
              makeProduct({ id: 'featured', featured: true, available: true }),
              makeProduct({ id: 'sold-out', featured: true, available: false }),
              makeProduct({ id: 'regular', featured: false, available: true }),
            ],
          }),
        ],
      }),
    )

    const store = useMenuStore()
    await store.loadMenu()

    expect(store.featuredProducts.map((product) => product.id)).toEqual(['featured'])
  })

  it('marks an empty catalog and finds categories and products by slug', async () => {
    mockedFetchMenu.mockResolvedValue(makeMenu({ categories: [] }))

    const store = useMenuStore()
    await store.loadMenu()

    expect(store.isEmpty).toBe(true)
    expect(store.findProductBySlug('brocado-cheddar')).toBeUndefined()
    expect(store.findCategoryBySlug('hamburgueres')).toBeUndefined()
  })
})
