import { useMemo } from 'react'
import { useProfileSettings } from '@/features/settings/use-profile-settings'

export type WorkspaceCollectionSummary = {
  name: string
  key: string
  itemCount: number
  scopeLabel: string
  statusLabel: string
  updatedLabel: string
  description: string
}

export type WorkspaceCollectionItem = {
  id: string
  name: string
  detail: string
  collectionName: string | null
}

type CollectionSeed = Omit<WorkspaceCollectionSummary, 'key'>

type PersistedWorkspaceCollectionsState = {
  collections?: string[]
  activeCollection?: string
  items?: WorkspaceCollectionItem[]
}

const collectionsSettingsKey = 'collections.workspace.v1'

const DEFAULT_COLLECTIONS: CollectionSeed[] = [
  {
    name: 'All Items',
    itemCount: 248,
    scopeLabel: 'Workspace default',
    statusLabel: 'Broadest scope',
    updatedLabel: 'Updated 5m ago',
    description: 'Everything currently tracked in Cabinet.',
  },
  {
    name: 'Watch List',
    itemCount: 18,
    scopeLabel: 'Priority lane',
    statusLabel: 'Needs review',
    updatedLabel: 'Updated 12m ago',
    description: 'Fast-moving cards and sets needing quick review.',
  },
  {
    name: 'Warehouse 1',
    itemCount: 64,
    scopeLabel: 'Primary storage',
    statusLabel: 'Stable',
    updatedLabel: 'Updated 32m ago',
    description: 'Shelved long-box inventory in the main warehouse.',
  },
  {
    name: 'Store 1',
    itemCount: 27,
    scopeLabel: 'Retail lane',
    statusLabel: 'Ready to sell',
    updatedLabel: 'Updated 48m ago',
    description: 'Shopfront display stock prepared for live selling.',
  },
  {
    name: 'Store 2',
    itemCount: 19,
    scopeLabel: 'Retail lane',
    statusLabel: 'Ready to sell',
    updatedLabel: 'Updated 1h ago',
    description: 'Overflow retail stock staged for the second store.',
  },
  {
    name: 'Overflow',
    itemCount: 11,
    scopeLabel: 'Overflow storage',
    statusLabel: 'Needs sorting',
    updatedLabel: 'Updated 2h ago',
    description: 'Backlog boxes waiting for proper collection placement.',
  },
]

const DEFAULT_ITEMS: WorkspaceCollectionItem[] = [
  {
    id: 'inventory-item-kobe-rookie',
    name: '1996 Topps Kobe Bryant rookie',
    detail: 'PSA candidate, Lakers lot',
    collectionName: 'Watch List',
  },
  {
    id: 'inventory-item-jordan-beam-team',
    name: '1992 Beam Team Michael Jordan',
    detail: 'Premium slab review queue',
    collectionName: 'Warehouse 1',
  },
  {
    id: 'inventory-item-pikachu-shadowless',
    name: 'Shadowless Pikachu',
    detail: 'Binder copy ready for retail',
    collectionName: 'Store 1',
  },
  {
    id: 'inventory-item-charizard-base',
    name: 'Base Set Charizard',
    detail: 'High-value vault candidate',
    collectionName: null,
  },
  {
    id: 'inventory-item-metal-raiders-pack',
    name: 'Yu-Gi-Oh Metal Raiders blister',
    detail: 'Sealed product awaiting assignment',
    collectionName: null,
  },
]

function normalizeCollectionName(value: string): string {
  return value.replace(/\s+/g, ' ').trim()
}

export function collectionKey(value: string): string {
  return normalizeCollectionName(value)
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
}

function buildDefaultSummary(name: string): WorkspaceCollectionSummary {
  const seed = DEFAULT_COLLECTIONS.find((entry) => entry.name === name)
  if (seed) {
    return {
      ...seed,
      key: collectionKey(seed.name),
    }
  }

  const normalized = normalizeCollectionName(name)
  return {
    name: normalized,
    key: collectionKey(normalized),
    itemCount: 0,
    scopeLabel: 'Custom collection',
    statusLabel: 'Active',
    updatedLabel: 'Updated just now',
    description:
      'Custom workspace collection created from the management surface.',
  }
}

