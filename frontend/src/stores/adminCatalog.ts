import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import { adminCatalogErrorMessage } from '../features/admin-catalog/errors'
import {
  createAdminAddon,
  createAdminCategory,
  createAdminProduct,
  createAdminZone,
  fetchAdminAddons,
  fetchAdminCategories,
  fetchAdminProducts,
  fetchAdminZones,
  updateAdminAddon,
  updateAdminCategory,
  updateAdminProduct,
  updateAdminZone,
} from '../services/adminCatalog'
import type { AdminPageMeta } from '../types/admin'
import type {
  AdminAddon,
  AdminAddonInput,
  AdminCatalogListQuery,
  AdminCategory,
  AdminCategoryInput,
  AdminProduct,
  AdminProductInput,
  AdminZone,
  AdminZoneInput,
} from '../types/adminCatalog'

export type AdminCatalogState = 'idle' | 'loading' | 'success' | 'error'

const DEFAULT_QUERY: AdminCatalogListQuery = {
  page: 1,
  pageSize: 20,
  sort: 'sortOrder',
  order: 'asc',
}

const EMPTY_META: AdminPageMeta = { page: 1, pageSize: 20, total: 0, totalPages: 0 }

interface ResourceApi<T, I> {
  list: (query: AdminCatalogListQuery) => Promise<{ data: T[]; meta: AdminPageMeta }>
  create: (input: I) => Promise<T>
  update: (id: string, input: Partial<I>) => Promise<T>
}

function createResource<T extends { id: string }, I>(
  api: ResourceApi<T, I>,
  defaultQuery: AdminCatalogListQuery = DEFAULT_QUERY,
) {
  const state = ref<AdminCatalogState>('idle')
  const items = ref<T[]>([])
  const meta = ref<AdminPageMeta>({ ...EMPTY_META })
  const filters = ref<AdminCatalogListQuery>({ ...defaultQuery })
  const error = ref('')
  const saving = ref(false)
  const saveError = ref('')
  const saveSuccess = ref('')

  const isEmpty = computed(() => state.value === 'success' && items.value.length === 0)
  const totalPages = computed(() => meta.value.totalPages)

  async function load(): Promise<void> {
    state.value = 'loading'
    error.value = ''
    try {
      const page = await api.list({ ...filters.value })
      items.value = page.data
      meta.value = page.meta
      state.value = 'success'
    } catch (cause: unknown) {
      items.value = []
      meta.value = { ...EMPTY_META }
      error.value = adminCatalogErrorMessage(cause, 'load')
      state.value = 'error'
    }
  }

  function normalizeFilters(next: Partial<AdminCatalogListQuery>): AdminCatalogListQuery {
    const result: AdminCatalogListQuery = { ...defaultQuery, ...next, page: 1 }
    if (!result.q?.trim()) {
      delete result.q
    }
    if (!result.categoryId) {
      delete result.categoryId
    }
    return result
  }

  async function applyFilters(next: Partial<AdminCatalogListQuery>): Promise<void> {
    filters.value = normalizeFilters(next)
    await load()
  }

  async function goToPage(page: number): Promise<void> {
    if (page < 1 || page > Math.max(totalPages.value, 1)) {
      return
    }
    filters.value = { ...filters.value, page }
    await load()
  }

  function clearMessages(): void {
    saveError.value = ''
    saveSuccess.value = ''
  }

  async function save(id: string | null, input: I | Partial<I>): Promise<T | null> {
    saving.value = true
    clearMessages()
    try {
      const saved = id ? await api.update(id, input as Partial<I>) : await api.create(input as I)
      await load()
      saveSuccess.value = id ? 'Alterações salvas.' : 'Cadastro criado.'
      return saved
    } catch (cause: unknown) {
      saveError.value = adminCatalogErrorMessage(cause, 'save')
      return null
    } finally {
      saving.value = false
    }
  }

  return {
    state,
    items,
    meta,
    filters,
    error,
    saving,
    saveError,
    saveSuccess,
    isEmpty,
    totalPages,
    load,
    applyFilters,
    goToPage,
    save,
    clearMessages,
  }
}

export const useAdminCategoriesStore = defineStore('adminCategories', () =>
  createResource<AdminCategory, AdminCategoryInput>({
    list: fetchAdminCategories,
    create: createAdminCategory,
    update: updateAdminCategory,
  }),
)

export const useAdminAddonsStore = defineStore('adminAddons', () =>
  createResource<AdminAddon, AdminAddonInput>({
    list: fetchAdminAddons,
    create: createAdminAddon,
    update: updateAdminAddon,
  }),
)

export const useAdminZonesStore = defineStore('adminZones', () =>
  createResource<AdminZone, AdminZoneInput>({
    list: fetchAdminZones,
    create: createAdminZone,
    update: updateAdminZone,
  }),
)

export const useAdminProductsStore = defineStore('adminProducts', () =>
  createResource<AdminProduct, AdminProductInput>({
    list: fetchAdminProducts,
    create: createAdminProduct,
    update: updateAdminProduct,
  }),
)
