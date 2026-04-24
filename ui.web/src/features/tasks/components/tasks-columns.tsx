import { type ColumnDef } from '@tanstack/react-table'
import { BarcodeIcon, ImageIcon } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { DataTableColumnHeader } from '@/components/data-table'
import { labels, priorities, statuses } from '../data/data'
import { type Task } from '../data/schema'
import { DataTableRowActions } from './data-table-row-actions'

export type TasksRoutePath =
  | '/_authenticated/inventory/'
  | '/_authenticated/wishlist/'

type TasksColumnsOptions = {
  routePath: TasksRoutePath
  onEditRow?: (task: Task) => void
  onPhotoRow?: (task: Task) => void
  onBarcodeRow?: (task: Task) => void
  onDeleteRow?: (task: Task) => void
  onWishlistMarkOwned?: (task: Task) => Promise<void>
  wishlistActionItemID?: string | null
}

const wishlistStatusLabels: Record<string, string> = {
  wishlist: 'Watching',
  discovered: 'Below target',
}

function formatWishlistStatus(status: string) {
  return wishlistStatusLabels[status] ?? status
}

export function getTasksColumns({
  routePath,
  onEditRow,
  onPhotoRow,
  onBarcodeRow,
  onDeleteRow,
  onWishlistMarkOwned,
  wishlistActionItemID,
}: TasksColumnsOptions): ColumnDef<Task>[] {
  const isInventoryRoute = routePath === '/_authenticated/inventory/'
  const isWishlistRoute = routePath === '/_authenticated/wishlist/'

  return [
    {
      id: 'select',
      header: ({ table }) => (
        <Checkbox
          checked={
            table.getIsAllPageRowsSelected() ||
            (table.getIsSomePageRowsSelected() && 'indeterminate')
          }
          onCheckedChange={(value) => table.toggleAllPageRowsSelected(!!value)}
          aria-label='Select all'
          className='translate-y-[2px]'
        />
      ),
      cell: ({ row }) => (
        <Checkbox
          checked={row.getIsSelected()}
          onCheckedChange={(value) => row.toggleSelected(!!value)}
          aria-label='Select row'
          className='translate-y-[2px]'
        />
      ),
      enableSorting: false,
      enableHiding: false,
    },
    {
      accessorKey: 'id',
      header: ({ column }) => (
        <DataTableColumnHeader
          column={column}
          title={
            isInventoryRoute ? 'Part #' : isWishlistRoute ? 'Item ID' : 'Task'
          }
        />
      ),
      cell: ({ row }) => <div className='w-[120px]'>{row.getValue('id')}</div>,
      enableSorting: false,
      enableHiding: false,
    },
    {
      accessorKey: 'title',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='Title' />
      ),
      meta: {
        className: 'ps-1 max-w-0 w-2/3',
        tdClassName: 'ps-4',
      },
      cell: ({ row }) => {
        const label = labels.find((label) => label.value === row.original.label)

        return (
          <div className='flex min-w-0 flex-col gap-1'>
            {!isInventoryRoute && !isWishlistRoute && label ? (
              <Badge variant='outline'>{label.label}</Badge>
            ) : null}
            <div className='flex space-x-2'>
              <span className='truncate font-medium'>
                {row.getValue('title')}
              </span>
            </div>
            {isWishlistRoute && row.original.notes ? (
              <span className='truncate text-xs text-muted-foreground'>
                {row.original.notes}
              </span>
            ) : null}
          </div>
        )
      },
    },
    {
      accessorKey: 'status',
      header: ({ column }) => (
        <DataTableColumnHeader
          column={column}
          title={
            isInventoryRoute
              ? 'Condition'
              : isWishlistRoute
                ? 'Watch Status'
                : 'Status'
          }
        />
      ),
      meta: { className: 'ps-1', tdClassName: 'ps-4' },
      cell: ({ row }) => {
        if (isInventoryRoute) {
          return (
            <div className='flex min-w-[100px] items-center gap-2'>
              <span className='capitalize'>
                {String(row.getValue('status'))}
              </span>
            </div>
          )
        }

        if (isWishlistRoute) {
          return (
            <div className='flex min-w-[120px] items-center gap-2'>
              <span>{formatWishlistStatus(row.original.status)}</span>
            </div>
          )
        }

        const status = statuses.find(
          (status) => status.value === row.getValue('status')
        )

        if (!status) {
          return null
        }

        return (
          <div className='flex w-[100px] items-center gap-2'>
            {status.icon && (
              <status.icon className='size-4 text-muted-foreground' />
            )}
            <span>{status.label}</span>
          </div>
        )
      },
      filterFn: (row, id, value) => {
        return value.includes(row.getValue(id))
      },
    },
    {
      accessorKey: 'priority',
      header: ({ column }) => (
        <DataTableColumnHeader
          column={column}
          title={
            isInventoryRoute
              ? 'Category'
              : isWishlistRoute
                ? 'Target Priority'
                : 'Priority'
          }
        />
      ),
      meta: { className: 'ps-1', tdClassName: 'ps-3' },
      cell: ({ row }) => {
        if (isInventoryRoute) {
          return <span>{row.original.label || 'Uncategorized'}</span>
        }

        const priority = priorities.find(
          (priority) => priority.value === row.getValue('priority')
        )

        if (!priority) {
          return null
        }

        return (
          <div className='flex items-center gap-2'>
            {priority.icon && (
              <priority.icon className='size-4 text-muted-foreground' />
            )}
            <span>{priority.label}</span>
          </div>
        )
      },
      filterFn: (row, id, value) => {
        if (isInventoryRoute) {
          return value.includes(row.original.label)
        }
        return value.includes(row.getValue(id))
      },
    },
    {
      id: 'actions',
      cell: ({ row }) => (
        <div className='flex items-center justify-end gap-1'>
          {isInventoryRoute ? (
            <>
              <Button
                type='button'
                variant='ghost'
                size='icon'
                className='h-8 w-8'
                data-testid='inventory-row-barcodes-action'
                aria-label={`Open barcodes for ${row.original.title}`}
                onClick={(event) => {
                  event.stopPropagation()
                  onBarcodeRow?.(row.original)
                }}
              >
                <BarcodeIcon className='h-4 w-4' />
              </Button>
              <Button
                type='button'
                variant='ghost'
                size='icon'
                className='h-8 w-8'
                data-testid='inventory-row-photos-action'
                aria-label={`Open photos for ${row.original.title}`}
                onClick={(event) => {
                  event.stopPropagation()
                  onPhotoRow?.(row.original)
                }}
              >
                <ImageIcon className='h-4 w-4' />
              </Button>
            </>
          ) : null}
          <DataTableRowActions
            row={row}
            routePath={routePath}
            onEditRow={onEditRow}
            onDeleteRow={onDeleteRow}
            onWishlistMarkOwned={onWishlistMarkOwned}
            wishlistActionItemID={wishlistActionItemID}
          />
        </div>
      ),
    },
  ]
}
