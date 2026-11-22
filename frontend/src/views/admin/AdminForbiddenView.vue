<script setup lang="ts">
import { RouterLink } from 'vue-router'
</script>

<template>
  <main id="conteudo" class="admin-forbidden">
    <p class="admin-forbidden__code">403</p>
    <h1>Acesso negado</h1>
    <p>Seu papel não permite abrir esta área do painel.</p>
    <RouterLink class="admin-forbidden__back" :to="{ name: 'admin-dashboard' }">
      Voltar ao painel
    </RouterLink>
  </main>
</template>

<style scoped>
.admin-forbidden {
  display: grid;
  justify-items: start;
  gap: 12px;
  padding: 32px 0;
}

.admin-forbidden__code {
  margin: 0;
  color: var(--status-error);
  font-family: var(--font-display);
  font-size: var(--font-size-display);
  line-height: var(--line-height-display);
}

.admin-forbidden h1 {
  margin: 0;
  font-family: var(--font-display);
  font-size: var(--font-size-display);
  line-height: var(--line-height-display);
}

.admin-forbidden p {
  margin: 0;
  color: var(--text-sub);
}

.admin-forbidden__back {
  display: inline-flex;
  min-height: 44px;
  align-items: center;
  color: var(--brand-neon);
  font-weight: 700;
}
</style>
