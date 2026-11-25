import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { fetchStoreStatus } from '../services/store'
import type { PublicStoreStatus } from '../types/store'
import ContactView from './ContactView.vue'

vi.mock('../services/store', () => ({ fetchStoreStatus: vi.fn() }))

const mockedFetchStoreStatus = vi.mocked(fetchStoreStatus)

const status: PublicStoreStatus = {
  name: 'Brocados Lanches',
  status: 'open',
  isOpenManual: true,
  timezone: 'America/Fortaleza',
  deliveryTime: '30-90 min',
  whatsapp: '+5585999999999',
  instagram: '@brocados',
  address: 'Rua de Teste, 10',
  openingHours: [
    {
      day: 'monday',
      closed: false,
      intervals: [{ open: '10:00', close: '22:00' }],
    },
    { day: 'tuesday', closed: true, intervals: [] },
  ],
  nextOpeningAt: null,
  updatedAt: '2026-11-22T12:00:00Z',
}

async function mountContact() {
  const pinia = createPinia()
  setActivePinia(pinia)

  return mount(ContactView, { global: { plugins: [pinia] } })
}

describe('ContactView', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows a loading state before the store data arrives', async () => {
    mockedFetchStoreStatus.mockResolvedValue(status)

    const wrapper = await mountContact()

    expect(wrapper.text()).toContain('Carregando contato')
  })

  it('renders only the contact data returned by the server', async () => {
    mockedFetchStoreStatus.mockResolvedValue(status)

    const wrapper = await mountContact()
    await flushPromises()

    expect(wrapper.text()).toContain('Brocados Lanches')
    expect(wrapper.text()).toContain('Aberta no momento')
    expect(wrapper.text()).toContain('30-90 min')
    expect(wrapper.text()).toContain('+5585999999999')
    expect(wrapper.text()).toContain('@brocados')
    expect(wrapper.text()).toContain('Rua de Teste, 10')
    expect(wrapper.text()).toContain('Segunda-feira: 10:00–22:00')
    expect(wrapper.text()).toContain('Terça-feira: Fechado')

    const whatsapp = wrapper.get('a[href="https://wa.me/5585999999999"]')
    expect(whatsapp.attributes('rel')).toBe('noopener noreferrer')
    expect(whatsapp.attributes('target')).toBe('_blank')
    expect(wrapper.find('a[href="@brocados"]').exists()).toBe(false)
  })

  it('keeps a safe state when the store request fails and retries on demand', async () => {
    mockedFetchStoreStatus
      .mockRejectedValueOnce(new Error('mongo down'))
      .mockResolvedValueOnce(status)

    const wrapper = await mountContact()
    await flushPromises()

    expect(wrapper.text()).toContain('Não foi possível carregar')
    expect(wrapper.text()).not.toContain('mongo down')
    expect(wrapper.text()).not.toContain('+5585999999999')

    await wrapper.get('button').trigger('click')
    await flushPromises()

    expect(mockedFetchStoreStatus).toHaveBeenCalledTimes(2)
    expect(wrapper.text()).toContain('+5585999999999')
  })
})
