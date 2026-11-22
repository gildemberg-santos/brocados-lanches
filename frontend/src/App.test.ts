import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { defineComponent, h } from 'vue'
import { createMemoryHistory, createRouter } from 'vue-router'
import { describe, expect, it } from 'vitest'

import App from './App.vue'

const StubView = defineComponent({
  name: 'StubView',
  render: () => h('div', 'stub'),
})

async function mountApp() {
  const pinia = createPinia()
  setActivePinia(pinia)

  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', component: StubView },
      { path: '/cardapio', component: StubView },
      { path: '/carrinho', component: StubView },
    ],
  })

  await router.push('/')
  await router.isReady()

  return mount(App, {
    global: {
      plugins: [pinia, router],
    },
  })
}

describe('App', () => {
  it('renders the shared header and the active route through RouterView', async () => {
    const wrapper = await mountApp()

    expect(wrapper.text()).toContain('BROCADOS')
    expect(wrapper.text()).toContain('stub')
  })

  it('offers a skip link to the main content', async () => {
    const wrapper = await mountApp()

    expect(wrapper.get('.skip-link').attributes('href')).toBe('#conteudo')
  })
})
