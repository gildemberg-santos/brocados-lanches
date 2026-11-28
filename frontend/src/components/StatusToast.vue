<script setup lang="ts">
import { onMounted, onUnmounted, watch } from 'vue'

const VISIBLE_MS = 4000

const props = defineProps<{ message: string }>()
const emit = defineEmits<{ close: [] }>()

let timer: ReturnType<typeof setTimeout> | null = null

function clearTimer(): void {
  if (timer !== null) {
    clearTimeout(timer)
    timer = null
  }
}

function scheduleClose(): void {
  clearTimer()
  timer = setTimeout(() => {
    timer = null
    emit('close')
  }, VISIBLE_MS)
}

onMounted(scheduleClose)
watch(() => props.message, scheduleClose)
onUnmounted(clearTimer)
</script>

<template>
  <aside class="status-toast" role="status" aria-live="polite" aria-atomic="true">
    <div class="status-toast__content">
      <p class="status-toast__title">Atualização do pedido</p>
      <p class="status-toast__message">{{ message }}</p>
    </div>
    <button
      type="button"
      class="status-toast__close"
      aria-label="Fechar notificação"
      @click="$emit('close')"
    >
      ×
    </button>
  </aside>
</template>

<style scoped>
.status-toast {
  position: fixed;
  right: 16px;
  bottom: calc(76px + env(safe-area-inset-bottom));
  z-index: 46;
  display: flex;
  width: min(360px, calc(100% - 32px));
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  padding: 16px;
  border: 1px solid var(--bg-border);
  border-left: 4px solid var(--brand-primary);
  border-radius: 12px;
  background: var(--bg-surface);
  color: var(--text-main);
}

.status-toast__title {
  margin: 0 0 4px;
  color: var(--brand-primary);
  font-size: var(--font-size-caption);
  font-weight: 800;
}

.status-toast__message {
  margin: 0;
  font-size: var(--font-size-body);
}

.status-toast__close {
  display: grid;
  width: 44px;
  height: 44px;
  flex: 0 0 auto;
  place-items: center;
  border: 1px solid var(--bg-border);
  border-radius: 999px;
  background: var(--bg-input);
  color: var(--text-main);
  font-family: var(--font-sans);
  font-size: var(--font-size-h2);
}

@media (min-width: 768px) {
  .status-toast {
    bottom: 16px;
  }
}
</style>
