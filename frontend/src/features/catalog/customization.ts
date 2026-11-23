import type { MenuAddon, MenuOptionGroup, MenuProduct } from '../../types/catalog'

export type OptionSelections = Record<string, string[]>

export function createOptionSelections(groups: MenuOptionGroup[]): OptionSelections {
  return Object.fromEntries(groups.map((group) => [group.id, []]))
}

export function minimumSelections(group: MenuOptionGroup): number {
  return Math.max(group.required ? 1 : 0, group.minSelections)
}

export function validateOptionGroup(group: MenuOptionGroup, selectedIds: string[]): string {
  if (group.options.length === 0) {
    return 'Este grupo está sem opções disponíveis.'
  }

  const availableIds = new Set(group.options.map((option) => option.id))
  const selectedCount = selectedIds.filter((id) => availableIds.has(id)).length
  const minimum = minimumSelections(group)

  if (selectedCount < minimum) {
    return minimum === 1 ? 'Escolha uma opção.' : `Escolha pelo menos ${minimum} opções.`
  }

  if (selectedCount > group.maxSelections) {
    return `Escolha no máximo ${group.maxSelections} opções.`
  }

  return ''
}

export function validateOptionGroups(
  groups: MenuOptionGroup[],
  selections: OptionSelections,
): Record<string, string> {
  const errors: Record<string, string> = {}

  for (const group of groups) {
    const message = validateOptionGroup(group, selections[group.id] ?? [])

    if (message.length > 0) {
      errors[group.id] = message
    }
  }

  return errors
}

export function isOptionSelectionValid(
  groups: MenuOptionGroup[],
  selections: OptionSelections,
): boolean {
  return Object.keys(validateOptionGroups(groups, selections)).length === 0
}

export function calculateConfiguredPriceCents(
  product: MenuProduct,
  groups: MenuOptionGroup[],
  selections: OptionSelections,
  addons: MenuAddon[],
  selectedAddonIds: string[],
): number {
  let total = product.priceCents

  const selectedOptionIds = new Set(Object.values(selections).flat())

  for (const group of groups) {
    for (const option of group.options) {
      if (option.available && selectedOptionIds.has(option.id)) {
        total += option.priceCents
      }
    }
  }

  const selectedAddons = new Set(selectedAddonIds)

  for (const addon of addons) {
    if (addon.available && selectedAddons.has(addon.id)) {
      total += addon.priceCents
    }
  }

  return total
}
