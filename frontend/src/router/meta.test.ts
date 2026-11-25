import { beforeEach, describe, expect, it } from 'vitest'
import type { RouteLocationNormalizedLoaded } from 'vue-router'

import { applyRouteMeta } from './meta'

function route(meta: Record<string, unknown>): RouteLocationNormalizedLoaded {
  return { meta } as unknown as RouteLocationNormalizedLoaded
}

function metaContent(name: string): string | null {
  return document.head.querySelector<HTMLMetaElement>(`meta[name="${name}"]`)?.content ?? null
}

describe('metadados por rota', () => {
  beforeEach(() => {
    document.head
      .querySelectorAll('meta[name="description"], meta[name="robots"]')
      .forEach((element) => element.remove())
    document.title = ''
  })

  it('usa título e descrição da rota', () => {
    applyRouteMeta(
      route({ title: 'Cardápio', description: 'Veja as opções do cardápio e escolha seu lanche.' }),
    )

    expect(document.title).toBe('Cardápio · Brocados Lanches')
    expect(metaContent('description')).toBe('Veja as opções do cardápio e escolha seu lanche.')
    expect(metaContent('robots')).toBe('index, follow')
  })

  it('usa valores padrão quando a rota não define metadados', () => {
    applyRouteMeta(route({}))

    expect(document.title).toBe('Brocados Lanches')
    expect(metaContent('description')).toContain('Cardápio')
  })

  it('marca rotas privadas e de acompanhamento como noindex', () => {
    applyRouteMeta(route({ title: 'Acompanhar pedido', noindex: true }))

    expect(metaContent('robots')).toBe('noindex, nofollow')
  })

  it('atualiza os mesmos elementos em navegações sucessivas', () => {
    applyRouteMeta(route({ title: 'Cardápio' }))
    applyRouteMeta(route({ title: 'Carrinho', noindex: true }))

    expect(document.title).toBe('Carrinho · Brocados Lanches')
    expect(document.head.querySelectorAll('meta[name="robots"]')).toHaveLength(1)
    expect(document.head.querySelectorAll('meta[name="description"]')).toHaveLength(1)
  })

  it('preserva a descrição própria de rotas institucionais', () => {
    const description = 'Canais de contato, endereço e horários publicados pela Brocados Lanches.'

    applyRouteMeta(route({}))
    applyRouteMeta(route({ title: 'Contato', description, noindex: false }))

    expect(document.title).toBe('Contato · Brocados Lanches')
    expect(metaContent('description')).toBe(description)
    expect(metaContent('robots')).toBe('index, follow')
  })
})
