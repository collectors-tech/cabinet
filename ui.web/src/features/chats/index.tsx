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
  preview_id: string
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
    'create_inventory_item' | 'create_wishlist_entry' | 'update_inventory_item'
  >('create_inventory_item')
  const [actionTargetItemID, setActionTargetItemID] = useState('')
  const [actionPreview, setActionPreview] = useState<ChatActionPreview | null>(
    null
  )
  const [applyResult, setApplyResult] = useState<ChatApplyResult | null>(null)
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
    return threads.filter((thread) => thread.title.toLowerCase().includes(query))
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
    async (profileID: string, preserveSelected = true) => {
      const response = await fetch(
        `/api/chat/threads?profile_id=${encodeURIComponent(profileID)}`
      )
      if (!response.ok) {
        throw new Error(`chat_threads_${response.status}`)
      }
      const payload = (await response.json()) as { threads?: ChatThread[] }
      const nextThreads = payload.threads ?? []
      setThreads(nextThreads)
      const nextSelected =
        preserveSelected &&
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
      await loadThreads(profileID, false)
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
            actionMode === 'update_inventory_item'
              ? actionTargetItemID.trim()
              : '',
          priority: actionMode === 'create_wishlist_entry' ? 'medium' : '',
        },
      }),
    })
    if (!response.ok) {
      setSendError(`chat_action_preview_${response.status}`)
      return
    }
    const preview = (await response.json()) as ChatActionPreview
    setActionPreview(preview)
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
      return
    }
    const result = (await response.json()) as ChatApplyResult
    setApplyResult(result)
    setConfirmApplyOpen(false)
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
    if (applyResult.wishlist_id) {
      return `${withPart} to wishlist ${applyResult.wishlist_id}`
    }
    if (applyResult.item_id) {
      return `${withPart} to item ${applyResult.item_id}`
    }
    return withPart
  })()

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

        <section className='grid h-full min-h-[520px] gap-4 lg:grid-cols-[320px_1fr]'>
          <div className='rounded-md border p-3'>
            <div className='mb-3 flex gap-2'>
              <Input
                data-testid='chat-new-thread-input'
                placeholder='New thread title'
                value={threadTitle}
                onChange={(event) => setThreadTitle(event.target.value)}
                disabled={threadCreationDisabled}
              />
              <Button
                data-testid='chat-create-thread-button'
                onClick={() => void createThread()}
                disabled={threadCreationDisabled || !threadTitle.trim()}
              >
                Create
              </Button>
            </div>
            <ScrollArea className='h-[420px]'>
              <div data-testid='chat-thread-list' className='space-y-1'>
                {threads.length === 0 && !loading ? (
                  <p className='rounded-md border border-dashed p-3 text-sm text-muted-foreground'>
                    No chat threads yet.
                  </p>
                ) : null}
                {threads.map((thread) => (
                  <button
                    key={thread.id}
                    type='button'
                    data-testid='chat-thread-item'
                    className={`w-full rounded-md border px-3 py-2 text-left text-sm ${
                      selectedThreadId === thread.id
                        ? 'border-primary bg-primary/5'
                        : 'hover:bg-muted/40'
                    }`}
                    onClick={() => {
                      setSelectedThreadId(thread.id)
                      void loadMessages(activeProfileId, thread.id)
                    }}
                  >
                    {thread.title}
                  </button>
                ))}
              </div>
            </ScrollArea>
          </div>

          <div className='rounded-md border p-3'>
            <div className='mb-3'>
              <h2 className='font-semibold' data-testid='chat-thread-title'>
                {selectedThread?.title ?? 'Select or create a thread'}
              </h2>
            </div>
            <ScrollArea className='h-[380px] rounded-md border p-3'>
              <div data-testid='chat-message-list' className='space-y-3'>
                {messagesLoading ? (
                  <p className='text-sm text-muted-foreground'>
                    Loading messages...
                  </p>
                ) : null}
                {!messagesLoading && selectedThread && messages.length === 0 ? (
                  <p className='text-sm text-muted-foreground'>
                    No messages in this thread yet.
                  </p>
                ) : null}
                {messages.map((message) => (
                  <div
                    key={message.id}
                    className='rounded-md border p-2 text-sm'
                  >
                    <p className='font-medium'>{prettyRole(message.role)}</p>
                    <p>{message.content}</p>
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

            <div className='mt-4 space-y-3 rounded-md border p-3'>
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

            <div className='mt-4 space-y-3 rounded-md border p-3'>
              <p className='text-sm font-medium'>Action Preview</p>
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
                </select>
              </label>
              {actionMode === 'update_inventory_item' ? (
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
              <div className='grid gap-2 sm:grid-cols-2'>
                <Input
                  data-testid='chat-preview-part-number'
                  value={actionPartNumber}
                  onChange={(event) => setActionPartNumber(event.target.value)}
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
                  disabled={
                    !selectedThreadId ||
                    messages.length === 0 ||
                    !actionPartNumber.trim() ||
                    !actionTitle.trim()
                  }
                >
                  Preview Action
                </Button>
                <Button
                  type='button'
                  variant='outline'
                  data-testid='chat-apply-action-button'
                  onClick={() => setConfirmApplyOpen(true)}
                  disabled={!selectedThreadId || !actionPreview?.id}
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
                  {actionPreview.status})
                </p>
              ) : null}
              {applyResult ? (
                <p data-testid='chat-action-apply-result' className='text-sm'>
                  {applyResultSummary}
                </p>
              ) : null}
            </div>
          </div>
        </section>
      </Main>

      <AlertDialog open={confirmApplyOpen} onOpenChange={setConfirmApplyOpen}>
        <AlertDialogContent data-testid='chat-apply-confirm-dialog'>
          <AlertDialogHeader>
            <AlertDialogTitle>Confirm Copilot Action</AlertDialogTitle>
            <AlertDialogDescription data-testid='chat-apply-confirm-summary'>
              {actionPreview
                ? `Apply ${actionPreview.action} with part_number=${String(actionPreview.payload?.part_number ?? 'n/a')} title=${String(actionPreview.payload?.title ?? 'n/a')}`
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
      </AlertDialog>
    </>
  )
}
