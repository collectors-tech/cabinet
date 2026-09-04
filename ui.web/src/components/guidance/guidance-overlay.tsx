import { useEffect, useState } from 'react'
import { useLocation } from '@tanstack/react-router'
import { ChevronLeft, ChevronRight, Pause, X } from 'lucide-react'
import {
  findUiTarget,
  uiGuidanceClearEventName,
  uiGuidanceEventName,
  type UiGuidanceRequest,
  type UiTarget,
} from '@/lib/ui-target-registry'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'

type ActiveGuidance = {
  request: UiGuidanceRequest
  target: UiTarget | null
  element: HTMLElement | null
  missingReason?: string
}

function routeMatches(pathname: string, target: UiTarget) {
  return pathname === target.route || pathname.startsWith(`${target.route}/`)
}

function focusTarget(element: HTMLElement, target: UiTarget) {
  element.scrollIntoView({
    behavior: 'smooth',
    block: target.scrollBehaviour,
    inline: 'nearest',
  })
  if (target.safeActions.includes('focus')) {
    element.focus({ preventScroll: true })
  }
}

export function GuidanceOverlay() {
  const location = useLocation()
  const [activeGuidance, setActiveGuidance] = useState<ActiveGuidance | null>(
    null
  )

  useEffect(() => {
    function handleGuidance(event: Event) {
      const detail = (event as CustomEvent<UiGuidanceRequest>).detail
      const targetId = detail?.targetId?.trim()
      if (!targetId) {
        return
      }

      const target = findUiTarget(targetId)
      if (!target) {
        setActiveGuidance({
          request: detail,
          target: null,
          element: null,
          missingReason: 'Target is not registered for guided walkthroughs.',
        })
        return
      }

      if (!routeMatches(location.pathname, target)) {
        setActiveGuidance({
          request: detail,
          target,
          element: null,
          missingReason: target.fallbackInstruction,
        })
        return
      }

      const element = document.querySelector<HTMLElement>(target.selector)
      if (!element) {
        setActiveGuidance({
          request: detail,
          target,
          element: null,
          missingReason: target.fallbackInstruction,
        })
        return
      }

      focusTarget(element, target)
      setActiveGuidance({ request: detail, target, element })
    }

    window.addEventListener(uiGuidanceEventName, handleGuidance)
    return () => {
      window.removeEventListener(uiGuidanceEventName, handleGuidance)
    }
  }, [location.pathname])

  useEffect(() => {
    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === 'Escape') {
        setActiveGuidance(null)
      }
    }
    function handleClearGuidance() {
      setActiveGuidance(null)
    }

    window.addEventListener('keydown', handleKeyDown)
    window.addEventListener(uiGuidanceClearEventName, handleClearGuidance)
    return () => {
      window.removeEventListener('keydown', handleKeyDown)
      window.removeEventListener(uiGuidanceClearEventName, handleClearGuidance)
    }
  }, [])

  const target = activeGuidance?.target
  const elementRect = activeGuidance?.element?.getBoundingClientRect()
  const style = elementRect
    ? {
        top: `${Math.max(8, elementRect.top - 6)}px`,
        left: `${Math.max(8, elementRect.left - 6)}px`,
        width: `${elementRect.width + 12}px`,
        height: `${elementRect.height + 12}px`,
      }
    : undefined

  if (!activeGuidance) {
    return null
  }

  return (
    <div
      aria-live='polite'
      className='pointer-events-none fixed inset-0 z-50'
      data-testid='ui-guidance-overlay'
    >
      {style ? (
        <div
          className='absolute rounded-md border-2 border-cyan-300 bg-cyan-300/10 shadow-[0_0_0_9999px_rgba(2,6,23,0.28)]'
          style={style}
          data-testid='ui-guidance-highlight'
        />
      ) : null}
      <section
        role='dialog'
        aria-label={activeGuidance.request.title || target?.label || 'Guidance'}
        className={cn(
          'pointer-events-auto fixed right-5 bottom-5 w-[min(22rem,calc(100vw-2.5rem))] rounded-md border border-slate-700 bg-slate-950 p-4 text-slate-100 shadow-xl',
          style && target?.calloutPlacement === 'top' ? 'bottom-20' : ''
        )}
        data-testid='ui-guidance-callout'
      >
        <div className='flex items-start justify-between gap-3'>
          <div className='min-w-0'>
            <p
              className='truncate text-sm font-semibold'
              data-testid='ui-guidance-title'
            >
              {activeGuidance.request.title ||
                target?.label ||
                'Missing target'}
            </p>
            <p className='mt-1 text-xs leading-5 text-slate-300'>
              {activeGuidance.request.instruction ||
                target?.description ||
                activeGuidance.missingReason}
            </p>
            {activeGuidance.missingReason ? (
              <p
                className='mt-2 rounded border border-amber-400/40 bg-amber-400/10 px-2 py-1 text-xs text-amber-100'
                data-testid='ui-guidance-fallback'
              >
                {activeGuidance.missingReason}
              </p>
            ) : null}
          </div>
          <Button
            type='button'
            variant='ghost'
            size='icon'
            className='h-8 w-8 shrink-0 text-slate-300 hover:bg-slate-800 hover:text-white'
            aria-label='Cancel guidance'
            title='Cancel guidance'
            data-testid='ui-guidance-cancel'
            onClick={() => setActiveGuidance(null)}
          >
            <X className='h-4 w-4' />
          </Button>
        </div>
        <div className='mt-3 flex items-center justify-between gap-2'>
          <div className='flex gap-1'>
            <Button
              type='button'
              variant='outline'
              size='icon'
              className='h-8 w-8 border-slate-700 bg-slate-900 text-slate-300'
              aria-label='Previous guidance step'
              title='Previous guidance step'
              data-testid='ui-guidance-previous'
              disabled
            >
              <ChevronLeft className='h-4 w-4' />
            </Button>
            <Button
              type='button'
              variant='outline'
              size='icon'
              className='h-8 w-8 border-slate-700 bg-slate-900 text-slate-300'
              aria-label='Next guidance step'
              title='Next guidance step'
              data-testid='ui-guidance-next'
              disabled
            >
              <ChevronRight className='h-4 w-4' />
            </Button>
          </div>
          <Button
            type='button'
            variant='outline'
            size='sm'
            className='border-slate-700 bg-slate-900 text-slate-200'
            data-testid='ui-guidance-pause'
            disabled
          >
            <Pause className='h-3.5 w-3.5' />
            Pause
          </Button>
        </div>
      </section>
    </div>
  )
}
