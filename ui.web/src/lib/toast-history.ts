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
  created_at: string
}

type ToastLike<TArgs extends readonly unknown[]> = {
  message?: (...args: TArgs) => unknown
  success?: (...args: TArgs) => unknown
  error?: (...args: TArgs) => unknown
  info?: (...args: TArgs) => unknown
  warning?: (...args: TArgs) => unknown
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

function summaryFromOptions(value: unknown) {
  if (!value || typeof value !== 'object') {
    return ''
  }
  const candidate = value as { description?: unknown }
  return normalizeText(candidate.description)
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

export function installToastHistoryCapture<TArgs extends readonly unknown[]>(
  toast: ToastLike<TArgs>
) {
  if (installed) {
    return
  }
  installed = true
  ;(
    ['message', 'success', 'error', 'info', 'warning'] as ToastHistoryLevel[]
  ).forEach((level) => {
    const original = toast[level]
    if (!original) {
      return
    }
    toast[level] = (...args: TArgs) => {
      recordToastHistory({
        level,
        title: normalizeText(args[0]),
        summary: summaryFromOptions(args[1]),
      })
      return original(...args)
    }
  })
}
