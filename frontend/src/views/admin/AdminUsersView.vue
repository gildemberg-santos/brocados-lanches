<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

import AdminUserForm from '../../components/admin/AdminUserForm.vue'
import AdminUserList from '../../components/admin/AdminUserList.vue'
import BaseButton from '../../components/BaseButton.vue'
import BaseModal from '../../components/BaseModal.vue'
import { useAdminAuthStore } from '../../stores/adminAuth'
import { useAdminUsersStore } from '../../stores/adminUsers'
import type { AdminUser, AdminUserCreateInput } from '../../types/adminUsers'

const store = useAdminUsersStore()
const auth = useAdminAuthStore()
const editing = ref<AdminUser | null>(null)
const showForm = ref(false)
const search = ref('')

const modalTitle = computed(() => (editing.value ? 'Editar usuário' : 'Novo usuário'))

function startNew(): void {
  editing.value = null
  showForm.value = true
  store.clearMessages()
}

function edit(user: AdminUser): void {
  editing.value = user
  showForm.value = true
  store.clearMessages()
}

function cancel(): void {
  showForm.value = false
  editing.value = null
  store.clearMessages()
}

async function submit(input: AdminUserCreateInput): Promise<void> {
  if (await store.save(input, editing.value?.id)) {
    cancel()
  }
}

async function toggle(user: AdminUser): Promise<void> {
  await store.toggle(user.id, !user.active)
}

function applySearch(): void {
  void store.applyFilters({ q: search.value.trim() })
}

onMounted(() => {
  if (store.state === 'idle') {
    void store.load()
  }
})
</script>

<template>
  <main class="admin-page">
    <p class="admin-page__kicker">Painel administrativo</p>
    <h1>Usuários</h1>
    <p class="admin-page__lead">
      Somente donos podem criar e editar contas do painel. A senha temporária deve ter pelo menos 12
      caracteres.
    </p>

    <BaseButton @click="startNew">Novo usuário</BaseButton>

    <form class="admin-page__search" @submit.prevent="applySearch">
      <label for="admin-users-search">Buscar por nome ou e-mail</label>
      <div>
        <input id="admin-users-search" v-model="search" type="search" maxlength="120" />
        <button type="submit">Buscar</button>
      </div>
    </form>

    <BaseModal :open="showForm" :title="modalTitle" @close="cancel">
      <AdminUserForm
        :item="editing"
        :saving="store.saving"
        :error="store.saveError"
        :success="store.saveSuccess"
        @submit="submit"
        @cancel="cancel"
      />
    </BaseModal>

    <AdminUserList
      :items="store.items"
      :state="store.state"
      :error="store.error"
      :isEmpty="store.isEmpty"
      :meta="store.meta"
      :actor-id="auth.user?.id ?? ''"
      @edit="edit"
      @toggle="toggle"
      @page="store.goToPage"
      @retry="store.load"
    />
  </main>
</template>

<style scoped>
.admin-page {
  display: grid;
  gap: 16px;
}

.admin-page__kicker {
  margin: 0;
  color: var(--brand-neon);
  font-size: var(--font-size-caption);
  font-weight: 800;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.admin-page h1 {
  margin: 8px 0;
  font-family: var(--font-display);
  font-size: var(--font-size-display);
  line-height: var(--line-height-display);
}

.admin-page__lead {
  margin: 0;
  color: var(--text-sub);
}

.admin-page__search {
  display: grid;
  gap: 8px;
}

.admin-page__search label {
  font-size: var(--font-size-caption);
  font-weight: 700;
}

.admin-page__search div {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.admin-page__search input {
  flex: 1 1 240px;
  min-height: 44px;
  padding: 10px 12px;
  border: 1px solid var(--bg-border);
  border-radius: 12px;
  background: var(--bg-input);
  color: var(--text-main);
  font-family: var(--font-sans);
  font-size: var(--font-size-body);
}

.admin-page__search button {
  min-height: 44px;
  padding: 10px 16px;
  border: 1px solid var(--brand-primary);
  border-radius: 12px;
  background: var(--brand-primary);
  color: var(--bg-primary);
  font-weight: 800;
}
</style>
