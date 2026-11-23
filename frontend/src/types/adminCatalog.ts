/** Tipos do catálogo administrativo. Valores monetários são sempre centavos inteiros. */

export interface AdminCategory {
  id: string
  name: string
  slug: string
  description: string
  sortOrder: number
  active: boolean
  createdAt: string
  updatedAt: string
}

export interface AdminAddon {
  id: string
  name: string
  slug: string
  priceCents: number
  active: boolean
  sortOrder: number
  createdAt: string
  updatedAt: string
}

export interface AdminZone {
  id: string
  name: string
  slug: string
  feeCents: number
  active: boolean
  sortOrder: number
  createdAt: string
  updatedAt: string
}

export interface AdminOption {
  id: string
  name: string
  priceCents: number
  available: boolean
  sortOrder: number
}

export interface AdminOptionGroup {
  id: string
  name: string
  required: boolean
  minSelections: number
  maxSelections: number
  options: AdminOption[]
}

export interface AdminProduct {
  id: string
  categoryId: string
  name: string
  slug: string
  description: string
  priceCents: number
  imageUrl: string
  imageAlt: string
  allergens: string[]
  available: boolean
  featured: boolean
  active: boolean
  sortOrder: number
  optionGroups: AdminOptionGroup[]
  addonIds: string[]
  createdAt: string
  updatedAt: string
}

export type AdminWeekday =
  'sunday' | 'monday' | 'tuesday' | 'wednesday' | 'thursday' | 'friday' | 'saturday'

export interface AdminTimeInterval {
  open: string
  close: string
}

export interface AdminOpeningHour {
  day: AdminWeekday
  closed: boolean
  intervals: AdminTimeInterval[]
}

export interface AdminStoreSettings {
  name: string
  isOpen: boolean
  openingHours: AdminOpeningHour[]
  deliveryTime: string
  whatsapp: string
  instagram: string
  address: string
  createdAt: string
  updatedAt: string
}

export interface AdminCatalogListQuery {
  page?: number
  pageSize?: number
  q?: string
  active?: boolean
  sort?: string
  order?: 'asc' | 'desc'
  categoryId?: string
}

export interface AdminCategoryInput {
  name: string
  slug: string
  description: string
  sortOrder: number
  active: boolean
}

export interface AdminAddonInput {
  name: string
  slug: string
  priceCents: number
  sortOrder: number
  active: boolean
}

export interface AdminZoneInput {
  name: string
  slug: string
  feeCents: number
  sortOrder: number
  active: boolean
}

export interface AdminProductInput {
  categoryId: string
  name: string
  slug: string
  description: string
  priceCents: number
  imageUrl: string
  imageAlt: string
  allergens: string[]
  available: boolean
  featured: boolean
  active: boolean
  sortOrder: number
  optionGroups: AdminOptionGroup[]
  addonIds: string[]
}

export type AdminStoreInput = Omit<AdminStoreSettings, 'createdAt' | 'updatedAt'>

export type AdminCatalogKind = 'category' | 'addon' | 'zone'

export type AdminCatalogInput = AdminCategoryInput | AdminAddonInput | AdminZoneInput
