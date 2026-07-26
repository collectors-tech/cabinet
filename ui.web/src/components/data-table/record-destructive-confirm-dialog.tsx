import { useEffect, useRef } from 'react'
import { Archive, RotateCcw, Trash2 } from 'lucide-react'
import type { LucideIcon } from 'lucide-react'
import {
  AlertDialog,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

export type RecordDestructiveAction =
  | 'archive'
  | 'delete'
  | 'restore'
  | 'permanent-delete'

type RecordDestructiveConfirmDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  action: RecordDestructiveAction
  recordLabel: string
  consequence?: string
  isSubmitting?: boolean
  onConfirm: () => void | Promise<void>
  returnFocusRef?: React.RefObject<HTMLElement | null>
  className?: string
  testId?: string
}

const actionCopy: Record<
  RecordDestructiveAction,
  {
    titleVerb: string
    buttonLabel: string
    submittingLabel: string
    consequence: string
    icon: LucideIcon
    destructive: boolean
  }
> = {
  archive: {
    titleVerb: 'Archive',
    buttonLabel: 'Archive',
    submittingLabel: 'Archiving',
    consequence: 'This archives the record without permanent removal.',
    icon: Archive,
    destructive: false,
  },
  delete: {
    titleVerb: 'Delete',
    buttonLabel: 'Delete',
    submittingLabel: 'Deleting',
    consequence: 'This soft deletes the record so it can be restored later.',
    icon: Trash2,
    destructive: true,
  },
  restore: {
    titleVerb: 'Restore',
    buttonLabel: 'Restore',
    submittingLabel: 'Restoring',
    consequence: 'This restores the record to the active table.',
    icon: RotateCcw,
    destructive: false,
  },
  'permanent-delete': {
    titleVerb: 'Permanently delete',
    buttonLabel: 'Permanently delete',
    submittingLabel: 'Deleting',
    consequence: 'This permanently deletes the record.',
    icon: Trash2,
    destructive: true,
  },
}

export function RecordDestructiveConfirmDialog({
  open,
  onOpenChange,
  action,
  recordLabel,
  consequence,
  isSubmitting = false,
  onConfirm,
  returnFocusRef,
  className,
  testId = 'record-destructive-confirm-dialog',
}: RecordDestructiveConfirmDialogProps) {
  const submitLockedRef = useRef(false)
  const copy = actionCopy[action]
  const Icon = copy.icon

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

  return (
    <AlertDialog
      open={open}
      onOpenChange={(nextOpen) => {
        if (!nextOpen && isSubmitting) {
          return
        }
        onOpenChange(nextOpen)
        if (!nextOpen) {
          returnFocusRef?.current?.focus()
        }
      }}
    >
      <AlertDialogContent
        className={cn('sm:max-w-md', className)}
        data-testid={testId}
      >
        <AlertDialogHeader className='text-start'>
          <div className='flex items-start gap-3'>
            <span
              className={cn(
                'mt-0.5 inline-flex size-9 shrink-0 items-center justify-center rounded-md border bg-muted',
                copy.destructive
                  ? 'text-destructive'
                  : 'text-muted-foreground'
              )}
              data-testid={`${testId}-icon`}
              aria-hidden='true'
            >
              <Icon className='size-4' />
            </span>
            <div className='min-w-0 space-y-1'>
              <AlertDialogTitle>
                {copy.titleVerb} {recordLabel}?
              </AlertDialogTitle>
              <AlertDialogDescription>
                {copy.consequence}
              </AlertDialogDescription>
            </div>
          </div>
        </AlertDialogHeader>
        {consequence ? (
          <p className='text-sm text-muted-foreground'>{consequence}</p>
        ) : null}
        <AlertDialogFooter>
          <AlertDialogCancel disabled={isSubmitting}>Cancel</AlertDialogCancel>
          <Button
            type='button'
            variant={copy.destructive ? 'destructive' : 'default'}
            disabled={isSubmitting}
            onClick={() => {
              if (isSubmitting || submitLockedRef.current) {
                return
              }
              submitLockedRef.current = true
              void Promise.resolve(onConfirm()).finally(() => {
                if (!isSubmitting) {
                  submitLockedRef.current = false
                }
              })
            }}
          >
            {isSubmitting ? copy.submittingLabel : copy.buttonLabel}
          </Button>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
