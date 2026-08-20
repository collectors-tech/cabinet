import { useCallback, useEffect, useMemo, useState } from 'react'
import {
  AlertTriangle,
  CheckCircle2,
  LoaderCircle,
  Settings2,
  ShieldCheck,
} from 'lucide-react'
import { cabinetProtectedFetch } from '@/lib/cabinet-session'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'

type AgentCapabilitySummary = {
  total?: number
  available?: number
  confirm_required?: number
  blocked_by_policy?: number
  setup_required?: number
  disabled?: number
  unavailable?: number
}

type AgentCapability = {
  skill_id?: string
  display_name?: string
  capability_state?: string
  next_action?: string
}

export type AgentCapabilitiesContext = {
  mode?: string
  source_surface?: string
  source_channel?: string
  summary?: AgentCapabilitySummary
  explanation?: {
    authority_mode?: string
    no_secret_values?: boolean
    capabilities?: AgentCapability[]
  }
}

export type AgentPlannerContext = {
  mode?: string
  provider?: string
  decision?: string
  skill_id?: string
  message?: string
  next_action?: string
  setup_next_action?: string
  recoverable?: boolean
  confirmation_state?: string
  error?: { code?: string; message?: string }
  preview_result?: {
    kind?: string
    action?: string
    status?: string
    preview_id?: string
    preview_status?: string
    expires_at?: string
    skill_id?: string
    target?: Record<string, unknown>
    confirmation_required?: boolean
    strong_confirmation_required?: boolean
    strong_confirmation_endpoint?: string
    strong_confirmation_request?: Record<string, unknown>
    mutation_applied?: boolean
    apply_endpoint?: string
    apply_request?: Record<string, unknown>
    cancel_endpoint?: string
    cancel_request?: Record<string, unknown>
    retrieval_endpoint?: string
  }
  execution_result?: {
    mutation_applied?: boolean
  }
}

export type AgentResponseState =
  | 'read_result'
  | 'clarification_required'
  | 'setup_required'
  | 'authority_blocked'
  | 'unsupported'
  | 'provider_unavailable'
  | 'retryable_failure'
  | 'preview_required'
  | 'preview_expired'
  | 'preview_failed'
  | 'preview_stale_target'
  | 'cancelled'
  | 'applied'

export type NormalizedAgentResponse = {
  state: AgentResponseState
  outcome:
    | 'success'
    | 'needs_input'
    | 'blocked'
    | 'preview'
    | 'cancelled'
    | 'applied'
    | 'failed'
  title: string
  message: string
  retryable: boolean
  original_intent?: string
  skill?: { id?: string; name?: string }
  source?: { surface?: string; channel?: string }
  next_action?: { kind?: string; label?: string; route?: string }
  preview?: {
    id?: string
    action?: string
    status?: string
    payload?: Record<string, unknown>
  }
  result_summary?: {
    kind?: string
    total?: number
    metrics?: Array<{
      id?: string
      label?: string
      value?: number
      route?: string
    }>
    items?: Array<{
      id?: string
      part_number?: string
      title?: string
      status?: string
      category?: string
      brand?: string
    }>
  }
}

export type AgentResponseMessage = {
  role?: string
  context?: {
    agent_capabilities?: AgentCapabilitiesContext
    agent_planner?: AgentPlannerContext
    agent_response?: NormalizedAgentResponse
  }
}

function latestAgentResponse(messages: AgentResponseMessage[]) {
  for (let index = messages.length - 1; index >= 0; index -= 1) {
    if (messages[index].role?.trim().toLowerCase() !== 'assistant') continue
    return messages[index].context
  }
  return undefined
}

