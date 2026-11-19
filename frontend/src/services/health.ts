import { get } from './http'

export interface HealthResponse {
  status: string
  checks?: {
    mongo?: string
  }
}

export function fetchHealth(): Promise<HealthResponse> {
  return get<HealthResponse>('/readyz')
}
