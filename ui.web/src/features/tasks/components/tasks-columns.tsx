import { useEffect, useState } from 'react'
import { type ColumnDef } from '@tanstack/react-table'
import {
  BarcodeIcon,
  CheckIcon,
  ImageIcon,
  MinusIcon,
  PlusIcon,
  TagsIcon,
  TrendingDownIcon,
  TrendingUpIcon,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Input } from '@/components/ui/input'
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
  onAssignCollectionRow?: (task: Task) => void
  onDeleteRow?: (task: Task) => void
  onWishlistInlineUpdate?: (
    task: Task,
    changes: WishlistInlineChanges
  ) => Promise<void>
  onWishlistPurchaseRow?: (task: Task) => void
}

type WishlistInlineChanges = {
  targetPrice?: number
  priority?: string
  owned?: boolean
  pricePaid?: number
  purchaseUrl?: string
  purchaseDate?: string
  purchaseCondition?: string
  quantity?: number
  neededQuantity?: number
}

function formatMoney(value: number | undefined) {
  if (typeof value !== 'number' || value <= 0) {
    return '-'
  }
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
  }).format(value)
}

function formatCostDraft(value: number | undefined) {
  if (typeof value !== 'number' || value <= 0) {
    return ''
  }
  return Number.isInteger(value) ? String(value) : value.toFixed(2)
}

function buildSparklinePoints(values: number[]) {
  const width = 88
  const height = 28
  const padding = 3
  const min = Math.min(...values)
  const max = Math.max(...values)
  const range = max - min || 1
  const step =
    values.length > 1 ? (width - padding * 2) / (values.length - 1) : 0

  return values
    .map((value, index) => {
      const x = padding + index * step
      const y =
        height - padding - ((value - min) / range) * (height - padding * 2)
      return `${x.toFixed(1)},${y.toFixed(1)}`
    })
    .join(' ')
}

function WishlistPriceSparkline({
  task,
  label,
}: {
  task: Task
  label: string
}) {
  const values = (task.priceHistory ?? []).filter(
    (value) => typeof value === 'number' && Number.isFinite(value)
  )

  if (values.length < 2) {
    return null
  }
  const sampleCount = task.priceSampleCount ?? values.length
  const dateRange =
    task.priceFirstDate && task.priceLatestDate
      ? `${task.priceFirstDate} to ${task.priceLatestDate}`
      : 'date range unavailable'
  const accessibleLabel = `${label}: ${sampleCount} price points, ${dateRange}, latest ${formatMoney(task.marketPrice)}`

  return (
    <svg
      viewBox='0 0 88 28'
      role='img'
      aria-label={accessibleLabel}
      className='h-7 w-[88px] rounded bg-slate-950/60'
      data-testid={`wishlist-price-sparkline-${task.id}`}
      preserveAspectRatio='none'
    >
      <polyline
        points={buildSparklinePoints(values)}
        fill='none'
        stroke='rgb(73 103 255)'
        strokeWidth='2.5'
        strokeLinecap='round'
        strokeLinejoin='round'
      />
    </svg>
  )
}

function WishlistPriceTrendCell({ task }: { task: Task }) {
  const trend = task.priceTrend ?? 'unknown'
  const trendConfig = {
    up: {
      label: 'Price trending up',
      icon: TrendingUpIcon,
      className: 'text-red-300',
    },
    steady: {
      label: 'Price steady',
      icon: MinusIcon,
      className: 'text-muted-foreground',
    },
    down: {
      label: 'Price trending down',
      icon: TrendingDownIcon,
      className: 'text-emerald-300',
    },
    unknown: {
      label: 'Price trend unknown',
      icon: MinusIcon,
      className: 'text-muted-foreground',
    },
  }[trend]
  const Icon = trendConfig.icon
  const hasHistory = (task.priceHistory ?? []).length >= 2
  const sampleCount = task.priceSampleCount ?? task.priceHistory?.length ?? 0
  const sourceText =
    task.priceSources && task.priceSources.length > 0
      ? task.priceSources.join(', ')
      : 'No source yet'
  const dateText =
    task.priceFirstDate && task.priceLatestDate
      ? `${task.priceFirstDate} - ${task.priceLatestDate}`
      : 'Awaiting pricing history'
  const stockText =
    typeof task.priceStockCount === 'number'
      ? `${task.priceStockCount} available`
      : 'stock unknown'

  return (
    <div
      className='flex min-w-[11rem] items-center gap-2'
      data-testid={`wishlist-price-trend-${task.id}`}
      aria-label={trendConfig.label}
      title={`${trendConfig.label}. ${sampleCount} points. ${dateText}. Sources: ${sourceText}. ${stockText}.`}
    >
      <Icon className={`size-4 ${trendConfig.className}`} />
      {hasHistory ? (
        <WishlistPriceSparkline task={task} label={trendConfig.label} />
      ) : (
        <span
          className='text-xs text-muted-foreground'
          data-testid={`wishlist-price-graph-empty-${task.id}`}
        >
          No history
        </span>
      )}
      <span
        className='sr-only'
        data-testid={`wishlist-price-graph-meta-${task.id}`}
      >
        {`${sampleCount} points ${dateText} ${sourceText} ${stockText}`}
      </span>
    </div>
  )
}

