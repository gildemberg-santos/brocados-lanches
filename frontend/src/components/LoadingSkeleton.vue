<script setup lang="ts">
withDefaults(
  defineProps<{
    label?: string
    count?: number
  }>(),
  {
    label: 'Carregando cardápio',
    count: 6,
  },
)
</script>

<template>
  <div class="loading-skeleton" role="status" aria-live="polite">
    <span class="loading-skeleton__label">{{ label }}</span>
    <div class="loading-skeleton__grid" aria-hidden="true">
      <div v-for="index in count" :key="index" class="loading-skeleton__card">
        <div class="loading-skeleton__media"></div>
        <div class="loading-skeleton__line loading-skeleton__line--title"></div>
        <div class="loading-skeleton__line"></div>
        <div class="loading-skeleton__line loading-skeleton__line--short"></div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.loading-skeleton {
  display: grid;
  gap: 14px;
}

.loading-skeleton__label {
  color: var(--text-sub);
  font-size: var(--font-size-caption);
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.loading-skeleton__grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.loading-skeleton__card {
  display: grid;
  gap: 10px;
  padding: 10px;
  border: 1px solid var(--bg-border);
  border-radius: 16px;
  background: var(--bg-surface);
}

.loading-skeleton__media,
.loading-skeleton__line {
  background: linear-gradient(90deg, var(--bg-input), var(--bg-border), var(--bg-input));
  background-size: 200% 100%;
  animation: skeleton-pulse 1400ms ease-in-out infinite;
}

.loading-skeleton__media {
  aspect-ratio: 4 / 3;
  border-radius: 10px;
}

.loading-skeleton__line {
  height: 12px;
  border-radius: 999px;
}

.loading-skeleton__line--title {
  width: 72%;
  height: 16px;
}

.loading-skeleton__line--short {
  width: 46%;
}

@keyframes skeleton-pulse {
  0% {
    background-position: 100% 0;
  }

  100% {
    background-position: -100% 0;
  }
}

@media (min-width: 900px) {
  .loading-skeleton__grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}

@media (prefers-reduced-motion: reduce) {
  .loading-skeleton__media,
  .loading-skeleton__line {
    animation: none;
  }
}
</style>
