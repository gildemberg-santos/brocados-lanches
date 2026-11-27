import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import AdminProductForm from './AdminProductForm.vue'
import { uploadProductImage } from '../../services/adminCatalog'

vi.mock('../../services/adminCatalog', () => ({
  uploadProductImage: vi.fn(),
}))

const mockedUploadProductImage = vi.mocked(uploadProductImage)

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

async function fillRequiredFields(wrapper: ReturnType<typeof mountForm>): Promise<void> {
  await wrapper.find('#product-category').setValue('a'.repeat(24))
  await wrapper.find('#product-name').setValue('Duplo Cheddar')
  await wrapper.find('#product-slug').setValue('duplo-cheddar')
  await wrapper.find('#product-description').setValue('Dois hambúrgueres com cheddar.')
  await wrapper.find('#product-price').setValue('29,90')
}

function selectImageFile(wrapper: ReturnType<typeof mountForm>, file: File): Promise<void> {
  const input = wrapper.get('#product-image-file').element as HTMLInputElement
  Object.defineProperty(input, 'files', { value: [file], configurable: true })
  return wrapper.get('#product-image-file').trigger('change')
}

describe('AdminProductForm', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })
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
    expect(wrapper.get('#product-price').attributes('aria-invalid')).toBe('true')
    expect(wrapper.get('#product-price').attributes('aria-describedby')).toBe(
      'product-priceCents-error',
    )
    expect(wrapper.get('#product-option-0-0-price').attributes('aria-invalid')).toBe('true')
    expect(wrapper.get('#product-option-0-0-price').attributes('aria-describedby')).toBe(
      'product-optionGroups-0-options-0-priceCents-error',
    )
  })

  it('bloqueia o envio sem categoria e sem descrição', async () => {
    const wrapper = mountForm()

    await wrapper.find('form').trigger('submit')

    expect(wrapper.emitted('submit')).toBeUndefined()
    expect(wrapper.text()).toContain('Selecione uma categoria')
    expect(wrapper.text()).toContain('Informe a descrição')
    expect(wrapper.get('#product-category').attributes('aria-invalid')).toBe('true')
    expect(wrapper.get('#product-category').attributes('aria-describedby')).toBe(
      'product-categoryId-error',
    )
    expect(wrapper.get('#product-description').attributes('aria-invalid')).toBe('true')
    expect(wrapper.get('#product-description').attributes('aria-describedby')).toBe(
      'product-description-error',
    )
  })

  it('aceita caminho relativo de imagem já cadastrado no banco', async () => {
    const wrapper = mountForm()
    await fillRequiredFields(wrapper)
    await wrapper.find('#product-image').setValue('/products/burger.png')
    await wrapper.find('#product-image-alt').setValue('Hambúrguer com cheddar')

    await wrapper.find('form').trigger('submit')

    const payload = wrapper.emitted('submit')?.[0]?.[0] as { imageUrl: string }
    expect(payload.imageUrl).toBe('/products/burger.png')
    expect(wrapper.find('#product-image').attributes('type')).toBe('text')
    expect(wrapper.find('.product-form__preview').attributes('src')).toBe('/products/burger.png')
  })

  it('faz upload da imagem e preenche o campo com a URL devolvida', async () => {
    const url = `/api/v1/uploads/${'c'.repeat(24)}`
    mockedUploadProductImage.mockResolvedValue({
      id: 'c'.repeat(24),
      url,
      contentType: 'image/png',
      sizeBytes: 128,
    })
    const wrapper = mountForm()
    const file = new File(['imagem'], 'burger.png', { type: 'image/png' })

    await selectImageFile(wrapper, file)
    await flushPromises()

    expect(mockedUploadProductImage).toHaveBeenCalledWith(file)
    expect((wrapper.find('#product-image').element as HTMLInputElement).value).toBe(url)
    expect(wrapper.text()).toContain('Imagem enviada.')
    expect(wrapper.find('#product-image-file').attributes('disabled')).toBeUndefined()
  })

  it('rejeita arquivo inválido no cliente sem chamar a API', async () => {
    const wrapper = mountForm()
    const file = new File(['<script>'], 'pagina.html', { type: 'text/html' })

    await selectImageFile(wrapper, file)
    await flushPromises()

    expect(mockedUploadProductImage).not.toHaveBeenCalled()
    expect(wrapper.text()).toContain('Use uma imagem JPEG, PNG, WebP ou AVIF.')
    expect(wrapper.get('#product-image-file').attributes('aria-invalid')).toBe('true')
    expect(wrapper.get('#product-image-file').attributes('aria-describedby')).toBe(
      'product-image-file-error',
    )
  })

  it('mostra mensagem de erro quando o upload falha', async () => {
    mockedUploadProductImage.mockRejectedValue(new Error('erro de rede'))
    const wrapper = mountForm()
    const file = new File(['imagem'], 'burger.png', { type: 'image/png' })

    await selectImageFile(wrapper, file)
    await flushPromises()

    expect(wrapper.text()).toContain('Não foi possível enviar a imagem. Tente novamente.')
    expect((wrapper.find('#product-image').element as HTMLInputElement).value).toBe('')
  })
})
