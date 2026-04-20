import { Button } from '@/components/ui/button'
import { useTranslation } from 'react-i18next'
import { ContentSection } from '../components/content-section'

export function SettingsOperations() {
  const { t } = useTranslation('pages')

  return (
    <ContentSection
      title={t('settings.operations.title')}
      desc={t('settings.operations.description')}
    >
      <div className='space-y-4 text-sm'>
        <div className='rounded-md border p-3'>
          <p className='font-medium'>Maintenance Window</p>
          <p className='text-muted-foreground'>
            Schedule maintenance tasks and control service-impacting operations.
          </p>
        </div>
        <div className='rounded-md border p-3'>
          <p className='font-medium'>Queue Controls</p>
          <p className='text-muted-foreground'>
            Pause and resume Market Watch and enrichment workers for active profile context.
          </p>
        </div>
        <div className='flex gap-2'>
          <Button variant='outline' size='sm' disabled>
            Pause Workers (Coming soon)
          </Button>
          <Button variant='outline' size='sm' disabled>
            Resume Workers (Coming soon)
          </Button>
        </div>
      </div>
    </ContentSection>
  )
}
