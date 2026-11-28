<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'

import BaseButton from '../../components/BaseButton.vue'
import OrderStatusBadge from '../../components/admin/OrderStatusBadge.vue'
import { adminRoleLabel, primaryAdminNav } from '../../features/admin/auth'
import { formatOrderDateTime } from '../../features/admin-orders/format'
import {
  FULFILLMENT_LABELS,
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

interface StatusChartItem {
  status: OrderStatus
  label: string
  count: number | null
  share: number
}

const auth = useAdminAuthStore()
const statusStore = useStoreStatusStore()
const settingsStore = useAdminStoreSettingsStore()
const ordersStore = useAdminOrdersStore()
const zonesStore = useDeliveryZonesStore()
const statusCounts = ref<Partial<Record<OrderStatus, number>>>({})
const productCounts = ref<{ available: number; total: number } | null>(null)
const fulfillmentCounts = ref<{ delivery: number | null; pickup: number | null }>({
  delivery: null,
  pickup: null,
})
const pendingKillSwitch = ref(false)

const shortcuts = computed(() =>
  primaryAdminNav(auth.role).filter((item) => item.name !== 'admin-dashboard'),
)

function formatCount(value: number | undefined | null): string {
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

const queuedOrdersValue = computed(() => {
  const values = OPEN_STATUSES.map((status) => statusCounts.value[status]).filter(
    (value): value is number => typeof value === 'number',
  )
  if (values.length === 0) {
    return null
  }
  return values.reduce((sum, value) => sum + value, 0)
})

const queuedOrdersLabel = computed(() => formatCount(queuedOrdersValue.value))

const totalOrdersValue = computed(() => {
  const values = ORDER_STATUSES.map((status) => statusCounts.value[status]).filter(
    (value): value is number => typeof value === 'number',
  )
  if (values.length === 0) {
    return null
  }
  return values.reduce((sum, value) => sum + value, 0)
})

const totalOrdersLabel = computed(() => formatCount(totalOrdersValue.value))

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

const statusChart = computed<StatusChartItem[]>(() => {
  const items = ORDER_STATUSES.map((status) => ({
    status,
    label: ORDER_STATUS_LABELS[status],
    count:
      typeof statusCounts.value[status] === 'number'
        ? (statusCounts.value[status] as number)
        : null,
  }))
  const max = items.reduce((current, item) => Math.max(current, item.count ?? 0), 0)
  return items.map((item) => ({
    ...item,
    share: max > 0 && item.count !== null ? (item.count / max) * 100 : 0,
  }))
})

const hasStatusChartData = computed(() => statusChart.value.some((item) => item.count !== null))

const fulfillmentTotal = computed(() => {
  const { delivery, pickup } = fulfillmentCounts.value
  if (typeof delivery !== 'number' || typeof pickup !== 'number') {
    return null
  }
  return delivery + pickup
})

const fulfillmentDeliveryShare = computed(() => {
  const total = fulfillmentTotal.value
  const delivery = fulfillmentCounts.value.delivery
  if (total === null || total <= 0 || typeof delivery !== 'number') {
    return 0
  }
  return Math.round((delivery / total) * 100)
})

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

const storeStatusTone = computed(() =>
  storeStatusLabel.value === 'Aberta' ? 'is-open' : 'is-neutral',
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
    const status = ORDER_STATUSES[index]
    if (status && result.status === 'fulfilled') {
      next[status] = result.value.total
    }
  })
  statusCounts.value = next
}

async function loadFulfillmentCounts(): Promise<void> {
  const [delivery, pickup] = await Promise.allSettled([
    fetchAdminOrders({ fulfillment: 'delivery', pageSize: 1 }),
    fetchAdminOrders({ fulfillment: 'pickup', pageSize: 1 }),
  ])
  fulfillmentCounts.value = {
    delivery: delivery.status === 'fulfilled' ? delivery.value.meta.total : null,
    pickup: pickup.status === 'fulfilled' ? pickup.value.meta.total : null,
  }
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
    loadFulfillmentCounts(),
    loadProductCounts(),
  ])
})
</script>

