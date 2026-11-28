<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'

import BaseButton from '../../components/BaseButton.vue'
import EmptyState from '../../components/EmptyState.vue'
import ErrorState from '../../components/ErrorState.vue'
import LoadingSkeleton from '../../components/LoadingSkeleton.vue'
import { validatePixInput } from '../../features/admin/pixForm'
import type { PixFieldErrors } from '../../features/admin/pixForm'
import { useAdminPixStore } from '../../stores/adminPix'
import type { AdminPixInput } from '../../types/adminPix'

const store = useAdminPixStore()
const errors = ref<PixFieldErrors>({})
const form = ref<AdminPixInput>({
  active: false,
  key: '',
  merchantName: '',
  merchantCity: '',
})

watch(
  () => store.settings,
  (settings) => {
    errors.value = {}
    if (!settings) {
      return
    }
    form.value = {
      active: settings.active,
      key: settings.key,
      merchantName: settings.merchantName,
      merchantCity: settings.merchantCity,
    }
  },
  { immediate: true },
)

function errorId(key: string): string {
  return `pix-${key}-error`
}

function errorAttrs(key: keyof AdminPixInput) {
  const message = errors.value[key]
  return {
    'aria-invalid': message ? ('true' as const) : undefined,
    'aria-describedby': message ? errorId(key) : undefined,
  }
}

async function submit(): Promise<void> {
  errors.value = validatePixInput(form.value)
  if (Object.keys(errors.value).length > 0) {
    return
  }
  await store.save({ ...form.value })
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
    <h1>Pix estático</h1>
    <p class="admin-page__lead">
      Após gravar o pedido, a confirmação gera localmente um BR Code com o total do pedido. Sem
      integração com PSP, o pagamento não tem confirmação automática: confira o comprovante ou o
      extrato antes de confirmar o pedido. Use uma chave aleatória (EVP) para não expor CPF, e-mail
      ou telefone.
    </p>

    <LoadingSkeleton v-if="store.state === 'loading'" label="Carregando Pix" :count="2" />

    <ErrorState v-else-if="store.state === 'error'" :message="store.error" @retry="store.load" />

    <EmptyState
      v-else-if="store.state === 'success' && !store.settings"
      title="Configuração indisponível"
      description="Não foi possível carregar os dados do Pix."
    />

    <form v-else-if="store.state === 'success'" class="pix-form" @submit.prevent="submit">
      <label class="pix-form__check">
        <input v-model="form.active" type="checkbox" />
        Oferecer Pix no checkout
      </label>

      <div class="pix-form__field">
        <label for="pix-key">Chave Pix</label>
        <input
          id="pix-key"
          v-model="form.key"
          type="text"
          maxlength="77"
          placeholder="EVP, CPF, CNPJ, e-mail ou telefone E.164"
          v-bind="errorAttrs('key')"
        />
        <p v-if="errors.key" :id="errorId('key')" class="pix-form__error" role="alert">
          {{ errors.key }}
        </p>
      </div>

      <div class="pix-form__grid">
        <div class="pix-form__field">
          <label for="pix-name">Nome do recebedor</label>
          <input
            id="pix-name"
            v-model="form.merchantName"
            type="text"
            v-bind="errorAttrs('merchantName')"
          />
          <p
            v-if="errors.merchantName"
            :id="errorId('merchantName')"
            class="pix-form__error"
            role="alert"
          >
            {{ errors.merchantName }}
          </p>
        </div>

        <div class="pix-form__field">
          <label for="pix-city">Cidade do recebedor</label>
          <input
            id="pix-city"
            v-model="form.merchantCity"
            type="text"
            v-bind="errorAttrs('merchantCity')"
          />
          <p
            v-if="errors.merchantCity"
            :id="errorId('merchantCity')"
            class="pix-form__error"
            role="alert"
          >
            {{ errors.merchantCity }}
          </p>
        </div>
      </div>

      <p class="pix-form__hint" role="status">
        O nome e a cidade são higienizados no BR Code (sem acentos ou símbolos). O TXID permanece
        <code>***</code>: nesta fase o QR não identifica pedido nem permite conciliação automática.
      </p>

      <p v-if="store.saveError" class="pix-form__message" role="alert">{{ store.saveError }}</p>
      <p v-if="store.saveSuccess" class="pix-form__message" role="status">
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

.pix-form {
  display: grid;
  gap: 18px;
  padding: 18px;
  border: 1px solid var(--bg-border);
  border-radius: 18px;
  background: var(--bg-surface);
}

.pix-form__check {
  display: flex;
  align-items: center;
  gap: 10px;
  min-height: 44px;
  font-weight: 700;
}

.pix-form__grid {
  display: grid;
  gap: 14px;
}

.pix-form__field {
  display: grid;
  gap: 8px;
}

.pix-form__field label {
  font-size: var(--font-size-caption);
  font-weight: 700;
}

.pix-form input:not([type='checkbox']) {
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

.pix-form__error {
  margin: 0;
  color: var(--status-error);
  font-size: var(--font-size-caption);
}

.pix-form__hint,
.pix-form__message {
  margin: 0;
  color: var(--text-sub);
  font-size: var(--font-size-caption);
  line-height: 1.5;
}

@media (min-width: 768px) {
  .pix-form__grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
