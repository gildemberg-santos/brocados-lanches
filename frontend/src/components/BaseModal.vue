<script setup lang="ts">
import { nextTick, onBeforeUnmount, ref, useId, watch } from 'vue'

const props = defineProps<{
  open: boolean
  title: string
}>()

const emit = defineEmits<{ close: [] }>()

const panel = ref<HTMLElement | null>(null)
const titleId = useId()

const FOCUSABLE_SELECTOR = [
  'a[href]',
  'button:not([disabled])',
  'input:not([disabled]):not([type="hidden"])',
  'select:not([disabled])',
  'textarea:not([disabled])',
  '[tabindex]:not([tabindex="-1"])',
].join(', ')

let previouslyFocused: HTMLElement | null = null
let previousOverflow = ''

function focusableElements(): HTMLElement[] {
  if (!panel.value) {
    return []
  }
  return Array.from(panel.value.querySelectorAll<HTMLElement>(FOCUSABLE_SELECTOR)).filter(
    (element) => !element.hasAttribute('hidden') && element.getAttribute('aria-hidden') !== 'true',
  )
}

function focusFirstField(): void {
  const field = panel.value?.querySelector<HTMLElement>(
    [
      'a[href]',
      'button:not([disabled])',
      'input:not([disabled]):not([type="hidden"])',
      'select:not([disabled])',
      'textarea:not([disabled])',
    ]
      .map((selector) => `.base-modal__body ${selector}`)
      .join(', '),
  )
  if (field) {
    field.focus()
    return
  }
  panel.value?.focus()
}

function handleKeydown(event: KeyboardEvent): void {
  if (event.key === 'Escape') {
    event.preventDefault()
    emit('close')
    return
  }

  if (event.key !== 'Tab') {
    return
  }

  const focusable = focusableElements()
  if (focusable.length === 0) {
    event.preventDefault()
    panel.value?.focus()
    return
  }

  const first = focusable[0]
  const last = focusable[focusable.length - 1]
  const active = document.activeElement

  if (event.shiftKey && active === first) {
    event.preventDefault()
    last.focus()
  } else if (!event.shiftKey && active === last) {
    event.preventDefault()
    first.focus()
  }
}

function release(): void {
  document.removeEventListener('keydown', handleKeydown)
  document.body.style.overflow = previousOverflow
  if (previouslyFocused?.isConnected) {
    previouslyFocused.focus()
  }
  previouslyFocused = null
}

watch(
  () => props.open,
  async (open) => {
    if (!open) {
      release()
      return
    }

    previouslyFocused =
      document.activeElement instanceof HTMLElement ? document.activeElement : null
    previousOverflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    document.addEventListener('keydown', handleKeydown)
    await nextTick()
    focusFirstField()
  },
  { immediate: true },
)

onBeforeUnmount(release)
</script>

<template>
  <div v-if="open" class="base-modal" data-testid="admin-modal">
    <div class="base-modal__backdrop" aria-hidden="true" @click="emit('close')"></div>

    <section
      ref="panel"
      class="base-modal__panel"
      role="dialog"
      aria-modal="true"
      :aria-labelledby="titleId"
      tabindex="-1"
    >
      <header class="base-modal__header">
        <h2 :id="titleId">{{ title }}</h2>
        <button type="button" class="base-modal__close" aria-label="Fechar" @click="emit('close')">
          ×
        </button>
      </header>

      <div class="base-modal__body">
        <slot />
      </div>
    </section>
  </div>
</template>

<style scoped>
.base-modal {
  position: fixed;
  inset: 0;
  z-index: 60;
  display: grid;
  place-items: center;
  padding: calc(16px + env(safe-area-inset-top)) 16px calc(16px + env(safe-area-inset-bottom));
}

.base-modal__backdrop {
  position: absolute;
  inset: 0;
  background: color-mix(in srgb, var(--bg-primary) 80%, transparent);
}

.base-modal__panel {
  position: relative;
  width: min(920px, 100%);
  /* `dvh` acompanha a barra do navegador e o teclado no mobile. */
  max-height: calc(100dvh - 32px - env(safe-area-inset-top) - env(safe-area-inset-bottom));
  overflow-y: auto;
  overscroll-behavior: contain;
  padding: 24px 24px calc(24px + env(safe-area-inset-bottom));
  border: 1px solid var(--bg-border);
  border-radius: 24px;
  background: var(--bg-surface);
  box-shadow: 0 24px 48px var(--bg-primary);
}

.base-modal__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--bg-border);
}

.base-modal__header h2 {
  margin: 0;
  font-family: var(--font-display);
  font-size: var(--font-size-h2);
  line-height: var(--line-height-display);
}

.base-modal__close {
  min-width: 44px;
  min-height: 44px;
  border: 1px solid var(--bg-border);
  border-radius: 999px;
  background: var(--bg-input);
  color: var(--text-main);
  font-family: var(--font-sans);
  font-size: var(--font-size-h2);
  line-height: 1;
}

.base-modal__close:hover {
  border-color: var(--brand-primary);
}

.base-modal__body {
  display: grid;
  gap: 16px;
}

@media (min-width: 768px) {
  .base-modal {
    padding: 32px;
  }

  .base-modal__panel {
    max-height: min(900px, calc(100dvh - 64px));
  }
}
</style>
