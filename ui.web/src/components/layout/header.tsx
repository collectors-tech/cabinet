import { useEffect, useState } from 'react'
import { MessageSquare } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useShellWorkspace } from '@/context/shell-workspace-provider'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import { SidebarTrigger } from '@/components/ui/sidebar'

type HeaderProps = React.HTMLAttributes<HTMLElement> & {
  fixed?: boolean
  ref?: React.Ref<HTMLElement>
}

export function Header({ className, fixed, children, ...props }: HeaderProps) {
  const [offset, setOffset] = useState(0)
  const { activeWorkspace, toggleAssistantWorkspace } = useShellWorkspace()
  const assistantActive = activeWorkspace === 'assistant'

  useEffect(() => {
    const onScroll = () => {
      setOffset(document.body.scrollTop || document.documentElement.scrollTop)
    }

    // Add scroll listener to the body
    document.addEventListener('scroll', onScroll, { passive: true })

    // Clean up the event listener on unmount
    return () => document.removeEventListener('scroll', onScroll)
  }, [])

  return (
    <header
      className={cn(
        'z-50 h-16',
        fixed && 'header-fixed peer/header sticky top-0 w-[inherit]',
        offset > 10 && fixed ? 'shadow' : 'shadow-none',
        className
      )}
      {...props}
    >
      <div
        className={cn(
          'relative flex h-full items-center gap-3 p-4 sm:gap-4',
          offset > 10 &&
            fixed &&
            'after:absolute after:inset-0 after:-z-10 after:bg-background/20 after:backdrop-blur-lg'
        )}
      >
        <SidebarTrigger variant='outline' className='max-md:scale-125' />
        <Separator orientation='vertical' className='h-6' />
        {children}
        <Button
          data-testid='shell-chat-toggle'
          type='button'
          variant={assistantActive ? 'default' : 'outline'}
          size='icon'
          aria-label={
            assistantActive
              ? 'Return to navigation workspace'
              : 'Open assistant workspace'
          }
          title={
            assistantActive
              ? 'Return to navigation workspace'
              : 'Open assistant workspace'
          }
          onClick={toggleAssistantWorkspace}
          className='shrink-0'
        >
          <MessageSquare className='h-4 w-4' />
        </Button>
      </div>
    </header>
  )
}
