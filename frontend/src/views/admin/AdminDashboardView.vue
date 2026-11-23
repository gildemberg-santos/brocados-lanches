<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'

import { adminRoleLabel, primaryAdminNav } from '../../features/admin/auth'
import { useAdminAuthStore } from '../../stores/adminAuth'

const auth = useAdminAuthStore()
const shortcuts = computed(() =>
  primaryAdminNav(auth.role).filter((item) => item.name !== 'admin-dashboard'),
)
</script>

<template>
  <main id="conteudo" class="admin-page">
    <p class="admin-page__kicker">Painel administrativo</p>
    <h1>Olá, {{ auth.userName }}</h1>
    <p class="admin-page__lead">
      Sessão ativa como {{ adminRoleLabel(auth.role) }}. A autorização de cada ação é validada no
      servidor.
    </p>

    <section class="admin-page__section" aria-labelledby="admin-shortcuts-title">
      <h2 id="admin-shortcuts-title">Atalhos</h2>
      <div class="admin-page__grid">
        <RouterLink
          v-for="item in shortcuts"
          :key="item.name"
          class="admin-page__card"
          :to="{ name: item.name }"
        >
          {{ item.label }}
        </RouterLink>
      </div>
    </section>

    <p class="admin-page__note">
      A gestão de catálogo e a fila de pedidos serão disponibilizadas nas próximas entregas.
    </p>
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
  margin: 0 0 28px;
  color: var(--text-sub);
}

.admin-page__section {
  display: grid;
  gap: 12px;
}

.admin-page__section h2 {
  margin: 0;
  font-size: var(--font-size-h2);
}

.admin-page__grid {
  display: grid;
  gap: 12px;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
}

.admin-page__card {
  display: flex;
  min-height: 88px;
  align-items: center;
  padding: 16px;
  border: 1px solid var(--bg-border);
  border-radius: 16px;
  background: var(--bg-surface);
  color: var(--text-main);
  font-weight: 700;
  text-decoration: none;
}

.admin-page__card:hover {
  border-color: var(--brand-primary);
}

.admin-page__note {
  margin: 28px 0 0;
  padding: 16px;
  border: 1px solid var(--bg-border);
  border-radius: 16px;
  color: var(--text-sub);
}
</style>
