<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'

import BaseButton from '../../components/BaseButton.vue'
import EmptyState from '../../components/EmptyState.vue'
import ErrorState from '../../components/ErrorState.vue'
import LoadingSkeleton from '../../components/LoadingSkeleton.vue'
import type { WebhookFieldErrors } from '../../features/admin/webhookForm'
import { validateWebhookInput } from '../../features/admin/webhookForm'
import { useAdminWebhookStore } from '../../stores/adminWebhook'
import type { AdminWebhookInput } from '../../types/adminWebhook'

const store = useAdminWebhookStore()
const errors = ref<WebhookFieldErrors>({ url: undefined })
const form = ref<AdminWebhookInput>({ url: '', active: false })

watch(
  () => store.settings,
  (settings) => {
    errors.value = { url: undefined }
    if (settings) {
      form.value = { url: settings.url, active: settings.active }
    }
  },
  { immediate: true },
)

async function submit(): Promise<void> {
  errors.value = validateWebhookInput(form.value.url, form.value.active)
  if (errors.value.url) {
    return
  }
  await store.save({ url: form.value.url.trim(), active: form.value.active })
}

onMounted(() => {
  if (store.state === 'idle') {
    void store.load()
  }
})
</script>

<template>
  <main class="admin-page">
    <p class="admin-page__kicker">Painel administrativo · Configurações da loja</p>
    <h1>Integrações · Webhook de notificações</h1>
    <p class="admin-page__lead">
      A cada pedido criado ou atualizado, o servidor envia um POST com número e mensagem. O envio é
      assíncrono e não bloqueia o pedido. A URL e o corpo podem conter dados pessoais: configure
      apenas um destino confiável.
    </p>

    <LoadingSkeleton v-if="store.state === 'loading'" label="Carregando integração" :count="2" />

    <ErrorState v-else-if="store.state === 'error'" :message="store.error" @retry="store.load" />

    <EmptyState
      v-else-if="store.state === 'success' && !store.settings"
      title="Integração indisponível"
      description="Não foi possível carregar as configurações do webhook."
    />

    <form v-else-if="store.state === 'success'" class="webhook-form" @submit.prevent="submit">
      <label class="webhook-form__check">
        <input v-model="form.active" type="checkbox" />
        Webhook ativo
      </label>

      <div class="webhook-form__field">
        <label for="webhook-url">URL do webhook</label>
        <input
          id="webhook-url"
          v-model="form.url"
          type="text"
          maxlength="2048"
          placeholder="https://integracao.exemplo.com/pedidos"
          :aria-invalid="errors.url ? 'true' : undefined"
          :aria-describedby="errors.url ? 'webhook-url-error' : undefined"
        />
        <p v-if="errors.url" id="webhook-url-error" class="webhook-form__error" role="alert">
          {{ errors.url }}
        </p>
      </div>

      <p v-if="store.saveError" class="webhook-form__message" role="alert">{{ store.saveError }}</p>
      <p v-if="store.saveSuccess" class="webhook-form__message" role="status">
        {{ store.saveSuccess }}
      </p>

      <BaseButton type="submit" :loading="store.saving" :disabled="store.saving">Salvar</BaseButton>
    </form>
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

.webhook-form {
  display: grid;
  gap: 18px;
  padding: 18px;
  border: 1px solid var(--bg-border);
  border-radius: 18px;
  background: var(--bg-surface);
}

.webhook-form__check {
  display: flex;
  align-items: center;
  gap: 10px;
  min-height: 44px;
  font-weight: 700;
}

.webhook-form__field {
  display: grid;
  gap: 8px;
}

.webhook-form__field label {
  font-size: var(--font-size-caption);
  font-weight: 700;
}

.webhook-form input[type='text'] {
  min-height: 44px;
  padding: 10px 12px;
  border: 1px solid var(--bg-border);
  border-radius: 12px;
  background: var(--bg-input);
  color: var(--text-main);
  font: inherit;
}

.webhook-form__error,
.webhook-form__message {
  margin: 0;
  color: var(--status-error);
  font-size: var(--font-size-caption);
}

.webhook-form__message[role='status'] {
  color: var(--status-open);
}
</style>
