<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'

import { adminRoleLabel, primaryAdminNav } from '../../features/admin/auth'
import { fetchAdminProducts } from '../../services/adminCatalog'
import { useAdminAuthStore } from '../../stores/adminAuth'
import { useAdminOrdersStore } from '../../stores/adminOrders'
import { useDeliveryZonesStore } from '../../stores/deliveryZones'
import { useStoreStatusStore } from '../../stores/storeStatus'

const UNAVAILABLE = '—'

const auth = useAdminAuthStore()
const statusStore = useStoreStatusStore()
const ordersStore = useAdminOrdersStore()
const zonesStore = useDeliveryZonesStore()
const activeProductCount = ref<number | null>(null)

const shortcuts = computed(() =>
  primaryAdminNav(auth.role).filter((item) => item.name !== 'admin-dashboard'),
)

function countOrDash(value: number | null): string {
  return value !== null && Number.isFinite(value) && value > 0 ? String(value) : UNAVAILABLE
}

const storeStatusLabel = computed(() => {
  if (statusStore.state !== 'success' || !statusStore.status) {
    return UNAVAILABLE
  }
  return statusStore.status.status === 'open' ? 'Aberta' : 'Fechada'
})

const queuedOrdersLabel = computed(() =>
  ordersStore.state === 'success' ? countOrDash(ordersStore.meta.total) : UNAVAILABLE,
)

const activeProductsLabel = computed(() => countOrDash(activeProductCount.value))

const activeZonesLabel = computed(() =>
  zonesStore.state === 'success' ? countOrDash(zonesStore.zones.length) : UNAVAILABLE,
)

async function loadActiveProducts(): Promise<void> {
  try {
    const page = await fetchAdminProducts({ active: true, pageSize: 1 })
    activeProductCount.value = page.meta.total
  } catch {
    activeProductCount.value = null
  }
}

onMounted(() => {
  void Promise.allSettled([
    statusStore.load(),
    ordersStore.state !== 'success' ? ordersStore.load() : Promise.resolve(),
    zonesStore.ensureLoaded(),
    loadActiveProducts(),
  ])
})
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

    <section class="admin-page__section" aria-labelledby="admin-overview-title">
      <h2 id="admin-overview-title">Visão geral</h2>
      <dl class="admin-page__metrics">
        <div class="admin-page__metric" data-testid="metric-store-status">
          <dt>Status da loja</dt>
          <dd :class="storeStatusLabel === 'Aberta' ? 'is-open' : undefined">
            {{ storeStatusLabel }}
          </dd>
        </div>
        <div class="admin-page__metric" data-testid="metric-queued-orders">
          <dt>Pedidos na fila</dt>
          <dd>{{ queuedOrdersLabel }}</dd>
        </div>
        <div class="admin-page__metric" data-testid="metric-active-products">
          <dt>Produtos ativos</dt>
          <dd>{{ activeProductsLabel }}</dd>
        </div>
        <div class="admin-page__metric" data-testid="metric-active-zones">
          <dt>Zonas ativas</dt>
          <dd>{{ activeZonesLabel }}</dd>
        </div>
      </dl>
      <p class="admin-page__hint">
        Indicadores lidos da API. O travessão aparece quando não há dado disponível.
      </p>
    </section>
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

.admin-page__metrics {
  display: grid;
  gap: 12px;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  margin: 0;
}

.admin-page__metric {
  display: grid;
  gap: 6px;
  padding: 16px;
  border: 1px solid var(--bg-border);
  border-radius: 16px;
  background: var(--bg-surface);
}

.admin-page__metric dt {
  color: var(--text-sub);
  font-size: var(--font-size-caption);
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.admin-page__metric dd {
  margin: 0;
  font-family: var(--font-display);
  font-size: var(--font-size-h1);
  line-height: 1.2;
}

.admin-page__metric dd.is-open {
  color: var(--status-open);
}

.admin-page__hint {
  margin: 0;
  color: var(--text-sub);
  font-size: var(--font-size-caption);
}
</style>
