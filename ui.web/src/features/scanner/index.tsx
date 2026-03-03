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
  const [providerHealth, setProviderHealth] = useState('unknown')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [actionStatus, setActionStatus] = useState<string | null>(null)
  const [actionFeedback, setActionFeedback] = useState<ActionFeedback | null>(null)
  const [newName, setNewName] = useState('')
  const [newKeywords, setNewKeywords] = useState('')

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
      setProviderHealth(healthPayload.status ?? 'unknown')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'scanner_load_failed')
      setQuerySets([])
      setFailures([])
      setProviderHealth('unknown')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void loadScanner()
  }, [loadScanner])

  const createQuerySet = async () => {
    if (!newName.trim()) {
      return
    }
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
    setActionStatus('query_set_created')
    setActionFeedback(null)
    setNewName('')
    setNewKeywords('')
    await loadScanner()
  }

  const runNow = async (querySetID: string) => {
    const response = await fetch('/api/scanner/run', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ query_set_id: querySetID }),
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
    setActionStatus(`run_started_${querySetID}`)
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
          <Button onClick={() => void createQuerySet()} data-testid='scanner-create-query'>
            Create Query Set
          </Button>
        </section>

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
                  </div>
                  <Button
                    size='sm'
                    data-testid={`scanner-run-${querySet.id}`}
                    onClick={() => void runNow(querySet.id)}
                  >
                    Run Now
                  </Button>
                </div>
              ))}
            </div>
          </section>
        ) : null}

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
