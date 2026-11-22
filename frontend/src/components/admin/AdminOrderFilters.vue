<script setup lang="ts">
import { ref } from 'vue'

import BaseButton from '../BaseButton.vue'
import { ORDER_STATUSES, ORDER_STATUS_LABELS } from '../../features/admin-orders/orderStatus'
import type { AdminOrderListQuery } from '../../types/order'

const props = defineProps<{ filters: AdminOrderListQuery }>()

const emit = defineEmits<{
  apply: [filters: Partial<AdminOrderListQuery>]
}>()

const status = ref(props.filters.status ?? '')
const fulfillment = ref(props.filters.fulfillment ?? '')
const search = ref(props.filters.q ?? '')
const from = ref(props.filters.from?.slice(0, 10) ?? '')
const to = ref(props.filters.to?.slice(0, 10) ?? '')

function apply(): void {
  emit('apply', {
    status: status.value || undefined,
    fulfillment: fulfillment.value ? (fulfillment.value as 'delivery' | 'pickup') : undefined,
    q: search.value.trim() || undefined,
    from: from.value ? `${from.value}T00:00:00Z` : undefined,
    to: to.value ? `${to.value}T23:59:59Z` : undefined,
  })
}

function clear(): void {
  status.value = ''
  fulfillment.value = ''
  search.value = ''
  from.value = ''
  to.value = ''
  emit('apply', {})
}
</script>

<template>
  <form class="admin-order-filters" aria-label="Filtros de pedidos" @submit.prevent="apply">
    <div class="admin-order-filters__field">
      <label for="order-search">Buscar</label>
      <input
        id="order-search"
        v-model="search"
        type="search"
        placeholder="Número ou dígitos do telefone"
      />
    </div>

    <div class="admin-order-filters__field">
      <label for="order-status">Status</label>
      <select id="order-status" v-model="status">
        <option value="">Todos</option>
        <option v-for="option in ORDER_STATUSES" :key="option" :value="option">
          {{ ORDER_STATUS_LABELS[option] }}
        </option>
      </select>
    </div>

    <div class="admin-order-filters__field">
      <label for="order-fulfillment">Forma de entrega</label>
      <select id="order-fulfillment" v-model="fulfillment">
        <option value="">Todas</option>
        <option value="delivery">Entrega</option>
        <option value="pickup">Retirada</option>
      </select>
    </div>

    <div class="admin-order-filters__field">
      <label for="order-from">De (UTC)</label>
      <input id="order-from" v-model="from" type="date" />
    </div>

    <div class="admin-order-filters__field">
      <label for="order-to">Até (UTC)</label>
      <input id="order-to" v-model="to" type="date" />
    </div>

    <div class="admin-order-filters__actions">
      <BaseButton type="submit">Filtrar</BaseButton>
      <BaseButton variant="ghost" @click="clear">Limpar</BaseButton>
    </div>
  </form>
</template>

<style scoped>
.admin-order-filters {
  display: grid;
  gap: 12px;
  padding: 18px;
  border: 1px solid var(--bg-border);
  border-radius: 18px;
  background: var(--bg-surface);
}

.admin-order-filters__field {
  display: grid;
  gap: 6px;
}

.admin-order-filters label {
  color: var(--text-sub);
  font-size: var(--font-size-caption);
  font-weight: 800;
}

.admin-order-filters input,
.admin-order-filters select {
  width: 100%;
  min-height: 44px;
  padding: 10px 12px;
  border: 1px solid var(--bg-border);
  border-radius: 12px;
  background: var(--bg-input);
  color: var(--text-main);
  font-family: var(--font-sans);
  font-size: var(--font-size-body);
}

.admin-order-filters__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

@media (min-width: 768px) {
  .admin-order-filters {
    grid-template-columns: 2fr 1fr 1fr 1fr 1fr auto;
    align-items: end;
  }
}
</style>