function NormalizedAgentResponseCard({
  response,
  testID,
  allowApply,
  onRetry,
  onApply,
  onAction,
}: {
  response: NormalizedAgentResponse
  testID: string
  allowApply: boolean
  onRetry?: (intent: string) => void | Promise<void>
  onApply?: () => void | Promise<void>
  onAction?: (response: NormalizedAgentResponse) => void | Promise<void>
}) {
  const action = response.next_action
  const canRetry =
    response.retryable === true &&
    action?.kind === 'retry' &&
    Boolean(response.original_intent?.trim())
  const canApply =
    allowApply &&
    response.state === 'preview_required' &&
    action?.kind === 'apply'
  return (
    <section
      className='rounded-md border border-slate-700 bg-slate-900 p-3 text-sm text-slate-100'
      data-testid={testID}
      data-agent-state={response.state}
      data-agent-outcome={response.outcome}
    >
      <p className='font-medium'>{response.title}</p>
      <p className='mt-1 text-xs text-slate-400'>
        {response.state.replace(/_/g, ' ')}
      </p>
      <p className='mt-2'>{response.message}</p>
      {response.state === 'read_result' && response.result_summary ? (
        <div
          className='mt-3 rounded-md border border-cyan-500/30 bg-cyan-950/20 p-2'
          data-testid={`${testID}-result-summary`}
        >
          <p className='text-xs text-cyan-100'>
            {response.result_summary.total === 0
              ? 'No matching records.'
              : `${response.result_summary.total ?? response.result_summary.items?.length ?? 0} matching record${(response.result_summary.total ?? response.result_summary.items?.length ?? 0) === 1 ? '' : 's'}`}
          </p>
          {response.result_summary.metrics?.length ? (
            <dl className='mt-2 grid gap-2 sm:grid-cols-2'>
              {response.result_summary.metrics.map((metric, index) => (
                <div
                  key={metric.id || `${metric.label || 'metric'}-${index}`}
                  className='rounded border border-slate-700 bg-slate-950/60 p-2'
                >
                  <dt className='text-xs text-slate-400'>{metric.label || metric.id}</dt>
                  <dd className='mt-1 font-medium text-slate-100'>{metric.value ?? 0}</dd>
                </div>
              ))}
            </dl>
          ) : null}
          {response.result_summary.items?.length ? (
            <ul className='mt-2 space-y-2'>
              {response.result_summary.items.map((item, index) => (
                <li
                  key={item.id || `${item.part_number || 'result'}-${index}`}
                  className='rounded border border-slate-700 bg-slate-950/60 p-2'
                >
                  <p className='font-medium text-slate-100'>
                    {[item.part_number, item.title].filter(Boolean).join(' - ')}
                  </p>
                  {[item.brand, item.category, item.status].some(Boolean) ? (
                    <p className='mt-1 text-xs text-slate-400'>
                      {[item.brand, item.category, item.status]
                        .filter(Boolean)
                        .join(' / ')}
                    </p>
                  ) : null}
                </li>
              ))}
            </ul>
          ) : null}
        </div>
      ) : null}
      {response.skill?.name || response.skill?.id ? (
        <p className='mt-2 text-xs text-slate-400'>
          Skill: {response.skill.name || response.skill.id}
        </p>
      ) : null}
      {response.source?.surface || response.source?.channel ? (
        <p className='text-xs text-slate-400'>
          Source: {response.source.surface || 'unknown'} /{' '}
          {response.source.channel || 'unknown'}
        </p>
      ) : null}
      {canRetry ? (
        <Button
          type='button'
          size='sm'
          variant='outline'
          className='mt-3'
          data-testid={`${testID.replace('-state', '')}-retry`}
          onClick={() => void onRetry?.(response.original_intent!.trim())}
        >
          Retry
        </Button>
      ) : canApply ? (
        <Button
          type='button'
          size='sm'
          className='mt-3'
          data-testid={`${testID.replace('-state', '')}-apply`}
          onClick={() => void onApply?.()}
        >
          Apply
        </Button>
      ) : action?.label ? (
        <Button
          type='button'
          size='sm'
          variant='outline'
          className='mt-3'
          onClick={() => void onAction?.(response)}
        >
          {action.label}
        </Button>
      ) : null}
    </section>
  )
}

function count(value: number | undefined) {
  return Number.isFinite(value) ? value : 0
}

