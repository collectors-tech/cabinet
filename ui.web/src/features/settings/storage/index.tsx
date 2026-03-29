import { useCallback, useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Link } from '@tanstack/react-router'
import { AlertTriangle } from 'lucide-react'
import { ContentSection } from '../components/content-section'
import { Button } from '@/components/ui/button'

type ActiveProfileResponse = {
  id?: string
}

type StorageResponse = {
  db_path?: string
  media_dir?: string
}

const LAST_KNOWN_STORAGE_KEY = 'cabinet.settings.storage.lastKnown'

export function SettingsStorage() {
  const { t } = useTranslation('pages')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [dbPath, setDbPath] = useState('')
  const [mediaDir, setMediaDir] = useState('')
  const [reindexPending, setReindexPending] = useState(false)
  const [rebuildPending, setRebuildPending] = useState(false)
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

  const actionsDisabled = loading || Boolean(error) || reindexPending || rebuildPending

  return (
    <ContentSection
      title={t('settings.storage.title')}
      desc={t('settings.storage.description')}
    >
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
        {error ? (
          <p className='text-xs text-muted-foreground'>
            Diagnostics actions are unavailable while storage info is degraded.
          </p>
        ) : null}
      </div>
    </ContentSection>
  )
}
