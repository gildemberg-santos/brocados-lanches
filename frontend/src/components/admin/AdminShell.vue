<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink, RouterView, useRouter } from 'vue-router'

import { adminRoleLabel, primaryAdminNav, visibleAdminNav } from '../../features/admin/auth'
import { useAdminAuthStore } from '../../stores/adminAuth'

const auth = useAdminAuthStore()
const router = useRouter()

const items = computed(() => visibleAdminNav(auth.role))
const primaryItems = computed(() => primaryAdminNav(auth.role))
const roleLabel = computed(() => adminRoleLabel(auth.role))

async function logout(): Promise<void> {
  await auth.logout()
  await router.replace({ name: 'admin-login' })
}
</script>

<template>
  <div class="admin-shell">
    <header class="admin-shell__header">
      <RouterLink class="admin-shell__brand" :to="{ name: 'admin-dashboard' }">
        <span class="admin-shell__brand-name">BROCADOS</span>
        <span class="admin-shell__brand-tag">ADMIN</span>
      </RouterLink>

      <div class="admin-shell__identity">
        <span class="admin-shell__user">{{ auth.userName || 'Operação' }}</span>
        <span class="admin-shell__role">{{ roleLabel }}</span>
        <button type="button" class="admin-shell__logout" @click="logout">Sair</button>
      </div>
    </header>

    <div class="admin-shell__body">
      <nav class="admin-shell__sidebar" aria-label="Navegação do painel">
        <RouterLink
          v-for="item in items"
          :key="item.name"
          class="admin-shell__link"
          :to="{ name: item.name }"
        >
          {{ item.label }}
        </RouterLink>
      </nav>

      <main id="conteudo" class="admin-shell__main">
        <RouterView />
      </main>
    </div>

    <nav class="admin-shell__mobile-nav" aria-label="Navegação principal do painel">
      <RouterLink
        v-for="item in primaryItems"
        :key="item.name"
        class="admin-shell__mobile-link"
        :to="{ name: item.name }"
      >
        {{ item.label }}
      </RouterLink>
    </nav>
  </div>
</template>

<style scoped>
.admin-shell {
  min-height: 100vh;
  background: var(--bg-primary);
}

.admin-shell__header {
  position: sticky;
  top: 0;
  z-index: 40;
  display: flex;
  min-height: 64px;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 8px 16px;
  border-bottom: 1px solid var(--bg-border);
  background: var(--bg-primary);
}

.admin-shell__brand {
  display: inline-flex;
  min-height: 44px;
  align-items: center;
  gap: 8px;
  color: var(--brand-primary);
  text-decoration: none;
}

.admin-shell__brand-name {
  font-family: var(--font-display);
  font-size: 26px;
  line-height: var(--line-height-display);
  letter-spacing: 0.04em;
}

.admin-shell__brand-tag {
  color: var(--text-sub);
  font-size: var(--font-size-caption);
  font-weight: 800;
  letter-spacing: 0.16em;
}

.admin-shell__identity {
  display: flex;
  align-items: center;
  gap: 10px;
}

.admin-shell__user {
  color: var(--text-main);
  font-weight: 700;
}

.admin-shell__role {
  padding: 4px 10px;
  border: 1px solid var(--bg-border);
  border-radius: 999px;
  color: var(--text-sub);
  font-size: var(--font-size-caption);
  font-weight: 700;
}

.admin-shell__logout {
  min-height: 44px;
  padding: 8px 16px;
  border: 1px solid var(--bg-border);
  border-radius: 999px;
  background: var(--bg-input);
  color: var(--text-main);
  font-weight: 700;
  cursor: pointer;
}

.admin-shell__logout:hover {
  border-color: var(--brand-primary);
}

.admin-shell__body {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
}

.admin-shell__sidebar {
  display: none;
}

.admin-shell__main {
  width: min(1120px, 100%);
  margin: 0 auto;
  padding: 24px 16px 104px;
}

.admin-shell__mobile-nav {
  position: fixed;
  right: 0;
  bottom: 0;
  left: 0;
  z-index: 40;
  display: flex;
  overflow-x: auto;
  border-top: 1px solid var(--bg-border);
  background: var(--bg-surface);
}

.admin-shell__mobile-link {
  display: grid;
  min-width: 88px;
  min-height: 56px;
  flex: 1;
  place-items: center;
  padding: 8px;
  color: var(--text-sub);
  font-size: var(--font-size-caption);
  font-weight: 700;
  text-align: center;
  text-decoration: none;
}

.admin-shell__mobile-link.router-link-exact-active {
  color: var(--brand-neon);
}

.admin-shell__link {
  display: flex;
  min-height: 44px;
  align-items: center;
  padding: 8px 14px;
  border-radius: 12px;
  color: var(--text-sub);
  font-weight: 700;
  text-decoration: none;
}

.admin-shell__link:hover,
.admin-shell__link.router-link-exact-active {
  background: var(--bg-surface);
  color: var(--text-main);
}

@media (min-width: 768px) {
  .admin-shell__header {
    padding-inline: 24px;
  }

  .admin-shell__body {
    grid-template-columns: 232px minmax(0, 1fr);
  }

  .admin-shell__sidebar {
    display: flex;
    flex-direction: column;
    gap: 4px;
    padding: 24px 12px;
    border-right: 1px solid var(--bg-border);
  }

  .admin-shell__main {
    padding: 32px 24px 80px;
  }

  .admin-shell__mobile-nav {
    display: none;
  }
}

@media (max-width: 419px) {
  .admin-shell__user {
    display: none;
  }
}
</style>
