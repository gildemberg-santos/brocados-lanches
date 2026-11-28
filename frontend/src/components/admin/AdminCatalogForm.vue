<script setup lang="ts">
import { ref, watch } from 'vue'

import BaseButton from '../BaseButton.vue'
import {
  moneyToCents,
  slugify,
  validateSimpleInput,
} from '../../features/admin-catalog/catalogForm'
import type { FieldErrors } from '../../features/admin-catalog/catalogForm'
import type {
  AdminAddon,
  AdminCatalogInput,
  AdminCatalogKind,
  AdminCategory,
  AdminZone,
} from '../../types/adminCatalog'

type CatalogItem = AdminCategory | AdminAddon | AdminZone

const props = defineProps<{
  kind: AdminCatalogKind
  item: CatalogItem | null
  saving: boolean
  error: string
  success: string
}>()

const emit = defineEmits<{ submit: [value: AdminCatalogInput]; cancel: [] }>()

const name = ref('')
const slug = ref('')
const description = ref('')
const amount = ref('')
const sortOrder = ref(0)
const active = ref(true)
const errors = ref<FieldErrors>({})

function errorKey(): string {
  return props.kind === 'addon' ? 'priceCents' : 'feeCents'
}

function errorId(key: string): string {
  return `catalog-${key.replace(/\./g, '-')}-error`
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

function centsToInput(cents: number): string {
  return (cents / 100).toFixed(2).replace('.', ',')
}

watch(
  () => props.item,
  (item) => {
    errors.value = {}
    name.value = item?.name ?? ''
    slug.value = item?.slug ?? ''
    description.value = item && 'description' in item ? item.description : ''
    amount.value =
      item && 'priceCents' in item
        ? centsToInput(item.priceCents)
        : item && 'feeCents' in item
          ? centsToInput(item.feeCents)
          : ''
    sortOrder.value = item?.sortOrder ?? 0
    active.value = item?.active ?? true
  },
  { immediate: true },
)

function submit(): void {
  const cents = moneyToCents(amount.value)
  if (props.kind !== 'category' && cents === null) {
    errors.value = {
      [props.kind === 'addon' ? 'priceCents' : 'feeCents']: 'Informe um valor válido em reais.',
    }
    return
  }

  const input: AdminCatalogInput =
    props.kind === 'category'
      ? {
          name: name.value.trim(),
          slug: slug.value.trim(),
          description: description.value.trim(),
          sortOrder: Number(sortOrder.value),
          active: active.value,
        }
      : props.kind === 'addon'
        ? {
            name: name.value.trim(),
            slug: slug.value.trim(),
            priceCents: cents ?? 0,
            sortOrder: Number(sortOrder.value),
            active: active.value,
          }
        : {
            name: name.value.trim(),
            slug: slug.value.trim(),
            feeCents: cents ?? 0,
            sortOrder: Number(sortOrder.value),
            active: active.value,
          }

  errors.value = validateSimpleInput(props.kind, input)
  if (Object.keys(errors.value).length > 0) {
    return
  }
  emit('submit', input)
}
</script>

<template>
  <form class="catalog-form" aria-label="Formulário de cadastro" @submit.prevent="submit">
    <div class="catalog-form__field">
      <label for="catalog-name">Nome</label>
      <input
        id="catalog-name"
        v-model="name"
        type="text"
        maxlength="120"
        v-bind="errorAttrs('name')"
      />
      <p v-if="errorFor('name')" :id="errorId('name')" class="catalog-form__error" role="alert">
        {{ errorFor('name') }}
      </p>
    </div>

    <div class="catalog-form__field">
      <label for="catalog-slug">Slug</label>
      <div class="catalog-form__slug">
        <input
          id="catalog-slug"
          v-model="slug"
          type="text"
          maxlength="80"
          v-bind="errorAttrs('slug')"
        />
        <button type="button" class="catalog-form__slug-button" @click="slug = slugify(name)">
          Gerar
        </button>
      </div>
      <p v-if="errorFor('slug')" :id="errorId('slug')" class="catalog-form__error" role="alert">
        {{ errorFor('slug') }}
      </p>
    </div>

    <div v-if="props.kind === 'category'" class="catalog-form__field">
      <label for="catalog-description">Descrição</label>
      <textarea
        id="catalog-description"
        v-model="description"
        rows="3"
        maxlength="2000"
        v-bind="errorAttrs('description')"
      />
      <p
        v-if="errorFor('description')"
        :id="errorId('description')"
        class="catalog-form__error"
        role="alert"
      >
        {{ errorFor('description') }}
      </p>
    </div>

    <div v-if="props.kind !== 'category'" class="catalog-form__field">
      <label for="catalog-amount">{{ props.kind === 'addon' ? 'Preço (R$)' : 'Taxa (R$)' }}</label>
      <input
        id="catalog-amount"
        v-model="amount"
        type="text"
        inputmode="decimal"
        placeholder="0,00"
        v-bind="errorAttrs(errorKey())"
      />
      <p
        v-if="errorFor(errorKey())"
        :id="errorId(errorKey())"
        class="catalog-form__error"
        role="alert"
      >
        {{ errorFor(errorKey()) }}
      </p>
    </div>

    <div class="catalog-form__field">
      <label for="catalog-sort">Ordem</label>
      <input
        id="catalog-sort"
        v-model.number="sortOrder"
        type="number"
        min="0"
        v-bind="errorAttrs('sortOrder')"
      />
      <p
        v-if="errorFor('sortOrder')"
        :id="errorId('sortOrder')"
        class="catalog-form__error"
        role="alert"
      >
        {{ errorFor('sortOrder') }}
      </p>
    </div>

    <label class="catalog-form__check">
      <input v-model="active" type="checkbox" />
      Ativo
    </label>

    <p v-if="props.error" class="catalog-form__message" role="alert">{{ props.error }}</p>
    <p v-if="props.success" class="catalog-form__message" role="status">{{ props.success }}</p>

    <div class="catalog-form__actions">
      <BaseButton type="submit" :loading="props.saving" :disabled="props.saving">
        Salvar
      </BaseButton>
      <BaseButton type="button" variant="ghost" @click="emit('cancel')">Cancelar</BaseButton>
    </div>
  </form>
</template>

<style scoped>
.catalog-form {
  display: grid;
  gap: 14px;
  padding: 18px;
  border: 1px solid var(--bg-border);
  border-radius: 18px;
  background: var(--bg-surface);
}

.catalog-form__field {
  display: grid;
  gap: 8px;
}

.catalog-form__field label {
  font-size: var(--font-size-caption);
  font-weight: 700;
}

.catalog-form input:not([type='checkbox']),
.catalog-form textarea,
.catalog-form__slug-button {
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

.catalog-form__slug {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 8px;
}

.catalog-form__slug-button {
  width: auto;
  font-weight: 700;
}

.catalog-form__check {
  display: flex;
  align-items: center;
  gap: 10px;
  min-height: 44px;
  font-size: var(--font-size-body);
  font-weight: 700;
}

.catalog-form__error,
.catalog-form__message {
  margin: 0;
  color: var(--status-error);
  font-size: var(--font-size-caption);
}

.catalog-form__message {
  padding: 10px 12px;
  border: 1px solid var(--bg-border);
  border-radius: 12px;
  color: var(--text-main);
}

.catalog-form__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
</style>
