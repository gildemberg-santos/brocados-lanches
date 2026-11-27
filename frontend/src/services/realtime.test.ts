import { afterEach, describe, expect, it, vi } from 'vitest'

import {
  resolveOrderEventsURL,
  subscribeOrderEvents,
  type OrderEvent,
  type OrderEventSocket,
} from './realtime'

class FakeSocket implements OrderEventSocket {
  static instances: FakeSocket[] = []

  onopen: (() => void) | null = null
  onmessage: ((event: { data: string }) => void) | null = null
  onclose: (() => void) | null = null
  onerror: (() => void) | null = null
  closed = false

  constructor(readonly url: string) {
    FakeSocket.instances.push(this)
  }

  emitOpen(): void {
    this.onopen?.()
  }

  emitMessage(data: string): void {
    this.onmessage?.({ data })
  }

  emitError(): void {
    this.onerror?.()
  }

  emitClose(): void {
    this.onclose?.()
  }

  close(): void {
    this.closed = true
  }
}

afterEach(() => {
  vi.useRealTimers()
  FakeSocket.instances = []
})

describe('resolveOrderEventsURL', () => {
  it('usa a origem atual quando não há base configurada', () => {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    expect(resolveOrderEventsURL('/api/v1/orders/abc/events')).toBe(
      `${protocol}//${window.location.host}/api/v1/orders/abc/events`,
    )
  })
})

describe('subscribeOrderEvents', () => {
  it('entrega somente eventos válidos e ignora mensagens desconhecidas', () => {
    const events: OrderEvent[] = []
    subscribeOrderEvents('/api/v1/orders/abc/events', (event) => events.push(event), {
      socketFactory: (url) => new FakeSocket(url),
    })

    const socket = FakeSocket.instances[0]!
    expect(socket.url).toContain('/api/v1/orders/abc/events')

    socket.emitMessage('{')
    socket.emitMessage('null')
    socket.emitMessage('{"status":"ready"}')
    socket.emitMessage(
      '{"type":"order.updated","orderNumber":42,"status":"ready","updatedAt":"2026-11-26T20:00:00Z"}',
    )

    expect(events).toEqual([
      {
        type: 'order.updated',
        orderId: undefined,
        publicId: undefined,
        orderNumber: 42,
        status: 'ready',
        updatedAt: '2026-11-26T20:00:00Z',
      },
    ])
  })

  it('informa o estado da conexão para a interface', async () => {
    vi.useFakeTimers()
    const statuses: string[] = []
    const dispose = subscribeOrderEvents('/api/v1/admin/orders/events', () => {}, {
      socketFactory: (url) => new FakeSocket(url),
      onStatus: (status) => statuses.push(status),
    })

    const socket = FakeSocket.instances[0]!
    socket.emitOpen()
    socket.emitClose()
    await vi.advanceTimersByTimeAsync(800)

    expect(statuses).toEqual(['connecting', 'connected', 'reconnecting', 'reconnecting'])
    dispose()
  })

  it('reconecta com backoff e zera o atraso após conexão bem-sucedida', async () => {
    vi.useFakeTimers()
    subscribeOrderEvents('/api/v1/admin/orders/events', () => {}, {
      socketFactory: (url) => new FakeSocket(url),
    })

    const first = FakeSocket.instances[0]!
    first.emitOpen()
    first.emitClose()

    await vi.advanceTimersByTimeAsync(799)
    expect(FakeSocket.instances).toHaveLength(1)
    await vi.advanceTimersByTimeAsync(1)
    expect(FakeSocket.instances).toHaveLength(2)

    const second = FakeSocket.instances[1]!
    second.emitClose()
    await vi.advanceTimersByTimeAsync(1600)
    expect(FakeSocket.instances).toHaveLength(3)

    const third = FakeSocket.instances[2]!
    third.emitOpen()
    third.emitClose()
    await vi.advanceTimersByTimeAsync(800)
    expect(FakeSocket.instances).toHaveLength(4)
  })

  it('trata erro como queda e reconecta sem duplicar temporizadores', async () => {
    vi.useFakeTimers()
    subscribeOrderEvents('/api/v1/admin/orders/events', () => {}, {
      socketFactory: (url) => new FakeSocket(url),
    })

    const socket = FakeSocket.instances[0]!
    socket.emitError()
    expect(socket.closed).toBe(true)
    socket.emitClose()
    socket.emitClose()

    await vi.advanceTimersByTimeAsync(800)
    expect(FakeSocket.instances).toHaveLength(2)
    await vi.advanceTimersByTimeAsync(2000)
    expect(FakeSocket.instances).toHaveLength(2)
  })

  it('dispose fecha a conexão e impede novas tentativas', async () => {
    vi.useFakeTimers()
    const dispose = subscribeOrderEvents('/api/v1/orders/abc/events', () => {}, {
      socketFactory: (url) => new FakeSocket(url),
    })

    const socket = FakeSocket.instances[0]!
    dispose()
    socket.emitClose()

    expect(socket.closed).toBe(true)
    await vi.advanceTimersByTimeAsync(30_000)
    expect(FakeSocket.instances).toHaveLength(1)
  })
})
