import { toDataURL } from 'qrcode'

import tokens from '../../../../design-system/tokens.json'

/**
 * Renderiza o mesmo BR Code em PNG dentro de uma Data URI. A imagem não muda
 * o payload: basta o app bancário ler o QR ou colar a string textual. As cores
 * vêm dos tokens do design system, sem introduzir padrão visual novo.
 */
export async function generatePixQrDataUri(brCode: string): Promise<string> {
  return toDataURL(brCode, {
    errorCorrectionLevel: 'M',
    margin: 4,
    width: 360,
    type: 'image/png',
    color: { dark: tokens.colors.background.primary, light: tokens.colors.text.main },
  })
}
