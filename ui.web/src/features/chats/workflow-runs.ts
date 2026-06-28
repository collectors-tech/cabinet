export type ChatWorkflowRun = {
  id: string
  workflow_id: string
  capability_id: string
  source_channel: string
  source_thread_id?: string
  source_message_id?: string
  status: string
  input?: Record<string, unknown>
  provider_trace?: Record<string, unknown>
  result?: Record<string, unknown>
  error?: Record<string, unknown>
  confirmation_state: string
  created_at: string
  updated_at: string
  started_at?: string
  completed_at?: string
}

export async function fetchChatWorkflowRuns(
  profileId: string,
  threadId: string
) {
  if (!profileId || !threadId) {
    return []
  }
  const response = await fetch(
    `/api/chat/workflow-runs?profile_id=${encodeURIComponent(profileId)}&thread_id=${encodeURIComponent(threadId)}`
  )
  if (!response.ok) {
    throw new Error(`chat_workflow_runs_${response.status}`)
  }
  const payload = (await response.json()) as { runs?: ChatWorkflowRun[] }
  return payload.runs ?? []
}

function stringValue(value: unknown) {
  return typeof value === 'string' && value.trim() ? value.trim() : ''
}

export function workflowRunResultSummary(run: ChatWorkflowRun) {
  const result = run.result ?? {}
  const error = run.error ?? {}
  const parts = [
    stringValue(result.route),
    stringValue(result.preview_id) ? `preview ${stringValue(result.preview_id)}` : '',
    stringValue(result.action),
    stringValue(result.item_id) ? `item ${stringValue(result.item_id)}` : '',
    stringValue(error.code),
    stringValue(error.message),
  ].filter(Boolean)
  if (parts.length > 0) {
    return parts.join(' | ')
  }
  return run.status === 'queued' || run.status === 'running'
    ? 'Awaiting next workflow update'
    : 'No result payload recorded'
}

export function workflowRunTimestamp(run: ChatWorkflowRun) {
  return run.completed_at || run.started_at || run.updated_at || run.created_at
}
