<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'

import type { MenuProduct } from '../types/catalog'
import PriceText from './PriceText.vue'
import ProductImage from './ProductImage.vue'

const props = defineProps<{
  product: MenuProduct
}>()

const accessibleLabel = computed(
  () => `Ver detalhes de ${props.product.name}${props.product.available ? '' : ' (esgotado)'}`,
)
</script>

<template>
  <article class="product-card" :class="{ 'product-card--sold-out': !product.available }">
    <RouterLink
      class="product-card__link"
      :to="{ name: 'product', params: { slug: product.slug } }"
      :aria-label="accessibleLabel"
    >
      <div class="product-card__media">
        <ProductImage
          :src="product.imageUrl"
          :alt="product.imageAlt.length > 0 ? product.imageAlt : product.name"
          :width="480"
          :height="360"
        />
        <span v-if="!product.available" class="product-card__badge">Esgotado</span>
      </div>

      <div class="product-card__body">
        <h3 class="product-card__name">{{ product.name }}</h3>
        <p class="product-card__description">{{ product.description }}</p>
        <div class="product-card__footer">
          <PriceText :value-cents="product.priceCents" />
          <span class="product-card__action">
            {{ product.available ? 'Ver item' : 'Indisponível' }}
          </span>
        </div>
      </div>
    </RouterLink>
  </article>
</template>

<style scoped>
.product-card {
  min-width: 0;
}

.product-card__link {
  display: grid;
  height: 100%;
  overflow: hidden;
  border: 1px solid var(--bg-border);
  border-radius: 18px;
  background: var(--bg-surface);
  color: inherit;
  text-decoration: none;
  transition:
    border-color 180ms ease,
    transform 180ms ease;
}

.product-card__link:hover {
  border-color: var(--brand-primary);
  transform: translateY(-2px);
}

.product-card__media {
  position: relative;
  aspect-ratio: 4 / 3;
  overflow: hidden;
  background: var(--bg-input);
}

.product-card__badge {
  position: absolute;
  top: 10px;
  left: 10px;
  padding: 5px 9px;
  border: 1px solid var(--status-error);
  border-radius: 999px;
  background: var(--bg-primary);
  color: var(--status-error);
  font-size: var(--font-size-caption);
  font-weight: 800;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

.product-card__body {
  display: grid;
  align-content: start;
  gap: 6px;
  padding: 12px;
}

.product-card__name {
  margin: 0;
  font-size: var(--font-size-h2);
  line-height: 1.25;
}

.product-card__description {
  display: -webkit-box;
  min-height: 3em;
  margin: 0;
  overflow: hidden;
  color: var(--text-sub);
  font-size: var(--font-size-caption);
  line-height: 1.5;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.product-card__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding-top: 4px;
}

.product-card__action {
  color: var(--brand-primary);
  font-size: var(--font-size-caption);
  font-weight: 800;
}

.product-card--sold-out .product-card__media,
.product-card--sold-out .product-card__body {
  opacity: 0.72;
}

.product-card--sold-out .product-card__action {
  color: var(--text-sub);
}

@media (prefers-reduced-motion: reduce) {
  .product-card__link {
    transition: none;
  }

  .product-card__link:hover {
    transform: none;
  }
}
</style>
