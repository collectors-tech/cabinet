import { useEffect, useState } from 'react'
import { Link, useSearch } from '@tanstack/react-router'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { AuthLayout } from '../auth-layout'
import { UserAuthForm } from './components/user-auth-form'

type RuntimeSetupStatus = {
  setup_required?: boolean
  config_path?: string
  default_storage_data_dir?: string
  default_runtime_host?: string
  default_runtime_port?: number
  default_runtime_port_mode?: 'auto' | 'fixed'
  default_runtime_url?: string
}

type RuntimeSetupCompletePayload = {
  ok?: boolean
  instance_name?: string
  profile_key?: string
  config_path?: string
  data_dir?: string
  media_dir?: string
  runtime_url?: string
  runtime_port?: number
}

type SetupFormState = {
  instanceName: string
  profileKey: string
  storageMode: 'exe_local' | 'custom'
  customDataDir: string
  portableMode: boolean
  runtimePortMode: 'auto' | 'fixed'
  runtimeFixedPort: number
  authMode: 'local' | 'clerk'
  clerkPublishableKey: string
  featureChat: boolean
  featureProviders: boolean
  featureScanner: boolean
}

type SetupEntryMode = 'welcome' | 'form' | 'import'

export function SignIn() {
  const { redirect } = useSearch({ from: '/(auth)/sign-in' })
  const [setupLoading, setSetupLoading] = useState(true)
  const [setupRequired, setSetupRequired] = useState(false)
  const [setupError, setSetupError] = useState<string | null>(null)
  const [setupConfigPath, setSetupConfigPath] = useState('')
  const [setupDefaultStorageDataDir, setSetupDefaultStorageDataDir] =
    useState('')
  const [setupDefaultRuntimeHost, setSetupDefaultRuntimeHost] =
    useState('127.0.0.1')
  const [setupDefaultRuntimePort, setSetupDefaultRuntimePort] = useState(17880)
  const [completingSetup, setCompletingSetup] = useState(false)
  const [setupStep, setSetupStep] = useState(0)
  const [setupCompleteState, setSetupCompleteState] =
    useState<RuntimeSetupCompletePayload | null>(null)
  const [setupCompleteFeedback, setSetupCompleteFeedback] = useState('')
  const [setupEntryMode, setSetupEntryMode] =
    useState<SetupEntryMode>('welcome')
  const [setupImportSourcePath, setSetupImportSourcePath] = useState('')
  const [setupForm, setSetupForm] = useState<SetupFormState>({
    instanceName: 'Primary',
    profileKey: 'primary',
    storageMode: 'exe_local',
    customDataDir: '',
    portableMode: false,
    runtimePortMode: 'auto',
    runtimeFixedPort: 17880,
    authMode: 'local',
    clerkPublishableKey: '',
    featureChat: true,
    featureProviders: true,
    featureScanner: true,
  })

  const totalSteps = 6
  const progressPercent = Math.round(((setupStep + 1) / totalSteps) * 100)

  useEffect(() => {
    let cancelled = false
    async function loadSetupStatus() {
      setSetupLoading(true)
      setSetupError(null)
      try {
        const response = await fetch('/api/runtime/setup-status')
        if (!response.ok) {
          throw new Error(`setup_status_failed_${response.status}`)
        }
        const payload = (await response.json()) as RuntimeSetupStatus
        if (!cancelled) {
          setSetupRequired(Boolean(payload.setup_required))
          setSetupConfigPath(payload.config_path ?? '')
          setSetupDefaultStorageDataDir(payload.default_storage_data_dir ?? '')
          setSetupDefaultRuntimeHost(
            payload.default_runtime_host ?? '127.0.0.1'
          )
          setSetupDefaultRuntimePort(payload.default_runtime_port ?? 17880)
          setSetupForm((previous) => ({
            ...previous,
            runtimePortMode: payload.default_runtime_port_mode ?? 'auto',
            runtimeFixedPort: payload.default_runtime_port ?? 17880,
          }))
        }
      } catch (error) {
        if (!cancelled) {
          setSetupError(
            error instanceof Error ? error.message : 'setup_status_failed'
          )
          setSetupRequired(false)
        }
      } finally {
        if (!cancelled) {
          setSetupLoading(false)
        }
      }
    }
    void loadSetupStatus()
    return () => {
      cancelled = true
    }
  }, [])

  useEffect(() => {
    if (setupRequired) {
      setSetupEntryMode('welcome')
      setSetupStep(0)
      setSetupImportSourcePath('')
      setSetupCompleteState(null)
    }
  }, [setupRequired])

  function selectedStorageDataDir() {
    if (setupForm.storageMode === 'custom') {
      return setupForm.customDataDir.trim()
    }
    return setupDefaultStorageDataDir.trim()
  }

  function selectedRuntimePort() {
    return setupForm.runtimePortMode === 'fixed'
      ? setupForm.runtimeFixedPort
      : setupDefaultRuntimePort
  }

  function selectedRuntimeURL() {
    return `http://${setupDefaultRuntimeHost}:${selectedRuntimePort()}`
  }

  function authReadinessLabel() {
    if (setupForm.authMode === 'local') {
      return 'Ready: Local auth'
    }
    return setupForm.clerkPublishableKey.trim() === ''
      ? 'Missing Clerk key'
      : 'Configured'
  }

  async function goToNextStep() {
    if (setupStep === 0 && setupForm.instanceName.trim() === '') {
      setSetupError('Instance name is required.')
      return
    }
    if (setupStep === 1) {
      if (
        setupForm.storageMode === 'custom' &&
        setupForm.customDataDir.trim() === ''
      ) {
        setSetupError('Custom storage path is required.')
        return
      }
      setCompletingSetup(true)
      try {
        const response = await fetch('/api/runtime/setup-storage-validate', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            data_dir: selectedStorageDataDir(),
          }),
        })
        const payload = (await response.json()) as {
          writable?: boolean
          message?: string
        }
        if (!response.ok || payload.writable !== true) {
          setSetupError(payload.message ?? 'Storage path is not writable.')
          return
        }
      } catch {
        setSetupError('Storage validation failed.')
        return
      } finally {
        setCompletingSetup(false)
      }
    }
    if (setupStep === 2) {
      if (
        setupForm.runtimePortMode === 'fixed' &&
        setupForm.runtimeFixedPort <= 0
      ) {
        setSetupError(
          'Fixed port value is required when runtime port mode is fixed.'
        )
        return
      }
    }
    if (setupStep === 3) {
      if (
        setupForm.authMode === 'clerk' &&
        setupForm.clerkPublishableKey.trim() === ''
      ) {
        setSetupError('Clerk publishable key is required.')
        return
      }
    }
    setSetupError(null)
    setSetupStep((previous) => Math.min(previous + 1, totalSteps - 1))
  }

  function goToPreviousStep() {
    setSetupError(null)
    setSetupStep((previous) => Math.max(previous - 1, 0))
  }

  async function submitSetupComplete(
    requestPayload: Record<string, unknown>,
    options?: { defaultsApplied?: boolean }
  ) {
    setCompletingSetup(true)
    setSetupError(null)
    try {
      const response = await fetch('/api/runtime/setup-complete', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(requestPayload),
      })
      if (!response.ok) {
        let message = `setup_complete_failed_${response.status}`
        try {
          const errorPayload = (await response.json()) as { message?: string }
          if (errorPayload?.message) {
            message = errorPayload.message
          }
        } catch {
          // no-op
        }
        throw new Error(message)
      }
      const completePayload =
        (await response.json()) as RuntimeSetupCompletePayload
      setSetupCompleteState(completePayload)
      setSetupCompleteFeedback(
        options?.defaultsApplied
          ? 'Defaults applied. You can refine these settings later in Settings.'
          : ''
      )
    } catch (error) {
      setSetupError(
        error instanceof Error ? error.message : 'setup_complete_failed'
      )
    } finally {
      setCompletingSetup(false)
    }
  }

  async function completeSetup() {
    if (
      setupForm.authMode === 'clerk' &&
      setupForm.clerkPublishableKey.trim() === ''
    ) {
      setSetupError('Clerk publishable key is required.')
      return
    }

    await submitSetupComplete({
      instance_name: setupForm.instanceName.trim(),
      profile_key: setupForm.profileKey.trim(),
      storage_mode: setupForm.storageMode,
      storage_data_dir: selectedStorageDataDir(),
      portable_mode: setupForm.portableMode,
      runtime_port_mode: setupForm.runtimePortMode,
      runtime_fixed_port:
        setupForm.runtimePortMode === 'fixed' ? setupForm.runtimeFixedPort : 0,
      auth_mode: setupForm.authMode,
      clerk_publishable_key:
        setupForm.authMode === 'clerk'
          ? setupForm.clerkPublishableKey.trim()
          : '',
      feature_chat: setupForm.featureChat,
      feature_providers: setupForm.featureProviders,
      feature_scanner: setupForm.featureScanner,
      bootstrap_workspace: 'Local Workspace',
      bootstrap_database_ref: 'Primary DB',
    })
  }

  async function applyDefaultsSetup() {
    await submitSetupComplete(
      {
        instance_name: 'Cabinet Local',
        profile_key: 'default',
        storage_mode: 'exe_local',
        storage_data_dir: setupDefaultStorageDataDir,
        portable_mode: false,
        runtime_port_mode: 'auto',
        runtime_fixed_port: 0,
        auth_mode: 'local',
        clerk_publishable_key: '',
        feature_chat: true,
        feature_providers: true,
        feature_scanner: true,
        bootstrap_workspace: 'Local Workspace',
        bootstrap_database_ref: 'Primary DB',
      },
      { defaultsApplied: true }
    )
  }

  function startSetupFlow() {
    setSetupError(null)
    setSetupEntryMode('form')
    setSetupStep(0)
  }

  function openSetupImportFlow() {
    setSetupError(null)
    setSetupEntryMode('import')
  }

  function backToSetupWelcome() {
    setSetupError(null)
    setSetupEntryMode('welcome')
  }

  async function importExistingSetupConfig() {
    setCompletingSetup(true)
    setSetupError(null)
    try {
      const response = await fetch('/api/runtime/setup-import', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          source_path: setupImportSourcePath.trim(),
        }),
      })
      if (!response.ok) {
        let message = `setup_import_failed_${response.status}`
        try {
          const payload = (await response.json()) as { message?: string }
          if (payload?.message) {
            message = payload.message
          }
        } catch {
          // no-op
        }
        throw new Error(message)
      }
      setSetupRequired(false)
      setSetupEntryMode('welcome')
      setSetupStep(0)
      setSetupImportSourcePath('')
    } catch (error) {
      setSetupError(
        error instanceof Error ? error.message : 'setup_import_failed'
      )
    } finally {
      setCompletingSetup(false)
    }
  }

  function startAppFromSetup() {
    setSetupCompleteState(null)
    setSetupCompleteFeedback('')
    setSetupRequired(false)
    setSetupStep(0)
  }

  function openConfigFolderFromSetup() {
    const targetPath = setupCompleteState?.config_path ?? 'unknown path'
    setSetupCompleteFeedback(
      `Config folder action requested for ${targetPath}.`
    )
  }

  if (setupLoading) {
    return (
      <AuthLayout>
        <Card className='gap-4'>
          <CardHeader>
            <CardTitle className='text-lg tracking-tight'>
              Setup Wizard
            </CardTitle>
            <CardDescription>Checking startup configuration...</CardDescription>
          </CardHeader>
        </Card>
      </AuthLayout>
    )
  }

  if (setupRequired) {
    if (setupCompleteState?.ok) {
      return (
        <AuthLayout>
          <Card className='gap-4' data-testid='setup-wizard-complete-state'>
            <CardHeader>
              <CardTitle className='text-lg tracking-tight'>
                Config complete
              </CardTitle>
              <CardDescription>
                Setup is complete. Start Cabinet with your configuration.
              </CardDescription>
            </CardHeader>
            <CardContent className='space-y-3'>
              <p className='text-sm text-muted-foreground'>
                Instance:{' '}
                <span
                  className='font-medium'
                  data-testid='setup-complete-instance-name'
                >
                  {setupCompleteState.instance_name ?? 'unknown'}
                </span>
              </p>
              <p className='text-sm text-muted-foreground'>
                Config path:{' '}
                <span
                  className='font-medium'
                  data-testid='setup-complete-config-path'
                >
                  {setupCompleteState.config_path ?? 'unknown'}
                </span>
              </p>
              <p className='text-sm text-muted-foreground'>
                Data directory:{' '}
                <span
                  className='font-medium'
                  data-testid='setup-complete-data-dir'
                >
                  {setupCompleteState.data_dir ?? 'unknown'}
                </span>
              </p>
              <p className='text-sm text-muted-foreground'>
                Media directory:{' '}
                <span
                  className='font-medium'
                  data-testid='setup-complete-media-dir'
                >
                  {setupCompleteState.media_dir ?? 'unknown'}
                </span>
              </p>
              <p className='text-sm text-muted-foreground'>
                Runtime URL:{' '}
                <span
                  className='font-medium'
                  data-testid='setup-complete-runtime-url'
                >
                  {setupCompleteState.runtime_url ?? 'unknown'}
                </span>
              </p>
              <p className='text-sm text-muted-foreground'>
                Runtime port:{' '}
                <span
                  className='font-medium'
                  data-testid='setup-complete-runtime-port'
                >
                  {setupCompleteState.runtime_port ?? 0}
                </span>
              </p>
              {setupCompleteFeedback ? (
                <p
                  className='text-xs text-muted-foreground'
                  data-testid='setup-complete-feedback'
                >
                  {setupCompleteFeedback}
                </p>
              ) : null}
              <div className='flex flex-wrap items-center gap-2'>
                <Button
                  data-testid='setup-open-cabinet'
                  onClick={startAppFromSetup}
                >
                  Open Cabinet
                </Button>
                <Button
                  variant='outline'
                  data-testid='setup-open-config-folder'
                  onClick={openConfigFolderFromSetup}
                >
                  Open Config Folder
                </Button>
                <Button
                  variant='outline'
                  data-testid='setup-finish'
                  onClick={startAppFromSetup}
                >
                  Finish
                </Button>
              </div>
            </CardContent>
          </Card>
        </AuthLayout>
      )
    }

    return (
      <AuthLayout>
        <Card className='gap-4' data-testid='setup-wizard'>
          <CardHeader>
            <CardTitle className='text-lg tracking-tight'>
              Setup Wizard
            </CardTitle>
            <CardDescription>
              Complete initial runtime setup before continuing to sign in.
            </CardDescription>
            {setupEntryMode === 'form' ? (
              <>
                <p
                  className='text-xs font-semibold tracking-wide text-muted-foreground uppercase'
                  data-testid='setup-step-indicator'
                >
                  STEP {setupStep + 1} OF {totalSteps}
                </p>
                <div className='flex items-center justify-between gap-2'>
                  <p className='text-sm text-muted-foreground'>
                    Setup Progress
                  </p>
                  <p
                    className='text-sm font-medium'
                    data-testid='setup-step-percent'
                  >
                    {progressPercent}%
                  </p>
                </div>
                <div
                  className='h-2 w-full overflow-hidden rounded-full bg-muted'
                  role='progressbar'
                  aria-valuemin={0}
                  aria-valuemax={100}
                  aria-valuenow={progressPercent}
                  data-testid='setup-progress-bar'
                >
                  <div
                    className='h-full bg-primary transition-all'
                    style={{ width: `${progressPercent}%` }}
                  />
                </div>
              </>
            ) : null}
          </CardHeader>
          <CardContent className='space-y-3'>
            <p className='text-sm text-muted-foreground'>
              Cabinet detected missing startup configuration. Create the setup
              file to unlock login.
            </p>
            {setupEntryMode === 'welcome' ? (
              <div className='flex flex-wrap items-center gap-2'>
                <Button data-testid='setup-start' onClick={startSetupFlow}>
                  Start Setup
                </Button>
                <Button
                  variant='secondary'
                  data-testid='setup-use-defaults'
                  disabled={completingSetup}
                  onClick={() => void applyDefaultsSetup()}
                >
                  {completingSetup ? 'Applying Defaults...' : 'Use Defaults'}
                </Button>
                <Button
                  variant='outline'
                  data-testid='setup-import-toggle'
                  onClick={openSetupImportFlow}
                >
                  Import Existing Config
                </Button>
              </div>
            ) : null}
            {setupEntryMode === 'import' ? (
              <div className='grid gap-3'>
                <label className='grid gap-1 text-sm'>
                  <span>Existing Config Path</span>
                  <input
                    className='h-9 rounded-md border bg-background px-3'
                    value={setupImportSourcePath}
                    onChange={(event) =>
                      setSetupImportSourcePath(event.target.value)
                    }
                    data-testid='setup-import-source-path'
                  />
                </label>
                <div className='flex flex-wrap items-center gap-2'>
                  <Button
                    data-testid='setup-import-submit'
                    disabled={completingSetup}
                    onClick={() => void importExistingSetupConfig()}
                  >
                    {completingSetup
                      ? 'Importing...'
                      : 'Import Existing Config'}
                  </Button>
                  <Button
                    variant='outline'
                    data-testid='setup-import-cancel'
                    disabled={completingSetup}
                    onClick={backToSetupWelcome}
                  >
                    Back
                  </Button>
                </div>
              </div>
            ) : null}
            {setupEntryMode === 'form' && setupStep === 0 ? (
              <div className='grid gap-3'>
                <label className='grid gap-1 text-sm'>
                  <span>Instance Name</span>
                  <input
                    className='h-9 rounded-md border bg-background px-3'
                    value={setupForm.instanceName}
                    onChange={(event) =>
                      setSetupForm((previous) => ({
                        ...previous,
                        instanceName: event.target.value,
                      }))
                    }
                    data-testid='setup-instance-name'
                  />
                </label>
                <label className='grid gap-1 text-sm'>
                  <span>Profile Key</span>
                  <input
                    className='h-9 rounded-md border bg-background px-3'
                    value={setupForm.profileKey}
                    onChange={(event) =>
                      setSetupForm((previous) => ({
                        ...previous,
                        profileKey: event.target.value,
                      }))
                    }
                    data-testid='setup-profile-key'
                  />
                </label>
                <p className='text-xs text-muted-foreground'>
                  Config path preview:{' '}
                  <span
                    className='font-medium'
                    data-testid='setup-config-path-preview'
                  >
                    {setupConfigPath || 'unknown'}
                  </span>
                </p>
              </div>
            ) : null}
            {setupEntryMode === 'form' && setupStep === 1 ? (
              <div className='grid gap-3'>
                <label className='grid gap-1 text-sm'>
                  <span>Storage Mode</span>
                  <select
                    className='h-9 rounded-md border bg-background px-3'
                    value={setupForm.storageMode}
                    onChange={(event) =>
                      setSetupForm((previous) => ({
                        ...previous,
                        storageMode:
                          event.target.value === 'custom'
                            ? 'custom'
                            : 'exe_local',
                      }))
                    }
                    data-testid='setup-storage-mode'
                  >
                    <option value='exe_local'>exe_local</option>
                    <option value='custom'>custom</option>
                  </select>
                </label>
                <p className='text-xs text-muted-foreground'>
                  Data directory preview:{' '}
                  <span
                    className='font-medium'
                    data-testid='setup-storage-data-dir-preview'
                  >
                    {selectedStorageDataDir() || 'unknown'}
                  </span>
                </p>
                {setupForm.storageMode === 'custom' ? (
                  <label className='grid gap-1 text-sm'>
                    <span>Custom Data Directory</span>
                    <input
                      className='h-9 rounded-md border bg-background px-3'
                      value={setupForm.customDataDir}
                      onChange={(event) =>
                        setSetupForm((previous) => ({
                          ...previous,
                          customDataDir: event.target.value,
                        }))
                      }
                      data-testid='setup-storage-custom-data-dir'
                    />
                  </label>
                ) : null}
                <label className='flex items-center gap-2 text-sm'>
                  <input
                    type='checkbox'
                    checked={setupForm.portableMode}
                    onChange={(event) =>
                      setSetupForm((previous) => ({
                        ...previous,
                        portableMode: event.target.checked,
                      }))
                    }
                    data-testid='setup-storage-portable-mode'
                  />
                  Portable mode
                </label>
              </div>
            ) : null}
            {setupEntryMode === 'form' && setupStep === 2 ? (
              <div className='grid gap-3'>
                <label className='grid gap-1 text-sm'>
                  <span>Runtime Port Mode</span>
                  <select
                    className='h-9 rounded-md border bg-background px-3'
                    value={setupForm.runtimePortMode}
                    onChange={(event) =>
                      setSetupForm((previous) => ({
                        ...previous,
                        runtimePortMode:
                          event.target.value === 'fixed' ? 'fixed' : 'auto',
                      }))
                    }
                    data-testid='setup-runtime-port-mode'
                  >
                    <option value='auto'>auto</option>
                    <option value='fixed'>fixed</option>
                  </select>
                </label>
                {setupForm.runtimePortMode === 'fixed' ? (
                  <label className='grid gap-1 text-sm'>
                    <span>Fixed Port</span>
                    <input
                      className='h-9 rounded-md border bg-background px-3'
                      type='number'
                      min={1}
                      value={setupForm.runtimeFixedPort}
                      onChange={(event) =>
                        setSetupForm((previous) => ({
                          ...previous,
                          runtimeFixedPort: Number(event.target.value),
                        }))
                      }
                      data-testid='setup-runtime-fixed-port'
                    />
                  </label>
                ) : null}
                <p className='text-xs text-muted-foreground'>
                  Resolved URL preview:{' '}
                  <span
                    className='font-medium'
                    data-testid='setup-runtime-url-preview'
                  >
                    {selectedRuntimeURL()}
                  </span>
                </p>
                <p
                  className='text-xs text-muted-foreground'
                  data-testid='setup-runtime-fallback-message'
                >
                  Auto mode allows runtime port fallback when another local
                  instance is already using the default port.
                </p>
              </div>
            ) : null}
            {setupEntryMode === 'form' && setupStep === 3 ? (
              <div className='grid gap-3'>
                <label className='grid gap-1 text-sm'>
                  <span>Auth Mode</span>
                  <select
                    className='h-9 rounded-md border bg-background px-3'
                    value={setupForm.authMode}
                    onChange={(event) =>
                      setSetupForm((previous) => ({
                        ...previous,
                        authMode:
                          event.target.value === 'clerk' ? 'clerk' : 'local',
                        clerkPublishableKey:
                          event.target.value === 'clerk'
                            ? previous.clerkPublishableKey
                            : '',
                      }))
                    }
                    data-testid='setup-auth-mode'
                  >
                    <option value='local'>local</option>
                    <option value='clerk'>clerk</option>
                  </select>
                </label>
                <p
                  className='text-xs text-muted-foreground'
                  data-testid='setup-auth-readiness'
                >
                  {authReadinessLabel()}
                </p>
                {setupForm.authMode === 'clerk' ? (
                  <label className='grid gap-1 text-sm'>
                    <span>Clerk Publishable Key</span>
                    <input
                      className='h-9 rounded-md border bg-background px-3'
                      value={setupForm.clerkPublishableKey}
                      onChange={(event) =>
                        setSetupForm((previous) => ({
                          ...previous,
                          clerkPublishableKey: event.target.value,
                        }))
                      }
                      data-testid='setup-clerk-publishable-key'
                    />
                  </label>
                ) : null}
              </div>
            ) : null}
            {setupEntryMode === 'form' && setupStep === 4 ? (
              <div className='grid gap-3'>
                <p
                  className='text-xs text-muted-foreground'
                  data-testid='setup-integrations-guidance'
                >
                  Optional baseline toggles. You can edit integrations any time
                  later in Settings.
                </p>
                <label className='flex items-center gap-2 text-sm'>
                  <input
                    type='checkbox'
                    checked={setupForm.featureScanner}
                    onChange={(event) =>
                      setSetupForm((previous) => ({
                        ...previous,
                        featureScanner: event.target.checked,
                      }))
                    }
                    data-testid='setup-feature-scanner'
                  />
                  Enable Market Watch
                </label>
                <label className='flex items-center gap-2 text-sm'>
                  <input
                    type='checkbox'
                    checked={setupForm.featureChat}
                    onChange={(event) =>
                      setSetupForm((previous) => ({
                        ...previous,
                        featureChat: event.target.checked,
                      }))
                    }
                    data-testid='setup-feature-chat'
                  />
                  Enable Chat
                </label>
                <label className='flex items-center gap-2 text-sm'>
                  <input
                    type='checkbox'
                    checked={setupForm.featureProviders}
                    onChange={(event) =>
                      setSetupForm((previous) => ({
                        ...previous,
                        featureProviders: event.target.checked,
                      }))
                    }
                    data-testid='setup-feature-providers'
                  />
                  Enable Providers
                </label>
              </div>
            ) : null}
            {setupEntryMode === 'form' && setupStep === 5 ? (
              <div className='rounded-md border bg-muted/40 p-3 text-sm'>
                <p>
                  <strong>Instance:</strong>{' '}
                  {setupForm.instanceName || 'Not set'}
                </p>
                <p>
                  <strong>Profile Key:</strong>{' '}
                  {setupForm.profileKey || 'Not set'}
                </p>
                <p>
                  <strong>Storage Mode:</strong> {setupForm.storageMode}
                </p>
                <p>
                  <strong>Data Dir:</strong>{' '}
                  {selectedStorageDataDir() || 'Not set'}
                </p>
                <p>
                  <strong>Portable:</strong>{' '}
                  {setupForm.portableMode ? 'enabled' : 'disabled'}
                </p>
                <p>
                  <strong>Runtime Mode:</strong> {setupForm.runtimePortMode}
                </p>
                <p>
                  <strong>Runtime Port:</strong> {selectedRuntimePort()}
                </p>
                <p>
                  <strong>Runtime URL:</strong> {selectedRuntimeURL()}
                </p>
                <p>
                  <strong>Auth Mode:</strong> {setupForm.authMode}
                </p>
                {setupForm.authMode === 'clerk' ? (
                  <p>
                    <strong>Clerk Key:</strong>{' '}
                    {setupForm.clerkPublishableKey ? 'Configured' : 'Missing'}
                  </p>
                ) : null}
                <p>
                  <strong>Features:</strong>{' '}
                  {[
                    setupForm.featureScanner ? 'scanner' : null,
                    setupForm.featureChat ? 'chat' : null,
                    setupForm.featureProviders ? 'providers' : null,
                  ]
                    .filter(Boolean)
                    .join(', ') || 'none'}
                </p>
              </div>
            ) : null}
            {setupError ? (
              <p
                className='text-sm text-destructive'
                data-testid='setup-wizard-error'
              >
                {setupError}
              </p>
            ) : null}
            {setupEntryMode === 'form' ? (
              <div className='flex items-center gap-2 pt-2'>
                <Button
                  variant='outline'
                  onClick={goToPreviousStep}
                  disabled={setupStep === 0 || completingSetup}
                  data-testid='setup-prev'
                >
                  Previous
                </Button>
                {setupStep < totalSteps - 1 ? (
                  <Button
                    onClick={() => void goToNextStep()}
                    disabled={completingSetup}
                    data-testid='setup-next'
                  >
                    Next
                  </Button>
                ) : (
                  <Button
                    onClick={() => void completeSetup()}
                    disabled={completingSetup}
                    data-testid='setup-complete'
                  >
                    {completingSetup
                      ? 'Creating Config...'
                      : 'Create Config & Launch'}
                  </Button>
                )}
              </div>
            ) : null}
          </CardContent>
        </Card>
      </AuthLayout>
    )
  }

  return (
    <AuthLayout>
      <Card className='gap-4'>
        <CardHeader>
          <CardTitle className='text-lg tracking-tight'>Sign in</CardTitle>
          <CardDescription>
            Enter your email and password below to <br />
            log into your account
          </CardDescription>
        </CardHeader>
        <CardContent className='space-y-4'>
          <div
            className='rounded-md border bg-muted/40 p-3 text-sm text-muted-foreground'
            data-testid='sign-in-profile-guidance'
          >
            <p className='font-medium text-foreground'>
              Sign in to unlock your Cabinet workspace.
            </p>
            <p>
              Your active database/profile controls Inventory, Wishlist,
              Collections, Settings, Chats, and Integrations after sign-in;
              collections live inside that profile.
            </p>
            <p>
              New here?{' '}
              <Link
                to='/sign-up'
                className='underline underline-offset-4 hover:text-primary'
              >
                Create account
              </Link>{' '}
              first, then choose or create the profile you want to work in.
            </p>
          </div>
          <UserAuthForm redirectTo={redirect} />
        </CardContent>
        <CardFooter>
          <div className='w-full space-y-3 px-8 text-center text-sm text-muted-foreground'>
            <p>
              New to Cabinet?{' '}
              <Link
                to='/sign-up'
                className='underline underline-offset-4 hover:text-primary'
              >
                Create account
              </Link>
            </p>
            <p>
              By clicking sign in, you agree to our{' '}
              <a
                href='/terms'
                className='underline underline-offset-4 hover:text-primary'
              >
                Terms of Service
              </a>{' '}
              and{' '}
              <a
                href='/privacy'
                className='underline underline-offset-4 hover:text-primary'
              >
                Privacy Policy
              </a>
              .
            </p>
          </div>
        </CardFooter>
      </Card>
    </AuthLayout>
  )
}
