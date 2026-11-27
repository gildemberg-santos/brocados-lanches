import { describe, expect, it, vi } from 'vitest'

import { buildCsp, cspNoncePlugin, CSP_NONCE_PLACEHOLDER } from './csp'

describe('CSP do frontend', () => {
  it('usa nonce, domínios do Google e nenhum script inline irrestrito', () => {
    const policy = buildCsp('nonce-de-teste')

    expect(policy).toContain(
      "script-src 'self' 'nonce-nonce-de-teste' https://www.googletagmanager.com",
    )
    expect(policy).toContain(
      "connect-src 'self' https://*.sentry.io https://www.googletagmanager.com https://www.google-analytics.com",
    )
    expect(policy).toContain("img-src 'self' data: https://www.google-analytics.com")
    expect(policy).not.toContain("script-src 'self' 'unsafe-inline'")
  })

  it('gera um nonce por resposta HTML e substitui o placeholder no index', () => {
    const plugin = cspNoncePlugin()
    const middlewares: Array<(req: unknown, res: unknown, next: () => void) => void> = []
    const server = {
      middlewares: {
        use: (handler: (req: unknown, res: unknown, next: () => void) => void) => {
          middlewares.push(handler)
        },
      },
    }
    plugin.configureServer?.(server as never, {} as never)
    expect(middlewares).toHaveLength(1)

    const html = `<script nonce="${CSP_NONCE_PLACEHOLDER}">gtm()</script>`
    const transform = plugin.transformIndexHtml
    if (typeof transform === 'function') {
      throw new TypeError('transformIndexHtml deveria ter ordem pre')
    }
    const handler = transform?.handler ?? transform

    const headers: Record<string, string> = {}
    const res = { setHeader: (name: string, value: string) => (headers[name] = value) }
    const req = { method: 'GET', headers: { accept: 'text/html' } }
    let transformed = ''

    middlewares[0]?.(req, res, () => {
      transformed = handler(html, { path: '/', filename: 'index.html' }) as string
    })

    const firstNonce = headers['Content-Security-Policy']?.match(/nonce-([^']+)/)?.[1] ?? ''
    expect(firstNonce).not.toBe('')
    expect(transformed).toContain(`nonce="${firstNonce}"`)
    expect(transformed).not.toContain(CSP_NONCE_PLACEHOLDER)

    const secondHeaders: Record<string, string> = {}
    const secondRes = { setHeader: (name: string, value: string) => (secondHeaders[name] = value) }
    middlewares[0]?.(req, secondRes, () => {
      transformed = handler(html, { path: '/', filename: 'index.html' }) as string
    })
    const secondNonce = secondHeaders['Content-Security-Policy']?.match(/nonce-([^']+)/)?.[1] ?? ''

    expect(secondNonce).not.toBe(firstNonce)
    expect(transformed).toContain(`nonce="${secondNonce}"`)
  })

  it('não altera requisições que não pedem HTML', () => {
    const plugin = cspNoncePlugin()
    const handlers: Array<(req: unknown, res: unknown, next: () => void) => void> = []
    plugin.configureServer?.(
      {
        middlewares: {
          use: (handler: (req: unknown, res: unknown, next: () => void) => void) => {
            handlers.push(handler)
          },
        },
      } as never,
      {} as never,
    )

    const setHeader = vi.fn()
    const next = vi.fn()
    handlers[0]?.({ method: 'GET', headers: { accept: 'application/json' } }, { setHeader }, next)

    expect(setHeader).not.toHaveBeenCalled()
    expect(next).toHaveBeenCalledTimes(1)
  })
})
