import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { defineComponent, h } from 'vue'
import { createMemoryHistory, createRouter } from 'vue-router'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { useAdminAuthStore } from '../../stores/adminAuth'
import AdminLoginView from './AdminLoginView.vue'

const StubView = defineComponent({ render: () => h('div') })

async function mountLogin(query = '') {
  const pinia = createPinia()
  setActivePinia(pinia)
  const auth = useAdminAuthStore()

  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', name: 'home', component: StubView },
      { path: '/admin/login', name: 'admin-login', component: StubView },
      { path: '/admin', name: 'admin-dashboard', component: StubView },
      { path: '/admin/pedidos', name: 'admin-orders', component: StubView },
    ],
  })

  await router.push('/admin/login' + query)
  await router.isReady()

  const wrapper = mount(AdminLoginView, {
    global: { plugins: [pinia, router] },
  })

  return { wrapper, auth, router }
}

describe('AdminLoginView', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('exibe formulário com rótulos e autocomplete', async () => {
    const { wrapper } = await mountLogin()

    expect(wrapper.get('label[for="admin-email"]').text()).toBe('E-mail')
    expect(wrapper.get('label[for="admin-password"]').text()).toBe('Senha')
    expect(wrapper.get('#admin-email').attributes('autocomplete')).toBe('username')
    expect(wrapper.get('#admin-password').attributes('autocomplete')).toBe('current-password')
  })

  it('valida campos vazios sem chamar a API', async () => {
    const { wrapper, auth } = await mountLogin()
    const login = vi.spyOn(auth, 'login')

    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(login).not.toHaveBeenCalled()
    expect(wrapper.get('[role="alert"]').text()).toContain('Informe o e-mail e a senha')
  })

  it('mostra erro genérico quando a API recusa as credenciais', async () => {
    const { wrapper, auth } = await mountLogin()
    vi.spyOn(auth, 'login').mockImplementation(async () => {
      auth.error = 'E-mail ou senha incorretos.'
      return false
    })

    await wrapper.get('#admin-email').setValue('dono@example.com')
    await wrapper.get('#admin-password').setValue('senha-errada')
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(wrapper.get('[role="alert"]').text()).toBe('E-mail ou senha incorretos.')
    expect(wrapper.get('#admin-password').element).toBeInstanceOf(HTMLInputElement)
    expect((wrapper.get('#admin-password').element as HTMLInputElement).value).toBe('senha-errada')
  })

  it('redireciona para o destino original após login', async () => {
    const { wrapper, auth, router } = await mountLogin('?redirect=%2Fadmin%2Fpedidos')
    vi.spyOn(auth, 'login').mockResolvedValue(true)

    await wrapper.get('#admin-email').setValue('dono@example.com')
    await wrapper.get('#admin-password').setValue('senha-forte-123')
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(router.currentRoute.value.path).toBe('/admin/pedidos')
  })

  it('ignora redirect que não começa com /admin', async () => {
    const { wrapper, auth, router } = await mountLogin('?redirect=https%3A%2F%2Fevil.example')
    vi.spyOn(auth, 'login').mockResolvedValue(true)

    await wrapper.get('#admin-email').setValue('dono@example.com')
    await wrapper.get('#admin-password').setValue('senha-forte-123')
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(router.currentRoute.value.path).toBe('/admin')
  })
})
