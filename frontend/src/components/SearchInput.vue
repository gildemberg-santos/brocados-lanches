<script setup lang="ts">
import { nextTick, ref } from 'vue'

withDefaults(
  defineProps<{
    modelValue: string
    label?: string
    placeholder?: string
  }>(),
  {
    label: 'Buscar no cardápio',
    placeholder: 'Buscar por nome ou descrição',
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const input = ref<HTMLInputElement | null>(null)

function handleInput(event: Event): void {
  const target = event.target as HTMLInputElement
  emit('update:modelValue', target.value)
}

async function clearSearch(): Promise<void> {
  emit('update:modelValue', '')
  await nextTick()
  input.value?.focus()
}
</script>

<template>
  <div class="search-input">
    <svg class="search-input__icon" viewBox="0 0 24 24" aria-hidden="true">
      <circle cx="11" cy="11" r="7" fill="none" stroke="currentColor" stroke-width="2" />
      <path d="m16.5 16.5 4 4" fill="none" stroke="currentColor" stroke-width="2" />
    </svg>
    <input
      ref="input"
      class="search-input__field"
      type="search"
      :value="modelValue"
      :aria-label="label"
      :placeholder="placeholder"
      autocomplete="off"
      autocapitalize="none"
      enterkeyhint="search"
      @input="handleInput"
    />
    <button
      v-if="modelValue.length > 0"
      class="search-input__clear"
      type="button"
      aria-label="Limpar busca"
      @click="clearSearch"
    >
      <svg viewBox="0 0 24 24" aria-hidden="true">
        <path d="m6 6 12 12M18 6 6 18" fill="none" stroke="currentColor" stroke-width="2" />
      </svg>
    </button>
  </div>
</template>

<style scoped>
.search-input {
  display: flex;
  min-height: 48px;
  align-items: center;
  gap: 10px;
  padding: 0 6px 0 14px;
  border: 1px solid var(--bg-border);
  border-radius: 14px;
  background: var(--bg-input);
}

.search-input:focus-within {
  border-color: var(--brand-neon);
}

.search-input__icon,
.search-input__clear svg {
  width: 20px;
  height: 20px;
  flex: 0 0 auto;
}

.search-input__icon {
  color: var(--text-sub);
}

.search-input__field {
  width: 100%;
  min-width: 0;
  min-height: 44px;
  padding: 0;
  border: 0;
  outline: 0;
  background: transparent;
  color: var(--text-main);
}

.search-input__field::placeholder {
  color: var(--text-sub);
}

.search-input__field::-webkit-search-cancel-button {
  display: none;
}

.search-input__clear {
  display: grid;
  width: 44px;
  height: 44px;
  flex: 0 0 auto;
  place-items: center;
  border: 0;
  border-radius: 999px;
  background: transparent;
  color: var(--text-sub);
  cursor: pointer;
}

.search-input__clear:hover {
  color: var(--text-main);
}
</style>
