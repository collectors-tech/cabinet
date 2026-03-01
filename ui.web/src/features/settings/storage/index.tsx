import { useCallback, useEffect, useState } from 'react'
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

export function SettingsStorage() {
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [dbPath, setDbPath] = useState('')
  const [mediaDir, setMediaDir] = useState('')

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
      setDbPath(storage.db_path?.trim() || 'Unavailable')
      setMediaDir(storage.media_dir?.trim() || 'Unavailable')
    } catch {
      setError('Storage information is unavailable right now.')
      setDbPath('')
      setMediaDir('')
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
          <Button variant='outline' size='sm'>Reindex Search</Button>
          <Button variant='outline' size='sm'>Rebuild Thumbnails</Button>
        </div>
      </div>
    </ContentSection>
  )
}