<template>
  <main class="admin-page">
    <header class="admin-page__hero">
      <div>
        <p class="admin-page__kicker">Painel administrativo</p>
        <h1>Olá, {{ auth.userName }}</h1>
        <p class="admin-page__lead">
          Acompanhe a operação em tempo real com indicadores consolidados. Sessão ativa como
          {{ adminRoleLabel(auth.role) }}.
        </p>
      </div>
      <div class="hero-status" :class="storeStatusTone">
        <span class="hero-status__label">Status efetivo</span>
        <strong class="hero-status__value">{{ storeStatusLabel }}</strong>
        <span class="hero-status__meta">Prazo: {{ deliveryTimeLabel }}</span>
      </div>
    </header>

    <section class="admin-page__section" aria-labelledby="admin-overview-title">
      <div class="section-head">
        <div>
          <p class="section-head__eyebrow">Resumo</p>
          <h2 id="admin-overview-title">Indicadores principais</h2>
        </div>
        <p class="section-head__hint">Dados lidos da API. O travessão indica indisponibilidade.</p>
      </div>

      <dl class="kpi-grid">
        <div class="kpi-card kpi-card--accent" data-testid="metric-store-status">
          <dt>Status da loja</dt>
          <dd :class="storeStatusLabel === 'Aberta' ? 'is-open' : undefined">
            {{ storeStatusLabel }}
          </dd>
          <span class="kpi-card__meta">Interruptor manual: {{ manualSwitchLabel }}</span>
        </div>
        <div class="kpi-card" data-testid="metric-queued-orders">
          <dt>Pedidos em andamento</dt>
          <dd>{{ queuedOrdersLabel }}</dd>
          <span class="kpi-card__meta">Aguardando operação</span>
        </div>
        <div class="kpi-card">
          <dt>Total de pedidos</dt>
          <dd data-testid="metric-total-orders">{{ totalOrdersLabel }}</dd>
          <span class="kpi-card__meta">Todos os status conhecidos</span>
        </div>
        <div class="kpi-card" data-testid="metric-active-products">
          <dt>Produtos disponíveis / total</dt>
          <dd>{{ productsLabel }}</dd>
          <span class="kpi-card__meta">Catálogo ativo</span>
        </div>
        <div class="kpi-card" data-testid="metric-active-zones">
          <dt>Zonas ativas</dt>
          <dd>{{ activeZonesLabel }}</dd>
          <span class="kpi-card__meta">Áreas de entrega</span>
        </div>
        <div class="kpi-card" data-testid="metric-delivery-time">
          <dt>Prazo de entrega</dt>
          <dd>{{ deliveryTimeLabel }}</dd>
          <span class="kpi-card__meta">Informado pela loja</span>
        </div>
      </dl>
    </section>

    <div class="dashboard-panels">
      <section class="panel" aria-labelledby="admin-status-chart-title">
        <div class="section-head">
          <div>
            <p class="section-head__eyebrow">Acompanhamento</p>
            <h2 id="admin-status-chart-title">Pedidos por status</h2>
          </div>
          <span class="panel__badge">{{ totalOrdersLabel }} no total</span>
        </div>

        <div
          v-if="hasStatusChartData"
          class="status-chart"
          role="img"
          aria-label="Gráfico de pedidos por status"
        >
          <div v-for="item in statusChart" :key="item.status" class="status-chart__row">
            <div class="status-chart__label">
              <span>{{ item.label }}</span>
              <strong>{{ formatCount(item.count) }}</strong>
            </div>
            <div class="status-chart__track" aria-hidden="true">
              <div
                class="status-chart__fill"
                :data-status="item.status"
                :style="{ width: `${item.share}%` }"
              />
            </div>
          </div>
        </div>
        <p v-else class="admin-page__hint">Contagens por status indisponíveis no momento.</p>

        <dl class="status-grid" aria-label="Contagens por status">
          <div
            v-for="item in statusBreakdown"
            :key="item.status"
            class="status-grid__item"
            :data-testid="`metric-status-${item.status}`"
          >
            <dt>{{ item.label }}</dt>
            <dd>{{ item.count }}</dd>
          </div>
        </dl>
      </section>

      <section class="panel" aria-labelledby="admin-operations-title">
        <div class="section-head">
          <div>
            <p class="section-head__eyebrow">Operação</p>
            <h2 id="admin-operations-title">Controle rápido</h2>
          </div>
        </div>

        <div class="kill-switch" data-testid="kill-switch">
          <div class="kill-switch__top">
            <div>
              <p class="kill-switch__eyebrow">Interruptor manual</p>
              <p class="kill-switch__status">
                <span class="status-dot" :class="storeStatusTone" aria-hidden="true"></span>
                {{ storeStatusLabel }}
              </p>
            </div>
            <span class="kill-switch__manual">Manual: {{ manualSwitchLabel }}</span>
          </div>

          <p class="admin-page__hint">
            A alteração passa pelo servidor, gera auditoria e só vale quando o horário permitir.
          </p>

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

          <p v-if="settingsStore.saveSuccess" class="admin-page__feedback" role="status">
            {{ settingsStore.saveSuccess }}
          </p>
          <p
            v-else-if="settingsStore.saveError"
            class="admin-page__feedback admin-page__feedback--error"
            role="alert"
          >
            {{ settingsStore.saveError }}
          </p>
        </div>

        <div class="fulfillment-panel" aria-labelledby="admin-fulfillment-title">
          <h3 id="admin-fulfillment-title">Distribuição por entrega</h3>
          <div class="fulfillment-bar" aria-hidden="true">
            <div
              class="fulfillment-bar__delivery"
              :style="{ width: `${fulfillmentDeliveryShare}%` }"
            ></div>
          </div>
          <dl class="fulfillment-legend">
            <div>
              <dt>Entrega</dt>
              <dd>{{ formatCount(fulfillmentCounts.delivery) }}</dd>
            </div>
            <div>
              <dt>Retirada</dt>
              <dd>{{ formatCount(fulfillmentCounts.pickup) }}</dd>
            </div>
          </dl>
          <p class="admin-page__hint">
            {{
              fulfillmentTotal === null
                ? 'Distribuição indisponível no momento.'
                : `${fulfillmentDeliveryShare}% entrega · ${100 - fulfillmentDeliveryShare}% retirada`
            }}
          </p>
        </div>

        <div class="shortcuts-panel" aria-labelledby="admin-shortcuts-title">
          <h3 id="admin-shortcuts-title">Atalhos</h3>
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
        </div>
      </section>
    </div>

    <section class="panel" aria-labelledby="admin-recent-orders-title">
      <div class="section-head">
        <div>
          <p class="section-head__eyebrow">Fila recente</p>
          <h2 id="admin-recent-orders-title">Pedidos recentes</h2>
        </div>
        <RouterLink class="panel__link" :to="{ name: 'admin-orders' }"
          >Ver todos os pedidos</RouterLink
        >
      </div>

      <p v-if="recentOrders.length === 0" class="admin-page__hint">
        Nenhum pedido recente disponível no momento.
      </p>
      <ol v-else class="recent-orders">
        <li v-for="order in recentOrders" :key="order.id" class="recent-orders__item">
          <div class="recent-orders__main">
            <RouterLink
              class="recent-orders__link"
              :to="{ name: 'admin-order-detail', params: { id: order.id } }"
            >
              #{{ order.orderNumber }}
            </RouterLink>
            <span class="recent-orders__meta">{{ FULFILLMENT_LABELS[order.fulfillment] }}</span>
          </div>
          <OrderStatusBadge :status="order.status" />
          <time :datetime="order.createdAt">{{ formatOrderDateTime(order.createdAt) }}</time>
        </li>
      </ol>
    </section>
  </main>
