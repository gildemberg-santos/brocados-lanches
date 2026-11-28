import { describe, expect, it } from 'vitest'

import { validatePixInput } from './pixForm'

const valid = {
  active: true,
  key: '123e4567-e89b-12d3-a456-426614174000',
  merchantName: 'Brocados Lanches',
  merchantCity: 'Aquiraz',
}

describe('validação da configuração do Pix', () => {
  it('aceita configuração completa', () => {
    expect(validatePixInput(valid)).toEqual({})
  })

  it('aceita rascunho desativado', () => {
    expect(
      validatePixInput({ ...valid, active: false, key: '', merchantName: '', merchantCity: '' }),
    ).toEqual({})
  })

  it('exige chave e dados do recebedor quando ativo', () => {
    const errors = validatePixInput({ ...valid, key: ' ', merchantName: '!', merchantCity: '' })

    expect(errors.key).toContain('chave Pix')
    expect(errors.merchantName).toContain('nome do recebedor')
    expect(errors.merchantCity).toContain('cidade')
  })

  it('rejeita chave com espaço e texto acima do limite sanitizado', () => {
    const errors = validatePixInput({
      ...valid,
      key: 'chave com espaço',
      merchantName: '12345678901234567890123456',
      merchantCity: 'São Paulo dos Campos',
    })

    expect(errors.key).toContain('sem espaços')
    expect(errors.merchantName).toContain('25')
    expect(errors.merchantCity).toContain('15')
  })

  it('aceita os formatos de chave do arranjo Pix', () => {
    for (const key of [
      '123.456.789-09',
      '12.345.678/0001-90',
      'contato@brocados.com.br',
      '+5585999999999',
    ]) {
      expect(validatePixInput({ ...valid, key })).toEqual({})
    }
  })

  it('rejeita chaves sem tipo reconhecido', () => {
    for (const key of ['abc', '@@@', '+abc']) {
      expect(validatePixInput({ ...valid, key }).key).toContain('chave Pix')
    }
  })

  it('aceita chave no limite do BR Code e rejeita uma caractere acima', () => {
    expect(validatePixInput({ ...valid, key: `${'a'.repeat(65)}@exemplo.com` })).toEqual({})
    expect(validatePixInput({ ...valid, key: `${'a'.repeat(66)}@exemplo.com` }).key).toContain(
      'chave Pix',
    )
  })

  it('rejeita texto que não sobrevive à higienização do BR Code', () => {
    expect(validatePixInput({ ...valid, merchantName: '日本語' }).merchantName).toContain(
      'nome do recebedor',
    )
  })

  it('rejeita caracteres de controle que o servidor também rejeita', () => {
    const errors = validatePixInput({
      ...valid,
      merchantName: 'Brocados\u0001',
      merchantCity: 'Aquiraz\u007f',
    })

    expect(errors.merchantName).toContain('caracteres de controle')
    expect(errors.merchantCity).toContain('caracteres de controle')
  })

  it('ignora símbolos que serão removidos pelo gerador', () => {
    expect(validatePixInput({ ...valid, merchantName: 'Brocados & Lanches' })).toEqual({})
  })
})
