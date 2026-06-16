type ShortcutID = 'command-palette' | 'sidebar-toggle'

type ShortcutDiagnosticsEntry = {
  id: ShortcutID
  requested: string
  fallback: string
  reason: 'invalid' | 'reserved' | 'duplicate'
}

const DEFAULT_SHORTCUTS: Record<ShortcutID, string> = {
  'command-palette': 'k',
  'sidebar-toggle': 'b',
}

const REGISTRATION_ORDER: ShortcutID[] = ['command-palette', 'sidebar-toggle']

const RESERVED_SHORTCUT_KEYS = new Set(['r', 'w', 't', 'l', ',', 'f'])

const OVERRIDES_STORAGE_KEY = 'cabinet.shortcuts.overrides'

let lastDiagnostics: ShortcutDiagnosticsEntry[] = []

function normalizeShortcutKey(value: unknown): string {
  if (typeof value !== 'string') {
    return ''
  }
  return value.trim().toLowerCase()
}

function isValidShortcutKey(value: string): boolean {
  return /^[a-z]$/.test(value)
}

function pickFallbackShortcut(id: ShortcutID, usedKeys: Set<string>): string {
  const candidates = [
    DEFAULT_SHORTCUTS[id],
    ...Object.values(DEFAULT_SHORTCUTS),
    'p',
    'g',
    'o',
  ]
  for (const candidate of candidates) {
    if (!isValidShortcutKey(candidate)) {
      continue
    }
    if (RESERVED_SHORTCUT_KEYS.has(candidate)) {
      continue
    }
    if (usedKeys.has(candidate)) {
      continue
    }
    return candidate
  }
  return DEFAULT_SHORTCUTS[id]
}

function readShortcutOverrides(): Partial<Record<ShortcutID, string>> {
  if (typeof window === 'undefined') {
    return {}
  }
  const raw = window.localStorage.getItem(OVERRIDES_STORAGE_KEY)
  if (!raw) {
    return {}
  }
  try {
    const parsed = JSON.parse(raw) as Partial<Record<ShortcutID, string>>
    return parsed && typeof parsed === 'object' ? parsed : {}
  } catch {
    return {}
  }
}

function applyDiagnosticsToWindow(
  diagnostics: ShortcutDiagnosticsEntry[],
  resolvedShortcuts: Record<ShortcutID, string>
) {
  if (typeof window === 'undefined') {
    return
  }
  ;(
    window as Window & {
      __cabinetShortcutDiagnostics?: ShortcutDiagnosticsEntry[]
    }
  ).__cabinetShortcutDiagnostics = diagnostics
  ;(
    window as Window & { __cabinetShortcuts?: Record<ShortcutID, string> }
  ).__cabinetShortcuts = resolvedShortcuts
}

function resolveShortcuts(): Record<ShortcutID, string> {
  const overrides = readShortcutOverrides()
  const usedKeys = new Set<string>()
  const resolved: Record<ShortcutID, string> = {
    'command-palette': DEFAULT_SHORTCUTS['command-palette'],
    'sidebar-toggle': DEFAULT_SHORTCUTS['sidebar-toggle'],
  }
  const diagnostics: ShortcutDiagnosticsEntry[] = []

  for (const id of REGISTRATION_ORDER) {
    const requested = normalizeShortcutKey(
      overrides[id] ?? DEFAULT_SHORTCUTS[id]
    )
    const isInvalid = !isValidShortcutKey(requested)
    const isReserved = RESERVED_SHORTCUT_KEYS.has(requested)
    const isDuplicate = usedKeys.has(requested)

    if (isInvalid || isReserved || isDuplicate) {
      const fallback = pickFallbackShortcut(id, usedKeys)
      diagnostics.push({
        id,
        requested: requested || String(overrides[id] ?? ''),
        fallback,
        reason: isInvalid ? 'invalid' : isReserved ? 'reserved' : 'duplicate',
      })
      resolved[id] = fallback
      usedKeys.add(fallback)
      continue
    }

    resolved[id] = requested
    usedKeys.add(requested)
  }

  lastDiagnostics = diagnostics
  applyDiagnosticsToWindow(diagnostics, resolved)
  return resolved
}

export function getShortcutKey(id: ShortcutID): string {
  return resolveShortcuts()[id]
}

export function getShortcutDiagnostics(): ShortcutDiagnosticsEntry[] {
  resolveShortcuts()
  return lastDiagnostics
}

if (typeof window !== 'undefined') {
  resolveShortcuts()
}
