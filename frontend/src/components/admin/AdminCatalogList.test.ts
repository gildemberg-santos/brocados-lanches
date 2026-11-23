import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'

import AdminCatalogList from './AdminCatalogList.vue'

const item = {
  id: 'a'.repeat(24),
  name: 'Cheddar',
  slug: 'cheddar',
  active: true,
  sortOrder: 2,
  priceCents: 350,
}

const meta = { page: 1, pageSize: 20, total: 2, totalPages: 2 }

function mountList(overrides: Record<string, unknown> = {}) {
  return mount(AdminCatalogList, {
    props: {
      kind: 'addon',
      items: [item],
      state: 'success',
      error: '',
      isEmpty: false,
      meta,
      ...overrides,
    },
  })
}

describe('AdminCatalogList', () => {
  it('mostra os dados do item e emite ações', async () => {
    const wrapper = mountList()

    expect(wrapper.text()).toContain('Cheddar')
    expect(wrapper.text()).toContain('cheddar')
    expect(wrapper.text()).toContain('R$\u00a03,50')
    expect(wrapper.text()).toContain('Ativo')

    const buttons = wrapper.findAll('.catalog-row__button')
    await buttons[0]?.trigger('click')
    await buttons[1]?.trigger('click')

    expect(wrapper.emitted('edit')?.[0]).toEqual([item])
    expect(wrapper.emitted('toggle')?.[0]).toEqual([item])
  })

  it('emite a página selecionada na paginação', async () => {
    const wrapper = mountList()

    await wrapper.findAll('.catalog-list__button')[1]?.trigger('click')

    expect(wrapper.emitted('page')?.[0]).toEqual([2])
  })

  it('mostra erro com ação de repetir', async () => {
    const wrapper = mountList({ state: 'error', error: 'Falha ao carregar.' })

    expect(wrapper.text()).toContain('Falha ao carregar.')
    await wrapper.find('button').trigger('click')
    expect(wrapper.emitted('retry')).toBeDefined()
  })
})
