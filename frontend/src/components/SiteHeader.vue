<script setup lang="ts">
import { RouterLink } from 'vue-router'

import { useCartStore } from '../stores/cart'

const cart = useCartStore()
</script>

<template>
  <header class="site-header">
    <div class="site-header__inner">
      <RouterLink class="site-header__brand" to="/" aria-label="Brocados Lanches, início">
        <span class="site-header__brand-name">BROCADOS</span>
        <span class="site-header__brand-accent">LANCHES</span>
      </RouterLink>

      <nav class="site-header__nav" aria-label="Navegação principal">
        <RouterLink class="site-header__link" to="/">Início</RouterLink>
        <RouterLink class="site-header__link" to="/cardapio">Cardápio</RouterLink>
        <RouterLink
          class="site-header__link site-header__link--cart"
          to="/carrinho"
          :aria-label="cart.count > 0 ? `Carrinho, ${cart.count} itens` : 'Carrinho'"
        >
          Carrinho
          <span v-if="cart.count > 0" class="site-header__badge" aria-hidden="true">
            {{ cart.count }}
          </span>
        </RouterLink>
      </nav>
    </div>
  </header>
</template>

<style scoped>
.site-header {
  position: sticky;
  top: 0;
  z-index: 40;
  border-bottom: 1px solid var(--bg-border);
  background: var(--bg-primary);
}

.site-header__inner {
  display: flex;
  width: min(1120px, 100%);
  min-height: 64px;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin: 0 auto;
  padding: 8px 16px;
}

.site-header__brand {
  display: inline-flex;
  min-height: 44px;
  align-items: center;
  gap: 6px;
  color: var(--brand-primary);
  text-decoration: none;
}

.site-header__brand-name {
  font-family: var(--font-display);
  font-size: 28px;
  line-height: var(--line-height-display);
  letter-spacing: 0.04em;
}

.site-header__brand-accent {
  color: var(--text-main);
  font-size: var(--font-size-caption);
  font-weight: 800;
  letter-spacing: 0.16em;
}

.site-header__nav {
  display: flex;
  align-items: center;
  gap: 4px;
}

.site-header__link {
  display: inline-flex;
  min-height: 44px;
  align-items: center;
  padding: 8px 12px;
  border-radius: 999px;
  color: var(--text-sub);
  font-weight: 700;
  text-decoration: none;
}

.site-header__link:hover,
.site-header__link.router-link-active {
  color: var(--text-main);
}

.site-header__link--cart {
  gap: 6px;
}

.site-header__badge {
  display: inline-grid;
  min-width: 24px;
  height: 24px;
  place-items: center;
  border-radius: 999px;
  background: var(--brand-primary);
  color: var(--bg-primary);
  font-size: var(--font-size-caption);
  font-weight: 800;
}

@media (max-width: 519px) {
  .site-header__brand-name {
    font-size: 26px;
  }

  .site-header__brand-accent {
    display: none;
  }

  .site-header__link {
    padding-inline: 8px;
  }

  .site-header__nav .site-header__link:first-child {
    display: none;
  }
}

@media (min-width: 768px) {
  .site-header__inner {
    padding-inline: 24px;
  }

  .site-header__brand-name {
    font-size: 34px;
  }
}
</style>
