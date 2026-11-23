import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import ProductImage from './ProductImage.vue'

describe('ProductImage', () => {
  it('renders an image with dimensions, alt and lazy loading', () => {
    const wrapper = mount(ProductImage, {
      props: {
        src: '/produtos/cheddar.webp',
        alt: 'Brocado Cheddar',
        width: 480,
        height: 360,
      },
    })

    const image = wrapper.get('img')
    expect(image.attributes('src')).toBe('/produtos/cheddar.webp')
    expect(image.attributes('alt')).toBe('Brocado Cheddar')
    expect(image.attributes('width')).toBe('480')
    expect(image.attributes('height')).toBe('360')
    expect(image.attributes('loading')).toBe('lazy')
  })

  it('loads the hero image eagerly when requested', () => {
    const wrapper = mount(ProductImage, {
      props: {
        src: '/produtos/cheddar.webp',
        alt: 'Brocado Cheddar',
        width: 480,
        height: 360,
        eager: true,
      },
    })

    expect(wrapper.get('img').attributes('loading')).toBe('eager')
  })

  it('falls back with an accessible label when there is no photo', () => {
    const wrapper = mount(ProductImage, {
      props: {
        src: '',
        alt: 'Brocado Cheddar',
        width: 480,
        height: 360,
      },
    })

    expect(wrapper.find('img').exists()).toBe(false)
    expect(wrapper.get('[role="img"]').attributes('aria-label')).toBe('Brocado Cheddar')
    expect(wrapper.text()).toContain('B')
    expect(wrapper.text()).toContain('Sem foto')
  })

  it('falls back when the image fails to load', async () => {
    const wrapper = mount(ProductImage, {
      props: {
        src: '/produtos/quebrada.webp',
        alt: 'Brocado Cheddar',
        width: 480,
        height: 360,
      },
    })

    await wrapper.get('img').trigger('error')

    expect(wrapper.find('img').exists()).toBe(false)
    expect(wrapper.get('[role="img"]').attributes('aria-label')).toBe('Brocado Cheddar')
  })
})
