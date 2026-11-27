/** Configuração da integração de webhook de notificações de pedidos. */
export interface AdminWebhookSettings {
  url: string
  active: boolean
  createdAt: string
  updatedAt: string
}

export interface AdminWebhookInput {
  url: string
  active: boolean
}
