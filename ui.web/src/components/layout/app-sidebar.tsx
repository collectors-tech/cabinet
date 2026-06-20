import { useEffect, useState, type DragEvent } from 'react'
import {
  ArrowDown,
  ArrowUp,
  Eye,
  EyeOff,
  GripVertical,
  Inbox,
  MessageSquare,
  PanelLeft,
  Pencil,
  X,
} from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { useAuthStore } from '@/stores/auth-store'
import { useLayout } from '@/context/layout-provider'
import { useShellWorkspace } from '@/context/shell-workspace-provider'
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
  const { collapsible, variant } = useLayout()
  const { state: sidebarState, isMobile, setOpen } = useSidebar()
  const { t } = useTranslation('nav')
  const { activeWorkspace, setActiveWorkspace } = useShellWorkspace()
  const isCollapsedSidebar = sidebarState === 'collapsed' && !isMobile
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
    if (
      activeWorkspace === 'navigation' &&
      sidebarState === 'collapsed' &&
      !isMobile
    ) {
      setOpen(true)
    }
  }, [activeWorkspace, isMobile, setOpen, sidebarState])

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
  const activeWorkspaceIcon =
    activeWorkspace === 'assistant'
      ? MessageSquare
      : activeWorkspace === 'inbox'
        ? Inbox
        : PanelLeft
  const ActiveWorkspaceIcon = activeWorkspaceIcon

  return (
    <Sidebar collapsible={collapsible} variant={variant}>
      <SidebarHeader>
        <TeamSwitcher teams={sidebarData.teams} />
        <div
          className='px-2 pb-2'
          data-collapsed={isCollapsedSidebar ? 'true' : 'false'}
          data-testid='shell-workspace-switcher'
        >
          <p
            className='mb-2 text-xs font-medium text-muted-foreground'
            data-testid='shell-workspace-label'
          >
            {isCollapsedSidebar ? 'Work' : 'Workspace'}
          </p>
          {isCollapsedSidebar ? (
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <button
                  type='button'
                  aria-label='Switch workspace'
                  data-testid='shell-workspace-menu-trigger'
                  className='inline-flex h-8 w-8 items-center justify-center rounded-md border text-muted-foreground hover:bg-muted'
                >
                  <ActiveWorkspaceIcon className='h-4 w-4' />
                </button>
              </DropdownMenuTrigger>
              <DropdownMenuContent side='right' align='start' sideOffset={8}>
                <DropdownMenuItem
                  data-testid='shell-workspace-menu-navigation'
                  onClick={() => setActiveWorkspace('navigation')}
                >
                  <PanelLeft className='h-4 w-4' />
                  <span>Nav</span>
                </DropdownMenuItem>
                <DropdownMenuItem
                  data-testid='shell-workspace-menu-assistant'
                  onClick={() => setActiveWorkspace('assistant')}
                >
                  <MessageSquare className='h-4 w-4' />
                  <span>Assistant</span>
                </DropdownMenuItem>
                <DropdownMenuItem
                  data-testid='shell-workspace-menu-inbox'
                  onClick={() => setActiveWorkspace('inbox')}
                >
                  <Inbox className='h-4 w-4' />
                  <span>Inbox</span>
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          ) : (
            <div className='grid grid-cols-3 gap-2'>
              <button
                type='button'
                data-testid='shell-workspace-navigation'
                data-active={
                  activeWorkspace === 'navigation' ? 'true' : 'false'
                }
                className='inline-flex items-center justify-center gap-1 rounded-md border px-2 py-1 text-xs hover:bg-muted data-[active=true]:bg-primary data-[active=true]:text-primary-foreground'
                onClick={() => setActiveWorkspace('navigation')}
              >
                <PanelLeft className='h-3.5 w-3.5' />
                Nav
              </button>
              <button
                type='button'
                data-testid='shell-workspace-assistant'
                data-active={activeWorkspace === 'assistant' ? 'true' : 'false'}
                className='inline-flex items-center justify-center gap-1 rounded-md border px-2 py-1 text-xs hover:bg-muted data-[active=true]:bg-primary data-[active=true]:text-primary-foreground'
                onClick={() => setActiveWorkspace('assistant')}
              >
                <MessageSquare className='h-3.5 w-3.5' />
                Assistant
              </button>
              <button
                type='button'
                data-testid='shell-workspace-inbox'
                data-active={activeWorkspace === 'inbox' ? 'true' : 'false'}
                className='inline-flex items-center justify-center gap-1 rounded-md border px-2 py-1 text-xs hover:bg-muted data-[active=true]:bg-primary data-[active=true]:text-primary-foreground'
                onClick={() => setActiveWorkspace('inbox')}
              >
                <Inbox className='h-3.5 w-3.5' />
                Inbox
              </button>
            </div>
          )}
        </div>

        {/* Replace <TeamSwitch /> with the following <AppTitle />
         /* if you want to use the normal app title instead of TeamSwitch dropdown */}
        {/* <AppTitle /> */}
      </SidebarHeader>
      <SidebarContent>
        {activeWorkspace === 'navigation'
          ? translatedNavGroups.map((props) => (
              <NavGroup key={props.title} {...props} />
            ))
          : null}
        {activeWorkspace === 'assistant' ? <AssistantWorkspacePanel /> : null}
        {activeWorkspace === 'inbox' ? <InboxWorkspacePanel /> : null}
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
                clearNavDragState()
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
            {navEditMode ? (
              <X className='h-4 w-4' />
            ) : (
              <Pencil className='h-4 w-4' />
            )}
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
                const dragPosition =
                  dragOverNavTarget?.key === key
                    ? dragOverNavTarget.position
                    : null
                return (
                  <div
                    key={key}
                    className='relative'
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
                        className='absolute inset-x-1 -top-1 h-0.5 rounded-full bg-primary'
                        data-testid={`sidebar-nav-drop-indicator-before-${key}`}
                      />
                    ) : null}
                    <div
                      className='flex items-center justify-between gap-2 rounded-md border border-transparent px-1 py-1'
                      data-testid={`sidebar-nav-edit-item-${key}`}
                      data-dragging={draggedNavKey === key ? 'true' : 'false'}
                    >
                      <div className='flex min-w-0 items-center gap-2'>
                        <button
                          type='button'
                          draggable
                          aria-label={`Drag ${item.title}`}
                          data-testid={`sidebar-nav-drag-handle-${key}`}
                          className='cursor-grab rounded border p-1 text-muted-foreground hover:bg-muted active:cursor-grabbing'
                          onDragStart={(event) =>
                            handleNavDragStart(key, event)
                          }
                          onDragEnd={clearNavDragState}
                        >
                          <GripVertical className='h-3 w-3' />
                        </button>
                        <span
                          className={
                            hidden ? 'truncate opacity-50' : 'truncate'
                          }
                        >
                          {item.title}
                        </span>
                      </div>
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
                    {dragPosition === 'after' ? (
                      <div
                        className='absolute inset-x-1 -bottom-1 h-0.5 rounded-full bg-primary'
                        data-testid={`sidebar-nav-drop-indicator-after-${key}`}
                      />
                    ) : null}
                  </div>
                )
              })}
            </div>
          ) : null}
        </div>
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
