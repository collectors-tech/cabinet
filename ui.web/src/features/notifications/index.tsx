import { useCallback, useEffect, useMemo, useState } from 'react'
import {
  Archive,
  Bell,
  CheckCheck,
  ChevronDown,
  ChevronUp,
  ExternalLink,
  Inbox,
  Mail,
  MailOpen,
  RefreshCw,
  Search,
} from 'lucide-react'
import { loadToastHistory, type ToastHistoryRecord } from '@/lib/toast-history'
import { cn } from '@/lib/utils'
import { useShellWorkspace } from '@/context/shell-workspace-provider'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Input } from '@/components/ui/input'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Header, HeaderTitle } from '@/components/layout/header'

type NotificationInboxItem = {
  id: string
  profile_id?: string
  thread_id?: string
  source?: string
  status: string
  title: string
  summary: string
  metadata?: {
    category?: string
    detail?: string
    source_label?: string
    assistant?: { provider?: string; model?: string }
    review_url?: string
    preview_id?: string
    confirmation_state?: string
    telegram_reply?: {
      review_url?: string
      confirmation_state?: string
    }
    item?: {
      id?: string
      title?: string
      part_number?: string
      href?: string
    }
    item_id?: string
    item_title?: string
    item_href?: string
    local_toast?: boolean
    level?: string
  }
  created_at?: string
  updated_at?: string
}

type InboxFilter = 'all' | 'unread' | 'assistant' | 'system'

const filterLabels: Record<InboxFilter, string> = {
  all: 'All',
  unread: 'Unread',
  assistant: 'Assistant',
  system: 'System',
}

const emptyStateByFilter: Record<InboxFilter, string> = {
  all: 'No notifications need review across this profile.',
  unread: 'No unread notifications are waiting.',
  assistant: 'No assistant handoffs or mentions are waiting.',
  system: 'No system or runtime notices are waiting.',
}

function normalizeStatus(status: string) {
  const normalized = status.trim().toLowerCase()
  return normalized === 'queued' ? 'unread' : normalized
}

function sourceLabel(source?: string) {
  const normalized = source?.trim().replace(/_/g, ' ')
  if (!normalized) {
    return 'Notification'
  }
  return normalized.replace(/\b\w/g, (letter) => letter.toUpperCase())
}

function categoryForItem(item: NotificationInboxItem) {
  const explicit = item.metadata?.category?.trim().toLowerCase()
  if (explicit === 'system' || explicit === 'runtime') {
    return 'system'
  }
  if (
    explicit === 'assistant' ||
    explicit === 'mention' ||
    item.source?.includes('assistant') ||
    item.source?.includes('telegram')
  ) {
    return 'assistant'
  }
  if (item.source?.includes('system') || item.source?.includes('runtime')) {
    return 'system'
  }
  return 'notification'
}

function formatTimestamp(value?: string) {
  if (!value) {
    return 'Recent'
  }
  const timestamp = new Date(value)
  if (Number.isNaN(timestamp.getTime())) {
    return 'Recent'
  }
  return timestamp.toLocaleString(undefined, {
    month: 'short',
    day: 'numeric',
    hour: 'numeric',
    minute: '2-digit',
  })
}

function targetLink(item: NotificationInboxItem) {
  const metadata = item.metadata
  const reviewHref =
    metadata?.review_url ?? metadata?.telegram_reply?.review_url ?? ''
  if (item.source === 'telegram_catalog_capture' && reviewHref) {
    return {
      href: reviewHref,
      label: 'Review capture',
    }
  }
  const href = metadata?.item?.href ?? metadata?.item_href
  const id = metadata?.item?.id ?? metadata?.item_id
  const title = metadata?.item?.title ?? metadata?.item_title
  const partNumber = metadata?.item?.part_number
  if (!href && !id && !title && !partNumber) {
    return null
  }
  return {
    href: href ?? `/inventory/?item=${encodeURIComponent(id ?? '')}`,
    label:
      [partNumber, title].filter(Boolean).join(' - ') ||
      title ||
      id ||
      'Open target',
  }
}

