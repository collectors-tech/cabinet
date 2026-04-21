import { useState } from 'react'
import { type Table } from '@tanstack/react-table'
import { AlertTriangle } from 'lucide-react'
import { toast } from 'sonner'
import { sleep } from '@/lib/utils'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { type Task } from '../data/schema'

type TaskMultiDeleteDialogProps<TData> = {
  open: boolean
  onOpenChange: (open: boolean) => void
  table: Table<TData>
  routePath: '/_authenticated/inventory/' | '/_authenticated/wishlist/'
  onWishlistBulkDelete?: (tasks: Task[]) => Promise<void>
  isLoading?: boolean
}

const CONFIRM_WORD = 'DELETE'

export function TasksMultiDeleteDialog<TData>({
  open,
  onOpenChange,
  table,
  routePath,
  onWishlistBulkDelete,
  isLoading = false,
}: TaskMultiDeleteDialogProps<TData>) {
  const [value, setValue] = useState('')
  const isWishlistRoute = routePath === '/_authenticated/wishlist/'

  const selectedRows = table.getFilteredSelectedRowModel().rows
  const selectedTasks = selectedRows.map((row) => row.original as Task)

  const handleDelete = () => {
    if (value.trim() !== CONFIRM_WORD) {
      toast.error(`Please type "${CONFIRM_WORD}" to confirm.`)
      return
    }

    if (isWishlistRoute && onWishlistBulkDelete) {
      void onWishlistBulkDelete(selectedTasks)
      setValue('')
      onOpenChange(false)
      return
    }

    onOpenChange(false)

    toast.promise(sleep(2000), {
      loading: 'Deleting tasks...',
      success: () => {
        setValue('')
        table.resetRowSelection()
        return `Deleted ${selectedRows.length} ${
          selectedRows.length > 1 ? 'tasks' : 'task'
        }`
      },
      error: 'Error',
    })
  }

  return (
    <ConfirmDialog
      open={open}
      onOpenChange={onOpenChange}
      handleConfirm={handleDelete}
      disabled={value.trim() !== CONFIRM_WORD}
      isLoading={isLoading}
      title={
        <span className='text-destructive'>
          <AlertTriangle
            className='me-1 inline-block stroke-destructive'
            size={18}
          />{' '}
          Delete {selectedRows.length}{' '}
          {selectedRows.length > 1
            ? isWishlistRoute
              ? 'wishlist entries'
              : 'tasks'
            : isWishlistRoute
              ? 'wishlist entry'
              : 'task'}
        </span>
      }
      desc={
        <div className='space-y-4'>
          <p className='mb-2'>
            {isWishlistRoute
              ? 'Are you sure you want to delete the selected wishlist entries?'
              : 'Are you sure you want to delete the selected tasks?'}{' '}
            <br />
            This action cannot be undone.
          </p>

          <Label className='my-4 flex flex-col items-start gap-1.5'>
            <span>Confirm by typing "{CONFIRM_WORD}":</span>
            <Input
              value={value}
              onChange={(e) => setValue(e.target.value)}
              placeholder={`Type "${CONFIRM_WORD}" to confirm.`}
            />
          </Label>

          <Alert variant='destructive'>
            <AlertTitle>Warning!</AlertTitle>
            <AlertDescription>
              Please be careful, this operation can not be rolled back.
            </AlertDescription>
          </Alert>
        </div>
      }
      confirmText='Delete'
      destructive
    />
  )
}
