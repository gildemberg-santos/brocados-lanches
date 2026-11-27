const emailPattern = /[\w.+-]+@[\w-]+\.[\w.-]+/g
const phonePattern = /(?:\+?55)?\(?\d{2}\)?[\s.-]?\d{4,5}-?\d{4}/g
const zipPattern = /\d{5}-?\d{3}/g
const waLinkPattern = /https?:\/\/(?:wa\.me|api\.whatsapp\.com)[^\s"'<>]*|wa\.me\/[^\s"'<>]*/gi
const secretPattern =
  /(token|authorization|cookie|idempotency-key|idempotencykey)([:=]\s*)[^\s,;]+/gi
const queryPattern = /\?[^\s"'<>]*/g
const sensitiveKeyPattern =
  /(phone|telefone|email|e-mail|address|endereco|cep|zip|token|authorization|cookie|idempotency|password|senha|whatsapp|walink|note|observacao|body|headers?)/i
const opaqueIdPattern = /^[a-f0-9]{24}$/i
const longIdPattern = /^[A-Za-z0-9_-]{16,}$/

/** Remove dados pessoais e segredos de qualquer texto que chega ao Sentry. */
export function sanitizeText(value: string): string {
  return value
    .replace(emailPattern, '[email]')
    .replace(phonePattern, '[telefone]')
    .replace(zipPattern, '[cep]')
    .replace(waLinkPattern, '[wa-link]')
    .replace(secretPattern, '$1$2[redacted]')
    .replace(queryPattern, '?[redacted]')
}

/** Caminho de baixa cardinalidade: sem query e sem identificadores opacos. */
export function sanitizePath(path: string): string {
  return (
    path
      .split(/[?#]/)[0]
      ?.split('/')
      .map((segment) =>
        opaqueIdPattern.test(segment) || longIdPattern.test(segment) ? ':id' : segment,
      )
      .join('/') ?? ''
  )
}

function sanitizeValue(value: unknown): unknown {
  if (typeof value === 'string') {
    return sanitizeText(value)
  }
  if (Array.isArray(value)) {
    return value.map((item) => sanitizeValue(item))
  }
  if (isRecord(value)) {
    return sanitizeRecord(value)
  }
  return value
}

export function sanitizeRecord(value: Record<string, unknown>): Record<string, unknown> {
  const result: Record<string, unknown> = {}
  for (const [key, item] of Object.entries(value)) {
    if (sensitiveKeyPattern.test(key)) {
      continue
    }
    result[key] = sanitizeValue(item)
  }
  return result
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}

interface SanitizableFrame {
  filename?: string
  abs_path?: string
  absPath?: string
  function?: string
  context_line?: string
  contextLine?: string
  pre_context?: string[]
  post_context?: string[]
  vars?: Record<string, unknown>
}

interface SanitizableEvent {
  user?: unknown
  request?: unknown
  breadcrumbs?: Array<{ message?: string; data?: Record<string, unknown> } | undefined>
  exception?: { values?: Array<{ value?: string; stacktrace?: { frames?: SanitizableFrame[] } }> }
  tags?: Record<string, unknown>
  extra?: Record<string, unknown>
  modules?: Record<string, string>
  fingerprint?: string[]
  contexts?: Record<string, Record<string, unknown> | undefined>
  message?: string
  transaction?: string
}

function sanitizeFrame(frame: SanitizableFrame): SanitizableFrame {
  return {
    ...frame,
    filename: frame.filename ? sanitizePath(frame.filename) : frame.filename,
    abs_path: frame.abs_path ? sanitizePath(frame.abs_path) : frame.abs_path,
    absPath: frame.absPath ? sanitizePath(frame.absPath) : frame.absPath,
    function: frame.function ? sanitizeText(frame.function) : frame.function,
    context_line: frame.context_line ? sanitizeText(frame.context_line) : frame.context_line,
    contextLine: frame.contextLine ? sanitizeText(frame.contextLine) : frame.contextLine,
    pre_context: frame.pre_context?.map(sanitizeText),
    post_context: frame.post_context?.map(sanitizeText),
    vars: frame.vars ? sanitizeRecord(frame.vars) : frame.vars,
  }
}

/** BeforeSend do frontend: nenhum evento leva user, request cru ou PII. */
export function sanitizeSentryEvent<T extends SanitizableEvent>(event: T): T {
  const sanitized: SanitizableEvent = event
  sanitized.user = undefined
  sanitized.request = undefined

  if (typeof sanitized.message === 'string') {
    sanitized.message = sanitizeText(sanitized.message)
  }
  if (typeof sanitized.transaction === 'string') {
    sanitized.transaction = sanitizePath(sanitized.transaction)
  }
  if (sanitized.tags) {
    sanitized.tags = sanitizeRecord(sanitized.tags)
  }
  if (sanitized.extra) {
    sanitized.extra = sanitizeRecord(sanitized.extra)
  }
  if (sanitized.contexts) {
    sanitized.contexts = Object.fromEntries(
      Object.entries(sanitized.contexts).map(([name, context]) => [
        name,
        context ? sanitizeRecord(context) : context,
      ]),
    )
  }
  if (sanitized.breadcrumbs) {
    sanitized.breadcrumbs = sanitized.breadcrumbs.map((breadcrumb) => {
      if (!breadcrumb) {
        return breadcrumb
      }
      return {
        ...breadcrumb,
        message: breadcrumb.message ? sanitizeText(breadcrumb.message) : breadcrumb.message,
        data: breadcrumb.data ? sanitizeRecord(breadcrumb.data) : breadcrumb.data,
      }
    })
  }
  if (sanitized.modules) {
    sanitized.modules = sanitizeRecord(sanitized.modules) as Record<string, string>
  }
  if (sanitized.fingerprint) {
    sanitized.fingerprint = sanitized.fingerprint.map(sanitizeText)
  }
  if (sanitized.exception?.values) {
    sanitized.exception.values = sanitized.exception.values.map((value) => ({
      ...value,
      value: value.value ? sanitizeText(value.value) : value.value,
      stacktrace: value.stacktrace?.frames
        ? { ...value.stacktrace, frames: value.stacktrace.frames.map(sanitizeFrame) }
        : value.stacktrace,
    }))
  }

  return event
}