function capabilityExamples(
  capabilities: AgentCapability[] | undefined,
  states: string[]
) {
  return (capabilities ?? [])
    .filter(
      (capability) =>
        states.includes(capability.capability_state?.trim() ?? '') &&
        Boolean(capability.display_name?.trim())
    )
    .slice(0, 3)
}

function AgentCapabilitiesCard({
  capabilities,
  testID,
}: {
  capabilities: AgentCapabilitiesContext
  testID: string
}) {
  const summary = capabilities.summary ?? {}
  const examples = capabilityExamples(capabilities.explanation?.capabilities, [
    'available',
    'confirm_required',
  ])
  return (
    <section
      className='rounded-lg border border-cyan-500/30 bg-cyan-950/20 p-3 text-sm'
      data-testid={testID}
      aria-label='Cabinet Agent capability summary'
    >
      <div className='flex flex-wrap items-center justify-between gap-2'>
        <div className='flex items-center gap-2 font-medium text-slate-100'>
          <ShieldCheck className='h-4 w-4 text-cyan-300' />
          Governed capabilities
        </div>
        <Badge variant='outline' className='border-cyan-500/40 text-cyan-100'>
          {count(summary.total)} skills
        </Badge>
      </div>
      <div className='mt-3 grid grid-cols-2 gap-2 text-xs sm:grid-cols-4'>
        <div className='rounded-md bg-slate-950/70 p-2'>
          <span className='block text-slate-400'>Available now</span>
          <strong className='text-slate-100'>{count(summary.available)}</strong>
        </div>
        <div className='rounded-md bg-slate-950/70 p-2'>
          <span className='block text-slate-400'>Confirmation required</span>
          <strong className='text-slate-100'>
            {count(summary.confirm_required)}
          </strong>
        </div>
        <div className='rounded-md bg-slate-950/70 p-2'>
          <span className='block text-slate-400'>Setup required</span>
          <strong className='text-slate-100'>
            {count(summary.setup_required)}
          </strong>
        </div>
        <div className='rounded-md bg-slate-950/70 p-2'>
          <span className='block text-slate-400'>Policy blocked</span>
          <strong className='text-slate-100'>
            {count(summary.blocked_by_policy)}
          </strong>
        </div>
      </div>
      {examples.length > 0 ? (
        <ul className='mt-3 space-y-1 text-xs text-slate-300'>
          {examples.map((capability) => (
            <li key={capability.skill_id} className='flex items-center gap-2'>
              <CheckCircle2 className='h-3.5 w-3.5 shrink-0 text-cyan-300' />
              <span>{capability.display_name}</span>
              <span className='text-slate-500'>
                {capability.capability_state === 'confirm_required'
                  ? 'review then confirm'
                  : 'read only'}
              </span>
            </li>
          ))}
        </ul>
      ) : null}
      <div className='mt-3 flex flex-wrap items-center justify-between gap-2 text-xs text-slate-400'>
        <span>
          Authority:{' '}
          {capabilities.explanation?.authority_mode || 'profile policy'}
        </span>
        <Button asChild type='button' size='sm' variant='outline'>
          <a href='/settings/skills'>
            <Settings2 className='h-3.5 w-3.5' />
            Manage Agent Skills
          </a>
        </Button>
      </div>
    </section>
  )
}

const setupDestinations = {
  configure_openai_api_key: {
    href: '/integrations',
    label: 'Open Integrations',
  },
  configure_openai_provider: {
    href: '/integrations',
    label: 'Open Integrations',
  },
  choose_supported_openai_model: {
    href: '/integrations',
    label: 'Open Integrations',
  },
  review_agent_skill_settings: {
    href: '/settings/skills',
    label: 'Open Agent settings',
  },
  review_profile_settings: {
    href: '/settings/profile',
    label: 'Open Profile settings',
  },
} as const

function setupDestination(setupNextAction: string | undefined) {
  const key = setupNextAction?.trim() as keyof typeof setupDestinations
  return setupDestinations[key]
}

