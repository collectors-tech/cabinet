import { useEffect, useState } from 'react'
import { Link } from '@tanstack/react-router'
import { useSearch } from '@tanstack/react-router'
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { AuthLayout } from '../auth-layout'
import { UserAuthForm } from './components/user-auth-form'

type RuntimeSetupStatus = {
  setup_required?: boolean
  config_path?: string
}

type RuntimeSetupCompletePayload = {
  ok?: boolean
  config_path?: string
  data_dir?: string
  media_dir?: string
  runtime_url?: string
  runtime_port?: number
}

type SetupFormState = {
  instanceName: string
  profileKey: string
  authMode: 'local' | 'clerk'
  clerkPublishableKey: string
}

type SetupEntryMode = 'welcome' | 'form' | 'import'

export function SignIn() {
  const { redirect } = useSearch({ from: '/(auth)/sign-in' })
  const [setupLoading, setSetupLoading] = useState(true)
  const [setupRequired, setSetupRequired] = useState(false)
  const [setupError, setSetupError] = useState<string | null>(null)
  const [setupConfigPath, setSetupConfigPath] = useState('')
  const [completingSetup, setCompletingSetup] = useState(false)
  const [setupStep, setSetupStep] = useState(0)
  const [setupCompleteState, setSetupCompleteState] =
    useState<RuntimeSetupCompletePayload | null>(null)
  const [setupEntryMode, setSetupEntryMode] = useState<SetupEntryMode>('welcome')
  const [setupImportSourcePath, setSetupImportSourcePath] = useState('')
  const [setupForm, setSetupForm] = useState<SetupFormState>({
    instanceName: 'Primary',
    profileKey: 'primary',
    authMode: 'local',
    clerkPublishableKey: '',
  })

  const totalSteps = 3
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

  function goToNextStep() {
    if (setupStep === 0 && setupForm.instanceName.trim() === '') {
      setSetupError('Instance name is required.')
      return
    }
    setSetupError(null)
    setSetupStep((previous) => Math.min(previous + 1, totalSteps - 1))
  }

  function goToPreviousStep() {
    setSetupError(null)
    setSetupStep((previous) => Math.max(previous - 1, 0))
  }

  async function completeSetup() {
    if (setupForm.authMode === 'clerk' && setupForm.clerkPublishableKey.trim() === '') {
      setSetupError('Clerk publishable key is required.')
      return
    }
    setCompletingSetup(true)
    setSetupError(null)
    try {
      const response = await fetch('/api/runtime/setup-complete', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          instance_name: setupForm.instanceName.trim(),
          profile_key: setupForm.profileKey.trim(),
          auth_mode: setupForm.authMode,
          clerk_publishable_key:
            setupForm.authMode === 'clerk'
              ? setupForm.clerkPublishableKey.trim()
              : '',
          runtime_port_mode: 'auto',
          bootstrap_workspace: 'Local Workspace',
          bootstrap_database_ref: 'Primary DB',
        }),
      })
      if (!response.ok) {
        let message = `setup_complete_failed_${response.status}`
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
      const payload = (await response.json()) as RuntimeSetupCompletePayload
      setSetupCompleteState(payload)
    } catch (error) {
      setSetupError(
        error instanceof Error ? error.message : 'setup_complete_failed'
      )
    } finally {
      setCompletingSetup(false)
    }
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
      setSetupError(error instanceof Error ? error.message : 'setup_import_failed')
    } finally {
      setCompletingSetup(false)
    }
  }

  function startAppFromSetup() {
    setSetupCompleteState(null)
    setSetupRequired(false)
    setSetupStep(0)
  }

  if (setupLoading) {
    return (
      <AuthLayout>
        <Card className='gap-4'>
          <CardHeader>
            <CardTitle className='text-lg tracking-tight'>Setup Wizard</CardTitle>
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
              <CardTitle className='text-lg tracking-tight'>Config complete</CardTitle>
              <CardDescription>
                Setup is complete. Start Cabinet with your configuration.
              </CardDescription>
            </CardHeader>
            <CardContent className='space-y-3'>
              <p className='text-sm text-muted-foreground'>
                Config path:{' '}
                <span className='font-medium' data-testid='setup-complete-config-path'>
                  {setupCompleteState.config_path ?? 'unknown'}
                </span>
              </p>
              <p className='text-sm text-muted-foreground'>
                Data directory:{' '}
                <span className='font-medium' data-testid='setup-complete-data-dir'>
                  {setupCompleteState.data_dir ?? 'unknown'}
                </span>
              </p>
              <p className='text-sm text-muted-foreground'>
                Media directory:{' '}
                <span className='font-medium' data-testid='setup-complete-media-dir'>
                  {setupCompleteState.media_dir ?? 'unknown'}
                </span>
              </p>
              <p className='text-sm text-muted-foreground'>
                Runtime URL:{' '}
                <span className='font-medium' data-testid='setup-complete-runtime-url'>
                  {setupCompleteState.runtime_url ?? 'unknown'}
                </span>
              </p>
              <p className='text-sm text-muted-foreground'>
                Runtime port:{' '}
                <span className='font-medium' data-testid='setup-complete-runtime-port'>
                  {setupCompleteState.runtime_port ?? 0}
                </span>
              </p>
              <Button data-testid='setup-start-app' onClick={startAppFromSetup}>
                Start App
              </Button>
            </CardContent>
          </Card>
        </AuthLayout>
      )
    }

    return (
      <AuthLayout>
        <Card className='gap-4' data-testid='setup-wizard'>
          <CardHeader>
            <CardTitle className='text-lg tracking-tight'>Setup Wizard</CardTitle>
            <CardDescription>
              Complete initial runtime setup before continuing to sign in.
            </CardDescription>
            {setupEntryMode === 'form' ? (
              <>
                <p
                  className='text-xs font-semibold uppercase tracking-wide text-muted-foreground'
                  data-testid='setup-step-indicator'
                >
                  STEP {setupStep + 1} OF {totalSteps}
                </p>
                <div className='flex items-center justify-between gap-2'>
                  <p className='text-sm text-muted-foreground'>Setup Progress</p>
                  <p className='text-sm font-medium' data-testid='setup-step-percent'>
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
              Cabinet detected missing startup configuration. Create the setup file
              to unlock login.
            </p>
            {setupEntryMode === 'welcome' ? (
              <div className='flex flex-wrap items-center gap-2'>
                <Button data-testid='setup-start' onClick={startSetupFlow}>
                  Start Setup
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
                    onChange={(event) => setSetupImportSourcePath(event.target.value)}
                    data-testid='setup-import-source-path'
                  />
                </label>
                <div className='flex flex-wrap items-center gap-2'>
                  <Button
                    data-testid='setup-import-submit'
                    disabled={completingSetup}
                    onClick={() => void importExistingSetupConfig()}
                  >
                    {completingSetup ? 'Importing...' : 'Import Existing Config'}
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
                  <span className='font-medium' data-testid='setup-config-path-preview'>
                    {setupConfigPath || 'unknown'}
                  </span>
                </p>
              </div>
            ) : null}
            {setupEntryMode === 'form' && setupStep === 1 ? (
              <div className='grid gap-3'>
                <label className='grid gap-1 text-sm'>
                  <span>Auth Mode</span>
                  <select
                    className='h-9 rounded-md border bg-background px-3'
                    value={setupForm.authMode}
                    onChange={(event) =>
                      setSetupForm((previous) => ({
                        ...previous,
                        authMode: event.target.value === 'clerk' ? 'clerk' : 'local',
                      }))
                    }
                    data-testid='setup-auth-mode'
                  >
                    <option value='local'>local</option>
                    <option value='clerk'>clerk</option>
                  </select>
                </label>
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
            {setupEntryMode === 'form' && setupStep === 2 ? (
              <div className='rounded-md border bg-muted/40 p-3 text-sm'>
                <p>
                  <strong>Instance:</strong> {setupForm.instanceName || 'Not set'}
                </p>
                <p>
                  <strong>Profile Key:</strong> {setupForm.profileKey || 'Not set'}
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
              </div>
            ) : null}
            {setupError ? (
              <p className='text-sm text-destructive' data-testid='setup-wizard-error'>
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
                    onClick={goToNextStep}
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
                    {completingSetup ? 'Completing Setup...' : 'Complete'}
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
        <CardContent>
          <UserAuthForm redirectTo={redirect} />
        </CardContent>
        <CardFooter>
          <div className='w-full space-y-3 px-8 text-center text-sm text-muted-foreground'>
            <p>
              New to Cabinet?{' '}
              <Link to='/sign-up' className='underline underline-offset-4 hover:text-primary'>
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
