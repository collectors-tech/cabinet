import { useCallback, useEffect, useRef, useState } from 'react'
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
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [dbPath, setDbPath] = useState('')
  const [mediaDir, setMediaDir] = useState('')
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

  return (
    <ContentSection
      title='Storage'
      desc='Review database/media paths and run storage maintenance actions.'
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
          <Button variant='outline' size='sm' disabled>
            Reindex Search (Diagnostics only)
          </Button>
          <Button variant='outline' size='sm' disabled>
            Rebuild Thumbnails (Diagnostics only)
          </Button>
        </div>
        {error ? (
          <p className='text-xs text-muted-foreground'>
            Diagnostics actions are unavailable while storage info is degraded.
          </p>
        ) : null}
      </div>
    </ContentSection>
  )
}
