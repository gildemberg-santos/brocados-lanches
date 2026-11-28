import { defineStore } from 'pinia'
import { ref } from 'vue'

import { adminCatalogErrorMessage } from '../features/admin-catalog/errors'
import { fetchAdminPix, updateAdminPix } from '../services/adminPix'
import type { AdminPixInput, AdminPixSettings } from '../types/adminPix'

export type AdminPixState = 'idle' | 'loading' | 'success' | 'error'

export const useAdminPixStore = defineStore('adminPix', () => {
  const state = ref<AdminPixState>('idle')
  const settings = ref<AdminPixSettings | null>(null)
  const error = ref('')
  const saving = ref(false)
  const saveError = ref('')
  const saveSuccess = ref('')

  async function load(): Promise<void> {
    state.value = 'loading'
    error.value = ''
    try {
      settings.value = await fetchAdminPix()
      state.value = 'success'
    } catch (cause: unknown) {
      settings.value = null
      error.value = adminCatalogErrorMessage(cause, 'load')
      state.value = 'error'
    }
  }

  async function save(input: AdminPixInput): Promise<boolean> {
    saving.value = true
    saveError.value = ''
    saveSuccess.value = ''
    try {
      settings.value = await updateAdminPix(input)
      saveSuccess.value = 'Configuração do Pix salva com sucesso.'
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
