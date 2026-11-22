const DATE_TIME_FORMATTER = new Intl.DateTimeFormat('pt-BR', {
  timeZone: 'America/Fortaleza',
  day: '2-digit',
  month: '2-digit',
  year: 'numeric',
  hour: '2-digit',
  minute: '2-digit',
})

/** Datas chegam em UTC; a operação é exibida em America/Fortaleza. */
export function formatOrderDateTime(value: string): string {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return '—'
  }
  return DATE_TIME_FORMATTER.format(date)
}

export function formatOrderTime(value: string): string {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return '—'
  }
  return new Intl.DateTimeFormat('pt-BR', {
    timeZone: 'America/Fortaleza',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date)
}
