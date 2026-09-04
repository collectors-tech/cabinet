import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useNavigate } from '@tanstack/react-router'
import {
  AssistantRuntimeProvider,
  type AppendMessage,
  useExternalStoreRuntime,
} from '@assistant-ui/react'
import {
  Bot,
  ExternalLink,
  MessageCircle,
  MessagesSquare,
  Mic,
  Paperclip,
  PanelLeft,
  Plus,
  Search as SearchIcon,
  ShieldAlert,
  Share2,
  Sparkles,
} from 'lucide-react'
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
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { ScrollArea } from '@/components/ui/scroll-area'
import { ConfigDrawer } from '@/components/config-drawer'
import { LanguageSwitch } from '@/components/language-switch'
import { Header, HeaderTitle } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { cabinetProtectedFetch } from '@/lib/cabinet-session'
import {
  CabinetAssistantUiComposer,
  CabinetAssistantUiMessageList,
} from './assistant-ui-adapter'
import {
  assistantAppendMessageText,
  cabinetMessageToAssistantUi,
} from './assistant-ui-adapter-utils'
import {
  AgentResponseCards,
  type AgentCapabilitiesContext,
  type NormalizedAgentResponse,
  type AgentPlannerContext,
} from './agent-response-cards'
import {
  fetchChatWorkflowRuns,
  type ChatWorkflowRun,
  workflowRunResultSummary,
  workflowRunTimestamp,
} from './workflow-runs'

type ChatThread = {
  id: string
  profile_id: string
  title: string
  metadata?: {
    provider?: string
    model?: string
    thread_semantics?: string
  }
  created_at: string
  updated_at: string
}

type ChatMessage = {
  id: string
  profile_id: string
  thread_id: string
  role: 'user' | 'assistant' | 'system'
  content: string
  attachments_json?: Array<{
    id: string
    filename: string
    mime_type: string
    size_bytes: number
    provenance: string
    source: string
    created_at: string
  }> | string
  context?: {
    route?: { pathname?: string; search?: string }
    profile?: { id?: string }
    assistant?: { provider?: string; model?: string }
    app_control?: ChatAppControlContext
    agent_capabilities?: AgentCapabilitiesContext
    agent_planner?: AgentPlannerContext
	agent_response?: NormalizedAgentResponse
  }
  created_at: string
}

type ChatAppControlContext = {
  capability_id?: string
  policy?: string
  route?: string
  setup_needed?: boolean
}

type ChatAttachment = {
  id: string
  profile_id: string
  thread_id: string
  filename: string
  mime_type: string
  size_bytes: number
  path: string
  created_at: string
}

type ChatActionPreview = {
  id: string
  profile_id: string
  thread_id: string
  action: string
  payload?: Record<string, unknown>
  status: string
  created_at: string
  applied_at?: string
}

type ChatApplyResult = {
  applied: boolean
  action: string
  item_id?: string
  wishlist_id?: string
  collection_name?: string
  part_number?: string
  title?: string
  preview_id: string
}

type AssistantDefaults = {
  provider: string
  model: string
}

function actionPreviewStorageKey(profileID: string, threadID: string) {
  if (!profileID || !threadID) {
    return ''
  }
  return `cabinet.chat.action-preview.${profileID}.${threadID}`
}

function readStoredActionPreview(key: string) {
  if (!key || typeof window === 'undefined') {
    return null
  }
  try {
    const raw = window.sessionStorage.getItem(key)
    if (!raw) {
      return null
    }
    return JSON.parse(raw) as ChatActionPreview
  } catch {
    return null
  }
}

function writeStoredActionPreview(key: string, preview: ChatActionPreview) {
  if (!key || typeof window === 'undefined') {
    return
  }
  window.sessionStorage.setItem(key, JSON.stringify(preview))
}

function clearStoredActionPreview(key: string) {
  if (!key || typeof window === 'undefined') {
    return
  }
  window.sessionStorage.removeItem(key)
}

function threadInitial(title: string) {
  return title.trim().charAt(0).toUpperCase() || 'C'
}

const promptChips = ['Weather', 'Code', 'Write', 'Analyze', 'Brainstorm']

function labelForRoute(route: string) {
  const normalized = route.replace(/^\/+/, '').replace(/\/+$/, '')
  if (!normalized) {
    return 'Open Cabinet'
  }
  return `Open ${normalized
    .split('/')
    .filter(Boolean)
    .map((segment) => segment.charAt(0).toUpperCase() + segment.slice(1))
    .join(' / ')}`
}

function latestAppControl(messages: ChatMessage[]) {
  const latestAssistantMessage = [...messages]
	.reverse()
	.find((message) => message.role === 'assistant')
  return latestAssistantMessage?.context?.app_control
}