function normalizeWorkspaceCollectionItem(
  item: WorkspaceCollectionItem
): WorkspaceCollectionItem {
  return {
    id: normalizeCollectionName(item.id),
    name: normalizeCollectionName(item.name),
    detail: normalizeCollectionName(item.detail),
    collectionName: item.collectionName
      ? normalizeCollectionName(item.collectionName)
      : null,
  }
}

function normalizeCollectionsList(value?: string[]): string[] {
  const normalized = (value ?? [])
    .map((entry) => normalizeCollectionName(entry))
    .filter(Boolean)

  return Array.from(new Set(['All Items', ...normalized]))
}

function defaultWorkspaceCollectionsState(): Required<PersistedWorkspaceCollectionsState> {
  return {
    collections: DEFAULT_COLLECTIONS.map((entry) => entry.name),
    activeCollection: 'All Items',
    items: DEFAULT_ITEMS.map(normalizeWorkspaceCollectionItem),
  }
}

function loadingWorkspaceCollectionsState(): Required<PersistedWorkspaceCollectionsState> {
  return {
    collections: ['All Items'],
    activeCollection: 'All Items',
    items: [],
  }
}

function parseWorkspaceCollectionsState(
  rawValue: string | undefined
): Required<PersistedWorkspaceCollectionsState> {
  if (!rawValue) {
    return defaultWorkspaceCollectionsState()
  }

  try {
    const parsed = JSON.parse(rawValue) as PersistedWorkspaceCollectionsState
    const normalizedCollections = normalizeCollectionsList(parsed.collections)
    const normalizedItems = Array.isArray(parsed.items)
      ? parsed.items
          .map(normalizeWorkspaceCollectionItem)
          .filter((item) => item.id && item.name)
      : DEFAULT_ITEMS.map(normalizeWorkspaceCollectionItem)

    const normalizedActive = normalizeCollectionName(
      parsed.activeCollection ?? 'All Items'
    )
    const activeCollection = normalizedCollections.includes(normalizedActive)
      ? normalizedActive
      : 'All Items'

    return {
      collections: normalizedCollections,
      activeCollection,
      items: normalizedItems,
    }
  } catch {
    return defaultWorkspaceCollectionsState()
  }
}

function serializeWorkspaceCollectionsState(
  collections: string[],
  activeCollection: string,
  items: WorkspaceCollectionItem[]
): string {
  const normalizedCollections = normalizeCollectionsList(collections)
  const normalizedItems = items
    .map(normalizeWorkspaceCollectionItem)
    .filter((item) => item.id && item.name)
  const normalizedActive = normalizeCollectionName(activeCollection)

  return JSON.stringify({
    collections: normalizedCollections,
    activeCollection: normalizedCollections.includes(normalizedActive)
      ? normalizedActive
      : 'All Items',
    items: normalizedItems,
  })
}