</template>

<style scoped>
.admin-page {
  display: grid;
  gap: 20px;
}

.admin-page__hero {
  display: grid;
  gap: 16px;
  padding: 20px;
  border: 1px solid var(--bg-border);
  border-radius: 20px;
  background:
    linear-gradient(
      135deg,
      color-mix(in srgb, var(--brand-primary) 14%, transparent),
      transparent 55%
    ),
    var(--bg-surface);
}

.admin-page__kicker,
.section-head__eyebrow,
.kill-switch__eyebrow {
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
  font-size: clamp(32px, 8vw, 48px);
  line-height: var(--line-height-display);
  letter-spacing: 0.02em;
}

.admin-page__lead {
  max-width: 62ch;
  margin: 0;
  color: var(--text-sub);
  line-height: 1.6;
}

.hero-status {
  display: grid;
  gap: 4px;
  align-content: center;
  padding: 16px;
  border: 1px solid var(--bg-border);
  border-radius: 16px;
  background: color-mix(in srgb, var(--bg-input) 88%, transparent);
}

.hero-status.is-open {
  border-color: color-mix(in srgb, var(--status-open) 45%, var(--bg-border));
}

.hero-status__label,
.hero-status__meta,
.section-head__hint,
.admin-page__hint,
.kpi-card__meta,
.panel__badge,
.recent-orders__meta {
  color: var(--text-sub);
  font-size: var(--font-size-caption);
}

