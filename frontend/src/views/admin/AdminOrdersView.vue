<script setup lang="ts">
import { onMounted } from 'vue'
import { RouterLink } from 'vue-router'

import AdminOrderFilters from '../../components/admin/AdminOrderFilters.vue'
import OrderStatusBadge from '../../components/admin/OrderStatusBadge.vue'
import EmptyState from '../../components/EmptyState.vue'
import ErrorState from '../../components/ErrorState.vue'
import LoadingSkeleton from '../../components/LoadingSkeleton.vue'
import PriceText from '../../components/PriceText.vue'
import { formatOrderDateTime } from '../../features/admin-orders/format'
import { FULFILLMENT_LABELS } from '../../features/admin-orders/orderStatus'
import { useAdminOrdersStore } from '../../stores/adminOrders'
import type { AdminOrderListQuery } from '../../types/order'

const store = useAdminOrdersStore()

function applyFilters(filters: Partial<AdminOrderListQuery>): void {
  void store.applyFilters(filters)
}

function goToPage(page: number): void {
  void store.goToPage(page)
}

onMounted(() => {
  if (store.state === 'idle') {
    void store.load()
  }
})
</script>

<template>
  <main id="conteudo" class="admin-page">
    <p class="admin-page__kicker">Painel administrativo</p>
    <h1>Pedidos</h1>
    <p class="admin-page__lead">
      A fila mostra dados mascarados. Abra o pedido para ver os dados completos e atualizar o
      status.
    </p>

    <AdminOrderFilters :filters="store.filters" @apply="applyFilters" />

    <LoadingSkeleton v-if="store.state === 'loading'" label="Carregando pedidos" :count="4" />

    <ErrorState v-else-if="store.state === 'error'" :message="store.error" @retry="store.load" />

    <EmptyState
      v-else-if="store.isEmpty"
      title="Nenhum pedido encontrado"
      description="Ajuste os filtros ou aguarde novos pedidos."
    />

    <template v-else-if="store.state === 'success'">
      <ol class="order-list">
        <li v-for="order in store.orders" :key="order.id" class="order-row">
          <div class="order-row__head">
            <RouterLink
              class="order-row__number"
              :to="{ name: 'admin-order-detail', params: { id: order.id } }"
            >
              #{{ order.orderNumber }}
            </RouterLink>
            <OrderStatusBadge :status="order.status" />
          </div>

          <dl class="order-row__meta">
            <div>
              <dt>Criado em</dt>
              <dd>{{ formatOrderDateTime(order.createdAt) }}</dd>
            </div>
            <div>
              <dt>Entrega</dt>
              <dd>{{ FULFILLMENT_LABELS[order.fulfillment] }}</dd>
            </div>
            <div>
              <dt>Cliente</dt>
              <dd>{{ order.customerName }} · {{ order.phone }}</dd>
            </div>
            <div>
              <dt>Itens</dt>
              <dd>{{ order.itemCount }}</dd>
            </div>
            <div>
              <dt>Total</dt>
              <dd><PriceText :value-cents="order.totalCents" /></dd>
            </div>
          </dl>

          <RouterLink
            class="order-row__detail"
            :to="{ name: 'admin-order-detail', params: { id: order.id } }"
          >
            Ver detalhes
          </RouterLink>
        </li>
      </ol>

      <nav v-if="store.totalPages > 0" class="order-pagination" aria-label="Paginação de pedidos">
        <button
          type="button"
          class="order-pagination__button"
          :disabled="store.meta.page <= 1"
          @click="goToPage(store.meta.page - 1)"
        >
          Anterior
        </button>
        <span aria-live="polite">Página {{ store.meta.page }} de {{ store.totalPages }}</span>
        <button
          type="button"
          class="order-pagination__button"
          :disabled="store.meta.page >= store.totalPages"
          @click="goToPage(store.meta.page + 1)"
        >
          Próxima
        </button>
      </nav>
    </template>
  </main>
</template>

<style scoped>
.admin-page {
  display: grid;
  gap: 16px;
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
  margin: 8px 0;
  font-family: var(--font-display);
  font-size: var(--font-size-display);
  line-height: var(--line-height-display);
  letter-spacing: 0.02em;
}

.admin-page__lead {
  margin: 0 0 8px;
  color: var(--text-sub);
}

.order-list {
  display: grid;
  gap: 12px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.order-row {
  display: grid;
  gap: 12px;
  padding: 18px;
  border: 1px solid var(--bg-border);
  border-radius: 18px;
  background: var(--bg-surface);
}

.order-row__head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.order-row__number {
  color: var(--text-main);
  font-family: var(--font-display);
  font-size: var(--font-size-h2);
  text-decoration: none;
}

.order-row__meta {
  display: grid;
  gap: 8px;
  margin: 0;
}

.order-row__meta div {
  display: flex;
  justify-content: space-between;
  gap: 12px;
}

.order-row__meta dt,
.order-row__meta dd {
  margin: 0;
  font-size: var(--font-size-caption);
}

.order-row__meta dt {
  color: var(--text-sub);
}

.order-row__detail {
  display: inline-flex;
  width: fit-content;
  min-height: 44px;
  align-items: center;
  color: var(--brand-primary);
  font-size: var(--font-size-caption);
  font-weight: 800;
  text-decoration: none;
}

.order-pagination {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: center;
  gap: 12px;
  color: var(--text-sub);
  font-size: var(--font-size-caption);
}

.order-pagination__button {
  min-height: 44px;
  padding: 10px 18px;
  border: 1px solid var(--bg-border);
  border-radius: 999px;
  background: var(--bg-surface);
  color: var(--text-main);
  font-family: var(--font-sans);
  font-size: var(--font-size-caption);
  font-weight: 700;
}

.order-pagination__button:disabled {
  opacity: 0.5;
}

@media (min-width: 768px) {
  .order-row__meta {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