function AgentPlannerCard({
  planner,
  testID,
  onPreviewStateChanged,
}: {
  planner: AgentPlannerContext
  testID: string
  onPreviewStateChanged?: () => void | Promise<void>
}) {
  const durablePreview = planner.preview_result
  const safeSetupDestination = setupDestination(planner.setup_next_action)
  const previewID = durablePreview?.preview_id?.trim() ?? ''
  const [resolvedPreview, setResolvedPreview] = useState({
    previewID: '',
    status: '',
    loaded: false,
  })
  const [previewError, setPreviewError] = useState('')
  const [previewBusy, setPreviewBusy] = useState(false)
  const [strongConfirmation, setStrongConfirmation] = useState<{
    confirmation_token: string
    expires_at: string
    action: string
    target: Record<string, unknown>
    impact: string[]
    recovery: string
  } | null>(null)
  const durableLifecycle = useMemo(() => {
    if (!previewID || durablePreview?.kind !== 'agent_skill_preview')
      return null
    const apply = durablePreview.apply_request
    const cancel = durablePreview.cancel_request
    const strongRequired = durablePreview.strong_confirmation_required === true
    const strongRequest = durablePreview.strong_confirmation_request
    const applyKeys = Object.keys(apply ?? {})
      .sort()
      .join(',')
    const cancelKeys = Object.keys(cancel ?? {})
      .sort()
      .join(',')
    const strongKeys = Object.keys(strongRequest ?? {})
      .sort()
      .join(',')
    if (
      durablePreview.apply_endpoint !== '/api/agent/skills/apply' ||
      durablePreview.cancel_endpoint !== '/api/agent/skills/cancel' ||
      applyKeys !== 'confirm,preview_id,profile_id' ||
      cancelKeys !== 'preview_id,profile_id' ||
      !/^asp_[a-f0-9]+$/.test(previewID) ||
      typeof apply?.profile_id !== 'string' ||
      apply.profile_id.trim() === '' ||
      cancel?.profile_id !== apply.profile_id ||
      apply?.preview_id !== previewID ||
      cancel?.preview_id !== previewID ||
      apply?.confirm !== true
    ) {
      return null
    }
    if (
      strongRequired &&
      (durablePreview.strong_confirmation_endpoint !==
        '/api/agent/skills/confirm-destructive' ||
        strongKeys !== 'preview_id,profile_id' ||
        strongRequest?.profile_id !== apply.profile_id ||
        strongRequest?.preview_id !== previewID)
    ) {
      return null
    }
    return {
      apply,
      cancel,
      strongConfirmation: strongRequired
        ? {
            endpoint: '/api/agent/skills/confirm-destructive' as const,
            request: strongRequest!,
          }
        : null,
    }
  }, [durablePreview, previewID])
  const durableLifecycleMode = durableLifecycle
    ? durableLifecycle.strongConfirmation
      ? 'strong'
      : 'standard'
    : ''
  const durableProfileID = String(
    durableLifecycle?.apply.profile_id ?? ''
  ).trim()
  const previewStatus =
    resolvedPreview.previewID === previewID && resolvedPreview.loaded
      ? resolvedPreview.status
      : ''
  const claimsDurablePreview =
    durablePreview?.kind === 'agent_skill_preview' &&
    durablePreview.confirmation_required === true
  const invalidPreviewContract = claimsDurablePreview && !durableLifecycle
  const previewPending =
    durableLifecycle && resolvedPreview.loaded && previewStatus === 'previewed'
  const lifecycleLabel =
    previewStatus === 'applied'
      ? 'Applied once'
      : previewStatus === 'cancelled'
        ? 'Cancelled safely'
        : previewStatus === 'expired'
          ? 'Preview expired'
          : ''
  const loadDurablePreview = useCallback(async () => {
    if (!durableLifecycleMode || !durableProfileID || !previewID) return
    setPreviewError('')
    try {
      const response = await cabinetProtectedFetch(
        `/api/agent/skills/preview?profile_id=${encodeURIComponent(durableProfileID)}&preview_id=${encodeURIComponent(previewID)}`,
        durableProfileID
      )
      if (!response.ok) {
        throw new Error(`agent_preview_${response.status}`)
      }
      const result = (await response.json()) as { preview_status?: string }
      const status = result.preview_status?.trim() ?? ''
      if (!['previewed', 'applied', 'cancelled', 'expired'].includes(status)) {
        throw new Error('agent_preview_invalid_status')
      }
      setResolvedPreview({ previewID, status, loaded: true })
    } catch {
      setResolvedPreview({ previewID, status: '', loaded: true })
      setPreviewError(
        'Cabinet could not verify this preview. Retry the request.'
      )
    }
  }, [durableLifecycleMode, durableProfileID, previewID])
  useEffect(() => {
    if (!durableLifecycleMode || !durableProfileID || !previewID) return
    setResolvedPreview({ previewID, status: '', loaded: false })
    setStrongConfirmation(null)
    void loadDurablePreview()
    const handlePreviewChanged = (event: Event) => {
      const changedPreviewID = (event as CustomEvent<{ previewID?: string }>)
        .detail?.previewID
      if (changedPreviewID === previewID) void loadDurablePreview()
    }
    window.addEventListener(
      'cabinet:agent-skill-preview-changed',
      handlePreviewChanged
    )
    return () =>
      window.removeEventListener(
        'cabinet:agent-skill-preview-changed',
        handlePreviewChanged
      )
  }, [durableLifecycleMode, durableProfileID, loadDurablePreview, previewID])
  const runPreviewAction = async (
    endpoint: '/api/agent/skills/apply' | '/api/agent/skills/cancel',
    body: Record<string, unknown>
  ) => {
    setPreviewBusy(true)
    setPreviewError('')
    try {
      const response = await cabinetProtectedFetch(endpoint, durableProfileID, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      })
      const result = (await response.json()) as {
        error?: string
        preview_status?: string
      }
      if (!response.ok)
        throw new Error(result.error || `agent_preview_${response.status}`)
      setResolvedPreview({
        previewID,
        status: result.preview_status?.trim() || '',
        loaded: true,
      })
      window.dispatchEvent(
        new CustomEvent('cabinet:agent-skill-preview-changed', {
          detail: { previewID },
        })
      )
      await onPreviewStateChanged?.()
    } catch (error) {
      setPreviewError(
        error instanceof Error ? error.message : 'agent_preview_request_failed'
      )
    } finally {
      setPreviewBusy(false)
    }
  }
  const reviewDestructiveAction = async () => {
    const contract = durableLifecycle?.strongConfirmation
    if (!contract) return
    setPreviewBusy(true)
    setPreviewError('')
    try {
      const response = await cabinetProtectedFetch(
        contract.endpoint,
        durableProfileID,
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(contract.request),
        }
      )
      const result = (await response.json()) as {
        error?: string
        confirmation_token?: string
        expires_at?: string
        action?: string
        target?: Record<string, unknown>
        impact?: string[]
        recovery?: string
      }
      if (
        !response.ok ||
        !result.confirmation_token ||
        !result.expires_at ||
        !result.action ||
        !Array.isArray(result.impact) ||
        result.impact.length === 0
      ) {
        throw new Error(
          result.error || `agent_strong_confirmation_${response.status}`
        )
      }
      setStrongConfirmation({
        confirmation_token: result.confirmation_token,
        expires_at: result.expires_at,
        action: result.action,
        target: result.target ?? {},
        impact: result.impact,
        recovery: result.recovery ?? '',
      })
    } catch (error) {
      setPreviewError(
        error instanceof Error
          ? error.message
          : 'agent_strong_confirmation_failed'
      )
    } finally {
      setPreviewBusy(false)
    }
  }
  const hasError =
    Boolean(planner.error?.code) ||
    invalidPreviewContract ||
    Boolean(previewError)
  const state = hasError
    ? 'Needs attention'
    : durableLifecycle && !resolvedPreview.loaded
      ? 'Verifying preview'
      : lifecycleLabel
        ? lifecycleLabel
        : planner.confirmation_state === 'preview_required'
          ? 'Review required'
          : planner.execution_result
            ? 'Completed safely'
            : 'Planned'
  return (
    <section
      className='rounded-lg border border-slate-700 bg-slate-900/80 p-3 text-sm'
      data-testid={testID}
      aria-label='Cabinet Agent plan result'
    >
      <div className='flex flex-wrap items-center justify-between gap-2'>
        <div className='flex items-center gap-2 font-medium text-slate-100'>
          {hasError ? (
            <AlertTriangle className='h-4 w-4 text-amber-300' />
          ) : (
            <ShieldCheck className='h-4 w-4 text-cyan-300' />
          )}
          Governed Agent plan
        </div>
        <Badge variant='outline'>{state}</Badge>
      </div>
      {invalidPreviewContract ? (
        <p className='mt-2 text-xs text-amber-100'>
          Cabinet rejected an invalid preview contract. Retry the request.
        </p>
      ) : null}
      {!invalidPreviewContract &&
      (planner.message || planner.error?.message) ? (
        <p className='mt-2 text-xs text-slate-300'>
          {planner.message || planner.error?.message}
        </p>
      ) : null}
      {!invalidPreviewContract && planner.preview_result?.action ? (
        <p className='mt-2 text-xs text-cyan-100'>
          Preview ready: {planner.preview_result.action}. Cabinet has not
          applied this change.
        </p>
      ) : null}
      {durableLifecycle ? (
        <div className='mt-3 rounded-md border border-cyan-500/30 bg-cyan-950/20 p-2 text-xs'>
          <p className='text-cyan-100'>
            Opaque preview {previewID}. Cabinet keeps the skill, parameters, and
            provenance server-side.
          </p>
          {!resolvedPreview.loaded ? (
            <p className='mt-2 flex items-center gap-1.5 text-slate-300'>
              <LoaderCircle className='h-3.5 w-3.5 animate-spin' />
              Verifying preview with Cabinet...
            </p>
          ) : null}
          {previewPending ? (
            <div className='mt-2 space-y-2'>
              {strongConfirmation ? (
                <div
                  className='rounded-md border border-red-400/50 bg-red-950/30 p-2 text-red-50'
                  data-testid={`${testID.replace(/-agent-planner-card$/, '')}-agent-strong-confirmation-impact`}
                >
                  <p className='font-medium'>
                    Destructive action: {strongConfirmation.action}
                  </p>
                  <ul className='mt-1 list-disc space-y-1 pl-4'>
                    {strongConfirmation.impact.map((impact) => (
                      <li key={impact}>{impact}</li>
                    ))}
                  </ul>
                  {Object.keys(strongConfirmation.target).length ? (
                    <p className='mt-2 text-xs break-all'>
                      Target: {JSON.stringify(strongConfirmation.target)}
                    </p>
                  ) : null}
                  <p className='mt-2 text-xs'>
                    Expires: {strongConfirmation.expires_at}
                  </p>
                  {strongConfirmation.recovery ? (
                    <p className='mt-1 text-xs'>
                      Recovery: {strongConfirmation.recovery}
                    </p>
                  ) : null}
                </div>
              ) : null}
              <div className='flex flex-wrap gap-2'>
                <Button
                  type='button'
                  size='sm'
                  variant={
                    durableLifecycle.strongConfirmation
                      ? 'destructive'
                      : 'default'
                  }
                  disabled={previewBusy}
                  data-testid={`${testID.replace(/-agent-planner-card$/, '')}-agent-preview-apply`}
                  onClick={() => {
                    if (
                      durableLifecycle.strongConfirmation &&
                      !strongConfirmation
                    ) {
                      void reviewDestructiveAction()
                      return
                    }
                    void runPreviewAction('/api/agent/skills/apply', {
                      ...durableLifecycle.apply,
                      ...(strongConfirmation
                        ? {
                            strong_confirmation_token:
                              strongConfirmation.confirmation_token,
                          }
                        : {}),
                    })
                  }}
                >
                  {previewBusy ? (
                    <LoaderCircle className='h-3.5 w-3.5 animate-spin' />
                  ) : null}
                  {durableLifecycle.strongConfirmation
                    ? strongConfirmation
                      ? 'Confirm destructive action'
                      : 'Review destructive action'
                    : 'Apply change'}
                </Button>
                <Button
                  type='button'
                  size='sm'
                  variant='outline'
                  disabled={previewBusy}
                  data-testid={`${testID.replace(/-agent-planner-card$/, '')}-agent-preview-cancel`}
                  onClick={() =>
                    void runPreviewAction(
                      '/api/agent/skills/cancel',
                      durableLifecycle.cancel
                    )
                  }
                >
                  Cancel preview
                </Button>
              </div>
            </div>
          ) : lifecycleLabel ? (
            <p className='mt-2 font-medium text-slate-100'>{lifecycleLabel}</p>
          ) : null}
          {previewError ? (
            <p className='mt-2 text-amber-200'>{previewError}</p>
          ) : null}
        </div>
      ) : null}
      {!invalidPreviewContract && planner.next_action ? (
        <p className='mt-2 text-xs text-amber-100'>
          Next: {planner.next_action}
        </p>
      ) : null}
      {!invalidPreviewContract && safeSetupDestination ? (
        <Button
          asChild
          type='button'
          size='sm'
          variant='outline'
          className='mt-3'
        >
          <a
            href={safeSetupDestination.href}
            data-testid={`${testID.replace(/-agent-planner-card$/, '')}-agent-setup-link`}
          >
            <Settings2 className='h-3.5 w-3.5' />
            {safeSetupDestination.label}
          </a>
        </Button>
      ) : null}
      {!invalidPreviewContract ? (
        <p className='mt-2 text-[11px] text-slate-500'>
          Provider {planner.provider || 'configured provider'} proposes; Cabinet
          validates authority, controls dispatch, and records the outcome.
        </p>
      ) : null}
    </section>
  )
}

