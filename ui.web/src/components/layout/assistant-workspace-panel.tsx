import { useEffect, useMemo, useState } from 'react'
import { useRouterState } from '@tanstack/react-router'
import { GitBranchPlus, Send } from 'lucide-react'
import { useAuthStore } from '@/stores/auth-store'
import { useShellWorkspace } from '@/context/shell-workspace-provider'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { ScrollArea } from '@/components/ui/scroll-area'

type ThreadMetadata = {
  provider?: string
  model?: string
  thread_semantics?: string
  forked_from_thread_id?: string
}

type Thread = {
  id: string
  title: string
  metadata?: ThreadMetadata
}

type Message = {
  id: string
  role: string
  content: string
  context?: {
    route?: { pathname?: string; search?: string }
    profile?: { id?: string }
    selection?: { active_workspace_collection?: string }
    assistant?: { provider?: string; model?: string }
  }
}

type AssistantProviderOption = {
  provider: string
  label: string
  models: { value: string; label: string }[]
}

const assistantProviderOptions: AssistantProviderOption[] = [
  {
    provider: 'openai',
    label: 'OpenAI',
    models: [
      { value: 'gpt-4o-mini', label: 'gpt-4o-mini' },
      { value: 'gpt-4.1-mini', label: 'gpt-4.1-mini' },
    ],
  },
  {
    provider: 'anthropic',
    label: 'Anthropic',
    models: [
      { value: 'claude-3-5-haiku', label: 'claude-3-5-haiku' },
      { value: 'claude-3-7-sonnet', label: 'claude-3-7-sonnet' },
    ],
  },
]

