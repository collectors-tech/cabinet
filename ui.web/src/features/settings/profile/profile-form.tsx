import { z } from 'zod'
import { useFieldArray, useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { type FormEvent, useEffect, useRef, useState } from 'react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import { ProfileContextBlocked } from '../components/profile-context-blocked'
import { useProfileSettings } from '../use-profile-settings'

const profileFormSchema = z.object({
  username: z
    .string('Please enter your username.')
    .min(2, 'Username must be at least 2 characters.')
    .max(30, 'Username must not be longer than 30 characters.'),
  email: z.email({
    error: (iss) =>
      iss.input === undefined
        ? 'Please select an email to display.'
        : undefined,
  }),
  bio: z.string().max(160).min(4),
  urls: z
    .array(
      z.object({
        value: z.url('Please enter a valid URL.'),
      })
    )
    .optional(),
})

type ProfileFormValues = z.infer<typeof profileFormSchema>

const defaultValues: Partial<ProfileFormValues> = {
  username: '',
  email: '',
  bio: 'I own a computer.',
  urls: [],
}

export function ProfileForm() {
  const {
    settings,
    loading,
    error,
    profileContextMissing,
    saving,
    saveSettings,
    reload,
  } =
    useProfileSettings()
  const [saveMessage, setSaveMessage] = useState<string | null>(null)
  const [saveError, setSaveError] = useState<string | null>(null)
  const [submitLocked, setSubmitLocked] = useState(false)
  const submitLockRef = useRef(false)
  const mutationLockRef = useRef(false)
  const form = useForm<ProfileFormValues>({
    resolver: zodResolver(profileFormSchema),
    defaultValues,
    mode: 'onChange',
  })

  const { fields, append } = useFieldArray({
    name: 'urls',
    control: form.control,
  })

  useEffect(() => {
    if (loading) {
      return
    }
    const storedURLs = settings['profile.urls']
    let urls: Array<{ value: string }> = []
    if (storedURLs) {
      try {
        const parsed = JSON.parse(storedURLs) as Array<{ value?: string }>
        urls = parsed
          .map((entry) => ({ value: entry.value?.trim() ?? '' }))
          .filter((entry) => entry.value !== '')
      } catch {
        urls = []
      }
    }
    form.reset({
      username: settings['profile.username'] ?? '',
      email: settings['profile.email'] ?? 'm@example.com',
      bio: settings['profile.bio'] ?? 'I own a computer.',
      urls,
    })
  }, [form, loading, settings])

  const handleSubmit = async (data: ProfileFormValues) => {
    if (mutationLockRef.current) {
      return
    }
    mutationLockRef.current = true
    setSaveMessage(null)
    setSaveError(null)
    try {
      await saveSettings({
        'profile.username': data.username.trim(),
        'profile.email': data.email.trim(),
        'profile.bio': data.bio.trim(),
        'profile.urls': JSON.stringify(data.urls ?? []),
      })
      setSaveMessage('Profile settings saved.')
    } catch (err) {
      setSaveError(err instanceof Error ? err.message : 'failed_to_save_profile')
    } finally {
      mutationLockRef.current = false
    }
  }

  const handleFormSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    if (submitLockRef.current) {
      return
    }
    submitLockRef.current = true
    setSubmitLocked(true)
    void (async () => {
      const isValid = await form.trigger()
      if (!isValid) {
        return
      }
      const values = form.getValues()
      await handleSubmit(values)
    })().finally(() => {
      submitLockRef.current = false
      setSubmitLocked(false)
    })
  }

  return (
    <Form {...form}>
      <form
        onSubmit={handleFormSubmit}
        className='space-y-8'
      >
        {error ? (
          profileContextMissing ? (
            <ProfileContextBlocked error={error} onRetry={() => void reload()} />
          ) : (
            <div className='rounded-md border border-destructive/40 bg-destructive/10 p-3 text-sm'>
              <p className='font-medium'>Failed to load profile settings.</p>
              <p className='mt-1 text-muted-foreground'>{error}</p>
              <Button
                type='button'
                variant='outline'
                size='sm'
                className='mt-3'
                onClick={() => void reload()}
              >
                Retry
              </Button>
            </div>
          )
        ) : null}
        {saveError ? (
          <div className='rounded-md border border-destructive/40 bg-destructive/10 p-3 text-sm text-destructive'>
            {saveError}
          </div>
        ) : null}
        {saveMessage ? (
          <div className='rounded-md border border-emerald-500/30 bg-emerald-500/10 p-3 text-sm text-emerald-700 dark:text-emerald-300'>
            {saveMessage}
          </div>
        ) : null}
        {profileContextMissing ? null : (
          <>
            <FormField
          control={form.control}
          name='username'
          render={({ field }) => (
            <FormItem>
              <FormLabel>Username</FormLabel>
              <FormControl>
                <Input placeholder='cabinet-user' {...field} />
              </FormControl>
              <FormDescription>
                This is your public display name. It can be your real name or a
                pseudonym. You can only change this once every 30 days.
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name='email'
          render={({ field }) => (
            <FormItem>
              <FormLabel>Email</FormLabel>
              <Select onValueChange={field.onChange} value={field.value}>
                <FormControl>
                  <SelectTrigger data-testid='settings-profile-email-trigger'>
                    <SelectValue
                      placeholder='Select a verified email to display'
                    />
                  </SelectTrigger>
                </FormControl>
                <SelectContent>
                  <SelectItem value='m@example.com'>m@example.com</SelectItem>
                  <SelectItem value='m@google.com'>m@google.com</SelectItem>
                  <SelectItem value='m@support.com'>m@support.com</SelectItem>
                </SelectContent>
              </Select>
              <FormDescription>
                This email is shown in your profile and account surfaces.
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name='bio'
          render={({ field }) => (
            <FormItem>
              <FormLabel>Bio</FormLabel>
              <FormControl>
                <Textarea
                  placeholder='Tell us a little bit about yourself'
                  className='resize-none'
                  {...field}
                />
              </FormControl>
              <FormDescription>
                You can <span>@mention</span> other users and organizations to
                link to them.
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />
        <div>
          {fields.map((field, index) => (
            <FormField
              control={form.control}
              key={field.id}
              name={`urls.${index}.value`}
              render={({ field }) => (
                <FormItem>
                  <FormLabel className={cn(index !== 0 && 'sr-only')}>
                    URLs
                  </FormLabel>
                  <FormDescription className={cn(index !== 0 && 'sr-only')}>
                    Add links to your website, blog, or social media profiles.
                  </FormDescription>
                  <FormControl className={cn(index !== 0 && 'mt-1.5')}>
                    <Input {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          ))}
          <Button
            type='button'
            variant='outline'
            size='sm'
            className='mt-2'
            onClick={() => append({ value: '' })}
          >
            Add URL
          </Button>
        </div>
            <Button
              type='submit'
              disabled={saving || loading || submitLocked}
              onClickCapture={() => {
                if (!submitLockRef.current && !submitLocked) {
                  setSubmitLocked(true)
                }
              }}
            >
              Update profile
            </Button>
          </>
        )}
      </form>
    </Form>
  )
}
