import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { defineComponent, h } from 'vue'
import { createMemoryHistory, createRouter } from 'vue-router'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { fetchAdminProducts } from '../../services/adminCatalog'
import { fetchAdminOrders } from '../../services/adminOrders'
import { fetchAdminStore, updateAdminStore } from '../../services/adminStore'
import { fetchDeliveryZones } from '../../services/deliveryZones'
import { fetchStoreStatus } from '../../services/store'
import { useAdminAuthStore } from '../../stores/adminAuth'
import type { AdminProduct, AdminStoreSettings } from '../../types/adminCatalog'
import type { AdminOrderListItem } from '../../types/order'
import type { PublicStoreStatus } from '../../types/store'
import AdminDashboardView from './AdminDashboardView.vue'

vi.mock('../../services/store', () => ({ fetchStoreStatus: vi.fn() }))
vi.mock('../../services/adminStore', () => ({
  fetchAdminStore: vi.fn(),
  updateAdminStore: vi.fn(),
}))
vi.mock('../../services/adminOrders', () => ({ fetchAdminOrders: vi.fn() }))
vi.mock('../../services/deliveryZones', () => ({ fetchDeliveryZones: vi.fn() }))
vi.mock('../../services/adminCatalog', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../../services/adminCatalog')>()
  return { ...actual, fetchAdminProducts: vi.fn() }
})

const mockedFetchStoreStatus = vi.mocked(fetchStoreStatus)
const mockedFetchAdminStore = vi.mocked(fetchAdminStore)
const mockedUpdateAdminStore = vi.mocked(updateAdminStore)
const mockedFetchAdminOrders = vi.mocked(fetchAdminOrders)
const mockedFetchAdminProducts = vi.mocked(fetchAdminProducts)
const mockedFetchDeliveryZones = vi.mocked(fetchDeliveryZones)

const StubView = defineComponent({ render: () => h('div') })

const storeStatus: PublicStoreStatus = {
  name: 'Brocados Lanches',
  status: 'open',
  isOpenManual: true,
  timezone: 'America/Fortaleza',
  deliveryTime: '30-90 min',
  whatsapp: '+5585999999999',
  instagram: '@brocados',
  address: 'Rua de Teste, 10',
  openingHours: [],
  pix: { active: false, key: '', merchantName: '', merchantCity: '' },
  nextOpeningAt: null,
  updatedAt: '2026-11-22T12:00:00Z',
}
const storeSettings: AdminStoreSettings = {
  name: 'Brocados Lanches',
  isOpen: true,
  openingHours: [],
  deliveryTime: '30-90 min',
  whatsapp: '+5585999999999',
  instagram: '@brocados',
  address: 'Rua de Teste, 10',
  createdAt: '2026-11-22T12:00:00Z',
  updatedAt: '2026-11-22T12:00:00Z',
}
const recentOrder: AdminOrderListItem = {
  id: 'order-1',
  orderNumber: 1042,
  publicId: 'public-123',
  status: 'confirmed',
  fulfillment: 'delivery',
  customerName: 'M***',
  phone: '***9999',
  itemCount: 2,
  subtotalCents: 5400,
  deliveryFeeCents: 800,
  totalCents: 6200,
  createdAt: '2026-11-22T02:30:00Z',
  updatedAt: '2026-11-22T03:10:00Z',
}
const totals: Record<string, number> = {
  pending: 2,
  confirmed: 3,
  preparing: 1,
  ready: 0,
  out_for_delivery: 1,
  picked_up: 1,
  delivered: 4,
  cancelled: 1,
}

function makeProduct(index: number, available: boolean): AdminProduct {
  return {
    id: `product-${index}`,
    categoryId: 'category-1',
    name: `Produto ${index}`,
    slug: `produto-${index}`,
    description: '',
    priceCents: 1000,
    imageUrl: '',
    imageAlt: '',
    allergens: [],
    available,
    featured: false,
    active: true,
    sortOrder: index,
    optionGroups: [],
    addonIds: [],
    createdAt: '2026-11-22T12:00:00Z',
    updatedAt: '2026-11-22T12:00:00Z',
  }
}

