import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { Plus, ScanSearch } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { ConfigDrawer } from '@/components/config-drawer'
import { LanguageSwitch } from '@/components/language-switch'
import { Header, HeaderTitle } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'

type QuerySet = {
  id: string
  name: string
  keywords: string[]
  exclusions?: string[]
  provider_scope?: string[]
  max_price?: number
  region?: string
  condition?: string
  schedule_cron?: string
  enabled?: boolean
  last_run_status?: 'never' | 'running' | 'succeeded' | 'failed'
  last_run_at?: string
  last_run_message?: string
  last_candidate_count?: number
}

type Failure = {
  id: string
  query_set_id: string
  provider: string
  message: string
}

type ActionFeedback = {
  summary: string
  actions: string[]
  diagnosticCode?: string
}

type Candidate = {
  id: string
  query_set_id: string
  listing_id: string
  title: string
  source?: string
  price?: number | string
  currency?: string
  url?: string
  source_url?: string
  stock_status?: string
  status?: string
  handoff_state?: string
  handoff_status?: string
}

type RunSummary = {
  page_count: number
  observed_page_size: number
  candidates_total: number
}

type RunMeta = {
  status: 'never' | 'running' | 'succeeded' | 'failed'
  ranAtISO?: string
}

type QuickScanQueueItem = {
  id: string
  fileName: string
  queuedAtISO: string
  status: 'Queued' | 'Linked'
  linkedToInventory: boolean
  confidencePct: number
  suggestions: string[]
  selectedSuggestion: string
  overrideUsed: boolean
}

type RecognitionCandidateInput = {
  id: string
  title: string
  confidence: number
  source: string
  provenance: string
  media_id?: string
  media_url?: string
  target?: 'inventory' | 'wishlist'
  override_id?: string
  override_by?: string
  override_note?: string
}

type RecognitionReview = {
  top_candidate: RecognitionCandidateInput
  alternates: RecognitionCandidateInput[]
  selected_candidate: RecognitionCandidateInput
  confidence_label: string
  requires_manual_review: boolean
  confirm_before_create: boolean
  target: 'inventory' | 'wishlist'
  media_evidence?: Record<string, string>
  provenance?: string[]
  manual_override_applied: boolean
}

type ScannerReviewApplyResult = {
  confirmation_state: 'required' | 'confirmed'
  target: 'inventory' | 'wishlist'
  review: RecognitionReview
  item?: {
    id: string
    title: string
    part_number?: string
    status?: string
  }
  wishlist_entry?: {
    id: string
    item_id: string
    owned: boolean
  }
}

const MARKET_WATCH_PROVIDER_OPTIONS = [
  'ebay',
  'amazon',
  'bonzaslotcars',
  'frontlinehobbies',
  'hobbytechtoys',
  'andrewshobbies',
  'voglers',
  'acercmodels',
  'mrtoys',
]

type ProviderMode = 'single' | 'multi'
type MarketWatchStatusFilter =
  | 'all'
  | 'never'
  | 'running'
  | 'succeeded'
  | 'failed'
type MarketWatchScheduleFilter = 'all' | 'scheduled' | 'manual'

type CreateQueryValidation = {
  name?: string
  keywords?: string
}

function parseErrorCode(payload: unknown, fallback: string): string {
  if (payload && typeof payload === 'object') {
    const values = payload as {
      error?: unknown
      error_code?: unknown
    }
    for (const value of [values.error_code, values.error]) {
      if (typeof value === 'string' && value.trim().length > 0) {
        return value.trim()
      }
    }
  }
  return fallback
}

function mapScannerActionError(
  operation: 'run' | 'retry',
  status: number,
  errorCode: string
): ActionFeedback {
  if (operation === 'run' && status === 400) {
    return {
      summary: 'Run failed due to query validation.',
      actions: [
        'Check query keywords and exclusions.',
        'Review provider health and credentials before retrying.',
      ],
      diagnosticCode: errorCode,
    }
  }
  if (operation === 'retry' && status === 400) {
    return {
      summary: 'Retry request was rejected due to invalid scanner state.',
      actions: [
        'Confirm the failed query set still exists.',
        'Re-run the query set directly after fixing provider inputs.',
      ],
      diagnosticCode: errorCode,
    }
  }
  if (status === 401 || status === 403) {
    return {
      summary: 'Market Watch action was denied.',
      actions: [
        'Review provider health and credentials before retrying.',
        'Sign in again if the provider is ready and the request is still denied.',
      ],
      diagnosticCode: errorCode,
    }
  }
  if (status >= 500) {
    return {
      summary: 'Market Watch service is temporarily unavailable.',
      actions: [
        'Retry shortly.',
        'Check diagnostics for provider/runtime health.',
      ],
      diagnosticCode: errorCode,
    }
  }
  return {
    summary: operation === 'run' ? 'Run failed.' : 'Retry failed.',
    actions: [
      'Verify provider health and credentials.',
      'Validate query set configuration, then retry.',
    ],
    diagnosticCode: errorCode,
  }
}

