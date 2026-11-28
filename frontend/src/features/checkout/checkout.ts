import { ApiError } from '../../services/http'
import type { CartItem } from '../../types/cart'
import type {
  CreateOrderRequest,
  Fulfillment,
  OrderQuoteRequest,
  PaymentMethod,
} from '../../types/order'

export type { Fulfillment, PaymentMethod }

export interface CheckoutForm {
  fulfillment: Fulfillment
  zoneId: string
  customerName: string
  phone: string
  street: string
  number: string
  complement: string
  district: string
  city: string
  state: string
  zipCode: string
  paymentMethod: PaymentMethod
  changeFor: string
  note: string
  saveCustomerProfile: boolean
}

export type CheckoutErrors = Partial<Record<keyof CheckoutForm, string>>

export const initialCheckoutForm: CheckoutForm = {
  fulfillment: 'delivery',
  zoneId: '',
  customerName: '',
  phone: '',
  street: '',
  number: '',
  complement: '',
  district: '',
  city: 'Aquiraz',
  state: 'CE',
  zipCode: '',
  paymentMethod: 'cash',
  changeFor: '',
  note: '',
  saveCustomerProfile: true,
}

export function createIdempotencyKey(): string {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID()
  }

  const bytes = new Uint8Array(16)
  if (typeof crypto !== 'undefined' && typeof crypto.getRandomValues === 'function') {
    crypto.getRandomValues(bytes)
  } else {
    for (let index = 0; index < bytes.length; index += 1) {
      bytes[index] = Math.floor(Math.random() * 256)
    }
  }
  return Array.from(bytes, (byte) => byte.toString(16).padStart(2, '0')).join('')
}

export function normalizePhone(value: string): string {
  const digits = value.replace(/\D/g, '')
  if (digits.length === 10 || digits.length === 11) {
    return `+55${digits}`
  }
  if ((digits.length === 12 || digits.length === 13) && digits.startsWith('55')) {
    return `+${digits}`
  }
  return digits === '' ? '' : `+${digits}`
}

/** DDDs em uso conforme Anatel; celular tem 9 dígitos e começa com 9, fixo tem 8. */
const BRAZILIAN_PHONE_PATTERN =
  /^\+55(1[1-9]|2[12478]|3[1-578]|4[1-9]|5[1345]|6[1-9]|7[134579]|8[1-9]|9[1-9])(9\d{8}|[2-5]\d{7})$/

export function isValidBrazilianPhone(value: string): boolean {
  return BRAZILIAN_PHONE_PATTERN.test(normalizePhone(value))
}

export function formatPhone(value: string): string {
  const digits = value.replace(/\D/g, '')
  // Quem cola o número em E.164 (+55) não pode ter o DDD trocado pelo 55 do país.
  const national =
    (digits.length === 12 || digits.length === 13) && digits.startsWith('55')
      ? digits.slice(2)
      : digits
  const masked = national.slice(0, 11)
  if (masked.length <= 2) {
    return masked === '' ? '' : `(${masked}`
  }
  if (masked.length <= 6) {
    return `(${masked.slice(0, 2)}) ${masked.slice(2)}`
  }
  if (masked.length <= 10) {
    return `(${masked.slice(0, 2)}) ${masked.slice(2, 6)}-${masked.slice(6)}`
  }
  return `(${masked.slice(0, 2)}) ${masked.slice(2, 7)}-${masked.slice(7)}`
}

/** Converte texto de dinheiro em centavos inteiros, sem parseFloat. */
export function parseCentsInput(value: string): number | null {
  const digits = value.replace(/\D/g, '')
  if (digits === '') {
    return null
  }

  const hasFractionSeparator = /[.,]/.test(value)
  const amount = Number(hasFractionSeparator ? digits : `${digits}00`)
  if (!Number.isSafeInteger(amount) || amount < 0) {
    return null
  }
  return amount
}

export function formatCents(cents: number): string {
  return new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(cents / 100)
}

export function buildQuoteRequest(items: CartItem[], form: CheckoutForm): OrderQuoteRequest {
  return {
    items: items.map((item) => ({
      productId: item.productId,
      quantity: item.quantity,
      optionIds: [...item.optionIds],
      addonIds: [...item.addonIds],
      note: item.note,
    })),
    fulfillment: form.fulfillment,
    zoneId: form.fulfillment === 'delivery' ? form.zoneId : undefined,
  }
}

