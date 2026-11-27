/** Regras de UI para o envio de imagens de produto no painel. */

export const MAX_IMAGE_UPLOAD_BYTES = 2 * 1024 * 1024

export const ACCEPTED_IMAGE_TYPES: readonly string[] = [
  'image/jpeg',
  'image/png',
  'image/webp',
  'image/avif',
]

export const IMAGE_UPLOAD_ACCEPT = 'image/jpeg,image/png,image/webp,image/avif'

/** Devolve uma mensagem em português quando o arquivo não pode ser enviado. */
export function validateImageFile(file: File): string {
  if (!ACCEPTED_IMAGE_TYPES.includes(file.type)) {
    return 'Use uma imagem JPEG, PNG, WebP ou AVIF.'
  }
  if (file.size === 0) {
    return 'O arquivo está vazio.'
  }
  if (file.size > MAX_IMAGE_UPLOAD_BYTES) {
    return 'A imagem deve ter no máximo 2 MB.'
  }
  return ''
}
