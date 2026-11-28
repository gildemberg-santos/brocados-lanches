<script setup lang="ts">
import { RouterLink } from 'vue-router'
</script>

<template>
  <main class="admin-page">
    <section class="admin-panel admin-forbidden">
      <p class="admin-forbidden__code">403</p>
      <p class="admin-page__kicker">Acesso restrito</p>
      <h1>Acesso negado</h1>
      <p class="admin-page__lead">Seu papel não permite abrir esta área do painel.</p>
      <RouterLink class="admin-forbidden__back" :to="{ name: 'admin-dashboard' }">
        Voltar ao painel
      </RouterLink>
    </section>
  </main>
</template>

<style scoped>
.admin-forbidden {
  justify-items: start;
  max-width: 720px;
  padding: 32px 24px;
}

.admin-forbidden__code {
  margin: 0;
  color: var(--status-error);
  font-family: var(--font-display);
  font-size: clamp(56px, 12vw, 96px);
  line-height: 0.9;
}

.admin-forbidden h1 {
  margin: 0;
}

.admin-forbidden__back {
  display: inline-flex;
  min-height: 44px;
  align-items: center;
  padding: 10px 18px;
  border: 1px solid var(--bg-border);
  border-radius: 999px;
  color: var(--brand-neon);
  font-weight: 700;
  text-decoration: none;
}

.admin-forbidden__back:hover {
  border-color: var(--brand-primary);
}
</style>
