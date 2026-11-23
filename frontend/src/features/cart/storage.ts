import type { CartItem, CartSnapshot } from '../../types/cart'
import { createCartKey, MAX_CART_NOTE_LENGTH, normalizeQuantity } from './cart'

export const CART_STORAGE_KEY = 'brocados.cart'
export const CART_SCHEMA_VERSION = 1

export interface LoadedCart {
  snapshot: CartSnapshot
  notice: string
}

function getStorage(): Storage | null {
  try {
    return typeof localStorage === 'undefined' ? null : localStorage
  } catch {
    return null
  }
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}

function parseStringArray(value: unknown): string[] | null {
  if (!Array.isArray(value) || value.some((entry) => typeof entry !== 'string')) {
    return null
  }

  return [...new Set(value as string[])].filter((entry) => entry.length > 0)
}

function parseCartItem(value: unknown): CartItem | null {
  if (!isRecord(value) || typeof value.productId !== 'string' || value.productId.length === 0) {
    return null
  }

  const optionIds = parseStringArray(value.optionIds)
  const addonIds = parseStringArray(value.addonIds)

  if (
    optionIds === null ||
    addonIds === null ||
    typeof value.quantity !== 'number' ||
    value.quantity < 1 ||
    typeof value.note !== 'string'
  ) {
    return null
  }

  const note = value.note.trim().slice(0, MAX_CART_NOTE_LENGTH)

  return {
    key: createCartKey({ productId: value.productId, optionIds, addonIds, note }),
    productId: value.productId,
    optionIds,
    addonIds,
    quantity: normalizeQuantity(value.quantity),
    note,
  }
}

function emptyCart(): CartSnapshot {
  return { version: CART_SCHEMA_VERSION, items: [] }
}

export function saveCart(snapshot: CartSnapshot): void {
  try {
    getStorage()?.setItem(CART_STORAGE_KEY, JSON.stringify(snapshot))
  } catch {
    // Armazenamento indisponível (modo privado ou cota cheia): o carrinho segue em memória.
  }
}

function resetCart(notice: string): LoadedCart {
  try {
    getStorage()?.removeItem(CART_STORAGE_KEY)
  } catch {
    // Nada a fazer se o armazenamento não puder ser alterado.
  }

  return { snapshot: emptyCart(), notice }
}

export function loadCart(): LoadedCart {
  let raw: string | null = null

  try {
    raw = getStorage()?.getItem(CART_STORAGE_KEY) ?? null
  } catch {
    raw = null
  }

  if (raw === null) {
    return { snapshot: emptyCart(), notice: '' }
  }

  let parsed: unknown

  try {
    parsed = JSON.parse(raw)
  } catch {
    return resetCart('Não foi possível ler o carrinho salvo. Ele foi reiniciado.')
  }

  if (!isRecord(parsed) || parsed.version !== CART_SCHEMA_VERSION || !Array.isArray(parsed.items)) {
    return resetCart('O formato do carrinho mudou. Os itens anteriores foram removidos.')
  }

  const items: CartItem[] = []

  for (const entry of parsed.items) {
    const item = parseCartItem(entry)

    if (item !== null) {
      items.push(item)
    }
  }

  const snapshot: CartSnapshot = { version: CART_SCHEMA_VERSION, items }

  if (items.length !== parsed.items.length) {
    saveCart(snapshot)

    return { snapshot, notice: 'Itens inválidos foram removidos do carrinho.' }
  }

  return { snapshot, notice: '' }
}
