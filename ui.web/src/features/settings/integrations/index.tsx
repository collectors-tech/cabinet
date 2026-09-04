import { useCallback, useEffect, useMemo, useState } from 'react'
import { KeyRound, Loader2, PlugZap, RotateCw, ShieldCheck } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { recordNotificationHistory } from '@/lib/toast-history'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { ContentSection } from '../components/content-section'
import { useProfileSettings } from '../use-profile-settings'

type MCPHTTPStatusResponse = {
  enabled?: boolean
  state?: string
  listen_addr?: string
  credential_configured?: boolean
  guidance?: string
  recovery_action?: string
  credential?: string
  last_diagnostic_outcome?: {
    operation_id?: string
    capability?: string
    method?: string
    input_class?: string
    outcome?: string
    error_class?: string
  }
}

type MCPCredentialResponse = {
  credential?: string
  credential_configured?: boolean
  configuration_guidance?: string
}

const defaultListenAddr = '127.0.0.1:17890'

export function SettingsIntegrations() {
  const { t } = useTranslation('pages')
  const { activeProfileId, loading: profileLoading } = useProfileSettings()
  const [status, setStatus] = useState<MCPHTTPStatusResponse | null>(null)
  const [listenAddr, setListenAddr] = useState(defaultListenAddr)
  const [credential, setCredential] = useState('')
  const [credentialNotice, setCredentialNotice] = useState(
    'No credential has been generated in this session.'
  )
  const [loading, setLoading] = useState(true)
  const [pendingAction, setPendingAction] = useState<
    'refresh' | 'config' | 'credential' | null
  >(null)
  const [error, setError] = useState<string | null>(null)

  const loadStatus = useCallback(async () => {
    if (!activeProfileId) {
      if (!profileLoading) {
        setError('Active profile is unavailable.')
        setLoading(false)
      }
      return
    }

    setLoading(true)
    setError(null)
    try {
      const response = await fetch(
        `/api/profiles/${encodeURIComponent(activeProfileId)}/mcp-http-status`
      )
      if (!response.ok) {
        throw new Error('mcp_status_unavailable')
      }
      const payload = (await response.json()) as MCPHTTPStatusResponse
      setStatus(payload)
      setListenAddr(payload.listen_addr?.trim() || defaultListenAddr)
    } catch {
      setStatus(null)
      setError('MCP transport status is unavailable right now.')
    } finally {
      setLoading(false)
    }
  }, [activeProfileId, profileLoading])

  useEffect(() => {
    void loadStatus()
  }, [loadStatus])

  const saveConfig = useCallback(
    async (enabled: boolean) => {
      if (!activeProfileId) {
        return
      }
      setPendingAction('config')
      setError(null)
      try {
        const response = await fetch(
          `/api/profiles/${encodeURIComponent(activeProfileId)}/mcp-http-config`,
          {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
              enabled,
              listen_addr: listenAddr.trim() || defaultListenAddr,
            }),
          }
        )
        if (!response.ok) {
          throw new Error('mcp_config_save_failed')
        }
        const payload = (await response.json()) as MCPHTTPStatusResponse
        setStatus(payload)
        setListenAddr(payload.listen_addr?.trim() || defaultListenAddr)
        recordNotificationHistory({
          id: enabled
            ? 'settings-integrations-mcp-http-enabled'
            : 'settings-integrations-mcp-http-disabled',
          level: 'success',
          title: enabled
            ? 'MCP HTTP transport enabled.'
            : 'MCP HTTP transport disabled.',
          summary: 'MCP transport status from Settings Integrations.',
          source_label: 'Settings Integrations',
          category: 'system',
        })
      } catch {
        setError('MCP transport configuration could not be saved.')
      } finally {
        setPendingAction(null)
      }
    },
    [activeProfileId, listenAddr]
  )

  const generateCredential = useCallback(async () => {
    if (!activeProfileId) {
      return
    }
    setPendingAction('credential')
    setError(null)
    setCredential('')
    try {
      const response = await fetch(
        `/api/profiles/${encodeURIComponent(activeProfileId)}/mcp-http-credential`,
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: '{}',
        }
      )
      if (!response.ok) {
        throw new Error('mcp_credential_failed')
      }
      const payload = (await response.json()) as MCPCredentialResponse
      setCredential(payload.credential?.trim() ?? '')
      setCredentialNotice(
        payload.configuration_guidance?.trim() ||
          'Credential generated for this profile.'
      )
      await loadStatus()
      recordNotificationHistory({
        id: 'settings-integrations-mcp-http-credential',
        level: 'success',
        title: 'MCP HTTP credential generated.',
        summary: 'MCP transport credential status from Settings Integrations.',
        source_label: 'Settings Integrations',
        category: 'system',
      })
    } catch {
      setCredentialNotice('MCP HTTP credential generation failed.')
      setError('MCP HTTP credential could not be generated.')
    } finally {
      setPendingAction(null)
    }
  }, [activeProfileId, loadStatus])

  const refreshStatus = useCallback(async () => {
    setPendingAction('refresh')
    setCredential('')
    setCredentialNotice('No credential has been generated in this session.')
    await loadStatus()
    setPendingAction(null)
  }, [loadStatus])

  const stateLabel = useMemo(() => {
    if (loading) {
      return 'Loading'
    }
    const state = status?.state?.trim()
    if (!state) {
      return 'Unavailable'
    }
    return state.charAt(0).toUpperCase() + state.slice(1)
  }, [loading, status?.state])

  const actionsDisabled =
    loading || profileLoading || !activeProfileId || !!pendingAction
  const refreshDisabled = profileLoading || !activeProfileId || !!pendingAction

  return (
    <ContentSection
      title={t('settings.integrations.title')}
      desc={t('settings.integrations.description')}
      contentClassName='lg:max-w-3xl'
    >
      <div className='space-y-4 text-sm'>
        {error ? (
          <div
            className='rounded-md border border-destructive/50 bg-destructive/10 p-3 text-destructive'
            data-testid='settings-integrations-mcp-error'
          >
            {error}
          </div>
        ) : null}

        <div
          className='space-y-4 rounded-md border p-3'
          data-testid='settings-integrations-mcp-card'
        >
          <div className='flex flex-wrap items-start justify-between gap-3'>
            <div className='space-y-1'>
              <div className='flex items-center gap-2'>
                <PlugZap className='h-4 w-4 text-muted-foreground' />
                <p className='font-medium'>Cabinet MCP</p>
              </div>
              <p className='text-muted-foreground'>
                Profile-bound local MCP transport for trusted clients.
              </p>
            </div>
            <Button
              variant='outline'
              size='sm'
              data-testid='settings-integrations-mcp-refresh'
              disabled={refreshDisabled}
              onClick={() => {
                void refreshStatus()
              }}
            >
              {pendingAction === 'refresh' ? (
                <Loader2 className='mr-2 h-4 w-4 animate-spin' />
              ) : (
                <RotateCw className='mr-2 h-4 w-4' />
              )}
              Refresh
            </Button>
          </div>

          <div className='grid gap-3 md:grid-cols-3'>
            <div className='rounded-md border bg-muted/20 p-3'>
              <p className='text-xs font-medium text-muted-foreground'>State</p>
              <p data-testid='settings-integrations-mcp-state'>{stateLabel}</p>
            </div>
            <div className='rounded-md border bg-muted/20 p-3'>
              <p className='text-xs font-medium text-muted-foreground'>
                Profile
              </p>
              <p
                className='break-all'
                data-testid='settings-integrations-mcp-profile'
              >
                {activeProfileId || 'Unavailable'}
              </p>
            </div>
            <div className='rounded-md border bg-muted/20 p-3'>
              <p className='text-xs font-medium text-muted-foreground'>
                Credential
              </p>
              <p data-testid='settings-integrations-mcp-credential-state'>
                {status?.credential_configured ? 'Configured' : 'Missing'}
              </p>
            </div>
          </div>

          <div className='flex flex-wrap items-center justify-between gap-3 rounded-md border p-3'>
            <div className='space-y-1'>
              <Label htmlFor='settings-integrations-mcp-enabled'>
                HTTP transport
              </Label>
              <p className='text-xs text-muted-foreground'>
                {status?.guidance ||
                  'Enable only for loopback clients bound to this profile.'}
              </p>
            </div>
            <Switch
              id='settings-integrations-mcp-enabled'
              data-testid='settings-integrations-mcp-enabled'
              checked={Boolean(status?.enabled)}
              disabled={actionsDisabled}
              onCheckedChange={(checked) => {
                void saveConfig(checked)
              }}
            />
          </div>

          <div className='grid gap-3 md:grid-cols-[1fr_auto]'>
            <div className='space-y-2'>
              <Label htmlFor='settings-integrations-mcp-listen'>
                Listen address
              </Label>
              <Input
                id='settings-integrations-mcp-listen'
                data-testid='settings-integrations-mcp-listen'
                value={listenAddr}
                disabled={actionsDisabled}
                onChange={(event) => {
                  setListenAddr(event.target.value)
                }}
              />
            </div>
            <div className='flex items-end'>
              <Button
                variant='outline'
                data-testid='settings-integrations-mcp-save-config'
                disabled={actionsDisabled}
                onClick={() => {
                  void saveConfig(Boolean(status?.enabled))
                }}
              >
                <ShieldCheck className='mr-2 h-4 w-4' />
                Save
              </Button>
            </div>
          </div>

          <div className='space-y-3 rounded-md border p-3'>
            <div className='flex flex-wrap items-start justify-between gap-3'>
              <div className='space-y-1'>
                <p className='font-medium'>HTTP bearer credential</p>
                <p className='text-xs text-muted-foreground'>
                  {credentialNotice}
                </p>
              </div>
              <Button
                variant='outline'
                size='sm'
                data-testid='settings-integrations-mcp-generate-credential'
                disabled={actionsDisabled}
                onClick={() => {
                  void generateCredential()
                }}
              >
                {pendingAction === 'credential' ? (
                  <Loader2 className='mr-2 h-4 w-4 animate-spin' />
                ) : (
                  <KeyRound className='mr-2 h-4 w-4' />
                )}
                Generate
              </Button>
            </div>
            <div
              className='rounded-md border bg-muted/20 p-3 font-mono text-xs break-all text-muted-foreground'
              data-testid='settings-integrations-mcp-generated-credential'
            >
              {credential || 'Credential appears only immediately after setup.'}
            </div>
          </div>

          <div className='rounded-md border bg-muted/20 p-3'>
            <p className='text-xs font-medium text-muted-foreground'>
              Last diagnostic
            </p>
            <p data-testid='settings-integrations-mcp-last-diagnostic'>
              {status?.last_diagnostic_outcome?.outcome
                ? [
                    status.last_diagnostic_outcome.outcome,
                    status.last_diagnostic_outcome.capability,
                    status.last_diagnostic_outcome.error_class,
                  ]
                    .filter(Boolean)
                    .join(' - ')
                : 'No diagnostic outcome recorded.'}
            </p>
          </div>
        </div>
      </div>
    </ContentSection>
  )
}
