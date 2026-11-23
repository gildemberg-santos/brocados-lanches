import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import LoadingSkeleton from './LoadingSkeleton.vue'

describe('LoadingSkeleton', () => {
  it('announces loading and renders the requested card count', () => {
    const wrapper = mount(LoadingSkeleton, {
      props: { label: 'Carregando produtos', count: 3 },
    })

    expect(wrapper.get('[role="status"]').text()).toContain('Carregando produtos')
    expect(wrapper.findAll('.loading-skeleton__card')).toHaveLength(3)
  })
})
