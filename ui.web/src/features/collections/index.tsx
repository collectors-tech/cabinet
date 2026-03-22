import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { LanguageSwitch } from '@/components/language-switch'
import { ConfigDrawer } from '@/components/config-drawer'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { useWorkspaceCollections } from '@/features/collections/use-workspace-collections'
import { ArrowDownAZ, ArrowUpAZ, Tag } from 'lucide-react'
import { useMemo, useState } from 'react'

type CollectionFilterMode = 'all' | 'active-lanes' | 'storage' | 'custom'
type CollectionSortMode = 'name-asc' | 'name-desc' | 'items-desc'

function isSystemCollection(collection: { name: string; scopeLabel: string }) {
  return (
    collection.name === 'All Items' ||
    collection.scopeLabel === 'Monitoring set' ||
    collection.scopeLabel === 'Intent shortlist' ||
    collection.scopeLabel === 'Storage location' ||
    collection.scopeLabel === 'Retail lane'
  )
}

export function Collections() {
  const {
    collectionSummaries,
    activeWorkspaceCollection,
    setActiveWorkspaceCollection,
    addCollection,
  } = useWorkspaceCollections()
  const [newCollectionName, setNewCollectionName] = useState('')
  const [createPanelOpen, setCreatePanelOpen] = useState(false)
  const [createError, setCreateError] = useState<string | null>(null)
  const [createOutcome, setCreateOutcome] = useState<string | null>(null)
  const [searchTerm, setSearchTerm] = useState('')
  const [filterMode, setFilterMode] = useState<CollectionFilterMode>('all')
  const [sortMode, setSortMode] = useState<CollectionSortMode>('name-asc')

  const availableCreateActions = useMemo(
    () => [
      {
        id: 'new-collection',
        label: 'New Collection',
        description: 'Open the primary create flow and name a collection explicitly.',
      },
      {
        id: 'starter-collections',
        label: 'Add Starter Collections',
        description: 'Seed a couple of ready-to-use collections for quick workspace setup.',
      },
    ],
    []
  )

  const filteredCollections = useMemo(() => {
    const normalizedSearch = searchTerm.trim().toLowerCase()

    const visible = collectionSummaries
      .filter((collection) => {
        if (!normalizedSearch) {
          return true
        }
        return [
          collection.name,
          collection.description,
          collection.scopeLabel,
          collection.statusLabel,
        ]
          .join(' ')
          .toLowerCase()
          .includes(normalizedSearch)
      })
      .filter((collection) => {
        if (filterMode === 'all') {
          return true
        }
        if (filterMode === 'active-lanes') {
          return collection.scopeLabel === 'Retail lane' || collection.scopeLabel === 'Monitoring set'
        }
        if (filterMode === 'storage') {
          return collection.scopeLabel === 'Storage location'
        }
        return !isSystemCollection(collection)
      })

    return [...visible].sort((left, right) => {
      if (sortMode === 'items-desc') {
        return right.itemCount - left.itemCount || left.name.localeCompare(right.name)
      }
      if (sortMode === 'name-desc') {
        return right.name.localeCompare(left.name)
      }
      return left.name.localeCompare(right.name)
    })
  }, [collectionSummaries, filterMode, searchTerm, sortMode])

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

      <Main className='space-y-4'>
        <div className='space-y-1'>
          <div className='flex items-center gap-2'>
            <Tag data-testid='collections-page-icon' className='h-5 w-5 text-muted-foreground' />
            <h1 className='text-2xl font-bold tracking-tight'>Collections</h1>
          </div>
          <p className='text-muted-foreground'>
            Manage workspace collections outside of sidebar chrome.
          </p>
        </div>
        <Card data-testid='collections-section'>
          <CardHeader>
            <div className='flex flex-wrap items-center justify-between gap-2'>
              <CardTitle>Workspace Collections</CardTitle>
              <div className='flex items-center gap-2'>
                <Button
                  type='button'
                  data-testid='collections-new-action'
                  onClick={() => {
                    setCreatePanelOpen(true)
                    setCreateError(null)
                    setCreateOutcome('New opens the primary create flow for a named collection.')
                  }}
                >
                  New
                </Button>
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button
                      type='button'
                      variant='outline'
                      data-testid='collections-create-menu-trigger'
                    >
                      Create
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align='end'>
                    <DropdownMenuItem
                      data-testid='collections-create-menu-new'
                      onClick={() => {
                        setCreatePanelOpen(true)
                        setCreateError(null)
                        setCreateOutcome('Create → New Collection opens the full naming flow on this page.')
                      }}
                    >
                      New Collection
                    </DropdownMenuItem>
                    <DropdownMenuItem
                      data-testid='collections-create-menu-starter'
                      onClick={() => {
                        addCollection('New Arrivals')
                        addCollection('Recently Graded')
                        setCreatePanelOpen(false)
                        setCreateError(null)
                        setCreateOutcome(
                          'Starter collections added. New Arrivals is now available as an active collection choice.'
                        )
                      }}
                    >
                      Add Starter Collections
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </div>
            </div>
            <CardDescription>
              Create, list, and switch active collection context.
            </CardDescription>
            <div
              className='rounded-md border border-border/60 bg-muted/20 p-3 text-sm'
              data-testid='collections-create-guidance'
            >
              <p className='font-medium'>How creation works here</p>
              <ul className='mt-2 space-y-1 text-muted-foreground'>
                {availableCreateActions.map((action) => (
                  <li key={action.id} data-testid={`collections-create-guidance-${action.id}`}>
                    <span className='font-medium text-foreground'>{action.label}:</span>{' '}
                    {action.description}
                  </li>
                ))}
              </ul>
            </div>
          </CardHeader>
          <CardContent className='space-y-3'>
            <div
              className='rounded-md border border-border/60 bg-muted/10 p-3'
              data-testid='collections-management-tools'
            >
              <div className='flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between'>
                <div className='flex flex-1 flex-col gap-3 md:flex-row'>
                  <div className='flex-1'>
                    <p className='mb-1 text-sm font-medium'>Find a collection</p>
                    <Input
                      data-testid='collections-search-input'
                      placeholder='Search collections...'
                      value={searchTerm}
                      onChange={(event) => setSearchTerm(event.target.value)}
                    />
                  </div>
                  <div>
                    <p className='mb-1 text-sm font-medium'>Filter</p>
                    <div className='flex flex-wrap gap-2'>
                      <Button
                        type='button'
                        size='sm'
                        variant={filterMode === 'all' ? 'default' : 'outline'}
                        data-testid='collections-filter-all'
                        onClick={() => setFilterMode('all')}
                      >
                        All
                      </Button>
                      <Button
                        type='button'
                        size='sm'
                        variant={filterMode === 'active-lanes' ? 'default' : 'outline'}
                        data-testid='collections-filter-active-lanes'
                        onClick={() => setFilterMode('active-lanes')}
                      >
                        Active lanes
                      </Button>
                      <Button
                        type='button'
                        size='sm'
                        variant={filterMode === 'storage' ? 'default' : 'outline'}
                        data-testid='collections-filter-storage'
                        onClick={() => setFilterMode('storage')}
                      >
                        Storage
                      </Button>
                      <Button
                        type='button'
                        size='sm'
                        variant={filterMode === 'custom' ? 'default' : 'outline'}
                        data-testid='collections-filter-custom'
                        onClick={() => setFilterMode('custom')}
                      >
                        Custom
                      </Button>
                    </div>
                  </div>
                </div>
                <div>
                  <p className='mb-1 text-sm font-medium'>Order</p>
                  <div className='flex flex-wrap gap-2'>
                    <Button
                      type='button'
                      size='sm'
                      variant={sortMode === 'name-asc' ? 'default' : 'outline'}
                      data-testid='collections-sort-name-asc'
                      onClick={() => setSortMode('name-asc')}
                    >
                      <ArrowUpAZ className='mr-1 size-4' /> A–Z
                    </Button>
                    <Button
                      type='button'
                      size='sm'
                      variant={sortMode === 'name-desc' ? 'default' : 'outline'}
                      data-testid='collections-sort-name-desc'
                      onClick={() => setSortMode('name-desc')}
                    >
                      <ArrowDownAZ className='mr-1 size-4' /> Z–A
                    </Button>
                    <Button
                      type='button'
                      size='sm'
                      variant={sortMode === 'items-desc' ? 'default' : 'outline'}
                      data-testid='collections-sort-items-desc'
                      onClick={() => setSortMode('items-desc')}
                    >
                      Largest first
                    </Button>
                  </div>
                </div>
              </div>
              <p
                className='mt-3 text-xs text-muted-foreground'
                data-testid='collections-management-summary'
              >
                Showing {filteredCollections.length} of {collectionSummaries.length} collections.
              </p>
            </div>
            {createOutcome ? (
              <div
                className='rounded-md border border-emerald-500/30 bg-emerald-500/10 px-3 py-2 text-sm'
                data-testid='collections-create-outcome'
              >
                {createOutcome}
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
                          setCreateOutcome(null)
                          return
                        }
                        addCollection(trimmedName)
                        setNewCollectionName('')
                        setCreatePanelOpen(false)
                        setCreateError(null)
                        setCreateOutcome(
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
                        setCreateOutcome('Collection creation cancelled. No collection was added.')
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
                return (
                  <button
                    key={collection.name}
                    type='button'
                    className='rounded-md border px-3 py-3 text-left hover:bg-muted'
                    data-testid={`collections-item-${collection.key}`}
                    data-state={isActive ? 'active' : 'inactive'}
                    onClick={() => setActiveWorkspaceCollection(collection.name)}
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
                )
              }) : (
                <div
                  className='rounded-md border border-dashed px-3 py-6 text-sm text-muted-foreground'
                  data-testid='collections-empty-state'
                >
                  No collections match the current search/filter combination.
                </div>
              )}
            </div>
          </CardContent>
        </Card>
      </Main>
    </>
  )
}
