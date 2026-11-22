<script setup lang="ts">
import { computed, ref, watch } from 'vue'

import BaseButton from '../BaseButton.vue'
import {
  ORDER_ACTION_LABELS,
  availableTransitions,
  isCancellationAllowed,
  validateStatusChange,
} from '../../features/admin-orders/orderStatus'
import type { AdminOrderStatusRequest } from '../../types/order'

const props = withDefaults(
  defineProps<{
    status: string
    fulfillment: 'delivery' | 'pickup'
    loading?: boolean
  }>(),
  { loading: false },
)

const emit = defineEmits<{
  change: [request: AdminOrderStatusRequest]
}>()

const cancelling = ref(false)
const reason = ref('')
const reasonError = ref('')

const transitions = computed(() => availableTransitions(props.status, props.fulfillment))
const canCancel = computed(() => isCancellationAllowed(props.status))
const hasActions = computed(() => transitions.value.length > 0)

watch(
  () => props.status,
  () => {
    cancelling.value = false
    reason.value = ''
    reasonError.value = ''
  },
)

function emitStatus(status: string, cancellationReason = ''): void {
  const result = validateStatusChange(status, cancellationReason)
  if (!result.valid) {
    reasonError.value = result.error
    return
  }
  reasonError.value = ''
  emit('change', result.request)
}

function confirmCancellation(): void {
  emitStatus('cancelled', reason.value)
}
</script>

<template>
  <section v-if="hasActions" class="status-actions" aria-labelledby="status-actions-title">
    <h2 id="status-actions-title">Atualizar status</h2>

    <div v-if="!cancelling" class="status-actions__buttons">
      <BaseButton
        v-for="next in transitions.filter((status) => status !== 'cancelled')"
        :key="next"
        variant="primary"
        :disabled="loading"
        @click="emitStatus(next)"
      >
        {{ ORDER_ACTION_LABELS[next] }}
      </BaseButton>
      <BaseButton v-if="canCancel" variant="ghost" :disabled="loading" @click="cancelling = true">
        {{ ORDER_ACTION_LABELS.cancelled }}
      </BaseButton>
    </div>

    <form v-else class="status-actions__cancel" @submit.prevent="confirmCancellation">
      <label for="cancel-reason">Motivo do cancelamento</label>
      <textarea
        id="cancel-reason"
        v-model="reason"
        rows="3"
        maxlength="280"
        :aria-invalid="reasonError ? 'true' : undefined"
        aria-describedby="cancel-reason-help cancel-reason-error"
      ></textarea>
      <p id="cancel-reason-help" class="status-actions__help">
        Use apenas informações operacionais; não registre dados pessoais.
      </p>
      <p v-if="reasonError" id="cancel-reason-error" class="status-actions__error" role="alert">
        {{ reasonError }}
      </p>
      <div class="status-actions__buttons">
        <BaseButton variant="primary" type="submit" :loading="loading"
          >Confirmar cancelamento</BaseButton
        >
        <BaseButton variant="secondary" :disabled="loading" @click="cancelling = false">
          Voltar
        </BaseButton>
      </div>
    </form>
  </section>
</template>

<style scoped>
.status-actions {
  display: grid;
  gap: 12px;
  padding: 20px;
  border: 1px solid var(--bg-border);
  border-radius: 18px;
  background: var(--bg-surface);
}

.status-actions h2 {
  margin: 0;
  font-size: var(--font-size-h2);
}

.status-actions__buttons {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.status-actions__cancel {
  display: grid;
  gap: 8px;
}

.status-actions__cancel label {
  font-size: var(--font-size-caption);
  font-weight: 800;
}

.status-actions__cancel textarea {
  width: 100%;
  min-height: 88px;
  padding: 12px;
  border: 1px solid var(--bg-border);
  border-radius: 12px;
  background: var(--bg-input);
  color: var(--text-main);
  font-family: var(--font-sans);
  font-size: var(--font-size-body);
}

.status-actions__help,
.status-actions__error {
  margin: 0;
  color: var(--text-sub);
  font-size: var(--font-size-caption);
}

.status-actions__error {
  color: var(--status-error);
  font-weight: 700;
}
</style>
