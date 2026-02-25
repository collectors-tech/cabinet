import { useEffect, useRef, useState } from 'react'
import { useAuth, useUser } from '@clerk/clerk-react'
import { Navigate, createFileRoute } from '@tanstack/react-router'
import { AuthenticatedLayout } from '@/components/layout/authenticated-layout'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { useAuthStore } from '@/stores/auth-store'

export const Route = createFileRoute('/clerk/_authenticated')({
  component: ClerkAuthenticatedRoute,
})

type BootstrapResponse = {
  user_id: string
  email: string
  plan: string
}

function ClerkAuthenticatedRoute() {
  const { isLoaded, isSignedIn, getToken } = useAuth()
  const { user } = useUser()
  const { auth } = useAuthStore()
  const [bootstrapping, setBootstrapping] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const bootstrapped = useRef(false)

  useEffect(() => {
    if (!isLoaded || !isSignedIn || bootstrapped.current) {
      return
    }
    bootstrapped.current = true

    const run = async () => {
      try {
        const token = await getToken()
        if (!token) {
          throw new Error('missing_clerk_token')
        }
        const response = await fetch('/api/auth/cloud/session/bootstrap', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ provider: 'clerk', token }),
        })
        if (!response.ok) {
          throw new Error(`bootstrap_failed_${response.status}`)
        }
        const payload = (await response.json()) as BootstrapResponse
        auth.setUser({
          accountNo: payload.user_id || user?.id || 'clerk-user',
          email:
            payload.email ||
            user?.primaryEmailAddress?.emailAddress ||
            'unknown@example.com',
          role: [payload.plan || 'free'],
          exp: Math.floor(Date.now() / 1000) + 60*60,
        })
        auth.setAccessToken(token)
        setError(null)
      } catch (err) {
        setError(err instanceof Error ? err.message : 'bootstrap_failed')
        auth.reset()
      } finally {
        setBootstrapping(false)
      }
    }

    void run()
  }, [auth, getToken, isLoaded, isSignedIn, user])

  if (!isLoaded) {
    return <div className='p-6 text-sm text-muted-foreground'>Loading identity...</div>
  }

  if (!isSignedIn) {
    return <Navigate to='/clerk/sign-in' />
  }

  if (bootstrapping) {
    return <div className='p-6 text-sm text-muted-foreground'>Bootstrapping session...</div>
  }

  if (error) {
    return (
      <div className='p-6'>
        <Alert>
          <AlertTitle>Session bootstrap failed</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      </div>
    )
  }

  return <AuthenticatedLayout />
}
