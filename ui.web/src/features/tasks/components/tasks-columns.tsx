import { useEffect, useState } from 'react'
import { type ColumnDef } from '@tanstack/react-table'
import {
  BarcodeIcon,
  ImageIcon,
  MinusIcon,
  PlusIcon,
  TagsIcon,
} from 'lucide-react'
import { Line, LineChart, XAxis, YAxis } from 'recharts'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
  type ChartConfig,
} from '@/components/ui/chart'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
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

function formatWishlistDate(value: string | undefined) {
  const trimmed = value?.trim()
  if (!trimmed) {
    return '-'
  }
  const datePart = trimmed.split('T')[0]?.split(' ')[0] ?? trimmed
  const parts = datePart.split('-').map((part) => Number(part))
  if (
    parts.length === 3 &&
    Number.isInteger(parts[0]) &&
    Number.isInteger(parts[1]) &&
    Number.isInteger(parts[2])
  ) {
    const [year, month, day] = parts
    return new Intl.DateTimeFormat('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      timeZone: 'UTC',
    }).format(new Date(Date.UTC(year, month - 1, day)))
  }
  return trimmed
}

function formatCostDraft(value: number | undefined) {
  if (typeof value !== 'number' || value < 0) {
    return ''
  }
  return Number.isInteger(value) ? String(value) : value.toFixed(2)
}

const wishlistStepperInputClassName =
  'h-8 border-x-0 rounded-none text-center [appearance:textfield] [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none'

function buildWishlistPricePointRows(task: Task, values: number[]) {
  const dates = task.priceHistoryDates ?? []
  return values
    .map((price, index) => ({
      date: dates[index] ?? `Point ${index + 1}`,
      price,
    }))
    .slice(-10)
}

const wishlistPriceChartConfig = {
  price: {
    label: 'Price',
    color: 'rgb(73 103 255)',
  },
} satisfies ChartConfig

