<script setup lang="ts">
import { RouterLink } from 'vue-router'

import type { CartLine } from '../types/cart'
import BaseButton from './BaseButton.vue'
import PriceText from './PriceText.vue'
import ProductImage from './ProductImage.vue'
import QuantityStepper from './QuantityStepper.vue'

defineProps<{
  line: CartLine
}>()

const emit = defineEmits<{
  'update:quantity': [quantity: number]
  remove: []
}>()
</script>

<template>
  <article class="cart-line">
    <RouterLink
      class="cart-line__media"
      :to="{ name: 'product', params: { slug: line.product.slug } }"
    >
      <ProductImage
        :src="line.product.imageUrl"
        :alt="line.product.imageAlt.length > 0 ? line.product.imageAlt : line.product.name"
        :width="160"
        :height="160"
      />
    </RouterLink>

    <div class="cart-line__content">
      <RouterLink
        class="cart-line__name"
        :to="{ name: 'product', params: { slug: line.product.slug } }"
      >
        {{ line.product.name }}
      </RouterLink>

      <p v-if="line.optionNames.length > 0" class="cart-line__choices">
        {{ line.optionNames.join(' · ') }}
      </p>
      <p v-if="line.addonNames.length > 0" class="cart-line__choices">
        Adicionais: {{ line.addonNames.join(', ') }}
      </p>
      <p v-if="line.note.length > 0" class="cart-line__note">Obs.: {{ line.note }}</p>

      <div class="cart-line__controls">
        <QuantityStepper
          :model-value="line.quantity"
          :label="`quantidade de ${line.product.name}`"
          @update:model-value="emit('update:quantity', $event)"
        />
        <BaseButton
          variant="ghost"
          class="cart-line__remove"
          :aria-label="`Remover ${line.product.name} do carrinho`"
          @click="emit('remove')"
        >
          Remover
        </BaseButton>
      </div>
    </div>

    <div class="cart-line__prices">
      <span class="cart-line__unit-label">Unitário</span>
      <PriceText :value-cents="line.unitPriceCents" />
      <span class="cart-line__unit-label">Item</span>
      <PriceText :value-cents="line.lineTotalCents" />
    </div>
  </article>
</template>

<style scoped>
.cart-line {
  display: grid;
  grid-template-columns: 76px minmax(0, 1fr);
  gap: 12px;
  padding: 12px;
  border: 1px solid var(--bg-border);
  border-radius: 18px;
  background: var(--bg-surface);
}

.cart-line__media {
  width: 76px;
  height: 76px;
  overflow: hidden;
  border-radius: 12px;
}

.cart-line__content {
  display: grid;
  min-width: 0;
  align-content: start;
  gap: 4px;
}

.cart-line__name {
  color: var(--text-main);
  font-weight: 800;
  text-decoration: none;
}

.cart-line__name:hover {
  color: var(--brand-primary);
}

.cart-line__choices,
.cart-line__note {
  margin: 0;
  color: var(--text-sub);
  font-size: var(--font-size-caption);
  line-height: 1.5;
}

.cart-line__note {
  color: var(--text-main);
}

.cart-line__controls {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  margin-top: 6px;
}

.cart-line__remove {
  min-height: 44px;
  padding-inline: 12px;
  font-size: var(--font-size-caption);
}

.cart-line__prices {
  display: grid;
  grid-column: 1 / -1;
  grid-template-columns: 1fr auto;
  align-items: center;
  gap: 4px 12px;
  padding-top: 10px;
  border-top: 1px solid var(--bg-border);
}

.cart-line__unit-label {
  color: var(--text-sub);
  font-size: var(--font-size-caption);
  font-weight: 700;
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

@media (min-width: 640px) {
  .cart-line {
    grid-template-columns: 96px minmax(0, 1fr) auto;
  }

  .cart-line__media {
    width: 96px;
    height: 96px;
  }

  .cart-line__prices {
    grid-column: auto;
    grid-template-columns: 1fr;
    align-content: start;
    justify-items: end;
    padding-top: 0;
    border-top: 0;
  }
}
</style>
