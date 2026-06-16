import * as React from 'react'
import { CheckIcon, PlusCircledIcon } from '@radix-ui/react-icons'
import { type Column } from '@tanstack/react-table'
import { cn } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandList,
  CommandSeparator,
} from '@/components/ui/command'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import { Separator } from '@/components/ui/separator'

type DataTableFacetedFilterProps<TData, TValue> = {
  column?: Column<TData, TValue>
  title?: string
  options: {
    label: string
    value: string
    icon?: React.ComponentType<{ className?: string }>
  }[]
  selectedValues?: Set<string>
  onSelectedValuesChange?: (values: string[]) => void
  singleSelect?: boolean
  testIdPrefix?: string
}

function filterOptionTestID(value: string) {
  return value
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
}

export function DataTableFacetedFilter<TData, TValue>({
  column,
  title,
  options,
  selectedValues: controlledSelectedValues,
  onSelectedValuesChange,
  singleSelect = false,
  testIdPrefix,
}: DataTableFacetedFilterProps<TData, TValue>) {
  const facets = column?.getFacetedUniqueValues()
  const lastSelectionAtRef = React.useRef(0)
  const selectedValues =
    controlledSelectedValues ?? new Set(column?.getFilterValue() as string[])

  const commitSelectedValues = (nextSelectedValues: Set<string>) => {
    const filterValues = Array.from(nextSelectedValues)
    if (onSelectedValuesChange) {
      onSelectedValuesChange(filterValues)
      return
    }
    column?.setFilterValue(filterValues.length ? filterValues : undefined)
  }

  const guardDuplicateSelection = (now: number) => {
    if (now - lastSelectionAtRef.current < 50) {
      return false
    }
    lastSelectionAtRef.current = now
    return true
  }

  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button
          variant='outline'
          size='sm'
          className='h-8 border-dashed'
          data-testid={testIdPrefix ? `${testIdPrefix}-trigger` : undefined}
        >
          <PlusCircledIcon className='size-4' />
          {title}
          {selectedValues?.size > 0 && (
            <>
              <Separator orientation='vertical' className='mx-2 h-4' />
              <Badge
                variant='secondary'
                className='rounded-sm px-1 font-normal lg:hidden'
              >
                {selectedValues.size}
              </Badge>
              <div className='hidden space-x-1 lg:flex'>
                {selectedValues.size > 2 ? (
                  <Badge
                    variant='secondary'
                    className='rounded-sm px-1 font-normal'
                  >
                    {selectedValues.size} selected
                  </Badge>
                ) : (
                  options
                    .filter((option) => selectedValues.has(option.value))
                    .map((option) => (
                      <Badge
                        variant='secondary'
                        key={option.value}
                        className='rounded-sm px-1 font-normal'
                      >
                        {option.label}
                      </Badge>
                    ))
                )}
              </div>
            </>
          )}
        </Button>
      </PopoverTrigger>
      <PopoverContent
        className='w-[200px] p-0'
        align='start'
        data-testid={testIdPrefix ? `${testIdPrefix}-content` : undefined}
      >
        <Command>
          <CommandInput placeholder={title} />
          <CommandList>
            <CommandEmpty>No results found.</CommandEmpty>
            <CommandGroup>
              {options.map((option) => {
                const isSelected = selectedValues.has(option.value)
                const selectOption = (
                  event: React.MouseEvent<HTMLButtonElement>
                ) => {
                  if (!guardDuplicateSelection(event.timeStamp)) {
                    return
                  }
                  const nextSelectedValues = new Set(selectedValues)
                  if (singleSelect) {
                    nextSelectedValues.clear()
                    nextSelectedValues.add(option.value)
                  } else if (isSelected) {
                    nextSelectedValues.delete(option.value)
                  } else {
                    nextSelectedValues.add(option.value)
                  }
                  commitSelectedValues(nextSelectedValues)
                }
                return (
                  <button
                    key={option.value}
                    type='button'
                    role='option'
                    aria-selected={isSelected}
                    className='relative flex w-full cursor-default items-center gap-2 rounded-sm px-2 py-1.5 text-start text-sm outline-hidden select-none hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:text-accent-foreground'
                    data-testid={
                      testIdPrefix
                        ? `${testIdPrefix}-option-${filterOptionTestID(option.value)}`
                        : undefined
                    }
                    onClick={selectOption}
                  >
                    <div
                      className={cn(
                        'flex size-4 items-center justify-center rounded-sm border border-primary',
                        isSelected
                          ? 'bg-primary text-primary-foreground'
                          : 'opacity-50 [&_svg]:invisible'
                      )}
                    >
                      <CheckIcon className={cn('h-4 w-4 text-background')} />
                    </div>
                    {option.icon && (
                      <option.icon className='size-4 text-muted-foreground' />
                    )}
                    <span>{option.label}</span>
                    {facets?.get(option.value) && (
                      <span className='ms-auto flex h-4 w-4 items-center justify-center font-mono text-xs'>
                        {facets.get(option.value)}
                      </span>
                    )}
                  </button>
                )
              })}
            </CommandGroup>
            {selectedValues.size > 0 && (
              <>
                <CommandSeparator />
                <CommandGroup>
                  <button
                    type='button'
                    data-testid={
                      testIdPrefix ? `${testIdPrefix}-clear` : undefined
                    }
                    onClick={() => commitSelectedValues(new Set())}
                    className='relative flex w-full cursor-default items-center justify-center rounded-sm px-2 py-1.5 text-center text-sm outline-hidden select-none hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:text-accent-foreground'
                  >
                    Clear filters
                  </button>
                </CommandGroup>
              </>
            )}
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  )
}
