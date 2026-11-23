import { parseCentsInput } from '../checkout/checkout'
import type {
  AdminAddonInput,
  AdminCategoryInput,
  AdminCatalogKind,
  AdminOpeningHour,
  AdminProductInput,
  AdminStoreInput,
  AdminTimeInterval,
  AdminWeekday,
  AdminZoneInput,
} from '../../types/adminCatalog'

const SLUG_PATTERN = /^[a-z0-9]+(?:-[a-z0-9]+)*$/
const OBJECT_ID_PATTERN = /^[a-f0-9]{24}$/i
const WHATSAPP_PATTERN = /^\+[1-9][0-9]{7,14}$/
const TIME_PATTERN = /^([01][0-9]|2[0-3]):[0-5][0-9]$/

export const WEEKDAYS: AdminWeekday[] = [
  'sunday',
  'monday',
  'tuesday',
  'wednesday',
  'thursday',
  'friday',
  'saturday',
]

export const WEEKDAY_LABELS: Record<AdminWeekday, string> = {
  sunday: 'Domingo',
  monday: 'Segunda-feira',
  tuesday: 'Terça-feira',
  wednesday: 'Quarta-feira',
  thursday: 'Quinta-feira',
  friday: 'Sexta-feira',
  saturday: 'Sábado',
}

export function slugify(value: string): string {
  return value
    .normalize('NFD')
    .replace(/[\u0300-\u036f]/g, '')
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
    .slice(0, 80)
    .replace(/-+$/g, '')
}

export function isValidSlug(value: string): boolean {
  return value.length > 0 && value.length <= 80 && SLUG_PATTERN.test(value)
}

export function isValidObjectId(value: string): boolean {
  return OBJECT_ID_PATTERN.test(value)
}

export function isValidTime(value: string): boolean {
  return TIME_PATTERN.test(value)
}

export function moneyToCents(value: string): number | null {
  return parseCentsInput(value)
}

export function allergensToText(allergens: string[]): string {
  return allergens.join(', ')
}

export function allergensFromText(value: string): string[] {
  return value
    .split(',')
    .map((item) => item.trim())
    .filter((item) => item.length > 0)
}

export function newLocalId(prefix: string): string {
  const random =
    globalThis.crypto && 'randomUUID' in globalThis.crypto
      ? globalThis.crypto.randomUUID()
      : `${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 10)}`
  return `${prefix}-${random}`
}

export function emptyInterval(): AdminTimeInterval {
  return { open: '', close: '' }
}

export function emptyOpeningHours(): AdminOpeningHour[] {
  return WEEKDAYS.map((day) => ({ day, closed: true, intervals: [] }))
}

function toMinutes(value: string): number {
  const [hours, minutes] = value.split(':').map(Number)
  return hours * 60 + minutes
}

function intervalsOverlap(intervals: AdminTimeInterval[]): boolean {
  const sorted = [...intervals].sort((a, b) => toMinutes(a.open) - toMinutes(b.open))
  for (let index = 1; index < sorted.length; index += 1) {
    const previousEnd = toMinutes(sorted[index - 1].close)
    const currentStart = toMinutes(sorted[index].open)
    const previousStart = toMinutes(sorted[index - 1].open)
    const end = previousEnd < previousStart ? previousEnd + 24 * 60 : previousEnd
    if (currentStart < end) {
      return true
    }
  }
  return false
}

export type FieldErrors = Record<string, string>

export function validateSimpleInput(
  kind: AdminCatalogKind,
  input: AdminCategoryInput | AdminAddonInput | AdminZoneInput,
): FieldErrors {
  const errors: FieldErrors = {}

  if (!input.name.trim()) {
    errors.name = 'Informe o nome.'
  } else if (input.name.length > 120) {
    errors.name = 'O nome deve ter no máximo 120 caracteres.'
  }

  if (!isValidSlug(input.slug)) {
    errors.slug = 'Use letras minúsculas, números e hífens simples.'
  }

  if (kind === 'category' && 'description' in input && input.description.length > 2000) {
    errors.description = 'A descrição deve ter no máximo 2000 caracteres.'
  }

  const amount =
    kind === 'addon' ? (input as AdminAddonInput).priceCents : (input as AdminZoneInput).feeCents
  if (kind !== 'category' && (!Number.isSafeInteger(amount) || amount < 0)) {
    errors[kind === 'addon' ? 'priceCents' : 'feeCents'] = 'Informe um valor maior ou igual a zero.'
  }

  if (!Number.isSafeInteger(input.sortOrder) || input.sortOrder < 0) {
    errors.sortOrder = 'A ordem deve ser um número inteiro positivo.'
  }

  return errors
}

