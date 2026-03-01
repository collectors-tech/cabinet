import { useCallback, useEffect, useMemo, useState } from 'react'
import { MessagesSquare, Send } from 'lucide-react'
import { ConfigDrawer } from '@/components/config-drawer'
import { Header } from '@/components/layout/header'
import { LanguageSwitch } from '@/components/language-switch'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Separator } from '@/components/ui/separator'

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

function prettyRole(role: ChatMessage['role']) {
  if (role === 'assistant') {
    return 'Assistant'
  }
  if (role === 'system') {
    return 'System'
  }
  return 'You'
}

export function Chats() {
  const [activeProfileId, setActiveProfileId] = useState('')
  const [threads, setThreads] = useState<ChatThread[]>([])
  const [selectedThreadId, setSelectedThreadId] = useState('')
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [threadTitle, setThreadTitle] = useState('')
  const [draft, setDraft] = useState('')
  const [loading, setLoading] = useState(true)
  const [messagesLoading, setMessagesLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [sendError, setSendError] = useState<string | null>(null)

  const selectedThread = useMemo(
    () => threads.find((thread) => thread.id === selectedThreadId) ?? null,
    [selectedThreadId, threads]
  )

  const loadMessages = useCallback(async (profileID: string, threadID: string) => {
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
  }, [])

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

  return (
    <>
      <Header>
        <Search />
        <div className='ms-auto flex items-center space-x-4'>
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
        <p className='text-muted-foreground'>
          Persistent profile-scoped conversation threads backed by Cabinet runtime.
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
              />
              <Button
                data-testid='chat-create-thread-button'
                onClick={() => void createThread()}
                disabled={!threadTitle.trim() || loading}
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
                  <p className='text-sm text-muted-foreground'>Loading messages...</p>
                ) : null}
                {!messagesLoading && selectedThread && messages.length === 0 ? (
                  <p className='text-sm text-muted-foreground'>
                    No messages in this thread yet.
                  </p>
                ) : null}
                {messages.map((message) => (
                  <div key={message.id} className='rounded-md border p-2 text-sm'>
                    <p className='font-medium'>{prettyRole(message.role)}</p>
                    <p>{message.content}</p>
                  </div>
                ))}
              </div>
            </ScrollArea>
            {sendError ? (
              <p className='mt-2 text-sm text-destructive' data-testid='chat-send-error'>
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
          </div>
        </section>
      </Main>
    </>
  )
}
