import { describe, expect, it } from 'vitest'

import { validateWebhookInput } from './webhookForm'

describe('validação do webhook', () => {
  it('exige URL quando ativo', () => {
    expect(validateWebhookInput('  ', true).url).toContain('Informe a URL')
  })

  it('aceita URL vazia quando inativo', () => {
    expect(validateWebhookInput('', false).url).toBeUndefined()
  })

  it('aceita qualquer esquema em qualquer host', () => {
    expect(validateWebhookInput('http://127.0.0.1:5678/hook', true).url).toBeUndefined()
    expect(validateWebhookInput('https://hooks.exemplo.invalid/path', true).url).toBeUndefined()
    expect(validateWebhookInput('ftp://integracao.exemplo.invalid/hook', true).url).toBeUndefined()
  })

  it('aceita esquema sem host', () => {
    expect(validateWebhookInput('custom:hook', true).url).toBeUndefined()
  })

  it('rejeita URL sem esquema', () => {
    expect(validateWebhookInput('somente/caminho', true).url).toContain('URL absoluta')
  })
})
