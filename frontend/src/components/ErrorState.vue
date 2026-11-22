<script setup lang="ts">
import BaseButton from './BaseButton.vue'

withDefaults(
  defineProps<{
    message: string
    retryLabel?: string
    title?: string
  }>(),
  {
    retryLabel: 'Tentar novamente',
    title: 'Não foi possível carregar',
  },
)

defineEmits<{
  retry: []
}>()
</script>

<template>
  <section class="error-state" role="alert">
    <div class="error-state__mark" aria-hidden="true">!</div>
    <div class="error-state__content">
      <h2>{{ title }}</h2>
      <p>{{ message }}</p>
    </div>
    <BaseButton variant="secondary" @click="$emit('retry')">{{ retryLabel }}</BaseButton>
  </section>
</template>

<style scoped>
.error-state {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 16px;
  padding: 20px;
  border: 1px solid var(--status-error);
  border-radius: 18px;
  background: var(--bg-surface);
}

.error-state__mark {
  display: grid;
  width: 40px;
  height: 40px;
  flex: 0 0 auto;
  place-items: center;
  border: 1px solid var(--status-error);
  border-radius: 999px;
  color: var(--status-error);
  font-weight: 800;
}

.error-state__content {
  flex: 1 1 220px;
}

.error-state h2 {
  margin: 0;
  font-size: var(--font-size-h2);
}

.error-state p {
  margin: 4px 0 0;
  color: var(--text-sub);
  line-height: 1.5;
}
</style>
