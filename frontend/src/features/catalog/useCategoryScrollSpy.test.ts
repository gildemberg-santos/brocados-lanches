import { flushPromises, mount } from '@vue/test-utils'
import { defineComponent, h, ref } from 'vue'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import type { MenuCategory } from '../../types/catalog'
import { categorySectionId } from './categorySections'
import { useCategoryScrollSpy } from './useCategoryScrollSpy'

const categories = ref<MenuCategory[]>([
  { ...makeCategory('combos', 'Combos', 'combos') },
  { ...makeCategory('hamburgueres', 'Hambúrgueres', 'hamburgueres') },
  { ...makeCategory('bebidas', 'Bebidas', 'bebidas') },
])

function makeCategory(id: string, name: string, slug: string): MenuCategory {
  return { id, name, slug, sortOrder: 0, products: [] }
}

const Harness = defineComponent({
  setup() {
    const enabled = ref(true)
    const spy = useCategoryScrollSpy(categories, enabled)

    return {
      enabled,
      categorySectionId,
      activeSlug: spy.activeSlug,
      remeasure: spy.remeasure,
      scrollToCategory: spy.scrollToCategory,
    }
  },
  render() {
    return h(
      'div',
      categories.value.map((category) =>
        h('section', {
          key: category.id,
          id: this.categorySectionId(category.slug),
          'data-category-section': category.slug,
        }),
      ),
    )
  },
})

type HarnessVm = {
  activeSlug: string
  remeasure: () => void
  scrollToCategory: (slug: string) => void
}

interface Rect {
  top: number
  bottom: number
}

const rects: Record<string, Rect> = {
  combos: { top: 300, bottom: 900 },
  hamburgueres: { top: 950, bottom: 1600 },
  bebidas: { top: 1650, bottom: 2100 },
}

function setScrollY(value: number): void {
  Object.defineProperty(window, 'scrollY', { configurable: true, writable: true, value })
}

function scrollTo(value: number): void {
  setScrollY(value)
  window.dispatchEvent(new Event('scroll'))
}

async function mountHarness() {
  const host = document.createElement('div')
  document.body.appendChild(host)
  const wrapper = mount(Harness, { attachTo: host })
  await flushPromises()

  for (const element of wrapper.findAll('[data-category-section]')) {
    const slug = element.attributes('data-category-section') ?? ''
    const rect = rects[slug] ?? { top: 0, bottom: 0 }
    Object.defineProperty(element.element, 'getBoundingClientRect', {
      configurable: true,
      value: () => ({ ...rect, left: 0, right: 0, width: 0, height: rect.bottom - rect.top }),
    })
  }

  setScrollY(0)
  Object.defineProperty(window, 'innerHeight', {
    configurable: true,
    writable: true,
    value: 700,
  })
  Object.defineProperty(document.documentElement, 'scrollHeight', {
    configurable: true,
    value: 2700,
  })

  const vm = wrapper.vm as unknown as HarnessVm
  vm.remeasure()

  return { wrapper, vm }
}

describe('useCategoryScrollSpy', () => {
  beforeEach(() => {
    Object.defineProperty(window, 'requestAnimationFrame', {
      configurable: true,
      writable: true,
      value: vi.fn(() => 1),
    })
    Object.defineProperty(window, 'cancelAnimationFrame', {
      configurable: true,
      writable: true,
      value: vi.fn(),
    })
  })

  afterEach(() => {
    vi.restoreAllMocks()
    document.body.innerHTML = ''
  })

  it('destaca a primeira categoria ao medir sem scroll prévio', async () => {
    const { vm } = await mountHarness()

    expect(vm.activeSlug).toBe('combos')

    scrollTo(500)
    expect(vm.activeSlug).toBe('combos')
  })

  it('acompanha o scroll para baixo, incluindo o fim da página', async () => {
    const { vm } = await mountHarness()

    scrollTo(500)
    expect(vm.activeSlug).toBe('combos')

    scrollTo(960)
    expect(vm.activeSlug).toBe('hamburgueres')

    scrollTo(1661)
    expect(vm.activeSlug).toBe('bebidas')

    scrollTo(2000)
    expect(vm.activeSlug).toBe('bebidas')
  })

  it('retrocede de categoria quando o scroll sobe', async () => {
    const { vm } = await mountHarness()

    scrollTo(2000)
    scrollTo(1490)
    expect(vm.activeSlug).toBe('hamburgueres')

    scrollTo(850)
    expect(vm.activeSlug).toBe('combos')
  })

  it('destaca a categoria clicada e rola até a seção', async () => {
    const scrollIntoView = vi.fn()
    Object.defineProperty(Element.prototype, 'scrollIntoView', {
      configurable: true,
      value: scrollIntoView,
    })

    const { wrapper, vm } = await mountHarness()

    vm.scrollToCategory('hamburgueres')

    expect(vm.activeSlug).toBe('hamburgueres')
    expect(scrollIntoView).toHaveBeenCalledWith({ behavior: 'smooth', block: 'start' })
    expect(wrapper.find(`#${categorySectionId('hamburgueres')}`).exists()).toBe(true)
  })

  it('volta ao topo quando Todos é escolhido', async () => {
    const scrollToTop = vi.fn()
    vi.spyOn(window, 'scrollTo').mockImplementation(scrollToTop)

    const { vm } = await mountHarness()

    vm.scrollToCategory('')

    expect(vm.activeSlug).toBe('combos')
    expect(scrollToTop).toHaveBeenCalledWith({ top: 0, behavior: 'smooth' })
  })

  it('mantém a escolha do clique até o usuário interagir', async () => {
    const { vm } = await mountHarness()

    vm.scrollToCategory('hamburgueres')
    scrollTo(500)
    expect(vm.activeSlug).toBe('hamburgueres')

    window.dispatchEvent(new Event('wheel'))
    scrollTo(500)
    expect(vm.activeSlug).toBe('combos')
  })

  it('reativa o spy quando o usuário rola com o espaço', async () => {
    const { vm } = await mountHarness()

    vm.scrollToCategory('bebidas')
    scrollTo(500)
    expect(vm.activeSlug).toBe('bebidas')

    window.dispatchEvent(new KeyboardEvent('keydown', { key: ' ' }))
    scrollTo(500)
    expect(vm.activeSlug).toBe('combos')
  })
})
