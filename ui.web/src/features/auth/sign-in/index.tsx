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

export function SignIn() {
  const { redirect } = useSearch({ from: '/(auth)/sign-in' })
  const [setupLoading, setSetupLoading] = useState(true)
  const [setupRequired, setSetupRequired] = useState(false)
  const [setupError, setSetupError] = useState<string | null>(null)
  const [completingSetup, setCompletingSetup] = useState(false)

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
    setCompletingSetup(true)
    setSetupError(null)
    try {
      const response = await fetch('/api/runtime/setup-complete', {
        method: 'POST',
      })
      if (!response.ok) {
        throw new Error(`setup_complete_failed_${response.status}`)
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
