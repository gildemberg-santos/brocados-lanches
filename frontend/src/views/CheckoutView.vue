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
  validateCheckoutForm,
} from '../features/checkout/checkout'
import { saveHandoffUrl } from '../features/checkout/handoff'
import type { CheckoutErrors, CheckoutForm } from '../features/checkout/checkout'
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
let quoteTimer: ReturnType<typeof setTimeout> | null = null

const isLoading = computed(
  () =>
    menu.state === 'idle' ||
    menu.state === 'loading' ||
    zones.state === 'idle' ||
    zones.state === 'loading',
)
const lines = computed(() => buildCartLines(cart.items, menu.products).lines)
const quoteTotalCents = computed(() => quote.value?.totalCents ?? null)
const canSubmit = computed(
  () => !cart.isEmpty && !submitting.value && !quoteLoading.value && quote.value !== null,
)

function quoteSignature(): string {
  return JSON.stringify(buildQuoteRequest(cart.items, form.value))
}

async function requestQuote(): Promise<void> {
  if (cart.isEmpty) {
    quote.value = null
    return
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
  quoteTimer = setTimeout(() => {
    void requestQuote()
  }, 250)
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

onMounted(async () => {
  await Promise.all([menu.ensureLoaded(), zones.ensureLoaded()])
  if (menu.state === 'success') {
    cart.removeUnavailable(buildCartLines(cart.items, menu.products).unavailableKeys)
  }
  await requestQuote()
})

onUnmounted(() => {
  if (quoteTimer !== null) {
    clearTimeout(quoteTimer)
  }
})
</script>

<template>
  <main id="conteudo" class="page-shell checkout-page">
    <RouterLink class="back-link" :to="{ name: 'cart' }">← Voltar ao carrinho</RouterLink>
    <h1>Finalizar pedido</h1>

    <div v-if="cart.isEmpty" class="checkout-empty" role="status">
      <p>Seu carrinho está vazio. Escolha os itens antes de finalizar.</p>
      <RouterLink class="text-link" :to="{ name: 'menu' }">Ver cardápio</RouterLink>
    </div>

    <LoadingSkeleton v-else-if="isLoading" label="Carregando checkout" :count="3" />

    <div v-else class="checkout-layout">
      <form class="checkout-form" novalidate @submit.prevent="submit">
        <section class="checkout-section" aria-labelledby="fulfillment-title">
          <h2 id="fulfillment-title">1. Recebimento</h2>
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
          <h2 id="address-title">2. Entrega</h2>
          <label class="field">
            <span>Zona de entrega</span>
            <select v-model="form.zoneId" name="zoneId">
              <option value="">Selecione o bairro</option>
              <option v-for="zone in zones.zones" :key="zone.id" :value="zone.id">
                {{ zone.name }} — {{ formatCents(zone.feeCents) }}
              </option>
            </select>
            <small v-if="errors.zoneId" class="field-error">{{ errors.zoneId }}</small>
          </label>

          <div class="field-grid">
            <label class="field field--wide">
              <span>Rua</span>
              <input
                v-model="form.street"
                name="street"
                type="text"
                autocomplete="street-address"
              />
              <small v-if="errors.street" class="field-error">{{ errors.street }}</small>
            </label>
            <label class="field">
              <span>Número</span>
              <input v-model="form.number" name="number" type="text" />
              <small v-if="errors.number" class="field-error">{{ errors.number }}</small>
            </label>
            <label class="field">
              <span>Complemento</span>
              <input v-model="form.complement" name="complement" type="text" />
            </label>
            <label class="field field--wide">
              <span>Bairro</span>
              <input v-model="form.district" name="district" type="text" />
              <small v-if="errors.district" class="field-error">{{ errors.district }}</small>
            </label>
            <label class="field">
              <span>CEP</span>
              <input v-model="form.zipCode" name="zipCode" type="text" inputmode="numeric" />
              <small v-if="errors.zipCode" class="field-error">{{ errors.zipCode }}</small>
            </label>
            <label class="field">
              <span>Cidade</span>
              <input v-model="form.city" name="city" type="text" />
              <small v-if="errors.city" class="field-error">{{ errors.city }}</small>
            </label>
            <label class="field">
              <span>Estado</span>
              <input v-model="form.state" name="state" type="text" maxlength="2" />
              <small v-if="errors.state" class="field-error">{{ errors.state }}</small>
            </label>
          </div>
        </section>

        <section class="checkout-section" aria-labelledby="customer-title">
          <h2>{{ form.fulfillment === 'delivery' ? '3' : '2' }}. Seus dados</h2>
          <label class="field">
            <span>Nome completo</span>
            <input
              v-model="form.customerName"
              name="customerName"
              type="text"
              autocomplete="name"
            />
            <small v-if="errors.customerName" class="field-error">{{ errors.customerName }}</small>
          </label>
          <label class="field">
            <span>Telefone</span>
            <input
              :value="form.phone"
              name="phone"
              type="tel"
              inputmode="tel"
              autocomplete="tel"
              placeholder="(85) 99999-9999"
              @input="form.phone = formatPhone(($event.target as HTMLInputElement).value)"
            />
            <small v-if="errors.phone" class="field-error">{{ errors.phone }}</small>
          </label>
        </section>

        <section class="checkout-section" aria-labelledby="payment-title">
          <h2>{{ form.fulfillment === 'delivery' ? '4' : '3' }}. Pagamento</h2>
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
              name="changeFor"
              type="text"
              inputmode="numeric"
              placeholder="100,00"
            />
            <small v-if="errors.changeFor" class="field-error">{{ errors.changeFor }}</small>
          </label>
          <label class="field">
            <span>Observação (opcional)</span>
            <textarea v-model="form.note" name="note" rows="3" maxlength="280"></textarea>
            <small v-if="errors.note" class="field-error">{{ errors.note }}</small>
          </label>
        </section>

        <ErrorState v-if="quoteError" :message="quoteError" @retry="requestQuote" />
        <ErrorState v-if="submitError" :message="submitError" />

        <div class="checkout-actions">
          <BaseButton type="submit" :loading="submitting" :disabled="!canSubmit">
            {{ submitting ? 'Enviando pedido…' : 'Enviar pedido' }}
          </BaseButton>
          <p v-if="!quote && !quoteError" class="checkout-hint">Calculando o total no servidor…</p>
        </div>
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
  background: var(--bg-base);
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
.summary-note,
.checkout-hint {
  margin: 0;
  color: var(--text-sub);
  font-size: var(--font-size-caption);
  line-height: 1.5;
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
