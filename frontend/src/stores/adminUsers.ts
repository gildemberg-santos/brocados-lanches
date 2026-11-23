import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import { adminCatalogErrorMessage } from '../features/admin-catalog/errors'
import {
  createAdminUser,
  fetchAdminAuditLogs,
  fetchAdminUsers,
  updateAdminUser,
} from '../services/adminUsers'
import type { AdminPageMeta } from '../types/admin'
import type {
  AdminAuditListQuery,
  AdminAuditLog,
  AdminUser,
  AdminUserCreateInput,
  AdminUserUpdateInput,
} from '../types/adminUsers'

export type AdminUsersState = 'idle' | 'loading' | 'success' | 'error'
export type AdminAuditState = 'idle' | 'loading' | 'success' | 'error'

const EMPTY_META: AdminPageMeta = { page: 1, pageSize: 20, total: 0, totalPages: 0 }

export const useAdminUsersStore = defineStore('adminUsers', () => {
  const state = ref<AdminUsersState>('idle')
  const items = ref<AdminUser[]>([])
  const meta = ref<AdminPageMeta>({ ...EMPTY_META })
  const filters = ref<AdminAuditListQuery>({ page: 1, pageSize: 20, sort: 'name', order: 'asc' })
  const error = ref('')
  const saving = ref(false)
  const saveError = ref('')
  const saveSuccess = ref('')

  const auditState = ref<AdminAuditState>('idle')
  const audits = ref<AdminAuditLog[]>([])
  const auditMeta = ref<AdminPageMeta>({ ...EMPTY_META })
  const auditFilters = ref<AdminAuditListQuery>({
    page: 1,
    pageSize: 20,
    sort: 'createdAt',
    order: 'desc',
  })
  const auditError = ref('')

  const isEmpty = computed(() => state.value === 'success' && items.value.length === 0)
  const auditIsEmpty = computed(() => auditState.value === 'success' && audits.value.length === 0)

  async function load(): Promise<void> {
    state.value = 'loading'
    error.value = ''
    try {
      const page = await fetchAdminUsers({ ...filters.value })
      items.value = page.data
      meta.value = page.meta
      state.value = 'success'
    } catch (cause: unknown) {
      items.value = []
      meta.value = { ...EMPTY_META }
      error.value = adminCatalogErrorMessage(cause, 'load')
      state.value = 'error'
    }
  }

  async function applyFilters(next: Partial<AdminAuditListQuery>): Promise<void> {
    filters.value = { ...filters.value, ...next, page: 1 }
    await load()
  }

  async function goToPage(page: number): Promise<void> {
    filters.value = { ...filters.value, page }
    await load()
  }

  function clearMessages(): void {
    saveError.value = ''
    saveSuccess.value = ''
  }

  async function save(input: AdminUserCreateInput, id?: string): Promise<boolean> {
    saving.value = true
    clearMessages()
    try {
      if (id) {
        await updateAdminUser(id, input as AdminUserUpdateInput)
        saveSuccess.value = 'Usuário atualizado.'
      } else {
        await createAdminUser(input)
        saveSuccess.value = 'Usuário criado com senha temporária.'
      }
      await load()
      return true
    } catch (cause: unknown) {
      saveError.value = adminCatalogErrorMessage(cause, 'save')
      return false
    } finally {
      saving.value = false
    }
  }

  async function toggle(id: string, active: boolean): Promise<void> {
    clearMessages()
    try {
      await updateAdminUser(id, { active })
      await load()
    } catch (cause: unknown) {
      saveError.value = adminCatalogErrorMessage(cause, 'save')
    }
  }

  async function loadAudit(): Promise<void> {
    auditState.value = 'loading'
    auditError.value = ''
    try {
      const page = await fetchAdminAuditLogs({ ...auditFilters.value })
      audits.value = page.data
      auditMeta.value = page.meta
      auditState.value = 'success'
    } catch (cause: unknown) {
      audits.value = []
      auditMeta.value = { ...EMPTY_META }
      auditError.value = adminCatalogErrorMessage(cause, 'load')
      auditState.value = 'error'
    }
  }

  async function applyAuditFilters(next: Partial<AdminAuditListQuery>): Promise<void> {
    auditFilters.value = { ...auditFilters.value, ...next, page: 1 }
    await loadAudit()
  }

  async function auditGoToPage(page: number): Promise<void> {
    auditFilters.value = { ...auditFilters.value, page }
    await loadAudit()
  }

  return {
    state,
    items,
    meta,
    filters,
    error,
    isEmpty,
    saving,
    saveError,
    saveSuccess,
    load,
    applyFilters,
    goToPage,
    clearMessages,
    save,
    toggle,
    auditState,
    audits,
    auditMeta,
    auditFilters,
    auditError,
    auditIsEmpty,
    loadAudit,
    applyAuditFilters,
    auditGoToPage,
  }
})
