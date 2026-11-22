<script setup lang="ts">
import { RouterLink } from 'vue-router'

import type { MenuProduct } from '../types/catalog'
import PriceText from './PriceText.vue'
import ProductImage from './ProductImage.vue'

defineProps<{
  product: MenuProduct
}>()
</script>

<template>
  <RouterLink
    class="product-card-compact"
    :class="{ 'product-card-compact--sold-out': !product.available }"
    :to="{ name: 'product', params: { slug: product.slug } }"
  >
    <span class="product-card-compact__media">
      <ProductImage
        :src="product.imageUrl"
        :alt="product.imageAlt.length > 0 ? product.imageAlt : product.name"
        :width="160"
        :height="160"
      />
    </span>
    <span class="product-card-compact__content">
      <strong>{{ product.name }}</strong>
      <span class="product-card-compact__description">{{ product.description }}</span>
      <span class="product-card-compact__footer">
        <PriceText :value-cents="product.priceCents" />
        <span v-if="!product.available" class="product-card-compact__badge">Esgotado</span>
      </span>
    </span>
  </RouterLink>
</template>

<style scoped>
.product-card-compact {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 12px;
  padding: 10px;
  border: 1px solid var(--bg-border);
  border-radius: 16px;
  background: var(--bg-surface);
  color: inherit;
  text-decoration: none;
  transition: border-color 180ms ease;
}

.product-card-compact:hover {
  border-color: var(--brand-primary);
}

.product-card-compact__media {
  width: 76px;
  height: 76px;
  flex: 0 0 auto;
  overflow: hidden;
  border-radius: 12px;
}

.product-card-compact__content {
  display: grid;
  min-width: 0;
  gap: 3px;
}

.product-card-compact strong {
  overflow: hidden;
  font-size: var(--font-size-body);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.product-card-compact__description {
  display: -webkit-box;
  overflow: hidden;
  color: var(--text-sub);
  font-size: var(--font-size-caption);
  line-height: 1.4;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 1;
}

.product-card-compact__footer {
  display: flex;
  align-items: center;
  gap: 8px;
}

.product-card-compact__badge {
  color: var(--status-error);
  font-size: var(--font-size-caption);
  font-weight: 700;
}

.product-card-compact--sold-out {
  opacity: 0.7;
}

@media (prefers-reduced-motion: reduce) {
  .product-card-compact {
    transition: none;
  }
}
</style>
