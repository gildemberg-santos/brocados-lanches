const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? ''

interface ApiErrorPayload {
  error?: {
    code?: string
    message?: string
  }
}

export class ApiError extends Error {
  readonly status: number
  readonly code: string

  constructor(status: number, code: string, message: string) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
  }
}

export async function get<T>(path: string): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    headers: {
      Accept: 'application/json',
    },
  })

  if (!response.ok) {
    let payload: ApiErrorPayload = {}

    try {
      payload = (await response.json()) as ApiErrorPayload
    } catch {
      payload = {}
    }

    throw new ApiError(
      response.status,
      payload.error?.code ?? 'HTTP_ERROR',
      payload.error?.message ?? 'Não foi possível concluir a requisição.',
    )
  }

  return (await response.json()) as T
}

export async function getData<T>(path: string): Promise<T> {
  const body = await get<{ data: T }>(path)
  return body.data
}
