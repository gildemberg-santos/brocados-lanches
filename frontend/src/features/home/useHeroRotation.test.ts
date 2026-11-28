import { mount } from '@vue/test-utils'
import { defineComponent, h, nextTick, ref, type Ref } from 'vue'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { makeCatalogProduct } from '../../test/fixtures/catalog'
import type { CatalogProduct } from '../../types/catalog'
import { HERO_ROTATION_INTERVAL_MS, useHeroRotation } from './useHeroRotation'

type Rotation = ReturnType<typeof useHeroRotation>

const INTERVAL_MS = 1000

function makeProducts(count: number): CatalogProduct[] {
  return Array.from({ length: count }, (_, position) =>
    makeCatalogProduct({ id: `product-${position}`, slug: `product-${position}` }),
  )
}

async function mountRotation(products: Ref<CatalogProduct[]>) {
  let rotation!: Rotation

  const Host = defineComponent({
    setup() {
      rotation = useHeroRotation(products, { intervalMs: INTERVAL_MS })
      return () => h('div')
    },
  })

  const wrapper = mount(Host)
  await nextTick()
  return { rotation, wrapper }
}

describe('useHeroRotation', () => {
  const originalMatchMedia = window.matchMedia
  const originalVisibility = Object.getOwnPropertyDescriptor(document, 'visibilityState')

  beforeEach(() => {
    vi.useFakeTimers()
    Object.defineProperty(window, 'matchMedia', {
      configurable: true,
      value: undefined,
      writable: true,
    })
    Object.defineProperty(document, 'visibilityState', {
      configurable: true,
      get: () => 'visible',
    })
  })

  afterEach(() => {
    vi.useRealTimers()
    Object.defineProperty(window, 'matchMedia', {
      configurable: true,
      value: originalMatchMedia,
      writable: true,
    })
    if (originalVisibility) {
      Object.defineProperty(document, 'visibilityState', originalVisibility)
    }
  })

  it('usa o intervalo padrão de 6 segundos para cada destaque', async () => {
    let rotation!: Rotation
    const Host = defineComponent({
      setup() {
        rotation = useHeroRotation(ref(makeProducts(2)))
        return () => h('div')
      },
    })

    const wrapper = mount(Host)
    await nextTick()

    expect(HERO_ROTATION_INTERVAL_MS).toBe(6000)

    vi.advanceTimersByTime(5999)
    expect(rotation.index.value).toBe(0)

    vi.advanceTimersByTime(1)
    expect(rotation.index.value).toBe(1)

    wrapper.unmount()
  })

  it('alterna todos os produtos em ciclo e reinicia no primeiro', async () => {
    const products = ref(makeProducts(3))
    const { rotation, wrapper } = await mountRotation(products)

    expect(rotation.current.value?.id).toBe('product-0')
    expect(rotation.canRotate.value).toBe(true)

    vi.advanceTimersByTime(INTERVAL_MS)
    expect(rotation.current.value?.id).toBe('product-1')

    vi.advanceTimersByTime(INTERVAL_MS)
    expect(rotation.current.value?.id).toBe('product-2')

    vi.advanceTimersByTime(INTERVAL_MS)
    expect(rotation.current.value?.id).toBe('product-0')

    wrapper.unmount()
  })

  it('pausa e retoma a rotação explicitamente', async () => {
    const { rotation, wrapper } = await mountRotation(ref(makeProducts(2)))

    rotation.togglePause()
    expect(rotation.paused.value).toBe(true)

    vi.advanceTimersByTime(INTERVAL_MS * 3)
    expect(rotation.index.value).toBe(0)

    rotation.togglePause()
    vi.advanceTimersByTime(INTERVAL_MS)
    expect(rotation.index.value).toBe(1)

    wrapper.unmount()
  })

  it('escolher um destaque reinicia o tempo de exposição daquele produto', async () => {
    const { rotation, wrapper } = await mountRotation(ref(makeProducts(3)))

    vi.advanceTimersByTime(INTERVAL_MS - 100)
    rotation.select(2)
    expect(rotation.index.value).toBe(2)

    vi.advanceTimersByTime(INTERVAL_MS - 1)
    expect(rotation.index.value).toBe(2)

    vi.advanceTimersByTime(1)
    expect(rotation.index.value).toBe(0)

    wrapper.unmount()
  })

  it('ignora índices inválidos', async () => {
    const { rotation, wrapper } = await mountRotation(ref(makeProducts(2)))

    rotation.select(-1)
    rotation.select(2)
    rotation.select(1.5)
    expect(rotation.index.value).toBe(0)

    wrapper.unmount()
  })

  it('pausa durante a interação e retoma ao sair do hero', async () => {
    const { rotation, wrapper } = await mountRotation(ref(makeProducts(2)))

    rotation.setInteracting(true)
    vi.advanceTimersByTime(INTERVAL_MS * 2)
    expect(rotation.index.value).toBe(0)

    rotation.setInteracting(false)
    vi.advanceTimersByTime(INTERVAL_MS)
    expect(rotation.index.value).toBe(1)

    wrapper.unmount()
  })

  it('não inicia a rotação automática com movimento reduzido', async () => {
    let listener: ((event: MediaQueryListEvent) => void) | undefined
    Object.defineProperty(window, 'matchMedia', {
      configurable: true,
      value: vi.fn(() => ({
        matches: true,
        addEventListener: vi.fn(
          (_event: string, callback: (event: MediaQueryListEvent) => void) => {
            listener = callback
          },
        ),
        removeEventListener: vi.fn(),
      })),
      writable: true,
    })

    const { rotation, wrapper } = await mountRotation(ref(makeProducts(3)))

    expect(rotation.reducedMotion.value).toBe(true)
    vi.advanceTimersByTime(INTERVAL_MS * 2)
    expect(rotation.index.value).toBe(0)

    rotation.select(1)
    expect(rotation.current.value?.id).toBe('product-1')

    listener?.({ matches: false } as MediaQueryListEvent)
    vi.advanceTimersByTime(INTERVAL_MS)
    expect(rotation.index.value).toBe(2)

    wrapper.unmount()
  })

  it('pausa com a aba oculta e retoma quando ela volta', async () => {
    const { rotation, wrapper } = await mountRotation(ref(makeProducts(3)))

    Object.defineProperty(document, 'visibilityState', {
      configurable: true,
      get: () => 'hidden',
    })
    document.dispatchEvent(new Event('visibilitychange'))

    vi.advanceTimersByTime(INTERVAL_MS * 2)
    expect(rotation.index.value).toBe(0)

    Object.defineProperty(document, 'visibilityState', {
      configurable: true,
      get: () => 'visible',
    })
    document.dispatchEvent(new Event('visibilitychange'))

    vi.advanceTimersByTime(INTERVAL_MS)
    expect(rotation.index.value).toBe(1)

    wrapper.unmount()
  })

  it('não gira com um único produto e ajusta o índice quando a lista encolhe', async () => {
    const products = ref(makeProducts(3))
    const { rotation, wrapper } = await mountRotation(products)

    rotation.select(2)
    products.value = products.value.slice(0, 1)
    await nextTick()

    expect(rotation.canRotate.value).toBe(false)
    expect(rotation.index.value).toBe(0)
    expect(rotation.current.value?.id).toBe('product-0')

    vi.advanceTimersByTime(INTERVAL_MS * 3)
    expect(rotation.index.value).toBe(0)

    wrapper.unmount()
  })

  it('limpa o timer ao desmontar', async () => {
    const products = ref(makeProducts(2))
    const { rotation, wrapper } = await mountRotation(products)

    wrapper.unmount()
    vi.advanceTimersByTime(INTERVAL_MS * 3)
    expect(rotation.index.value).toBe(0)
  })
})
