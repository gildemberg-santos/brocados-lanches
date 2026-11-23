import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'

import AdminCatalogFilters from './AdminCatalogFilters.vue'

describe('AdminCatalogFilters', () => {
  it('emite busca e situação preenchidas', async () => {
    const wrapper = mount(AdminCatalogFilters, { props: { filters: {} } })

    await wrapper.find('#catalog-search').setValue('hambúrguer')
    await wrapper.find('#catalog-active').setValue('true')
    await wrapper.find('form').trigger('submit')

    expect(wrapper.emitted('apply')).toEqual([[{ q: 'hambúrguer', active: true }]])
  })

  it('limpa os filtros', async () => {
    const wrapper = mount(AdminCatalogFilters, {
      props: { filters: { q: 'antigo', active: false } },
    })

    await wrapper.findAll('button')[1]?.trigger('click')

    expect(wrapper.emitted('apply')).toEqual([[{}]])
  })
})
