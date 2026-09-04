import {
  type KeyboardEvent,
  type MouseEvent,
  type ReactNode,
  useCallback,
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
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { DataTablePagination, DataTableToolbar } from '@/components/data-table'
import { useProfileSettings } from '@/features/settings/use-profile-settings'
import { priorities, statuses } from '../data/data'
import { type Task } from '../data/schema'
import { DataTableBulkActions } from './data-table-bulk-actions'
import { getTasksColumns } from './tasks-columns'
import { WishlistThumbnail } from './wishlist-thumbnail'

type TasksRoutePath = '/_authenticated/inventory/' | '/_authenticated/wishlist/'

type DataTableProps = {
  data: Task[]
  routePath: TasksRoutePath
  currentRecordID?: string
  onRecordFocus?: (itemID: string, recordID: string, title: string) => void
  onOpenDetailsRow?: (task: Task, navigationRows?: Task[]) => void
  onEditRow?: (task: Task, navigationRows?: Task[]) => void
  onPhotoRow?: (task: Task) => void
  onBarcodeRow?: (task: Task) => void
  onAssignCollectionRow?: (task: Task) => void
  onDeleteRow?: (task: Task) => void
  onRestoreRow?: (task: Task) => void
  onWishlistBulkStatusChange?: (tasks: Task[], status: string) => Promise<void>
  onWishlistBulkPriorityChange?: (
    tasks: Task[],
    priority: string
  ) => Promise<void>
  onWishlistBulkDelete?: (tasks: Task[]) => Promise<void>
  onWishlistExport?: (tasks: Task[]) => void
  onWishlistInlineUpdate?: (
    task: Task,
    changes: {
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
  ) => Promise<void>
  isWishlistMutating?: boolean
  customFilters?: ReactNode
  wishlistCollectionOptions?: string[]
}

type ViewMode = 'rows' | 'cards'

const inventoryItemDragDataType = 'application/x-cabinet-inventory-item-id'

type InventorySavedView = {
  id: string
  name: string
  globalFilter: string
  statusFilters: string[]
  categoryFilters: string[]
  itemTypeFilters: string[]
  packagingGradeFilters: string[]
  sorting: Array<{ id: string; desc: boolean }>
  viewMode: ViewMode
}

function inventoryItemRowTestID(task: Task): string | undefined {
  return task.itemID ? `inventory-item-row-${task.itemID}` : undefined
}

const inventorySavedViewsSettingsKey = 'inventory.saved-views.v1'

function parseInventorySavedViews(
  value: string | undefined
): InventorySavedView[] {
  if (!value) {
    return []
  }

  try {
    const parsed = JSON.parse(value) as unknown
    if (!Array.isArray(parsed)) {
      return []
    }

    return parsed.flatMap((entry) => {
      if (!entry || typeof entry !== 'object') {
        return []
      }
      const candidate = entry as Record<string, unknown>
      const id = typeof candidate.id === 'string' ? candidate.id.trim() : ''
      const name =
        typeof candidate.name === 'string' ? candidate.name.trim() : ''
      if (!id || !name) {
        return []
      }

      const normalizeStringArray = (raw: unknown) =>
        Array.isArray(raw)
          ? raw
              .filter((item): item is string => typeof item === 'string')
              .map((item) => item.trim())
              .filter((item) => item !== '')
          : []

      const sorting = Array.isArray(candidate.sorting)
        ? candidate.sorting.flatMap((sortEntry) => {
            if (!sortEntry || typeof sortEntry !== 'object') {
              return []
            }
            const sortCandidate = sortEntry as Record<string, unknown>
            const columnID =
              typeof sortCandidate.id === 'string'
                ? sortCandidate.id.trim()
                : ''
            if (!columnID) {
              return []
            }
            return [
              {
                id: columnID,
                desc: Boolean(sortCandidate.desc),
              },
            ]
          })
        : []

      return [
        {
          id,
          name,
          globalFilter:
            typeof candidate.globalFilter === 'string'
              ? candidate.globalFilter
              : '',
          statusFilters: normalizeStringArray(candidate.statusFilters),
          categoryFilters: normalizeStringArray(candidate.categoryFilters),
          itemTypeFilters: normalizeStringArray(candidate.itemTypeFilters),
          packagingGradeFilters: normalizeStringArray(
            candidate.packagingGradeFilters
          ),
          sorting,
          viewMode: candidate.viewMode === 'cards' ? 'cards' : 'rows',
        },
      ]
    })
  } catch {
    return []
  }
}

function serializeInventorySavedViews(views: InventorySavedView[]) {
  return JSON.stringify(
    views.map((view) => ({
      id: view.id,
      name: view.name,
      globalFilter: view.globalFilter,
      statusFilters: view.statusFilters,
      categoryFilters: view.categoryFilters,
      itemTypeFilters: view.itemTypeFilters,
      packagingGradeFilters: view.packagingGradeFilters,
      sorting: view.sorting,
      viewMode: view.viewMode,
    }))
  )
}

function formatWishlistStatus(status: string) {
  if (status === 'wishlist') {
    return 'Watching'
  }
  if (status === 'discovered') {
    return 'Below target'
  }
  return status
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

function formatMoneyDraft(value: number | undefined) {
  if (typeof value !== 'number' || value <= 0) {
    return ''
  }
  return Number.isInteger(value) ? String(value) : value.toFixed(2)
}

function defaultPurchasePrice(task: Task) {
  return task.pricePaid || task.targetPrice || task.marketPrice
}

function defaultPurchaseQuantity(task: Task) {
  return task.quantity && task.quantity > 0 ? task.quantity : 1
}

function todayISODate() {
  return new Date().toISOString().slice(0, 10)
}

export function TasksTable({
  data,
  routePath,
  currentRecordID,
  onRecordFocus,
  onOpenDetailsRow,
  onEditRow,
  onPhotoRow,
  onBarcodeRow,
  onAssignCollectionRow,
  onDeleteRow,
  onRestoreRow,
  onWishlistBulkStatusChange,
  onWishlistBulkPriorityChange,
  onWishlistBulkDelete,
  onWishlistExport,
  onWishlistInlineUpdate,
  isWishlistMutating,
  customFilters,
  wishlistCollectionOptions = [],
}: DataTableProps) {
  const isInventoryRoute = routePath === '/_authenticated/inventory/'
  const isWishlistRoute = routePath === '/_authenticated/wishlist/'
  const [purchaseTask, setPurchaseTask] = useState<Task | null>(null)
  const [purchasePricePaid, setPurchasePricePaid] = useState('')
  const [purchaseUrl, setPurchaseUrl] = useState('')
  const [purchaseDate, setPurchaseDate] = useState(todayISODate())
  const [purchaseQuantity, setPurchaseQuantity] = useState('0')
  const [purchaseCondition, setPurchaseCondition] = useState('')
  const priceClearedOnFocusRef = useRef(false)
  const openPurchaseDialog = useCallback((task: Task) => {
    setPurchaseTask(task)
    setPurchasePricePaid(formatMoneyDraft(defaultPurchasePrice(task)))
    setPurchaseUrl(task.purchaseUrl ?? '')
    setPurchaseDate(task.purchaseDate || todayISODate())
    setPurchaseQuantity(String(defaultPurchaseQuantity(task)))
    setPurchaseCondition(task.purchaseCondition ?? '')
    priceClearedOnFocusRef.current = false
  }, [])
  const columns = useMemo(
    () =>
      getTasksColumns({
        routePath,
        onEditRow,
        onPhotoRow,
        onBarcodeRow,
        onAssignCollectionRow,
        onDeleteRow,
        onRestoreRow,
        onWishlistInlineUpdate,
        onWishlistPurchaseRow: openPurchaseDialog,
      }),
    [
      routePath,
      onEditRow,
      onPhotoRow,
      onBarcodeRow,
      onAssignCollectionRow,
      onDeleteRow,
      onRestoreRow,
      onWishlistInlineUpdate,
      openPurchaseDialog,
    ]
  )

  const route =
    routePath === '/_authenticated/inventory/'
      ? getRouteApi('/_authenticated/inventory/')
      : getRouteApi('/_authenticated/wishlist/')

  const storageKey = isWishlistRoute
    ? 'cabinet.viewMode.wishlist'
    : 'cabinet.viewMode.inventory'
  const {
    activeProfileId: activeInventoryProfileId,
    settings: inventoryProfileSettings,
    saveSettings: saveInventoryProfileSettings,
    loading: inventoryProfileSettingsLoading,
    saving: inventoryProfileSettingsSaving,
  } = useProfileSettings()

  const [rowSelection, setRowSelection] = useState({})
  const [sorting, setSorting] = useState<SortingState>([])
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>(
    (): VisibilityState => {
      if (isWishlistRoute) {
        return { collectionName: false, status: false }
      }
      return {}
    }
  )
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
  const [saveViewDialogOpen, setSaveViewDialogOpen] = useState(false)
  const [saveViewName, setSaveViewName] = useState('')
  const [savedViewFeedback, setSavedViewFeedback] = useState<string | null>(
    null
  )
  const [savedViewError, setSavedViewError] = useState<string | null>(null)
  const [activeSavedViewID, setActiveSavedViewID] = useState('')
  const clickTimerRef = useRef<number | null>(null)
  const saveViewNameInputRef = useRef<HTMLInputElement>(null)

  const routeSearch = route.useSearch()
  const routeNavigate = route.useNavigate()

  const {
    globalFilter,
    onGlobalFilterChange,
    columnFilters,
    onColumnFiltersChange,
    pagination,
    onPaginationChange,
    ensurePageInRange,
  } = useTableUrlState({
    search: routeSearch,
    navigate: routeNavigate,
    pagination: { defaultPage: 1, defaultPageSize: 10 },
    globalFilter: { enabled: true, key: 'filter' },
    columnFilters: [
      ...(isWishlistRoute
        ? []
        : [
            {
              columnId: 'status',
              searchKey: 'status',
              type: 'array' as const,
            },
          ]),
      { columnId: 'priority', searchKey: 'priority', type: 'array' as const },
      {
        columnId: 'collectionName',
        searchKey: 'collection',
        type: 'array' as const,
      },
      ...(isInventoryRoute
        ? [
            {
              columnId: 'itemType',
              searchKey: 'itemType',
              type: 'array' as const,
            },
            {
              columnId: 'packagingGradeType',
              searchKey: 'packagingGrade',
              type: 'array' as const,
            },
          ]
        : []),
    ],
  })
  const [wishlistStatusFilters, setWishlistStatusFilters] = useState<string[]>(
    () => {
      if (!isWishlistRoute || typeof window === 'undefined') {
        return []
      }
      const stored = window.localStorage.getItem(
        'cabinet.wishlist.statusFilters'
      )
      if (stored) {
        try {
          const parsed = JSON.parse(stored) as unknown
          if (Array.isArray(parsed)) {
            return parsed.filter(
              (value): value is string =>
                typeof value === 'string' && value.trim() !== ''
            )
          }
        } catch {
          // Fall through to route search values below.
        }
      }
      const rawStatus = (routeSearch as Record<string, unknown>).status
      const routeStatuses = Array.isArray(rawStatus)
        ? rawStatus.filter(
            (value): value is string =>
              typeof value === 'string' && value.trim() !== ''
          )
        : []
      return routeStatuses.length > 0
        ? routeStatuses
        : ['wishlist', 'discovered']
    }
  )

  const getArrayColumnFilterValue = useCallback(
    (id: string) => {
      const found = columnFilters.find((filter) => filter.id === id)
      return Array.isArray(found?.value)
        ? found.value
            .filter((value): value is string => typeof value === 'string')
            .map((value) => value.trim())
            .filter((value) => value !== '')
        : []
    },
    [columnFilters]
  )

  const setArrayColumnFilterValue = useCallback(
    (id: string, values: string[]) => {
      const normalizedValues = values
        .map((value) => value.trim())
        .filter((value) => value !== '')
      onColumnFiltersChange((current) => [
        ...current.filter((filter) => filter.id !== id),
        ...(normalizedValues.length > 0
          ? [{ id, value: normalizedValues }]
          : []),
      ])
    },
    [onColumnFiltersChange]
  )

  useEffect(() => {
    if (!isWishlistRoute || typeof window === 'undefined') {
      return
    }
    if (wishlistStatusFilters.length === 0) {
      window.localStorage.removeItem('cabinet.wishlist.statusFilters')
      return
    }
    window.localStorage.setItem(
      'cabinet.wishlist.statusFilters',
      JSON.stringify(wishlistStatusFilters)
    )
  }, [isWishlistRoute, wishlistStatusFilters])

  const filteredData = useMemo(() => {
    if (!isWishlistRoute) {
      return data
    }
    const activeFilters =
      wishlistStatusFilters.length > 0
        ? wishlistStatusFilters
        : ['wishlist', 'discovered']
    return data.filter((task) => activeFilters.includes(task.status))
  }, [data, isWishlistRoute, wishlistStatusFilters])

  const table = useReactTable({
    data: filteredData,
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
      const id = row.original.id.toLowerCase()
      const partNumber = (row.original.partNumber ?? '').toLowerCase()
      const title = String(row.getValue('title')).toLowerCase()
      const itemType = (row.original.itemType ?? '').toLowerCase()
      const condition = (row.original.condition ?? '').toLowerCase()
      const packagingGrade = (
        row.original.packagingGradeType ?? ''
      ).toLowerCase()
      const category = (row.original.label ?? '').toLowerCase()
      const searchValue = String(filterValue).toLowerCase()
      return (
        id.includes(searchValue) ||
        partNumber.includes(searchValue) ||
        title.includes(searchValue) ||
        itemType.includes(searchValue) ||
        condition.includes(searchValue) ||
        packagingGrade.includes(searchValue) ||
        category.includes(searchValue)
      )
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

  const inventorySavedViews = useMemo(
    () =>
      isInventoryRoute
        ? parseInventorySavedViews(
            inventoryProfileSettings[inventorySavedViewsSettingsKey]
          )
        : [],
    [inventoryProfileSettings, isInventoryRoute]
  )

  const visibleRecordIDs = useMemo(
    () => table.getRowModel().rows.map((row) => row.original.id),
    [table]
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
      target.closest(
        'button, a, input, select, textarea, [role="checkbox"], [data-sidebar="menu-action"]'
      )
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
    const record = data.find((item) => item.id === id)
    if (record) {
      onRecordFocus?.(record.itemID ?? record.id, record.id, record.title)
    }
    if (
      (isInventoryRoute || routePath === '/_authenticated/wishlist/') &&
      record &&
      onOpenDetailsRow
    ) {
      setSelectedRecordContext(id)
      if (clickTimerRef.current !== null) {
        window.clearTimeout(clickTimerRef.current)
      }
      clickTimerRef.current = window.setTimeout(() => {
        onOpenDetailsRow(
          record,
          table.getRowModel().rows.map((row) => row.original)
        )
      }, 180)
      return
    }
    if (isInventoryRoute || routePath === '/_authenticated/wishlist/') {
      return
    }
    if (clickTimerRef.current !== null) {
      window.clearTimeout(clickTimerRef.current)
    }
    clickTimerRef.current = window.setTimeout(() => {
      openDetails(id)
    }, 180)
  }

  const handleRowDoubleClick = (id: string, event: MouseEvent<HTMLElement>) => {
    if (isInteractiveTarget(event.target)) {
      return
    }
    const record = data.find((item) => item.id === id)
    if (record) {
      onRecordFocus?.(record.itemID ?? record.id, record.id, record.title)
    }
    if (clickTimerRef.current !== null) {
      window.clearTimeout(clickTimerRef.current)
      clickTimerRef.current = null
    }
    if (record && onEditRow) {
      onEditRow(
        record,
        table.getRowModel().rows.map((row) => row.original)
      )
      return
    }
    if (record && onOpenDetailsRow) {
      onOpenDetailsRow(
        record,
        table.getRowModel().rows.map((row) => row.original)
      )
      return
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
    setSelectedRecordContext(
      visibleRecordIDs[nextIndex] ?? selectedRecordID ?? ''
    )
  }

  const currentInventoryViewSnapshot = useMemo(() => {
    return {
      globalFilter: (globalFilter ?? '').trim(),
      statusFilters: getArrayColumnFilterValue('status'),
      categoryFilters: getArrayColumnFilterValue('priority'),
      itemTypeFilters: getArrayColumnFilterValue('itemType'),
      packagingGradeFilters: getArrayColumnFilterValue('packagingGradeType'),
      sorting: sorting.map((entry) => ({
        id: entry.id,
        desc: Boolean(entry.desc),
      })),
      viewMode,
    }
  }, [getArrayColumnFilterValue, globalFilter, sorting, viewMode])

  const persistInventorySavedViews = useCallback(
    async (nextViews: InventorySavedView[]) => {
      await saveInventoryProfileSettings({
        ...inventoryProfileSettings,
        [inventorySavedViewsSettingsKey]:
          serializeInventorySavedViews(nextViews),
      })
    },
    [inventoryProfileSettings, saveInventoryProfileSettings]
  )

  const applyInventorySavedView = useCallback(
    (view: InventorySavedView) => {
      onGlobalFilterChange?.(view.globalFilter)
      onColumnFiltersChange([
        ...(view.statusFilters.length > 0
          ? [{ id: 'status', value: view.statusFilters }]
          : []),
        ...(view.categoryFilters.length > 0
          ? [{ id: 'priority', value: view.categoryFilters }]
          : []),
        ...(view.itemTypeFilters.length > 0
          ? [{ id: 'itemType', value: view.itemTypeFilters }]
          : []),
        ...(view.packagingGradeFilters.length > 0
          ? [{ id: 'packagingGradeType', value: view.packagingGradeFilters }]
          : []),
      ])
      setSorting(view.sorting)
      setViewMode(view.viewMode)
      setActiveSavedViewID(view.id)
      setSavedViewError(null)
      setSavedViewFeedback(`Applied saved view: ${view.name}`)
    },
    [onColumnFiltersChange, onGlobalFilterChange]
  )

  const handleSaveInventoryView = useCallback(async () => {
    const name = saveViewName.trim()
    if (!isInventoryRoute || !activeInventoryProfileId || name === '') {
      return
    }

    const matchingView = inventorySavedViews.find(
      (view) => view.name.toLowerCase() === name.toLowerCase()
    )
    const nextView: InventorySavedView = {
      id:
        matchingView?.id ??
        `inventory-view-${name.toLowerCase().replace(/[^a-z0-9]+/g, '-')}-${Date.now()}`,
      name,
      ...currentInventoryViewSnapshot,
    }

    setSavedViewError(null)

    try {
      await persistInventorySavedViews([
        ...inventorySavedViews.filter((view) => view.id !== matchingView?.id),
        nextView,
      ])
      setActiveSavedViewID(nextView.id)
      setSaveViewDialogOpen(false)
      setSaveViewName('')
      setSavedViewFeedback(`Saved view: ${name}`)
    } catch {
      setSavedViewError(
        'Saved view failed to persist. Retry once profile settings are available.'
      )
    }
  }, [
    activeInventoryProfileId,
    currentInventoryViewSnapshot,
    inventorySavedViews,
    isInventoryRoute,
    persistInventorySavedViews,
    saveViewName,
  ])

  const handleDeleteInventoryView = useCallback(async () => {
    if (!activeSavedViewID) {
      return
    }
    const currentView = inventorySavedViews.find(
      (view) => view.id === activeSavedViewID
    )
    if (!currentView) {
      return
    }
    setSavedViewError(null)
    try {
      await persistInventorySavedViews(
        inventorySavedViews.filter((view) => view.id !== activeSavedViewID)
      )
      setActiveSavedViewID('')
      setSavedViewFeedback(`Deleted saved view: ${currentView.name}`)
    } catch {
      setSavedViewError(
        'Saved view failed to delete. Retry once profile settings are available.'
      )
    }
  }, [activeSavedViewID, inventorySavedViews, persistInventorySavedViews])

  useEffect(() => {
    if (
      activeSavedViewID !== '' &&
      !inventorySavedViews.some((view) => view.id === activeSavedViewID)
    ) {
      setActiveSavedViewID('')
    }
  }, [activeSavedViewID, inventorySavedViews])

  const statusFilterOptions = isInventoryRoute
    ? Array.from(
        new Set(
          data
            .map((task) => task.status)
            .filter((value): value is string => Boolean(value?.trim()))
        )
      ).map((value) => ({ label: value, value }))
    : routePath === '/_authenticated/wishlist/'
      ? [
          { label: 'Watching', value: 'wishlist' },
          { label: 'Below target', value: 'discovered' },
          { label: 'Deleted', value: 'deleted' },
        ]
      : statuses
  const categoryFilterOptions = isInventoryRoute
    ? Array.from(
        new Set(
          data
            .flatMap((task) => String(task.label || '').split(','))
            .map((value) => value.trim())
            .filter((value) => value !== '')
        )
      ).map((value) => ({ label: value, value }))
    : priorities
  const itemTypeFilterOptions = isInventoryRoute
    ? Array.from(
        new Set(
          data
            .map((task) => task.itemType)
            .filter((value): value is string => Boolean(value?.trim()))
        )
      ).map((value) => ({ label: value, value }))
    : []
  const packagingGradeFilterOptions = isInventoryRoute
    ? Array.from(
        new Set(
          data
            .map((task) => task.packagingGradeType)
            .filter((value): value is string => Boolean(value?.trim()))
        )
      ).map((value) => ({ label: value, value }))
    : []
  const wishlistCollectionFilterOptions =
    routePath === '/_authenticated/wishlist/'
      ? [
          { label: 'All Items', value: 'All Items' },
          ...Array.from(
            new Set(
              [
                ...wishlistCollectionOptions,
                ...data
                  .map((task) => task.collectionName?.trim())
                  .filter((value): value is string => Boolean(value)),
              ].map((value) => value.trim())
            )
          )
            .filter((value) => value !== 'All Items')
            .map((value) => ({ label: value, value })),
        ]
      : []

  const savePurchaseDetails = useCallback(async () => {
    if (!purchaseTask) {
      return
    }

    const parsedPrice = Number(purchasePricePaid.trim())
    if (Number.isNaN(parsedPrice) || parsedPrice < 0) {
      setPurchasePricePaid(formatMoneyDraft(defaultPurchasePrice(purchaseTask)))
      return
    }
    const parsedQuantity = Number(purchaseQuantity.trim())
    if (
      !Number.isInteger(parsedQuantity) ||
      Number.isNaN(parsedQuantity) ||
      parsedQuantity < 0
    ) {
      setPurchaseQuantity(String(defaultPurchaseQuantity(purchaseTask)))
      return
    }

    await onWishlistInlineUpdate?.(purchaseTask, {
      owned: true,
      delivered: purchaseTask.delivered,
      pricePaid: parsedPrice,
      purchaseUrl: purchaseUrl.trim(),
      purchaseDate: purchaseDate || todayISODate(),
      purchaseCondition: purchaseCondition.trim(),
      quantity: parsedQuantity,
    })
    setPurchaseTask(null)
  }, [
    onWishlistInlineUpdate,
    purchaseCondition,
    purchaseDate,
    purchasePricePaid,
    purchaseQuantity,
    purchaseTask,
    purchaseUrl,
  ])

  const viewModeControls = (
    <div className='flex shrink-0 items-center gap-2'>
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
  )

  return (
    <div
      data-testid={
        isInventoryRoute ? 'inventory-table-stack' : 'wishlist-table-stack'
      }
      className={cn(
        'max-sm:has-[div[role="toolbar"]]:mb-16',
        'flex min-h-0 flex-1 flex-col gap-4'
      )}
    >
      <DataTableToolbar
        table={table}
        toolbarTestId={
          isInventoryRoute
            ? 'inventory-table-toolbar'
            : 'wishlist-table-toolbar'
        }
        searchPlaceholder={
          isInventoryRoute
            ? 'Filter by title, part number, type, condition, or packaging...'
            : 'Filter by title or part number...'
        }
        searchInputTestId={
          isInventoryRoute
            ? 'inventory-table-search-input'
            : 'wishlist-table-search-input'
        }
        filters={
          isInventoryRoute
            ? [
                {
                  columnId: 'status',
                  title: 'Condition',
                  options: statusFilterOptions,
                },
                {
                  columnId: 'priority',
                  title: 'Category',
                  options: categoryFilterOptions,
                },
                {
                  columnId: 'itemType',
                  title: 'Item type',
                  options: itemTypeFilterOptions,
                  testIdPrefix: 'inventory-table-item-type',
                },
                {
                  columnId: 'packagingGradeType',
                  title: 'Packaging',
                  options: packagingGradeFilterOptions,
                  testIdPrefix: 'inventory-table-packaging',
                },
              ]
            : [
                {
                  columnId: 'status',
                  title: 'Status',
                  options: statusFilterOptions,
                  singleSelect: true,
                  testIdPrefix: 'wishlist-table-status',
                  selectedValues: new Set(wishlistStatusFilters),
                  onSelectedValuesChange: setWishlistStatusFilters,
                },
                {
                  columnId: 'priority',
                  title: 'Priority',
                  options: categoryFilterOptions,
                },
                {
                  columnId: 'collectionName',
                  title: 'Collection',
                  options: wishlistCollectionFilterOptions,
                  singleSelect: true,
                  testIdPrefix: 'wishlist-table-collection',
                  selectedValues: new Set(
                    getArrayColumnFilterValue('collectionName')
                  ),
                  onSelectedValuesChange: (values) =>
                    setArrayColumnFilterValue('collectionName', values),
                },
              ]
        }
        customFilters={customFilters}
        actions={!isInventoryRoute ? viewModeControls : undefined}
      />

      {isInventoryRoute ? (
        <div className='flex flex-wrap items-center justify-between gap-2'>
          <div className='flex flex-wrap items-center gap-2'>
            <select
              className='h-9 min-w-[12rem] rounded-md border bg-background px-2 text-sm'
              data-testid='inventory-saved-view-select'
              value={activeSavedViewID}
              disabled={inventoryProfileSettingsLoading}
              onChange={(event) => {
                const nextID = event.target.value
                if (nextID === '') {
                  setActiveSavedViewID('')
                  setSavedViewFeedback(null)
                  setSavedViewError(null)
                  return
                }
                const nextView = inventorySavedViews.find(
                  (view) => view.id === nextID
                )
                if (!nextView) {
                  return
                }
                applyInventorySavedView(nextView)
              }}
            >
              <option value=''>Saved views</option>
              {inventorySavedViews.map((view) => (
                <option key={view.id} value={view.id}>
                  {view.name}
                </option>
              ))}
            </select>
            <Button
              type='button'
              size='sm'
              variant='outline'
              data-testid='inventory-saved-view-save'
              disabled={
                inventoryProfileSettingsLoading ||
                inventoryProfileSettingsSaving
              }
              onClick={() => {
                setSavedViewFeedback(null)
                setSavedViewError(null)
                setSaveViewName((previous) =>
                  previous !== ''
                    ? previous
                    : activeSavedViewID !== ''
                      ? (inventorySavedViews.find(
                          (view) => view.id === activeSavedViewID
                        )?.name ?? '')
                      : ''
                )
                setSaveViewDialogOpen(true)
              }}
            >
              Save View
            </Button>
            <Button
              type='button'
              size='sm'
              variant='outline'
              data-testid='inventory-saved-view-delete'
              disabled={
                activeSavedViewID === '' ||
                inventoryProfileSettingsLoading ||
                inventoryProfileSettingsSaving
              }
              onClick={() => void handleDeleteInventoryView()}
            >
              Delete View
            </Button>
            {savedViewFeedback ? (
              <p
                className='text-xs text-muted-foreground'
                data-testid='inventory-saved-view-feedback'
              >
                {savedViewFeedback}
              </p>
            ) : null}
            {savedViewError ? (
              <p
                className='text-xs text-destructive'
                data-testid='inventory-saved-view-error'
              >
                {savedViewError}
              </p>
            ) : null}
          </div>
          {viewModeControls}
        </div>
      ) : null}

      {viewMode === 'rows' ? (
        <div
          className='min-h-0 flex-1 overflow-auto rounded-md border'
          data-testid={
            isInventoryRoute
              ? 'inventory-table-surface'
              : 'wishlist-table-surface'
          }
        >
          <Table
            className={cn(
              isInventoryRoute ? 'min-w-[88rem] table-fixed' : 'min-w-[42rem]',
              routePath === '/_authenticated/wishlist/' ? 'min-w-[56rem]' : ''
            )}
          >
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
                    data-testid={inventoryItemRowTestID(row.original)}
                    data-state={row.getIsSelected() && 'selected'}
                    draggable={
                      routePath === '/_authenticated/inventory/' &&
                      Boolean(row.original.itemID)
                    }
                    className={cn(
                      currentRecordID ===
                        (row.original.itemID ?? row.original.id)
                        ? 'bg-primary/5'
                        : undefined
                    )}
                    onDragStart={(event) => {
                      const itemID = row.original.itemID
                      if (
                        routePath !== '/_authenticated/inventory/' ||
                        !itemID
                      ) {
                        return
                      }
                      event.dataTransfer.effectAllowed = 'move'
                      event.dataTransfer.setData(
                        inventoryItemDragDataType,
                        itemID
                      )
                      event.dataTransfer.setData('text/plain', itemID)
                    }}
                    onClick={(event) => handleRowClick(row.original.id, event)}
                    onDoubleClick={(event) =>
                      handleRowDoubleClick(row.original.id, event)
                    }
                  >
                    {row.getVisibleCells().map((cell) => (
                      <TableCell
                        key={cell.id}
                        className={cn(
                          isInventoryRoute
                            ? 'max-w-0 overflow-hidden text-ellipsis'
                            : undefined,
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
        <div className='grid min-h-0 flex-1 auto-rows-max gap-3 overflow-auto sm:grid-cols-2 lg:grid-cols-4'>
          {table.getRowModel().rows?.length ? (
            table.getRowModel().rows.map((row) => (
              <div
                key={row.id}
                data-testid={inventoryItemRowTestID(row.original)}
                draggable={
                  routePath === '/_authenticated/inventory/' &&
                  Boolean(row.original.itemID)
                }
                className={cn(
                  'space-y-2 rounded-md border',
                  routePath === '/_authenticated/wishlist/' ? 'p-3' : 'p-4',
                  currentRecordID === (row.original.itemID ?? row.original.id)
                    ? 'border-primary/60 bg-primary/5'
                    : 'cursor-pointer'
                )}
                data-state={row.getIsSelected() && 'selected'}
                onDragStart={(event) => {
                  const itemID = row.original.itemID
                  if (routePath !== '/_authenticated/inventory/' || !itemID) {
                    return
                  }
                  event.dataTransfer.effectAllowed = 'move'
                  event.dataTransfer.setData(inventoryItemDragDataType, itemID)
                  event.dataTransfer.setData('text/plain', itemID)
                }}
                onClick={(event) => handleRowClick(row.original.id, event)}
                onDoubleClick={(event) =>
                  handleRowDoubleClick(row.original.id, event)
                }
              >
                {routePath === '/_authenticated/wishlist/' ? (
                  <WishlistThumbnail task={row.original} variant='card' />
                ) : null}
                <div className='flex items-start justify-between gap-2'>
                  <div className='min-w-0 space-y-1'>
                    <p className='truncate text-xs text-muted-foreground'>
                      {row.original.id}
                    </p>
                    <p className='line-clamp-2 text-sm leading-snug font-medium'>
                      {row.original.title}
                    </p>
                  </div>
                  <Checkbox
                    checked={row.getIsSelected()}
                    onCheckedChange={(checked) =>
                      row.toggleSelected(Boolean(checked))
                    }
                    aria-label={`Select ${row.original.title}`}
                  />
                </div>
                <div className='flex flex-wrap gap-1.5 text-xs text-muted-foreground'>
                  <span>
                    Status:{' '}
                    {routePath === '/_authenticated/wishlist/'
                      ? formatWishlistStatus(row.original.status)
                      : row.original.status}
                  </span>
                  <span>Priority: {row.original.priority}</span>
                  <span>Category: {row.original.label}</span>
                  {routePath === '/_authenticated/wishlist/' ? (
                    <>
                      <span
                        data-testid={`wishlist-card-purchased-${row.original.id}`}
                      >
                        Purchased: {row.original.owned ? 'Yes' : 'No'}
                      </span>
                      <span
                        data-testid={`wishlist-card-delivered-${row.original.id}`}
                      >
                        Delivered: {row.original.delivered ? 'Yes' : 'No'}
                      </span>
                      <span
                        data-testid={`wishlist-card-date-added-${row.original.id}`}
                      >
                        Date added:{' '}
                        {formatWishlistDate(row.original.wishlistCreatedAt)}
                      </span>
                      <span
                        data-testid={`wishlist-card-date-updated-${row.original.id}`}
                        title='Latest pricing refresh date'
                      >
                        Updated:{' '}
                        {formatWishlistDate(
                          row.original.wishlistPriceUpdatedAt
                        )}
                      </span>
                    </>
                  ) : null}
                </div>
                {(routePath === '/_authenticated/wishlist/' ||
                  routePath === '/_authenticated/inventory/') &&
                row.original.notes ? (
                  <p
                    className='line-clamp-2 text-xs text-muted-foreground'
                    data-testid={
                      routePath === '/_authenticated/inventory/'
                        ? `inventory-card-notes-${row.original.itemID ?? row.original.id}`
                        : undefined
                    }
                  >
                    Notes: {row.original.notes}
                  </p>
                ) : null}
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
      <DataTableBulkActions
        table={table}
        routePath={routePath}
        onWishlistBulkStatusChange={onWishlistBulkStatusChange}
        onWishlistBulkPriorityChange={onWishlistBulkPriorityChange}
        onWishlistBulkDelete={onWishlistBulkDelete}
        onWishlistExport={onWishlistExport}
        isWishlistMutating={isWishlistMutating}
      />
      <Dialog
        open={purchaseTask !== null}
        onOpenChange={(open) => {
          if (!open) {
            setPurchaseTask(null)
          }
        }}
      >
        <DialogContent
          data-testid='wishlist-purchase-dialog'
          onOpenAutoFocus={(event) => event.preventDefault()}
        >
          <DialogHeader>
            <DialogTitle>Add purchase details</DialogTitle>
            <DialogDescription>
              Record what you paid and where this wishlist item was purchased.
            </DialogDescription>
          </DialogHeader>
          <div className='grid gap-4 sm:grid-cols-2'>
            <label className='space-y-2 text-sm font-medium'>
              <span>Price paid</span>
              <Input
                type='number'
                min='0'
                step='0.01'
                inputMode='decimal'
                value={purchasePricePaid}
                data-testid='wishlist-purchase-price-paid'
                onFocus={() => {
                  if (!priceClearedOnFocusRef.current) {
                    setPurchasePricePaid('')
                    priceClearedOnFocusRef.current = true
                  }
                }}
                onChange={(event) => setPurchasePricePaid(event.target.value)}
              />
            </label>
            <label className='space-y-2 text-sm font-medium'>
              <span>Purchase date</span>
              <Input
                type='date'
                value={purchaseDate}
                data-testid='wishlist-purchase-date'
                onChange={(event) => setPurchaseDate(event.target.value)}
              />
            </label>
            <label className='space-y-2 text-sm font-medium'>
              <span>Qty</span>
              <Input
                type='number'
                min='0'
                step='1'
                inputMode='numeric'
                value={purchaseQuantity}
                data-testid='wishlist-purchase-quantity'
                onChange={(event) => setPurchaseQuantity(event.target.value)}
              />
            </label>
            <label className='space-y-2 text-sm font-medium'>
              <span>Condition</span>
              <Input
                value={purchaseCondition}
                data-testid='wishlist-purchase-condition'
                placeholder='New, boxed, loose...'
                onChange={(event) => setPurchaseCondition(event.target.value)}
              />
            </label>
            <label className='space-y-2 text-sm font-medium sm:col-span-2'>
              <span>URL</span>
              <Input
                value={purchaseUrl}
                data-testid='wishlist-purchase-url'
                placeholder='https://'
                onChange={(event) => setPurchaseUrl(event.target.value)}
              />
            </label>
          </div>
          <DialogFooter>
            <Button
              type='button'
              variant='outline'
              onClick={() => setPurchaseTask(null)}
            >
              Cancel
            </Button>
            <Button
              type='button'
              data-testid='wishlist-purchase-save'
              onClick={() => void savePurchaseDetails()}
            >
              Save purchase
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
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
              {selectedRecord.notes ? (
                <p>
                  <strong>Notes:</strong> {selectedRecord.notes}
                </p>
              ) : null}
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
      <Dialog open={saveViewDialogOpen} onOpenChange={setSaveViewDialogOpen}>
        <DialogContent
          data-testid='inventory-saved-view-dialog'
          onOpenAutoFocus={(event) => {
            event.preventDefault()
            saveViewNameInputRef.current?.focus()
          }}
        >
          <DialogHeader>
            <DialogTitle>Save Inventory View</DialogTitle>
            <DialogDescription>
              Save the current inventory search, filters, sorting, and layout
              for this profile.
            </DialogDescription>
          </DialogHeader>
          <div className='space-y-2'>
            <label
              className='text-sm font-medium'
              htmlFor='inventory-saved-view-name'
            >
              View name
            </label>
            <Input
              ref={saveViewNameInputRef}
              id='inventory-saved-view-name'
              data-testid='inventory-saved-view-name'
              value={saveViewName}
              onChange={(event) => setSaveViewName(event.target.value)}
              placeholder='Used Road Cars'
            />
          </div>
          <DialogFooter>
            <Button
              type='button'
              variant='outline'
              onClick={() => {
                setSaveViewDialogOpen(false)
                setSaveViewName('')
              }}
            >
              Cancel
            </Button>
            <Button
              type='button'
              data-testid='inventory-saved-view-submit'
              disabled={
                saveViewName.trim() === '' ||
                inventoryProfileSettingsLoading ||
                inventoryProfileSettingsSaving
              }
              onClick={() => void handleSaveInventoryView()}
            >
              Save View
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
