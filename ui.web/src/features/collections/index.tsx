import { useMemo, useState } from 'react'
import {
  type ColumnDef,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
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
import { Header } from '@/components/layout/header'
import { Input } from '@/components/ui/input'
import { LanguageSwitch } from '@/components/language-switch'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { Separator } from '@/components/ui/separator'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { ThemeSwitch } from '@/components/theme-switch'
import {
  type WorkspaceCollectionSummary,
  collectionKey,
  useWorkspaceCollections,
} from './use-workspace-collections'

const collectionsHeaderDescription =
  'Manage collection rows and item placement from the shared Cabinet table surface.'

type CollectionRow = WorkspaceCollectionSummary

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
      header: 'Collection',
      cell: ({ row }) => (
        <div className='space-y-1'>
          <div className='font-medium' data-testid={`collections-row-name-${row.original.key}`}>
            {row.original.name}
          </div>
          <div className='text-xs text-muted-foreground'>{row.original.description}</div>
        </div>
      ),
    },
    {
      accessorKey: 'itemCount',
      header: 'Items',
      cell: ({ row }) => (
        <span data-testid={`collections-row-count-${row.original.key}`}>{row.original.itemCount}</span>
      ),
    },
    {
      accessorKey: 'scopeLabel',
      header: 'Scope',
    },
    {
      accessorKey: 'statusLabel',
      header: 'Status',
    },
    {
      accessorKey: 'updatedLabel',
      header: 'Updated',
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

  const [sorting, setSorting] = useState<SortingState>([{ id: 'name', desc: false }])
  const [globalFilter, setGlobalFilter] = useState('')
  const [selectedCollectionID, setSelectedCollectionID] = useState<string | null>(null)
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
        : collectionItems.filter((item) => item.collectionName === selectedCollectionName),
    [collectionItems, selectedCollectionName]
  )

  const columns = useMemo(
    () =>
      buildCollectionColumns({
        onEdit: (row) => {
          setSelectedCollectionID(row.key)
          setEditValue(row.name)
          setEditOpen(true)
        },
        onDelete: (row) => {
          setSelectedCollectionID(row.key)
          setDeleteOpen(true)
        },
      }),
    []
  )

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
  })

  const filteredCount = table.getFilteredRowModel().rows.length

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
        <h1
          className='pointer-events-auto absolute left-1/2 top-1/2 hidden -translate-x-1/2 -translate-y-1/2 items-center gap-2 whitespace-nowrap text-lg font-bold tracking-tight lg:flex'
          data-testid='collections-header-title'
          title={collectionsHeaderDescription}
          aria-label={`Collections - ${collectionsHeaderDescription}`}
        >
          <Tag
            aria-hidden='true'
            className='h-5 w-5 text-muted-foreground'
            data-testid='collections-page-icon'
          />
          Collections
        </h1>
        <div className='ml-auto flex min-w-0 items-center gap-3'>
          <div
            className='flex min-w-0 flex-wrap items-center justify-end gap-2'
            data-testid='collections-global-header-actions'
          >
            <div className='mr-1 flex min-w-0 flex-wrap items-center gap-x-3 gap-y-1'>
              <span className='text-xs font-medium text-muted-foreground'>
                Active:
              </span>
              <span
                className='text-xs font-medium text-muted-foreground'
                data-testid='collections-active-context'
              >
                {selectedCollectionName}
              </span>
            </div>
            <Button data-testid='collections-new-action' onClick={() => setCreateOpen(true)}>
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
        <div
          className='grid gap-4'
          data-testid='collections-workspace'
        >
          <Card data-testid='collections-section'>
            <CardHeader>
              <CardTitle>Collections table</CardTitle>
              <CardDescription>
                Browse, create, rename, and remove collection rows from the same management surface.
              </CardDescription>
            </CardHeader>
            <CardContent className='space-y-4'>
              <div className='flex items-center gap-3' data-testid='collections-management-tools'>
                <Input
                  value={globalFilter}
                  onChange={(event) => setGlobalFilter(event.target.value)}
                  placeholder='Filter collections...'
                  data-testid='collections-search-input'
                />
                <div className='text-sm text-muted-foreground' data-testid='collections-management-summary'>
                  Showing {filteredCount} of {rows.length} collections.
                </div>
              </div>

              <div className='rounded-md border' data-testid='collections-shared-table'>
                <Table>
                  <TableHeader>
                    {table.getHeaderGroups().map((headerGroup) => (
                      <TableRow key={headerGroup.id}>
                        {headerGroup.headers.map((header) => (
                          <TableHead key={header.id}>
                            {header.isPlaceholder
                              ? null
                              : flexRender(header.column.columnDef.header, header.getContext())}
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
            </CardContent>
          </Card>

          <Card data-testid='collections-members-panel'>
            <CardHeader>
              <CardTitle>Collection members</CardTitle>
              <CardDescription>
                Review inventory items assigned to the selected collection. Assign or
                move items from Inventory row actions.
              </CardDescription>
            </CardHeader>
            <CardContent className='space-y-4'>
              <div className='rounded-md border' data-testid='collections-members-table'>
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Item</TableHead>
                      <TableHead>Details</TableHead>
                      <TableHead>Current collection</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {selectedCollectionItems.length ? (
                      selectedCollectionItems.map((item) => (
                        <TableRow
                          key={item.id}
                          data-testid={`collections-member-row-${item.id}`}
                        >
                          <TableCell
                            className='font-medium'
                            data-testid={`collections-member-${collectionKey(item.name)}`}
                          >
                            <span data-testid={`collections-member-name-${collectionKey(item.name)}`}>
                              {item.name}
                            </span>
                          </TableCell>
                          <TableCell className='text-muted-foreground'>
                            {item.detail}
                          </TableCell>
                          <TableCell
                            className='text-muted-foreground'
                            data-testid={`collections-member-current-${collectionKey(item.name)}`}
                          >
                            Currently in {item.collectionName ?? 'Unassigned'}.
                          </TableCell>
                        </TableRow>
                      ))
                    ) : (
                      <TableRow data-testid='collections-members-empty-row'>
                        <TableCell colSpan={3} className='h-24 text-center text-muted-foreground'>
                          No items are currently assigned to {selectedCollectionName}.
                        </TableCell>
                      </TableRow>
                    )}
                  </TableBody>
                </Table>
              </div>
            </CardContent>
          </Card>
        </div>
      </Main>

      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent data-testid='collections-create-dialog'>
          <DialogHeader>
            <DialogTitle>Create collection</DialogTitle>
            <DialogDescription>Add a new collection row to the management table.</DialogDescription>
          </DialogHeader>
          <Input
            value={createValue}
            onChange={(event) => setCreateValue(event.target.value)}
            placeholder='Collection name'
            data-testid='collections-create-input'
          />
          <DialogFooter>
            <Button type='button' variant='outline' onClick={() => setCreateOpen(false)}>
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

      <Dialog open={editOpen} onOpenChange={setEditOpen}>
        <DialogContent data-testid='collections-edit-dialog'>
          <DialogHeader>
            <DialogTitle>Edit collection</DialogTitle>
            <DialogDescription>Rename the selected collection through the row workflow.</DialogDescription>
          </DialogHeader>
          <Input
            value={editValue}
            onChange={(event) => setEditValue(event.target.value)}
            placeholder='New collection name'
            data-testid='collections-edit-input'
          />
          <DialogFooter>
            <Button type='button' variant='outline' onClick={() => setEditOpen(false)}>
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
              ? `Delete ${selectedRow.name}? Assigned items will remain in Cabinet and become unassigned.`
              : 'No collection is selected.'}
          </p>
          <DialogFooter>
            <Button type='button' variant='outline' onClick={() => setDeleteOpen(false)}>
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
