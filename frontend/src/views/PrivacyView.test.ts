import { mount } from '@vue/test-utils'
import { defineComponent, h } from 'vue'
import { createMemoryHistory, createRouter } from 'vue-router'
import { beforeEach, describe, expect, it } from 'vitest'

import { CHECKOUT_PREFERENCES_KEY } from '../features/checkout/preferences'
import { ORDER_ACCESS_TOKEN_KEY } from '../features/orders/orderAccessToken'
import PrivacyView from './PrivacyView.vue'

const StubView = defineComponent({ name: 'StubView', render: () => h('div') })

async function mountPrivacy() {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', component: StubView },
      { path: '/contato', component: StubView },
    ],
  })
  await router.push('/')
  await router.isReady()

  return mount(PrivacyView, { global: { plugins: [router] } })
}

describe('PrivacyView', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('presents the approved text as the current version', async () => {
    const wrapper = await mountPrivacy()
    const text = wrapper.text()

    expect(wrapper.get('h1').text()).toBe('Privacidade')
    expect(wrapper.get('[role="note"]').text()).toContain('Versão vigente')
    expect(wrapper.get('.info-page__date').text()).toContain('Publicada em 25/11/2026.')
    expect(wrapper.get('.info-page__date').text()).toContain('Atualizada em 27/11/2026.')
    expect(text).not.toContain('Rascunho')
    expect(text).not.toContain('pendente de E0')
    expect(text).not.toContain('sujeito a aprovação')
  })

  it('describes only documented data and does not invent legal details', async () => {
    const wrapper = await mountPrivacy()
    const text = wrapper.text()

    expect(text).toContain('nome, telefone, endereço de entrega e observação do pedido')
    expect(text).toContain('canais já publicados pela loja')
    expect(text).toContain('BR Code na confirmação, com o total do pedido já gravado')
    expect(text).not.toContain('BR Code com o total cotado')
    expect(wrapper.find('a[href="/contato"]').exists()).toBe(true)
    expect(text).not.toContain('CNPJ')
    expect(text).not.toContain('+55')
    expect(text).not.toContain('@')
  })

  it('explains the real data flow, retention and rights', async () => {
    const wrapper = await mountPrivacy()
    const text = wrapper.text()

    expect(wrapper.findAll('section h2')).toHaveLength(8)
    expect(text).toContain('Aquiraz–CE')
    expect(text).toContain('hash Argon2id')
    expect(text).toContain('link wa.me')
    expect(text).toContain('Google Tag Manager')
    expect(text).toContain('em todas as visitas')
    expect(text).not.toContain('Não usamos analytics')
    expect(text).toContain('atualizações automáticas')
    expect(text).toContain('nenhum dado pessoal trafega por ele')
    expect(text).toContain('24 meses')
    expect(text).toContain('7 dias')
    expect(text).toContain('12 meses')
    expect(text).toContain('30 dias')
    expect(text).toContain('anonimização ou eliminação')
  })

  it('descreve as preferências salvas e o que fica de fora', async () => {
    const wrapper = await mountPrivacy()
    const text = wrapper.text()

    expect(text).toContain('preferências do checkout salvas neste navegador')
    expect(text).toContain('Observação e valor de troco não são salvos')
    expect(text).toContain('Apagar preferências salvas neste navegador')
    expect(text).toContain('até você apagar nesta página')
    expect(text).toContain('perfil de conveniência no servidor, identificado pelo telefone')
    expect(text).toContain('a resposta não traz taxa, histórico de pedidos, observação ou troco')
    expect(text).toContain('buscar e reaproveitar no próximo pedido')
    expect(text).toContain('Esse tratamento pode ser desativado no checkout')
    expect(text).toContain('perfil de conveniência no servidor: 24 meses')
    expect(text).toContain('token opaco de acesso aos pedidos')
    expect(text).toContain('Authorization')
    expect(text).toContain('histórico de pedidos: permitir que você consulte novamente')
  })

  it('descreve o monitoramento de erros sem prometer dados pessoais', async () => {
    const wrapper = await mountPrivacy()
    const text = wrapper.text()

    expect(text).toContain('relatórios de falhas inesperadas da API e do navegador')
    expect(text).toContain('Sentry: serviço de monitoramento de erros')
    expect(text).toContain('eventos de erro no Sentry: até 30 dias')
    expect(text).toContain('não coleta identificação de usuário')
    expect(text).not.toContain('Session Replay')
    expect(text).not.toContain('rastreamento de performance ativo')
  })

  it('apaga as preferências salvas neste navegador', async () => {
    localStorage.setItem(CHECKOUT_PREFERENCES_KEY, JSON.stringify({ version: 1 }))
    const wrapper = await mountPrivacy()
    const button = wrapper.get('button')

    expect(button.attributes('disabled')).toBeUndefined()

    await button.trigger('click')

    expect(localStorage.getItem(CHECKOUT_PREFERENCES_KEY)).toBeNull()
    expect(wrapper.get('[role="status"]').text()).toBe('Preferências apagadas deste navegador.')
    expect(wrapper.get('button').attributes('disabled')).toBeDefined()
  })

  it('apaga o token local e avisa que o histórico deixará de existir', async () => {
    localStorage.setItem(
      ORDER_ACCESS_TOKEN_KEY,
      JSON.stringify({ version: 1, token: 'a'.repeat(43) }),
    )
    const wrapper = await mountPrivacy()
    const button = wrapper.findAll('button')[1]

    expect(button.text()).toBe('Apagar acesso aos meus pedidos')
    expect(button.attributes('disabled')).toBeUndefined()

    await button.trigger('click')

    expect(localStorage.getItem(ORDER_ACCESS_TOKEN_KEY)).toBeNull()
    expect(wrapper.get('[role="status"]').text()).toBe(
      'Acesso aos pedidos apagado deste navegador. Faça um novo pedido para criar outro acesso.',
    )
  })

  it('does not copy providers or features from the reference policy', async () => {
    const wrapper = await mountPrivacy()
    const text = wrapper.text()

    for (const name of ['Stripe', 'Firebase', 'Clarity', 'Anthropic', 'LocationIQ', 'Microsoft']) {
      expect(text).not.toContain(name)
    }
    expect(text).not.toContain('geocodificação de endereços')
    expect(text).not.toContain('inteligência artificial para importar')
  })
})
