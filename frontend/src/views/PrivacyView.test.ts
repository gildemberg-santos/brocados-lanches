import { mount } from '@vue/test-utils'
import { defineComponent, h } from 'vue'
import { createMemoryHistory, createRouter } from 'vue-router'
import { beforeEach, describe, expect, it } from 'vitest'

import { CHECKOUT_PREFERENCES_KEY } from '../features/checkout/preferences'
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
    expect(wrapper.get('.info-page__date').text()).toContain('Atualizada em 26/11/2026.')
    expect(text).not.toContain('Rascunho')
    expect(text).not.toContain('pendente de E0')
    expect(text).not.toContain('sujeito a aprovação')
  })

  it('describes only documented data and does not invent legal details', async () => {
    const wrapper = await mountPrivacy()
    const text = wrapper.text()

    expect(text).toContain('nome, telefone, endereço de entrega e observação do pedido')
    expect(text).toContain('canais já publicados pela loja')
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
    expect(text).toContain('Não usamos analytics')
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
