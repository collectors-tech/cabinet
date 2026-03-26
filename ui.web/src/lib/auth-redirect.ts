export function normalizeAuthRedirectTarget(target: string | null | undefined) {
  const trimmed = target?.trim() ?? ''

  if (!trimmed || trimmed === '/' || trimmed === '/?') {
    return undefined
  }

  return trimmed
}