export function useWorkspaceCollections() {
  const { activeProfileId, loading, settings, saveSettings } =
    useProfileSettings()
  const persistedState = useMemo(
    () =>
      loading
        ? loadingWorkspaceCollectionsState()
        : parseWorkspaceCollectionsState(settings[collectionsSettingsKey]),
    [loading, settings]
  )

  const workspaceCollections = persistedState.collections
  const activeWorkspaceCollection = persistedState.activeCollection
  const workspaceItems = persistedState.items

  const persistWorkspaceCollectionsState = async (
    nextState: Required<PersistedWorkspaceCollectionsState>
  ) => {
    if (!activeProfileId) {
      throw new Error('active_profile_missing')
    }

    await saveSettings({
      ...settings,
      [collectionsSettingsKey]: serializeWorkspaceCollectionsState(
        nextState.collections,
        nextState.activeCollection,
        nextState.items
      ),
    })
  }

  const collectionSummaries = useMemo(() => {
    return workspaceCollections.map((name) => {
      const base = buildDefaultSummary(name)
      const assignedCount =
        name === 'All Items'
          ? workspaceItems.length
          : workspaceItems.filter((item) => item.collectionName === name).length

      return {
        ...base,
        itemCount:
          name === 'All Items'
            ? Math.max(base.itemCount, assignedCount)
            : assignedCount,
        updatedLabel:
          assignedCount > 0 && name !== 'All Items'
            ? 'Updated just now'
            : base.updatedLabel,
      }
    })
  }, [workspaceCollections, workspaceItems])

  const addCollection = async (value: string): Promise<string | null> => {
    const normalized = normalizeCollectionName(value)
    if (!normalized) {
      return null
    }
    const exists = workspaceCollections.some(
      (collection) => collection.toLowerCase() === normalized.toLowerCase()
    )
    if (exists) {
      return null
    }

    await persistWorkspaceCollectionsState({
      collections: [...workspaceCollections, normalized],
      activeCollection: normalized,
      items: workspaceItems,
    })

    return normalized
  }

  const renameCollection = async (
    currentName: string,
    nextName: string
  ): Promise<string | null> => {
    const normalizedCurrent = normalizeCollectionName(currentName)
    const normalizedNext = normalizeCollectionName(nextName)
    const currentKey = collectionKey(normalizedCurrent)
    const nextKey = collectionKey(normalizedNext)
    if (
      !normalizedCurrent ||
      !normalizedNext ||
      normalizedCurrent === 'All Items'
    ) {
      return null
    }
    const exists = workspaceCollections.some(
      (collection) =>
        collectionKey(collection) === nextKey &&
        collectionKey(collection) !== currentKey
    )
    if (exists) {
      return null
    }

    await persistWorkspaceCollectionsState({
      collections: workspaceCollections.map((collection) =>
        collectionKey(collection) === currentKey ? normalizedNext : collection
      ),
      activeCollection:
        collectionKey(activeWorkspaceCollection) === currentKey
          ? normalizedNext
          : activeWorkspaceCollection,
      items: workspaceItems.map((item) =>
        item.collectionName && collectionKey(item.collectionName) === currentKey
          ? { ...item, collectionName: normalizedNext }
          : item
      ),
    })

    return normalizedNext
  }

  const removeCollection = async (name: string): Promise<boolean> => {
    const normalized = normalizeCollectionName(name)
    const normalizedKey = collectionKey(normalized)
    if (!normalized || normalized === 'All Items') {
      return false
    }
    const exists = workspaceCollections.some(
      (collection) => collectionKey(collection) === normalizedKey
    )
    if (!exists) {
      return false
    }

    await persistWorkspaceCollectionsState({
      collections: workspaceCollections.filter(
        (collection) => collectionKey(collection) !== normalizedKey
      ),
      activeCollection:
        collectionKey(activeWorkspaceCollection) === normalizedKey
          ? 'All Items'
          : activeWorkspaceCollection,
      items: workspaceItems.map((item) =>
        item.collectionName &&
        collectionKey(item.collectionName) === normalizedKey
          ? { ...item, collectionName: null }
          : item
      ),
    })

    return true
  }

  const assignItemToCollection = (
    itemID: string,
    collectionName: string
  ): Promise<WorkspaceCollectionItem | null> => {
    const normalizedCollection = normalizeCollectionName(collectionName)
    if (
      !itemID ||
      !normalizedCollection ||
      normalizedCollection === 'All Items'
    ) {
      return Promise.resolve(null)
    }
    if (
      !workspaceCollections.some(
        (collection) => collection === normalizedCollection
      )
    ) {
      return Promise.resolve(null)
    }

    const updatedItems = workspaceItems.map((item) =>
      item.id === itemID
        ? { ...item, collectionName: normalizedCollection }
        : item
    )
    const updatedItem = updatedItems.find((item) => item.id === itemID) ?? null

    if (!updatedItem) {
      return Promise.resolve(null)
    }

    return persistWorkspaceCollectionsState({
      collections: workspaceCollections,
      activeCollection: activeWorkspaceCollection,
      items: updatedItems,
    }).then(() => updatedItem)
  }

  const assignWorkspaceItemToCollection = (
    item: WorkspaceCollectionItem,
    collectionName: string
  ): Promise<WorkspaceCollectionItem | null> => {
    const normalizedCollection = normalizeCollectionName(collectionName)
    if (
      !item.id ||
      !normalizedCollection ||
      normalizedCollection === 'All Items'
    ) {
      return Promise.resolve(null)
    }
    if (
      !workspaceCollections.some(
        (collection) => collection === normalizedCollection
      )
    ) {
      return Promise.resolve(null)
    }

    const normalizedItem = normalizeWorkspaceCollectionItem({
      ...item,
      collectionName: normalizedCollection,
    })
    const existingItem = workspaceItems.some(
      (workspaceItem) => workspaceItem.id === normalizedItem.id
    )
    const updatedItems = existingItem
      ? workspaceItems.map((workspaceItem) =>
          workspaceItem.id === normalizedItem.id
            ? normalizedItem
            : workspaceItem
        )
      : [...workspaceItems, normalizedItem]

    return persistWorkspaceCollectionsState({
      collections: workspaceCollections,
      activeCollection: activeWorkspaceCollection,
      items: updatedItems,
    }).then(() => normalizedItem)
  }

  const ensureWorkspaceCollectionAndAssignItem = (
    item: WorkspaceCollectionItem,
    collectionName: string
  ): Promise<WorkspaceCollectionItem | null> => {
    const normalizedCollection = normalizeCollectionName(collectionName)
    if (
      !item.id ||
      !normalizedCollection ||
      normalizedCollection === 'All Items'
    ) {
      return Promise.resolve(null)
    }

    const existingCollection =
      workspaceCollections.find(
        (collection) =>
          collection.toLowerCase() === normalizedCollection.toLowerCase()
      ) ?? null
    const targetCollection = existingCollection ?? normalizedCollection
    const nextCollections = existingCollection
      ? workspaceCollections
      : [...workspaceCollections, targetCollection]
    const normalizedItem = normalizeWorkspaceCollectionItem({
      ...item,
      collectionName: targetCollection,
    })
    const existingItem = workspaceItems.some(
      (workspaceItem) => workspaceItem.id === normalizedItem.id
    )
    const updatedItems = existingItem
      ? workspaceItems.map((workspaceItem) =>
          workspaceItem.id === normalizedItem.id
            ? normalizedItem
            : workspaceItem
        )
      : [...workspaceItems, normalizedItem]

    return persistWorkspaceCollectionsState({
      collections: nextCollections,
      activeCollection: targetCollection,
      items: updatedItems,
    }).then(() => normalizedItem)
  }

  const unassignItemFromCollection = (
    itemID: string
  ): Promise<WorkspaceCollectionItem | null> => {
    if (!itemID) {
      return Promise.resolve(null)
    }

    const updatedItems = workspaceItems.map((item) =>
      item.id === itemID ? { ...item, collectionName: null } : item
    )
    const updatedItem = updatedItems.find((item) => item.id === itemID) ?? null

    if (!updatedItem) {
      return Promise.resolve(null)
    }

    return persistWorkspaceCollectionsState({
      collections: workspaceCollections,
      activeCollection: activeWorkspaceCollection,
      items: updatedItems,
    }).then(() => updatedItem)
  }

  const collectionItems = useMemo(() => workspaceItems, [workspaceItems])

  return {
    workspaceCollections,
    activeWorkspaceCollection,
    setActiveWorkspaceCollection: async (nextCollection: string) => {
      const normalizedCollection = normalizeCollectionName(nextCollection)
      if (!normalizedCollection) {
        return
      }
      const safeCollection = workspaceCollections.includes(normalizedCollection)
        ? normalizedCollection
        : 'All Items'
      await persistWorkspaceCollectionsState({
        collections: workspaceCollections,
        activeCollection: safeCollection,
        items: workspaceItems,
      })
    },
    addCollection,
    renameCollection,
    removeCollection,
    collectionSummaries,
    collectionItems,
    assignItemToCollection,
    assignWorkspaceItemToCollection,
    ensureWorkspaceCollectionAndAssignItem,
    unassignItemFromCollection,
  }
}
