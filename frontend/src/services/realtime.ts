/**
 * Cliente WebSocket compartilhado pelo acompanhamento público e pelo painel.
 * O canal só transporta sinais mínimos (identificador, número, status e data);
 * a recarga dos dados completos continua feita pelo servidor via HTTP.
 */

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? ''
const INITIAL_RECONNECT_DELAY_MS = 800
const MAX_RECONNECT_DELAY_MS = 10_000

export interface OrderEvent {
  type: string
  orderId?: string
  publicId?: string
  orderNumber?: number
  status?: string
  updatedAt?: string
}

/** Subconjunto da API WebSocket usado pelo cliente; facilita testes sem rede. */
export interface OrderEventSocket {
  onopen: (() => void) | null
  onmessage: ((event: { data: string }) => void) | null
  onclose: (() => void) | null
  onerror: (() => void) | null
  close(): void
}

export type OrderEventConnectionStatus = 'connecting' | 'connected' | 'reconnecting'

export interface SubscribeOptions {
  socketFactory?: (url: string) => OrderEventSocket
  onStatus?: (status: OrderEventConnectionStatus) => void
}

export function resolveOrderEventsURL(path: string): string {
  const base = String(API_BASE_URL).trim().replace(/\/$/, '')
  if (/^wss?:\/\//i.test(base)) return `${base}${path}`
  if (/^https?:\/\//i.test(base)) return `${base.replace(/^http/i, 'ws')}${path}`
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  return `${protocol}//${window.location.host}${path}`
}

function parseOrderEvent(raw: string): OrderEvent | null {
  try {
    const value: unknown = JSON.parse(raw)
    if (typeof value !== 'object' || value === null) return null
    const candidate = value as Record<string, unknown>
    if (typeof candidate.type !== 'string') return null
    return {
      type: candidate.type,
      orderId: typeof candidate.orderId === 'string' ? candidate.orderId : undefined,
      publicId: typeof candidate.publicId === 'string' ? candidate.publicId : undefined,
      orderNumber: typeof candidate.orderNumber === 'number' ? candidate.orderNumber : undefined,
      status: typeof candidate.status === 'string' ? candidate.status : undefined,
      updatedAt: typeof candidate.updatedAt === 'string' ? candidate.updatedAt : undefined,
    }
  } catch {
    return null
  }
}

/**
 * Abre o canal e reconecta com backoff exponencial até `dispose()` ser chamado.
 * Erros de parse e mensagens desconhecidas são ignorados sem derrubar a conexão.
 */
export function subscribeOrderEvents(
  path: string,
  onEvent: (event: OrderEvent) => void,
  options: SubscribeOptions = {},
): () => void {
  const createSocket =
    options.socketFactory ?? ((url: string) => new WebSocket(url) as OrderEventSocket)
  const url = resolveOrderEventsURL(path)
  let socket: OrderEventSocket | null = null
  let timer: ReturnType<typeof setTimeout> | null = null
  let disposed = false
  let attempt = 0

  function scheduleReconnect(): void {
    if (disposed || timer !== null) return
    const delay = Math.min(INITIAL_RECONNECT_DELAY_MS * 2 ** attempt, MAX_RECONNECT_DELAY_MS)
    attempt += 1
    timer = setTimeout(() => {
      timer = null
      connect()
    }, delay)
  }

  function connect(): void {
    if (disposed) return
    options.onStatus?.(attempt === 0 ? 'connecting' : 'reconnecting')
    const current = createSocket(url)
    socket = current
    current.onopen = () => {
      attempt = 0
      options.onStatus?.('connected')
    }
    current.onmessage = (message) => {
      const event = parseOrderEvent(message.data)
      if (event) onEvent(event)
    }
    current.onerror = () => {
      // O navegador dispara `close` logo após `error`; encerrar aqui evita
      // depender de implementações que não fecham sozinhas.
      if (socket === current) current.close()
    }
    current.onclose = () => {
      if (socket !== current) return
      socket = null
      current.onopen = null
      current.onmessage = null
      current.onclose = null
      current.onerror = null
      options.onStatus?.('reconnecting')
      scheduleReconnect()
    }
  }

  function dispose(): void {
    disposed = true
    if (timer !== null) {
      clearTimeout(timer)
      timer = null
    }
    const current = socket
    socket = null
    if (current) {
      current.onopen = null
      current.onmessage = null
      current.onclose = null
      current.onerror = null
      current.close()
    }
  }

  connect()
  return dispose
}
