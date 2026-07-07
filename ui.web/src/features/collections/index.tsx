import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import {
  type ColumnDef,
  flexRender,
  getCoreRowModel,
  getFacetedRowModel,
  getFacetedUniqueValues,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  type SortingState,
  useReactTable,
} from '@tanstack/react-table'
import { Eye, Pencil, Plus, Tag, Trash2 } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Separator } from '@/components/ui/separator'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'
import { Textarea } from '@/components/ui/textarea'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import {
  DataTableColumnHeader,
  DataTablePagination,
  DataTableToolbar,
} from '@/components/data-table'
import { LanguageSwitch } from '@/components/language-switch'
import { Header, HeaderTitle } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { recordNotificationHistory } from '@/lib/toast-history'
import {
  type WorkspaceCollectionItem,
  type WorkspaceCollectionSummary,
  collectionKey,
  useWorkspaceCollections,
} from './use-workspace-collections'

const collectionsHeaderDescription =
  'Manage collection rows and item placement from the shared Cabinet table surface.'

type CollectionRow = WorkspaceCollectionSummary
type CollectionMemberRow = WorkspaceCollectionItem

const tableClassName = 'w-full table-fixed'
const tableCellClassName = 'max-w-0 truncate'
const tableHeaderClassName = 'max-w-0 truncate'
const actionsCellClassName = 'w-[8rem] min-w-[8rem]'

function recordCollectionsStatusHistory({
  id,
  level = 'success',
  title,
}: {
  id: string
  level?: 'success' | 'error' | 'warning'
  title: string
}) {
  recordNotificationHistory({
    id,
    level,
    title,
    summary: 'Collections workspace status from Collections.',
    source_label: 'Collections',
    category: 'notification',
  })
}

type InventoryCatalogItem = {
  id?: string
  part_number?: string
  title?: string
  brand?: string
  category?: string
  description?: string
  status?: string
}

function normalizeString(value: unknown): string {
  return typeof value === 'string' ? value.trim() : ''
}

function collectionMemberFromInventoryItem(
  item: InventoryCatalogItem
): CollectionMemberRow | null {
  const id = normalizeString(item.id) || normalizeString(item.part_number)
  const name = normalizeString(item.title) || normalizeString(item.part_number)
  if (!id || !name) {
    return null
  }

  const detail =
    [
      normalizeString(item.part_number),
      normalizeString(item.brand),
      normalizeString(item.category),
    ]
      .filter(Boolean)
      .join(' · ') ||
    normalizeString(item.description) ||
    normalizeString(item.status) ||
    'Inventory item'

  return {
    id,
    name,
    detail,
    collectionName: null,
  }
}

function isEditableShortcutTarget(target: EventTarget | null) {
  if (!(target instanceof HTMLElement)) {
    return false
  }

  return Boolean(
    target.closest(
      'input, textarea, select, [contenteditable="true"], [role="textbox"]'
    )
  )
}

