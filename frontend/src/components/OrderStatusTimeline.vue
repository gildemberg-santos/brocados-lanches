<script setup lang="ts">
import { computed } from 'vue'

import { formatOrderDateTime } from '../features/admin-orders/format'
import { orderStatusLabel } from '../features/checkout/checkout'
import type { Fulfillment } from '../types/order'

const props = defineProps<{
  status: string
  fulfillment: Fulfillment
  createdAt: string
  updatedAt: string
}>()

type StepState = 'done' | 'current' | 'pending'

interface TimelineStep {
  key: string
  label: string
  state: StepState
  time: string
  dateTime: string
}

const DELIVERY_STEPS = [
  'pending',
  'confirmed',
  'preparing',
  'ready',
  'out_for_delivery',
  'delivered',
]
const PICKUP_STEPS = ['pending', 'confirmed', 'preparing', 'ready', 'picked_up']

function stepFor(key: string, state: StepState, dateTime: string): TimelineStep {
  return {
    key,
    label: key === 'pending' ? 'Pedido registrado' : orderStatusLabel(key),
    state,
    time: dateTime ? formatOrderDateTime(dateTime) : '',
    dateTime,
  }
}

const steps = computed<TimelineStep[]>(() => {
  const sequence = props.fulfillment === 'pickup' ? PICKUP_STEPS : DELIVERY_STEPS
  const currentIndex = sequence.indexOf(props.status)

  // Cancelamento e status fora da sequência não permitem inventar etapas.
  if (currentIndex < 0) {
    return [stepFor(props.status, 'current', props.updatedAt)]
  }

  return sequence.map((key, index) => {
    const state: StepState =
      index < currentIndex ? 'done' : index === currentIndex ? 'current' : 'pending'
    const dateTime = index === 0 ? props.createdAt : state === 'current' ? props.updatedAt : ''
    return stepFor(key, state, dateTime)
  })
})
</script>

<template>
  <section class="timeline" aria-labelledby="timeline-title">
    <h2 id="timeline-title">Andamento do pedido</h2>
    <ol class="timeline__list">
      <li
        v-for="step in steps"
        :key="step.key"
        class="timeline__step"
        :class="`is-${step.state}`"
        :aria-current="step.state === 'current' ? 'step' : undefined"
      >
        <span class="timeline__label">{{ step.label }}</span>
        <time v-if="step.time" :datetime="step.dateTime">{{ step.time }}</time>
        <span v-else-if="step.state === 'pending'" class="timeline__hint">Aguardando</span>
      </li>
    </ol>
  </section>
</template>

<style scoped>
.timeline {
  display: grid;
  gap: 10px;
}

.timeline h2 {
  margin: 0;
  font-size: var(--font-size-h2);
}

.timeline__list {
  display: grid;
  gap: 12px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.timeline__step {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 12px;
  padding-left: 14px;
  border-left: 2px solid var(--bg-border);
  font-size: var(--font-size-caption);
}

.timeline__step.is-done {
  border-left-color: var(--status-open);
}

.timeline__step.is-current {
  border-left-color: var(--brand-primary);
  color: var(--text-main);
  font-weight: 800;
}

.timeline__step.is-pending {
  color: var(--text-sub);
}

.timeline__hint {
  color: var(--text-sub);
}
</style>
