<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { RouterLink, RouterView, useRouter } from 'vue-router'

import { adminRoleLabel, visibleAdminNav } from '../../features/admin/auth'
import { useAdminAuthStore } from '../../stores/adminAuth'

const auth = useAdminAuthStore()
const router = useRouter()

const items = computed(() => visibleAdminNav(auth.role))
const roleLabel = computed(() => adminRoleLabel(auth.role))
const isDrawerOpen = ref(false)
const menuButton = ref<HTMLButtonElement | null>(null)
const drawer = ref<HTMLElement | null>(null)
const drawerCloseButton = ref<HTMLButtonElement | null>(null)

async function openDrawer(): Promise<void> {
  isDrawerOpen.value = true
  await nextTick()
  drawerCloseButton.value?.focus()
}

async function closeDrawer(): Promise<void> {
  isDrawerOpen.value = false
  await nextTick()
  menuButton.value?.focus()
}

function handleDrawerKeydown(event: KeyboardEvent): void {
  if (event.key === 'Escape') {
    event.preventDefault()
    void closeDrawer()
    return
  }

  if (event.key !== 'Tab' || !drawer.value) {
    return
  }

  const focusable = Array.from(
    drawer.value.querySelectorAll<HTMLElement>('a[href], button:not([disabled])'),
  )
  const first = focusable[0]
  const last = focusable[focusable.length - 1]
  if (!first || !last) {
    return
  }

  if (event.shiftKey && document.activeElement === first) {
    event.preventDefault()
    last.focus()
  } else if (!event.shiftKey && document.activeElement === last) {
    event.preventDefault()
    first.focus()
  }
}

watch(isDrawerOpen, (open) => {
  if (open) {
    window.addEventListener('keydown', handleDrawerKeydown)
  } else {
    window.removeEventListener('keydown', handleDrawerKeydown)
  }
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', handleDrawerKeydown)
})

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

      <button
        ref="menuButton"
        type="button"
        class="admin-shell__menu-button"
        aria-controls="admin-drawer"
        :aria-expanded="isDrawerOpen"
        data-testid="admin-menu-button"
        @click="openDrawer"
      >
        Menu
      </button>

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

      <div id="conteudo" class="admin-shell__main" tabindex="-1">
        <RouterView />
      </div>
    </div>

    <template v-if="isDrawerOpen">
      <div class="admin-shell__backdrop" data-testid="admin-drawer-backdrop" @click="closeDrawer" />

      <aside
        id="admin-drawer"
        ref="drawer"
        class="admin-shell__drawer"
        role="dialog"
        aria-modal="true"
        aria-label="Navegação do painel"
        data-testid="admin-drawer"
      >
        <div class="admin-shell__drawer-header">
          <strong>Menu</strong>
          <button
            ref="drawerCloseButton"
            type="button"
            class="admin-shell__drawer-close"
            data-testid="admin-drawer-close"
            @click="closeDrawer"
          >
            Fechar
          </button>
        </div>

        <nav class="admin-shell__drawer-nav" aria-label="Navegação completa do painel">
          <RouterLink
            v-for="item in items"
            :key="item.name"
            class="admin-shell__link"
            :to="{ name: item.name }"
            @click="closeDrawer"
          >
            {{ item.label }}
          </RouterLink>
        </nav>
      </aside>
    </template>
  </div>
</template>

<style scoped>
.admin-shell {
  min-height: 100vh;
  min-height: 100dvh;
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
  padding: calc(8px + env(safe-area-inset-top)) 16px 8px;
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
  padding: 24px 16px 48px;
}

.admin-shell__menu-button {
  display: inline-flex;
  min-height: 44px;
  align-items: center;
  padding: 8px 14px;
  border: 1px solid var(--bg-border);
  border-radius: 999px;
  background: var(--bg-input);
  color: var(--text-main);
  font-weight: 700;
  cursor: pointer;
}

.admin-shell__menu-button:hover {
  border-color: var(--brand-primary);
}

.admin-shell__backdrop {
  position: fixed;
  inset: 0;
  z-index: 50;
  background: color-mix(in srgb, var(--bg-primary) 80%, transparent);
}

.admin-shell__drawer {
  position: fixed;
  top: 0;
  right: 0;
  bottom: 0;
  z-index: 51;
  display: flex;
  width: min(320px, 85vw);
  flex-direction: column;
  overflow-y: auto;
  padding: 16px 16px calc(16px + env(safe-area-inset-bottom));
  border-left: 1px solid var(--bg-border);
  background: var(--bg-surface);
}

.admin-shell__drawer-header {
  display: flex;
  min-height: 44px;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--bg-border);
  font-family: var(--font-display);
  font-size: var(--font-size-h2);
  letter-spacing: 0.04em;
}

.admin-shell__drawer-close {
  min-height: 44px;
  padding: 8px 14px;
  border: 1px solid var(--bg-border);
  border-radius: 999px;
  background: var(--bg-input);
  color: var(--text-main);
  font-family: var(--font-sans);
  font-size: var(--font-size-caption);
  font-weight: 700;
  cursor: pointer;
}

.admin-shell__drawer-close:hover {
  border-color: var(--brand-primary);
}

.admin-shell__drawer-nav {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding-bottom: 24px;
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

  .admin-shell__menu-button,
  .admin-shell__backdrop,
  .admin-shell__drawer {
    display: none;
  }
}

@media (max-width: 419px) {
  .admin-shell__user,
  .admin-shell__role {
    display: none;
  }
}
</style>
