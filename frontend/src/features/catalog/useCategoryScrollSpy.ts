import { nextTick, onMounted, onUnmounted, ref, watch, type Ref } from 'vue'

import type { MenuCategory } from '../../types/catalog'
import {
  CATEGORY_SECTION_ATTRIBUTE,
  categorySectionId,
  resolveActiveCategory,
  type CategoryScrollDirection,
  type CategorySectionMetric,
} from './categorySections'

/** Frames estáveis antes de o spy voltar a seguir um scroll programado. */
export const PROGRAMMATIC_STABLE_FRAMES = 20

/**
 * Acompanha a seção visível do cardápio e mantém a categoria ativa em `activeSlug`.
 * As posições são medidas no mount, no resize e quando o catálogo muda; o scroll
 * só compara números já capturados.
 */
export function useCategoryScrollSpy(categories: Ref<MenuCategory[]>, enabled: Ref<boolean>) {
  const activeSlug = ref('')

  let metrics: CategorySectionMetric[] = []
  let lastScrollY = 0
  let programmaticScroll = false
  let stableFrames = 0
  let lastStableScrollY = 0
  let frameId: number | null = null

  function prefersReducedMotion(): boolean {
    return (
      typeof window.matchMedia === 'function' &&
      window.matchMedia('(prefers-reduced-motion: reduce)').matches
    )
  }

  function scrollBehavior(): ScrollBehavior {
    return prefersReducedMotion() ? 'auto' : 'smooth'
  }

  function maxScroll(): number {
    return Math.max(0, document.documentElement.scrollHeight - window.innerHeight)
  }

  function measure(): void {
    if (!enabled.value) {
      metrics = []
      return
    }

    metrics = Array.from(document.querySelectorAll<HTMLElement>(`[${CATEGORY_SECTION_ATTRIBUTE}]`))
      .map((element) => {
        const rect = element.getBoundingClientRect()
        return {
          slug: element.getAttribute(CATEGORY_SECTION_ATTRIBUTE) ?? '',
          top: rect.top + window.scrollY,
          bottom: rect.bottom + window.scrollY,
        }
      })
      .filter((section) => section.slug.length > 0)

    lastScrollY = window.scrollY
  }

  function cancelStabilityWatch(): void {
    if (frameId !== null && typeof window.cancelAnimationFrame === 'function') {
      window.cancelAnimationFrame(frameId)
    }
    frameId = null
  }

  function resumeSpy(): void {
    if (!programmaticScroll) {
      return
    }
    programmaticScroll = false
    cancelStabilityWatch()
    measure()
  }

  function watchScrollStability(): void {
    cancelStabilityWatch()
    lastStableScrollY = window.scrollY
    stableFrames = 0

    if (typeof window.requestAnimationFrame !== 'function') {
      programmaticScroll = false
      return
    }

    const tick = (): void => {
      if (!programmaticScroll || !enabled.value) {
        return
      }

      const current = window.scrollY
      if (Math.abs(current - lastStableScrollY) < 1) {
        stableFrames += 1
      } else {
        stableFrames = 0
        lastStableScrollY = current
      }

      if (stableFrames >= PROGRAMMATIC_STABLE_FRAMES) {
        resumeSpy()
        return
      }

      frameId = window.requestAnimationFrame(tick)
    }

    frameId = window.requestAnimationFrame(tick)
  }

  function onScroll(): void {
    if (!enabled.value || programmaticScroll || metrics.length === 0) {
      return
    }

    const current = window.scrollY
    const direction: CategoryScrollDirection = current >= lastScrollY ? 'down' : 'up'
    lastScrollY = current

    const slug = resolveActiveCategory(metrics, current, direction, maxScroll())
    if (slug.length > 0) {
      activeSlug.value = slug
    }
  }

  function onUserInteraction(): void {
    resumeSpy()
  }

  function onKeydown(event: KeyboardEvent): void {
    if (event.key === ' ') {
      resumeSpy()
    }
  }

  /**
   * Destaca a categoria na hora e rola até a seção correspondente. Durante o
   * scroll programado o spy fica pausado para não sobrescrever a escolha do
   * clique; voltas do usuário ou estabilidade reativam o acompanhamento.
   */
  function scrollToCategory(slug: string): void {
    if (!enabled.value) {
      return
    }

    measure()
    const behavior = scrollBehavior()
    programmaticScroll = true

    if (slug.length === 0) {
      activeSlug.value = metrics[0]?.slug ?? ''
      window.scrollTo({ top: 0, behavior })
    } else {
      activeSlug.value = slug
      document.getElementById(categorySectionId(slug))?.scrollIntoView({ behavior, block: 'start' })
    }

    watchScrollStability()
  }

  onMounted(() => {
    window.addEventListener('scroll', onScroll, { passive: true })
    window.addEventListener('wheel', onUserInteraction, { passive: true })
    window.addEventListener('touchmove', onUserInteraction, { passive: true })
    window.addEventListener('pointerdown', onUserInteraction, { passive: true })
    window.addEventListener('keydown', onKeydown)
    window.addEventListener('resize', measure)
    void nextTick(measure)
  })

  onUnmounted(() => {
    window.removeEventListener('scroll', onScroll)
    window.removeEventListener('wheel', onUserInteraction)
    window.removeEventListener('touchmove', onUserInteraction)
    window.removeEventListener('pointerdown', onUserInteraction)
    window.removeEventListener('keydown', onKeydown)
    window.removeEventListener('resize', measure)
    cancelStabilityWatch()
  })

  watch(categories, () => {
    void nextTick(measure)
  })

  watch(enabled, async (value) => {
    if (value) {
      await nextTick()
      measure()
      return
    }

    metrics = []
    programmaticScroll = false
    cancelStabilityWatch()
  })

  return {
    activeSlug,
    scrollToCategory,
    remeasure: measure,
  }
}
