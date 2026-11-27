import { describe, expect, it, vi } from 'vitest'

import type { OrderEvent } from '../../services/realtime'
import { useOrderStatusNotifications } from './statusNotifications'

function setup(currentStatus: () => string | null | undefined = () => 'confirmed') {
  const playSound = vi.fn(() => true)
  const labels: Record<string, string> = {
    confirmed: 'Confirmado',
    ready: 'Pronto',
    cancelled: 'Cancelado',
  }
  const notifications = useOrderStatusNotifications({
    currentStatus,
    labelForStatus: (status) => labels[status] ?? status,
    matchesOrder: (event) => !event.publicId || event.publicId === 'public-123',
    playSound,
  })
  return { ...notifications, playSound }
}

describe('useOrderStatusNotifications', () => {
  it('notifica e toca som quando o status realmente muda', () => {
    const { notification, handleEvent, playSound } = setup()

    handleEvent({
      type: 'order.updated',
      publicId: 'public-123',
      orderNumber: 1042,
      status: 'ready',
      updatedAt: '2026-11-27T12:00:00Z',
    })

    expect(notification.value).toEqual({
      id: expect.stringContaining('public-123|ready|2026-11-27T12:00:00Z'),
      message: `Pedido #${1042}: Pronto`,
    })
    expect(playSound).toHaveBeenCalledTimes(1)
  })

  it('ignora conexão, mensagens desconhecidas e evento sem status', () => {
    const { notification, handleEvent, playSound } = setup()
    const events: OrderEvent[] = [
      { type: 'connected' },
      { type: 'order.deleted', status: 'cancelled' },
      { type: 'order.updated' },
    ]

    for (const event of events) handleEvent(event)

    expect(notification.value).toBeNull()
    expect(playSound).not.toHaveBeenCalled()
  })

  it('ignora carga inicial, pedido de outro cliente e status já conhecido', () => {
    const loading = setup(() => undefined)
    const otherOrder = setup()
    const knownStatus = setup(() => 'ready')
    const event: OrderEvent = {
      type: 'order.updated',
      publicId: 'public-123',
      orderNumber: 1042,
      status: 'ready',
    }

    loading.handleEvent(event)
    otherOrder.handleEvent({ ...event, publicId: 'outro-pedido' })
    knownStatus.handleEvent(event)

    expect(loading.notification.value).toBeNull()
    expect(otherOrder.notification.value).toBeNull()
    expect(knownStatus.notification.value).toBeNull()
    expect(loading.playSound).not.toHaveBeenCalled()
    expect(otherOrder.playSound).not.toHaveBeenCalled()
    expect(knownStatus.playSound).not.toHaveBeenCalled()
  })

  it('não repete notificação para o mesmo evento', () => {
    const { notification, handleEvent, playSound } = setup()
    const event: OrderEvent = {
      type: 'order.updated',
      orderId: 'order-1',
      status: 'cancelled',
      updatedAt: '2026-11-27T12:05:00Z',
    }

    handleEvent(event)
    const firstId = notification.value?.id
    handleEvent(event)

    expect(notification.value?.id).toBe(firstId)
    expect(playSound).toHaveBeenCalledTimes(1)
  })

  it('anuncia ação local e não repete o aviso quando o WebSocket confirma', () => {
    const { notification, handleEvent, announce, playSound } = setup(() => 'ready')
    const event: OrderEvent = {
      type: 'order.updated',
      orderId: 'order-1',
      orderNumber: 1042,
      status: 'ready',
      updatedAt: '2026-11-27T12:10:00Z',
    }

    announce(event)
    expect(notification.value?.message).toBe(`Pedido #${1042}: Pronto`)

    handleEvent(event)
    expect(playSound).toHaveBeenCalledTimes(1)
  })

  it('permite fechar a notificação', () => {
    const { notification, handleEvent, dismiss } = setup()

    handleEvent({ type: 'order.updated', publicId: 'public-123', status: 'ready' })
    dismiss()

    expect(notification.value).toBeNull()
  })
})
