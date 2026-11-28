import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import { makeCategory } from '../test/fixtures/catalog'
import MenuCategoryNav from './MenuCategoryNav.vue'

const categories = [
  makeCategory({ id: 'category-1', name: 'Hambúrgueres', slug: 'hamburgueres' }),
  makeCategory({ id: 'category-2', name: 'Bebidas', slug: 'bebidas' }),
]

describe('MenuCategoryNav', () => {
  it('marca somente a categoria ativa com aria-pressed', () => {
    const wrapper = mount(MenuCategoryNav, {
      props: { categories, activeSlug: 'bebidas' },
    })

    const pills = wrapper.findAll('button')

    expect(wrapper.get('nav').attributes('aria-label')).toBe('Categorias do cardápio')
    expect(pills).toHaveLength(2)
    expect(pills[0]?.attributes('aria-pressed')).toBe('false')
    expect(pills[1]?.attributes('aria-pressed')).toBe('true')
  })

  it('emite o slug da categoria escolhida', async () => {
    const wrapper = mount(MenuCategoryNav, {
      props: { categories, activeSlug: '' },
    })

    await wrapper.findAll('button')[0]?.trigger('click')

    expect(wrapper.emitted('select')).toEqual([['hamburgueres']])
  })
})