function sortNotifications(items: NotificationInboxItem[]) {
  return [...items].sort((a, b) => {
    const aStatus = normalizeStatus(a.status) === 'unread' ? 0 : 1
    const bStatus = normalizeStatus(b.status) === 'unread' ? 0 : 1
    if (aStatus !== bStatus) {
      return aStatus - bStatus
    }
    return (
      new Date(b.updated_at ?? b.created_at ?? 0).getTime() -
      new Date(a.updated_at ?? a.created_at ?? 0).getTime()
    )
  })
}

function itemFromToastHistory(
  record: ToastHistoryRecord
): NotificationInboxItem {
  return {
    id: `toast:${record.id}`,
    source: 'toast_history',
    status: 'read',
    title: record.title,
    summary: record.summary || `${sourceLabel(record.level)} toast`,
    metadata: {
      category: 'system',
      source_label: 'Toast History',
      detail: record.summary || record.title,
      local_toast: true,
      level: record.level,
    },
    created_at: record.created_at,
    updated_at: record.created_at,
  }
}

function isLocalToastItem(item: NotificationInboxItem) {
  return item.metadata?.local_toast === true || item.id.startsWith('toast:')
}

export function NotificationInbox() {
  const { activeProfileId, setActiveWorkspace } = useShellWorkspace()
  const [items, setItems] = useState<NotificationInboxItem[]>([])
  const [toastItems, setToastItems] = useState<NotificationInboxItem[]>([])
  const [filter, setFilter] = useState<InboxFilter>('all')
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedItemId, setSelectedItemId] = useState('')
  const [selectedIds, setSelectedIds] = useState<string[]>([])
  const [expandedIds, setExpandedIds] = useState<string[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [updating, setUpdating] = useState(false)

  const loadItems = useCallback(async () => {
    if (!activeProfileId) {
      return
    }
    setLoading(true)
    setError('')
    try {
      const response = await fetch(
        `/api/chat/inbox?profile_id=${encodeURIComponent(activeProfileId)}`
      )
      if (!response.ok) {
        throw new Error('notification_inbox_load_failed')
      }
      const payload = (await response.json()) as {
        items?: NotificationInboxItem[]
      }
      setItems(sortNotifications(payload.items ?? []))
    } catch (err) {
      setError(
        err instanceof Error ? err.message : 'notification_inbox_load_failed'
      )
    } finally {
      setLoading(false)
    }
  }, [activeProfileId])

  useEffect(() => {
    const loadToastItems = () => {
      setToastItems(loadToastHistory().map(itemFromToastHistory))
    }
    loadToastItems()
    window.addEventListener('cabinet:toast-history', loadToastItems)
    window.addEventListener('storage', loadToastItems)
    return () => {
      window.removeEventListener('cabinet:toast-history', loadToastItems)
      window.removeEventListener('storage', loadToastItems)
    }
  }, [])

  async function updateItems(
    ids: string[],
    status: 'read' | 'unread' | 'archived'
  ) {
    if (!activeProfileId || ids.length === 0) {
      return
    }
    setUpdating(true)
    setError('')
    try {
      const localIds = ids.filter((id) =>
        [...items, ...toastItems].some(
          (item) => item.id === id && isLocalToastItem(item)
        )
      )
      const remoteIds = ids.filter((id) => !localIds.includes(id))
      const updatedItems = await Promise.all(
        remoteIds.map(async (id) => {
          const response = await fetch(
            `/api/chat/inbox/${encodeURIComponent(id)}`,
            {
              method: 'PATCH',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify({
                profile_id: activeProfileId,
                status,
              }),
            }
          )
          if (!response.ok) {
            throw new Error('notification_inbox_update_failed')
          }
          return (await response.json()) as NotificationInboxItem
        })
      )
      setItems((current) => {
        if (status === 'archived') {
          return current.filter((item) => !remoteIds.includes(item.id))
        }
        return sortNotifications(
          current.map((item) => {
            const updated = updatedItems.find(
              (candidate) => candidate.id === item.id
            )
            return updated ? { ...item, ...updated } : item
          })
        )
      })
      setToastItems((current) => {
        if (status === 'archived') {
          return current.filter((item) => !localIds.includes(item.id))
        }
        return sortNotifications(
          current.map((item) =>
            localIds.includes(item.id)
              ? { ...item, status, updated_at: new Date().toISOString() }
              : item
          )
        )
      })
      setSelectedIds((current) => current.filter((id) => !ids.includes(id)))
    } catch (err) {
      setError(
        err instanceof Error ? err.message : 'notification_inbox_update_failed'
      )
    } finally {
      setUpdating(false)
    }
  }

  useEffect(() => {
    void loadItems()
  }, [loadItems])

  useEffect(() => {
    setSelectedIds([])
  }, [filter])

  const allItems = useMemo(
    () => sortNotifications([...items, ...toastItems]),
    [items, toastItems]
  )

  const visibleItems = useMemo(() => {
    const query = searchQuery.trim().toLowerCase()
    return allItems.filter((item) => {
      const status = normalizeStatus(item.status)
      if (status === 'archived') {
        return false
      }
      const matchesFilter =
        filter === 'all' ||
        (filter === 'unread' && status === 'unread') ||
        (filter === 'assistant' && categoryForItem(item) === 'assistant') ||
        (filter === 'system' && categoryForItem(item) === 'system')
      const matchesSearch =
        !query ||
        [
          item.title,
          item.summary,
          item.metadata?.detail,
          item.metadata?.source_label,
          item.source,
        ]
          .filter(Boolean)
          .some((value) => String(value).toLowerCase().includes(query))
      return matchesFilter && matchesSearch
    })
  }, [allItems, filter, searchQuery])

  const counts = useMemo(() => {
    const active = allItems.filter(
      (item) => normalizeStatus(item.status) !== 'archived'
    )
    return {
      all: active.length,
      unread: active.filter((item) => normalizeStatus(item.status) === 'unread')
        .length,
      assistant: active.filter((item) => categoryForItem(item) === 'assistant')
        .length,
      system: active.filter((item) => categoryForItem(item) === 'system')
        .length,
    }
  }, [allItems])

  const selectedItem = useMemo(() => {
    return (
      visibleItems.find((item) => item.id === selectedItemId) ??
      visibleItems[0] ??
      null
    )
  }, [selectedItemId, visibleItems])

  useEffect(() => {
    setSelectedItemId((current) => {
      if (visibleItems.some((item) => item.id === current)) {
        return current
      }
      return visibleItems[0]?.id ?? ''
    })
  }, [visibleItems])

  const allVisibleSelected =
    visibleItems.length > 0 &&
    visibleItems.every((item) => selectedIds.includes(item.id))

  function toggleSelection(id: string) {
    setSelectedIds((current) =>
      current.includes(id)
        ? current.filter((candidate) => candidate !== id)
        : [...current, id]
    )
  }

  function toggleExpanded(id: string) {
    setExpandedIds((current) =>
      current.includes(id)
        ? current.filter((candidate) => candidate !== id)
        : [...current, id]
    )
  }

  function openCompactInbox() {
    setActiveWorkspace('inbox')
  }

  return (
    <>
      <Header fixed>
        <HeaderTitle
          title='Notification Inbox'
          description='Operational notifications, assistant handoffs, and system notices'
          icon={Bell}
          testId='notification-inbox-header-title'
          iconTestId='notification-inbox-page-icon'
        />
        <div
          className='ms-auto flex items-center gap-2'
          data-header-title-avoid='true'
        >
          <Button
            type='button'
            variant='outline'
            onClick={() => void loadItems()}
            disabled={loading}
            data-testid='notification-inbox-refresh'
          >
            <RefreshCw className='h-4 w-4' />
            Refresh
          </Button>
          <Button
            type='button'
            variant='outline'
            onClick={openCompactInbox}
            data-testid='notification-inbox-open-compact'
          >
            <Inbox className='h-4 w-4' />
            Compact Inbox
          </Button>
        </div>
      </Header>
      <main
        className='flex h-[calc(100svh-4rem)] min-h-0 flex-col gap-4 overflow-hidden px-4 py-4 sm:px-6 lg:px-8'
        data-testid='notification-inbox-page'
        data-layout='full-height-split'
      >
        <section className='space-y-3'>
          <div className='grid gap-3 sm:grid-cols-2 xl:grid-cols-4'>
            {(['all', 'unread', 'assistant', 'system'] as InboxFilter[]).map(
              (key) => (
                <div
                  key={key}
                  className='rounded-md border bg-card px-3 py-2'
                  data-testid={`notification-inbox-count-${key}`}
                >
                  <p className='text-xs font-medium text-muted-foreground'>
                    {filterLabels[key]}
                  </p>
                  <p className='text-2xl font-semibold'>{counts[key]}</p>
                </div>
              )
            )}
          </div>
        </section>

        <section className='flex min-h-0 flex-1 flex-col gap-3'>
          <div className='flex flex-col gap-3 rounded-md border bg-card p-3 xl:flex-row xl:items-center xl:justify-between'>
            <Tabs
              value={filter}
              onValueChange={(value) => setFilter(value as InboxFilter)}
            >
              <TabsList data-testid='notification-inbox-filters'>
                {(
                  ['all', 'unread', 'assistant', 'system'] as InboxFilter[]
                ).map((key) => (
                  <TabsTrigger
                    key={key}
                    value={key}
                    data-testid={`notification-inbox-filter-${key}`}
                  >
                    {filterLabels[key]}
                  </TabsTrigger>
                ))}
              </TabsList>
            </Tabs>
            <div className='relative min-w-0 flex-1 xl:max-w-sm'>
              <Search className='pointer-events-none absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-muted-foreground' />
              <Input
                value={searchQuery}
                onChange={(event) => setSearchQuery(event.target.value)}
                placeholder='Search notifications'
                className='pl-9'
                data-testid='notification-inbox-search'
              />
            </div>
            <div className='flex flex-wrap items-center gap-2'>
              <label
                className='flex items-center gap-2 text-sm'
                data-testid='notification-inbox-select-all'
              >
                <Checkbox
                  checked={allVisibleSelected}
                  disabled={visibleItems.length === 0}
                  onCheckedChange={(checked) => {
                    setSelectedIds(
                      checked ? visibleItems.map((item) => item.id) : []
                    )
                  }}
                />
                Select visible
              </label>
              <Button
                type='button'
                variant='outline'
                disabled={selectedIds.length === 0 || updating}
                onClick={() => void updateItems(selectedIds, 'read')}
                data-testid='notification-inbox-bulk-read'
              >
                <CheckCheck className='h-4 w-4' />
                Mark read
              </Button>
              <Button
                type='button'
                variant='outline'
                disabled={selectedIds.length === 0 || updating}
                onClick={() => void updateItems(selectedIds, 'archived')}
                data-testid='notification-inbox-bulk-archive'
              >
                <Archive className='h-4 w-4' />
                Archive
              </Button>
            </div>
          </div>

          {error ? (
            <Alert
              variant='destructive'
              data-testid='notification-inbox-error-state'
            >
              <AlertTitle>Notification Inbox could not update.</AlertTitle>
              <AlertDescription>
                {error}
                <Button
                  type='button'
                  variant='outline'
                  size='sm'
                  className='ml-3'
                  onClick={() => void loadItems()}
                  data-testid='notification-inbox-retry'
                >
                  Retry
                </Button>
              </AlertDescription>
            </Alert>
          ) : null}

          <div className='grid min-h-0 flex-1 gap-3 lg:grid-cols-[minmax(0,1fr)_minmax(20rem,24rem)]'>
            <div
              className='min-h-0 overflow-y-auto pr-1'
              data-testid='notification-inbox-list-pane'
            >
              {loading ? (
                <div
                  className='rounded-md border bg-card p-6 text-sm text-muted-foreground'
                  data-testid='notification-inbox-loading-state'
                >
                  Loading Notification Inbox...
                </div>
              ) : null}

              {!loading && visibleItems.length === 0 ? (
                <div
                  className='rounded-md border bg-card p-6 text-sm text-muted-foreground'
                  data-testid='notification-inbox-empty-state'
                >
                  <p className='font-medium text-foreground'>
                    {filterLabels[filter]} is clear.
                  </p>
                  <p>{emptyStateByFilter[filter]}</p>
                </div>
              ) : null}

              {!loading && visibleItems.length > 0 ? (
                <div
                  className='min-h-[720px] space-y-2'
                  data-testid='notification-inbox-list'
                >
                  {visibleItems.map((item) => {
                    const status = normalizeStatus(item.status)
                    const read = status === 'read'
                    const selected = selectedIds.includes(item.id)
                    const expanded = expandedIds.includes(item.id)
                    const link = targetLink(item)
                    return (
                      <article
                        key={item.id}
                        className={cn(
                          'rounded-md border bg-card p-3 transition-colors',
                          !read && 'border-primary/40 bg-primary/5',
                          selectedItem?.id === item.id &&
                            'ring-2 ring-primary/30'
                        )}
                        onClick={() => setSelectedItemId(item.id)}
                        data-testid='notification-inbox-row'
                        data-status={status}
                        data-category={categoryForItem(item)}
                      >
                        <div className='grid gap-3 md:grid-cols-[auto_1fr_auto] md:items-start'>
                          <Checkbox
                            checked={selected}
                            aria-label={`Select ${item.title}`}
                            onCheckedChange={() => toggleSelection(item.id)}
                            data-testid='notification-inbox-row-select'
                          />
                          <div className='min-w-0 space-y-2'>
                            <div className='flex flex-wrap items-center gap-2'>
                              {read ? (
                                <MailOpen className='h-4 w-4 text-muted-foreground' />
                              ) : (
                                <Mail className='h-4 w-4 text-primary' />
                              )}
                              <h3
                                className='font-semibold'
                                data-testid='notification-inbox-row-title'
                              >
                                {item.title}
                              </h3>
                              <Badge
                                variant={read ? 'outline' : 'secondary'}
                                data-testid='notification-inbox-row-status'
                              >
                                {status}
                              </Badge>
                              <Badge
                                variant='outline'
                                data-testid='notification-inbox-row-category'
                              >
                                {categoryForItem(item)}
                              </Badge>
                            </div>
                            <p className='text-sm text-muted-foreground'>
                              {item.summary}
                            </p>
                            <div className='flex flex-wrap gap-3 text-xs text-muted-foreground'>
                              <span data-testid='notification-inbox-row-source'>
                                {item.metadata?.source_label ??
                                  sourceLabel(item.source)}
                              </span>
                              <span>{formatTimestamp(item.created_at)}</span>
                              {link ? (
                                <a
                                  className='inline-flex items-center gap-1 text-primary underline-offset-4 hover:underline'
                                  href={link.href}
                                  data-testid='notification-inbox-row-link'
                                >
                                  <ExternalLink className='h-3.5 w-3.5' />
                                  {link.label}
                                </a>
                              ) : null}
                            </div>
                          </div>
                          <div className='flex flex-wrap justify-start gap-2 md:justify-end'>
                            <Button
                              type='button'
                              variant='outline'
                              size='sm'
                              onClick={() =>
                                void updateItems(
                                  [item.id],
                                  read ? 'unread' : 'read'
                                )
                              }
                              disabled={updating}
                              data-testid={
                                read
                                  ? 'notification-inbox-row-unread'
                                  : 'notification-inbox-row-read'
                              }
                            >
                              {read ? (
                                <Mail className='h-4 w-4' />
                              ) : (
                                <MailOpen className='h-4 w-4' />
                              )}
                              {read ? 'Unread' : 'Read'}
                            </Button>
                            <Button
                              type='button'
                              variant='outline'
                              size='sm'
                              onClick={() =>
                                void updateItems([item.id], 'archived')
                              }
                              disabled={updating}
                              data-testid='notification-inbox-row-archive'
                            >
                              <Archive className='h-4 w-4' />
                              Archive
                            </Button>
                            <Button
                              type='button'
                              variant='outline'
                              size='sm'
                              onClick={() => toggleExpanded(item.id)}
                              aria-expanded={expanded}
                              data-testid='notification-inbox-row-expand'
                            >
                              {expanded ? (
                                <ChevronUp className='h-4 w-4' />
                              ) : (
                                <ChevronDown className='h-4 w-4' />
                              )}
                              Details
                            </Button>
                          </div>
                        </div>
                        {expanded ? (
                          <div
                            className='mt-3 rounded-md border bg-background p-3 text-sm'
                            data-testid='notification-inbox-row-detail'
                          >
                            <p>
                              {item.metadata?.detail ||
                                item.metadata?.confirmation_state ||
                                item.summary}
                            </p>
                            <p className='mt-2 text-xs text-muted-foreground'>
                              Source:{' '}
                              {item.metadata?.source_label ??
                                sourceLabel(item.source)}
                            </p>
                          </div>
                        ) : null}
                      </article>
                    )
                  })}
                </div>
              ) : null}
            </div>
            <aside
              className='min-h-0 overflow-y-auto rounded-md border bg-card p-4'
              data-testid='notification-inbox-detail-pane'
            >
              {selectedItem ? (
                <div className='space-y-4'>
                  <div className='space-y-2'>
                    <div className='flex flex-wrap items-center gap-2'>
                      <Badge
                        variant={
                          normalizeStatus(selectedItem.status) === 'read'
                            ? 'outline'
                            : 'secondary'
                        }
                      >
                        {normalizeStatus(selectedItem.status)}
                      </Badge>
                      <Badge variant='outline'>
                        {categoryForItem(selectedItem)}
                      </Badge>
                    </div>
                    <h2 className='text-lg font-semibold'>
                      {selectedItem.title}
                    </h2>
                    <p className='text-sm text-muted-foreground'>
                      {selectedItem.summary}
                    </p>
                  </div>
                  <div className='space-y-2 text-sm'>
                    <p>
                      {selectedItem.metadata?.detail ||
                        selectedItem.metadata?.confirmation_state ||
                        selectedItem.summary}
                    </p>
                    <dl className='grid gap-2 text-xs text-muted-foreground'>
                      <div>
                        <dt className='font-medium text-foreground'>Source</dt>
                        <dd>
                          {selectedItem.metadata?.source_label ??
                            sourceLabel(selectedItem.source)}
                        </dd>
                      </div>
                      <div>
                        <dt className='font-medium text-foreground'>Created</dt>
                        <dd>{formatTimestamp(selectedItem.created_at)}</dd>
                      </div>
                    </dl>
                  </div>
                  {targetLink(selectedItem) ? (
                    <Button asChild variant='outline' size='sm'>
                      <a href={targetLink(selectedItem)?.href}>
                        <ExternalLink className='h-4 w-4' />
                        {targetLink(selectedItem)?.label}
                      </a>
                    </Button>
                  ) : null}
                </div>
              ) : (
                <p className='text-sm text-muted-foreground'>
                  Select a notification to inspect its details.
                </p>
              )}
            </aside>
          </div>
        </section>
      </main>
    </>
  )
}
