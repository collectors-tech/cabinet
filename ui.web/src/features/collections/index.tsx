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
import { ArrowRightLeft, Pencil, Plus, Tag, Trash2 } from 'lucide-react'
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
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
  type WorkspaceCollectionItem,
  type WorkspaceCollectionSummary,
  collectionKey,
  useWorkspaceCollections,
} from './use-workspace-collections'

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
    assignItemToCollection,
    unassignItemFromCollection,
  } = useWorkspaceCollections()

  const [sorting, setSorting] = useState<SortingState>([{ id: 'name', desc: false }])
  const [globalFilter, setGlobalFilter] = useState('')
  const [selectedCollectionID, setSelectedCollectionID] = useState<string | null>(null)
  const [createOpen, setCreateOpen] = useState(false)
  const [editOpen, setEditOpen] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [createValue, setCreateValue] = useState('')
  const [editValue, setEditValue] = useState('')
  const [assignmentItemID, setAssignmentItemID] = useState('')
  const [moveTargets, setMoveTargets] = useState<Record<string, string>>({})

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
  const isProtectedCollection = selectedCollectionName === 'All Items'

  const selectedCollectionItems = useMemo(
    () =>
      collectionItems.filter((item) => item.collectionName === selectedCollectionName),
    [collectionItems, selectedCollectionName]
  )

  const assignmentCandidates = useMemo(
    () =>
      collectionItems.filter((item) => item.collectionName !== selectedCollectionName),
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

  async function handleAssignToSelectedCollection() {
    if (!assignmentItemID || !selectedRow || isProtectedCollection) {
      return
    }
    const updated = await assignItemToCollection(assignmentItemID, selectedRow.name)
    if (!updated) {
      toast.error('Select an item and a valid collection before saving.')
      return
    }
    setAssignmentItemID('')
    toast.success(`${updated.name} assigned to ${selectedRow.name}.`)
  }

  async function handleMoveItem(item: WorkspaceCollectionItem) {
    const destination = moveTargets[item.id]
    if (!destination) {
      toast.error('Choose a destination collection before moving the item.')
      return
    }
    const updated = await assignItemToCollection(item.id, destination)
    if (!updated) {
      toast.error('Item move failed.')
      return
    }
    setMoveTargets((current) => {
      const next = { ...current }
      delete next[item.id]
      return next
    })
    toast.success(`${item.name} moved to ${destination}.`)
  }

  async function handleUnassignItem(item: WorkspaceCollectionItem) {
    const updated = await unassignItemFromCollection(item.id)
    if (!updated) {
      toast.error('Item release failed.')
      return
    }
    setMoveTargets((current) => {
      const next = { ...current }
      delete next[item.id]
      return next
    })
    toast.success(`${item.name} removed from ${selectedCollectionName}.`)
  }

  return (
    <>
      <Header>
        <Search />
        <div className='ml-auto flex items-center gap-4'>
          <ThemeSwitch />
          <LanguageSwitch />
          <ProfileDropdown />
        </div>
      </Header>

      <Main className='flex flex-1 flex-col gap-3 sm:gap-4'>
        <div
          className='flex flex-wrap items-center justify-between gap-2'
          data-testid='collections-page-header'
        >
          <div className='flex min-w-0 flex-wrap items-center gap-x-3 gap-y-1'>
            <Tag
              className='h-5 w-5 text-muted-foreground'
              data-testid='collections-page-icon'
            />
            <h1 className='text-xl font-bold tracking-tight'>Collections</h1>
            <span className='text-xs font-medium text-muted-foreground'>
              {selectedCollectionName}
            </span>
          </div>
          <Button data-testid='collections-new-action' onClick={() => setCreateOpen(true)}>
            <Plus className='mr-2 h-4 w-4' />
            New collection
          </Button>
        </div>

        <div
          className='grid gap-6 lg:grid-cols-[1.5fr,1fr]'
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

          <div className='space-y-6'>
            <Card data-testid='collections-selected-panel'>
              <CardHeader>
                <CardTitle>Selected collection</CardTitle>
                <CardDescription>
                  Keep the active collection context visible while you manage rows and item placement.
                </CardDescription>
              </CardHeader>
              <CardContent className='space-y-3'>
                {selectedRow ? (
                  <>
                    <div className='font-medium' data-testid='collections-selected-name'>
                      {selectedRow.name}
                    </div>
                    <div className='text-sm text-muted-foreground'>
                      {selectedRow.description}
                    </div>
                    <div className='text-sm text-muted-foreground' data-testid='collections-active-context-message'>
                      Active collection is {selectedRow.name}. This choice persists for the current signed-in profile.
                    </div>
                    <div className='text-xs text-muted-foreground' data-testid='collections-active-context-persistence'>
                      Persists for this signed-in profile across refresh.
                    </div>
                  </>
                ) : (
                  <div className='text-sm text-muted-foreground'>Select a collection row to manage it here.</div>
                )}
              </CardContent>
            </Card>

            <Card data-testid='collections-assignment-panel'>
              <CardHeader>
                <CardTitle>Assign items</CardTitle>
                <CardDescription>
                  Add an item into the selected collection or move an item into this collection from another lane.
                </CardDescription>
              </CardHeader>
              <CardContent className='space-y-4'>
                {isProtectedCollection ? (
                  <div className='text-sm text-muted-foreground' data-testid='collections-assignment-disabled'>
                    Select a specific collection to assign items. “All Items” stays as the global overview.
                  </div>
                ) : (
                  <>
                    <Select value={assignmentItemID} onValueChange={setAssignmentItemID}>
                      <SelectTrigger data-testid='collections-assignment-select'>
                        <SelectValue placeholder='Choose an item to assign' />
                      </SelectTrigger>
                      <SelectContent>
                        {assignmentCandidates.map((item) => (
                          <SelectItem key={item.id} value={item.id}>
                            {item.name} ({item.collectionName ?? 'Unassigned'})
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                    <Button
                      type='button'
                      onClick={() => {
                        void handleAssignToSelectedCollection()
                      }}
                      data-testid='collections-assignment-submit'
                    >
                      Assign to {selectedCollectionName}
                    </Button>
                  </>
                )}
              </CardContent>
            </Card>

            <Card data-testid='collections-members-panel'>
              <CardHeader>
                <CardTitle>Collection members</CardTitle>
                <CardDescription>
                  Review assigned items and move them deterministically between collections.
                </CardDescription>
              </CardHeader>
              <CardContent className='space-y-3'>
                {selectedCollectionItems.length ? (
                  selectedCollectionItems.map((item) => (
                    <div
                      key={item.id}
                      className='rounded-md border p-3'
                      data-testid={`collections-member-${collectionKey(item.name)}`}
                    >
                      <div className='font-medium' data-testid={`collections-member-name-${collectionKey(item.name)}`}>
                        {item.name}
                      </div>
                      <div className='text-sm text-muted-foreground'>{item.detail}</div>
                      <div
                        className='mt-2 text-xs text-muted-foreground'
                        data-testid={`collections-member-current-${collectionKey(item.name)}`}
                      >
                        Currently in {item.collectionName ?? 'Unassigned'}.
                      </div>
                      <div className='mt-3 flex items-center gap-3'>
                        <Select
                          value={moveTargets[item.id] ?? ''}
                          onValueChange={(value) =>
                            setMoveTargets((current) => ({ ...current, [item.id]: value }))
                          }
                        >
                          <SelectTrigger data-testid={`collections-move-target-${collectionKey(item.name)}`}>
                            <SelectValue placeholder='Move to...' />
                          </SelectTrigger>
                          <SelectContent>
                            {rows
                              .filter((row) => row.name !== 'All Items' && row.name !== item.collectionName)
                              .map((row) => (
                                <SelectItem key={row.key} value={row.name}>
                                  {row.name}
                                </SelectItem>
                              ))}
                          </SelectContent>
                        </Select>
                        <Button
                          type='button'
                          variant='outline'
                          data-testid={`collections-move-submit-${collectionKey(item.name)}`}
                          onClick={() => {
                            void handleMoveItem(item)
                          }}
                        >
                          <ArrowRightLeft className='mr-2 h-4 w-4' />
                          Move item
                        </Button>
                        <Button
                          type='button'
                          variant='outline'
                          data-testid={`collections-unassign-submit-${collectionKey(item.name)}`}
                          onClick={() => {
                            void handleUnassignItem(item)
                          }}
                        >
                          <Trash2 className='mr-2 h-4 w-4' />
                          Unassign
                        </Button>
                      </div>
                    </div>
                  ))
                ) : (
                  <div className='text-sm text-muted-foreground' data-testid='collections-members-empty'>
                    No items are currently assigned to {selectedCollectionName}.
                  </div>
                )}
              </CardContent>
            </Card>
          </div>
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