function WishlistCostCell({
  task,
  onWishlistInlineUpdate,
}: {
  task: Task
  onWishlistInlineUpdate?: (
    task: Task,
    changes: WishlistInlineChanges
  ) => Promise<void>
}) {
  const [draft, setDraft] = useState(formatCostDraft(task.targetPrice))

  useEffect(() => {
    setDraft(formatCostDraft(task.targetPrice))
  }, [task.targetPrice])

  const persist = async () => {
    const trimmed = draft.trim()
    const nextValue = trimmed === '' ? 0 : Number(trimmed)
    if (Number.isNaN(nextValue) || nextValue < 0) {
      setDraft(formatCostDraft(task.targetPrice))
      return
    }
    if ((task.targetPrice ?? 0) === nextValue) {
      return
    }
    await onWishlistInlineUpdate?.(task, { targetPrice: nextValue })
  }

  return (
    <Input
      type='number'
      min='0'
      step='0.01'
      inputMode='decimal'
      value={draft}
      data-testid={`wishlist-cost-input-${task.id}`}
      aria-label={`Cost for ${task.title}`}
      className='h-8 w-[6.5rem]'
      onClick={(event) => event.stopPropagation()}
      onDoubleClick={(event) => event.stopPropagation()}
      onChange={(event) => setDraft(event.target.value)}
      onBlur={() => {
        void persist()
      }}
      onKeyDown={(event) => {
        if (event.key === 'Enter') {
          event.preventDefault()
          event.currentTarget.blur()
        }
      }}
    />
  )
}

function WishlistPriorityCell({
  task,
  onWishlistInlineUpdate,
}: {
  task: Task
  onWishlistInlineUpdate?: (
    task: Task,
    changes: WishlistInlineChanges
  ) => Promise<void>
}) {
  const persistPriority = (nextPriority: string) => {
    if (nextPriority === task.priority) {
      return
    }
    void onWishlistInlineUpdate?.(task, { priority: nextPriority })
  }

  return (
    <select
      className='h-8 min-w-[6.75rem] rounded-md border bg-background px-2 text-sm'
      value={task.priority}
      data-testid={`wishlist-priority-select-${task.id}`}
      aria-label={`Priority for ${task.title}`}
      onClick={(event) => event.stopPropagation()}
      onDoubleClick={(event) => event.stopPropagation()}
      onChange={(event) => {
        persistPriority(event.target.value)
      }}
    >
      {priorities.map((priority) => (
        <option key={priority.value} value={priority.value}>
          {priority.label}
        </option>
      ))}
    </select>
  )
}

