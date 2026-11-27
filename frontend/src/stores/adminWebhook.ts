import { defineStore } from 'pinia'
import { ref } from 'vue'

import { adminCatalogErrorMessage } from '../features/admin-catalog/errors'
import { fetchAdminWebhook, updateAdminWebhook } from '../services/adminWebhook'
import type { AdminWebhookInput, AdminWebhookSettings } from '../types/adminWebhook'

export type AdminWebhookState = 'idle' | 'loading' | 'success' | 'error'

export const useAdminWebhookStore = defineStore('adminWebhook', () => {
  const state = ref<AdminWebhookState>('idle')
  const settings = ref<AdminWebhookSettings | null>(null)
  const error = ref('')
  const saving = ref(false)
  const saveError = ref('')
  const saveSuccess = ref('')

  async function load(): Promise<void> {
    state.value = 'loading'
    error.value = ''
    try {
      settings.value = await fetchAdminWebhook()
      state.value = 'success'
    } catch (cause: unknown) {
      settings.value = null
      error.value = adminCatalogErrorMessage(cause, 'load')
      state.value = 'error'
    }
  }

  async function save(input: AdminWebhookInput): Promise<boolean> {
    saving.value = true
    saveError.value = ''
    saveSuccess.value = ''
    try {
      settings.value = await updateAdminWebhook(input)
      saveSuccess.value = 'Integração salva com sucesso.'
      return true
    } catch (cause: unknown) {
      saveError.value = adminCatalogErrorMessage(cause, 'save')
      return false
    } finally {
      saving.value = false
    }
  }

  return { state, settings, error, saving, saveError, saveSuccess, load, save }
})
