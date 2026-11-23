<script setup lang="ts">
import { onMounted, ref } from 'vue'

import AdminCatalogFilters from '../../components/admin/AdminCatalogFilters.vue'
import AdminCatalogForm from '../../components/admin/AdminCatalogForm.vue'
import AdminCatalogList from '../../components/admin/AdminCatalogList.vue'
import type { AdminCatalogRow } from '../../components/admin/AdminCatalogList.vue'
import BaseButton from '../../components/BaseButton.vue'
import { useAdminAddonsStore } from '../../stores/adminCatalog'
import type { AdminAddon, AdminCatalogInput, AdminCatalogListQuery } from '../../types/adminCatalog'

const store = useAdminAddonsStore()
const editing = ref<AdminAddon | null>(null)
const showForm = ref(false)

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
  <main id="conteudo" class="admin-page">
    <p class="admin-page__kicker">Painel administrativo</p>
    <h1>Adicionais</h1>
    <p class="admin-page__lead">
      Informe o preço em reais; a API recebe e grava o valor em centavos.
    </p>

    <BaseButton @click="startNew">Novo adicional</BaseButton>

    <AdminCatalogFilters :filters="store.filters" @apply="applyFilters" />

    <AdminCatalogForm
      v-if="showForm"
      kind="addon"
      :item="editing"
      :saving="store.saving"
      :error="store.saveError"
      :success="store.saveSuccess"
      @submit="submit"
      @cancel="cancel"
    />

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
