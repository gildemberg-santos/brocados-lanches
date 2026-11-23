import type {
  CatalogProduct,
  Menu,
  MenuAddon,
  MenuCategory,
  MenuOption,
  MenuOptionGroup,
  MenuProduct,
} from '../../types/catalog'

export function makeOption(overrides: Partial<MenuOption> = {}): MenuOption {
  return {
    id: 'option-1',
    name: 'Ao ponto',
    priceCents: 0,
    available: true,
    ...overrides,
  }
}

export function makeOptionGroup(overrides: Partial<MenuOptionGroup> = {}): MenuOptionGroup {
  return {
    id: 'group-1',
    name: 'Ponto da carne',
    required: true,
    minSelections: 1,
    maxSelections: 1,
    options: [makeOption()],
    ...overrides,
  }
}

export function makeAddon(overrides: Partial<MenuAddon> = {}): MenuAddon {
  return {
    id: 'addon-1',
    name: 'Bacon crocante',
    priceCents: 500,
    available: true,
    ...overrides,
  }
}

export function makeProduct(overrides: Partial<MenuProduct> = {}): MenuProduct {
  return {
    id: 'product-1',
    name: 'Brocado Cheddar',
    slug: 'brocado-cheddar',
    description: 'Blend artesanal com cheddar.',
    priceCents: 2300,
    imageUrl: '',
    imageAlt: '',
    allergens: [],
    available: true,
    featured: true,
    sortOrder: 0,
    optionGroups: [],
    addons: [],
    ...overrides,
  }
}

export function makeCatalogProduct(overrides: Partial<CatalogProduct> = {}): CatalogProduct {
  return {
    ...makeProduct(),
    categoryId: 'category-1',
    categoryName: 'Hambúrgueres',
    categorySlug: 'hamburgueres',
    ...overrides,
  }
}

export function makeCategory(overrides: Partial<MenuCategory> = {}): MenuCategory {
  return {
    id: 'category-1',
    name: 'Hambúrgueres',
    slug: 'hamburgueres',
    sortOrder: 0,
    products: [makeProduct()],
    ...overrides,
  }
}

export function makeMenu(overrides: Partial<Menu> = {}): Menu {
  return {
    categories: [makeCategory()],
    updatedAt: '2026-11-21T12:00:00Z',
    ...overrides,
  }
}
