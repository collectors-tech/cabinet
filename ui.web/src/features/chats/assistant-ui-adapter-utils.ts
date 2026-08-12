import { type AppendMessage, type ThreadMessageLike } from '@assistant-ui/react'

export type CabinetAssistantUiMessage = {
  id: string
  role: string
  content: string
  context?: {
    route?: { pathname?: string; search?: string }
    profile?: { id?: string }
    assistant?: { provider?: string; model?: string }
    app_control?: unknown
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
