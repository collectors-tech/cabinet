import {
  normalizeDisplayOption,
  normalizeDisplayOptions,
} from './item-type-condition-scales'

export const inventoryPackagingGradesSettingsKey = 'grading.enums.packaging'

export const defaultInventoryPackagingGrades = [
  'sealed_mint',
  'sealed_good',
  'opened_complete',
  'loose',
]

export function parsePackagingGradeOptions(
  value: string | null | undefined
): string[] {
  if (!value) {
    return defaultInventoryPackagingGrades
  }

  try {
    const parsed = JSON.parse(value) as unknown
    if (Array.isArray(parsed)) {
      const normalized = normalizeDisplayOptions(
        parsed.filter((entry): entry is string => typeof entry === 'string')
      )
      return normalized.length > 0 ? normalized : defaultInventoryPackagingGrades
    }
  } catch {
    return defaultInventoryPackagingGrades
  }

  return defaultInventoryPackagingGrades
}

export function serializePackagingGradeOptions(values: string[]): string {
  return JSON.stringify(normalizeDisplayOptions(values))
}

export function normalizePackagingGradeName(value: string): string {
  return normalizeDisplayOption(value)
}
