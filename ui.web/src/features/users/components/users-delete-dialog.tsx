'use client'

import { useState } from 'react'
import { AlertTriangle } from 'lucide-react'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { type User } from '../data/schema'

type UserDeleteDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  currentRow: User
  onDeleted: () => Promise<void> | void
}

export function UsersDeleteDialog({
  open,
  onOpenChange,
  currentRow,
  onDeleted,
}: UserDeleteDialogProps) {
  const [value, setValue] = useState('')
  const [error, setError] = useState<string | null>(null)

  const handleDelete = async () => {
    setError(null)
    if (value.trim() !== currentRow.username) return

    try {
      const response = await fetch(`/api/users/${currentRow.id}`, {
        method: 'DELETE',
      })
      if (!response.ok) {
        throw new Error(`users_delete_failed_${response.status}`)
      }
      await onDeleted()
      setValue('')
      onOpenChange(false)
    } catch (deleteError) {
      const message =
        deleteError instanceof Error
          ? deleteError.message
          : 'users_delete_failed'
      setError(message)
    }
  }

  return (
    <ConfirmDialog
      open={open}
      onOpenChange={onOpenChange}
      handleConfirm={handleDelete}
      disabled={value.trim() !== currentRow.username}
      title={
        <span className='text-destructive'>
          <AlertTriangle
            className='me-1 inline-block stroke-destructive'
            size={18}
          />{' '}
          Delete User
        </span>
      }
      desc={
        <div className='space-y-4'>
          <p className='mb-2'>
            Are you sure you want to delete{' '}
            <span className='font-bold'>{currentRow.username}</span>?
            <br />
            This action will permanently remove the user with the role of{' '}
            <span className='font-bold'>
              {currentRow.role.toUpperCase()}
            </span>{' '}
            from the system. This cannot be undone.
          </p>

          <Label className='my-2'>
            Username:
            <Input
              value={value}
              onChange={(e) => setValue(e.target.value)}
              placeholder='Enter username to confirm deletion.'
            />
          </Label>

          <Alert variant='destructive'>
            <AlertTitle>Warning!</AlertTitle>
            <AlertDescription>
              Please be careful, this operation can not be rolled back.
            </AlertDescription>
          </Alert>
          {error ? <p className='text-sm font-medium text-destructive'>{error}</p> : null}
        </div>
      }
      confirmText='Delete'
      destructive
    />
  )
}
