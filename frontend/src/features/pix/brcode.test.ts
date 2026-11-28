import { describe, expect, it } from 'vitest'

import {
  buildPixBrCode,
  calculateCrc16,
  formatPixAmount,
  formatTlv,
  sanitizePixText,
} from './brcode'

const pixKey = '123e4567-e89b-12d3-a456-426614174000'

describe('BR Code do Pix estático', () => {
  it('formata TLV com tamanho de dois dígitos', () => {
    expect(formatTlv('58', 'BR')).toBe('5802BR')
  })

  it('calcula CRC16-CCITT-FALSE com quatro hexadecimais maiúsculos', () => {
    expect(calculateCrc16('5802BR')).toMatch(/^6304[0-9A-F]{4}$/)
    expect(calculateCrc16('5802BR')).toBe(calculateCrc16('5802BR'))
  })

  it('higieniza acentos e limita os campos de texto', () => {
    expect(sanitizePixText('São Paulo!', 15)).toBe('Sao Paulo')
    expect(sanitizePixText('Brocados & Lanches', 25)).toBe('Brocados  Lanches')
    expect(sanitizePixText('Aquiraz', 3)).toBe('Aqu')
  })

  it('converte centavos sem ponto flutuante e respeita o limite da tag 54', () => {
    expect(formatPixAmount(4990)).toBe('49.90')
    expect(formatPixAmount(5)).toBe('0.05')
    expect(formatPixAmount(10000)).toBe('100.00')
    expect(() => formatPixAmount(0)).toThrow(/centavos positivo/)
    expect(() => formatPixAmount(1234567890123456)).toThrow(/limite da tag 54/)
  })

  it('gera payload com valor fixo e txid padrão', () => {
    const brCode = buildPixBrCode({
      key: pixKey,
      merchantName: 'BROCADOS LANCHES',
      merchantCity: 'Aquiraz',
      amountCents: 4990,
    })

    expect(brCode).toBe(
      '00020126580014br.gov.bcb.pix0136123e4567-e89b-12d3-a456-426614174000520400005303986540549.905802BR5916BROCADOS LANCHES6007Aquiraz62070503***630476B5',
    )
  })

  it('aceita chave de 77 caracteres e rejeita 78 para manter o TLV em dois dígitos', () => {
    const limite = buildPixBrCode({
      key: `${'a'.repeat(65)}@exemplo.com`,
      merchantName: 'Brocados',
      merchantCity: 'Aquiraz',
    })

    expect(limite).toContain('2699')
    expect(limite.slice(-8)).toBe(calculateCrc16(limite.slice(0, -8)))
    expect(() =>
      buildPixBrCode({
        key: `${'a'.repeat(66)}@exemplo.com`,
        merchantName: 'Brocados',
        merchantCity: 'Aquiraz',
      }),
    ).toThrow(/chave Pix/)
  })

  it('rejeita chaves que não são CPF, CNPJ, e-mail, telefone E.164 ou EVP', () => {
    for (const key of ['abc', '@@@', '+abc']) {
      expect(() =>
        buildPixBrCode({ key, merchantName: 'Brocados', merchantCity: 'Aquiraz' }),
      ).toThrow(/chave Pix/)
    }
  })

  it('omite a tag 54 quando o valor é livre', () => {
    const brCode = buildPixBrCode({
      key: pixKey,
      merchantName: 'Brocados',
      merchantCity: 'Aquiraz',
    })

    expect(brCode).not.toContain('5405')
    expect(brCode).toContain('5802BR')
    expect(brCode.slice(-8)).toBe(calculateCrc16(brCode.slice(0, -8)))
  })

  it('usa txid higienizado quando informado', () => {
    const brCode = buildPixBrCode({
      key: 'pagamento@brocados.example',
      merchantName: 'Brocados',
      merchantCity: 'Aquiraz',
      amountCents: 1200,
      txid: 'PED-1234',
    })

    expect(brCode).toContain('62110507PED1234')
    expect(brCode.slice(-8)).toBe(calculateCrc16(brCode.slice(0, -8)))
  })

  it('rejeita configuração incompleta ou inválida', () => {
    expect(() =>
      buildPixBrCode({ key: '', merchantName: 'Brocados', merchantCity: 'Aquiraz' }),
    ).toThrow(/chave Pix/)

    expect(() =>
      buildPixBrCode({
        key: 'chave com espaco',
        merchantName: 'Brocados',
        merchantCity: 'Aquiraz',
      }),
    ).toThrow(/chave Pix/)

    expect(() =>
      buildPixBrCode({ key: pixKey, merchantName: '!!!', merchantCity: 'Aquiraz' }),
    ).toThrow(/recebedor/)
  })
})
