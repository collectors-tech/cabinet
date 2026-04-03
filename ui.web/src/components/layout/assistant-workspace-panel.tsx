import { useEffect, useMemo, useState } from 'react'
import { useRouterState } from '@tanstack/react-router'
import { Send } from 'lucide-react'
import { useAuthStore } from '@/stores/auth-store'
import { useShellWorkspace } from '@/context/shell-workspace-provider'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { ScrollArea } from '@/components/ui/scroll-area'

type Thread = {
  id: string
  title: string
}

type Message = {
  id: string
  role: string
  content: string
  context?: {
    route?: { pathname?: string; search?: string }
    profile?: { id?: string }
    selection?: { active_workspace_collection?: string }
  }
}

function activeCollectionKey(profileScope: string) {
  return `cabinet.workspace.collections.active.${profileScope || 'local'}`
}

function assistantThreadKey(profileId: string) {
  return `cabinet.assistant.workspace.thread.${profileId || 'local'}`
}

export function AssistantWorkspacePanel() {
  const { activeProfileId } = useShellWorkspace()
  const authUser = useAuthStore((state) => state.auth.user)
  const location = useRouterState({
    select: (state) => ({
      pathname: state.location.pathname,
      search: state.location.searchStr,
    }),
  })
  const profileScope = useMemo(
    () => authUser?.email || authUser?.accountNo || 'local',
    [authUser?.accountNo, authUser?.email]
  )
  const [threadId, setThreadId] = useState('')
  const [messages, setMessages] = useState<Message[]>([])
  const [draft, setDraft] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [sending, setSending] = useState(false)

  const routeContext = useMemo(
    () => ({
      pathname: location.pathname || '/',
      search: location.search || '',
    }),
    [location.pathname, location.search]
  )

  const selectionContext = useMemo(() => {
    try {
      return (
        window.localStorage.getItem(activeCollectionKey(profileScope)) ||
        'All Items'
      )
    } catch {
      return 'All Items'
    }
  }, [profileScope, messages.length, location.pathname, location.search])

  async function ensureThread(profileId: string) {
    const storageKey = assistantThreadKey(profileId)
    let nextThreadID = ''
    try {
      nextThreadID = window.localStorage.getItem(storageKey) || ''
    } catch {
      nextThreadID = ''
    }

    if (!nextThreadID) {
      const createResp = await fetch('/api/chat/threads', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          profile_id: profileId,
          title: 'Assistant Workspace',
        }),
      })
      if (!createResp.ok) {
        throw new Error('failed_to_create_assistant_thread')
      }
      const created = (await createResp.json()) as Thread
      nextThreadID = created.id
      try {
        window.localStorage.setItem(storageKey, nextThreadID)
      } catch {
        // ignore storage issues
      }
    }

    setThreadId(nextThreadID)
    return nextThreadID
  }

  async function loadMessages(profileId: string, targetThreadId: string) {
    const resp = await fetch(
      `/api/chat/messages?profile_id=${encodeURIComponent(profileId)}&thread_id=${encodeURIComponent(targetThreadId)}`
    )
    if (!resp.ok) {
      throw new Error('failed_to_load_assistant_messages')
    }
    const payload = (await resp.json()) as { messages?: Message[] }
    setMessages(payload.messages ?? [])
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
        const ensuredThread = await ensureThread(activeProfileId)
        if (cancelled) {
          return
        }
        await loadMessages(activeProfileId, ensuredThread)
      } catch (err) {
        if (!cancelled) {
          setError(
            err instanceof Error
              ? err.message
              : 'assistant_workspace_bootstrap_failed'
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

  async function sendMessage() {
    if (!activeProfileId || !threadId || !draft.trim()) {
      return
    }
    setSending(true)
    setError('')
    try {
      const response = await fetch('/api/chat/messages', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          profile_id: activeProfileId,
          thread_id: threadId,
          role: 'user',
          content: draft.trim(),
          context: {
            route: routeContext,
            profile: { id: activeProfileId },
            selection: {
              active_workspace_collection: selectionContext,
            },
          },
        }),
      })
      if (!response.ok) {
        throw new Error('failed_to_send_assistant_message')
      }
      setDraft('')
      await loadMessages(activeProfileId, threadId)
    } catch (err) {
      setError(
        err instanceof Error ? err.message : 'failed_to_send_assistant_message'
      )
    } finally {
      setSending(false)
    }
  }

  return (
    <div
      className='space-y-3 px-2 py-2'
      data-testid='shell-assistant-workspace'
    >
      <div
        className='rounded-md border bg-card p-3'
        data-testid='shell-chat-rail'
      >
        <h2 className='font-semibold'>Assistant Workspace</h2>
        <p className='mt-2 text-sm text-muted-foreground'>
          Persistent route-aware helper workspace for guided actions.
        </p>
        <dl className='mt-3 space-y-2 text-xs text-muted-foreground'>
          <div>
            <dt className='font-medium text-foreground'>Profile scope</dt>
            <dd data-testid='shell-assistant-profile-scope'>
              {activeProfileId}
            </dd>
          </div>
          <div>
            <dt className='font-medium text-foreground'>Current route</dt>
            <dd data-testid='shell-assistant-route-context'>{`${routeContext.pathname}${routeContext.search}`}</dd>
          </div>
          <div>
            <dt className='font-medium text-foreground'>Selection</dt>
            <dd data-testid='shell-assistant-selection-context'>
              {selectionContext}
            </dd>
          </div>
        </dl>
        <p
          className='mt-3 text-xs text-muted-foreground'
          data-testid='shell-assistant-boundary-note'
        >
          Thread continuity persists across authenticated route changes until an
          explicit reset boundary.
        </p>
      </div>

      <div className='rounded-md border bg-card p-3'>
        <div className='mb-2 flex items-center justify-between gap-2'>
          <p className='text-sm font-medium'>Assistant Thread</p>
          <span
            className='text-xs text-muted-foreground'
            data-testid='shell-assistant-thread-id'
          >
            {threadId || 'bootstrapping'}
          </span>
        </div>
        <ScrollArea className='h-44 rounded-md border p-2'>
          <div className='space-y-2' data-testid='shell-assistant-message-list'>
            {loading ? (
              <p className='text-sm text-muted-foreground'>
                Loading assistant workspace...
              </p>
            ) : null}
            {!loading && messages.length === 0 ? (
              <p className='text-sm text-muted-foreground'>
                No assistant messages yet.
              </p>
            ) : null}
            {messages.map((message) => (
              <div key={message.id} className='rounded border p-2 text-xs'>
                <p className='font-medium text-muted-foreground uppercase'>
                  {message.role}
                </p>
                <p>{message.content}</p>
              </div>
            ))}
          </div>
        </ScrollArea>
        {error ? (
          <p
            className='mt-2 text-xs text-destructive'
            data-testid='shell-assistant-error'
          >
            {error}
          </p>
        ) : null}
        <div className='mt-3 flex gap-2'>
          <Input
            data-testid='shell-assistant-compose-input'
            value={draft}
            onChange={(event) => setDraft(event.target.value)}
            placeholder='Ask Assistant about the current route...'
            disabled={!threadId || loading || sending}
          />
          <Button
            data-testid='shell-assistant-send-button'
            onClick={() => void sendMessage()}
            disabled={!threadId || loading || sending || !draft.trim()}
          >
            <Send className='mr-1 h-4 w-4' />
            Send
          </Button>
        </div>
      </div>
    </div>
  )
}
