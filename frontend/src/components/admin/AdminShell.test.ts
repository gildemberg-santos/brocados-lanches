import { flushPromises, mount, type VueWrapper } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { defineComponent, h } from 'vue'
import { createMemoryHistory, createRouter } from 'vue-router'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

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
    attachTo: document.body,
    global: { plugins: [pinia, router] },
  })

  return { wrapper, router, auth }
}

async function openDrawer(wrapper: VueWrapper): Promise<void> {
  await wrapper.get('[data-testid="admin-menu-button"]').trigger('click')
  await flushPromises()
}

describe('AdminShell', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    document.body.innerHTML = ''
  })

  it('mostra o usuário, o papel e a navegação do operador', async () => {
    const { wrapper } = await mountShell('operator')

    expect(wrapper.text()).toContain('Dono da Loja')
    expect(wrapper.text()).toContain('Operador')
    expect(wrapper.get('nav[aria-label="Navegação do painel"]').text()).toContain('Produtos')
    expect(wrapper.get('nav[aria-label="Navegação do painel"]').text()).not.toContain('Usuários')
    expect(wrapper.find('[data-testid="admin-drawer"]').exists()).toBe(false)

    await openDrawer(wrapper)

    const drawer = wrapper.get('[data-testid="admin-drawer"]')
    expect(drawer.get('nav').text()).toContain('Pedidos')
    expect(drawer.text()).not.toContain('Usuários')
    expect(wrapper.get('[data-testid="admin-menu-button"]').attributes('aria-expanded')).toBe(
      'true',
    )
  })

  it('mostra itens de owner no menu lateral e no drawer', async () => {
    const { wrapper } = await mountShell('owner')

    expect(wrapper.get('nav[aria-label="Navegação do painel"]').text()).toContain('Usuários')
    expect(wrapper.get('nav[aria-label="Navegação do painel"]').text()).toContain('Auditoria')

    await openDrawer(wrapper)

    const drawer = wrapper.get('[data-testid="admin-drawer"]')
    expect(drawer.text()).toContain('Usuários')
    expect(drawer.text()).toContain('Auditoria')
    expect(drawer.get('nav').findAll('a')).toHaveLength(adminNavItems.length)
  })

  it('move o foco para o drawer ao abrir e o devolve ao botão ao fechar com Escape', async () => {
    const { wrapper } = await mountShell('owner')
    const menuButton = wrapper.get('[data-testid="admin-menu-button"]')

    await openDrawer(wrapper)

    expect(document.activeElement).toBe(wrapper.get('[data-testid="admin-drawer-close"]').element)

    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }))
    await flushPromises()

    expect(wrapper.find('[data-testid="admin-drawer"]').exists()).toBe(false)
    expect(document.activeElement).toBe(menuButton.element)
    expect(menuButton.attributes('aria-expanded')).toBe('false')
  })

  it('fecha o drawer ao clicar no fundo', async () => {
    const { wrapper } = await mountShell('operator')

    await openDrawer(wrapper)
    await wrapper.get('[data-testid="admin-drawer-backdrop"]').trigger('click')
    await flushPromises()

    expect(wrapper.find('[data-testid="admin-drawer"]').exists()).toBe(false)
  })

  it('mantém o foco dentro do drawer ao tabular', async () => {
    const { wrapper } = await mountShell('owner')

    await openDrawer(wrapper)

    const closeButton = wrapper.get('[data-testid="admin-drawer-close"]').element
    const links = wrapper.get('[data-testid="admin-drawer"]').findAll('nav a')
    const lastLink = links[links.length - 1]?.element
    expect(lastLink).toBeDefined()

    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'Tab', shiftKey: true }))
    expect(document.activeElement).toBe(lastLink)

    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'Tab' }))
    expect(document.activeElement).toBe(closeButton)
  })

  it('oferece um destino único e focável para o skip-link', async () => {
    const { wrapper } = await mountShell('owner')
    const target = wrapper.get('#conteudo')

    expect(target.element.tagName).toBe('DIV')
    expect(target.attributes('tabindex')).toBe('-1')
    expect(wrapper.findAll('#conteudo')).toHaveLength(1)
    expect(wrapper.findAll('main')).toHaveLength(0)
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
