const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? ''

export interface ApiErrorDetail {
  field?: string
  code?: string
  message?: string
}

interface ApiErrorPayload {
  error?: {
    code?: string
    message?: string
    details?: ApiErrorDetail[]
  }
}

interface RequestOptions {
  headers?: Record<string, string>
  idempotencyKey?: string
}

export class ApiError extends Error {
  readonly status: number
  readonly code: string
  readonly details: ApiErrorDetail[]

  constructor(status: number, code: string, message: string, details: ApiErrorDetail[] = []) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
    this.details = details
  }
}

export async function get<T>(path: string): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    headers: {
      Accept: 'application/json',
    },
    credentials: 'include',
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
      payload.error?.details ?? [],
    )
  }

  return (await response.json()) as T
}

export async function getData<T>(path: string): Promise<T> {
  const body = await get<{ data: T }>(path)
  return body.data
}

async function parseError(response: Response): Promise<ApiError> {
  let payload: ApiErrorPayload = {}

  try {
    payload = (await response.json()) as ApiErrorPayload
  } catch {
    payload = {}
  }

  return new ApiError(
    response.status,
    payload.error?.code ?? 'HTTP_ERROR',
    payload.error?.message ?? 'Não foi possível concluir a requisição.',
    payload.error?.details ?? [],
  )
}

export async function post<T>(
  path: string,
  value: unknown,
  options: RequestOptions = {},
): Promise<T> {
  const headers: Record<string, string> = {
    Accept: 'application/json',
    'Content-Type': 'application/json',
    ...options.headers,
  }
  if (options.idempotencyKey) {
    headers['Idempotency-Key'] = options.idempotencyKey
  }

  const response = await fetch(`${API_BASE_URL}${path}`, {
    method: 'POST',
    headers,
    credentials: 'include',
    body: JSON.stringify(value),
  })

  if (!response.ok) {
    throw await parseError(response)
  }

  return (await response.json()) as T
}

export async function postData<T>(
  path: string,
  value: unknown,
  options: RequestOptions = {},
): Promise<T> {
  const body = await post<{ data: T }>(path, value, options)
  return body.data
}

/** POST sem corpo, para endpoints que respondem 204 (ex.: logout). */
export async function postEmpty(path: string): Promise<void> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    method: 'POST',
    headers: {
      Accept: 'application/json',
    },
    credentials: 'include',
  })

  if (!response.ok) {
    throw await parseError(response)
  }
}

export async function patch<T>(path: string, value: unknown): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    method: 'PATCH',
    headers: {
      Accept: 'application/json',
      'Content-Type': 'application/json',
    },
    credentials: 'include',
    body: JSON.stringify(value),
  })

  if (!response.ok) {
    throw await parseError(response)
  }

  return (await response.json()) as T
}

export async function patchData<T>(path: string, value: unknown): Promise<T> {
  const body = await patch<{ data: T }>(path, value)
  return body.data
}
