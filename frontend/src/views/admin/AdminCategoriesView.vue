<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

import AdminCatalogFilters from '../../components/admin/AdminCatalogFilters.vue'
import AdminCatalogForm from '../../components/admin/AdminCatalogForm.vue'
import AdminCatalogList from '../../components/admin/AdminCatalogList.vue'
import type { AdminCatalogRow } from '../../components/admin/AdminCatalogList.vue'
import BaseButton from '../../components/BaseButton.vue'
import BaseModal from '../../components/BaseModal.vue'
import { useAdminCategoriesStore } from '../../stores/adminCatalog'
import type {
  AdminCatalogInput,
  AdminCatalogListQuery,
  AdminCategory,
} from '../../types/adminCatalog'

const store = useAdminCategoriesStore()
const editing = ref<AdminCategory | null>(null)
const showForm = ref(false)

const modalTitle = computed(() => (editing.value ? 'Editar categoria' : 'Nova categoria'))

function applyFilters(filters: Partial<AdminCatalogListQuery>): void {
  void store.applyFilters(filters)
}

function goToPage(page: number): void {
  void store.goToPage(page)
}

function startNew(): void {
  editing.value = null
  showForm.value = true
  store.clearMessages()
}

function edit(item: AdminCatalogRow): void {
  editing.value = item as AdminCategory
  showForm.value = true
  store.clearMessages()
}

function cancel(): void {
  showForm.value = false
  editing.value = null
  store.clearMessages()
}

async function submit(input: AdminCatalogInput): Promise<void> {
  if (!('description' in input)) {
    return
  }
  const saved = await store.save(editing.value?.id ?? null, input)
  if (saved) {
    cancel()
  }
}

async function toggle(item: AdminCatalogRow): Promise<void> {
  await store.save(item.id, { active: !item.active })
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
    <h1>Categorias</h1>
    <p class="admin-page__lead">
      Organize o cardápio. Use a ordem para definir a sequência exibida.
    </p>

    <BaseButton @click="startNew">Nova categoria</BaseButton>

    <AdminCatalogFilters :filters="store.filters" @apply="applyFilters" />

    <BaseModal :open="showForm" :title="modalTitle" @close="cancel">
      <AdminCatalogForm
        kind="category"
        :item="editing"
        :saving="store.saving"
        :error="store.saveError"
        :success="store.saveSuccess"
        @submit="submit"
        @cancel="cancel"
      />
    </BaseModal>

    <AdminCatalogList
      kind="category"
      :items="store.items"
      :state="store.state"
      :error="store.error"
      :isEmpty="store.isEmpty"
      :meta="store.meta"
      @edit="edit"
      @toggle="toggle"
      @page="goToPage"
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
</style>
