import { useCallback, useEffect, useMemo, useState } from 'react'
import {
  AssistantRuntimeProvider,
  type AppendMessage,
  useExternalStoreRuntime,
} from '@assistant-ui/react'
import {
  Bot,
  MessageCircle,
  MessagesSquare,
  Mic,
  Paperclip,
  PanelLeft,
  Plus,
  Search as SearchIcon,
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
import { Input } from '@/components/ui/input'
import { ScrollArea } from '@/components/ui/scroll-area'
import { ConfigDrawer } from '@/components/config-drawer'
import { LanguageSwitch } from '@/components/language-switch'
import { Header, HeaderTitle } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import {
  assistantAppendMessageText,
  CabinetAssistantUiComposer,
  CabinetAssistantUiMessageList,
  cabinetMessageToAssistantUi,
} from './assistant-ui-adapter'

type ChatThread = {
  id: string
  profile_id: string
  title: string
  created_at: string
  updated_at: string
}

type ChatMessage = {
  id: string
  profile_id: string
  thread_id: string
  role: 'user' | 'assistant' | 'system'
  content: string
  created_at: string
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

export function Chats() {
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
  const [loading, setLoading] = useState(true)
  const [messagesLoading, setMessagesLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [sendError, setSendError] = useState<string | null>(null)
  const [pendingAttachment, setPendingAttachment] = useState<File | null>(null)
  const [attachments, setAttachments] = useState<ChatAttachment[]>([])
  const [actionPartNumber, setActionPartNumber] = useState('CHAT-001')
  const [actionTitle, setActionTitle] = useState('Chat Created Item')
  const [actionMode, setActionMode] = useState<
    | 'create_inventory_item'
    | 'create_wishlist_entry'
    | 'update_inventory_item'
    | 'assign_collection_item'
  >('create_inventory_item')
  const [actionTargetItemID, setActionTargetItemID] = useState('')
  const [actionCollectionName, setActionCollectionName] = useState('Store 1')
  const [actionPreview, setActionPreview] = useState<ChatActionPreview | null>(
    null
  )
  const [applyResult, setApplyResult] = useState<ChatApplyResult | null>(null)
  const [applyNotice, setApplyNotice] = useState('')
  const [confirmApplyOpen, setConfirmApplyOpen] = useState(false)

  const selectedThread = useMemo(
    () => threads.find((thread) => thread.id === selectedThreadId) ?? null,
    [selectedThreadId, threads]
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

  const threadCreationDisabled = loading || Boolean(error) || !activeProfileId

  const loadMessages = useCallback(
    async (profileID: string, threadID: string) => {
      if (!profileID || !threadID) {
        setMessages([])
        return
      }
      setMessagesLoading(true)
      setSendError(null)
      try {
        const response = await fetch(
          `/api/chat/messages?profile_id=${encodeURIComponent(profileID)}&thread_id=${encodeURIComponent(threadID)}`
        )
        if (!response.ok) {
          throw new Error(`chat_messages_${response.status}`)
        }
        const payload = (await response.json()) as { messages?: ChatMessage[] }
        setMessages(payload.messages ?? [])
      } catch (err) {
        setSendError(
          err instanceof Error ? err.message : 'failed_to_load_chat_messages'
        )
        setMessages([])
      } finally {
        setMessagesLoading(false)
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
    clearStoredActionPreview(selectedActionPreviewStorageKey)
  }, [activeProfileId, selectedThreadId, selectedActionPreviewStorageKey])

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
    setThreadTitle('')
    await loadThreads(activeProfileId, false)
  }

  const sendMessageContent = useCallback(
    async (messageDraft: string) => {
      const content = messageDraft.trim()
      if (!activeProfileId || !selectedThreadId || !content) {
        return
      }
      setSendError(null)
      const response = await fetch('/api/chat/messages', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          profile_id: activeProfileId,
          thread_id: selectedThreadId,
          role: 'user',
          content,
          context: {
            route: { pathname: '/chats/' },
            profile: { id: activeProfileId },
            assistant: assistantDefaults,
          },
        }),
      })
      if (!response.ok) {
        setSendError(`chat_message_create_${response.status}`)
        return
      }
      await loadThreads(activeProfileId)
    },
    [activeProfileId, assistantDefaults, loadThreads, selectedThreadId]
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

  const uploadAttachment = async () => {
    if (!activeProfileId || !selectedThreadId || !pendingAttachment) {
      return
    }
    setSendError(null)
    const form = new FormData()
    form.set('profile_id', activeProfileId)
    form.set('thread_id', selectedThreadId)
    form.set('file', pendingAttachment)

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
    setPendingAttachment(null)
  }

  const previewCreateItemAction = async () => {
    if (!activeProfileId || !selectedThreadId || messages.length === 0) {
      return
    }
    setSendError(null)
    setApplyResult(null)
    setApplyNotice('')
    const response = await fetch('/api/chat/actions/preview', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        profile_id: activeProfileId,
        thread_id: selectedThreadId,
        action: actionMode,
        payload: {
          part_number: actionPartNumber.trim(),
          title: actionTitle.trim(),
          brand: 'AFX',
          category: 'General',
          item_id:
            actionMode === 'update_inventory_item' ||
            actionMode === 'assign_collection_item'
              ? actionTargetItemID.trim()
              : '',
          collection_name:
            actionMode === 'assign_collection_item'
              ? actionCollectionName.trim()
              : '',
          priority: actionMode === 'create_wishlist_entry' ? 'medium' : '',
          assistant_provider: assistantDefaults.provider,
          assistant_model: assistantDefaults.model,
        },
      }),
    })
    if (!response.ok) {
      setSendError(`chat_action_preview_${response.status}`)
      return
    }
    const preview = (await response.json()) as ChatActionPreview
    setActionPreview(preview)
    writeStoredActionPreview(selectedActionPreviewStorageKey, preview)
  }

  const applyPreviewAction = async () => {
    if (!activeProfileId || !selectedThreadId || !actionPreview?.id) {
      return
    }
    setSendError(null)
    const response = await fetch('/api/chat/actions/apply', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        profile_id: activeProfileId,
        thread_id: selectedThreadId,
        preview_id: actionPreview.id,
        confirm: true,
      }),
    })
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
    const response = await fetch('/api/chat/actions/cancel', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        profile_id: activeProfileId,
        thread_id: selectedThreadId,
        preview_id: actionPreview.id,
      }),
    })
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

  const previewDisabled =
    !selectedThreadId ||
    messages.length === 0 ||
    !actionPartNumber.trim() ||
    !actionTitle.trim() ||
    ((actionMode === 'update_inventory_item' ||
      actionMode === 'assign_collection_item') &&
      !actionTargetItemID.trim()) ||
    (actionMode === 'assign_collection_item' && !actionCollectionName.trim())

  return (
    <>
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

      <Main fixed className='overflow-hidden'>
        <div className='sr-only'>
          <h1>Chats</h1>
          <p data-testid='chat-workspace-description'>
            Persistent profile-scoped conversation threads backed by Cabinet
            runtime.
          </p>
          <p data-testid='chat-workspace-boundary-note'>
            Use Assistant for AI-guided help and actions; use Chats for durable
            conversation threads.
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
          className='grid h-full min-h-[620px] overflow-hidden border border-slate-800 bg-[#05060a] text-slate-100 shadow-2xl'
          style={{ gridTemplateColumns: '300px minmax(0, 1fr)' }}
          data-testid='chat-layout'
          data-visual-contract='assistant-ui-example'
        >
          <aside
            className='flex min-h-0 flex-col border-b border-slate-800 bg-[#070910] p-3 lg:border-e lg:border-b-0'
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
              onClick={() =>
                document
                  .querySelector<HTMLInputElement>(
                    '[data-testid="chat-new-thread-input"]'
                  )
                  ?.focus()
              }
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
            <div className='mb-3 flex gap-2'>
              <Input
                data-testid='chat-new-thread-input'
                placeholder='New thread title'
                value={threadTitle}
                onChange={(event) => setThreadTitle(event.target.value)}
                disabled={threadCreationDisabled}
                className='border-slate-800 bg-slate-900 text-slate-100 placeholder:text-slate-500'
              />
              <Button
                data-testid='chat-create-thread-button'
                onClick={() => void createThread()}
                disabled={threadCreationDisabled || !threadTitle.trim()}
                variant='outline'
                className='border-slate-700 bg-slate-900 text-slate-100 hover:bg-slate-800'
              >
                Create
              </Button>
            </div>
            <ScrollArea className='min-h-0 flex-1'>
              <div data-testid='chat-thread-list' className='space-y-1'>
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
                    onClick={() =>
                      document
                        .querySelector<HTMLInputElement>(
                          '[data-testid="chat-new-thread-input"]'
                        )
                        ?.focus()
                    }
                  >
                    Start a conversation
                  </Button>
                </div>
              </div>
            ) : (
              <>
                <div
                  className='flex min-h-[500px] flex-1 flex-col'
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
                        onClick={() =>
                          document
                            .querySelector<HTMLInputElement>(
                              '[data-testid="chat-new-thread-input"]'
                            )
                            ?.focus()
                        }
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
                      className='min-h-[360px] flex-1 rounded-none border-0 bg-transparent px-4 py-6'
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
                          className='flex h-full min-h-[320px] items-center justify-center'
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
                        }}
                      />
                    </ScrollArea>
                    <div
                      className='relative z-10 mx-auto mt-auto w-full max-w-3xl rounded-2xl border border-slate-800 bg-slate-950 p-3 shadow-xl'
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
                            {assistantDefaults.provider} /{' '}
                            {assistantDefaults.model}
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
                <div className='relative z-0 mt-4 space-y-3 rounded-md border border-slate-800 bg-slate-950 p-3'>
                  <p className='text-sm font-medium'>Attachments</p>
                  <div className='flex items-center gap-2'>
                    <Input
                      type='file'
                      data-testid='chat-attachment-input'
                      disabled={!selectedThreadId}
                      onChange={(event) =>
                        setPendingAttachment(event.target.files?.[0] ?? null)
                      }
                    />
                    <Button
                      type='button'
                      data-testid='chat-upload-attachment-button'
                      disabled={!selectedThreadId || !pendingAttachment}
                      onClick={() => void uploadAttachment()}
                    >
                      Upload
                    </Button>
                  </div>
                  <div
                    data-testid='chat-attachment-list'
                    className='space-y-1 text-sm'
                  >
                    {attachments.length === 0 ? (
                      <p className='text-muted-foreground'>
                        No attachments uploaded.
                      </p>
                    ) : (
                      attachments.map((attachment) => (
                        <p key={attachment.id}>{attachment.filename}</p>
                      ))
                    )}
                  </div>
                </div>

                <div
                  className='relative z-0 mt-3 space-y-3 rounded-md border border-slate-800/80 bg-slate-950/70 p-3 opacity-90'
                  data-testid='chat-tool-card-container'
                  data-visual-priority='secondary'
                >
                  <div className='flex flex-wrap items-start justify-between gap-2'>
                    <p className='text-sm font-medium'>Action Preview</p>
                    <p
                      className='rounded-md border bg-muted/30 px-2 py-1 text-xs text-muted-foreground'
                      data-testid='chat-assistant-defaults'
                    >
                      Assistant default: {assistantDefaults.provider} /{' '}
                      {assistantDefaults.model}
                    </p>
                  </div>
                  <label className='grid gap-1 text-sm'>
                    <span>Action Mode</span>
                    <select
                      data-testid='chat-preview-action-mode'
                      className='h-9 rounded-md border bg-background px-3'
                      value={actionMode}
                      onChange={(event) =>
                        setActionMode(
                          event.target.value as
                            | 'create_inventory_item'
                            | 'create_wishlist_entry'
                            | 'update_inventory_item'
                            | 'assign_collection_item'
                        )
                      }
                      disabled={!selectedThreadId}
                    >
                      <option value='create_inventory_item'>
                        create_inventory_item
                      </option>
                      <option value='create_wishlist_entry'>
                        create_wishlist_entry
                      </option>
                      <option value='update_inventory_item'>
                        update_inventory_item
                      </option>
                      <option value='assign_collection_item'>
                        assign_collection_item
                      </option>
                    </select>
                  </label>
                  {actionMode === 'update_inventory_item' ||
                  actionMode === 'assign_collection_item' ? (
                    <Input
                      data-testid='chat-preview-target-item-id'
                      value={actionTargetItemID}
                      onChange={(event) =>
                        setActionTargetItemID(event.target.value)
                      }
                      placeholder='Existing item ID'
                      disabled={!selectedThreadId}
                    />
                  ) : null}
                  {actionMode === 'assign_collection_item' ? (
                    <Input
                      data-testid='chat-preview-collection-name'
                      value={actionCollectionName}
                      onChange={(event) =>
                        setActionCollectionName(event.target.value)
                      }
                      placeholder='Collection name'
                      disabled={!selectedThreadId}
                    />
                  ) : null}
                  <div className='grid gap-2 sm:grid-cols-2'>
                    <Input
                      data-testid='chat-preview-part-number'
                      value={actionPartNumber}
                      onChange={(event) =>
                        setActionPartNumber(event.target.value)
                      }
                      placeholder='Part number'
                      disabled={!selectedThreadId}
                    />
                    <Input
                      data-testid='chat-preview-title'
                      value={actionTitle}
                      onChange={(event) => setActionTitle(event.target.value)}
                      placeholder='Item title'
                      disabled={!selectedThreadId}
                    />
                  </div>
                  <div className='flex flex-wrap gap-2'>
                    <Button
                      type='button'
                      data-testid='chat-preview-action-button'
                      onClick={() => void previewCreateItemAction()}
                      disabled={previewDisabled}
                    >
                      Preview Action
                    </Button>
                    <Button
                      type='button'
                      variant='outline'
                      data-testid='chat-apply-action-button'
                      onClick={() => setConfirmApplyOpen(true)}
                      disabled={
                        !selectedThreadId ||
                        !actionPreview?.id ||
                        actionPreviewStatusLabel !== 'pending'
                      }
                    >
                      Apply Action
                    </Button>
                  </div>
                  {actionPreview ? (
                    <p
                      data-testid='chat-action-preview'
                      className='text-sm text-muted-foreground'
                    >
                      Preview {actionPreview.id}: {actionPreview.action} (
                      {actionPreviewStatusLabel}) via{' '}
                      {String(
                        actionPreview.payload?.assistant_provider ?? 'openai'
                      )}{' '}
                      /{' '}
                      {String(
                        actionPreview.payload?.assistant_model ?? 'gpt-4o-mini'
                      )}
                      {actionPreviewTargetSummary
                        ? ` - ${actionPreviewTargetSummary}`
                        : ''}
                    </p>
                  ) : null}
                  {applyResult ? (
                    <p
                      data-testid='chat-action-apply-result'
                      className='text-sm'
                    >
                      {applyResultSummary}
                    </p>
                  ) : null}
                  {applyNotice ? (
                    <p
                      data-testid='chat-action-apply-notice'
                      className='text-sm text-muted-foreground'
                    >
                      {applyNotice}
                    </p>
                  ) : null}
                </div>
              </>
            )}
          </div>
        </section>
      </Main>

      <AlertDialog open={confirmApplyOpen} onOpenChange={setConfirmApplyOpen}>
        <AlertDialogContent data-testid='chat-apply-confirm-dialog'>
          <AlertDialogHeader>
            <AlertDialogTitle>Confirm Copilot Action</AlertDialogTitle>
            <AlertDialogDescription data-testid='chat-apply-confirm-summary'>
              {actionPreview
                ? `Apply ${actionPreview.action} with ${actionPreviewTargetSummary || `part_number=${String(actionPreview.payload?.part_number ?? 'n/a')} title=${String(actionPreview.payload?.title ?? 'n/a')}`} assistant=${String(actionPreview.payload?.assistant_provider ?? 'openai')}/${String(actionPreview.payload?.assistant_model ?? 'gpt-4o-mini')}`
                : 'No action preview selected.'}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel
              data-testid='chat-apply-confirm-cancel'
              onClick={() => void cancelPreviewApply()}
            >
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
      </AlertDialog>
    </>
  )
}
