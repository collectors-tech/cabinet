import { useCallback, useMemo } from 'react'
import { useProfileSettings } from '@/features/settings/use-profile-settings'

export type WorkspaceCollectionSummary = {
  name: string
  key: string
  itemCount: number
  scopeLabel: string
  statusLabel: string
  updatedLabel: string
  description: string
  deletedAt: string | null
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
  collectionMetadata?: Record<string, WorkspaceCollectionMetadata>
}

export type WorkspaceCollectionMetadata = {
  scopeLabel?: string
  statusLabel?: string
  updatedLabel?: string
  description?: string
  deletedAt?: string | null
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
    deletedAt: null,
  },
  {
    name: 'Watch List',
    itemCount: 18,
    scopeLabel: 'Priority lane',
    statusLabel: 'Needs review',
    updatedLabel: 'Updated 12m ago',
    description: 'Fast-moving cards and sets needing quick review.',
    deletedAt: null,
  },
  {
    name: 'Warehouse 1',
    itemCount: 64,
    scopeLabel: 'Primary storage',
    statusLabel: 'Stable',
    updatedLabel: 'Updated 32m ago',
    description: 'Shelved long-box inventory in the main warehouse.',
    deletedAt: null,
  },
  {
    name: 'Store 1',
    itemCount: 27,
    scopeLabel: 'Retail lane',
    statusLabel: 'Ready to sell',
    updatedLabel: 'Updated 48m ago',
    description: 'Shopfront display stock prepared for live selling.',
    deletedAt: null,
  },
  {
    name: 'Store 2',
    itemCount: 19,
    scopeLabel: 'Retail lane',
    statusLabel: 'Ready to sell',
    updatedLabel: 'Updated 1h ago',
    description: 'Overflow retail stock staged for the second store.',
    deletedAt: null,
  },
  {
    name: 'Overflow',
    itemCount: 11,
    scopeLabel: 'Overflow storage',
    statusLabel: 'Needs sorting',
    updatedLabel: 'Updated 2h ago',
    description: 'Backlog boxes waiting for proper collection placement.',
    deletedAt: null,
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

function buildDefaultSummary(
  name: string,
  metadata?: WorkspaceCollectionMetadata
): WorkspaceCollectionSummary {
  const seed = DEFAULT_COLLECTIONS.find((entry) => entry.name === name)
  if (seed) {
    return {
      ...seed,
      key: collectionKey(seed.name),
      ...metadata,
      deletedAt: metadata?.deletedAt ?? null,
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
    ...metadata,
    deletedAt: metadata?.deletedAt ?? null,
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
    collectionMetadata: {},
  }
}

function loadingWorkspaceCollectionsState(): Required<PersistedWorkspaceCollectionsState> {
  return {
    collections: ['All Items'],
    activeCollection: 'All Items',
    items: [],
    collectionMetadata: {},
  }
}

function normalizeCollectionMetadata(
  value?: Record<string, WorkspaceCollectionMetadata>
): Record<string, WorkspaceCollectionMetadata> {
  const normalized: Record<string, WorkspaceCollectionMetadata> = {}
  Object.entries(value ?? {}).forEach(([rawKey, metadata]) => {
    const key = collectionKey(rawKey)
    if (!key || !metadata) {
      return
    }
    normalized[key] = {
      scopeLabel: normalizeCollectionName(metadata.scopeLabel ?? ''),
      statusLabel: normalizeCollectionName(metadata.statusLabel ?? ''),
      updatedLabel: normalizeCollectionName(metadata.updatedLabel ?? ''),
      description: normalizeCollectionName(metadata.description ?? ''),
      deletedAt:
        typeof metadata.deletedAt === 'string' && metadata.deletedAt.trim()
          ? metadata.deletedAt.trim()
          : null,
    }
  })
  return normalized
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
    const collectionMetadata = normalizeCollectionMetadata(
      parsed.collectionMetadata
    )
    const activeCollection =
      normalizedCollections.includes(normalizedActive) &&
      !collectionMetadata[collectionKey(normalizedActive)]?.deletedAt
        ? normalizedActive
        : 'All Items'

    return {
      collections: normalizedCollections,
      activeCollection,
      items: normalizedItems,
      collectionMetadata,
    }
  } catch {
    return defaultWorkspaceCollectionsState()
  }
}

function serializeWorkspaceCollectionsState(
  collections: string[],
  activeCollection: string,
  items: WorkspaceCollectionItem[],
  collectionMetadata: Record<string, WorkspaceCollectionMetadata>
): string {
  const normalizedCollections = normalizeCollectionsList(collections)
  const normalizedItems = items
    .map(normalizeWorkspaceCollectionItem)
    .filter((item) => item.id && item.name)
  const normalizedActive = normalizeCollectionName(activeCollection)
  const normalizedMetadata = normalizeCollectionMetadata(collectionMetadata)
  const activeMetadata = normalizedMetadata[collectionKey(normalizedActive)]

  return JSON.stringify({
    collections: normalizedCollections,
    activeCollection:
      normalizedCollections.includes(normalizedActive) &&
      !activeMetadata?.deletedAt
      ? normalizedActive
      : 'All Items',
    items: normalizedItems,
    collectionMetadata: normalizedMetadata,
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
  const collectionMetadata = persistedState.collectionMetadata
  const activeWorkspaceCollections = useMemo(
    () =>
      workspaceCollections.filter(
        (collection) => !collectionMetadata[collectionKey(collection)]?.deletedAt
      ),
    [collectionMetadata, workspaceCollections]
  )

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
        nextState.items,
        nextState.collectionMetadata
      ),
    })
  }

  const buildSummaries = useCallback(
    (collections: string[]) =>
      collections.map((name) => {
        const base = buildDefaultSummary(
          name,
          collectionMetadata[collectionKey(name)]
        )
        const assignedCount =
          name === 'All Items'
            ? workspaceItems.length
            : workspaceItems.filter((item) => item.collectionName === name)
                .length

        return {
          ...base,
          itemCount:
            name === 'All Items'
              ? Math.max(base.itemCount, assignedCount)
              : assignedCount,
          updatedLabel: base.deletedAt
            ? 'Deleted'
            : assignedCount > 0 && name !== 'All Items'
              ? 'Updated just now'
              : base.updatedLabel,
        }
      }),
    [collectionMetadata, workspaceItems]
  )

  const collectionSummaries = useMemo(
    () => buildSummaries(activeWorkspaceCollections),
    [activeWorkspaceCollections, buildSummaries]
  )
  const deletedCollectionSummaries = useMemo(
    () =>
      buildSummaries(
        workspaceCollections.filter(
          (collection) =>
            !!collectionMetadata[collectionKey(collection)]?.deletedAt
        )
      ),
    [buildSummaries, collectionMetadata, workspaceCollections]
  )

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
      collectionMetadata,
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

    const { [currentKey]: currentMetadata, ...remainingMetadata } =
      collectionMetadata

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
      collectionMetadata: {
        ...remainingMetadata,
        [nextKey]: currentMetadata ?? {},
      },
    })

    return normalizedNext
  }

  const updateCollectionMetadata = async (
    currentName: string,
    metadata: WorkspaceCollectionMetadata
  ): Promise<boolean> => {
    const normalizedCurrent = normalizeCollectionName(currentName)
    const currentKey = collectionKey(normalizedCurrent)
    if (!normalizedCurrent || normalizedCurrent === 'All Items') {
      return false
    }
    if (!workspaceCollections.some((collection) => collectionKey(collection) === currentKey)) {
      return false
    }

    await persistWorkspaceCollectionsState({
      collections: workspaceCollections,
      activeCollection: activeWorkspaceCollection,
      items: workspaceItems,
      collectionMetadata: {
        ...collectionMetadata,
        [currentKey]: {
          ...(collectionMetadata[currentKey] ?? {}),
          ...metadata,
        },
      },
    })
    return true
  }

  const updateCollectionDetails = async (
    currentName: string,
    nextName: string,
    metadata: WorkspaceCollectionMetadata
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

    const { [currentKey]: currentMetadata, ...remainingMetadata } =
      collectionMetadata
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
      collectionMetadata: {
        ...remainingMetadata,
        [nextKey]: {
          ...(currentMetadata ?? {}),
          ...metadata,
        },
      },
    })
    return normalizedNext
  }

  const removeCollection = async (
    name: string,
    destinationName?: string | null
  ): Promise<boolean> => {
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
    const normalizedDestination = normalizeCollectionName(destinationName ?? '')
    const validDestination =
      normalizedDestination &&
      normalizedDestination !== 'All Items' &&
      collectionKey(normalizedDestination) !== normalizedKey &&
      activeWorkspaceCollections.some(
        (collection) => collectionKey(collection) === collectionKey(normalizedDestination)
      )
        ? normalizedDestination
        : null

    await persistWorkspaceCollectionsState({
      collections: workspaceCollections,
      activeCollection:
        collectionKey(activeWorkspaceCollection) === normalizedKey
          ? 'All Items'
          : activeWorkspaceCollection,
      items: workspaceItems.map((item) =>
        item.collectionName &&
        collectionKey(item.collectionName) === normalizedKey
          ? { ...item, collectionName: validDestination }
          : item
      ),
      collectionMetadata: {
        ...collectionMetadata,
        [normalizedKey]: {
          ...(collectionMetadata[normalizedKey] ?? {}),
          deletedAt: new Date().toISOString(),
          statusLabel: 'Deleted',
        },
      },
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
      !activeWorkspaceCollections.some(
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
        collectionMetadata,
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
      !activeWorkspaceCollections.some(
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
      collectionMetadata,
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
      activeWorkspaceCollections.find(
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
      collectionMetadata,
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
      collectionMetadata,
    }).then(() => updatedItem)
  }

  const collectionItems = useMemo(() => workspaceItems, [workspaceItems])

  return {
    workspaceCollections: activeWorkspaceCollections,
    allWorkspaceCollections: workspaceCollections,
    activeWorkspaceCollection,
    setActiveWorkspaceCollection: async (nextCollection: string) => {
      const normalizedCollection = normalizeCollectionName(nextCollection)
      if (!normalizedCollection) {
        return
      }
      const safeCollection = activeWorkspaceCollections.includes(normalizedCollection)
        ? normalizedCollection
        : 'All Items'
      await persistWorkspaceCollectionsState({
        collections: workspaceCollections,
        activeCollection: safeCollection,
        items: workspaceItems,
        collectionMetadata,
      })
    },
    addCollection,
    renameCollection,
    updateCollectionMetadata,
    updateCollectionDetails,
    removeCollection,
    collectionSummaries,
    deletedCollectionSummaries,
    collectionItems,
    assignItemToCollection,
    assignWorkspaceItemToCollection,
    ensureWorkspaceCollectionAndAssignItem,
    unassignItemFromCollection,
  }
}
