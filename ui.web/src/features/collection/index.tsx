import { useCallback, useEffect, useMemo, useState } from 'react'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { ConfigDrawer } from '@/components/config-drawer'
import { Header } from '@/components/layout/header'
import { LanguageSwitch } from '@/components/language-switch'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { TasksTable } from '@/features/tasks/components/tasks-table'
import { TasksDialogs } from '@/features/tasks/components/tasks-dialogs'
import { TasksProvider } from '@/features/tasks/components/tasks-provider'
import { tasks } from '@/features/tasks/data/tasks'
import { type Task } from '@/features/tasks/data/schema'

type CollectionWorkspaceProps = {
  title?: string
  description?: string
  routePath: '/_authenticated/inventory/' | '/_authenticated/wishlist/'
}

const folderNames = [
  'All Items',
  'Watch List',
  'Wishlist Focus',
  'Store 1',
  'Store 2',
  'Warehouse 1',
]

export function Collection({
  title = 'Collection',
  description = 'Command your inventory and move from folders to item actions quickly.',
  routePath,
}: CollectionWorkspaceProps) {
  const [tableData, setTableData] = useState<Task[]>(tasks)
  const [loading, setLoading] = useState(false)
  const [loadError, setLoadError] = useState<string | null>(null)

  const loadInventoryItems = useCallback(async () => {
    if (routePath !== '/_authenticated/inventory/') {
      setTableData(tasks)
      setLoadError(null)
      return
    }
    setLoading(true)
    setLoadError(null)
    try {
      const response = await fetch('/api/items')
      if (!response.ok) {
        throw new Error(`items_api_${response.status}`)
      }
      const payload = (await response.json()) as {
        items?: Array<{
          id?: string
          part_number?: string
          title?: string
          status?: string
          category?: string
        }>
      }
      const mapped: Task[] = (payload.items ?? []).map((item, index) => ({
        id: item.part_number?.trim() || item.id?.trim() || `ITEM-${index + 1}`,
        title: item.title?.trim() || 'Untitled item',
        status: item.status?.trim() || 'todo',
        label: item.category?.trim() || 'feature',
        priority: 'medium',
      }))
      setTableData(mapped)
    } catch {
      setLoadError(
        'Inventory failed to load. Retry and confirm runtime API availability.'
      )
      setTableData([])
    } finally {
      setLoading(false)
    }
  }, [routePath])

  useEffect(() => {
    void loadInventoryItems()
  }, [loadInventoryItems])

  const summary = useMemo(
    () => ({
      folders: folderNames.length,
      items: tableData.length,
      activeBrand: 'All',
      activeCategory: 'All',
    }),
    [tableData.length]
  )

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

      <Main className='space-y-4'>
        <div className='space-y-2'>
          <h1 className='text-2xl font-bold tracking-tight'>{title}</h1>
          <p className='text-muted-foreground'>{description}</p>
        </div>

        <Card>
          <CardHeader className='pb-3'>
            <CardTitle className='text-base'>Command Row</CardTitle>
            <CardDescription>
              Search, quick actions, and create flows for collection operations.
            </CardDescription>
          </CardHeader>
          <CardContent className='flex flex-wrap gap-2'>
            <Input
              className='w-full max-w-sm'
              placeholder='Search product code, title, grading...'
            />
            <Button>Add Item</Button>
            <Button variant='outline'>Add Folder</Button>
            <Button variant='outline'>Bulk Edit</Button>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className='pb-2'>
            <CardTitle className='text-base'>Summary Strip</CardTitle>
          </CardHeader>
          <CardContent className='flex flex-wrap items-center gap-4 text-sm'>
            <span>
              Folders: <strong>{summary.folders}</strong>
            </span>
            <span>
              Items: <strong>{summary.items}</strong>
            </span>
            <span>
              Active Brand: <strong>{summary.activeBrand}</strong>
            </span>
            <span>
              Active Category: <strong>{summary.activeCategory}</strong>
            </span>
          </CardContent>
        </Card>

        <div className='grid grid-cols-1 gap-4 lg:grid-cols-12'>
          <Card className='lg:col-span-3'>
            <CardHeader>
              <CardTitle>Folders</CardTitle>
              <CardDescription>
                Browse folders before drilling into results.
              </CardDescription>
            </CardHeader>
            <CardContent className='space-y-2'>
              {folderNames.map((folder) => (
                <Button
                  key={folder}
                  className='w-full justify-start'
                  variant={folder === 'All Items' ? 'default' : 'outline'}
                >
                  {folder}
                </Button>
              ))}
            </CardContent>
          </Card>

          <Card className='lg:col-span-9'>
            <CardHeader>
              <CardTitle>Collection Browser</CardTitle>
              <CardDescription>
                Rows and cards share the same filters, sort, and pagination.
              </CardDescription>
            </CardHeader>
            <CardContent className='space-y-4'>
              {loading ? (
                <div className='rounded-md border p-6 text-sm text-muted-foreground'>
                  Loading inventory...
                </div>
              ) : null}
              {loadError ? (
                <div
                  className='rounded-md border border-destructive/40 bg-destructive/10 p-4 text-sm'
                  data-testid='inventory-load-error'
                >
                  <p className='font-medium'>Inventory load failed</p>
                  <p className='mt-1 text-muted-foreground'>{loadError}</p>
                  <Button
                    className='mt-3'
                    variant='outline'
                    size='sm'
                    onClick={() => void loadInventoryItems()}
                  >
                    Retry
                  </Button>
                </div>
              ) : null}
              <TasksTable data={tableData} routePath={routePath} />
            </CardContent>
          </Card>
        </div>
      </Main>
      <TasksDialogs />
    </TasksProvider>
  )
}
