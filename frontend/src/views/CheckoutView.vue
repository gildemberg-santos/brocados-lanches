<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { RouterLink, useRouter } from 'vue-router'

import BaseButton from '../components/BaseButton.vue'
import ErrorState from '../components/ErrorState.vue'
import LoadingSkeleton from '../components/LoadingSkeleton.vue'
import PriceText from '../components/PriceText.vue'
import { buildCartLines } from '../features/cart/cart'
import {
  buildCreateOrderRequest,
  buildQuoteRequest,
  checkoutApiMessage,
  createIdempotencyKey,
  formatCents,
  formatPhone,
  initialCheckoutForm,
  normalizePhone,
  validateCheckoutForm,
} from '../features/checkout/checkout'
import { applyCustomerProfile, canLookupCustomer } from '../features/checkout/customer'
import { saveHandoffUrl } from '../features/checkout/handoff'
import {
  applyCheckoutPreferences,
  checkoutPreferencesMatchPhone,
  clearCheckoutPreferences,
  loadCheckoutPreferences,
  saveCheckoutPreferences,
} from '../features/checkout/preferences'
import type { CheckoutErrors, CheckoutForm } from '../features/checkout/checkout'
import type { CheckoutPreferences } from '../features/checkout/preferences'
import { fetchCustomerProfile } from '../services/customers'
import { createOrder, fetchOrderQuote } from '../services/orders'
import { useCartStore } from '../stores/cart'
import { useDeliveryZonesStore } from '../stores/deliveryZones'
import { useMenuStore } from '../stores/menu'
import type { OrderQuote } from '../types/order'

const router = useRouter()
const cart = useCartStore()
const menu = useMenuStore()
const zones = useDeliveryZonesStore()

const form = ref<CheckoutForm>({ ...initialCheckoutForm })
const errors = ref<CheckoutErrors>({})
const quote = ref<OrderQuote | null>(null)
const quotedRequest = ref('')
const quoteLoading = ref(false)
const quoteError = ref('')
const submitting = ref(false)
const submitError = ref('')
const idempotencyKey = ref(createIdempotencyKey())
const lastRequestSignature = ref('')
const savedPreferences = ref<CheckoutPreferences | null>(null)
const profileLookupLoading = ref(false)
const profileLookupMessage = ref('')
let quoteTimer: ReturnType<typeof setTimeout> | null = null
let profileLookupTimer: ReturnType<typeof setTimeout> | null = null
let profileLookupSequence = 0
let lastLookedUpPhone = ''

const isLoading = computed(
  () =>
    menu.state === 'idle' ||
    menu.state === 'loading' ||
    zones.state === 'idle' ||
    zones.state === 'loading',
)
const lines = computed(() => buildCartLines(cart.items, menu.products).lines)
const missingDeliveryZone = computed(
  () => form.value.fulfillment === 'delivery' && form.value.zoneId === '',
)
const canQuote = computed(() => !cart.isEmpty && !missingDeliveryZone.value)
const quoteTotalCents = computed(() => quote.value?.totalCents ?? null)
const canSubmit = computed(
  () => !cart.isEmpty && !submitting.value && !quoteLoading.value && quote.value !== null,
)

function quoteSignature(): string {
  return JSON.stringify(buildQuoteRequest(cart.items, form.value))
}

async function requestQuote(): Promise<void> {
  if (!canQuote.value) {
    quote.value = null
    quotedRequest.value = ''
    quoteError.value = ''
    return
  }

  if (quoteTimer !== null) {
    clearTimeout(quoteTimer)
    quoteTimer = null
  }

  quoteLoading.value = true
  quoteError.value = ''
  try {
    const request = buildQuoteRequest(cart.items, form.value)
    const result = await fetchOrderQuote(request)
    quote.value = result
    quotedRequest.value = JSON.stringify(request)
  } catch (error: unknown) {
    quote.value = null
    quotedRequest.value = ''
    quoteError.value = checkoutApiMessage(error)
  } finally {
    quoteLoading.value = false
  }
}

function scheduleQuote(): void {
  if (quoteTimer !== null) {
    clearTimeout(quoteTimer)
  }
  if (!canQuote.value) {
    return
  }
  quoteTimer = setTimeout(() => {
    void requestQuote()
  }, 250)
}

