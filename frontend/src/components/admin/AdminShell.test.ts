import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { defineComponent, h } from 'vue'
import { createMemoryHistory, createRouter } from 'vue-router'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { adminNavItems } from '../../features/admin/auth'
import { useAdminAuthStore } from '../../stores/adminAuth'
import type { AdminUser } from '../../types/admin'
import AdminShell from './AdminShell.vue'

const StubView = defineComponent({ render: () => h('div', 'conteudo') })

async function mountShell(role: AdminUser['role']) {
  const pinia = createPinia()
  setActivePinia(pinia)

  const auth = useAdminAuthStore()
  auth.user = {
    id: 'user-1',
    name: 'Dono da Loja',
    email: 'dono@example.com',
    role,
  }

  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', name: 'home', component: StubView },
      { path: '/admin/login', name: 'admin-login', component: StubView },
      {
        path: '/admin',
        component: StubView,
        children: [
          { path: '', name: 'admin-dashboard', component: StubView },
          ...adminNavItems
            .filter((item) => item.name !== 'admin-dashboard')
            .map((item) => ({ path: item.name, name: item.name, component: StubView })),
          { path: 'sem-acesso', name: 'admin-forbidden', component: StubView },
        ],
      },
    ],
  })

  await router.push('/admin')
  await router.isReady()

  const wrapper = mount(AdminShell, {
    global: { plugins: [pinia, router] },
  })

  return { wrapper, router, auth }
}

describe('AdminShell', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('mostra o usuário, o papel e a navegação do operador', async () => {
    const { wrapper } = await mountShell('operator')

    expect(wrapper.text()).toContain('Dono da Loja')
    expect(wrapper.text()).toContain('Operador')
    expect(wrapper.get('nav[aria-label="Navegação do painel"]').text()).toContain('Produtos')
    expect(wrapper.get('nav[aria-label="Navegação do painel"]').text()).not.toContain('Usuários')
    expect(wrapper.get('nav[aria-label="Navegação principal do painel"]').text()).toContain(
      'Pedidos',
    )
  })

  it('mostra itens de owner no menu lateral', async () => {
    const { wrapper } = await mountShell('owner')

    expect(wrapper.get('nav[aria-label="Navegação do painel"]').text()).toContain('Usuários')
    expect(wrapper.get('nav[aria-label="Navegação do painel"]').text()).toContain('Auditoria')
  })

  it('encerra a sessão e volta para o login', async () => {
    const { wrapper, router, auth } = await mountShell('owner')
    const logout = vi.spyOn(auth, 'logout').mockResolvedValue()

    await wrapper.get('.admin-shell__logout').trigger('click')
    await flushPromises()

    expect(logout).toHaveBeenCalledTimes(1)
    expect(router.currentRoute.value.name).toBe('admin-login')
  })
})
