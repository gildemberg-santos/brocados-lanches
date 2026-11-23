<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    modelValue: number
    min?: number
    max?: number
    label?: string
  }>(),
  {
    min: 1,
    max: 99,
    label: 'quantidade',
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: number]
}>()

function step(delta: number): void {
  emit('update:modelValue', Math.min(props.max, Math.max(props.min, props.modelValue + delta)))
}
</script>

<template>
  <div class="quantity-stepper" role="group" :aria-label="label">
    <button
      type="button"
      class="quantity-stepper__button"
      :aria-label="`Diminuir ${label}`"
      :disabled="modelValue <= min"
      @click="step(-1)"
    >
      −
    </button>
    <output class="quantity-stepper__value" aria-live="polite">{{ modelValue }}</output>
    <button
      type="button"
      class="quantity-stepper__button"
      :aria-label="`Aumentar ${label}`"
      :disabled="modelValue >= max"
      @click="step(1)"
    >
      +
    </button>
  </div>
</template>

<style scoped>
.quantity-stepper {
  display: inline-flex;
  min-height: 44px;
  align-items: center;
  gap: 2px;
  border: 1px solid var(--bg-border);
  border-radius: 999px;
  background: var(--bg-input);
}

.quantity-stepper__button {
  width: 44px;
  height: 44px;
  border: 0;
  border-radius: 999px;
  background: transparent;
  color: var(--brand-primary);
  font-size: var(--font-size-h2);
  font-weight: 800;
  cursor: pointer;
}

.quantity-stepper__button:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.quantity-stepper__value {
  min-width: 24px;
  color: var(--text-main);
  font-variant-numeric: tabular-nums;
  font-weight: 800;
  text-align: center;
}
</style>
