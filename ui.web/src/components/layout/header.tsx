import {
  type ComponentType,
  useEffect,
  useLayoutEffect,
  useRef,
  useState,
} from 'react'
import { MessageSquare } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useShellWorkspace } from '@/context/shell-workspace-context'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import { SidebarTrigger, useSidebar } from '@/components/ui/sidebar'

type HeaderProps = React.HTMLAttributes<HTMLElement> & {
  fixed?: boolean
  ref?: React.Ref<HTMLElement>
}

type HeaderTitleProps = {
  title: string
  description?: string
  icon: ComponentType<{ className?: string; 'aria-hidden'?: true }>
  testId: string
  iconTestId?: string
  className?: string
}

export function HeaderTitle({
  title,
  description,
  icon: Icon,
  testId,
  iconTestId,
  className,
}: HeaderTitleProps) {
  const titleText = title.trim()
  const hint = description?.trim()
  const titleRef = useRef<HTMLHeadingElement | null>(null)
  const [isCrowded, setIsCrowded] = useState(false)

  useLayoutEffect(() => {
    const titleElement = titleRef.current
    const headerElement = titleElement?.closest('header')
    if (!titleElement || !headerElement) {
      return
    }

    let frame = 0
    const measure = () => {
      cancelAnimationFrame(frame)
      frame = requestAnimationFrame(() => {
        const titleRect = titleElement.getBoundingClientRect()
        const avoidElements = Array.from(
          headerElement.querySelectorAll('[data-header-title-avoid="true"]')
        )
        const overlapsHeaderControls = avoidElements.some((element) => {
          const controlRect = element.getBoundingClientRect()
          const hasHorizontalOverlap =
            titleRect.right > controlRect.left - 8 &&
            titleRect.left < controlRect.right + 8
          const hasVerticalOverlap =
            titleRect.bottom > controlRect.top &&
            titleRect.top < controlRect.bottom

          return hasHorizontalOverlap && hasVerticalOverlap
        })

        setIsCrowded(overlapsHeaderControls)
      })
    }

    const resizeObserver = new ResizeObserver(measure)
    resizeObserver.observe(headerElement)
    resizeObserver.observe(titleElement)
    Array.from(
      headerElement.querySelectorAll('[data-header-title-avoid="true"]')
    ).forEach((element) => resizeObserver.observe(element))

    window.addEventListener('resize', measure)
    measure()

    return () => {
      cancelAnimationFrame(frame)
      resizeObserver.disconnect()
      window.removeEventListener('resize', measure)
    }
  }, [titleText])

  return (
    <div
      className={cn(
        'pointer-events-none z-10 flex max-w-28 min-w-20 shrink-0 justify-center md:absolute md:top-1/2 md:left-1/2 md:max-w-[min(34rem,42vw)] md:min-w-0 md:-translate-x-1/2 md:-translate-y-1/2',
        isCrowded && 'md:opacity-0',
        className
      )}
      data-crowded={isCrowded ? 'true' : 'false'}
    >
      <h1
        ref={titleRef}
        className='flex min-w-0 items-center justify-center gap-2 truncate text-center text-lg font-bold tracking-tight'
        data-testid={testId}
        data-centered='true'
        data-crowded={isCrowded ? 'true' : 'false'}
        title={hint || titleText}
        aria-hidden={isCrowded ? true : undefined}
        aria-label={
          isCrowded ? undefined : hint ? `${titleText} - ${hint}` : titleText
        }
      >
        <Icon
          aria-hidden
          className='h-5 w-5 shrink-0 text-muted-foreground'
          data-testid={iconTestId}
        />
        <span className='truncate'>{titleText}</span>
      </h1>
    </div>
  )
}

export function Header({ className, fixed, children, ...props }: HeaderProps) {
  const [offset, setOffset] = useState(0)
  const { activeWorkspace, toggleAssistantWorkspace } = useShellWorkspace()
  const { isMobile, setOpenMobile } = useSidebar()
  const assistantActive = activeWorkspace === 'assistant'

  const handleAssistantToggle = () => {
    if (!assistantActive && isMobile) {
      setOpenMobile(true)
    }
    toggleAssistantWorkspace()
  }

  const handleAssistantToggleKeyDown = (
    event: React.KeyboardEvent<HTMLButtonElement>
  ) => {
    if (event.key !== 'Enter' && event.key !== ' ') return
    event.preventDefault()
    handleAssistantToggle()
  }

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
        <Button
          data-testid='shell-chat-toggle'
          type='button'
          variant={assistantActive ? 'default' : 'outline'}
          size='icon'
          aria-label={
            assistantActive ? 'Close Cabinet Agent' : 'Open Cabinet Agent'
          }
          title={assistantActive ? 'Close Cabinet Agent' : 'Open Cabinet Agent'}
          onClick={handleAssistantToggle}
          onKeyDown={handleAssistantToggleKeyDown}
          className='shrink-0'
        >
          <MessageSquare className='h-4 w-4' />
        </Button>
        {children}
      </div>
    </header>
  )
}
