<script setup lang="ts">
import EmptyState from '../EmptyState.vue'
import ErrorState from '../ErrorState.vue'
import LoadingSkeleton from '../LoadingSkeleton.vue'
import type { AdminAuditState } from '../../stores/adminUsers'
import type { AdminPageMeta } from '../../types/admin'
import type { AdminAuditLog } from '../../types/adminUsers'

const props = defineProps<{
  items: AdminAuditLog[]
  state: AdminAuditState
  error: string
  isEmpty: boolean
  meta: AdminPageMeta
}>()

const emit = defineEmits<{
  page: [page: number]
  retry: []
}>()

const entityLabels: Record<string, string> = {
  categories: 'Categorias',
  products: 'Produtos',
  addons: 'Adicionais',
  delivery_zones: 'Zonas de entrega',
  store_settings: 'Configuração da loja',
  orders: 'Pedidos',
  admin_users: 'Usuários do painel',
}

function label(entity: string): string {
  return entityLabels[entity] ?? entity
}

function formatDate(value: string): string {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }
  return new Intl.DateTimeFormat('pt-BR', {
    dateStyle: 'short',
    timeStyle: 'short',
    timeZone: 'America/Fortaleza',
  }).format(date)
}

function goToPage(page: number): void {
  emit('page', page)
}
</script>

<template>
  <section class="audit-list" aria-label="Registro de auditoria">
    <LoadingSkeleton v-if="props.state === 'loading'" label="Carregando auditoria" :count="3" />
    <ErrorState v-else-if="props.state === 'error'" :message="props.error" @retry="emit('retry')" />
    <EmptyState
      v-else-if="props.isEmpty"
      title="Nenhuma alteração registrada"
      description="As mutações do painel aparecem aqui."
    />

    <template v-else-if="props.state === 'success'">
      <ol class="audit-list__items">
        <li v-for="log in props.items" :key="log.id" class="audit-row">
          <strong>{{ label(log.entity) }}</strong>
          <span>{{ log.action === 'create' ? 'Criação' : 'Atualização' }}</span>
          <span>{{ formatDate(log.createdAt) }}</span>
          <span class="audit-row__actor">Registro: {{ log.entityId }}</span>
        </li>
      </ol>

      <nav class="audit-list__pagination" aria-label="Paginação de auditoria">
        <button
          type="button"
          class="audit-list__button"
          :disabled="props.meta.page <= 1"
          @click="goToPage(props.meta.page - 1)"
        >
          Anterior
        </button>
        <span>Página {{ props.meta.page }} de {{ Math.max(props.meta.totalPages, 1) }}</span>
        <button
          type="button"
          class="audit-list__button"
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
.audit-list {
  display: grid;
  gap: 14px;
}

.audit-list__items {
  display: grid;
  gap: 10px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.audit-row {
  display: grid;
  gap: 4px;
  padding: 14px;
  border: 1px solid var(--bg-border);
  border-radius: 14px;
  background: var(--bg-surface);
  font-size: var(--font-size-caption);
}

.audit-row strong {
  font-size: var(--font-size-body);
}

.audit-row span {
  color: var(--text-sub);
}

.audit-row__actor {
  overflow-wrap: anywhere;
}

.audit-list__pagination {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px;
}

.audit-list__button {
  min-height: 44px;
  padding: 10px 12px;
  border: 1px solid var(--bg-border);
  border-radius: 12px;
  background: var(--bg-input);
  color: var(--text-main);
  font-weight: 700;
}

.audit-list__button:disabled {
  opacity: 0.5;
}

.audit-list__pagination span {
  color: var(--text-sub);
  font-size: var(--font-size-caption);
}
</style>
