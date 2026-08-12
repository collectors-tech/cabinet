import {
  ComposerPrimitive,
  ThreadPrimitive,
  type MessageState,
} from '@assistant-ui/react'
import { Bot, Send, UserRound } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import {
  type CabinetAssistantUiComposerState,
  type CabinetAssistantUiMessage,
} from './assistant-ui-adapter-utils'

function assistantUiMessageText(message: MessageState) {
  return message.content
    .filter((part) => part.type === 'text')
    .map((part) => part.text)
    .join('\n')
}

type CabinetAssistantUiMessageListProps = {
  messages: CabinetAssistantUiMessage[]
  testIds?: {
    root?: string
    messagePrimitive?: string
    userBubble?: string
    assistantBubble?: string
  }
}

export function CabinetAssistantUiMessageList({
  messages,
  testIds,
}: CabinetAssistantUiMessageListProps) {
  return (
    <ThreadPrimitive.Root
      data-testid={testIds?.root ?? 'shell-assistant-ui-adapter'}
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
                  data-testid={
                    testIds?.messagePrimitive ??
                    'shell-assistant-ui-message-primitive'
                  }
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
                        ? (testIds?.userBubble ??
                          'shell-assistant-message-bubble-user')
                        : (testIds?.assistantBubble ??
                          'shell-assistant-message-bubble-assistant')
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
  placeholder?: string
  testIds?: {
    root?: string
    input?: string
    sendButton?: string
  }
}

export function CabinetAssistantUiComposer({
  composer,
  placeholder = 'Ask Cabinet to update, find, or link records...',
  testIds,
}: CabinetAssistantUiComposerProps) {
  return (
    <ComposerPrimitive.Root
      className='flex items-center gap-2 rounded-2xl border bg-muted/20 p-1'
      data-testid={testIds?.root ?? 'shell-assistant-ui-composer-primitive'}
      data-sending={composer.sending ? 'true' : 'false'}
    >
      <ComposerPrimitive.Input
        data-testid={testIds?.input ?? 'shell-assistant-compose-input'}
        placeholder={placeholder}
        disabled={composer.disabled}
        className='max-h-32 min-h-9 flex-1 resize-none overflow-y-auto border-0 bg-transparent px-3 py-2 text-sm shadow-none outline-none focus-visible:ring-0'
      />
      <ComposerPrimitive.Send asChild>
        <Button
          type='button'
          size='icon'
          data-testid={testIds?.sendButton ?? 'shell-assistant-send-button'}
          aria-label='Send assistant message'
          title='Send assistant message'
          disabled={composer.disabled}
        >
          <Send className='h-4 w-4' />
        </Button>
      </ComposerPrimitive.Send>
    </ComposerPrimitive.Root>
  )
}
