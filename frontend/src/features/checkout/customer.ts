import type { CustomerProfileLookup, PaymentMethod } from '../../types/order'
import { isValidBrazilianPhone, type CheckoutForm } from './checkout'

const paymentMethods: readonly PaymentMethod[] = ['cash', 'card', 'pix']

/** Só consulta o servidor com telefone brasileiro completo. */
export function canLookupCustomer(phone: string): boolean {
  return isValidBrazilianPhone(phone)
}

/**
 * Aplica a sugestão do servidor sem sobrescrever identidade já digitada,
 * observação ou troco. Zona inativa é descartada; preço e taxa continuam
 * sendo calculados pela cotação autoritativa.
 */
export function applyCustomerProfile(
  form: CheckoutForm,
  profile: CustomerProfileLookup,
  activeZoneIds: readonly string[] = [],
): CheckoutForm {
  const next: CheckoutForm = { ...form }

  if (form.customerName.trim() === '' && profile.name?.trim()) {
    next.customerName = profile.name
  }
  if (profile.fulfillment === 'delivery' || profile.fulfillment === 'pickup') {
    next.fulfillment = profile.fulfillment
  }
  if (profile.zone) {
    next.zoneId = activeZoneIds.includes(profile.zone.id) ? profile.zone.id : ''
  }
  if (profile.address) {
    // A API omite campos vazios opcionais; o formulário exige strings para
    // formatar e enviar o pedido sem quebrar no trim.
    next.street = profile.address.street ?? ''
    next.number = profile.address.number ?? ''
    next.complement = profile.address.complement ?? ''
    next.district = profile.address.district ?? ''
    next.city = profile.address.city ?? ''
    next.state = profile.address.state ?? ''
    next.zipCode = profile.address.zipCode ?? ''
  }
  if (profile.paymentMethod && paymentMethods.includes(profile.paymentMethod)) {
    next.paymentMethod = profile.paymentMethod
  }
  return next
}
