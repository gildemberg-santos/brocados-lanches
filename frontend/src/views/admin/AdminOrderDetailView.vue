<script setup lang="ts">
import { computed, onMounted, watch } from 'vue'
import { RouterLink, useRoute } from 'vue-router'

import OrderStatusActions from '../../components/admin/OrderStatusActions.vue'
import OrderStatusBadge from '../../components/admin/OrderStatusBadge.vue'
import PriceText from '../../components/PriceText.vue'
import ErrorState from '../../components/ErrorState.vue'
import LoadingSkeleton from '../../components/LoadingSkeleton.vue'
import { formatOrderDateTime } from '../../features/admin-orders/format'
import {
  FULFILLMENT_LABELS,
  ORDER_STATUS_LABELS,
  isOrderStatus,
} from '../../features/admin-orders/orderStatus'
import { useAdminOrdersStore } from '../../stores/adminOrders'
import type { AdminOrderStatusRequest, PaymentMethod } from '../../types/order'

const PAYMENT_LABELS: Record<PaymentMethod, string> = {
  cash: 'Dinheiro',
  card: 'Cartão',
  pix: 'Pix',
}

const route = useRoute()
const store = useAdminOrdersStore()
const orderId = computed(() => String(route.params.id ?? ''))
const order = computed(() => store.detail)

function statusLabelFor(status: string): string {
  return isOrderStatus(status) ? ORDER_STATUS_LABELS[status] : status
}

const statusLabel = computed(() => statusLabelFor(order.value?.status ?? ''))

function changeStatus(request: AdminOrderStatusRequest): void {
  void store.changeStatus(orderId.value, request)
}

function load(): void {
  void store.loadDetail(orderId.value, true)
}

onMounted(() => {
  if (!order.value || order.value.id !== orderId.value) {
    void store.loadDetail(orderId.value)
  }
})

watch(orderId, (next) => {
  store.resetDetail()
  void store.loadDetail(next)
})
</script>

<template>
  <main id="conteudo" class="admin-page">
    <RouterLink class="admin-page__back" :to="{ name: 'admin-orders' }"> Voltar à fila </RouterLink>
    <p class="admin-page__kicker">Painel administrativo</p>

    <LoadingSkeleton v-if="store.detailState === 'loading'" label="Carregando pedido" :count="3" />

    <ErrorState
      v-else-if="store.detailState === 'error'"
      :message="store.detailError"
      @retry="load"
    />

    <template v-else-if="order">
      <div class="admin-page__head">
        <h1>Pedido #{{ order.orderNumber }}</h1>
        <OrderStatusBadge :status="order.status" />
      </div>
      <p class="admin-page__lead">
        {{ FULFILLMENT_LABELS[order.fulfillment] }} · aberto em
        {{ formatOrderDateTime(order.createdAt) }} · status: {{ statusLabel }}
      </p>

      <p v-if="store.actionSuccess" class="admin-page__success" role="status">
        {{ store.actionSuccess }}
      </p>
      <p v-if="store.actionError" class="admin-page__error" role="alert">
        {{ store.actionError }}
      </p>

      <OrderStatusActions
        :status="order.status"
        :fulfillment="order.fulfillment"
        :loading="store.actionLoading"
        @change="changeStatus"
      />

      <section class="admin-page__section" aria-labelledby="order-customer-title">
        <h2 id="order-customer-title">Cliente e entrega</h2>
        <dl class="detail-list">
          <div>
            <dt>Nome</dt>
            <dd>{{ order.customer.name }}</dd>
          </div>
          <div>
            <dt>Telefone</dt>
            <dd>{{ order.customer.phone }}</dd>
          </div>
          <div v-if="order.zone">
            <dt>Zona</dt>
            <dd>{{ order.zone.name }}</dd>
          </div>
          <div v-if="order.address">
            <dt>Endereço</dt>
            <dd>
              {{ order.address.street }}, {{ order.address.number }}
              <template v-if="order.address.complement"> — {{ order.address.complement }}</template>
              — {{ order.address.district }}, {{ order.address.city }}/{{ order.address.state }} —
              CEP {{ order.address.zipCode }}
            </dd>
          </div>
          <div>
            <dt>Pagamento</dt>
            <dd>
              {{ PAYMENT_LABELS[order.payment.method] }}
              <template v-if="order.payment.changeForCents">
                · troco para <PriceText :value-cents="order.payment.changeForCents" />
              </template>
            </dd>
          </div>
          <div v-if="order.note">
            <dt>Observação</dt>
            <dd>{{ order.note }}</dd>
          </div>
        </dl>
      </section>

      <section class="admin-page__section" aria-labelledby="order-items-title">
        <h2 id="order-items-title">Itens</h2>
        <ul class="detail-items">
          <li v-for="(item, index) in order.items" :key="`${item.productId}-${index}`">
            <strong>{{ item.quantity }}× {{ item.name }}</strong>
            <span>{{
              item.options.map((option) => option.name).join(', ') || 'Sem adicionais'
            }}</span>
            <span v-if="item.note">Obs.: {{ item.note }}</span>
            <span><PriceText :value-cents="item.lineTotalCents" /></span>
          </li>
        </ul>
        <dl class="detail-totals">
          <div>
            <dt>Subtotal</dt>
            <dd><PriceText :value-cents="order.subtotalCents" /></dd>
          </div>
          <div>
            <dt>Taxa de entrega</dt>
            <dd><PriceText :value-cents="order.deliveryFeeCents" /></dd>
          </div>
          <div class="detail-totals__final">
            <dt>Total</dt>
            <dd><PriceText :value-cents="order.totalCents" /></dd>
          </div>
        </dl>
      </section>

      <section class="admin-page__section" aria-labelledby="order-history-title">
        <h2 id="order-history-title">Histórico de status</h2>
        <ol class="detail-history">
          <li v-for="(change, index) in order.statusHistory" :key="`${change.at}-${index}`">
            <span>{{ statusLabelFor(change.status) }}</span>
            <time :datetime="change.at">{{ formatOrderDateTime(change.at) }}</time>
          </li>
        </ol>
      </section>

      <RouterLink
        class="admin-page__back"
        :to="{ name: 'order', params: { publicId: order.publicId } }"
      >
        Ver acompanhamento público
      </RouterLink>
    </template>
  </main>
