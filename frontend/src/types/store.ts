import type { AdminOpeningHour } from './adminCatalog'

/** Status efetivo da loja, calculado pelo servidor (GET /store). */
export interface PublicStoreStatus {
  name: string
  status: 'open' | 'closed'
  isOpenManual: boolean
  timezone: string
  deliveryTime: string
  whatsapp: string
  instagram: string
  address: string
  openingHours: AdminOpeningHour[]
  nextOpeningAt: string | null
  updatedAt: string
}
