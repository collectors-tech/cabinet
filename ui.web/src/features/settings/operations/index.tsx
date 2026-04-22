import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { ContentSection } from '../components/content-section'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Input } from '@/components/ui/input'
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

type ImportSnapshotRequest = {
  snapshot: {
    schema_version?: number
    items?: unknown[]
  }
}

type ImportApplyAction = 'merge' | 'create' | 'skip'

type CSVImportRequest = {
  csv: string
  mapping: Record<string, string>
}

type CSVImportMappingState = {
  brand: string
  category: string
  part_number: string
  title: string
}

function parseImportSnapshotRequest(rawInput: string): ImportSnapshotRequest {
  const parsed = JSON.parse(rawInput) as {
    snapshot?: {
      schema_version?: number
      items?: unknown[]
    }
    schema_version?: number
    items?: unknown[]
  }
  if (parsed && typeof parsed === 'object' && parsed.snapshot) {
    return parsed as ImportSnapshotRequest
  }
  return { snapshot: parsed }
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
  const [importApplyPending, setImportApplyPending] = useState(false)
  const [exportCsvPending, setExportCsvPending] = useState(false)
  const [importCsvDryRunPending, setImportCsvDryRunPending] = useState(false)
  const [importCsvApplyPending, setImportCsvApplyPending] = useState(false)
  const [importJsonInput, setImportJsonInput] = useState(
    '{\n  "snapshot": {\n    "schema_version": 1,\n    "items": []\n  }\n}'
  )
  const [importCsvInput, setImportCsvInput] = useState(
    'brand,category,part_number,title\nAFX,Slot,CSV-001,Example Item'
  )
  const [importCsvMapping, setImportCsvMapping] = useState<CSVImportMappingState>({
    brand: '',
    category: '',
    part_number: '',
    title: '',
  })
  const [lastExportSummary, setLastExportSummary] = useState<ExportSnapshotResponse | null>(null)
  const [importSummary, setImportSummary] = useState<ImportDryRunSummaryResponse | null>(null)
  const [csvStatus, setCsvStatus] = useState<string>('No CSV import or export action has run yet.')
  const [csvTone, setCsvTone] = useState<'default' | 'destructive'>('default')
  const [csvSummary, setCsvSummary] = useState<ImportDryRunSummaryResponse | null>(null)
  const [logsExportPending, setLogsExportPending] = useState(false)
  const [logsStatus, setLogsStatus] = useState<string>('No diagnostics export has run yet.')
  const [logsTone, setLogsTone] = useState<'default' | 'destructive'>('default')
  const [logsPreview, setLogsPreview] = useState<string | null>(null)
  const [importDefaultAction, setImportDefaultAction] =
    useState<ImportApplyAction>('merge')
  const [importCsvDefaultAction, setImportCsvDefaultAction] =
    useState<ImportApplyAction>('merge')
  const [lastDryRunRequest, setLastDryRunRequest] = useState<ImportSnapshotRequest | null>(null)
  const [lastCsvDryRunRequest, setLastCsvDryRunRequest] = useState<CSVImportRequest | null>(null)

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

  const buildCsvImportRequest = useCallback((): CSVImportRequest => {
    const mapping = Object.fromEntries(
      Object.entries(importCsvMapping)
        .map(([field, value]) => [field, value.trim()])
        .filter(([, value]) => value !== '')
    ) as Record<string, string>

    return {
      csv: importCsvInput,
      mapping,
    }
  }, [importCsvInput, importCsvMapping])

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
    setLastDryRunRequest(null)
    setDataTone('default')
    try {
      const body = parseImportSnapshotRequest(importJsonInput)
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
      setLastDryRunRequest(body)
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

  const runImportJsonApply = useCallback(async () => {
    if (!lastDryRunRequest) {
      return
    }
    setImportApplyPending(true)
    setDataTone('default')
    try {
      const response = await fetch('/api/data/import/json/apply', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          snapshot: lastDryRunRequest.snapshot,
          options: {
            default_action: importDefaultAction,
          },
        }),
      })
      if (!response.ok) {
        throw new Error('failed_to_apply_import')
      }
      setDataStatus('Import applied successfully.')
      setImportSummary(null)
      setLastDryRunRequest(null)
    } catch {
      setDataTone('destructive')
      setDataStatus('Import apply failed.')
    } finally {
      setImportApplyPending(false)
    }
  }, [importDefaultAction, lastDryRunRequest])

  const runExportCsv = useCallback(async () => {
    setExportCsvPending(true)
    setCsvTone('default')
    try {
      const response = await fetch('/api/data/export/csv/items')
      if (!response.ok) {
        throw new Error('failed_to_export_csv')
      }
      const csvText = await response.text()
      const rowCount = Math.max(
        0,
        csvText
          .split(/\r?\n/)
          .map((line) => line.trim())
          .filter((line) => line !== '').length - 1
      )
      setCsvStatus(`Exported CSV with ${rowCount} item row${rowCount === 1 ? '' : 's'}.`)
    } catch {
      setCsvTone('destructive')
      setCsvStatus('CSV export failed.')
    } finally {
      setExportCsvPending(false)
    }
  }, [])

  const runExportLogs = useCallback(async () => {
    setLogsExportPending(true)
    setLogsTone('default')
    try {
      const response = await fetch('/api/logs/export')
      if (!response.ok) {
        throw new Error('failed_to_export_logs')
      }
      const text = await response.text()
      const preview = text.trim()
      setLogsPreview(preview || null)
      setLogsStatus('Exported runtime logs successfully.')
    } catch {
      setLogsPreview(null)
      setLogsTone('destructive')
      setLogsStatus('Runtime logs export failed.')
    } finally {
      setLogsExportPending(false)
    }
  }, [])

  const runImportCsvDryRun = useCallback(async () => {
    setImportCsvDryRunPending(true)
    setCsvSummary(null)
    setLastCsvDryRunRequest(null)
    setCsvTone('default')
    try {
      const body = buildCsvImportRequest()
      const response = await fetch('/api/data/import/csv/dry-run', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      })
      if (!response.ok) {
        throw new Error('failed_to_dry_run_csv_import')
      }
      const payload = (await response.json()) as ImportDryRunSummaryResponse
      setCsvSummary(payload)
      setLastCsvDryRunRequest(body)
      const conflictCount = Number(payload.conflicts ?? 0)
      setCsvStatus(
        `CSV dry-run completed with ${conflictCount} conflict${conflictCount === 1 ? '' : 's'}.`
      )
    } catch {
      setCsvSummary(null)
      setCsvTone('destructive')
      setCsvStatus('CSV dry-run failed.')
    } finally {
      setImportCsvDryRunPending(false)
    }
  }, [buildCsvImportRequest])

  const runImportCsvApply = useCallback(async () => {
    if (!lastCsvDryRunRequest) {
      return
    }
    setImportCsvApplyPending(true)
    setCsvTone('default')
    try {
      const response = await fetch('/api/data/import/csv/apply', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          csv_import: lastCsvDryRunRequest,
          options: {
            default_action: importCsvDefaultAction,
          },
        }),
      })
      if (!response.ok) {
        throw new Error('failed_to_apply_csv_import')
      }
      setCsvStatus('CSV import applied successfully.')
      setCsvSummary(null)
      setLastCsvDryRunRequest(null)
    } catch {
      setCsvTone('destructive')
      setCsvStatus('CSV import apply failed.')
    } finally {
      setImportCsvApplyPending(false)
    }
  }, [importCsvDefaultAction, lastCsvDryRunRequest])

  const runtimeAddress = runtimeInfo
    ? `${runtimeInfo.runtime_host?.trim() || '127.0.0.1'}:${runtimeInfo.runtime_port ?? 17880}`
    : 'Unavailable'
  const updateKeyStatus = runtimeInfo?.update_public_key_configured
    ? 'Configured'
    : 'Missing'
  const importActionsDisabled =
    loading ||
    Boolean(error) ||
    exportPending ||
    importDryRunPending ||
    importApplyPending
  const csvActionsDisabled =
    loading ||
    Boolean(error) ||
    exportCsvPending ||
    importCsvDryRunPending ||
    importCsvApplyPending
  const logsActionsDisabled = loading || Boolean(error) || logsExportPending

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
          data-testid='settings-operations-logs-card'
        >
          <div className='flex flex-wrap items-start justify-between gap-3'>
            <div>
              <p className='font-medium'>Diagnostics logs</p>
              <p className='text-muted-foreground'>
                Export the current redacted runtime log snapshot for recovery and support workflows.
              </p>
            </div>
            <Button
              variant='outline'
              size='sm'
              data-testid='settings-operations-export-logs'
              disabled={logsActionsDisabled}
              onClick={() => {
                void runExportLogs()
              }}
            >
              {logsExportPending ? 'Exporting Logs…' : 'Export Logs'}
            </Button>
          </div>

          <p
            data-testid='settings-operations-logs-status'
            className={logsTone === 'destructive' ? 'text-sm text-destructive' : 'text-sm text-muted-foreground'}
          >
            {logsStatus}
          </p>

          <div
            className='rounded-md border bg-muted/20 p-3 text-xs text-muted-foreground'
            data-testid='settings-operations-logs-preview'
          >
            {logsPreview ?? 'No log export preview yet. Export logs to review the current redacted snapshot.'}
          </div>
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
              disabled={importActionsDisabled}
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
                setImportSummary(null)
                setLastDryRunRequest(null)
              }}
              className='min-h-36 font-mono text-xs'
              spellCheck={false}
            />
            <div className='flex flex-wrap items-center justify-between gap-3'>
              <div className='space-y-1'>
                <label className='text-xs font-medium text-muted-foreground'>
                  Conflict default action
                </label>
                <Select
                  value={importDefaultAction}
                  onValueChange={(value) => {
                    setImportDefaultAction(value as ImportApplyAction)
                  }}
                >
                  <SelectTrigger
                    className='w-44'
                    data-testid='settings-operations-import-default-action'
                  >
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value='merge'>Merge into existing item</SelectItem>
                    <SelectItem value='create'>Create new item</SelectItem>
                    <SelectItem value='skip'>Skip conflicting item</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className='flex gap-2'>
                <Button
                  variant='outline'
                  size='sm'
                  data-testid='settings-operations-import-json-apply'
                  disabled={importActionsDisabled || !lastDryRunRequest || !importSummary}
                  onClick={() => {
                    void runImportJsonApply()
                  }}
                >
                  {importApplyPending ? 'Applying…' : 'Apply Import'}
                </Button>
                <Button
                  variant='outline'
                  size='sm'
                  data-testid='settings-operations-import-json-dry-run'
                  disabled={importActionsDisabled}
                  onClick={() => {
                    void runImportJsonDryRun()
                  }}
                >
                  {importDryRunPending ? 'Running Dry-Run…' : 'Run Dry-Run'}
                </Button>
              </div>
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

        <div
          className='rounded-md border p-3 space-y-3'
          data-testid='settings-operations-csv-card'
        >
          <div className='flex flex-wrap items-start justify-between gap-3'>
            <div>
              <p className='font-medium'>CSV import and export</p>
              <p className='text-muted-foreground'>
                Export item rows as CSV and dry-run CSV imports with the default item column mapping.
              </p>
            </div>
            <Button
              variant='outline'
              size='sm'
              data-testid='settings-operations-export-csv'
              disabled={csvActionsDisabled}
              onClick={() => {
                void runExportCsv()
              }}
            >
              {exportCsvPending ? 'Exporting…' : 'Export CSV'}
            </Button>
          </div>

          <p
            data-testid='settings-operations-csv-status'
            className={csvTone === 'destructive' ? 'text-sm text-destructive' : 'text-sm text-muted-foreground'}
          >
            {csvStatus}
          </p>

          <div className='space-y-2'>
            <label
              htmlFor='settings-operations-import-csv-input'
              className='text-sm font-medium'
            >
              CSV import dry-run
            </label>
            <Textarea
              id='settings-operations-import-csv-input'
              data-testid='settings-operations-import-csv-input'
              value={importCsvInput}
              onChange={(event) => {
                setImportCsvInput(event.target.value)
                setCsvSummary(null)
                setLastCsvDryRunRequest(null)
              }}
              className='min-h-36 font-mono text-xs'
              spellCheck={false}
            />
            <div className='space-y-2'>
              <p className='text-xs font-medium text-muted-foreground'>
                CSV column mapping
              </p>
              <div className='grid gap-2 md:grid-cols-2'>
                <div className='space-y-1'>
                  <label
                    htmlFor='settings-operations-import-csv-mapping-brand'
                    className='text-xs text-muted-foreground'
                  >
                    Brand column
                  </label>
                  <Input
                    id='settings-operations-import-csv-mapping-brand'
                    data-testid='settings-operations-import-csv-mapping-brand'
                    value={importCsvMapping.brand}
                    onChange={(event) => {
                      setImportCsvMapping((current) => ({
                        ...current,
                        brand: event.target.value,
                      }))
                      setCsvSummary(null)
                      setLastCsvDryRunRequest(null)
                    }}
                    placeholder='brand'
                  />
                </div>
                <div className='space-y-1'>
                  <label
                    htmlFor='settings-operations-import-csv-mapping-category'
                    className='text-xs text-muted-foreground'
                  >
                    Category column
                  </label>
                  <Input
                    id='settings-operations-import-csv-mapping-category'
                    data-testid='settings-operations-import-csv-mapping-category'
                    value={importCsvMapping.category}
                    onChange={(event) => {
                      setImportCsvMapping((current) => ({
                        ...current,
                        category: event.target.value,
                      }))
                      setCsvSummary(null)
                      setLastCsvDryRunRequest(null)
                    }}
                    placeholder='category'
                  />
                </div>
                <div className='space-y-1'>
                  <label
                    htmlFor='settings-operations-import-csv-mapping-part-number'
                    className='text-xs text-muted-foreground'
                  >
                    Part number column
                  </label>
                  <Input
                    id='settings-operations-import-csv-mapping-part-number'
                    data-testid='settings-operations-import-csv-mapping-part-number'
                    value={importCsvMapping.part_number}
                    onChange={(event) => {
                      setImportCsvMapping((current) => ({
                        ...current,
                        part_number: event.target.value,
                      }))
                      setCsvSummary(null)
                      setLastCsvDryRunRequest(null)
                    }}
                    placeholder='part_number'
                  />
                </div>
                <div className='space-y-1'>
                  <label
                    htmlFor='settings-operations-import-csv-mapping-title'
                    className='text-xs text-muted-foreground'
                  >
                    Title column
                  </label>
                  <Input
                    id='settings-operations-import-csv-mapping-title'
                    data-testid='settings-operations-import-csv-mapping-title'
                    value={importCsvMapping.title}
                    onChange={(event) => {
                      setImportCsvMapping((current) => ({
                        ...current,
                        title: event.target.value,
                      }))
                      setCsvSummary(null)
                      setLastCsvDryRunRequest(null)
                    }}
                    placeholder='title'
                  />
                </div>
              </div>
              <p className='text-xs text-muted-foreground'>
                Leave fields blank to use the default CSV headers.
              </p>
            </div>
            <div className='flex flex-wrap items-center justify-between gap-3'>
              <div className='space-y-1'>
                <label className='text-xs font-medium text-muted-foreground'>
                  Conflict default action
                </label>
                <Select
                  value={importCsvDefaultAction}
                  onValueChange={(value) => {
                    setImportCsvDefaultAction(value as ImportApplyAction)
                  }}
                >
                  <SelectTrigger
                    className='w-44'
                    data-testid='settings-operations-import-csv-default-action'
                  >
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value='merge'>Merge into existing item</SelectItem>
                    <SelectItem value='create'>Create new item</SelectItem>
                    <SelectItem value='skip'>Skip conflicting item</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className='flex gap-2'>
                <Button
                  variant='outline'
                  size='sm'
                  data-testid='settings-operations-import-csv-apply'
                  disabled={csvActionsDisabled || !lastCsvDryRunRequest || !csvSummary}
                  onClick={() => {
                    void runImportCsvApply()
                  }}
                >
                  {importCsvApplyPending ? 'Applying CSV…' : 'Apply CSV Import'}
                </Button>
                <Button
                  variant='outline'
                  size='sm'
                  data-testid='settings-operations-import-csv-dry-run'
                  disabled={csvActionsDisabled}
                  onClick={() => {
                    void runImportCsvDryRun()
                  }}
                >
                  {importCsvDryRunPending ? 'Running CSV Dry-Run…' : 'Run CSV Dry-Run'}
                </Button>
              </div>
            </div>
          </div>

          <div
            className='rounded-md border bg-muted/20 p-3 text-sm text-muted-foreground'
            data-testid='settings-operations-csv-summary'
          >
            {csvSummary ? (
              <div className='space-y-2'>
                <p>
                  {csvSummary.total_items ?? 0} items, {csvSummary.new_items ?? 0} new,{' '}
                  {csvSummary.conflicts ?? 0} conflict
                  {(csvSummary.conflicts ?? 0) === 1 ? '' : 's'}.
                </p>
                {csvSummary.conflict_details?.length ? (
                  <div className='space-y-1'>
                    {csvSummary.conflict_details.map((detail) => (
                      <p key={`csv-${detail.part_number || 'unknown'}-${detail.existing_id || 'missing'}`}>
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
              <p>No CSV dry-run summary yet. Paste CSV rows and run a dry-run to review conflicts.</p>
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
