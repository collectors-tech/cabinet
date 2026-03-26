import {
  type DragEvent,
  type KeyboardEvent,
  type MouseEvent,
  useEffect,
  useMemo,
  useRef,
  useState,
} from 'react'
import { getRouteApi } from '@tanstack/react-router'
import {
  type SortingState,
  type VisibilityState,
  flexRender,
  getCoreRowModel,
  getFacetedRowModel,
  getFacetedUniqueValues,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  useReactTable,
} from '@tanstack/react-table'
import { cn } from '@/lib/utils'
import { useTableUrlState } from '@/hooks/use-table-url-state'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { DataTablePagination, DataTableToolbar } from '@/components/data-table'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { priorities, statuses } from '../data/data'
import { type Task } from '../data/schema'
import { DataTableBulkActions } from './data-table-bulk-actions'
import { tasksColumns as columns } from './tasks-columns'

type TasksRoutePath = '/_authenticated/inventory/' | '/_authenticated/wishlist/'

type DataTableProps = {
  data: Task[]
  routePath: TasksRoutePath
  getRowTestId?: (record: Task) => string | undefined
  onRowDragStart?: (record: Task, event: DragEvent<HTMLElement>) => void
  onRowDragEnd?: (record: Task, event: DragEvent<HTMLElement>) => void
}

type ViewMode = 'rows' | 'cards'

