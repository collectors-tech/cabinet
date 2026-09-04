import { useEffect, useState, type DragEvent } from 'react'
import { useLocation } from '@tanstack/react-router'
import {
  ArrowDown,
  ArrowUp,
  Bell,
  Eye,
  EyeOff,
  GripVertical,
  MessageSquare,
  MoreHorizontal,
  PanelLeft,
  SearchIcon,
  Settings,
  SlidersHorizontal,
} from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { useAuthStore } from '@/stores/auth-store'
import { useLayout } from '@/context/layout-provider'
import { useShellWorkspace } from '@/context/shell-workspace-context'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarRail,
  useSidebar,
} from '@/components/ui/sidebar'
import { AssistantWorkspacePanel } from './assistant-workspace-panel'
// import { AppTitle } from './app-title'
import { sidebarData } from './data/sidebar-data'
import { InboxWorkspacePanel } from './inbox-workspace-panel'
import { NavGroup } from './nav-group'
import { NavUser } from './nav-user'
import { SearchWorkspacePanel } from './search-workspace-panel'
import { TeamSwitcher } from './team-switcher'
import { type NavCollapsible, type NavItem } from './types'

type NavPreference = {
  order: number
  hidden: boolean
}

const GLOBAL_NAV_PREFERENCES_KEY = 'cabinet.nav.preferences.global'

function navKeyForTitle(title: string): string {
  return title.trim().toLowerCase().replace(/\s+/g, '-')
}

function moveKeyToIndex(order: string[], key: string, targetIndex: number) {
  const fromIndex = order.findIndex((itemKey) => itemKey === key)
  if (fromIndex < 0) {
    return order
  }

  const clampedTargetIndex = Math.max(
    0,
    Math.min(targetIndex, order.length - 1)
  )
  if (fromIndex === clampedTargetIndex) {
    return order
  }

  const next = [...order]
  const [moved] = next.splice(fromIndex, 1)
  next.splice(clampedTargetIndex, 0, moved)
  return next
}

