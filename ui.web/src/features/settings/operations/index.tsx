import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { ContentSection } from '../components/content-section'
import { Button } from '@/components/ui/button'

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

export function SettingsOperations() {
  const { t } = useTranslation('pages')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [runtimeInfo, setRuntimeInfo] = useState<RuntimeResponse | null>(null)
  const [recoveryInfo, setRecoveryInfo] = useState<RuntimeRecoveryResponse | null>(null)

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
