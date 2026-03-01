import { z } from 'zod'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useEffect, useState } from 'react'
import { Link } from '@tanstack/react-router'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import { Switch } from '@/components/ui/switch'
import { useProfileSettings } from '../use-profile-settings'

const notificationsFormSchema = z.object({
  type: z.enum(['all', 'mentions', 'none'], {
    error: (iss) =>
      iss.input === undefined
        ? 'Please select a notification type.'
        : undefined,
  }),
  mobile: z.boolean().default(false).optional(),
  communication_emails: z.boolean().default(false).optional(),
  social_emails: z.boolean().default(false).optional(),
  marketing_emails: z.boolean().default(false).optional(),
  security_emails: z.boolean(),
})

type NotificationsFormValues = z.infer<typeof notificationsFormSchema>

const defaultValues: Partial<NotificationsFormValues> = {
  type: 'mentions',
  mobile: false,
  communication_emails: false,
  marketing_emails: false,
  social_emails: true,
  security_emails: true,
}

export function NotificationsForm() {
  const { settings, loading, error, saving, saveSettings, reload } =
    useProfileSettings()
  const [saveMessage, setSaveMessage] = useState<string | null>(null)
  const [saveError, setSaveError] = useState<string | null>(null)
  const form = useForm<NotificationsFormValues>({
    resolver: zodResolver(notificationsFormSchema),
    defaultValues,
  })

  useEffect(() => {
    if (loading) {
      return
    }
    const storedType = settings['notifications.type']
    form.reset({
      type:
        storedType === 'all' || storedType === 'mentions' || storedType === 'none'
          ? storedType
          : 'mentions',
      mobile: settings['notifications.mobile'] === 'true',
      communication_emails:
        settings['notifications.communication_emails'] === 'true',
      social_emails: settings['notifications.social_emails'] === 'true',
      marketing_emails: settings['notifications.marketing_emails'] === 'true',
      security_emails: settings['notifications.security_emails'] !== 'false',
    })
  }, [form, loading, settings])

  const handleSubmit = async (data: NotificationsFormValues) => {
    setSaveMessage(null)
    setSaveError(null)
    try {
      await saveSettings({
        'notifications.type': data.type,
        'notifications.mobile': String(data.mobile ?? false),
        'notifications.communication_emails': String(
          data.communication_emails ?? false
        ),
        'notifications.social_emails': String(data.social_emails ?? false),
        'notifications.marketing_emails': String(data.marketing_emails ?? false),
        'notifications.security_emails': String(data.security_emails),
      })
      setSaveMessage('Notification settings saved.')
    } catch (err) {
      setSaveError(
        err instanceof Error ? err.message : 'failed_to_save_notifications'
      )
    }
  }

  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(handleSubmit)}
        className='space-y-8'
      >
        {error ? (
          <div className='rounded-md border border-destructive/40 bg-destructive/10 p-3 text-sm'>
            <p className='font-medium'>Failed to load notification settings.</p>
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
        <FormField
          control={form.control}
          name='type'
          render={({ field }) => (
            <FormItem className='relative space-y-3'>
              <FormLabel>Notify me about...</FormLabel>
              <FormControl>
                <RadioGroup
                  onValueChange={field.onChange}
                  value={field.value}
                  className='flex flex-col gap-2'
                >
                  <FormItem className='flex items-center'>
                    <FormControl>
                      <RadioGroupItem value='all' />
                    </FormControl>
                    <FormLabel className='font-normal'>
                      All new messages
                    </FormLabel>
                  </FormItem>
                  <FormItem className='flex items-center'>
                    <FormControl>
                      <RadioGroupItem value='mentions' />
                    </FormControl>
                    <FormLabel className='font-normal'>
                      Direct messages and mentions
                    </FormLabel>
                  </FormItem>
                  <FormItem className='flex items-center'>
                    <FormControl>
                      <RadioGroupItem value='none' />
                    </FormControl>
                    <FormLabel className='font-normal'>Nothing</FormLabel>
                  </FormItem>
                </RadioGroup>
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <div className='relative'>
          <h3 className='mb-4 text-lg font-medium'>Email Notifications</h3>
          <div className='space-y-4'>
            <FormField
              control={form.control}
              name='communication_emails'
              render={({ field }) => (
                <FormItem className='flex flex-row items-center justify-between rounded-lg border p-4'>
                  <div className='space-y-0.5'>
                    <FormLabel className='text-base'>
                      Communication emails
                    </FormLabel>
                    <FormDescription>
                      Receive emails about your account activity.
                    </FormDescription>
                  </div>
                  <FormControl>
                    <Switch
                      data-testid='settings-notifications-communication'
                      checked={field.value}
                      onCheckedChange={field.onChange}
                    />
                  </FormControl>
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='marketing_emails'
              render={({ field }) => (
                <FormItem className='flex flex-row items-center justify-between rounded-lg border p-4'>
                  <div className='space-y-0.5'>
                    <FormLabel className='text-base'>
                      Marketing emails
                    </FormLabel>
                    <FormDescription>
                      Receive emails about new products, features, and more.
                    </FormDescription>
                  </div>
                  <FormControl>
                    <Switch
                      data-testid='settings-notifications-marketing'
                      checked={field.value}
                      onCheckedChange={field.onChange}
                    />
                  </FormControl>
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='social_emails'
              render={({ field }) => (
                <FormItem className='flex flex-row items-center justify-between rounded-lg border p-4'>
                  <div className='space-y-0.5'>
                    <FormLabel className='text-base'>Social emails</FormLabel>
                    <FormDescription>
                      Receive emails for friend requests, follows, and more.
                    </FormDescription>
                  </div>
                  <FormControl>
                    <Switch
                      data-testid='settings-notifications-social'
                      checked={field.value}
                      onCheckedChange={field.onChange}
                    />
                  </FormControl>
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='security_emails'
              render={({ field }) => (
                <FormItem className='flex flex-row items-center justify-between rounded-lg border p-4'>
                  <div className='space-y-0.5'>
                    <FormLabel className='text-base'>Security emails</FormLabel>
                    <FormDescription>
                      Receive emails about your account activity and security.
                    </FormDescription>
                  </div>
                  <FormControl>
                    <Switch
                      data-testid='settings-notifications-security'
                      checked={field.value}
                      onCheckedChange={field.onChange}
                      disabled
                      aria-readonly
                    />
                  </FormControl>
                </FormItem>
              )}
            />
          </div>
        </div>
        <FormField
          control={form.control}
          name='mobile'
          render={({ field }) => (
            <FormItem className='relative flex flex-row items-start'>
              <FormControl>
                <Checkbox
                  checked={field.value}
                  onCheckedChange={field.onChange}
                />
              </FormControl>
              <div className='space-y-1 leading-none'>
                <FormLabel>
                  Use different settings for my mobile devices
                </FormLabel>
                <FormDescription>
                  You can manage your mobile notifications in the{' '}
                  <Link
                    to='/settings'
                    className='underline decoration-dashed underline-offset-4 hover:decoration-solid'
                  >
                    mobile settings
                  </Link>{' '}
                  page.
                </FormDescription>
              </div>
            </FormItem>
          )}
        />
        <Button type='submit' disabled={saving || loading}>
          Update notifications
        </Button>
      </form>
    </Form>
  )
}
