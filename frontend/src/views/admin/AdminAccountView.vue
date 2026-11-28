<script setup lang="ts">
import { useRouter } from 'vue-router'

import BaseButton from '../../components/BaseButton.vue'
import { useAdminAuthStore } from '../../stores/adminAuth'

const auth = useAdminAuthStore()
const router = useRouter()

async function logout(): Promise<void> {
  await auth.logout()
  await router.push({ name: 'admin-login' })
}
</script>

<template>
  <main class="admin-page">
    <header class="admin-hero">
      <div>
        <p class="admin-page__kicker">Painel administrativo</p>
        <h1>Minha conta</h1>
        <p class="admin-page__lead">Dados da sessão atual e acesso rápido para sair do painel.</p>
      </div>
      <div class="admin-hero__meta">
        <span class="admin-hero__meta-label">Papel</span>
        <strong class="admin-hero__meta-value">{{
          auth.role === 'owner' ? 'Dono' : auth.role === 'operator' ? 'Operador' : '—'
        }}</strong>
        <span class="admin-hero__meta-note">{{ auth.user?.email ?? '—' }}</span>
      </div>
    </header>

    <section class="admin-panel account-card">
      <div class="admin-section-head">
        <div>
          <p class="admin-section__eyebrow">Sessão</p>
          <h2>Dados da conta</h2>
        </div>
      </div>
      <dl>
        <div>
          <dt>Nome</dt>
          <dd>{{ auth.userName || '—' }}</dd>
        </div>
        <div>
          <dt>E-mail</dt>
          <dd>{{ auth.user?.email ?? '—' }}</dd>
        </div>
        <div>
          <dt>Papel</dt>
          <dd>
            {{ auth.role === 'owner' ? 'Dono' : auth.role === 'operator' ? 'Operador' : '—' }}
          </dd>
        </div>
      </dl>
      <p class="account-card__note admin-empty-inline">
        A troca de senha e o MFA não fazem parte da V1. Para alterar seus dados, peça a um dono no
        painel de usuários.
      </p>
      <div class="admin-toolbar">
        <BaseButton :loading="auth.loading" @click="logout">Sair do painel</BaseButton>
      </div>
    </section>
  </main>
</template>

<style scoped>
.account-card dl {
  display: grid;
  gap: 10px;
  margin: 0;
}

.account-card dl div {
  display: grid;
  gap: 2px;
  padding: 12px 0;
  border-bottom: 1px solid var(--bg-border);
}

.account-card dl div:last-child {
  padding-bottom: 0;
  border-bottom: 0;
}

.account-card dt {
  color: var(--text-sub);
  font-size: var(--font-size-caption);
}

.account-card dd {
  margin: 0;
  overflow-wrap: anywhere;
  font-weight: 700;
}

.account-card__note {
  margin: 0;
}
</style>
