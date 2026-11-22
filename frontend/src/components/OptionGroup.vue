<script setup lang="ts">
import { computed } from 'vue'

import { minimumSelections } from '../features/catalog/customization'
import type { MenuOptionGroup } from '../types/catalog'
import PriceText from './PriceText.vue'

const props = defineProps<{
  group: MenuOptionGroup
  selectedIds: string[]
  error?: string
}>()

const emit = defineEmits<{
  'update:selectedIds': [ids: string[]]
}>()

const isSingleChoice = computed(() => props.group.maxSelections <= 1)
const minimum = computed(() => minimumSelections(props.group))
const errorId = computed(() => `group-${props.group.id}-error`)

const helperText = computed(() => {
  if (props.group.options.length === 0) {
    return 'Nenhuma opção disponível.'
  }

  if (isSingleChoice.value) {
    return props.group.required ? 'Obrigatório · escolha 1' : 'Opcional · escolha 1'
  }

  if (minimum.value > 0) {
    return `Obrigatório · escolha de ${minimum.value} a ${props.group.maxSelections}`
  }

  return `Opcional · escolha até ${props.group.maxSelections}`
})

function isSelected(optionId: string): boolean {
  return props.selectedIds.includes(optionId)
}

function isMaxReached(): boolean {
  return props.selectedIds.length >= props.group.maxSelections
}

function selectSingle(optionId: string): void {
  emit('update:selectedIds', [optionId])
}

function toggleMultiple(optionId: string, event: Event): void {
  const target = event.target as HTMLInputElement

  if (target.checked) {
    if (isMaxReached()) {
      target.checked = false
      return
    }

    emit('update:selectedIds', [...props.selectedIds, optionId])
    return
  }

  emit(
    'update:selectedIds',
    props.selectedIds.filter((id) => id !== optionId),
  )
}
</script>

<template>
  <fieldset
    class="option-group"
    :aria-required="group.required ? 'true' : undefined"
    :aria-describedby="error ? errorId : undefined"
  >
    <legend class="option-group__legend">
      <span>{{ group.name }}</span>
      <span class="option-group__helper">{{ helperText }}</span>
    </legend>

    <div v-if="group.options.length > 0" class="option-group__options">
      <label
        v-for="option in group.options"
        :key="option.id"
        class="option-group__option"
        :class="{ 'option-group__option--disabled': !option.available }"
      >
        <span class="option-group__choice">
          <input
            v-if="isSingleChoice"
            class="option-group__input"
            type="radio"
            :name="`group-${group.id}`"
            :value="option.id"
            :checked="isSelected(option.id)"
            :disabled="!option.available"
            @change="selectSingle(option.id)"
          />
          <input
            v-else
            class="option-group__input"
            type="checkbox"
            :value="option.id"
            :checked="isSelected(option.id)"
            :disabled="!option.available || (isMaxReached() && !isSelected(option.id))"
            @change="toggleMultiple(option.id, $event)"
          />
          <span>{{ option.name }}</span>
        </span>

        <span class="option-group__price">
          <template v-if="option.priceCents > 0">
            + <PriceText :value-cents="option.priceCents" />
          </template>
          <template v-else>Grátis</template>
        </span>
      </label>
    </div>

    <p v-else class="option-group__empty">Este grupo está sem opções disponíveis.</p>

    <p v-if="error" :id="errorId" class="option-group__error" role="alert">{{ error }}</p>
  </fieldset>
</template>

<style scoped>
.option-group {
  min-width: 0;
  margin: 0;
  padding: 16px;
  scroll-margin-top: 88px;
  border: 1px solid var(--bg-border);
  border-radius: 16px;
  background: var(--bg-surface);
}

.option-group__legend {
  display: flex;
  width: 100%;
  flex-wrap: wrap;
  align-items: baseline;
  justify-content: space-between;
  gap: 6px 12px;
  padding: 0;
  font-size: var(--font-size-h2);
  font-weight: 800;
}

.option-group__helper {
  color: var(--text-sub);
  font-size: var(--font-size-caption);
  font-weight: 600;
}

.option-group__options {
  display: grid;
  gap: 2px;
  margin-top: 10px;
}

.option-group__option {
  display: flex;
  min-height: 44px;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 6px 0;
  cursor: pointer;
}

.option-group__choice {
  display: inline-flex;
  min-width: 0;
  align-items: center;
  gap: 10px;
}

.option-group__input {
  width: 18px;
  height: 18px;
  flex: 0 0 auto;
  accent-color: var(--brand-primary);
}

.option-group__price {
  display: inline-flex;
  flex: 0 0 auto;
  align-items: baseline;
  gap: 2px;
  color: var(--text-sub);
  font-size: var(--font-size-caption);
  font-weight: 700;
}

.option-group__option--disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.option-group__empty,
.option-group__error {
  margin: 10px 0 0;
  font-size: var(--font-size-caption);
  line-height: 1.5;
}

.option-group__empty {
  color: var(--text-sub);
}

.option-group__error {
  color: var(--status-error);
  font-weight: 700;
}
</style>
