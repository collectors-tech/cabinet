import { Button } from '@/components/ui/button'
import { ContentSection } from '../components/content-section'

export function SettingsBilling() {
  return (
    <ContentSection
      title='Billing'
      desc='Review subscription status and entitlement configuration.'
    >
      <div className='space-y-4 text-sm'>
        <div className='rounded-md border p-3'>
          <p className='font-medium'>Plan</p>
          <p className='text-muted-foreground'>
            Billing controls are visible here and sync with cloud entitlement state.
          </p>
        </div>
        <div className='rounded-md border p-3'>
          <p className='font-medium'>License Status</p>
          <p className='text-muted-foreground'>
            Check current license tier and renewal state for this account.
          </p>
        </div>
        <Button variant='outline' size='sm' disabled>
          Open Billing Portal (Coming soon)
        </Button>
      </div>
    </ContentSection>
  )
}
