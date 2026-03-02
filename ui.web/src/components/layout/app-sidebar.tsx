import { useEffect, useState } from 'react'
import { ArrowDown, ArrowUp, Eye, EyeOff, Pencil, X } from 'lucide-react'
import { useLayout } from '@/context/layout-provider'
import { useAuthStore } from '@/stores/auth-store'
import { useTranslation } from 'react-i18next'
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarRail,
} from '@/components/ui/sidebar'
// import { AppTitle } from './app-title'
import { sidebarData } from './data/sidebar-data'
import { type NavCollapsible, type NavItem } from './types'
import { NavGroup } from './nav-group'
import { NavUser } from './nav-user'
import { TeamSwitcher } from './team-switcher'

type NavPreference = {
  order: number
  hidden: boolean
}

const GLOBAL_NAV_PREFERENCES_KEY = 'cabinet.nav.preferences.global'

function navKeyForTitle(title: string): string {
  return title.trim().toLowerCase().replace(/\s+/g, '-')
}

export function AppSidebar() {
  const { collapsible, variant } = useLayout()
  const { t } = useTranslation('nav')
  const authUser = useAuthStore((state) => state.auth.user)
  const sidebarUser = authUser
    ? {
        name: authUser.accountNo || sidebarData.user.name,
        email: authUser.email || sidebarData.user.email,
        avatar: sidebarData.user.avatar,
      }
    : sidebarData.user

  const normalizeNavKey = (title: string) => navKeyForTitle(title)
  const [navEditMode, setNavEditMode] = useState(false)
  const [navEditOrder, setNavEditOrder] = useState<string[]>([])
  const [navPreferences, setNavPreferences] = useState<Record<string, NavPreference>>(() => {
    try {
      const raw = window.localStorage.getItem(GLOBAL_NAV_PREFERENCES_KEY)
      if (!raw) {
        return {}
      }
      const parsed = JSON.parse(raw) as Record<string, NavPreference>
      return parsed && typeof parsed === 'object' ? parsed : {}
    } catch {
      return {}
    }
  })
  const [runtimeMeta, setRuntimeMeta] = useState<{
    appVersion: string
    buildDate: string
  }>({
    appVersion: 'unknown',
    buildDate: 'unknown',
  })
  const [workspaceCollections, setWorkspaceCollections] = useState<string[]>([
    'All Items',
    'Watch List',
    'Wishlist Focus',
    'Store 1',
    'Store 2',
    'Warehouse 1',
  ])
  const [activeWorkspaceCollection, setActiveWorkspaceCollection] = useState('All Items')
  const [newCollectionName, setNewCollectionName] = useState('')
  const [collectionInputOpen, setCollectionInputOpen] = useState(false)

  useEffect(() => {
    let cancelled = false
    async function loadRuntimeMeta() {
      try {
        const resp = await fetch('/api/runtime')
        if (!resp.ok) {
          return
        }
        const payload = (await resp.json()) as {
          app_version?: string
          build_date?: string
        }
        if (cancelled) {
          return
        }
        setRuntimeMeta({
          appVersion: payload.app_version?.trim() || 'unknown',
          buildDate: payload.build_date?.trim() || 'unknown',
        })
      } catch {
        // Keep unknown fallback metadata when runtime endpoint is unavailable.
      }
    }
    void loadRuntimeMeta()
    return () => {
      cancelled = true
    }
  }, [])

  useEffect(() => {
    const profileScope = authUser?.email || authUser?.accountNo || 'local'
    const storageKey = `cabinet.nav.preferences.${profileScope}`
    try {
      const raw = window.localStorage.getItem(storageKey)
      if (!raw) {
        return
      }
      const parsed = JSON.parse(raw) as Record<string, NavPreference>
      if (parsed && typeof parsed === 'object') {
        setNavPreferences(parsed)
      }
    } catch {}
  }, [authUser?.accountNo, authUser?.email])

  useEffect(() => {
    const profileScope = authUser?.email || authUser?.accountNo || 'local'
    const listKey = `cabinet.workspace.collections.${profileScope}`
    const activeKey = `cabinet.workspace.collections.active.${profileScope}`
    try {
      const raw = window.localStorage.getItem(listKey)
      if (raw) {
        const parsed = JSON.parse(raw) as string[]
        if (Array.isArray(parsed) && parsed.length > 0) {
          const sanitized = parsed
            .map((entry) => entry.trim())
            .filter((entry) => entry !== '')
          if (sanitized.length > 0) {
            setWorkspaceCollections(sanitized)
          }
        }
      }
      const active = window.localStorage.getItem(activeKey)?.trim()
      if (active) {
        setActiveWorkspaceCollection(active)
      }
    } catch {
      // Fallback to in-memory defaults.
    }
  }, [authUser?.accountNo, authUser?.email])

  useEffect(() => {
    const profileScope = authUser?.email || authUser?.accountNo || 'local'
    const storageKey = `cabinet.nav.preferences.${profileScope}`
    window.localStorage.setItem(GLOBAL_NAV_PREFERENCES_KEY, JSON.stringify(navPreferences))
    window.localStorage.setItem(storageKey, JSON.stringify(navPreferences))
  }, [authUser?.accountNo, authUser?.email, navPreferences])

  useEffect(() => {
    const profileScope = authUser?.email || authUser?.accountNo || 'local'
    const listKey = `cabinet.workspace.collections.${profileScope}`
    const activeKey = `cabinet.workspace.collections.active.${profileScope}`
    window.localStorage.setItem(listKey, JSON.stringify(workspaceCollections))
    window.localStorage.setItem(activeKey, activeWorkspaceCollection)
  }, [authUser?.accountNo, authUser?.email, workspaceCollections, activeWorkspaceCollection])

  const primaryItems = sidebarData.navGroups[0]?.items ?? []
  const orderForPrimaryItem = (item: NavItem) => {
    const key = navKeyForTitle(item.title)
    const defaultOrder = primaryItems.findIndex((candidate) => candidate.title === item.title)
    return navPreferences[key]?.order ?? defaultOrder
  }

  const movePrimaryItem = (key: string, direction: 'up' | 'down') => {
    setNavEditOrder((current) => {
      const fromIndex = current.findIndex((itemKey) => itemKey === key)
      if (fromIndex < 0) {
        return current
      }

      const toIndex = direction === 'up' ? fromIndex - 1 : fromIndex + 1
      if (toIndex < 0 || toIndex >= current.length) {
        return current
      }

      const swapped = [...current]
      const [moved] = swapped.splice(fromIndex, 1)
      swapped.splice(toIndex, 0, moved)
      return swapped
    })
  }

  const togglePrimaryVisibility = (key: string) => {
    setNavPreferences((current) => ({
      ...current,
      [key]: {
        order:
          current[key]?.order ??
          primaryItems.findIndex((item) => navKeyForTitle(item.title) === key),
        hidden: !(current[key]?.hidden ?? false),
      },
    }))
  }

  const orderedPrimaryItems = [...primaryItems].sort(
    (left, right) => orderForPrimaryItem(left) - orderForPrimaryItem(right)
  )
  const orderedPrimaryKeys = orderedPrimaryItems.map((item) => navKeyForTitle(item.title))
  const primaryItemsByKey = new Map(
    primaryItems.map((item) => [navKeyForTitle(item.title), item] as const)
  )

  const configuredPrimaryItems = orderedPrimaryItems
    .filter((item) => {
      const key = navKeyForTitle(item.title)
      return !navPreferences[key]?.hidden
    })

  const translateItem = (item: NavItem): NavItem => {
    if ('items' in item) {
      const collapsible = item as NavCollapsible
      return {
        ...collapsible,
        title: t(`items.${normalizeNavKey(collapsible.title)}`, {
          defaultValue: collapsible.title,
        }),
        items: collapsible.items.map((nested) => ({
          ...nested,
          title: t(`items.${normalizeNavKey(nested.title)}`, {
            defaultValue: nested.title,
          }),
        })),
      }
    }

    return {
      ...item,
      title: t(`items.${normalizeNavKey(item.title)}`, {
        defaultValue: item.title,
      }),
    }
  }

  const sidebarGroupsWithPreferences = sidebarData.navGroups.map((group, index) => ({
    ...group,
    items: index === 0 ? configuredPrimaryItems : group.items,
  }))

  const translatedNavGroups = sidebarGroupsWithPreferences.map((group) => ({
    ...group,
    title: t(`groups.${normalizeNavKey(group.title)}`, { defaultValue: group.title }),
    items: group.items.map(translateItem),
  }))

  const saveNewCollection = () => {
    const trimmed = newCollectionName.trim()
    if (!trimmed) {
      return
    }
    const exists = workspaceCollections.some(
      (collection) => collection.toLowerCase() === trimmed.toLowerCase()
    )
    if (!exists) {
      setWorkspaceCollections((current) => [...current, trimmed])
    }
    setCollectionInputOpen(false)
    setNewCollectionName('')
  }

  return (
    <Sidebar collapsible={collapsible} variant={variant}>
      <SidebarHeader>
        <TeamSwitcher teams={sidebarData.teams} />

        {/* Replace <TeamSwitch /> with the following <AppTitle />
         /* if you want to use the normal app title instead of TeamSwitch dropdown */}
        {/* <AppTitle /> */}
      </SidebarHeader>
      <SidebarContent>
        <div
          className='mx-2 mt-1 space-y-2 rounded-md border p-2'
          data-testid='workspace-collections-panel'
        >
          <div className='flex items-center justify-between gap-2'>
            <h3
              className='text-xs font-semibold uppercase tracking-wide text-muted-foreground'
              data-testid='workspace-collections-heading'
            >
              Collections
            </h3>
            <button
              type='button'
              className='inline-flex h-7 items-center rounded-md border px-2 text-xs hover:bg-muted'
              data-testid='workspace-add-collection'
              onClick={() => setCollectionInputOpen((open) => !open)}
            >
              Add Collection
            </button>
          </div>
          {collectionInputOpen ? (
            <div className='flex items-center gap-2'>
              <input
                className='h-7 w-full rounded-md border bg-background px-2 text-xs'
                data-testid='workspace-new-collection-name'
                value={newCollectionName}
                onChange={(event) => setNewCollectionName(event.target.value)}
                placeholder='Collection name'
              />
              <button
                type='button'
                className='inline-flex h-7 items-center rounded-md border px-2 text-xs hover:bg-muted'
                data-testid='workspace-save-collection'
                onClick={saveNewCollection}
              >
                Save
              </button>
            </div>
          ) : null}
          <div className='space-y-1'>
            {workspaceCollections.map((collection) => {
              const key = collection.trim().toLowerCase().replace(/\s+/g, '-')
              const isActive = activeWorkspaceCollection === collection
              return (
                <button
                  key={collection}
                  type='button'
                  className='flex w-full items-center justify-start rounded-md border px-2 py-1 text-xs hover:bg-muted'
                  data-testid={`workspace-collection-item-${key}`}
                  data-state={isActive ? 'active' : 'inactive'}
                  onClick={() => setActiveWorkspaceCollection(collection)}
                >
                  {collection}
                </button>
              )
            })}
          </div>
        </div>
        {translatedNavGroups.map((props) => (
          <NavGroup key={props.title} {...props} />
        ))}
      </SidebarContent>
      <SidebarFooter>
        <div className='px-2 pb-2'>
          <button
            type='button'
            className='inline-flex h-8 w-8 items-center justify-center rounded-full border text-muted-foreground hover:bg-muted'
            data-testid='sidebar-nav-edit-toggle'
            onClick={() => {
              setNavEditMode((open) => {
                const nextOpen = !open
                if (nextOpen) {
                  setNavEditOrder(orderedPrimaryKeys)
                } else {
                  setNavPreferences((current) => {
                    const nextPreferences = { ...current }
                    navEditOrder.forEach((itemKey, index) => {
                      nextPreferences[itemKey] = {
                        order: index,
                        hidden: current[itemKey]?.hidden ?? false,
                      }
                    })
                    return nextPreferences
                  })
                }
                return nextOpen
              })
            }}
          >
            {navEditMode ? <X className='h-4 w-4' /> : <Pencil className='h-4 w-4' />}
          </button>
          {navEditMode ? (
            <div
              className='mt-2 space-y-1 rounded-md border p-2 text-xs'
              data-testid='sidebar-nav-edit-panel'
            >
              {navEditOrder.map((key) => {
                const item = primaryItemsByKey.get(key)
                if (!item) {
                  return null
                }
                const hidden = navPreferences[key]?.hidden ?? false
                return (
                  <div
                    key={key}
                    className='flex items-center justify-between gap-1'
                    data-testid={`sidebar-nav-edit-item-${key}`}
                  >
                    <span className={hidden ? 'opacity-50' : ''}>{item.title}</span>
                    <div className='flex items-center gap-1'>
                      <button
                        type='button'
                        data-testid={`sidebar-nav-move-up-${key}`}
                        className='rounded border p-1 hover:bg-muted'
                        onClick={() => movePrimaryItem(key, 'up')}
                      >
                        <ArrowUp className='h-3 w-3' />
                      </button>
                      <button
                        type='button'
                        data-testid={`sidebar-nav-move-down-${key}`}
                        className='rounded border p-1 hover:bg-muted'
                        onClick={() => movePrimaryItem(key, 'down')}
                      >
                        <ArrowDown className='h-3 w-3' />
                      </button>
                      <button
                        type='button'
                        data-testid={`sidebar-nav-visibility-${key}`}
                        className='rounded border p-1 hover:bg-muted'
                        onClick={() => togglePrimaryVisibility(key)}
                      >
                        {hidden ? (
                          <EyeOff className='h-3 w-3' />
                        ) : (
                          <Eye className='h-3 w-3' />
                        )}
                      </button>
                    </div>
                  </div>
                )
              })}
            </div>
          ) : null}
        </div>
        <NavUser user={sidebarUser} />
        <div
          className='px-2 pb-2 text-xs text-muted-foreground'
          data-testid='sidebar-runtime-meta'
        >
          <p>
            Version: <span data-testid='sidebar-app-version'>{runtimeMeta.appVersion}</span>
          </p>
          <p>
            Build Date: <span data-testid='sidebar-build-date'>{runtimeMeta.buildDate}</span>
          </p>
        </div>
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  )
}