export function Scanner() {
  const [querySets, setQuerySets] = useState<QuerySet[]>([])
  const [failures, setFailures] = useState<Failure[]>([])
  const [candidatesByQuerySet, setCandidatesByQuerySet] = useState<
    Record<string, Candidate[]>
  >({})
  const [runSummaryByQuerySet, setRunSummaryByQuerySet] = useState<
    Record<string, RunSummary>
  >({})
  const [runMetaByQuerySet, setRunMetaByQuerySet] = useState<
    Record<string, RunMeta>
  >({})
  const [providerHealth, setProviderHealth] = useState('unknown')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [actionStatus, setActionStatus] = useState<string | null>(null)
  const [actionFeedback, setActionFeedback] = useState<ActionFeedback | null>(
    null
  )
  const [newName, setNewName] = useState('')
  const [newKeywords, setNewKeywords] = useState('')
  const [newScheduleCron, setNewScheduleCron] = useState('0 */6 * * *')
  const [createValidation, setCreateValidation] =
    useState<CreateQueryValidation>({})
  const [providerMode, setProviderMode] = useState<ProviderMode>('single')
  const [singleProvider, setSingleProvider] = useState('ebay')
  const [multiProviders, setMultiProviders] = useState<string[]>([
    'ebay',
    'amazon',
  ])
  const [providerValidation, setProviderValidation] = useState<string | null>(
    null
  )
  const [viewMode, setViewMode] = useState<'cards' | 'table'>('cards')
  const [tableProviderFilter, setTableProviderFilter] = useState('all')
  const [tableStatusFilter, setTableStatusFilter] =
    useState<MarketWatchStatusFilter>('all')
  const [tableScheduleFilter, setTableScheduleFilter] =
    useState<MarketWatchScheduleFilter>('all')
  const [tableAttentionOnly, setTableAttentionOnly] = useState(false)
  const [tableResultsOnly, setTableResultsOnly] = useState(false)
  const [selectedOutputQuerySetID, setSelectedOutputQuerySetID] = useState<
    string | null
  >(null)
  const [editingQuerySetID, setEditingQuerySetID] = useState<string | null>(
    null
  )
  const [editingName, setEditingName] = useState('')
  const [editingKeywords, setEditingKeywords] = useState('')
  const [editingScheduleCron, setEditingScheduleCron] = useState('')
  const [handoffStatus, setHandoffStatus] = useState<string | null>(null)
  const [quickScanStatus, setQuickScanStatus] = useState<string | null>(null)
  const [quickScanQueue, setQuickScanQueue] = useState<QuickScanQueueItem[]>([])
  const [pendingApplyScanID, setPendingApplyScanID] = useState<string | null>(
    null
  )
  const [quickApplyTargetByScanID, setQuickApplyTargetByScanID] = useState<
    Record<string, 'inventory' | 'wishlist'>
  >({})
  const [quickReviewByScanID, setQuickReviewByScanID] = useState<
    Record<string, RecognitionReview>
  >({})
  const [quickApplyResultByScanID, setQuickApplyResultByScanID] = useState<
    Record<string, ScannerReviewApplyResult>
  >({})
  const [quickCategoryView, setQuickCategoryView] = useState<'cards' | 'table'>(
    'cards'
  )
  const quickScanFileInputRef = useRef<HTMLInputElement | null>(null)

  const loadScanner = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const [querySetsRes, failuresRes, healthRes] = await Promise.all([
        fetch('/api/scanner/query-sets'),
        fetch('/api/scanner/failures'),
        fetch('/api/provider/health?provider=ebay'),
      ])
      if (!querySetsRes.ok) {
        throw new Error(`query_sets_${querySetsRes.status}`)
      }
      if (!failuresRes.ok) {
        throw new Error(`failures_${failuresRes.status}`)
      }
      if (!healthRes.ok) {
        throw new Error(`provider_health_${healthRes.status}`)
      }

      const querySetPayload = (await querySetsRes.json()) as {
        query_sets?: QuerySet[]
      }
      const failuresPayload = (await failuresRes.json()) as {
        failures?: Failure[]
      }
      const healthPayload = (await healthRes.json()) as { status?: string }

      setQuerySets(querySetPayload.query_sets ?? [])
      setFailures(failuresPayload.failures ?? [])
      setCandidatesByQuerySet({})
      setRunSummaryByQuerySet({})
      setRunMetaByQuerySet(
        Object.fromEntries(
          (querySetPayload.query_sets ?? []).map((querySet) => [
            querySet.id,
            {
              status: querySet.last_run_status ?? 'never',
              ranAtISO: querySet.last_run_at,
            },
          ])
        )
      )
      setProviderHealth(healthPayload.status ?? 'unknown')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'scanner_load_failed')
      setQuerySets([])
      setFailures([])
      setCandidatesByQuerySet({})
      setRunSummaryByQuerySet({})
      setRunMetaByQuerySet({})
      setProviderHealth('unknown')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void loadScanner()
  }, [loadScanner])

  useEffect(() => {
    if (typeof window === 'undefined') {
      return
    }
    const barcode = new URLSearchParams(window.location.search)
      .get('barcode')
      ?.trim()
    if (!barcode) {
      return
    }
    setNewName(`Barcode ${barcode}`)
    setNewKeywords(barcode)
    setActionStatus(`barcode_lookup_ready_${barcode}`)
    setActionFeedback({
      summary: 'Barcode lookup is ready for Market Watch.',
      actions: [
        'Review provider scope before creating the query set.',
        'Create the query set or edit the barcode keyword first.',
      ],
    })
  }, [])

  const resolveProviderScope = () => {
    if (providerMode === 'single') {
      const normalized = singleProvider.trim().toLowerCase()
      return normalized ? [normalized] : []
    }
    return Array.from(
      new Set(
        multiProviders
          .map((provider) => provider.trim().toLowerCase())
          .filter((provider) => provider.length > 0)
      )
    )
  }

  const createQuerySet = async () => {
    const keywords = newKeywords
      .split(',')
      .map((value) => value.trim())
      .filter(Boolean)
    const nextValidation: CreateQueryValidation = {}
    if (!newName.trim()) {
      nextValidation.name = 'Query set name is required.'
    }
    if (keywords.length === 0) {
      nextValidation.keywords =
        'Enter at least one keyword before creating a query set.'
    }
    if (nextValidation.name || nextValidation.keywords) {
      setCreateValidation(nextValidation)
      setActionStatus('create_query_set_validation_failed')
      setActionFeedback({
        summary: 'Create Query Set requires the highlighted fields.',
        actions: [
          nextValidation.name ?? 'Provide a query set name.',
          nextValidation.keywords ?? 'Provide at least one keyword.',
        ],
      })
      return
    }
    setCreateValidation({})
    const providerScope = resolveProviderScope()
    if (providerScope.length === 0) {
      setProviderValidation('Select at least one provider')
      return
    }
    setProviderValidation(null)
    const response = await fetch('/api/scanner/query-sets', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        name: newName.trim(),
        keywords,
        exclusions: [],
        provider_scope: providerScope,
        schedule_cron: newScheduleCron.trim(),
        enabled: true,
      }),
    })
    if (!response.ok) {
      setActionStatus('create_query_set_failed')
      setActionFeedback(
        mapScannerActionError(
          'run',
          response.status,
          `create_query_set_${response.status}`
        )
      )
      return
    }
    const createdQuerySet = (await response.json()) as QuerySet
    setActionStatus('query_set_created')
    setActionFeedback(null)
    setQuerySets((current) => [...current, createdQuerySet])
    setCreateValidation({})
    setRunMetaByQuerySet((current) => ({
      ...current,
      [createdQuerySet.id]: { status: 'never' },
    }))
    setNewName('')
    setNewKeywords('')
    setNewScheduleCron('0 */6 * * *')
  }

  const startEditQuerySet = (querySet: QuerySet) => {
    setEditingQuerySetID(querySet.id)
    setEditingName(querySet.name)
    setEditingKeywords((querySet.keywords ?? []).join(', '))
    setEditingScheduleCron(querySet.schedule_cron ?? '')
  }

  const cancelEditQuerySet = () => {
    setEditingQuerySetID(null)
    setEditingName('')
    setEditingKeywords('')
    setEditingScheduleCron('')
  }

  const saveQuerySet = async (querySet: QuerySet) => {
    if (editingQuerySetID !== querySet.id) {
      return
    }
    const payload: QuerySet = {
      ...querySet,
      name: editingName.trim() || querySet.name,
      keywords: editingKeywords
        .split(',')
        .map((value) => value.trim())
        .filter(Boolean),
      schedule_cron: editingScheduleCron.trim(),
    }
    const response = await fetch(
      `/api/scanner/query-sets/${encodeURIComponent(querySet.id)}`,
      {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      }
    )
    if (!response.ok) {
      setActionStatus('update_query_set_failed')
      setActionFeedback(
        mapScannerActionError(
          'run',
          response.status,
          `update_query_set_${response.status}`
        )
      )
      return
    }
    const updated = (await response.json()) as QuerySet
    setQuerySets((current) =>
      current.map((item) => (item.id === updated.id ? updated : item))
    )
    setActionStatus(`query_set_updated_${updated.id}`)
    setActionFeedback(null)
    cancelEditQuerySet()
  }

  const deleteQuerySet = async (querySetID: string) => {
    const response = await fetch(
      `/api/scanner/query-sets/${encodeURIComponent(querySetID)}`,
      {
        method: 'DELETE',
      }
    )
    if (!response.ok) {
      setActionStatus('delete_query_set_failed')
      setActionFeedback(
        mapScannerActionError(
          'run',
          response.status,
          `delete_query_set_${response.status}`
        )
      )
      return
    }
    setQuerySets((current) => current.filter((item) => item.id !== querySetID))
    setActionStatus(`query_set_deleted_${querySetID}`)
    setActionFeedback(null)
    if (selectedOutputQuerySetID === querySetID) {
      setSelectedOutputQuerySetID(null)
    }
  }

  const runScheduledRefresh = async () => {
    if (querySets.length === 0) {
      setActionStatus('scheduled_run_blocked_empty')
      setActionFeedback({
        summary: 'Run Scheduled Refresh needs at least one runnable query set.',
        actions: [
          'Create a query set first.',
          'Add keywords so the first scheduled run has valid criteria.',
        ],
      })
      return
    }
    const response = await fetch('/api/scanner/run/scheduled', {
      method: 'POST',
    })
    if (!response.ok) {
      const fallbackCode = `scheduled_run_failed_${response.status}`
      let code = fallbackCode
      try {
        code = parseErrorCode(await response.json(), fallbackCode)
      } catch {
        code = fallbackCode
      }
      setActionStatus('scheduled_run_failed')
      setActionFeedback(mapScannerActionError('run', response.status, code))
      return
    }
    const payload = (await response.json()) as {
      run_id?: string
      query_sets_executed?: number
      candidates_collected?: number
      failures?: number
    }
    setActionStatus(`scheduled_run_completed_${payload.run_id ?? 'unknown'}`)
    setActionFeedback({
      summary: 'Scheduled refresh completed.',
      actions: [
        `Query sets executed: ${payload.query_sets_executed ?? 0}`,
        `Candidates collected: ${payload.candidates_collected ?? 0}`,
        `Failures: ${payload.failures ?? 0}`,
      ],
    })
    await loadScanner()
  }

  const handoffToDiscoveries = async (querySet: QuerySet) => {
    const query = encodeURIComponent((querySet.keywords ?? [])[0] ?? '')
    const response = await fetch(`/api/discovery/not-in-collection?q=${query}`)
    if (!response.ok) {
      setHandoffStatus(`discoveries_handoff_failed_${response.status}`)
      return
    }
    const payload = (await response.json()) as {
      items?: Array<{ candidate_id: string }>
    }
    setHandoffStatus(`discoveries_handoff_ok_${payload.items?.length ?? 0}`)
  }

  const handoffFirstCandidateToWishlist = async (querySetID: string) => {
    const candidates = candidatesByQuerySet[querySetID] ?? []
    const firstCandidate = candidates.find(
      (candidate) => (candidate.id ?? '').trim() !== ''
    )
    if (!firstCandidate || !firstCandidate.id) {
      setHandoffStatus('wishlist_handoff_no_candidate')
      return
    }
    const response = await fetch('/api/discovery/action', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        candidate_id: firstCandidate.id,
        type: 'add_to_wishlist',
        payload: {
          source: 'market_watch',
          query_set_id: querySetID,
        },
      }),
    })
    if (!response.ok) {
      setHandoffStatus(`wishlist_handoff_failed_${response.status}`)
      return
    }
    setHandoffStatus(`wishlist_handoff_ok_${firstCandidate.id}`)
  }

  const handoffFirstCandidateToInventory = async (querySetID: string) => {
    const candidates = candidatesByQuerySet[querySetID] ?? []
    const firstCandidate = candidates.find(
      (candidate) => (candidate.id ?? '').trim() !== ''
    )
    if (!firstCandidate || !firstCandidate.id) {
      setHandoffStatus('inventory_handoff_no_candidate')
      return
    }
    const response = await fetch('/api/discovery/action', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        candidate_id: firstCandidate.id,
        type: 'create_item',
        payload: {
          source: 'market_watch',
          query_set_id: querySetID,
        },
      }),
    })
    if (!response.ok) {
      setHandoffStatus(`inventory_handoff_failed_${response.status}`)
      return
    }
    setHandoffStatus(`inventory_handoff_ok_${firstCandidate.id}`)
  }

  const runNow = async (querySet: QuerySet) => {
    setRunMetaByQuerySet((current) => ({
      ...current,
      [querySet.id]: { status: 'running' },
    }))
    const providerScope = Array.isArray(querySet.provider_scope)
      ? querySet.provider_scope
          .map((value) => value.trim().toLowerCase())
          .filter(Boolean)
      : []
    const providerRunRoute =
      providerScope.length === 1
        ? providerScope[0] === 'ebay'
          ? { provider: 'ebay', url: '/api/providers/ebay/run' }
          : providerScope[0] === 'bonzaslotcars' || providerScope[0] === 'bonza'
            ? { provider: 'bonza', url: '/api/providers/bonza/run' }
            : null
        : null
    if (providerRunRoute) {
      const providerResponse = await fetch(providerRunRoute.url, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          query_set_id: querySet.id,
        }),
      })
      if (!providerResponse.ok) {
        const fallbackCode = `${providerRunRoute.provider}_run_failed_${providerResponse.status}`
        let code = fallbackCode
        try {
          code = parseErrorCode(await providerResponse.json(), fallbackCode)
        } catch {
          code = fallbackCode
        }
        setActionStatus('run_failed')
        setActionFeedback(
          mapScannerActionError('run', providerResponse.status, code)
        )
        setRunMetaByQuerySet((current) => ({
          ...current,
          [querySet.id]: { status: 'failed' },
        }))
        return
      }
      const payload = (await providerResponse.json()) as {
        candidates?: Candidate[]
        run_summary?: RunSummary
      }
      setCandidatesByQuerySet((current) => ({
        ...current,
        [querySet.id]: payload.candidates ?? [],
      }))
      if (payload.run_summary) {
        setRunSummaryByQuerySet((current) => ({
          ...current,
          [querySet.id]: payload.run_summary as RunSummary,
        }))
      }
      setActionStatus(`${providerRunRoute.provider}_run_started_${querySet.id}`)
      setActionFeedback(null)
      setRunMetaByQuerySet((current) => ({
        ...current,
        [querySet.id]: {
          status: 'succeeded',
          ranAtISO: new Date().toISOString(),
        },
      }))
      return
    }
    const response = await fetch('/api/scanner/run', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        query_set_id: querySet.id,
        provider_scope: providerScope,
      }),
    })
    if (!response.ok) {
      const fallbackCode = `run_failed_${response.status}`
      let code = fallbackCode
      try {
        code = parseErrorCode(await response.json(), fallbackCode)
      } catch {
        code = fallbackCode
      }
      setActionStatus('run_failed')
      setActionFeedback(mapScannerActionError('run', response.status, code))
      setRunMetaByQuerySet((current) => ({
        ...current,
        [querySet.id]: { status: 'failed' },
      }))
      return
    }
    const candidatesResponse = await fetch(
      `/api/scanner/candidates?query_set_id=${encodeURIComponent(querySet.id)}`
    )
    if (candidatesResponse.ok) {
      const payload = (await candidatesResponse.json()) as {
        candidates?: Candidate[]
      }
      setCandidatesByQuerySet((current) => ({
        ...current,
        [querySet.id]: payload.candidates ?? [],
      }))
    }
    setRunSummaryByQuerySet((current) => {
      if (!(querySet.id in current)) {
        return current
      }
      const next = { ...current }
      delete next[querySet.id]
      return next
    })
    setActionStatus(`run_started_${querySet.id}`)
    setActionFeedback(null)
    setRunMetaByQuerySet((current) => ({
      ...current,
      [querySet.id]: {
        status: 'succeeded',
        ranAtISO: new Date().toISOString(),
      },
    }))
  }

  const retryFailure = async (querySetID: string) => {
    const response = await fetch('/api/scanner/failures/retry', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ query_set_id: querySetID }),
    })
    if (!response.ok) {
      const fallbackCode = `retry_failed_${response.status}`
      let code = fallbackCode
      try {
        code = parseErrorCode(await response.json(), fallbackCode)
      } catch {
        code = fallbackCode
      }
      setActionStatus('retry_failed')
      setActionFeedback(mapScannerActionError('retry', response.status, code))
      setRunMetaByQuerySet((current) => ({
        ...current,
        [querySetID]: { status: 'failed' },
      }))
      return
    }
    setActionStatus(`retry_requested_${querySetID}`)
    setActionFeedback({
      summary: 'Retry requested.',
      actions: [
        'Refreshing Market Watch failure state.',
        'Review provider health if the query remains failed.',
      ],
    })
    setRunMetaByQuerySet((current) => ({
      ...current,
      [querySetID]: { status: 'running' },
    }))
    await loadScanner()
  }

  const hasNoQuerySets = useMemo(
    () => !loading && !error && querySets.length === 0,
    [error, loading, querySets.length]
  )

  const formatRunStatus = (querySetID: string) =>
    runMetaByQuerySet[querySetID]?.status ?? 'never'

  const formatTableStatus = (querySet: QuerySet) => {
    const status = formatRunStatus(querySet.id)
    if (status === 'failed' && querySet.last_run_message) {
      return `${status}: ${querySet.last_run_message}`
    }
    return status
  }

  const formatScheduleMode = (querySet: QuerySet) => {
    if (querySet.enabled === false) {
      return 'Manual / paused'
    }
    if (querySet.schedule_cron?.trim()) {
      return `Scheduled: ${querySet.schedule_cron.trim()}`
    }
    return 'Manual'
  }

  const formatResultCount = (querySetID: string) => {
    const summary = runSummaryByQuerySet[querySetID]
    if (summary) {
      return String(summary.candidates_total)
    }
    const count = candidatesByQuerySet[querySetID]?.length ?? 0
    if (count > 0) {
      return String(count)
    }
    const persistedCount =
      querySets.find((querySet) => querySet.id === querySetID)
        ?.last_candidate_count ?? 0
    return String(persistedCount)
  }

  const formatRunTime = (querySetID: string) => {
    const ranAtISO = runMetaByQuerySet[querySetID]?.ranAtISO
    if (!ranAtISO) {
      return 'Never'
    }
    const value = new Date(ranAtISO)
    if (Number.isNaN(value.getTime())) {
      return 'Never'
    }
    return value.toLocaleString()
  }

  const formatOutputSummary = (querySetID: string) => {
    const summary = runSummaryByQuerySet[querySetID]
    if (summary) {
      return `Pages: ${summary.page_count} | Candidates: ${summary.candidates_total}`
    }
    const count = candidatesByQuerySet[querySetID]?.length ?? 0
    if (count > 0) {
      return `Candidates: ${count}`
    }
    const persistedCount =
      querySets.find((querySet) => querySet.id === querySetID)
        ?.last_candidate_count ?? 0
    if (persistedCount > 0) {
      return `Candidates: ${persistedCount}`
    }
    return 'No output'
  }

  const formatCandidatePrice = (candidate: Candidate) => {
    if (candidate.price === undefined || candidate.price === null) {
      return 'Not provided'
    }
    const price =
      typeof candidate.price === 'number'
        ? candidate.price.toFixed(2)
        : candidate.price.trim()
    if (!price) {
      return 'Not provided'
    }
    return candidate.currency?.trim()
      ? `${price} ${candidate.currency.trim()}`
      : price
  }

  const formatCandidateSource = (candidate: Candidate) =>
    candidate.url?.trim() ||
    candidate.source_url?.trim() ||
    candidate.listing_id?.trim() ||
    'Not provided'

  const formatCandidateStock = (candidate: Candidate) =>
    candidate.stock_status?.trim() || candidate.status?.trim() || 'Not provided'

  const formatCandidateHandoff = (candidate: Candidate) =>
    candidate.handoff_state?.trim() ||
    candidate.handoff_status?.trim() ||
    'Not handed off'

  const querySetHasResults = (querySetID: string) =>
    Number(formatResultCount(querySetID)) > 0

  const resetTableFilters = () => {
    setTableProviderFilter('all')
    setTableStatusFilter('all')
    setTableScheduleFilter('all')
    setTableAttentionOnly(false)
    setTableResultsOnly(false)
  }

  const filteredQuerySets = querySets.filter((querySet) => {
    const providerScope =
      querySet.provider_scope && querySet.provider_scope.length > 0
        ? querySet.provider_scope
        : ['ebay']
    const status = formatRunStatus(querySet.id)
    const isScheduled =
      querySet.enabled !== false && Boolean(querySet.schedule_cron?.trim())
    const hasFailure = status === 'failed' || Boolean(querySet.last_run_message)

    if (
      tableProviderFilter !== 'all' &&
      !providerScope.includes(tableProviderFilter)
    ) {
      return false
    }
    if (tableStatusFilter !== 'all' && status !== tableStatusFilter) {
      return false
    }
    if (tableScheduleFilter === 'scheduled' && !isScheduled) {
      return false
    }
    if (tableScheduleFilter === 'manual' && isScheduled) {
      return false
    }
    if (tableAttentionOnly && !hasFailure) {
      return false
    }
    if (tableResultsOnly && !querySetHasResults(querySet.id)) {
      return false
    }
    return true
  })

  const hasActiveTableFilters =
    tableProviderFilter !== 'all' ||
    tableStatusFilter !== 'all' ||
    tableScheduleFilter !== 'all' ||
    tableAttentionOnly ||
    tableResultsOnly
  const hasProviderAttention =
    !loading &&
    !error &&
    providerHealth !== 'unknown' &&
    providerHealth !== 'ok'

  const latestRunHistory = filteredQuerySets.map((querySet) => ({
    id: querySet.id,
    name: querySet.name,
    status: formatRunStatus(querySet.id),
    ranAt: formatRunTime(querySet.id),
    output: formatOutputSummary(querySet.id),
  }))

  const launchQuickScan = () => {
    const isMobileViewport =
      typeof window !== 'undefined' && window.innerWidth <= 768
    const hasCameraAPI =
      typeof navigator !== 'undefined' &&
      typeof navigator.mediaDevices !== 'undefined' &&
      typeof navigator.mediaDevices.getUserMedia === 'function'

    if (isMobileViewport) {
      setQuickScanStatus('Mobile quick capture ready')
    } else if (hasCameraAPI) {
      setQuickScanStatus(
        'Desktop quick capture ready (camera available + upload fallback)'
      )
    } else {
      setQuickScanStatus('Desktop quick capture ready (upload fallback active)')
    }
    quickScanFileInputRef.current?.click()
  }

  const queueQuickScanFile = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    if (!file) {
      return
    }
    const normalized =
      file.name
        .replace(/\.[^.]+$/, '')
        .trim()
        .toLowerCase() || 'scan'
    const suggestions = [
      `${normalized} (primary match)`,
      `${normalized} (alt: foil variant)`,
      `${normalized} (alt: promo variant)`,
    ]
    const confidencePct = Math.max(
      52,
      Math.min(96, 96 - (file.name.length % 32))
    )
    const entry: QuickScanQueueItem = {
      id: `${Date.now()}-${file.name}`,
      fileName: file.name,
      queuedAtISO: new Date().toISOString(),
      status: 'Queued',
      linkedToInventory: false,
      confidencePct,
      suggestions,
      selectedSuggestion: suggestions[0],
      overrideUsed: false,
    }
    setQuickScanQueue((current) => [entry, ...current].slice(0, 12))
    setQuickScanStatus(`Quick Scan queued: ${file.name}`)
    event.target.value = ''
  }

  const selectQuickScanAlternative = (itemID: string) => {
    setQuickScanQueue((current) =>
      current.map((item) =>
        item.id === itemID
          ? {
              ...item,
              selectedSuggestion:
                item.suggestions[
                  (item.suggestions.indexOf(item.selectedSuggestion) + 1) %
                    item.suggestions.length
                ],
              overrideUsed: true,
            }
          : item
      )
    )
    setQuickScanStatus(
      'Manual override selected. Confirm apply to mutate inventory link state.'
    )
  }

  const markQuickScanLinked = (itemID: string) => {
    setQuickScanQueue((current) =>
      current.map((item) =>
        item.id === itemID
          ? { ...item, linkedToInventory: true, status: 'Linked' }
          : item
      )
    )
    setQuickScanStatus('Scan marked linked after explicit review.')
  }

  const quickScanCandidatesForApply = (
    item: QuickScanQueueItem,
    target: 'inventory' | 'wishlist'
  ): RecognitionCandidateInput[] =>
    item.suggestions.map((suggestion, index) => ({
      id: `${item.fileName}-${index + 1}`.replace(/\s+/g, '-'),
      title: suggestion,
      confidence: Math.max(0.1, (item.confidencePct - index * 12) / 100),
      source: index === 0 ? 'quick-scan-upload' : 'quick-scan-alternate',
      provenance: index === 0 ? 'ui-upload-preview' : 'ui-manual-review',
      media_id: `quick-scan:${item.id}`,
      media_url: `cabinet://quick-scan/${encodeURIComponent(item.fileName)}`,
      target,
      ...(suggestion === item.selectedSuggestion && item.overrideUsed
        ? {
            override_id: `${item.fileName}-${index + 1}`.replace(/\s+/g, '-'),
            override_by: 'scanner-reviewer',
            override_note: 'Selected alternate before confirmed apply',
          }
        : {}),
    }))

  const reviewQuickScanApply = async (itemID: string) => {
    const item = quickScanQueue.find((scan) => scan.id === itemID)
    if (!item) {
      return
    }
    const target = quickApplyTargetByScanID[itemID] ?? 'inventory'
    setPendingApplyScanID(itemID)
    setQuickScanStatus('Reviewing scanner apply preview before any write.')
    const response = await fetch('/api/scanner/recognition-review/apply', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        target,
        confirmed: false,
        candidates: quickScanCandidatesForApply(item, target),
      }),
    })
    const payload = (await response.json()) as
      | ScannerReviewApplyResult
      | { review?: RecognitionReview; error?: string }
    if (response.status !== 409 || !('review' in payload) || !payload.review) {
      setQuickScanStatus(
        `Review preview failed: ${
          'error' in payload && payload.error ? payload.error : response.status
        }`
      )
      return
    }
    setQuickReviewByScanID((current) => ({
      ...current,
      [itemID]: payload.review as RecognitionReview,
    }))
    setQuickScanStatus(
      'Confirmation required. Review confidence, provenance, and selected target before applying.'
    )
  }

  const confirmQuickScanApply = async () => {
    if (!pendingApplyScanID) {
      return
    }
    const item = quickScanQueue.find((scan) => scan.id === pendingApplyScanID)
    if (!item) {
      return
    }
    const target = quickApplyTargetByScanID[pendingApplyScanID] ?? 'inventory'
    const response = await fetch('/api/scanner/recognition-review/apply', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        target,
        confirmed: true,
        candidates: quickScanCandidatesForApply(item, target),
      }),
    })
    if (!response.ok) {
      setQuickScanStatus(`Confirmed scanner apply failed: ${response.status}`)
      return
    }
    const result = (await response.json()) as ScannerReviewApplyResult
    const reloadResponse = await fetch(
      `/api/items?status=${encodeURIComponent(
        target === 'wishlist' ? 'wishlist' : 'owned'
      )}`
    )
    if (!reloadResponse.ok) {
      setQuickScanStatus(
        `Confirmed scanner apply saved, but ${target} reload failed: ${reloadResponse.status}`
      )
      return
    }
    setQuickScanQueue((current) =>
      current.map((item) =>
        item.id === pendingApplyScanID
          ? { ...item, linkedToInventory: true, status: 'Linked' }
          : item
      )
    )
    setQuickApplyResultByScanID((current) => ({
      ...current,
      [pendingApplyScanID]: result,
    }))
    setQuickScanStatus(
      `Scanner ${target} write applied after explicit confirmation.`
    )
    setPendingApplyScanID(null)
  }

  const cancelQuickScanApply = () => {
    setPendingApplyScanID(null)
    setQuickScanStatus('Apply mutation cancelled.')
  }

  const recentUnlinkedQuickScans = useMemo(
    () =>
      quickScanQueue
        .filter((item) => !item.linkedToInventory)
        .sort((a, b) => b.queuedAtISO.localeCompare(a.queuedAtISO)),
    [quickScanQueue]
  )

  return (
    <>
      <Header fixed>
        <Search />
        <HeaderTitle
          title='Market Watch'
          description='Provider scans, listing candidates, and discovery queues.'
          icon={ScanSearch}
          testId='market-watch-header-title'
          iconTestId='market-watch-page-icon'
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

      <Main className='space-y-4'>
        <div>
          <h1 className='text-2xl font-bold tracking-tight'>Market Watch</h1>
          <p className='text-muted-foreground'>
            Manage provider query sets, run market watch searches, and recover
            from provider failures.
          </p>
        </div>

        <div
          className='rounded-md border p-3 text-sm'
          data-testid='scanner-provider-health'
        >
          Provider health (eBay): <strong>{providerHealth}</strong>
        </div>

        <section className='grid gap-2 md:grid-cols-3'>
          <div className='space-y-1'>
            <Input
              value={newName}
              onChange={(event) => {
                setNewName(event.target.value)
                setCreateValidation((current) => ({
                  ...current,
                  name: undefined,
                }))
              }}
              placeholder='Query set name'
              aria-invalid={createValidation.name ? 'true' : 'false'}
              data-testid='scanner-new-query-name'
            />
            {createValidation.name ? (
              <p
                className='text-xs text-destructive'
                data-testid='scanner-new-query-name-validation'
              >
                {createValidation.name}
              </p>
            ) : null}
          </div>
          <div className='space-y-1'>
            <Input
              value={newKeywords}
              onChange={(event) => {
                setNewKeywords(event.target.value)
                setCreateValidation((current) => ({
                  ...current,
                  keywords: undefined,
                }))
              }}
              placeholder='Keywords (comma-separated)'
              aria-invalid={createValidation.keywords ? 'true' : 'false'}
              data-testid='scanner-new-query-keywords'
            />
            {createValidation.keywords ? (
              <p
                className='text-xs text-destructive'
                data-testid='scanner-new-query-keywords-validation'
              >
                {createValidation.keywords}
              </p>
            ) : null}
          </div>
          <Input
            value={newScheduleCron}
            onChange={(event) => setNewScheduleCron(event.target.value)}
            placeholder='Schedule cron (e.g. 0 */6 * * *)'
            data-testid='scanner-new-query-schedule'
          />
          <select
            className='h-9 rounded-md border bg-background px-3 text-sm'
            value={providerMode}
            onChange={(event) => {
              setProviderMode(event.target.value as ProviderMode)
              setProviderValidation(null)
            }}
            data-testid='market-watch-provider-mode'
          >
            <option value='single'>single</option>
            <option value='multi'>multi</option>
          </select>
          {providerMode === 'single' ? (
            <select
              className='h-9 rounded-md border bg-background px-3 text-sm md:col-span-2'
              value={singleProvider}
              onChange={(event) => {
                setSingleProvider(event.target.value)
                setProviderValidation(null)
              }}
              data-testid='market-watch-provider-single'
            >
              {MARKET_WATCH_PROVIDER_OPTIONS.map((provider) => (
                <option key={provider} value={provider}>
                  {provider}
                </option>
              ))}
            </select>
          ) : (
            <div className='flex flex-wrap gap-2 md:col-span-2'>
              {MARKET_WATCH_PROVIDER_OPTIONS.map((provider) => (
                <label
                  key={provider}
                  className='inline-flex items-center gap-2 text-sm'
                >
                  <input
                    type='checkbox'
                    checked={multiProviders.includes(provider)}
                    onChange={(event) => {
                      setProviderValidation(null)
                      setMultiProviders((current) => {
                        if (event.target.checked) {
                          return current.includes(provider)
                            ? current
                            : [...current, provider]
                        }
                        return current.filter((value) => value !== provider)
                      })
                    }}
                    data-testid={`market-watch-provider-checkbox-${provider}`}
                  />
                  <span>{provider}</span>
                </label>
              ))}
            </div>
          )}
          <Button
            onClick={() => void createQuerySet()}
            data-testid='scanner-create-query'
          >
            Create Query Set
          </Button>
          <Button
            variant='outline'
            onClick={() => void runScheduledRefresh()}
            data-testid='scanner-run-scheduled-refresh'
          >
            Run Scheduled Refresh
          </Button>
        </section>
        <section className='flex flex-wrap items-center gap-2'>
          <Button
            type='button'
            size='sm'
            aria-label='Create Market Watch query'
            title='Create Market Watch query'
            data-testid='market-watch-toolbar-create-query'
            onClick={() => void createQuerySet()}
          >
            <Plus className='h-4 w-4' aria-hidden='true' />
          </Button>
          <Button
            type='button'
            size='sm'
            data-testid='card-scanner-quick-scan'
            onClick={launchQuickScan}
          >
            Quick Scan
          </Button>
          <input
            ref={quickScanFileInputRef}
            type='file'
            accept='image/*'
            capture='environment'
            className='hidden'
            data-testid='card-scanner-quick-file-input'
            onChange={queueQuickScanFile}
          />
          <Button
            type='button'
            size='sm'
            variant={viewMode === 'cards' ? 'default' : 'outline'}
            data-testid='market-watch-view-mode-cards'
            onClick={() => setViewMode('cards')}
          >
            Cards
          </Button>
          <Button
            type='button'
            size='sm'
            variant={viewMode === 'table' ? 'default' : 'outline'}
            data-testid='market-watch-view-mode-table'
            onClick={() => setViewMode('table')}
          >
            Table
          </Button>
          {quickScanStatus ? (
            <span
              className='text-xs text-muted-foreground'
              data-testid='card-scanner-quick-scan-status'
            >
              {quickScanStatus}
            </span>
          ) : (
            <span className='text-xs text-muted-foreground'>
              Quick Scan supports one-tap mobile capture and desktop upload
              fallback.
            </span>
          )}
        </section>
        {querySets.length > 0 ? (
          <section
            className='rounded-md border p-3 text-sm'
            data-testid='market-watch-table-filters'
          >
            <div className='grid gap-2 md:grid-cols-[1fr_1fr_1fr_auto_auto_auto]'>
              <label className='grid gap-1 text-xs font-medium'>
                Provider
                <select
                  className='h-9 rounded-md border bg-background px-3 text-sm font-normal'
                  value={tableProviderFilter}
                  data-testid='market-watch-filter-provider'
                  onChange={(event) =>
                    setTableProviderFilter(event.target.value)
                  }
                >
                  <option value='all'>All providers</option>
                  {MARKET_WATCH_PROVIDER_OPTIONS.map((provider) => (
                    <option key={provider} value={provider}>
                      {provider}
                    </option>
                  ))}
                </select>
              </label>
              <label className='grid gap-1 text-xs font-medium'>
                Status
                <select
                  className='h-9 rounded-md border bg-background px-3 text-sm font-normal'
                  value={tableStatusFilter}
                  data-testid='market-watch-filter-status'
                  onChange={(event) =>
                    setTableStatusFilter(
                      event.target.value as MarketWatchStatusFilter
                    )
                  }
                >
                  <option value='all'>All statuses</option>
                  <option value='never'>Never run</option>
                  <option value='running'>Running</option>
                  <option value='succeeded'>Succeeded</option>
                  <option value='failed'>Failed</option>
                </select>
              </label>
              <label className='grid gap-1 text-xs font-medium'>
                Schedule
                <select
                  className='h-9 rounded-md border bg-background px-3 text-sm font-normal'
                  value={tableScheduleFilter}
                  data-testid='market-watch-filter-schedule'
                  onChange={(event) =>
                    setTableScheduleFilter(
                      event.target.value as MarketWatchScheduleFilter
                    )
                  }
                >
                  <option value='all'>All schedules</option>
                  <option value='scheduled'>Scheduled</option>
                  <option value='manual'>Manual or paused</option>
                </select>
              </label>
              <label className='flex items-center gap-2 text-xs font-medium md:self-end md:pb-2'>
                <input
                  type='checkbox'
                  checked={tableAttentionOnly}
                  data-testid='market-watch-filter-attention'
                  onChange={(event) =>
                    setTableAttentionOnly(event.target.checked)
                  }
                />
                Needs attention
              </label>
              <label className='flex items-center gap-2 text-xs font-medium md:self-end md:pb-2'>
                <input
                  type='checkbox'
                  checked={tableResultsOnly}
                  data-testid='market-watch-filter-results'
                  onChange={(event) =>
                    setTableResultsOnly(event.target.checked)
                  }
                />
                Has results
              </label>
              <Button
                type='button'
                size='sm'
                variant='outline'
                className='md:self-end'
                data-testid='market-watch-filter-reset'
                onClick={resetTableFilters}
              >
                Reset
              </Button>
            </div>
            <p
              className='mt-2 text-xs text-muted-foreground'
              data-testid='market-watch-filter-summary'
            >
              Showing {filteredQuerySets.length} of {querySets.length} Market
              Watch queries
            </p>
          </section>
        ) : null}
        {latestRunHistory.length > 0 ? (
          <section
            className='rounded-md border p-3 text-sm'
            data-testid='market-watch-run-history'
          >
            <div className='flex flex-wrap items-center justify-between gap-2'>
              <p className='font-medium'>Latest run history</p>
              <p className='text-xs text-muted-foreground'>
                Hydrated from saved query metadata
              </p>
            </div>
            <ul className='mt-2 divide-y text-xs'>
              {latestRunHistory.map((row) => (
                <li
                  key={row.id}
                  className='grid gap-1 py-2 md:grid-cols-[1.2fr_0.8fr_1fr_1fr]'
                  data-testid={`market-watch-run-history-${row.id}`}
                >
                  <span className='font-medium'>{row.name}</span>
                  <span className='capitalize'>{row.status}</span>
                  <span className='text-muted-foreground'>{row.ranAt}</span>
                  <span className='text-muted-foreground'>{row.output}</span>
                </li>
              ))}
            </ul>
          </section>
        ) : null}
        <section
          className='rounded-md border p-2 text-xs'
          data-testid='card-scanner-queue'
        >
          {quickScanQueue.length === 0 ? (
            <p className='text-muted-foreground'>No quick-scan items queued.</p>
          ) : (
            <ul className='space-y-1'>
              {quickScanQueue.map((item) => (
                <li
                  key={item.id}
                  className='flex flex-wrap items-center justify-between gap-2'
                >
                  <span>{item.fileName}</span>
                  <span className='text-muted-foreground'>{item.status}</span>
                  {quickApplyResultByScanID[item.id]?.item ? (
                    <span
                      className='basis-full text-[11px] text-muted-foreground'
                      data-testid={`card-scanner-apply-result-${item.id}`}
                    >
                      Created {quickApplyResultByScanID[item.id].target} item:{' '}
                      {quickApplyResultByScanID[item.id].item?.title}
                    </span>
                  ) : null}
                </li>
              ))}
            </ul>
          )}
        </section>
        <section
          className='rounded-md border p-3'
          data-testid='card-scanner-quick-category'
        >
          <div className='flex flex-wrap items-center justify-between gap-2'>
            <p className='text-sm font-medium'>Recent Unlinked Scans</p>
            <div className='flex items-center gap-2'>
              <Button
                type='button'
                size='sm'
                variant={quickCategoryView === 'cards' ? 'default' : 'outline'}
                data-testid='card-scanner-quick-category-view-cards'
                onClick={() => setQuickCategoryView('cards')}
              >
                Cards
              </Button>
              <Button
                type='button'
                size='sm'
                variant={quickCategoryView === 'table' ? 'default' : 'outline'}
                data-testid='card-scanner-quick-category-view-table'
                onClick={() => setQuickCategoryView('table')}
              >
                Table
              </Button>
            </div>
          </div>
          {recentUnlinkedQuickScans.length === 0 ? (
            <p className='mt-2 text-xs text-muted-foreground'>
              No recent unlinked scans. Add quick-scan items to review here.
            </p>
          ) : null}
          {recentUnlinkedQuickScans.length > 0 &&
          quickCategoryView === 'cards' ? (
            <ul
              className='mt-3 space-y-2'
              data-testid='card-scanner-unlinked-cards-list'
            >
              {recentUnlinkedQuickScans.map((item) => (
                <li
                  key={item.id}
                  className='rounded-md border p-2'
                  data-testid={`card-scanner-unlinked-item-${item.id}`}
                >
                  <div className='flex flex-wrap items-center justify-between gap-2'>
                    <div>
                      <p className='text-xs font-medium'>{item.fileName}</p>
                      <p
                        className='text-[11px] text-muted-foreground'
                        data-testid={`card-scanner-confidence-${item.id}`}
                      >
                        Confidence: {item.confidencePct}%
                      </p>
                      <p
                        className='text-[11px] text-muted-foreground'
                        data-testid={`card-scanner-suggestion-${item.id}`}
                      >
                        Suggestion: {item.selectedSuggestion}
                      </p>
                      {quickReviewByScanID[item.id] ? (
                        <p
                          className='text-[11px] text-muted-foreground'
                          data-testid={`card-scanner-review-summary-${item.id}`}
                        >
                          Review:{' '}
                          {quickReviewByScanID[item.id].confidence_label}{' '}
                          confidence, target{' '}
                          {quickReviewByScanID[item.id].target},{' '}
                          {quickReviewByScanID[item.id].confirm_before_create
                            ? 'confirm-before-create required'
                            : 'confirmation not required'}
                        </p>
                      ) : null}
                      {quickApplyResultByScanID[item.id]?.item ? (
                        <p
                          className='text-[11px] text-muted-foreground'
                          data-testid={`card-scanner-apply-result-${item.id}`}
                        >
                          Created {quickApplyResultByScanID[item.id].target}{' '}
                          item: {quickApplyResultByScanID[item.id].item?.title}
                        </p>
                      ) : null}
                      <p className='text-[11px] text-muted-foreground'>
                        Queued: {new Date(item.queuedAtISO).toLocaleString()}
                      </p>
                    </div>
                    <div className='flex flex-wrap items-center gap-2'>
                      <select
                        className='h-8 rounded-md border bg-background px-2 text-xs'
                        value={quickApplyTargetByScanID[item.id] ?? 'inventory'}
                        data-testid={`card-scanner-apply-target-${item.id}`}
                        onChange={(event) =>
                          setQuickApplyTargetByScanID((current) => ({
                            ...current,
                            [item.id]: event.target.value as
                              | 'inventory'
                              | 'wishlist',
                          }))
                        }
                      >
                        <option value='inventory'>Inventory</option>
                        <option value='wishlist'>Wishlist</option>
                      </select>
                      <Button
                        type='button'
                        size='sm'
                        variant='outline'
                        data-testid={`card-scanner-mark-linked-${item.fileName}`}
                        onClick={() => markQuickScanLinked(item.id)}
                      >
                        Mark Linked
                      </Button>
                      <Button
                        type='button'
                        size='sm'
                        variant='outline'
                        data-testid={`card-scanner-override-${item.id}`}
                        onClick={() => selectQuickScanAlternative(item.id)}
                      >
                        Use Alternative
                      </Button>
                      <Button
                        type='button'
                        size='sm'
                        variant='outline'
                        data-testid={`card-scanner-review-apply-${item.id}`}
                        onClick={() => void reviewQuickScanApply(item.id)}
                      >
                        Review Apply
                      </Button>
                    </div>
                  </div>
                  {item.overrideUsed ? (
                    <p
                      className='mt-2 text-[11px] text-amber-500'
                      data-testid={`card-scanner-override-flag-${item.id}`}
                    >
                      Manual override active
                    </p>
                  ) : null}
                  {pendingApplyScanID === item.id ? (
                    <div
                      className='mt-2 rounded-md border border-primary/40 bg-primary/5 p-2 text-[11px]'
                      data-testid={`card-scanner-apply-confirmation-${item.id}`}
                    >
                      <p className='mb-2'>
                        Confirm apply to link this scan to inventory using
                        selected suggestion.
                      </p>
                      <div className='flex flex-wrap gap-2'>
                        <Button
                          type='button'
                          size='sm'
                          data-testid={`card-scanner-confirm-apply-${item.id}`}
                          onClick={() => void confirmQuickScanApply()}
                        >
                          Confirm Apply
                        </Button>
                        <Button
                          type='button'
                          size='sm'
                          variant='outline'
                          data-testid={`card-scanner-cancel-apply-${item.id}`}
                          onClick={cancelQuickScanApply}
                        >
                          Cancel
                        </Button>
                      </div>
                    </div>
                  ) : null}
                </li>
              ))}
            </ul>
          ) : null}
          {recentUnlinkedQuickScans.length > 0 &&
          quickCategoryView === 'table' ? (
            <div className='mt-3 overflow-x-auto'>
              <table
                className='w-full text-xs'
                data-testid='card-scanner-unlinked-table'
              >
                <thead className='text-left'>
                  <tr>
                    <th className='px-2 py-1'>File</th>
                    <th className='px-2 py-1'>Confidence</th>
                    <th className='px-2 py-1'>Suggestion</th>
                    <th className='px-2 py-1'>Queued At</th>
                    <th className='px-2 py-1'>Status</th>
                  </tr>
                </thead>
                <tbody>
                  {recentUnlinkedQuickScans.map((item) => (
                    <tr key={item.id} className='border-t'>
                      <td className='px-2 py-1'>{item.fileName}</td>
                      <td className='px-2 py-1'>{item.confidencePct}%</td>
                      <td className='px-2 py-1'>{item.selectedSuggestion}</td>
                      <td className='px-2 py-1'>
                        {new Date(item.queuedAtISO).toLocaleString()}
                      </td>
                      <td className='px-2 py-1'>{item.status}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : null}
        </section>
        {providerValidation ? (
          <p
            className='text-sm text-destructive'
            data-testid='market-watch-provider-validation'
          >
            {providerValidation}
          </p>
        ) : null}

        {loading ? (
          <div
            className='rounded-md border p-4 text-sm text-muted-foreground'
            data-testid='scanner-loading-state'
          >
            Loading Market Watch workspace data, provider health, and failure
            history...
          </div>
        ) : null}

        {error ? (
          <div
            className='rounded-md border border-destructive/40 bg-destructive/10 p-4 text-sm'
            data-testid='scanner-error-state'
          >
            <p className='font-medium'>Market Watch data is unavailable.</p>
            <p className='mt-1 text-muted-foreground'>{error}</p>
            <Button
              className='mt-3'
              variant='outline'
              size='sm'
              data-testid='scanner-error-retry'
              onClick={() => void loadScanner()}
            >
              Retry
            </Button>
          </div>
        ) : null}

        {hasProviderAttention ? (
          <div
            className='rounded-md border border-amber-300 bg-amber-50 p-4 text-sm text-amber-950 dark:border-amber-700 dark:bg-amber-950/30 dark:text-amber-100'
            data-testid='market-watch-provider-attention-state'
          >
            <p className='font-medium'>Provider needs attention.</p>
            <p className='mt-1'>
              eBay provider health is <strong>{providerHealth}</strong>.
            </p>
            <ul className='mt-2 list-disc ps-4'>
              <li>Review provider credentials and API access.</li>
              <li>Retry failed Market Watch runs after provider health recovers.</li>
            </ul>
          </div>
        ) : null}

        {hasNoQuerySets ? (
          <div
            className='rounded-md border border-dashed p-4 text-sm text-muted-foreground'
            data-testid='scanner-empty-state'
          >
            No query sets found. Create your first query set to start Market
            Watch runs.
          </div>
        ) : null}

        {actionStatus ? (
          <p data-testid='scanner-action-status' className='text-sm'>
            {actionStatus}
          </p>
        ) : null}
        {handoffStatus ? (
          <p data-testid='scanner-handoff-status' className='text-sm'>
            {handoffStatus}
          </p>
        ) : null}
        {actionFeedback ? (
          <div
            className='rounded-md border p-3 text-sm'
            data-testid='scanner-action-feedback'
          >
            <p className='font-medium'>{actionFeedback.summary}</p>
            <ul className='mt-2 list-disc ps-4 text-muted-foreground'>
              {actionFeedback.actions.map((action) => (
                <li key={action}>{action}</li>
              ))}
            </ul>
            {actionFeedback.diagnosticCode ? (
              <details
                className='mt-2 text-xs text-muted-foreground'
                data-testid='scanner-action-diagnostics'
              >
                <summary>Diagnostics</summary>
                <p className='mt-1'>{actionFeedback.diagnosticCode}</p>
              </details>
            ) : null}
          </div>
        ) : null}

        {querySets.length > 0 && filteredQuerySets.length === 0 ? (
          <section
            className='rounded-md border border-dashed p-4 text-sm text-muted-foreground'
            data-testid='market-watch-filter-empty'
          >
            No Market Watch queries match the current filters.
            {hasActiveTableFilters ? (
              <Button
                type='button'
                size='sm'
                variant='outline'
                className='ms-3'
                data-testid='market-watch-filter-empty-reset'
                onClick={resetTableFilters}
              >
                Reset filters
              </Button>
            ) : null}
          </section>
        ) : null}

        {filteredQuerySets.length > 0 && viewMode === 'cards' ? (
          <section
            className='rounded-md border'
            data-testid='scanner-query-list'
          >
            <div className='divide-y'>
              {filteredQuerySets.map((querySet) => (
                <div
                  key={querySet.id}
                  className='flex flex-wrap items-center justify-between gap-2 p-3'
                >
                  <div>
                    {editingQuerySetID === querySet.id ? (
                      <div className='grid gap-2'>
                        <Input
                          value={editingName}
                          onChange={(event) =>
                            setEditingName(event.target.value)
                          }
                          data-testid={`scanner-edit-name-${querySet.id}`}
                        />
                        <Input
                          value={editingKeywords}
                          onChange={(event) =>
                            setEditingKeywords(event.target.value)
                          }
                          data-testid={`scanner-edit-keywords-${querySet.id}`}
                        />
                        <Input
                          value={editingScheduleCron}
                          onChange={(event) =>
                            setEditingScheduleCron(event.target.value)
                          }
                          data-testid={`scanner-edit-schedule-${querySet.id}`}
                        />
                      </div>
                    ) : (
                      <>
                        <p className='font-medium'>{querySet.name}</p>
                        <p className='text-xs text-muted-foreground'>
                          {querySet.keywords?.join(', ') || 'no keywords'}
                        </p>
                      </>
                    )}
                    <p
                      className='text-xs text-muted-foreground'
                      data-testid={`scanner-query-providers-${querySet.id}`}
                    >
                      {(querySet.provider_scope ?? []).join(', ') || 'ebay'}
                    </p>
                    <p
                      className='text-xs text-muted-foreground'
                      data-testid={`scanner-query-schedule-${querySet.id}`}
                    >
                      Schedule: {querySet.schedule_cron || 'none'}
                    </p>
                  </div>
                  <div className='flex flex-wrap gap-2'>
                    <Button
                      size='sm'
                      data-testid={`scanner-run-${querySet.id}`}
                      onClick={() => void runNow(querySet)}
                    >
                      Run Now
                    </Button>
                    {editingQuerySetID === querySet.id ? (
                      <>
                        <Button
                          size='sm'
                          variant='outline'
                          data-testid={`scanner-save-${querySet.id}`}
                          onClick={() => void saveQuerySet(querySet)}
                        >
                          Save
                        </Button>
                        <Button
                          size='sm'
                          variant='ghost'
                          data-testid={`scanner-cancel-edit-${querySet.id}`}
                          onClick={cancelEditQuerySet}
                        >
                          Cancel
                        </Button>
                      </>
                    ) : (
                      <Button
                        size='sm'
                        variant='outline'
                        data-testid={`scanner-edit-${querySet.id}`}
                        onClick={() => startEditQuerySet(querySet)}
                      >
                        Edit
                      </Button>
                    )}
                    <Button
                      size='sm'
                      variant='outline'
                      data-testid={`scanner-delete-${querySet.id}`}
                      onClick={() => void deleteQuerySet(querySet.id)}
                    >
                      Delete
                    </Button>
                    <Button
                      size='sm'
                      variant='outline'
                      data-testid={`scanner-handoff-discoveries-${querySet.id}`}
                      onClick={() => void handoffToDiscoveries(querySet)}
                    >
                      Send to Discoveries
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          </section>
        ) : null}

        {filteredQuerySets.length > 0 && viewMode === 'table' ? (
          <section
            className='rounded-md border'
            data-testid='market-watch-query-table'
          >
            <div className='overflow-x-auto'>
              <table className='w-full text-sm'>
                <thead className='bg-muted/30 text-left'>
                  <tr>
                    <th className='px-3 py-2 font-medium'>Query Name</th>
                    <th className='px-3 py-2 font-medium'>Terms</th>
                    <th className='px-3 py-2 font-medium'>Provider Scope</th>
                    <th className='px-3 py-2 font-medium'>Schedule</th>
                    <th className='px-3 py-2 font-medium'>Latest Status</th>
                    <th className='px-3 py-2 font-medium'>Last Run Time</th>
                    <th className='px-3 py-2 font-medium'>Result Count</th>
                    <th className='px-3 py-2 font-medium'>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {filteredQuerySets.map((querySet) => (
                    <tr key={querySet.id} className='border-t'>
                      <td className='px-3 py-2'>{querySet.name}</td>
                      <td className='px-3 py-2'>
                        {querySet.keywords?.join(', ') || 'no keywords'}
                      </td>
                      <td className='px-3 py-2'>
                        {(querySet.provider_scope ?? []).join(', ') || 'ebay'}
                      </td>
                      <td className='px-3 py-2'>
                        {formatScheduleMode(querySet)}
                      </td>
                      <td className='px-3 py-2 capitalize'>
                        {formatTableStatus(querySet)}
                      </td>
                      <td className='px-3 py-2'>
                        {formatRunTime(querySet.id)}
                      </td>
                      <td className='px-3 py-2'>
                        {formatResultCount(querySet.id)}
                      </td>
                      <td className='px-3 py-2'>
                        <Button
                          type='button'
                          size='sm'
                          variant='outline'
                          data-testid={`market-watch-open-output-${querySet.id}`}
                          onClick={() =>
                            setSelectedOutputQuerySetID(querySet.id)
                          }
                        >
                          Inspect Output
                        </Button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </section>
        ) : null}

        {selectedOutputQuerySetID ? (
          <section
            className='rounded-md border p-3'
            data-testid='market-watch-output-detail'
          >
            <div className='flex items-start justify-between gap-3'>
              <div>
                <p className='text-sm font-medium'>
                  {querySets.find(
                    (querySet) => querySet.id === selectedOutputQuerySetID
                  )?.name ?? selectedOutputQuerySetID}
                </p>
                <p className='mt-1 text-xs text-muted-foreground'>
                  Provider Scope:{' '}
                  {(
                    querySets.find(
                      (querySet) => querySet.id === selectedOutputQuerySetID
                    )?.provider_scope ?? ['ebay']
                  ).join(', ')}
                </p>
                <p className='text-xs text-muted-foreground'>
                  Last run at: {formatRunTime(selectedOutputQuerySetID)}
                </p>
                <p className='text-xs text-muted-foreground'>
                  Last run status: {formatRunStatus(selectedOutputQuerySetID)}
                </p>
                {querySets.find(
                  (querySet) => querySet.id === selectedOutputQuerySetID
                )?.last_run_message ? (
                  <p className='text-xs text-muted-foreground'>
                    Latest failure:{' '}
                    {
                      querySets.find(
                        (querySet) => querySet.id === selectedOutputQuerySetID
                      )?.last_run_message
                    }
                  </p>
                ) : null}
              </div>
              <Button
                type='button'
                size='sm'
                variant='ghost'
                onClick={() => setSelectedOutputQuerySetID(null)}
              >
                Close
              </Button>
            </div>
            {runSummaryByQuerySet[selectedOutputQuerySetID] ? (
              <div className='mt-3 rounded-md border p-2 text-xs'>
                <p>
                  Pages scanned:{' '}
                  {runSummaryByQuerySet[selectedOutputQuerySetID].page_count}
                </p>
                <p>
                  Candidates:{' '}
                  {
                    runSummaryByQuerySet[selectedOutputQuerySetID]
                      .candidates_total
                  }
                </p>
                <p>
                  Observed page size:{' '}
                  {
                    runSummaryByQuerySet[selectedOutputQuerySetID]
                      .observed_page_size
                  }
                </p>
              </div>
            ) : null}
            <div className='mt-3'>
              <p className='text-xs font-medium'>Latest output items</p>
              {(candidatesByQuerySet[selectedOutputQuerySetID] ?? []).length ===
              0 ? (
                <p
                  className='text-xs text-muted-foreground'
                  data-testid='market-watch-output-no-results'
                >
                  No output available yet.
                  <span className='ms-1'>
                    Run this query or adjust provider scope to collect visible
                    results.
                  </span>
                </p>
              ) : (
                <div className='mt-2 overflow-x-auto'>
                  <table
                    className='w-full text-xs'
                    data-testid='market-watch-output-results-table'
                  >
                    <thead className='bg-muted/30 text-left'>
                      <tr>
                        <th className='px-2 py-1 font-medium'>Provider</th>
                        <th className='px-2 py-1 font-medium'>Result</th>
                        <th className='px-2 py-1 font-medium'>Price</th>
                        <th className='px-2 py-1 font-medium'>Source</th>
                        <th className='px-2 py-1 font-medium'>Stock</th>
                        <th className='px-2 py-1 font-medium'>Handoff</th>
                      </tr>
                    </thead>
                    <tbody>
                      {(
                        candidatesByQuerySet[selectedOutputQuerySetID] ?? []
                      ).map((candidate) => (
                        <tr
                          key={candidate.id || candidate.listing_id}
                          className='border-t'
                        >
                          <td className='px-2 py-1'>
                            {candidate.source ?? 'unknown'}
                          </td>
                          <td className='px-2 py-1'>{candidate.title}</td>
                          <td className='px-2 py-1'>
                            {formatCandidatePrice(candidate)}
                          </td>
                          <td className='px-2 py-1'>
                            {formatCandidateSource(candidate)}
                          </td>
                          <td className='px-2 py-1'>
                            {formatCandidateStock(candidate)}
                          </td>
                          <td className='px-2 py-1'>
                            {formatCandidateHandoff(candidate)}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
            <div className='mt-3 flex flex-wrap gap-2'>
              <Button
                type='button'
                size='sm'
                variant='outline'
                data-testid={`scanner-handoff-wishlist-${selectedOutputQuerySetID}`}
                onClick={() =>
                  void handoffFirstCandidateToWishlist(selectedOutputQuerySetID)
                }
              >
                Add First Result to Wishlist
              </Button>
              <Button
                type='button'
                size='sm'
                variant='outline'
                data-testid={`scanner-handoff-inventory-${selectedOutputQuerySetID}`}
                onClick={() =>
                  void handoffFirstCandidateToInventory(
                    selectedOutputQuerySetID
                  )
                }
              >
                Add First Result to Inventory
              </Button>
              <Button
                type='button'
                size='sm'
                variant='outline'
                data-testid={`scanner-handoff-discoveries-detail-${selectedOutputQuerySetID}`}
                onClick={() => {
                  const current = querySets.find(
                    (querySet) => querySet.id === selectedOutputQuerySetID
                  )
                  if (current) {
                    void handoffToDiscoveries(current)
                  }
                }}
              >
                Open Discoveries Handoff
              </Button>
            </div>
          </section>
        ) : null}

        {Object.entries(candidatesByQuerySet).map(
          ([querySetID, candidates]) => (
            <section
              key={querySetID}
              className='rounded-md border p-3'
              data-testid={`scanner-candidates-${querySetID}`}
            >
              {runSummaryByQuerySet[querySetID] ? (
                <p
                  className='mb-2 text-xs text-muted-foreground'
                  data-testid={`scanner-run-summary-${querySetID}`}
                >
                  Pages: {runSummaryByQuerySet[querySetID].page_count} •
                  Candidates:{' '}
                  {runSummaryByQuerySet[querySetID].candidates_total} • Observed
                  page size:{' '}
                  {runSummaryByQuerySet[querySetID].observed_page_size}
                </p>
              ) : null}
              <p className='mb-2 text-sm font-medium'>Candidates</p>
              {candidates.length === 0 ? (
                <p className='text-xs text-muted-foreground'>
                  No candidates returned.
                </p>
              ) : (
                <ul className='space-y-1 text-xs text-muted-foreground'>
                  {candidates.map((candidate) => (
                    <li key={candidate.id || candidate.listing_id}>
                      {candidate.title} ({candidate.source ?? 'unknown'})
                    </li>
                  ))}
                </ul>
              )}
            </section>
          )
        )}

        {failures.length > 0 ? (
          <section className='rounded-md border' data-testid='scanner-failures'>
            <div className='divide-y'>
              {failures.map((failure) => (
                <div
                  key={failure.id}
                  className='flex flex-wrap items-center justify-between gap-2 p-3'
                >
                  <div>
                    <p className='font-medium'>{failure.provider}</p>
                    <p className='text-xs text-muted-foreground'>
                      {failure.message}
                    </p>
                  </div>
                  <Button
                    size='sm'
                    variant='outline'
                    data-testid={`scanner-retry-${failure.query_set_id}`}
                    onClick={() => void retryFailure(failure.query_set_id)}
                  >
                    Retry
                  </Button>
                </div>
              ))}
            </div>
          </section>
        ) : null}
      </Main>
    </>
  )
}
