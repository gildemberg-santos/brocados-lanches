import type { MenuCategory, MenuProduct } from '../../types/catalog'
import type { PublicStoreStatus } from '../../types/store'

type JsonLdObject = Record<string, unknown>

const SCHEMA_CONTEXT = 'https://schema.org'

/** Converte centavos inteiros para o formato decimal exigido pelo schema.org, sem ponto flutuante. */
export function centsToSchemaPrice(cents: number): string {
  const sign = cents < 0 ? '-' : ''
  const absolute = Math.abs(cents)
  return `${sign}${Math.trunc(absolute / 100)}.${String(absolute % 100).padStart(2, '0')}`
}

function weekday(day: string): string {
  return `${SCHEMA_CONTEXT}/${day.charAt(0).toUpperCase()}${day.slice(1)}`
}

/** JSON-LD da loja usando somente dados devolvidos por GET /store. */
export function buildRestaurantJsonLd(status: PublicStoreStatus): JsonLdObject {
  const data: JsonLdObject = {
    '@context': SCHEMA_CONTEXT,
    '@type': 'Restaurant',
    name: status.name,
  }

  const telephone = status.whatsapp.trim()
  if (telephone.length > 0) {
    data.telephone = telephone
  }

  const streetAddress = status.address.trim()
  if (streetAddress.length > 0) {
    data.address = { '@type': 'PostalAddress', streetAddress }
  }

  const openingHours = status.openingHours.flatMap((day) =>
    day.closed
      ? []
      : day.intervals.map((interval) => ({
          '@type': 'OpeningHoursSpecification',
          dayOfWeek: weekday(day.day),
          opens: interval.open,
          closes: interval.close,
        })),
  )
  if (openingHours.length > 0) {
    data.openingHoursSpecification = openingHours
  }

  return data
}

function buildMenuItem(product: MenuProduct): JsonLdObject {
  const item: JsonLdObject = {
    '@type': 'MenuItem',
    name: product.name,
    offers: {
      '@type': 'Offer',
      price: centsToSchemaPrice(product.priceCents),
      priceCurrency: 'BRL',
      availability: product.available
        ? `${SCHEMA_CONTEXT}/InStock`
        : `${SCHEMA_CONTEXT}/OutOfStock`,
    },
  }

  if (product.description.trim().length > 0) {
    item.description = product.description.trim()
  }
  if (product.imageUrl.trim().length > 0) {
    item.image = product.imageUrl.trim()
  }

  return item
}

/** JSON-LD do cardápio, agrupado pelas categorias devolvidas por GET /menu. */
export function buildMenuJsonLd(categories: MenuCategory[]): JsonLdObject | null {
  const sections = categories
    .map((category) => ({
      '@type': 'MenuSection',
      name: category.name,
      hasMenuItem: category.products.map(buildMenuItem),
    }))
    .filter((section) => section.hasMenuItem.length > 0)

  if (sections.length === 0) {
    return null
  }

  return {
    '@context': SCHEMA_CONTEXT,
    '@type': 'Menu',
    hasMenuSection: sections,
  }
}

/** JSON-LD de um produto individual; útil na rota de detalhe. */
export function buildProductJsonLd(product: MenuProduct): JsonLdObject {
  const data: JsonLdObject = {
    '@context': SCHEMA_CONTEXT,
    '@type': 'Product',
    name: product.name,
    offers: {
      '@type': 'Offer',
      price: centsToSchemaPrice(product.priceCents),
      priceCurrency: 'BRL',
      availability: product.available
        ? `${SCHEMA_CONTEXT}/InStock`
        : `${SCHEMA_CONTEXT}/OutOfStock`,
    },
  }

  if (product.description.trim().length > 0) {
    data.description = product.description.trim()
  }
  if (product.imageUrl.trim().length > 0) {
    data.image = product.imageUrl.trim()
  }

  return data
}

/** Encapsula um ou mais objetos no formato aceito por consumidores de JSON-LD. */
export function serializeJsonLd(objects: JsonLdObject[]): string {
  if (objects.length === 0) {
    return ''
  }
  const value =
    objects.length === 1 ? objects[0] : { '@context': SCHEMA_CONTEXT, '@graph': objects }
  return JSON.stringify(value)
}