function WishlistPriceSparkline({
  task,
  label,
}: {
  task: Task
  label: string
}) {
  const [isDialogOpen, setIsDialogOpen] = useState(false)
  const values = (task.priceHistory ?? []).filter(
    (value) => typeof value === 'number' && Number.isFinite(value)
  )

  if (values.length < 2) {
    return null
  }
  const latestPointRows = buildWishlistPricePointRows(task, values)
  const chartRows = values.map((price, index) => ({
    date: task.priceHistoryDates?.[index] ?? `Point ${index + 1}`,
    price,
  }))
  const sampleCount = task.priceSampleCount ?? values.length
  const dateRange =
    task.priceFirstDate && task.priceLatestDate
      ? `${task.priceFirstDate} to ${task.priceLatestDate}`
      : 'date range unavailable'
  const accessibleLabel = `${label}: ${sampleCount} price points, ${dateRange}, latest ${formatMoney(task.marketPrice)}`

  return (
    <>
      <button
        type='button'
        data-testid={`wishlist-price-chart-open-${task.id}`}
        className='rounded bg-slate-950/60 p-0 ring-offset-background transition-colors outline-none hover:bg-slate-900 focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2'
        aria-label={`Open ${task.title} price graph`}
        title={accessibleLabel}
        onClick={(event) => {
          event.stopPropagation()
          setIsDialogOpen(true)
        }}
        onDoubleClick={(event) => event.stopPropagation()}
      >
        <ChartContainer
          config={wishlistPriceChartConfig}
          data-testid={`wishlist-price-sparkline-${task.id}`}
          role='img'
          aria-label={accessibleLabel}
          className='h-7 w-[88px]'
        >
          <LineChart
            accessibilityLayer
            data={chartRows}
            margin={{ top: 3, right: 3, bottom: 3, left: 3 }}
          >
            <XAxis dataKey='date' hide />
            <YAxis dataKey='price' hide domain={['dataMin', 'dataMax']} />
            <Line
              type='monotone'
              dataKey='price'
              stroke='var(--color-price)'
              strokeWidth={2.5}
              dot={false}
              activeDot={false}
              isAnimationActive={false}
            />
          </LineChart>
        </ChartContainer>
      </button>
      <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
        <DialogContent
          className='sm:max-w-3xl'
          data-testid='wishlist-price-chart-dialog'
          closeButtonTestId='wishlist-price-chart-dialog-close'
        >
          <DialogHeader>
            <DialogTitle>{task.title} price history</DialogTitle>
            <DialogDescription>
              Latest {latestPointRows.length} of {sampleCount} price points.
              {dateRange !== 'date range unavailable' ? ` ${dateRange}.` : ''}
            </DialogDescription>
          </DialogHeader>
          <ChartContainer
            config={wishlistPriceChartConfig}
            data-testid='wishlist-price-chart-large'
            className='h-72 w-full rounded-md border bg-slate-950/40 p-3'
          >
            <LineChart
              accessibilityLayer
              data={chartRows}
              margin={{ top: 16, right: 20, bottom: 8, left: 8 }}
            >
              <XAxis
                dataKey='date'
                tickLine={false}
                axisLine={false}
                tickMargin={8}
                minTickGap={24}
              />
              <YAxis
                dataKey='price'
                tickLine={false}
                axisLine={false}
                tickMargin={8}
                width={56}
                domain={['dataMin', 'dataMax']}
                tickFormatter={(value) => formatMoney(Number(value))}
              />
              <ChartTooltip
                cursor={false}
                content={
                  <ChartTooltipContent
                    indicator='line'
                    labelFormatter={(value) => String(value)}
                    formatter={(value) => formatMoney(Number(value))}
                  />
                }
              />
              <Line
                type='monotone'
                dataKey='price'
                stroke='var(--color-price)'
                strokeWidth={3}
                dot={{ r: 3, fill: 'rgb(129 146 255)' }}
                activeDot={{ r: 5, fill: 'rgb(129 146 255)' }}
                isAnimationActive={false}
              />
            </LineChart>
          </ChartContainer>
          <div
            className='rounded-md border bg-card/40 p-3'
            data-testid={`wishlist-price-chart-points-${task.id}`}
          >
            <p className='mb-2 text-sm font-semibold'>Latest 10 price points</p>
            <ul className='grid gap-1 text-sm sm:grid-cols-2'>
              {latestPointRows.map((point) => (
                <li key={`${point.date}-${point.price}`}>
                  <span className='text-muted-foreground'>{point.date}</span>{' '}
                  <span className='font-medium'>
                    {formatMoney(point.price)}
                  </span>
                </li>
              ))}
            </ul>
          </div>
          <DialogFooter>
            <Button
              type='button'
              variant='outline'
              onClick={() => setIsDialogOpen(false)}
            >
              Close
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}

function getWishlistTrendConfig(trend: Task['priceTrend']) {
  return {
    up: {
      label: 'Price trending up',
      marker: '↑',
      className: 'text-red-300',
    },
    steady: {
      label: 'Price steady',
      marker: '-',
      className: 'text-muted-foreground',
    },
    down: {
      label: 'Price trending down',
      marker: '↓',
      className: 'text-emerald-300',
    },
    unknown: {
      label: 'Price trend unknown',
      marker: '-',
      className: 'text-muted-foreground',
    },
  }[trend ?? 'unknown']
}

function WishlistPriceTrendCell({ task }: { task: Task }) {
  const trendConfig = getWishlistTrendConfig(task.priceTrend)
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
      className='flex min-w-[8rem] items-center gap-2'
      data-testid={`wishlist-price-trend-${task.id}`}
      aria-label={trendConfig.label}
      title={`${trendConfig.label}. ${sampleCount} points. ${dateText}. Sources: ${sourceText}. ${stockText}.`}
    >
      <span
        className={`inline-flex w-4 justify-center text-sm leading-none font-semibold ${trendConfig.className}`}
        data-testid={`wishlist-price-trend-marker-${task.id}`}
        aria-hidden='true'
      >
        {trendConfig.marker}
      </span>
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

function hashWishlistThumbnailKey(value: string) {
  return [...value].reduce((hash, char) => {
    return (hash * 31 + char.charCodeAt(0)) % 360
  }, 23)
}

function WishlistThumbnail({ task }: { task: Task }) {
  const key = task.itemID?.trim() || task.id.trim() || task.title.trim()
  const hue = hashWishlistThumbnailKey(key)
  const accentHue = (hue + 46) % 360
  const initials = task.title
    .split(/\s+/)
    .map((word) => word[0])
    .filter(Boolean)
    .slice(0, 2)
    .join('')
    .toUpperCase()

  if (task.thumbnailUrl) {
    return (
      <img
        src={task.thumbnailUrl}
        alt=''
        aria-hidden='true'
        data-testid={`wishlist-thumbnail-${task.id}`}
        data-thumbnail-key={key}
        className='h-8 w-8 shrink-0 rounded-md border object-cover'
      />
    )
  }

  return (
    <span
      aria-hidden='true'
      data-testid={`wishlist-thumbnail-${task.id}`}
      data-thumbnail-key={key}
      className='relative flex h-8 w-8 shrink-0 items-center justify-center overflow-hidden rounded-md border text-[0.625rem] font-semibold text-white shadow-sm'
      style={{
        background: `linear-gradient(135deg, hsl(${hue} 68% 42%), hsl(${accentHue} 72% 36%))`,
      }}
    >
      <span className='absolute inset-x-0 top-0 h-1/3 bg-white/18' />
      <span className='absolute right-0 bottom-0 h-4 w-4 rounded-tl-full bg-black/20' />
      <span className='relative'>{initials || 'W'}</span>
    </span>
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

  const persistValue = async (rawValue: string) => {
    const trimmed = rawValue.trim()
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

  const persist = () => persistValue(draft)

  const stepCost = (direction: -1 | 1) => {
    const parsed = Number(draft.trim())
    const base = Number.isFinite(parsed) ? parsed : (task.targetPrice ?? 0)
    const nextValue = Math.max(0, Number((base + direction * 0.01).toFixed(2)))
    const nextDraft = formatCostDraft(nextValue)
    setDraft(nextDraft)
    void persistValue(nextDraft)
  }

  return (
    <div
      data-testid={`wishlist-cost-stepper-${task.id}`}
      className='flex w-[8.75rem] items-center'
      onClick={(event) => event.stopPropagation()}
      onDoubleClick={(event) => event.stopPropagation()}
    >
      <Button
        type='button'
        variant='outline'
        size='icon'
        className='h-8 w-8 shrink-0 rounded-r-none'
        data-testid={`wishlist-cost-decrease-${task.id}`}
        aria-label={`Decrease cost for ${task.title}`}
        onMouseDown={(event) => event.preventDefault()}
        onClick={() => stepCost(-1)}
      >
        <MinusIcon className='h-4 w-4' />
      </Button>
      <Input
        type='number'
        min='0'
        step='0.01'
        inputMode='decimal'
        value={draft}
        data-testid={`wishlist-cost-input-${task.id}`}
        aria-label={`Cost for ${task.title}`}
        className={wishlistStepperInputClassName}
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
      <Button
        type='button'
        variant='outline'
        size='icon'
        className='h-8 w-8 shrink-0 rounded-l-none'
        data-testid={`wishlist-cost-increase-${task.id}`}
        aria-label={`Increase cost for ${task.title}`}
        onMouseDown={(event) => event.preventDefault()}
        onClick={() => stepCost(1)}
      >
        <PlusIcon className='h-4 w-4' />
      </Button>
    </div>
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
  onWishlistPurchaseRow,
}: {
  task: Task
  onWishlistPurchaseRow?: (task: Task) => void
}) {
  return (
    <div
      className='flex min-w-[2.5rem] items-center'
      onClick={(event) => event.stopPropagation()}
      onDoubleClick={(event) => event.stopPropagation()}
    >
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
  minValue,
  onWishlistInlineUpdate,
}: {
  task: Task
  field: 'quantity' | 'neededQuantity'
  label: string
  value: number | undefined
  minValue: number
  onWishlistInlineUpdate?: (
    task: Task,
    changes: WishlistInlineChanges
  ) => Promise<void>
}) {
  const normalizedValue = value ?? minValue
  const [draft, setDraft] = useState(String(normalizedValue))

  useEffect(() => {
    setDraft(String(normalizedValue))
  }, [normalizedValue])

  const persistValue = async (rawValue: string) => {
    const parsed = Number(rawValue.trim())
    if (!Number.isInteger(parsed) || parsed < minValue) {
      setDraft(String(value ?? minValue))
      return
    }
    if ((value ?? 0) === parsed) {
      return
    }
    await onWishlistInlineUpdate?.(task, { [field]: parsed })
  }

  const persist = () => persistValue(draft)

  const stepNumber = (direction: -1 | 1) => {
    const parsed = Number(draft.trim())
    const base = Number.isInteger(parsed) ? parsed : (value ?? minValue)
    const nextValue = Math.max(minValue, base + direction)
    const nextDraft = String(nextValue)
    setDraft(nextDraft)
    void persistValue(nextDraft)
  }

  const testIDPrefix = field === 'quantity' ? 'wishlist-qty' : 'wishlist-needs'

  return (
    <div
      data-testid={`${testIDPrefix}-stepper-${task.id}`}
      className='flex w-[7rem] items-center'
      onClick={(event) => event.stopPropagation()}
      onDoubleClick={(event) => event.stopPropagation()}
    >
      <Button
        type='button'
        variant='outline'
        size='icon'
        className='h-8 w-8 shrink-0 rounded-r-none'
        data-testid={`${testIDPrefix}-decrease-${task.id}`}
        aria-label={`Decrease ${label.toLowerCase()} for ${task.title}`}
        onMouseDown={(event) => event.preventDefault()}
        onClick={() => stepNumber(-1)}
      >
        <MinusIcon className='h-4 w-4' />
      </Button>
      <Input
        type='number'
        min={String(minValue)}
        step='1'
        inputMode='numeric'
        value={draft}
        data-testid={`${testIDPrefix}-input-${task.id}`}
        aria-label={`${label} for ${task.title}`}
        className={wishlistStepperInputClassName}
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
      <Button
        type='button'
        variant='outline'
        size='icon'
        className='h-8 w-8 shrink-0 rounded-l-none'
        data-testid={`${testIDPrefix}-increase-${task.id}`}
        aria-label={`Increase ${label.toLowerCase()} for ${task.title}`}
        onMouseDown={(event) => event.preventDefault()}
        onClick={() => stepNumber(1)}
      >
        <PlusIcon className='h-4 w-4' />
      </Button>
    </div>
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
                  className: 'w-[14rem] max-w-[14rem]',
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
          : 'ps-1 max-w-0 w-2/3',
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
                  className='block min-w-0 max-w-full truncate font-medium'
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
            meta: { className: 'ps-1', tdClassName: 'ps-4' },
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
              <DataTableColumnHeader column={column} title='Owned' />
            ),
            meta: {
              className: 'w-16 min-w-16 ps-1',
              tdClassName: 'ps-4 pe-3',
            },
            cell: ({ row }) => (
              <WishlistOwnedCell
                task={row.original}
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
            meta: { className: 'ps-1', tdClassName: 'ps-4' },
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
            meta: { className: 'ps-1', tdClassName: 'ps-4' },
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
      meta: { className: 'ps-1', tdClassName: 'ps-3' },
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
        className: isInventoryRoute ? 'w-40' : undefined,
        tdClassName: isInventoryRoute ? 'max-w-none' : undefined,
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
