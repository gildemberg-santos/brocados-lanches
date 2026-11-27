import { ref } from 'vue'

import type { OrderEvent } from '../../services/realtime'
import { playNotificationSound } from './notificationSound'

export interface StatusNotification {
  id: string
  message: string
}

export interface StatusNotificationOptions {
  /**
   * Status já conhecido pela tela. `undefined` significa "ainda carregando" e
   * bloqueia a notificação; `null` significa "pedido não está nesta página".
   */
  currentStatus: (event: OrderEvent) => string | null | undefined
  labelForStatus: (status: string) => string
  /** Quando informado, filtra eventos de outros pedidos. */
  matchesOrder?: (event: OrderEvent) => boolean
  /** Injetado nos testes; em produção usa o Web Audio do navegador. */
  playSound?: () => boolean
}

/**
 * Converte apenas mudanças reais de status em notificação visual e sonora.
 * `connected`, reconexões, eventos duplicados e mensagens desconhecidas são ignorados.
 */
export function useOrderStatusNotifications(options: StatusNotificationOptions) {
  const notification = ref<StatusNotification | null>(null)
  const seen = new Set<string>()
  let sequence = 0

  /** Usado tanto pelo WebSocket quanto pela confirmação de uma ação local no painel. */
  function announce(event: OrderEvent): void {
    if (event.type !== 'order.updated' || !event.status) return
    if (options.matchesOrder && !options.matchesOrder(event)) return

    // `updatedAt` distingue duas transições para o mesmo status (ex.: cancelamentos).
    const key = `${event.orderId ?? event.publicId ?? ''}|${event.status}|${event.updatedAt ?? ''}`
    if (seen.has(key)) return
    seen.add(key)

    const orderNumber = event.orderNumber ? ` #${event.orderNumber}` : ''
    notification.value = {
      id: `${key}|${sequence++}`,
      message: `Pedido${orderNumber}: ${options.labelForStatus(event.status)}`,
    }
    ;(options.playSound ?? playNotificationSound)()
  }

  function handleEvent(event: OrderEvent): void {
    if (event.type !== 'order.updated' || !event.status) return
    const current = options.currentStatus(event)
    if (current === undefined || current === event.status) return
    announce(event)
  }

  function dismiss(): void {
    notification.value = null
  }

  return { notification, handleEvent, announce, dismiss }
}
