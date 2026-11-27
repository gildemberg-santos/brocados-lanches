import { ApiError } from '../../services/http'
import type { AdminPageMeta } from '../../types/admin'

export const DEFAULT_ORDER_HISTORY_PAGE_SIZE = 20

export const EMPTY_ORDER_HISTORY_META: AdminPageMeta = {
  page: 1,
  pageSize: DEFAULT_ORDER_HISTORY_PAGE_SIZE,
  total: 0,
  totalPages: 0,
}

/** Não devolve detalhe do erro nem dados pessoais na mensagem da tela. */
export function orderHistoryErrorMessage(error: unknown): string {
  if (error instanceof ApiError) {
    if (error.code === 'RATE_LIMITED') {
      return 'Muitas consultas em seguida. Aguarde um momento e tente novamente.'
    }
    if (error.code === 'VALIDATION_ERROR') {
      return 'Não foi possível consultar o histórico neste navegador. Faça um novo pedido para criar um acesso.'
    }
  }
  return 'Não foi possível carregar seus pedidos. Tente novamente.'
}
