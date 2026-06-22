export const TOAST_HISTORY_STORAGE_KEY = 'cabinet.toastHistory.v1'

export type ToastHistoryLevel =
  | 'message'
  | 'success'
  | 'error'
  | 'info'
  | 'warning'

export type ToastHistoryRecord = {
  id: string
  level: ToastHistoryLevel
  title: string
  summary?: string
  source_label?: string
  category?: string
  created_at: string
}

type ToastMethod = (...args: unknown[]) => unknown

type ToastLike = Partial<Record<ToastHistoryLevel, ToastMethod>> & {
  promise?: ToastMethod
}

const MAX_TOAST_HISTORY = 100
let installed = false

function isBrowser() {
  return typeof window !== 'undefined' && Boolean(window.localStorage)
}

function normalizeText(value: unknown) {
  if (typeof value === 'string') {
    return value
  }
  if (typeof value === 'number' || typeof value === 'boolean') {
    return String(value)
  }
  return ''
}

function titleFromValue(value: unknown, fallback: string) {
  return normalizeText(value).trim() || fallback
}

function summaryFromOptions(value: unknown) {
  if (!value || typeof value !== 'object') {
    return ''
  }
  const candidate = value as { description?: unknown }
  return normalizeText(candidate.description)
}

function promiseMessage(
  value: unknown,
  fallback: string
): { title: string; summary?: string } {
  if (typeof value === 'string') {
    return { title: value }
  }
  if (!value || typeof value !== 'object') {
    return { title: fallback }
  }
  const candidate = value as {
    loading?: unknown
    success?: unknown
    error?: unknown
    description?: unknown
  }
  const title =
    titleFromValue(candidate.loading, '') ||
    titleFromValue(candidate.success, '') ||
    titleFromValue(candidate.error, '') ||
    fallback
  const summary = normalizeText(candidate.description)
  return { title, summary: summary || undefined }
}

function toastId() {
  if (typeof crypto !== 'undefined' && 'randomUUID' in crypto) {
    return crypto.randomUUID()
  }
  return `${Date.now()}-${Math.random().toString(16).slice(2)}`
}

export function loadToastHistory(): ToastHistoryRecord[] {
  if (!isBrowser()) {
    return []
  }
  try {
    const raw = window.localStorage.getItem(TOAST_HISTORY_STORAGE_KEY)
    if (!raw) {
      return []
    }
    const parsed = JSON.parse(raw) as ToastHistoryRecord[]
    if (!Array.isArray(parsed)) {
      return []
    }
    return parsed.filter((record) => Boolean(record?.id && record?.title))
  } catch {
    return []
  }
}

export function saveToastHistory(records: ToastHistoryRecord[]) {
  if (!isBrowser()) {
    return
  }
  window.localStorage.setItem(
    TOAST_HISTORY_STORAGE_KEY,
    JSON.stringify(records.slice(0, MAX_TOAST_HISTORY))
  )
  window.dispatchEvent(new Event('cabinet:toast-history'))
}

export function recordToastHistory(
  record: Omit<ToastHistoryRecord, 'id' | 'created_at'>
) {
  const title = record.title.trim()
  if (!title) {
    return
  }
  saveToastHistory([
    {
      ...record,
      id: toastId(),
      title,
      created_at: new Date().toISOString(),
    },
    ...loadToastHistory(),
  ])
}

export function recordNotificationHistory(
  record: Omit<ToastHistoryRecord, 'id' | 'created_at' | 'level'> & {
    level?: ToastHistoryLevel
  }
) {
  recordToastHistory({
    level: record.level ?? 'info',
    title: record.title,
    summary: record.summary,
    source_label: record.source_label,
    category: record.category,
  })
}

function recordPromiseResolution(
  value: unknown,
  level: 'success' | 'error',
  fallback: string
) {
  if (typeof value === 'string') {
    recordToastHistory({
      level,
      title: value,
      summary: 'Promise toast settled and was preserved in Inbox history.',
    })
    return
  }
  recordToastHistory({
    level,
    title: fallback,
    summary: 'Promise toast settled and was preserved in Inbox history.',
  })
}

export function installToastHistoryCapture(toast: unknown) {
  if (installed) {
    return
  }
  installed = true
  const toastLike = toast as ToastLike
  ;(
    ['message', 'success', 'error', 'info', 'warning'] as ToastHistoryLevel[]
  ).forEach((level) => {
    const original = toastLike[level]
    if (!original) {
      return
    }
    toastLike[level] = (...args: unknown[]) => {
      recordToastHistory({
        level,
        title: normalizeText(args[0]),
        summary: summaryFromOptions(args[1]),
      })
      return original(...args)
    }
  })
  const originalPromise = toastLike.promise
  if (originalPromise) {
    toastLike.promise = (...args: unknown[]) => {
      const [, messages] = args
      const loading = promiseMessage(messages, 'Async notification started')
      recordToastHistory({
        level: 'info',
        title: loading.title,
        summary:
          loading.summary ||
          'Promise toast started and was preserved in Inbox history.',
      })

      Promise.resolve(args[0])
        .then(() => {
          const success =
            messages && typeof messages === 'object'
              ? (messages as { success?: unknown }).success
              : undefined
          recordPromiseResolution(
            success,
            'success',
            'Async notification completed'
          )
        })
        .catch(() => {
          const error =
            messages && typeof messages === 'object'
              ? (messages as { error?: unknown }).error
              : undefined
          recordPromiseResolution(error, 'error', 'Async notification failed')
        })

      return originalPromise(...args)
    }
  }
}
