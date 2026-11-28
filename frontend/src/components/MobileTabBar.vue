<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'

import { useCartStore } from '../stores/cart'

const cart = useCartStore()

const cartLabel = computed(() => (cart.count > 0 ? `Carrinho, ${cart.count} itens` : 'Carrinho'))
</script>

<template>
  <nav class="mobile-tab-bar" aria-label="Navegação inferior" data-testid="mobile-tab-bar">
    <RouterLink class="mobile-tab-bar__link" to="/">
      <span class="mobile-tab-bar__label">Início</span>
    </RouterLink>
    <RouterLink class="mobile-tab-bar__link" to="/cardapio">
      <span class="mobile-tab-bar__label">Cardápio</span>
    </RouterLink>
    <RouterLink class="mobile-tab-bar__link" to="/meus-pedidos">
      <span class="mobile-tab-bar__label">Pedidos</span>
    </RouterLink>
    <RouterLink
      class="mobile-tab-bar__link mobile-tab-bar__link--cart"
      to="/carrinho"
      :aria-label="cartLabel"
    >
      <span class="mobile-tab-bar__label">
        Carrinho
        <span v-if="cart.count > 0" class="mobile-tab-bar__badge" aria-hidden="true">
          {{ cart.count }}
        </span>
      </span>
    </RouterLink>
  </nav>
</template>

<style scoped>
.mobile-tab-bar {
  position: fixed;
  right: 0;
  bottom: 0;
  left: 0;
  z-index: 40;
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  padding-bottom: env(safe-area-inset-bottom);
  border-top: 1px solid var(--bg-border);
  background: var(--bg-surface);
}

.mobile-tab-bar__link {
  position: relative;
  display: flex;
  min-width: 0;
  min-height: 60px;
  align-items: center;
  justify-content: center;
  padding: 8px 4px;
  color: var(--text-sub);
  font-size: var(--font-size-caption);
  font-weight: 700;
  text-align: center;
  text-decoration: none;
}

.mobile-tab-bar__link:hover,
.mobile-tab-bar__link.router-link-active {
  color: var(--brand-primary);
}

.mobile-tab-bar__link.router-link-active::before {
  position: absolute;
  top: 0;
  width: 28px;
  height: 3px;
  background: var(--brand-primary);
  content: '';
}

.mobile-tab-bar__label {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.mobile-tab-bar__badge {
  display: inline-grid;
  min-width: 20px;
  height: 20px;
  place-items: center;
  padding-inline: 4px;
  border-radius: 999px;
  background: var(--brand-primary);
  color: var(--bg-primary);
  font-size: var(--font-size-caption);
  font-weight: 800;
  line-height: 1;
}

@media (min-width: 768px) {
  .mobile-tab-bar {
    display: none;
  }
}
</style>
