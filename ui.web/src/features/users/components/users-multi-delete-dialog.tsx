'use client'

import { useState } from 'react'
import { type Table } from '@tanstack/react-table'
import { AlertTriangle } from 'lucide-react'
import { toast } from 'sonner'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { ConfirmDialog } from '@/components/confirm-dialog'

type UserMultiDeleteDialogProps<TData> = {
  open: boolean
  onOpenChange: (open: boolean) => void
  table: Table<TData>
  onDeleted: () => Promise<void> | void
}

const CONFIRM_WORD = 'DELETE'

function usersDeleteHistory(id: string, title: string, summary?: string) {
  return {
    history: {
      id: `${id}-${Date.now()}`,
      title,
      summary,
      source_label: 'Users delete dialog',
      category: 'users',
    },
  }
}

export function UsersMultiDeleteDialog<TData>({
  open,
  onOpenChange,
  table,
  onDeleted,
}: UserMultiDeleteDialogProps<TData>) {
  const [value, setValue] = useState('')

  const selectedRows = table.getFilteredSelectedRowModel().rows

  const handleDelete = () => {
    if (value.trim() !== CONFIRM_WORD) {
      toast.error(
        `Please type "${CONFIRM_WORD}" to confirm.`,
        usersDeleteHistory(
          'users-delete-confirmation-invalid',
          'Users delete confirmation blocked',
          'Users delete confirmation validation feedback was preserved in Inbox history.'
        ) as never
      )
      return
    }

    toast.promise(
      async () => {
        await Promise.all(
          selectedRows.map((row) =>
            fetch(`/api/users/${(row.original as { id: string }).id}`, {
              method: 'DELETE',
            }).then((response) => {
              if (!response.ok) {
                throw new Error(`users_bulk_delete_failed_${response.status}`)
              }
            })
          )
        )
        await onDeleted()
        setValue('')
        table.resetRowSelection()
        onOpenChange(false)
      },
      {
        loading: 'Deleting users...',
        success: `Deleted ${selectedRows.length} ${
          selectedRows.length > 1 ? 'users' : 'user'
        }`,
        error: 'Error',
        ...usersDeleteHistory(
          'users-bulk-delete',
          'Delete selected users',
          'Users bulk delete feedback was preserved in Inbox history.'
        ),
      }
    )
  }

  return (
    <ConfirmDialog
      open={open}
      onOpenChange={onOpenChange}
      handleConfirm={handleDelete}
      disabled={value.trim() !== CONFIRM_WORD}
      title={
        <span className='text-destructive'>
          <AlertTriangle
            className='me-1 inline-block stroke-destructive'
            size={18}
          />{' '}
          Delete {selectedRows.length}{' '}
          {selectedRows.length > 1 ? 'users' : 'user'}
        </span>
      }
      desc={
        <div className='space-y-4'>
          <p className='mb-2'>
            Are you sure you want to delete the selected users? <br />
            This action cannot be undone.
          </p>

          <Label className='my-4 flex flex-col items-start gap-1.5'>
            <span className=''>Confirm by typing "{CONFIRM_WORD}":</span>
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
