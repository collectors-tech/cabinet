export const inventoryCategoryOptionsSettingsKey =
  'inventory.category-options.v1'

export const defaultInventoryCategoryOptions = [
  'General',
  'Cars',
  'Diecast',
  'Slot Car',
  'Trading Card',
  'Action Figure',
  'Comic',
  'Model Kit',
]

export function normalizeCategoryName(value: string): string {
  return value.trim().replace(/\s+/g, ' ')
}

export function normalizeCategoryOptions(values: string[]): string[] {
  const seen = new Set<string>()
  return values
    .map(normalizeCategoryName)
    .filter((value) => {
      if (value === '') {
        return false
      }
      const key = value.toLowerCase()
      if (seen.has(key)) {
        return false
      }
      seen.add(key)
      return true
    })
}

export function parseCategoryOptions(value: string | null | undefined): string[] {
  if (!value) {
    return defaultInventoryCategoryOptions
  }

  try {
    const parsed = JSON.parse(value) as unknown
    if (Array.isArray(parsed)) {
      const parsedOptions = normalizeCategoryOptions(
        parsed.filter((entry): entry is string => typeof entry === 'string')
      )
      return parsedOptions.length > 0
        ? parsedOptions
        : defaultInventoryCategoryOptions
    }
  } catch {
    return defaultInventoryCategoryOptions
  }

  return defaultInventoryCategoryOptions
}

export function serializeCategoryOptions(values: string[]): string {
  return JSON.stringify(normalizeCategoryOptions(values))
}

export function splitCategoryValue(value: string): string[] {
  return normalizeCategoryOptions(value.split(','))
}

export function joinCategoryValue(values: string[]): string {
  return normalizeCategoryOptions(values).join(', ')
}