function WishlistOwnedCell({
  task,
  onWishlistInlineUpdate,
  onWishlistPurchaseRow,
}: {
  task: Task
  onWishlistInlineUpdate?: (
    task: Task,
    changes: WishlistInlineChanges
  ) => Promise<void>
  onWishlistPurchaseRow?: (task: Task) => void
}) {
  const owned = Boolean(task.owned)

  return (
    <div
      className='flex min-w-[5rem] items-center gap-1'
      onClick={(event) => event.stopPropagation()}
      onDoubleClick={(event) => event.stopPropagation()}
    >
      <button
        type='button'
        role='checkbox'
        aria-checked={owned}
        data-state={owned ? 'checked' : 'unchecked'}
        data-testid={`wishlist-owned-checkbox-${task.id}`}
        aria-label={`Owned status for ${task.title}`}
        className={cn(
          'relative flex h-7 w-12 items-center justify-center rounded-md border transition-colors',
          owned
            ? 'border-slate-700 bg-slate-950 shadow-inner'
            : 'border-border bg-background hover:bg-accent/40'
        )}
        onClick={() => {
          void onWishlistInlineUpdate?.(task, { owned: !owned })
        }}
      >
        <CheckIcon
          aria-hidden='true'
          className={cn(
            'h-4 w-4 transition-opacity',
            owned ? 'opacity-100' : 'opacity-0'
          )}
          data-testid={`wishlist-owned-tick-${task.id}`}
        />
      </button>
      <Button
        type='button'
        variant='outline'
        size='icon'
        className='h-8 w-8'
        data-testid={`wishlist-purchase-open-${task.id}`}
        aria-label={`Add purchase details for ${task.title}`}
        onClick={() => onWishlistPurchaseRow?.(task)}
      >
        <PlusIcon className='h-4 w-4' />
      </Button>
    </div>
  )
}

function WishlistPricePaidCell({ task }: { task: Task }) {
  return (
    <span
      className='min-w-[76px] font-medium'
      data-testid={`wishlist-price-paid-value-${task.id}`}
    >
      {formatMoney(task.pricePaid)}
    </span>
  )
}

function WishlistNumberCell({
  task,
  field,
  label,
  value,
  onWishlistInlineUpdate,
}: {
  task: Task
  field: 'quantity' | 'neededQuantity'
  label: string
  value: number | undefined
  onWishlistInlineUpdate?: (
    task: Task,
    changes: WishlistInlineChanges
  ) => Promise<void>
}) {
  const [draft, setDraft] = useState(String(value ?? 0))

  useEffect(() => {
    setDraft(String(value ?? 0))
  }, [value])

  const persist = async () => {
    const parsed = Number(draft.trim())
    if (!Number.isInteger(parsed) || parsed < 0) {
      setDraft(String(value ?? 0))
      return
    }
    if ((value ?? 0) === parsed) {
      return
    }
    await onWishlistInlineUpdate?.(task, { [field]: parsed })
  }

  return (
    <Input
      type='number'
      min='0'
      step='1'
      inputMode='numeric'
      value={draft}
      data-testid={
        field === 'quantity'
          ? `wishlist-qty-input-${task.id}`
          : `wishlist-needs-input-${task.id}`
      }
      aria-label={`${label} for ${task.title}`}
      className='h-8 w-[4.5rem]'
      onClick={(event) => event.stopPropagation()}
      onDoubleClick={(event) => event.stopPropagation()}
      onChange={(event) => setDraft(event.target.value)}
      onBlur={() => {
        void persist()
      }}
      onKeyDown={(event) => {
        if (event.key === 'Enter') {
          event.preventDefault()
          event.currentTarget.blur()
        }
      }}
    />
  )
}

export function getTasksColumns({
  routePath,
  onEditRow,
  onPhotoRow,
  onBarcodeRow,
  onAssignCollectionRow,
  onDeleteRow,
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
            cell: ({ row }) => (
              <div className='w-[120px]'>{row.getValue('id')}</div>
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
          : 'ps-1 max-w-0 w-2/3',
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
    ...(isWishlistRoute
      ? []
      : [
          {
            accessorKey: 'status',
            header: ({ column }) => (
              <DataTableColumnHeader
                column={column}
                title={isInventoryRoute ? 'Condition' : 'Status'}
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
            accessorKey: 'owned',
            header: ({ column }) => (
              <DataTableColumnHeader column={column} title='Owned' />
            ),
            meta: { className: 'ps-1', tdClassName: 'ps-4' },
            cell: ({ row }) => (
              <WishlistOwnedCell
                task={row.original}
                onWishlistInlineUpdate={onWishlistInlineUpdate}
                onWishlistPurchaseRow={onWishlistPurchaseRow}
              />
            ),
            enableSorting: false,
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
                task={row.original}
                field='quantity'
                label='Quantity'
                value={row.original.quantity}
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
                task={row.original}
                field='neededQuantity'
                label='Needs'
                value={row.original.neededQuantity}
                onWishlistInlineUpdate={onWishlistInlineUpdate}
              />
            ),
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
      meta: { className: 'ps-1', tdClassName: 'ps-3' },
      cell: ({ row }) => {
        if (isInventoryRoute) {
          return <span>{row.original.label || 'Uncategorized'}</span>
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
