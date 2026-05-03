import { createFileRoute } from '@tanstack/react-router'
import { priorities } from '@/features/tasks/data/data'
import { Wishlist } from '@/features/wishlist'

const wishlistStatuses = ['wishlist', 'discovered'] as const
const priorityValues = priorities.map((priority) => priority.value)

type WishlistSearch = {
  page?: number
  pageSize?: number
  status?: (typeof wishlistStatuses)[number][]
  priority?: (typeof priorityValues)[number][]
  collection?: string[]
  filter?: string
}

function normalizeSearchArray(value: unknown): string[] {
  if (Array.isArray(value)) {
    return value.flatMap((entry) => normalizeSearchArray(entry))
  }
  if (typeof value !== 'string') {
    return []
  }
  const trimmed = value.trim()
  if (!trimmed) {
    return []
  }
  try {
    const parsed = JSON.parse(trimmed)
    return Array.isArray(parsed)
      ? parsed.flatMap((entry) => normalizeSearchArray(entry))
      : [String(parsed)]
  } catch {
    return [trimmed]
  }
}

function normalizeNumber(value: unknown, fallback: number) {
  if (typeof value === 'number' && Number.isFinite(value)) {
    return value
  }
  if (typeof value === 'string' && value.trim() !== '') {
    const parsed = Number(value)
    return Number.isFinite(parsed) ? parsed : fallback
  }
  return fallback
}

function normalizeEnumArray<T extends string>(
  value: unknown,
  allowedValues: readonly T[]
): T[] {
  const allowed = new Set<string>(allowedValues)
  return normalizeSearchArray(value).filter((entry): entry is T =>
    allowed.has(entry)
  )
}

export const Route = createFileRoute('/_authenticated/wishlist/')({
  validateSearch: (search): WishlistSearch => {
    const page = normalizeNumber(search.page, 1)
    const pageSize = normalizeNumber(search.pageSize, 10)
    const status = normalizeEnumArray(search.status, wishlistStatuses)
    const priority = normalizeEnumArray(search.priority, priorityValues)
    const collection = normalizeSearchArray(search.collection)
    const filter = typeof search.filter === 'string' ? search.filter : ''

    return {
      page: page > 1 ? page : undefined,
      pageSize: pageSize !== 10 ? pageSize : undefined,
      status: status.length > 0 ? status : undefined,
      priority: priority.length > 0 ? priority : undefined,
      collection: collection.length > 0 ? collection : undefined,
      filter: filter.trim() ? filter : undefined,
    }
  },
  component: Wishlist,
})