async function mountDashboard() {
  const pinia = createPinia()
  setActivePinia(pinia)
  const auth = useAdminAuthStore()
  auth.user = {
    id: 'user-1',
    name: 'Dono',
    email: 'dono@example.com',
    role: 'owner',
  }

  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/admin', name: 'admin-dashboard', component: StubView },
      { path: '/admin/pedidos', name: 'admin-orders', component: StubView },
      { path: '/admin/pedidos/:id', name: 'admin-order-detail', component: StubView },
      { path: '/admin/produtos', name: 'admin-products', component: StubView },
      { path: '/admin/loja', name: 'admin-store', component: StubView },
      { path: '/admin/conta', name: 'admin-account', component: StubView },
    ],
  })
  await router.push('/admin')
  await router.isReady()

  const wrapper = mount(AdminDashboardView, {
    global: { plugins: [pinia, router] },
  })
  await flushPromises()

  return wrapper
}

function metric(wrapper: Awaited<ReturnType<typeof mountDashboard>>, testId: string): string {
  return wrapper.get(`[data-testid="${testId}"] dd`).text().trim()
}

describe('AdminDashboardView', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockedFetchStoreStatus.mockResolvedValue(storeStatus)
    mockedFetchAdminStore.mockResolvedValue(storeSettings)
    mockedUpdateAdminStore.mockImplementation(async (input) => ({
      ...storeSettings,
      ...input,
      updatedAt: '2026-11-22T13:00:00Z',
    }))
    mockedFetchAdminOrders.mockImplementation(async (query = {}) => {
      const status = query.status ?? ''
      const fulfillment = query.fulfillment ?? ''
      const fulfillmentTotals: Record<string, number> = { delivery: 8, pickup: 5 }
      const total = status
        ? (totals[status] ?? 0)
        : fulfillment
          ? (fulfillmentTotals[fulfillment] ?? 0)
          : 13
      return {
        data: status || fulfillment ? [] : [recentOrder],
        meta: {
          page: query.page ?? 1,
          pageSize: query.pageSize ?? 20,
          total,
          totalPages: total > 0 ? 1 : 0,
        },
      }
    })
    mockedFetchAdminProducts.mockResolvedValue({
      data: Array.from({ length: 37 }, (_, index) => makeProduct(index + 1, index < 30)),
      meta: { page: 1, pageSize: 100, total: 37, totalPages: 1 },
    })
    mockedFetchDeliveryZones.mockResolvedValue([
      { id: 'zone-1', name: 'Centro', feeCents: 800, sortOrder: 1 },
      { id: 'zone-2', name: 'Aquiraz', feeCents: 1000, sortOrder: 2 },
      { id: 'zone-3', name: 'Iguape', feeCents: 1200, sortOrder: 3 },
      { id: 'zone-4', name: 'Eusébio', feeCents: 1500, sortOrder: 4 },
      { id: 'zone-5', name: 'Pacajus', feeCents: 1800, sortOrder: 5 },
    ])
  })

  it('mostra atalhos, indicadores, gráficos e pedidos recentes sem PII', async () => {
    const wrapper = await mountDashboard()

    expect(wrapper.get('h1').text()).toBe('Olá, Dono')
    expect(wrapper.text()).toContain('Administrador')
    expect(wrapper.text()).toContain('Pedidos')
    expect(wrapper.text()).toContain('Produtos')

    expect(metric(wrapper, 'metric-store-status')).toBe('Aberta')
    expect(metric(wrapper, 'metric-queued-orders')).toBe('7')
    expect(wrapper.get('[data-testid="metric-total-orders"]').text()).toBe('13')
    expect(metric(wrapper, 'metric-active-products')).toBe('30 / 37')
    expect(metric(wrapper, 'metric-active-zones')).toBe('5')
    expect(metric(wrapper, 'metric-delivery-time')).toBe('30-90 min')
    expect(metric(wrapper, 'metric-status-pending')).toBe('2')
    expect(metric(wrapper, 'metric-status-delivered')).toBe('4')

    expect(wrapper.findAll('.status-chart__row')).toHaveLength(8)
    expect(wrapper.text()).toContain('Entrega')
    expect(wrapper.text()).toContain('Retirada')
    expect(wrapper.text()).toContain('Pedidos recentes')
    expect(wrapper.get('.recent-orders__link').text()).toContain('1042')
    expect(wrapper.get('.recent-orders__link').attributes('href')).toContain(
      '/admin/pedidos/order-1',
    )
    expect(wrapper.text()).not.toContain('M***')
    expect(wrapper.text()).not.toContain('***9999')
  })

  it('aciona o kill switch com confirmação e atualiza o status efetivo', async () => {
    const wrapper = await mountDashboard()

    await wrapper.get('[data-testid="kill-switch"] button').trigger('click')
    expect(wrapper.text()).toContain('Confirmar fechamento')

    const buttons = wrapper.findAll('[data-testid="kill-switch"] button')
    const confirm = buttons.find((button) => button.text().includes('Confirmar fechamento'))
    await confirm!.trigger('click')
    await flushPromises()

    expect(mockedUpdateAdminStore).toHaveBeenCalledWith(
      expect.objectContaining({ isOpen: false, name: 'Brocados Lanches' }),
    )
    expect(mockedUpdateAdminStore.mock.calls[0][0]).not.toHaveProperty('createdAt')
    expect(mockedUpdateAdminStore.mock.calls[0][0]).not.toHaveProperty('updatedAt')
    expect(mockedFetchStoreStatus).toHaveBeenCalledTimes(2)
    expect(wrapper.text()).toContain('Configurações salvas.')
  })

  it('mostra o status efetivo fechado informado pelo servidor', async () => {
    mockedFetchStoreStatus.mockResolvedValue({ ...storeStatus, status: 'closed' })

    const wrapper = await mountDashboard()

    expect(metric(wrapper, 'metric-store-status')).toBe('Fechada')
    expect(wrapper.text()).toContain('Controle rápido')
  })

  it('usa travessão quando as fontes falham', async () => {
    mockedFetchStoreStatus.mockRejectedValue(new Error('boom'))
    mockedFetchAdminStore.mockRejectedValue(new Error('boom'))
    mockedFetchAdminOrders.mockRejectedValue(new Error('boom'))
    mockedFetchAdminProducts.mockRejectedValue(new Error('boom'))
    mockedFetchDeliveryZones.mockRejectedValue(new Error('boom'))

    const wrapper = await mountDashboard()

    expect(metric(wrapper, 'metric-store-status')).toBe('—')
    expect(metric(wrapper, 'metric-queued-orders')).toBe('—')
    expect(wrapper.get('[data-testid="metric-total-orders"]').text()).toBe('—')
    expect(metric(wrapper, 'metric-active-products')).toBe('—')
    expect(metric(wrapper, 'metric-active-zones')).toBe('—')
    expect(metric(wrapper, 'metric-delivery-time')).toBe('—')
    expect(metric(wrapper, 'metric-status-pending')).toBe('—')
    expect(wrapper.text()).toContain('Contagens por status indisponíveis')
    expect(wrapper.text()).not.toContain('boom')
  })

  it('exibe zero quando a API devolve contadores zerados', async () => {
    mockedFetchAdminOrders.mockResolvedValue({
      data: [],
      meta: { page: 1, pageSize: 20, total: 0, totalPages: 0 },
    })
    mockedFetchAdminProducts.mockResolvedValue({
      data: [],
      meta: { page: 1, pageSize: 100, total: 0, totalPages: 0 },
    })
    mockedFetchDeliveryZones.mockResolvedValue([])

    const wrapper = await mountDashboard()

    expect(metric(wrapper, 'metric-queued-orders')).toBe('0')
    expect(metric(wrapper, 'metric-active-products')).toBe('0 / 0')
    expect(metric(wrapper, 'metric-active-zones')).toBe('0')
    expect(wrapper.text()).toContain('Nenhum pedido recente disponível')
  })
})
