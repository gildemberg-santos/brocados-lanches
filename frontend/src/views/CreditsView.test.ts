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
      { path: '/cardapio', component: StubView },
      { path: '/contato', component: StubView },
    ],
  })
  await router.push('/')
  await router.isReady()

  return mount(CreditsView, { global: { plugins: [router] } })
}

describe('CreditsView', () => {
  it('credits the author using the supplied curriculum', async () => {
    const wrapper = await mountCredits()
    const text = wrapper.text()

    expect(wrapper.get('h1').text()).toBe('Créditos')
    expect(wrapper.find('[role="note"]').exists()).toBe(false)
    expect(text).not.toContain('em avaliação')
    expect(text).not.toContain('textos definitivos aguardam validação')
    expect(text).toContain('Gildemberg Santos')
    expect(text).toContain('Desenvolvedor de software com mais de uma década de experiência')
    expect(text).toContain('Leadster')
    expect(text).toContain('Aquiraz')
    expect(text).not.toContain('Nenhum crédito é afirmado antes dessa confirmação')
  })

  it('keeps only public professional links and avoids personal contact data', async () => {
    const wrapper = await mountCredits()
    const text = wrapper.text()

    const professionalLinks = wrapper.get('nav[aria-label="Links profissionais"]')
    const linkedin = professionalLinks.get(
      'a[href="https://www.linkedin.com/in/gildemberg-santos/"]',
    )
    const github = professionalLinks.get('a[href="https://github.com/gildemberg-santos"]')

    expect(professionalLinks.findAll('a')).toHaveLength(2)
    for (const link of [linkedin, github]) {
      expect(link.attributes('target')).toBe('_blank')
      expect(link.attributes('rel')).toBe('noopener noreferrer')
    }
    expect(text).not.toContain('99136-5507')
    expect(text).not.toContain('CNPJ')
    expect(text).not.toContain('+55')
    expect(text).not.toContain('@')
  })

  it('describes the real implementation, publication date and pending store media', async () => {
    const wrapper = await mountCredits()
    const text = wrapper.text()

    expect(text).toContain('Go + Chi')
    expect(text).toContain('Vue 3 + Vite + TypeScript')
    expect(text).toContain('MongoDB')
    expect(text).toContain('Bacharelado em Ciência da Computação')
    expect(text).toContain('imagens de produto usadas no desenvolvimento são ilustrativas')
    expect(text).toContain('publicados em 25/11/2026')
    expect(text).not.toContain('noindex')
    expect(wrapper.find('a[href="/contato"]').exists()).toBe(true)
    expect(wrapper.find('a[href="/cardapio"]').exists()).toBe(true)
    expect(wrapper.find('a[href="/"]').exists()).toBe(true)
  })
})
