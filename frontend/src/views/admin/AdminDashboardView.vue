<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'

import BaseButton from '../../components/BaseButton.vue'
import OrderStatusBadge from '../../components/admin/OrderStatusBadge.vue'
import { adminRoleLabel, primaryAdminNav } from '../../features/admin/auth'
import { formatOrderDateTime } from '../../features/admin-orders/format'
import {
  ORDER_STATUSES,
  ORDER_STATUS_LABELS,
  type OrderStatus,
} from '../../features/admin-orders/orderStatus'
import { fetchAdminProducts } from '../../services/adminCatalog'
import { fetchAdminOrders } from '../../services/adminOrders'
import { useAdminAuthStore } from '../../stores/adminAuth'
import { useAdminOrdersStore } from '../../stores/adminOrders'
import { useAdminStoreSettingsStore } from '../../stores/adminStoreSettings'
import { useDeliveryZonesStore } from '../../stores/deliveryZones'
import { useStoreStatusStore } from '../../stores/storeStatus'
import type { AdminStoreInput } from '../../types/adminCatalog'

const UNAVAILABLE = '—'
const MAX_PRODUCT_PAGES = 20
const OPEN_STATUSES: OrderStatus[] = [
  'pending',
  'confirmed',
  'preparing',
  'ready',
  'out_for_delivery',
]

const auth = useAdminAuthStore()
const statusStore = useStoreStatusStore()
const settingsStore = useAdminStoreSettingsStore()
const ordersStore = useAdminOrdersStore()
const zonesStore = useDeliveryZonesStore()
const statusCounts = ref<Partial<Record<OrderStatus, number>>>({})
const productCounts = ref<{ available: number; total: number } | null>(null)
const pendingKillSwitch = ref(false)

const shortcuts = computed(() =>
  primaryAdminNav(auth.role).filter((item) => item.name !== 'admin-dashboard'),
)

function formatCount(value: number | undefined): string {
  return typeof value === 'number' && Number.isFinite(value) ? String(value) : UNAVAILABLE
}

const storeStatusLabel = computed(() => {
  if (statusStore.state !== 'success' || !statusStore.status) {
    return UNAVAILABLE
  }
  return statusStore.status.status === 'open' ? 'Aberta' : 'Fechada'
})

const manualSwitchLabel = computed(() => {
  if (settingsStore.state !== 'success' || !settingsStore.settings) {
    return UNAVAILABLE
  }
  return settingsStore.settings.isOpen ? 'Ligado' : 'Desligado'
})

const deliveryTimeLabel = computed(() => statusStore.status?.deliveryTime.trim() || UNAVAILABLE)

const queuedOrdersLabel = computed(() => {
  const values = OPEN_STATUSES.map((status) => statusCounts.value[status]).filter(
    (value): value is number => typeof value === 'number',
  )
  if (values.length === 0) {
    return UNAVAILABLE
  }
  return String(values.reduce((sum, value) => sum + value, 0))
})

const productsLabel = computed(() =>
  productCounts.value
    ? `${productCounts.value.available} / ${productCounts.value.total}`
    : UNAVAILABLE,
)

const activeZonesLabel = computed(() =>
  zonesStore.state === 'success' ? formatCount(zonesStore.zones.length) : UNAVAILABLE,
)

const statusBreakdown = computed(() =>
  ORDER_STATUSES.map((status) => ({
    status,
    label: ORDER_STATUS_LABELS[status],
    count: formatCount(statusCounts.value[status]),
  })),
)

const recentOrders = computed(() =>
  ordersStore.state === 'success' ? ordersStore.orders.slice(0, 5) : [],
)

const canToggleKillSwitch = computed(
  () => settingsStore.state === 'success' && settingsStore.settings !== null,
)
const killSwitchLabel = computed(() =>
  settingsStore.settings?.isOpen ? 'Fechar loja' : 'Abrir loja',
)
const confirmationLabel = computed(() =>
  settingsStore.settings?.isOpen ? 'Confirmar fechamento' : 'Confirmar abertura',
)

async function loadStatusCounts(): Promise<void> {
  const results = await Promise.allSettled(
    ORDER_STATUSES.map(async (status) => {
      const page = await fetchAdminOrders({ status, pageSize: 1 })
      return { status, total: page.meta.total }
    }),
  )
  const next: Partial<Record<OrderStatus, number>> = {}
  results.forEach((result, index) => {
    if (result.status === 'fulfilled') {
      next[ORDER_STATUSES[index]] = result.value.total
    }
  })
  statusCounts.value = next
}

async function loadProductCounts(): Promise<void> {
  try {
    let available = 0
    let total = 0
    let totalPages = 1

    for (let pageNumber = 1; pageNumber <= totalPages && pageNumber <= MAX_PRODUCT_PAGES;) {
      const page = await fetchAdminProducts({ active: true, page: pageNumber, pageSize: 100 })
      total = page.meta.total
      totalPages = page.meta.totalPages
      available += page.data.filter((product) => product.available).length
      pageNumber += 1
    }

    productCounts.value = totalPages > MAX_PRODUCT_PAGES ? null : { available, total }
  } catch {
    productCounts.value = null
  }
}

async function toggleKillSwitch(): Promise<void> {
  const settings = settingsStore.settings
  if (!settings) {
    return
  }

  const input: AdminStoreInput = {
    name: settings.name,
    isOpen: !settings.isOpen,
    openingHours: settings.openingHours,
    deliveryTime: settings.deliveryTime,
    whatsapp: settings.whatsapp,
    instagram: settings.instagram,
    address: settings.address,
  }
  if (await settingsStore.save(input)) {
    pendingKillSwitch.value = false
    await statusStore.refresh()
  }
}

