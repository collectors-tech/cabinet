import { useEffect, useMemo, useState } from 'react'
import { useAuthStore } from '@/stores/auth-store'

const DEFAULT_COLLECTIONS = [
  'All Items',
  'Watch List',
  'Wishlist Focus',
  'Store 1',
  'Store 2',
  'Warehouse 1',
]

type CollectionSummary = {
  name: string
  key: string
  itemCount: number
  updatedLabel: string
  scopeLabel: string
  statusLabel: string
  description: string
}

function normalizeCollectionName(value: string): string {
  return value.trim().replace(/\s+/g, ' ')
}

export function collectionKey(value: string): string {
  return normalizeCollectionName(value).toLowerCase().replace(/[^a-z0-9]+/g, '-')
}

function deriveCollectionSummary(name: string, index: number): CollectionSummary {
  const normalized = normalizeCollectionName(name)
  const lower = normalized.toLowerCase()

  if (lower === 'all items') {
    return {
      name: normalized,
      key: collectionKey(normalized),
      itemCount: 248,
      updatedLabel: 'Updated 5m ago',
      scopeLabel: 'Workspace default',
      statusLabel: 'Broadest scope',
      description: 'Everything currently tracked in this workspace profile.',
    }
  }

  if (lower.includes('watch')) {
    return {
      name: normalized,
      key: collectionKey(normalized),
      itemCount: 34,
      updatedLabel: 'Updated 12m ago',
      scopeLabel: 'Monitoring set',
      statusLabel: 'Needs review',
      description: 'Fast-moving cards/items that benefit from repeated checking.',
    }
  }

  if (lower.includes('wishlist')) {
    return {
      name: normalized,
      key: collectionKey(normalized),
      itemCount: 18,
      updatedLabel: 'Updated 1h ago',
      scopeLabel: 'Intent shortlist',
      statusLabel: 'Prioritized',
      description: 'Targets staged for acquisition or tighter pricing decisions.',
    }
  }

  if (lower.includes('warehouse')) {
    return {
      name: normalized,
      key: collectionKey(normalized),
      itemCount: 126 + index * 3,
      updatedLabel: 'Updated yesterday',
      scopeLabel: 'Storage location',
      statusLabel: 'Archived storage',
      description: 'Longer-term storage grouping with slower movement and larger volume.',
    }
  }

  if (lower.includes('store')) {
    return {
      name: normalized,
      key: collectionKey(normalized),
      itemCount: 42 + index * 5,
      updatedLabel: 'Updated 20m ago',
      scopeLabel: 'Retail lane',
      statusLabel: 'Active rotation',
      description: 'Operational selling/stock lane with frequent active-context switching.',
    }
  }

  return {
    name: normalized,
    key: collectionKey(normalized),
    itemCount: 12 + index * 2,
    updatedLabel: 'Updated recently',
    scopeLabel: 'Custom collection',
    statusLabel: 'User managed',
    description: 'Custom workspace collection created for a focused subset of items.',
  }
}

export function useWorkspaceCollections() {
  const authUser = useAuthStore((state) => state.auth.user)
  const profileScope = useMemo(
    () => authUser?.email || authUser?.accountNo || 'local',
    [authUser?.accountNo, authUser?.email]
  )
  const [workspaceCollections, setWorkspaceCollections] =
    useState<string[]>(DEFAULT_COLLECTIONS)
  const [activeWorkspaceCollection, setActiveWorkspaceCollection] =
    useState('All Items')

  useEffect(() => {
    const listKey = `cabinet.workspace.collections.${profileScope}`
    const activeKey = `cabinet.workspace.collections.active.${profileScope}`
    try {
      const raw = window.localStorage.getItem(listKey)
      if (raw) {
        const parsed = JSON.parse(raw) as string[]
        if (Array.isArray(parsed) && parsed.length > 0) {
          const sanitized = parsed
            .map((entry) => normalizeCollectionName(entry))
            .filter((entry) => entry !== '')
          if (sanitized.length > 0) {
            setWorkspaceCollections(sanitized)
          }
        }
      }
      const active = normalizeCollectionName(
        window.localStorage.getItem(activeKey) || ''
      )
      if (active) {
        setActiveWorkspaceCollection(active)
      }
    } catch {
      setWorkspaceCollections(DEFAULT_COLLECTIONS)
      setActiveWorkspaceCollection('All Items')
    }
  }, [profileScope])

  useEffect(() => {
    const listKey = `cabinet.workspace.collections.${profileScope}`
    const activeKey = `cabinet.workspace.collections.active.${profileScope}`
    window.localStorage.setItem(listKey, JSON.stringify(workspaceCollections))
    window.localStorage.setItem(activeKey, activeWorkspaceCollection)
  }, [profileScope, workspaceCollections, activeWorkspaceCollection])

  const addCollection = (rawName: string) => {
    const trimmed = normalizeCollectionName(rawName)
    if (!trimmed) {
      return null
    }
    const existing = workspaceCollections.find(
      (value) => value.toLowerCase() === trimmed.toLowerCase()
    )
    if (existing) {
      setActiveWorkspaceCollection(existing)
      return existing
    }
    setWorkspaceCollections((current) => [...current, trimmed])
    setActiveWorkspaceCollection(trimmed)
    return trimmed
  }

  const renameCollection = (currentName: string, nextName: string) => {
    const normalizedCurrent = normalizeCollectionName(currentName)
    const normalizedNext = normalizeCollectionName(nextName)
    if (!normalizedCurrent || !normalizedNext) {
      return null
    }
    const target = workspaceCollections.find((value) => value === normalizedCurrent)
    if (!target || normalizedCurrent === 'All Items') {
      return null
    }
    const existing = workspaceCollections.find(
      (value) => value.toLowerCase() === normalizedNext.toLowerCase()
    )
    if (existing && existing !== normalizedCurrent) {
      return existing
    }
    setWorkspaceCollections((current) =>
      current.map((value) => (value === normalizedCurrent ? normalizedNext : value))
    )
    if (activeWorkspaceCollection === normalizedCurrent) {
      setActiveWorkspaceCollection(normalizedNext)
    }
    return normalizedNext
  }

  const removeCollection = (name: string) => {
    const normalized = normalizeCollectionName(name)
    if (!normalized || normalized === 'All Items') {
      return false
    }
    const exists = workspaceCollections.includes(normalized)
    if (!exists) {
      return false
    }
    setWorkspaceCollections((current) => current.filter((value) => value !== normalized))
    if (activeWorkspaceCollection === normalized) {
      setActiveWorkspaceCollection('All Items')
    }
    return true
  }

  const collectionSummaries = useMemo(
    () => workspaceCollections.map((name, index) => deriveCollectionSummary(name, index)),
    [workspaceCollections]
  )

  return {
    workspaceCollections,
    collectionSummaries,
    activeWorkspaceCollection,
    setActiveWorkspaceCollection,
    addCollection,
    renameCollection,
    removeCollection,
  }
}
