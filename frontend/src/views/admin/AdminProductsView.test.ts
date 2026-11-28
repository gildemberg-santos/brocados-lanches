import { createPinia, setActivePinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import AdminProductsView from './AdminProductsView.vue'
import {
  fetchAdminAddons,
  fetchAdminCategories,
  fetchAdminProducts,
  updateAdminProduct,
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

const product = {
  id: 'd'.repeat(24),
  categoryId: category.id,
  name: 'Duplo Cheddar',
  slug: 'duplo-cheddar',
  description: 'Descrição',
  priceCents: 2990,
  imageUrl: '',
  imageAlt: '',
  allergens: [],
  available: true,
  featured: false,
  active: true,
  sortOrder: 0,
  optionGroups: [],
  addonIds: [addon.id],
  createdAt: '',
  updatedAt: '',
}

beforeEach(() => {
  setActivePinia(createPinia())
  vi.clearAllMocks()
  vi.mocked(fetchAdminProducts).mockResolvedValue({
    data: [product],
    meta: { page: 1, pageSize: 20, total: 1, totalPages: 1 },
  })
  vi.mocked(fetchAdminCategories).mockResolvedValue({
    data: [category],
    meta: { page: 1, pageSize: 100, total: 1, totalPages: 1 },
  })
  vi.mocked(fetchAdminAddons).mockResolvedValue({
    data: [addon],
    meta: { page: 1, pageSize: 100, total: 1, totalPages: 1 },
  })
})

describe('AdminProductsView', () => {
  it('carrega produtos, categorias e adicionais', async () => {
    const wrapper = mount(AdminProductsView)
    await flushPromises()

    expect(wrapper.get('h1').text()).toBe('Produtos')
    expect(wrapper.text()).toContain('Duplo Cheddar')
    expect(wrapper.text()).toContain('Hambúrgueres')
    expect(wrapper.text()).toContain('R$\u00a029,90')
    expect(wrapper.find('#product-category-filter').findAll('option').length).toBe(2)
  })

  it('renderiza contagem zero quando o backend devolve listas nulas', async () => {
    vi.mocked(fetchAdminProducts).mockResolvedValue({
      data: [
        {
          ...product,
          allergens: null as unknown as typeof product.allergens,
          optionGroups: null as unknown as typeof product.optionGroups,
          addonIds: null as unknown as typeof product.addonIds,
        },
      ],
      meta: { page: 1, pageSize: 20, total: 1, totalPages: 1 },
    })
    const wrapper = mount(AdminProductsView)
    await flushPromises()

    expect(wrapper.text()).toContain('0 grupo(s) · 0 adicional(is)')
  })

  it('abre o formulário com categoria, adicional e preço em reais', async () => {
    const wrapper = mount(AdminProductsView)
    await flushPromises()

    await wrapper
      .findAll('.product-row__button')
      .filter((button) => button.text() === 'Editar')[0]
      ?.trigger('click')

    expect(wrapper.get('[role="dialog"]').attributes('aria-modal')).toBe('true')
    expect(wrapper.get('[data-testid="admin-modal"] h2').text()).toBe('Editar produto')
    expect((wrapper.find('#product-price').element as HTMLInputElement).value).toBe('29,90')
    expect(wrapper.find('#product-category').findAll('option').length).toBe(2)
    expect(wrapper.find('fieldset').text()).toContain('Cheddar')
  })

  it('fecha o modal com ESC sem salvar', async () => {
    const wrapper = mount(AdminProductsView)
    await flushPromises()

    await wrapper
      .findAll('.product-row__button')
      .filter((button) => button.text() === 'Editar')[0]
      ?.trigger('click')
    expect(wrapper.find('[role="dialog"]').exists()).toBe(true)

    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }))
    await flushPromises()

    expect(wrapper.find('[role="dialog"]').exists()).toBe(false)
    expect(updateAdminProduct).not.toHaveBeenCalled()
  })

  it('filtra por categoria', async () => {
    const wrapper = mount(AdminProductsView)
    await flushPromises()

    await wrapper.find('#product-category-filter').setValue(category.id)
    await flushPromises()

    expect(fetchAdminProducts).toHaveBeenLastCalledWith(
      expect.objectContaining({ categoryId: category.id, page: 1 }),
    )
  })

  it('desativa um produto', async () => {
    vi.mocked(updateAdminProduct).mockResolvedValue({ ...product, active: false })
    const wrapper = mount(AdminProductsView)
    await flushPromises()

    await wrapper
      .findAll('.product-row__button')
      .filter((button) => button.text() === 'Desativar')[0]
      ?.trigger('click')
    await flushPromises()

    expect(updateAdminProduct).toHaveBeenCalledWith(product.id, { active: false })
  })
})