function buildCollectionColumns({
  onView,
  onEdit,
  onDelete,
}: {
  onView: (row: CollectionRow) => void
  onEdit: (row: CollectionRow) => void
  onDelete: (row: CollectionRow) => void
}): ColumnDef<CollectionRow>[] {
  return [
    {
      accessorKey: 'name',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='Collection' />
      ),
      cell: ({ row }) => (
        <div className='min-w-0 space-y-1'>
          <div
            className='truncate font-medium'
            data-testid={`collections-row-name-${row.original.key}`}
          >
            {row.original.name}
          </div>
          <div className='truncate text-xs text-muted-foreground'>
            {row.original.description}
          </div>
          {row.original.deletedAt ? (
            <div
              className='text-xs font-medium text-destructive'
              data-testid={`collections-row-deleted-${row.original.key}`}
            >
              Deleted collection
            </div>
          ) : null}
        </div>
      ),
    },
    {
      accessorKey: 'itemCount',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='Items' />
      ),
      cell: ({ row }) => (
        <span
          className='block truncate'
          data-testid={`collections-row-count-${row.original.key}`}
        >
          {row.original.itemCount}
        </span>
      ),
    },
    {
      accessorKey: 'scopeLabel',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='Scope' />
      ),
    },
    {
      accessorKey: 'statusLabel',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='Status' />
      ),
    },
    {
      accessorKey: 'updatedLabel',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='Updated' />
      ),
    },
    {
      id: 'actions',
      header: () => <div className='text-right'>Actions</div>,
      cell: ({ row }) => (
        <div className='flex items-center justify-end gap-2'>
          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                type='button'
                size='icon'
                variant='outline'
                data-testid={`collections-row-view-${row.original.key}`}
                aria-label={`View ${row.original.name} in inventory`}
                title={`View ${row.original.name} in inventory`}
                onClick={(event) => {
                  event.stopPropagation()
                  onView(row.original)
                }}
                onDoubleClick={(event) => event.stopPropagation()}
              >
                <Eye className='h-4 w-4' aria-hidden='true' />
              </Button>
            </TooltipTrigger>
            <TooltipContent>View in inventory</TooltipContent>
          </Tooltip>
          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                type='button'
                size='icon'
                variant='outline'
                data-testid={`collections-row-edit-${row.original.key}`}
                aria-label={`Edit ${row.original.name}`}
                title={`Edit ${row.original.name}`}
                onClick={(event) => {
                  event.stopPropagation()
                  onEdit(row.original)
                }}
                onDoubleClick={(event) => event.stopPropagation()}
              >
                <Pencil className='h-4 w-4' aria-hidden='true' />
              </Button>
            </TooltipTrigger>
            <TooltipContent>Edit collection</TooltipContent>
          </Tooltip>
          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                type='button'
                size='icon'
                variant='outline'
                data-testid={`collections-row-delete-${row.original.key}`}
                aria-label={`Delete ${row.original.name}`}
                title={`Delete ${row.original.name}`}
                onClick={(event) => {
                  event.stopPropagation()
                  onDelete(row.original)
                }}
                onDoubleClick={(event) => event.stopPropagation()}
              >
                <Trash2 className='h-4 w-4' aria-hidden='true' />
              </Button>
            </TooltipTrigger>
            <TooltipContent>Delete collection</TooltipContent>
          </Tooltip>
        </div>
      ),
    },
  ]
}

function buildCollectionMemberColumns(): ColumnDef<CollectionMemberRow>[] {
  return [
    {
      accessorKey: 'name',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='Item' />
      ),
      cell: ({ row }) => (
        <span
          className='block truncate font-medium'
          data-testid={`collections-member-${collectionKey(row.original.name)}`}
        >
          <span
            className='block truncate'
            data-testid={`collections-member-name-${collectionKey(row.original.name)}`}
          >
            {row.original.name}
          </span>
        </span>
      ),
    },
    {
      accessorKey: 'detail',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='Details' />
      ),
      cell: ({ row }) => (
        <span className='block truncate text-muted-foreground'>
          {row.original.detail}
        </span>
      ),
    },
    {
      accessorKey: 'collectionName',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='Current collection' />
      ),
      cell: ({ row }) => (
        <span
          className='block truncate text-muted-foreground'
          data-testid={`collections-member-current-${collectionKey(row.original.name)}`}
        >
          Currently in {row.original.collectionName ?? 'Unassigned'}.
        </span>
      ),
    },
  ]
}

