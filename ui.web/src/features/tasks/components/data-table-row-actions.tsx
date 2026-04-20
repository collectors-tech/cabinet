import { DotsHorizontalIcon } from '@radix-ui/react-icons'
import { type Row } from '@tanstack/react-table'
import { Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuSeparator,
  DropdownMenuShortcut,
  DropdownMenuSub,
  DropdownMenuSubContent,
  DropdownMenuSubTrigger,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { labels } from '../data/data'
import { type Task, taskSchema } from '../data/schema'
import { useTasks } from './tasks-provider'

type DataTableRowActionsProps<TData> = {
  row: Row<TData>
  routePath: '/_authenticated/inventory/' | '/_authenticated/wishlist/'
  onWishlistMarkOwned?: (task: Task) => Promise<void>
  wishlistActionItemID?: string | null
}

export function DataTableRowActions<TData>({
  row,
  routePath,
  onWishlistMarkOwned,
  wishlistActionItemID,
}: DataTableRowActionsProps<TData>) {
  const task = taskSchema.parse(row.original)
  const isWishlistRoute = routePath === '/_authenticated/wishlist/'
  const isWishlistActionPending = wishlistActionItemID === task.id

  const { setOpen, setCurrentRow } = useTasks()

  return (
    <DropdownMenu modal={false}>
      <DropdownMenuTrigger asChild>
        <Button
          variant='ghost'
          className='flex h-8 w-8 p-0 data-[state=open]:bg-muted'
          data-testid='task-row-actions-trigger'
        >
          <DotsHorizontalIcon className='h-4 w-4' />
          <span className='sr-only'>Open menu</span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align='end' className='w-[160px]'>
        {isWishlistRoute ? (
          <>
            <DropdownMenuItem
              data-testid='wishlist-mark-owned-action'
              disabled={!task.wishlistEntryID || isWishlistActionPending}
              onClick={() => {
                if (onWishlistMarkOwned) {
                  void onWishlistMarkOwned(task)
                }
              }}
            >
              {isWishlistActionPending ? 'Moving...' : 'Mark owned'}
            </DropdownMenuItem>
            <DropdownMenuSeparator />
          </>
        ) : null}
        <DropdownMenuItem
          onClick={() => {
            setCurrentRow(task)
            setOpen('update')
          }}
        >
          Edit
        </DropdownMenuItem>
        <DropdownMenuItem disabled>Make a copy</DropdownMenuItem>
        <DropdownMenuItem disabled>Favorite</DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuSub>
          <DropdownMenuSubTrigger>Labels</DropdownMenuSubTrigger>
          <DropdownMenuSubContent>
            <DropdownMenuRadioGroup value={task.label}>
              {labels.map((label) => (
                <DropdownMenuRadioItem key={label.value} value={label.value}>
                  {label.label}
                </DropdownMenuRadioItem>
              ))}
            </DropdownMenuRadioGroup>
          </DropdownMenuSubContent>
        </DropdownMenuSub>
        <DropdownMenuSeparator />
        <DropdownMenuItem
          onClick={() => {
            setCurrentRow(task)
            setOpen('delete')
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
