import { useCallback, useEffect, useMemo, useState } from 'react'
import {
  MessageCircle,
  MessagesSquare,
  Plus,
  Search as SearchIcon,
  Send,
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
import { Separator } from '@/components/ui/separator'
import { ConfigDrawer } from '@/components/config-drawer'
import { LanguageSwitch } from '@/components/language-switch'
import { Header, HeaderTitle } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'

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

function prettyRole(role: ChatMessage['role']) {
  if (role === 'assistant') {
    return 'Assistant'
  }
  if (role === 'system') {
    return 'System'
  }
  return 'You'
}

function threadInitial(title: string) {
  return title.trim().charAt(0).toUpperCase() || 'C'
}

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
  const [draft, setDraft] = useState('')
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
  const filteredThreads = useMemo(() => {
    const query = threadSearch.trim().toLowerCase()
    if (!query) {
      return threads
    }
    return threads.filter((thread) =>
      thread.title.toLowerCase().includes(query)
    )
  }, [threadSearch, threads])
  const selectedActionPreviewStorageKey = useMemo(
    () => actionPreviewStorageKey(activeProfileId, selectedThreadId),
    [activeProfileId, selectedThreadId]
  )

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

  const sendMessage = async () => {
    const content = draft.trim()
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
      }),
    })
    if (!response.ok) {
      setSendError(`chat_message_create_${response.status}`)
      return
    }
    setDraft('')
    await loadThreads(activeProfileId)
  }

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

      <Main fixed>
        <div className='flex items-center gap-2'>
          <h1 className='text-2xl font-bold tracking-tight'>Chats</h1>
          <MessagesSquare className='h-5 w-5 text-muted-foreground' />
        </div>
        <p
          className='text-muted-foreground'
          data-testid='chat-workspace-description'
        >
          Persistent profile-scoped conversation threads backed by Cabinet
          runtime.
        </p>
        <p
          className='text-sm text-muted-foreground'
          data-testid='chat-workspace-boundary-note'
        >
          Use Assistant for AI-guided help and actions; use Chats for durable
          conversation threads.
        </p>
        <Separator className='my-4' />

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
          className='grid h-full min-h-[620px] overflow-hidden rounded-md border bg-zinc-950 text-zinc-100 lg:grid-cols-[320px_minmax(0,1fr)]'
          data-testid='chat-layout'
        >
          <aside
            className='flex min-h-0 flex-col border-b border-zinc-800 bg-zinc-950 lg:border-r lg:border-b-0'
            data-testid='chat-conversation-rail'
          >
            <div className='space-y-3 border-b border-zinc-800 p-3'>
              <label className='relative block'>
                <SearchIcon className='pointer-events-none absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-zinc-500' />
                <Input
                  data-testid='chat-conversation-search'
                  placeholder='Search messages'
                  value={threadSearch}
                  onChange={(event) => setThreadSearch(event.target.value)}
                  className='h-9 border-zinc-800 bg-zinc-900 pl-9 text-sm text-zinc-100 placeholder:text-zinc-500'
                />
              </label>
              <div className='flex gap-2'>
                <Input
                  data-testid='chat-new-thread-input'
                  placeholder='New thread title'
                  value={threadTitle}
                  onChange={(event) => setThreadTitle(event.target.value)}
                  disabled={threadCreationDisabled}
                  className='h-9 border-zinc-800 bg-zinc-900 text-zinc-100 placeholder:text-zinc-500'
                />
                <Button
                  data-testid='chat-create-thread-button'
                  onClick={() => void createThread()}
                  disabled={threadCreationDisabled || !threadTitle.trim()}
                  size='icon'
                  aria-label='Create thread'
                  className='h-9 w-9 shrink-0'
                >
                  <Plus className='h-4 w-4' />
                </Button>
              </div>
            </div>
            <ScrollArea className='min-h-0 flex-1'>
              <div data-testid='chat-thread-list' className='space-y-1 p-2'>
                {threads.length === 0 && !loading ? (
                  <p className='rounded-md border border-dashed border-zinc-800 p-3 text-sm text-zinc-500'>
                    No chat threads yet.
                  </p>
                ) : null}
                {threads.length > 0 && filteredThreads.length === 0 ? (
                  <p className='rounded-md border border-dashed border-zinc-800 p-3 text-sm text-zinc-500'>
                    No matching conversations.
                  </p>
                ) : null}
                {filteredThreads.map((thread) => (
                  <button
                    key={thread.id}
                    type='button'
                    data-testid='chat-thread-item'
                    className={`flex w-full items-center gap-3 rounded-md border px-3 py-2 text-left text-sm transition ${
                      selectedThreadId === thread.id
                        ? 'border-primary bg-zinc-900 text-zinc-50'
                        : 'border-transparent text-zinc-300 hover:border-zinc-800 hover:bg-zinc-900/70'
                    }`}
                    onClick={() => {
                      setSelectedThreadId(thread.id)
                      void loadMessages(activeProfileId, thread.id)
                    }}
                  >
                    <span
                      className='flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-zinc-800 text-xs font-semibold text-zinc-100'
                      data-testid='chat-thread-avatar'
                    >
                      {threadInitial(thread.title)}
                    </span>
                    <span className='min-w-0 flex-1'>
                      <span className='block truncate font-medium'>
                        {thread.title}
                      </span>
                      <span
                        className='block truncate text-xs text-zinc-500'
                        data-testid='chat-thread-preview'
                      >
                        {selectedThreadId === thread.id && messages[0]
                          ? messages[0].content
                          : 'No messages yet'}
                      </span>
                    </span>
                  </button>
                ))}
              </div>
            </ScrollArea>
          </aside>

          <div className='flex min-h-0 flex-col bg-zinc-950'>
            {selectedThread ? (
              <>
                <div className='border-b border-zinc-800 p-4'>
                  <h2
                    className='font-semibold text-zinc-50'
                    data-testid='chat-thread-title'
                  >
                    {selectedThread.title}
                  </h2>
                </div>
                <ScrollArea className='min-h-[240px] flex-1 p-4'>
                  <div data-testid='chat-message-list' className='space-y-3'>
                    {messagesLoading ? (
                      <p className='text-sm text-zinc-500'>
                        Loading messages...
                      </p>
                    ) : null}
                    {!messagesLoading && messages.length === 0 ? (
                      <p className='text-sm text-zinc-500'>
                        No messages in this thread yet.
                      </p>
                    ) : null}
                    {messages.map((message) => (
                      <div
                        key={message.id}
                        className='rounded-md border border-zinc-800 bg-zinc-900/80 p-2 text-sm'
                      >
                        <p className='font-medium text-zinc-100'>
                          {prettyRole(message.role)}
                        </p>
                        <p className='text-zinc-300'>{message.content}</p>
                      </div>
                    ))}
                  </div>
                </ScrollArea>
                {sendError ? (
                  <p
                    className='mt-2 text-sm text-destructive'
                    data-testid='chat-send-error'
                  >
                    {sendError}
                  </p>
                ) : null}
                <div className='mt-3 flex gap-2'>
                  <Input
                    data-testid='chat-compose-input'
                    value={draft}
                    onChange={(event) => setDraft(event.target.value)}
                    placeholder='Type your message...'
                    disabled={!selectedThreadId}
                    className='border-zinc-800 bg-zinc-900 text-zinc-100 placeholder:text-zinc-500'
                  />
                  <Button
                    data-testid='chat-send-button'
                    onClick={() => void sendMessage()}
                    disabled={!selectedThreadId || !draft.trim()}
                  >
                    <Send className='mr-1 h-4 w-4' />
                    Send
                  </Button>
                </div>

                <div className='mt-4 space-y-3 rounded-md border border-zinc-800 p-3'>
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
                      <p className='text-zinc-500'>No attachments uploaded.</p>
                    ) : (
                      attachments.map((attachment) => (
                        <p key={attachment.id}>{attachment.filename}</p>
                      ))
                    )}
                  </div>
                </div>

                <div className='mt-4 space-y-3 rounded-md border border-zinc-800 p-3'>
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
                      className='h-9 rounded-md border border-zinc-800 bg-zinc-900 px-3'
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
                      className='border-zinc-800 bg-zinc-900 text-zinc-100 placeholder:text-zinc-500'
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
                      className='border-zinc-800 bg-zinc-900 text-zinc-100 placeholder:text-zinc-500'
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
                      className='border-zinc-800 bg-zinc-900 text-zinc-100 placeholder:text-zinc-500'
                    />
                    <Input
                      data-testid='chat-preview-title'
                      value={actionTitle}
                      onChange={(event) => setActionTitle(event.target.value)}
                      placeholder='Item title'
                      disabled={!selectedThreadId}
                      className='border-zinc-800 bg-zinc-900 text-zinc-100 placeholder:text-zinc-500'
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
            ) : (
              <div
                className='flex min-h-[520px] flex-1 items-center justify-center p-8'
                data-testid='chat-empty-workspace-state'
              >
                <div className='mx-auto max-w-sm text-center'>
                  <div className='mx-auto flex h-14 w-14 items-center justify-center rounded-full border border-zinc-800 bg-zinc-900'>
                    <MessageCircle className='h-7 w-7 text-zinc-400' />
                  </div>
                  <h2
                    className='mt-5 text-xl font-semibold text-zinc-50'
                    data-testid='chat-thread-title'
                  >
                    Select a conversation
                  </h2>
                  <p className='mt-2 text-sm text-zinc-500'>
                    Choose an existing thread or create a new one to continue a
                    durable Cabinet conversation.
                  </p>
                  <Button
                    type='button'
                    data-testid='chat-empty-workspace-action'
                    className='mt-5'
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
