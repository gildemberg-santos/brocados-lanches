import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'

import AdminCatalogForm from './AdminCatalogForm.vue'

describe('AdminCatalogForm', () => {
  it('gera o slug a partir do nome e envia uma categoria', async () => {
    const wrapper = mount(AdminCatalogForm, {
      props: { kind: 'category', item: null, saving: false, error: '', success: '' },
    })

    await wrapper.find('#catalog-name').setValue('Hambúrgueres da Casa')
    await wrapper.findAll('.catalog-form__slug-button')[0]?.trigger('click')
    await wrapper.find('form').trigger('submit')

    expect(wrapper.emitted('submit')?.[0]).toEqual([
      {
        name: 'Hambúrgueres da Casa',
        slug: 'hamburgueres-da-casa',
        description: '',
        sortOrder: 0,
        active: true,
      },
    ])
  })

  it('bloqueia o envio com slug inválido', async () => {
    const wrapper = mount(AdminCatalogForm, {
      props: { kind: 'addon', item: null, saving: false, error: '', success: '' },
    })

    await wrapper.find('#catalog-name').setValue('Cheddar')
    await wrapper.find('#catalog-slug').setValue('Cheddar')
    await wrapper.find('#catalog-amount').setValue('3,50')
    await wrapper.find('form').trigger('submit')

    expect(wrapper.emitted('submit')).toBeUndefined()
    expect(wrapper.text()).toContain('letras minúsculas')
  })

  it('bloqueia o envio com valor monetário inválido', async () => {
    const wrapper = mount(AdminCatalogForm, {
      props: { kind: 'zone', item: null, saving: false, error: '', success: '' },
    })

    await wrapper.find('#catalog-name').setValue('Centro')
    await wrapper.find('#catalog-slug').setValue('centro')
    await wrapper.find('#catalog-amount').setValue('abc')
    await wrapper.find('form').trigger('submit')

    expect(wrapper.emitted('submit')).toBeUndefined()
    expect(wrapper.text()).toContain('valor válido')
  })

  it('converte preço em reais para centavos e mantém o item em edição', async () => {
    const wrapper = mount(AdminCatalogForm, {
      props: {
        kind: 'addon',
        item: {
          id: 'a'.repeat(24),
          name: 'Cheddar',
          slug: 'cheddar',
          priceCents: 350,
          sortOrder: 1,
          active: false,
          createdAt: '',
          updatedAt: '',
        },
        saving: false,
        error: '',
        success: '',
      },
    })

    expect((wrapper.find('#catalog-amount').element as HTMLInputElement).value).toBe('3,50')
    await wrapper.find('#catalog-amount').setValue('4,25')
    await wrapper.find('form').trigger('submit')

    expect(wrapper.emitted('submit')?.[0]?.[0]).toMatchObject({
      priceCents: 425,
      sortOrder: 1,
      active: false,
    })
  })
})
