import { createPinia, setActivePinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import AdminCategoriesView from './AdminCategoriesView.vue'
import {
  createAdminCategory,
  fetchAdminCategories,
  updateAdminCategory,
} from '../../services/adminCatalog'

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

const category = {
  id: 'a'.repeat(24),
  name: 'Hambúrgueres',
  slug: 'hamburgeres',
  description: '',
  sortOrder: 0,
  active: true,
  createdAt: '',
  updatedAt: '',
}

beforeEach(() => {
  setActivePinia(createPinia())
  vi.clearAllMocks()
  vi.mocked(fetchAdminCategories).mockResolvedValue({
    data: [category],
    meta: { page: 1, pageSize: 20, total: 1, totalPages: 1 },
  })
})

describe('AdminCategoriesView', () => {
  it('carrega e exibe as categorias', async () => {
    const wrapper = mount(AdminCategoriesView)
    await flushPromises()

    expect(wrapper.get('h1').text()).toBe('Categorias')
    expect(wrapper.text()).toContain('Hambúrgueres')
    expect(wrapper.text()).toContain('hamburgeres')
  })

  it('abre o formulário e cria uma categoria', async () => {
    vi.mocked(createAdminCategory).mockResolvedValue(category)
    const wrapper = mount(AdminCategoriesView)
    await flushPromises()

    await wrapper
      .findAll('button')
      .filter((button) => button.text() === 'Nova categoria')[0]
      ?.trigger('click')
    await wrapper.find('#catalog-name').setValue('Novo item')
    await wrapper.find('#catalog-slug').setValue('novo-item')
    await wrapper.find('.catalog-form').trigger('submit')
    await flushPromises()

    expect(createAdminCategory).toHaveBeenCalledWith(
      expect.objectContaining({ name: 'Novo item', slug: 'novo-item' }),
    )
  })

  it('desativa uma categoria existente', async () => {
    vi.mocked(updateAdminCategory).mockResolvedValue({ ...category, active: false })
    const wrapper = mount(AdminCategoriesView)
    await flushPromises()

    await wrapper
      .findAll('.catalog-row__button')
      .filter((button) => button.text() === 'Desativar')[0]
      ?.trigger('click')
    await flushPromises()

    expect(updateAdminCategory).toHaveBeenCalledWith(category.id, { active: false })
  })
})
