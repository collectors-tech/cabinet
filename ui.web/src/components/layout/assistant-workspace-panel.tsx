import { useEffect, useMemo, useRef, useState } from 'react'
import { useRouterState } from '@tanstack/react-router'
import {
  Bot,
  CheckCircle2,
  ExternalLink,
  GitBranchPlus,
  RotateCcw,
  Send,
  ShieldAlert,
  Sparkles,
  UserRound,
  Wand2,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { useAuthStore } from '@/stores/auth-store'
import { useShellWorkspace } from '@/context/shell-workspace-provider'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Badge } from '@/components/ui/badge'
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
    assistant_handoff?: { status?: string; inbox_item_id?: string }
  }
}

type ActionPreview = {
  id: string
  action: string
  status: string
  payload?: { part_number?: string; title?: string }
}

type ApplyActionResult = {
  applied: boolean
  action: string
  item_id?: string
  wishlist_id?: string
  preview_id: string
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

function resultLink(result: ApplyActionResult | null) {
  if (!result) {
    return null
  }
  if (result.item_id) {
    return {
      href: `/inventory/?item=${encodeURIComponent(result.item_id)}`,
      label: 'Open item',
    }
  }
  if (result.wishlist_id) {
    return {
      href: `/wishlist/?item=${encodeURIComponent(result.wishlist_id)}`,
      label: 'Open wishlist item',
    }
  }
  return null
}

async function loadAssistantDefaultSettings(profileId: string) {
  const response = await fetch(`/api/profiles/${profileId}/settings`)
  if (!response.ok) {
    throw new Error(`profile_settings_${response.status}`)
  }
  const payload = (await response.json()) as {
    settings?: Record<string, string>
  }
  const settings = payload.settings ?? {}
  const nextProvider =
    settings.assistant_default_provider?.trim() || 'openai'
  const nextModel =
    settings.assistant_default_model?.trim() ||
    defaultModelForProvider(nextProvider)
  return { nextProvider, nextModel }
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
  const manualProviderModelChangeRef = useRef(false)
  const [actionPartNumber, setActionPartNumber] = useState('ASSIST-001')
  const [actionTitle, setActionTitle] = useState('Assistant Proposed Item')
  const [actionPreview, setActionPreview] = useState<ActionPreview | null>(null)
  const [executionState, setExecutionState] = useState<
    'idle' | 'queued' | 'running' | 'success' | 'failure'
  >('idle')
  const [applyResult, setApplyResult] = useState<ApplyActionResult | null>(null)
  const [confirmApplyOpen, setConfirmApplyOpen] = useState(false)
  const [permissionGuidance, setPermissionGuidance] = useState(
    'Read-only browsing is always allowed. Structured mutations are preview-first and confirmation-required before any apply call runs.'
  )

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
    options?: { semantics?: string; forkedFromThreadId?: string }
  ) {
    const semantics =
      options?.semantics?.trim() || 'assistant_workspace_session'
    const forkedFromThreadId = options?.forkedFromThreadId?.trim() || ''
    const metadata: ThreadMetadata = {
      provider: nextProvider,
      model: nextModel,
      thread_semantics: semantics,
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
    if (!activeProfileId) return

    setLoading(true)
    setError('')
    void (async () => {
      try {
        const { nextProvider: storedProvider, nextModel: storedModel } =
          await loadAssistantDefaultSettings(activeProfileId)
        try {
          window.localStorage.setItem(
            assistantProviderKey(activeProfileId),
            storedProvider
          )
          window.localStorage.setItem(
            assistantModelKey(activeProfileId),
            storedModel
          )
        } catch {
          // ignore storage issues
        }
        if (!cancelled && !manualProviderModelChangeRef.current) {
          setProvider(storedProvider)
          setModel(storedModel)
        }
        const ensuredThread = await ensureThread(
          activeProfileId,
          storedProvider,
          storedModel
        )
        if (cancelled) return
        const threads = await loadThreads(activeProfileId)
        const activeThread = threads.find(
          (thread) => thread.id === ensuredThread
        )
        if (
          activeThread?.metadata &&
          !cancelled &&
          !manualProviderModelChangeRef.current
        ) {
          setThreadMetadata(activeThread.metadata)
          setProvider(activeThread.metadata.provider || storedProvider)
          setModel(activeThread.metadata.model || storedModel)
        }
        if (manualProviderModelChangeRef.current) return
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
        if (!cancelled) setLoading(false)
      }
    })()

    return () => {
      cancelled = true
    }
  }, [activeProfileId])

  useEffect(() => {
    let cancelled = false
    if (!activeProfileId || loading || manualProviderModelChangeRef.current) {
      return
    }

    void (async () => {
      try {
        const { nextProvider, nextModel } =
          await loadAssistantDefaultSettings(activeProfileId)
        if (cancelled) return
        const threadSemantics = threadMetadata.thread_semantics || ''
        const allowsDefaultSync =
          threadSemantics === '' ||
          threadSemantics === 'assistant_workspace_session'
        if (!allowsDefaultSync) {
          return
        }
        if (
          (threadMetadata.provider && threadMetadata.provider !== provider) ||
          (threadMetadata.model && threadMetadata.model !== model)
        ) {
          return
        }
        if (provider === nextProvider && model === nextModel) {
          return
        }
        try {
          window.localStorage.setItem(
            assistantProviderKey(activeProfileId),
            nextProvider
          )
          window.localStorage.setItem(
            assistantModelKey(activeProfileId),
            nextModel
          )
        } catch {
          // ignore storage issues
        }
        setProvider(nextProvider)
        setModel(nextModel)
        setThreadMetadata((current) => ({
          ...current,
          provider: nextProvider,
          model: nextModel,
        }))
      } catch {
        // best-effort live sync only
      }
    })()

    return () => {
      cancelled = true
    }
  }, [
    activeProfileId,
    loading,
    location.pathname,
    location.search,
    provider,
    model,
    threadMetadata.thread_semantics,
  ])

  async function handleProviderChange(nextProvider: string) {
    if (!activeProfileId) return
    const nextModel = defaultModelForProvider(nextProvider)
    manualProviderModelChangeRef.current = true
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
        {
          semantics: 'fork_on_provider_model_change',
          forkedFromThreadId: previousThreadId,
        }
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
    if (!activeProfileId) return
    manualProviderModelChangeRef.current = true
    setModel(nextModel)
    setSending(true)
    setError('')
    try {
      const previousThreadId = threadId
      const newThreadId = await createAssistantThread(
        activeProfileId,
        provider,
        nextModel,
        {
          semantics: 'fork_on_provider_model_change',
          forkedFromThreadId: previousThreadId,
        }
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

  async function handleNewThread() {
    if (!activeProfileId) return
    manualProviderModelChangeRef.current = true
    setSending(true)
    setError('')
    try {
      const newThreadId = await createAssistantThread(
        activeProfileId,
        provider,
        model,
        {
          semantics: 'manual_new_thread',
        }
      )
      setDraft('')
      setActionPreview(null)
      setApplyResult(null)
      setExecutionState('idle')
      await loadMessages(activeProfileId, newThreadId)
    } catch (err) {
      setError(
        err instanceof Error ? err.message : 'failed_to_reset_assistant_thread'
      )
    } finally {
      setSending(false)
    }
  }

  async function sendMessage() {
    if (!activeProfileId || !threadId || !draft.trim()) return
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
            selection: { active_workspace_collection: selectionContext },
            assistant: { provider, model },
          },
        }),
      })
      if (!response.ok) throw new Error('failed_to_send_assistant_message')
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

  async function previewAction() {
    if (!activeProfileId || !threadId) return
    setExecutionState('queued')
    setError('')
    setApplyResult(null)
    setPermissionGuidance(
      'Structured mutations are preview-only until you explicitly confirm apply.'
    )
    try {
      const response = await fetch('/api/chat/actions/preview', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          profile_id: activeProfileId,
          thread_id: threadId,
          action: 'create_item_stub',
          payload: {
            part_number: actionPartNumber.trim(),
            title: actionTitle.trim(),
            brand: 'AFX',
            category: 'General',
          },
        }),
      })
      if (!response.ok) throw new Error(`assistant_preview_${response.status}`)
      const preview = (await response.json()) as ActionPreview
      setActionPreview(preview)
      setExecutionState('running')
    } catch (err) {
      setExecutionState('failure')
      setError(err instanceof Error ? err.message : 'assistant_preview_failed')
      setPermissionGuidance(
        'This action could not be previewed under the active policy. Read-only browsing remains available; mutation preview/apply may be unavailable.'
      )
    }
  }

  async function applyPreviewAction() {
    if (!activeProfileId || !threadId || !actionPreview?.id) return
    setExecutionState('running')
    setError('')
    try {
      const response = await fetch('/api/chat/actions/apply', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          profile_id: activeProfileId,
          thread_id: threadId,
          preview_id: actionPreview.id,
          confirm: true,
        }),
      })
      if (!response.ok) throw new Error(`assistant_apply_${response.status}`)
      const result = (await response.json()) as ApplyActionResult
      setApplyResult(result)
      setExecutionState('success')
      setConfirmApplyOpen(false)
    } catch (err) {
      setExecutionState('failure')
      setError(err instanceof Error ? err.message : 'assistant_apply_failed')
      setPermissionGuidance(
        'Apply is confirm-required. If apply remains blocked, the active policy may be preview-only for this action class.'
      )
    }
  }

  const applyLink = resultLink(applyResult)

  return (
    <div
      className='space-y-3 px-2 py-2'
      data-testid='shell-assistant-workspace'
    >
      <section
        className='overflow-hidden rounded-2xl border bg-card'
        data-testid='shell-assistant-codex-chat'
      >
        <div
          className='border-b bg-muted/20 p-3'
          data-testid='shell-chat-rail'
        >
          <div className='flex items-start justify-between gap-2'>
            <div className='min-w-0'>
              <div className='flex items-center gap-2'>
                <Sparkles className='h-4 w-4 text-primary' />
                <h2 className='font-semibold'>Assistant</h2>
              </div>
              <p className='mt-1 text-xs text-muted-foreground'>
                Route-aware agent for database work, evidence checks, and item
                links.
              </p>
            </div>
            <Button
              type='button'
              variant='outline'
              size='icon'
              data-testid='shell-assistant-new-thread'
              aria-label='New assistant thread'
              onClick={() => void handleNewThread()}
              disabled={loading || sending || !activeProfileId}
            >
              <RotateCcw className='h-3.5 w-3.5' />
            </Button>
          </div>

          <div className='mt-3 flex flex-wrap gap-2 text-[11px]'>
            <Badge
              variant='outline'
              data-testid='shell-assistant-context-chip'
              className='max-w-full justify-start gap-1 truncate'
            >
              <GitBranchPlus className='h-3 w-3 shrink-0' />
              <span
                className='truncate'
                data-testid='shell-assistant-route-context'
              >{`${routeContext.pathname}${routeContext.search}`}</span>
            </Badge>
            <Badge
              variant='secondary'
              data-testid='shell-assistant-model-chip'
              className='max-w-full justify-start gap-1 truncate'
            >
              <Bot className='h-3 w-3 shrink-0' />
              <span
                className='truncate'
                data-testid='shell-assistant-thread-provider'
              >
                {threadMetadata.provider || provider}
              </span>
              <span>/</span>
              <span
                className='truncate'
                data-testid='shell-assistant-thread-model'
              >
                {threadMetadata.model || model}
              </span>
            </Badge>
          </div>

          <div className='mt-3 grid grid-cols-2 gap-2 text-xs'>
            <label className='space-y-1'>
              <span className='font-medium text-foreground'>Provider</span>
              <select
                data-testid='shell-assistant-provider-select'
                className='w-full rounded-md border bg-background px-2 py-1'
                value={provider}
                onChange={(e) => void handleProviderChange(e.target.value)}
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
                onChange={(e) => void handleModelChange(e.target.value)}
              >
                {availableModels.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>
            </label>
          </div>

          <div className='mt-3 grid gap-1 text-[11px] text-muted-foreground'>
            <span>
              Profile:{' '}
              <span data-testid='shell-assistant-profile-scope'>
                {activeProfileId}
              </span>
            </span>
            <span data-testid='shell-assistant-selection-context'>
              Collection: {selectionContext}
            </span>
            <span data-testid='shell-assistant-thread-id'>
              {threadId || 'bootstrapping'}
            </span>
            <span data-testid='shell-assistant-boundary-note'>
              Thread continuity persists across authenticated route changes
              until an explicit reset boundary.
            </span>
            <span data-testid='shell-assistant-thread-semantics'>
              Provider/model changes fork a new assistant thread; manual reset
              creates a clean thread for the current profile.
            </span>
          </div>
        </div>

        <ScrollArea className='h-80 p-3'>
          <div
            className='space-y-4 pb-2'
            data-testid='shell-assistant-message-list'
          >
            {loading ? (
              <p className='text-sm text-muted-foreground'>
                Loading assistant workspace...
              </p>
            ) : null}
            {!loading && messages.length === 0 ? (
              <div className='rounded-2xl border border-dashed p-3 text-sm text-muted-foreground'>
                Ask Cabinet to update records, create drafts, search inventory,
                and return links to the items it touched.
              </div>
            ) : null}
            {messages.map((message) => {
              const isUser = message.role === 'user'
              return (
                <div
                  key={message.id}
                  className={cn('flex', isUser ? 'justify-end' : 'justify-start')}
                >
                  <div
                    className={cn(
                      'max-w-[92%] rounded-2xl px-3 py-2 text-sm leading-relaxed',
                      isUser
                        ? 'bg-muted text-foreground'
                        : 'bg-transparent px-0 text-muted-foreground'
                    )}
                    data-testid={
                      isUser
                        ? 'shell-assistant-message-bubble-user'
                        : 'shell-assistant-message-bubble-assistant'
                    }
                  >
                    <div className='mb-1 flex items-center gap-1.5 text-[11px] font-medium uppercase tracking-wide text-muted-foreground'>
                      {isUser ? (
                        <UserRound className='h-3 w-3' />
                      ) : (
                        <Bot className='h-3 w-3' />
                      )}
                      {isUser ? 'You' : 'Assistant'}
                    </div>
                    <p>{message.content}</p>
                  </div>
                </div>
              )
            })}

            <div
              className='rounded-2xl border bg-muted/10 p-3 text-xs'
              data-testid='shell-assistant-execution-panel'
            >
              <div className='flex items-center justify-between gap-2'>
                <div className='flex items-center gap-2 font-medium'>
                  <Wand2 className='h-4 w-4 text-primary' />
                  Agent actions
                </div>
                <Badge
                  variant='outline'
                  className='uppercase'
                  data-testid='shell-assistant-execution-state'
                >
                  {executionState}
                </Badge>
              </div>
              <p
                className='mt-2 text-muted-foreground'
                data-testid='shell-assistant-permission-guidance'
              >
                {permissionGuidance}
              </p>
              <div className='mt-3 grid gap-2'>
                <Input
                  data-testid='shell-assistant-preview-part-number'
                  value={actionPartNumber}
                  onChange={(e) => setActionPartNumber(e.target.value)}
                  placeholder='Part number'
                  disabled={!threadId || sending}
                />
                <Input
                  data-testid='shell-assistant-preview-title'
                  value={actionTitle}
                  onChange={(e) => setActionTitle(e.target.value)}
                  placeholder='Item title'
                  disabled={!threadId || sending}
                />
              </div>
              <div className='mt-3 flex flex-wrap gap-2'>
                <Button
                  type='button'
                  size='sm'
                  data-testid='shell-assistant-preview-action'
                  onClick={() => void previewAction()}
                  disabled={
                    !threadId ||
                    !actionPartNumber.trim() ||
                    !actionTitle.trim() ||
                    sending
                  }
                >
                  Preview
                </Button>
                <Button
                  type='button'
                  size='sm'
                  variant='outline'
                  data-testid='shell-assistant-apply-action'
                  onClick={() => setConfirmApplyOpen(true)}
                  disabled={!actionPreview?.id || sending}
                >
                  Apply
                </Button>
              </div>
              {actionPreview ? (
                <div
                  className='mt-3 rounded-xl border bg-background p-2'
                  data-testid='shell-assistant-action-card'
                >
                  <div
                    className='text-xs'
                    data-testid='shell-assistant-action-preview'
                  >
                    Preview {actionPreview.action} ({actionPreview.status}) for{' '}
                    {actionPreview.payload?.part_number} /{' '}
                    {actionPreview.payload?.title}
                  </div>
                </div>
              ) : null}
              {applyResult ? (
                <div
                  className='mt-3 rounded-xl border bg-background p-2'
                  data-testid='shell-assistant-apply-result'
                >
                  <div className='flex items-start gap-2'>
                    <CheckCircle2 className='mt-0.5 h-4 w-4 text-primary' />
                    <div className='min-w-0 flex-1'>
                      <p>
                        Applied {applyResult.action}{' '}
                        {applyResult.item_id ? `to ${applyResult.item_id}` : ''}
                      </p>
                      {applyLink ? (
                        <Button
                          type='button'
                          variant='link'
                          size='sm'
                          asChild
                          className='h-auto px-0 text-xs'
                        >
                          <a
                            href={applyLink.href}
                            data-testid='shell-assistant-result-link'
                          >
                            <ExternalLink className='h-3.5 w-3.5' />
                            {applyLink.label}
                          </a>
                        </Button>
                      ) : null}
                    </div>
                  </div>
                </div>
              ) : null}
              <div
                className='mt-3 flex items-start gap-2 rounded-xl border border-dashed p-2 text-muted-foreground'
                data-testid='shell-assistant-permission-boundary'
              >
                <ShieldAlert className='mt-0.5 h-4 w-4 shrink-0' />
                <div>
                  <p className='font-medium text-foreground'>
                    Permission boundary
                  </p>
                  <p>
                    Read-only is always allowed. Mutations are preview-first,
                    confirm-required, and may still be unavailable under the
                    active policy.
                  </p>
                </div>
              </div>
            </div>
          </div>
        </ScrollArea>

        {error ? (
          <p
            className='px-3 pb-2 text-xs text-destructive'
            data-testid='shell-assistant-error'
          >
            {error}
          </p>
        ) : null}

        <div className='border-t bg-background/80 p-3'>
          <div className='flex items-center gap-2 rounded-2xl border bg-muted/20 p-1'>
            <Input
              data-testid='shell-assistant-compose-input'
              value={draft}
              onChange={(e) => setDraft(e.target.value)}
              placeholder='Ask Cabinet to update, find, or link records...'
              disabled={!threadId || loading || sending}
              className='border-0 bg-transparent shadow-none focus-visible:ring-0'
            />
            <Button
              type='button'
              size='icon'
              data-testid='shell-assistant-send-button'
              aria-label='Send assistant message'
              onClick={() => void sendMessage()}
              disabled={!threadId || loading || sending || !draft.trim()}
            >
              <Send className='h-4 w-4' />
            </Button>
          </div>
        </div>
      </section>

      <AlertDialog open={confirmApplyOpen} onOpenChange={setConfirmApplyOpen}>
        <AlertDialogContent data-testid='shell-assistant-apply-confirm-dialog'>
          <AlertDialogHeader>
            <AlertDialogTitle>Confirm Assistant Action</AlertDialogTitle>
            <AlertDialogDescription data-testid='shell-assistant-apply-confirm-summary'>
              {actionPreview
                ? `Apply ${actionPreview.action} with part_number=${String(actionPreview.payload?.part_number ?? 'n/a')} title=${String(actionPreview.payload?.title ?? 'n/a')}`
                : 'No action preview selected.'}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel data-testid='shell-assistant-apply-cancel'>
              Cancel
            </AlertDialogCancel>
            <AlertDialogAction
              data-testid='shell-assistant-apply-confirm'
              onClick={(event) => {
                event.preventDefault()
                void applyPreviewAction()
              }}
            >
              Confirm Apply
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
