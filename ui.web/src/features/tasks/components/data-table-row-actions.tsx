import { DotsHorizontalIcon } from '@radix-ui/react-icons'
import { type Row } from '@tanstack/react-table'
import { Trash2 } from 'lucide-react'
import { useState } from 'react'
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

  return (
    <DropdownMenu modal={false} open={open} onOpenChange={setOpen}>
      <DropdownMenuTrigger asChild>
        <Button
          type='button'
          variant='ghost'
          className='flex h-8 w-8 p-0 data-[state=open]:bg-muted'
          data-testid='task-row-actions-trigger'
          onPointerDown={(event) => {
            event.preventDefault()
            event.stopPropagation()
            setOpen(true)
          }}
          onMouseDown={(event) => {
            event.preventDefault()
            event.stopPropagation()
            setOpen(true)
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
        onClick={(event) => {
          event.stopPropagation()
        }}
      >
        <DropdownMenuItem
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
          onSelect={(event) => {
            event.stopPropagation()
            onDeleteRow?.(task)
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
