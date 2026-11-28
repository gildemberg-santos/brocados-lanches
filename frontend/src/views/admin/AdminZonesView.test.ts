import { createPinia, setActivePinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import AdminZonesView from './AdminZonesView.vue'
import { fetchAdminZones, updateAdminZone } from '../../services/adminCatalog'

vi.mock('../../services/adminCatalog', () => ({
  fetchAdminCategories: vi.fn(),
  fetchAdminAddons: vi.fn(),
  fetchAdminZones: vi.fn(),
  fetchAdminProducts: vi.fn(),
  createAdminCategory: vi.fn(),
  createAdminAddon: vi.fn(),
  createAdminZone: vi.fn(),
  createAdminProduct: vi.fn(),
  updateAdminCategory: vi.fn(),
  updateAdminAddon: vi.fn(),
  updateAdminZone: vi.fn(),
  updateAdminProduct: vi.fn(),
}))

const zone = {
  id: 'c'.repeat(24),
  name: 'Centro',
  slug: 'centro',
  feeCents: 500,
  active: true,
  sortOrder: 0,
  createdAt: '',
  updatedAt: '',
}

beforeEach(() => {
  setActivePinia(createPinia())
  vi.clearAllMocks()
  vi.mocked(fetchAdminZones).mockResolvedValue({
    data: [zone],
    meta: { page: 1, pageSize: 20, total: 1, totalPages: 1 },
  })
})

describe('AdminZonesView', () => {
  it('exibe local e taxa formatada', async () => {
    const wrapper = mount(AdminZonesView)
    await flushPromises()

    expect(wrapper.get('h1').text()).toBe('Locais e taxas')
    expect(wrapper.text()).toContain('Centro')
    expect(wrapper.text()).toContain('R$\u00a05,00')
  })

  it('edita a taxa em reais', async () => {
    const wrapper = mount(AdminZonesView)
    await flushPromises()

    await wrapper
      .findAll('.catalog-row__button')
      .filter((button) => button.text() === 'Editar')[0]
      ?.trigger('click')

    expect(wrapper.get('[role="dialog"]').attributes('aria-modal')).toBe('true')
    expect(wrapper.get('[data-testid="admin-modal"] h2').text()).toBe('Editar local e taxa')
    expect((wrapper.find('#catalog-amount').element as HTMLInputElement).value).toBe('5,00')
  })

  it('desativa uma zona', async () => {
    vi.mocked(updateAdminZone).mockResolvedValue({ ...zone, active: false })
    const wrapper = mount(AdminZonesView)
    await flushPromises()

    await wrapper
      .findAll('.catalog-row__button')
      .filter((button) => button.text() === 'Desativar')[0]
      ?.trigger('click')
    await flushPromises()

    expect(updateAdminZone).toHaveBeenCalledWith(zone.id, { active: false })
  })
})