export function Collections() {
  const {
    activeWorkspaceCollection,
    setActiveWorkspaceCollection,
    addCollection,
    updateCollectionDetails,
    removeCollection,
    collectionItems,
    collectionSummaries,
    deletedCollectionSummaries,
  } = useWorkspaceCollections()

  const [sorting, setSorting] = useState<SortingState>([
    { id: 'name', desc: false },
  ])
  const [globalFilter, setGlobalFilter] = useState('')
  const [memberSorting, setMemberSorting] = useState<SortingState>([
    { id: 'name', desc: false },
  ])
  const [memberGlobalFilter, setMemberGlobalFilter] = useState('')
  const [selectedCollectionID, setSelectedCollectionID] = useState<
    string | null
  >(null)
  const [createOpen, setCreateOpen] = useState(false)
  const [editOpen, setEditOpen] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [showDeletedCollections, setShowDeletedCollections] = useState(false)
  const [createValue, setCreateValue] = useState('')
  const [createError, setCreateError] = useState('')
  const [createSubmitting, setCreateSubmitting] = useState(false)
  const createInputRef = useRef<HTMLInputElement>(null)
  const [editValue, setEditValue] = useState('')
  const [editScopeValue, setEditScopeValue] = useState('')
  const [editStatusValue, setEditStatusValue] = useState('')
  const [editDescriptionValue, setEditDescriptionValue] = useState('')
  const [deleteDestination, setDeleteDestination] = useState('')
  const [inventoryMembers, setInventoryMembers] = useState<
    CollectionMemberRow[]
  >([])

  useEffect(() => {
    const controller = new AbortController()

    async function loadInventoryMembers() {
      try {
        const response = await fetch('/api/items', {
          signal: controller.signal,
        })
        if (!response.ok) {
          return
        }
        const payload = (await response.json()) as { items?: unknown[] }
        const nextMembers = (payload.items ?? [])
          .map((item) =>
            collectionMemberFromInventoryItem(item as InventoryCatalogItem)
          )
          .filter((item): item is CollectionMemberRow => item !== null)
        setInventoryMembers(nextMembers)
      } catch (error) {
        void error
      }
    }

    void loadInventoryMembers()

    return () => controller.abort()
  }, [])

  const collectionAssignmentByID = useMemo(() => {
    const assignments = new Map<string, string | null>()
    collectionItems.forEach((item) => {
      assignments.set(item.id, item.collectionName)
    })
    return assignments
  }, [collectionItems])

  const memberRows = useMemo(() => {
    if (inventoryMembers.length === 0) {
      return collectionItems
    }

    return inventoryMembers.map((item) => ({
      ...item,
      collectionName: collectionAssignmentByID.get(item.id) ?? null,
    }))
  }, [collectionAssignmentByID, collectionItems, inventoryMembers])

  const activeRows = useMemo(
    () =>
      collectionSummaries.map((summary) => ({
        ...summary,
        itemCount:
          summary.name === 'All Items'
            ? memberRows.length
            : memberRows.filter((item) => item.collectionName === summary.name)
                .length,
      })),
    [collectionSummaries, memberRows]
  )
  const deletedRows = useMemo(
    () =>
      deletedCollectionSummaries.map((summary) => ({
        ...summary,
        itemCount: 0,
      })),
    [deletedCollectionSummaries]
  )
  const rows = showDeletedCollections ? deletedRows : activeRows

  const selectedRow = useMemo(
    () =>
      rows.find((row) => row.key === selectedCollectionID) ??
      rows.find((row) => row.name === activeWorkspaceCollection) ??
      rows[0] ??
      null,
    [activeWorkspaceCollection, rows, selectedCollectionID]
  )

  const selectedCollectionName = selectedRow?.name ?? 'All Items'
  const selectedCollectionAssignedCount = useMemo(
    () =>
      selectedRow && selectedRow.name !== 'All Items'
        ? memberRows.filter((item) => item.collectionName === selectedRow.name)
            .length
        : 0,
    [memberRows, selectedRow]
  )
  const deleteDestinationOptions = useMemo(
    () =>
      activeRows.filter(
        (row) =>
          row.name !== 'All Items' &&
          row.key !== selectedRow?.key &&
          !row.deletedAt
      ),
    [activeRows, selectedRow]
  )

  const selectedCollectionItems = useMemo(
    () =>
      selectedCollectionName === 'All Items'
        ? memberRows
        : memberRows.filter(
            (item) => item.collectionName === selectedCollectionName
          ),
    [memberRows, selectedCollectionName]
  )

  const handleSelectCollection = useCallback(
    async (row: CollectionRow) => {
      setSelectedCollectionID(row.key)
      if (row.deletedAt) {
        return
      }
      await setActiveWorkspaceCollection(row.name)
    },
    [setActiveWorkspaceCollection]
  )

  const handleViewCollection = useCallback(
    async (row: CollectionRow) => {
      await handleSelectCollection(row)
      window.location.assign('/inventory/')
    },
    [handleSelectCollection]
  )

  const openEditPanel = useCallback((row: CollectionRow) => {
    setSelectedCollectionID(row.key)
    setEditValue(row.name)
    setEditScopeValue(row.scopeLabel)
    setEditStatusValue(row.deletedAt ? 'Deleted' : row.statusLabel)
    setEditDescriptionValue(row.description)
    setEditOpen(true)
  }, [])

  const columns = useMemo(
    () =>
      buildCollectionColumns({
        onView: (row) => {
          void handleViewCollection(row)
        },
        onEdit: openEditPanel,
        onDelete: (row) => {
          setSelectedCollectionID(row.key)
          setDeleteDestination('')
          setDeleteOpen(true)
        },
      }),
    [handleViewCollection, openEditPanel]
  )
  const memberColumns = useMemo(() => buildCollectionMemberColumns(), [])

  const table = useReactTable({
    data: rows,
    columns,
    state: {
      sorting,
      globalFilter,
    },
    onSortingChange: setSorting,
    onGlobalFilterChange: setGlobalFilter,
    globalFilterFn: (row, _columnId, filterValue) => {
      const searchValue = String(filterValue).trim().toLowerCase()
      if (!searchValue) {
        return true
      }
      return [
        row.original.name,
        row.original.description,
        row.original.scopeLabel,
        row.original.statusLabel,
        String(row.original.itemCount),
      ].some((value) => value.toLowerCase().includes(searchValue))
    },
    getCoreRowModel: getCoreRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getFacetedRowModel: getFacetedRowModel(),
    getFacetedUniqueValues: getFacetedUniqueValues(),
  })

  const membersTable = useReactTable({
    data: selectedCollectionItems,
    columns: memberColumns,
    state: {
      sorting: memberSorting,
      globalFilter: memberGlobalFilter,
    },
    onSortingChange: setMemberSorting,
    onGlobalFilterChange: setMemberGlobalFilter,
    globalFilterFn: (row, _columnId, filterValue) => {
      const searchValue = String(filterValue).trim().toLowerCase()
      if (!searchValue) {
        return true
      }
      return [
        row.original.name,
        row.original.detail,
        row.original.collectionName ?? 'Unassigned',
      ].some((value) => value.toLowerCase().includes(searchValue))
    },
    getCoreRowModel: getCoreRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getFacetedRowModel: getFacetedRowModel(),
    getFacetedUniqueValues: getFacetedUniqueValues(),
  })

  const filteredCount = table.getFilteredRowModel().rows.length
  const filteredMemberCount = membersTable.getFilteredRowModel().rows.length
  const visibleRows = table
    .getRowModel()
    .rows.map((tableRow) => tableRow.original)
  const selectedVisibleIndex = selectedRow
    ? visibleRows.findIndex((row) => row.key === selectedRow.key)
    : -1
  const canNavigatePrevious = selectedVisibleIndex > 0
  const canNavigateNext =
    selectedVisibleIndex >= 0 && selectedVisibleIndex < visibleRows.length - 1

  const setCreateDialogOpen = useCallback((open: boolean) => {
    setCreateOpen(open)
    if (!open) {
      setCreateValue('')
      setCreateError('')
      setCreateSubmitting(false)
    }
  }, [])

  useEffect(() => {
    function handleCommandCreateCollection() {
      setCreateDialogOpen(true)
    }

    window.addEventListener(
      'cabinet:create-collection',
      handleCommandCreateCollection
    )

    const searchParams = new URLSearchParams(window.location.search)
    if (searchParams.get('create') === 'collection') {
      setCreateDialogOpen(true)
      searchParams.delete('create')
      const nextSearch = searchParams.toString()
      window.history.replaceState(
        null,
        '',
        `${window.location.pathname}${nextSearch ? `?${nextSearch}` : ''}`
      )
    }

    return () =>
      window.removeEventListener(
        'cabinet:create-collection',
        handleCommandCreateCollection
      )
  }, [setCreateDialogOpen])

  useEffect(() => {
    function handleShortcut(event: KeyboardEvent) {
      if (
        event.key.toLowerCase() !== 'n' ||
        !event.ctrlKey ||
        event.metaKey ||
        event.altKey ||
        event.shiftKey ||
        isEditableShortcutTarget(event.target)
      ) {
        return
      }

      event.preventDefault()
      setCreateDialogOpen(true)
    }

    window.addEventListener('keydown', handleShortcut)
    return () => window.removeEventListener('keydown', handleShortcut)
  }, [setCreateDialogOpen])

  function navigateEditPanel(offset: number) {
    if (selectedVisibleIndex < 0) {
      return
    }
    const nextRow = visibleRows[selectedVisibleIndex + offset]
    if (!nextRow) {
      return
    }
    setSelectedCollectionID(nextRow.key)
    setEditValue(nextRow.name)
    setEditScopeValue(nextRow.scopeLabel)
    setEditStatusValue(nextRow.deletedAt ? 'Deleted' : nextRow.statusLabel)
    setEditDescriptionValue(nextRow.description)
  }

  async function handleCreateCollection() {
    if (createSubmitting) {
      return
    }
    setCreateSubmitting(true)
    setCreateError('')
    try {
      const created = await addCollection(createValue)
      if (!created) {
        setCreateError('Enter a unique collection name.')
        createInputRef.current?.focus()
        return
      }
      setSelectedCollectionID(collectionKey(created))
      setCreateValue('')
      setCreateDialogOpen(false)
      const message = `${created} created and set as the active collection.`
      toast.success(message)
      recordCollectionsStatusHistory({
        id: 'collections-create-success',
        title: message,
      })
    } finally {
      setCreateSubmitting(false)
    }
  }

  async function handleRenameCollection() {
    if (!selectedRow) {
      return
    }
    const originalName = selectedRow.name
    const renamed = await updateCollectionDetails(originalName, editValue, {
      scopeLabel: editScopeValue,
      statusLabel: editStatusValue,
      description: editDescriptionValue,
      updatedLabel: 'Updated just now',
    })
    if (!renamed) {
      const message = 'Rename failed. Use a unique non-empty collection name.'
      toast.error(message)
      recordCollectionsStatusHistory({
        id: 'collections-rename-error',
        level: 'error',
        title: message,
      })
      return
    }
    setSelectedCollectionID(collectionKey(renamed))
    setEditOpen(false)
    const message =
      originalName === renamed
        ? `${renamed} metadata updated.`
        : `${originalName} renamed to ${renamed}.`
    toast.success(message)
    recordCollectionsStatusHistory({
      id: 'collections-rename-success',
      title: message,
    })
  }

  async function handleDeleteCollection() {
    if (!selectedRow) {
      return
    }
    const removedName = selectedRow.name
    const removed = await removeCollection(removedName, deleteDestination)
    if (!removed) {
      const message = 'Collection removal failed.'
      toast.error(message)
      recordCollectionsStatusHistory({
        id: 'collections-remove-error',
        level: 'error',
        title: message,
      })
      return
    }
    setSelectedCollectionID(collectionKey('All Items'))
    setDeleteOpen(false)
    setDeleteDestination('')
    const message = `${removedName} hidden from active workspace collections.`
    toast.success(message)
    recordCollectionsStatusHistory({
      id: 'collections-remove-success',
      title: message,
    })
  }

  return (
    <>
      <Header fixed data-testid='collections-shell-header'>
        <Search />
        <HeaderTitle
          title='Collections'
          description={collectionsHeaderDescription}
          icon={Tag}
          testId='collections-header-title'
          iconTestId='collections-page-icon'
        />
        <div
          className='ms-auto flex min-w-0 items-center gap-3'
          data-header-title-avoid='true'
        >
          <div
            className='flex min-w-0 flex-wrap items-center justify-end gap-2'
            data-testid='collections-global-header-actions'
          >
            <Button
              data-testid='collections-new-action'
              size='icon'
              aria-label='New collection'
              title='New collection'
              onClick={() => setCreateDialogOpen(true)}
            >
              <Plus className='h-4 w-4' aria-hidden='true' />
            </Button>
          </div>
          <Separator
            orientation='vertical'
            className='h-6'
            data-testid='collections-header-action-separator'
          />
          <div className='flex items-center gap-4'>
            <ThemeSwitch />
            <LanguageSwitch />
            <ProfileDropdown />
          </div>
        </div>
      </Header>

      <Main fixed className='min-h-0 gap-3 sm:gap-4'>
        <div
          className='grid min-h-0 flex-1 grid-rows-[minmax(0,0.9fr)_minmax(0,1.1fr)] gap-4'
          data-testid='collections-workspace'
        >
          <Card
            className='flex min-h-0 flex-col overflow-hidden'
            data-testid='collections-section'
          >
            <CardContent className='flex min-h-0 flex-1 flex-col gap-3'>
              <div
                className='space-y-3'
                data-testid='collections-management-tools'
              >
                <div data-testid='collections-table-toolbar'>
                  <DataTableToolbar
                    table={table}
                    searchPlaceholder='Filter collections...'
                    searchInputTestId='collections-search-input'
                  />
                </div>
                <div
                  className='flex flex-wrap items-center gap-2 text-sm text-muted-foreground'
                  data-testid='collections-management-summary'
                >
                  <span>
                    Showing {filteredCount} of {rows.length} collections.
                  </span>
                  <span data-testid='collections-active-context'>
                    {selectedCollectionName}
                  </span>
                  <Button
                    type='button'
                    variant={showDeletedCollections ? 'default' : 'outline'}
                    size='sm'
                    data-testid='collections-deleted-filter-toggle'
                    onClick={() => {
                      setShowDeletedCollections((current) => !current)
                      setSelectedCollectionID(null)
                    }}
                  >
                    {showDeletedCollections ? 'Showing deleted' : 'Show deleted'}
                  </Button>
                </div>
              </div>

              <div
                className='min-h-0 flex-1 overflow-auto rounded-md border'
                data-table-surface='true'
                data-testid='collections-shared-table'
              >
                <Table className={tableClassName}>
                  <TableHeader>
                    {table.getHeaderGroups().map((headerGroup) => (
                      <TableRow key={headerGroup.id}>
                        {headerGroup.headers.map((header) => (
                          <TableHead
                            key={header.id}
                            className={
                              header.column.id === 'actions'
                                ? actionsCellClassName
                                : tableHeaderClassName
                            }
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
                    {table.getRowModel().rows.length ? (
                      table.getRowModel().rows.map((row) => {
                        const isSelected = row.original.key === selectedRow?.key
                        return (
                          <TableRow
                            key={row.id}
                            className='cursor-pointer'
                            data-state={isSelected ? 'selected' : undefined}
                            data-testid={`collections-row-${row.original.key}`}
                            onClick={() => {
                              void handleSelectCollection(row.original)
                            }}
                            onDoubleClick={() => {
                              openEditPanel(row.original)
                            }}
                          >
                            {row.getVisibleCells().map((cell) => (
                              <TableCell
                                key={cell.id}
                                className={
                                  cell.column.id === 'actions'
                                    ? actionsCellClassName
                                    : tableCellClassName
                                }
                              >
                                {flexRender(
                                  cell.column.columnDef.cell,
                                  cell.getContext()
                                )}
                              </TableCell>
                            ))}
                          </TableRow>
                        )
                      })
                    ) : (
                      <TableRow>
                        <TableCell
                          colSpan={columns.length}
                          className='h-24 text-center'
                        >
                          No collections match the current filter.
                        </TableCell>
                      </TableRow>
                    )}
                  </TableBody>
                </Table>
              </div>
              <div
                className='mt-auto'
                data-testid='collections-table-pagination'
              >
                <DataTablePagination table={table} />
              </div>
            </CardContent>
          </Card>

          <Card
            className='flex min-h-0 flex-col overflow-hidden'
            data-testid='collections-members-panel'
          >
            <CardHeader className='shrink-0 py-3'>
              <CardTitle>Collection members</CardTitle>
              <CardDescription>
                Review inventory items assigned to the selected collection.
                Assign or move items from Inventory row actions.
              </CardDescription>
            </CardHeader>
            <CardContent className='flex min-h-0 flex-1 flex-col gap-3'>
              <div className='space-y-3'>
                <div data-testid='collections-members-table-toolbar'>
                  <DataTableToolbar
                    table={membersTable}
                    searchPlaceholder='Filter collection members...'
                    searchInputTestId='collections-members-search-input'
                  />
                </div>
                <p
                  className='text-sm text-muted-foreground'
                  data-testid='collections-members-summary'
                >
                  Showing {filteredMemberCount} of{' '}
                  {selectedCollectionItems.length} items.
                </p>
              </div>
              <div
                className='min-h-0 flex-1 overflow-auto rounded-md border'
                data-table-surface='true'
                data-testid='collections-members-table'
              >
                <Table className={tableClassName}>
                  <TableHeader>
                    {membersTable.getHeaderGroups().map((headerGroup) => (
                      <TableRow key={headerGroup.id}>
                        {headerGroup.headers.map((header) => (
                          <TableHead
                            key={header.id}
                            className={tableHeaderClassName}
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
                    {membersTable.getRowModel().rows.length ? (
                      membersTable.getRowModel().rows.map((row) => (
                        <TableRow
                          key={row.id}
                          data-testid={`collections-member-row-${row.original.id}`}
                        >
                          {row.getVisibleCells().map((cell) => (
                            <TableCell
                              key={cell.id}
                              className={tableCellClassName}
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
                      <TableRow data-testid='collections-members-empty-row'>
                        <TableCell
                          colSpan={memberColumns.length}
                          className='h-24 text-center text-muted-foreground'
                        >
                          {selectedCollectionItems.length
                            ? 'No collection members match the current filter.'
                            : `No items are currently assigned to ${selectedCollectionName}.`}
                        </TableCell>
                      </TableRow>
                    )}
                  </TableBody>
                </Table>
              </div>
              <div
                className='mt-auto'
                data-testid='collections-members-table-pagination'
              >
                <DataTablePagination table={membersTable} />
              </div>
            </CardContent>
          </Card>
        </div>
      </Main>

      <Dialog open={createOpen} onOpenChange={setCreateDialogOpen}>
        <DialogContent data-testid='collections-create-dialog'>
          <form
            className='space-y-4'
            onSubmit={(event) => {
              event.preventDefault()
              void handleCreateCollection()
            }}
          >
            <DialogHeader>
              <DialogTitle>Create collection</DialogTitle>
              <DialogDescription>
                Add a new collection row to the management table.
              </DialogDescription>
            </DialogHeader>
            <div className='space-y-2'>
              <Input
                ref={createInputRef}
                value={createValue}
                onChange={(event) => {
                  setCreateValue(event.target.value)
                  if (createError) {
                    setCreateError('')
                  }
                }}
                onKeyDown={(event) => {
                  if (event.key !== 'Enter') {
                    return
                  }
                  event.preventDefault()
                  void handleCreateCollection()
                }}
                placeholder='Collection name'
                aria-invalid={createError ? 'true' : undefined}
                aria-describedby={
                  createError ? 'collections-create-error' : undefined
                }
                data-testid='collections-create-input'
              />
              {createError ? (
                <p
                  id='collections-create-error'
                  className='text-sm text-destructive'
                  data-testid='collections-create-error'
                >
                  {createError}
                </p>
              ) : null}
            </div>
            <DialogFooter>
              <Button
                type='button'
                variant='outline'
                onClick={() => setCreateDialogOpen(false)}
                disabled={createSubmitting}
              >
                Cancel
              </Button>
              <Button
                type='submit'
                disabled={createSubmitting}
                data-testid='collections-create-submit'
              >
                {createSubmitting ? 'Saving...' : 'Save collection'}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <Sheet open={editOpen} onOpenChange={setEditOpen}>
        <SheetContent
          className='flex flex-col'
          data-testid='collections-edit-panel'
          data-side='right'
          side='right'
        >
          <SheetHeader className='text-start'>
            <SheetTitle>Edit collection</SheetTitle>
            <SheetDescription>
              Edit the selected collection metadata through the row workflow.
            </SheetDescription>
            <div className='flex flex-wrap items-center gap-2 pt-2'>
              <Button
                type='button'
                variant='outline'
                size='sm'
                data-testid='collections-edit-previous'
                disabled={!canNavigatePrevious}
                onClick={() => navigateEditPanel(-1)}
              >
                Previous
              </Button>
              <Button
                type='button'
                variant='outline'
                size='sm'
                data-testid='collections-edit-next'
                disabled={!canNavigateNext}
                onClick={() => navigateEditPanel(1)}
              >
                Next
              </Button>
            </div>
          </SheetHeader>
          <div className='flex-1 space-y-4 overflow-auto px-4'>
            <label className='space-y-2 text-sm font-medium'>
              <span>Collection name</span>
              <Input
                value={editValue}
                onChange={(event) => setEditValue(event.target.value)}
                placeholder='New collection name'
                data-testid='collections-edit-input'
              />
            </label>
            <label className='space-y-2 text-sm font-medium'>
              <span>Scope</span>
              <Input
                value={editScopeValue}
                onChange={(event) => setEditScopeValue(event.target.value)}
                placeholder='Collection scope'
                data-testid='collections-edit-scope-input'
              />
            </label>
            <label className='space-y-2 text-sm font-medium'>
              <span>Status</span>
              <Input
                value={editStatusValue}
                onChange={(event) => setEditStatusValue(event.target.value)}
                placeholder='Collection status'
                data-testid='collections-edit-status-input'
              />
            </label>
            <label className='space-y-2 text-sm font-medium'>
              <span>Description</span>
              <Textarea
                value={editDescriptionValue}
                onChange={(event) => setEditDescriptionValue(event.target.value)}
                placeholder='Collection description'
                data-testid='collections-edit-description-input'
              />
            </label>
          </div>
          <SheetFooter className='gap-2'>
            <Button
              type='button'
              variant='outline'
              onClick={() => setEditOpen(false)}
            >
              Cancel
            </Button>
            <Button
              type='button'
              onClick={() => {
                void handleRenameCollection()
              }}
              data-testid='collections-edit-submit'
            >
              Save collection
            </Button>
          </SheetFooter>
        </SheetContent>
      </Sheet>

      <Dialog open={deleteOpen} onOpenChange={setDeleteOpen}>
        <DialogContent data-testid='collections-delete-dialog'>
          <DialogHeader>
            <DialogTitle>Delete collection</DialogTitle>
          <DialogDescription>
              Hide the selected collection from active collection views.
            </DialogDescription>
          </DialogHeader>
          <div className='space-y-3 text-sm text-muted-foreground'>
            <p data-testid='collections-delete-message'>
              {selectedRow
                ? selectedRow.name === 'All Items'
                  ? 'All Items is protected and cannot be deleted.'
                  : selectedCollectionAssignedCount > 0
                    ? `${selectedRow.name} has ${selectedCollectionAssignedCount} assigned item${selectedCollectionAssignedCount === 1 ? '' : 's'}. Choose a destination collection to move them, or leave the destination blank to remove only the collection membership.`
                    : `${selectedRow.name} will be hidden from active collections and can be reviewed in the deleted filter.`
                : 'No collection is selected.'}
            </p>
            {selectedRow &&
            selectedRow.name !== 'All Items' &&
            selectedCollectionAssignedCount > 0 ? (
              <label className='block space-y-2 font-medium text-foreground'>
                <span>Move assigned items to</span>
                <select
                  value={deleteDestination}
                  onChange={(event) => setDeleteDestination(event.target.value)}
                  className='h-9 w-full rounded-md border border-input bg-background px-3 text-sm'
                  data-testid='collections-delete-destination-select'
                >
                  <option value=''>No destination, remove membership only</option>
                  {deleteDestinationOptions.map((row) => (
                    <option key={row.key} value={row.name}>
                      {row.name}
                    </option>
                  ))}
                </select>
              </label>
            ) : null}
          </div>
          <DialogFooter>
            <Button
              type='button'
              variant='outline'
              onClick={() => setDeleteOpen(false)}
            >
              Cancel
            </Button>
            <Button
              type='button'
              variant='destructive'
              onClick={() => {
                void handleDeleteCollection()
              }}
              data-testid='collections-delete-submit'
            >
              Hide collection
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
