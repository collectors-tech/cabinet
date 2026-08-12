import { type ColumnDef } from '@tanstack/react-table'
import { BarcodeIcon, ImageIcon, RotateCcwIcon, TagsIcon } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { DataTableColumnHeader } from '@/components/data-table'
import { labels, priorities, statuses } from '../data/data'
import { type Task } from '../data/schema'
import { DataTableRowActions } from './data-table-row-actions'
import {
  WishlistCostCell,
  WishlistNumberCell,
  WishlistPricePaidCell,
  WishlistPriceTrendCell,
  WishlistPriorityCell,
  WishlistPurchasedCell,
  type WishlistInlineChanges,
} from './tasks-column-cells'
import { formatMoney, formatWishlistDate } from './tasks-column-formatters'
import { WishlistThumbnail } from './wishlist-thumbnail'

export type TasksRoutePath =
  | '/_authenticated/inventory/'
  | '/_authenticated/wishlist/'

type TasksColumnsOptions = {
  routePath: TasksRoutePath
  onEditRow?: (task: Task) => void
  onPhotoRow?: (task: Task) => void
  onBarcodeRow?: (task: Task) => void
  onAssignCollectionRow?: (task: Task) => void
  onDeleteRow?: (task: Task) => void
  onRestoreRow?: (task: Task) => void
  onWishlistInlineUpdate?: (
    task: Task,
    changes: WishlistInlineChanges
  ) => Promise<void>
  onWishlistPurchaseRow?: (task: Task) => void
}
export function getTasksColumns({
  routePath,
  onEditRow,
  onPhotoRow,
  onBarcodeRow,
  onAssignCollectionRow,
  onDeleteRow,
  onRestoreRow,
  onWishlistInlineUpdate,
  onWishlistPurchaseRow,
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
      meta: { className: 'w-12' },
      enableSorting: false,
      enableHiding: false,
    },
    ...(isWishlistRoute
      ? []
      : [
          {
            accessorKey: 'id',
            header: ({ column }) => (
              <DataTableColumnHeader
                column={column}
                title={isInventoryRoute ? 'Part #' : 'Task'}
              />
            ),
            meta: isInventoryRoute
              ? {
                  className: 'w-[12rem] max-w-[12rem]',
                  tdClassName: 'max-w-0',
                }
              : undefined,
            cell: ({ row }) => (
              <span
                className='block max-w-full truncate'
                data-testid='inventory-row-part-number'
                title={String(row.getValue('id'))}
              >
                {row.getValue('id')}
              </span>
            ),
            enableSorting: false,
            enableHiding: false,
          } satisfies ColumnDef<Task>,
        ]),
    {
      accessorKey: 'title',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='Title' />
      ),
      meta: {
        className: isWishlistRoute
          ? 'ps-1 min-w-[12rem]'
          : 'ps-1 w-[28rem] max-w-[28rem]',
        tdClassName: 'ps-4',
      },
      cell: ({ row }) => {
        const label = labels.find((label) => label.value === row.original.label)

        return (
          <div className='flex min-w-0 items-center gap-2'>
            {isWishlistRoute ? <WishlistThumbnail task={row.original} /> : null}
            <div className='flex min-w-0 flex-col gap-1'>
              {!isInventoryRoute && !isWishlistRoute && label ? (
                <Badge variant='outline'>{label.label}</Badge>
              ) : null}
              <div className='flex min-w-0 space-x-2'>
                <span
                  className='block max-w-full min-w-0 truncate font-medium'
                  title={String(row.getValue('title'))}
                >
                  {row.getValue('title')}
                </span>
              </div>
              {isWishlistRoute && row.original.notes ? (
                <span className='truncate text-xs text-muted-foreground'>
                  {row.original.notes}
                </span>
              ) : null}
            </div>
          </div>
        )
      },
    },
    ...(isWishlistRoute
      ? [
          {
            accessorKey: 'status',
            header: 'Status',
            cell: ({ row }) => <span>{row.original.status}</span>,
            filterFn: (row, id, value) => {
              return value.includes(row.getValue(id))
            },
            enableHiding: true,
          } satisfies ColumnDef<Task>,
        ]
      : [
          {
            accessorKey: 'status',
            header: ({ column }) => (
              <DataTableColumnHeader
                column={column}
                title={isInventoryRoute ? 'Condition' : 'Status'}
              />
            ),
            meta: isInventoryRoute
              ? {
                  className: 'w-[13rem] max-w-[13rem] ps-1',
                  tdClassName: 'ps-4',
                }
              : { className: 'ps-1', tdClassName: 'ps-4' },
            cell: ({ row }) => {
              if (isInventoryRoute) {
                return (
                  <div className='flex min-w-0 items-center gap-2'>
                    <span className='block max-w-full truncate capitalize'>
                      {String(row.getValue('status'))}
                    </span>
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
          } satisfies ColumnDef<Task>,
        ]),
    ...(isWishlistRoute
      ? [
          {
            accessorKey: 'collectionName',
            header: 'Collection',
            cell: ({ row }) => (
              <span>{row.original.collectionName ?? 'Unassigned'}</span>
            ),
            filterFn: (row, id, value) => {
              if (Array.isArray(value) && value.includes('All Items')) {
                return true
              }
              return value.includes(row.getValue(id))
            },
            enableHiding: true,
          } satisfies ColumnDef<Task>,
        ]
      : []),
    ...(isWishlistRoute
      ? [
          {
            accessorKey: 'owned',
            header: ({ column }) => (
              <DataTableColumnHeader column={column} title='Purchased' />
            ),
            meta: {
              className: 'w-[7.5rem] min-w-[7.5rem] ps-1',
              tdClassName: 'ps-4 pe-3',
            },
            cell: ({ row }) => (
              <WishlistPurchasedCell
                task={row.original}
                onWishlistPurchaseRow={onWishlistPurchaseRow}
              />
            ),
            enableSorting: false,
          } satisfies ColumnDef<Task>,
          {
            accessorKey: 'label',
            header: ({ column }) => (
              <DataTableColumnHeader column={column} title='Category' />
            ),
            meta: { className: 'ps-1', tdClassName: 'ps-4' },
            cell: ({ row }) => (
              <span data-testid={`wishlist-category-${row.original.id}`}>
                {row.original.label || 'General'}
              </span>
            ),
          } satisfies ColumnDef<Task>,
          {
            accessorKey: 'pricePaid',
            header: ({ column }) => (
              <DataTableColumnHeader column={column} title='Price Paid' />
            ),
            meta: { className: 'ps-1', tdClassName: 'ps-4' },
            cell: ({ row }) => <WishlistPricePaidCell task={row.original} />,
          } satisfies ColumnDef<Task>,
          {
            accessorKey: 'marketPrice',
            header: ({ column }) => (
              <DataTableColumnHeader column={column} title='Market Price' />
            ),
            meta: { className: 'ps-1', tdClassName: 'ps-4' },
            cell: ({ row }) => (
              <span
                className='min-w-[76px] font-medium'
                data-testid={`wishlist-market-price-${row.original.id}`}
              >
                {formatMoney(row.original.marketPrice)}
              </span>
            ),
          } satisfies ColumnDef<Task>,
          {
            accessorKey: 'wishlistCreatedAt',
            header: ({ column }) => (
              <DataTableColumnHeader column={column} title='Date added' />
            ),
            meta: { className: 'ps-1', tdClassName: 'ps-4' },
            cell: ({ row }) => (
              <span
                className='text-sm whitespace-nowrap'
                data-testid={`wishlist-date-added-${row.original.id}`}
              >
                {formatWishlistDate(row.original.wishlistCreatedAt)}
              </span>
            ),
          } satisfies ColumnDef<Task>,
          {
            accessorKey: 'wishlistPriceUpdatedAt',
            header: ({ column }) => (
              <DataTableColumnHeader column={column} title='Updated' />
            ),
            meta: { className: 'ps-1', tdClassName: 'ps-4' },
            cell: ({ row }) => (
              <span
                className='text-sm whitespace-nowrap'
                data-testid={`wishlist-date-updated-${row.original.id}`}
                title='Latest pricing refresh date'
              >
                {formatWishlistDate(row.original.wishlistPriceUpdatedAt)}
              </span>
            ),
          } satisfies ColumnDef<Task>,
          {
            accessorKey: 'priceTrend',
            header: ({ column }) => (
              <DataTableColumnHeader column={column} title='Price Graph' />
            ),
            meta: { className: 'ps-1', tdClassName: 'ps-4' },
            cell: ({ row }) => <WishlistPriceTrendCell task={row.original} />,
            enableSorting: false,
          } satisfies ColumnDef<Task>,
          {
            accessorKey: 'targetPrice',
            header: ({ column }) => (
              <DataTableColumnHeader column={column} title='Cost' />
            ),
            meta: { className: 'ps-1', tdClassName: 'ps-4' },
            cell: ({ row }) => (
              <WishlistCostCell
                key={`${row.original.id}-${row.original.targetPrice ?? 0}`}
                task={row.original}
                onWishlistInlineUpdate={onWishlistInlineUpdate}
              />
            ),
          } satisfies ColumnDef<Task>,
          {
            accessorKey: 'quantity',
            header: ({ column }) => (
              <DataTableColumnHeader column={column} title='Qty' />
            ),
            meta: { className: 'ps-1', tdClassName: 'ps-4' },
            cell: ({ row }) => (
              <WishlistNumberCell
                key={`${row.original.id}-quantity-${row.original.quantity ?? 0}`}
                task={row.original}
                field='quantity'
                label='Quantity'
                value={row.original.quantity}
                minValue={0}
                onWishlistInlineUpdate={onWishlistInlineUpdate}
              />
            ),
          } satisfies ColumnDef<Task>,
          {
            accessorKey: 'neededQuantity',
            header: ({ column }) => (
              <DataTableColumnHeader column={column} title='Needs' />
            ),
            meta: { className: 'ps-1', tdClassName: 'ps-4' },
            cell: ({ row }) => (
              <WishlistNumberCell
                key={`${row.original.id}-needed-${row.original.neededQuantity ?? 1}`}
                task={row.original}
                field='neededQuantity'
                label='Needs'
                value={row.original.neededQuantity}
                minValue={1}
                onWishlistInlineUpdate={onWishlistInlineUpdate}
              />
            ),
          } satisfies ColumnDef<Task>,
        ]
      : []),
    ...(isInventoryRoute
      ? [
          {
            accessorKey: 'itemType',
            header: ({ column }) => (
              <DataTableColumnHeader column={column} title='Item type' />
            ),
            meta: {
              className: 'w-[10rem] max-w-[10rem] ps-1',
              tdClassName: 'ps-4',
            },
            cell: ({ row }) => (
              <span
                className='block max-w-full truncate'
                data-testid='inventory-row-item-type'
                title={row.original.itemType || 'Unclassified'}
              >
                {row.original.itemType || 'Unclassified'}
              </span>
            ),
            filterFn: (row, id, value) => {
              return value.includes(row.getValue(id))
            },
          } satisfies ColumnDef<Task>,
          {
            accessorKey: 'packagingGradeType',
            header: ({ column }) => (
              <DataTableColumnHeader column={column} title='Packaging' />
            ),
            meta: {
              className: 'w-[10rem] max-w-[10rem] ps-1',
              tdClassName: 'ps-4',
            },
            cell: ({ row }) => (
              <span
                className='block max-w-full truncate'
                data-testid='inventory-row-packaging-grade'
                title={row.original.packagingGradeType || 'Unset'}
              >
                {row.original.packagingGradeType || 'Unset'}
              </span>
            ),
            filterFn: (row, id, value) => {
              return value.includes(row.getValue(id))
            },
          } satisfies ColumnDef<Task>,
        ]
      : []),
    {
      accessorKey: 'priority',
      header: ({ column }) => (
        <DataTableColumnHeader
          column={column}
          title={isInventoryRoute ? 'Category' : 'Priority'}
        />
      ),
      meta: isInventoryRoute
        ? {
            className: 'w-[10rem] max-w-[10rem] ps-1',
            tdClassName: 'ps-3',
          }
        : { className: 'ps-1', tdClassName: 'ps-3' },
      cell: ({ row }) => {
        if (isInventoryRoute) {
          return (
            <span className='block max-w-full truncate'>
              {row.original.label || 'Uncategorized'}
            </span>
          )
        }

        if (isWishlistRoute) {
          return (
            <WishlistPriorityCell
              task={row.original}
              onWishlistInlineUpdate={onWishlistInlineUpdate}
            />
          )
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
          const categories = String(row.original.label || '')
            .split(',')
            .map((entry) => entry.trim())
            .filter(Boolean)
          return categories.some((category) => value.includes(category))
        }
        return value.includes(row.getValue(id))
      },
    },
    {
      id: 'actions',
      meta: {
        className: isInventoryRoute
          ? 'w-44'
          : isWishlistRoute
            ? 'sticky right-0 z-20 w-12 bg-background text-right'
            : undefined,
        tdClassName: isInventoryRoute
          ? 'max-w-none'
          : isWishlistRoute
            ? 'max-w-none'
            : undefined,
      },
      cell: ({ row }) => (
        <div className='flex items-center justify-end gap-1'>
          {isInventoryRoute ? (
            <>
              <Button
                type='button'
                variant='ghost'
                size='icon'
                className='h-8 w-8'
                data-testid='inventory-row-assign-collection-action'
                aria-label={`Assign ${row.original.title} to a collection`}
                onClick={(event) => {
                  event.stopPropagation()
                  onAssignCollectionRow?.(row.original)
                }}
              >
                <TagsIcon className='h-4 w-4' />
              </Button>
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
          {isWishlistRoute && row.original.deleted ? (
            <Button
              type='button'
              variant='outline'
              size='sm'
              className='h-8 px-2'
              aria-label={`Restore ${row.original.title}`}
              onClick={(event) => {
                event.stopPropagation()
                onRestoreRow?.(row.original)
              }}
            >
              <RotateCcwIcon className='mr-1 h-4 w-4' />
              Restore
            </Button>
          ) : null}
          <DataTableRowActions
            row={row}
            routePath={routePath}
            onEditRow={onEditRow}
            onDeleteRow={onDeleteRow}
          />
        </div>
      ),
    },
  ]
}
