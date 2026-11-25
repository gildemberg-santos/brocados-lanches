import { mount } from '@vue/test-utils'
import { defineComponent, h } from 'vue'
import { createMemoryHistory, createRouter } from 'vue-router'
import { describe, expect, it } from 'vitest'

import NotFoundView from './NotFoundView.vue'

const StubView = defineComponent({ name: 'StubView', render: () => h('div') })

async function mountNotFound() {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', component: StubView },
      { path: '/cardapio', component: StubView },
      { path: '/:pathMatch(.*)*', component: StubView },
    ],
  })
  await router.push('/url-desconhecida')
  await router.isReady()

  return mount(NotFoundView, { global: { plugins: [router] } })
}

describe('NotFoundView', () => {
  it('explains that the page was not found', async () => {
    const wrapper = await mountNotFound()

    expect(wrapper.get('h2').text()).toBe('Página não encontrada')
    expect(wrapper.text()).toContain('O endereço acessado não existe ou foi movido')
  })

  it('offers links back to the home page and the menu', async () => {
    const wrapper = await mountNotFound()

    expect(wrapper.get('a[href="/"]').text()).toBe('Voltar para o início')
    expect(wrapper.get('a[href="/cardapio"]').text()).toBe('Ver cardápio')
  })
})
