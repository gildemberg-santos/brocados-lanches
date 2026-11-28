import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import PixPaymentPanel from './PixPaymentPanel.vue'
import { generatePixQrDataUri } from '../features/pix/qrcode'

vi.mock('../features/pix/qrcode', () => ({ generatePixQrDataUri: vi.fn() }))

const props = {
  brCode: '00020126...6304ABCD',
  amountCents: 4990,
  merchantName: 'Brocados Lanches',
  merchantCity: 'Aquiraz',
}

describe('PixPaymentPanel', () => {
  beforeEach(() => {
    vi.mocked(generatePixQrDataUri).mockReset()
    vi.mocked(generatePixQrDataUri).mockResolvedValue('data:image/png;base64,abc')
  })

  it('mostra o QR Code e o Pix Copia e Cola do mesmo payload', async () => {
    const wrapper = mount(PixPaymentPanel, { props })
    await flushPromises()

    expect(generatePixQrDataUri).toHaveBeenCalledWith(props.brCode)
    expect(wrapper.get('img').attributes('src')).toBe('data:image/png;base64,abc')
    expect((wrapper.get('[data-testid="pix-brcode"]').element as HTMLTextAreaElement).value).toBe(
      props.brCode,
    )
    expect(wrapper.text()).toContain('Brocados Lanches')
    expect(wrapper.text()).toContain('Aquiraz')
  })

  it('descarta QR antigo se o payload mudar durante a geração', async () => {
    let resolveFirst: (value: string) => void = () => {}
    const firstQr = new Promise<string>((resolve) => {
      resolveFirst = resolve
    })
    vi.mocked(generatePixQrDataUri)
      .mockReturnValueOnce(firstQr)
      .mockResolvedValue('data:image/png;base64,novo')

    const wrapper = mount(PixPaymentPanel, { props })
    await flushPromises()
    await wrapper.setProps({ brCode: '00020126...6304EF01' })
    await flushPromises()
    expect(wrapper.get('img').attributes('src')).toBe('data:image/png;base64,novo')

    resolveFirst('data:image/png;base64,antigo')
    await flushPromises()
    expect(wrapper.get('img').attributes('src')).toBe('data:image/png;base64,novo')
  })

  it('avisa quando a imagem não pode ser gerada', async () => {
    vi.mocked(generatePixQrDataUri).mockRejectedValue(new Error('falha'))
    const wrapper = mount(PixPaymentPanel, { props })
    await flushPromises()

    expect(wrapper.find('img').exists()).toBe(false)
    expect(wrapper.text()).toContain('Use o Pix Copia e Cola')
  })

  it('copia o código para a área de transferência', async () => {
    const writeText = vi.fn().mockResolvedValue(undefined)
    Object.defineProperty(navigator, 'clipboard', {
      configurable: true,
      value: { writeText },
    })
    const wrapper = mount(PixPaymentPanel, { props })
    await flushPromises()

    await wrapper.get('button').trigger('click')
    await flushPromises()

    expect(writeText).toHaveBeenCalledWith(props.brCode)
    expect(wrapper.text()).toContain('Código Pix copiado')
  })
})
