<script setup lang="ts">
import { computed, onMounted } from 'vue'

import EmptyState from '../components/EmptyState.vue'
import ErrorState from '../components/ErrorState.vue'
import LoadingSkeleton from '../components/LoadingSkeleton.vue'
import { useStoreStatusStore } from '../stores/storeStatus'
import type { AdminWeekday } from '../types/adminCatalog'

const store = useStoreStatusStore()
const status = computed(() => store.status)
const isLoading = computed(() => store.state === 'idle' || store.state === 'loading')
const statusLabel = computed(() =>
  status.value?.status === 'open' ? 'Aberta no momento' : 'Fechada no momento',
)

const WEEKDAY_LABELS: Record<AdminWeekday, string> = {
  sunday: 'Domingo',
  monday: 'Segunda-feira',
  tuesday: 'Terça-feira',
  wednesday: 'Quarta-feira',
  thursday: 'Quinta-feira',
  friday: 'Sexta-feira',
  saturday: 'Sábado',
}

const hasPublishedChannels = computed(() => {
  const current = status.value
  if (!current) {
    return false
  }
  return (
    [current.whatsapp, current.instagram, current.address, current.deliveryTime].some(
      (value) => value.trim().length > 0,
    ) || current.openingHours.length > 0
  )
})

function formatHours(hour: {
  day: AdminWeekday
  closed: boolean
  intervals: { open: string; close: string }[]
}): string {
  const label = WEEKDAY_LABELS[hour.day]
  const intervals = hour.intervals
    .map((interval) => `${interval.open}–${interval.close}`)
    .join(', ')
  return `${label}: ${hour.closed || intervals.length === 0 ? 'Fechado' : intervals}`
}

function whatsappHref(value: string): string {
  const digits = value.replace(/\D/g, '')
  return digits.length > 0 ? `https://wa.me/${digits}` : ''
}

function instagramHref(value: string): string {
  return /^https?:\/\//i.test(value) ? value : ''
}

function retry(): void {
  void store.refresh()
}

onMounted(() => {
  void store.load()
})
</script>

<template>
  <main id="conteudo" tabindex="-1" class="page-shell contact-page">
    <h1>Contato</h1>
    <p class="contact-page__lead">Estes dados vêm do status publicado pela loja no servidor.</p>

    <LoadingSkeleton v-if="isLoading" label="Carregando contato" :count="3" />

    <ErrorState v-else-if="store.state === 'error'" :message="store.error" @retry="retry" />

    <template v-else-if="status">
      <section class="contact-page__summary" aria-labelledby="contact-status-title">
        <h2 id="contact-status-title">{{ status.name }}</h2>
        <p>{{ statusLabel }}</p>
        <p v-if="status.deliveryTime">Tempo de entrega informado: {{ status.deliveryTime }}</p>
      </section>

      <EmptyState
        v-if="!hasPublishedChannels"
        title="Contato indisponível"
        description="A loja ainda não publicou canais de contato. Tente novamente em instantes."
      />

      <dl v-else class="contact-page__list">
        <div v-if="whatsappHref(status.whatsapp)" class="contact-page__item">
          <dt>WhatsApp</dt>
          <dd>
            <span>{{ status.whatsapp }}</span>
            <a :href="whatsappHref(status.whatsapp)" target="_blank" rel="noopener noreferrer">
              Abrir WhatsApp
            </a>
          </dd>
        </div>

        <div v-if="status.instagram" class="contact-page__item">
          <dt>Instagram</dt>
          <dd>
            <span>{{ status.instagram }}</span>
            <a
              v-if="instagramHref(status.instagram)"
              :href="instagramHref(status.instagram)"
              target="_blank"
              rel="noopener noreferrer"
            >
              Abrir Instagram
            </a>
          </dd>
        </div>

        <div v-if="status.address" class="contact-page__item">
          <dt>Endereço</dt>
          <dd>{{ status.address }}</dd>
        </div>

        <div v-if="status.openingHours.length > 0" class="contact-page__item">
          <dt>Horários</dt>
          <dd>
            <ul>
              <li v-for="hour in status.openingHours" :key="hour.day">{{ formatHours(hour) }}</li>
            </ul>
          </dd>
        </div>
      </dl>
    </template>
  </main>
</template>

<style scoped>
.contact-page {
  display: grid;
  gap: 20px;
}

.contact-page h1 {
  margin: 0;
  font-family: var(--font-display);
  font-size: clamp(var(--font-size-h1), 7vw, var(--font-size-display));
  line-height: var(--line-height-display);
}

.contact-page__lead,
.contact-page__summary p,
.contact-page__item dd {
  margin: 0;
  color: var(--text-sub);
  line-height: 1.6;
}

.contact-page__summary,
.contact-page__item {
  display: grid;
  gap: 6px;
  padding: 16px;
  border: 1px solid var(--bg-border);
  border-radius: 16px;
  background: var(--bg-surface);
}

.contact-page__summary h2 {
  margin: 0;
  font-size: var(--font-size-h2);
}

.contact-page__list {
  display: grid;
  gap: 12px;
  margin: 0;
}

.contact-page dt {
  color: var(--brand-primary);
  font-size: var(--font-size-caption);
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.contact-page__item dd {
  display: grid;
  gap: 4px;
}

.contact-page__item ul {
  display: grid;
  gap: 4px;
  margin: 0;
  padding-left: 20px;
}

.contact-page__item a {
  display: inline-flex;
  min-height: 44px;
  align-items: center;
  color: var(--brand-neon);
  font-weight: 700;
}
</style>
