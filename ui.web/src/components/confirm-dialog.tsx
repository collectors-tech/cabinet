import { useEffect, useRef } from 'react'
import { cn } from '@/lib/utils'
import { recordNotificationHistory } from '@/lib/toast-history'
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

type ConfirmDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  title: React.ReactNode
  disabled?: boolean
  desc: React.JSX.Element | string
  cancelBtnText?: string
  confirmText?: React.ReactNode
  destructive?: boolean
  handleConfirm: () => void
  isLoading?: boolean
  className?: string
  contentTestId?: string
  cancelTestId?: string
  confirmTestId?: string
  children?: React.ReactNode
}

function textFromNode(value: React.ReactNode, fallback: string) {
  if (typeof value === 'string') {
    return value
  }
  if (typeof value === 'number' || typeof value === 'boolean') {
    return String(value)
  }
  return fallback
}

export function ConfirmDialog(props: ConfirmDialogProps) {
  const {
    title,
    desc,
    children,
    className,
    confirmText,
    cancelBtnText,
    destructive,
    isLoading,
    disabled = false,
    handleConfirm,
    contentTestId,
    cancelTestId,
    confirmTestId,
    ...actions
  } = props
  const wasOpen = useRef(false)

  useEffect(() => {
    if (!actions.open || wasOpen.current) {
      wasOpen.current = Boolean(actions.open)
      return
    }
    wasOpen.current = true
    recordNotificationHistory({
      level: destructive ? 'warning' : 'info',
      title: textFromNode(title, 'Confirmation dialog'),
      summary: textFromNode(
        desc,
        'A confirmation dialog opened and was preserved in Inbox history.'
      ),
      source_label: 'Dialog History',
      category: 'system',
    })
  }, [actions.open, desc, destructive, title])

  return (
    <AlertDialog {...actions}>
      <AlertDialogContent
        className={cn(className && className)}
        data-testid={contentTestId}
      >
        <AlertDialogHeader className='text-start'>
          <AlertDialogTitle>{title}</AlertDialogTitle>
          <AlertDialogDescription asChild>
            <div>{desc}</div>
          </AlertDialogDescription>
        </AlertDialogHeader>
        {children}
        <AlertDialogFooter>
          <AlertDialogCancel disabled={isLoading} data-testid={cancelTestId}>
            {cancelBtnText ?? 'Cancel'}
          </AlertDialogCancel>
          <Button
            variant={destructive ? 'destructive' : 'default'}
            onClick={handleConfirm}
            disabled={disabled || isLoading}
            data-testid={confirmTestId}
          >
            {confirmText ?? 'Continue'}
          </Button>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
