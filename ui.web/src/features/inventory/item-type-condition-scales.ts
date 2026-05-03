export const inventoryItemTypeConditionScalesSettingsKey =
  'grading.enums.item_type_condition_scales'

export type InventoryItemTypeConditionScale = {
  item_type: string
  conditions: string[]
}

export const defaultInventoryItemTypeConditionScales: InventoryItemTypeConditionScale[] =
  [
    {
      item_type: 'Slot Cars',
      conditions: [
        '10+ - New, in packaging',
        '10 - New, with packaging separate',
        '9 - New, no packaging',
        '8 - Like new',
        '7 - Minor track-wear',
        '6 - Bumper-wear & scratches',
        '5 - Worn, with scratches & nicks',
        '4 - Cut wheel wells, but nice',
        '3 - Cut badly, but good runner',
        '2 - Good for parts only',
        '1 - Good for nothing',
      ],
    },
    {
      item_type: 'Trading Cards',
      conditions: [
        'Mint (M)',
        'Near Mint (NM)',
        'Excellent (EX)',
        'Good (GD)',
        'Light Played (LP)',
        'Played (PL)',
        'Poor (PO)',
      ],
    },
  ]

export function normalizeDisplayOption(value: string): string {
  return value.trim().replace(/\s+/g, ' ')
}

export function normalizeDisplayOptions(values: string[]): string[] {
  const seen = new Set<string>()
  return values
    .map(normalizeDisplayOption)
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

export function normalizeItemTypeConditionScales(
  values: InventoryItemTypeConditionScale[]
): InventoryItemTypeConditionScale[] {
  const seen = new Set<string>()
  const normalized = values
    .map((scale) => ({
      item_type: normalizeDisplayOption(scale.item_type),
      conditions: normalizeDisplayOptions(scale.conditions ?? []),
    }))
    .filter((scale) => {
      if (scale.item_type === '' || scale.conditions.length === 0) {
        return false
      }
      const key = scale.item_type.toLowerCase()
      if (seen.has(key)) {
        return false
      }
      seen.add(key)
      return true
    })

  return normalized.length > 0
    ? normalized
    : defaultInventoryItemTypeConditionScales
}

export function parseItemTypeConditionScales(
  value: string | null | undefined
): InventoryItemTypeConditionScale[] {
  if (!value) {
    return defaultInventoryItemTypeConditionScales
  }

  try {
    const parsed = JSON.parse(value) as unknown
    if (Array.isArray(parsed)) {
      return normalizeItemTypeConditionScales(
        parsed
          .map((entry) => {
            if (!entry || typeof entry !== 'object') {
              return null
            }
            const candidate = entry as {
              item_type?: unknown
              conditions?: unknown
            }
            return {
              item_type:
                typeof candidate.item_type === 'string'
                  ? candidate.item_type
                  : '',
              conditions: Array.isArray(candidate.conditions)
                ? candidate.conditions.filter(
                    (condition): condition is string =>
                      typeof condition === 'string'
                  )
                : [],
            }
          })
          .filter(
            (entry): entry is InventoryItemTypeConditionScale => entry !== null
          )
      )
    }
  } catch {
    return defaultInventoryItemTypeConditionScales
  }

  return defaultInventoryItemTypeConditionScales
}

export function serializeItemTypeConditionScales(
  values: InventoryItemTypeConditionScale[]
): string {
  return JSON.stringify(normalizeItemTypeConditionScales(values))
}

export function itemTypeOptions(
  values: InventoryItemTypeConditionScale[]
): string[] {
  return normalizeItemTypeConditionScales(values).map((scale) => scale.item_type)
}

export function conditionsForItemType(
  values: InventoryItemTypeConditionScale[],
  itemType: string
): string[] {
  const normalizedType = normalizeDisplayOption(itemType).toLowerCase()
  const scales = normalizeItemTypeConditionScales(values)
  const matched =
    scales.find((scale) => scale.item_type.toLowerCase() === normalizedType) ??
    scales[0]
  return matched?.conditions ?? []
}

export function inferItemTypeFromCategory(category: string): string {
  const normalized = category.toLowerCase()
  if (normalized.includes('trading') || normalized.includes('card')) {
    return 'Trading Cards'
  }
  if (normalized.includes('slot') || normalized.includes('car')) {
    return 'Slot Cars'
  }
  return defaultInventoryItemTypeConditionScales[0]?.item_type ?? 'Slot Cars'
}
