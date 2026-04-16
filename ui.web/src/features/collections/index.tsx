import { useMemo, useState } from 'react'
import {
  type ColumnDef,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  type SortingState,
  useReactTable,
} from '@tanstack/react-table'
import { Pencil, Plus, Trash2 } from 'lucide-react'
import { toast } from 'sonner'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { ConfigDrawer } from '@/components/config-drawer'
import { DataTableColumnHeader, DataTablePagination, DataTableToolbar } from '@/components/data-table'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Header } from '@/components/layout/header'
import { Input } from '@/components/ui/input'
import { LanguageSwitch } from '@/components/language-switch'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { ThemeSwitch } from '@/components/theme-switch'
import { useWorkspaceCollections } from './use-workspace-collections'

type CollectionRow = {
  id: string
  name: string
  itemCount: number
  scopeLabel: string
  statusLabel: string
  updatedLabel: string
  description: string
}

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
          <div className='font-medium' data-testid={`collections-row-name-${row.original.id}`}>
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
        <span data-testid={`collections-row-count-${row.original.id}`}>
          {row.original.itemCount}
        </span>
      ),
    },
    {
      accessorKey: 'scopeLabel',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='Scope' />
      ),
      cell: ({ row }) => <Badge variant='secondary'>{row.original.scopeLabel}</Badge>,
    },
    {
      accessorKey: 'statusLabel',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='Status' />
      ),
      cell: ({ row }) => <Badge variant='outline'>{row.original.statusLabel}</Badge>,
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
      enableSorting: false,
      enableHiding: false,
      cell: ({ row }) => (
        <div className='flex items-center justify-end gap-2'>
          <Button
            type='button'
            variant='outline'
            size='sm'
            data-testid={`collections-row-edit-${row.original.id}`}
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
            variant='outline'
            size='sm'
            data-testid={`collections-row-delete-${row.original.id}`}
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

export function Collections() {
  const {
    workspaceCollections,
    activeWorkspaceCollection,
    setActiveWorkspaceCollection,
    addCollection,
    renameCollection,
    removeCollection,
    collectionSummaries,
  } = useWorkspaceCollections()
  const [globalFilter, setGlobalFilter] = useState('')
  const [sorting, setSorting] = useState<SortingState>([
    { id: 'name', desc: false },
  ])
  const [createOpen, setCreateOpen] = useState(false)
  const [editOpen, setEditOpen] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [createValue, setCreateValue] = useState('')
  const [editValue, setEditValue] = useState('')
  const [selectedCollectionID, setSelectedCollectionID] = useState<string | null>(null)

  const rows = useMemo<CollectionRow[]>(() => {
    return collectionSummaries.map((summary) => ({
      id: summary.key,
      name: summary.name,
      itemCount: summary.itemCount,
      scopeLabel: summary.scopeLabel,
      statusLabel: summary.statusLabel,
      updatedLabel: summary.updatedLabel,
      description: summary.description,
    }))
  }, [collectionSummaries])

  const selectedRow = useMemo(() => {
    return (
      rows.find((row) => row.id === selectedCollectionID) ??
      rows.find((row) => row.name === activeWorkspaceCollection) ??
      rows[0] ??
      null
    )
  }, [activeWorkspaceCollection, rows, selectedCollectionID])

  const columns = useMemo(
    () =>
      buildCollectionColumns({
        onEdit: (row) => {
          setSelectedCollectionID(row.id)
          setEditValue(row.name)
          setEditOpen(true)
        },
        onDelete: (row) => {
          setSelectedCollectionID(row.id)
          setDeleteOpen(true)
        },
      }),
    []
  )

<<<<<<< HEAD
  const table = useReactTable({
    data: rows,
    columns,
    state: {
      globalFilter,
      sorting,
    },
    onGlobalFilterChange: setGlobalFilter,
    onSortingChange: setSorting,
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
  })

  const handleSelectRow = (row: CollectionRow) => {
    setSelectedCollectionID(row.id)
    setActiveWorkspaceCollection(row.name)
  }

  const handleCreateCollection = () => {
    const created = addCollection(createValue)
    if (!created) {
      toast.error('Collection name must be unique and non-empty.')
      return
    }
    setActiveWorkspaceCollection(created)
    setSelectedCollectionID(
      collectionSummaries.find((summary) => summary.name === created)?.key ?? null
    )
    setCreateValue('')
    setCreateOpen(false)
    toast.success(`Created collection ${created}.`)
  }
=======
  const activeCollectionSummary = useMemo(
    () =>
      collectionSummaries.find((collection) => collection.name === activeWorkspaceCollection) ??
      collectionSummaries[0],
    [activeWorkspaceCollection, collectionSummaries]
  )
  const activeCollectionName =
    activeWorkspaceCollection.trim() || activeCollectionSummary?.name || 'All Items'

  useEffect(() => {
    setActiveContextMessage(
      `Active collection is ${activeCollectionName}. This choice persists for the current signed-in profile.`
    )
  }, [activeCollectionName])
>>>>>>> 28e0916 (#443 Fix collections persistence and tag icon contracts)

  const handleRenameCollection = () => {
    if (!selectedRow) {
      return
    }
    const renamed = renameCollection(selectedRow.name, editValue)
    if (!renamed) {
      toast.error('Rename failed. Use a unique non-empty collection name.')
      return
    }
    setActiveWorkspaceCollection(renamed)
    setSelectedCollectionID(
      collectionSummaries.find((summary) => summary.name === renamed)?.key ?? selectedRow.id
    )
    setEditOpen(false)
    setEditValue('')
    toast.success(`Renamed collection to ${renamed}.`)
  }

  const handleDeleteCollection = () => {
    if (!selectedRow) {
      return
    }
    const deleted = removeCollection(selectedRow.name)
    if (!deleted) {
      toast.error('Collection could not be deleted.')
      return
    }
    setDeleteOpen(false)
    setSelectedCollectionID(null)
    toast.success(`Deleted collection ${selectedRow.name}.`)
  }

  const filteredCount = table.getFilteredRowModel().rows.length

  return (
    <>
      <Header fixed>
        <Search />
        <div className='ms-auto flex items-center space-x-4'>
          <LanguageSwitch />
          <ThemeSwitch />
          <ConfigDrawer />
          <ProfileDropdown />
        </div>
      </Header>

<<<<<<< HEAD
      <Main className='flex flex-1 flex-col gap-4 sm:gap-6'>
        <div className='flex flex-wrap items-end justify-between gap-3'>
          <div>
=======
      <Main className='space-y-4'>
        <div className='space-y-1'>
          <div className='flex items-center gap-2'>
            <Tag
              data-testid='collections-page-icon'
              data-lucide='tag'
              className='h-5 w-5 text-muted-foreground'
            />
>>>>>>> 28e0916 (#443 Fix collections persistence and tag icon contracts)
            <h1 className='text-2xl font-bold tracking-tight'>Collections</h1>
            <p className='text-muted-foreground'>
              Manage collection rows through the shared Cabinet table surface.
            </p>
          </div>
          <Button
            type='button'
            data-testid='collections-new-action'
            onClick={() => setCreateOpen(true)}
          >
            <Plus className='mr-2 h-4 w-4' />
            New Collection
          </Button>
        </div>

        <div className='grid gap-4 xl:grid-cols-[minmax(0,1fr)_320px]'>
          <Card>
            <CardHeader>
              <CardTitle>Collection management table</CardTitle>
              <CardDescription>
                Browse, create, and edit collection rows from the standard table workflow.
              </CardDescription>
            </CardHeader>
            <CardContent className='space-y-4'>
              <DataTableToolbar
                table={table}
                searchPlaceholder='Filter collections...'
                filters={[]}
              />
              <div className='rounded-md border' data-testid='collections-shared-table'>
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
                    {table.getRowModel().rows.length > 0 ? (
                      table.getRowModel().rows.map((row) => {
                        const isActive = selectedRow?.id === row.original.id
                        return (
                          <TableRow
                            key={row.id}
                            data-testid={`collections-row-${row.original.id}`}
                            data-state={isActive ? 'selected' : undefined}
                            className='cursor-pointer'
                            onClick={() => handleSelectRow(row.original)}
                          >
                            {row.getVisibleCells().map((cell) => (
                              <TableCell key={cell.id}>
                                {flexRender(cell.column.columnDef.cell, cell.getContext())}
                              </TableCell>
                            ))}
                          </TableRow>
                        )
                      })
                    ) : (
                      <TableRow>
                        <TableCell colSpan={columns.length} className='h-24 text-center'>
                          No collections match the current filter.
                        </TableCell>
                      </TableRow>
                    )}
                  </TableBody>
                </Table>
              </div>
              <DataTablePagination table={table} />
              <p className='text-xs text-muted-foreground' data-testid='collections-filtered-count'>
                Showing {filteredCount} of {workspaceCollections.length} collections.
              </p>
<<<<<<< HEAD
            </CardContent>
          </Card>

          <Card data-testid='collections-selected-panel'>
            <CardHeader>
              <CardTitle>Selected collection</CardTitle>
              <CardDescription>
                Keep the current management context visible while working the table.
              </CardDescription>
            </CardHeader>
            <CardContent className='space-y-3'>
              {selectedRow ? (
                <>
                  <div>
                    <div className='text-sm text-muted-foreground'>Name</div>
                    <div className='font-medium' data-testid='collections-selected-name'>
                      {selectedRow.name}
=======
            </div>
            {collectionSummaries.length ? (
              <div
                className='rounded-md border border-primary/30 bg-primary/5 px-3 py-3 text-sm'
                data-testid='collections-active-context-panel'
              >
                <div className='flex flex-wrap items-start justify-between gap-3'>
                  <div>
                    <p className='font-medium'>Current active collection</p>
                    <p
                      className='mt-1 text-base font-semibold text-foreground'
                      data-testid='collections-active-context-name'
                    >
                      {activeCollectionName}
                    </p>
                    <p
                      className='mt-1 text-muted-foreground'
                      data-testid='collections-active-context-message'
                    >
                      {activeContextMessage}
                    </p>
                  </div>
                  <div
                    className='rounded-full border border-primary/30 px-2 py-1 text-xs text-primary'
                    data-testid='collections-active-context-persistence'
                  >
                    Persists for this signed-in profile
                  </div>
                </div>
              </div>
            ) : null}
            {createError ? (
              <div
                className='rounded-md border border-destructive/40 bg-destructive/10 px-3 py-2 text-sm text-destructive'
                data-testid='collections-create-error'
              >
                {createError}
              </div>
            ) : null}
            {createPanelOpen ? (
              <div
                className='rounded-md border bg-muted/20 p-3'
                data-testid='collections-create-panel'
              >
                <div className='space-y-3'>
                  <div>
                    <p className='font-medium' data-testid='collections-create-panel-title'>
                      Create a collection
                    </p>
                    <p
                      className='text-sm text-muted-foreground'
                      data-testid='collections-create-panel-description'
                    >
                      Saving creates the collection immediately, closes this panel, and makes the new collection active.
                    </p>
                  </div>
                  <div className='flex flex-wrap gap-2'>
                    <Input
                      data-testid='collections-new-input'
                      placeholder='Collection name'
                      value={newCollectionName}
                      onChange={(event) => {
                        setNewCollectionName(event.target.value)
                        if (createError) {
                          setCreateError(null)
                        }
                      }}
                    />
                    <Button
                      data-testid='collections-new-save'
                      onClick={() => {
                        const trimmedName = newCollectionName.trim()
                        if (!trimmedName) {
                          setCreateError('Enter a collection name before saving.')
                          return
                        }
                        addCollection(trimmedName)
                        setNewCollectionName('')
                        setCreatePanelOpen(false)
                        setEditingCollectionName(null)
                        setPendingRemoveName(null)
                        setCreateError(null)
                        toast.success(
                          `${trimmedName} created and set as the active collection.`
                        )
                      }}
                    >
                      Save
                    </Button>
                    <Button
                      variant='outline'
                      data-testid='collections-new-cancel'
                      onClick={() => {
                        setCreatePanelOpen(false)
                        setNewCollectionName('')
                        setCreateError(null)
                        toast.message('Collection creation cancelled. No collection was added.')
                      }}
                    >
                      Cancel
                    </Button>
                  </div>
                </div>
              </div>
            ) : null}
            {editingCollectionName ? (
              <div className='rounded-md border bg-muted/20 p-3' data-testid='collections-rename-panel'>
                <div className='space-y-3'>
                  <div>
                    <p className='font-medium'>Rename collection</p>
                    <p className='text-sm text-muted-foreground'>
                      Rename updates the collection label in place and preserves active context when applicable.
                    </p>
                  </div>
                  <div className='flex flex-wrap gap-2'>
                    <Input
                      data-testid='collections-rename-input'
                      value={renameValue}
                      onChange={(event) => {
                        setRenameValue(event.target.value)
                        if (createError) {
                          setCreateError(null)
                        }
                      }}
                    />
                    <Button
                      data-testid='collections-rename-save'
                      onClick={() => {
                        const renamed = renameCollection(editingCollectionName, renameValue)
                        if (!renamed) {
                          setCreateError('Enter a new collection name before saving rename.')
                          return
                        }
                        setEditingCollectionName(null)
                        setRenameValue('')
                        setCreateError(null)
                        toast.success(`${editingCollectionName} renamed to ${renamed}.`)
                      }}
                    >
                      Save rename
                    </Button>
                    <Button
                      variant='outline'
                      data-testid='collections-rename-cancel'
                      onClick={() => {
                        setEditingCollectionName(null)
                        setRenameValue('')
                        setCreateError(null)
                        toast.message('Collection rename cancelled. No changes were made.')
                      }}
                    >
                      Cancel
                    </Button>
                  </div>
                </div>
              </div>
            ) : null}
            {pendingRemoveName ? (
              <div className='rounded-md border bg-muted/20 p-3' data-testid='collections-remove-panel'>
                <div className='space-y-3'>
                  <div>
                    <p className='font-medium'>Remove collection</p>
                    <p className='text-sm text-muted-foreground'>
                      Confirm removal of {pendingRemoveName}. Active context falls back to All Items if needed.
                    </p>
                  </div>
                  <div className='flex flex-wrap gap-2'>
                    <Button
                      variant='destructive'
                      data-testid='collections-remove-confirm'
                      onClick={() => {
                        const removed = removeCollection(pendingRemoveName)
                        if (!removed) {
                          setCreateError('Collection removal could not be completed.')
                          return
                        }
                        const removedName = pendingRemoveName
                        setPendingRemoveName(null)
                        setEditingCollectionName(null)
                        setRenameValue('')
                        setCreateError(null)
                        toast.success(`${removedName} removed from workspace collections.`)
                      }}
                    >
                      Confirm remove
                    </Button>
                    <Button
                      variant='outline'
                      data-testid='collections-remove-cancel'
                      onClick={() => {
                        setPendingRemoveName(null)
                        toast.message('Collection removal cancelled. No collection was removed.')
                      }}
                    >
                      Cancel
                    </Button>
                  </div>
                </div>
              </div>
            ) : null}
            <div className='grid grid-cols-1 gap-2 md:grid-cols-2 lg:grid-cols-3'>
              {filteredCollections.length ? filteredCollections.map((collection) => {
                const isActive = activeWorkspaceCollection === collection.name
                const isProtected = collection.name === 'All Items'
                return (
                  <div
                    key={collection.name}
                    className={isActive ? 'rounded-md border-2 border-primary bg-primary/5 px-3 py-3' : 'rounded-md border px-3 py-3'}
                    data-testid={`collections-item-${collection.key}`}
                    data-state={isActive ? 'active' : 'inactive'}
                  >
                    <button
                      type='button'
                      data-testid={`collections-select-trigger-${collection.key}`}
                      className='w-full text-left hover:bg-muted/40'
                      onClick={() => {
                        setActiveWorkspaceCollection(collection.name)
                        setActiveContextMessage(
                          `Active collection switched to ${collection.name}. This choice persists for the current signed-in profile.`
                        )
                      }}
                    >
                      <div className='flex items-start justify-between gap-3'>
                        <div className='min-w-0'>
                          <div
                            className='font-medium'
                            data-testid={`collections-item-title-${collection.key}`}
                          >
                            {collection.name}
                          </div>
                          <div
                            className='mt-1 text-sm text-muted-foreground'
                            data-testid={`collections-item-description-${collection.key}`}
                          >
                            {collection.description}
                          </div>
                        </div>
                        <div
                          className='rounded-full border px-2 py-1 text-xs text-muted-foreground'
                          data-testid={`collections-item-status-${collection.key}`}
                        >
                          {collection.statusLabel}
                        </div>
                      </div>
                      <div
                        className='mt-3 flex flex-wrap gap-x-3 gap-y-1 text-xs text-muted-foreground'
                        data-testid={`collections-item-metadata-${collection.key}`}
                      >
                        <span data-testid={`collections-item-count-${collection.key}`}>
                          {collection.itemCount} items
                        </span>
                        <span data-testid={`collections-item-scope-${collection.key}`}>
                          {collection.scopeLabel}
                        </span>
                        <span data-testid={`collections-item-updated-${collection.key}`}>
                          {collection.updatedLabel}
                        </span>
                      </div>
                    </button>
                    <div className='mt-3 flex flex-wrap gap-2'>
                      <Button
                        type='button'
                        size='sm'
                        variant='outline'
                        data-testid={`collections-rename-trigger-${collection.key}`}
                        disabled={isProtected}
                        onClick={() => {
                          setEditingCollectionName(collection.name)
                          setRenameValue(collection.name)
                          setPendingRemoveName(null)
                          setCreateError(null)
                        }}
                      >
                        <Pencil className='mr-1 size-4' /> Rename
                      </Button>
                      <Button
                        type='button'
                        size='sm'
                        variant='outline'
                        data-testid={`collections-remove-trigger-${collection.key}`}
                        disabled={isProtected}
                        onClick={() => {
                          setPendingRemoveName(collection.name)
                          setEditingCollectionName(null)
                          setRenameValue('')
                          setCreateError(null)
                        }}
                      >
                        <Trash2 className='mr-1 size-4' /> Remove
                      </Button>
>>>>>>> 28e0916 (#443 Fix collections persistence and tag icon contracts)
                    </div>
                  </div>
                  <div className='grid grid-cols-2 gap-3 text-sm'>
                    <div>
                      <div className='text-muted-foreground'>Items</div>
                      <div data-testid='collections-selected-count'>{selectedRow.itemCount}</div>
                    </div>
                    <div>
                      <div className='text-muted-foreground'>Updated</div>
                      <div>{selectedRow.updatedLabel}</div>
                    </div>
                  </div>
                  <div className='flex flex-wrap gap-2'>
                    <Badge variant='secondary'>{selectedRow.scopeLabel}</Badge>
                    <Badge variant='outline'>{selectedRow.statusLabel}</Badge>
                  </div>
                  <p className='text-sm text-muted-foreground'>
                    {selectedRow.description}
                  </p>
                  <div className='flex gap-2'>
                    <Button
                      type='button'
                      variant='outline'
                      data-testid='collections-selected-edit'
                      onClick={() => {
                        setEditValue(selectedRow.name)
                        setEditOpen(true)
                      }}
                    >
                      <Pencil className='mr-2 h-4 w-4' />
                      Edit
                    </Button>
                    <Button
                      type='button'
                      variant='outline'
                      data-testid='collections-selected-delete'
                      onClick={() => setDeleteOpen(true)}
                    >
                      <Trash2 className='mr-2 h-4 w-4' />
                      Delete
                    </Button>
                  </div>
                </>
              ) : (
                <p className='text-sm text-muted-foreground'>
                  Select a collection row to manage it here.
                </p>
              )}
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
            data-testid='collections-create-input'
            placeholder='Collection name'
            value={createValue}
            onChange={(event) => setCreateValue(event.target.value)}
          />
          <DialogFooter>
            <Button type='button' variant='outline' onClick={() => setCreateOpen(false)}>
              Cancel
            </Button>
            <Button
              type='button'
              data-testid='collections-create-submit'
              onClick={handleCreateCollection}
            >
              Create
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={editOpen} onOpenChange={setEditOpen}>
        <DialogContent data-testid='collections-edit-dialog'>
          <DialogHeader>
            <DialogTitle>Edit collection</DialogTitle>
            <DialogDescription>
              Rename the selected collection through the row management workflow.
            </DialogDescription>
          </DialogHeader>
          <Input
            data-testid='collections-edit-input'
            placeholder='Collection name'
            value={editValue}
            onChange={(event) => setEditValue(event.target.value)}
          />
          <DialogFooter>
            <Button type='button' variant='outline' onClick={() => setEditOpen(false)}>
              Cancel
            </Button>
            <Button
              type='button'
              data-testid='collections-edit-submit'
              onClick={handleRenameCollection}
            >
              Save
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

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
              ? `Delete ${selectedRow.name}? This should immediately remove it from the collections table.`
              : 'No collection is selected.'}
          </p>
          <DialogFooter>
            <Button type='button' variant='outline' onClick={() => setDeleteOpen(false)}>
              Cancel
            </Button>
            <Button
              type='button'
              variant='destructive'
              data-testid='collections-delete-submit'
              onClick={handleDeleteCollection}
            >
              Delete
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
