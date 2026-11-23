import type { RouteLocationNormalizedLoaded } from 'vue-router'

const SITE_NAME = 'Brocados Lanches'
const DEFAULT_DESCRIPTION = 'Cardápio, pedidos e acompanhamento da Brocados Lanches.'

function upsertMeta(name: 'description' | 'robots', content: string): void {
  let element = document.head.querySelector<HTMLMetaElement>(`meta[name="${name}"]`)
  if (!element) {
    element = document.createElement('meta')
    element.name = name
    document.head.appendChild(element)
  }
  element.content = content
}

/** Aplica título e descrição por rota sem dependências de terceiros. */
export function applyRouteMeta(route: RouteLocationNormalizedLoaded): void {
  const rawTitle = route.meta.title
  const title = typeof rawTitle === 'string' && rawTitle.length > 0 ? rawTitle : SITE_NAME
  document.title = title === SITE_NAME ? title : `${title} · ${SITE_NAME}`

  const rawDescription = route.meta.description
  upsertMeta(
    'description',
    typeof rawDescription === 'string' && rawDescription.length > 0
      ? rawDescription
      : DEFAULT_DESCRIPTION,
  )
  // Telas administrativas e de pedido não têm valor de indexação pública.
  upsertMeta('robots', route.meta.noindex === true ? 'noindex, nofollow' : 'index, follow')
}
