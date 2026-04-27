import { useCallback, useMemo, useState } from 'react'
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
import { Pencil, Plus, Tag, Trash2 } from 'lucide-react'
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
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { LanguageSwitch } from '@/components/language-switch'
import { Header, HeaderTitle } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import {
  DataTableColumnHeader,
  DataTablePagination,
  DataTableToolbar,
} from '@/components/data-table'
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

function buildCollectionColumns({
  onEdit,
  onDelete,
}: {
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
        <div className='space-y-1'>
          <div
            className='font-medium'
            data-testid={`collections-row-name-${row.original.key}`}
          >
            {row.original.name}
          </div>
          <div className='text-xs text-muted-foreground'>
            {row.original.description}
          </div>
        </div>
      ),
    },
    {
      accessorKey: 'itemCount',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='Items' />
      ),
      cell: ({ row }) => (
        <span data-testid={`collections-row-count-${row.original.key}`}>
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
      header: 'Actions',
      cell: ({ row }) => (
        <div className='flex items-center justify-end gap-2'>
          <Button
            type='button'
            size='sm'
            variant='outline'
            data-testid={`collections-row-edit-${row.original.key}`}
            onClick={(event) => {
              event.stopPropagation()
              onEdit(row.original)
            }}
          >
            <Pencil className='mr-2 h-4 w-4' />
            Edit
          </Button>
          <Button
            type='button'
            size='sm'
            variant='outline'
            data-testid={`collections-row-delete-${row.original.key}`}
            onClick={(event) => {
              event.stopPropagation()
              onDelete(row.original)
            }}
          >
            <Trash2 className='mr-2 h-4 w-4' />
            Delete
          </Button>
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
          className='font-medium'
          data-testid={`collections-member-${collectionKey(row.original.name)}`}
        >
          <span
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
        <span className='text-muted-foreground'>{row.original.detail}</span>
      ),
    },
    {
      accessorKey: 'collectionName',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='Current collection' />
      ),
      cell: ({ row }) => (
        <span
          className='text-muted-foreground'
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
    renameCollection,
    removeCollection,
    collectionItems,
    collectionSummaries,
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
  const [createValue, setCreateValue] = useState('')
  const [editValue, setEditValue] = useState('')

  const rows = useMemo(() => collectionSummaries, [collectionSummaries])

  const selectedRow = useMemo(
    () =>
      rows.find((row) => row.key === selectedCollectionID) ??
      rows.find((row) => row.name === activeWorkspaceCollection) ??
      rows[0] ??
      null,
    [activeWorkspaceCollection, rows, selectedCollectionID]
  )

  const selectedCollectionName = selectedRow?.name ?? 'All Items'

  const selectedCollectionItems = useMemo(
    () =>
      selectedCollectionName === 'All Items'
        ? collectionItems
        : collectionItems.filter(
            (item) => item.collectionName === selectedCollectionName
          ),
    [collectionItems, selectedCollectionName]
  )

  const openEditPanel = useCallback((row: CollectionRow) => {
    setSelectedCollectionID(row.key)
    setEditValue(row.name)
    setEditOpen(true)
  }, [])

  const columns = useMemo(
    () =>
      buildCollectionColumns({
        onEdit: openEditPanel,
        onDelete: (row) => {
          setSelectedCollectionID(row.key)
          setDeleteOpen(true)
        },
      }),
    [openEditPanel]
  )
  const memberColumns = useMemo(() => buildCollectionMemberColumns(), [])

  // eslint-disable-next-line react-hooks/incompatible-library
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

  // eslint-disable-next-line react-hooks/incompatible-library
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
  }

  async function handleSelectCollection(row: CollectionRow) {
    setSelectedCollectionID(row.key)
    await setActiveWorkspaceCollection(row.name)
  }

  async function handleCreateCollection() {
    const created = await addCollection(createValue)
    if (!created) {
      toast.error('Collection name must be unique and non-empty.')
      return
    }
    setSelectedCollectionID(collectionKey(created))
    setCreateValue('')
    setCreateOpen(false)
    toast.success(`${created} created and set as the active collection.`)
  }

  async function handleRenameCollection() {
    if (!selectedRow) {
      return
    }
    const renamed = await renameCollection(selectedRow.name, editValue)
    if (!renamed) {
      toast.error('Rename failed. Use a unique non-empty collection name.')
      return
    }
    setSelectedCollectionID(collectionKey(renamed))
    setEditOpen(false)
    toast.success(`${selectedRow.name} renamed to ${renamed}.`)
  }

  async function handleDeleteCollection() {
    if (!selectedRow) {
      return
    }
    const removedName = selectedRow.name
    const removed = await removeCollection(removedName)
    if (!removed) {
      toast.error('Collection removal failed.')
      return
    }
    setSelectedCollectionID(collectionKey('All Items'))
    setDeleteOpen(false)
    toast.success(`${removedName} removed from workspace collections.`)
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
              onClick={() => setCreateOpen(true)}
            >
              <Plus className='mr-2 h-4 w-4' />
              New collection
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

      <Main className='flex flex-1 flex-col gap-3 sm:gap-4'>
        <div className='grid gap-4' data-testid='collections-workspace'>
          <Card data-testid='collections-section'>
            <CardContent className='space-y-4'>
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
                  <span>Showing {filteredCount} of {rows.length} collections.</span>
                  <span data-testid='collections-active-context'>
                    Active: {selectedCollectionName}
                  </span>
                </div>
              </div>

              <div
                className='overflow-x-auto rounded-md border'
                data-testid='collections-shared-table'
              >
                <Table>
                  <TableHeader>
                    {table.getHeaderGroups().map((headerGroup) => (
                      <TableRow key={headerGroup.id}>
                        {headerGroup.headers.map((header) => (
                          <TableHead key={header.id}>
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
                              <TableCell key={cell.id}>
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
            </CardContent>
          </Card>

          <Card data-testid='collections-members-panel'>
            <CardHeader>
              <CardTitle>Collection members</CardTitle>
              <CardDescription>
                Review inventory items assigned to the selected collection.
                Assign or move items from Inventory row actions.
              </CardDescription>
            </CardHeader>
            <CardContent className='space-y-4'>
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
                className='overflow-x-auto rounded-md border'
                data-testid='collections-members-table'
              >
                <Table>
                  <TableHeader>
                    {membersTable.getHeaderGroups().map((headerGroup) => (
                      <TableRow key={headerGroup.id}>
                        {headerGroup.headers.map((header) => (
                          <TableHead key={header.id}>
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
                            <TableCell key={cell.id}>
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
              <DataTablePagination table={membersTable} />
            </CardContent>
          </Card>
        </div>
      </Main>

      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent data-testid='collections-create-dialog'>
          <DialogHeader>
            <DialogTitle>Create collection</DialogTitle>
            <DialogDescription>
              Add a new collection row to the management table.
            </DialogDescription>
          </DialogHeader>
          <Input
            value={createValue}
            onChange={(event) => setCreateValue(event.target.value)}
            placeholder='Collection name'
            data-testid='collections-create-input'
          />
          <DialogFooter>
            <Button
              type='button'
              variant='outline'
              onClick={() => setCreateOpen(false)}
            >
              Cancel
            </Button>
            <Button
              type='button'
              onClick={() => {
                void handleCreateCollection()
              }}
              data-testid='collections-create-submit'
            >
              Save collection
            </Button>
          </DialogFooter>
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
              Rename the selected collection through the row workflow.
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
          <div className='flex-1 px-4'>
            <label className='space-y-2 text-sm font-medium'>
              <span>Collection name</span>
              <Input
                value={editValue}
                onChange={(event) => setEditValue(event.target.value)}
                placeholder='New collection name'
                data-testid='collections-edit-input'
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
              Save rename
            </Button>
          </SheetFooter>
        </SheetContent>
      </Sheet>

      <Dialog open={deleteOpen} onOpenChange={setDeleteOpen}>
        <DialogContent data-testid='collections-delete-dialog'>
          <DialogHeader>
            <DialogTitle>Delete collection</DialogTitle>
            <DialogDescription>
              Remove the selected collection row from the shared table surface.
            </DialogDescription>
          </DialogHeader>
          <p className='text-sm text-muted-foreground'>
            {selectedRow
              ? `Delete ${selectedRow.name}? Assigned items will remain in Cabinet and become unassigned.`
              : 'No collection is selected.'}
          </p>
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
              Delete collection
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
