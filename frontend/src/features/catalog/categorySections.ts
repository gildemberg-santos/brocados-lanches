/** Limiares do scroll spy de categorias, em pixels. */
export const CATEGORY_TOP_THRESHOLD = 20
export const CATEGORY_EDGE_OFFSET = 10
export const CATEGORY_LAST_EDGE_OFFSET = -90
export const CATEGORY_BOTTOM_TOLERANCE = 1

/** Atributo usado para medir as seções sem depender do texto do rótulo. */
export const CATEGORY_SECTION_ATTRIBUTE = 'data-category-section'

export interface CategorySectionMetric {
  slug: string
  top: number
  bottom: number
}

export type CategoryScrollDirection = 'up' | 'down'

/**
 * ID estável da seção de uma categoria. O slug vem da API e é único por
 * cardápio, então a âncora sobrevive a mudanças de nome da categoria.
 */
export function categorySectionId(slug: string): string {
  return `categoria-${slug}`
}

/**
 * Resolve a categoria ativa a partir das posições já medidas das seções.
 *
 * O scroll para baixo avança quando o topo de uma seção cruza 10 px; o scroll
 * para cima retrocede quando a base da seção ainda passa de 10 px. A última
 * seção usa -90 px porque normalmente não há conteúdo suficiente após ela.
 * Topo e fim do documento são casos especiais para não deixar nenhuma aba
 * ativa durante o carrossel inicial nem no fim da página.
 */
export function resolveActiveCategory(
  sections: CategorySectionMetric[],
  scrollY: number,
  direction: CategoryScrollDirection,
  maxScroll: number,
): string {
  if (sections.length === 0) {
    return ''
  }

  const first = sections[0]!
  const last = sections[sections.length - 1]!

  if (scrollY <= CATEGORY_TOP_THRESHOLD) {
    return first.slug
  }

  if (Math.abs(maxScroll - scrollY) <= CATEGORY_BOTTOM_TOLERANCE) {
    return last.slug
  }

  if (direction === 'up') {
    for (const section of sections) {
      const offset = section.slug === last.slug ? CATEGORY_LAST_EDGE_OFFSET : CATEGORY_EDGE_OFFSET
      if (section.bottom - scrollY > offset) {
        return section.slug
      }
    }
    return first.slug
  }

  let active = first.slug
  for (const section of sections) {
    if (section.top - scrollY > CATEGORY_EDGE_OFFSET) {
      break
    }
    active = section.slug
  }
  return active
}
