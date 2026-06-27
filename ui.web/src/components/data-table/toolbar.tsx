import { type ReactNode } from 'react'
import { Cross2Icon } from '@radix-ui/react-icons'
import { type Table } from '@tanstack/react-table'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { DataTableFacetedFilter } from './faceted-filter'
import { DataTableViewOptions } from './view-options'

type DataTableToolbarProps<TData> = {
  table: Table<TData>
  searchPlaceholder?: string
  searchKey?: string
  searchInputTestId?: string
  toolbarTestId?: string
  actions?: ReactNode
  customFilters?: ReactNode
  filters?: {
    columnId: string
    title: string
    singleSelect?: boolean
    testIdPrefix?: string
    selectedValues?: Set<string>
    onSelectedValuesChange?: (values: string[]) => void
    options: {
      label: string
      value: string
      icon?: React.ComponentType<{ className?: string }>
    }[]
  }[]
}

export function DataTableToolbar<TData>({
  table,
  searchPlaceholder = 'Filter...',
  searchKey,
  searchInputTestId,
  toolbarTestId,
  actions,
  customFilters,
  filters = [],
}: DataTableToolbarProps<TData>) {
  const isFiltered =
    table.getState().columnFilters.length > 0 || table.getState().globalFilter

  return (
    <div
      className='flex flex-wrap items-center justify-between gap-2'
      data-testid={toolbarTestId}
    >
      <div className='flex flex-1 flex-col-reverse items-start gap-y-2 sm:flex-row sm:flex-wrap sm:items-center sm:gap-2'>
        {searchKey ? (
          <Input
            placeholder={searchPlaceholder}
            data-testid={searchInputTestId}
            value={
              (table.getColumn(searchKey)?.getFilterValue() as string) ?? ''
            }
            onChange={(event) =>
              table.getColumn(searchKey)?.setFilterValue(event.target.value)
            }
            className='h-8 w-[150px] lg:w-[250px]'
          />
        ) : (
          <Input
            placeholder={searchPlaceholder}
            data-testid={searchInputTestId}
            value={table.getState().globalFilter ?? ''}
            onChange={(event) => table.setGlobalFilter(event.target.value)}
            className='h-8 w-[150px] lg:w-[250px]'
          />
        )}
        <div className='flex flex-wrap gap-2'>
          {filters.map((filter) => {
            const column = table.getColumn(filter.columnId)
            if (!column) return null
            return (
              <DataTableFacetedFilter
                key={filter.columnId}
                column={column}
                title={filter.title}
                options={filter.options}
                singleSelect={filter.singleSelect}
                testIdPrefix={filter.testIdPrefix}
                selectedValues={filter.selectedValues}
                onSelectedValuesChange={filter.onSelectedValuesChange}
              />
            )
          })}
        </div>
        {customFilters}
        {isFiltered && (
          <Button
            variant='ghost'
            onClick={() => {
              table.resetColumnFilters()
              table.setGlobalFilter('')
            }}
            className='h-8 px-2 lg:px-3'
          >
            Reset
            <Cross2Icon className='ms-2 h-4 w-4' />
          </Button>
        )}
      </div>
      <div className='ms-auto flex flex-wrap items-center justify-end gap-2'>
        {actions}
        <DataTableViewOptions table={table} />
      </div>
    </div>
  )
}
