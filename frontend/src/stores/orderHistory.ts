import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import {
  DEFAULT_ORDER_HISTORY_PAGE_SIZE,
  EMPTY_ORDER_HISTORY_META,
  orderHistoryErrorMessage,
} from '../features/orders/history'
import { loadOrderAccessToken } from '../features/orders/orderAccessToken'
import { fetchOrderHistory } from '../services/orders'
import type { AdminPageMeta } from '../types/admin'
import type { TrackedOrder } from '../types/order'

export type OrderHistoryState = 'idle' | 'loading' | 'success' | 'error'

export const useOrderHistoryStore = defineStore('orderHistory', () => {
  const state = ref<OrderHistoryState>('idle')
  const orders = ref<TrackedOrder[]>([])
  const meta = ref<AdminPageMeta>({ ...EMPTY_ORDER_HISTORY_META })
  const error = ref('')
  const hasToken = ref(false)

  const isEmpty = computed(() => state.value === 'success' && orders.value.length === 0)
  const totalPages = computed(() => meta.value.totalPages)

  async function load(): Promise<void> {
    const accessToken = loadOrderAccessToken()
    hasToken.value = accessToken !== null
    if (!accessToken) {
      orders.value = []
      meta.value = { ...EMPTY_ORDER_HISTORY_META }
      error.value = ''
      state.value = 'success'
      return
    }

    state.value = 'loading'
    error.value = ''
    try {
      const page = await fetchOrderHistory(accessToken, {
        page: meta.value.page,
        pageSize: DEFAULT_ORDER_HISTORY_PAGE_SIZE,
      })
      orders.value = page.data
      meta.value = page.meta
      state.value = 'success'
    } catch (cause: unknown) {
      orders.value = []
      meta.value = { ...EMPTY_ORDER_HISTORY_META }
      error.value = orderHistoryErrorMessage(cause)
      state.value = 'error'
    }
  }

  async function goToPage(page: number): Promise<void> {
    if (page < 1 || page > Math.max(totalPages.value, 1)) {
      return
    }
    meta.value = { ...meta.value, page }
    await load()
  }

  return {
    state,
    orders,
    meta,
    error,
    hasToken,
    isEmpty,
    totalPages,
    load,
    goToPage,
  }
})
