<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

import AdminCatalogFilters from '../../components/admin/AdminCatalogFilters.vue'
import AdminCatalogForm from '../../components/admin/AdminCatalogForm.vue'
import AdminCatalogList from '../../components/admin/AdminCatalogList.vue'
import type { AdminCatalogRow } from '../../components/admin/AdminCatalogList.vue'
import BaseButton from '../../components/BaseButton.vue'
import BaseModal from '../../components/BaseModal.vue'
import { useAdminAddonsStore } from '../../stores/adminCatalog'
import type { AdminAddon, AdminCatalogInput, AdminCatalogListQuery } from '../../types/adminCatalog'

const store = useAdminAddonsStore()
const editing = ref<AdminAddon | null>(null)
const showForm = ref(false)

const modalTitle = computed(() => (editing.value ? 'Editar adicional' : 'Novo adicional'))

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
  editing.value = item as AdminAddon
  showForm.value = true
  store.clearMessages()
}

function cancel(): void {
  showForm.value = false
  editing.value = null
  store.clearMessages()
}

async function submit(input: AdminCatalogInput): Promise<void> {
  if (!('priceCents' in input)) {
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
    <header class="admin-hero">
      <div>
        <p class="admin-page__kicker">Painel administrativo</p>
        <h1>Adicionais</h1>
        <p class="admin-page__lead">
          Gerencie complementos do cardápio. O preço é informado em reais e gravado em centavos.
        </p>
      </div>
      <div class="admin-hero__meta">
        <span class="admin-hero__meta-label">Total no filtro</span>
        <strong class="admin-hero__meta-value">{{ store.meta.total }}</strong>
        <span class="admin-hero__meta-note"
          >{{ store.items.filter((i) => i.active).length }} ativo(s) nesta página</span
        >
      </div>
    </header>

    <section class="admin-panel">
      <div class="admin-section-head">
        <div>
          <p class="admin-section__eyebrow">Catálogo</p>
          <h2>Gerenciar adicionais</h2>
        </div>
        <BaseButton @click="startNew">Novo adicional</BaseButton>
      </div>
      <AdminCatalogFilters :filters="store.filters" @apply="applyFilters" />
    </section>

    <BaseModal :open="showForm" :title="modalTitle" @close="cancel">
      <AdminCatalogForm
        kind="addon"
        :item="editing"
        :saving="store.saving"
        :error="store.saveError"
        :success="store.saveSuccess"
        @submit="submit"
        @cancel="cancel"
      />
    </BaseModal>

    <AdminCatalogList
      kind="addon"
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
.admin-panel :deep(.catalog-filters) {
  padding: 0;
  border: 0;
  background: transparent;
}
</style>
