<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRoute } from 'vue-router'

import BaseButton from '../components/BaseButton.vue'
import ErrorState from '../components/ErrorState.vue'
import LoadingSkeleton from '../components/LoadingSkeleton.vue'
import OrderStatusTimeline from '../components/OrderStatusTimeline.vue'
import PriceText from '../components/PriceText.vue'
import { formatOrderDateTime } from '../features/admin-orders/format'
import { checkoutApiMessage, orderStatusLabel } from '../features/checkout/checkout'
import { readHandoffUrl } from '../features/checkout/handoff'
import { fetchTrackedOrder } from '../services/orders'
import { useStoreStatusStore } from '../stores/storeStatus'
import type { TrackedOrder } from '../types/order'

const route = useRoute()
const store = useStoreStatusStore()
const publicId = computed(() => String(route.params.publicId ?? ''))
const order = ref<TrackedOrder | null>(null)
const loading = ref(true)
const error = ref('')
const handoffUrl = ref('')
const linkCopied = ref(false)

const statusLabel = computed(() => (order.value ? orderStatusLabel(order.value.status) : ''))
const isFinalStatus = computed(() =>
  order.value ? ['delivered', 'picked_up', 'cancelled'].includes(order.value.status) : false,
)
const pageTitle = computed(() => {
  if (loading.value) return 'Carregando pedido'
  if (error.value) return 'Acompanhamento indisponível'
  if (!order.value) return 'Acompanhe seu pedido'
  if (order.value.status === 'cancelled') return 'Pedido cancelado'
  if (isFinalStatus.value) return 'Pedido finalizado'
  return handoffUrl.value ? 'Pedido registrado' : 'Acompanhe seu pedido'
})

const storeStatus = computed(() => (store.state === 'success' ? store.status : null))
const storeWhatsappUrl = computed(() => {
  const digits = (storeStatus.value?.whatsapp ?? '').replace(/\D/g, '')
  return digits ? `https://wa.me/${digits}` : ''
})
const storeInstagramUrl = computed(() => {
  const value = storeStatus.value?.instagram ?? ''
  return /^https?:\/\//i.test(value) ? value : ''
})
const storeAddress = computed(() => storeStatus.value?.address.trim() ?? '')
const hasStoreContact = computed(
  () =>
    storeWhatsappUrl.value.length > 0 ||
    storeInstagramUrl.value.length > 0 ||
    storeAddress.value.length > 0,
)

async function copyHandoffUrl(): Promise<void> {
  try {
    await navigator.clipboard.writeText(handoffUrl.value)
    linkCopied.value = true
  } catch {
    linkCopied.value = false
  }
}

async function load(): Promise<void> {
  loading.value = true
  error.value = ''
  try {
    order.value = await fetchTrackedOrder(publicId.value)
    if (!handoffUrl.value) {
      handoffUrl.value = readHandoffUrl(publicId.value)
    }
  } catch (cause: unknown) {
    order.value = null
    error.value = checkoutApiMessage(cause)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  const state = window.history.state
  if (typeof state?.whatsappUrl === 'string') {
    handoffUrl.value = state.whatsappUrl
  }
  void load()
  void store.load()
})
</script>

