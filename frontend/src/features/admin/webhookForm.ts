export type WebhookFieldErrors = Record<'url', string | undefined>

const MAX_URL_LENGTH = 2048

/** Validação de UX; o servidor continua sendo a fonte da verdade. */
export function validateWebhookInput(url: string, active: boolean): WebhookFieldErrors {
  const value = url.trim()
  if (active && value.length === 0) {
    return { url: 'Informe a URL do webhook ou desative a integração.' }
  }
  if (value.length > MAX_URL_LENGTH) {
    return { url: `A URL deve ter no máximo ${MAX_URL_LENGTH} caracteres.` }
  }
  if (value.length > 0) {
    try {
      const parsed = new URL(value)
      if (!parsed.protocol) {
        return { url: 'Informe uma URL absoluta, incluindo o esquema.' }
      }
    } catch {
      return { url: 'Informe uma URL absoluta, incluindo o esquema.' }
    }
  }
  return { url: undefined }
}
