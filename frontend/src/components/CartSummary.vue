<script setup lang="ts">
import PriceText from './PriceText.vue'

defineProps<{
  subtotalCents: number
  itemCount: number
}>()
</script>

<template>
  <aside class="cart-summary" aria-labelledby="cart-summary-title">
    <h2 id="cart-summary-title">Resumo do pedido</h2>

    <dl class="cart-summary__rows">
      <div class="cart-summary__row cart-summary__row--total">
        <dt>Subtotal ({{ itemCount }})</dt>
        <dd><PriceText :value-cents="subtotalCents" /></dd>
      </div>
    </dl>

    <p class="cart-summary__note">
      A taxa de entrega aparece no checkout, depois que você escolher entrega ou retirada. O total é
      calculado no servidor.
    </p>

    <div v-if="$slots.action" class="cart-summary__action">
      <slot name="action" />
    </div>
  </aside>
</template>

<style scoped>
.cart-summary {
  display: grid;
  gap: 14px;
  align-content: start;
  padding: 18px;
  border: 1px solid var(--bg-border);
  border-radius: 18px;
  background: var(--bg-surface);
}

.cart-summary h2 {
  margin: 0;
  font-size: var(--font-size-h1);
}

.cart-summary__rows {
  display: grid;
  gap: 10px;
  margin: 0;
}

.cart-summary__row {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  justify-content: space-between;
  gap: 4px 12px;
}

.cart-summary__row dt {
  color: var(--text-sub);
  font-weight: 700;
}

.cart-summary__row dd {
  display: grid;
  justify-items: end;
  gap: 2px;
  margin: 0;
}

.cart-summary__row--total {
  padding-top: 12px;
  border-top: 1px solid var(--bg-border);
}

.cart-summary__row--total dt {
  color: var(--text-main);
  font-weight: 800;
}

.cart-summary__note {
  margin: 0;
  color: var(--text-sub);
  font-size: var(--font-size-caption);
  line-height: 1.5;
}

.cart-summary__action {
  display: grid;
  gap: 8px;
}

.cart-summary__action :deep(.base-button) {
  width: 100%;
}
</style>
