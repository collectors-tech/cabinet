import { useEffect, useState } from 'react'
import { z } from 'zod'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Link, useNavigate } from '@tanstack/react-router'
import { Loader2, LogIn } from 'lucide-react'
import { toast } from 'sonner'
import { IconFacebook, IconGithub } from '@/assets/brand-icons'
import { useAuthStore } from '@/stores/auth-store'
import { sleep, cn } from '@/lib/utils'
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
  providers?: ProviderOption[]
}

export function UserAuthForm({
  className,
  redirectTo,
  ...props
}: UserAuthFormProps) {
  const [isLoading, setIsLoading] = useState(false)
  const [passkeyLoading, setPasskeyLoading] = useState(false)
  const [passkeyError, setPasskeyError] = useState<string | null>(null)
  const [identityMode, setIdentityMode] = useState('local')
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

  function onSubmit(data: z.infer<typeof formSchema>) {
    setIsLoading(true)

    toast.promise(sleep(2000), {
      loading: 'Signing in...',
      success: () => {
        setIsLoading(false)

        // Mock successful authentication with expiry computed at success time
        const mockUser = {
          accountNo: 'ACC001',
          email: data.email,
          role: ['user'],
          exp: Date.now() + 24 * 60 * 60 * 1000, // 24 hours from now
        }

        // Set user and access token
        auth.setUser(mockUser)
        auth.setAccessToken('mock-access-token')

        // Redirect to the stored location or default to canonical dashboard
        const targetPath = redirectTo || '/dashboard'
        navigate({ to: targetPath, replace: true })

        return `Welcome back, ${data.email}!`
      },
      error: 'Error',
    })
  }

  async function signInWithPasskey() {
    setPasskeyError(null)
    setPasskeyLoading(true)
    try {
      const maybePublicKeyCredential = (
        window as Window & { PublicKeyCredential?: unknown }
      ).PublicKeyCredential
      if (!maybePublicKeyCredential || !navigator.credentials?.get) {
        throw new Error(
          'Passkey sign-in is unavailable on this device. Use password or provider sign-in.'
        )
      }

      const credential = await navigator.credentials.get({
        publicKey: {
          challenge: new Uint8Array([1, 2, 3, 4]),
          timeout: 60000,
          userVerification: 'preferred',
        } as PublicKeyCredentialRequestOptions,
      })

      if (!credential) {
        throw new Error(
          'Passkey sign-in failed. Use password or provider sign-in.'
        )
      }

      const mockUser = {
        accountNo: 'ACC001',
        email: 'passkey@cabinet.local',
        role: ['user'],
        exp: Date.now() + 24 * 60 * 60 * 1000,
      }

      auth.setUser(mockUser)
      auth.setAccessToken('mock-passkey-access-token')
      const targetPath = redirectTo || '/dashboard'
      navigate({ to: targetPath, replace: true })
    } catch (error) {
      setPasskeyError(
        error instanceof Error
          ? error.message
          : 'Passkey sign-in failed. Use password or provider sign-in.'
      )
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
          return
        }
        const payload = (await response.json()) as ProviderOptionsPayload
        if (cancelled) {
          return
        }
        const nextMode =
          payload.identity_mode === 'clerk' ? 'clerk' : 'local'
        setIdentityMode(nextMode)
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
        // Keep deterministic local defaults on fetch failure.
      }
    }
    void loadProviderOptions()
    return () => {
      cancelled = true
    }
  }, [])

  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(onSubmit)}
        className={cn('grid gap-3', className)}
        {...props}
      >
        <FormField
          control={form.control}
          name='email'
          render={({ field }) => (
            <FormItem>
              <FormLabel>Email</FormLabel>
              <FormControl>
                <Input placeholder='name@example.com' {...field} />
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
              <FormLabel>Password</FormLabel>
              <FormControl>
                <PasswordInput placeholder='********' toggleTestId='sign-in-password-toggle' {...field} />
              </FormControl>
              <FormMessage />
              <Link
                to='/forgot-password'
                className='absolute end-0 -top-0.5 text-sm font-medium text-muted-foreground hover:opacity-75'
              >
                Forgot password?
              </Link>
            </FormItem>
          )}
        />
        <Button className='mt-2' disabled={isLoading}>
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
          {providerOptions.map((provider) => (
            <Button
              key={provider.id}
              variant='outline'
              type='button'
              disabled={isLoading || !provider.enabled}
              data-testid={`provider-${provider.id}`}
            >
              {provider.label}
            </Button>
          ))}
        </div>

        <div className='grid grid-cols-2 gap-2'>
          <Button variant='outline' type='button' disabled={isLoading}>
            <IconGithub className='h-4 w-4' /> GitHub
          </Button>
          <Button variant='outline' type='button' disabled={isLoading}>
            <IconFacebook className='h-4 w-4' /> Facebook
          </Button>
        </div>
      </form>
    </Form>
  )
}
