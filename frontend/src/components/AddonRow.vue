<script setup lang="ts">
import type { MenuAddon } from '../types/catalog'
import PriceText from './PriceText.vue'

defineProps<{
  addon: MenuAddon
  selected: boolean
}>()

const emit = defineEmits<{
  'update:selected': [selected: boolean]
}>()

function handleChange(event: Event): void {
  const target = event.target as HTMLInputElement
  emit('update:selected', target.checked)
}
</script>

<template>
  <label class="addon-row" :class="{ 'addon-row--disabled': !addon.available }">
    <span class="addon-row__name">
      <input
        class="addon-row__input"
        type="checkbox"
        :checked="selected"
        :disabled="!addon.available"
        @change="handleChange"
      />
      <span>{{ addon.name }}</span>
    </span>
    <span class="addon-row__price"> + <PriceText :value-cents="addon.priceCents" /> </span>
  </label>
</template>

<style scoped>
.addon-row {
  display: flex;
  min-height: 44px;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 6px 0;
  cursor: pointer;
}

.addon-row__name {
  display: inline-flex;
  min-width: 0;
  align-items: center;
  gap: 10px;
}

.addon-row__input {
  width: 18px;
  height: 18px;
  flex: 0 0 auto;
  accent-color: var(--brand-primary);
}

.addon-row__price {
  display: inline-flex;
  flex: 0 0 auto;
  align-items: baseline;
  gap: 2px;
  color: var(--text-sub);
  font-size: var(--font-size-caption);
  font-weight: 700;
}

.addon-row--disabled {
  opacity: 0.55;
  cursor: not-allowed;
}
</style>
