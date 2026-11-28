import { createPinia, setActivePinia } from 'pinia'
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import AdminPixView from './AdminPixView.vue'
import { fetchAdminPix, updateAdminPix } from '../../services/adminPix'

vi.mock('../../services/adminPix', () => ({ fetchAdminPix: vi.fn(), updateAdminPix: vi.fn() }))

const settings = {
  active: false,
  key: '',
  merchantName: '',
  merchantCity: '',
  createdAt: '2026-11-27T12:00:00Z',
  updatedAt: '2026-11-27T12:00:00Z',
}

beforeEach(() => {
  setActivePinia(createPinia())
  vi.clearAllMocks()
  vi.mocked(fetchAdminPix).mockResolvedValue(settings)
})

describe('AdminPixView', () => {
  it('carrega a configuração existente', async () => {
    const wrapper = mount(AdminPixView)
    await flushPromises()

    expect(wrapper.get('h1').text()).toBe('Pix estático')
    expect((wrapper.get('#pix-key').element as HTMLInputElement).value).toBe('')
    expect(wrapper.text()).toContain('confira o comprovante')
  })

  it('salva chave, nome e cidade do recebedor', async () => {
    vi.mocked(updateAdminPix).mockResolvedValue({
      ...settings,
      active: true,
      key: '123e4567-e89b-12d3-a456-426614174000',
      merchantName: 'Brocados Lanches',
      merchantCity: 'Aquiraz',
    })
    const wrapper = mount(AdminPixView)
    await flushPromises()

    await wrapper.get('input[type="checkbox"]').setValue(true)
    await wrapper.get('#pix-key').setValue('123e4567-e89b-12d3-a456-426614174000')
    await wrapper.get('#pix-name').setValue('Brocados Lanches')
    await wrapper.get('#pix-city').setValue('Aquiraz')
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(updateAdminPix).toHaveBeenCalledWith({
      active: true,
      key: '123e4567-e89b-12d3-a456-426614174000',
      merchantName: 'Brocados Lanches',
      merchantCity: 'Aquiraz',
    })
    expect(wrapper.text()).toContain('salva com sucesso')
  })

  it('não recarrega a configuração depois de salvar', async () => {
    vi.mocked(updateAdminPix).mockResolvedValue({
      ...settings,
      active: true,
      key: '123e4567-e89b-12d3-a456-426614174000',
      merchantName: 'Brocados Lanches',
      merchantCity: 'Aquiraz',
    })
    const wrapper = mount(AdminPixView)
    await flushPromises()

    await wrapper.get('input[type="checkbox"]').setValue(true)
    await wrapper.get('#pix-key').setValue('123e4567-e89b-12d3-a456-426614174000')
    await wrapper.get('#pix-name').setValue('Brocados Lanches')
    await wrapper.get('#pix-city').setValue('Aquiraz')
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(fetchAdminPix).toHaveBeenCalledTimes(1)
  })

  it('aceita símbolos que desaparecem na higienização do BR Code', async () => {
    const merchantName = '***Brocados Lanches*****'
    const merchantCity = '*****Aquiraz*****'
    vi.mocked(updateAdminPix).mockResolvedValue({
      ...settings,
      active: true,
      key: '123e4567-e89b-12d3-a456-426614174000',
      merchantName,
      merchantCity,
    })
    const wrapper = mount(AdminPixView)
    await flushPromises()

    await wrapper.get('input[type="checkbox"]').setValue(true)
    await wrapper.get('#pix-key').setValue('123e4567-e89b-12d3-a456-426614174000')
    await wrapper.get('#pix-name').setValue(merchantName)
    await wrapper.get('#pix-city').setValue(merchantCity)
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(updateAdminPix).toHaveBeenCalledWith({
      active: true,
      key: '123e4567-e89b-12d3-a456-426614174000',
      merchantName,
      merchantCity,
    })
  })

  it('bloqueia o envio quando a chave é inválida', async () => {
    const wrapper = mount(AdminPixView)
    await flushPromises()

    await wrapper.get('input[type="checkbox"]').setValue(true)
    await wrapper.get('#pix-key').setValue('chave com espaço')
    await wrapper.get('form').trigger('submit')
    await flushPromises()

    expect(updateAdminPix).not.toHaveBeenCalled()
    expect(wrapper.get('#pix-key-error').text()).toContain('sem espaços')
    expect(wrapper.get('#pix-key').attributes('aria-invalid')).toBe('true')
  })
})
