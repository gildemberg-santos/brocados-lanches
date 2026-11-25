<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'

import BaseButton from '../../components/BaseButton.vue'
import EmptyState from '../../components/EmptyState.vue'
import ErrorState from '../../components/ErrorState.vue'
import LoadingSkeleton from '../../components/LoadingSkeleton.vue'
import {
  emptyInterval,
  emptyOpeningHours,
  validateStoreInput,
  WEEKDAY_LABELS,
} from '../../features/admin-catalog/catalogForm'
import type { FieldErrors } from '../../features/admin-catalog/catalogForm'
import { useAdminStoreSettingsStore } from '../../stores/adminStoreSettings'
import { useStoreStatusStore } from '../../stores/storeStatus'
import type { AdminStoreInput } from '../../types/adminCatalog'

const store = useAdminStoreSettingsStore()
const statusStore = useStoreStatusStore()
const errors = ref<FieldErrors>({})

function errorId(key: string): string {
  return `store-${key.replace(/\./g, '-')}-error`
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

const form = ref<AdminStoreInput>({
  name: '',
  isOpen: true,
  openingHours: emptyOpeningHours(),
  deliveryTime: '',
  whatsapp: '',
  instagram: '',
  address: '',
})

watch(
  () => store.settings,
  (settings) => {
    errors.value = {}
    if (!settings) {
      return
    }
    form.value = {
      name: settings.name,
      isOpen: settings.isOpen,
      openingHours:
        settings.openingHours.length > 0
          ? settings.openingHours.map((hour) => ({
              ...hour,
              intervals: hour.intervals.map((interval) => ({ ...interval })),
            }))
          : emptyOpeningHours(),
      deliveryTime: settings.deliveryTime,
      whatsapp: settings.whatsapp,
      instagram: settings.instagram,
      address: settings.address,
    }
  },
  { immediate: true },
)

function addInterval(dayIndex: number): void {
  form.value.openingHours[dayIndex]?.intervals.push(emptyInterval())
}

function removeInterval(dayIndex: number, intervalIndex: number): void {
  form.value.openingHours[dayIndex]?.intervals.splice(intervalIndex, 1)
}

function formatNextOpening(value: string): string {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return 'a confirmar'
  }
  return new Intl.DateTimeFormat('pt-BR', {
    dateStyle: 'short',
    timeStyle: 'short',
    timeZone: 'America/Fortaleza',
  }).format(date)
}

async function submit(): Promise<void> {
  errors.value = validateStoreInput(form.value)
  if (Object.keys(errors.value).length > 0) {
    return
  }
  if (await store.save({ ...form.value })) {
    await statusStore.refresh()
  }
}

onMounted(() => {
  if (store.state === 'idle') {
    void store.load()
  }
  void statusStore.load()
})
</script>

