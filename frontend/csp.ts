import { AsyncLocalStorage } from 'node:async_hooks'
import { randomBytes } from 'node:crypto'

import type { Plugin } from 'vite'

export const CSP_NONCE_PLACEHOLDER = '__CSP_NONCE__'

const GTM_ORIGIN = 'https://www.googletagmanager.com'
const ANALYTICS_ORIGIN = 'https://www.google-analytics.com'

/** Política usada no desenvolvimento; o Nginx replica as diretivas em produção. */
export function buildCsp(nonce: string): string {
  return [
    "default-src 'self'",
    `script-src 'self' 'nonce-${nonce}' ${GTM_ORIGIN}`,
    "style-src 'self' 'unsafe-inline'",
    `img-src 'self' data: ${ANALYTICS_ORIGIN}`,
    "font-src 'self'",
    `connect-src 'self' https://*.sentry.io ${GTM_ORIGIN} ${ANALYTICS_ORIGIN}`,
    "base-uri 'self'",
    "form-action 'self'",
    "frame-ancestors 'none'",
  ].join('; ')
}

const nonceStorage = new AsyncLocalStorage<string>()

/**
 * Gera um nonce por resposta HTML no Vite e injeta o mesmo valor no snippet
 * do GTM antes que os outros middlewares sirvam o index.html.
 */
export function cspNoncePlugin(): Plugin {
  return {
    name: 'brocados-csp-nonce',
    configureServer(server) {
      server.middlewares.use((req, res, next) => {
        const accept = req.headers.accept
        const wantsHtml = typeof accept === 'string' && accept.includes('text/html')
        if ((req.method !== 'GET' && req.method !== 'HEAD') || !wantsHtml) {
          next()
          return
        }

        const nonce = randomBytes(16).toString('base64')
        res.setHeader('Content-Security-Policy', buildCsp(nonce))
        nonceStorage.run(nonce, () => next())
      })
    },
    transformIndexHtml: {
      order: 'pre',
      handler(html) {
        const nonce = nonceStorage.getStore()
        return nonce ? html.replaceAll(CSP_NONCE_PLACEHOLDER, nonce) : undefined
      },
    },
  }
}
