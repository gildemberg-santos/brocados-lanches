<script setup lang="ts">
import { onMounted } from 'vue'
import { RouterLink } from 'vue-router'

import OrderStatusBadge from '../components/admin/OrderStatusBadge.vue'
import BaseButton from '../components/BaseButton.vue'
import EmptyState from '../components/EmptyState.vue'
import ErrorState from '../components/ErrorState.vue'
import LoadingSkeleton from '../components/LoadingSkeleton.vue'
import PriceText from '../components/PriceText.vue'
import { formatOrderDateTime } from '../features/admin-orders/format'
import { FULFILLMENT_LABELS } from '../features/admin-orders/orderStatus'
import { orderStatusLabel } from '../features/checkout/checkout'
import { useOrderHistoryStore } from '../stores/orderHistory'

const store = useOrderHistoryStore()

function refresh(): void {
  void store.load()
}

function goToPage(page: number): void {
  void store.goToPage(page)
}

onMounted(() => {
  if (store.state === 'idle') {
    refresh()
  }
})
</script>

<template>
  <main id="conteudo" tabindex="-1" class="page-shell history-page">
    <h1>Meus pedidos</h1>
    <p class="history-page__lead">
      Consulte novamente os pedidos feitos neste navegador. O histórico não usa senha nem telefone.
    </p>

    <div class="history-page__toolbar">
      <p>
        O acesso fica apenas neste dispositivo. Limpar os dados do navegador remove o histórico.
      </p>
      <BaseButton variant="secondary" :loading="store.state === 'loading'" @click="refresh">
        Atualizar
      </BaseButton>
    </div>

    <LoadingSkeleton v-if="store.state === 'loading'" label="Carregando pedidos" :count="3" />

    <ErrorState v-else-if="store.state === 'error'" :message="store.error" @retry="refresh" />

    <EmptyState
      v-else-if="!store.hasToken"
      title="Nenhum histórico neste navegador"
      description="Quando você fizer um pedido, um acesso será guardado aqui para consultá-lo depois."
    >
      <template #actions>
        <RouterLink class="history-page__cta" to="/cardapio">Ver cardápio</RouterLink>
      </template>
    </EmptyState>

    <EmptyState
      v-else-if="store.isEmpty"
      title="Você ainda não fez pedidos"
      description="Escolha seus lanches e finalize um pedido para vê-lo aqui."
    >
      <template #actions>
        <RouterLink class="history-page__cta" to="/cardapio">Ver cardápio</RouterLink>
      </template>
    </EmptyState>

    <template v-else-if="store.state === 'success'">
      <ul class="history-list">
        <li v-for="order in store.orders" :key="order.publicId">
          <RouterLink
            class="history-card"
            :to="{ name: 'order', params: { publicId: order.publicId } }"
          >
            <span class="history-card__head">
              <strong>Pedido #{{ order.orderNumber }}</strong>
              <OrderStatusBadge :status="order.status" />
            </span>
            <span class="history-card__date">
              <time :datetime="order.createdAt">{{ formatOrderDateTime(order.createdAt) }}</time>
            </span>
            <span class="history-card__summary">
              {{ order.items.map((item) => `${item.quantity}× ${item.name}`).join(', ') }}
            </span>
            <span class="history-card__foot">
              <span
                >{{ FULFILLMENT_LABELS[order.fulfillment] }} ·
                {{ orderStatusLabel(order.status) }}</span
              >
              <strong><PriceText :value-cents="order.totalCents" /></strong>
            </span>
            <span class="history-card__action">Ver acompanhamento</span>
          </RouterLink>
        </li>
      </ul>

      <nav v-if="store.totalPages > 0" class="history-pagination" aria-label="Paginação de pedidos">
        <button
          type="button"
          class="history-pagination__button"
          :disabled="store.meta.page <= 1"
          @click="goToPage(store.meta.page - 1)"
        >
          Anterior
        </button>
        <span aria-live="polite">Página {{ store.meta.page }} de {{ store.totalPages }}</span>
        <button
          type="button"
          class="history-pagination__button"
          :disabled="store.meta.page >= store.totalPages"
          @click="goToPage(store.meta.page + 1)"
        >
          Próxima
        </button>
      </nav>

      <p class="history-page__privacy">
        Para apagar este acesso, use a opção de limpar dados em
        <RouterLink class="history-page__inline-link" to="/privacidade">Privacidade</RouterLink>. O
        histórico também ficará indisponível neste navegador.
      </p>
    </template>
  </main>
</template>

<style scoped>
.history-page {
  display: grid;
  gap: 16px;
}

.history-page h1 {
  margin: 0;
  font-family: var(--font-display);
  font-size: var(--font-size-display);
  line-height: var(--line-height-display);
  letter-spacing: 0.02em;
}

.history-page__lead {
  max-width: 62ch;
  margin: 0;
  color: var(--text-sub);
  line-height: 1.6;
}

.history-page__toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.history-page__toolbar p {
  max-width: 58ch;
  margin: 0;
  color: var(--text-sub);
  font-size: var(--font-size-caption);
}

.history-page__cta {
  display: inline-flex;
  min-height: 44px;
  align-items: center;
  padding: 10px 20px;
  border: 1px solid var(--bg-border);
  border-radius: 999px;
  color: var(--brand-primary);
  font-weight: 700;
  text-decoration: none;
}

.history-page__cta:hover {
  color: var(--brand-hover);
}

.history-list {
  display: grid;
  gap: 12px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.history-card {
  display: grid;
  gap: 10px;
  padding: 18px;
  border: 1px solid var(--bg-border);
  border-radius: 18px;
  background: var(--bg-surface);
  color: var(--text-main);
  text-decoration: none;
}

.history-card:hover,
.history-card:focus-visible {
  border-color: var(--brand-primary);
}

.history-card__head,
.history-card__foot {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.history-card__date,
.history-card__summary {
  color: var(--text-sub);
  font-size: var(--font-size-caption);
}

.history-card__summary {
  line-height: 1.5;
}

.history-card__foot strong {
  font-size: var(--font-size-h2);
}

.history-card__action {
  color: var(--brand-primary);
  font-weight: 700;
}

.history-pagination {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: center;
  gap: 12px;
  color: var(--text-sub);
  font-size: var(--font-size-caption);
}

.history-pagination__button {
  min-height: 44px;
  padding: 10px 18px;
  border: 1px solid var(--bg-border);
  border-radius: 999px;
  background: var(--bg-input);
  color: var(--text-main);
  font: inherit;
  font-weight: 700;
  cursor: pointer;
}

.history-pagination__button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.history-page__privacy {
  margin: 0;
  color: var(--text-sub);
  font-size: var(--font-size-caption);
}

.history-page__inline-link {
  color: var(--brand-primary);
}
</style>