.hero-status__value {
  font-family: var(--font-display);
  font-size: var(--font-size-h1);
  line-height: 1.1;
}

.hero-status.is-open .hero-status__value,
.kpi-card dd.is-open,
.status-dot.is-open {
  color: var(--status-open);
}

.admin-page__section,
.panel {
  display: grid;
  gap: 16px;
}

.section-head {
  display: flex;
  flex-wrap: wrap;
  align-items: end;
  justify-content: space-between;
  gap: 10px 16px;
}

.section-head h2,
.shortcuts-panel h3,
.fulfillment-panel h3 {
  margin: 4px 0 0;
  font-size: var(--font-size-h2);
}

.section-head__hint {
  margin: 0;
}

.kpi-grid {
  display: grid;
  gap: 12px;
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
  margin: 0;
}

.kpi-card,
.panel,
.admin-page__card {
  border: 1px solid var(--bg-border);
  border-radius: 16px;
  background: var(--bg-surface);
}

.kpi-card {
  display: grid;
  gap: 6px;
  padding: 16px;
}

.kpi-card--accent {
  background: linear-gradient(
    180deg,
    color-mix(in srgb, var(--brand-primary) 10%, var(--bg-surface)),
    var(--bg-surface)
  );
}

.kpi-card dt {
  color: var(--text-sub);
  font-size: var(--font-size-caption);
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.kpi-card dd {
  margin: 0;
  font-family: var(--font-display);
  font-size: clamp(24px, 5vw, 34px);
  line-height: 1.1;
  font-variant-numeric: tabular-nums;
}

.dashboard-panels {
  display: grid;
  gap: 16px;
}

.panel {
  padding: 16px;
}

.panel__badge {
  padding: 6px 10px;
  border: 1px solid var(--bg-border);
  border-radius: 999px;
  font-weight: 700;
}

.panel__link {
  display: inline-flex;
  min-height: 44px;
  align-items: center;
  color: var(--brand-primary);
  font-size: var(--font-size-caption);
  font-weight: 800;
  text-decoration: none;
}

.status-chart {
  display: grid;
  gap: 10px;
}

.status-chart__row {
  display: grid;
  gap: 6px;
}

.status-chart__label {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  font-size: var(--font-size-caption);
}

.status-chart__label span {
  color: var(--text-sub);
}

.status-chart__label strong {
  font-variant-numeric: tabular-nums;
}

.status-chart__track {
  height: 10px;
  overflow: hidden;
  border-radius: 999px;
  background: var(--bg-input);
}

.status-chart__fill {
  height: 100%;
  min-width: 0;
  border-radius: inherit;
  background: var(--brand-primary);
  transition: width 180ms ease;
}

.status-chart__fill[data-status='pending'] {
  background: var(--text-sub);
}

.status-chart__fill[data-status='confirmed'],
.status-chart__fill[data-status='preparing'] {
  background: var(--brand-primary);
}

.status-chart__fill[data-status='ready'],
.status-chart__fill[data-status='out_for_delivery'] {
  background: var(--brand-neon);
}

.status-chart__fill[data-status='delivered'],
.status-chart__fill[data-status='picked_up'] {
  background: var(--status-open);
}

.status-chart__fill[data-status='cancelled'] {
  background: var(--status-error);
}

.status-grid {
  display: grid;
  gap: 10px;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  margin: 0;
}

.status-grid__item {
  padding: 12px;
  border: 1px solid var(--bg-border);
  border-radius: 12px;
  background: var(--bg-input);
}

.status-grid__item dt {
  color: var(--text-sub);
  font-size: var(--font-size-caption);
}

.status-grid__item dd {
  margin: 4px 0 0;
  font-size: var(--font-size-h2);
  font-weight: 800;
  font-variant-numeric: tabular-nums;
}

.kill-switch,
.fulfillment-panel,
.shortcuts-panel {
  display: grid;
  gap: 12px;
}

.kill-switch__top {
  display: flex;
  flex-wrap: wrap;
  align-items: start;
  justify-content: space-between;
  gap: 12px;
}

.kill-switch__status {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 6px 0 0;
  font-size: var(--font-size-h2);
  font-weight: 800;
}

.status-dot {
  width: 10px;
  height: 10px;
  border-radius: 999px;
  background: var(--text-sub);
}

.status-dot.is-open {
  background: var(--status-open);
}

.kill-switch__manual {
  padding: 6px 10px;
  border: 1px solid var(--bg-border);
  border-radius: 999px;
  color: var(--text-sub);
  font-size: var(--font-size-caption);
  font-weight: 700;
}

.kill-switch__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.admin-page__hint,
.admin-page__feedback {
  margin: 0;
  font-size: var(--font-size-caption);
}

.admin-page__feedback {
  padding: 10px 12px;
  border: 1px solid var(--bg-border);
  border-radius: 12px;
  color: var(--text-main);
}

.admin-page__feedback--error {
  color: var(--status-error);
}

.fulfillment-bar {
  height: 12px;
  overflow: hidden;
  border-radius: 999px;
  background: var(--bg-input);
}

.fulfillment-bar__delivery {
  height: 100%;
  background: linear-gradient(90deg, var(--brand-primary), var(--brand-neon));
}

.fulfillment-legend {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
  margin: 0;
}

.fulfillment-legend dt {
  color: var(--text-sub);
  font-size: var(--font-size-caption);
}

.fulfillment-legend dd {
  margin: 2px 0 0;
  font-weight: 800;
  font-variant-numeric: tabular-nums;
}

.admin-page__grid {
  display: grid;
  gap: 12px;
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
}

.admin-page__card {
  display: flex;
  min-height: 72px;
  align-items: center;
  padding: 16px;
  color: var(--text-main);
  font-weight: 700;
  text-decoration: none;
}

.admin-page__card:hover,
.panel__link:hover,
.recent-orders__link:hover {
  border-color: var(--brand-primary);
  color: var(--brand-primary);
}

.recent-orders {
  display: grid;
  gap: 10px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.recent-orders__item {
  display: grid;
  gap: 10px;
  align-items: center;
  padding: 14px 16px;
  border: 1px solid var(--bg-border);
  border-radius: 14px;
  background: var(--bg-input);
}

.recent-orders__main {
  display: grid;
  gap: 2px;
}

.recent-orders__link {
  display: inline-flex;
  width: fit-content;
  min-height: 44px;
  align-items: center;
  color: var(--text-main);
  font-size: var(--font-size-h2);
  font-weight: 800;
  text-decoration: none;
}

.recent-orders__item time {
  color: var(--text-sub);
  font-size: var(--font-size-caption);
}

@media (min-width: 768px) {
  .admin-page__hero {
    grid-template-columns: minmax(0, 1fr) auto;
    align-items: stretch;
  }

  .dashboard-panels {
    grid-template-columns: minmax(0, 1.45fr) minmax(0, 1fr);
  }

  .recent-orders__item {
    grid-template-columns: minmax(0, 1fr) auto auto;
  }

  .recent-orders__item time {
    justify-self: end;
  }
}

@media (min-width: 1120px) {
  .status-grid {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
}

@media (prefers-reduced-motion: reduce) {
  .status-chart__fill {
    transition: none;
  }
}
</style>
