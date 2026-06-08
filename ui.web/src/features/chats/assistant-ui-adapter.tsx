import {
  ComposerPrimitive,
  ThreadPrimitive,
  type AppendMessage,
  type MessageState,
  type ThreadMessageLike,
} from '@assistant-ui/react'
import { Bot, Send, UserRound } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'

export type CabinetAssistantUiMessage = {
  id: string
  role: string
  content: string
  context?: {
    route?: { pathname?: string; search?: string }
    profile?: { id?: string }
    assistant?: { provider?: string; model?: string }
  }
}

export type CabinetAssistantUiComposerState = {
  disabled: boolean
  sending: boolean
}

export function cabinetMessageToAssistantUi(
  message: CabinetAssistantUiMessage
): ThreadMessageLike {
  const role =
    message.role === 'assistant' || message.role === 'system'
      ? message.role
      : 'user'
  return {
    id: message.id,
    role,
    content: message.content,
    metadata: {
      custom: {
        cabinet_message_id: message.id,
        cabinet_route: message.context?.route,
        cabinet_profile: message.context?.profile,
        cabinet_assistant: message.context?.assistant,
      },
    },
  }
}

export function assistantAppendMessageText(message: AppendMessage) {
  return message.content
    .filter((part) => part.type === 'text')
    .map((part) => part.text)
    .join('\n')
    .trim()
}

function assistantUiMessageText(message: MessageState) {
  return message.content
    .filter((part) => part.type === 'text')
    .map((part) => part.text)
    .join('\n')
}

type CabinetAssistantUiMessageListProps = {
  messages: CabinetAssistantUiMessage[]
}

export function CabinetAssistantUiMessageList({
  messages,
}: CabinetAssistantUiMessageListProps) {
  return (
    <ThreadPrimitive.Root
      data-testid='shell-assistant-ui-adapter'
      data-message-count={messages.length}
    >
      <ThreadPrimitive.Viewport asChild>
        <div className='space-y-4 pb-2'>
          <ThreadPrimitive.Messages>
            {({ message }) => {
              const isUser = message.role === 'user'
              return (
                <div
                  className={cn(
                    'flex',
                    isUser ? 'justify-end' : 'justify-start'
                  )}
                  data-testid='shell-assistant-ui-message-primitive'
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
                    <div className='mb-1 flex items-center gap-1.5 text-[11px] font-medium text-muted-foreground uppercase'>
                      {isUser ? (
                        <UserRound className='h-3 w-3' />
                      ) : (
                        <Bot className='h-3 w-3' />
                      )}
                      {isUser ? 'You' : 'Assistant'}
                    </div>
                    <p>{assistantUiMessageText(message)}</p>
                  </div>
                </div>
              )
            }}
          </ThreadPrimitive.Messages>
        </div>
      </ThreadPrimitive.Viewport>
    </ThreadPrimitive.Root>
  )
}

type CabinetAssistantUiComposerProps = {
  composer: CabinetAssistantUiComposerState
}

export function CabinetAssistantUiComposer({
  composer,
}: CabinetAssistantUiComposerProps) {
  return (
    <ComposerPrimitive.Root
      className='flex items-center gap-2 rounded-2xl border bg-muted/20 p-1'
      data-testid='shell-assistant-ui-composer-primitive'
      data-sending={composer.sending ? 'true' : 'false'}
    >
      <ComposerPrimitive.Input
        data-testid='shell-assistant-compose-input'
        placeholder='Ask Cabinet to update, find, or link records...'
        disabled={composer.disabled}
        className='min-h-9 flex-1 resize-none border-0 bg-transparent px-3 py-2 text-sm shadow-none outline-none focus-visible:ring-0'
      />
      <ComposerPrimitive.Send asChild>
        <Button
          type='button'
          size='icon'
          data-testid='shell-assistant-send-button'
          aria-label='Send assistant message'
          disabled={composer.disabled}
        >
          <Send className='h-4 w-4' />
        </Button>
      </ComposerPrimitive.Send>
    </ComposerPrimitive.Root>
  )
}
