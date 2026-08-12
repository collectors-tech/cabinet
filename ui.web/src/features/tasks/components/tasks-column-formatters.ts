import { type Task } from '../data/schema'

export function formatMoney(value: number | undefined) {
  if (typeof value !== 'number' || value <= 0) {
    return '-'
  }
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
  }).format(value)
}

export function formatWishlistDate(value: string | undefined) {
  const trimmed = value?.trim()
  if (!trimmed) {
    return '-'
  }
  const datePart = trimmed.split('T')[0]?.split(' ')[0] ?? trimmed
  const parts = datePart.split('-').map((part) => Number(part))
  if (
    parts.length === 3 &&
    Number.isInteger(parts[0]) &&
    Number.isInteger(parts[1]) &&
    Number.isInteger(parts[2])
  ) {
    const [year, month, day] = parts
    return new Intl.DateTimeFormat('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      timeZone: 'UTC',
    }).format(new Date(Date.UTC(year, month - 1, day)))
  }
  return trimmed
}

export function buildWishlistPricePointRows(task: Task, values: number[]) {
  const dates = task.priceHistoryDates ?? []
  return values
    .map((price, index) => ({
      date: dates[index] ?? `Point ${index + 1}`,
      price,
    }))
    .slice(-10)
}
