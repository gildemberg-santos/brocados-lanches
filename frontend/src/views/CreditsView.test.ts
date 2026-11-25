import { mount } from '@vue/test-utils'
import { defineComponent, h } from 'vue'
import { createMemoryHistory, createRouter } from 'vue-router'
import { describe, expect, it } from 'vitest'

import CreditsView from './CreditsView.vue'

const StubView = defineComponent({ name: 'StubView', render: () => h('div') })

async function mountCredits() {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', component: StubView },
      { path: '/contato', component: StubView },
    ],
  })
  await router.push('/')
  await router.isReady()

  return mount(CreditsView, { global: { plugins: [router] } })
}

describe('CreditsView', () => {
  it('flags the pending confirmation instead of claiming authorship', async () => {
    const wrapper = await mountCredits()

    expect(wrapper.get('h1').text()).toBe('Créditos')
    expect(wrapper.get('[role="note"]').text()).toContain('Página institucional em avaliação')
    expect(wrapper.text()).toContain('Nenhum crédito é afirmado antes dessa confirmação')
  })

  it('links to the published contact page without inventing details', async () => {
    const wrapper = await mountCredits()
    const text = wrapper.text()

    expect(wrapper.find('a[href="/contato"]').exists()).toBe(true)
    expect(text).toContain('noindex')
    expect(text).not.toContain('CNPJ')
    expect(text).not.toContain('+55')
    expect(text).not.toContain('@')
  })
})
