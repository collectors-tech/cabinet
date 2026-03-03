import { Link } from '@tanstack/react-router'
import { Button } from '@/components/ui/button'

type ProfileContextBlockedProps = {
  error: string
  onRetry: () => void
}

export function ProfileContextBlocked({
  error,
  onRetry,
}: ProfileContextBlockedProps) {
  return (
    <div
      data-testid='settings-profile-context-blocked'
      className='rounded-md border border-destructive/40 bg-destructive/10 p-3 text-sm'
    >
      <p className='font-medium'>Active profile is required.</p>
      <p className='mt-1 text-muted-foreground'>{error}</p>
      <p className='mt-1 text-muted-foreground'>
        Select or create a profile before editing settings.
      </p>
      <div className='mt-3 flex gap-2'>
        <Button type='button' variant='outline' size='sm' onClick={onRetry}>
          Retry
        </Button>
        <Button asChild type='button' variant='default' size='sm'>
          <Link to='/'>Create or Select Profile</Link>
        </Button>
      </div>
    </div>
  )
}
