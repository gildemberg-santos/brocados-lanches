import { mount } from '@vue/test-utils'
import { defineComponent, h } from 'vue'
import { createMemoryHistory, createRouter } from 'vue-router'
import { describe, expect, it } from 'vitest'

import AdminForbiddenView from './AdminForbiddenView.vue'
import AdminPlaceholderView from './AdminPlaceholderView.vue'

const StubView = defineComponent({ render: () => h('div') })

describe('telas estáticas do admin', () => {
  it('mostra 403 com volta ao painel', async () => {
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { path: '/admin/sem-acesso', name: 'admin-forbidden', component: StubView },
        { path: '/admin', name: 'admin-dashboard', component: StubView },
      ],
    })
    await router.push('/admin/sem-acesso')
    await router.isReady()

    const wrapper = mount(AdminForbiddenView, { global: { plugins: [router] } })

    expect(wrapper.text()).toContain('403')
    expect(wrapper.get('h1').text()).toBe('Acesso negado')
    expect(wrapper.text()).toContain('Voltar ao painel')
  })

  it('mostra o título da rota em área em preparação', async () => {
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [
        {
          path: '/admin/produtos',
          name: 'admin-products',
          component: StubView,
          meta: { title: 'Produtos' },
        },
        { path: '/admin', name: 'admin-dashboard', component: StubView },
      ],
    })
    await router.push('/admin/produtos')
    await router.isReady()

    const wrapper = mount(AdminPlaceholderView, { global: { plugins: [router] } })

    expect(wrapper.get('h1').text()).toBe('Produtos')
    expect(wrapper.text()).toContain('em preparação')
  })
})
