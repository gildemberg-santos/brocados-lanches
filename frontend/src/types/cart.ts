import type { CatalogProduct } from './catalog'

/** Estado persistido do carrinho: apenas IDs, escolhas e quantidades. */
export interface CartItem {
  key: string
  productId: string
  optionIds: string[]
  addonIds: string[]
  quantity: number
  note: string
}

export interface CartSnapshot {
  version: number
  items: CartItem[]
}

/** Item enriquecido com dados do catálogo; nunca persistido. */
export interface CartLine {
  key: string
  product: CatalogProduct
  optionNames: string[]
  addonNames: string[]
  note: string
  quantity: number
  unitPriceCents: number
  lineTotalCents: number
}