export function AppSidebar() {
  const location = useLocation()
  const { collapsible, variant } = useLayout()
  const { state: sidebarState, isMobile, setOpen } = useSidebar()
  const { t } = useTranslation('nav')
  const { activeWorkspace, setActiveWorkspace } = useShellWorkspace()
  const isCollapsedSidebar = sidebarState === 'collapsed' && !isMobile
  const authUser = useAuthStore((state) => state.auth.user)

  useEffect(() => {
    if (activeWorkspace !== 'navigation') {
      setOpen(true)
    }
  }, [activeWorkspace, setOpen])

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
  const [navDraftPreferences, setNavDraftPreferences] = useState<
    Record<string, NavPreference>
  >({})
  const [draggedNavKey, setDraggedNavKey] = useState<string | null>(null)
  const [dragOverNavTarget, setDragOverNavTarget] = useState<{
    key: string
    position: 'before' | 'after'
  } | null>(null)
  const [navPreferences, setNavPreferences] = useState<
    Record<string, NavPreference>
  >(() => {
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
    const syncPreferences = () => {
      try {
        const raw = window.localStorage.getItem(storageKey)
        if (!raw) {
          setNavPreferences({})
          return
        }
        const parsed = JSON.parse(raw) as Record<string, NavPreference>
        setNavPreferences(parsed && typeof parsed === 'object' ? parsed : {})
      } catch {
        setNavPreferences({})
      }
    }

    const timeoutId = window.setTimeout(syncPreferences, 0)
    return () => {
      window.clearTimeout(timeoutId)
    }
  }, [authUser?.accountNo, authUser?.email])

  useEffect(() => {
    const profileScope = authUser?.email || authUser?.accountNo || 'local'
    const storageKey = `cabinet.nav.preferences.${profileScope}`
    window.localStorage.setItem(
      GLOBAL_NAV_PREFERENCES_KEY,
      JSON.stringify(navPreferences)
    )
    window.localStorage.setItem(storageKey, JSON.stringify(navPreferences))
  }, [authUser?.accountNo, authUser?.email, navPreferences])

  const primaryItems = sidebarData.navGroups[0]?.items ?? []
  const orderForPrimaryItem = (item: NavItem) => {
    const key = navKeyForTitle(item.title)
    const defaultOrder = primaryItems.findIndex(
      (candidate) => candidate.title === item.title
    )
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

      return moveKeyToIndex(current, key, toIndex)
    })
  }

  const openNavEditor = () => {
    clearNavDragState()
    setOpen(true)
    setActiveWorkspace('navigation')
    setNavDraftPreferences(navPreferences)
    setNavEditOrder(orderedPrimaryKeys)
    setNavEditMode(true)
  }

  const closeNavEditor = () => {
    clearNavDragState()
    setNavEditMode(false)
    setNavDraftPreferences({})
    setNavEditOrder([])
  }

  const applyNavEditor = () => {
    setNavPreferences(() => {
      const nextPreferences: Record<string, NavPreference> = {}
      navEditOrder.forEach((itemKey, index) => {
        nextPreferences[itemKey] = {
          order: index,
          hidden: navDraftPreferences[itemKey]?.hidden ?? false,
        }
      })
      return nextPreferences
    })
    closeNavEditor()
  }

  const handleNavDragStart = (
    key: string,
    event: DragEvent<HTMLButtonElement>
  ) => {
    event.dataTransfer.effectAllowed = 'move'
    event.dataTransfer.setData('text/plain', key)
    setDraggedNavKey(key)
  }

  const handleNavDragOver = (key: string, event: DragEvent<HTMLDivElement>) => {
    if (!draggedNavKey || draggedNavKey === key) {
      return
    }

    event.preventDefault()
    event.dataTransfer.dropEffect = 'move'

    const bounds = event.currentTarget.getBoundingClientRect()
    const midpoint = bounds.top + bounds.height / 2
    const position = event.clientY < midpoint ? 'before' : 'after'
    setDragOverNavTarget({ key, position })
  }

  const handleNavDrop = (key: string, event: DragEvent<HTMLDivElement>) => {
    event.preventDefault()
    const droppedKey = event.dataTransfer.getData('text/plain') || draggedNavKey
    if (!droppedKey || droppedKey === key) {
      setDraggedNavKey(null)
      setDragOverNavTarget(null)
      return
    }

    setNavEditOrder((current) => {
      const targetIndex = current.findIndex((itemKey) => itemKey === key)
      if (targetIndex < 0) {
        return current
      }
      const insertionIndex =
        dragOverNavTarget?.key === key && dragOverNavTarget.position === 'after'
          ? targetIndex + 1
          : targetIndex
      return moveKeyToIndex(current, droppedKey, insertionIndex)
    })
    setDraggedNavKey(null)
    setDragOverNavTarget(null)
  }

  const clearNavDragState = () => {
    setDraggedNavKey(null)
    setDragOverNavTarget(null)
  }

  const togglePrimaryVisibility = (key: string) => {
    setNavDraftPreferences((current) => ({
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
  const orderedPrimaryKeys = orderedPrimaryItems.map((item) =>
    navKeyForTitle(item.title)
  )
  const primaryItemsByKey = new Map(
    primaryItems.map((item) => [navKeyForTitle(item.title), item] as const)
  )

  const configuredPrimaryItems = orderedPrimaryItems.filter((item) => {
    const key = navKeyForTitle(item.title)
    return !navPreferences[key]?.hidden
  })
  const visibleDraftItemCount = navEditOrder.filter(
    (key) => !(navDraftPreferences[key]?.hidden ?? false)
  ).length
  const hasHiddenDraftItems = navEditOrder.some(
    (key) => navDraftPreferences[key]?.hidden ?? false
  )
  const resetNavEditorDefaults = () => {
    clearNavDragState()
    setNavEditOrder(primaryItems.map((item) => navKeyForTitle(item.title)))
    setNavDraftPreferences({})
  }
  const restoreHiddenDraftItems = () => {
    setNavDraftPreferences((current) => {
      const next = { ...current }
      navEditOrder.forEach((itemKey, index) => {
        next[itemKey] = {
          order: current[itemKey]?.order ?? index,
          hidden: false,
        }
      })
      return next
    })
  }

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

  const sidebarGroupsWithPreferences = sidebarData.navGroups.map(
    (group, index) => ({
      ...group,
      items: index === 0 ? configuredPrimaryItems : group.items,
    })
  )

  const translatedNavGroups = sidebarGroupsWithPreferences.map((group) => ({
    ...group,
    title: t(`groups.${normalizeNavKey(group.title)}`, {
      defaultValue: group.title,
    }),
    items: group.items.map(translateItem),
  }))
  const markNotificationInboxOpening = () => {
    try {
      window.localStorage.setItem(
        'cabinet.notification_inbox.origin_route',
        window.location.pathname || '/'
      )
    } catch {
      // Keep navigation working when browser storage is unavailable.
    }
    setActiveWorkspace('navigation')
  }
  const inboxRouteActive = location.pathname.startsWith('/inbox')
  const inboxActive = inboxRouteActive || activeWorkspace === 'inbox'
  const navigationActive = !inboxActive && activeWorkspace === 'navigation'
  const searchActive = !inboxActive && activeWorkspace === 'search'
  const assistantActive = !inboxActive && activeWorkspace === 'assistant'

  return (
    <Sidebar
      collapsible={collapsible}
      variant={variant}
      className={
        assistantActive
          ? 'max-w-full overflow-x-hidden data-[state=open]:animate-none'
          : undefined
      }
    >
      <SidebarHeader>
        <TeamSwitcher teams={sidebarData.teams} />
        <div
          className='px-2 pb-2'
          data-collapsed={isCollapsedSidebar ? 'true' : 'false'}
          data-testid='shell-workspace-switcher'
        >
          <div
            className={`inline-flex gap-1 rounded-lg border border-slate-700/70 bg-slate-950 p-1 shadow-sm ${
              isCollapsedSidebar ? 'flex-col' : 'items-center'
            }`}
            aria-label='Workspace tools'
            data-testid='shell-workspace-icon-rail'
          >
            <button
              type='button'
              aria-label='Navigation workspace'
              title='Navigation workspace'
              data-testid='shell-workspace-navigation'
              data-active={navigationActive ? 'true' : 'false'}
              className='inline-flex h-8 w-8 items-center justify-center rounded-md text-slate-300 transition hover:bg-slate-800 hover:text-white focus-visible:ring-2 focus-visible:ring-slate-200 focus-visible:outline-none data-[active=true]:bg-slate-800 data-[active=true]:text-white data-[active=true]:ring-1 data-[active=true]:ring-slate-500'
              onClick={() => setActiveWorkspace('navigation')}
            >
              <PanelLeft className='h-4 w-4' aria-hidden />
            </button>
            <button
              type='button'
              aria-label='Search workspace'
              title='Search workspace'
              data-testid='shell-workspace-search'
              data-active={searchActive ? 'true' : 'false'}
              className='inline-flex h-8 w-8 items-center justify-center rounded-md text-slate-300 transition hover:bg-slate-800 hover:text-white focus-visible:ring-2 focus-visible:ring-slate-200 focus-visible:outline-none data-[active=true]:bg-slate-800 data-[active=true]:text-white data-[active=true]:ring-1 data-[active=true]:ring-slate-500'
              onClick={() => {
                setOpen(true)
                setActiveWorkspace('search')
              }}
            >
              <SearchIcon className='h-4 w-4' aria-hidden />
            </button>
            <button
              type='button'
              aria-label='Cabinet Agent'
              title='Cabinet Agent'
              data-testid='shell-workspace-assistant'
              data-active={assistantActive ? 'true' : 'false'}
              className='inline-flex h-8 w-8 items-center justify-center rounded-md text-slate-300 transition hover:bg-slate-800 hover:text-white focus-visible:ring-2 focus-visible:ring-slate-200 focus-visible:outline-none data-[active=true]:bg-slate-800 data-[active=true]:text-white data-[active=true]:ring-1 data-[active=true]:ring-slate-500'
              onClick={() => setActiveWorkspace('assistant')}
            >
              <MessageSquare className='h-4 w-4' aria-hidden />
            </button>
            <a
              href='/inbox'
              aria-label='Open notification inbox'
              title='Open notification inbox'
              data-testid='shell-workspace-bell'
              data-active={inboxActive ? 'true' : 'false'}
              className='relative inline-flex h-8 w-8 items-center justify-center rounded-md text-slate-300 transition hover:bg-slate-800 hover:text-white focus-visible:ring-2 focus-visible:ring-slate-200 focus-visible:outline-none data-[active=true]:bg-slate-800 data-[active=true]:text-white data-[active=true]:ring-1 data-[active=true]:ring-slate-500'
              onClick={markNotificationInboxOpening}
            >
              <Bell className='h-4 w-4' aria-hidden />
              <span
                aria-hidden
                className='absolute top-1.5 right-1.5 h-1.5 w-1.5 rounded-full bg-rose-400 ring-1 ring-slate-950'
                data-testid='shell-workspace-bell-badge'
              />
            </a>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <button
                  type='button'
                  aria-label='Open workspace menu'
                  title='Open workspace menu'
                  data-testid='shell-workspace-menu-trigger'
                  className='inline-flex h-8 w-8 items-center justify-center rounded-md text-slate-300 transition hover:bg-slate-800 hover:text-white focus-visible:ring-2 focus-visible:ring-slate-200 focus-visible:outline-none data-[state=open]:bg-slate-800 data-[state=open]:text-white'
                >
                  <MoreHorizontal className='h-4 w-4' aria-hidden />
                </button>
              </DropdownMenuTrigger>
              <DropdownMenuContent
                align='start'
                side={isCollapsedSidebar ? 'right' : 'bottom'}
                className='w-48'
              >
                <DropdownMenuItem
                  onSelect={openNavEditor}
                  data-testid='shell-workspace-menu-customise-nav'
                >
                  <SlidersHorizontal className='h-4 w-4' aria-hidden />
                  <span>Customise Nav</span>
                </DropdownMenuItem>
                <DropdownMenuItem asChild>
                  <a
                    href='/settings/display'
                    data-testid='shell-workspace-menu-settings'
                  >
                    <Settings className='h-4 w-4' aria-hidden />
                    <span>Settings</span>
                  </a>
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>

        {/* Replace <TeamSwitch /> with the following <AppTitle />
         /* if you want to use the normal app title instead of TeamSwitch dropdown */}
        {/* <AppTitle /> */}
      </SidebarHeader>
      <SidebarContent
        className={
          navEditMode
            ? 'overflow-hidden'
            : activeWorkspace === 'assistant'
              ? 'max-w-full min-w-0 overflow-x-hidden overflow-y-auto'
              : undefined
        }
      >
        {navEditMode ? (
          <div
            className='flex min-h-full flex-col overflow-hidden px-2 py-2'
            data-testid='sidebar-nav-edit-panel'
          >
            <div className='shrink-0 border-b border-sidebar-border/70 pb-2'>
              <h2 className='text-sm font-semibold'>Customise Nav</h2>
              <p className='mt-1 text-xs leading-5 text-muted-foreground'>
                Reorder and hide primary nav without changing routes or
                permissions.
              </p>
              <p
                className='mt-2 text-xs font-medium'
                data-testid='sidebar-nav-visible-count'
              >
                {visibleDraftItemCount} items visible in the sidebar
              </p>
            </div>
            <div
              className='min-h-0 flex-1 space-y-1 overflow-y-auto py-2 pr-1'
              data-testid='sidebar-nav-edit-list'
            >
              {navEditOrder.map((key, index) => {
                const item = primaryItemsByKey.get(key)
                if (!item) {
                  return null
                }
                const hidden = navDraftPreferences[key]?.hidden ?? false
                const dragPosition =
                  dragOverNavTarget?.key === key
                    ? dragOverNavTarget.position
                    : null
                const isFirst = index === 0
                const isLast = index === navEditOrder.length - 1
                return (
                  <div
                    key={key}
                    className='relative scroll-my-16'
                    data-testid={`sidebar-nav-edit-dropzone-${key}`}
                    onDragOver={(event) => handleNavDragOver(key, event)}
                    onDrop={(event) => handleNavDrop(key, event)}
                    onDragLeave={() => {
                      if (dragOverNavTarget?.key === key) {
                        setDragOverNavTarget(null)
                      }
                    }}
                  >
                    {dragPosition === 'before' ? (
                      <div
                        className='mx-1 mb-1 h-0.5 rounded-full bg-primary'
                        data-testid={`sidebar-nav-drop-indicator-before-${key}`}
                      />
                    ) : null}
                    <div
                      className='rounded-md border border-sidebar-border bg-sidebar-accent/30 p-2'
                      data-testid={`sidebar-nav-edit-item-${key}`}
                      data-dragging={draggedNavKey === key ? 'true' : 'false'}
                    >
                      <div className='flex min-w-0 items-center gap-2'>
                        <button
                          type='button'
                          draggable
                          aria-label={`Drag ${item.title}`}
                          title={`Drag ${item.title}`}
                          data-testid={`sidebar-nav-drag-handle-${key}`}
                          className='cursor-grab rounded border border-sidebar-border p-1 text-muted-foreground hover:bg-sidebar-accent active:cursor-grabbing'
                          onDragStart={(event) =>
                            handleNavDragStart(key, event)
                          }
                          onDragEnd={clearNavDragState}
                        >
                          <GripVertical className='h-3 w-3' aria-hidden />
                        </button>
                        <div className='min-w-0 flex-1'>
                          <p
                            className={
                              hidden
                                ? 'truncate text-xs font-medium opacity-50'
                                : 'truncate text-xs font-medium'
                            }
                          >
                            {item.title}
                          </p>
                          <p
                            className='truncate text-[11px] text-muted-foreground'
                            data-testid={`sidebar-nav-stable-id-${key}`}
                          >
                            {key}
                          </p>
                        </div>
                      </div>
                      <div className='mt-2 grid grid-cols-3 gap-1'>
                        <button
                          type='button'
                          aria-label={`Move ${item.title} up`}
                          title={`Move ${item.title} up`}
                          data-testid={`sidebar-nav-move-up-${key}`}
                          disabled={isFirst}
                          className='inline-flex h-7 items-center justify-center rounded border border-sidebar-border text-xs hover:bg-sidebar-accent disabled:cursor-not-allowed disabled:opacity-45'
                          onClick={() => movePrimaryItem(key, 'up')}
                        >
                          <ArrowUp className='h-3 w-3' aria-hidden />
                          <span className='sr-only'>Move up</span>
                        </button>
                        <button
                          type='button'
                          aria-label={`Move ${item.title} down`}
                          title={`Move ${item.title} down`}
                          data-testid={`sidebar-nav-move-down-${key}`}
                          disabled={isLast}
                          className='inline-flex h-7 items-center justify-center rounded border border-sidebar-border text-xs hover:bg-sidebar-accent disabled:cursor-not-allowed disabled:opacity-45'
                          onClick={() => movePrimaryItem(key, 'down')}
                        >
                          <ArrowDown className='h-3 w-3' aria-hidden />
                          <span className='sr-only'>Move down</span>
                        </button>
                        <button
                          type='button'
                          data-testid={`sidebar-nav-visibility-${key}`}
                          aria-label={`${hidden ? 'Show' : 'Hide'} ${item.title}`}
                          className='inline-flex h-7 items-center justify-center gap-1 rounded border border-sidebar-border px-1 text-xs hover:bg-sidebar-accent'
                          onClick={() => togglePrimaryVisibility(key)}
                        >
                          {hidden ? (
                            <Eye className='h-3 w-3' aria-hidden />
                          ) : (
                            <EyeOff className='h-3 w-3' aria-hidden />
                          )}
                          <span>{hidden ? 'Show' : 'Hide'}</span>
                        </button>
                      </div>
                    </div>
                    {dragPosition === 'after' ? (
                      <div
                        className='mx-1 mt-1 h-0.5 rounded-full bg-primary'
                        data-testid={`sidebar-nav-drop-indicator-after-${key}`}
                      />
                    ) : null}
                  </div>
                )
              })}
            </div>
          </div>
        ) : activeWorkspace === 'navigation' ? (
          translatedNavGroups.map((props) => (
            <NavGroup key={props.title} {...props} />
          ))
        ) : activeWorkspace === 'search' ? (
          <SearchWorkspacePanel />
        ) : activeWorkspace === 'inbox' ? (
          <InboxWorkspacePanel />
        ) : null}
        {activeWorkspace === 'assistant' ? <AssistantWorkspacePanel /> : null}
      </SidebarContent>
      <SidebarFooter>
        {navEditMode ? (
          <div
            className='grid grid-cols-2 gap-1 border-t border-sidebar-border/70 pt-2'
            data-testid='sidebar-nav-edit-footer'
          >
            <button
              type='button'
              className='rounded border border-sidebar-border px-2 py-1.5 text-xs hover:bg-sidebar-accent disabled:cursor-not-allowed disabled:opacity-45'
              data-testid='sidebar-nav-restore-hidden'
              disabled={!hasHiddenDraftItems}
              onClick={restoreHiddenDraftItems}
            >
              Restore hidden items
            </button>
            <button
              type='button'
              className='rounded border border-sidebar-border px-2 py-1.5 text-xs hover:bg-sidebar-accent'
              data-testid='sidebar-nav-reset-defaults'
              onClick={resetNavEditorDefaults}
            >
              Reset defaults
            </button>
            <button
              type='button'
              className='rounded border border-sidebar-border px-2 py-1.5 text-xs hover:bg-sidebar-accent'
              data-testid='sidebar-nav-cancel'
              onClick={closeNavEditor}
            >
              Cancel
            </button>
            <button
              type='button'
              className='rounded bg-primary px-2 py-1.5 text-xs text-primary-foreground hover:bg-primary/90'
              data-testid='sidebar-nav-apply'
              onClick={applyNavEditor}
            >
              Apply
            </button>
          </div>
        ) : null}
        <NavUser user={sidebarUser} />
        {!isCollapsedSidebar ? (
          <div
            className='px-2 pb-2 text-xs text-muted-foreground'
            data-testid='sidebar-runtime-meta'
          >
            <p>
              Version:{' '}
              <span data-testid='sidebar-app-version'>
                {runtimeMeta.appVersion}
              </span>
            </p>
            <p>
              Build Date:{' '}
              <span data-testid='sidebar-build-date'>
                {runtimeMeta.buildDate}
              </span>
            </p>
          </div>
        ) : null}
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  )
}
