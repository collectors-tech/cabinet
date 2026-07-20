import { type ComponentType, type SVGProps, useEffect, useState } from 'react'
import { z } from 'zod'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Link, useNavigate } from '@tanstack/react-router'
import { Loader2, LogIn, MonitorCheck, ShieldCheck } from 'lucide-react'
import { toast } from 'sonner'
import {
  IconApple,
  IconFacebook,
  IconGithub,
  IconGoogle,
  IconMicrosoft,
} from '@/assets/brand-icons'
import { useAuthStore } from '@/stores/auth-store'
import { recordNotificationHistory } from '@/lib/toast-history'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { PasswordInput } from '@/components/password-input'

const formSchema = z.object({
  email: z.email({
    error: (iss) => (iss.input === '' ? 'Please enter your email' : undefined),
  }),
  password: z
    .string()
    .min(1, 'Please enter your password')
    .min(7, 'Password must be at least 7 characters long'),
})

interface UserAuthFormProps extends React.HTMLAttributes<HTMLFormElement> {
  redirectTo?: string
}

type ProviderOption = {
  id: string
  label: string
  enabled: boolean
}

type ProviderOptionsPayload = {
  identity_mode?: string
  zitadel_configured?: boolean
  zitadel_login_path?: string
  providers?: ProviderOption[]
}

const providerIcons: Record<string, ComponentType<SVGProps<SVGSVGElement>>> = {
  google: IconGoogle,
  apple: IconApple,
  microsoft: IconMicrosoft,
}

const LOCAL_DEVICE_SESSION_TOKEN = 'cabinet-local-device-session-v1'

function normalizePasskeyError(error: unknown) {
  const fallback = 'Passkey sign-in failed. Use password or provider sign-in.'
  if (!(error instanceof Error)) {
    return fallback
  }

  const message = error.message.trim()
  const lower = message.toLowerCase()
  if (
    lower.includes('invalid domain') ||
    lower.includes('invalid rp id') ||
    lower.includes('relying party') ||
    lower.includes('origin mismatch') ||
    lower.includes('domain mismatch')
  ) {
    return 'Passkey sign-in is not available on this domain yet. Use password or provider sign-in.'
  }

  return message || fallback
}

function authToastHistory(id: string, title: string, summary?: string) {
  return {
    history: {
      id,
      title,
      summary,
      source_label: 'Auth sign-in',
      category: 'auth',
    },
  } as Record<string, unknown>
}

