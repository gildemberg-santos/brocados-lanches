import { normalizePhone, type CheckoutForm, type Fulfillment, type PaymentMethod } from './checkout'

export const CHECKOUT_PREFERENCES_KEY = 'brocados:checkout-preferences'

const PREFERENCES_VERSION = 1
const MAX_TEXT_LENGTH = 160
const FULFILLMENTS: readonly Fulfillment[] = ['delivery', 'pickup']
const PAYMENT_METHODS: readonly PaymentMethod[] = ['pix', 'card', 'cash']

/** Observação e troco ficam de fora: mudam a cada pedido e não são preferências. */
export type CheckoutPreferences = Pick<
  CheckoutForm,
  | 'fulfillment'
  | 'zoneId'
  | 'customerName'
  | 'phone'
  | 'street'
  | 'number'
  | 'complement'
  | 'district'
  | 'city'
  | 'state'
  | 'zipCode'
  | 'paymentMethod'
>

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}

function readText(value: unknown, maxLength = MAX_TEXT_LENGTH): string {
  return typeof value === 'string' ? value.trim().slice(0, maxLength) : ''
}

function readId(value: unknown): string {
  const text = readText(value, 64)
  return /^[A-Za-z0-9_-]+$/.test(text) ? text : ''
}

function readState(value: unknown): string {
  const text = readText(value, 2).toUpperCase()
  return /^[A-Z]{2}$/.test(text) ? text : ''
}

function readEnum<T extends string>(value: unknown, allowed: readonly T[], fallback: T): T {
  return allowed.find((candidate) => candidate === value) ?? fallback
}

export function buildCheckoutPreferences(form: CheckoutForm): CheckoutPreferences {
  return {
    fulfillment: form.fulfillment,
    zoneId: readId(form.zoneId),
    customerName: readText(form.customerName),
    phone: readText(form.phone, 20),
    street: readText(form.street),
    number: readText(form.number, 20),
    complement: readText(form.complement),
    district: readText(form.district),
    city: readText(form.city, 80),
    state: readState(form.state),
    zipCode: readText(form.zipCode, 20),
    paymentMethod: form.paymentMethod,
  }
}

/** Chamado somente depois que o pedido foi gravado no servidor. */
export function saveCheckoutPreferences(form: CheckoutForm): void {
  if (typeof localStorage === 'undefined') {
    return
  }

  try {
    localStorage.setItem(
      CHECKOUT_PREFERENCES_KEY,
      JSON.stringify({ version: PREFERENCES_VERSION, preferences: buildCheckoutPreferences(form) }),
    )
  } catch {
    // Preferências são conveniência: falha de armazenamento não afeta o pedido.
  }
}

export function loadCheckoutPreferences(): CheckoutPreferences | null {
  if (typeof localStorage === 'undefined') {
    return null
  }

  try {
    const raw = localStorage.getItem(CHECKOUT_PREFERENCES_KEY)
    if (!raw) {
      return null
    }

    const parsed: unknown = JSON.parse(raw)
    if (!isRecord(parsed) || parsed.version !== PREFERENCES_VERSION) {
      return null
    }

    const stored = parsed.preferences
    if (!isRecord(stored)) {
      return null
    }

    return {
      fulfillment: readEnum(stored.fulfillment, FULFILLMENTS, 'delivery'),
      zoneId: readId(stored.zoneId),
      customerName: readText(stored.customerName),
      phone: readText(stored.phone, 20),
      street: readText(stored.street),
      number: readText(stored.number, 20),
      complement: readText(stored.complement),
      district: readText(stored.district),
      city: readText(stored.city, 80),
      state: readState(stored.state),
      zipCode: readText(stored.zipCode, 20),
      paymentMethod: readEnum(stored.paymentMethod, PAYMENT_METHODS, 'cash'),
    }
  } catch {
    return null
  }
}

/** Só completa os dados associados quando o telefone identifica o mesmo cliente. */
export function checkoutPreferencesMatchPhone(
  preferences: CheckoutPreferences,
  phone: string,
): boolean {
  const savedPhone = normalizePhone(preferences.phone)
  return savedPhone !== '' && savedPhone === normalizePhone(phone)
}

/** Aplica o que depende do cliente identificado; nunca sobrescreve nome/telefone. */
export function applyCheckoutPreferences(
  form: CheckoutForm,
  preferences: CheckoutPreferences,
  activeZoneIds: readonly string[] = [],
): CheckoutForm {
  return {
    ...form,
    fulfillment: preferences.fulfillment,
    zoneId: activeZoneIds.includes(preferences.zoneId) ? preferences.zoneId : '',
    street: preferences.street,
    number: preferences.number,
    complement: preferences.complement,
    district: preferences.district,
    city: preferences.city,
    state: preferences.state,
    zipCode: preferences.zipCode,
    paymentMethod: preferences.paymentMethod,
  }
}

export function hasCheckoutPreferences(): boolean {
  if (typeof localStorage === 'undefined') {
    return false
  }

  try {
    return localStorage.getItem(CHECKOUT_PREFERENCES_KEY) !== null
  } catch {
    return false
  }
}

export function clearCheckoutPreferences(): void {
  if (typeof localStorage === 'undefined') {
    return
  }

  try {
    localStorage.removeItem(CHECKOUT_PREFERENCES_KEY)
  } catch {
    // Nada a fazer: o botão apenas informa que não foi possível apagar.
  }
}
