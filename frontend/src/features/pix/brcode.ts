/**
 * Gerador local de BR Code (EMV QRCPS-MPM) para Pix estático.
 *
 * O payload é montado somente com dados públicos devolvidos por GET /store e
 * com o total calculado no servidor. Nenhuma chave, taxa ou total é inventado
 * no navegador.
 */

const PIX_GUI = 'br.gov.bcb.pix'
const PIX_KEY_MAX_LENGTH = 77
const PIX_AMOUNT_MAX_LENGTH = 13

// Os quatro tipos de chave aceitos pelo arranjo Pix: CPF/CNPJ (com ou sem
// máscara), e-mail, telefone em E.164 e EVP (UUID).
const PIX_KEY_PATTERNS: readonly RegExp[] = [
  /^(?:[0-9]{11}|[0-9]{14}|[0-9]{3}\.[0-9]{3}\.[0-9]{3}-[0-9]{2}|[0-9]{2}\.[0-9]{3}\.[0-9]{3}\/[0-9]{4}-[0-9]{2})$/,
  /^[A-Za-z0-9._%+-]+@[A-Za-z0-9-]+(?:\.[A-Za-z0-9-]+)*\.[A-Za-z]{2,}$/,
  /^\+[1-9][0-9]{7,14}$/,
  /^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$/,
]
const ACCENT_REPLACEMENTS: Readonly<Record<string, string>> = {
  á: 'a',
  à: 'a',
  ã: 'a',
  â: 'a',
  ä: 'a',
  é: 'e',
  ê: 'e',
  è: 'e',
  ë: 'e',
  í: 'i',
  ì: 'i',
  î: 'i',
  ï: 'i',
  ó: 'o',
  ô: 'o',
  õ: 'o',
  ò: 'o',
  ö: 'o',
  ú: 'u',
  ù: 'u',
  û: 'u',
  ü: 'u',
  ç: 'c',
  ñ: 'n',
  Á: 'A',
  À: 'A',
  Ã: 'A',
  Â: 'A',
  Ä: 'A',
  É: 'E',
  Ê: 'E',
  È: 'E',
  Ë: 'E',
  Í: 'I',
  Ì: 'I',
  Î: 'I',
  Ï: 'I',
  Ó: 'O',
  Ô: 'O',
  Õ: 'O',
  Ò: 'O',
  Ö: 'O',
  Ú: 'U',
  Ù: 'U',
  Û: 'U',
  Ü: 'U',
  Ç: 'C',
  Ñ: 'N',
}

export interface PixBrCodeInput {
  /** Chave Pix completa: CPF, CNPJ, e-mail, telefone E.164 ou EVP. */
  key: string
  merchantName: string
  merchantCity: string
  /** Total já calculado pelo servidor, em centavos. Ausente = valor livre. */
  amountCents?: number
  /** TXID alfanumérico de até 25 caracteres. Padrão: ***. */
  txid?: string
}

/** Formata um campo TLV com tag e tamanho de dois dígitos. */
export function formatTlv(tag: string, value: string): string {
  return `${tag}${String(value.length).padStart(2, '0')}${value}`
}

/** BTC/CRC16-CCITT-FALSE: polinômio 0x1021, inicial 0xFFFF, maiúsculas. */
export function calculateCrc16(payload: string): string {
  const bytes = new TextEncoder().encode(`${payload}6304`)
  let crc = 0xffff
  for (const byte of bytes) {
    crc ^= byte << 8
    for (let bit = 0; bit < 8; bit += 1) {
      crc = (crc & 0x8000) !== 0 ? ((crc << 1) ^ 0x1021) & 0xffff : (crc << 1) & 0xffff
    }
  }
  return `6304${crc.toString(16).toUpperCase().padStart(4, '0')}`
}

/** Detecta caracteres de controle rejeitados pela validação do servidor. */
export function hasPixControlChars(value: string): boolean {
  return /[\u0000-\u001f\u007f-\u009f]/.test(value)
}

/** Conta os caracteres que permaneceriam após a higienização do campo. */
export function countSanitizedPixChars(value: string): number {
  const normalized = Array.from(value.trim())
    .map((character) => ACCENT_REPLACEMENTS[character] ?? character)
    .join('')
  return Array.from(normalized.replace(/[^a-zA-Z0-9 ]/g, '')).length
}

/** Remove acentos e caracteres especiais, limitando o campo ao padrão EMV. */
export function sanitizePixText(value: string, maxLength: number): string {
  let normalized = ''
  for (const character of value.trim()) {
    normalized += ACCENT_REPLACEMENTS[character] ?? character
  }
  return normalized
    .replace(/[^a-zA-Z0-9 ]/g, '')
    .slice(0, maxLength)
    .trim()
}

/** Valida os quatro formatos de chave aceitos pelo arranjo Pix. */
export function isValidPixKey(key: string): boolean {
  const trimmed = key.trim()
  if (trimmed === '' || trimmed.length > PIX_KEY_MAX_LENGTH) {
    return false
  }
  return PIX_KEY_PATTERNS.some((pattern) => pattern.test(trimmed))
}

/** Converte centavos inteiros para o formato exigido pela tag 54. */
export function formatPixAmount(amountCents: number): string {
  if (!Number.isSafeInteger(amountCents) || amountCents <= 0) {
    throw new Error('O valor do Pix deve ser um número inteiro de centavos positivo.')
  }
  const reais = Math.floor(amountCents / 100)
  const centavos = amountCents % 100
  const amount = `${reais}.${String(centavos).padStart(2, '0')}`
  if (amount.length > PIX_AMOUNT_MAX_LENGTH) {
    throw new Error('O valor do Pix excede o limite da tag 54.')
  }
  return amount
}

/** Monta o payload completo, pronto para Pix Copia e Cola. */
export function buildPixBrCode(input: PixBrCodeInput): string {
  const key = input.key.trim()
  if (!isValidPixKey(key)) {
    throw new Error('A chave Pix configurada é inválida.')
  }

  const merchantName = sanitizePixText(input.merchantName, 25)
  const merchantCity = sanitizePixText(input.merchantCity, 15)
  if (merchantName === '' || merchantCity === '') {
    throw new Error('Os dados do recebedor Pix estão incompletos.')
  }

  const txid =
    input.txid === undefined || input.txid === '***'
      ? '***'
      : sanitizePixText(input.txid, 25) || '***'

  const merchantAccountInfo = formatTlv('00', PIX_GUI) + formatTlv('01', key)
  let payload =
    formatTlv('00', '01') +
    formatTlv('26', merchantAccountInfo) +
    formatTlv('52', '0000') +
    formatTlv('53', '986')

  if (input.amountCents !== undefined) {
    payload += formatTlv('54', formatPixAmount(input.amountCents))
  }

  payload +=
    formatTlv('58', 'BR') +
    formatTlv('59', merchantName) +
    formatTlv('60', merchantCity) +
    formatTlv('62', formatTlv('05', txid))

  return payload + calculateCrc16(payload)
}
