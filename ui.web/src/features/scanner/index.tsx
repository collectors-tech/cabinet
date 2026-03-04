import { useCallback, useEffect, useMemo, useState } from 'react'
import { ConfigDrawer } from '@/components/config-drawer'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { LanguageSwitch } from '@/components/language-switch'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'

type QuerySet = {
  id: string
  name: string
  keywords: string[]
  provider_scope?: string[]
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
}

type RunSummary = {
  page_count: number
  observed_page_size: number
  candidates_total: number
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

function parseErrorCode(payload: unknown, fallback: string): string {
  if (payload && typeof payload === 'object' && 'error' in payload) {
    const value = (payload as { error?: unknown }).error
    if (typeof value === 'string' && value.trim().length > 0) {
      return value.trim()
    }
  }
  return fallback
}

function mapScannerActionError(operation: 'run' | 'retry', status: number, errorCode: string): ActionFeedback {
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
      actions: ['Sign in again and confirm profile access permissions.'],
      diagnosticCode: errorCode,
    }
  }
  if (status >= 500) {
    return {
      summary: 'Market Watch service is temporarily unavailable.',
      actions: ['Retry shortly.', 'Check diagnostics for provider/runtime health.'],
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
  const [candidatesByQuerySet, setCandidatesByQuerySet] = useState<Record<string, Candidate[]>>({})
  const [runSummaryByQuerySet, setRunSummaryByQuerySet] = useState<Record<string, RunSummary>>({})
  const [providerHealth, setProviderHealth] = useState('unknown')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [actionStatus, setActionStatus] = useState<string | null>(null)
  const [actionFeedback, setActionFeedback] = useState<ActionFeedback | null>(null)
  const [newName, setNewName] = useState('')
  const [newKeywords, setNewKeywords] = useState('')
  const [providerMode, setProviderMode] = useState<ProviderMode>('single')
  const [singleProvider, setSingleProvider] = useState('ebay')
  const [multiProviders, setMultiProviders] = useState<string[]>(['ebay', 'amazon'])
  const [providerValidation, setProviderValidation] = useState<string | null>(null)

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

      const querySetPayload = (await querySetsRes.json()) as { query_sets?: QuerySet[] }
      const failuresPayload = (await failuresRes.json()) as { failures?: Failure[] }
      const healthPayload = (await healthRes.json()) as { status?: string }

      setQuerySets(querySetPayload.query_sets ?? [])
      setFailures(failuresPayload.failures ?? [])
      setCandidatesByQuerySet({})
      setRunSummaryByQuerySet({})
      setProviderHealth(healthPayload.status ?? 'unknown')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'scanner_load_failed')
      setQuerySets([])
      setFailures([])
      setCandidatesByQuerySet({})
      setRunSummaryByQuerySet({})
      setProviderHealth('unknown')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void loadScanner()
  }, [loadScanner])

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
    if (!newName.trim()) {
      return
    }
    const providerScope = resolveProviderScope()
    if (providerScope.length === 0) {
      setProviderValidation('Select at least one provider')
      return
    }
    setProviderValidation(null)
    const keywords = newKeywords
      .split(',')
      .map((value) => value.trim())
      .filter(Boolean)
    const response = await fetch('/api/scanner/query-sets', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        name: newName.trim(),
        keywords,
        exclusions: [],
        provider_scope: providerScope,
        enabled: true,
      }),
    })
    if (!response.ok) {
      setActionStatus('create_query_set_failed')
      setActionFeedback(
        mapScannerActionError('run', response.status, `create_query_set_${response.status}`)
      )
      return
    }
    const createdQuerySet = (await response.json()) as QuerySet
    setActionStatus('query_set_created')
    setActionFeedback(null)
    setQuerySets((current) => [...current, createdQuerySet])
    setNewName('')
    setNewKeywords('')
  }

  const runNow = async (querySet: QuerySet) => {
    const providerScope = Array.isArray(querySet.provider_scope)
      ? querySet.provider_scope.map((value) => value.trim().toLowerCase()).filter(Boolean)
      : []
    const isBonzaOnly =
      providerScope.length === 1 &&
      (providerScope[0] === 'bonzaslotcars' || providerScope[0] === 'bonza')
    if (isBonzaOnly) {
      const bonzaResponse = await fetch('/api/providers/bonza/run', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          query_set_id: querySet.id,
        }),
      })
      if (!bonzaResponse.ok) {
        const fallbackCode = `bonza_run_failed_${bonzaResponse.status}`
        let code = fallbackCode
        try {
          code = parseErrorCode(await bonzaResponse.json(), fallbackCode)
        } catch {
          code = fallbackCode
        }
        setActionStatus('run_failed')
        setActionFeedback(mapScannerActionError('run', bonzaResponse.status, code))
        return
      }
      const payload = (await bonzaResponse.json()) as {
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
      setActionStatus(`bonza_run_started_${querySet.id}`)
      setActionFeedback(null)
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
      return
    }
    const candidatesResponse = await fetch(
      `/api/scanner/candidates?query_set_id=${encodeURIComponent(querySet.id)}`
    )
    if (candidatesResponse.ok) {
      const payload = (await candidatesResponse.json()) as { candidates?: Candidate[] }
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
      return
    }
    setActionStatus(`retry_requested_${querySetID}`)
    setActionFeedback(null)
    await loadScanner()
  }

  const hasNoQuerySets = useMemo(() => !loading && !error && querySets.length === 0, [
    error,
    loading,
    querySets.length,
  ])

  return (
    <>
      <Header fixed>
        <Search />
        <div className='ms-auto flex items-center space-x-4'>
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
            Manage provider query sets, run market watch searches, and recover from provider failures.
          </p>
        </div>

        <div className='rounded-md border p-3 text-sm' data-testid='scanner-provider-health'>
          Provider health (eBay): <strong>{providerHealth}</strong>
        </div>

        <section className='grid gap-2 md:grid-cols-3'>
          <Input
            value={newName}
            onChange={(event) => setNewName(event.target.value)}
            placeholder='Query set name'
            data-testid='scanner-new-query-name'
          />
          <Input
            value={newKeywords}
            onChange={(event) => setNewKeywords(event.target.value)}
            placeholder='Keywords (comma-separated)'
            data-testid='scanner-new-query-keywords'
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
                <label key={provider} className='inline-flex items-center gap-2 text-sm'>
                  <input
                    type='checkbox'
                    checked={multiProviders.includes(provider)}
                    onChange={(event) => {
                      setProviderValidation(null)
                      setMultiProviders((current) => {
                        if (event.target.checked) {
                          return current.includes(provider) ? current : [...current, provider]
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
          <Button onClick={() => void createQuerySet()} data-testid='scanner-create-query'>
            Create Query Set
          </Button>
        </section>
        {providerValidation ? (
          <p className='text-sm text-destructive' data-testid='market-watch-provider-validation'>
            {providerValidation}
          </p>
        ) : null}

        {loading ? (
          <div className='rounded-md border p-4 text-sm text-muted-foreground'>
            Loading Market Watch workspace...
          </div>
        ) : null}

        {error ? (
          <div
            className='rounded-md border border-destructive/40 bg-destructive/10 p-4 text-sm'
            data-testid='scanner-error-state'
          >
            <p className='font-medium'>Market Watch data is unavailable.</p>
            <p className='mt-1 text-muted-foreground'>{error}</p>
            <Button className='mt-3' variant='outline' size='sm' onClick={() => void loadScanner()}>
              Retry
            </Button>
          </div>
        ) : null}

        {hasNoQuerySets ? (
          <div
            className='rounded-md border border-dashed p-4 text-sm text-muted-foreground'
            data-testid='scanner-empty-state'
          >
            No query sets found. Create your first query set to start Market Watch runs.
          </div>
        ) : null}

        {actionStatus ? (
          <p data-testid='scanner-action-status' className='text-sm'>
            {actionStatus}
          </p>
        ) : null}
        {actionFeedback ? (
          <div className='rounded-md border p-3 text-sm' data-testid='scanner-action-feedback'>
            <p className='font-medium'>{actionFeedback.summary}</p>
            <ul className='mt-2 list-disc ps-4 text-muted-foreground'>
              {actionFeedback.actions.map((action) => (
                <li key={action}>{action}</li>
              ))}
            </ul>
            {actionFeedback.diagnosticCode ? (
              <details className='mt-2 text-xs text-muted-foreground' data-testid='scanner-action-diagnostics'>
                <summary>Diagnostics</summary>
                <p className='mt-1'>{actionFeedback.diagnosticCode}</p>
              </details>
            ) : null}
          </div>
        ) : null}

        {querySets.length > 0 ? (
          <section className='rounded-md border' data-testid='scanner-query-list'>
            <div className='divide-y'>
              {querySets.map((querySet) => (
                <div key={querySet.id} className='flex flex-wrap items-center justify-between gap-2 p-3'>
                  <div>
                    <p className='font-medium'>{querySet.name}</p>
                    <p className='text-xs text-muted-foreground'>
                      {querySet.keywords?.join(', ') || 'no keywords'}
                    </p>
                    <p
                      className='text-xs text-muted-foreground'
                      data-testid={`scanner-query-providers-${querySet.id}`}
                    >
                      {(querySet.provider_scope ?? []).join(', ') || 'ebay'}
                    </p>
                  </div>
                  <Button
                    size='sm'
                    data-testid={`scanner-run-${querySet.id}`}
                    onClick={() => void runNow(querySet)}
                  >
                    Run Now
                  </Button>
                </div>
              ))}
            </div>
          </section>
        ) : null}

        {Object.entries(candidatesByQuerySet).map(([querySetID, candidates]) => (
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
                Pages: {runSummaryByQuerySet[querySetID].page_count} • Candidates:{' '}
                {runSummaryByQuerySet[querySetID].candidates_total} • Observed page size:{' '}
                {runSummaryByQuerySet[querySetID].observed_page_size}
              </p>
            ) : null}
            <p className='mb-2 text-sm font-medium'>Candidates</p>
            {candidates.length === 0 ? (
              <p className='text-xs text-muted-foreground'>No candidates returned.</p>
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
        ))}

        {failures.length > 0 ? (
          <section className='rounded-md border' data-testid='scanner-failures'>
            <div className='divide-y'>
              {failures.map((failure) => (
                <div key={failure.id} className='flex flex-wrap items-center justify-between gap-2 p-3'>
                  <div>
                    <p className='font-medium'>{failure.provider}</p>
                    <p className='text-xs text-muted-foreground'>{failure.message}</p>
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
