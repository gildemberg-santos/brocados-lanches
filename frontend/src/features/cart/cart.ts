import type { CartItem, CartLine } from '../../types/cart'
import type { CatalogProduct } from '../../types/catalog'
import { calculateConfiguredPriceCents, validateOptionGroup } from '../catalog/customization'
import type { OptionSelections } from '../catalog/customization'

export const MAX_CART_QUANTITY = 99
export const MAX_CART_NOTE_LENGTH = 140

export function normalizeQuantity(value: number, max = MAX_CART_QUANTITY): number {
  if (!Number.isFinite(value)) {
    return 1
  }

  return Math.min(max, Math.max(1, Math.trunc(value)))
}

export function createCartKey(
  input: Pick<CartItem, 'productId' | 'optionIds' | 'addonIds' | 'note'>,
): string {
  return [
    input.productId,
    [...input.optionIds].sort().join(','),
    [...input.addonIds].sort().join(','),
    input.note.trim().toLocaleLowerCase('pt-BR'),
  ].join('|')
}

function sanitizeSelections(
  product: CatalogProduct,
  selections: OptionSelections,
): OptionSelections | null {
  const sanitized: OptionSelections = {}

  for (const group of product.optionGroups) {
    const availableIds = new Set(
      group.options.filter((option) => option.available).map((option) => option.id),
    )
    const ids = [...new Set(selections[group.id] ?? [])]

    if (ids.some((id) => !availableIds.has(id)) || validateOptionGroup(group, ids).length > 0) {
      return null
    }

    sanitized[group.id] = ids
  }

  return sanitized
}

function sanitizeAddonIds(product: CatalogProduct, addonIds: string[]): string[] | null {
  const availableIds = new Set(
    product.addons.filter((addon) => addon.available).map((addon) => addon.id),
  )
  const ids = [...new Set(addonIds)]

  return ids.some((id) => !availableIds.has(id)) ? null : ids
}

export function createCartItem(
  product: CatalogProduct,
  selections: OptionSelections,
  addonIds: string[],
  note: string,
  quantity = 1,
): CartItem | null {
  if (!product.available) {
    return null
  }

  const sanitizedSelections = sanitizeSelections(product, selections)
  const sanitizedAddonIds = sanitizeAddonIds(product, addonIds)

  if (sanitizedSelections === null || sanitizedAddonIds === null) {
    return null
  }

  const optionIds = [...new Set(Object.values(sanitizedSelections).flat())]
  const sanitizedNote = note.trim().slice(0, MAX_CART_NOTE_LENGTH)

  return {
    key: createCartKey({
      productId: product.id,
      optionIds,
      addonIds: sanitizedAddonIds,
      note: sanitizedNote,
    }),
    productId: product.id,
    optionIds,
    addonIds: sanitizedAddonIds,
    quantity: normalizeQuantity(quantity),
    note: sanitizedNote,
  }
}

export interface BuiltCart {
  lines: CartLine[]
  unavailableKeys: string[]
}

export function buildCartLines(items: CartItem[], products: CatalogProduct[]): BuiltCart {
  const lines: CartLine[] = []
  const unavailableKeys: string[] = []

  for (const item of items) {
    const product = products.find((candidate) => candidate.id === item.productId)
    const selections =
      product === undefined ? null : sanitizeSelections(product, selectionsFromItem(product, item))
    const addonIds = product === undefined ? null : sanitizeAddonIds(product, item.addonIds)

    if (product === undefined || selections === null || addonIds === null) {
      unavailableKeys.push(item.key)
      continue
    }

    const selectedOptionIds = new Set(Object.values(selections).flat())
    const optionNames = product.optionGroups.flatMap((group) => {
      const names = group.options
        .filter((option) => selectedOptionIds.has(option.id))
        .map((option) => option.name)

      return names.length > 0 ? [`${group.name}: ${names.join(', ')}`] : []
    })
    const addonNames = product.addons
      .filter((addon) => addonIds.includes(addon.id))
      .map((addon) => addon.name)
    const unitPriceCents = calculateConfiguredPriceCents(
      product,
      product.optionGroups,
      selections,
      product.addons,
      addonIds,
    )

    lines.push({
      key: item.key,
      product,
      optionNames,
      addonNames,
      note: item.note,
      quantity: item.quantity,
      unitPriceCents,
      lineTotalCents: unitPriceCents * item.quantity,
    })
  }

  return { lines, unavailableKeys }
}

function selectionsFromItem(product: CatalogProduct, item: CartItem): OptionSelections {
  return Object.fromEntries(
    product.optionGroups.map((group) => [
      group.id,
      group.options
        .filter((option) => item.optionIds.includes(option.id))
        .map((option) => option.id),
    ]),
  )
}

export function calculateSubtotalCents(lines: CartLine[]): number {
  return lines.reduce((total, line) => total + line.lineTotalCents, 0)
}