</template>

<style scoped>
.admin-page {
  display: grid;
  gap: 16px;
}

.admin-page__back {
  display: inline-flex;
  width: fit-content;
  min-height: 44px;
  align-items: center;
  color: var(--brand-primary);
  font-size: var(--font-size-caption);
  font-weight: 800;
  text-decoration: none;
}

.admin-page__kicker {
  margin: 0;
  color: var(--brand-neon);
  font-size: var(--font-size-caption);
  font-weight: 800;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.admin-page h1 {
  margin: 0;
  font-family: var(--font-display);
  font-size: var(--font-size-display);
  line-height: var(--line-height-display);
  letter-spacing: 0.02em;
}

.admin-page__head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px;
}

.admin-page__lead {
  margin: 0;
  color: var(--text-sub);
}

.admin-page__success,
.admin-page__error {
  margin: 0;
  padding: 12px 16px;
  border: 1px solid var(--status-open);
  border-radius: 12px;
  font-size: var(--font-size-caption);
  font-weight: 700;
}

.admin-page__error {
  border-color: var(--status-error);
  color: var(--status-error);
}

.admin-page__section {
  display: grid;
  gap: 12px;
  padding: 20px;
  border: 1px solid var(--bg-border);
  border-radius: 18px;
  background: var(--bg-surface);
}

.admin-page__section h2 {
  margin: 0;
  font-size: var(--font-size-h2);
}

.detail-list,
.detail-totals {
  display: grid;
  gap: 8px;
  margin: 0;
}

.detail-list div,
.detail-totals div {
  display: flex;
  justify-content: space-between;
  gap: 12px;
}

.detail-list dt,
.detail-list dd,
.detail-totals dt,
.detail-totals dd {
  margin: 0;
  font-size: var(--font-size-caption);
}

.detail-list dt,
.detail-totals dt {
  color: var(--text-sub);
  flex: 0 0 auto;
}

.detail-list dd {
  text-align: right;
}

.detail-items {
  display: grid;
  gap: 10px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.detail-items li {
  display: grid;
  gap: 4px;
  padding-bottom: 10px;
  border-bottom: 1px solid var(--bg-border);
  font-size: var(--font-size-caption);
}

.detail-items span {
  color: var(--text-sub);
}

.detail-totals__final {
  padding-top: 10px;
  border-top: 1px solid var(--bg-border);
  font-weight: 800;
}

.detail-history {
  display: grid;
  gap: 8px;
  margin: 0;
  padding-left: 18px;
  font-size: var(--font-size-caption);
}

.detail-history li {
  display: flex;
  flex-wrap: wrap;
  justify-content: space-between;
  gap: 8px;
}

.detail-history time {
  color: var(--text-sub);
}
</style>
