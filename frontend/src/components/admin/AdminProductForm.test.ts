import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'

import AdminProductForm from './AdminProductForm.vue'

const categories = [
  {
    id: 'a'.repeat(24),
    name: 'Hambúrgueres',
    slug: 'hamburgeres',
    description: '',
    sortOrder: 0,
    active: true,
    createdAt: '',
    updatedAt: '',
  },
]

const addons = [
  {
    id: 'b'.repeat(24),
    name: 'Cheddar',
    slug: 'cheddar',
    priceCents: 350,
    active: true,
    sortOrder: 0,
    createdAt: '',
    updatedAt: '',
  },
]

function mountForm() {
  return mount(AdminProductForm, {
    props: {
      item: null,
      categories,
      addons,
      saving: false,
      error: '',
      success: '',
    },
  })
}

describe('AdminProductForm', () => {
  it('envia produto com grupos, opções e adicionais em centavos', async () => {
    const wrapper = mountForm()

    await wrapper.find('#product-category').setValue('a'.repeat(24))
    await wrapper.find('#product-name').setValue('Duplo Cheddar')
    await wrapper.find('#product-slug').setValue('duplo-cheddar')
    await wrapper.find('#product-description').setValue('Dois hambúrgueres com cheddar.')
    await wrapper.find('#product-price').setValue('29,90')
    await wrapper.findAll('fieldset')[0]?.find('input[type="checkbox"]').setValue(true)
    await wrapper
      .findAll('button')
      .filter((button) => button.text() === 'Adicionar grupo')[0]
      ?.trigger('click')
    await wrapper
      .findAll('button')
      .filter((button) => button.text() === 'Adicionar opção')[0]
      ?.trigger('click')
    await wrapper.find('input[placeholder="Nome do grupo"]').setValue('Ponto da carne')
    await wrapper.find('input[placeholder="Nome da opção"]').setValue('Ao ponto')
    await wrapper.findAll('input[placeholder="0,00"]')[1]?.setValue('2,50')
    await wrapper.find('form').trigger('submit')

    const emitted = wrapper.emitted('submit')
    expect(emitted).toBeDefined()
    const payload = emitted?.[0]?.[0] as {
      priceCents: number
      addonIds: string[]
      optionGroups: Array<{ options: Array<{ priceCents: number }> }>
    }
    expect(payload.priceCents).toBe(2990)
    expect(payload.addonIds).toEqual(['b'.repeat(24)])
    expect(payload.optionGroups[0]?.options[0]?.priceCents).toBe(250)
  })

  it('bloqueia o envio com preço ou preço de opção inválido', async () => {
    const wrapper = mountForm()

    await wrapper.find('#product-category').setValue('a'.repeat(24))
    await wrapper.find('#product-name').setValue('Duplo Cheddar')
    await wrapper.find('#product-slug').setValue('duplo-cheddar')
    await wrapper.find('#product-description').setValue('Dois hambúrgueres com cheddar.')
    await wrapper.find('#product-price').setValue('abc')
    await wrapper
      .findAll('button')
      .filter((button) => button.text() === 'Adicionar grupo')[0]
      ?.trigger('click')
    await wrapper
      .findAll('button')
      .filter((button) => button.text() === 'Adicionar opção')[0]
      ?.trigger('click')
    await wrapper.find('input[placeholder="Nome do grupo"]').setValue('Ponto da carne')
    await wrapper.find('input[placeholder="Nome da opção"]').setValue('Ao ponto')
    await wrapper.findAll('input[placeholder="0,00"]')[1]?.setValue('xyz')
    await wrapper.find('form').trigger('submit')

    expect(wrapper.emitted('submit')).toBeUndefined()
    expect(wrapper.text()).toContain('Informe um preço válido em reais.')
  })

  it('bloqueia o envio sem categoria e sem descrição', async () => {
    const wrapper = mountForm()

    await wrapper.find('form').trigger('submit')

    expect(wrapper.emitted('submit')).toBeUndefined()
    expect(wrapper.text()).toContain('Selecione uma categoria')
    expect(wrapper.text()).toContain('Informe a descrição')
  })
})
