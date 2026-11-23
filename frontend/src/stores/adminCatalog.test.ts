import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import {
  createAdminCategory,
  fetchAdminCategories,
  fetchAdminProducts,
  updateAdminCategory,
} from '../services/adminCatalog'
import { useAdminCategoriesStore, useAdminProductsStore } from './adminCatalog'
import { ApiError } from '../services/http'

vi.mock('../services/adminCatalog', () => ({
  fetchAdminCategories: vi.fn(),
  fetchAdminAddons: vi.fn(),
  fetchAdminZones: vi.fn(),
  fetchAdminProducts: vi.fn(),
  createAdminCategory: vi.fn(),
  createAdminAddon: vi.fn(),
  createAdminZone: vi.fn(),
  createAdminProduct: vi.fn(),
  updateAdminCategory: vi.fn(),
  updateAdminAddon: vi.fn(),
  updateAdminZone: vi.fn(),
  updateAdminProduct: vi.fn(),
}))

const mockedFetchCategories = vi.mocked(fetchAdminCategories)
const mockedFetchProducts = vi.mocked(fetchAdminProducts)
const mockedCreateCategory = vi.mocked(createAdminCategory)
const mockedUpdateCategory = vi.mocked(updateAdminCategory)

const category = {
  id: 'a'.repeat(24),
  name: 'Hambúrgueres',
  slug: 'hamburgeres',
  description: '',
  sortOrder: 0,
  active: true,
  createdAt: '2026-11-23T00:00:00Z',
  updatedAt: '2026-11-23T00:00:00Z',
}

const page = {
  data: [category],
  meta: { page: 1, pageSize: 20, total: 1, totalPages: 1 },
}

describe('adminCatalog stores', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    mockedFetchCategories.mockResolvedValue(page)
    mockedFetchProducts.mockResolvedValue({ data: [], meta: page.meta })
  })

  it('carrega categorias e expõe estado de sucesso', async () => {
    const store = useAdminCategoriesStore()

    await store.load()

    expect(store.state).toBe('success')
    expect(store.items).toEqual([category])
    expect(store.totalPages).toBe(1)
    expect(store.isEmpty).toBe(false)
  })

  it('mostra mensagem amigável quando o carregamento falha', async () => {
    mockedFetchCategories.mockRejectedValue(new ApiError(401, 'UNAUTHORIZED', 'x'))
    const store = useAdminCategoriesStore()

    await store.load()

    expect(store.state).toBe('error')
    expect(store.error).toContain('sessão expirou')
    expect(store.items).toEqual([])
  })

  it('normaliza filtros ao aplicar e volta para a primeira página', async () => {
    const store = useAdminCategoriesStore()
    await store.load()

    await store.applyFilters({ page: 3, q: '  ', sort: 'name' })

    expect(mockedFetchCategories).toHaveBeenLastCalledWith({
      page: 1,
      pageSize: 20,
      sort: 'name',
      order: 'asc',
    })
  })

  it('cria, recarrega a lista e avisa o operador', async () => {
    mockedCreateCategory.mockResolvedValue(category)
    const store = useAdminCategoriesStore()
    await store.load()
    const input = {
      name: 'Hambúrgueres',
      slug: 'hamburgeres',
      description: '',
      sortOrder: 0,
      active: true,
    }

    await expect(store.save(null, input)).resolves.toEqual(category)

    expect(mockedCreateCategory).toHaveBeenCalledWith(input)
    expect(mockedFetchCategories).toHaveBeenCalledTimes(2)
    expect(store.saveSuccess).toContain('criado')
  })

  it('edita um registro existente e reporta erro de validação', async () => {
    const store = useAdminCategoriesStore()
    await store.load()

    mockedUpdateCategory.mockRejectedValue(
      new ApiError(422, 'VALIDATION_ERROR', 'Dados inválidos.', [
        { field: 'slug', message: 'Slug já existe.' },
      ]),
    )

    await expect(store.save(category.id, { active: false })).resolves.toBeNull()
    expect(store.saveError).toBe('Slug já existe.')
  })

  it('mantém o filtro por categoria no store de produtos', async () => {
    const store = useAdminProductsStore()

    await store.applyFilters({ categoryId: 'b'.repeat(24), active: true })

    expect(mockedFetchProducts).toHaveBeenCalledWith({
      page: 1,
      pageSize: 20,
      sort: 'sortOrder',
      order: 'asc',
      categoryId: 'b'.repeat(24),
      active: true,
    })
  })
})
