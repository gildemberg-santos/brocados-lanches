import { describe, expect, it } from 'vitest'

import { validateImageFile } from './imageUpload'

function imageFile(type: string, size: number): File {
  return new File([new Uint8Array(size)], 'produto.png', { type })
}

describe('validateImageFile', () => {
  it('aceita os formatos permitidos dentro do limite', () => {
    for (const type of ['image/jpeg', 'image/png', 'image/webp', 'image/avif']) {
      expect(validateImageFile(imageFile(type, 128))).toBe('')
    }
  })

  it('rejeita formato fora da lista', () => {
    expect(validateImageFile(imageFile('text/html', 128))).toBe(
      'Use uma imagem JPEG, PNG, WebP ou AVIF.',
    )
  })

  it('rejeita arquivo vazio', () => {
    expect(validateImageFile(imageFile('image/png', 0))).toBe('O arquivo está vazio.')
  })

  it('rejeita arquivo acima de 2 MB', () => {
    expect(validateImageFile(imageFile('image/png', 2 * 1024 * 1024 + 1))).toBe(
      'A imagem deve ter no máximo 2 MB.',
    )
  })
})
