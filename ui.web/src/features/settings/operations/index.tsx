import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { ContentSection } from '../components/content-section'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'

type RuntimeResponse = {
  app_version?: string
  build_date?: string
  bind_mode?: string
  runtime_host?: string
  runtime_port?: number
  update_channel?: string
  update_public_key_configured?: boolean
}

type RuntimeRecoveryResponse = {
  recovery_required?: boolean
}

type ExportSnapshotResponse = {
  schema_version?: number
  exported_at?: string
  items?: Array<{
    part_number?: string
    title?: string
  }>
}

type ImportConflictDetail = {
  part_number?: string
  existing_id?: string
}

type ImportDryRunSummaryResponse = {
  total_items?: number
  new_items?: number
  conflicts?: number
  conflict_details?: ImportConflictDetail[]
}

export function SettingsOperations() {
  const { t } = useTranslation('pages')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [runtimeInfo, setRuntimeInfo] = useState<RuntimeResponse | null>(null)
  const [recoveryInfo, setRecoveryInfo] = useState<RuntimeRecoveryResponse | null>(null)
  const [dataStatus, setDataStatus] = useState<string>('No import or export action has run yet.')
  const [dataTone, setDataTone] = useState<'default' | 'destructive'>('default')
  const [exportPending, setExportPending] = useState(false)
  const [importDryRunPending, setImportDryRunPending] = useState(false)
  const [importJsonInput, setImportJsonInput] = useState(
    '{\n  "snapshot": {\n    "schema_version": 1,\n    "items": []\n  }\n}'
  )
  const [lastExportSummary, setLastExportSummary] = useState<ExportSnapshotResponse | null>(null)
  const [importSummary, setImportSummary] = useState<ImportDryRunSummaryResponse | null>(null)

  const loadOperations = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const [runtimeResp, recoveryResp] = await Promise.all([
        fetch('/api/runtime'),
        fetch('/api/runtime/recovery'),
      ])
      if (!runtimeResp.ok || !recoveryResp.ok) {
        throw new Error('runtime_operations_unavailable')
      }
      const runtimePayload = (await runtimeResp.json()) as RuntimeResponse
      const recoveryPayload = (await recoveryResp.json()) as RuntimeRecoveryResponse
      setRuntimeInfo(runtimePayload)
      setRecoveryInfo(recoveryPayload)
    } catch {
      setRuntimeInfo(null)
      setRecoveryInfo(null)
      setError('Runtime information is unavailable right now.')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void loadOperations()
  }, [loadOperations])

  const runExportJson = useCallback(async () => {
    setExportPending(true)
    setDataTone('default')
    try {
      const response = await fetch('/api/data/export/json')
      if (!response.ok) {
        throw new Error('failed_to_export_json')
      }
      const payload = (await response.json()) as ExportSnapshotResponse
      const itemCount = Array.isArray(payload.items) ? payload.items.length : 0
      setLastExportSummary(payload)
      setDataStatus(`Exported ${itemCount} item snapshot${itemCount === 1 ? '' : 's'}.`)
    } catch {
      setLastExportSummary(null)
      setDataTone('destructive')
      setDataStatus('Export failed. Retry when runtime data services are healthy.')
    } finally {
      setExportPending(false)
    }
  }, [])

  const runImportJsonDryRun = useCallback(async () => {
    setImportDryRunPending(true)
    setImportSummary(null)
    setDataTone('default')
    try {
      const parsed = JSON.parse(importJsonInput) as {
        snapshot?: {
          schema_version?: number
          items?: unknown[]
        }
        schema_version?: number
        items?: unknown[]
      }
      const body =
        parsed && typeof parsed === 'object' && parsed.snapshot
          ? parsed
          : { snapshot: parsed }
      const response = await fetch('/api/data/import/json/dry-run', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      })
      if (!response.ok) {
        throw new Error('failed_to_dry_run_import')
      }
      const payload = (await response.json()) as ImportDryRunSummaryResponse
      setImportSummary(payload)
      const conflictCount = Number(payload.conflicts ?? 0)
      setDataStatus(
        `Import dry-run completed with ${conflictCount} conflict${conflictCount === 1 ? '' : 's'}.`
      )
    } catch {
      setImportSummary(null)
      setDataTone('destructive')
      setDataStatus('Import dry-run failed.')
    } finally {
      setImportDryRunPending(false)
    }
  }, [importJsonInput])

  const runtimeAddress = runtimeInfo
    ? `${runtimeInfo.runtime_host?.trim() || '127.0.0.1'}:${runtimeInfo.runtime_port ?? 17880}`
    : 'Unavailable'
  const updateKeyStatus = runtimeInfo?.update_public_key_configured
    ? 'Configured'
    : 'Missing'

  return (
    <ContentSection
      title={t('settings.operations.title')}
      desc={t('settings.operations.description')}
    >
      <div className='space-y-4 text-sm'>
        {error ? (
          <div
            className='rounded-md border border-destructive/50 bg-destructive/10 p-3 text-destructive'
            data-testid='settings-operations-runtime-error'
          >
            <p className='font-medium'>{error}</p>
            <p className='mt-1 text-xs text-muted-foreground'>
              Check runtime health and retry without leaving the Operations screen.
            </p>
          </div>
        ) : null}

        <div
          className='rounded-md border p-3 space-y-2'
          data-testid='settings-operations-runtime-card'
        >
          <div className='flex items-start justify-between gap-3'>
            <div>
              <p className='font-medium'>Runtime</p>
              <p className='text-muted-foreground'>
                Live runtime version, bind mode, and update channel for this lane.
              </p>
            </div>
            <Button
              variant='outline'
              size='sm'
              data-testid='settings-operations-retry'
              disabled={loading}
              onClick={() => {
                void loadOperations()
              }}
            >
              {loading ? 'Refreshing…' : 'Refresh'}
            </Button>
          </div>
          <p>Version: {loading ? 'Loading runtime...' : runtimeInfo?.app_version || 'Unavailable'}</p>
          <p>Build date: {loading ? 'Loading runtime...' : runtimeInfo?.build_date || 'Unavailable'}</p>
          <p>Address: {loading ? 'Loading runtime...' : runtimeAddress}</p>
          <p>Bind mode: {loading ? 'Loading runtime...' : runtimeInfo?.bind_mode || 'Unavailable'}</p>
          <p>Update channel: {loading ? 'Loading runtime...' : runtimeInfo?.update_channel || 'Unavailable'}</p>
          <p>Update signing key: {loading ? 'Loading runtime...' : updateKeyStatus}</p>
        </div>

        <div
          className='rounded-md border p-3 space-y-2'
          data-testid='settings-operations-recovery-card'
        >
          <p className='font-medium'>Recovery state</p>
          <p className='text-muted-foreground'>
            Surface whether the last runtime shutdown requires recovery attention.
          </p>
          <p>
            {loading
              ? 'Loading recovery state...'
              : recoveryInfo?.recovery_required
                ? 'Recovery required'
              : 'No recovery required'}
          </p>
        </div>

        <div
          className='rounded-md border p-3 space-y-3'
          data-testid='settings-operations-data-card'
        >
          <div className='flex flex-wrap items-start justify-between gap-3'>
            <div>
              <p className='font-medium'>Data import and export</p>
              <p className='text-muted-foreground'>
                Export a JSON snapshot and dry-run a JSON import before applying any workspace changes.
              </p>
            </div>
            <Button
              variant='outline'
              size='sm'
              data-testid='settings-operations-export-json'
              disabled={loading || Boolean(error) || exportPending || importDryRunPending}
              onClick={() => {
                void runExportJson()
              }}
            >
              {exportPending ? 'Exporting…' : 'Export JSON'}
            </Button>
          </div>

          <p
            data-testid='settings-operations-data-status'
            className={dataTone === 'destructive' ? 'text-sm text-destructive' : 'text-sm text-muted-foreground'}
          >
            {dataStatus}
          </p>

          {lastExportSummary ? (
            <div className='rounded-md border bg-muted/20 p-3 text-sm text-muted-foreground'>
              <p>
                Snapshot schema {lastExportSummary.schema_version ?? 1} with{' '}
                {lastExportSummary.items?.length ?? 0} item
                {(lastExportSummary.items?.length ?? 0) === 1 ? '' : 's'} exported at{' '}
                {lastExportSummary.exported_at || 'unknown time'}.
              </p>
            </div>
          ) : null}

          <div className='space-y-2'>
            <label
              htmlFor='settings-operations-import-json-input'
              className='text-sm font-medium'
            >
              JSON import dry-run
            </label>
            <Textarea
              id='settings-operations-import-json-input'
              data-testid='settings-operations-import-json-input'
              value={importJsonInput}
              onChange={(event) => {
                setImportJsonInput(event.target.value)
              }}
              className='min-h-36 font-mono text-xs'
              spellCheck={false}
            />
            <div className='flex justify-end'>
              <Button
                variant='outline'
                size='sm'
                data-testid='settings-operations-import-json-dry-run'
                disabled={loading || Boolean(error) || exportPending || importDryRunPending}
                onClick={() => {
                  void runImportJsonDryRun()
                }}
              >
                {importDryRunPending ? 'Running Dry-Run…' : 'Run Dry-Run'}
              </Button>
            </div>
          </div>

          <div
            className='rounded-md border bg-muted/20 p-3 text-sm text-muted-foreground'
            data-testid='settings-operations-import-summary'
          >
            {importSummary ? (
              <div className='space-y-2'>
                <p>
                  {importSummary.total_items ?? 0} items, {importSummary.new_items ?? 0} new,{' '}
                  {importSummary.conflicts ?? 0} conflict
                  {(importSummary.conflicts ?? 0) === 1 ? '' : 's'}.
                </p>
                {importSummary.conflict_details?.length ? (
                  <div className='space-y-1'>
                    {importSummary.conflict_details.map((detail) => (
                      <p key={`${detail.part_number || 'unknown'}-${detail.existing_id || 'missing'}`}>
                        {detail.part_number || 'Unknown part'} already exists as{' '}
                        {detail.existing_id || 'unknown item'}.
                      </p>
                    ))}
                  </div>
                ) : (
                  <p>No conflicts detected.</p>
                )}
              </div>
            ) : (
              <p>No dry-run summary yet. Paste a snapshot and run a dry-run to review conflicts.</p>
            )}
          </div>
        </div>

        <div className='rounded-md border p-3'>
          <p className='font-medium'>Queue Controls</p>
          <p className='text-muted-foreground'>
            Pause and resume Market Watch and enrichment workers for active profile context.
          </p>
          <div className='mt-3 flex gap-2'>
            <Button variant='outline' size='sm' disabled>
              Pause Workers (Coming soon)
            </Button>
            <Button variant='outline' size='sm' disabled>
              Resume Workers (Coming soon)
            </Button>
          </div>
        </div>
      </div>
    </ContentSection>
  )
}
