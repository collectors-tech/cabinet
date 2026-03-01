import { AlertTriangle } from 'lucide-react'
import { ContentSection } from '../components/content-section'
import { Button } from '@/components/ui/button'

export function SettingsStorage() {
  return (
    <ContentSection
      title='Storage'
      desc='Review database/media paths and run storage maintenance actions.'
    >
      <div className='space-y-4 text-sm'>
        <div className='rounded-md border p-3'>
          <p className='font-medium'>Database location</p>
          <p className='text-muted-foreground'>Profile-scoped local database path is managed by runtime configuration.</p>
        </div>
        <div className='rounded-md border p-3'>
          <p className='font-medium'>Media location</p>
          <p className='text-muted-foreground'>Photo/media storage remains local-first under the active profile workspace.</p>
        </div>
        <div className='rounded-md border border-amber-300/40 bg-amber-50/20 p-3 text-amber-950 dark:text-amber-200'>
          <div className='flex items-start gap-2'>
            <AlertTriangle className='mt-0.5 h-4 w-4' />
            <p>
              Storage repair and migration actions are restricted to diagnostics workflows.
            </p>
          </div>
        </div>
        <div className='flex gap-2'>
          <Button variant='outline' size='sm'>Reindex Search</Button>
          <Button variant='outline' size='sm'>Rebuild Thumbnails</Button>
        </div>
      </div>
    </ContentSection>
  )
}
