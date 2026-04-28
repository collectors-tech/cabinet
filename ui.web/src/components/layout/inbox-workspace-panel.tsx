import { useEffect, useState } from 'react'
import {
  Archive,
  ExternalLink,
  Inbox,
  Mail,
  MailOpen,
  RefreshCw,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { useShellWorkspace } from '@/context/shell-workspace-provider'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { ScrollArea } from '@/components/ui/scroll-area'

type InboxItem = {
  id: string
  profile_id?: string
  thread_id: string
  source?: string
  status: string
  title: string
  summary: string
  metadata?: {
    assistant?: { provider?: string; model?: string }
    item?: {
      id?: string
      title?: string
      part_number?: string
      href?: string
    }
    item_id?: string
    item_title?: string
    item_href?: string
  }
  created_at?: string
  updated_at?: string
}

function assistantThreadKey(profileId: string) {
  return `cabinet.assistant.workspace.thread.${profileId || 'local'}`
}

function assistantProviderKey(profileId: string) {
  return `cabinet.assistant.workspace.provider.${profileId || 'local'}`
}

function assistantModelKey(profileId: string) {
  return `cabinet.assistant.workspace.model.${profileId || 'local'}`
}

function normalizeInboxStatus(status: string) {
  const normalized = status.trim().toLowerCase()
  return normalized === 'queued' ? 'unread' : normalized
}

function formatDaysOld(value?: string) {
  if (!value) {
    return 'recent'
  }
  const created = new Date(value)
  if (Number.isNaN(created.getTime())) {
    return 'recent'
  }
  const diffMs = Date.now() - created.getTime()
  const days = Math.max(0, Math.floor(diffMs / 86_400_000))
  if (days === 0) {
    return 'today'
  }
  if (days === 1) {
    return '1 day old'
  }
  return `${days} days old`
}

function sourceLabel(source?: string) {
  const normalized = source?.trim().replace(/_/g, ' ')
  if (!normalized) {
    return 'Notification'
  }
  return normalized.replace(/\b\w/g, (letter) => letter.toUpperCase())
}

function itemLink(item: InboxItem) {
  const metadata = item.metadata
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
      [partNumber, title].filter(Boolean).join(' · ') ||
      title ||
      id ||
      'Open item',
  }
}

