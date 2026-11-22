<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink, useRoute } from 'vue-router'

const route = useRoute()
const title = computed(() =>
  typeof route.meta.title === 'string' ? route.meta.title : 'Área do painel',
)
</script>

<template>
  <main id="conteudo" class="admin-page">
    <p class="admin-page__kicker">Painel administrativo</p>
    <h1>{{ title }}</h1>
    <p class="admin-page__lead">
      Esta área está em preparação. Ela será ligada à API administrativa na próxima entrega.
    </p>
    <RouterLink class="admin-page__back" :to="{ name: 'admin-dashboard' }">
      Voltar ao painel
    </RouterLink>
  </main>
</template>

<style scoped>
.admin-page__kicker {
  margin: 0;
  color: var(--brand-neon);
  font-size: var(--font-size-caption);
  font-weight: 800;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.admin-page h1 {
  margin: 8px 0;
  font-family: var(--font-display);
  font-size: var(--font-size-display);
  line-height: var(--line-height-display);
  letter-spacing: 0.02em;
}

.admin-page__lead {
  margin: 0 0 24px;
  color: var(--text-sub);
}

.admin-page__back {
  display: inline-flex;
  min-height: 44px;
  align-items: center;
  color: var(--brand-neon);
  font-weight: 700;
}
</style>
