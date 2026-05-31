import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Link } from '@tanstack/react-router'
import { AlertTriangle, Download } from 'lucide-react'
import { ContentSection } from '../components/content-section'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'

type ActiveProfileResponse = {
  id?: string
}

type StorageResponse = {
  db_path?: string
  media_dir?: string
}

type BackupInfo = {
  path?: string
  file_name?: string
  size_bytes?: number
  created_at?: string
  archive_format?: string
  download_url?: string
  integrity_check?: string
}

type BackupListResponse = {
  backups?: BackupInfo[]
}

type BackupSortKey = 'created_at' | 'file_name'
type BackupSortDirection = 'asc' | 'desc'

const LAST_KNOWN_STORAGE_KEY = 'cabinet.settings.storage.lastKnown'

export function SettingsStorage() {
  const { t } = useTranslation('pages')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [dbPath, setDbPath] = useState('')
  const [mediaDir, setMediaDir] = useState('')
  const [reindexPending, setReindexPending] = useState(false)
  const [rebuildPending, setRebuildPending] = useState(false)
  const [repairPending, setRepairPending] = useState(false)
  const [repairResult, setRepairResult] = useState<string | null>(null)
  const [repairTone, setRepairTone] = useState<'default' | 'destructive'>('default')
  const [backupList, setBackupList] = useState<BackupInfo[]>([])
  const [backupsLoading, setBackupsLoading] = useState(true)
  const [backupPending, setBackupPending] = useState(false)
  const [restorePending, setRestorePending] = useState(false)
  const [backupError, setBackupError] = useState<string | null>(null)
  const [backupSortKey, setBackupSortKey] = useState<BackupSortKey>('created_at')
  const [backupSortDirection, setBackupSortDirection] =
    useState<BackupSortDirection>('desc')
  const [restoreConfirmOpen, setRestoreConfirmOpen] = useState(false)
  const [selectedBackupPath, setSelectedBackupPath] = useState<string | null>(null)
  const [actionStatus, setActionStatus] = useState<string | null>(null)
  const [actionTone, setActionTone] = useState<'default' | 'destructive'>('default')
  const [lastKnown, setLastKnown] = useState<StorageResponse | null>(() => {
    if (typeof window === 'undefined') {
      return null
    }
    const raw = window.localStorage.getItem(LAST_KNOWN_STORAGE_KEY)
    if (!raw) {
      return null
    }
    try {
      const parsed = JSON.parse(raw) as StorageResponse
      if (!parsed.db_path && !parsed.media_dir) {
        return null
      }
      return parsed
    } catch {
      return null
    }
  })
  const lastKnownRef = useRef<StorageResponse | null>(lastKnown)

  const loadStorage = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const activeProfileResp = await fetch('/api/profiles/active')
      if (!activeProfileResp.ok) {
        throw new Error('active_profile_unavailable')
      }
      const activeProfile = (await activeProfileResp.json()) as ActiveProfileResponse
      const profileID = activeProfile.id?.trim()
      if (!profileID) {
        throw new Error('active_profile_unavailable')
      }

      const storageResp = await fetch(`/api/profiles/${profileID}/storage`)
      if (!storageResp.ok) {
        throw new Error('storage_unavailable')
      }
      const storage = (await storageResp.json()) as StorageResponse
      const next = {
        db_path: storage.db_path?.trim() || 'Unavailable',
        media_dir: storage.media_dir?.trim() || 'Unavailable',
      }
      setLastKnown(next)
      lastKnownRef.current = next
      if (typeof window !== 'undefined') {
        window.localStorage.setItem(LAST_KNOWN_STORAGE_KEY, JSON.stringify(next))
      }
      setDbPath(next.db_path)
      setMediaDir(next.media_dir)
    } catch {
      setError('Storage information is unavailable right now.')
      setDbPath(lastKnownRef.current?.db_path || 'Unavailable')
      setMediaDir(lastKnownRef.current?.media_dir || 'Unavailable')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void loadStorage()
  }, [loadStorage])

  const loadBackups = useCallback(async () => {
    setBackupsLoading(true)
    setBackupError(null)
    try {
      const response = await fetch('/api/backup/list')
      if (!response.ok) {
        throw new Error('failed_to_list_backups')
      }
      const payload = (await response.json()) as BackupListResponse
      setBackupList(
        (payload.backups ?? []).filter(
          (backup) => typeof backup.path === 'string' && backup.path.trim() !== ''
        )
      )
    } catch {
      setBackupList([])
      setBackupError('Backup information is unavailable right now.')
    } finally {
      setBackupsLoading(false)
    }
  }, [])

  useEffect(() => {
    void loadBackups()
  }, [loadBackups])

  const runReindex = useCallback(async () => {
    setReindexPending(true)
    setActionStatus(null)
    setActionTone('default')
    try {
      const response = await fetch('/api/data/reindex', { method: 'POST' })
      if (!response.ok) {
        throw new Error('failed_to_reindex')
      }
      setActionStatus('Search reindex completed successfully.')
    } catch {
      setActionTone('destructive')
      setActionStatus('Search reindex failed. Try again when runtime diagnostics are healthy.')
    } finally {
      setReindexPending(false)
    }
  }, [])

  const runRebuild = useCallback(async () => {
    setRebuildPending(true)
    setActionStatus(null)
    setActionTone('default')
    try {
      const response = await fetch('/api/data/rebuild-thumbnails', { method: 'POST' })
      if (!response.ok) {
        throw new Error('failed_to_rebuild_thumbnails')
      }
      const payload = (await response.json()) as {
        rebuilt_items?: number
        rebuilt_photos?: number
      }
      const rebuiltItems = Number(payload.rebuilt_items ?? 0)
      const rebuiltPhotos = Number(payload.rebuilt_photos ?? 0)
      setActionStatus(
        `Thumbnail rebuild completed for ${rebuiltPhotos} photo${rebuiltPhotos === 1 ? '' : 's'} across ${rebuiltItems} item${rebuiltItems === 1 ? '' : 's'}.`
      )
    } catch {
      setActionTone('destructive')
      setActionStatus(
        'Thumbnail rebuild failed. Check diagnostics health and try again.'
      )
    } finally {
      setRebuildPending(false)
    }
  }, [])

  const runRepair = useCallback(async () => {
    setRepairPending(true)
    setRepairResult(null)
    setRepairTone('default')
    try {
      const response = await fetch('/api/data/repair', { method: 'POST' })
      if (!response.ok) {
        throw new Error('failed_to_repair_check')
      }
      const payload = (await response.json()) as {
        integrity_check?: string
      }
      const integrityCheck = payload.integrity_check?.trim() || 'unknown'
      if (integrityCheck.toLowerCase() === 'ok') {
        setRepairResult(`Database integrity check passed. Result: ${integrityCheck}`)
      } else {
        setRepairResult(`Database integrity check reported issues. Result: ${integrityCheck}`)
      }
    } catch {
      setRepairTone('destructive')
      setRepairResult('Database integrity check failed. Check diagnostics and try again.')
    } finally {
      setRepairPending(false)
    }
  }, [])

  const runBackup = useCallback(async () => {
    setBackupPending(true)
    setActionStatus(null)
    setActionTone('default')
    try {
      const response = await fetch('/api/backup/run', { method: 'POST' })
      if (!response.ok) {
        throw new Error('failed_to_create_backup')
      }
      const payload = (await response.json()) as { backup?: BackupInfo }
      const fileName = payload.backup?.file_name?.trim() || 'backup snapshot'
      const integrityCheck = payload.backup?.integrity_check?.trim() || 'unknown'
      const archiveFormat = payload.backup?.archive_format?.trim().toUpperCase() || 'archive'
      setActionStatus(`Backup created successfully: ${fileName}. ${archiveFormat} ready for download. Integrity check: ${integrityCheck}.`)
      await loadBackups()
    } catch {
      setActionTone('destructive')
      setActionStatus('Backup creation failed. Try again when runtime storage is healthy.')
    } finally {
      setBackupPending(false)
    }
  }, [loadBackups])

  const runRestore = useCallback(async () => {
    if (!selectedBackupPath) {
      return
    }
    setRestorePending(true)
    setActionStatus(null)
    setActionTone('default')
    try {
      const response = await fetch('/api/backup/restore', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ backup_path: selectedBackupPath, confirm_restore: true }),
      })
      if (!response.ok) {
        throw new Error('failed_to_restore_backup')
      }
      setRestoreConfirmOpen(false)
      setSelectedBackupPath(null)
      const payload = (await response.json()) as {
        restore?: { restored_at?: string; integrity_check?: string }
      }
      const integrityCheck = payload.restore?.integrity_check?.trim() || 'unknown'
      setActionStatus(`Backup restored successfully. Integrity check: ${integrityCheck}.`)
      await Promise.all([loadStorage(), loadBackups()])
    } catch {
      setRestoreConfirmOpen(false)
      setSelectedBackupPath(null)
      setActionTone('destructive')
      setActionStatus('Backup restore failed. Verify the backup and try again.')
    } finally {
      setRestorePending(false)
    }
  }, [loadBackups, loadStorage, selectedBackupPath])

  const sortedBackupList = useMemo(() => {
    return [...backupList].sort((left, right) => {
      const direction = backupSortDirection === 'asc' ? 1 : -1
      if (backupSortKey === 'created_at') {
        const leftTime = Date.parse(left.created_at ?? '')
        const rightTime = Date.parse(right.created_at ?? '')
        const leftValue = Number.isFinite(leftTime) ? leftTime : 0
        const rightValue = Number.isFinite(rightTime) ? rightTime : 0
        if (leftValue !== rightValue) {
          return (leftValue - rightValue) * direction
        }
      }
      return backupDisplayName(left)
        .localeCompare(backupDisplayName(right), undefined, {
          numeric: true,
          sensitivity: 'base',
        }) * direction
    })
  }, [backupList, backupSortDirection, backupSortKey])

  const toggleBackupSort = useCallback((key: BackupSortKey) => {
    setBackupSortKey((currentKey) => {
      if (currentKey === key) {
        setBackupSortDirection((currentDirection) =>
          currentDirection === 'asc' ? 'desc' : 'asc'
        )
        return currentKey
      }
      setBackupSortDirection(key === 'created_at' ? 'desc' : 'asc')
      return key
    })
  }, [])

  const actionsDisabled =
    loading ||
    Boolean(error) ||
    reindexPending ||
    rebuildPending ||
    repairPending ||
    backupPending ||
    restorePending

  return (
    <ContentSection
      title={t('settings.storage.title')}
      desc={t('settings.storage.description')}
    >
      <div className='space-y-4 text-sm'>
        <div className='space-y-4 text-sm'>
          {error ? (
            <div className='rounded-md border border-destructive/50 bg-destructive/10 p-3 text-destructive'>
              <p className='font-medium'>{error}</p>
              <p className='mt-1 text-xs text-muted-foreground'>
                Check active profile selection and runtime connectivity, then retry.
              </p>
              <Button
                variant='outline'
                size='sm'
                className='mt-3'
                data-testid='settings-storage-retry'
                onClick={() => {
                  void loadStorage()
                }}
              >
                Retry
              </Button>
              <Button asChild variant='default' size='sm' className='mt-3 ml-2'>
                <Link to='/'>Create or Select Profile</Link>
              </Button>
            </div>
          ) : null}
          <div className='rounded-md border p-3'>
            <p className='font-medium'>Database location</p>
            <p className='text-muted-foreground'>
              {loading ? 'Loading storage paths...' : dbPath}
            </p>
          </div>
          <div className='rounded-md border p-3'>
            <p className='font-medium'>Media location</p>
            <p className='text-muted-foreground'>
              {loading ? 'Loading storage paths...' : mediaDir}
            </p>
          </div>
          <div className='rounded-md border border-amber-300/40 bg-amber-50/20 p-3 text-amber-950 dark:text-amber-200'>
            <div className='flex items-start gap-2'>
              <AlertTriangle className='mt-0.5 h-4 w-4' />
              <p>
                Storage repair and migration actions are restricted to diagnostics workflows.
              </p>
            </div>
          </div>
          <div className='rounded-md border p-3 space-y-3'>
            <div className='flex items-start justify-between gap-3'>
              <div>
                <p className='font-medium'>Database integrity</p>
                <p className='text-xs text-muted-foreground'>
                  Run an integrity check before deeper diagnostics or restore workflows.
                </p>
              </div>
              <Button
                variant='outline'
                size='sm'
                data-testid='settings-storage-repair-run'
                disabled={actionsDisabled}
                onClick={() => {
                  void runRepair()
                }}
              >
                {repairPending ? 'Running Integrity Check…' : 'Run Integrity Check'}
              </Button>
            </div>
            <p
              data-testid='settings-storage-repair-result'
              className={repairTone === 'destructive' ? 'text-sm text-destructive' : 'text-sm text-muted-foreground'}
            >
              {repairResult ?? 'No integrity check has been run yet.'}
            </p>
          </div>
          <div className='flex gap-2'>
            <Button
              variant='outline'
              size='sm'
              disabled={actionsDisabled}
              onClick={() => {
                void runReindex()
              }}
            >
              {reindexPending ? 'Reindexing Search…' : 'Reindex Search'}
            </Button>
            <Button
              variant='outline'
              size='sm'
              disabled={actionsDisabled}
              onClick={() => {
                void runRebuild()
              }}
            >
              {rebuildPending ? 'Rebuilding Thumbnails…' : 'Rebuild Thumbnails'}
            </Button>
          </div>
          {actionStatus ? (
            <p
              data-testid='settings-storage-action-status'
              className={actionTone === 'destructive' ? 'text-sm text-destructive' : 'text-sm text-muted-foreground'}
            >
              {actionStatus}
            </p>
          ) : null}
          <div className='rounded-md border p-3 space-y-3'>
            <div>
              <p className='font-medium'>Data exports</p>
              <p className='text-xs text-muted-foreground'>
                Download the active profile as a JSON snapshot or item CSV.
              </p>
            </div>
            <div className='flex flex-wrap gap-2'>
              {actionsDisabled ? (
                <>
                  <Button variant='outline' size='sm' disabled>
                    <Download className='mr-2 h-4 w-4' />
                    JSON Snapshot
                  </Button>
                  <Button variant='outline' size='sm' disabled>
                    <Download className='mr-2 h-4 w-4' />
                    Item CSV
                  </Button>
                </>
              ) : (
                <>
                  <Button asChild variant='outline' size='sm'>
                    <a
                      data-testid='settings-storage-export-json'
                      href='/api/data/export/json'
                      download='cabinet-data-snapshot.json'
                    >
                      <Download className='mr-2 h-4 w-4' />
                      JSON Snapshot
                    </a>
                  </Button>
                  <Button asChild variant='outline' size='sm'>
                    <a
                      data-testid='settings-storage-export-csv'
                      href='/api/data/export/csv/items'
                      download='cabinet-items.csv'
                    >
                      <Download className='mr-2 h-4 w-4' />
                      Item CSV
                    </a>
                  </Button>
                </>
              )}
            </div>
          </div>
          <div className='rounded-md border p-3 space-y-3'>
            <div className='flex items-start justify-between gap-3'>
              <div>
                <p className='font-medium'>Backups</p>
                <p className='text-xs text-muted-foreground'>
                  Create a backup snapshot and restore an existing backup when needed.
                </p>
              </div>
              <Button
                variant='outline'
                size='sm'
                data-testid='settings-storage-backup-run'
                disabled={actionsDisabled || backupsLoading}
                onClick={() => {
                  void runBackup()
                }}
              >
                {backupPending ? 'Creating Backup…' : 'Create Backup'}
              </Button>
            </div>
            {backupsLoading ? (
              <p className='text-sm text-muted-foreground'>Loading backups...</p>
            ) : null}
            {backupError ? (
              <p className='text-sm text-destructive'>{backupError}</p>
            ) : null}
            {!backupsLoading && !backupError && backupList.length === 0 ? (
              <p className='text-sm text-muted-foreground'>
                No backups available yet. Create one to capture the current workspace.
              </p>
            ) : null}
            {!backupsLoading && backupList.length > 0 ? (
              <div className='overflow-x-auto rounded-md border'>
                <table
                  className='w-full min-w-[760px] table-fixed text-left text-sm'
                  data-testid='settings-storage-backup-table'
                >
                  <thead className='bg-muted/40 text-xs uppercase text-muted-foreground'>
                    <tr>
                      <th className='w-[28%] px-3 py-2 font-medium'>
                        <button
                          type='button'
                          className='font-medium text-foreground hover:underline'
                          data-testid='settings-storage-backup-sort-file'
                          onClick={() => {
                            toggleBackupSort('file_name')
                          }}
                        >
                          Filename {backupSortKey === 'file_name' ? sortIndicator(backupSortDirection) : ''}
                        </button>
                      </th>
                      <th className='w-[18%] px-3 py-2 font-medium'>
                        <button
                          type='button'
                          className='font-medium text-foreground hover:underline'
                          data-testid='settings-storage-backup-sort-created'
                          onClick={() => {
                            toggleBackupSort('created_at')
                          }}
                        >
                          Created {backupSortKey === 'created_at' ? sortIndicator(backupSortDirection) : ''}
                        </button>
                      </th>
                      <th className='w-[17%] px-3 py-2 font-medium'>Backup source</th>
                      <th className='w-[12%] px-3 py-2 font-medium'>Archive size</th>
                      <th className='w-[12%] px-3 py-2 font-medium'>Status</th>
                      <th className='w-[13%] px-3 py-2 text-right font-medium'>Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {sortedBackupList.map((backup) => (
                      <tr
                        key={backup.path}
                        className='border-t align-top'
                        data-testid='settings-storage-backup-row'
                      >
                        <td className='px-3 py-2'>
                          <p className='break-all font-medium text-foreground'>
                            {backupDisplayName(backup)}
                          </p>
                          <p className='mt-1 break-all text-xs text-muted-foreground'>
                            {backup.path}
                          </p>
                        </td>
                        <td className='px-3 py-2 text-muted-foreground'>
                          {formatBackupTimestamp(backup.created_at)}
                        </td>
                        <td className='px-3 py-2 text-muted-foreground'>
                          {backupSourceLabel(backup)}
                        </td>
                        <td className='px-3 py-2 text-muted-foreground'>
                          {formatBackupSize(backup.size_bytes)}
                        </td>
                        <td className='px-3 py-2 text-muted-foreground'>
                          {formatBackupStatus(backup)}
                        </td>
                        <td className='px-3 py-2'>
                          <div className='flex flex-wrap justify-end gap-2'>
                            {backup.download_url ? (
                              <Button asChild variant='outline' size='sm'>
                                <a
                                  data-testid='settings-storage-backup-download'
                                  href={backup.download_url}
                                  download={backup.file_name || undefined}
                                >
                                  <Download className='mr-2 h-4 w-4' />
                                  Download
                                </a>
                              </Button>
                            ) : null}
                            <Button
                              variant='outline'
                              size='sm'
                              data-testid='settings-storage-backup-restore'
                              disabled={actionsDisabled}
                              onClick={() => {
                                setSelectedBackupPath(backup.path ?? null)
                                setRestoreConfirmOpen(true)
                              }}
                            >
                              Restore
                            </Button>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            ) : null}
          </div>
          {error ? (
            <p className='text-xs text-muted-foreground'>
              Diagnostics actions are unavailable while storage info is degraded.
            </p>
          ) : null}
        </div>
        <Dialog open={restoreConfirmOpen} onOpenChange={setRestoreConfirmOpen}>
          <DialogContent data-testid='settings-storage-restore-confirm'>
            <div className='space-y-4'>
              <DialogHeader>
                <DialogTitle>Restore Backup</DialogTitle>
              </DialogHeader>
              <p className='text-sm text-muted-foreground'>
                Restore the selected backup snapshot? This replaces the current runtime database with the selected backup.
              </p>
              <p className='rounded-md border p-2 text-xs text-muted-foreground break-all'>
                {selectedBackupPath ?? 'No backup selected'}
              </p>
              <DialogFooter>
                <Button
                  variant='outline'
                  onClick={() => {
                    setRestoreConfirmOpen(false)
                    setSelectedBackupPath(null)
                  }}
                >
                  Cancel
                </Button>
                <Button
                  data-testid='settings-storage-restore-submit'
                  disabled={restorePending || !selectedBackupPath}
                  onClick={() => {
                    void runRestore()
                  }}
                >
                  {restorePending ? 'Restoring…' : 'Restore Backup'}
                </Button>
              </DialogFooter>
            </div>
          </DialogContent>
        </Dialog>
      </div>
    </ContentSection>
  )
}

function formatBackupSize(sizeBytes?: number) {
  if (!Number.isFinite(sizeBytes) || !sizeBytes || sizeBytes <= 0) {
    return 'size unavailable'
  }
  if (sizeBytes < 1024) {
    return `${sizeBytes} B`
  }
  if (sizeBytes < 1024 * 1024) {
    return `${(sizeBytes / 1024).toFixed(1)} KB`
  }
  return `${(sizeBytes / (1024 * 1024)).toFixed(1)} MB`
}

function formatBackupFormat(format?: string) {
  const normalized = format?.trim().toLowerCase()
  if (normalized === 'zip') {
    return 'ZIP archive'
  }
  if (normalized === 'db') {
    return 'legacy DB snapshot'
  }
  return 'archive format unavailable'
}

function backupDisplayName(backup: BackupInfo) {
  return backup.file_name?.trim() || backup.path?.trim() || 'backup artifact'
}

function formatBackupTimestamp(timestamp?: string) {
  const parsed = Date.parse(timestamp ?? '')
  if (!Number.isFinite(parsed)) {
    return 'created time unavailable'
  }
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(parsed))
}

function backupSourceLabel(backup: BackupInfo) {
  const name = backupDisplayName(backup).toLowerCase()
  const format = backup.archive_format?.trim().toLowerCase()
  if (format === 'zip' || name.startsWith('cabinet-backup-')) {
    return 'Generated ZIP archive'
  }
  if (format === 'db' || name.endsWith('.db')) {
    return 'Legacy database snapshot'
  }
  return formatBackupFormat(backup.archive_format)
}

function formatBackupStatus(backup: BackupInfo) {
  const integrity = backup.integrity_check?.trim()
  if (integrity) {
    return integrity.toLowerCase() === 'ok' ? 'Valid' : integrity
  }
  return 'Validity unknown'
}

function sortIndicator(direction: BackupSortDirection) {
  return direction === 'asc' ? '↑' : '↓'
}
