<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

import AdminCatalogFilters from '../../components/admin/AdminCatalogFilters.vue'
import AdminProductForm from '../../components/admin/AdminProductForm.vue'
import BaseButton from '../../components/BaseButton.vue'
import BaseModal from '../../components/BaseModal.vue'
import EmptyState from '../../components/EmptyState.vue'
import ErrorState from '../../components/ErrorState.vue'
import LoadingSkeleton from '../../components/LoadingSkeleton.vue'
import PriceText from '../../components/PriceText.vue'
import {
  useAdminAddonsStore,
  useAdminCategoriesStore,
  useAdminProductsStore,
} from '../../stores/adminCatalog'
import type {
  AdminCatalogListQuery,
  AdminProduct,
  AdminProductInput,
} from '../../types/adminCatalog'

const products = useAdminProductsStore()
const categories = useAdminCategoriesStore()
const addons = useAdminAddonsStore()
const editing = ref<AdminProduct | null>(null)
const showForm = ref(false)

const modalTitle = computed(() => (editing.value ? 'Editar produto' : 'Novo produto'))

const categoryNames = computed(() => {
  const names = new Map<string, string>()
  for (const category of categories.items) {
    names.set(category.id, category.name)
  }
  return names
})

function applyFilters(filters: Partial<AdminCatalogListQuery>): void {
  void products.applyFilters(filters)
}

function count(items: readonly unknown[] | null | undefined): number {
  return Array.isArray(items) ? items.length : 0
}

function selectCategory(event: Event): void {
  const value = (event.target as HTMLSelectElement).value
  void products.applyFilters({ ...products.filters, categoryId: value || undefined })
}

function goToPage(page: number): void {
  void products.goToPage(page)
}

function startNew(): void {
  editing.value = null
  showForm.value = true
  products.clearMessages()
}

function edit(item: AdminProduct): void {
  editing.value = item
  showForm.value = true
  products.clearMessages()
}

function cancel(): void {
  showForm.value = false
  editing.value = null
  products.clearMessages()
}

async function submit(input: AdminProductInput): Promise<void> {
  const saved = await products.save(editing.value?.id ?? null, input)
  if (saved) {
    cancel()
  }
}

async function toggle(item: AdminProduct): Promise<void> {
  await products.save(item.id, { active: !item.active })
}

onMounted(() => {
  if (products.state === 'idle') {
    void products.load()
  }
  if (categories.state === 'idle') {
    void categories.applyFilters({ pageSize: 100 })
  }
  if (addons.state === 'idle') {
    void addons.applyFilters({ pageSize: 100 })
  }
})
</script>

