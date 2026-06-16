import { useState } from 'react'
import { DotsHorizontalIcon } from '@radix-ui/react-icons'
import { type Row } from '@tanstack/react-table'
import { Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuShortcut,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { type Task, taskSchema } from '../data/schema'

type DataTableRowActionsProps<TData> = {
  row: Row<TData>
  routePath: '/_authenticated/inventory/' | '/_authenticated/wishlist/'
  onEditRow?: (task: Task) => void
  onDeleteRow?: (task: Task) => void
}

export function DataTableRowActions<TData>({
  row,
  routePath,
  onEditRow,
  onDeleteRow,
}: DataTableRowActionsProps<TData>) {
  const task = taskSchema.parse(row.original)
  const isWishlistRoute = routePath === '/_authenticated/wishlist/'
  const [open, setOpen] = useState(false)
  const rowActionId = task.itemID?.trim() || task.id
  const handleDeleteRow = () => {
    onDeleteRow?.(task)
  }

  return (
    <DropdownMenu modal={false} open={open} onOpenChange={setOpen}>
      <DropdownMenuTrigger asChild>
        <Button
          type='button'
          variant='ghost'
          className='flex h-8 w-8 p-0 data-[state=open]:bg-muted'
          data-testid='task-row-actions-trigger'
          data-row-id={rowActionId}
          aria-label={`Open actions for ${task.title}`}
          onPointerDown={(event) => {
            event.stopPropagation()
          }}
          onClick={(event) => {
            event.stopPropagation()
            setOpen(true)
          }}
        >
          <DotsHorizontalIcon className='h-4 w-4' />
          <span className='sr-only'>Open menu</span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent
        align='end'
        className='w-[160px]'
        data-testid='task-row-actions-menu'
        data-row-id={rowActionId}
        onPointerDown={(event) => {
          event.stopPropagation()
        }}
        onClick={(event) => {
          event.stopPropagation()
        }}
      >
        <DropdownMenuItem
          data-testid='task-row-action-edit'
          data-row-id={rowActionId}
          onSelect={(event) => {
            event.stopPropagation()
            onEditRow?.(task)
          }}
        >
          Edit
        </DropdownMenuItem>
        {!isWishlistRoute ? (
          <>
            <DropdownMenuItem disabled>Make a copy</DropdownMenuItem>
            <DropdownMenuItem disabled>Favorite</DropdownMenuItem>
            <DropdownMenuSeparator />
          </>
        ) : null}
        <DropdownMenuItem
          data-testid='task-row-action-delete'
          data-row-id={rowActionId}
          onClick={(event) => {
            event.stopPropagation()
            handleDeleteRow()
          }}
          onSelect={(event) => {
            event.stopPropagation()
            handleDeleteRow()
          }}
        >
          Delete
          <DropdownMenuShortcut>
            <Trash2 size={16} />
          </DropdownMenuShortcut>
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
