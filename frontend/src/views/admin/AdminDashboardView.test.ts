import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { defineComponent, h } from 'vue'
import { createMemoryHistory, createRouter } from 'vue-router'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { fetchAdminProducts } from '../../services/adminCatalog'
import { fetchAdminOrders } from '../../services/adminOrders'
import { fetchDeliveryZones } from '../../services/deliveryZones'
import { fetchStoreStatus } from '../../services/store'
import { useAdminAuthStore } from '../../stores/adminAuth'
import type { PublicStoreStatus } from '../../types/store'
import AdminDashboardView from './AdminDashboardView.vue'

vi.mock('../../services/store', () => ({ fetchStoreStatus: vi.fn() }))
vi.mock('../../services/adminOrders', () => ({ fetchAdminOrders: vi.fn() }))
vi.mock('../../services/deliveryZones', () => ({ fetchDeliveryZones: vi.fn() }))
vi.mock('../../services/adminCatalog', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../../services/adminCatalog')>()
  return { ...actual, fetchAdminProducts: vi.fn() }
})

const mockedFetchStoreStatus = vi.mocked(fetchStoreStatus)
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
  nextOpeningAt: null,
  updatedAt: '2026-11-22T12:00:00Z',
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
      { path: '/admin/produtos', name: 'admin-products', component: StubView },
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
    mockedFetchAdminOrders.mockResolvedValue({
      data: [],
      meta: { page: 1, pageSize: 20, total: 8, totalPages: 1 },
    })
    mockedFetchAdminProducts.mockResolvedValue({
      data: [],
      meta: { page: 1, pageSize: 1, total: 37, totalPages: 37 },
    })
    mockedFetchDeliveryZones.mockResolvedValue([
      { id: 'zone-1', name: 'Centro', feeCents: 800, sortOrder: 1 },
      { id: 'zone-2', name: 'Aquiraz', feeCents: 1000, sortOrder: 2 },
      { id: 'zone-3', name: 'Iguape', feeCents: 1200, sortOrder: 3 },
      { id: 'zone-4', name: 'Eusébio', feeCents: 1500, sortOrder: 4 },
      { id: 'zone-5', name: 'Pacajus', feeCents: 1800, sortOrder: 5 },
    ])
  })

  it('sauda o usuário, mostra atalhos e indicadores vindos da API', async () => {
    const wrapper = await mountDashboard()

    expect(wrapper.get('h1').text()).toBe('Olá, Dono')
    expect(wrapper.text()).toContain('Administrador')
    expect(wrapper.text()).toContain('Pedidos')
    expect(wrapper.text()).toContain('Produtos')
    expect(wrapper.text()).not.toContain('serão disponibilizadas nas próximas entregas')

    expect(metric(wrapper, 'metric-store-status')).toBe('Aberta')
    expect(metric(wrapper, 'metric-queued-orders')).toBe('8')
    expect(metric(wrapper, 'metric-active-products')).toBe('37')
    expect(metric(wrapper, 'metric-active-zones')).toBe('5')
    expect(mockedFetchAdminProducts).toHaveBeenCalledWith({ active: true, pageSize: 1 })
  })

  it('mostra a loja fechada quando o servidor informa status fechado', async () => {
    mockedFetchStoreStatus.mockResolvedValue({ ...storeStatus, status: 'closed' })

    const wrapper = await mountDashboard()

    expect(metric(wrapper, 'metric-store-status')).toBe('Fechada')
  })

  it('usa travessão quando as fontes falham ou estão vazias', async () => {
    mockedFetchStoreStatus.mockRejectedValue(new Error('boom'))
    mockedFetchAdminOrders.mockRejectedValue(new Error('boom'))
    mockedFetchAdminProducts.mockRejectedValue(new Error('boom'))
    mockedFetchDeliveryZones.mockResolvedValue([])

    const wrapper = await mountDashboard()

    expect(metric(wrapper, 'metric-store-status')).toBe('—')
    expect(metric(wrapper, 'metric-queued-orders')).toBe('—')
    expect(metric(wrapper, 'metric-active-products')).toBe('—')
    expect(metric(wrapper, 'metric-active-zones')).toBe('—')
    expect(wrapper.text()).not.toContain('boom')
  })

  it('não inventa números quando a API devolve contadores zerados', async () => {
    mockedFetchAdminOrders.mockResolvedValue({
      data: [],
      meta: { page: 1, pageSize: 20, total: 0, totalPages: 0 },
    })
    mockedFetchAdminProducts.mockResolvedValue({
      data: [],
      meta: { page: 1, pageSize: 1, total: 0, totalPages: 0 },
    })
    mockedFetchDeliveryZones.mockResolvedValue([])

    const wrapper = await mountDashboard()

    expect(metric(wrapper, 'metric-store-status')).toBe('Aberta')
    expect(metric(wrapper, 'metric-queued-orders')).toBe('—')
    expect(metric(wrapper, 'metric-active-products')).toBe('—')
    expect(metric(wrapper, 'metric-active-zones')).toBe('—')
  })
})