function applySavedPreferences(preferences: CheckoutPreferences): void {
  form.value = applyCheckoutPreferences(
    form.value,
    preferences,
    zones.zones.map((zone) => zone.id),
  )
}

function clearProfileLookup(): void {
  if (profileLookupTimer !== null) {
    clearTimeout(profileLookupTimer)
    profileLookupTimer = null
  }
  profileLookupSequence += 1
  profileLookupLoading.value = false
  profileLookupMessage.value = ''
}

function scheduleCustomerLookup(): void {
  clearProfileLookup()
  if (!form.value.saveCustomerProfile || !canLookupCustomer(form.value.phone)) {
    return
  }

  const phone = form.value.phone
  const phoneKey = normalizePhone(phone)
  if (phoneKey === lastLookedUpPhone) {
    return
  }

  const sequence = profileLookupSequence
  profileLookupTimer = setTimeout(() => {
    void requestCustomerLookup(phone, phoneKey, sequence)
  }, 500)
}

function lookupStillCurrent(phoneKey: string, sequence: number): boolean {
  return (
    sequence === profileLookupSequence &&
    form.value.saveCustomerProfile &&
    normalizePhone(form.value.phone) === phoneKey
  )
}

async function requestCustomerLookup(
  phone: string,
  phoneKey: string,
  sequence: number,
): Promise<void> {
  profileLookupTimer = null
  if (!lookupStillCurrent(phoneKey, sequence)) {
    return
  }

  profileLookupLoading.value = true
  profileLookupMessage.value = 'Buscando dados salvos para este telefone...'
  try {
    const response = await fetchCustomerProfile(phone)
    if (!lookupStillCurrent(phoneKey, sequence)) {
      return
    }
    lastLookedUpPhone = phoneKey
    if (response.found && response.profile) {
      form.value = applyCustomerProfile(
        form.value,
        response.profile,
        zones.zones.map((zone) => zone.id),
      )
      profileLookupMessage.value = 'Dados encontrados. Confira e ajuste se necessário.'
    } else {
      profileLookupMessage.value = ''
    }
  } catch {
    if (!lookupStillCurrent(phoneKey, sequence)) {
      return
    }
    lastLookedUpPhone = ''
    profileLookupMessage.value = 'Não foi possível buscar seus dados agora. Preencha manualmente.'
  } finally {
    if (sequence === profileLookupSequence) {
      profileLookupLoading.value = false
    }
  }
}

async function submit(): Promise<void> {
  submitError.value = ''
  errors.value = validateCheckoutForm(form.value, quoteTotalCents.value)

  const firstField = Object.keys(errors.value)[0]
  if (firstField) {
    await nextTick()
    document.querySelector<HTMLElement>(`[name="${firstField}"]`)?.focus()
    return
  }
  if (!quote.value) {
    submitError.value = 'Não foi possível calcular o total. Tente novamente.'
    return
  }

  const request = buildCreateOrderRequest(cart.items, form.value)
  // A mesma chave deve ser reutilizada em retries: se a resposta se perdeu,
  // o servidor devolve o pedido já gravado em vez de criar outro. Uma nova
  // chave só faz sentido quando o conteúdo do pedido muda.
  const requestSignature = JSON.stringify(request)
  if (requestSignature !== lastRequestSignature.value) {
    idempotencyKey.value = createIdempotencyKey()
    lastRequestSignature.value = requestSignature
  }
  if (quoteSignature() !== quotedRequest.value) {
    await requestQuote()
    if (!quote.value) {
      submitError.value = 'O carrinho mudou. Revise o pedido e tente novamente.'
      return
    }
  }

  submitting.value = true
  try {
    const created = await createOrder(request, idempotencyKey.value)
    saveHandoffUrl(created.publicId, created.whatsappUrl)
    if (form.value.saveCustomerProfile) {
      saveCheckoutPreferences(form.value)
    } else {
      clearCheckoutPreferences()
      savedPreferences.value = null
    }
    cart.clear()
    await router.push({
      name: 'order',
      params: { publicId: created.publicId },
      state: { whatsappUrl: created.whatsappUrl },
    })
  } catch (error: unknown) {
    submitError.value = checkoutApiMessage(error)
  } finally {
    submitting.value = false
  }
}

