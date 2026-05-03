import * as React from 'react'
import * as RechartsPrimitive from 'recharts'
import { type TooltipPayloadEntry } from 'recharts'
import { cn } from '@/lib/utils'

export type ChartConfig = {
  [key: string]: {
    label?: React.ReactNode
    color?: string
  }
}

const ChartContext = React.createContext<{ config: ChartConfig } | null>(null)

function useChart() {
  const context = React.useContext(ChartContext)
  if (!context) {
    throw new Error('useChart must be used within a <ChartContainer />')
  }
  return context
}

function ChartStyle({ id, config }: { id: string; config: ChartConfig }) {
  const colorConfig = Object.entries(config).filter(([, item]) => item.color)
  if (!colorConfig.length) {
    return null
  }

  return (
    <style
      dangerouslySetInnerHTML={{
        __html: colorConfig
          .map(
            ([key, item]) =>
              `[data-chart=${id}] { --color-${key}: ${item.color}; }`
          )
          .join('\n'),
      }}
    />
  )
}

function ChartContainer({
  id,
  className,
  children,
  config,
  ...props
}: React.ComponentProps<'div'> & {
  config: ChartConfig
  children: React.ComponentProps<
    typeof RechartsPrimitive.ResponsiveContainer
  >['children']
}) {
  const uniqueId = React.useId()
  const chartId = `chart-${id ?? uniqueId.replace(/:/g, '')}`

  return (
    <ChartContext.Provider value={{ config }}>
      <div
        data-chart={chartId}
        data-slot='chart'
        className={cn(
          "[&_.recharts-cartesian-axis-tick_text]:fill-muted-foreground [&_.recharts-grid_line]:stroke-border/50 [&_.recharts-tooltip-cursor]:stroke-border [&_.recharts-surface]:outline-none",
          className
        )}
        {...props}
      >
        <ChartStyle id={chartId} config={config} />
        <RechartsPrimitive.ResponsiveContainer>
          {children}
        </RechartsPrimitive.ResponsiveContainer>
      </div>
    </ChartContext.Provider>
  )
}

const ChartTooltip = RechartsPrimitive.Tooltip

function ChartTooltipContent({
  active,
  payload,
  className,
  label,
  hideLabel = false,
  hideIndicator = false,
  indicator = 'dot',
  labelFormatter,
  formatter,
}: React.ComponentProps<'div'> & {
  active?: boolean
  payload?: ReadonlyArray<TooltipPayloadEntry<number, string>>
  label?: string | number
  hideLabel?: boolean
  hideIndicator?: boolean
  indicator?: 'dot' | 'line' | 'dashed'
  labelFormatter?: (label: string | number) => React.ReactNode
  formatter?: (value: number, name: string) => React.ReactNode
}) {
  const { config } = useChart()

  if (!active || !payload?.length) {
    return null
  }

  return (
    <div
      className={cn(
        'grid min-w-32 gap-1.5 rounded-lg border bg-background px-2.5 py-1.5 text-xs shadow-xl',
        className
      )}
    >
      {!hideLabel && label !== undefined ? (
        <div className='font-medium'>
          {labelFormatter ? labelFormatter(label) : label}
        </div>
      ) : null}
      <div className='grid gap-1.5'>
        {payload.map((item) => {
          const key = String(item.dataKey ?? item.name ?? '')
          const labelText = config[key]?.label ?? item.name ?? key
          const color = item.color ?? config[key]?.color
          const indicatorClass =
            indicator === 'line'
              ? 'h-0.5 w-2.5'
              : indicator === 'dashed'
                ? 'h-0 w-2.5 border-t-2 border-dashed bg-transparent'
                : 'size-2 rounded-full'

          return (
            <div
              key={`${key}-${item.value}`}
              className='flex items-center gap-2'
            >
              {!hideIndicator ? (
                <span
                  className={indicatorClass}
                  style={{ backgroundColor: color, borderColor: color }}
                />
              ) : null}
              <span className='text-muted-foreground'>{labelText}</span>
              <span className='ml-auto font-mono font-medium tabular-nums text-foreground'>
                {formatter && typeof item.value === 'number'
                  ? formatter(item.value, key)
                  : item.value}
              </span>
            </div>
          )
        })}
      </div>
    </div>
  )
}

export { ChartContainer, ChartTooltip, ChartTooltipContent }
