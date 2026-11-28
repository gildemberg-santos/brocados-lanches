<script setup lang="ts">
import { onMounted, ref } from 'vue'

import AdminAuditList from '../../components/admin/AdminAuditList.vue'
import { useAdminUsersStore } from '../../stores/adminUsers'

const store = useAdminUsersStore()
const entity = ref('')
const action = ref('')

function applyFilters(): void {
  void store.applyAuditFilters({
    entity: entity.value || undefined,
    action: action.value || undefined,
  })
}

function clearFilters(): void {
  entity.value = ''
  action.value = ''
  void store.applyAuditFilters({ entity: undefined, action: undefined })
}

onMounted(() => {
  if (store.auditState === 'idle') {
    void store.loadAudit()
  }
})
</script>

<template>
  <main class="admin-page">
    <header class="admin-hero">
      <div>
        <p class="admin-page__kicker">Painel administrativo</p>
        <h1>Auditoria</h1>
        <p class="admin-page__lead">
          Registro somente leitura das alterações feitas no painel, sem dados pessoais.
        </p>
      </div>
      <div class="admin-hero__meta">
        <span class="admin-hero__meta-label">Registros no filtro</span>
        <strong class="admin-hero__meta-value">{{ store.auditMeta.total }}</strong>
        <span class="admin-hero__meta-note">Somente leitura</span>
      </div>
    </header>

    <section class="admin-panel">
      <div class="admin-section-head">
        <div>
          <p class="admin-section__eyebrow">Rastreabilidade</p>
          <h2>Filtrar registros</h2>
        </div>
      </div>
      <form class="audit-filters" @submit.prevent="applyFilters">
        <div class="audit-filters__field">
          <label for="audit-entity">Entidade</label>
          <select id="audit-entity" v-model="entity">
            <option value="">Todas</option>
            <option value="products">Produtos</option>
            <option value="categories">Categorias</option>
            <option value="addons">Adicionais</option>
            <option value="delivery_zones">Zonas de entrega</option>
            <option value="store_settings">Configuração da loja</option>
            <option value="orders">Pedidos</option>
            <option value="admin_users">Usuários do painel</option>
          </select>
        </div>
        <div class="audit-filters__field">
          <label for="audit-action">Ação</label>
          <select id="audit-action" v-model="action">
            <option value="">Todas</option>
            <option value="create">Criação</option>
            <option value="update">Atualização</option>
          </select>
        </div>
        <button type="submit">Filtrar</button>
        <button type="button" @click="clearFilters">Limpar</button>
      </form>
    </section>

    <AdminAuditList
      :items="store.audits"
      :state="store.auditState"
      :error="store.auditError"
      :isEmpty="store.auditIsEmpty"
      :meta="store.auditMeta"
      @page="store.auditGoToPage"
      @retry="store.loadAudit"
    />
  </main>
</template>

<style scoped>
/* layout herdado de admin.css */

.audit-filters {
  display: flex;
  flex-wrap: wrap;
  align-items: end;
  gap: 12px;
}

.audit-filters__field {
  display: grid;
  gap: 6px;
}

.audit-filters label {
  font-size: var(--font-size-caption);
  font-weight: 700;
}

.audit-filters select,
.audit-filters button {
  min-height: 44px;
  padding: 10px 12px;
  border: 1px solid var(--bg-border);
  border-radius: 12px;
  background: var(--bg-input);
  color: var(--text-main);
  font-family: var(--font-sans);
  font-size: var(--font-size-body);
  font-weight: 700;
}

.audit-filters button[type='submit'] {
  border-color: var(--brand-primary);
  background: var(--brand-primary);
  color: var(--bg-primary);
}
</style>