watch(
  () => [form.value.fulfillment, form.value.zoneId, cart.items] as const,
  () => {
    quote.value = null
    scheduleQuote()
  },
  { deep: true },
)

watch(
  () => form.value.phone,
  (phone) => {
    const preferences = savedPreferences.value
    if (preferences && checkoutPreferencesMatchPhone(preferences, phone)) {
      clearProfileLookup()
      applySavedPreferences(preferences)
      return
    }
    scheduleCustomerLookup()
  },
)

watch(
  () => form.value.saveCustomerProfile,
  (save) => {
    if (!save) {
      clearProfileLookup()
      clearCheckoutPreferences()
      savedPreferences.value = null
      return
    }
    scheduleCustomerLookup()
  },
)

onMounted(async () => {
  await Promise.all([menu.ensureLoaded(), zones.ensureLoaded()])

  const preferences = loadCheckoutPreferences()
  savedPreferences.value = preferences
  if (preferences) {
    form.value.customerName = preferences.customerName
    form.value.phone = preferences.phone
    if (checkoutPreferencesMatchPhone(preferences, form.value.phone)) {
      applySavedPreferences(preferences)
    }
  }

  if (menu.state === 'success') {
    cart.removeUnavailable(buildCartLines(cart.items, menu.products).unavailableKeys)
  }
  await requestQuote()
  if (quoteTimer !== null) {
    clearTimeout(quoteTimer)
    quoteTimer = null
  }
})

onUnmounted(() => {
  if (quoteTimer !== null) {
    clearTimeout(quoteTimer)
  }
  if (profileLookupTimer !== null) {
    clearTimeout(profileLookupTimer)
  }
  // Ignora a resposta de uma busca em voo quando a tela é fechada.
  profileLookupSequence += 1
})
</script>

