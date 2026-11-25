import { mount } from '@vue/test-utils'
import { defineComponent, h } from 'vue'
import { createMemoryHistory, createRouter } from 'vue-router'
import { describe, expect, it } from 'vitest'

import AdminForbiddenView from './AdminForbiddenView.vue'

const StubView = defineComponent({ render: () => h('div') })

describe('AdminForbiddenView', () => {
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
})
