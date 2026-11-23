<script setup lang="ts">
import EmptyState from '../EmptyState.vue'
import ErrorState from '../ErrorState.vue'
import LoadingSkeleton from '../LoadingSkeleton.vue'
import type { AdminUsersState } from '../../stores/adminUsers'
import type { AdminPageMeta } from '../../types/admin'
import type { AdminUser } from '../../types/adminUsers'

const props = defineProps<{
  items: AdminUser[]
  state: AdminUsersState
  error: string
  isEmpty: boolean
  meta: AdminPageMeta
  actorId: string
}>()

const emit = defineEmits<{
  edit: [user: AdminUser]
  toggle: [user: AdminUser]
  page: [page: number]
  retry: []
}>()

function goToPage(page: number): void {
  emit('page', page)
}
</script>

<template>
  <section class="user-list" aria-label="Usuários do painel">
    <LoadingSkeleton v-if="props.state === 'loading'" label="Carregando usuários" :count="3" />
    <ErrorState v-else-if="props.state === 'error'" :message="props.error" @retry="emit('retry')" />
    <EmptyState
      v-else-if="props.isEmpty"
      title="Nenhum usuário encontrado"
      description="Ajuste os filtros ou cadastre um novo usuário."
    />

    <template v-else-if="props.state === 'success'">
      <ol class="user-list__items">
        <li v-for="user in props.items" :key="user.id" class="user-row">
          <div class="user-row__head">
            <h2>{{ user.name }}</h2>
            <span class="user-row__badge" :class="{ 'user-row__badge--off': !user.active }">
              {{ user.active ? 'Ativo' : 'Inativo' }}
            </span>
          </div>
          <dl class="user-row__meta">
            <div>
              <dt>E-mail</dt>
              <dd>{{ user.email }}</dd>
            </div>
            <div>
              <dt>Papel</dt>
              <dd>{{ user.role === 'owner' ? 'Dono' : 'Operador' }}</dd>
            </div>
          </dl>
          <div class="user-row__actions">
            <button type="button" class="user-row__button" @click="emit('edit', user)">
              Editar
            </button>
            <button
              v-if="user.id !== props.actorId"
              type="button"
              class="user-row__button"
              @click="emit('toggle', user)"
            >
              {{ user.active ? 'Desativar' : 'Ativar' }}
            </button>
          </div>
        </li>
      </ol>

      <nav class="user-list__pagination" aria-label="Paginação de usuários">
        <button
          type="button"
          class="user-row__button"
          :disabled="props.meta.page <= 1"
          @click="goToPage(props.meta.page - 1)"
        >
          Anterior
        </button>
        <span>Página {{ props.meta.page }} de {{ Math.max(props.meta.totalPages, 1) }}</span>
        <button
          type="button"
          class="user-row__button"
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
.user-list {
  display: grid;
  gap: 14px;
}

.user-list__items {
  display: grid;
  gap: 12px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.user-row {
  display: grid;
  gap: 12px;
  padding: 16px;
  border: 1px solid var(--bg-border);
  border-radius: 18px;
  background: var(--bg-surface);
}

.user-row__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.user-row h2 {
  margin: 0;
  font-size: var(--font-size-h2);
}

.user-row__badge {
  padding: 4px 10px;
  border-radius: 999px;
  background: var(--bg-input);
  color: var(--status-open);
  font-size: var(--font-size-caption);
  font-weight: 800;
}

.user-row__badge--off {
  color: var(--status-error);
}

.user-row__meta {
  display: grid;
  gap: 8px;
  margin: 0;
}

.user-row__meta div {
  display: grid;
  gap: 2px;
}

.user-row__meta dt {
  color: var(--text-sub);
  font-size: var(--font-size-caption);
}

.user-row__meta dd {
  margin: 0;
  overflow-wrap: anywhere;
}

.user-row__actions,
.user-list__pagination {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px;
}

.user-row__button {
  min-height: 44px;
  padding: 10px 12px;
  border: 1px solid var(--bg-border);
  border-radius: 12px;
  background: var(--bg-input);
  color: var(--text-main);
  font-weight: 700;
}

.user-row__button:disabled {
  opacity: 0.5;
}

.user-list__pagination span {
  color: var(--text-sub);
  font-size: var(--font-size-caption);
}

@media (min-width: 768px) {
  .user-row__meta {
    grid-template-columns: 2fr 1fr;
  }
}
</style>
