import { useEffect } from 'react'
import { useNavigate, useRouter } from '@tanstack/react-router'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'

type GeneralErrorProps = React.HTMLAttributes<HTMLDivElement> & {
  minimal?: boolean
  error?: unknown
}

export function GeneralError({
  className,
  minimal = false,
  error,
}: GeneralErrorProps) {
  const navigate = useNavigate()
  const { history } = useRouter()

  useEffect(() => {
    if (typeof window === 'undefined' || !error) {
      return
    }

    const message =
      error instanceof Error ? error.message : String(error ?? '')
    const isChunkLoadError =
      /dynamically imported module/i.test(message) ||
      /loading chunk/i.test(message) ||
      /importing a module script failed/i.test(message)

    if (!isChunkLoadError) {
      return
    }

    const reloadKey = 'cabinet.chunk-reload-once'
    if (window.sessionStorage.getItem(reloadKey) === '1') {
      window.sessionStorage.removeItem(reloadKey)
      return
    }
    window.sessionStorage.setItem(reloadKey, '1')
    window.location.reload()
  }, [error])

  return (
    <div className={cn('h-svh w-full', className)}>
      <div className='m-auto flex h-full w-full flex-col items-center justify-center gap-2'>
        {!minimal && (
          <h1 className='text-[7rem] leading-tight font-bold'>500</h1>
        )}
        <span className='font-medium'>Oops! Something went wrong {`:')`}</span>
        <p className='text-center text-muted-foreground'>
          We apologize for the inconvenience. <br /> Please try again later.
        </p>
        {!minimal && (
          <div className='mt-6 flex gap-4'>
            <Button variant='outline' onClick={() => history.go(-1)}>
              Go Back
            </Button>
            <Button onClick={() => navigate({ to: '/' })}>Back to Home</Button>
          </div>
        )}
      </div>
    </div>
  )
}
