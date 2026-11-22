<script setup lang="ts">
defineProps<{
  label: string
  active: boolean
}>()

defineEmits<{
  select: []
}>()
</script>

<template>
  <button
    class="category-pill"
    :class="{ 'category-pill--active': active }"
    type="button"
    :aria-pressed="active"
    @click="$emit('select')"
  >
    {{ label }}
  </button>
</template>

<style scoped>
.category-pill {
  min-height: 44px;
  padding: 10px 18px;
  border: 1px solid var(--bg-border);
  border-radius: 999px;
  background: var(--bg-surface);
  color: var(--text-sub);
  font-weight: 700;
  white-space: nowrap;
  cursor: pointer;
  transition:
    background-color 180ms ease,
    border-color 180ms ease,
    color 180ms ease;
}

.category-pill:hover {
  border-color: var(--brand-primary);
  color: var(--text-main);
}

.category-pill--active {
  border-color: var(--brand-primary);
  background: var(--brand-primary);
  color: var(--bg-primary);
}

@media (prefers-reduced-motion: reduce) {
  .category-pill {
    transition: none;
  }
}
</style>
