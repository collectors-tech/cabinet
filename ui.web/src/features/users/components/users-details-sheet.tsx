import { Mail, ShieldAlert, Trash2, UserPen } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'
import { type User } from '../data/schema'
import { useUsers } from './users-provider'

type UsersDetailsSheetProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  user: User | null
  users: User[]
  onMutated: () => Promise<void> | void
}

const dateFormatter = new Intl.DateTimeFormat('en-AU', {
  dateStyle: 'medium',
  timeStyle: 'short',
})

function formatValue(value: string | Date | null | undefined) {
  if (!value) {
    return 'Not provided'
  }
  if (value instanceof Date) {
    return Number.isNaN(value.getTime())
      ? 'Not provided'
      : dateFormatter.format(value)
  }
  return value.trim() === '' ? 'Not provided' : value
}

function isProtectedLocalOwner(user: User) {
  return (
    user.username.toLowerCase().startsWith('owner_') &&
    user.email.toLowerCase().endsWith('@cabinet.local') &&
    user.role === 'admin'
  )
}

function wouldLeaveNoActiveAdmin(
  users: User[],
  target: User,
  nextRole: User['role'] | '',
  nextStatus: User['status'] | ''
) {
  return !users.some((user) => {
    const role = user.id === target.id ? nextRole : user.role
    const status = user.id === target.id ? nextStatus : user.status
    return role === 'admin' && status === 'active'
  })
}

function DetailRow({ label, value }: { label: string; value: string }) {
  return (
    <div className='grid grid-cols-[8rem_1fr] gap-3 border-b py-2 text-sm last:border-b-0'>
      <dt className='text-muted-foreground'>{label}</dt>
      <dd
        className={cn(
          'min-w-0 font-medium break-words',
          value === 'Not provided' && 'font-normal text-muted-foreground'
        )}
      >
        {value}
      </dd>
    </div>
  )
}

export function UsersDetailsSheet({
  open,
  onOpenChange,
  user,
  users,
  onMutated,
}: UsersDetailsSheetProps) {
  const { setOpen, setCurrentRow } = useUsers()

  if (!user) {
    return null
  }

  const protectedOwner = isProtectedLocalOwner(user)
  const cannotRemove =
    protectedOwner || wouldLeaveNoActiveAdmin(users, user, '', '')
  const cannotDowngrade = wouldLeaveNoActiveAdmin(
    users,
    user,
    'view',
    user.status
  )
  const canDeactivate = !wouldLeaveNoActiveAdmin(
    users,
    user,
    user.role,
    'inactive'
  )
  const statusAction = user.status === 'active' ? 'inactive' : 'active'
  const statusLabel =
    user.status === 'active' ? 'Deactivate user' : 'Activate user'

  const updateUser = async (payload: Partial<User>) => {
    const response = await fetch(`/api/users/${user.id}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    })
    if (!response.ok) {
      throw new Error(`users_update_failed_${response.status}`)
    }
    await onMutated()
  }

  const resendInvitation = async () => {
    const response = await fetch('/api/users/invite', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        email: user.email,
        role: user.role,
        desc: 'Resend user invitation',
      }),
    })
    if (!response.ok) {
      throw new Error(`users_resend_invite_failed_${response.status}`)
    }
    await onMutated()
  }

  const openEdit = () => {
    setCurrentRow(user)
    setOpen('edit')
  }

  const openDelete = () => {
    setCurrentRow(user)
    setOpen('delete')
  }

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent
        className='w-full sm:max-w-xl'
        data-testid='users-details-sheet'
      >
        <SheetHeader className='border-b pe-12 text-start'>
          <SheetTitle>{user.username}</SheetTitle>
          <SheetDescription>
            User details, access state, and available management actions.
          </SheetDescription>
          <div className='mt-2 flex flex-wrap gap-2'>
            <Badge variant='outline' className='capitalize'>
              {user.status}
            </Badge>
            <Badge variant='secondary' className='capitalize'>
              {user.role}
            </Badge>
            {protectedOwner ? (
              <Badge
                variant='outline'
                className='border-amber-300 bg-amber-100/50 text-amber-900'
              >
                Protected local owner
              </Badge>
            ) : null}
          </div>
        </SheetHeader>

        <div className='min-h-0 flex-1 overflow-y-auto px-4'>
          <dl data-testid='users-details-fields'>
            <DetailRow label='Username' value={formatValue(user.username)} />
            <DetailRow
              label='Name'
              value={formatValue(`${user.firstName} ${user.lastName}`.trim())}
            />
            <DetailRow label='Email' value={formatValue(user.email)} />
            <DetailRow
              label='Phone number'
              value={formatValue(user.phoneNumber)}
            />
            <DetailRow label='Status' value={formatValue(user.status)} />
            <DetailRow label='Role' value={formatValue(user.role)} />
            <DetailRow
              label='Invitation state'
              value={
                user.status === 'invited'
                  ? 'Invitation pending'
                  : 'No pending invitation'
              }
            />
            <DetailRow label='Created' value={formatValue(user.createdAt)} />
            <DetailRow
              label='Last active'
              value={
                user.status === 'active'
                  ? formatValue(user.updatedAt)
                  : 'Not provided'
              }
            />
            <DetailRow
              label='Last updated'
              value={formatValue(user.updatedAt)}
            />
          </dl>

          {cannotRemove || cannotDowngrade || !canDeactivate ? (
            <div
              className='mt-4 rounded-md border border-amber-300 bg-amber-100/40 p-3 text-sm text-amber-950'
              data-testid='users-protection-message'
            >
              <div className='flex items-start gap-2'>
                <ShieldAlert className='mt-0.5 h-4 w-4 shrink-0' />
                <p>
                  Protected admin safeguards prevent removing or degrading the
                  last active admin for this workspace.
                </p>
              </div>
            </div>
          ) : null}
        </div>

        <SheetFooter className='border-t sm:flex-col'>
          <div className='grid gap-2 sm:grid-cols-2'>
            <Button
              type='button'
              variant='outline'
              onClick={openEdit}
              data-testid='users-details-edit-action'
            >
              <UserPen className='h-4 w-4' />
              Edit role
            </Button>
            <Button
              type='button'
              variant='outline'
              disabled={user.status === 'active' && !canDeactivate}
              onClick={() => void updateUser({ status: statusAction })}
              data-testid='users-details-status-action'
            >
              {statusLabel}
            </Button>
            <Button
              type='button'
              variant='outline'
              disabled={user.status !== 'invited'}
              onClick={() => void resendInvitation()}
              data-testid='users-details-resend-action'
            >
              <Mail className='h-4 w-4' />
              Resend invitation
            </Button>
            <Button
              type='button'
              variant='destructive'
              disabled={cannotRemove}
              onClick={openDelete}
              data-testid='users-details-delete-action'
            >
              <Trash2 className='h-4 w-4' />
              Remove user
            </Button>
          </div>
          {cannotRemove ? (
            <p className='text-xs text-muted-foreground'>
              Remove user is unavailable for the protected owner or last active
              admin.
            </p>
          ) : null}
        </SheetFooter>
      </SheetContent>
    </Sheet>
  )
}
