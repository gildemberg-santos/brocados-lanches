import { createPinia, setActivePinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import AdminAddonsView from './AdminAddonsView.vue'
import { fetchAdminAddons, updateAdminAddon } from '../../services/adminCatalog'

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

const addon = {
  id: 'b'.repeat(24),
  name: 'Cheddar',
  slug: 'cheddar',
  priceCents: 350,
  active: true,
  sortOrder: 0,
  createdAt: '',
  updatedAt: '',
}

beforeEach(() => {
  setActivePinia(createPinia())
  vi.clearAllMocks()
  vi.mocked(fetchAdminAddons).mockResolvedValue({
    data: [addon],
    meta: { page: 1, pageSize: 20, total: 1, totalPages: 1 },
  })
})

describe('AdminAddonsView', () => {
  it('exibe adicionais com preço formatado', async () => {
    const wrapper = mount(AdminAddonsView)
    await flushPromises()

    expect(wrapper.get('h1').text()).toBe('Adicionais')
    expect(wrapper.text()).toContain('Cheddar')
    expect(wrapper.text()).toContain('R$\u00a03,50')
  })

  it('abre o formulário de edição com o preço em reais', async () => {
    const wrapper = mount(AdminAddonsView)
    await flushPromises()

    await wrapper
      .findAll('.catalog-row__button')
      .filter((button) => button.text() === 'Editar')[0]
      ?.trigger('click')

    expect(wrapper.get('[role="dialog"]').attributes('aria-modal')).toBe('true')
    expect(wrapper.get('[data-testid="admin-modal"] h2').text()).toBe('Editar adicional')
    expect((wrapper.find('#catalog-amount').element as HTMLInputElement).value).toBe('3,50')
  })

  it('desativa um adicional', async () => {
    vi.mocked(updateAdminAddon).mockResolvedValue({ ...addon, active: false })
    const wrapper = mount(AdminAddonsView)
    await flushPromises()

    await wrapper
      .findAll('.catalog-row__button')
      .filter((button) => button.text() === 'Desativar')[0]
      ?.trigger('click')
    await flushPromises()

    expect(updateAdminAddon).toHaveBeenCalledWith(addon.id, { active: false })
  })
})
