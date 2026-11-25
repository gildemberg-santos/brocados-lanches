<script setup lang="ts">
import { RouterLink } from 'vue-router'

import PriceText from './PriceText.vue'

defineProps<{
  count: number
  subtotalCents: number
}>()
</script>

<template>
  <RouterLink class="cart-summary-bar" :to="{ name: 'cart' }" data-testid="mobile-cart-bar">
    <span class="cart-summary-bar__count">Ver carrinho ({{ count }})</span>
    <PriceText :value-cents="subtotalCents" />
  </RouterLink>
</template>

<style scoped>
.cart-summary-bar {
  position: fixed;
  right: 12px;
  bottom: 12px;
  left: 12px;
  z-index: 30;
  display: flex;
  min-height: 56px;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 12px 20px;
  border: 1px solid var(--bg-border);
  border-radius: 999px;
  background: var(--brand-primary);
  color: var(--bg-primary);
  font-size: var(--font-size-caption);
  font-weight: 800;
  text-decoration: none;
}

.cart-summary-bar:hover {
  background: var(--brand-hover);
}

.cart-summary-bar :deep(.price-text) {
  color: var(--bg-primary);
}

@media (min-width: 768px) {
  .cart-summary-bar {
    display: none;
  }
}
</style>
