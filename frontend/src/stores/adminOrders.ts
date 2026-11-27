import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import { adminOrderErrorMessage } from '../features/admin-orders/orderStatus'
import { fetchAdminOrder, fetchAdminOrders, updateAdminOrderStatus } from '../services/adminOrders'
import type { AdminPageMeta } from '../types/admin'
import type {
  AdminOrderDetail,
  AdminOrderListItem,
  AdminOrderListQuery,
  AdminOrderStatusRequest,
} from '../types/order'

export type AdminOrdersState = 'idle' | 'loading' | 'success' | 'error'
export type AdminOrderDetailState = 'idle' | 'loading' | 'success' | 'error'

const DEFAULT_PAGE_SIZE = 20

const DEFAULT_FILTERS: AdminOrderListQuery = {
  page: 1,
  pageSize: DEFAULT_PAGE_SIZE,
  sort: 'createdAt',
  order: 'desc',
}

const EMPTY_META: AdminPageMeta = {
  page: 1,
  pageSize: DEFAULT_PAGE_SIZE,
  total: 0,
  totalPages: 0,
}

export const useAdminOrdersStore = defineStore('adminOrders', () => {
  const state = ref<AdminOrdersState>('idle')
  const orders = ref<AdminOrderListItem[]>([])
  const meta = ref<AdminPageMeta>({ ...EMPTY_META })
  const filters = ref<AdminOrderListQuery>({ ...DEFAULT_FILTERS })
  const error = ref('')

  const detailState = ref<AdminOrderDetailState>('idle')
  const detail = ref<AdminOrderDetail | null>(null)
  const detailError = ref('')
  const actionLoading = ref(false)
  const actionError = ref('')
  const actionSuccess = ref('')

  const isEmpty = computed(() => state.value === 'success' && orders.value.length === 0)
  const totalPages = computed(() => meta.value.totalPages)

  async function load(): Promise<void> {
    state.value = 'loading'
    error.value = ''
    try {
      const page = await fetchAdminOrders({ ...filters.value })
      orders.value = page.data
      meta.value = page.meta
      state.value = 'success'
    } catch (cause: unknown) {
      orders.value = []
      meta.value = { ...EMPTY_META }
      error.value = adminOrderErrorMessage(cause, 'load')
      state.value = 'error'
    }
  }

  /** Recarrega a fila após um evento sem exibir skeleton nem perder a lista atual. */
  async function refreshFromEvent(): Promise<void> {
    if (state.value !== 'success') return
    try {
      const page = await fetchAdminOrders({ ...filters.value })
      orders.value = page.data
      meta.value = page.meta
    } catch {
      // Mantém a última lista conhecida; a reconexão vai tentar novamente.
    }
  }

  function normalizeFilters(next: Partial<AdminOrderListQuery>): AdminOrderListQuery {
    const result: AdminOrderListQuery = {
      ...DEFAULT_FILTERS,
      ...next,
      page: 1,
    }
    if (!result.status) delete result.status
    if (!result.fulfillment) delete result.fulfillment
    if (!result.q?.trim()) delete result.q
    if (!result.from) delete result.from
    if (!result.to) delete result.to
    return result
  }

  async function applyFilters(next: Partial<AdminOrderListQuery>): Promise<void> {
    filters.value = normalizeFilters(next)
    await load()
  }

  async function goToPage(page: number): Promise<void> {
    if (page < 1 || page > Math.max(totalPages.value, 1)) {
      return
    }
    filters.value = { ...filters.value, page }
    await load()
  }

  async function loadDetail(id: string, force = false): Promise<void> {
    if (!id || (detailState.value === 'success' && !force)) {
      return
    }
    detailState.value = 'loading'
    detailError.value = ''
    actionError.value = ''
    actionSuccess.value = ''
    try {
      detail.value = await fetchAdminOrder(id)
      detailState.value = 'success'
    } catch (cause: unknown) {
      detail.value = null
      detailError.value = adminOrderErrorMessage(cause, 'load')
      detailState.value = 'error'
    }
  }

  /** Atualiza o detalhe aberto quando outro operador muda o status do pedido. */
  async function refreshDetailFromEvent(id: string): Promise<void> {
    if (
      actionLoading.value ||
      detailState.value !== 'success' ||
      detail.value?.id !== id ||
      !detail.value
    ) {
      return
    }
    try {
      const next = await fetchAdminOrder(id)
      if (next.status !== detail.value.status || next.updatedAt !== detail.value.updatedAt) {
        detail.value = next
      }
    } catch {
      // Mantém o detalhe atual em caso de falha transitória.
    }
  }

  function resetDetail(): void {
    detailState.value = 'idle'
    detail.value = null
    detailError.value = ''
    actionError.value = ''
    actionSuccess.value = ''
  }

  async function changeStatus(id: string, request: AdminOrderStatusRequest): Promise<boolean> {
    actionLoading.value = true
    actionError.value = ''
    actionSuccess.value = ''
    try {
      detail.value = await updateAdminOrderStatus(id, request)
      const index = orders.value.findIndex((item) => item.id === id)
      if (index >= 0) {
        orders.value[index] = {
          ...orders.value[index],
          status: detail.value.status,
          updatedAt: detail.value.updatedAt,
        }
      }
      actionSuccess.value = 'Status do pedido atualizado.'
      return true
    } catch (cause: unknown) {
      actionError.value = adminOrderErrorMessage(cause, 'status')
      return false
    } finally {
      actionLoading.value = false
    }
  }

  return {
    state,
    orders,
    meta,
    filters,
    error,
    isEmpty,
    totalPages,
    detailState,
    detail,
    detailError,
    actionLoading,
    actionError,
    actionSuccess,
    load,
    refreshFromEvent,
    applyFilters,
    goToPage,
    loadDetail,
    refreshDetailFromEvent,
    resetDetail,
    changeStatus,
  }
})
