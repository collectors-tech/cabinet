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

function normalizeCollectionName(value: string): string {
  return value.trim().replace(/\s+/g, ' ')
}

export function collectionKey(value: string): string {
  return normalizeCollectionName(value).toLowerCase().replace(/[^a-z0-9]+/g, '-')
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

  return {
    workspaceCollections,
    activeWorkspaceCollection,
    setActiveWorkspaceCollection,
    addCollection,
  }
}
