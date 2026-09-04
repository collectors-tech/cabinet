import { type AppendMessage, type ThreadMessageLike } from '@assistant-ui/react'

export type CabinetAssistantUiMessage = {
	id: string
	role: string
	content: string
	attachments_json?: CabinetMessageAttachment[] | string
  context?: {
    route?: { pathname?: string; search?: string }
    profile?: { id?: string }
    assistant?: { provider?: string; model?: string }
    app_control?: unknown
  }
}

export type CabinetMessageAttachment = {
	id: string
	filename: string
	mime_type: string
	size_bytes: number
	provenance: string
	source: string
	created_at: string
}

export function cabinetMessageAttachments(
	message: CabinetAssistantUiMessage
): CabinetMessageAttachment[] {
	const value = message.attachments_json
	if (Array.isArray(value)) return value
	if (typeof value !== 'string' || !value.trim()) return []
	try {
		const parsed: unknown = JSON.parse(value)
		return Array.isArray(parsed) ? (parsed as CabinetMessageAttachment[]) : []
	} catch {
		return []
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
