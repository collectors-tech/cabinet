import { useEffect, useState } from 'react'
import { Inbox } from 'lucide-react'
import { useShellWorkspace } from '@/context/shell-workspace-provider'
import { Button } from '@/components/ui/button'
import { ScrollArea } from '@/components/ui/scroll-area'

type InboxItem = {
  id: string
  thread_id: string
  status: string
  title: string
  summary: string
  metadata?: {
    assistant?: { provider?: string; model?: string }
  }
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

export function InboxWorkspacePanel() {
  const { activeProfileId, setActiveWorkspace } = useShellWorkspace()
  const [items, setItems] = useState<InboxItem[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

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

  return (
    <div className='space-y-3 px-2 py-2' data-testid='shell-inbox-workspace'>
      <div className='rounded-md border bg-card p-3'>
        <div className='mb-2 flex items-center gap-2'>
          <Inbox className='h-4 w-4' />
          <h2 className='font-semibold'>Inbox Workspace</h2>
        </div>
        <p className='text-sm text-muted-foreground'>
          Notifications and asynchronous assistant outcomes will surface here.
        </p>
      </div>
      <div className='rounded-md border bg-card p-3'>
        <p className='mb-2 text-sm font-medium'>Assistant outcomes</p>
        <ScrollArea className='h-56 rounded-md border p-2'>
          <div className='space-y-2' data-testid='shell-inbox-item-list'>
            {loading ? (
              <p className='text-sm text-muted-foreground'>Loading Inbox...</p>
            ) : null}
            {!loading && items.length === 0 ? (
              <p className='text-sm text-muted-foreground'>
                No inbox items yet.
              </p>
            ) : null}
            {items.map((item) => (
              <div
                key={item.id}
                className='rounded border p-2 text-xs'
                data-testid='shell-inbox-item'
              >
                <div className='flex items-center justify-between gap-2'>
                  <p className='font-medium'>{item.title}</p>
                  <span
                    data-testid='shell-inbox-item-status'
                    className='text-muted-foreground uppercase'
                  >
                    {item.status}
                  </span>
                </div>
                <p className='mt-1 text-muted-foreground'>{item.summary}</p>
                <Button
                  type='button'
                  size='sm'
                  variant='outline'
                  className='mt-2'
                  data-testid='shell-inbox-open-assistant'
                  onClick={() => openInAssistant(item)}
                >
                  Open in Assistant
                </Button>
              </div>
            ))}
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