export function Chats() {
  const navigate = useNavigate()
  const [activeProfileId, setActiveProfileId] = useState('')
  const [assistantDefaults, setAssistantDefaults] = useState<AssistantDefaults>(
    {
      provider: 'openai',
      model: 'gpt-4o-mini',
    }
  )
  const [threads, setThreads] = useState<ChatThread[]>([])
  const [selectedThreadId, setSelectedThreadId] = useState('')
  const [threadSearch, setThreadSearch] = useState('')
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [threadTitle, setThreadTitle] = useState('')
  const [newThreadDialogOpen, setNewThreadDialogOpen] = useState(false)
  const [loading, setLoading] = useState(true)
  const [messagesLoading, setMessagesLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [sendError, setSendError] = useState<string | null>(null)
  const [attachments, setAttachments] = useState<ChatAttachment[]>([])
  const [actionPreview, setActionPreview] = useState<ChatActionPreview | null>(
    null
  )
  const [applyResult, setApplyResult] = useState<ChatApplyResult | null>(null)
  const [applyNotice, setApplyNotice] = useState('')
  const [confirmApplyOpen, setConfirmApplyOpen] = useState(false)
  const applyActionButtonRef = useRef<HTMLButtonElement>(null)
  const [workflowRuns, setWorkflowRuns] = useState<ChatWorkflowRun[]>([])
  const [workflowRunsLoading, setWorkflowRunsLoading] = useState(false)
  const [workflowRunsError, setWorkflowRunsError] = useState('')
  const [externalReviewPreviewId, setExternalReviewPreviewId] = useState('')

  const selectedThread = useMemo(
    () => threads.find((thread) => thread.id === selectedThreadId) ?? null,
    [selectedThreadId, threads]
  )
  const selectedAssistant = useMemo(
    () => ({
      provider:
        selectedThread?.metadata?.provider?.trim() ||
        assistantDefaults.provider,
      model: selectedThread?.metadata?.model?.trim() || assistantDefaults.model,
    }),
    [assistantDefaults, selectedThread]
  )
  const selectedActionPreviewStorageKey = useMemo(
    () => actionPreviewStorageKey(activeProfileId, selectedThreadId),
    [activeProfileId, selectedThreadId]
  )
  const filteredThreads = useMemo(() => {
    const query = threadSearch.trim().toLowerCase()
    if (!query) {
      return threads
    }
    return threads.filter((thread) =>
      thread.title.toLowerCase().includes(query)
    )
  }, [threadSearch, threads])

  const appControl = useMemo(() => latestAppControl(messages), [messages])
  const navigationRoute = appControl?.route?.trim() ?? ''
  const setupNeeded = Boolean(appControl?.setup_needed)

  const threadCreationDisabled = loading || Boolean(error) || !activeProfileId

  const loadMessages = useCallback(
    async (profileID: string, threadID: string) => {
      if (!profileID || !threadID) {
        setMessages([])
        setWorkflowRuns([])
        return
      }
      setMessagesLoading(true)
      setWorkflowRunsLoading(true)
      setSendError(null)
      try {
        const [response, runs] = await Promise.all([
          cabinetProtectedFetch(
            `/api/chat/messages?profile_id=${encodeURIComponent(profileID)}&thread_id=${encodeURIComponent(threadID)}`,
            profileID
          ),
          fetchChatWorkflowRuns(profileID, threadID),
        ])
        if (!response.ok) {
          throw new Error(`chat_messages_${response.status}`)
        }
        const payload = (await response.json()) as { messages?: ChatMessage[] }
        setMessages(payload.messages ?? [])
        setWorkflowRuns(runs)
        setWorkflowRunsError('')
      } catch (err) {
        setSendError(
          err instanceof Error ? err.message : 'failed_to_load_chat_messages'
        )
        setMessages([])
        setWorkflowRuns([])
        setWorkflowRunsError(
          err instanceof Error ? err.message : 'failed_to_load_workflow_runs'
        )
      } finally {
        setMessagesLoading(false)
        setWorkflowRunsLoading(false)
      }
    },
    []
  )

  const loadThreads = useCallback(
    async (
      profileID: string,
      preserveSelected = true,
      preferredThreadID = ''
    ) => {
      const response = await fetch(
        `/api/chat/threads?profile_id=${encodeURIComponent(profileID)}`
      )
      if (!response.ok) {
        throw new Error(`chat_threads_${response.status}`)
      }
      const payload = (await response.json()) as { threads?: ChatThread[] }
      const nextThreads = payload.threads ?? []
      setThreads(nextThreads)
      const requestedThread = preferredThreadID.trim()
      const nextSelected =
        requestedThread &&
        nextThreads.some((thread) => thread.id === requestedThread)
          ? requestedThread
          : preserveSelected &&
              nextThreads.some((thread) => thread.id === selectedThreadId)
            ? selectedThreadId
            : (nextThreads[0]?.id ?? '')
      setSelectedThreadId(nextSelected)
      await loadMessages(profileID, nextSelected)
    },
    [loadMessages, selectedThreadId]
  )

  const loadBootstrap = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const activeResponse = await fetch('/api/profiles/active')
      if (!activeResponse.ok) {
        throw new Error(`active_profile_${activeResponse.status}`)
      }
      const active = (await activeResponse.json()) as { id?: string }
      const profileID = active.id?.trim() ?? ''
      if (!profileID) {
        throw new Error('active_profile_missing')
      }
      setActiveProfileId(profileID)
      const settingsResponse = await fetch(
        `/api/profiles/${profileID}/settings`
      )
      if (settingsResponse.ok) {
        const settingsPayload = (await settingsResponse.json()) as {
          settings?: Record<string, string>
        }
        const settings = settingsPayload.settings ?? {}
        setAssistantDefaults({
          provider: settings.assistant_default_provider?.trim() || 'openai',
          model: settings.assistant_default_model?.trim() || 'gpt-4o-mini',
        })
      } else {
        setAssistantDefaults({ provider: 'openai', model: 'gpt-4o-mini' })
      }
      const requestedThread =
        typeof window !== 'undefined'
          ? (new URLSearchParams(window.location.search).get('thread_id') ?? '')
          : ''
      setExternalReviewPreviewId(
        typeof window !== 'undefined'
          ? (new URLSearchParams(window.location.search).get('preview_id') ??
              '')
          : ''
      )
      await loadThreads(profileID, false, requestedThread)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'failed_to_bootstrap_chat')
      setThreads([])
      setMessages([])
      setSelectedThreadId('')
    } finally {
      setLoading(false)
    }
  }, [loadThreads])

  useEffect(() => {
    void loadBootstrap()
  }, [loadBootstrap])

  useEffect(() => {
    if (!selectedThreadId || typeof window === 'undefined') {
      return
    }
    const search = new URLSearchParams(window.location.search)
    if (search.get('focus') !== 'composer') {
      return
    }
    const frame = window.requestAnimationFrame(() => {
      document
        .querySelector<HTMLElement>('[data-testid="chat-compose-input"]')
        ?.focus()
    })
    return () => window.cancelAnimationFrame(frame)
  }, [selectedThreadId])

  useEffect(() => {
    setAttachments([])
    setApplyResult(null)
    setApplyNotice('')
    setConfirmApplyOpen(false)
    const storedPreview = readStoredActionPreview(
      selectedActionPreviewStorageKey
    )
    if (
      storedPreview?.profile_id === activeProfileId &&
      storedPreview.thread_id === selectedThreadId &&
      (!storedPreview.status || storedPreview.status === 'previewed')
    ) {
      setActionPreview(storedPreview)
      return
    }
    setActionPreview(null)
    if (!externalReviewPreviewId.trim()) {
      clearStoredActionPreview(selectedActionPreviewStorageKey)
    }
  }, [
    activeProfileId,
    externalReviewPreviewId,
    selectedActionPreviewStorageKey,
    selectedThreadId,
  ])

  useEffect(() => {
    const previewID = externalReviewPreviewId.trim()
    if (!activeProfileId || !selectedThreadId || !previewID) {
      return
    }
    const controller = new AbortController()
    const loadReviewPreview = async () => {
      try {
        const response = await cabinetProtectedFetch(
          `/api/chat/actions/preview?profile_id=${encodeURIComponent(activeProfileId)}&preview_id=${encodeURIComponent(previewID)}`,
          activeProfileId,
          { signal: controller.signal }
        )
        if (!response.ok) {
          throw new Error(`chat_action_preview_${response.status}`)
        }
        const preview = (await response.json()) as ChatActionPreview
        if (
          preview.profile_id === activeProfileId &&
          preview.thread_id === selectedThreadId
        ) {
          setActionPreview(preview)
          writeStoredActionPreview(selectedActionPreviewStorageKey, preview)
        }
      } catch (err) {
        if (!controller.signal.aborted) {
          setSendError(
            err instanceof Error
              ? err.message
              : 'failed_to_load_chat_action_preview'
          )
        }
      }
    }
    void loadReviewPreview()
    return () => controller.abort()
  }, [
    activeProfileId,
    externalReviewPreviewId,
    selectedActionPreviewStorageKey,
    selectedThreadId,
  ])

  const createThread = async () => {
    const title = threadTitle.trim()
    if (!activeProfileId || !title) {
      return
    }
    setSendError(null)
    const response = await fetch('/api/chat/threads', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        profile_id: activeProfileId,
        title,
      }),
    })
    if (!response.ok) {
      setSendError(`chat_thread_create_${response.status}`)
      return
    }
    const createdThread = (await response.json()) as ChatThread
    setThreadTitle('')
    setNewThreadDialogOpen(false)
    await loadThreads(activeProfileId, false, createdThread.id)
  }

  const sendMessageContent = useCallback(
    async (messageDraft: string) => {
      const content = messageDraft.trim()
      if (!activeProfileId || !selectedThreadId || !content) {
        return
      }
      setSendError(null)
      const response = await cabinetProtectedFetch(
        '/api/chat/messages',
        activeProfileId,
        {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          profile_id: activeProfileId,
          thread_id: selectedThreadId,
          role: 'user',
          content,
          attachment_ids: attachments.map((attachment) => attachment.id),
          agent_context: {
            profile_id: activeProfileId,
            thread_id: selectedThreadId,
            route_id: '/chats/',
            surface_id: 'chats.main',
            source_channel: 'in-app',
            permission_state: 'ask_before_local_changes',
            setup_state: 'ready',
            intent_text: content,
          },
          context: {
            route: { pathname: '/chats/' },
            profile: { id: activeProfileId },
            assistant: selectedAssistant,
          },
        }),
        }
      )
      if (!response.ok) {
        setSendError(`chat_message_create_${response.status}`)
        return
      }
      setAttachments([])
      await loadThreads(activeProfileId)
    },
    [
      activeProfileId,
      attachments,
      loadThreads,
      selectedAssistant,
      selectedThreadId,
    ]
  )

  const handleAssistantUiNewMessage = useCallback(
    async (message: AppendMessage) => {
      await sendMessageContent(assistantAppendMessageText(message))
    },
    [sendMessageContent]
  )

  const chatAssistantRuntime = useExternalStoreRuntime<ChatMessage>({
    messages,
    convertMessage: cabinetMessageToAssistantUi,
    isLoading: messagesLoading,
    isRunning: false,
    isSendDisabled: !activeProfileId || !selectedThreadId,
    onNew: handleAssistantUiNewMessage,
    adapters: {
      threadList: {
        threadId: selectedThreadId,
        isLoading: loading,
        threads: threads.map((thread) => ({
          id: thread.id,
          title: thread.title,
          status: 'regular' as const,
        })),
        onSwitchToNewThread: () => {
          document
            .querySelector<HTMLInputElement>(
              '[data-testid="chat-new-thread-input"]'
            )
            ?.focus()
        },
        onSwitchToThread: async (threadId) => {
          setSelectedThreadId(threadId)
          await loadMessages(activeProfileId, threadId)
        },
      },
    },
  })

  const uploadAttachment = async (file: File | null | undefined) => {
    if (!activeProfileId || !selectedThreadId || !file) {
      return
    }
    setSendError(null)
    const form = new FormData()
    form.set('profile_id', activeProfileId)
    form.set('thread_id', selectedThreadId)
    form.set('file', file)

    const response = await fetch('/api/chat/attachments', {
      method: 'POST',
      body: form,
    })
    if (!response.ok) {
      setSendError(`chat_attachment_upload_${response.status}`)
      return
    }
    const attachment = (await response.json()) as ChatAttachment
    setAttachments((current) => [attachment, ...current])
  }

  const applyPreviewAction = async () => {
    if (!activeProfileId || !selectedThreadId || !actionPreview?.id) {
      return
    }
    setSendError(null)
    const response = await cabinetProtectedFetch(
      '/api/chat/actions/apply',
      activeProfileId,
      {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        profile_id: activeProfileId,
        thread_id: selectedThreadId,
        preview_id: actionPreview.id,
        confirm: true,
      }),
      }
    )
    if (!response.ok) {
      setSendError(`chat_action_apply_${response.status}`)
      setApplyResult(null)
      setApplyNotice('Action apply failed; preview remains pending.')
      setConfirmApplyOpen(false)
      return
    }
    const result = (await response.json()) as ChatApplyResult
    setApplyResult(result)
    setApplyNotice('')
    setConfirmApplyOpen(false)
    clearStoredActionPreview(selectedActionPreviewStorageKey)
    await loadMessages(activeProfileId, selectedThreadId)
  }

  const cancelPreviewApply = async () => {
    if (!activeProfileId || !selectedThreadId || !actionPreview?.id) {
      setConfirmApplyOpen(false)
      return
    }
    setSendError(null)
    const response = await cabinetProtectedFetch(
      '/api/chat/actions/cancel',
      activeProfileId,
      {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        profile_id: activeProfileId,
        thread_id: selectedThreadId,
        preview_id: actionPreview.id,
      }),
      }
    )
    if (!response.ok) {
      setSendError(`chat_action_cancel_${response.status}`)
      setApplyNotice('Action cancel failed; preview remains pending.')
      setConfirmApplyOpen(false)
      return
    }
    const result = (await response.json()) as ChatApplyResult
    setConfirmApplyOpen(false)
    setApplyResult(null)
    setActionPreview((current) =>
      current ? { ...current, status: 'cancelled' } : current
    )
    clearStoredActionPreview(selectedActionPreviewStorageKey)
    setApplyNotice(
      result.applied
        ? 'Action cancel returned an unexpected applied result.'
        : 'Action apply canceled; no mutation applied.'
    )
    await loadMessages(activeProfileId, selectedThreadId)
  }

  const applyResultSummary = (() => {
    if (!applyResult) {
      return ''
    }
    const previewPart =
      typeof actionPreview?.payload?.part_number === 'string'
        ? actionPreview.payload.part_number
        : ''
    const base = `Applied ${applyResult.action}`
    const withPart = previewPart ? `${base} (${previewPart})` : base
    const changedFields = [
      applyResult.part_number ? `part_number=${applyResult.part_number}` : '',
      applyResult.title ? `title=${applyResult.title}` : '',
    ]
      .filter(Boolean)
      .join(' ')
    if (applyResult.wishlist_id) {
      return `${withPart} to wishlist ${applyResult.wishlist_id}`
    }
    if (applyResult.collection_name && applyResult.item_id) {
      return `${withPart} to collection ${applyResult.collection_name} for item ${applyResult.item_id}`
    }
    if (applyResult.item_id) {
      return changedFields
        ? `${withPart} to item ${applyResult.item_id} with ${changedFields}`
        : `${withPart} to item ${applyResult.item_id}`
    }
    return withPart
  })()

  const actionPreviewTargetSummary = (() => {
    if (!actionPreview) {
      return ''
    }
    const payload = actionPreview.payload ?? {}
    const targetItem =
      typeof payload.item_id === 'string' && payload.item_id.trim()
        ? payload.item_id.trim()
        : ''
    const collection =
      typeof payload.collection_name === 'string' &&
      payload.collection_name.trim()
        ? payload.collection_name.trim()
        : ''
    const partNumber =
      typeof payload.part_number === 'string' && payload.part_number.trim()
        ? payload.part_number.trim()
        : ''
    const title =
      typeof payload.title === 'string' && payload.title.trim()
        ? payload.title.trim()
        : ''
    return [
      targetItem ? `target=${targetItem}` : '',
      collection ? `collection=${collection}` : '',
      partNumber ? `part_number=${partNumber}` : '',
      title ? `title=${title}` : '',
    ]
      .filter(Boolean)
      .join(' ')
  })()

  const actionPreviewStatusLabel =
    !actionPreview?.status || actionPreview.status === 'previewed'
      ? 'pending'
      : actionPreview.status

  return (
    <AlertDialog open={confirmApplyOpen} onOpenChange={setConfirmApplyOpen}>
      <Header>
        <Search />
        <HeaderTitle
          title='Chats'
          description='Persistent profile-scoped conversation threads backed by Cabinet runtime.'
          icon={MessagesSquare}
          testId='chats-header-title'
          iconTestId='chats-page-icon'
        />
        <div
          className='ms-auto flex items-center space-x-4'
          data-header-title-avoid='true'
        >
          <LanguageSwitch />
          <ThemeSwitch />
          <ConfigDrawer />
          <ProfileDropdown />
        </div>
      </Header>

      <Main fixed className='min-h-0 overflow-hidden px-2 py-2 sm:px-4 sm:py-4'>
        <div className='sr-only'>
          <h1>Chats</h1>
          <p data-testid='chat-workspace-description'>
            Persistent profile-scoped conversation threads backed by Cabinet
            runtime.
          </p>
          <p data-testid='chat-workspace-boundary-note'>
            Cabinet Agent keeps the same governed conversation, context, and
            action reviews in this full workspace and the contextual panel.
          </p>
        </div>

        {error ? (
          <div
            className='rounded-md border border-destructive/40 bg-destructive/10 p-3 text-sm'
            data-testid='chat-bootstrap-error'
          >
            <p className='font-medium'>Chat service is unavailable.</p>
            <p className='mt-1 text-muted-foreground'>{error}</p>
            <Button
              variant='outline'
              size='sm'
              className='mt-3'
              onClick={() => void loadBootstrap()}
            >
              Retry
            </Button>
          </div>
        ) : null}

        <section
          className='grid h-full min-h-0 grid-cols-1 overflow-hidden border border-slate-800 bg-[#05060a] text-slate-100 shadow-2xl lg:grid-cols-[300px_minmax(0,1fr)]'
          data-testid='chat-layout'
          data-visual-contract='assistant-ui-example'
        >
          <aside
            className='hidden min-h-0 flex-col border-b border-slate-800 bg-[#070910] p-3 lg:flex lg:border-e lg:border-b-0'
            data-testid='chat-conversation-rail'
          >
            <div className='mb-4 flex items-center gap-3'>
              <div className='flex h-9 w-9 items-center justify-center rounded-md bg-cyan-400 text-slate-950'>
                <Sparkles className='h-4 w-4' />
              </div>
              <div className='min-w-0'>
                <p className='truncate text-sm font-semibold'>Cabinet Agent</p>
                <p className='truncate text-xs text-slate-400'>
                  Threads and governed actions
                </p>
              </div>
            </div>
            <Button
              type='button'
              data-testid='chat-new-thread-action'
              className='mb-3 w-full justify-start gap-2 bg-slate-100 text-slate-950 hover:bg-white'
              onClick={() => setNewThreadDialogOpen(true)}
              disabled={threadCreationDisabled}
            >
              <Plus className='h-4 w-4' />
              New Thread
            </Button>
            <div className='relative mb-3'>
              <SearchIcon className='pointer-events-none absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-slate-500' />
              <Input
                data-testid='chat-conversation-search'
                placeholder='Search messages'
                value={threadSearch}
                onChange={(event) => setThreadSearch(event.target.value)}
                className='border-slate-800 bg-slate-900 ps-9 text-slate-100 placeholder:text-slate-500'
              />
            </div>
            <ScrollArea className='min-h-0 flex-1'>
              <div data-testid='chat-thread-list' className='min-h-px space-y-1'>
                {threads.length === 0 && !loading ? (
                  <p className='rounded-md border border-dashed border-slate-800 p-3 text-sm text-slate-500'>
                    No chat threads yet.
                  </p>
                ) : null}
                {filteredThreads.map((thread) => (
                  <button
                    key={thread.id}
                    type='button'
                    data-testid='chat-thread-item'
                    className={`flex w-full items-center gap-3 rounded-md border px-3 py-2 text-left text-sm transition ${
                      selectedThreadId === thread.id
                        ? 'border-cyan-400/60 bg-cyan-400/10 text-cyan-50'
                        : 'border-slate-800 bg-slate-900/70 text-slate-200 hover:bg-slate-800'
                    }`}
                    onClick={() => {
                      setSelectedThreadId(thread.id)
                      void loadMessages(activeProfileId, thread.id)
                    }}
                  >
                    <span
                      className='flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-slate-800 text-xs font-semibold text-cyan-200'
                      data-testid='chat-thread-avatar'
                    >
                      {threadInitial(thread.title)}
                    </span>
                    <span className='min-w-0 flex-1'>
                      <span className='block truncate font-medium'>
                        {thread.title}
                      </span>
                      <span
                        className='block truncate text-xs text-slate-500'
                        data-testid='chat-thread-preview'
                      >
                        No messages yet
                      </span>
                    </span>
                  </button>
                ))}
              </div>
            </ScrollArea>
          </aside>

          <div className='flex min-h-0 flex-col bg-[#0f1117] p-3'>
            {!selectedThread ? (
              <div
                className='flex min-h-[520px] flex-1 items-center justify-center rounded-lg border border-slate-800 bg-slate-950'
                data-testid='chat-empty-workspace-state'
              >
                <div className='mx-auto max-w-sm text-center'>
                  <div className='mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full border border-slate-700 bg-slate-900'>
                    <MessageCircle className='h-6 w-6 text-cyan-300' />
                  </div>
                  <h2 className='text-xl font-semibold text-slate-100'>
                    How can I help you today?
                  </h2>
                  <p className='mt-2 text-sm text-slate-400'>
                    Choose an existing thread or create a new one to continue a
                    durable Cabinet conversation.
                  </p>
                  <Button
                    type='button'
                    className='mt-4'
                    data-testid='chat-empty-workspace-action'
                    onClick={() => setNewThreadDialogOpen(true)}
                    disabled={threadCreationDisabled}
                  >
                    Start a conversation
                  </Button>
                </div>
              </div>
            ) : (
              <>
                <div
                  className='flex min-h-0 flex-1 flex-col overflow-hidden [@media(max-height:500px)]:overflow-y-auto'
                  data-testid='chat-main-surface'
                >
                  <div
                    className='mb-3 flex items-center justify-between gap-3 border-b border-slate-800/80 bg-transparent px-2 pb-3'
                    data-testid='chat-main-topbar'
                  >
                    <div className='min-w-0'>
                      <div className='flex items-center gap-2 text-xs text-slate-500'>
                        <PanelLeft className='h-3.5 w-3.5' />
                        Assistant-ui workspace
                      </div>
                      <h2
                        className='truncate font-semibold text-slate-100'
                        data-testid='chat-thread-title'
                      >
                        {selectedThread.title}
                      </h2>
                    </div>
                    <div className='flex items-center gap-2'>
                      <Button
                        type='button'
                        size='sm'
                        data-testid='chat-new-chat-button'
                        className='gap-2 bg-slate-100 text-slate-950 hover:bg-white'
                        onClick={() => setNewThreadDialogOpen(true)}
                        disabled={threadCreationDisabled}
                      >
                        <Plus className='h-4 w-4' />
                        New Chat
                      </Button>
                      <Button
                        type='button'
                        size='icon'
                        variant='outline'
                        data-testid='chat-share-export-button'
                        aria-label='Share or export chat'
                        className='border-slate-700 bg-slate-900 text-slate-100 hover:bg-slate-800'
                      >
                        <Share2 className='h-4 w-4' />
                      </Button>
                    </div>
                  </div>
                  <AssistantRuntimeProvider runtime={chatAssistantRuntime}>
                    <ScrollArea
                      className='min-h-0 flex-1 rounded-none border-0 bg-transparent px-3 py-3 sm:px-4 sm:py-6 [@media(max-height:500px)]:min-h-32 [@media(max-height:500px)]:shrink-0'
                      data-testid='chat-main-canvas'
                    >
                      {messagesLoading ? (
                        <p className='text-sm text-slate-400'>
                          Loading messages...
                        </p>
                      ) : null}
                      {!messagesLoading &&
                      selectedThread &&
                      messages.length === 0 ? (
                        <div
                          className='flex h-full min-h-0 items-center justify-center'
                          data-testid='chat-empty-thread-state'
                        >
                          <div className='text-center'>
                            <div className='mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full border border-slate-700 bg-slate-900'>
                              <Sparkles className='h-5 w-5 text-cyan-300' />
                            </div>
                            <p className='text-xl font-semibold text-slate-100'>
                              How can I help you today?
                            </p>
                            <p className='mt-2 text-sm text-slate-400'>
                              No messages in this thread yet.
                            </p>
                          </div>
                        </div>
                      ) : null}
                      <CabinetAssistantUiMessageList
                        messages={messages}
                        testIds={{
                          root: 'chat-message-list',
                          messagePrimitive:
                            'chat-assistant-ui-message-primitive',
                          userBubble: 'chat-message-bubble-user',
                          assistantBubble: 'chat-message-bubble-assistant',
                          attachment: 'chat-message-attachment',
                        }}
                      />
                      <AgentResponseCards
                        messages={messages}
                        testIDPrefix='chat'
						onRetry={sendMessageContent}
						onApply={() => setConfirmApplyOpen(true)}
						onAction={(response) => {
						  const route = response.next_action?.route?.trim()
						  if (route) void navigate({ to: route })
						}}
                        onPreviewStateChanged={() =>
                          loadMessages(activeProfileId, selectedThreadId)
                        }
                      />
                      {navigationRoute ? (
                        <div
                          className='rounded-md border border-slate-800 bg-slate-900 p-3 text-sm text-slate-100'
                          data-testid='chat-navigation-action'
                        >
                          <div className='flex items-start gap-2'>
                            <ExternalLink className='mt-0.5 h-4 w-4 text-cyan-300' />
                            <div className='min-w-0 flex-1'>
                              <p className='font-medium'>
                                {labelForRoute(navigationRoute)}
                              </p>
                              <p
                                className='mt-1 text-xs text-slate-400'
                                data-testid='chat-navigation-reason'
                              >
                                Cabinet planned this as a read-only navigation
                                action from the chat thread.
                              </p>
                              <Button
                                type='button'
                                size='sm'
                                variant='outline'
                                className='mt-2 border-slate-700 bg-slate-950 text-slate-100 hover:bg-slate-800'
                                data-testid='chat-navigation-action-open'
                                onClick={() =>
                                  void navigate({ to: navigationRoute })
                                }
                              >
                                <ExternalLink className='h-3.5 w-3.5' />
                                Open screen
                              </Button>
                            </div>
                          </div>
                        </div>
                      ) : null}
                      {setupNeeded ? (
                        <div
                          className='rounded-md border border-amber-500/40 bg-amber-950/30 p-3 text-sm text-amber-100'
                          data-testid='chat-setup-needed-guidance'
                        >
                          <div className='flex items-start gap-2'>
                            <ShieldAlert className='mt-0.5 h-4 w-4 text-amber-300' />
                            <p>
                              Provider setup is needed before Cabinet can run
                              this assistant action.
                            </p>
                          </div>
                        </div>
                      ) : null}
                      {actionPreview ? (
                        <div
                          className='rounded-md border border-slate-800 bg-slate-900 p-3 text-sm text-slate-100'
                          data-testid='chat-action-preview'
                        >
                          <div className='flex flex-wrap items-start justify-between gap-2'>
                            <div>
                              <p className='font-medium'>
                                Preview {actionPreview.id}:{' '}
                                {actionPreview.action}
                              </p>
                              <p className='mt-1 text-xs text-slate-400'>
                                {actionPreviewStatusLabel} via{' '}
                                {String(
                                  actionPreview.payload?.assistant_provider ??
                                    'openai'
                                )}{' '}
                                /{' '}
                                {String(
                                  actionPreview.payload?.assistant_model ??
                                    'gpt-4o-mini'
                                )}
                                {actionPreviewTargetSummary
                                  ? ` - ${actionPreviewTargetSummary}`
                                  : ''}
                              </p>
                            </div>
                            <div className='flex flex-wrap gap-2'>
                              <Button
                                ref={applyActionButtonRef}
                                type='button'
                                size='sm'
                                variant='outline'
                                className='border-slate-700 bg-slate-950 text-slate-100 hover:bg-slate-800'
                                data-testid='chat-apply-action-button'
                                onClick={() => setConfirmApplyOpen(true)}
                                onKeyDown={(event) => {
                                  if (
                                    !event.repeat &&
                                    (event.key === 'Enter' || event.key === ' ')
                                  ) {
                                    event.preventDefault()
                                    event.currentTarget.click()
                                  }
                                }}
                                disabled={
                                  !selectedThreadId ||
                                  !actionPreview.id ||
                                  actionPreviewStatusLabel !== 'pending'
                                }
                              >
                                Apply
                              </Button>
                              <Button
                                type='button'
                                size='sm'
                                variant='ghost'
                                className='text-slate-300 hover:bg-slate-800 hover:text-slate-100'
                                data-testid='chat-cancel-action-button'
                                onClick={() => void cancelPreviewApply()}
                                disabled={
                                  !selectedThreadId ||
                                  !actionPreview.id ||
                                  actionPreviewStatusLabel !== 'pending'
                                }
                              >
                                Cancel
                              </Button>
                            </div>
                          </div>
                          {applyResult ? (
                            <p
                              data-testid='chat-action-apply-result'
                              className='mt-2 text-xs text-slate-300'
                            >
                              {applyResultSummary}
                            </p>
                          ) : null}
                          {applyNotice ? (
                            <p
                              data-testid='chat-action-apply-notice'
                              className='mt-2 text-xs text-slate-400'
                            >
                              {applyNotice}
                            </p>
                          ) : null}
                        </div>
                      ) : null}
                      <div
                        className='mt-3 rounded-md border border-slate-800 bg-slate-900 p-3 text-sm text-slate-100'
                        data-testid='chat-action-timeline'
                      >
                        <div className='mb-2 flex items-center justify-between gap-2'>
                          <p className='font-medium'>Action Timeline</p>
                          <span
                            className='rounded-md border border-slate-700 bg-slate-950 px-2 py-1 text-xs text-slate-400 uppercase'
                            data-testid='chat-action-timeline-count'
                          >
                            {workflowRuns.length} records
                          </span>
                        </div>
                        {workflowRunsLoading ? (
                          <p className='text-xs text-slate-400'>
                            Loading workflow runs...
                          </p>
                        ) : null}
                        {!workflowRunsLoading && workflowRuns.length === 0 ? (
                          <p
                            className='text-xs text-slate-400'
                            data-testid='chat-action-timeline-empty'
                          >
                            Durable workflow records appear here after Cabinet
                            plans, previews, applies, cancels, or fails an
                            assistant action.
                          </p>
                        ) : null}
                        {workflowRunsError && workflowRuns.length === 0 ? (
                          <p
                            className='text-xs text-red-300'
                            data-testid='chat-action-timeline-error'
                          >
                            {workflowRunsError}
                          </p>
                        ) : null}
                        <div className='space-y-2'>
                          {workflowRuns.map((run) => (
                            <div
                              key={run.id}
                              className='rounded border border-slate-800 bg-slate-950 px-3 py-2'
                              data-testid='chat-action-timeline-run'
                              data-workflow-status={run.status}
                              data-capability-id={run.capability_id}
                            >
                              <div className='flex flex-wrap items-center justify-between gap-2'>
                                <span className='font-medium text-slate-100'>
                                  {run.capability_id}
                                </span>
                                <span className='rounded border border-slate-700 px-2 py-0.5 text-[11px] text-cyan-200 uppercase'>
                                  {run.status}
                                </span>
                              </div>
                              <p className='mt-1 text-xs text-slate-400'>
                                {run.workflow_id} / {run.confirmation_state}
                              </p>
                              <p className='mt-1 text-xs text-slate-300'>
                                {workflowRunResultSummary(run)}
                              </p>
                              <p className='mt-1 text-[11px] text-slate-500'>
                                {workflowRunTimestamp(run)}
                              </p>
                            </div>
                          ))}
                        </div>
                      </div>
                    </ScrollArea>
                    <div
                      className='relative z-10 mx-auto mt-auto max-h-[55vh] w-full max-w-3xl shrink-0 overflow-y-auto rounded-2xl border border-slate-800 bg-slate-950 p-3 shadow-xl sm:max-h-none sm:overflow-visible'
                      data-testid='chat-composer-shell'
                      data-position='bottom-center'
                    >
                      <div className='mb-3 flex flex-wrap items-center justify-between gap-2'>
                        <div
                          className='flex flex-wrap gap-2'
                          data-testid='chat-prompt-chips'
                        >
                          {promptChips.map((chip) => (
                            <button
                              key={chip}
                              type='button'
                              className='rounded-full border border-slate-700 px-3 py-1 text-xs text-slate-300 hover:border-cyan-400/70 hover:text-cyan-100'
                              disabled={!selectedThreadId}
                            >
                              {chip}
                            </button>
                          ))}
                        </div>
                        <div
                          className='flex items-center gap-2 text-xs text-slate-400'
                          data-testid='chat-model-selector-row'
                        >
                          <Bot className='h-3.5 w-3.5 text-cyan-300' />
                          <span data-testid='chat-model-selector'>
                            {selectedAssistant.provider} /{' '}
                            {selectedAssistant.model}
                          </span>
                        </div>
                      </div>
                      <CabinetAssistantUiComposer
                        composer={{
                          disabled: !selectedThreadId,
                          sending: false,
                        }}
                        placeholder='Send a message... (@ to mention, / for commands)'
                        testIds={{
                          root: 'chat-assistant-ui-composer-primitive',
                          input: 'chat-compose-input',
                          sendButton: 'chat-send-button',
                        }}
                      />
                      <Input
                        type='file'
                        data-testid='chat-attachment-input'
                        className='sr-only'
                        disabled={!selectedThreadId}
                        onChange={(event) => {
                          const file = event.target.files?.[0]
                          event.target.value = ''
                          void uploadAttachment(file)
                        }}
                      />
                      <div className='mt-2 flex items-center justify-between gap-2'>
                        <Button
                          type='button'
                          size='sm'
                          variant='ghost'
                          className='gap-2 text-slate-300 hover:bg-slate-800 hover:text-slate-100'
                          data-testid='chat-composer-attachment-button'
                          onClick={() =>
                            document
                              .querySelector<HTMLInputElement>(
                                '[data-testid="chat-attachment-input"]'
                              )
                              ?.click()
                          }
                          disabled={!selectedThreadId}
                        >
                          <Paperclip className='h-4 w-4' />
                          Attach
                        </Button>
                        <Button
                          type='button'
                          size='icon'
                          variant='ghost'
                          className='text-slate-300 hover:bg-slate-800 hover:text-slate-100'
                          data-testid='chat-voice-control'
                          aria-label='Voice input unavailable'
                          disabled
                        >
                          <Mic className='h-4 w-4' />
                        </Button>
                      </div>
                      {attachments.length > 0 ? (
                        <div
                          className='mt-2 space-y-2'
                          data-testid='chat-attachment-list'
                        >
                          {attachments.map((attachment) => (
                            <div
                              key={attachment.id}
                              className='flex items-center justify-between rounded-lg border border-slate-700 bg-slate-900 px-3 py-2 text-xs text-slate-300'
                            >
                              <span>{attachment.filename}</span>
                              <Button
                                type='button'
                                size='sm'
                                variant='ghost'
                                data-testid='chat-remove-attachment-button'
                                onClick={() =>
                                  setAttachments((current) =>
                                    current.filter(
                                      (item) => item.id !== attachment.id
                                    )
                                  )
                                }
                              >
                                Remove
                              </Button>
                            </div>
                          ))}
                        </div>
                      ) : null}
                    </div>
                  </AssistantRuntimeProvider>
                </div>
                {sendError ? (
                  <p
                    className='mt-2 text-sm text-destructive'
                    data-testid='chat-send-error'
                  >
                    {sendError}
                  </p>
                ) : null}
              </>
            )}
          </div>
        </section>
      </Main>

      <AlertDialogContent
        data-testid='chat-apply-confirm-dialog'
        onCloseAutoFocus={(event) => {
          event.preventDefault()
          applyActionButtonRef.current?.focus()
        }}
      >
          <AlertDialogHeader>
            <AlertDialogTitle>Confirm Copilot Action</AlertDialogTitle>
            <AlertDialogDescription data-testid='chat-apply-confirm-summary'>
              {actionPreview
                ? `Apply ${actionPreview.action} with ${actionPreviewTargetSummary || `part_number=${String(actionPreview.payload?.part_number ?? 'n/a')} title=${String(actionPreview.payload?.title ?? 'n/a')}`} assistant=${String(actionPreview.payload?.assistant_provider ?? 'openai')}/${String(actionPreview.payload?.assistant_model ?? 'gpt-4o-mini')}`
                : 'No action preview selected.'}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel data-testid='chat-apply-confirm-cancel'>
              Cancel
            </AlertDialogCancel>
            <AlertDialogAction
              data-testid='chat-apply-confirm-submit'
              onClick={(event) => {
                event.preventDefault()
                void applyPreviewAction()
              }}
            >
              Confirm Apply
            </AlertDialogAction>
          </AlertDialogFooter>
      </AlertDialogContent>
      <Dialog
        open={newThreadDialogOpen}
        onOpenChange={(open) => {
          setNewThreadDialogOpen(open)
          if (!open) {
            setThreadTitle('')
          }
        }}
      >
        <DialogContent data-testid='chat-new-thread-dialog'>
          <DialogHeader>
            <DialogTitle>Start a new Chat</DialogTitle>
            <DialogDescription>
              Name the conversation so Cabinet can keep its messages, context,
              attachments, and governed actions together.
            </DialogDescription>
          </DialogHeader>
          <form
            className='space-y-4'
            onSubmit={(event) => {
              event.preventDefault()
              void createThread()
            }}
          >
            <Input
              data-testid='chat-new-thread-input'
              aria-label='Chat title'
              autoFocus
              placeholder='For example, organise my latest purchases'
              value={threadTitle}
              onChange={(event) => setThreadTitle(event.target.value)}
              disabled={threadCreationDisabled}
            />
            <DialogFooter>
              <Button
                type='submit'
                data-testid='chat-create-thread-button'
                disabled={threadCreationDisabled || !threadTitle.trim()}
              >
                Create Chat
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </AlertDialog>
  )
}
