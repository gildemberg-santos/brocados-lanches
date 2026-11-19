import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { fetchHealth } from '../services/health'
import HomeView from './HomeView.vue'

vi.mock('../services/health', () => ({
  fetchHealth: vi.fn(),
}))

const mockedFetchHealth = vi.mocked(fetchHealth)

describe('HomeView', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('shows the connected state after a successful health check', async () => {
    mockedFetchHealth.mockResolvedValue({
      status: 'ok',
      checks: { mongo: 'ok' },
    })

    const wrapper = mount(HomeView)
    await flushPromises()

    expect(wrapper.text()).toContain('Estrutura inicial')
    expect(wrapper.text()).toContain('API: conectada')
    expect(wrapper.text()).toContain('MongoDB: ok')
  })

  it('shows the offline state when the health check fails', async () => {
    mockedFetchHealth.mockRejectedValue(new Error('network error'))

    const wrapper = mount(HomeView)
    await flushPromises()

    expect(wrapper.text()).toContain('API: indisponível')
    expect(wrapper.text()).toContain('Não foi possível falar com a API.')
  })
})
