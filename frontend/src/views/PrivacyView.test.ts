import { mount } from '@vue/test-utils'
import { defineComponent, h } from 'vue'
import { createMemoryHistory, createRouter } from 'vue-router'
import { describe, expect, it } from 'vitest'

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
  it('states that the definitive text still needs approval', async () => {
    const wrapper = await mountPrivacy()

    expect(wrapper.get('h1').text()).toBe('Privacidade')
    expect(wrapper.get('[role="note"]').text()).toContain('Texto sujeito a aprovação')
    expect(wrapper.text()).toContain('aprovação do cliente e do jurídico')
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
})
