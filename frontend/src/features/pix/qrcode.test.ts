import { beforeEach, describe, expect, it, vi } from 'vitest'

import tokens from '../../../../design-system/tokens.json'

import { generatePixQrDataUri } from './qrcode'

const toDataURL = vi.fn()

vi.mock('qrcode', () => ({ toDataURL: (...args: unknown[]) => toDataURL(...args) }))

describe('geração do QR Code do Pix', () => {
  beforeEach(() => {
    toDataURL.mockReset()
    toDataURL.mockResolvedValue('data:image/png;base64,iVBORw0KGgo=')
  })

  it('gera uma Data URI PNG a partir do BR Code', async () => {
    const brCode = '00020126...6304ABCD'

    await expect(generatePixQrDataUri(brCode)).resolves.toBe('data:image/png;base64,iVBORw0KGgo=')
    expect(toDataURL).toHaveBeenCalledWith(brCode, {
      errorCorrectionLevel: 'M',
      margin: 4,
      width: 360,
      type: 'image/png',
      color: { dark: tokens.colors.background.primary, light: tokens.colors.text.main },
    })
  })

  it('propaga falha de renderização para o componente avisar o cliente', async () => {
    toDataURL.mockRejectedValue(new Error('canvas indisponível'))

    await expect(generatePixQrDataUri('payload')).rejects.toThrow('canvas indisponível')
  })
})
