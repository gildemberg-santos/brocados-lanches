<script setup lang="ts">
import { ref, watch } from 'vue'

import BaseButton from '../BaseButton.vue'
import type { AdminUser, AdminUserCreateInput, AdminRole } from '../../types/adminUsers'

const props = defineProps<{
  item: AdminUser | null
  saving: boolean
  error: string
  success: string
}>()

const emit = defineEmits<{
  submit: [input: AdminUserCreateInput]
  cancel: []
}>()

const name = ref('')
const email = ref('')
const password = ref('')
const role = ref<AdminRole>('operator')
const errors = ref<Record<string, string>>({})

function errorId(key: string): string {
  return `admin-user-${key}-error`
}

function errorFor(key: string): string {
  return errors.value[key] ?? ''
}

function errorAttrs(key: string) {
  const message = errorFor(key)
  return {
    'aria-invalid': message ? ('true' as const) : undefined,
    'aria-describedby': message ? errorId(key) : undefined,
  }
}

watch(
  () => props.item,
  (item) => {
    name.value = item?.name ?? ''
    email.value = item?.email ?? ''
    password.value = ''
    role.value = item?.role ?? 'operator'
    errors.value = {}
  },
  { immediate: true },
)

function submit(): void {
  const nextErrors: Record<string, string> = {}
  if (!name.value.trim()) {
    nextErrors.name = 'Informe o nome.'
  }
  if (!email.value.trim()) {
    nextErrors.email = 'Informe o e-mail.'
  }
  if (!props.item && password.value.length < 12) {
    nextErrors.password = 'A senha temporária precisa ter ao menos 12 caracteres.'
  }
  errors.value = nextErrors
  if (Object.keys(nextErrors).length > 0) {
    return
  }
  emit('submit', {
    name: name.value.trim(),
    email: email.value.trim(),
    password: password.value,
    role: role.value,
  })
}
</script>

<template>
  <form class="user-form" @submit.prevent="submit">
    <div class="user-form__field">
      <label for="admin-user-name">Nome</label>
      <input
        id="admin-user-name"
        v-model="name"
        type="text"
        maxlength="120"
        v-bind="errorAttrs('name')"
      />
      <p v-if="errorFor('name')" :id="errorId('name')" class="user-form__error" role="alert">
        {{ errorFor('name') }}
      </p>
    </div>

    <div class="user-form__field">
      <label for="admin-user-email">E-mail</label>
      <input
        id="admin-user-email"
        v-model="email"
        type="email"
        maxlength="254"
        v-bind="errorAttrs('email')"
      />
      <p v-if="errorFor('email')" :id="errorId('email')" class="user-form__error" role="alert">
        {{ errorFor('email') }}
      </p>
    </div>

    <div v-if="!item" class="user-form__field">
      <label for="admin-user-password">Senha temporária</label>
      <input
        id="admin-user-password"
        v-model="password"
        type="password"
        minlength="12"
        v-bind="errorAttrs('password')"
      />
      <p
        v-if="errorFor('password')"
        :id="errorId('password')"
        class="user-form__error"
        role="alert"
      >
        {{ errorFor('password') }}
      </p>
    </div>

    <div class="user-form__field">
      <label for="admin-user-role">Papel</label>
      <select id="admin-user-role" v-model="role">
        <option value="operator">Operador</option>
        <option value="owner">Dono</option>
      </select>
    </div>

    <p v-if="error" class="user-form__error" role="alert">{{ error }}</p>
    <p v-if="success" class="user-form__success" role="status">{{ success }}</p>

    <div class="user-form__actions">
      <BaseButton type="submit" :loading="saving" :disabled="saving">Salvar</BaseButton>
      <button type="button" class="user-form__cancel" @click="emit('cancel')">Cancelar</button>
    </div>
  </form>
</template>

<style scoped>
.user-form {
  display: grid;
  gap: 14px;
  padding: 18px;
  border: 1px solid var(--bg-border);
  border-radius: 18px;
  background: var(--bg-surface);
}

.user-form__field {
  display: grid;
  gap: 8px;
}

.user-form__field label {
  font-size: var(--font-size-caption);
  font-weight: 700;
}

.user-form input,
.user-form select,
.user-form__cancel {
  min-height: 44px;
  padding: 10px 12px;
  border: 1px solid var(--bg-border);
  border-radius: 12px;
  background: var(--bg-input);
  color: var(--text-main);
  font-family: var(--font-sans);
  font-size: var(--font-size-body);
}

.user-form__error {
  margin: 0;
  color: var(--status-error);
  font-size: var(--font-size-caption);
}

.user-form__success {
  margin: 0;
  color: var(--status-open);
  font-size: var(--font-size-caption);
}

.user-form__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.user-form__cancel {
  width: fit-content;
  font-weight: 700;
}
</style>
