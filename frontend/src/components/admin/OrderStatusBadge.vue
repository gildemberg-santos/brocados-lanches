<script setup lang="ts">
import { computed } from 'vue'

import { ORDER_STATUS_LABELS, isOrderStatus } from '../../features/admin-orders/orderStatus'

const props = defineProps<{ status: string }>()

const label = computed(() =>
  isOrderStatus(props.status) ? ORDER_STATUS_LABELS[props.status] : props.status,
)
</script>

<template>
  <span class="order-status-badge" :data-status="status">{{ label }}</span>
</template>

<style scoped>
.order-status-badge {
  display: inline-flex;
  width: fit-content;
  align-items: center;
  padding: 6px 12px;
  border: 1px solid var(--bg-border);
  border-radius: 999px;
  color: var(--text-sub);
  font-size: var(--font-size-caption);
  font-weight: 800;
}

.order-status-badge[data-status='confirmed'],
.order-status-badge[data-status='preparing'] {
  color: var(--brand-primary);
}

.order-status-badge[data-status='ready'],
.order-status-badge[data-status='out_for_delivery'] {
  color: var(--brand-neon);
}

.order-status-badge[data-status='delivered'],
.order-status-badge[data-status='picked_up'] {
  border-color: var(--status-open);
  color: var(--status-open);
}

.order-status-badge[data-status='cancelled'] {
  border-color: var(--status-error);
  color: var(--status-error);
}
</style>
