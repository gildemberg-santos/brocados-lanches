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
    <p class="admin-page__kicker">Painel administrativo</p>
    <h1>Minha conta</h1>

    <section class="account-card">
      <h2>Dados da sessão</h2>
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
      <p class="account-card__note">
        A troca de senha e o MFA não fazem parte da V1. Para alterar seus dados, peça a um dono no
        painel de usuários.
      </p>
      <BaseButton :loading="auth.loading" @click="logout">Sair do painel</BaseButton>
    </section>
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

.account-card {
  display: grid;
  gap: 14px;
  padding: 18px;
  border: 1px solid var(--bg-border);
  border-radius: 18px;
  background: var(--bg-surface);
}

.account-card h2 {
  margin: 0;
  font-size: var(--font-size-h2);
}

.account-card dl {
  display: grid;
  gap: 10px;
  margin: 0;
}

.account-card dl div {
  display: grid;
  gap: 2px;
}

.account-card dt {
  color: var(--text-sub);
  font-size: var(--font-size-caption);
}

.account-card dd {
  margin: 0;
  overflow-wrap: anywhere;
}

.account-card__note {
  margin: 0;
  color: var(--text-sub);
  font-size: var(--font-size-caption);
}
</style>
