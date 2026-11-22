import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import { adminAuthErrorMessage } from '../features/admin/auth'
import { fetchAdminUser, loginAdmin, logoutAdmin } from '../services/adminAuth'
import { ApiError } from '../services/http'
import type { AdminRole, AdminUser } from '../types/admin'

export const useAdminAuthStore = defineStore('adminAuth', () => {
  const user = ref<AdminUser | null>(null)
  const loading = ref(false)
  const initialized = ref(false)
  const error = ref('')

  let bootstrapPromise: Promise<void> | null = null

  const isAuthenticated = computed(() => user.value !== null)
  const role = computed<AdminRole | null>(() => user.value?.role ?? null)
  const isOwner = computed(() => role.value === 'owner')
  const userName = computed(() => user.value?.name ?? '')

  function clear(): void {
    user.value = null
    error.value = ''
    initialized.value = true
    bootstrapPromise = null
  }

  async function bootstrap(force = false): Promise<void> {
    if (initialized.value && !force) {
      return
    }
    if (bootstrapPromise) {
      return bootstrapPromise
    }

    loading.value = true
    error.value = ''

    const pending = (async () => {
      try {
        user.value = await fetchAdminUser()
      } catch (cause: unknown) {
        user.value = null
        // Sessão ausente/expirada é estado normal, não erro de carregamento.
        if (!(cause instanceof ApiError && cause.status === 401)) {
          error.value = 'Não foi possível verificar a sessão. Tente novamente.'
        }
      } finally {
        initialized.value = true
        loading.value = false
        bootstrapPromise = null
      }
    })()

    bootstrapPromise = pending
    return pending
  }

  async function login(email: string, password: string): Promise<boolean> {
    loading.value = true
    error.value = ''

    try {
      user.value = await loginAdmin(email, password)
      initialized.value = true
      return true
    } catch (cause: unknown) {
      user.value = null
      error.value = adminAuthErrorMessage(cause)
      return false
    } finally {
      loading.value = false
    }
  }

  async function logout(): Promise<void> {
    loading.value = true
    error.value = ''

    try {
      await logoutAdmin()
    } catch {
      // A sessão local é encerrada mesmo se a API estiver indisponível.
    } finally {
      user.value = null
      initialized.value = true
      loading.value = false
    }
  }

  return {
    user,
    loading,
    initialized,
    error,
    isAuthenticated,
    role,
    isOwner,
    userName,
    clear,
    bootstrap,
    login,
    logout,
  }
})
