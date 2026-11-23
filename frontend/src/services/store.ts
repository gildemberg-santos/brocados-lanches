import { getData } from './http'
import type { PublicStoreStatus } from '../types/store'

const BASE_PATH = '/api/v1/store'

export function fetchStoreStatus(): Promise<PublicStoreStatus> {
  return getData<PublicStoreStatus>(BASE_PATH)
}