export function UserAuthForm({
  className,
  redirectTo,
  ...props
}: UserAuthFormProps) {
  const [isLoading, setIsLoading] = useState(false)
  const [passkeyLoading, setPasskeyLoading] = useState(false)
  const [passkeyError, setPasskeyError] = useState<string | null>(null)
  const [identityMode, setIdentityMode] = useState('loading')
  const [zitadelConfigured, setZitadelConfigured] = useState(false)
  const [zitadelLoginPath, setZitadelLoginPath] = useState(
    '/api/auth/zitadel/login'
  )
  const [providerOptions, setProviderOptions] = useState<ProviderOption[]>([
    { id: 'google', label: 'Google', enabled: true },
    { id: 'apple', label: 'Apple', enabled: true },
    { id: 'microsoft', label: 'Microsoft', enabled: true },
  ])
  const navigate = useNavigate()
  const { auth } = useAuthStore()

  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      email: '',
      password: '',
    },
  })

  function openLocalWorkspace() {
    setIsLoading(true)
    const localUser = {
      accountNo: 'LOCAL001',
      email: 'local-device@cabinet.local',
      role: ['local-device'],
      exp: Date.now() + 24 * 60 * 60 * 1000,
    }

    auth.setUser(localUser)
    auth.setAccessToken(LOCAL_DEVICE_SESSION_TOKEN)
    const targetPath = redirectTo || '/dashboard'
    navigate({ to: targetPath, replace: true })
    setIsLoading(false)
  }

  function onSubmit(data: z.infer<typeof formSchema>) {
    if (identityMode === 'local') {
      openLocalWorkspace()
      return
    }

    setIsLoading(true)
    const message =
      'Cloud sign-in is unavailable until Cabinet can verify the configured identity provider.'
    toast.error(message, {
      ...authToastHistory(
        'auth-sign-in-cloud-unavailable',
        'Cloud sign-in unavailable',
        `Cloud sign-in was blocked for ${data.email}`
      ),
    })
    recordNotificationHistory({
      id: 'auth-sign-in-cloud-unavailable',
      level: 'warning',
      title: 'Cloud sign-in unavailable',
      summary: message,
      source_label: 'Auth sign-in',
      category: 'auth',
    })
    setIsLoading(false)
  }

  function startZitadelLogin() {
    setIsLoading(true)
    const query = new URLSearchParams()
    if (redirectTo) {
      query.set('return_to', redirectTo)
    }
    const target = query.size > 0 ? `${zitadelLoginPath}?${query}` : zitadelLoginPath
    window.location.assign(target)
  }

  async function signInWithPasskey() {
    setPasskeyError(null)
    setPasskeyLoading(true)
    try {
      throw new Error(
        'Passkey sign-in is not enabled for the local beta. Use local-device mode or a configured cloud identity provider.'
      )
    } catch (error) {
      const message = normalizePasskeyError(error)
      setPasskeyError(message)
      recordNotificationHistory({
        id: 'auth-passkey-sign-in-failed',
        level: 'warning',
        title: 'Passkey sign-in failed',
        summary: message,
        source_label: 'Auth sign-in',
        category: 'auth',
      })
    } finally {
      setPasskeyLoading(false)
    }
  }

  useEffect(() => {
    let cancelled = false
    async function loadProviderOptions() {
      try {
        const response = await fetch('/api/auth/provider-options')
        if (!response.ok) {
          if (!cancelled) {
            setIdentityMode('unavailable')
          }
          return
        }
        const payload = (await response.json()) as ProviderOptionsPayload
        if (cancelled) {
          return
        }
        const nextMode =
          payload.identity_mode === 'zitadel'
            ? 'zitadel'
            : payload.identity_mode === 'clerk'
              ? 'clerk'
              : 'local'
        setIdentityMode(nextMode)
        setZitadelConfigured(Boolean(payload.zitadel_configured))
        if (payload.zitadel_login_path?.startsWith('/')) {
          setZitadelLoginPath(payload.zitadel_login_path)
        }
        if (Array.isArray(payload.providers) && payload.providers.length > 0) {
          setProviderOptions(
            payload.providers.map((provider) => ({
              id: String(provider.id || '').toLowerCase(),
              label: String(provider.label || provider.id || ''),
              enabled: Boolean(provider.enabled),
            }))
          )
        }
      } catch {
        if (!cancelled) {
          setIdentityMode('unavailable')
        }
      }
    }
    void loadProviderOptions()
    return () => {
      cancelled = true
    }
  }, [])

  if (identityMode === 'loading' || identityMode === 'unavailable') {
    return (
      <div className={cn('grid gap-3', className)}>
        <div
          className='rounded-md border bg-muted/40 p-4 text-sm'
          role={identityMode === 'unavailable' ? 'alert' : 'status'}
        >
          <div className='flex items-center gap-2 font-medium'>
            {identityMode === 'loading' ? (
              <Loader2 className='h-4 w-4 animate-spin' aria-hidden='true' />
            ) : (
              <ShieldCheck className='h-4 w-4' aria-hidden='true' />
            )}
            {identityMode === 'loading'
              ? 'Checking identity configuration'
              : 'Identity configuration unavailable'}
          </div>
          <p className='mt-2 text-muted-foreground'>
            {identityMode === 'loading'
              ? 'Cabinet is verifying this environment before offering a sign-in method.'
              : 'Cabinet could not verify a safe sign-in method. Refresh after the runtime is available.'}
          </p>
        </div>
      </div>
    )
  }

  if (identityMode === 'zitadel') {
    return (
      <div className={cn('grid gap-3', className)}>
        <div
          className='rounded-md border bg-muted/40 p-4 text-sm'
          data-testid='zitadel-auth-boundary'
        >
          <div className='flex items-center gap-2 font-medium'>
            <ShieldCheck className='h-4 w-4' aria-hidden='true' />
            Cabinet secure account
          </div>
          <p className='mt-2 text-muted-foreground'>
            Continue to the Cabinet-branded identity service. Cabinet never
            stores your password or provider tokens in this browser; it
            receives only an opaque secure session cookie.
          </p>
        </div>
        <Button
          className='mt-2'
          type='button'
          disabled={isLoading || !zitadelConfigured}
          data-testid='zitadel-sign-in'
          onClick={startZitadelLogin}
        >
          {isLoading ? <Loader2 className='animate-spin' /> : <ShieldCheck />}
          Continue securely
        </Button>
        {!zitadelConfigured ? (
          <p className='text-sm text-destructive' role='alert'>
            Secure account sign-in is not configured for this environment.
          </p>
        ) : (
          <p className='text-xs text-muted-foreground'>
            Account creation and recovery are available in the next secure
            step.
          </p>
        )}
        <p
          className='text-xs text-muted-foreground'
          data-testid='identity-mode-indicator'
        >
          Identity mode: zitadel
        </p>
      </div>
    )
  }

  if (identityMode === 'local') {
    return (
      <div className={cn('grid gap-3', className)}>
        <div
          className='rounded-md border bg-muted/40 p-3 text-sm'
          data-testid='local-device-auth-boundary'
        >
          <p className='font-medium'>Local device mode</p>
          <p className='text-muted-foreground'>
            Opens this device's local Cabinet workspace. This does not verify a
            password, passkey, cloud account, or encrypted-at-rest lock.
          </p>
        </div>
        <Button
          className='mt-2'
          type='button'
          disabled={isLoading}
          data-testid='open-local-workspace'
          onClick={openLocalWorkspace}
        >
          {isLoading ? <Loader2 className='animate-spin' /> : <MonitorCheck />}
          Open local workspace
        </Button>
        <Button
          className='mt-1'
          variant='outline'
          type='button'
          data-testid='passkey-signin'
          disabled={passkeyLoading}
          onClick={() => void signInWithPasskey()}
        >
          {passkeyLoading ? <Loader2 className='animate-spin' /> : null}
          Passkey unavailable
        </Button>
        {passkeyError ? (
          <p className='text-sm text-destructive' data-testid='passkey-error'>
            {passkeyError}
          </p>
        ) : null}
        <Link
          to='/forgot-password'
          data-testid='sign-in-forgot-password-link'
          className='text-sm font-medium text-muted-foreground underline-offset-4 hover:underline'
        >
          Forgot password?
        </Link>
        <p
          className='text-xs text-muted-foreground'
          data-testid='identity-mode-indicator'
        >
          Identity mode: local-device
        </p>
      </div>
    )
  }

  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(onSubmit)}
        noValidate
        className={cn('grid gap-3', className)}
        {...props}
      >
        <FormField
          control={form.control}
          name='email'
          render={({ field }) => (
            <FormItem>
              <FormLabel htmlFor='email'>Email</FormLabel>
              <FormControl>
                <Input
                  id='email'
                  type='email'
                  autoComplete='username'
                  placeholder='name@example.com'
                  {...field}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name='password'
          render={({ field }) => (
            <FormItem className='relative'>
              <FormLabel htmlFor='password'>Password</FormLabel>
              <FormControl>
                <PasswordInput
                  id='password'
                  autoComplete='current-password'
                  placeholder='********'
                  toggleTestId='sign-in-password-toggle'
                  {...field}
                />
              </FormControl>
              <FormMessage />
              <Link
                to='/forgot-password'
                data-testid='sign-in-forgot-password-link'
                className='absolute end-0 -top-0.5 text-sm font-medium text-muted-foreground hover:opacity-75'
              >
                Forgot password?
              </Link>
            </FormItem>
          )}
        />
        <Button className='mt-2' type='submit' disabled={isLoading}>
          {isLoading ? <Loader2 className='animate-spin' /> : <LogIn />}
          Sign in
        </Button>

        <Button
          className='mt-1'
          variant='outline'
          type='button'
          data-testid='passkey-signin'
          disabled={passkeyLoading}
          onClick={() => void signInWithPasskey()}
        >
          {passkeyLoading ? <Loader2 className='animate-spin' /> : null}
          Sign in with Passkey
        </Button>
        {passkeyError ? (
          <p className='text-sm text-destructive' data-testid='passkey-error'>
            {passkeyError}
          </p>
        ) : null}

        <div className='relative my-2'>
          <div className='absolute inset-0 flex items-center'>
            <span className='w-full border-t' />
          </div>
          <div className='relative flex justify-center text-xs uppercase'>
            <span className='bg-background px-2 text-muted-foreground'>
              Or continue with
            </span>
          </div>
        </div>

        <p
          className='text-xs text-muted-foreground'
          data-testid='identity-mode-indicator'
        >
          Identity mode: {identityMode}
        </p>

        <div className='grid grid-cols-3 gap-2'>
          {providerOptions.map((provider) => {
            const ProviderIcon = providerIcons[provider.id]
            return (
              <Button
                key={provider.id}
                variant='outline'
                type='button'
                disabled={isLoading || !provider.enabled}
                data-testid={`provider-${provider.id}`}
              >
                {ProviderIcon ? (
                  <ProviderIcon
                    aria-hidden='true'
                    data-testid={`provider-${provider.id}-icon`}
                    className='h-4 w-4 shrink-0'
                  />
                ) : null}
                <span data-testid={`provider-${provider.id}-label`}>
                  {provider.label}
                </span>
              </Button>
            )
          })}
        </div>

        <div className='grid grid-cols-2 gap-2'>
          <Button
            variant='outline'
            type='button'
            disabled
            data-testid='sign-in-provider-github'
          >
            <IconGithub className='h-4 w-4' /> GitHub
          </Button>
          <Button
            variant='outline'
            type='button'
            disabled
            data-testid='sign-in-provider-facebook'
          >
            <IconFacebook className='h-4 w-4' /> Facebook
          </Button>
        </div>
      </form>
    </Form>
  )
}
