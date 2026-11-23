import { defineStore } from 'pinia'
import { ref } from 'vue'

import { adminCatalogErrorMessage } from '../features/admin-catalog/errors'
import { fetchAdminStore, updateAdminStore } from '../services/adminStore'
import type { AdminStoreInput, AdminStoreSettings } from '../types/adminCatalog'

export type AdminStoreSettingsState = 'idle' | 'loading' | 'success' | 'error'

export const useAdminStoreSettingsStore = defineStore('adminStoreSettings', () => {
  const state = ref<AdminStoreSettingsState>('idle')
  const settings = ref<AdminStoreSettings | null>(null)
  const error = ref('')
  const saving = ref(false)
  const saveError = ref('')
  const saveSuccess = ref('')

  async function load(): Promise<void> {
    state.value = 'loading'
    error.value = ''
    try {
      settings.value = await fetchAdminStore()
      state.value = 'success'
    } catch (cause: unknown) {
      settings.value = null
      error.value = adminCatalogErrorMessage(cause, 'load')
      state.value = 'error'
    }
  }

  function clearMessages(): void {
    saveError.value = ''
    saveSuccess.value = ''
  }

  async function save(input: AdminStoreInput): Promise<boolean> {
    saving.value = true
    clearMessages()
    try {
      settings.value = await updateAdminStore(input)
      saveSuccess.value = 'Configurações salvas.'
      return true
    } catch (cause: unknown) {
      saveError.value = adminCatalogErrorMessage(cause, 'save')
      return false
    } finally {
      saving.value = false
    }
  }

  return { state, settings, error, saving, saveError, saveSuccess, load, save, clearMessages }
})
