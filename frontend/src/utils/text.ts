/**
 * Normaliza texto para busca: remove acentos, ignora caixa e espaços nas
 * extremidades. A normalização é apenas de apresentação; o servidor continua
 * responsável pelos dados que retorna.
 */
export function normalizeText(value: string): string {
  return value
    .normalize('NFD')
    .replace(/\p{Diacritic}/gu, '')
    .toLowerCase()
    .trim()
}

export function splitSearchTerms(value: string): string[] {
  const normalized = normalizeText(value)

  return normalized.length === 0 ? [] : normalized.split(/\s+/)
}
