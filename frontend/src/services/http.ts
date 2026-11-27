import { reportHttpError } from '../features/monitoring/sentry'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? ''
const DEFAULT_HTTP_TIMEOUT_MS = 10_000
const TIMEOUT_MESSAGE = 'A conexão demorou muito. Verifique sua conexão e tente novamente.'

function resolveHttpTimeoutMs(): number {
  const configured: unknown = import.meta.env.VITE_HTTP_TIMEOUT_MS
  const value =
    typeof configured === 'string' || typeof configured === 'number'
      ? Number(configured)
      : Number.NaN

  return Number.isFinite(value) && value > 0 ? value : DEFAULT_HTTP_TIMEOUT_MS
}

const HTTP_TIMEOUT_MS = resolveHttpTimeoutMs()

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

/**
 * Limita o tempo total da requisição (incluindo a leitura do corpo) e converte
 * abortes por timeout em um erro amigável, sem expor detalhes da falha.
 */
async function withTimeout<T>(request: (signal: AbortSignal) => Promise<T>): Promise<T> {
  const controller = new AbortController()
  const timer = setTimeout(() => controller.abort(), HTTP_TIMEOUT_MS)

  try {
    return await request(controller.signal)
  } catch (error: unknown) {
    if (controller.signal.aborted) {
      throw new ApiError(0, 'TIMEOUT', TIMEOUT_MESSAGE)
    }
    throw error
  } finally {
    clearTimeout(timer)
  }
}

/** Mede e reporta a falha sem expor headers, corpo ou URL completa. */
async function performRequest<T>(
  method: string,
  path: string,
  request: (signal: AbortSignal) => Promise<T>,
): Promise<T> {
  const started = performance.now()
  try {
    return await withTimeout(request)
  } catch (error: unknown) {
    reportHttpError({ method, path, error, durationMs: performance.now() - started })
    throw error
  }
}

export async function get<T>(path: string, options: RequestOptions = {}): Promise<T> {
  return performRequest('GET', path, async (signal) => {
    const response = await fetch(`${API_BASE_URL}${path}`, {
      headers: {
        Accept: 'application/json',
        ...options.headers,
      },
      credentials: 'include',
      signal,
    })

    if (!response.ok) {
      throw await parseError(response)
    }

    return (await response.json()) as T
  })
}

export async function getData<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const body = await get<{ data: T }>(path, options)
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

  return performRequest('POST', path, async (signal) => {
    const response = await fetch(`${API_BASE_URL}${path}`, {
      method: 'POST',
      headers,
      credentials: 'include',
      body: JSON.stringify(value),
      signal,
    })

    if (!response.ok) {
      throw await parseError(response)
    }

    return (await response.json()) as T
  })
}

export async function postData<T>(
  path: string,
  value: unknown,
  options: RequestOptions = {},
): Promise<T> {
  const body = await post<{ data: T }>(path, value, options)
  return body.data
}

/** POST multipart para upload; o navegador define o Content-Type com o boundary. */
export async function postForm<T>(path: string, formData: FormData): Promise<T> {
  return performRequest('POST', path, async (signal) => {
    const response = await fetch(`${API_BASE_URL}${path}`, {
      method: 'POST',
      headers: { Accept: 'application/json' },
      credentials: 'include',
      body: formData,
      signal,
    })

    if (!response.ok) {
      throw await parseError(response)
    }

    return (await response.json()) as T
  })
}

export async function postFormData<T>(path: string, formData: FormData): Promise<T> {
  const body = await postForm<{ data: T }>(path, formData)
  return body.data
}

/** POST sem corpo, para endpoints que respondem 204 (ex.: logout). */
export async function postEmpty(path: string): Promise<void> {
  return performRequest('POST', path, async (signal) => {
    const response = await fetch(`${API_BASE_URL}${path}`, {
      method: 'POST',
      headers: {
        Accept: 'application/json',
      },
      credentials: 'include',
      signal,
    })

    if (!response.ok) {
      throw await parseError(response)
    }
  })
}

export async function patch<T>(path: string, value: unknown): Promise<T> {
  return performRequest('PATCH', path, async (signal) => {
    const response = await fetch(`${API_BASE_URL}${path}`, {
      method: 'PATCH',
      headers: {
        Accept: 'application/json',
        'Content-Type': 'application/json',
      },
      credentials: 'include',
      body: JSON.stringify(value),
      signal,
    })

    if (!response.ok) {
      throw await parseError(response)
    }

    return (await response.json()) as T
  })
}

export async function patchData<T>(path: string, value: unknown): Promise<T> {
  const body = await patch<{ data: T }>(path, value)
  return body.data
}
