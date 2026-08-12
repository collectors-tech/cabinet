import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useNavigate, useRouterState } from '@tanstack/react-router'
import {
  AssistantModalPrimitive,
  AssistantRuntimeProvider,
  type AppendMessage,
  useExternalStoreRuntime,
} from '@assistant-ui/react'
import {
  Bot,
  CheckCircle2,
  ChevronLeft,
  ChevronRight,
  ExternalLink,
  GitBranchPlus,
  MessageSquarePlus,
  Paperclip,
  Pause,
  Play,
  ShieldAlert,
  Sparkles,
  SkipForward,
  VolumeX,
  Wand2,
  X,
} from 'lucide-react'
import { useAuthStore } from '@/stores/auth-store'
import { cabinetProtectedFetch } from '@/lib/cabinet-session'
import {
  dispatchShellCommand,
  type ShellCommandEvent,
  type ShellCommandType,
} from '@/lib/shell-command-bus'
import { useShellWorkspace } from '@/context/shell-workspace-context'
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
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { ScrollArea } from '@/components/ui/scroll-area'
import {
  CabinetAssistantUiComposer,
  CabinetAssistantUiMessageList,
} from '@/features/chats/assistant-ui-adapter'
import {
  assistantAppendMessageText,
  cabinetMessageToAssistantUi,
} from '@/features/chats/assistant-ui-adapter-utils'
import {
  fetchChatWorkflowRuns,
  type ChatWorkflowRun,
  workflowRunResultSummary,
  workflowRunTimestamp,
} from '@/features/chats/workflow-runs'

type ThreadMetadata = {
  provider?: string
  model?: string
  thread_semantics?: string
  forked_from_thread_id?: string
}

type Thread = {
  id: string
  title: string
  metadata?: ThreadMetadata
}

