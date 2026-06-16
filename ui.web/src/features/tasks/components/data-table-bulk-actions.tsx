import { type ComponentType, useState } from 'react'
import { type Table } from '@tanstack/react-table'
import { ArrowUpDown, CircleArrowUp, Download, Trash2 } from 'lucide-react'
import { toast } from 'sonner'
import { sleep } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { DataTableBulkActions as BulkActionsToolbar } from '@/components/data-table'
import { priorities, statuses } from '../data/data'
import { type Task } from '../data/schema'
import { TasksMultiDeleteDialog } from './tasks-multi-delete-dialog'

interface DataTableBulkActionsProps<TData> {
  table: Table<TData>
  routePath: '/_authenticated/inventory/' | '/_authenticated/wishlist/'
  onWishlistBulkStatusChange?: (tasks: Task[], status: string) => Promise<void>
  onWishlistBulkPriorityChange?: (
    tasks: Task[],
    priority: string
  ) => Promise<void>
  onWishlistBulkDelete?: (tasks: Task[]) => Promise<void>
  onWishlistExport?: (tasks: Task[]) => void
  isWishlistMutating?: boolean
}

export function DataTableBulkActions<TData>({
  table,
  routePath,
  onWishlistBulkStatusChange,
  onWishlistBulkPriorityChange,
  onWishlistBulkDelete,
  onWishlistExport,
  isWishlistMutating = false,
}: DataTableBulkActionsProps<TData>) {
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)
  const selectedRows = table.getFilteredSelectedRowModel().rows
  const selectedTasks = selectedRows.map((row) => row.original as Task)
  const isWishlistRoute = routePath === '/_authenticated/wishlist/'
  const wishlistStatusOptions = [
    { label: 'Watching', value: 'wishlist' },
    { label: 'Below target', value: 'discovered' },
  ]

  const handleBulkStatusChange = (status: string) => {
    if (isWishlistRoute && onWishlistBulkStatusChange) {
      void onWishlistBulkStatusChange(selectedTasks, status)
      return
    }

    toast.promise(sleep(2000), {
      loading: 'Updating status...',
      success: () => {
        table.resetRowSelection()
        return `Status updated to "${status}" for ${selectedTasks.length} task${selectedTasks.length > 1 ? 's' : ''}.`
      },
      error: 'Error',
    })
    table.resetRowSelection()
  }

  const handleBulkPriorityChange = (priority: string) => {
    if (isWishlistRoute && onWishlistBulkPriorityChange) {
      void onWishlistBulkPriorityChange(selectedTasks, priority)
      return
    }

    toast.promise(sleep(2000), {
      loading: 'Updating priority...',
      success: () => {
        table.resetRowSelection()
        return `Priority updated to "${priority}" for ${selectedTasks.length} task${selectedTasks.length > 1 ? 's' : ''}.`
      },
      error: 'Error',
    })
    table.resetRowSelection()
  }

  const handleBulkExport = () => {
    if (isWishlistRoute && onWishlistExport) {
      onWishlistExport(selectedTasks)
      table.resetRowSelection()
      return
    }

    toast.promise(sleep(2000), {
      loading: 'Exporting tasks...',
      success: () => {
        table.resetRowSelection()
        return `Exported ${selectedTasks.length} task${selectedTasks.length > 1 ? 's' : ''} to CSV.`
      },
      error: 'Error',
    })
    table.resetRowSelection()
  }

  return (
    <>
      <BulkActionsToolbar
        table={table}
        entityName={isWishlistRoute ? 'wishlist entry' : 'task'}
      >
        <DropdownMenu>
          <Tooltip>
            <TooltipTrigger asChild>
              <DropdownMenuTrigger asChild>
                <Button
                  variant='outline'
                  size='icon'
                  className='size-8'
                  aria-label='Update status'
                  title='Update status'
                  disabled={isWishlistMutating}
                >
                  <CircleArrowUp />
                  <span className='sr-only'>Update status</span>
                </Button>
              </DropdownMenuTrigger>
            </TooltipTrigger>
            <TooltipContent>
              <p>Update status</p>
            </TooltipContent>
          </Tooltip>
          <DropdownMenuContent sideOffset={14}>
            {(isWishlistRoute ? wishlistStatusOptions : statuses).map(
              (status) => {
                const StatusIcon =
                  'icon' in status
                    ? (status.icon as ComponentType<{ className?: string }>)
                    : null

                return (
                  <DropdownMenuItem
                    key={status.value}
                    defaultValue={status.value}
                    onClick={() => handleBulkStatusChange(status.value)}
                  >
                    {StatusIcon ? (
                      <StatusIcon className='size-4 text-muted-foreground' />
                    ) : null}
                    {status.label}
                  </DropdownMenuItem>
                )
              }
            )}
          </DropdownMenuContent>
        </DropdownMenu>

        <DropdownMenu>
          <Tooltip>
            <TooltipTrigger asChild>
              <DropdownMenuTrigger asChild>
                <Button
                  variant='outline'
                  size='icon'
                  className='size-8'
                  aria-label='Update priority'
                  title='Update priority'
                  disabled={isWishlistMutating}
                >
                  <ArrowUpDown />
                  <span className='sr-only'>Update priority</span>
                </Button>
              </DropdownMenuTrigger>
            </TooltipTrigger>
            <TooltipContent>
              <p>Update priority</p>
            </TooltipContent>
          </Tooltip>
          <DropdownMenuContent sideOffset={14}>
            {priorities.map((priority) => (
              <DropdownMenuItem
                key={priority.value}
                defaultValue={priority.value}
                onClick={() => handleBulkPriorityChange(priority.value)}
              >
                {priority.icon && (
                  <priority.icon className='size-4 text-muted-foreground' />
                )}
                {priority.label}
              </DropdownMenuItem>
            ))}
          </DropdownMenuContent>
        </DropdownMenu>

        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant='outline'
              size='icon'
              onClick={() => handleBulkExport()}
              className='size-8'
              aria-label={
                isWishlistRoute ? 'Export wishlist entries' : 'Export tasks'
              }
              title={
                isWishlistRoute ? 'Export wishlist entries' : 'Export tasks'
              }
              disabled={isWishlistMutating}
            >
              <Download />
              <span className='sr-only'>
                {isWishlistRoute ? 'Export wishlist entries' : 'Export tasks'}
              </span>
            </Button>
          </TooltipTrigger>
          <TooltipContent>
            <p>
              {isWishlistRoute ? 'Export wishlist entries' : 'Export tasks'}
            </p>
          </TooltipContent>
        </Tooltip>

        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant='destructive'
              size='icon'
              onClick={() => setShowDeleteConfirm(true)}
              className='size-8'
              aria-label={
                isWishlistRoute
                  ? 'Delete selected wishlist entries'
                  : 'Delete selected tasks'
              }
              title={
                isWishlistRoute
                  ? 'Delete selected wishlist entries'
                  : 'Delete selected tasks'
              }
              disabled={isWishlistMutating}
            >
              <Trash2 />
              <span className='sr-only'>
                {isWishlistRoute
                  ? 'Delete selected wishlist entries'
                  : 'Delete selected tasks'}
              </span>
            </Button>
          </TooltipTrigger>
          <TooltipContent>
            <p>
              {isWishlistRoute
                ? 'Delete selected wishlist entries'
                : 'Delete selected tasks'}
            </p>
          </TooltipContent>
        </Tooltip>
      </BulkActionsToolbar>

      <TasksMultiDeleteDialog
        open={showDeleteConfirm}
        onOpenChange={setShowDeleteConfirm}
        table={table}
        routePath={routePath}
        onWishlistBulkDelete={onWishlistBulkDelete}
        isLoading={isWishlistMutating}
      />
    </>
  )
}
