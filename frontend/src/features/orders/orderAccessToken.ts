export const ORDER_ACCESS_TOKEN_KEY = 'brocados:order-access-token'

const TOKEN_VERSION = 1
const TOKEN_LENGTH = 43
const TOKEN_PATTERN = /^[A-Za-z0-9_-]{43}$/

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}

/** O token é opaco: só conferimos a forma antes de guardar ou enviar. */
export function isValidOrderAccessToken(value: string): boolean {
  return value.length === TOKEN_LENGTH && TOKEN_PATTERN.test(value)
}

/** Chamado somente depois que o pedido foi gravado no servidor. */
export function saveOrderAccessToken(token: string): void {
  if (typeof localStorage === 'undefined') {
    return
  }

  const normalized = token.trim()
  if (!isValidOrderAccessToken(normalized)) {
    return
  }

  try {
    localStorage.setItem(
      ORDER_ACCESS_TOKEN_KEY,
      JSON.stringify({ version: TOKEN_VERSION, token: normalized }),
    )
  } catch {
    // Histórico é conveniência: falha de armazenamento não afeta o pedido.
  }
}

export function loadOrderAccessToken(): string | null {
  if (typeof localStorage === 'undefined') {
    return null
  }

  try {
    const raw = localStorage.getItem(ORDER_ACCESS_TOKEN_KEY)
    if (!raw) {
      return null
    }

    const parsed: unknown = JSON.parse(raw)
    if (!isRecord(parsed) || parsed.version !== TOKEN_VERSION) {
      return null
    }

    const token = typeof parsed.token === 'string' ? parsed.token.trim() : ''
    return isValidOrderAccessToken(token) ? token : null
  } catch {
    return null
  }
}

export function hasOrderAccessToken(): boolean {
  return loadOrderAccessToken() !== null
}

export function clearOrderAccessToken(): void {
  if (typeof localStorage === 'undefined') {
    return
  }

  try {
    localStorage.removeItem(ORDER_ACCESS_TOKEN_KEY)
  } catch {
    // Nada a fazer se o navegador bloquear a remoção.
  }
}
