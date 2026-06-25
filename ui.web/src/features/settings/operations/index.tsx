import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import { recordNotificationHistory } from '@/lib/toast-history'
import { ContentSection } from '../components/content-section'
import { useProfileSettings } from '../use-profile-settings'

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

type RuntimeSetupImportResponse = {
  ok?: boolean
  setup_required?: boolean
  instance_name?: string
  profile_key?: string
  config_path?: string
  runtime_url?: string
  runtime_port?: number
}

type RecoveryResetResponse = {
  session_id?: string
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

type ImportApplySummaryResponse = {
  total_items?: number
  created?: number
  merged?: number
  skipped?: number
  failed?: number
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

const queueWorkerScheduleSettingKey = 'scanner_schedule'
const queueResumeScheduleSettingKey = 'operations_queue_resume_schedule'
const pausedQueueWorkerSchedule = 'manual'
const defaultQueueWorkerResumeSchedule = '0 */6 * * *'

function normalizeQueueWorkerSchedule(value?: string): string {
  const trimmed = value?.trim() ?? ''
  return trimmed === '' ? pausedQueueWorkerSchedule : trimmed
}

function buildQueueStatusMessage(
  schedule: string,
  resumeSchedule: string,
  profileSettingsLoading: boolean,
  profileSettingsError: string | null
): string {
  if (profileSettingsLoading) {
    return 'Loading worker scheduling…'
  }
  if (profileSettingsError) {
    return 'Queue controls are unavailable right now.'
  }
  if (schedule === pausedQueueWorkerSchedule) {
    return `Workers paused. Resume will restore schedule ${resumeSchedule}.`
  }
  return `Workers scheduled: ${schedule}`
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

function formatImportApplySummary(
  prefix: string,
  summary: ImportApplySummaryResponse
): string {
  const totalItems = Number(summary.total_items ?? 0)
  const created = Number(summary.created ?? 0)
  const merged = Number(summary.merged ?? 0)
  const skipped = Number(summary.skipped ?? 0)
  const failed = Number(summary.failed ?? 0)

  return `${prefix}: ${totalItems} item${totalItems === 1 ? '' : 's'}, ${created} created, ${merged} merged, ${skipped} skipped, ${failed} failed.`
}

export function SettingsOperations() {
  const { t } = useTranslation('pages')
  const {
    activeProfileId,
    settings: profileSettings,
    loading: profileSettingsLoading,
    error: profileSettingsError,
    saving: profileSettingsSaving,
    saveSettings,
  } = useProfileSettings()
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [runtimeInfo, setRuntimeInfo] = useState<RuntimeResponse | null>(null)
  const [recoveryInfo, setRecoveryInfo] =
    useState<RuntimeRecoveryResponse | null>(null)
  const [setupImportSourcePath, setSetupImportSourcePath] = useState('')
  const [setupImportPending, setSetupImportPending] = useState(false)
  const [setupImportStatus, setSetupImportStatus] = useState(
    'No runtime setup import has been run yet.'
  )
  const [setupImportTone, setSetupImportTone] = useState<
    'default' | 'destructive'
  >('default')
  const [setupImportSummary, setSetupImportSummary] =
    useState<RuntimeSetupImportResponse | null>(null)
  const [dataStatus, setDataStatus] = useState<string>(
    'No import or export action has run yet.'
  )
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
  const [importCsvMapping, setImportCsvMapping] =
    useState<CSVImportMappingState>({
      brand: '',
      category: '',
      part_number: '',
      title: '',
    })
  const [lastExportSummary, setLastExportSummary] =
    useState<ExportSnapshotResponse | null>(null)
  const [importSummary, setImportSummary] =
    useState<ImportDryRunSummaryResponse | null>(null)
  const [csvStatus, setCsvStatus] = useState<string>(
    'No CSV import or export action has run yet.'
  )
  const [csvTone, setCsvTone] = useState<'default' | 'destructive'>('default')
  const [csvSummary, setCsvSummary] =
    useState<ImportDryRunSummaryResponse | null>(null)
  const [logsExportPending, setLogsExportPending] = useState(false)
  const [logsStatus, setLogsStatus] = useState<string>(
    'No diagnostics export has run yet.'
  )
  const [logsTone, setLogsTone] = useState<'default' | 'destructive'>('default')
  const [logsPreview, setLogsPreview] = useState<string | null>(null)
  const [importDefaultAction, setImportDefaultAction] =
    useState<ImportApplyAction>('merge')
  const [importCsvDefaultAction, setImportCsvDefaultAction] =
    useState<ImportApplyAction>('merge')
  const [lastDryRunRequest, setLastDryRunRequest] =
    useState<ImportSnapshotRequest | null>(null)
  const [lastCsvDryRunRequest, setLastCsvDryRunRequest] =
    useState<CSVImportRequest | null>(null)
  const [queuePendingAction, setQueuePendingAction] = useState<
    'pause' | 'resume' | null
  >(null)
  const [queueStatusOverride, setQueueStatusOverride] = useState<string | null>(
    null
  )
  const [queueTone, setQueueTone] = useState<'default' | 'destructive'>(
    'default'
  )
  const [recoveryPassphraseInput, setRecoveryPassphraseInput] = useState('')
  const [recoveryPassphrasePending, setRecoveryPassphrasePending] =
    useState(false)
  const [recoveryResetPending, setRecoveryResetPending] = useState(false)
  const [authRecoveryStatus, setAuthRecoveryStatus] = useState(
    'No recovery passphrase or reset action has run yet.'
  )
  const [authRecoveryTone, setAuthRecoveryTone] = useState<
    'default' | 'destructive'
  >('default')
  const [authRecoverySummary, setAuthRecoverySummary] =
    useState<RecoveryResetResponse | null>(null)

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
      const recoveryPayload =
        (await recoveryResp.json()) as RuntimeRecoveryResponse
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
      setDataStatus(
        `Exported ${itemCount} item snapshot${itemCount === 1 ? '' : 's'}.`
      )
    } catch {
      setLastExportSummary(null)
      setDataTone('destructive')
      setDataStatus(
        'Export failed. Retry when runtime data services are healthy.'
      )
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
      setDataStatus(
        'Import dry-run failed. No records were changed; fix the JSON snapshot and run dry-run again.'
      )
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
      const payload = (await response.json()) as ImportApplySummaryResponse
      setDataStatus(formatImportApplySummary('Import applied', payload))
      setImportSummary(null)
      setLastDryRunRequest(null)
    } catch {
      setDataTone('destructive')
      setDataStatus(
        'Import apply failed. No records were changed; review the dry-run summary and retry when data services are healthy.'
      )
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
      setCsvStatus(
        `Exported CSV with ${rowCount} item row${rowCount === 1 ? '' : 's'}.`
      )
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
      recordNotificationHistory({
        id: 'settings-operations-logs-export-success',
        level: 'success',
        title: 'Exported runtime logs successfully.',
        summary: 'Diagnostics logs status from Settings Operations.',
        source_label: 'Settings Operations',
        category: 'system',
      })
    } catch {
      setLogsPreview(null)
      setLogsTone('destructive')
      setLogsStatus('Runtime logs export failed.')
      recordNotificationHistory({
        id: 'settings-operations-logs-export-failed',
        level: 'error',
        title: 'Runtime logs export failed.',
        summary: 'Diagnostics logs status from Settings Operations.',
        source_label: 'Settings Operations',
        category: 'system',
      })
    } finally {
      setLogsExportPending(false)
    }
  }, [])

  const runSetupImport = useCallback(async () => {
    setSetupImportPending(true)
    setSetupImportTone('default')
    try {
      const response = await fetch('/api/runtime/setup-import', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          source_path: setupImportSourcePath.trim(),
        }),
      })
      if (!response.ok) {
        throw new Error('failed_to_import_runtime_setup')
      }
      const payload = (await response.json()) as RuntimeSetupImportResponse
      setSetupImportSummary(payload)
      setSetupImportStatus('Runtime setup imported successfully.')
      recordNotificationHistory({
        id: 'settings-operations-setup-import-success',
        level: 'success',
        title: 'Runtime setup imported successfully.',
        summary:
          'Runtime setup import status from Settings Operations was preserved in Inbox history.',
        source_label: 'Settings Operations',
        category: 'system',
      })
      await loadOperations()
    } catch {
      setSetupImportSummary(null)
      setSetupImportTone('destructive')
      setSetupImportStatus('Runtime setup import failed.')
      recordNotificationHistory({
        id: 'settings-operations-setup-import-failed',
        level: 'error',
        title: 'Runtime setup import failed.',
        summary:
          'Runtime setup import failure from Settings Operations was preserved in Inbox history.',
        source_label: 'Settings Operations',
        category: 'system',
      })
    } finally {
      setSetupImportPending(false)
    }
  }, [loadOperations, setupImportSourcePath])

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
      setCsvStatus(
        'CSV dry-run failed. No records were changed; fix the CSV rows or mapping and run dry-run again.'
      )
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
      const payload = (await response.json()) as ImportApplySummaryResponse
      setCsvStatus(formatImportApplySummary('CSV import applied', payload))
      setCsvSummary(null)
      setLastCsvDryRunRequest(null)
    } catch {
      setCsvTone('destructive')
      setCsvStatus(
        'CSV import apply failed. No records were changed; review the CSV dry-run summary and retry when data services are healthy.'
      )
    } finally {
      setImportCsvApplyPending(false)
    }
  }, [importCsvDefaultAction, lastCsvDryRunRequest])

  const queueWorkerSchedule = normalizeQueueWorkerSchedule(
    profileSettings[queueWorkerScheduleSettingKey]
  )
  const queuePaused = queueWorkerSchedule === pausedQueueWorkerSchedule
  const queueResumeSchedule =
    profileSettings[queueResumeScheduleSettingKey]?.trim() ||
    (queuePaused ? defaultQueueWorkerResumeSchedule : queueWorkerSchedule)
  const queueStatus =
    queueStatusOverride ??
    buildQueueStatusMessage(
      queueWorkerSchedule,
      queueResumeSchedule,
      profileSettingsLoading,
      profileSettingsError
    )

  const runPauseWorkers = useCallback(async () => {
    if (!activeProfileId) {
      setQueueTone('destructive')
      setQueueStatusOverride(
        'Queue controls are unavailable without an active profile.'
      )
      return
    }

    const resumeSchedule =
      queueWorkerSchedule === pausedQueueWorkerSchedule
        ? queueResumeSchedule
        : queueWorkerSchedule

    setQueuePendingAction('pause')
    setQueueTone('default')
    try {
      await saveSettings({
        ...profileSettings,
        [queueWorkerScheduleSettingKey]: pausedQueueWorkerSchedule,
        [queueResumeScheduleSettingKey]: resumeSchedule,
      })
      setQueueStatusOverride(
        `Workers paused. Resume will restore schedule ${resumeSchedule}.`
      )
    } catch {
      setQueueTone('destructive')
      setQueueStatusOverride(
        'Failed to pause workers. Retry when profile settings are available.'
      )
    } finally {
      setQueuePendingAction(null)
    }
  }, [
    activeProfileId,
    profileSettings,
    queueResumeSchedule,
    queueWorkerSchedule,
    saveSettings,
  ])

  const runResumeWorkers = useCallback(async () => {
    if (!activeProfileId) {
      setQueueTone('destructive')
      setQueueStatusOverride(
        'Queue controls are unavailable without an active profile.'
      )
      return
    }

    setQueuePendingAction('resume')
    setQueueTone('default')
    try {
      await saveSettings({
        ...profileSettings,
        [queueWorkerScheduleSettingKey]: queueResumeSchedule,
        [queueResumeScheduleSettingKey]: queueResumeSchedule,
      })
      setQueueStatusOverride(`Workers scheduled: ${queueResumeSchedule}`)
    } catch {
      setQueueTone('destructive')
      setQueueStatusOverride(
        'Failed to resume workers. Retry when profile settings are available.'
      )
    } finally {
      setQueuePendingAction(null)
    }
  }, [activeProfileId, profileSettings, queueResumeSchedule, saveSettings])

  const runSetRecoveryPassphrase = useCallback(async () => {
    if (!activeProfileId) {
      setAuthRecoveryTone('destructive')
      setAuthRecoveryStatus(
        'Recovery access is unavailable without an active profile.'
      )
      return
    }

    setRecoveryPassphrasePending(true)
    setAuthRecoveryTone('default')
    try {
      const response = await fetch('/api/auth/recovery/passphrase', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          profile_id: activeProfileId,
          passphrase: recoveryPassphraseInput,
        }),
      })
      if (!response.ok) {
        throw new Error('failed_to_set_recovery_passphrase')
      }
      setAuthRecoverySummary(null)
      setAuthRecoveryStatus('Recovery passphrase saved.')
    } catch {
      setAuthRecoveryTone('destructive')
      setAuthRecoveryStatus('Recovery passphrase save failed.')
    } finally {
      setRecoveryPassphrasePending(false)
    }
  }, [activeProfileId, recoveryPassphraseInput])

  const runBeginRecoveryReset = useCallback(async () => {
    if (!activeProfileId) {
      setAuthRecoveryTone('destructive')
      setAuthRecoveryStatus(
        'Recovery access is unavailable without an active profile.'
      )
      return
    }

    setRecoveryResetPending(true)
    setAuthRecoveryTone('default')
    try {
      const response = await fetch('/api/auth/recovery/reset/begin', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          profile_id: activeProfileId,
          passphrase: recoveryPassphraseInput,
        }),
      })
      if (!response.ok) {
        throw new Error('failed_to_begin_recovery_reset')
      }
      const payload = (await response.json()) as RecoveryResetResponse
      setAuthRecoverySummary(payload)
      setAuthRecoveryStatus('Recovery reset session started.')
    } catch {
      setAuthRecoverySummary(null)
      setAuthRecoveryTone('destructive')
      setAuthRecoveryStatus('Recovery reset could not be started.')
    } finally {
      setRecoveryResetPending(false)
    }
  }, [activeProfileId, recoveryPassphraseInput])

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
  const setupImportActionsDisabled =
    loading ||
    Boolean(error) ||
    setupImportPending ||
    setupImportSourcePath.trim() === ''
  const queueActionsDisabled =
    loading ||
    Boolean(error) ||
    profileSettingsLoading ||
    Boolean(profileSettingsError) ||
    profileSettingsSaving ||
    queuePendingAction !== null
  const authRecoveryActionsDisabled =
    loading ||
    Boolean(error) ||
    profileSettingsLoading ||
    Boolean(profileSettingsError) ||
    recoveryPassphraseInput.trim() === '' ||
    recoveryPassphrasePending ||
    recoveryResetPending

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
              Check runtime health and retry without leaving the Operations
              screen.
            </p>
          </div>
        ) : null}

        <div
          className='space-y-2 rounded-md border p-3'
          data-testid='settings-operations-runtime-card'
        >
          <div className='flex items-start justify-between gap-3'>
            <div>
              <p className='font-medium'>Runtime</p>
              <p className='text-muted-foreground'>
                Live runtime version, bind mode, and update channel for this
                lane.
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
          <p>
            Version:{' '}
            {loading
              ? 'Loading runtime...'
              : runtimeInfo?.app_version || 'Unavailable'}
          </p>
          <p>
            Build date:{' '}
            {loading
              ? 'Loading runtime...'
              : runtimeInfo?.build_date || 'Unavailable'}
          </p>
          <p>Address: {loading ? 'Loading runtime...' : runtimeAddress}</p>
          <p>
            Bind mode:{' '}
            {loading
              ? 'Loading runtime...'
              : runtimeInfo?.bind_mode || 'Unavailable'}
          </p>
          <p>
            Update channel:{' '}
            {loading
              ? 'Loading runtime...'
              : runtimeInfo?.update_channel || 'Unavailable'}
          </p>
          <p>
            Update signing key:{' '}
            {loading ? 'Loading runtime...' : updateKeyStatus}
          </p>
        </div>

        <div
          className='space-y-2 rounded-md border p-3'
          data-testid='settings-operations-recovery-card'
        >
          <p className='font-medium'>Recovery state</p>
          <p className='text-muted-foreground'>
            Surface whether the last runtime shutdown requires recovery
            attention.
          </p>
          <p>
            {loading
              ? 'Loading recovery state...'
              : recoveryInfo?.recovery_required
                ? 'Recovery required'
                : 'No recovery required'}
          </p>

          <div className='space-y-2 border-t pt-3'>
            <label
              htmlFor='settings-operations-setup-import-source'
              className='text-sm font-medium'
            >
              Runtime setup import
            </label>
            <p className='text-muted-foreground'>
              Import a runtime setup config from disk to recover a missing or
              damaged local setup.
            </p>
            <Input
              id='settings-operations-setup-import-source'
              data-testid='settings-operations-setup-import-source'
              value={setupImportSourcePath}
              onChange={(event) => {
                setSetupImportSourcePath(event.target.value)
                setSetupImportSummary(null)
              }}
              placeholder='C:\\cabinet\\recovery\\setup-import-source.json'
            />
            <div className='flex justify-end'>
              <Button
                variant='outline'
                size='sm'
                data-testid='settings-operations-setup-import-submit'
                disabled={setupImportActionsDisabled}
                onClick={() => {
                  void runSetupImport()
                }}
              >
                {setupImportPending
                  ? 'Importing Setup…'
                  : 'Import Setup Config'}
              </Button>
            </div>
            <p
              data-testid='settings-operations-setup-import-status'
              className={
                setupImportTone === 'destructive'
                  ? 'text-sm text-destructive'
                  : 'text-sm text-muted-foreground'
              }
            >
              {setupImportStatus}
            </p>
            <div
              className='rounded-md border bg-muted/20 p-3 text-sm text-muted-foreground'
              data-testid='settings-operations-setup-import-summary'
            >
              {setupImportSummary ? (
                <div className='space-y-1'>
                  <p>
                    Instance:{' '}
                    {setupImportSummary.instance_name || 'Unknown instance'}
                  </p>
                  <p>
                    Profile:{' '}
                    {setupImportSummary.profile_key || 'Unknown profile'}
                  </p>
                  <p>
                    Config:{' '}
                    {setupImportSummary.config_path || 'Unknown config path'}
                  </p>
                  <p>
                    Runtime URL:{' '}
                    {setupImportSummary.runtime_url || 'Unknown runtime URL'}
                  </p>
                </div>
              ) : (
                <p>No imported setup summary yet.</p>
              )}
            </div>
          </div>
        </div>

        <div
          className='space-y-3 rounded-md border p-3'
          data-testid='settings-operations-auth-recovery-card'
        >
          <div className='space-y-1'>
            <p className='font-medium'>Recovery Access</p>
            <p className='text-muted-foreground'>
              Set a recovery passphrase and begin a recovery reset session for
              the active profile without leaving Operations.
            </p>
          </div>

          <div className='space-y-2'>
            <label
              htmlFor='settings-operations-recovery-passphrase-input'
              className='text-sm font-medium'
            >
              Recovery passphrase
            </label>
            <Input
              id='settings-operations-recovery-passphrase-input'
              data-testid='settings-operations-recovery-passphrase-input'
              type='password'
              value={recoveryPassphraseInput}
              onChange={(event) => {
                setRecoveryPassphraseInput(event.target.value)
              }}
              placeholder='Enter recovery passphrase'
            />
          </div>

          <div className='flex gap-2'>
            <Button
              variant='outline'
              size='sm'
              data-testid='settings-operations-recovery-passphrase-submit'
              disabled={authRecoveryActionsDisabled}
              onClick={() => {
                void runSetRecoveryPassphrase()
              }}
            >
              {recoveryPassphrasePending
                ? 'Saving Passphrase…'
                : 'Save Recovery Passphrase'}
            </Button>
            <Button
              variant='outline'
              size='sm'
              data-testid='settings-operations-recovery-reset-submit'
              disabled={authRecoveryActionsDisabled}
              onClick={() => {
                void runBeginRecoveryReset()
              }}
            >
              {recoveryResetPending
                ? 'Starting Recovery Reset…'
                : 'Begin Recovery Reset'}
            </Button>
          </div>

          <p
            data-testid='settings-operations-auth-recovery-status'
            className={
              authRecoveryTone === 'destructive'
                ? 'text-sm text-destructive'
                : 'text-sm text-muted-foreground'
            }
          >
            {authRecoveryStatus}
          </p>

          <div
            className='rounded-md border bg-muted/20 p-3 text-sm text-muted-foreground'
            data-testid='settings-operations-auth-recovery-summary'
          >
            {authRecoverySummary ? (
              <div className='space-y-1'>
                <p>
                  Recovery session:{' '}
                  {authRecoverySummary.session_id || 'Unknown session'}
                </p>
                <p>Profile: {activeProfileId || 'Unknown profile'}</p>
              </div>
            ) : (
              <p>No recovery reset session has been started yet.</p>
            )}
          </div>
        </div>

        <div
          className='space-y-3 rounded-md border p-3'
          data-testid='settings-operations-logs-card'
        >
          <div className='flex flex-wrap items-start justify-between gap-3'>
            <div>
              <p className='font-medium'>Diagnostics logs</p>
              <p className='text-muted-foreground'>
                Export the current redacted runtime log snapshot for recovery
                and support workflows.
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
            className={
              logsTone === 'destructive'
                ? 'text-sm text-destructive'
                : 'text-sm text-muted-foreground'
            }
          >
            {logsStatus}
          </p>

          <div
            className='rounded-md border bg-muted/20 p-3 text-xs text-muted-foreground'
            data-testid='settings-operations-logs-preview'
          >
            {logsPreview ??
              'No log export preview yet. Export logs to review the current redacted snapshot.'}
          </div>
        </div>

        <div
          className='space-y-3 rounded-md border p-3'
          data-testid='settings-operations-data-card'
        >
          <div className='flex flex-wrap items-start justify-between gap-3'>
            <div>
              <p className='font-medium'>Data import and export</p>
              <p className='text-muted-foreground'>
                Export a JSON snapshot and dry-run a JSON import before applying
                any workspace changes.
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
            className={
              dataTone === 'destructive'
                ? 'text-sm text-destructive'
                : 'text-sm text-muted-foreground'
            }
          >
            {dataStatus}
          </p>

          {lastExportSummary ? (
            <div className='rounded-md border bg-muted/20 p-3 text-sm text-muted-foreground'>
              <p>
                Snapshot schema {lastExportSummary.schema_version ?? 1} with{' '}
                {lastExportSummary.items?.length ?? 0} item
                {(lastExportSummary.items?.length ?? 0) === 1 ? '' : 's'}{' '}
                exported at {lastExportSummary.exported_at || 'unknown time'}.
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
                    <SelectItem value='merge'>
                      Merge into existing item
                    </SelectItem>
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
                  disabled={
                    importActionsDisabled ||
                    !lastDryRunRequest ||
                    !importSummary
                  }
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
                  {importSummary.total_items ?? 0} items,{' '}
                  {importSummary.new_items ?? 0} new,{' '}
                  {importSummary.conflicts ?? 0} conflict
                  {(importSummary.conflicts ?? 0) === 1 ? '' : 's'}.
                </p>
                {importSummary.conflict_details?.length ? (
                  <div className='space-y-1'>
                    {importSummary.conflict_details.map((detail) => (
                      <p
                        key={`${detail.part_number || 'unknown'}-${detail.existing_id || 'missing'}`}
                      >
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
              <p>
                No dry-run summary yet. Paste a snapshot and run a dry-run to
                review conflicts.
              </p>
            )}
          </div>
        </div>

        <div
          className='space-y-3 rounded-md border p-3'
          data-testid='settings-operations-csv-card'
        >
          <div className='flex flex-wrap items-start justify-between gap-3'>
            <div>
              <p className='font-medium'>CSV import and export</p>
              <p className='text-muted-foreground'>
                Export item rows as CSV and dry-run CSV imports with the default
                item column mapping.
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
            className={
              csvTone === 'destructive'
                ? 'text-sm text-destructive'
                : 'text-sm text-muted-foreground'
            }
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
                    <SelectItem value='merge'>
                      Merge into existing item
                    </SelectItem>
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
                  disabled={
                    csvActionsDisabled || !lastCsvDryRunRequest || !csvSummary
                  }
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
                  {importCsvDryRunPending
                    ? 'Running CSV Dry-Run…'
                    : 'Run CSV Dry-Run'}
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
                  {csvSummary.total_items ?? 0} items,{' '}
                  {csvSummary.new_items ?? 0} new, {csvSummary.conflicts ?? 0}{' '}
                  conflict
                  {(csvSummary.conflicts ?? 0) === 1 ? '' : 's'}.
                </p>
                {csvSummary.conflict_details?.length ? (
                  <div className='space-y-1'>
                    {csvSummary.conflict_details.map((detail) => (
                      <p
                        key={`csv-${detail.part_number || 'unknown'}-${detail.existing_id || 'missing'}`}
                      >
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
              <p>
                No CSV dry-run summary yet. Paste CSV rows and run a dry-run to
                review conflicts.
              </p>
            )}
          </div>
        </div>

        <div
          className='space-y-3 rounded-md border p-3'
          data-testid='settings-operations-queue-card'
        >
          <p className='font-medium'>Queue Controls</p>
          <p className='text-muted-foreground'>
            Pause and resume Market Watch and enrichment workers for active
            profile context.
          </p>
          <p
            data-testid='settings-operations-queue-status'
            className={
              queueTone === 'destructive'
                ? 'text-sm text-destructive'
                : 'text-sm text-muted-foreground'
            }
          >
            {queueStatus}
          </p>
          <div className='flex gap-2'>
            <Button
              variant='outline'
              size='sm'
              data-testid='settings-operations-queue-pause'
              disabled={queueActionsDisabled || queuePaused}
              onClick={() => {
                void runPauseWorkers()
              }}
            >
              {queuePendingAction === 'pause'
                ? 'Pausing Workers…'
                : 'Pause Workers'}
            </Button>
            <Button
              variant='outline'
              size='sm'
              data-testid='settings-operations-queue-resume'
              disabled={queueActionsDisabled || !queuePaused}
              onClick={() => {
                void runResumeWorkers()
              }}
            >
              {queuePendingAction === 'resume'
                ? 'Resuming Workers…'
                : 'Resume Workers'}
            </Button>
          </div>
        </div>
      </div>
    </ContentSection>
  )
}
