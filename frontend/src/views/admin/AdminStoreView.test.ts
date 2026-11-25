import { createPinia, setActivePinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import AdminStoreView from './AdminStoreView.vue'
import { fetchAdminStore, updateAdminStore } from '../../services/adminStore'
import { fetchStoreStatus } from '../../services/store'
import type { AdminStoreSettings } from '../../types/adminCatalog'
import type { PublicStoreStatus } from '../../types/store'

vi.mock('../../services/adminStore', () => ({
  fetchAdminStore: vi.fn(),
  updateAdminStore: vi.fn(),
}))
vi.mock('../../services/store', () => ({ fetchStoreStatus: vi.fn() }))

const settings: AdminStoreSettings = {
  name: 'Brocados Lanches',
  isOpen: true,
  openingHours: [
    { day: 'sunday', closed: true, intervals: [] },
    { day: 'monday', closed: false, intervals: [{ open: '11:00', close: '15:00' }] },
  ],
  deliveryTime: '35-50 min',
  whatsapp: '+5585999999999',
  instagram: '@brocados',
  address: 'Rua de Teste, 10',
  createdAt: '',
  updatedAt: '',
}

const storeStatus: PublicStoreStatus = {
  name: 'Brocados Lanches',
  status: 'closed',
  isOpenManual: true,
  timezone: 'America/Fortaleza',
  deliveryTime: '35-50 min',
  whatsapp: '+5585999999999',
  instagram: '@brocados',
  address: 'Rua de Teste, 10',
  openingHours: [],
  nextOpeningAt: '2026-11-23T02:00:00Z',
  updatedAt: '2026-11-22T12:00:00Z',
}

beforeEach(() => {
  setActivePinia(createPinia())
  vi.clearAllMocks()
  vi.mocked(fetchAdminStore).mockResolvedValue(settings)
  vi.mocked(fetchStoreStatus).mockResolvedValue(storeStatus)
})

describe('AdminStoreView', () => {
  it('carrega as configurações e o horário existente', async () => {
    const wrapper = mount(AdminStoreView)
    await flushPromises()

    expect(wrapper.get('h1').text()).toBe('Loja')
    expect((wrapper.find('#store-name').element as HTMLInputElement).value).toBe('Brocados Lanches')
    expect((wrapper.find('#store-whatsapp').element as HTMLInputElement).value).toBe(
      '+5585999999999',
    )
    expect(wrapper.text()).toContain('Segunda-feira')
    const times = wrapper.findAll('input[type="time"]')
    expect(times).toHaveLength(2)
    expect((times[0]?.element as HTMLInputElement).value).toBe('11:00')
  })

  it('mostra o status efetivo calculado pelo servidor', async () => {
    const wrapper = mount(AdminStoreView)
    await flushPromises()

    const badge = wrapper.get('[data-testid="store-effective-status"]')
    expect(badge.text()).toContain('Loja fechada')
    expect(badge.text()).toContain('America/Fortaleza')
    expect(badge.text()).toContain('Próxima abertura')
  })

  it('recarrega o status efetivo depois de salvar', async () => {
    vi.mocked(updateAdminStore).mockResolvedValue(settings)
    const wrapper = mount(AdminStoreView)
    await flushPromises()
    expect(fetchStoreStatus).toHaveBeenCalledOnce()

    await wrapper.find('form').trigger('submit')
    await flushPromises()

    expect(fetchStoreStatus).toHaveBeenCalledTimes(2)
  })

  it('bloqueia o envio com WhatsApp inválido', async () => {
    const wrapper = mount(AdminStoreView)
    await flushPromises()

    await wrapper.find('#store-whatsapp').setValue('(85) 99999-9999')
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    expect(updateAdminStore).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('formato internacional')
    expect(wrapper.get('#store-whatsapp').attributes('aria-invalid')).toBe('true')
    expect(wrapper.get('#store-whatsapp').attributes('aria-describedby')).toBe(
      'store-whatsapp-error',
    )
    expect(wrapper.get('#store-whatsapp-error').text()).toContain('formato internacional')
  })

  it('associa erros de horário ao intervalo correspondente', async () => {
    const wrapper = mount(AdminStoreView)
    await flushPromises()

    await wrapper.find('input[type="time"]').setValue('')
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    expect(updateAdminStore).not.toHaveBeenCalled()
    expect(wrapper.get('input[type="time"]').attributes('aria-invalid')).toBe('true')
    expect(wrapper.get('input[type="time"]').attributes('aria-describedby')).toBe(
      'store-openingHours-1-intervals-0-open-error',
    )
    expect(wrapper.get('#store-openingHours-1-intervals-0-open-error').text()).toBeTruthy()
  })

  it('salva as configurações alteradas', async () => {
    vi.mocked(updateAdminStore).mockResolvedValue({ ...settings, isOpen: false })
    const wrapper = mount(AdminStoreView)
    await flushPromises()

    const toggle = wrapper.find('.store-form__check--highlight input')
    await toggle.setValue(false)
    await wrapper.find('form').trigger('submit')
    await flushPromises()

    expect(updateAdminStore).toHaveBeenCalledWith(expect.objectContaining({ isOpen: false }))
    expect(wrapper.text()).toContain('Configurações salvas.')
  })

  it('oferece os sete dias quando a loja ainda não tem horários cadastrados', async () => {
    vi.mocked(fetchAdminStore).mockResolvedValue({ ...settings, openingHours: [] })
    const wrapper = mount(AdminStoreView)
    await flushPromises()

    expect(wrapper.findAll('.store-form__day')).toHaveLength(7)
    expect(wrapper.text()).toContain('Domingo')
    expect(wrapper.text()).toContain('Sábado')

    await wrapper.find('form').trigger('submit')
    await flushPromises()

    expect(updateAdminStore).toHaveBeenCalledWith(
      expect.objectContaining({
        openingHours: expect.arrayContaining([
          expect.objectContaining({ day: 'sunday', closed: true }),
        ]),
      }),
    )
  })

  it('adiciona e remove intervalos em um dia aberto', async () => {
    const wrapper = mount(AdminStoreView)
    await flushPromises()

    const day = wrapper.findAll('.store-form__day')[1]
    if (!day) {
      throw new Error('Dia não encontrado')
    }
    const addButton = day
      .findAll('button')
      .filter((button) => button.text() === 'Adicionar intervalo')[0]
    await addButton?.trigger('click')
    expect(wrapper.findAll('input[type="time"]')).toHaveLength(4)

    const removeButtons = wrapper
      .findAll('.store-form__day')[1]
      ?.findAll('button')
      .filter((button) => button.text() === 'Remover')
    await removeButtons?.[1]?.trigger('click')
    expect(wrapper.findAll('input[type="time"]')).toHaveLength(2)
  })
})
