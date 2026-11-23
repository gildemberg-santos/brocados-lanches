export interface MenuOption {
  id: string
  name: string
  priceCents: number
  available: boolean
}

export interface MenuOptionGroup {
  id: string
  name: string
  required: boolean
  minSelections: number
  maxSelections: number
  options: MenuOption[]
}

export interface MenuAddon {
  id: string
  name: string
  priceCents: number
  available: boolean
}

export interface MenuProduct {
  id: string
  name: string
  slug: string
  description: string
  priceCents: number
  imageUrl: string
  imageAlt: string
  allergens: string[]
  available: boolean
  featured: boolean
  sortOrder: number
  optionGroups: MenuOptionGroup[]
  addons: MenuAddon[]
}

export interface MenuCategory {
  id: string
  name: string
  slug: string
  sortOrder: number
  products: MenuProduct[]
}

export interface Menu {
  categories: MenuCategory[]
  updatedAt?: string
}

export interface CatalogProduct extends MenuProduct {
  categoryId: string
  categoryName: string
  categorySlug: string
}