<template>
  <main class="admin-page">
    <p class="admin-page__kicker">Painel administrativo</p>
    <h1>Produtos</h1>
    <p class="admin-page__lead">
      Cadastre preço, opções obrigatórias, adicionais permitidos e disponibilidade.
    </p>

    <BaseButton @click="startNew">Novo produto</BaseButton>

    <div class="admin-page__filters">
      <div class="admin-page__field">
        <label for="product-category-filter">Categoria</label>
        <select id="product-category-filter" @change="selectCategory">
          <option value="">Todas</option>
          <option v-for="category in categories.items" :key="category.id" :value="category.id">
            {{ category.name }}
          </option>
        </select>
      </div>
      <AdminCatalogFilters :filters="products.filters" @apply="applyFilters" />
    </div>

    <BaseModal :open="showForm" :title="modalTitle" @close="cancel">
      <AdminProductForm
        :item="editing"
        :categories="categories.items"
        :addons="addons.items"
        :saving="products.saving"
        :error="products.saveError"
        :success="products.saveSuccess"
        @submit="submit"
        @cancel="cancel"
      />
    </BaseModal>

    <LoadingSkeleton v-if="products.state === 'loading'" label="Carregando produtos" :count="3" />

    <ErrorState
      v-else-if="products.state === 'error'"
      :message="products.error"
      @retry="products.load"
    />

    <EmptyState
      v-else-if="products.isEmpty"
      title="Nenhum produto encontrado"
      description="Ajuste os filtros ou cadastre um novo produto."
    />

    <template v-else-if="products.state === 'success'">
      <ol class="product-list">
        <li v-for="item in products.items" :key="item.id" class="product-row">
          <div class="product-row__head">
            <h2>{{ item.name }}</h2>
            <PriceText :value-cents="item.priceCents" />
          </div>
          <p class="product-row__slug">{{ item.slug }}</p>
          <dl class="product-row__meta">
            <div>
              <dt>Categoria</dt>
              <dd>{{ categoryNames.get(item.categoryId) ?? '—' }}</dd>
            </div>
            <div>
              <dt>Situação</dt>
              <dd>
                {{ item.active ? 'Ativo' : 'Inativo' }} ·
                {{ item.available ? 'Disponível' : 'Esgotado' }}
              </dd>
            </div>
            <div>
              <dt>Opções</dt>
              <dd>
                {{ count(item.optionGroups) }} grupo(s) · {{ count(item.addonIds) }} adicional(is)
              </dd>
            </div>
          </dl>
          <div class="product-row__actions">
            <button type="button" class="product-row__button" @click="edit(item)">Editar</button>
            <button type="button" class="product-row__button" @click="toggle(item)">
              {{ item.active ? 'Desativar' : 'Ativar' }}
            </button>
          </div>
        </li>
      </ol>

      <nav
        v-if="products.totalPages > 0"
        class="product-pagination"
        aria-label="Paginação de produtos"
      >
        <button
          type="button"
          :disabled="products.meta.page <= 1"
          @click="goToPage(products.meta.page - 1)"
        >
          Anterior
        </button>
        <span aria-live="polite">Página {{ products.meta.page }} de {{ products.totalPages }}</span>
        <button
          type="button"
          :disabled="products.meta.page >= products.totalPages"
          @click="goToPage(products.meta.page + 1)"
        >
          Próxima
        </button>
      </nav>
    </template>
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

.admin-page__filters {
  display: grid;
  gap: 12px;
}

.admin-page__field {
  display: grid;
  gap: 8px;
}

.admin-page__field label {
  font-size: var(--font-size-caption);
  font-weight: 700;
}

.admin-page__field select {
  min-height: 44px;
  padding: 10px 12px;
  border: 1px solid var(--bg-border);
  border-radius: 12px;
  background: var(--bg-input);
  color: var(--text-main);
  font-family: var(--font-sans);
}

.product-list {
  display: grid;
  gap: 12px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.product-row {
  display: grid;
  gap: 12px;
  padding: 18px;
  border: 1px solid var(--bg-border);
  border-radius: 18px;
  background: var(--bg-surface);
}

.product-row__head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.product-row h2 {
  margin: 0;
  font-family: var(--font-display);
  font-size: var(--font-size-h2);
}

.product-row__slug {
  margin: 0;
  color: var(--text-sub);
  font-size: var(--font-size-caption);
}

.product-row__meta {
  display: grid;
  gap: 8px;
  margin: 0;
}

.product-row__meta div {
  display: flex;
  justify-content: space-between;
  gap: 12px;
}

.product-row__meta dt,
.product-row__meta dd {
  margin: 0;
  font-size: var(--font-size-caption);
}

.product-row__meta dt {
  color: var(--text-sub);
}

.product-row__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.product-row__button,
.product-pagination button {
  min-height: 44px;
  padding: 10px 16px;
  border: 1px solid var(--bg-border);
  border-radius: 999px;
  background: var(--bg-input);
  color: var(--text-main);
  font-family: var(--font-sans);
  font-size: var(--font-size-caption);
  font-weight: 700;
}

.product-pagination {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: center;
  gap: 12px;
  color: var(--text-sub);
  font-size: var(--font-size-caption);
}

@media (min-width: 768px) {
  .product-row__meta {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}
</style>
