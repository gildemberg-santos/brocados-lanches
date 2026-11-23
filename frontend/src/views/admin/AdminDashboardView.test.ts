import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { defineComponent, h } from 'vue'
import { createMemoryHistory, createRouter } from 'vue-router'
import { describe, expect, it } from 'vitest'

import { useAdminAuthStore } from '../../stores/adminAuth'
import AdminDashboardView from './AdminDashboardView.vue'

const StubView = defineComponent({ render: () => h('div') })

describe('AdminDashboardView', () => {
  it('sauda o usuário e mostra atalhos permitidos ao papel', async () => {
    const pinia = createPinia()
    setActivePinia(pinia)
    const auth = useAdminAuthStore()
    auth.user = {
      id: 'user-1',
      name: 'Dono',
      email: 'dono@example.com',
      role: 'owner',
    }

    const router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { path: '/admin', name: 'admin-dashboard', component: StubView },
        { path: '/admin/pedidos', name: 'admin-orders', component: StubView },
        { path: '/admin/produtos', name: 'admin-products', component: StubView },
        { path: '/admin/conta', name: 'admin-account', component: StubView },
      ],
    })
    await router.push('/admin')
    await router.isReady()

    const wrapper = mount(AdminDashboardView, {
      global: { plugins: [pinia, router] },
    })

    expect(wrapper.get('h1').text()).toBe('Olá, Dono')
    expect(wrapper.text()).toContain('Administrador')
    expect(wrapper.text()).toContain('Pedidos')
    expect(wrapper.text()).toContain('Produtos')
  })
})
