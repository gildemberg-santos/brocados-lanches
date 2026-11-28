import { mount } from '@vue/test-utils'
import { defineComponent, h } from 'vue'
import { createMemoryHistory, createRouter } from 'vue-router'
import { describe, expect, it } from 'vitest'

import TermsView from './TermsView.vue'

const StubView = defineComponent({ name: 'StubView', render: () => h('div') })

async function mountTerms() {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', component: StubView },
      { path: '/contato', component: StubView },
      { path: '/privacidade', component: StubView },
    ],
  })
  await router.push('/')
  await router.isReady()

  return mount(TermsView, { global: { plugins: [router] } })
}

describe('TermsView', () => {
  it('presents the approved terms as the current version', async () => {
    const wrapper = await mountTerms()
    const text = wrapper.text()

    expect(wrapper.get('h1').text()).toBe('Termos de uso')
    expect(wrapper.get('[role="note"]').text()).toContain('Versão vigente')
    expect(wrapper.get('.info-page__date').text()).toBe('Publicados em 25/11/2026.')
    expect(text).not.toContain('Rascunho')
    expect(text).not.toContain('pendente de E0')
    expect(text).not.toContain('sujeito a aprovação')
  })

  it('keeps the rules factual and never invents business data', async () => {
    const wrapper = await mountTerms()
    const text = wrapper.text()

    expect(text).toContain('gravado no servidor')
    expect(text).toContain('nenhum dado de cartão é solicitado')
    expect(wrapper.find('a[href="/contato"]').exists()).toBe(true)
    expect(text).not.toContain('CNPJ')
    expect(text).not.toContain('+55')
    expect(text).not.toContain('@')
  })

  it('describes the real order, payment and store rules', async () => {
    const wrapper = await mountTerms()
    const text = wrapper.text()

    expect(wrapper.findAll('section h2')).toHaveLength(10)
    expect(text).toContain('Aquiraz–CE')
    expect(text).toContain('código público opaco')
    expect(text).toContain('chave de idempotência')
    expect(text).toContain('interruptor manual desligado')
    expect(text).toContain('congelados no pedido')
    expect(text).toContain('taxa de entrega é zero')
    expect(text).toContain('pedido é gravado no servidor primeiro')
    expect(text).toContain('confirmação mostra o QR Code')
    expect(text).not.toContain('checkout mostra o QR Code')
    expect(text).toContain('estados finais')
    expect(text).toContain('Política de Privacidade')
    expect(wrapper.find('a[href="/privacidade"]').exists()).toBe(true)
    expect(text).toContain('token opaco de acesso')
    expect(text).toContain('Meus pedidos')
    expect(text).toContain('sem cadastro nem consulta por telefone')
  })

  it('does not copy subscription rules from the reference terms', async () => {
    const wrapper = await mountTerms()
    const text = wrapper.text()

    for (const phrase of [
      'MenuAqui',
      'Stripe',
      'R$ 89,90',
      'teste gratuito',
      'Plano Pro',
      'cartão de crédito',
      'comissão',
    ]) {
      expect(text).not.toContain(phrase)
    }
  })
})