type Message = {
  id: string
  role: string
  content: string
  context?: {
    route?: { pathname?: string; search?: string }
    profile?: { id?: string }
    selection?: { active_workspace_collection?: string }
    assistant?: { provider?: string; model?: string }
    assistant_handoff?: { status?: string; inbox_item_id?: string }
    app_control?: AppControlContext
  }
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

type ActionPreview = {
  id: string
  action: string
  status: string
  payload?: { part_number?: string; title?: string; item_id?: string }
}

type AppControlContext = {
  capability_id?: string
  policy?: string
  route?: string
  setup_needed?: boolean
  preview?: ActionPreview
  guided_workflow?: GuidedWorkflowPlan
  workflow_run?: {
    id?: string
    status?: string
    confirmation_state?: string
  }
}

type GuidedWorkflowStep = {
  id: string
  title: string
  instruction: string
  route?: string
  ui_target_id?: string
  command: string
}

type GuidedWorkflowPlan = {
  recipe_id: string
  title: string
  mode: 'explain' | 'show_me' | 'do_it_with_me' | 'do_it_for_me'
  route: string
  steps: GuidedWorkflowStep[]
  mutation_boundary: string
  completion_criteria: string
}

type ApplyActionResult = {
  applied: boolean
  action: string
  item_id?: string
  wishlist_id?: string
  preview_id: string
}

type AgentSkillPreview = {
  skill_id: string
  status: string
  safety_level: string
  allowed: boolean
  confirmation_required: boolean
  blocker?: string
  preview_target?: Record<string, string>
  source_surface?: string
  source_channel?: string
}

type AgentSkillAuthorityError = {
  error?: string
  authority?: {
    entry_point?: string
    next_action?: string
    blocker?: string
  }
}

type AgentSkillApplyResult = {
  skill_id: string
  mutation_applied: boolean
  target?: {
    operation?: string
    provider_id?: string
    connection_status?: string
    secret_redacted?: boolean
    external_write_claimed?: boolean
    next_action?: string
  }
  source_surface?: string
  source_channel?: string
  source_thread_id?: string
  source_message_id?: string
}

type AgentSelectedRecord = {
  type?: string
  id?: string
  label?: string
  title?: string
  route_id?: string
}

type InboxAgentContext = {
  profile_id?: string
  thread_id?: string
  route_id?: string
  surface_id?: string
  source_channel?: string
  selected_notification?: {
    id?: string
    source?: string
  }
  source_thread_id?: string
  source_message_id?: string
}

type AgentSkillOption = {
  id: string
  label: string
  surface: string
  primaryLabel: string
  contextLabel: string
  secretLabel: string
}

const agentSkillOptions = [
  {
    id: 'cabinet.dashboard.summarise_activity',
    label: 'Summarise dashboard',
    surface: 'dashboard.home',
    primaryLabel: 'Time window',
    contextLabel: 'Workspace ID',
    secretLabel: 'Optional note',
  },
  {
    id: 'cabinet.integrations.test_connection',
    label: 'Test provider connection',
    surface: 'settings.integrations.provider.card',
    primaryLabel: 'Provider',
    contextLabel: 'Setup step',
    secretLabel: 'Secret input is redacted',
  },
  {
    id: 'cabinet.integrations.configure_provider',
    label: 'Configure provider',
    surface: 'settings.integrations.provider.card',
    primaryLabel: 'Provider',
    contextLabel: 'Setup step',
    secretLabel: 'Secret input is redacted',
  },
  {
    id: 'cabinet.settings.update_profile',
    label: 'Update profile settings',
    surface: 'settings.profile.form',
    primaryLabel: 'Display currency',
    contextLabel: 'Timezone',
    secretLabel: 'Private note',
  },
  {
    id: 'cabinet.settings.update_account',
    label: 'Update account settings',
    surface: 'settings.account.form',
    primaryLabel: 'Account email',
    contextLabel: 'Locale',
    secretLabel: 'Private account note',
  },
  {
    id: 'cabinet.settings.update_appearance',
    label: 'Update appearance settings',
    surface: 'settings.appearance.form',
    primaryLabel: 'Setting key',
    contextLabel: 'Setting scope',
    secretLabel: 'Setting value',
  },
  {
    id: 'cabinet.storage.configure_backup',
    label: 'Configure backup storage',
    surface: 'settings.storage.backup',
    primaryLabel: 'Backup target',
    contextLabel: 'Backup schedule',
    secretLabel: 'Private storage note',
  },
  {
    id: 'cabinet.maintenance.run_safe_check',
    label: 'Run data maintenance check',
    surface: 'settings.data.maintenance',
    primaryLabel: 'Maintenance scope',
    contextLabel: 'Check level',
    secretLabel: 'Private maintenance note',
  },
  {
    id: 'cabinet.market_watch.run_watch',
    label: 'Run saved watch',
    surface: 'market_watch.saved_watch.row',
    primaryLabel: 'Provider',
    contextLabel: 'Saved watch ID',
    secretLabel: 'Optional note',
  },
  {
    id: 'cabinet.market_watch.handoff_result',
    label: 'Handoff watch result',
    surface: 'market_watch.result.card',
    primaryLabel: 'Provider',
    contextLabel: 'Result ID',
    secretLabel: 'Destination',
  },
  {
    id: 'cabinet.purchases.create_order',
    label: 'Create purchase order',
    surface: 'purchases.inbox.capture',
    primaryLabel: 'Purchase source',
    contextLabel: 'Item ID',
    secretLabel: 'Tracking or source URL',
  },
  {
    id: 'cabinet.inventory.search_items',
    label: 'Search inventory items',
    surface: 'inventory.workspace.search',
    primaryLabel: 'Search query',
    contextLabel: 'Status or filter',
    secretLabel: 'Optional note',
  },
  {
    id: 'cabinet.inventory.create_item',
    label: 'Create inventory item',
    surface: 'inventory.quick-create',
    primaryLabel: 'Part number',
    contextLabel: 'Title',
    secretLabel: 'Source URL or note',
  },
  {
    id: 'cabinet.inventory.update_item',
    label: 'Update inventory item',
    surface: 'inventory.item.detail',
    primaryLabel: 'Item ID',
    contextLabel: 'New title',
    secretLabel: 'Notes',
  },
  {
    id: 'cabinet.inventory.attach_media',
    label: 'Attach inventory media',
    surface: 'inventory.media.assignment',
    primaryLabel: 'Item ID',
    contextLabel: 'Media ID',
    secretLabel: 'Optional note',
  },
  {
    id: 'cabinet.inventory.assign_to_collection',
    label: 'Assign item to collection',
    surface: 'inventory.collection.assignment',
    primaryLabel: 'Item ID',
    contextLabel: 'Collection name',
    secretLabel: 'Optional note',
  },
  {
    id: 'cabinet.wishlist.create_entry',
    label: 'Create wishlist entry',
    surface: 'wishlist.intent.capture',
    primaryLabel: 'Title',
    contextLabel: 'Target price',
    secretLabel: 'Source URL or note',
  },
  {
    id: 'cabinet.wishlist.mark_purchased',
    label: 'Mark wishlist purchased',
    surface: 'wishlist.entry.row',
    primaryLabel: 'Wishlist entry ID',
    contextLabel: 'Quantity',
    secretLabel: 'Purchase URL or note',
  },
  {
    id: 'cabinet.collections.create',
    label: 'Create collection',
    surface: 'collections.workspace.create',
    primaryLabel: 'Collection name',
    contextLabel: 'Description',
    secretLabel: 'Optional note',
  },
  {
    id: 'cabinet.collections.assign_item',
    label: 'Assign item to collection',
    surface: 'collections.workspace.assignment',
    primaryLabel: 'Collection ID or name',
    contextLabel: 'Item ID',
    secretLabel: 'Optional note',
  },
  {
    id: 'cabinet.media.search',
    label: 'Search media',
    surface: 'media.workspace.search',
    primaryLabel: 'Search query',
    contextLabel: 'Filter',
    secretLabel: 'Optional note',
  },
  {
    id: 'cabinet.media.upload_or_import',
    label: 'Import media',
    surface: 'media.workspace.import',
    primaryLabel: 'Source URL or path',
    contextLabel: 'Filename',
    secretLabel: 'Notes',
  },
  {
    id: 'cabinet.media.attach_to_item',
    label: 'Attach media to item',
    surface: 'media.workspace.assignment',
    primaryLabel: 'Media ID',
    contextLabel: 'Item ID',
    secretLabel: 'Notes',
  },
  {
    id: 'cabinet.media.review_unlinked',
    label: 'Review unlinked media',
    surface: 'media.workspace.unlinked',
    primaryLabel: 'Search query',
    contextLabel: 'Filter',
    secretLabel: 'Optional note',
  },
  {
    id: 'cabinet.media.update_notes',
    label: 'Update media notes',
    surface: 'media.workspace.metadata',
    primaryLabel: 'Media ID',
    contextLabel: 'Notes',
    secretLabel: 'Additional notes',
  },
  {
    id: 'cabinet.media.detach_from_item',
    label: 'Detach media from item',
    surface: 'media.workspace.assignment',
    primaryLabel: 'Media ID',
    contextLabel: 'Item ID',
    secretLabel: 'Reason',
  },
  {
    id: 'cabinet.discoveries.search',
    label: 'Search discoveries',
    surface: 'discoveries.result.list',
    primaryLabel: 'Provider',
    contextLabel: 'Search query',
    secretLabel: 'Optional note',
  },
  {
    id: 'cabinet.discoveries.review_result',
    label: 'Review discovery result',
    surface: 'discoveries.result.card',
    primaryLabel: 'Provider',
    contextLabel: 'Result ID',
    secretLabel: 'Optional note',
  },
  {
    id: 'cabinet.discoveries.dismiss_result',
    label: 'Dismiss discovery result',
    surface: 'discoveries.result.card',
    primaryLabel: 'Provider',
    contextLabel: 'Result ID',
    secretLabel: 'Notes',
  },
  {
    id: 'cabinet.discoveries.send_to_wishlist',
    label: 'Send discovery to wishlist',
    surface: 'discoveries.result.card',
    primaryLabel: 'Provider',
    contextLabel: 'Result ID',
    secretLabel: 'Notes',
  },
  {
    id: 'cabinet.discoveries.create_purchase',
    label: 'Create purchase from discovery',
    surface: 'discoveries.result.card',
    primaryLabel: 'Provider',
    contextLabel: 'Result ID',
    secretLabel: 'Notes',
  },
  {
    id: 'cabinet.discoveries.create_or_update_inventory_candidate',
    label: 'Create inventory candidate',
    surface: 'discoveries.result.card',
    primaryLabel: 'Provider',
    contextLabel: 'Result ID',
    secretLabel: 'Notes',
  },
] satisfies AgentSkillOption[]

type NavigationAction = {
  id: string
  label: string
  target: string
  reason: string
}

type AssistantProviderOption = {
  provider: string
  label: string
  models: { value: string; label: string }[]
}

const assistantProviderOptions: AssistantProviderOption[] = [
  {
    provider: 'openai',
    label: 'OpenAI',
    models: [
      { value: 'gpt-4o-mini', label: 'gpt-4o-mini' },
      { value: 'gpt-4.1-mini', label: 'gpt-4.1-mini' },
    ],
  },
  {
    provider: 'anthropic',
    label: 'Anthropic',
    models: [
      { value: 'claude-3-5-haiku', label: 'claude-3-5-haiku' },
      { value: 'claude-3-7-sonnet', label: 'claude-3-7-sonnet' },
    ],
  },
]

function activeCollectionKey(profileScope: string) {
  return `cabinet.workspace.collections.active.${profileScope || 'local'}`
}

function assistantThreadKey(profileId: string) {
  return `cabinet.assistant.workspace.thread.${profileId || 'local'}`
}

function assistantProviderKey(profileId: string) {
  return `cabinet.assistant.workspace.provider.${profileId || 'local'}`
}

function assistantModelKey(profileId: string) {
  return `cabinet.assistant.workspace.model.${profileId || 'local'}`
}

function agentSelectedRecordKey(profileId: string) {
  return `cabinet.agent.selected_record.${profileId || 'local'}`
}

function routeMatchesAgentSelectedRecord(
  recordRoute: string,
  currentRoute: string
) {
  const normalizedRecordRoute = recordRoute.replace(/\/+$/, '')
  const normalizedCurrentRoute = currentRoute.replace(/\/+$/, '')
  return (
    normalizedRecordRoute !== '' &&
    normalizedCurrentRoute === normalizedRecordRoute
  )
}

function loadAgentSelectedRecord(
  profileId: string,
  routePath: string
): AgentSelectedRecord | null {
  if (typeof window === 'undefined') return null
  for (const key of [
    agentSelectedRecordKey(profileId),
    agentSelectedRecordKey('local'),
  ]) {
    try {
      const raw = window.localStorage.getItem(key)
      if (!raw) continue
      const parsed = JSON.parse(raw) as AgentSelectedRecord
      if (
        parsed?.type?.trim() &&
        parsed?.id?.trim() &&
        routeMatchesAgentSelectedRecord(parsed.route_id || '', routePath)
      ) {
        return parsed
      }
    } catch {
      // Ignore malformed stale context and continue with the next fallback.
    }
  }
  return null
}

function surfaceIDForAgentSelectedRecord(record: AgentSelectedRecord | null) {
  if (!record) {
    return 'chats.side-panel'
  }
  if (record.type === 'inventory_item') {
    return 'inventory.item.detail'
  }
  return `${record.type}.detail`
}

function loadInboxAgentContext(
  profileId: string,
  threadId: string
): InboxAgentContext | null {
  if (typeof window === 'undefined' || !profileId || !threadId) return null
  try {
    const raw = window.localStorage.getItem(
      'cabinet.agent.inbox_notification_context'
    )
    if (!raw) return null
    const parsed = JSON.parse(raw) as InboxAgentContext
    if (parsed.profile_id !== profileId || parsed.thread_id !== threadId) {
      return null
    }
    if (!parsed.selected_notification?.id?.trim()) {
      return null
    }
    return parsed
  } catch {
    return null
  }
}

function defaultModelForProvider(provider: string) {
  return (
    assistantProviderOptions.find((option) => option.provider === provider)
      ?.models[0]?.value || 'gpt-4o-mini'
  )
}

function resultLink(result: ApplyActionResult | null) {
  if (!result) {
    return null
  }
  if (result.item_id) {
    return {
      href: `/inventory/?item=${encodeURIComponent(result.item_id)}`,
      label: 'Open item',
    }
  }
  if (result.wishlist_id) {
    return {
      href: `/wishlist/?item=${encodeURIComponent(result.wishlist_id)}`,
      label: 'Open wishlist item',
    }
  }
  return null
}

function labelForRoute(route: string) {
  const normalized = route.replace(/^\/+/, '').replace(/\/+$/, '')
  if (!normalized) {
    return 'Open Cabinet'
  }
  return `Open ${normalized
    .split('/')
    .filter(Boolean)
    .map((segment) => segment.charAt(0).toUpperCase() + segment.slice(1))
    .join(' / ')}`
}

function navigationActionFromAppControl(
  appControl: AppControlContext | undefined
): NavigationAction | null {
  const route = appControl?.route?.trim()
  if (!route) {
    return null
  }
  return {
    id: appControl?.capability_id || 'navigate.open_surface',
    label: labelForRoute(route),
    target: route,
    reason:
      appControl?.policy === 'preview-before-apply'
        ? 'Cabinet planned this as a read-only navigation action from the assistant thread.'
        : 'Cabinet planned this route action from the assistant thread.',
  }
}

function latestAppControl(messages: Message[]) {
  return [...messages].reverse().find((message) => message.context?.app_control)
    ?.context?.app_control
}

async function loadAssistantDefaultSettings(profileId: string) {
  const response = await fetch(`/api/profiles/${profileId}/settings`)
  if (!response.ok) {
    throw new Error(`profile_settings_${response.status}`)
  }
  const payload = (await response.json()) as {
    settings?: Record<string, string>
  }
  const settings = payload.settings ?? {}
  const nextProvider = settings.assistant_default_provider?.trim() || 'openai'
  const nextModel =
    settings.assistant_default_model?.trim() ||
    defaultModelForProvider(nextProvider)
  return { nextProvider, nextModel }
}

export function AssistantWorkspacePanel() {
  const { activeProfileId, setActiveWorkspace } = useShellWorkspace()
  const navigate = useNavigate()
  const authUser = useAuthStore((state) => state.auth.user)
  const location = useRouterState({
    select: (state) => ({
      pathname: state.location.pathname,
      search: state.location.searchStr,
    }),
  })
  const profileScope = useMemo(
    () => authUser?.email || authUser?.accountNo || 'local',
    [authUser?.accountNo, authUser?.email]
  )
  const [threadId, setThreadId] = useState('')
  const [threadMetadata, setThreadMetadata] = useState<ThreadMetadata>({})
  const [threads, setThreads] = useState<Thread[]>([])
  const [messages, setMessages] = useState<Message[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [sending, setSending] = useState(false)
  const [pendingAttachment, setPendingAttachment] = useState<File | null>(null)
  const [attachments, setAttachments] = useState<ChatAttachment[]>([])
  const [provider, setProvider] = useState('openai')
  const [model, setModel] = useState('gpt-4o-mini')
  const manualProviderModelChangeRef = useRef(false)
  const [actionPartNumber, setActionPartNumber] = useState('ASSIST-001')
  const [actionTitle, setActionTitle] = useState('Assistant Proposed Item')
  const [actionPreview, setActionPreview] = useState<ActionPreview | null>(null)
  const [agentSkillID, setAgentSkillID] = useState(agentSkillOptions[0].id)
  const [agentSkillProvider, setAgentSkillProvider] = useState('ebay')
  const [agentSkillSetupStep, setAgentSkillSetupStep] = useState('oauth')
  const [agentSkillSecret, setAgentSkillSecret] = useState('')
  const [agentSkillPreview, setAgentSkillPreview] =
    useState<AgentSkillPreview | null>(null)
  const [agentSkillResult, setAgentSkillResult] =
    useState<AgentSkillApplyResult | null>(null)
  const [confirmTarget, setConfirmTarget] = useState<'action' | 'agent-skill'>(
    'action'
  )
  const [executionState, setExecutionState] = useState<
    'idle' | 'queued' | 'running' | 'success' | 'failure'
  >('idle')
  const [applyResult, setApplyResult] = useState<ApplyActionResult | null>(null)
  const [confirmApplyOpen, setConfirmApplyOpen] = useState(false)
  const [permissionGuidance, setPermissionGuidance] = useState(
    'Read-only browsing is always allowed. Structured mutations are preview-first and confirmation-required before any apply call runs.'
  )
  const [navigationAction, setNavigationAction] =
    useState<NavigationAction | null>(null)
  const [commandEvents, setCommandEvents] = useState<ShellCommandEvent[]>([])
  const [workflowRuns, setWorkflowRuns] = useState<ChatWorkflowRun[]>([])
  const [workflowRunsLoading, setWorkflowRunsLoading] = useState(false)
  const [workflowRunsError, setWorkflowRunsError] = useState('')
  const [guidedPaused, setGuidedPaused] = useState(false)

  const routeContext = useMemo(
    () => ({
      pathname: location.pathname || '/',
      search: location.search || '',
    }),
    [location.pathname, location.search]
  )

  const selectionContext = useMemo(() => {
    try {
      return (
        window.localStorage.getItem(activeCollectionKey(profileScope)) ||
        'All Items'
      )
    } catch {
      return 'All Items'
    }
  }, [profileScope])
  const selectedAgentRecordContext = useMemo(
    () => loadAgentSelectedRecord(activeProfileId, routeContext.pathname),
    [activeProfileId, routeContext.pathname]
  )
  const inboxAgentContext = useMemo(
    () => loadInboxAgentContext(activeProfileId, threadId),
    [activeProfileId, threadId]
  )
  const agentContextEnvelope = useMemo(
    () => ({
      profile_id: activeProfileId,
      workspace_id: profileScope,
      route_id: inboxAgentContext?.route_id || routeContext.pathname,
      surface_id:
        inboxAgentContext?.surface_id ||
        surfaceIDForAgentSelectedRecord(selectedAgentRecordContext),
      selected_record: selectedAgentRecordContext
        ? {
            type: selectedAgentRecordContext.type,
            id: selectedAgentRecordContext.id,
          }
        : undefined,
      thread_id: threadId,
      source_channel: inboxAgentContext?.source_channel || 'in-app',
      selected_notification: inboxAgentContext?.selected_notification
        ? {
            id: inboxAgentContext.selected_notification.id,
            source: inboxAgentContext.selected_notification.source,
          }
        : undefined,
      source_thread_id: inboxAgentContext?.source_thread_id,
      source_message_id: inboxAgentContext?.source_message_id,
      permission_state: 'ask_before_local_changes',
      setup_state: 'ready',
    }),
    [
      activeProfileId,
      inboxAgentContext,
      profileScope,
      routeContext.pathname,
      selectedAgentRecordContext,
      threadId,
    ]
  )

  const availableModels = useMemo(
    () =>
      assistantProviderOptions.find((option) => option.provider === provider)
        ?.models || [],
    [provider]
  )
  const appControl = useMemo(() => latestAppControl(messages), [messages])
  const guidedWorkflow = appControl?.guided_workflow
  const backendNavigationAction = useMemo(
    () => navigationActionFromAppControl(appControl),
    [appControl]
  )
  const displayedNavigationAction = navigationAction ?? backendNavigationAction
  const displayedPermissionGuidance = appControl?.setup_needed
    ? 'Provider setup is needed before Cabinet can run this assistant action.'
    : permissionGuidance
  const selectedAgentSkillOption = useMemo(
    () =>
      agentSkillOptions.find((option) => option.id === agentSkillID) ??
      agentSkillOptions[0],
    [agentSkillID]
  )

  const selectedThreadTitle = useMemo(
    () =>
      threads.find((thread) => thread.id === threadId)?.title ||
      'Assistant chat',
    [threadId, threads]
  )

  function inferNavigationAction(prompt: string): NavigationAction | null {
    const normalized = prompt.toLowerCase()
    if (
      normalized.includes('layout') &&
      (normalized.includes('config') ||
        normalized.includes('settings') ||
        normalized.includes('configure'))
    ) {
      return {
        id: 'open-layout-config',
        label: 'Open layout settings',
        target: '/settings/display',
        reason:
          'The request mentions layout configuration, so Cabinet can open the display settings surface.',
      }
    }
    return null
  }

  const createAssistantThread = useCallback(async (
    profileId: string,
    nextProvider: string,
    nextModel: string,
    options?: { semantics?: string; forkedFromThreadId?: string }
  ) => {
    const semantics =
      options?.semantics?.trim() || 'assistant_workspace_session'
    const forkedFromThreadId = options?.forkedFromThreadId?.trim() || ''
    const metadata: ThreadMetadata = {
      provider: nextProvider,
      model: nextModel,
      thread_semantics: semantics,
    }
    if (forkedFromThreadId.trim()) {
      metadata.forked_from_thread_id = forkedFromThreadId.trim()
    }
    const createResp = await fetch('/api/chat/threads', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        profile_id: profileId,
        title: `Assistant Workspace (${nextProvider} / ${nextModel})`,
        metadata,
      }),
    })
    if (!createResp.ok) {
      throw new Error('failed_to_create_assistant_thread')
    }
    const created = (await createResp.json()) as Thread
    setThreadId(created.id)
    setThreadMetadata(created.metadata ?? metadata)
    setThreads((current) => [
      created,
      ...current.filter((thread) => thread.id !== created.id),
    ])
    try {
      window.localStorage.setItem(assistantThreadKey(profileId), created.id)
      window.localStorage.setItem(assistantProviderKey(profileId), nextProvider)
      window.localStorage.setItem(assistantModelKey(profileId), nextModel)
    } catch {
      // ignore storage issues
    }
    return created.id
  }, [])

  const ensureThread = useCallback(async (
    profileId: string,
    nextProvider: string,
    nextModel: string
  ) => {
    const storageKey = assistantThreadKey(profileId)
    let nextThreadID = ''
    try {
      nextThreadID = window.localStorage.getItem(storageKey) || ''
    } catch {
      nextThreadID = ''
    }

    if (!nextThreadID) {
      nextThreadID = await createAssistantThread(
        profileId,
        nextProvider,
        nextModel
      )
    }

    setThreadId(nextThreadID)
    return nextThreadID
  }, [createAssistantThread])

  const loadMessages = useCallback(async (
    profileId: string,
    targetThreadId: string
  ) => {
    setWorkflowRunsLoading(true)
    try {
      const [resp, runs] = await Promise.all([
        cabinetProtectedFetch(
          `/api/chat/messages?profile_id=${encodeURIComponent(profileId)}&thread_id=${encodeURIComponent(targetThreadId)}`,
          profileId
        ),
        fetchChatWorkflowRuns(profileId, targetThreadId),
      ])
      if (!resp.ok) {
        throw new Error('failed_to_load_assistant_messages')
      }
      const payload = (await resp.json()) as { messages?: Message[] }
      setMessages(payload.messages ?? [])
      setWorkflowRuns(runs)
      setWorkflowRunsError('')
    } catch (err) {
      setWorkflowRuns([])
      setWorkflowRunsError(
        err instanceof Error ? err.message : 'failed_to_load_workflow_runs'
      )
      throw err
    } finally {
      setWorkflowRunsLoading(false)
    }
  }, [])

  const loadThreads = useCallback(async (profileId: string) => {
    const resp = await fetch(
      `/api/chat/threads?profile_id=${encodeURIComponent(profileId)}`
    )
    if (!resp.ok) {
      throw new Error('failed_to_load_assistant_threads')
    }
    const payload = (await resp.json()) as { threads?: Thread[] }
    const nextThreads = payload.threads ?? []
    setThreads(nextThreads)
    return nextThreads
  }, [])

  useEffect(() => {
    const preview = appControl?.preview
    if (!preview?.id || actionPreview?.id === preview.id) {
      return
    }
    setActionPreview(preview)
    setApplyResult(null)
    setExecutionState('running')
    setPermissionGuidance(
      'Cabinet created this preview from the assistant thread. Confirm before any mutation is applied.'
    )
  }, [actionPreview?.id, appControl?.preview])

  async function handleSelectThread(nextThreadId: string) {
    if (!activeProfileId || !nextThreadId || nextThreadId === threadId) return
    const selected = threads.find((thread) => thread.id === nextThreadId)
    manualProviderModelChangeRef.current = true
    setThreadId(nextThreadId)
    setThreadMetadata(selected?.metadata ?? {})
    if (selected?.metadata?.provider) setProvider(selected.metadata.provider)
    if (selected?.metadata?.model) setModel(selected.metadata.model)
    setMessages([])
    setPendingAttachment(null)
    setAttachments([])
    setActionPreview(null)
    setApplyResult(null)
    setExecutionState('idle')
    setWorkflowRuns([])
    setWorkflowRunsError('')
    setNavigationAction(null)
    setError('')
    setLoading(true)
    try {
      window.localStorage.setItem(
        assistantThreadKey(activeProfileId),
        nextThreadId
      )
      await loadMessages(activeProfileId, nextThreadId)
    } catch (err) {
      setError(
        err instanceof Error ? err.message : 'failed_to_load_assistant_messages'
      )
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    let cancelled = false
    if (!activeProfileId) return

    setLoading(true)
    setError('')
    void (async () => {
      try {
        const { nextProvider: storedProvider, nextModel: storedModel } =
          await loadAssistantDefaultSettings(activeProfileId)
        try {
          window.localStorage.setItem(
            assistantProviderKey(activeProfileId),
            storedProvider
          )
          window.localStorage.setItem(
            assistantModelKey(activeProfileId),
            storedModel
          )
        } catch {
          // ignore storage issues
        }
        if (!cancelled && !manualProviderModelChangeRef.current) {
          setProvider(storedProvider)
          setModel(storedModel)
        }
        const ensuredThread = await ensureThread(
          activeProfileId,
          storedProvider,
          storedModel
        )
        if (cancelled) return
        const threads = await loadThreads(activeProfileId)
        const activeThread = threads.find(
          (thread) => thread.id === ensuredThread
        )
        if (
          activeThread?.metadata &&
          !cancelled &&
          !manualProviderModelChangeRef.current
        ) {
          setThreadMetadata(activeThread.metadata)
          setProvider(activeThread.metadata.provider || storedProvider)
          setModel(activeThread.metadata.model || storedModel)
        }
        if (manualProviderModelChangeRef.current) return
        await loadMessages(activeProfileId, ensuredThread)
      } catch (err) {
        if (!cancelled) {
          setError(
            err instanceof Error
              ? err.message
              : 'assistant_workspace_bootstrap_failed'
          )
        }
      } finally {
        if (!cancelled) setLoading(false)
      }
    })()

    return () => {
      cancelled = true
    }
  }, [activeProfileId, ensureThread, loadMessages, loadThreads])

  useEffect(() => {
    let cancelled = false
    if (!activeProfileId || loading || manualProviderModelChangeRef.current) {
      return
    }

    void (async () => {
      try {
        const { nextProvider, nextModel } =
          await loadAssistantDefaultSettings(activeProfileId)
        if (cancelled) return
        const threadSemantics = threadMetadata.thread_semantics || ''
        const allowsDefaultSync =
          threadSemantics === '' ||
          threadSemantics === 'assistant_workspace_session'
        if (!allowsDefaultSync) {
          return
        }
        if (
          (threadMetadata.provider && threadMetadata.provider !== provider) ||
          (threadMetadata.model && threadMetadata.model !== model)
        ) {
          return
        }
        if (provider === nextProvider && model === nextModel) {
          return
        }
        try {
          window.localStorage.setItem(
            assistantProviderKey(activeProfileId),
            nextProvider
          )
          window.localStorage.setItem(
            assistantModelKey(activeProfileId),
            nextModel
          )
        } catch {
          // ignore storage issues
        }
        setProvider(nextProvider)
        setModel(nextModel)
        setThreadMetadata((current) => ({
          ...current,
          provider: nextProvider,
          model: nextModel,
        }))
      } catch {
        // best-effort live sync only
      }
    })()

    return () => {
      cancelled = true
    }
  }, [
    activeProfileId,
    loading,
    provider,
    model,
    threadMetadata.model,
    threadMetadata.provider,
    threadMetadata.thread_semantics,
  ])

  async function handleProviderChange(nextProvider: string) {
    if (!activeProfileId) return
    const nextModel = defaultModelForProvider(nextProvider)
    manualProviderModelChangeRef.current = true
    setProvider(nextProvider)
    setModel(nextModel)
    setSending(true)
    setError('')
    try {
      const previousThreadId = threadId
      const newThreadId = await createAssistantThread(
        activeProfileId,
        nextProvider,
        nextModel,
        {
          semantics: 'fork_on_provider_model_change',
          forkedFromThreadId: previousThreadId,
        }
      )
      await loadMessages(activeProfileId, newThreadId)
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : 'failed_to_change_assistant_provider'
      )
    } finally {
      setSending(false)
    }
  }

  async function handleModelChange(nextModel: string) {
    if (!activeProfileId) return
    manualProviderModelChangeRef.current = true
    setModel(nextModel)
    setSending(true)
    setError('')
    try {
      const previousThreadId = threadId
      const newThreadId = await createAssistantThread(
        activeProfileId,
        provider,
        nextModel,
        {
          semantics: 'fork_on_provider_model_change',
          forkedFromThreadId: previousThreadId,
        }
      )
      await loadMessages(activeProfileId, newThreadId)
    } catch (err) {
      setError(
        err instanceof Error ? err.message : 'failed_to_change_assistant_model'
      )
    } finally {
      setSending(false)
    }
  }

  async function handleNewThread() {
    if (!activeProfileId) return
    manualProviderModelChangeRef.current = true
    setSending(true)
    setError('')
    try {
      const newThreadId = await createAssistantThread(
        activeProfileId,
        provider,
        model,
        {
          semantics: 'manual_new_thread',
        }
      )
      setActionPreview(null)
      setApplyResult(null)
      setNavigationAction(null)
      setCommandEvents([])
      setMessages([])
      setPendingAttachment(null)
      setAttachments([])
      setExecutionState('idle')
      await loadMessages(activeProfileId, newThreadId)
    } catch (err) {
      setError(
        err instanceof Error ? err.message : 'failed_to_reset_assistant_thread'
      )
    } finally {
      setSending(false)
    }
  }

  const sendAssistantMessage = useCallback(
    async (messageDraft: string) => {
      const normalizedDraft = messageDraft.trim()
      if (!activeProfileId || !threadId || !normalizedDraft) return
      setSending(true)
      setError('')
      setNavigationAction(null)
      try {
        const response = await cabinetProtectedFetch(
          '/api/chat/messages',
          activeProfileId,
          {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            profile_id: activeProfileId,
            thread_id: threadId,
            role: 'user',
            content: normalizedDraft,
            attachment_ids: attachments.map((attachment) => attachment.id),
            agent_context: agentContextEnvelope,
            context: {
              route: routeContext,
              profile: { id: activeProfileId },
              selection: { active_workspace_collection: selectionContext },
              assistant: { provider, model },
            },
          }),
          }
        )
        if (!response.ok) throw new Error('failed_to_send_assistant_message')
        setAttachments([])
        setNavigationAction(inferNavigationAction(normalizedDraft))
        await loadMessages(activeProfileId, threadId)
      } catch (err) {
        setError(
          err instanceof Error
            ? err.message
            : 'failed_to_send_assistant_message'
        )
      } finally {
        setSending(false)
      }
    },
    [
      activeProfileId,
      agentContextEnvelope,
      attachments,
      loadMessages,
      model,
      provider,
      routeContext,
      selectionContext,
      threadId,
    ]
  )

  const uploadAttachment = async () => {
    if (!activeProfileId || !threadId || !pendingAttachment) {
      return
    }
    setError('')
    const form = new FormData()
    form.set('profile_id', activeProfileId)
    form.set('thread_id', threadId)
    form.set('file', pendingAttachment)

    try {
      const response = await fetch('/api/chat/attachments', {
        method: 'POST',
        body: form,
      })
      if (!response.ok)
        throw new Error(`assistant_attachment_${response.status}`)
      const attachment = (await response.json()) as ChatAttachment
      setAttachments((current) => [attachment, ...current])
      setPendingAttachment(null)
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : 'assistant_attachment_upload_failed'
      )
    }
  }

  const removeAttachment = (attachmentID: string) => {
    setAttachments((current) =>
      current.filter((attachment) => attachment.id !== attachmentID)
    )
  }

  const appendCommandEvent = useCallback((event: ShellCommandEvent) => {
    setCommandEvents((current) => [...current.slice(-19), event])
  }, [])

  const runNavigationAction = useCallback(
    async (action: NavigationAction) => {
      await dispatchShellCommand(
        {
          id: `${action.id}:${action.target}`,
          type: 'navigate.open_surface',
          route: action.target,
        },
        {
          navigate: async (route) => {
            await navigate({ to: route })
            setActiveWorkspace('assistant')
          },
          emit: appendCommandEvent,
        }
      )
    },
    [appendCommandEvent, navigate, setActiveWorkspace]
  )

  const runHighlightCommand = useCallback(
    async (targetId: string) => {
      await dispatchShellCommand(
        {
          id: `highlight:${targetId}`,
          type: 'ui.highlight_target',
          targetId,
          title: 'Inventory workspace',
          instruction:
            'Guided walkthroughs use registered target ids before they highlight controls or ask for changes.',
        },
        {
          navigate: async (route) => {
            await navigate({ to: route })
          },
          emit: appendCommandEvent,
        }
      )
    },
    [appendCommandEvent, navigate]
  )

  const cancelGuidance = useCallback(async () => {
    await dispatchShellCommand(
      {
        id: 'walkthrough:cancel',
        type: 'walkthrough.cancel',
      },
      {
        navigate: async (route) => {
          await navigate({ to: route })
        },
        emit: appendCommandEvent,
      }
    )
  }, [appendCommandEvent, navigate])

  const dispatchGuidedControl = useCallback(
    async (type: ShellCommandType) => {
      await dispatchShellCommand(
        {
          id: `walkthrough:${type}`,
          type,
        },
        {
          navigate: async (route) => {
            await navigate({ to: route })
          },
          emit: appendCommandEvent,
        }
      )
    },
    [appendCommandEvent, navigate]
  )

  const dispatchGuidedStep = useCallback(
    async (step: GuidedWorkflowStep) => {
      await dispatchShellCommand(
        {
          id: `walkthrough:${step.id}`,
          type: step.command as ShellCommandType,
          route: step.route,
          targetId: step.ui_target_id,
          title: step.title,
          instruction: step.instruction,
        },
        {
          navigate: async (route) => {
            await navigate({ to: route })
            setActiveWorkspace('assistant')
          },
          emit: appendCommandEvent,
        }
      )
    },
    [appendCommandEvent, navigate, setActiveWorkspace]
  )

  const runGuidedWalkthrough = useCallback(
    async (plan: GuidedWorkflowPlan) => {
      setGuidedPaused(false)
      for (const step of plan.steps) {
        if (step.command === 'navigate.open_surface') {
          await dispatchGuidedStep(step)
        }
        if (step.command === 'ui.highlight_target') {
          await dispatchGuidedStep(step)
        }
        if (
          step.command === 'chat.action.preview' ||
          step.command === 'chat.action.confirm_apply'
        ) {
          await dispatchGuidedStep(step)
        }
      }
    },
    [dispatchGuidedStep]
  )

  const handleAssistantUiNewMessage = useCallback(
    async (message: AppendMessage) => {
      await sendAssistantMessage(assistantAppendMessageText(message))
    },
    [sendAssistantMessage]
  )

  const assistantRuntime = useExternalStoreRuntime<Message>({
    messages,
    convertMessage: cabinetMessageToAssistantUi,
    isLoading: loading,
    isRunning: sending,
    isSendDisabled: !activeProfileId || !threadId || loading || sending,
    onNew: handleAssistantUiNewMessage,
    adapters: {
      threadList: {
        threadId,
        isLoading: loading,
        threads: threads.map((thread) => ({
          id: thread.id,
          title: thread.title,
          status: 'regular' as const,
          custom: {
            provider: thread.metadata?.provider,
            model: thread.metadata?.model,
            thread_semantics: thread.metadata?.thread_semantics,
          },
        })),
        onSwitchToNewThread: handleNewThread,
        onSwitchToThread: handleSelectThread,
      },
    },
  })

  async function previewAction() {
    if (!activeProfileId || !threadId) return
    setExecutionState('queued')
    setError('')
    setApplyResult(null)
    setPermissionGuidance(
      'Structured mutations are preview-only until you explicitly confirm apply.'
    )
    try {
      const response = await cabinetProtectedFetch(
        '/api/chat/actions/preview',
        activeProfileId,
        {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          profile_id: activeProfileId,
          thread_id: threadId,
          action: 'create_item_stub',
          payload: {
            part_number: actionPartNumber.trim(),
            title: actionTitle.trim(),
            brand: 'AFX',
            category: 'General',
          },
        }),
        }
      )
      if (!response.ok) throw new Error(`assistant_preview_${response.status}`)
      const preview = (await response.json()) as ActionPreview
      setActionPreview(preview)
      setExecutionState('running')
      await loadMessages(activeProfileId, threadId)
    } catch (err) {
      setExecutionState('failure')
      setWorkflowRunsLoading(false)
      setError(err instanceof Error ? err.message : 'assistant_preview_failed')
      setPermissionGuidance(
        'This action could not be previewed under the active policy. Read-only browsing remains available; mutation preview/apply may be unavailable.'
      )
    }
  }

  function agentSkillParameters() {
    const primary = agentSkillProvider.trim()
    const context = agentSkillSetupStep.trim()
    const secretOrTarget = agentSkillSecret.trim()
    const params: Record<string, string | Record<string, string>> = {}
    if (agentSkillID === 'cabinet.dashboard.summarise_activity') {
      if (primary) {
        params.window = primary
      }
      if (context) {
        params.workspace_id = context
      }
      if (secretOrTarget) {
        params.notes = secretOrTarget
      }
      return params
    }
    if (agentSkillID.startsWith('cabinet.integrations.')) {
      params.provider_id = primary
      if (context) {
        params.setup_step = context
      }
      if (secretOrTarget) {
        params.provider_secret = secretOrTarget
      }
      return params
    }
    if (agentSkillID === 'cabinet.market_watch.run_watch') {
      params.provider_id = primary
      if (context) {
        params.watch_id = context
      }
      if (secretOrTarget) {
        params.note = secretOrTarget
      }
      return params
    }
    if (agentSkillID === 'cabinet.market_watch.handoff_result') {
      params.provider_id = primary
      if (context) {
        params.result_id = context
      }
      if (secretOrTarget) {
        params.destination = secretOrTarget
      }
      return params
    }
    if (agentSkillID === 'cabinet.purchases.create_order') {
      if (primary) {
        params.purchase_source = primary
        params.source = primary
      }
      if (context) {
        params.item_id = context
      }
      if (secretOrTarget) {
        params.tracking_number = secretOrTarget
        params.source_url = secretOrTarget
      }
    }
    if (agentSkillID.startsWith('cabinet.inventory.')) {
      if (agentSkillID === 'cabinet.inventory.search_items') {
        if (primary) {
          params.query = primary
        }
        if (context) {
          params.status = context
          params.filter = context
        }
      } else if (agentSkillID === 'cabinet.inventory.create_item') {
        if (primary) {
          params.part_number = primary
        }
        if (context) {
          params.title = context
        }
        if (secretOrTarget) {
          params.source_url = secretOrTarget
          params.notes = secretOrTarget
        }
      } else if (agentSkillID === 'cabinet.inventory.update_item') {
        if (primary) {
          params.item_id = primary
        }
        if (context) {
          params.title = context
        }
        if (secretOrTarget) {
          params.notes = secretOrTarget
        }
      } else if (agentSkillID === 'cabinet.inventory.attach_media') {
        if (primary) {
          params.item_id = primary
        }
        if (context) {
          params.media_id = context
        }
        if (secretOrTarget) {
          params.notes = secretOrTarget
        }
      } else if (agentSkillID === 'cabinet.inventory.assign_to_collection') {
        if (primary) {
          params.item_id = primary
        }
        if (context) {
          params.collection_name = context
        }
        if (secretOrTarget) {
          params.notes = secretOrTarget
        }
      }
      return params
    }
    if (agentSkillID.startsWith('cabinet.wishlist.')) {
      if (agentSkillID === 'cabinet.wishlist.create_entry') {
        if (primary) {
          params.title = primary
        }
        if (context) {
          params.target_price = context
        }
        if (secretOrTarget) {
          params.source_url = secretOrTarget
          params.notes = secretOrTarget
        }
      } else if (agentSkillID === 'cabinet.wishlist.mark_purchased') {
        if (primary) {
          params.entry_id = primary
          params.wishlist_entry_id = primary
        }
        if (context) {
          params.quantity = context
        }
        if (secretOrTarget) {
          params.purchase_url = secretOrTarget
          params.notes = secretOrTarget
        }
      } else {
        if (primary) {
          params.entry_id = primary
          params.wishlist_entry_id = primary
          params.query = primary
        }
        if (context) {
          params.title = context
        }
        if (secretOrTarget) {
          params.notes = secretOrTarget
        }
      }
      return params
    }
    if (agentSkillID.startsWith('cabinet.collections.')) {
      if (
        agentSkillID === 'cabinet.collections.create' ||
        agentSkillID === 'cabinet.collections.update_metadata'
      ) {
        if (primary) {
          params.collection_name = primary
          params.name = primary
        }
        if (context) {
          params.description = context
        }
      } else if (agentSkillID === 'cabinet.collections.assign_item') {
        if (primary) {
          params.collection_id = primary
          params.collection_name = primary
        }
        if (context) {
          params.item_id = context
        }
      } else {
        if (primary) {
          params.collection_id = primary
          params.collection_name = primary
          params.query = primary
        }
        if (context) {
          params.destination_collection_id = context
          params.destination_collection_name = context
        }
      }
      if (secretOrTarget) {
        params.notes = secretOrTarget
      }
      return params
    }
    if (agentSkillID === 'cabinet.settings.update_profile') {
      params.settings_profile = {
        display_currency: primary,
        timezone: context,
        profile_private_note: secretOrTarget,
      }
      return params
    }
    if (agentSkillID === 'cabinet.settings.update_account') {
      params.settings_account = {
        account_email: primary,
        locale: context,
        account_private_note: secretOrTarget,
      }
      return params
    }
    if (agentSkillID === 'cabinet.settings.update_appearance') {
      params.setting_key = primary
      params.setting_scope = context || 'appearance'
      params.setting_value = secretOrTarget
      return params
    }
    if (agentSkillID === 'cabinet.storage.configure_backup') {
      if (primary) {
        params.backup_target = primary
      }
      if (context) {
        params.backup_schedule = context
      }
      if (secretOrTarget) {
        params.storage_note = secretOrTarget
      }
      return params
    }
    if (agentSkillID === 'cabinet.maintenance.run_safe_check') {
      if (primary) {
        params.maintenance_scope = primary
      }
      if (context) {
        params.check_level = context
      }
      if (secretOrTarget) {
        params.maintenance_note = secretOrTarget
      }
      return params
    }
    if (agentSkillID === 'cabinet.media.search') {
      if (primary) {
        params.query = primary
      }
      if (context) {
        params.filter = context
      }
      if (secretOrTarget) {
        params.notes = secretOrTarget
      }
      return params
    }
    if (agentSkillID === 'cabinet.media.review_unlinked') {
      if (primary) {
        params.query = primary
      }
      params.filter = context || 'unlinked'
      if (secretOrTarget) {
        params.notes = secretOrTarget
      }
      return params
    }
    if (agentSkillID === 'cabinet.media.upload_or_import') {
      if (primary) {
        params.source_url = primary
        params.file_path = primary
      }
      if (context) {
        params.filename = context
      }
      if (secretOrTarget) {
        params.notes = secretOrTarget
      }
      return params
    }
    if (
      agentSkillID === 'cabinet.media.attach_to_item' ||
      agentSkillID === 'cabinet.media.detach_from_item'
    ) {
      if (primary) {
        params.media_id = primary
      }
      if (context) {
        params.item_id = context
        params.target_item = context
      }
      if (secretOrTarget) {
        params.notes = secretOrTarget
      }
      return params
    }
    if (agentSkillID === 'cabinet.media.update_notes') {
      if (primary) {
        params.media_id = primary
      }
      params.notes = secretOrTarget || context
      return params
    }
    if (agentSkillID.startsWith('cabinet.discoveries.')) {
      if (primary) {
        params.provider_id = primary
      }
      if (agentSkillID === 'cabinet.discoveries.search') {
        if (context) {
          params.query = context
        }
      } else if (context) {
        params.result_id = context
        params.candidate_id = context
      }
      if (secretOrTarget) {
        params.notes = secretOrTarget
        params.destination = secretOrTarget
      }
      return params
    }
    return params
  }

  async function previewAgentSkill() {
    if (!activeProfileId || !threadId) return
    setExecutionState('queued')
    setError('')
    setAgentSkillResult(null)
    setPermissionGuidance(
      'Agent Skill work is preview-first and keeps provider secrets out of result text.'
    )
    try {
      const response = await cabinetProtectedFetch(
        '/api/agent/skills/preview',
        activeProfileId,
        {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          profile_id: activeProfileId,
          skill_id: agentSkillID,
          source_surface: selectedAgentSkillOption.surface,
          source_channel: 'in-app',
          source_thread_id: threadId,
          source_message_id: 'assistant-workspace-agent-skill',
          agent_context: agentContextEnvelope,
          parameters: agentSkillParameters(),
        }),
        }
      )
      if (!response.ok) {
        const payload = (await response
          .json()
          .catch(() => null)) as AgentSkillAuthorityError | null
        const blocker = payload?.error || payload?.authority?.blocker
        if (blocker) {
          throw new Error(
            [
              blocker,
              payload?.authority?.entry_point
                ? `entry=${payload.authority.entry_point}`
                : '',
              payload?.authority?.next_action || '',
            ]
              .filter(Boolean)
              .join(' - ')
          )
        }
        throw new Error(`agent_skill_preview_${response.status}`)
      }
      const preview = (await response.json()) as AgentSkillPreview
      setAgentSkillPreview(preview)
      setExecutionState(preview.confirmation_required ? 'running' : 'success')
    } catch (err) {
      const message =
        err instanceof Error ? err.message : 'agent_skill_preview_failed'
      setExecutionState('failure')
      setWorkflowRunsLoading(false)
      setError(message)
      setPermissionGuidance(
        message.includes(' - ')
          ? message
          : 'Agent Skill preview could not be prepared under the current policy.'
      )
    }
  }

  async function applyPreviewAction() {
    if (!activeProfileId || !threadId || !actionPreview?.id) return
    setExecutionState('running')
    setError('')
    try {
      const response = await cabinetProtectedFetch(
        '/api/chat/actions/apply',
        activeProfileId,
        {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          profile_id: activeProfileId,
          thread_id: threadId,
          preview_id: actionPreview.id,
          confirm: true,
        }),
        }
      )
      if (!response.ok) throw new Error(`assistant_apply_${response.status}`)
      const result = (await response.json()) as ApplyActionResult
      setApplyResult(result)
      setExecutionState('success')
      setConfirmApplyOpen(false)
      await loadMessages(activeProfileId, threadId)
    } catch (err) {
      setExecutionState('failure')
      setWorkflowRunsLoading(false)
      setError(err instanceof Error ? err.message : 'assistant_apply_failed')
      setPermissionGuidance(
        'Apply is confirm-required. If apply remains blocked, the active policy may be preview-only for this action class.'
      )
    }
  }

  async function applyAgentSkillPreview() {
    if (!activeProfileId || !threadId || !agentSkillPreview?.skill_id) return
    setExecutionState('running')
    setError('')
    try {
      const response = await cabinetProtectedFetch(
        '/api/agent/skills/apply',
        activeProfileId,
        {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          profile_id: activeProfileId,
          skill_id: agentSkillPreview.skill_id,
          confirm: true,
          source_surface: selectedAgentSkillOption.surface,
          source_channel: 'in-app',
          source_thread_id: threadId,
          source_message_id: 'assistant-workspace-agent-skill',
          agent_context: agentContextEnvelope,
          parameters: agentSkillParameters(),
        }),
        }
      )
      if (!response.ok) {
        const payload = (await response
          .json()
          .catch(() => null)) as AgentSkillAuthorityError | null
        const blocker = payload?.error || payload?.authority?.blocker
        if (blocker) {
          throw new Error(
            [
              blocker,
              payload?.authority?.entry_point
                ? `entry=${payload.authority.entry_point}`
                : '',
              payload?.authority?.next_action || '',
            ]
              .filter(Boolean)
              .join(' - ')
          )
        }
        throw new Error(`agent_skill_apply_${response.status}`)
      }
      const result = (await response.json()) as AgentSkillApplyResult
      setAgentSkillResult(result)
      setExecutionState('success')
      setConfirmApplyOpen(false)
    } catch (err) {
      const message =
        err instanceof Error ? err.message : 'agent_skill_apply_failed'
      setExecutionState('failure')
      setWorkflowRunsLoading(false)
      setError(message)
      setPermissionGuidance(
        message.includes(' - ')
          ? message
          : 'Agent Skill apply is confirm-required and may stay blocked until setup targets are complete.'
      )
    }
  }

  const applyLink = resultLink(applyResult)

  return (
    <div className='px-2 py-2' data-testid='shell-assistant-workspace'>
      <AssistantRuntimeProvider runtime={assistantRuntime}>
        <AssistantModalPrimitive.Root
          defaultOpen
          unstable_openOnRunStart={false}
        >
          <AssistantModalPrimitive.Anchor
            className='block'
            data-testid='shell-assistant-modal-anchor'
          >
            <AssistantModalPrimitive.Trigger asChild>
              <Button
                type='button'
                variant='outline'
                className='w-full justify-start gap-2'
                data-testid='shell-assistant-modal-trigger'
              >
                <Sparkles className='h-4 w-4' />
                Assistant workspace
              </Button>
            </AssistantModalPrimitive.Trigger>
          </AssistantModalPrimitive.Anchor>
          <AssistantModalPrimitive.Content
            side='right'
            align='start'
            sideOffset={8}
            collisionPadding={12}
            className='z-[1000] flex h-[min(46rem,calc(100vh-6rem))] w-[min(22rem,calc(100vw-1.5rem))] overflow-hidden rounded-xl border border-slate-800 bg-slate-950 text-slate-100 shadow-2xl outline-none'
            data-testid='shell-assistant-modal-content'
          >
            <section
              className='flex min-h-0 flex-1 flex-col overflow-hidden'
              data-testid='shell-assistant-codex-chat'
            >
              <div
                className='border-b border-slate-800 bg-slate-950 p-3'
                data-testid='shell-chat-rail'
              >
                <div className='flex items-center justify-between gap-2'>
                  <h2
                    className='min-w-0 text-xl font-semibold tracking-normal text-slate-100'
                    data-testid='shell-assistant-panel-title'
                  >
                    <span>Chat</span>
                  </h2>
                  <div className='flex items-center gap-1'>
                    <Button
                      type='button'
                      variant='ghost'
                      size='icon'
                      data-testid='shell-assistant-new-thread'
                      aria-label='New assistant thread'
                      title='New assistant thread'
                      className='h-8 w-8 text-slate-300 hover:bg-slate-800 hover:text-white'
                      onClick={() => void handleNewThread()}
                      disabled={loading || sending || !activeProfileId}
                    >
                      <MessageSquarePlus className='h-3.5 w-3.5' />
                    </Button>
                    <Button
                      type='button'
                      variant='ghost'
                      size='icon'
                      data-testid='shell-assistant-mute-toggle'
                      aria-label='Mute assistant workspace updates'
                      title='Mute assistant workspace updates'
                      className='h-8 w-8 text-slate-300 hover:bg-slate-800 hover:text-white'
                    >
                      <VolumeX className='h-3.5 w-3.5' />
                    </Button>
                    <Button
                      type='button'
                      variant='ghost'
                      size='icon'
                      data-testid='shell-assistant-close'
                      aria-label='Close assistant workspace'
                      title='Close assistant workspace'
                      className='h-8 w-8 text-slate-300 hover:bg-slate-800 hover:text-white'
                      onClick={() => setActiveWorkspace('navigation')}
                    >
                      <X className='h-3.5 w-3.5' />
                    </Button>
                  </div>
                </div>

                <div
                  className='mt-3 rounded-lg border border-slate-800 bg-slate-900/70 p-3'
                  data-testid='shell-assistant-identity-card'
                >
                  <div className='flex items-start gap-2'>
                    <div className='flex h-8 w-8 shrink-0 items-center justify-center rounded-md border border-slate-700 bg-slate-950 text-cyan-300'>
                      <Bot className='h-4 w-4' />
                    </div>
                    <div className='min-w-0 flex-1'>
                      <p
                        className='truncate text-sm font-medium text-slate-100'
                        data-testid='shell-assistant-agent-name'
                      >
                        Cabinet Agent
                      </p>
                      <p
                        className='text-xs text-slate-400'
                        data-testid='shell-assistant-agent-role'
                      >
                        Single app-control agent
                      </p>
                      <p
                        className='mt-1 flex items-center gap-1 text-[11px] text-emerald-300'
                        data-testid='shell-assistant-runtime-state'
                      >
                        <span className='h-1.5 w-1.5 rounded-full bg-emerald-300' />
                        {activeProfileId
                          ? 'Agent runtime connected'
                          : 'Waiting for profile'}
                      </p>
                    </div>
                  </div>
                </div>

                <label className='mt-3 block space-y-1 text-xs'>
                  <span className='font-medium text-slate-300'>
                    Conversation
                  </span>
                  <select
                    data-testid='shell-assistant-thread-select'
                    className='w-full rounded-md border border-slate-700 bg-slate-900 px-2 py-1.5 text-slate-100'
                    value={threadId}
                    onChange={(e) => void handleSelectThread(e.target.value)}
                    disabled={loading || sending || threads.length === 0}
                  >
                    {threads.length === 0 ? (
                      <option value=''>No assistant chats yet</option>
                    ) : null}
                    {threads.map((thread) => (
                      <option key={thread.id} value={thread.id}>
                        {thread.title}
                      </option>
                    ))}
                  </select>
                </label>

                <div className='mt-3 flex flex-wrap gap-2 text-[11px]'>
                  <Badge
                    variant='outline'
                    data-testid='shell-assistant-context-chip'
                    className='max-w-full justify-start gap-1 truncate border-slate-700 bg-slate-900 text-slate-300'
                  >
                    <GitBranchPlus className='h-3 w-3 shrink-0' />
                    <span
                      className='truncate'
                      data-testid='shell-assistant-route-context'
                    >{`${routeContext.pathname}${routeContext.search}`}</span>
                  </Badge>
                  <Badge
                    variant='secondary'
                    data-testid='shell-assistant-model-chip'
                    className='max-w-full justify-start gap-1 truncate bg-slate-800 text-slate-200'
                  >
                    <Bot className='h-3 w-3 shrink-0' />
                    <span
                      className='truncate'
                      data-testid='shell-assistant-thread-provider'
                    >
                      {threadMetadata.provider || provider}
                    </span>
                    <span>/</span>
                    <span
                      className='truncate'
                      data-testid='shell-assistant-thread-model'
                    >
                      {threadMetadata.model || model}
                    </span>
                  </Badge>
                </div>

                <div className='mt-3 grid grid-cols-2 gap-2 text-xs'>
                  <label className='space-y-1'>
                    <span className='font-medium text-slate-300'>Provider</span>
                    <select
                      data-testid='shell-assistant-provider-select'
                      className='w-full rounded-md border border-slate-700 bg-slate-900 px-2 py-1 text-slate-100'
                      value={provider}
                      onChange={(e) =>
                        void handleProviderChange(e.target.value)
                      }
                    >
                      {assistantProviderOptions.map((option) => (
                        <option key={option.provider} value={option.provider}>
                          {option.label}
                        </option>
                      ))}
                    </select>
                  </label>
                  <label className='space-y-1'>
                    <span className='font-medium text-slate-300'>Model</span>
                    <select
                      data-testid='shell-assistant-model-select'
                      className='w-full rounded-md border border-slate-700 bg-slate-900 px-2 py-1 text-slate-100'
                      value={model}
                      onChange={(e) => void handleModelChange(e.target.value)}
                    >
                      {availableModels.map((option) => (
                        <option key={option.value} value={option.value}>
                          {option.label}
                        </option>
                      ))}
                    </select>
                  </label>
                </div>

                <div className='mt-3 grid gap-1 text-[11px] text-slate-400'>
                  <span>
                    Profile:{' '}
                    <span data-testid='shell-assistant-profile-scope'>
                      {activeProfileId}
                    </span>
                  </span>
                  <span data-testid='shell-assistant-selection-context'>
                    Collection: {selectionContext}
                  </span>
                  <span data-testid='shell-assistant-selected-record-context'>
                    Selected:{' '}
                    {selectedAgentRecordContext
                      ? `${selectedAgentRecordContext.type}:${selectedAgentRecordContext.label || selectedAgentRecordContext.id}`
                      : 'None'}
                  </span>
                  <span data-testid='shell-assistant-thread-id'>
                    {threadId || 'bootstrapping'}
                  </span>
                  <span data-testid='shell-assistant-selected-thread-title'>
                    Conversation: {selectedThreadTitle}
                  </span>
                  <span data-testid='shell-assistant-boundary-note'>
                    Thread continuity persists across authenticated route
                    changes until an explicit reset boundary.
                  </span>
                  <span data-testid='shell-assistant-thread-semantics'>
                    Provider/model changes fork a new assistant thread; manual
                    reset creates a clean thread for the current profile.
                  </span>
                </div>
                {displayedNavigationAction ? (
                  <div
                    className='mt-3 rounded-md border border-slate-800 bg-slate-900 p-3 text-sm'
                    data-testid='shell-assistant-navigation-action'
                  >
                    <div className='flex items-start gap-2'>
                      <ExternalLink className='mt-0.5 h-4 w-4 text-primary' />
                      <div className='min-w-0 flex-1'>
                        <p className='font-medium'>
                          {displayedNavigationAction.label}
                        </p>
                        <p
                          className='mt-1 text-xs text-muted-foreground'
                          data-testid='shell-assistant-navigation-reason'
                        >
                          {displayedNavigationAction.reason}
                        </p>
                        <Button
                          type='button'
                          size='sm'
                          variant='outline'
                          className='mt-2 border-slate-700 bg-slate-950 text-slate-100 hover:bg-slate-800'
                          data-testid='shell-assistant-navigation-action-open'
                          onClick={() =>
                            void runNavigationAction(displayedNavigationAction)
                          }
                        >
                          <ExternalLink className='h-3.5 w-3.5' />
                          Open screen
                        </Button>
                        {displayedNavigationAction.target.startsWith(
                          '/inventory'
                        ) ? (
                          <Button
                            type='button'
                            size='sm'
                            variant='outline'
                            className='mt-2 ml-2 border-slate-700 bg-slate-950 text-slate-100 hover:bg-slate-800'
                            data-testid='shell-assistant-navigation-action-highlight'
                            onClick={() =>
                              void runHighlightCommand('inventory.surface')
                            }
                          >
                            <Sparkles className='h-3.5 w-3.5' />
                            Show target
                          </Button>
                        ) : null}
                      </div>
                    </div>
                  </div>
                ) : null}
                {guidedWorkflow ? (
                  <div
                    className='mt-3 rounded-md border border-cyan-800/70 bg-cyan-950/30 p-3 text-sm'
                    data-testid='shell-assistant-guided-walkthrough'
                    data-guided-mode={guidedWorkflow.mode}
                    data-guided-recipe={guidedWorkflow.recipe_id}
                  >
                    <div className='flex items-start gap-2'>
                      <Sparkles className='mt-0.5 h-4 w-4 shrink-0 text-cyan-300' />
                      <div className='min-w-0 flex-1'>
                        <p
                          className='font-medium text-slate-100'
                          data-testid='shell-assistant-guided-title'
                        >
                          {guidedWorkflow.title}
                        </p>
                        <p
                          className='mt-1 text-xs text-slate-300'
                          data-testid='shell-assistant-guided-boundary'
                        >
                          {guidedWorkflow.mutation_boundary}
                        </p>
                        <div
                          className='mt-2 flex flex-wrap gap-1'
                          data-testid='shell-assistant-guided-steps'
                        >
                          {guidedWorkflow.steps.map((step) => (
                            <span
                              key={step.id}
                              className='rounded border border-cyan-800/60 bg-slate-950 px-2 py-1 text-[11px] text-slate-300'
                              data-testid='shell-assistant-guided-step'
                              data-guided-command={step.command}
                              data-guided-target={step.ui_target_id || ''}
                            >
                              {step.title}
                            </span>
                          ))}
                        </div>
                        <div className='mt-3 flex flex-wrap gap-2'>
                          <Button
                            type='button'
                            size='sm'
                            variant='outline'
                            className='border-cyan-800 bg-slate-950 text-slate-100 hover:bg-slate-900'
                            data-testid='shell-assistant-guided-start'
                            onClick={() =>
                              void runGuidedWalkthrough(guidedWorkflow)
                            }
                          >
                            <Play className='h-3.5 w-3.5' />
                            Start
                          </Button>
                          <Button
                            type='button'
                            size='icon'
                            variant='outline'
                            className='h-8 w-8 border-slate-700 bg-slate-950 text-slate-300'
                            aria-label='Previous walkthrough step'
                            title='Previous walkthrough step'
                            data-testid='shell-assistant-guided-back'
                            onClick={() => {
                              void dispatchGuidedControl(
                                'walkthrough.step_back'
                              )
                              if (guidedWorkflow.steps[0]) {
                                void dispatchGuidedStep(guidedWorkflow.steps[0])
                              }
                            }}
                          >
                            <ChevronLeft className='h-3.5 w-3.5' />
                          </Button>
                          <Button
                            type='button'
                            size='icon'
                            variant='outline'
                            className='h-8 w-8 border-slate-700 bg-slate-950 text-slate-300'
                            aria-label={
                              guidedPaused
                                ? 'Resume walkthrough'
                                : 'Pause walkthrough'
                            }
                            title={
                              guidedPaused
                                ? 'Resume walkthrough'
                                : 'Pause walkthrough'
                            }
                            data-testid='shell-assistant-guided-pause'
                            data-guided-paused={guidedPaused ? 'true' : 'false'}
                            onClick={() => {
                              const nextPaused = !guidedPaused
                              setGuidedPaused(nextPaused)
                              void dispatchGuidedControl(
                                nextPaused
                                  ? 'walkthrough.pause'
                                  : 'walkthrough.resume'
                              )
                            }}
                          >
                            {guidedPaused ? (
                              <Play className='h-3.5 w-3.5' />
                            ) : (
                              <Pause className='h-3.5 w-3.5' />
                            )}
                          </Button>
                          <Button
                            type='button'
                            size='icon'
                            variant='outline'
                            className='h-8 w-8 border-slate-700 bg-slate-950 text-slate-300'
                            aria-label='Skip walkthrough step'
                            title='Skip walkthrough step'
                            data-testid='shell-assistant-guided-skip'
                            onClick={() => {
                              void dispatchGuidedControl('walkthrough.skip')
                              if (
                                guidedWorkflow.steps[
                                  guidedWorkflow.steps.length - 1
                                ]
                              ) {
                                void dispatchGuidedStep(
                                  guidedWorkflow.steps[
                                    guidedWorkflow.steps.length - 1
                                  ]
                                )
                              }
                            }}
                          >
                            <SkipForward className='h-3.5 w-3.5' />
                          </Button>
                          <Button
                            type='button'
                            size='icon'
                            variant='outline'
                            className='h-8 w-8 border-slate-700 bg-slate-950 text-slate-300'
                            aria-label='Next walkthrough step'
                            title='Next walkthrough step'
                            data-testid='shell-assistant-guided-next'
                            onClick={() =>
                              guidedWorkflow.steps[1]
                                ? void dispatchGuidedStep(
                                    guidedWorkflow.steps[1]
                                  )
                                : undefined
                            }
                          >
                            <ChevronRight className='h-3.5 w-3.5' />
                          </Button>
                          <Button
                            type='button'
                            size='sm'
                            variant='outline'
                            className='border-slate-700 bg-slate-950 text-slate-100 hover:bg-slate-900'
                            data-testid='shell-assistant-guided-cancel'
                            onClick={() => void cancelGuidance()}
                          >
                            <X className='h-3.5 w-3.5' />
                            Cancel
                          </Button>
                        </div>
                      </div>
                    </div>
                  </div>
                ) : null}
              </div>

              <ScrollArea className='min-h-0 flex-1 bg-slate-950 p-3'>
                <div
                  className='space-y-4 pb-2'
                  data-testid='shell-assistant-message-list'
                >
                  {loading ? (
                    <p className='text-sm text-slate-400'>
                      Loading assistant workspace...
                    </p>
                  ) : null}
                  {!loading && messages.length === 0 ? (
                    <div className='rounded-lg border border-dashed border-slate-700 bg-slate-900/60 p-3 text-sm text-slate-400'>
                      Ask Cabinet to update records, create drafts, search
                      inventory, and return links to the items it touched.
                    </div>
                  ) : null}
                  <CabinetAssistantUiMessageList messages={messages} />

                  <div
                    className='rounded-lg border border-slate-800 bg-slate-900/70 p-3 text-xs'
                    data-testid='shell-assistant-execution-panel'
                  >
                    <div className='flex items-center justify-between gap-2'>
                      <div className='flex items-center gap-2 font-medium'>
                        <Wand2 className='h-4 w-4 text-cyan-300' />
                        Agent actions
                      </div>
                      <Badge
                        variant='outline'
                        className='border-slate-700 bg-slate-950 text-slate-300 uppercase'
                        data-testid='shell-assistant-execution-state'
                      >
                        {executionState}
                      </Badge>
                    </div>
                    <p
                      className='mt-2 text-slate-400'
                      data-testid='shell-assistant-permission-guidance'
                    >
                      {displayedPermissionGuidance}
                    </p>
                    <div className='mt-3 grid gap-2'>
                      <Input
                        data-testid='shell-assistant-preview-part-number'
                        className='border-slate-700 bg-slate-950 text-slate-100'
                        value={actionPartNumber}
                        onChange={(e) => setActionPartNumber(e.target.value)}
                        placeholder='Part number'
                        disabled={!threadId || sending}
                      />
                      <Input
                        data-testid='shell-assistant-preview-title'
                        className='border-slate-700 bg-slate-950 text-slate-100'
                        value={actionTitle}
                        onChange={(e) => setActionTitle(e.target.value)}
                        placeholder='Item title'
                        disabled={!threadId || sending}
                      />
                    </div>
                    <div className='mt-3 flex flex-wrap gap-2'>
                      <Button
                        type='button'
                        size='sm'
                        data-testid='shell-assistant-preview-action'
                        onClick={() => void previewAction()}
                        disabled={
                          !threadId ||
                          !actionPartNumber.trim() ||
                          !actionTitle.trim() ||
                          sending
                        }
                      >
                        Preview
                      </Button>
                      <Button
                        type='button'
                        size='sm'
                        variant='outline'
                        data-testid='shell-assistant-apply-action'
                        onClick={() => setConfirmApplyOpen(true)}
                        disabled={!actionPreview?.id || sending}
                      >
                        Apply
                      </Button>
                    </div>
                    {actionPreview ? (
                      <div
                        className='mt-3 rounded-lg border border-slate-800 bg-slate-950 p-2'
                        data-testid='shell-assistant-action-card'
                      >
                        <div
                          className='text-xs'
                          data-testid='shell-assistant-action-preview'
                        >
                          Preview {actionPreview.action} ({actionPreview.status}
                          ) for {actionPreview.payload?.part_number} /{' '}
                          {actionPreview.payload?.title}
                        </div>
                      </div>
                    ) : null}
                    {applyResult ? (
                      <div
                        className='mt-3 rounded-lg border border-slate-800 bg-slate-950 p-2'
                        data-testid='shell-assistant-apply-result'
                      >
                        <div className='flex items-start gap-2'>
                          <CheckCircle2 className='mt-0.5 h-4 w-4 text-primary' />
                          <div className='min-w-0 flex-1'>
                            <p>
                              Applied {applyResult.action}{' '}
                              {applyResult.item_id
                                ? `to ${applyResult.item_id}`
                                : ''}
                            </p>
                            {applyLink ? (
                              <Button
                                type='button'
                                variant='link'
                                size='sm'
                                asChild
                                className='h-auto px-0 text-xs'
                              >
                                <a
                                  href={applyLink.href}
                                  data-testid='shell-assistant-result-link'
                                >
                                  <ExternalLink className='h-3.5 w-3.5' />
                                  {applyLink.label}
                                </a>
                              </Button>
                            ) : null}
                          </div>
                        </div>
                      </div>
                    ) : null}
                    <div
                      className='mt-3 rounded-lg border border-slate-800 bg-slate-950 p-3'
                      data-testid='shell-assistant-agent-skill-panel'
                    >
                      <div className='flex items-center justify-between gap-2'>
                        <p className='font-medium text-slate-100'>
                          Agent Skill
                        </p>
                        <Badge
                          variant='outline'
                          className='border-slate-700 bg-slate-900 text-slate-300'
                          data-testid='shell-assistant-agent-skill-policy'
                        >
                          preview-first
                        </Badge>
                      </div>
                      <div className='mt-3 grid gap-2'>
                        <select
                          data-testid='shell-assistant-agent-skill-select'
                          className='w-full rounded-md border border-slate-700 bg-slate-900 px-2 py-1.5 text-slate-100'
                          value={agentSkillID}
                          onChange={(event) => {
                            setAgentSkillID(event.target.value)
                            setAgentSkillPreview(null)
                            setAgentSkillResult(null)
                          }}
                          disabled={!threadId || sending}
                        >
                          {agentSkillOptions.map((option) => (
                            <option key={option.id} value={option.id}>
                              {option.label}
                            </option>
                          ))}
                        </select>
                        <Input
                          data-testid='shell-assistant-agent-skill-provider'
                          className='border-slate-700 bg-slate-900 text-slate-100'
                          value={agentSkillProvider}
                          onChange={(event) =>
                            setAgentSkillProvider(event.target.value)
                          }
                          placeholder={selectedAgentSkillOption.primaryLabel}
                          disabled={!threadId || sending}
                        />
                        <Input
                          data-testid='shell-assistant-agent-skill-setup-step'
                          className='border-slate-700 bg-slate-900 text-slate-100'
                          value={agentSkillSetupStep}
                          onChange={(event) =>
                            setAgentSkillSetupStep(event.target.value)
                          }
                          placeholder={selectedAgentSkillOption.contextLabel}
                          disabled={!threadId || sending}
                        />
                        <Input
                          data-testid='shell-assistant-agent-skill-secret'
                          className='border-slate-700 bg-slate-900 text-slate-100'
                          value={agentSkillSecret}
                          onChange={(event) =>
                            setAgentSkillSecret(event.target.value)
                          }
                          placeholder={selectedAgentSkillOption.secretLabel}
                          disabled={!threadId || sending}
                        />
                      </div>
                      <div className='mt-3 flex flex-wrap gap-2'>
                        <Button
                          type='button'
                          size='sm'
                          variant='outline'
                          data-testid='shell-assistant-agent-skill-preview'
                          onClick={() => void previewAgentSkill()}
                          disabled={
                            !threadId || !agentSkillProvider.trim() || sending
                          }
                        >
                          Preview skill
                        </Button>
                        <Button
                          type='button'
                          size='sm'
                          variant='outline'
                          data-testid='shell-assistant-agent-skill-apply'
                          onClick={() => {
                            setConfirmTarget('agent-skill')
                            setConfirmApplyOpen(true)
                          }}
                          disabled={!agentSkillPreview || sending}
                        >
                          Apply skill
                        </Button>
                      </div>
                      {agentSkillPreview ? (
                        <div
                          className='mt-3 rounded-md border border-slate-800 bg-slate-900 p-2 text-slate-300'
                          data-testid='shell-assistant-agent-skill-preview-card'
                        >
                          {agentSkillPreview.skill_id} /{' '}
                          {agentSkillPreview.safety_level} /{' '}
                          {agentSkillPreview.blocker || 'ready'}
                        </div>
                      ) : null}
                      {agentSkillResult ? (
                        <div
                          className='mt-3 rounded-md border border-slate-800 bg-slate-900 p-2 text-slate-300'
                          data-testid='shell-assistant-agent-skill-result'
                        >
                          {agentSkillResult.target?.operation} / mutation:{' '}
                          {String(agentSkillResult.mutation_applied)} / secret
                          redacted:{' '}
                          {String(agentSkillResult.target?.secret_redacted)}
                        </div>
                      ) : null}
                    </div>
                    <div
                      className='mt-3 flex items-start gap-2 rounded-lg border border-dashed border-slate-700 p-2 text-slate-400'
                      data-testid='shell-assistant-permission-boundary'
                    >
                      <ShieldAlert className='mt-0.5 h-4 w-4 shrink-0' />
                      <div>
                        <p className='font-medium text-slate-100'>
                          Permission boundary
                        </p>
                        <p>
                          Read-only is always allowed. Mutations are
                          preview-first, confirm-required, and may still be
                          unavailable under the active policy.
                        </p>
                      </div>
                    </div>
                  </div>
                </div>
              </ScrollArea>

              {error ? (
                <p
                  className='px-3 pb-2 text-xs text-red-300'
                  data-testid='shell-assistant-error'
                >
                  {error}
                </p>
              ) : null}

              <div className='border-t border-slate-800 bg-slate-950 p-3'>
                <div
                  className='mb-3 space-y-2 rounded-lg border border-slate-800 bg-slate-900/70 p-2 text-xs text-slate-300'
                  data-testid='shell-assistant-attachment-panel'
                >
                  <div className='flex items-center gap-2'>
                    <Button
                      type='button'
                      variant='outline'
                      size='sm'
                      className='border-slate-700 bg-slate-950 text-slate-100 hover:bg-slate-800'
                      data-testid='shell-assistant-attachment-picker'
                      onClick={() =>
                        document
                          .querySelector<HTMLInputElement>(
                            '[data-testid="shell-assistant-attachment-input"]'
                          )
                          ?.click()
                      }
                      disabled={!threadId || loading || sending}
                    >
                      <Paperclip className='h-3.5 w-3.5' />
                      Attach
                    </Button>
                    <Input
                      data-testid='shell-assistant-attachment-input'
                      type='file'
                      className='hidden'
                      onChange={(event) =>
                        setPendingAttachment(event.target.files?.[0] ?? null)
                      }
                    />
                    <Button
                      type='button'
                      variant='outline'
                      size='sm'
                      className='border-slate-700 bg-slate-950 text-slate-100 hover:bg-slate-800'
                      data-testid='shell-assistant-attachment-upload'
                      onClick={() => void uploadAttachment()}
                      disabled={!threadId || !pendingAttachment || sending}
                    >
                      Upload
                    </Button>
                    {pendingAttachment ? (
                      <span
                        className='min-w-0 truncate text-slate-400'
                        data-testid='shell-assistant-pending-attachment'
                      >
                        {pendingAttachment.name}
                      </span>
                    ) : null}
                  </div>
                  {attachments.length > 0 ? (
                    <div
                      className='space-y-1'
                      data-testid='shell-assistant-attachment-list'
                    >
                      {attachments.map((attachment) => (
                        <div
                          key={attachment.id}
                          className='flex items-center justify-between gap-2 rounded border border-slate-800 bg-slate-950 px-2 py-1'
                          data-attachment-id={attachment.id}
                        >
                          <span className='min-w-0 truncate'>
                            {attachment.filename}
                          </span>
                          <span className='shrink-0 text-[10px] text-slate-500'>
                            {attachment.mime_type || 'file'} /{' '}
                            {attachment.size_bytes} bytes
                          </span>
                          <Button
                            type='button'
                            size='icon'
                            variant='ghost'
                            className='h-6 w-6 shrink-0 text-slate-500 hover:bg-slate-800 hover:text-slate-100'
                            data-testid='shell-assistant-remove-attachment'
                            aria-label={`Remove ${attachment.filename}`}
                            onClick={() => removeAttachment(attachment.id)}
                          >
                            <X className='h-3.5 w-3.5' />
                          </Button>
                        </div>
                      ))}
                    </div>
                  ) : null}
                </div>
                <CabinetAssistantUiComposer
                  composer={{
                    disabled: !threadId || loading || sending,
                    sending,
                  }}
                />
                <details
                  className='mt-3 rounded-lg border border-slate-800 bg-slate-900/60 px-3 py-2 text-xs text-slate-300'
                  data-testid='shell-assistant-action-timeline'
                >
                  <summary className='cursor-pointer list-none font-medium'>
                    Action Timeline
                  </summary>
                  <div
                    className='mt-2 max-h-36 space-y-2 overflow-y-auto text-slate-400'
                    data-testid='shell-assistant-command-timeline'
                  >
                    {workflowRunsLoading ? (
                      <p>Loading durable workflow records...</p>
                    ) : null}
                    {!workflowRunsLoading &&
                    workflowRuns.length === 0 &&
                    commandEvents.length === 0 ? (
                      <p>
                        Durable workflow records appear here after Cabinet
                        plans, previews, applies, cancels, or fails an assistant
                        action.
                      </p>
                    ) : null}
                    {workflowRunsError && workflowRuns.length === 0 ? (
                      <p
                        className='text-red-300'
                        data-testid='shell-assistant-action-timeline-error'
                      >
                        {workflowRunsError}
                      </p>
                    ) : null}
                    {workflowRuns.map((run) => (
                      <div
                        key={run.id}
                        className='rounded border border-slate-800 bg-slate-950 px-2 py-1'
                        data-testid='shell-assistant-workflow-run'
                        data-workflow-status={run.status}
                        data-capability-id={run.capability_id}
                      >
                        <div className='flex items-center justify-between gap-2'>
                          <span className='font-medium text-slate-200'>
                            {run.status}
                          </span>
                          <span className='text-[10px] text-slate-500 uppercase'>
                            {run.confirmation_state}
                          </span>
                        </div>
                        <p>{run.capability_id}</p>
                        <p className='text-slate-500'>
                          {workflowRunResultSummary(run)}
                        </p>
                        <p className='text-[10px] text-slate-600'>
                          {workflowRunTimestamp(run)}
                        </p>
                      </div>
                    ))}
                    {commandEvents.map((event, index) => (
                      <div
                        key={`${event.id}-${event.status}-${index}`}
                        className='rounded border border-slate-800 bg-slate-950 px-2 py-1'
                        data-testid='shell-assistant-command-event'
                        data-command-type={event.type}
                        data-command-status={event.status}
                      >
                        <span className='font-medium text-slate-200'>
                          {event.status}
                        </span>{' '}
                        {event.type}: {event.message}
                      </div>
                    ))}
                  </div>
                  {commandEvents.length > 0 ? (
                    <Button
                      type='button'
                      size='sm'
                      variant='outline'
                      className='mt-2 border-slate-700 bg-slate-950 text-slate-100 hover:bg-slate-800'
                      data-testid='shell-assistant-command-cancel'
                      onClick={() => void cancelGuidance()}
                    >
                      Cancel guidance
                    </Button>
                  ) : null}
                </details>
              </div>
            </section>
          </AssistantModalPrimitive.Content>
        </AssistantModalPrimitive.Root>
      </AssistantRuntimeProvider>

      <AlertDialog open={confirmApplyOpen} onOpenChange={setConfirmApplyOpen}>
        <AlertDialogContent data-testid='shell-assistant-apply-confirm-dialog'>
          <AlertDialogHeader>
            <AlertDialogTitle>Confirm Assistant Action</AlertDialogTitle>
            <AlertDialogDescription data-testid='shell-assistant-apply-confirm-summary'>
              {confirmTarget === 'agent-skill' && agentSkillPreview
                ? `Apply ${agentSkillPreview.skill_id} for ${selectedAgentSkillOption.primaryLabel.toLowerCase()}=${agentSkillProvider.trim()} from ${selectedAgentSkillOption.surface}`
                : actionPreview
                  ? `Apply ${actionPreview.action} with part_number=${String(actionPreview.payload?.part_number ?? 'n/a')} title=${String(actionPreview.payload?.title ?? 'n/a')}`
                  : 'No action preview selected.'}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel data-testid='shell-assistant-apply-cancel'>
              Cancel
            </AlertDialogCancel>
            <AlertDialogAction
              data-testid='shell-assistant-apply-confirm'
              onClick={(event) => {
                event.preventDefault()
                if (confirmTarget === 'agent-skill') {
                  void applyAgentSkillPreview()
                } else {
                  void applyPreviewAction()
                }
              }}
            >
              Confirm Apply
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
