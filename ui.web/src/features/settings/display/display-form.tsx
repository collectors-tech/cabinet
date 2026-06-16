import { useEffect, useState } from 'react'
import { z } from 'zod'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
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
import { ProfileContextBlocked } from '../components/profile-context-blocked'
import { useProfileSettings } from '../use-profile-settings'

const items = [
  {
    id: 'recents',
    label: 'Recents',
  },
  {
    id: 'home',
    label: 'Home',
  },
  {
    id: 'applications',
    label: 'Applications',
  },
  {
    id: 'desktop',
    label: 'Desktop',
  },
  {
    id: 'downloads',
    label: 'Downloads',
  },
  {
    id: 'documents',
    label: 'Documents',
  },
] as const

const displayFormSchema = z.object({
  items: z.array(z.string()).refine((value) => value.some((item) => item), {
    message: 'You have to select at least one item.',
  }),
})

type DisplayFormValues = z.infer<typeof displayFormSchema>

const defaultValues: Partial<DisplayFormValues> = {
  items: ['recents', 'home'],
}

export function DisplayForm() {
  const {
    settings,
    loading,
    error,
    profileContextMissing,
    saving,
    saveSettings,
    reload,
  } = useProfileSettings()
  const [saveMessage, setSaveMessage] = useState<string | null>(null)
  const [saveError, setSaveError] = useState<string | null>(null)
  const form = useForm<DisplayFormValues>({
    resolver: zodResolver(displayFormSchema),
    defaultValues,
  })

  useEffect(() => {
    if (loading) {
      return
    }
    const savedItems = settings['display.items']
    if (!savedItems) {
      form.reset(defaultValues)
      return
    }
    const parsed = savedItems
      .split(',')
      .map((entry) => entry.trim())
      .filter(Boolean)
    form.reset({
      items: parsed.length > 0 ? parsed : defaultValues.items,
    })
  }, [form, loading, settings])

  const handleSubmit = async (data: DisplayFormValues) => {
    setSaveMessage(null)
    setSaveError(null)
    try {
      await saveSettings({
        'display.items': data.items.join(','),
      })
      setSaveMessage('Display settings saved.')
    } catch (err) {
      setSaveError(
        err instanceof Error ? err.message : 'failed_to_save_display'
      )
    }
  }

  return (
    <Form {...form}>
      <form
        onSubmit={(event) => {
          const selectedItems = form.getValues('items') ?? []
          if (selectedItems.length === 0) {
            event.preventDefault()
            form.setError('items', {
              type: 'manual',
              message: 'You have to select at least one item.',
            })
            return
          }
          void form.handleSubmit(handleSubmit)(event)
        }}
        className='space-y-8'
      >
        {error ? (
          profileContextMissing ? (
            <ProfileContextBlocked
              error={error}
              onRetry={() => void reload()}
            />
          ) : (
            <div className='rounded-md border border-destructive/40 bg-destructive/10 p-3 text-sm'>
              <p className='font-medium'>Failed to load display settings.</p>
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
              name='items'
              render={({ field }) => (
                <FormItem>
                  <div className='mb-4'>
                    <FormLabel className='text-base'>Sidebar</FormLabel>
                    <FormDescription>
                      Select the items you want to display in the sidebar.
                    </FormDescription>
                  </div>
                  {items.map((item) => {
                    const selectedItems = field.value ?? []
                    const checkboxID = `settings-display-${item.id}`
                    return (
                      <FormItem
                        key={item.id}
                        className='flex flex-row items-start'
                      >
                        <FormControl>
                          <Checkbox
                            id={checkboxID}
                            data-testid={`settings-display-${item.id}`}
                            checked={selectedItems.includes(item.id)}
                            onCheckedChange={(checked) => {
                              field.onChange(
                                checked
                                  ? Array.from(
                                      new Set([...selectedItems, item.id])
                                    )
                                  : selectedItems.filter(
                                      (value) => value !== item.id
                                    )
                              )
                            }}
                          />
                        </FormControl>
                        <FormLabel htmlFor={checkboxID} className='font-normal'>
                          {item.label}
                        </FormLabel>
                      </FormItem>
                    )
                  })}
                  <Button
                    type='button'
                    variant='outline'
                    size='sm'
                    className='mt-3'
                    disabled={loading}
                    onClick={() => {
                      field.onChange([])
                      form.clearErrors('items')
                    }}
                  >
                    Clear selection
                  </Button>
                  <FormMessage />
                </FormItem>
              )}
            />
            <Button type='submit' disabled={saving || loading}>
              Update display
            </Button>
          </>
        )}
      </form>
    </Form>
  )
}