function activeCollectionKey(profileScope: string) {
  return `cabinet.workspace.collections.active.${profileScope || 'local'}`
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

function defaultModelForProvider(provider: string) {
  return (
    assistantProviderOptions.find((option) => option.provider === provider)
      ?.models[0]?.value || 'gpt-4o-mini'
  )
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
  const [threadMetadata, setThreadMetadata] = useState<ThreadMetadata>({})
  const [messages, setMessages] = useState<Message[]>([])
  const [draft, setDraft] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [sending, setSending] = useState(false)
  const [provider, setProvider] = useState('openai')
  const [model, setModel] = useState('gpt-4o-mini')

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

  const availableModels = useMemo(
    () =>
      assistantProviderOptions.find((option) => option.provider === provider)
        ?.models || [],
    [provider]
  )

  async function createAssistantThread(
    profileId: string,
    nextProvider: string,
    nextModel: string,
    forkedFromThreadId = ''
  ) {
    const metadata: ThreadMetadata = {
      provider: nextProvider,
      model: nextModel,
      thread_semantics: 'fork_on_provider_model_change',
    }
    if (forkedFromThreadId.trim()) {
      metadata.forked_from_thread_id = forkedFromThreadId.trim()
    }
    const createResp = await fetch('/api/chat/threads', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        profile_id: profileId,
        title: `Assistant Workspace (${nextProvider} / ${nextModel})`,
        metadata,
      }),
    })
    if (!createResp.ok) {
      throw new Error('failed_to_create_assistant_thread')
    }
    const created = (await createResp.json()) as Thread
    setThreadId(created.id)
    setThreadMetadata(created.metadata ?? metadata)
    try {
      window.localStorage.setItem(assistantThreadKey(profileId), created.id)
      window.localStorage.setItem(assistantProviderKey(profileId), nextProvider)
      window.localStorage.setItem(assistantModelKey(profileId), nextModel)
    } catch {
      // ignore storage issues
    }
    return created.id
  }

  async function ensureThread(
    profileId: string,
    nextProvider: string,
    nextModel: string
  ) {
    const storageKey = assistantThreadKey(profileId)
    let nextThreadID = ''
    try {
      nextThreadID = window.localStorage.getItem(storageKey) || ''
    } catch {
      nextThreadID = ''
    }

    if (!nextThreadID) {
      nextThreadID = await createAssistantThread(
        profileId,
        nextProvider,
        nextModel
      )
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

  async function loadThreads(profileId: string) {
    const resp = await fetch(
      `/api/chat/threads?profile_id=${encodeURIComponent(profileId)}`
    )
    if (!resp.ok) {
      throw new Error('failed_to_load_assistant_threads')
    }
    const payload = (await resp.json()) as { threads?: Thread[] }
    return payload.threads ?? []
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
        const storedProvider =
          window.localStorage.getItem(assistantProviderKey(activeProfileId)) ||
          'openai'
        const storedModel =
          window.localStorage.getItem(assistantModelKey(activeProfileId)) ||
          defaultModelForProvider(storedProvider)
        if (!cancelled) {
          setProvider(storedProvider)
          setModel(storedModel)
        }
        const ensuredThread = await ensureThread(
          activeProfileId,
          storedProvider,
          storedModel
        )
        if (cancelled) {
          return
        }
        const threads = await loadThreads(activeProfileId)
        const activeThread = threads.find(
          (thread) => thread.id === ensuredThread
        )
        if (activeThread?.metadata && !cancelled) {
          setThreadMetadata(activeThread.metadata)
          setProvider(activeThread.metadata.provider || storedProvider)
          setModel(activeThread.metadata.model || storedModel)
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

  async function handleProviderChange(nextProvider: string) {
    if (!activeProfileId) {
      return
    }
    const nextModel = defaultModelForProvider(nextProvider)
    setProvider(nextProvider)
    setModel(nextModel)
    setSending(true)
    setError('')
    try {
      const previousThreadId = threadId
      const newThreadId = await createAssistantThread(
        activeProfileId,
        nextProvider,
        nextModel,
        previousThreadId
      )
      await loadMessages(activeProfileId, newThreadId)
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : 'failed_to_change_assistant_provider'
      )
    } finally {
      setSending(false)
    }
  }

  async function handleModelChange(nextModel: string) {
    if (!activeProfileId) {
      return
    }
    setModel(nextModel)
    setSending(true)
    setError('')
    try {
      const previousThreadId = threadId
      const newThreadId = await createAssistantThread(
        activeProfileId,
        provider,
        nextModel,
        previousThreadId
      )
      await loadMessages(activeProfileId, newThreadId)
    } catch (err) {
      setError(
        err instanceof Error ? err.message : 'failed_to_change_assistant_model'
      )
    } finally {
      setSending(false)
    }
  }

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
            assistant: {
              provider,
              model,
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
        <div className='mt-3 grid grid-cols-2 gap-2 text-xs'>
          <label className='space-y-1'>
            <span className='font-medium text-foreground'>Provider</span>
            <select
              data-testid='shell-assistant-provider-select'
              className='w-full rounded-md border bg-background px-2 py-1'
              value={provider}
              onChange={(event) =>
                void handleProviderChange(event.target.value)
              }
            >
              {assistantProviderOptions.map((option) => (
                <option key={option.provider} value={option.provider}>
                  {option.label}
                </option>
              ))}
            </select>
          </label>
          <label className='space-y-1'>
            <span className='font-medium text-foreground'>Model</span>
            <select
              data-testid='shell-assistant-model-select'
              className='w-full rounded-md border bg-background px-2 py-1'
              value={model}
              onChange={(event) => void handleModelChange(event.target.value)}
            >
              {availableModels.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          </label>
        </div>
        <p
          className='mt-3 text-xs text-muted-foreground'
          data-testid='shell-assistant-boundary-note'
        >
          Thread continuity persists across authenticated route changes until an
          explicit reset boundary.
        </p>
        <p
          className='mt-2 text-xs text-muted-foreground'
          data-testid='shell-assistant-thread-semantics'
        >
          Provider/model changes fork a new assistant thread and record the
          resulting provider/model in thread metadata.
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
        <div className='mb-2 flex items-center gap-2 text-xs text-muted-foreground'>
          <GitBranchPlus className='h-3.5 w-3.5' />
          <span data-testid='shell-assistant-thread-provider'>
            {threadMetadata.provider || provider}
          </span>
          <span>/</span>
          <span data-testid='shell-assistant-thread-model'>
            {threadMetadata.model || model}
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
