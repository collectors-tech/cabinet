import { ConfigDrawer } from '@/components/config-drawer'
import { Header } from '@/components/layout/header'
import { LanguageSwitch } from '@/components/language-switch'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Input } from '@/components/ui/input'
import {
  collectionKey,
  useWorkspaceCollections,
} from '@/features/collections/use-workspace-collections'
import { useEffect, useState } from 'react'
import { TasksDialogs } from './components/tasks-dialogs'
import { TasksProvider } from './components/tasks-provider'
import { TasksTable } from './components/tasks-table'
import { tasks } from './data/tasks'
import { type Task } from './data/schema'

type TasksProps = {
  title?: string
  description?: string
  routePath?: '/_authenticated/inventory/' | '/_authenticated/wishlist/'
}

export function Tasks({
  title = 'Tasks',
  description = "Here's a list of your tasks for this month!",
  routePath = '/_authenticated/inventory/',
}: TasksProps) {
  const {
    workspaceCollections,
    activeWorkspaceCollection,
    setActiveWorkspaceCollection,
    addCollection,
  } = useWorkspaceCollections()
  const [inlineCollectionInputOpen, setInlineCollectionInputOpen] = useState(false)
  const [inlineCollectionName, setInlineCollectionName] = useState('')
  const [inlineCollectionValidationMessage, setInlineCollectionValidationMessage] =
    useState('')
  const [tableData, setTableData] = useState<Task[]>(tasks)
  const isWishlistRoute = routePath === '/_authenticated/wishlist/'

  useEffect(() => {
    if (!isWishlistRoute) {
      setTableData(tasks)
      return
    }

    let cancelled = false
    const loadWishlistData = async () => {
      try {
        const [wishlistResponse, itemsResponse] = await Promise.all([
          fetch('/api/wishlist'),
          fetch('/api/items'),
        ])
        if (!wishlistResponse.ok || !itemsResponse.ok) {
          throw new Error('wishlist_bootstrap_failed')
        }
        const wishlistPayload = (await wishlistResponse.json()) as {
          items?: Array<{
            id?: string
            item_id?: string
            priority?: string
            below_target_now?: boolean
          }>
        }
        const itemsPayload = (await itemsResponse.json()) as {
          items?: Array<{
            id?: string
            title?: string
            part_number?: string
          }>
        }
        const itemTitleByID = new Map<string, string>()
        ;(itemsPayload.items ?? []).forEach((item) => {
          const itemID = item.id?.trim()
          if (!itemID) {
            return
          }
          const label = item.title?.trim() || item.part_number?.trim() || itemID
          itemTitleByID.set(itemID, label)
        })
        const mapped: Task[] = (wishlistPayload.items ?? []).map((entry, index) => {
          const itemID = entry.item_id?.trim() || `wishlist-item-${index + 1}`
          return {
            id: itemID,
            title: itemTitleByID.get(itemID) || `Wishlist item ${index + 1}`,
            status: entry.below_target_now ? 'discovered' : 'wishlist',
            label: 'collection',
            priority: entry.priority?.trim() || 'medium',
          }
        })
        if (!cancelled) {
          setTableData(mapped)
        }
      } catch {
        if (!cancelled) {
          setTableData([])
        }
      }
    }

    void loadWishlistData()
    return () => {
      cancelled = true
    }
  }, [isWishlistRoute])

  return (
    <TasksProvider>
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
        <div className='flex flex-wrap items-end justify-between gap-2'>
          <div>
            <h2 className='text-2xl font-bold tracking-tight'>{title}</h2>
            <p className='text-muted-foreground'>{description}</p>
          </div>
          <div className='flex items-center gap-2'>
            <Button
              type='button'
              data-testid='wishlist-new-action'
            >
              New
            </Button>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button
                  type='button'
                  variant='outline'
                  data-testid='wishlist-create-menu-trigger'
                >
                  Create
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align='end'>
                <DropdownMenuItem
                  data-testid='wishlist-create-menu-entry'
                >
                  New Wishlist Entry
                </DropdownMenuItem>
                <DropdownMenuItem
                  data-testid='wishlist-create-menu-import'
                >
                  Import Wishlist
                </DropdownMenuItem>
                {isWishlistRoute ? (
                  <DropdownMenuItem
                    data-testid='wishlist-create-menu-collection'
                    onClick={() => {
                      setInlineCollectionValidationMessage('')
                      setInlineCollectionInputOpen(true)
                    }}
                  >
                    New Collection
                  </DropdownMenuItem>
                ) : null}
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>
        {isWishlistRoute ? (
          <div
            className='rounded-md border p-3'
            data-testid='wishlist-inline-picker'
          >
            <div className='grid gap-2 md:grid-cols-[minmax(0,1fr)_auto_auto] md:items-center'>
              <select
                className='h-9 rounded-md border bg-background px-2 text-sm'
                value={activeWorkspaceCollection}
                onChange={(event) =>
                  setActiveWorkspaceCollection(event.target.value)
                }
              >
                {workspaceCollections.map((collection) => (
                  <option
                    key={collection}
                    value={collection}
                    data-testid={`wishlist-inline-picker-option-${collectionKey(collection)}`}
                  >
                    {collection}
                  </option>
                ))}
              </select>
              <span
                className='text-sm text-muted-foreground'
                data-testid='wishlist-inline-picker-selected'
              >
                {activeWorkspaceCollection}
              </span>
              <Button
                type='button'
                variant='outline'
                data-testid='wishlist-inline-add-new'
                onClick={() => {
                  setInlineCollectionValidationMessage('')
                  setInlineCollectionInputOpen((open) => !open)
                }}
              >
                + New Collection
              </Button>
            </div>
            {inlineCollectionInputOpen ? (
              <div className='mt-2'>
                <div className='flex gap-2'>
                  <Input
                    data-testid='wishlist-inline-new-name'
                    placeholder='Collection name'
                    aria-invalid={
                      inlineCollectionValidationMessage ? 'true' : 'false'
                    }
                    value={inlineCollectionName}
                    onChange={(event) => {
                      setInlineCollectionName(event.target.value)
                      if (inlineCollectionValidationMessage) {
                        setInlineCollectionValidationMessage('')
                      }
                    }}
                  />
                  <Button
                    type='button'
                    data-testid='wishlist-inline-save'
                    onClick={() => {
                      const created = addCollection(inlineCollectionName)
                      if (!created) {
                        setInlineCollectionValidationMessage(
                          'Collection name is required.'
                        )
                        return
                      }
                      setInlineCollectionValidationMessage('')
                      setActiveWorkspaceCollection(created)
                      setInlineCollectionName('')
                      setInlineCollectionInputOpen(false)
                    }}
                  >
                    Save
                  </Button>
                </div>
                {inlineCollectionValidationMessage ? (
                  <p
                    className='mt-2 text-sm text-destructive'
                    data-testid='wishlist-inline-validation'
                    role='alert'
                  >
                    {inlineCollectionValidationMessage}
                  </p>
                ) : null}
              </div>
            ) : null}
          </div>
        ) : null}
        <TasksTable data={tableData} routePath={routePath} />
      </Main>

      <TasksDialogs />
    </TasksProvider>
  )
}
