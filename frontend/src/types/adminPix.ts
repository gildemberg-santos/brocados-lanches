/** Configuração do BR Code estático exibido na confirmação de pedidos pagos com Pix. */
export interface AdminPixSettings {
  active: boolean
  key: string
  merchantName: string
  merchantCity: string
  createdAt: string
  updatedAt: string
}

export interface AdminPixInput {
  active: boolean
  key: string
  merchantName: string
  merchantCity: string
}

/** Dados públicos necessários para montar o BR Code no navegador. */
export interface PublicPixSettings {
  active: boolean
  key: string
  merchantName: string
  merchantCity: string
}
