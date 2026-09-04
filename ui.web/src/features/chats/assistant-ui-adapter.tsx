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
	cabinetMessageAttachments,
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
		attachment?: string
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
			{messages.flatMap((message) =>
				cabinetMessageAttachments(message).map((attachment) => (
					<div
						key={`${message.id}:${attachment.id}`}
						className='ml-auto max-w-[92%] rounded-lg border border-slate-700 bg-slate-900 px-3 py-2 text-xs text-slate-300'
						data-testid={
							testIds?.attachment ?? 'shell-assistant-message-attachment'
						}
						data-attachment-id={attachment.id}
						data-message-id={message.id}
					>
						<p className='font-medium text-slate-100'>
							{attachment.filename}
						</p>
						<p>
							{attachment.mime_type || 'file'} / {attachment.size_bytes}{' '}
							bytes
						</p>
						<p>
							{attachment.source === 'telegram'
								? 'Received from Telegram'
								: 'Uploaded in Cabinet'}
						</p>
					</div>
				))
			)}
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
  classNames?: {
    root?: string
    input?: string
    sendButton?: string
  }
}

export function CabinetAssistantUiComposer({
  composer,
  placeholder = 'Ask Cabinet to update, find, or link records...',
  testIds,
  classNames,
}: CabinetAssistantUiComposerProps) {
  return (
    <ComposerPrimitive.Root
      className={cn(
        'flex min-w-0 items-center gap-2 rounded-2xl border bg-muted/20 p-1',
        classNames?.root
      )}
      data-testid={testIds?.root ?? 'shell-assistant-ui-composer-primitive'}
      data-sending={composer.sending ? 'true' : 'false'}
    >
      <ComposerPrimitive.Input
        data-testid={testIds?.input ?? 'shell-assistant-compose-input'}
        placeholder={placeholder}
        disabled={composer.disabled}
        className={cn(
          'max-h-32 min-h-9 min-w-0 flex-1 resize-none overflow-y-auto border-0 bg-transparent px-3 py-2 text-sm shadow-none outline-none focus-visible:ring-0',
          classNames?.input
        )}
      />
      <ComposerPrimitive.Send asChild>
        <Button
          type='button'
          size='icon'
          data-testid={testIds?.sendButton ?? 'shell-assistant-send-button'}
          aria-label='Send assistant message'
          title='Send assistant message'
          disabled={composer.disabled}
          className={classNames?.sendButton}
        >
          <Send className='h-4 w-4' />
        </Button>
      </ComposerPrimitive.Send>
    </ComposerPrimitive.Root>
  )
}
