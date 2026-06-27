import { useEffect, useState } from 'react'
import { z } from 'zod'
import { useForm } from 'react-hook-form'
import { ChevronDownIcon } from '@radix-ui/react-icons'
import { zodResolver } from '@hookform/resolvers/zod'
import { fonts } from '@/config/fonts'
import { useTranslation } from 'react-i18next'
import { cn } from '@/lib/utils'
import { useFont } from '@/context/font-provider'
import { useTheme } from '@/context/theme-provider'
import { Button, buttonVariants } from '@/components/ui/button'
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
import { ProfileContextBlocked } from '../components/profile-context-blocked'
import { recordSettingsFeedbackHistory } from '../settings-feedback-history'
import { useProfileSettings } from '../use-profile-settings'

const appearanceFormSchema = z.object({
  theme: z.enum(['light', 'dark']),
  font: z.enum(fonts),
  language: z.enum(['en', 'zh', 'ja']),
})

type AppearanceFormValues = z.infer<typeof appearanceFormSchema>

export function AppearanceForm() {
  const { i18n, t } = useTranslation('common')
  const {
    settings,
    loading,
    error,
    profileContextMissing,
    saving,
    saveSettings,
    reload,
  } = useProfileSettings()
  const { font, setFont } = useFont()
  const { theme, setTheme } = useTheme()
  const [saveMessage, setSaveMessage] = useState<string | null>(null)
  const [saveError, setSaveError] = useState<string | null>(null)

  const defaultValues: Partial<AppearanceFormValues> = {
    theme: theme as 'light' | 'dark',
    font,
    language:
      i18n.resolvedLanguage === 'zh' || i18n.resolvedLanguage === 'ja'
        ? (i18n.resolvedLanguage as 'zh' | 'ja')
        : 'en',
  }

  const form = useForm<AppearanceFormValues>({
    resolver: zodResolver(appearanceFormSchema),
    defaultValues,
  })

  useEffect(() => {
    if (loading) {
      return
    }
    form.reset({
      theme:
        settings['appearance.theme'] === 'light' ||
        settings['appearance.theme'] === 'dark'
          ? (settings['appearance.theme'] as 'light' | 'dark')
          : (theme as 'light' | 'dark'),
      font: fonts.includes(
        settings['appearance.font'] as (typeof fonts)[number]
      )
        ? (settings['appearance.font'] as (typeof fonts)[number])
        : font,
      language:
        settings['appearance.language'] === 'zh' ||
        settings['appearance.language'] === 'ja' ||
        settings['appearance.language'] === 'en'
          ? (settings['appearance.language'] as 'en' | 'zh' | 'ja')
          : i18n.resolvedLanguage === 'zh' || i18n.resolvedLanguage === 'ja'
            ? (i18n.resolvedLanguage as 'zh' | 'ja')
            : 'en',
    })
  }, [font, form, i18n.resolvedLanguage, loading, settings, theme])

  async function onSubmit(data: AppearanceFormValues) {
    setSaveMessage(null)
    setSaveError(null)
    try {
      await saveSettings({
        'appearance.theme': data.theme,
        'appearance.font': data.font,
        'appearance.language': data.language,
      })
      if (data.font != font) setFont(data.font)
      if (data.theme != theme) setTheme(data.theme)
      if (data.language !== i18n.language) {
        await i18n.changeLanguage(data.language)
      }
      setSaveMessage('Appearance settings saved.')
      recordSettingsFeedbackHistory({
        id: 'settings-appearance-save-success',
        level: 'success',
        title: 'Appearance settings saved.',
        summary:
          'Appearance preference save feedback was preserved for review.',
        source: 'appearance',
        sourceLabel: 'Appearance',
      })
    } catch (err) {
      const message =
        err instanceof Error ? err.message : 'failed_to_save_appearance'
      setSaveError(message)
      recordSettingsFeedbackHistory({
        id: 'settings-appearance-save-failed',
        level: 'error',
        title: 'Appearance settings failed to save.',
        summary: message,
        source: 'appearance',
        sourceLabel: 'Appearance',
      })
    }
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className='space-y-8'>
        {error ? (
          profileContextMissing ? (
            <ProfileContextBlocked
              error={error}
              onRetry={() => void reload()}
            />
          ) : (
            <div className='rounded-md border border-destructive/40 bg-destructive/10 p-3 text-sm'>
              <p className='font-medium'>Failed to load appearance settings.</p>
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
              name='font'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Font</FormLabel>
                  <div className='relative w-max'>
                    <FormControl>
                      <select
                        className={cn(
                          buttonVariants({ variant: 'outline' }),
                          'w-[200px] appearance-none font-normal capitalize',
                          'dark:bg-background dark:hover:bg-background'
                        )}
                        {...field}
                      >
                        {fonts.map((font) => (
                          <option key={font} value={font}>
                            {font}
                          </option>
                        ))}
                      </select>
                    </FormControl>
                    <ChevronDownIcon className='absolute end-3 top-2.5 h-4 w-4 opacity-50' />
                  </div>
                  <FormDescription className='font-manrope'>
                    Set the font you want to use in the dashboard.
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='language'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Language</FormLabel>
                  <div className='relative w-max'>
                    <FormControl>
                      <select
                        data-testid='appearance-language-select'
                        className={cn(
                          buttonVariants({ variant: 'outline' }),
                          'w-[200px] appearance-none font-normal',
                          'dark:bg-background dark:hover:bg-background'
                        )}
                        {...field}
                      >
                        <option value='en'>English</option>
                        <option value='zh'>Chinese</option>
                        <option value='ja'>Japanese</option>
                      </select>
                    </FormControl>
                    <ChevronDownIcon className='absolute end-3 top-2.5 h-4 w-4 opacity-50' />
                  </div>
                  <FormDescription>
                    Select your preferred language for the dashboard UI.
                  </FormDescription>
                  <p
                    className='text-xs text-muted-foreground'
                    data-testid='appearance-language-fallback-sample'
                  >
                    {t('appearance.sampleText')}
                  </p>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name='theme'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Theme</FormLabel>
                  <FormDescription>
                    Select the theme for the dashboard.
                  </FormDescription>
                  <FormMessage />
                  <RadioGroup
                    onValueChange={field.onChange}
                    defaultValue={field.value}
                    className='grid max-w-md grid-cols-2 gap-8 pt-2'
                  >
                    <FormItem>
                      <FormLabel className='[&:has([data-state=checked])>div]:border-primary'>
                        <FormControl>
                          <RadioGroupItem value='light' className='sr-only' />
                        </FormControl>
                        <div className='items-center rounded-md border-2 border-muted p-1 hover:border-accent'>
                          <div className='space-y-2 rounded-sm bg-[#ecedef] p-2'>
                            <div className='space-y-2 rounded-md bg-white p-2 shadow-xs'>
                              <div className='h-2 w-[80px] rounded-lg bg-[#ecedef]' />
                              <div className='h-2 w-[100px] rounded-lg bg-[#ecedef]' />
                            </div>
                            <div className='flex items-center space-x-2 rounded-md bg-white p-2 shadow-xs'>
                              <div className='h-4 w-4 rounded-full bg-[#ecedef]' />
                              <div className='h-2 w-[100px] rounded-lg bg-[#ecedef]' />
                            </div>
                            <div className='flex items-center space-x-2 rounded-md bg-white p-2 shadow-xs'>
                              <div className='h-4 w-4 rounded-full bg-[#ecedef]' />
                              <div className='h-2 w-[100px] rounded-lg bg-[#ecedef]' />
                            </div>
                          </div>
                        </div>
                        <span className='block w-full p-2 text-center font-normal'>
                          Light
                        </span>
                      </FormLabel>
                    </FormItem>
                    <FormItem>
                      <FormLabel className='[&:has([data-state=checked])>div]:border-primary'>
                        <FormControl>
                          <RadioGroupItem value='dark' className='sr-only' />
                        </FormControl>
                        <div className='items-center rounded-md border-2 border-muted bg-popover p-1 hover:bg-accent hover:text-accent-foreground'>
                          <div className='space-y-2 rounded-sm bg-slate-950 p-2'>
                            <div className='space-y-2 rounded-md bg-slate-800 p-2 shadow-xs'>
                              <div className='h-2 w-[80px] rounded-lg bg-slate-400' />
                              <div className='h-2 w-[100px] rounded-lg bg-slate-400' />
                            </div>
                            <div className='flex items-center space-x-2 rounded-md bg-slate-800 p-2 shadow-xs'>
                              <div className='h-4 w-4 rounded-full bg-slate-400' />
                              <div className='h-2 w-[100px] rounded-lg bg-slate-400' />
                            </div>
                            <div className='flex items-center space-x-2 rounded-md bg-slate-800 p-2 shadow-xs'>
                              <div className='h-4 w-4 rounded-full bg-slate-400' />
                              <div className='h-2 w-[100px] rounded-lg bg-slate-400' />
                            </div>
                          </div>
                        </div>
                        <span className='block w-full p-2 text-center font-normal'>
                          Dark
                        </span>
                      </FormLabel>
                    </FormItem>
                  </RadioGroup>
                </FormItem>
              )}
            />

            <Button type='submit' disabled={saving || loading}>
              Update preferences
            </Button>
          </>
        )}
      </form>
    </Form>
  )
}
