<script setup lang="ts">
import { computed, ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'

import BaseButton from '../../components/BaseButton.vue'
import { useAdminAuthStore } from '../../stores/adminAuth'

const auth = useAdminAuthStore()
const router = useRouter()
const route = useRoute()

const email = ref('')
const password = ref('')
const submitting = ref(false)
const error = ref('')

const redirect = computed(() => {
  const value = route.query.redirect
  return typeof value === 'string' && value.startsWith('/admin') ? value : '/admin'
})

async function submit(): Promise<void> {
  if (submitting.value) {
    return
  }

  if (!email.value.trim() || !password.value) {
    error.value = 'Informe o e-mail e a senha para entrar.'
    return
  }

  submitting.value = true
  error.value = ''

  const authenticated = await auth.login(email.value.trim(), password.value)
  submitting.value = false

  if (!authenticated) {
    error.value = auth.error
    return
  }

  await router.replace(redirect.value)
}
</script>

<template>
  <main id="conteudo" tabindex="-1" class="admin-login">
    <section class="admin-login__card" aria-labelledby="admin-login-title">
      <p class="admin-login__kicker">Área da operação</p>
      <h1 id="admin-login-title">Acessar painel</h1>
      <p class="admin-login__hint">Use o e-mail e a senha fornecidos pela gestão da loja.</p>

      <form class="admin-login__form" @submit.prevent="submit">
        <div class="admin-login__field">
          <label for="admin-email">E-mail</label>
          <input
            id="admin-email"
            v-model="email"
            type="email"
            name="email"
            autocomplete="username"
            inputmode="email"
            required
          />
        </div>

        <div class="admin-login__field">
          <label for="admin-password">Senha</label>
          <input
            id="admin-password"
            v-model="password"
            type="password"
            name="password"
            autocomplete="current-password"
            required
          />
        </div>

        <p v-if="error" class="admin-login__error" role="alert">{{ error }}</p>

        <BaseButton type="submit" :loading="submitting" :disabled="submitting">Entrar</BaseButton>
      </form>

      <RouterLink class="admin-login__back" :to="{ name: 'home' }"> Voltar ao cardápio </RouterLink>
    </section>
  </main>
</template>

<style scoped>
.admin-login {
  display: grid;
  min-height: 100vh;
  min-height: 100dvh;
  place-items: center;
  padding: calc(24px + env(safe-area-inset-top)) 16px calc(64px + env(safe-area-inset-bottom));
}

.admin-login__card {
  width: min(420px, 100%);
  padding: 28px 24px;
  border: 1px solid var(--bg-border);
  border-radius: 24px;
  background: var(--bg-surface);
}

.admin-login__kicker {
  margin: 0;
  color: var(--brand-neon);
  font-size: var(--font-size-caption);
  font-weight: 800;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.admin-login h1 {
  margin: 8px 0;
  font-family: var(--font-display);
  font-size: var(--font-size-display);
  line-height: var(--line-height-display);
  letter-spacing: 0.02em;
}

.admin-login__hint {
  margin: 0 0 24px;
  color: var(--text-sub);
}

.admin-login__form {
  display: grid;
  gap: 16px;
}

.admin-login__field {
  display: grid;
  gap: 8px;
}

.admin-login__field label {
  font-weight: 700;
}

.admin-login__field input {
  width: 100%;
  min-height: 48px;
  padding: 12px 14px;
  border: 1px solid var(--bg-border);
  border-radius: 12px;
  background: var(--bg-input);
  color: var(--text-main);
}

.admin-login__field input:focus-visible {
  border-color: var(--brand-primary);
}

.admin-login__error {
  margin: 0;
  padding: 12px 14px;
  border: 1px solid var(--status-error);
  border-radius: 12px;
  color: var(--text-main);
}

.admin-login__back {
  display: inline-flex;
  min-height: 44px;
  align-items: center;
  margin-top: 16px;
  color: var(--brand-neon);
  font-weight: 700;
}
</style>