export function buildCreateOrderRequest(
  items: CartItem[],
  form: CheckoutForm,
  quotedTotalCents: number,
): CreateOrderRequest {
  const request: CreateOrderRequest = {
    ...buildQuoteRequest(items, form),
    expectedTotalCents: quotedTotalCents,
    customer: {
      name: (form.customerName ?? '').trim(),
      phone: normalizePhone(form.phone),
    },
    payment: {
      method: form.paymentMethod,
    },
    note: (form.note ?? '').trim(),
    saveCustomerProfile: form.saveCustomerProfile,
  }

  if (form.fulfillment === 'delivery') {
    request.address = {
      street: (form.street ?? '').trim(),
      number: (form.number ?? '').trim(),
      complement: (form.complement ?? '').trim(),
      district: (form.district ?? '').trim(),
      city: (form.city ?? '').trim(),
      state: (form.state ?? '').trim().toUpperCase(),
      zipCode: (form.zipCode ?? '').trim(),
    }
  }
  if (form.paymentMethod === 'cash') {
    const changeForCents = parseCentsInput(form.changeFor)
    if (changeForCents !== null) {
      request.payment.changeForCents = changeForCents
    }
  }

  return request
}

export function validateCheckoutForm(
  form: CheckoutForm,
  totalCents: number | null,
): CheckoutErrors {
  const errors: CheckoutErrors = {}

  if (form.customerName.trim().length < 3) {
    errors.customerName = 'Informe seu nome completo.'
  }
  if (!isValidBrazilianPhone(form.phone)) {
    errors.phone = 'Informe um telefone brasileiro válido, com DDD.'
  }

  if (form.fulfillment === 'delivery') {
    if (!form.zoneId) {
      errors.zoneId = 'Escolha uma zona de entrega.'
    }
    if (!form.street.trim()) errors.street = 'Informe a rua.'
    if (!form.number.trim()) errors.number = 'Informe o número.'
    if (!form.district.trim()) errors.district = 'Informe o bairro.'
    if (!form.city.trim()) errors.city = 'Informe a cidade.'
    if (!/^[A-Za-z]{2}$/.test(form.state.trim())) errors.state = 'Use a sigla do estado.'
    if (form.zipCode.replace(/\D/g, '').length !== 8)
      errors.zipCode = 'Informe o CEP com 8 dígitos.'
  }

  if (form.paymentMethod === 'cash' && totalCents !== null) {
    const changeForCents = parseCentsInput(form.changeFor)
    if (changeForCents === null || changeForCents < totalCents) {
      errors.changeFor = 'Informe um valor maior ou igual ao total do pedido.'
    }
  }
  if (form.note.length > 280) {
    errors.note = 'A observação pode ter até 280 caracteres.'
  }

  return errors
}

export function checkoutApiMessage(error: unknown): string {
  if (isPixUnavailableError(error)) {
    return 'O Pix não está disponível no momento. Escolha outra forma de pagamento.'
  }
  if (error instanceof ApiError) {
    const messages: Record<string, string> = {
      NOT_FOUND: 'Pedido não encontrado ou código de acompanhamento inválido.',
      STORE_CLOSED: 'A loja está fechada no momento. Volte ao cardápio para escolher outros itens.',
      ZONE_UNAVAILABLE: 'Essa zona de entrega não está disponível. Escolha outra.',
      PRODUCT_UNAVAILABLE: 'Um dos itens do carrinho não está mais disponível.',
      CONFLICT: 'Este pedido já foi enviado. Verifique a confirmação antes de tentar novamente.',
      VALIDATION_ERROR: 'Revise os dados informados e tente novamente.',
      RATE_LIMITED: 'Muitas tentativas. Aguarde um momento e tente novamente.',
    }
    return messages[error.code] ?? error.message
  }
  return 'Não foi possível enviar o pedido. Seus dados foram mantidos; tente novamente.'
}

function isPixUnavailableError(error: unknown): boolean {
  return (
    error instanceof ApiError &&
    error.code === 'VALIDATION_ERROR' &&
    error.details.some((detail) => detail.field === 'payment.method')
  )
}

/** Identifica o conflito de cotação para o checkout recotar sem apagar o carrinho. */
export function isQuoteOutdatedError(error: unknown): boolean {
  return (
    error instanceof ApiError &&
    error.code === 'CONFLICT' &&
    error.details.some(
      (detail) => detail.field === 'expectedTotalCents' && detail.code === 'OUTDATED',
    )
  )
}

export function orderStatusLabel(status: string): string {
  const labels: Record<string, string> = {
    pending: 'Aguardando confirmação',
    confirmed: 'Confirmado',
    preparing: 'Em preparo',
    ready: 'Pronto',
    out_for_delivery: 'Saiu para entrega',
    picked_up: 'Retirado',
    delivered: 'Entregue',
    cancelled: 'Cancelado',
  }
  return labels[status] ?? status
}
