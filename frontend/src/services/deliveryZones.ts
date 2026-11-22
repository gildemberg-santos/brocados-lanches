import type { DeliveryZone } from '../types/delivery'

import { getData } from './http'

export function fetchDeliveryZones(): Promise<DeliveryZone[]> {
  return getData<DeliveryZone[]>('/api/v1/delivery-zones')
}
