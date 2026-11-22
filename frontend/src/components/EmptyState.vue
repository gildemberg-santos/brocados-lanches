<script setup lang="ts">
defineProps<{
  title: string
  description: string
}>()
</script>

<template>
  <section class="empty-state" role="status">
    <div class="empty-state__mark" aria-hidden="true">—</div>
    <h2>{{ title }}</h2>
    <p>{{ description }}</p>
    <div v-if="$slots.actions" class="empty-state__actions">
      <slot name="actions" />
    </div>
  </section>
</template>

<style scoped>
.empty-state {
  display: grid;
  justify-items: center;
  gap: 10px;
  padding: clamp(28px, 7vw, 56px) 20px;
  border: 1px dashed var(--bg-border);
  border-radius: 20px;
  background: var(--bg-surface);
  text-align: center;
}

.empty-state__mark {
  display: grid;
  width: 52px;
  height: 52px;
  place-items: center;
  border: 1px solid var(--bg-border);
  border-radius: 999px;
  color: var(--brand-primary);
  font-family: var(--font-display);
  font-size: var(--font-size-display);
  line-height: var(--line-height-display);
}

.empty-state h2 {
  margin: 0;
  font-size: var(--font-size-h1);
}

.empty-state p {
  max-width: 46ch;
  margin: 0;
  color: var(--text-sub);
  line-height: 1.6;
}

.empty-state__actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 10px;
  margin-top: 8px;
}
</style>
