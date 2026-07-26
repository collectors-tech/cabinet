import { X } from 'lucide-react'
import type { LucideIcon } from 'lucide-react'
import { useEffect, useRef } from 'react'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { cn } from '@/lib/utils'

type RecordEditDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  title: string
  description: string
  icon?: LucideIcon
  children: React.ReactNode
  isDirty?: boolean
  isSubmitting?: boolean
  serverError?: string
  submitLabel?: string
  submittingLabel?: string
  cancelLabel?: string
  onSubmit: () => void | Promise<void>
  returnFocusRef?: React.RefObject<HTMLElement | null>
  className?: string
  testId?: string
}

export function RecordEditDialog({
  open,
  onOpenChange,
  title,
  description,
  icon: Icon,
  children,
  isDirty = false,
  isSubmitting = false,
  serverError,
  submitLabel = 'Save',
  submittingLabel = 'Saving',
  cancelLabel = 'Cancel',
  onSubmit,
  returnFocusRef,
  className,
  testId = 'record-edit-dialog',
}: RecordEditDialogProps) {
  const submitLockedRef = useRef(false)
  const closeBlocked = isDirty || isSubmitting

  useEffect(() => {
    if (!isSubmitting) {
      submitLockedRef.current = false
    }
  }, [isSubmitting])

  useEffect(() => {
    if (open) {
      return
    }
    const returnTarget = returnFocusRef?.current
    if (!returnTarget) {
      return
    }
    window.setTimeout(() => returnTarget.focus(), 0)
  }, [open, returnFocusRef])

  function requestClose() {
    if (closeBlocked) {
      return
    }
    onOpenChange(false)
    returnFocusRef?.current?.focus()
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(nextOpen) => {
        if (!nextOpen && closeBlocked) {
          return
        }
        onOpenChange(nextOpen)
        if (!nextOpen) {
          returnFocusRef?.current?.focus()
        }
      }}
    >
      <DialogContent
        className={cn('sm:max-w-xl', className)}
        data-testid={testId}
        showCloseButton={false}
        onEscapeKeyDown={(event) => {
          if (closeBlocked) {
            event.preventDefault()
          }
        }}
        onInteractOutside={(event) => {
          if (closeBlocked) {
            event.preventDefault()
          }
        }}
      >
        <DialogHeader className='pe-10'>
          <div className='flex items-start gap-3'>
            {Icon ? (
              <span
                className='mt-0.5 inline-flex size-9 shrink-0 items-center justify-center rounded-md border bg-muted text-muted-foreground'
                data-testid={`${testId}-icon`}
                aria-hidden='true'
              >
                <Icon className='size-4' />
              </span>
            ) : null}
            <div className='min-w-0 space-y-1'>
              <DialogTitle>{title}</DialogTitle>
              <DialogDescription>{description}</DialogDescription>
            </div>
          </div>
        </DialogHeader>
        <Button
          type='button'
          variant='ghost'
          size='icon'
          className='absolute end-4 top-4 size-8'
          aria-label={`Close ${title}`}
          disabled={isSubmitting}
          onClick={requestClose}
          data-testid={`${testId}-close`}
        >
          <X aria-hidden='true' />
        </Button>
        <form
          className='space-y-4'
          onSubmit={(event) => {
            event.preventDefault()
            if (isSubmitting || submitLockedRef.current) {
              return
            }
            submitLockedRef.current = true
            void Promise.resolve(onSubmit()).finally(() => {
              if (!isSubmitting) {
                submitLockedRef.current = false
              }
            })
          }}
        >
          {children}
          {serverError ? (
            <Alert variant='destructive' role='alert'>
              <AlertTitle>Could not save record</AlertTitle>
              <AlertDescription>{serverError}</AlertDescription>
            </Alert>
          ) : null}
          {closeBlocked ? (
            <Alert data-testid={`${testId}-dirty-warning`}>
              <AlertTitle>Unsaved changes</AlertTitle>
              <AlertDescription>
                Save the record or clear the changes before closing this dialog.
              </AlertDescription>
            </Alert>
          ) : null}
          <DialogFooter>
            <Button
              type='button'
              variant='outline'
              onClick={requestClose}
              disabled={isSubmitting}
            >
              {cancelLabel}
            </Button>
            <Button type='submit' disabled={isSubmitting}>
              {isSubmitting ? submittingLabel : submitLabel}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
