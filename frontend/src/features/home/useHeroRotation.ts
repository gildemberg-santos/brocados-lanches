import { computed, onMounted, onUnmounted, ref, watch, type Ref } from 'vue'

import type { CatalogProduct } from '../../types/catalog'

export const HERO_ROTATION_INTERVAL_MS = 6000

interface HeroRotationOptions {
  intervalMs?: number
}

/**
 * Alterna os produtos em destaque do hero em ciclo, dando a cada um o mesmo
 * tempo de exposição. A rotação para quando o usuário pausa, interage com o
 * hero, a aba fica oculta ou o sistema prefere movimento reduzido.
 */
export function useHeroRotation(
  products: Ref<CatalogProduct[]>,
  options: HeroRotationOptions = {},
) {
  const intervalMs = options.intervalMs ?? HERO_ROTATION_INTERVAL_MS
  const index = ref(0)
  const paused = ref(false)
  const reducedMotion = ref(false)
  const interacting = ref(false)
  const hidden = ref(false)

  const canRotate = computed(() => products.value.length > 1)
  const current = computed<CatalogProduct | null>(() => {
    if (products.value.length === 0) {
      return null
    }
    return products.value[index.value] ?? products.value[0] ?? null
  })

  let timer: ReturnType<typeof setInterval> | null = null
  let motionQuery: MediaQueryList | null = null

  function clearTimer(): void {
    if (timer === null) {
      return
    }
    clearInterval(timer)
    timer = null
  }

  function advance(): void {
    if (!canRotate.value) {
      return
    }
    index.value = (index.value + 1) % products.value.length
  }

  function startTimer(): void {
    if (
      timer !== null ||
      !canRotate.value ||
      paused.value ||
      reducedMotion.value ||
      interacting.value ||
      hidden.value
    ) {
      return
    }
    timer = setInterval(advance, intervalMs)
  }

  function syncTimer(): void {
    clearTimer()
    startTimer()
  }

  function select(nextIndex: number): void {
    if (
      !canRotate.value ||
      !Number.isInteger(nextIndex) ||
      nextIndex < 0 ||
      nextIndex >= products.value.length
    ) {
      return
    }
    index.value = nextIndex
    syncTimer()
  }

  function togglePause(): void {
    paused.value = !paused.value
    syncTimer()
  }

  function setInteracting(value: boolean): void {
    if (interacting.value === value) {
      return
    }
    interacting.value = value
    syncTimer()
  }

  function onVisibilityChange(): void {
    hidden.value = document.visibilityState === 'hidden'
    syncTimer()
  }

  function onMotionChange(event: MediaQueryListEvent): void {
    reducedMotion.value = event.matches
    syncTimer()
  }

  watch(products, (list) => {
    if (index.value >= list.length) {
      index.value = 0
    }
    syncTimer()
  })

  onMounted(() => {
    if (typeof window.matchMedia === 'function') {
      motionQuery = window.matchMedia('(prefers-reduced-motion: reduce)')
      reducedMotion.value = motionQuery.matches
      motionQuery.addEventListener('change', onMotionChange)
    }
    hidden.value = document.visibilityState === 'hidden'
    document.addEventListener('visibilitychange', onVisibilityChange)
    syncTimer()
  })

  onUnmounted(() => {
    clearTimer()
    motionQuery?.removeEventListener('change', onMotionChange)
    document.removeEventListener('visibilitychange', onVisibilityChange)
  })

  return {
    index,
    current,
    canRotate,
    paused,
    reducedMotion,
    select,
    togglePause,
    setInteracting,
  }
}