export function InboxWorkspacePanel() {
  const { activeProfileId, setActiveWorkspace } = useShellWorkspace()
  const [items, setItems] = useState<InboxItem[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [updatingItemId, setUpdatingItemId] = useState('')

  async function loadItems(profileId: string) {
    const response = await fetch(
      `/api/chat/inbox?profile_id=${encodeURIComponent(profileId)}`
    )
    if (!response.ok) {
      throw new Error('failed_to_load_inbox_items')
    }
    const payload = (await response.json()) as { items?: InboxItem[] }
    setItems(payload.items ?? [])
  }

  async function refreshInbox() {
    if (!activeProfileId) {
      return
    }
    setLoading(true)
    setError('')
    try {
      await loadItems(activeProfileId)
    } catch (err) {
      setError(
        err instanceof Error ? err.message : 'failed_to_load_inbox_items'
      )
    } finally {
      setLoading(false)
    }
  }

  function openChats() {
    window.location.assign('/chats')
  }

  async function updateInboxStatus(
    item: InboxItem,
    status: 'read' | 'unread' | 'archived'
  ) {
    if (!activeProfileId) {
      return
    }
    setUpdatingItemId(item.id)
    setError('')
    try {
      const response = await fetch(
        `/api/chat/inbox/${encodeURIComponent(item.id)}`,
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
        throw new Error('failed_to_update_inbox_item')
      }
      const updated = (await response.json()) as InboxItem
      setItems((current) => {
        if (normalizeInboxStatus(updated.status) === 'archived') {
          return current.filter((candidate) => candidate.id !== item.id)
        }
        return current.map((candidate) =>
          candidate.id === item.id ? { ...candidate, ...updated } : candidate
        )
      })
    } catch (err) {
      setError(
        err instanceof Error ? err.message : 'failed_to_update_inbox_item'
      )
    } finally {
      setUpdatingItemId('')
    }
  }

  useEffect(() => {
    let cancelled = false
    if (!activeProfileId) {
      return
    }
    setLoading(true)
    setError('')
    void (async () => {
      try {
        await loadItems(activeProfileId)
      } catch (err) {
        if (!cancelled) {
          setError(
            err instanceof Error ? err.message : 'failed_to_load_inbox_items'
          )
        }
      } finally {
        if (!cancelled) {
          setLoading(false)
        }
      }
    })()
    return () => {
      cancelled = true
    }
  }, [activeProfileId])

  function openInAssistant(item: InboxItem) {
    try {
      window.localStorage.setItem(
        assistantThreadKey(activeProfileId),
        item.thread_id
      )
      if (item.metadata?.assistant?.provider) {
        window.localStorage.setItem(
          assistantProviderKey(activeProfileId),
          item.metadata.assistant.provider
        )
      }
      if (item.metadata?.assistant?.model) {
        window.localStorage.setItem(
          assistantModelKey(activeProfileId),
          item.metadata.assistant.model
        )
      }
    } catch {
      // ignore storage issues
    }
    setActiveWorkspace('assistant')
  }

  const visibleItems = items.filter(
    (item) => normalizeInboxStatus(item.status) !== 'archived'
  )

  return (
    <div className='space-y-3 px-2 py-2' data-testid='shell-inbox-workspace'>
      <div className='rounded-md border bg-card p-3'>
        <div className='mb-2 flex items-center gap-2'>
          <Inbox className='h-4 w-4' />
          <h2 className='font-semibold'>Inbox</h2>
        </div>
        <p className='text-sm text-muted-foreground'>
          Notifications and asynchronous assistant outcomes will surface here as
          a simple catch-up list.
        </p>
      </div>
      <div className='rounded-md border bg-card p-3'>
        <div className='mb-2 flex items-center justify-between gap-2'>
          <div>
            <p className='text-sm font-medium'>Catch-up notifications</p>
            <p className='text-xs text-muted-foreground'>
              Read, archive, or jump to the linked item when one is available.
            </p>
          </div>
          <Button
            type='button'
            size='sm'
            variant='outline'
            data-testid='shell-inbox-refresh'
            disabled={loading}
            onClick={() => void refreshInbox()}
          >
            <RefreshCw className='h-3.5 w-3.5' />
            Refresh
          </Button>
        </div>
        <ScrollArea className='h-80 rounded-md border p-2'>
          <div className='space-y-2' data-testid='shell-inbox-item-list'>
            {loading ? (
              <p className='text-sm text-muted-foreground'>Loading Inbox...</p>
            ) : null}
            {!loading && visibleItems.length === 0 ? (
              <div className='space-y-3 text-sm text-muted-foreground'>
                <p>Caught up. No inbox items yet.</p>
                <div className='flex flex-wrap gap-2'>
                  <Button
                    type='button'
                    size='sm'
                    variant='outline'
                    data-testid='shell-inbox-open-chats'
                    onClick={openChats}
                  >
                    Open Chats
                  </Button>
                  <Button
                    type='button'
                    size='sm'
                    variant='outline'
                    data-testid='shell-inbox-open-assistant-workspace'
                    onClick={() => setActiveWorkspace('assistant')}
                  >
                    Open Assistant Workspace
                  </Button>
                </div>
              </div>
            ) : null}
            {visibleItems.map((item) => {
              const status = normalizeInboxStatus(item.status)
              const read = status === 'read'
              const link = itemLink(item)
              const updating = updatingItemId === item.id
              return (
                <article
                  key={item.id}
                  className={cn(
                    'rounded-lg border p-3 text-xs transition-colors',
                    read
                      ? 'bg-card/60 text-muted-foreground'
                      : 'bg-card text-foreground shadow-sm'
                  )}
                  data-testid='shell-inbox-notification-card'
                  data-status={status}
                >
                  <div className='flex items-start justify-between gap-2'>
                    <div className='min-w-0 space-y-1'>
                      <div className='flex flex-wrap items-center gap-2'>
                        {read ? (
                          <MailOpen className='h-3.5 w-3.5 text-muted-foreground' />
                        ) : (
                          <Mail className='h-3.5 w-3.5 text-primary' />
                        )}
                        <h3 className='font-semibold text-foreground'>
                          {item.title}
                        </h3>
                        <Badge
                          variant={read ? 'outline' : 'secondary'}
                          data-testid='shell-inbox-item-status'
                          className='capitalize'
                        >
                          {status}
                        </Badge>
                      </div>
                      <p className='text-muted-foreground'>{item.summary}</p>
                    </div>
                    <span className='shrink-0 text-muted-foreground'>
                      {formatDaysOld(item.created_at)}
                    </span>
                  </div>
                  <div className='mt-3 flex flex-wrap items-center gap-2 text-muted-foreground'>
                    <span>{sourceLabel(item.source)}</span>
                    {link ? (
                      <Button
                        type='button'
                        size='sm'
                        variant='link'
                        asChild
                        className='h-auto px-0 text-xs'
                      >
                        <a href={link.href} data-testid='shell-inbox-item-link'>
                          <ExternalLink className='h-3.5 w-3.5' />
                          {link.label}
                        </a>
                      </Button>
                    ) : null}
                  </div>
                  <div className='mt-3 flex flex-wrap gap-2'>
                    {read ? (
                      <Button
                        type='button'
                        size='sm'
                        variant='outline'
                        data-testid='shell-inbox-mark-unread'
                        disabled={updating}
                        onClick={() => void updateInboxStatus(item, 'unread')}
                      >
                        <Mail className='h-3.5 w-3.5' />
                        Mark unread
                      </Button>
                    ) : (
                      <Button
                        type='button'
                        size='sm'
                        variant='outline'
                        data-testid='shell-inbox-mark-read'
                        disabled={updating}
                        onClick={() => void updateInboxStatus(item, 'read')}
                      >
                        <MailOpen className='h-3.5 w-3.5' />
                        Mark read
                      </Button>
                    )}
                    <Button
                      type='button'
                      size='sm'
                      variant='outline'
                      data-testid='shell-inbox-archive'
                      disabled={updating}
                      onClick={() => void updateInboxStatus(item, 'archived')}
                    >
                      <Archive className='h-3.5 w-3.5' />
                      Archive
                    </Button>
                    <Button
                      type='button'
                      size='sm'
                      variant='outline'
                      data-testid='shell-inbox-open-assistant'
                      disabled={!item.thread_id}
                      onClick={() => openInAssistant(item)}
                    >
                      Open in Assistant
                    </Button>
                  </div>
                </article>
              )
            })}
          </div>
        </ScrollArea>
        {error ? (
          <p
            data-testid='shell-inbox-error'
            className='mt-2 text-xs text-destructive'
          >
            {error}
          </p>
        ) : null}
      </div>
    </div>
  )
}
