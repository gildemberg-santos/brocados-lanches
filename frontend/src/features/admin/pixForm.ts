import {
  countSanitizedPixChars,
  hasPixControlChars,
  isValidPixKey,
  sanitizePixText,
} from '../pix/brcode'
import type { AdminPixInput } from '../../types/adminPix'

export type PixFieldErrors = Partial<Record<keyof AdminPixInput, string>>

const NAME_MAX_LENGTH = 25
const CITY_MAX_LENGTH = 15

/** Validação de UX; o servidor continua sendo a fonte da verdade. */
export function validatePixInput(input: AdminPixInput): PixFieldErrors {
  const errors: PixFieldErrors = {}
  const key = input.key.trim()

  if (input.active && key === '') {
    errors.key = 'Informe a chave Pix ou desative o QR Code.'
  } else if (key !== '' && !isValidPixKey(key)) {
    errors.key = 'Use uma chave Pix válida e sem espaços: CPF, CNPJ, e-mail, telefone E.164 ou EVP.'
  }

  const merchantName = sanitizePixText(input.merchantName, NAME_MAX_LENGTH)
  if (input.active && hasPixControlChars(input.merchantName)) {
    errors.merchantName = 'Remova os caracteres de controle do nome.'
  } else if (input.active && merchantName === '') {
    errors.merchantName = 'Informe o nome do recebedor.'
  } else if (countSanitizedPixChars(input.merchantName) > NAME_MAX_LENGTH) {
    errors.merchantName = `Use até ${NAME_MAX_LENGTH} caracteres após remover acentos e símbolos.`
  }

  const merchantCity = sanitizePixText(input.merchantCity, CITY_MAX_LENGTH)
  if (input.active && hasPixControlChars(input.merchantCity)) {
    errors.merchantCity = 'Remova os caracteres de controle da cidade.'
  } else if (input.active && merchantCity === '') {
    errors.merchantCity = 'Informe a cidade do recebedor.'
  } else if (countSanitizedPixChars(input.merchantCity) > CITY_MAX_LENGTH) {
    errors.merchantCity = `Use até ${CITY_MAX_LENGTH} caracteres após remover acentos e símbolos.`
  }

  return errors
}
