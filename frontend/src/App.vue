<script setup lang="ts">
import { computed } from 'vue'
import { RouterView, useRoute } from 'vue-router'

import MobileTabBar from './components/MobileTabBar.vue'
import SiteFooter from './components/SiteFooter.vue'
import SiteHeader from './components/SiteHeader.vue'
import StructuredData from './components/StructuredData.vue'
import { useCartStore } from './stores/cart'

const route = useRoute()
const cart = useCartStore()
// O login também faz parte do painel, embora não use o meta `admin`.
const isAdminRoute = computed(() => route.meta.admin === true || route.path.startsWith('/admin'))
// Checkout e produto escondem a barra para dar espaço ao CTA principal.
const showTabBar = computed(() => !isAdminRoute.value && route.meta.hideTabBar !== true)
// A home exibe a barra de carrinho acima da navegação inferior.
const showCartBar = computed(() => !isAdminRoute.value && route.name === 'home' && !cart.isEmpty)
const footerBottomSpace = computed<'none' | 'tab' | 'cart'>(() => {
  if (showCartBar.value) {
    return 'cart'
  }
  return showTabBar.value ? 'tab' : 'none'
})
</script>

<template>
  <a class="skip-link" href="#conteudo">Pular para o conteúdo</a>
  <SiteHeader v-if="!isAdminRoute" />
  <StructuredData v-if="!isAdminRoute" />
  <RouterView />
  <SiteFooter v-if="!isAdminRoute" />
  <!-- Espaço em fluxo para as barras fixas: o rodapé termina acima delas. -->
  <div
    v-if="!isAdminRoute && footerBottomSpace !== 'none'"
    class="mobile-bottom-spacer"
    :class="`mobile-bottom-spacer--${footerBottomSpace}`"
    data-testid="mobile-bottom-spacer"
    aria-hidden="true"
  />
  <MobileTabBar v-if="showTabBar" />
</template>

<style scoped>
.mobile-bottom-spacer {
  background: var(--bg-primary);
}

.mobile-bottom-spacer--tab {
  height: calc(72px + env(safe-area-inset-bottom));
}

.mobile-bottom-spacer--cart {
  height: calc(140px + env(safe-area-inset-bottom));
}

@media (min-width: 768px) {
  .mobile-bottom-spacer {
    display: none;
  }
}
</style>