<template>
  <main id="conteudo" tabindex="-1" class="page-shell checkout-page">
    <RouterLink class="back-link" :to="{ name: 'cart' }">← Voltar ao carrinho</RouterLink>
    <h1>Finalizar pedido</h1>
    <p class="checkout-preferences-note">
      Com sua autorização, podemos buscar e salvar suas escolhas neste navegador e no servidor para
      agilizar o próximo pedido. Você pode desativar no formulário e apagar na
      <RouterLink to="/privacidade">política de privacidade</RouterLink>.
    </p>

    <div v-if="cart.isEmpty" class="checkout-empty" role="status">
      <p>Seu carrinho está vazio. Escolha os itens antes de finalizar.</p>
      <RouterLink class="text-link" :to="{ name: 'menu' }">Ver cardápio</RouterLink>
    </div>

    <LoadingSkeleton v-else-if="isLoading" label="Carregando checkout" :count="3" />

    <div v-else class="checkout-layout">
      <form class="checkout-form" novalidate @submit.prevent="submit">
        <section class="checkout-section" aria-labelledby="customer-title">
          <h2 id="customer-title">1. Seus dados</h2>
          <p class="checkout-preferences-note">
            Informe o telefone para preencher as escolhas salvas neste navegador ou no servidor.
          </p>
          <label class="field">
            <span>Nome completo</span>
            <input
              v-model="form.customerName"
              id="checkout-customerName"
              name="customerName"
              type="text"
              autocomplete="name"
              :aria-invalid="errors.customerName ? 'true' : undefined"
              :aria-describedby="errors.customerName ? 'checkout-customerName-error' : undefined"
            />
            <small v-if="errors.customerName" id="checkout-customerName-error" class="field-error">
              {{ errors.customerName }}
            </small>
          </label>
          <label class="field">
            <span>Telefone</span>
            <input
              :value="form.phone"
              id="checkout-phone"
              name="phone"
              type="tel"
              inputmode="tel"
              autocomplete="tel"
              placeholder="(85) 99999-9999"
              :aria-invalid="errors.phone ? 'true' : undefined"
              :aria-describedby="errors.phone ? 'checkout-phone-error' : undefined"
              @input="form.phone = formatPhone(($event.target as HTMLInputElement).value)"
            />
            <small v-if="errors.phone" id="checkout-phone-error" class="field-error">
              {{ errors.phone }}
            </small>
          </label>
          <label class="choice choice--opt-in">
            <input v-model="form.saveCustomerProfile" name="saveCustomerProfile" type="checkbox" />
            <span>Buscar e salvar meus dados para agilizar meus pedidos</span>
          </label>
          <p
            v-if="profileLookupLoading || profileLookupMessage"
            class="checkout-preferences-note"
            role="status"
            aria-live="polite"
          >
            {{ profileLookupMessage || 'Buscando dados salvos para este telefone...' }}
          </p>
        </section>

        <section class="checkout-section" aria-labelledby="fulfillment-title">
          <h2 id="fulfillment-title">2. Recebimento</h2>
          <div class="choice-group">
            <label class="choice">
              <input v-model="form.fulfillment" name="fulfillment" type="radio" value="delivery" />
              <span>Entrega</span>
            </label>
            <label class="choice">
              <input v-model="form.fulfillment" name="fulfillment" type="radio" value="pickup" />
              <span>Retirada na loja</span>
            </label>
          </div>
        </section>

        <section
          v-if="form.fulfillment === 'delivery'"
          class="checkout-section"
          aria-labelledby="address-title"
        >
          <h2 id="address-title">3. Entrega</h2>
          <label class="field">
            <span>Zona de entrega</span>
            <select
              v-model="form.zoneId"
              id="checkout-zoneId"
              name="zoneId"
              :aria-invalid="errors.zoneId ? 'true' : undefined"
              :aria-describedby="errors.zoneId ? 'checkout-zoneId-error' : undefined"
            >
              <option value="">Selecione o bairro</option>
              <option v-for="zone in zones.zones" :key="zone.id" :value="zone.id">
                {{ zone.name }} — {{ formatCents(zone.feeCents) }}
              </option>
            </select>
            <small v-if="errors.zoneId" id="checkout-zoneId-error" class="field-error">
              {{ errors.zoneId }}
            </small>
          </label>

          <div class="field-grid">
            <label class="field field--wide">
              <span>Rua</span>
              <input
                v-model="form.street"
                id="checkout-street"
                name="street"
                type="text"
                autocomplete="street-address"
                :aria-invalid="errors.street ? 'true' : undefined"
                :aria-describedby="errors.street ? 'checkout-street-error' : undefined"
              />
              <small v-if="errors.street" id="checkout-street-error" class="field-error">
                {{ errors.street }}
              </small>
            </label>
            <label class="field">
              <span>Número</span>
              <input
                v-model="form.number"
                id="checkout-number"
                name="number"
                type="text"
                :aria-invalid="errors.number ? 'true' : undefined"
                :aria-describedby="errors.number ? 'checkout-number-error' : undefined"
              />
              <small v-if="errors.number" id="checkout-number-error" class="field-error">
                {{ errors.number }}
              </small>
            </label>
            <label class="field">
              <span>Complemento</span>
              <input v-model="form.complement" name="complement" type="text" />
            </label>
            <label class="field field--wide">
              <span>Bairro</span>
              <input
                v-model="form.district"
                id="checkout-district"
                name="district"
                type="text"
                :aria-invalid="errors.district ? 'true' : undefined"
                :aria-describedby="errors.district ? 'checkout-district-error' : undefined"
              />
              <small v-if="errors.district" id="checkout-district-error" class="field-error">
                {{ errors.district }}
              </small>
            </label>
            <label class="field">
              <span>CEP</span>
              <input
                v-model="form.zipCode"
                id="checkout-zipCode"
                name="zipCode"
                type="text"
                inputmode="numeric"
                :aria-invalid="errors.zipCode ? 'true' : undefined"
                :aria-describedby="errors.zipCode ? 'checkout-zipCode-error' : undefined"
              />
              <small v-if="errors.zipCode" id="checkout-zipCode-error" class="field-error">
                {{ errors.zipCode }}
              </small>
            </label>
            <label class="field">
              <span>Cidade</span>
              <input
                v-model="form.city"
                id="checkout-city"
                name="city"
                type="text"
                :aria-invalid="errors.city ? 'true' : undefined"
                :aria-describedby="errors.city ? 'checkout-city-error' : undefined"
              />
              <small v-if="errors.city" id="checkout-city-error" class="field-error">
                {{ errors.city }}
              </small>
            </label>
            <label class="field">
              <span>Estado</span>
              <input
                v-model="form.state"
                id="checkout-state"
                name="state"
                type="text"
                maxlength="2"
                :aria-invalid="errors.state ? 'true' : undefined"
                :aria-describedby="errors.state ? 'checkout-state-error' : undefined"
              />
              <small v-if="errors.state" id="checkout-state-error" class="field-error">
                {{ errors.state }}
              </small>
            </label>
          </div>
        </section>

        <section class="checkout-section" aria-labelledby="payment-title">
          <h2 id="payment-title">{{ form.fulfillment === 'delivery' ? '4' : '3' }}. Pagamento</h2>
          <div class="choice-group">
            <label class="choice">
              <input v-model="form.paymentMethod" name="paymentMethod" type="radio" value="cash" />
              <span>Dinheiro</span>
            </label>
            <label class="choice">
              <input v-model="form.paymentMethod" name="paymentMethod" type="radio" value="card" />
              <span>Cartão</span>
            </label>
            <label class="choice">
              <input v-model="form.paymentMethod" name="paymentMethod" type="radio" value="pix" />
              <span>Pix</span>
            </label>
          </div>
          <p class="payment-note">
            Pagamento na entrega ou retirada. Nenhum dado de cartão é solicitado.
          </p>
          <label v-if="form.paymentMethod === 'cash'" class="field">
            <span>Troco para</span>
            <input
              v-model="form.changeFor"
              id="checkout-changeFor"
              name="changeFor"
              type="text"
              inputmode="numeric"
              placeholder="100,00"
              :aria-invalid="errors.changeFor ? 'true' : undefined"
              :aria-describedby="errors.changeFor ? 'checkout-changeFor-error' : undefined"
            />
            <small v-if="errors.changeFor" id="checkout-changeFor-error" class="field-error">
              {{ errors.changeFor }}
            </small>
          </label>
          <label class="field">
            <span>Observação (opcional)</span>
            <textarea
              v-model="form.note"
              id="checkout-note"
              name="note"
              rows="3"
              maxlength="280"
              :aria-invalid="errors.note ? 'true' : undefined"
              :aria-describedby="errors.note ? 'checkout-note-error' : undefined"
            ></textarea>
            <small v-if="errors.note" id="checkout-note-error" class="field-error">
              {{ errors.note }}
            </small>
          </label>
        </section>

        <section class="checkout-section" aria-labelledby="review-title">
          <h2 id="review-title">
            {{ form.fulfillment === 'delivery' ? '5' : '4' }}. Revisão e envio
          </h2>
          <p class="review-note">
            Confira o resumo ao lado. O total abaixo foi calculado no servidor e será recalculado ao
            gravar o pedido.
          </p>
          <dl class="review-total" data-testid="review-total">
            <div>
              <dt>Total cotado</dt>
              <dd v-if="quote"><PriceText :value-cents="quote.totalCents" /></dd>
              <dd v-else-if="missingDeliveryZone">Escolha a zona para calcular</dd>
              <dd v-else>Aguardando cotação</dd>
            </div>
          </dl>

          <ErrorState v-if="quoteError" :message="quoteError" @retry="requestQuote" />
          <ErrorState v-if="submitError" :message="submitError" />

          <div class="checkout-actions">
            <BaseButton type="submit" :loading="submitting" :disabled="!canSubmit">
              {{ submitting ? 'Enviando pedido…' : 'Enviar pedido' }}
            </BaseButton>
            <p v-if="missingDeliveryZone" class="checkout-hint">
              Escolha uma zona de entrega para calcular o total.
            </p>
            <p v-else-if="!quote && !quoteError" class="checkout-hint">
              Calculando o total no servidor…
            </p>
          </div>
        </section>
      </form>

      <aside class="checkout-summary" aria-labelledby="checkout-summary-title">
        <h2 id="checkout-summary-title">Resumo</h2>
        <ul class="summary-items">
          <li v-for="line in lines" :key="line.key">
            {{ line.quantity }}× {{ line.product.name }}
          </li>
        </ul>
        <dl class="summary-totals">
          <div>
            <dt>Subtotal</dt>
            <dd><PriceText :value-cents="quote?.subtotalCents ?? 0" /></dd>
          </div>
          <div>
            <dt>Taxa de entrega</dt>
            <dd><PriceText :value-cents="quote?.deliveryFeeCents ?? 0" /></dd>
          </div>
          <div class="summary-totals__final">
            <dt>Total</dt>
            <dd><PriceText :value-cents="quote?.totalCents ?? 0" /></dd>
          </div>
        </dl>
        <p class="summary-note">
          O total é calculado no servidor e confirmado antes de gravar o pedido.
        </p>
      </aside>
    </div>
  </main>
