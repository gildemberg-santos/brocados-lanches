import { createApp } from 'vue'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { initSentry, reportHttpError, reportRouterError } from './sentry'

const { initMock, breadcrumbMock, captureMock, scope } = vi.hoisted(() => ({
  initMock: vi.fn(),
  breadcrumbMock: vi.fn(),
  captureMock: vi.fn(),
  scope: {
    setLevel: vi.fn(),
    setTag: vi.fn(),
    setContext: vi.fn(),
  },
}))

vi.mock('@sentry/vue', () => ({
  init: initMock,
  addBreadcrumb: breadcrumbMock,
  captureException: captureMock,
  withScope: (callback: (value: unknown) => void) => callback(scope),
}))

describe('monitoring', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    vi.unstubAllEnvs()
  })

  it('não inicializa sem DSN', () => {
    expect(initSentry(createApp({ template: '<div />' }))).toBe(false)
    expect(initMock).not.toHaveBeenCalled()
  })

  it('configura somente erros e sem coleta de PII quando há DSN', () => {
    vi.stubEnv('VITE_SENTRY_DSN', 'https://key@sentry.example.com/1')
    vi.stubEnv('VITE_SENTRY_ENVIRONMENT', 'production')
    vi.stubEnv('VITE_SENTRY_RELEASE', 'abc123')

    expect(initSentry(createApp({ template: '<div />' }))).toBe(true)

    const options = initMock.mock.calls[0]?.[0] as Record<string, unknown>
    expect(options).toMatchObject({
      dsn: 'https://key@sentry.example.com/1',
      environment: 'production',
      release: 'abc123',
      tracesSampleRate: 0,
      attachProps: false,
      dataCollection: {
        userInfo: false,
        cookies: false,
        httpHeaders: false,
        httpBodies: [],
        urlQueryParams: false,
      },
    })

    const beforeSend = options.beforeSend as (
      event: Record<string, unknown>,
    ) => Record<string, unknown>
    const sanitized = beforeSend({
      user: { email: 'dono@teste.local' },
      message: '+5585999999999',
      exception: { values: [{ value: 'dono@teste.local' }] },
    })
    expect(sanitized.user).toBeUndefined()
    expect(JSON.stringify(sanitized)).not.toContain('dono@teste.local')
  })

  it('não captura 4xx e não expõe caminho com identificador ou query', () => {
    reportHttpError({
      method: 'GET',
      path: '/api/v1/orders/507f1f77bcf86cd799439011?phone=5585999999999',
      error: { status: 422 },
      durationMs: 12.4,
    })

    expect(captureMock).not.toHaveBeenCalled()
    expect(breadcrumbMock).toHaveBeenCalledWith({
      category: 'http',
      message: 'GET /api/v1/orders/:id',
      level: 'warning',
      data: { method: 'GET', path: '/api/v1/orders/:id', status: 422, duration_ms: 12 },
    })
  })

  it('captura 5xx e timeout com contexto sanitizado', () => {
    const error = new Error('falha interna')

    reportHttpError({
      method: 'POST',
      path: '/api/v1/orders',
      error: { status: 500 },
      durationMs: 30,
    })
    reportHttpError({ method: 'POST', path: '/orders?phone=1', error, durationMs: 40 })

    expect(captureMock).toHaveBeenCalledTimes(2)
    expect(scope.setTag).toHaveBeenCalledWith('http.status', '500')
    expect(scope.setTag).toHaveBeenCalledWith('http.status', '0')
    expect(scope.setContext).toHaveBeenCalledWith('http', {
      method: 'POST',
      path: '/api/v1/orders',
      status: 500,
      duration_ms: 30,
    })
    expect(JSON.stringify(scope.setContext.mock.calls)).not.toContain('phone=1')
  })

  it('captura erro de rota apenas com o nome sanitizado', () => {
    reportRouterError(new Error('chunk falhou'), 'admin-order-detail')

    expect(captureMock).toHaveBeenCalledTimes(1)
    expect(scope.setTag).toHaveBeenCalledWith('route', 'admin-order-detail')
  })
})
