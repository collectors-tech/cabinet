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

type ToastHistoryMetadata = {
  id?: string
  level?: ToastHistoryLevel
  title?: string
  summary?: string
  source_label?: string
  category?: string
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

function historyMetadataFromOptions(value: unknown): ToastHistoryMetadata {
  if (!value || typeof value !== 'object') {
    return {}
  }
  const candidate = value as { history?: ToastHistoryMetadata }
  return candidate.history ?? {}
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

function promiseHistoryMetadata(value: unknown): ToastHistoryMetadata {
  if (!value || typeof value !== 'object') {
    return {}
  }
  const candidate = value as { history?: ToastHistoryMetadata }
  return candidate.history ?? {}
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
  record: Omit<ToastHistoryRecord, 'id' | 'created_at'> & { id?: string }
) {
  const title = record.title.trim()
  if (!title) {
    return
  }
  const id = record.id ?? toastId()
  saveToastHistory([
    {
      ...record,
      id,
      title,
      created_at: new Date().toISOString(),
    },
    ...loadToastHistory().filter((existing) => existing.id !== id),
  ])
}

export function recordNotificationHistory(
  record: Omit<ToastHistoryRecord, 'id' | 'created_at' | 'level'> & {
    id?: string
    level?: ToastHistoryLevel
  }
) {
  recordToastHistory({
    id: record.id,
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
  fallback: string,
  history: ToastHistoryMetadata = {}
) {
  const title = titleFromValue(value, fallback)
  if (title !== fallback || typeof value === 'string') {
    recordToastHistory({
      id: history.id ? `${history.id}-${level}` : undefined,
      level,
      title,
      summary:
        history.summary ||
        'Promise toast settled and was preserved in Inbox history.',
      source_label: history.source_label,
      category: history.category,
    })
    return
  }
  recordToastHistory({
    id: history.id ? `${history.id}-${level}` : undefined,
    level,
    title: fallback,
    summary:
      history.summary ||
      'Promise toast settled and was preserved in Inbox history.',
    source_label: history.source_label,
    category: history.category,
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
      const history = historyMetadataFromOptions(args[1])
      recordToastHistory({
        id: history.id,
        level: history.level ?? level,
        title: titleFromValue(history.title, normalizeText(args[0])),
        summary: history.summary || summaryFromOptions(args[1]),
        source_label: history.source_label,
        category: history.category,
      })
      return original(...args)
    }
  })
  const originalPromise = toastLike.promise
  if (originalPromise) {
    toastLike.promise = (...args: unknown[]) => {
      const [, messages] = args
      const history = promiseHistoryMetadata(messages)
      const loading = promiseMessage(messages, 'Async notification started')
      recordToastHistory({
        id: history.id,
        level: 'info',
        title: history.title ?? loading.title,
        summary:
          history.summary ||
          loading.summary ||
          'Promise toast started and was preserved in Inbox history.',
        source_label: history.source_label,
        category: history.category,
      })

      let forwardedArgs = args
      let successCapturedByCallback = false
      let errorCapturedByCallback = false
      if (messages && typeof messages === 'object') {
        const wrappedMessages = { ...(messages as Record<string, unknown>) }
        if (typeof wrappedMessages.success === 'function') {
          const originalSuccess = wrappedMessages.success as (
            value: unknown
          ) => unknown
          successCapturedByCallback = true
          wrappedMessages.success = (value: unknown) => {
            const result = originalSuccess(value)
            recordPromiseResolution(
              result,
              'success',
              'Async notification completed',
              history
            )
            return result
          }
        }
        if (typeof wrappedMessages.error === 'function') {
          const originalError = wrappedMessages.error as (
            value: unknown
          ) => unknown
          errorCapturedByCallback = true
          wrappedMessages.error = (value: unknown) => {
            const result = originalError(value)
            recordPromiseResolution(
              result,
              'error',
              'Async notification failed',
              history
            )
            return result
          }
        }
        forwardedArgs = [args[0], wrappedMessages, ...args.slice(2)]
      }

      Promise.resolve(args[0])
        .then(() => {
          if (successCapturedByCallback) {
            return
          }
          const success =
            messages && typeof messages === 'object'
              ? (messages as { success?: unknown }).success
              : undefined
          recordPromiseResolution(
            success,
            'success',
            'Async notification completed',
            history
          )
        })
        .catch(() => {
          if (errorCapturedByCallback) {
            return
          }
          const error =
            messages && typeof messages === 'object'
              ? (messages as { error?: unknown }).error
              : undefined
          recordPromiseResolution(
            error,
            'error',
            'Async notification failed',
            history
          )
        })

      return originalPromise(...forwardedArgs)
    }
  }
}
