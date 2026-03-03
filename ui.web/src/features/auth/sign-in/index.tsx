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
}

type SetupFormState = {
  instanceName: string
  profileKey: string
  authMode: 'local' | 'clerk'
  clerkPublishableKey: string
}

export function SignIn() {
  const { redirect } = useSearch({ from: '/(auth)/sign-in' })
  const [setupLoading, setSetupLoading] = useState(true)
  const [setupRequired, setSetupRequired] = useState(false)
  const [setupError, setSetupError] = useState<string | null>(null)
  const [completingSetup, setCompletingSetup] = useState(false)
  const [setupForm, setSetupForm] = useState<SetupFormState>({
    instanceName: 'Primary',
    profileKey: 'primary',
    authMode: 'local',
    clerkPublishableKey: '',
  })

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
          // Ignore parse failures and keep fallback.
        }
        throw new Error(message)
      }
      setSetupRequired(false)
    } catch (error) {
      setSetupError(
        error instanceof Error ? error.message : 'setup_complete_failed'
      )
    } finally {
      setCompletingSetup(false)
    }
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
    return (
      <AuthLayout>
        <Card className='gap-4' data-testid='setup-wizard'>
          <CardHeader>
            <CardTitle className='text-lg tracking-tight'>Setup Wizard</CardTitle>
            <CardDescription>
              Complete initial runtime setup before continuing to sign in.
            </CardDescription>
          </CardHeader>
          <CardContent className='space-y-3'>
            <p className='text-sm text-muted-foreground'>
              Cabinet detected missing startup configuration. Create the setup file
              to unlock login.
            </p>
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
            {setupError ? (
              <p className='text-sm text-destructive' data-testid='setup-wizard-error'>
                {setupError}
              </p>
            ) : null}
            <Button
              onClick={() => void completeSetup()}
              disabled={completingSetup}
              data-testid='setup-wizard-complete'
            >
              {completingSetup ? 'Completing Setup...' : 'Complete Setup'}
            </Button>
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