<template>
  <main id="conteudo" class="page-shell order-page">
    <h1>{{ pageTitle }}</h1>

    <LoadingSkeleton v-if="loading" label="Carregando pedido" :count="3" />

    <ErrorState v-else-if="error" :message="error" @retry="load" />

    <template v-else-if="order">
      <section class="order-hero" aria-labelledby="order-number">
        <p class="order-kicker">Pedido #{{ order.orderNumber }}</p>
        <h2 id="order-number">{{ statusLabel }}</h2>
        <p v-if="order.estimatedTime && !isFinalStatus" class="order-time">
          Tempo estimado: {{ order.estimatedTime }}
        </p>
        <p class="order-id">Código de acompanhamento: {{ order.publicId }}</p>
        <p class="order-registered">
          Pedido registrado no servidor em
          <time :datetime="order.createdAt">{{ formatOrderDateTime(order.createdAt) }}</time>.
          Guarde o código acima para acompanhar o status.
        </p>
      </section>

      <section v-if="handoffUrl" class="order-handoff" aria-labelledby="handoff-title">
        <h2 id="handoff-title">Envie o pedido para a loja</h2>
        <p>
          Seu pedido já está salvo. Abra o WhatsApp para enviar os detalhes — se o envio falhar,
          você pode usar este link novamente.
        </p>
        <div class="handoff-actions">
          <a class="handoff-link" :href="handoffUrl" target="_blank" rel="noopener noreferrer">
            Abrir WhatsApp
          </a>
          <BaseButton variant="secondary" @click="copyHandoffUrl">
            {{ linkCopied ? 'Link copiado' : 'Copiar link' }}
          </BaseButton>
        </div>
        <p v-if="linkCopied" class="handoff-feedback" role="status">
          Link copiado. Você pode colá-lo no WhatsApp.
        </p>
      </section>

      <section class="order-detail" aria-labelledby="order-detail-title">
        <h2 id="order-detail-title">Resumo do pedido</h2>
        <ul class="order-items">
          <li v-for="(item, index) in order.items" :key="`${item.name}-${index}`">
            {{ item.quantity }}× {{ item.name }}
          </li>
        </ul>
        <dl class="order-totals">
          <div>
            <dt>Subtotal</dt>
            <dd><PriceText :value-cents="order.subtotalCents" /></dd>
          </div>
          <div>
            <dt>Taxa de entrega</dt>
            <dd><PriceText :value-cents="order.deliveryFeeCents" /></dd>
          </div>
          <div class="order-totals__final">
            <dt>Total</dt>
            <dd><PriceText :value-cents="order.totalCents" /></dd>
          </div>
        </dl>
        <p class="order-receipt">
          {{
            order.fulfillment === 'pickup' ? 'Retirada na loja' : 'Entrega em endereço cadastrado'
          }}
        </p>
      </section>

      <OrderStatusTimeline
        :status="order.status"
        :fulfillment="order.fulfillment"
        :created-at="order.createdAt"
        :updated-at="order.updatedAt"
      />

      <section v-if="hasStoreContact" class="order-contact" aria-labelledby="order-contact-title">
        <h2 id="order-contact-title">Fale com a loja</h2>
        <p>Tire dúvidas sobre o pedido #{{ order.orderNumber }} pelos canais publicados.</p>
        <div class="order-contact__actions">
          <a
            v-if="storeWhatsappUrl"
            class="text-link"
            :href="storeWhatsappUrl"
            target="_blank"
            rel="noopener noreferrer"
          >
            Abrir WhatsApp da loja
          </a>
          <a
            v-if="storeInstagramUrl"
            class="text-link"
            :href="storeInstagramUrl"
            target="_blank"
            rel="noopener noreferrer"
          >
            Abrir Instagram da loja
          </a>
        </div>
        <p v-if="storeAddress">Endereço da loja: {{ storeAddress }}</p>
      </section>

      <RouterLink class="text-link" :to="{ name: 'menu' }">Voltar ao cardápio</RouterLink>
    </template>
  </main>
</template>

<style scoped>
.order-page {
  display: grid;
  gap: 16px;
}

.order-page h1 {
  margin: 0;
  font-family: var(--font-display);
  font-size: var(--font-size-display);
  font-weight: 400;
  line-height: var(--line-height-display);
}

.order-hero,
.order-handoff,
.order-detail,
.order-contact {
  display: grid;
  gap: 10px;
  padding: 20px;
  border: 1px solid var(--bg-border);
  border-radius: 18px;
  background: var(--bg-surface);
}

.order-kicker {
  margin: 0;
  color: var(--text-sub);
  font-size: var(--font-size-caption);
  font-weight: 700;
}

.order-hero h2,
.order-handoff h2,
.order-detail h2,
.order-contact h2 {
  margin: 0;
  font-size: var(--font-size-h2);
}

.order-time,
.order-id,
.order-registered,
.order-handoff p,
.order-receipt,
.order-contact p {
  margin: 0;
  color: var(--text-sub);
  font-size: var(--font-size-caption);
  line-height: 1.5;
}

.order-id {
  overflow-wrap: anywhere;
}

.handoff-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.order-contact__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.handoff-feedback {
  margin: 0;
  color: var(--text-sub);
  font-size: var(--font-size-caption);
}

.handoff-link {
  display: inline-flex;
  width: fit-content;
  min-height: 44px;
  align-items: center;
  justify-content: center;
  padding: 12px 22px;
  border-radius: 999px;
  background: var(--brand-primary);
  color: var(--bg-primary);
  font-weight: 700;
  text-decoration: none;
}

.order-items {
  display: grid;
  gap: 6px;
  margin: 0;
  padding-left: 18px;
  font-size: var(--font-size-caption);
}

.order-totals {
  display: grid;
  gap: 8px;
  margin: 0;
}

.order-totals div {
  display: flex;
  justify-content: space-between;
  gap: 12px;
}

.order-totals dt,
.order-totals dd {
  margin: 0;
  font-size: var(--font-size-caption);
}

.order-totals__final {
  padding-top: 10px;
  border-top: 1px solid var(--bg-border);
  font-weight: 800;
}

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
</style>
