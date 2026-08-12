import { useState } from 'react'
import { MinusIcon, PlusIcon } from 'lucide-react'
import { Line, LineChart, XAxis, YAxis } from 'recharts'
import { Button } from '@/components/ui/button'
import {
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
  type ChartConfig,
} from '@/components/ui/chart'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { priorities } from '../data/data'
import { type Task } from '../data/schema'
import {
  buildWishlistPricePointRows,
  formatMoney,
} from './tasks-column-formatters'

export type WishlistInlineChanges = {
  targetPrice?: number
  priority?: string
  owned?: boolean
  delivered?: boolean
  pricePaid?: number
  purchaseUrl?: string
  purchaseDate?: string
  purchaseCondition?: string
  quantity?: number
  neededQuantity?: number
}

function formatCostDraft(value: number | undefined) {
  if (typeof value !== 'number' || value < 0) {
    return ''
  }
  return Number.isInteger(value) ? String(value) : value.toFixed(2)
}

const wishlistStepperInputClassName =
  'h-8 border-x-0 rounded-none text-center [appearance:textfield] [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none'

const wishlistPriceChartConfig = {
  price: {
    label: 'Price',
    color: 'rgb(73 103 255)',
  },
} satisfies ChartConfig

export function WishlistPriceSparkline({
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

export function WishlistPriceTrendCell({ task }: { task: Task }) {
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

export function WishlistCostCell({
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

export function WishlistPriorityCell({
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

export function WishlistPurchasedCell({
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
        aria-label={
          task.owned
            ? `Edit purchase details for ${task.title}`
            : `Add purchase details for ${task.title}`
        }
        onClick={() => onWishlistPurchaseRow?.(task)}
      >
        <PlusIcon className='h-4 w-4' />
        <span className='sr-only'>{task.owned ? 'Purchased' : 'Add'}</span>
      </Button>
      <span
        className='ml-2 text-xs text-muted-foreground'
        data-testid={`wishlist-purchased-state-${task.id}`}
      >
        {task.owned ? 'Yes' : 'No'}
      </span>
    </div>
  )
}

export function WishlistPricePaidCell({ task }: { task: Task }) {
  return (
    <span
      className='min-w-[76px] font-medium'
      data-testid={`wishlist-price-paid-value-${task.id}`}
    >
      {formatMoney(task.pricePaid)}
    </span>
  )
}

export function WishlistNumberCell({
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