</template>

<style scoped>
.checkout-page {
  display: grid;
  gap: 16px;
}

.checkout-page h1 {
  margin: 0;
  font-family: var(--font-display);
  font-size: var(--font-size-display);
  font-weight: 400;
  line-height: var(--line-height-display);
}

.checkout-preferences-note {
  margin: 0;
  color: var(--text-sub);
  font-size: var(--font-size-caption);
}

.back-link,
.text-link {
  display: inline-flex;
  width: fit-content;
  min-height: 44px;
  align-items: center;
  color: var(--brand-primary);
  font-size: var(--font-size-caption);
  font-weight: 800;
  text-decoration: none;
}

.checkout-empty {
  display: grid;
  gap: 8px;
  padding: 20px;
  border: 1px solid var(--bg-border);
  border-radius: 16px;
}

.checkout-layout {
  display: grid;
  gap: 20px;
  align-items: start;
}

.checkout-form {
  display: grid;
  gap: 14px;
}

.checkout-section {
  display: grid;
  gap: 12px;
  padding: 16px;
  border: 1px solid var(--bg-border);
  border-radius: 16px;
  background: var(--bg-surface);
}

.checkout-section h2,
.checkout-summary h2 {
  margin: 0;
  font-size: var(--font-size-h2);
}

.choice-group {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.choice {
  display: inline-flex;
  min-height: 44px;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  border: 1px solid var(--bg-border);
  border-radius: 12px;
  font-size: var(--font-size-caption);
  font-weight: 700;
}

.choice:has(input:checked) {
  border-color: var(--brand-primary);
}

.choice--opt-in {
  width: 100%;
  align-items: flex-start;
}

.choice--opt-in input {
  margin-top: 4px;
}

.field {
  display: grid;
  gap: 6px;
  font-size: var(--font-size-caption);
  font-weight: 700;
}

.field input,
.field select,
.field textarea {
  width: 100%;
  min-height: 44px;
  padding: 10px 12px;
  border: 1px solid var(--bg-border);
  border-radius: 10px;
  background: var(--bg-input);
  color: var(--text-main);
  font: inherit;
  font-weight: 400;
}

.field-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.field--wide {
  grid-column: 1 / -1;
}

.field-error {
  color: var(--status-error);
  font-size: var(--font-size-caption);
  font-weight: 500;
}

.payment-note,
.review-note,
.summary-note,
.checkout-hint {
  margin: 0;
  color: var(--text-sub);
  font-size: var(--font-size-caption);
  line-height: 1.5;
}

.review-total {
  margin: 0;
  padding: 12px 16px;
  border: 1px solid var(--bg-border);
  border-radius: 12px;
}

.review-total div {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 12px;
}

.review-total dt {
  color: var(--text-sub);
  font-size: var(--font-size-caption);
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.review-total dd {
  margin: 0;
  font-size: var(--font-size-h2);
  font-weight: 800;
}

.checkout-actions {
  display: grid;
  gap: 8px;
  justify-items: center;
}

.checkout-summary {
  display: grid;
  gap: 14px;
  padding: 18px;
  border: 1px solid var(--bg-border);
  border-radius: 16px;
  background: var(--bg-surface);
}

.summary-items {
  display: grid;
  gap: 6px;
  margin: 0;
  padding-left: 18px;
  color: var(--text-sub);
  font-size: var(--font-size-caption);
}

.summary-totals {
  display: grid;
  gap: 8px;
  margin: 0;
}

.summary-totals div {
  display: flex;
  justify-content: space-between;
  gap: 12px;
}

.summary-totals dt,
.summary-totals dd {
  margin: 0;
  font-size: var(--font-size-caption);
}

.summary-totals__final {
  padding-top: 10px;
  border-top: 1px solid var(--bg-border);
  font-weight: 800;
}

@media (min-width: 900px) {
  .checkout-layout {
    grid-template-columns: minmax(0, 1.6fr) minmax(280px, 1fr);
  }
}
</style>
