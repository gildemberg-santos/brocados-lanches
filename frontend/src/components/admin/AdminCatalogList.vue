<script setup lang="ts">
import EmptyState from '../EmptyState.vue'
import ErrorState from '../ErrorState.vue'
import LoadingSkeleton from '../LoadingSkeleton.vue'
import PriceText from '../PriceText.vue'
import type { AdminCatalogState } from '../../stores/adminCatalog'
import type { AdminPageMeta } from '../../types/admin'
import type { AdminCatalogKind } from '../../types/adminCatalog'

export interface AdminCatalogRow {
  id: string
  name: string
  slug: string
  active: boolean
  sortOrder: number
  description?: string
  priceCents?: number
  feeCents?: number
}

const props = defineProps<{
  kind: AdminCatalogKind
  items: AdminCatalogRow[]
  state: AdminCatalogState
  error: string
  isEmpty: boolean
  meta: AdminPageMeta
}>()

const emit = defineEmits<{
  edit: [item: AdminCatalogRow]
  toggle: [item: AdminCatalogRow]
  page: [page: number]
  retry: []
}>()

const titles: Record<AdminCatalogKind, string> = {
  category: 'Categorias',
  addon: 'Adicionais',
  zone: 'Locais e taxas',
}

function goToPage(page: number): void {
  emit('page', page)
}
</script>

<template>
  <section class="catalog-list" :aria-label="titles[props.kind]">
    <LoadingSkeleton
      v-if="props.state === 'loading'"
      :label="'Carregando ' + titles[props.kind]"
      :count="3"
    />

    <ErrorState v-else-if="props.state === 'error'" :message="props.error" @retry="emit('retry')" />

    <EmptyState
      v-else-if="props.isEmpty"
      :title="`Nenhum item encontrado`"
      description="Ajuste os filtros ou cadastre um novo item."
    />

    <template v-else-if="props.state === 'success'">
      <ol class="catalog-list__items">
        <li v-for="item in props.items" :key="item.id" class="catalog-row">
          <div class="catalog-row__head">
            <h2>{{ item.name }}</h2>
            <span class="catalog-row__status" :class="{ 'catalog-row__status--off': !item.active }">
              {{ item.active ? 'Ativo' : 'Inativo' }}
            </span>
          </div>

          <dl class="catalog-row__meta">
            <div>
              <dt>Slug</dt>
              <dd>{{ item.slug }}</dd>
            </div>
            <div>
              <dt>Ordem</dt>
              <dd>{{ item.sortOrder }}</dd>
            </div>
            <div v-if="props.kind === 'category'">
              <dt>Descrição</dt>
              <dd>{{ item.description || '—' }}</dd>
            </div>
            <div v-if="props.kind === 'addon'">
              <dt>Preço</dt>
              <dd><PriceText :value-cents="item.priceCents ?? 0" /></dd>
            </div>
            <div v-if="props.kind === 'zone'">
              <dt>Taxa</dt>
              <dd><PriceText :value-cents="item.feeCents ?? 0" /></dd>
            </div>
          </dl>

          <div class="catalog-row__actions">
            <button type="button" class="catalog-row__button" @click="emit('edit', item)">
              Editar
            </button>
            <button type="button" class="catalog-row__button" @click="emit('toggle', item)">
              {{ item.active ? 'Desativar' : 'Ativar' }}
            </button>
          </div>
        </li>
      </ol>

      <nav v-if="props.meta.totalPages > 0" class="catalog-list__pagination" aria-label="Paginação">
        <button
          type="button"
          class="catalog-list__button"
          :disabled="props.meta.page <= 1"
          @click="goToPage(props.meta.page - 1)"
        >
          Anterior
        </button>
        <span aria-live="polite">Página {{ props.meta.page }} de {{ props.meta.totalPages }}</span>
        <button
          type="button"
          class="catalog-list__button"
          :disabled="props.meta.page >= props.meta.totalPages"
          @click="goToPage(props.meta.page + 1)"
        >
          Próxima
        </button>
      </nav>
    </template>
  </section>
</template>

<style scoped>
.catalog-list {
  display: grid;
  gap: 16px;
}

.catalog-list__items {
  display: grid;
  gap: 12px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.catalog-row {
  display: grid;
  gap: 12px;
  padding: 18px;
  border: 1px solid var(--bg-border);
  border-radius: 18px;
  background: var(--bg-surface);
}

.catalog-row__head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.catalog-row h2 {
  margin: 0;
  font-family: var(--font-display);
  font-size: var(--font-size-h2);
}

.catalog-row__status {
  padding: 4px 10px;
  border-radius: 999px;
  background: var(--brand-primary);
  color: var(--bg-primary);
  font-size: var(--font-size-caption);
  font-weight: 800;
}

.catalog-row__status--off {
  background: var(--bg-input);
  color: var(--text-sub);
}

.catalog-row__meta {
  display: grid;
  gap: 8px;
  margin: 0;
}

.catalog-row__meta div {
  display: flex;
  justify-content: space-between;
  gap: 12px;
}

.catalog-row__meta dt,
.catalog-row__meta dd {
  margin: 0;
  font-size: var(--font-size-caption);
}

.catalog-row__meta dt {
  color: var(--text-sub);
}

.catalog-row__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.catalog-row__button,
.catalog-list__button {
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

.catalog-row__button:disabled,
.catalog-list__button:disabled {
  opacity: 0.5;
}

.catalog-list__pagination {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: center;
  gap: 12px;
  color: var(--text-sub);
  font-size: var(--font-size-caption);
}

@media (min-width: 768px) {
  .catalog-row__meta {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
