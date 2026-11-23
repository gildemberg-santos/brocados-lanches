import { get, patchData, postData } from './http'
import type { AdminPage } from '../types/admin'
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

const BASE_PATH = '/api/v1/admin'

function buildQuery(query: AdminCatalogListQuery): string {
  const params = new URLSearchParams()
  const values: Record<string, string | number | boolean | undefined> = {
    page: query.page,
    pageSize: query.pageSize,
    q: query.q,
    active: query.active,
    sort: query.sort,
    order: query.order,
    categoryId: query.categoryId,
  }

  for (const [key, value] of Object.entries(values)) {
    if (value !== undefined && value !== '') {
      params.set(key, String(value))
    }
  }

  const serialized = params.toString()
  return serialized ? `?${serialized}` : ''
}

function path(resource: string, query: AdminCatalogListQuery = {}): string {
  return `${BASE_PATH}/${resource}${buildQuery(query)}`
}

export function fetchAdminCategories(
  query: AdminCatalogListQuery = {},
): Promise<AdminPage<AdminCategory>> {
  return get<AdminPage<AdminCategory>>(path('categories', query))
}

export function createAdminCategory(input: AdminCategoryInput): Promise<AdminCategory> {
  return postData<AdminCategory>(path('categories'), input)
}

export function updateAdminCategory(
  id: string,
  input: Partial<AdminCategoryInput>,
): Promise<AdminCategory> {
  return patchData<AdminCategory>(`${path('categories')}/${encodeURIComponent(id)}`, input)
}

export function fetchAdminAddons(
  query: AdminCatalogListQuery = {},
): Promise<AdminPage<AdminAddon>> {
  return get<AdminPage<AdminAddon>>(path('addons', query))
}

export function createAdminAddon(input: AdminAddonInput): Promise<AdminAddon> {
  return postData<AdminAddon>(path('addons'), input)
}

export function updateAdminAddon(id: string, input: Partial<AdminAddonInput>): Promise<AdminAddon> {
  return patchData<AdminAddon>(`${path('addons')}/${encodeURIComponent(id)}`, input)
}

export function fetchAdminZones(query: AdminCatalogListQuery = {}): Promise<AdminPage<AdminZone>> {
  return get<AdminPage<AdminZone>>(path('delivery-zones', query))
}

export function createAdminZone(input: AdminZoneInput): Promise<AdminZone> {
  return postData<AdminZone>(path('delivery-zones'), input)
}

export function updateAdminZone(id: string, input: Partial<AdminZoneInput>): Promise<AdminZone> {
  return patchData<AdminZone>(`${path('delivery-zones')}/${encodeURIComponent(id)}`, input)
}

export function fetchAdminProducts(
  query: AdminCatalogListQuery = {},
): Promise<AdminPage<AdminProduct>> {
  return get<AdminPage<AdminProduct>>(path('products', query))
}

export function createAdminProduct(input: AdminProductInput): Promise<AdminProduct> {
  return postData<AdminProduct>(path('products'), input)
}

export function updateAdminProduct(
  id: string,
  input: Partial<AdminProductInput>,
): Promise<AdminProduct> {
  return patchData<AdminProduct>(`${path('products')}/${encodeURIComponent(id)}`, input)
}
