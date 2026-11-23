import { createPinia, setActivePinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import AdminAccountView from './AdminAccountView.vue'
import { useAdminAuthStore } from '../../stores/adminAuth'

const push = vi.fn()
vi.mock('vue-router', () => ({
  useRouter: () => ({ push }),
}))

beforeEach(() => {
  setActivePinia(createPinia())
  push.mockClear()
})

describe('AdminAccountView', () => {
  it('mostra os dados do usuário da sessão', () => {
    const auth = useAdminAuthStore()
    auth.user = {
      id: 'u1',
      name: 'Dono',
      email: 'dono@example.com',
      role: 'owner',
    }

    const wrapper = mount(AdminAccountView)

    expect(wrapper.text()).toContain('Dono')
    expect(wrapper.text()).toContain('dono@example.com')
    expect(wrapper.text()).toContain('MFA não fazem parte da V1')
  })

  it('encerra a sessão e volta para o login', async () => {
    const auth = useAdminAuthStore()
    auth.user = {
      id: 'u1',
      name: 'Dono',
      email: 'dono@example.com',
      role: 'owner',
    }
    const logout = vi.spyOn(auth, 'logout').mockResolvedValue()
    const wrapper = mount(AdminAccountView)

    await wrapper.findAll('button')[0]?.trigger('click')
    await flushPromises()

    expect(logout).toHaveBeenCalledOnce()
    expect(push).toHaveBeenCalledWith({ name: 'admin-login' })
  })
})
