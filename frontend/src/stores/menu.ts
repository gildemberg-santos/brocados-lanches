import { computed, ref } from 'vue'
import { defineStore } from 'pinia'

import { fetchMenu } from '../services/menu'
import type { CatalogProduct, Menu, MenuCategory } from '../types/catalog'

export type MenuState = 'idle' | 'loading' | 'success' | 'error'

const LOAD_ERROR_MESSAGE = 'Não foi possível carregar o cardápio. Tente novamente.'

export const useMenuStore = defineStore('menu', () => {
  const state = ref<MenuState>('idle')
  const menu = ref<Menu | null>(null)
  const error = ref('')

  const categories = computed<MenuCategory[]>(() => menu.value?.categories ?? [])

  const products = computed<CatalogProduct[]>(() =>
    categories.value.flatMap((category) =>
      category.products.map((product) => ({
        ...product,
        categoryId: category.id,
        categoryName: category.name,
        categorySlug: category.slug,
      })),
    ),
  )

  const featuredProducts = computed<CatalogProduct[]>(() =>
    products.value.filter((product) => product.available && product.featured),
  )

  const isEmpty = computed(() => state.value === 'success' && products.value.length === 0)

  async function loadMenu(force = false): Promise<void> {
    if (state.value === 'loading') {
      return
    }

    if (!force && state.value === 'success') {
      return
    }

    state.value = 'loading'
    error.value = ''

    try {
      menu.value = await fetchMenu()
      state.value = 'success'
    } catch {
      state.value = 'error'
      error.value = LOAD_ERROR_MESSAGE
    }
  }

  function ensureLoaded(): Promise<void> {
    if (state.value === 'success' || state.value === 'loading') {
      return Promise.resolve()
    }

    return loadMenu()
  }

  function findProductBySlug(slug: string): CatalogProduct | undefined {
    return products.value.find((product) => product.slug === slug)
  }

  function findCategoryBySlug(slug: string): MenuCategory | undefined {
    return categories.value.find((category) => category.slug === slug)
  }

  return {
    state,
    menu,
    error,
    categories,
    products,
    featuredProducts,
    isEmpty,
    loadMenu,
    ensureLoaded,
    findProductBySlug,
    findCategoryBySlug,
  }
})
