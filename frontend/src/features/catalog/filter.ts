import type { CatalogProduct } from '../../types/catalog'
import { normalizeText, splitSearchTerms } from '../../utils/text'

export function matchesProductQuery(product: CatalogProduct, query: string): boolean {
  const terms = splitSearchTerms(query)

  if (terms.length === 0) {
    return true
  }

  const haystack = normalizeText(`${product.name} ${product.description}`)

  return terms.every((term) => haystack.includes(term))
}

export function filterProducts(
  products: CatalogProduct[],
  categorySlug: string,
  query: string,
): CatalogProduct[] {
  return products.filter((product) => {
    const matchesCategory = categorySlug.length === 0 || product.categorySlug === categorySlug

    return matchesCategory && matchesProductQuery(product, query)
  })
}
