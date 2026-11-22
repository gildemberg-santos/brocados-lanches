<script setup lang="ts">
import { computed } from 'vue'

defineOptions({ inheritAttrs: false })

const props = withDefaults(
  defineProps<{
    variant?: 'primary' | 'secondary' | 'ghost'
    type?: 'button' | 'submit' | 'reset'
    disabled?: boolean
    loading?: boolean
  }>(),
  {
    variant: 'primary',
    type: 'button',
    disabled: false,
    loading: false,
  },
)

const isDisabled = computed(() => props.disabled || props.loading)
</script>

<template>
  <button
    v-bind="$attrs"
    :type="type"
    class="base-button"
    :class="`base-button--${variant}`"
    :disabled="isDisabled"
    :aria-busy="loading ? 'true' : undefined"
  >
    <span v-if="loading" class="base-button__spinner" aria-hidden="true"></span>
    <slot />
  </button>
</template>

<style scoped>
.base-button {
  display: inline-flex;
  min-height: 44px;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 12px 22px;
  border: 1px solid transparent;
  border-radius: 999px;
  font-weight: 700;
  line-height: 1.2;
  text-decoration: none;
  cursor: pointer;
  transition:
    background-color 180ms ease,
    border-color 180ms ease,
    color 180ms ease,
    opacity 180ms ease;
}

.base-button--primary {
  background: var(--brand-primary);
  color: var(--bg-primary);
}

.base-button--primary:hover:not(:disabled) {
  background: var(--brand-hover);
}

.base-button--secondary {
  border-color: var(--bg-border);
  background: var(--bg-input);
  color: var(--text-main);
}

.base-button--secondary:hover:not(:disabled) {
  border-color: var(--brand-primary);
}

.base-button--ghost {
  background: transparent;
  color: var(--brand-neon);
}

.base-button--ghost:hover:not(:disabled) {
  color: var(--brand-primary);
}

.base-button:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.base-button__spinner {
  width: 16px;
  height: 16px;
  flex: 0 0 auto;
  border: 2px solid currentColor;
  border-right-color: transparent;
  border-radius: 999px;
  animation: button-spin 700ms linear infinite;
}

@keyframes button-spin {
  to {
    transform: rotate(360deg);
  }
}

@media (prefers-reduced-motion: reduce) {
  .base-button {
    transition: none;
  }

  .base-button__spinner {
    animation-duration: 1500ms;
  }
}
</style>
