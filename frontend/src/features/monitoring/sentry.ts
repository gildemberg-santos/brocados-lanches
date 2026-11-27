import * as Sentry from '@sentry/vue'
import type { App } from 'vue'

import { sanitizePath, sanitizeSentryEvent, sanitizeText } from './sentryScrub'

/** Inicializa apenas com DSN configurado; dev/test ficam sem envio de eventos. */
export function initSentry(app: App): boolean {
  const dsn = import.meta.env.VITE_SENTRY_DSN
  if (!dsn) {
    return false
  }

  Sentry.init({
    app,
    dsn,
    environment: import.meta.env.VITE_SENTRY_ENVIRONMENT ?? 'development',
    release: import.meta.env.VITE_SENTRY_RELEASE,
    tracesSampleRate: 0,
    attachProps: false,
    dataCollection: {
      userInfo: false,
      cookies: false,
      httpHeaders: false,
      httpBodies: [],
      urlQueryParams: false,
    },
    beforeSend: (event) => sanitizeSentryEvent(event),
  })

  return true
}

interface HttpFailure {
  method: string
  path: string
  error: unknown
  durationMs: number
}

function apiStatus(error: unknown): number {
  if (typeof error === 'object' && error !== null && 'status' in error) {
    const status: unknown = error.status
    if (typeof status === 'number') {
      return status
    }
  }
  return 0
}

/**
 * Ponto único de captura do cliente HTTP. 4xx são regras de negócio: viram
 * breadcrumb. Somente 5xx e falhas de rede/timeout viram evento no Sentry.
 */
export function reportHttpError({ method, path, error, durationMs }: HttpFailure): void {
  const status = apiStatus(error)
  const safePath = sanitizePath(path)
  const data = {
    method,
    path: safePath,
    status,
    duration_ms: Math.round(durationMs),
  }

  Sentry.addBreadcrumb({
    category: 'http',
    message: `${method} ${safePath}`,
    level: 'warning',
    data,
  })

  if (status >= 500 || status === 0) {
    Sentry.withScope((scope) => {
      scope.setLevel('error')
      scope.setTag('http.status', String(status))
      scope.setContext('http', data)
      Sentry.captureException(error)
    })
  }
}

/** Erros de carregamento de rota usam apenas o nome da rota, nunca a URL. */
export function reportRouterError(error: unknown, routeName: unknown): void {
  Sentry.withScope((scope) => {
    scope.setTag(
      'route',
      typeof routeName === 'string' && routeName ? sanitizeText(routeName) : 'unknown',
    )
    Sentry.captureException(error)
  })
}
