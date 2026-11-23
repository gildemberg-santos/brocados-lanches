import { defineStore } from 'pinia'
import { ref } from 'vue'

import { adminCatalogErrorMessage } from '../features/admin-catalog/errors'
import { fetchStoreStatus } from '../services/store'
import type { PublicStoreStatus } from '../types/store'

export type StoreStatusState = 'idle' | 'loading' | 'success' | 'error'

export const useStoreStatusStore = defineStore('storeStatus', () => {
  const state = ref<StoreStatusState>('idle')
  const status = ref<PublicStoreStatus | null>(null)
  const error = ref('')

  async function refresh(): Promise<void> {
    state.value = 'loading'
    error.value = ''
    try {
      status.value = await fetchStoreStatus()
      state.value = 'success'
    } catch (cause: unknown) {
      status.value = null
      error.value = adminCatalogErrorMessage(cause, 'load')
      state.value = 'error'
    }
  }

  async function load(): Promise<void> {
    if (state.value === 'idle') {
      await refresh()
    }
  }

  return { state, status, error, load, refresh }
})