<template>
  <main class="admin-page">
    <p class="admin-page__kicker">Painel administrativo</p>
    <h1>Loja</h1>
    <p class="admin-page__lead">
      O interruptor manual e os horários precisam permitir o pedido. A operação usa
      <strong>America/Fortaleza</strong>; alterações de taxa não afetam pedidos antigos.
    </p>

    <div
      v-if="statusStore.state === 'success' && statusStore.status"
      class="store-status"
      data-testid="store-effective-status"
    >
      <strong :class="statusStore.status.status === 'open' ? 'is-open' : 'is-closed'">
        {{ statusStore.status.status === 'open' ? 'Loja aberta' : 'Loja fechada' }}
      </strong>
      <span>Status calculado pelo servidor em {{ statusStore.status.timezone }}.</span>
      <span v-if="statusStore.status.nextOpeningAt">
        Próxima abertura: {{ formatNextOpening(statusStore.status.nextOpeningAt) }}
      </span>
    </div>

    <LoadingSkeleton v-if="store.state === 'loading'" label="Carregando configurações" :count="2" />

    <ErrorState v-else-if="store.state === 'error'" :message="store.error" @retry="store.load" />

    <EmptyState
      v-else-if="store.state === 'success' && !store.settings"
      title="Configurações indisponíveis"
      description="Não foi possível carregar os dados da loja."
    />

    <form v-else-if="store.state === 'success'" class="store-form" @submit.prevent="submit">
      <label class="store-form__check store-form__check--highlight">
        <input v-model="form.isOpen" type="checkbox" />
        Loja aberta para pedidos (interruptor manual)
      </label>

      <div class="store-form__grid">
        <div class="store-form__field">
          <label for="store-name">Nome da loja</label>
          <input
            id="store-name"
            v-model="form.name"
            type="text"
            maxlength="120"
            v-bind="errorAttrs('name')"
          />
          <p v-if="errorFor('name')" :id="errorId('name')" class="store-form__error" role="alert">
            {{ errorFor('name') }}
          </p>
        </div>

        <div class="store-form__field">
          <label for="store-delivery">Prazo de entrega</label>
          <input
            id="store-delivery"
            v-model="form.deliveryTime"
            type="text"
            maxlength="60"
            v-bind="errorAttrs('deliveryTime')"
          />
          <p
            v-if="errorFor('deliveryTime')"
            :id="errorId('deliveryTime')"
            class="store-form__error"
            role="alert"
          >
            {{ errorFor('deliveryTime') }}
          </p>
        </div>

        <div class="store-form__field">
          <label for="store-whatsapp">WhatsApp (E.164)</label>
          <input
            id="store-whatsapp"
            v-model="form.whatsapp"
            type="tel"
            placeholder="+5585999999999"
            v-bind="errorAttrs('whatsapp')"
          />
          <p
            v-if="errorFor('whatsapp')"
            :id="errorId('whatsapp')"
            class="store-form__error"
            role="alert"
          >
            {{ errorFor('whatsapp') }}
          </p>
        </div>

        <div class="store-form__field">
          <label for="store-instagram">Instagram</label>
          <input id="store-instagram" v-model="form.instagram" type="text" maxlength="100" />
        </div>

        <div class="store-form__field store-form__field--wide">
          <label for="store-address">Endereço</label>
          <input
            id="store-address"
            v-model="form.address"
            type="text"
            maxlength="300"
            v-bind="errorAttrs('address')"
          />
          <p
            v-if="errorFor('address')"
            :id="errorId('address')"
            class="store-form__error"
            role="alert"
          >
            {{ errorFor('address') }}
          </p>
        </div>
      </div>

      <fieldset class="store-form__section">
        <legend>Horários de funcionamento</legend>
        <div v-for="(hour, dayIndex) in form.openingHours" :key="hour.day" class="store-form__day">
          <strong>{{ WEEKDAY_LABELS[hour.day] }}</strong>
          <label class="store-form__check">
            <input v-model="hour.closed" type="checkbox" />
            Fechado
          </label>

          <template v-if="!hour.closed">
            <div
              v-for="(interval, intervalIndex) in hour.intervals"
              :key="intervalIndex"
              class="store-form__interval"
            >
              <label>
                Abertura
                <input
                  :id="`store-${dayIndex}-${intervalIndex}-open`"
                  v-model="interval.open"
                  type="time"
                  v-bind="errorAttrs(`openingHours.${dayIndex}.intervals.${intervalIndex}.open`)"
                />
              </label>
              <label>
                Fechamento
                <input
                  :id="`store-${dayIndex}-${intervalIndex}-close`"
                  v-model="interval.close"
                  type="time"
                  v-bind="errorAttrs(`openingHours.${dayIndex}.intervals.${intervalIndex}.close`)"
                />
              </label>
              <button
                type="button"
                class="store-form__small"
                @click="removeInterval(dayIndex, intervalIndex)"
              >
                Remover
              </button>
              <p
                v-if="errorFor(`openingHours.${dayIndex}.intervals.${intervalIndex}.open`)"
                :id="errorId(`openingHours.${dayIndex}.intervals.${intervalIndex}.open`)"
                class="store-form__error"
                role="alert"
              >
                {{ errorFor(`openingHours.${dayIndex}.intervals.${intervalIndex}.open`) }}
              </p>
              <p
                v-if="errorFor(`openingHours.${dayIndex}.intervals.${intervalIndex}.close`)"
                :id="errorId(`openingHours.${dayIndex}.intervals.${intervalIndex}.close`)"
                class="store-form__error"
                role="alert"
              >
                {{ errorFor(`openingHours.${dayIndex}.intervals.${intervalIndex}.close`) }}
              </p>
            </div>

            <button type="button" class="store-form__small" @click="addInterval(dayIndex)">
              Adicionar intervalo
            </button>
            <p v-if="errors[`openingHours.${dayIndex}`]" class="store-form__error" role="alert">
              {{ errors[`openingHours.${dayIndex}`] }}
            </p>
          </template>
        </div>
      </fieldset>

      <p v-if="store.saveError" class="store-form__message" role="alert">{{ store.saveError }}</p>
      <p v-if="store.saveSuccess" class="store-form__message" role="status">
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

.store-status {
  display: grid;
  gap: 4px;
  padding: 12px;
  border: 1px solid var(--bg-border);
  border-radius: 12px;
  background: var(--bg-surface);
  font-size: var(--font-size-caption);
}

.store-status strong {
  font-size: var(--font-size-body);
}

.store-status .is-open {
  color: var(--status-open);
}

.store-status .is-closed {
  color: var(--status-error);
}

.store-form {
  display: grid;
  gap: 18px;
  padding: 18px;
  border: 1px solid var(--bg-border);
  border-radius: 18px;
  background: var(--bg-surface);
}

.store-form__grid {
  display: grid;
  gap: 14px;
}

.store-form__field {
  display: grid;
  gap: 8px;
}

.store-form__field label {
  font-size: var(--font-size-caption);
  font-weight: 700;
}

.store-form input,
.store-form__small {
  min-height: 44px;
  padding: 10px 12px;
  border: 1px solid var(--bg-border);
  border-radius: 12px;
  background: var(--bg-input);
  color: var(--text-main);
  font-family: var(--font-sans);
  font-size: var(--font-size-body);
}

.store-form__small {
  width: fit-content;
  font-weight: 700;
}

.store-form__check {
  display: flex;
  align-items: center;
  gap: 10px;
  min-height: 44px;
  font-weight: 700;
}

.store-form__check--highlight {
  padding: 12px;
  border: 1px solid var(--brand-primary);
  border-radius: 12px;
}

.store-form__section {
  display: grid;
  gap: 12px;
  padding: 16px;
  border: 1px solid var(--bg-border);
  border-radius: 18px;
}

.store-form__section legend {
  padding: 0 8px;
  font-weight: 800;
}

.store-form__day {
  display: grid;
  gap: 10px;
  padding: 12px;
  border: 1px solid var(--bg-border);
  border-radius: 14px;
}

.store-form__interval {
  display: grid;
  gap: 8px;
}

.store-form__interval label {
  display: grid;
  gap: 6px;
  font-size: var(--font-size-caption);
}

.store-form__error {
  margin: 0;
  color: var(--status-error);
  font-size: var(--font-size-caption);
}

.store-form__message {
  margin: 0;
  padding: 10px 12px;
  border: 1px solid var(--bg-border);
  border-radius: 12px;
  font-size: var(--font-size-caption);
}

@media (min-width: 768px) {
  .store-form__grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .store-form__field--wide {
    grid-column: 1 / -1;
  }

  .store-form__interval {
    grid-template-columns: 1fr 1fr auto;
    align-items: end;
  }
}
</style>
