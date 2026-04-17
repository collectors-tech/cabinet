import { useEffect, useMemo, useState } from 'react'
import { useAuthStore } from '@/stores/auth-store'

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
    description: 'Custom workspace collection created from the management surface.',
  }
}

function profileScopeKey(profileID?: string | null, userEmail?: string | null): string {
  return normalizeCollectionName(profileID || userEmail || 'default-profile')
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
}

export function useWorkspaceCollections() {
  const userEmail = useAuthStore((state) => state.auth.user?.email)
  const profileScope = profileScopeKey(null, userEmail)
  const listKey = `cabinet.workspace.collections.${profileScope}`
  const activeKey = `cabinet.workspace.collections.active.${profileScope}`
  const itemsKey = `cabinet.workspace.collectionItems.${profileScope}`

  const [workspaceCollections, setWorkspaceCollections] = useState<string[]>(
    DEFAULT_COLLECTIONS.map((entry) => entry.name)
  )
  const [activeWorkspaceCollection, setActiveWorkspaceCollection] = useState('All Items')
  const [workspaceItems, setWorkspaceItems] = useState<WorkspaceCollectionItem[]>(DEFAULT_ITEMS)

  useEffect(() => {
    if (typeof window === 'undefined') {
      return
    }

    const persistedCollections = window.localStorage.getItem(listKey)
    if (persistedCollections) {
      try {
        const parsed = JSON.parse(persistedCollections) as string[]
        const normalized = parsed
          .map((entry) => normalizeCollectionName(entry))
          .filter(Boolean)
        if (normalized.length) {
          const unique = Array.from(new Set(['All Items', ...normalized]))
          setWorkspaceCollections(unique)
        }
      } catch {
        // Ignore malformed persisted collections and fall back to defaults.
      }
    } else {
      setWorkspaceCollections(DEFAULT_COLLECTIONS.map((entry) => entry.name))
    }

    const persistedActive = window.localStorage.getItem(activeKey)
    if (persistedActive) {
      const normalizedActive = normalizeCollectionName(persistedActive)
      if (normalizedActive) {
        setActiveWorkspaceCollection(normalizedActive)
      }
    } else {
      setActiveWorkspaceCollection('All Items')
    }

    const persistedItems = window.localStorage.getItem(itemsKey)
    if (persistedItems) {
      try {
        const parsed = JSON.parse(persistedItems) as WorkspaceCollectionItem[]
        if (Array.isArray(parsed) && parsed.length) {
          setWorkspaceItems(
            parsed.map((item) => ({
              ...item,
              name: normalizeCollectionName(item.name),
              detail: normalizeCollectionName(item.detail),
              collectionName: item.collectionName
                ? normalizeCollectionName(item.collectionName)
                : null,
            }))
          )
        }
      } catch {
        // Ignore malformed persisted items and fall back to defaults.
      }
    } else {
      setWorkspaceItems(DEFAULT_ITEMS)
    }
  }, [activeKey, itemsKey, listKey])

  useEffect(() => {
    if (typeof window === 'undefined') {
      return
    }
    window.localStorage.setItem(listKey, JSON.stringify(workspaceCollections))
  }, [listKey, workspaceCollections])

  useEffect(() => {
    if (typeof window === 'undefined') {
      return
    }
    window.localStorage.setItem(activeKey, activeWorkspaceCollection)
  }, [activeKey, activeWorkspaceCollection])

  useEffect(() => {
    if (typeof window === 'undefined') {
      return
    }
    window.localStorage.setItem(itemsKey, JSON.stringify(workspaceItems))
  }, [itemsKey, workspaceItems])

  const collectionSummaries = useMemo(() => {
    return workspaceCollections.map((name) => {
      const base = buildDefaultSummary(name)
      const assignedCount =
        name === 'All Items'
          ? workspaceItems.length
          : workspaceItems.filter((item) => item.collectionName === name).length

      return {
        ...base,
        itemCount: name === 'All Items' ? Math.max(base.itemCount, assignedCount) : assignedCount,
        updatedLabel:
          assignedCount > 0 && name !== 'All Items' ? 'Updated just now' : base.updatedLabel,
      }
    })
  }, [workspaceCollections, workspaceItems])

  const addCollection = (value: string): string | null => {
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
    setWorkspaceCollections((current) => [...current, normalized])
    setActiveWorkspaceCollection(normalized)
    return normalized
  }

  const renameCollection = (currentName: string, nextName: string): string | null => {
    const normalizedCurrent = normalizeCollectionName(currentName)
    const normalizedNext = normalizeCollectionName(nextName)
    if (!normalizedCurrent || !normalizedNext || normalizedCurrent === 'All Items') {
      return null
    }
    const exists = workspaceCollections.some(
      (collection) =>
        collection.toLowerCase() === normalizedNext.toLowerCase() &&
        collection.toLowerCase() !== normalizedCurrent.toLowerCase()
    )
    if (exists) {
      return null
    }

    setWorkspaceCollections((current) =>
      current.map((collection) =>
        collection.toLowerCase() === normalizedCurrent.toLowerCase()
          ? normalizedNext
          : collection
      )
    )
    setWorkspaceItems((current) =>
      current.map((item) =>
        item.collectionName?.toLowerCase() === normalizedCurrent.toLowerCase()
          ? { ...item, collectionName: normalizedNext }
          : item
      )
    )
    if (activeWorkspaceCollection.toLowerCase() === normalizedCurrent.toLowerCase()) {
      setActiveWorkspaceCollection(normalizedNext)
    }
    return normalizedNext
  }

  const removeCollection = (name: string): boolean => {
    const normalized = normalizeCollectionName(name)
    if (!normalized || normalized === 'All Items') {
      return false
    }
    const exists = workspaceCollections.some(
      (collection) => collection.toLowerCase() === normalized.toLowerCase()
    )
    if (!exists) {
      return false
    }

    setWorkspaceCollections((current) =>
      current.filter((collection) => collection.toLowerCase() !== normalized.toLowerCase())
    )
    setWorkspaceItems((current) =>
      current.map((item) =>
        item.collectionName?.toLowerCase() === normalized.toLowerCase()
          ? { ...item, collectionName: null }
          : item
      )
    )
    if (activeWorkspaceCollection.toLowerCase() === normalized.toLowerCase()) {
      setActiveWorkspaceCollection('All Items')
    }
    return true
  }

  const assignItemToCollection = (itemID: string, collectionName: string): WorkspaceCollectionItem | null => {
    const normalizedCollection = normalizeCollectionName(collectionName)
    if (!itemID || !normalizedCollection || normalizedCollection === 'All Items') {
      return null
    }
    if (!workspaceCollections.some((collection) => collection === normalizedCollection)) {
      return null
    }

    let updatedItem: WorkspaceCollectionItem | null = null
    setWorkspaceItems((current) =>
      current.map((item) => {
        if (item.id !== itemID) {
          return item
        }
        updatedItem = { ...item, collectionName: normalizedCollection }
        return updatedItem
      })
    )
    return updatedItem
  }

  const collectionItems = useMemo(() => workspaceItems, [workspaceItems])

  return {
    workspaceCollections,
    activeWorkspaceCollection,
    setActiveWorkspaceCollection,
    addCollection,
    renameCollection,
    removeCollection,
    collectionSummaries,
    collectionItems,
    assignItemToCollection,
  }
}
