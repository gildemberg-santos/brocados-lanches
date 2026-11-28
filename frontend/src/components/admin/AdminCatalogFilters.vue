<script setup lang="ts">
import { ref } from 'vue'

import BaseButton from '../BaseButton.vue'
import type { AdminCatalogListQuery } from '../../types/adminCatalog'

const props = defineProps<{ filters: AdminCatalogListQuery }>()
const emit = defineEmits<{ apply: [value: Partial<AdminCatalogListQuery>] }>()

const search = ref(props.filters.q ?? '')
const active = ref(
  props.filters.active === undefined ? '' : props.filters.active ? 'true' : 'false',
)

function apply(): void {
  emit('apply', {
    q: search.value.trim() || undefined,
    active: active.value === '' ? undefined : active.value === 'true',
  })
}

function clear(): void {
  search.value = ''
  active.value = ''
  emit('apply', {})
}
</script>

<template>
  <form class="catalog-filters" aria-label="Filtros de catálogo" @submit.prevent="apply">
    <div class="catalog-filters__field">
      <label for="catalog-search">Buscar</label>
      <input id="catalog-search" v-model="search" type="search" placeholder="Nome ou slug" />
    </div>

    <div class="catalog-filters__field">
      <label for="catalog-active">Situação</label>
      <select id="catalog-active" v-model="active">
        <option value="">Todos</option>
        <option value="true">Ativos</option>
        <option value="false">Inativos</option>
      </select>
    </div>

    <div class="catalog-filters__actions">
      <BaseButton type="submit">Filtrar</BaseButton>
      <BaseButton type="button" variant="ghost" @click="clear">Limpar</BaseButton>
    </div>
  </form>
</template>

<style scoped>
.catalog-filters {
  display: grid;
  gap: 12px;
  padding: 16px;
  border: 1px solid var(--bg-border);
  border-radius: 16px;
  background: color-mix(in srgb, var(--bg-input) 55%, var(--bg-surface));
}

.catalog-filters__field {
  display: grid;
  gap: 8px;
}

.catalog-filters__field label {
  font-size: var(--font-size-caption);
  font-weight: 700;
}

.catalog-filters input,
.catalog-filters select {
  width: 100%;
  min-height: 44px;
  padding: 10px 12px;
  border: 1px solid var(--bg-border);
  border-radius: 12px;
  background: var(--bg-input);
  color: var(--text-main);
  font-family: var(--font-sans);
  font-size: var(--font-size-body);
}

.catalog-filters__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

@media (min-width: 768px) {
  .catalog-filters {
    grid-template-columns: 2fr 1fr auto;
    align-items: end;
  }
}
</style>
