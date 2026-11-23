import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import { makeAddon } from '../test/fixtures/catalog'
import AddonRow from './AddonRow.vue'

describe('AddonRow', () => {
  it('emits the checked state and shows the price', async () => {
    const wrapper = mount(AddonRow, {
      props: {
        addon: makeAddon(),
        selected: false,
      },
    })

    expect(wrapper.text()).toContain('Bacon crocante')
    expect(wrapper.text()).toContain('5,00')
    expect(wrapper.get('input').attributes('disabled')).toBeUndefined()

    await wrapper.get('input').setValue(true)

    expect(wrapper.emitted('update:selected')).toEqual([[true]])
  })

  it('disables an unavailable addon', () => {
    const wrapper = mount(AddonRow, {
      props: {
        addon: makeAddon({ available: false }),
        selected: false,
      },
    })

    expect(wrapper.get('input').attributes('disabled')).toBeDefined()
    expect(wrapper.classes()).toContain('addon-row--disabled')
  })
})