onMounted(() => {
  void Promise.allSettled([
    statusStore.load(),
    settingsStore.load(),
    ordersStore.state !== 'success' ? ordersStore.load() : Promise.resolve(),
    zonesStore.ensureLoaded(),
    loadStatusCounts(),
    loadProductCounts(),
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

    <section class="admin-page__section" aria-labelledby="admin-kill-switch-title">
      <h2 id="admin-kill-switch-title">Controle da loja</h2>
      <div class="kill-switch" data-testid="kill-switch">
        <dl class="kill-switch__info">
          <div>
            <dt>Status efetivo</dt>
            <dd>{{ storeStatusLabel }}</dd>
          </div>
          <div>
            <dt>Interruptor manual</dt>
            <dd>{{ manualSwitchLabel }}</dd>
          </div>
        </dl>
        <div v-if="pendingKillSwitch" class="kill-switch__actions">
          <BaseButton :loading="settingsStore.saving" @click="toggleKillSwitch">
            {{ confirmationLabel }}
          </BaseButton>
          <BaseButton variant="secondary" @click="pendingKillSwitch = false">Cancelar</BaseButton>
        </div>
        <BaseButton
          v-else
          :disabled="!canToggleKillSwitch || settingsStore.saving"
          @click="pendingKillSwitch = true"
        >
          {{ killSwitchLabel }}
        </BaseButton>
      </div>
      <p v-if="settingsStore.saveSuccess" class="admin-page__hint" role="status">
        {{ settingsStore.saveSuccess }}
      </p>
      <p v-else-if="settingsStore.saveError" class="admin-page__hint" role="alert">
        {{ settingsStore.saveError }}
      </p>
      <p class="admin-page__hint">
        A alteração passa pelo servidor, gera auditoria e só vale quando o horário permitir.
      </p>
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
          <dt>Pedidos em andamento</dt>
          <dd>{{ queuedOrdersLabel }}</dd>
        </div>
        <div class="admin-page__metric" data-testid="metric-active-products">
          <dt>Produtos disponíveis / total</dt>
          <dd>{{ productsLabel }}</dd>
        </div>
        <div class="admin-page__metric" data-testid="metric-active-zones">
          <dt>Zonas ativas</dt>
          <dd>{{ activeZonesLabel }}</dd>
        </div>
        <div class="admin-page__metric" data-testid="metric-delivery-time">
          <dt>Prazo de entrega</dt>
          <dd>{{ deliveryTimeLabel }}</dd>
        </div>
      </dl>
      <p class="admin-page__hint">
        Indicadores lidos da API. O travessão aparece quando não há dado disponível.
      </p>
    </section>

    <section class="admin-page__section" aria-labelledby="admin-status-counts-title">
      <h2 id="admin-status-counts-title">Pedidos por status</h2>
      <dl class="admin-page__metrics">
        <div
          v-for="item in statusBreakdown"
          :key="item.status"
          class="admin-page__metric"
          :data-testid="`metric-status-${item.status}`"
        >
          <dt>{{ item.label }}</dt>
          <dd>{{ item.count }}</dd>
        </div>
      </dl>
    </section>

    <section class="admin-page__section" aria-labelledby="admin-recent-orders-title">
      <h2 id="admin-recent-orders-title">Pedidos recentes</h2>
      <p v-if="recentOrders.length === 0" class="admin-page__hint">
        Nenhum pedido recente disponível no momento.
      </p>
      <ol v-else class="recent-orders">
        <li v-for="order in recentOrders" :key="order.id" class="recent-orders__item">
          <RouterLink
            class="recent-orders__link"
            :to="{ name: 'admin-order-detail', params: { id: order.id } }"
          >
            #{{ order.orderNumber }}
          </RouterLink>
          <OrderStatusBadge :status="order.status" />
          <time :datetime="order.createdAt">{{ formatOrderDateTime(order.createdAt) }}</time>
        </li>
      </ol>
      <RouterLink class="admin-page__card admin-page__card--link" :to="{ name: 'admin-orders' }">
        Ver todos os pedidos
      </RouterLink>
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

.admin-page__card--link {
  min-height: 44px;
  border-radius: 999px;
}

.kill-switch {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 16px;
  border: 1px solid var(--bg-border);
  border-radius: 16px;
  background: var(--bg-surface);
}

.kill-switch__info {
  display: flex;
  flex-wrap: wrap;
  gap: 24px;
  margin: 0;
}

.kill-switch__info dt {
  color: var(--text-sub);
  font-size: var(--font-size-caption);
  font-weight: 700;
}

.kill-switch__info dd {
  margin: 0;
  font-size: var(--font-size-body);
  font-weight: 700;
}

.kill-switch__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
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

.recent-orders {
  display: grid;
  gap: 10px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.recent-orders__item {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  border: 1px solid var(--bg-border);
  border-radius: 12px;
  background: var(--bg-surface);
  font-size: var(--font-size-caption);
}

.recent-orders__item time {
  margin-left: auto;
  color: var(--text-sub);
}

.recent-orders__link {
  display: inline-flex;
  min-height: 44px;
  align-items: center;
  color: var(--brand-primary);
  font-weight: 800;
  text-decoration: none;
}
</style>