export function AgentResponseCards({
  messages,
  testIDPrefix,
  onPreviewStateChanged,
  onRetry,
  onApply,
  onAction,
}: {
  messages: AgentResponseMessage[]
  testIDPrefix: string
  onPreviewStateChanged?: () => void | Promise<void>
  onRetry?: (intent: string) => void | Promise<void>
  onApply?: () => void | Promise<void>
  onAction?: (response: NormalizedAgentResponse) => void | Promise<void>
}) {
  const response = latestAgentResponse(messages)
  const normalized = response?.agent_response
  const normalizedDurablePreview =
    normalized?.state === 'preview_required' && response?.agent_planner
      ? response.agent_planner
      : undefined
  const capabilities = normalized ? undefined : response?.agent_capabilities
  const planner =
    normalizedDurablePreview ??
    (normalized ? undefined : response?.agent_planner)
  if (!capabilities && !planner && !normalized) return null
  return (
    <div className='space-y-3' data-testid={`${testIDPrefix}-agent-responses`}>
      {normalized ? (
        <NormalizedAgentResponseCard
          response={normalized}
          testID={`${testIDPrefix}-agent-response-state`}
          allowApply={!normalizedDurablePreview}
          onRetry={onRetry}
          onApply={onApply}
          onAction={onAction}
        />
      ) : null}
      {planner ? (
        <AgentPlannerCard
          planner={planner}
          testID={`${testIDPrefix}-agent-planner-card`}
          onPreviewStateChanged={onPreviewStateChanged}
        />
      ) : null}
      {capabilities ? (
        <AgentCapabilitiesCard
          capabilities={capabilities}
          testID={`${testIDPrefix}-agent-capabilities-card`}
        />
      ) : null}
    </div>
  )
}
