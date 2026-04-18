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

      <Main className='flex flex-1 flex-col gap-4 sm:gap-6'>
        <div className='flex flex-wrap items-end justify-between gap-3'>
          <div>
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