export function validateProductInput(input: AdminProductInput): FieldErrors {
  const errors: FieldErrors = {}

  if (!isValidObjectId(input.categoryId)) {
    errors.categoryId = 'Selecione uma categoria.'
  }
  if (!input.name.trim()) {
    errors.name = 'Informe o nome.'
  }
  if (!isValidSlug(input.slug)) {
    errors.slug = 'Use letras minúsculas, números e hífens simples.'
  }
  if (!input.description.trim()) {
    errors.description = 'Informe a descrição.'
  }
  if (!Number.isSafeInteger(input.priceCents) || input.priceCents < 0) {
    errors.priceCents = 'Informe um preço maior ou igual a zero.'
  }
  if (input.imageUrl && !input.imageAlt.trim()) {
    errors.imageAlt = 'Descreva a imagem para quem usa leitor de tela.'
  }
  if (input.allergens.some((item) => item.length > 60)) {
    errors.allergens = 'Cada alergênico deve ter no máximo 60 caracteres.'
  }
  if (input.addonIds.some((id) => !isValidObjectId(id))) {
    errors.addonIds = 'Selecione adicionais válidos.'
  }
  if (!Number.isSafeInteger(input.sortOrder) || input.sortOrder < 0) {
    errors.sortOrder = 'A ordem deve ser um número inteiro positivo.'
  }

  input.optionGroups.forEach((group, groupIndex) => {
    const prefix = `optionGroups.${groupIndex}`
    if (!group.name.trim()) {
      errors[`${prefix}.name`] = 'Informe o nome do grupo.'
    }
    if (!Number.isSafeInteger(group.minSelections) || group.minSelections < 0) {
      errors[`${prefix}.minSelections`] = 'Mínimo inválido.'
    }
    if (!Number.isSafeInteger(group.maxSelections) || group.maxSelections < 0) {
      errors[`${prefix}.maxSelections`] = 'Máximo inválido.'
    } else if (group.maxSelections > 0 && group.maxSelections < group.minSelections) {
      errors[`${prefix}.maxSelections`] = 'O máximo não pode ser menor que o mínimo.'
    }

    group.options.forEach((option, optionIndex) => {
      const optionPrefix = `${prefix}.options.${optionIndex}`
      if (!option.name.trim()) {
        errors[`${optionPrefix}.name`] = 'Informe o nome da opção.'
      }
      if (!Number.isSafeInteger(option.priceCents) || option.priceCents < 0) {
        errors[`${optionPrefix}.priceCents`] = 'Informe um preço válido.'
      }
    })
  })

  return errors
}

export function validateStoreInput(input: AdminStoreInput): FieldErrors {
  const errors: FieldErrors = {}

  if (!input.name.trim()) {
    errors.name = 'Informe o nome da loja.'
  }
  if (!input.deliveryTime.trim()) {
    errors.deliveryTime = 'Informe o prazo de entrega.'
  }
  if (!WHATSAPP_PATTERN.test(input.whatsapp)) {
    errors.whatsapp = 'Use o WhatsApp em formato internacional, por exemplo +5585999999999.'
  }
  if (!input.address.trim()) {
    errors.address = 'Informe o endereço da loja.'
  }

  input.openingHours.forEach((hour, dayIndex) => {
    if (hour.closed) {
      return
    }
    if (hour.intervals.length === 0) {
      errors[`openingHours.${dayIndex}`] = 'Adicione pelo menos um intervalo para o dia aberto.'
      return
    }

    hour.intervals.forEach((interval, intervalIndex) => {
      const prefix = `openingHours.${dayIndex}.intervals.${intervalIndex}`
      if (!isValidTime(interval.open)) {
        errors[`${prefix}.open`] = 'Use o formato HH:MM.'
      }
      if (!isValidTime(interval.close)) {
        errors[`${prefix}.close`] = 'Use o formato HH:MM.'
      } else if (isValidTime(interval.open) && interval.open === interval.close) {
        errors[`${prefix}.close`] = 'O fechamento deve ser diferente da abertura.'
      }
    })

    const complete = hour.intervals.filter(
      (interval) => isValidTime(interval.open) && isValidTime(interval.close),
    )
    if (complete.length === hour.intervals.length && intervalsOverlap(complete)) {
      errors[`openingHours.${dayIndex}`] = 'Os intervalos do dia não podem se sobrepor.'
    }
  })

  return errors
}

export function firstFieldError(errors: FieldErrors): string {
  const values = Object.values(errors)
  return values[0] ?? ''
}