export function TasksTable({
  data,
  routePath,
  getRowTestId,
  onRowDragStart,
  onRowDragEnd,
}: DataTableProps) {
  const route =
    routePath === '/_authenticated/inventory/'
      ? getRouteApi('/_authenticated/inventory/')
      : getRouteApi('/_authenticated/wishlist/')

  const storageKey =
    routePath === '/_authenticated/wishlist/'
      ? 'cabinet.viewMode.wishlist'
      : 'cabinet.viewMode.inventory'

  const [rowSelection, setRowSelection] = useState({})
  const [sorting, setSorting] = useState<SortingState>([])
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({})
  const [viewMode, setViewMode] = useState<ViewMode>(() => {
    if (typeof window === 'undefined') {
      return 'rows'
    }
    const saved = window.localStorage.getItem(storageKey)
    return saved === 'cards' ? 'cards' : 'rows'
  })
  const [selectedRecordID, setSelectedRecordID] = useState<string | null>(null)
  const [detailsOpen, setDetailsOpen] = useState(false)
  const [editOpen, setEditOpen] = useState(false)
  const clickTimerRef = useRef<number | null>(null)

  const {
    globalFilter,
    onGlobalFilterChange,
    columnFilters,
    onColumnFiltersChange,
    pagination,
    onPaginationChange,
    ensurePageInRange,
  } = useTableUrlState({
    search: route.useSearch(),
    navigate: route.useNavigate(),
    pagination: { defaultPage: 1, defaultPageSize: 10 },
    globalFilter: { enabled: true, key: 'filter' },
    columnFilters: [
      { columnId: 'status', searchKey: 'status', type: 'array' },
      { columnId: 'priority', searchKey: 'priority', type: 'array' },
    ],
  })

  // eslint-disable-next-line react-hooks/incompatible-library
  const table = useReactTable({
    data,
    columns,
    state: {
      sorting,
      columnVisibility,
      rowSelection,
      columnFilters,
      globalFilter,
      pagination,
    },
    enableRowSelection: true,
    onRowSelectionChange: setRowSelection,
    onSortingChange: setSorting,
    onColumnVisibilityChange: setColumnVisibility,
    globalFilterFn: (row, _columnId, filterValue) => {
      const id = String(row.getValue('id')).toLowerCase()
      const title = String(row.getValue('title')).toLowerCase()
      const searchValue = String(filterValue).toLowerCase()
      return id.includes(searchValue) || title.includes(searchValue)
    },
    getCoreRowModel: getCoreRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFacetedRowModel: getFacetedRowModel(),
    getFacetedUniqueValues: getFacetedUniqueValues(),
    onPaginationChange,
    onGlobalFilterChange,
    onColumnFiltersChange,
  })

  const pageCount = table.getPageCount()
  useEffect(() => {
    ensurePageInRange(pageCount)
  }, [pageCount, ensurePageInRange])

  useEffect(() => {
    if (typeof window === 'undefined') {
      return
    }
    window.localStorage.setItem(storageKey, viewMode)
  }, [storageKey, viewMode])

  useEffect(() => {
    return () => {
      if (clickTimerRef.current !== null) {
        window.clearTimeout(clickTimerRef.current)
      }
    }
  }, [])

  const selectedRecord = useMemo(
    () => data.find((item) => item.id === selectedRecordID) ?? null,
    [data, selectedRecordID]
  )

  const visibleRecordIDs = useMemo(
    () => table.getRowModel().rows.map((row) => row.original.id),
    [table, data, rowSelection, columnFilters, globalFilter, pagination, sorting]
  )

  const selectedVisibleIndex = selectedRecordID
    ? visibleRecordIDs.findIndex((id) => id === selectedRecordID)
    : -1

  const setSelectedRecordContext = (id: string) => {
    setSelectedRecordID(id)
    if (typeof window !== 'undefined') {
      const url = new URL(window.location.href)
      url.searchParams.set('selected', id)
      window.history.replaceState({}, '', url.toString())
    }
  }

  const isInteractiveTarget = (target: EventTarget | null) => {
    if (!(target instanceof HTMLElement)) {
      return false
    }
    return Boolean(
      target.closest('button, a, input, select, textarea, [role="checkbox"], [data-sidebar="menu-action"]')
    )
  }

  const openDetails = (id: string) => {
    setSelectedRecordContext(id)
    setDetailsOpen(true)
  }

  const openEdit = (id: string) => {
    setSelectedRecordContext(id)
    setDetailsOpen(false)
    setEditOpen(true)
  }

  const handleRowClick = (id: string, event: MouseEvent<HTMLElement>) => {
    if (isInteractiveTarget(event.target)) {
      return
    }
    if (clickTimerRef.current !== null) {
      window.clearTimeout(clickTimerRef.current)
    }
    clickTimerRef.current = window.setTimeout(() => {
      openDetails(id)
    }, 180)
  }

  const handleRowDoubleClick = (
    id: string,
    event: MouseEvent<HTMLElement>
  ) => {
    if (isInteractiveTarget(event.target)) {
      return
    }
    if (clickTimerRef.current !== null) {
      window.clearTimeout(clickTimerRef.current)
      clickTimerRef.current = null
    }
    openEdit(id)
  }

  const handleViewModeKeyDown = (
    mode: ViewMode,
    event: KeyboardEvent<HTMLButtonElement>
  ) => {
    if (event.key !== 'Enter' && event.key !== ' ') {
      return
    }
    event.preventDefault()
    setViewMode(mode)
  }

  const openAdjacentRecord = (offset: number) => {
    if (selectedVisibleIndex < 0) {
      return
    }
    const nextIndex = selectedVisibleIndex + offset
    if (nextIndex < 0 || nextIndex >= visibleRecordIDs.length) {
      return
    }
    setSelectedRecordContext(visibleRecordIDs[nextIndex] ?? selectedRecordID ?? '')
  }

  return (
    <div
      className={cn(
        'max-sm:has-[div[role="toolbar"]]:mb-16',
        'flex flex-1 flex-col gap-4'
      )}
    >
      <DataTableToolbar
        table={table}
        searchPlaceholder='Filter by title or ID...'
        filters={[
          {
            columnId: 'status',
            title: 'Status',
            options: statuses,
          },
          {
            columnId: 'priority',
            title: 'Priority',
            options: priorities,
          },
        ]}
      />

      <div className='flex items-center justify-end gap-2'>
        <Button
          size='sm'
          variant={viewMode === 'rows' ? 'default' : 'outline'}
          onClick={() => setViewMode('rows')}
          onKeyDown={(event) => handleViewModeKeyDown('rows', event)}
          aria-pressed={viewMode === 'rows'}
          aria-label='Switch to rows view'
        >
          Rows
        </Button>
        <Button
          size='sm'
          variant={viewMode === 'cards' ? 'default' : 'outline'}
          onClick={() => setViewMode('cards')}
          onKeyDown={(event) => handleViewModeKeyDown('cards', event)}
          aria-pressed={viewMode === 'cards'}
          aria-label='Switch to cards view'
        >
          Cards
        </Button>
      </div>

      {viewMode === 'rows' ? (
        <div className='overflow-hidden rounded-md border'>
          <Table className='min-w-xl'>
            <TableHeader>
              {table.getHeaderGroups().map((headerGroup) => (
                <TableRow key={headerGroup.id}>
                  {headerGroup.headers.map((header) => (
                    <TableHead
                      key={header.id}
                      colSpan={header.colSpan}
                      className={cn(
                        header.column.columnDef.meta?.className,
                        header.column.columnDef.meta?.thClassName
                      )}
                    >
                      {header.isPlaceholder
                        ? null
                        : flexRender(
                            header.column.columnDef.header,
                            header.getContext()
                          )}
                    </TableHead>
                  ))}
                </TableRow>
              ))}
            </TableHeader>
            <TableBody>
              {table.getRowModel().rows?.length ? (
                table.getRowModel().rows.map((row) => (
                  <TableRow
                    key={row.id}
                    data-testid={getRowTestId?.(row.original)}
                    data-state={row.getIsSelected() && 'selected'}
                    draggable={Boolean(onRowDragStart && row.original.recordID)}
                    onDragStart={(event) => onRowDragStart?.(row.original, event)}
                    onDragEnd={(event) => onRowDragEnd?.(row.original, event)}
                    onClick={(event) => handleRowClick(row.original.id, event)}
                    onDoubleClick={(event) =>
                      handleRowDoubleClick(row.original.id, event)
                    }
                  >
                    {row.getVisibleCells().map((cell) => (
                      <TableCell
                        key={cell.id}
                        className={cn(
                          cell.column.columnDef.meta?.className,
                          cell.column.columnDef.meta?.tdClassName
                        )}
                      >
                        {flexRender(
                          cell.column.columnDef.cell,
                          cell.getContext()
                        )}
                      </TableCell>
                    ))}
                  </TableRow>
                ))
              ) : (
                <TableRow>
                  <TableCell
                    colSpan={columns.length}
                    className='h-24 text-center'
                  >
                    No results.
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </div>
      ) : (
        <div className='grid gap-3 sm:grid-cols-2 xl:grid-cols-3'>
          {table.getRowModel().rows?.length ? (
            table.getRowModel().rows.map((row) => (
              <div
                key={row.id}
                className='space-y-2 rounded-md border p-4'
                data-testid={getRowTestId?.(row.original)}
                data-state={row.getIsSelected() && 'selected'}
                draggable={Boolean(onRowDragStart && row.original.recordID)}
                onDragStart={(event) => onRowDragStart?.(row.original, event)}
                onDragEnd={(event) => onRowDragEnd?.(row.original, event)}
              >
                <div className='flex items-start justify-between gap-2'>
                  <div className='space-y-1'>
                    <p className='text-xs text-muted-foreground'>
                      {row.original.id}
                    </p>
                    <p className='font-medium'>{row.original.title}</p>
                  </div>
                  <Checkbox
                    checked={row.getIsSelected()}
                    onCheckedChange={(checked) =>
                      row.toggleSelected(Boolean(checked))
                    }
                    aria-label={`Select ${row.original.title}`}
                  />
                </div>
                <div className='flex flex-wrap gap-2 text-xs text-muted-foreground'>
                  <span>Status: {row.original.status}</span>
                  <span>Priority: {row.original.priority}</span>
                  <span>Type: {row.original.label}</span>
                </div>
              </div>
            ))
          ) : (
            <div className='rounded-md border p-6 text-sm text-muted-foreground'>
              No results.
            </div>
          )}
        </div>
      )}

      <DataTablePagination table={table} className='mt-auto' />
      <DataTableBulkActions table={table} />
      <Dialog open={detailsOpen} onOpenChange={setDetailsOpen}>
        <DialogContent data-testid='row-details-modal'>
          <DialogHeader>
            <DialogTitle>Row Details</DialogTitle>
            <DialogDescription>
              Inspect selected record context.
            </DialogDescription>
          </DialogHeader>
          {selectedRecord ? (
            <div className='space-y-1 text-sm'>
              <p>
                <strong>ID:</strong> {selectedRecord.id}
              </p>
              <p>
                <strong>Title:</strong> {selectedRecord.title}
              </p>
              <p>
                <strong>Status:</strong> {selectedRecord.status}
              </p>
            </div>
          ) : null}
        </DialogContent>
      </Dialog>
      <Dialog open={editOpen} onOpenChange={setEditOpen}>
        <DialogContent data-testid='row-edit-modal'>
          <DialogHeader>
            <DialogTitle>Edit Row</DialogTitle>
            <DialogDescription>
              Update selected row and navigate adjacent rows.
            </DialogDescription>
          </DialogHeader>
          {selectedRecord ? (
            <div className='space-y-1 text-sm'>
              <p>
                <strong>ID:</strong> {selectedRecord.id}
              </p>
              <p>
                <strong>Title:</strong> {selectedRecord.title}
              </p>
            </div>
          ) : null}
          <DialogFooter>
            <Button
              type='button'
              variant='outline'
              onClick={() => openAdjacentRecord(-1)}
            >
              Previous
            </Button>
            <Button
              type='button'
              variant='outline'
              onClick={() => openAdjacentRecord(1)}
            >
              Next
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
