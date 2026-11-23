import { beforeEach, describe, expect, it, vi } from 'vitest'

import { get, patchData, postData } from './http'
import {
  createAdminCategory,
  fetchAdminAddons,
  fetchAdminCategories,
  fetchAdminProducts,
  fetchAdminZones,
  updateAdminProduct,
  updateAdminZone,
} from './adminCatalog'

vi.mock('./http', () => ({
  get: vi.fn(),
  patchData: vi.fn(),
  postData: vi.fn(),
}))

const mockedGet = vi.mocked(get)
const mockedPatchData = vi.mocked(patchData)
const mockedPostData = vi.mocked(postData)

describe('adminCatalog service', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('carrega listagens sem query quando nenhum filtro é informado', async () => {
    const page = { data: [], meta: { page: 1, pageSize: 20, total: 0, totalPages: 0 } }
    mockedGet.mockResolvedValue(page)

    await expect(fetchAdminCategories()).resolves.toEqual(page)
    expect(mockedGet).toHaveBeenCalledWith('/api/v1/admin/categories')
  })

  it('usa o caminho correto para adicionais, zonas e produtos', async () => {
    mockedGet.mockResolvedValue({ data: [], meta: {} })

    await fetchAdminAddons()
    await fetchAdminZones()
    await fetchAdminProducts()

    expect(mockedGet).toHaveBeenNthCalledWith(1, '/api/v1/admin/addons')
    expect(mockedGet).toHaveBeenNthCalledWith(2, '/api/v1/admin/delivery-zones')
    expect(mockedGet).toHaveBeenNthCalledWith(3, '/api/v1/admin/products')
  })

  it('serializa filtros de catálogo, incluindo active=false', async () => {
    mockedGet.mockResolvedValue({ data: [], meta: {} })

    await fetchAdminProducts({
      page: 2,
      pageSize: 10,
      q: 'hambúrguer',
      active: false,
      sort: 'priceCents',
      order: 'desc',
      categoryId: 'a'.repeat(24),
    })

    expect(mockedGet).toHaveBeenCalledWith(
      `/api/v1/admin/products?page=2&pageSize=10&q=hamb%C3%BArguer&active=false&sort=priceCents&order=desc&categoryId=${'a'.repeat(24)}`,
    )
  })

  it('cria e atualiza categorias e produtos com o corpo informado', async () => {
    mockedPostData.mockResolvedValue({ id: 'new' })
    mockedPatchData.mockResolvedValue({ id: 'updated' })
    const input = {
      name: 'Categoria',
      slug: 'categoria',
      description: '',
      sortOrder: 0,
      active: true,
    }

    await expect(createAdminCategory(input)).resolves.toEqual({ id: 'new' })
    expect(mockedPostData).toHaveBeenCalledWith('/api/v1/admin/categories', input)

    await expect(updateAdminProduct('abc/123', { active: false })).resolves.toEqual({
      id: 'updated',
    })
    expect(mockedPatchData).toHaveBeenCalledWith('/api/v1/admin/products/abc%2F123', {
      active: false,
    })
  })

  it('codifica o identificador na atualização de zonas', async () => {
    mockedPatchData.mockResolvedValue({ id: 'abc' })

    await updateAdminZone('abc/123', { feeCents: 500 })

    expect(mockedPatchData).toHaveBeenCalledWith('/api/v1/admin/delivery-zones/abc%2F123', {
      feeCents: 500,
    })
  })
})
