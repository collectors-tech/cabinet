import { useMemo } from 'react'
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
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { TasksTable } from '@/features/tasks/components/tasks-table'
import { TasksDialogs } from '@/features/tasks/components/tasks-dialogs'
import { TasksProvider } from '@/features/tasks/components/tasks-provider'
import { tasks } from '@/features/tasks/data/tasks'

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
  const summary = useMemo(
    () => ({
      folders: folderNames.length,
      items: tasks.length,
      activeBrand: 'All',
      activeCategory: 'All',
    }),
    []
  )

  return (
    <TasksProvider>
      <Header fixed>
        <Search />
        <div className='ms-auto flex items-center space-x-4'>
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
              <TasksTable data={tasks} routePath={routePath} />
            </CardContent>
          </Card>
        </div>
      </Main>
      <TasksDialogs />
    </TasksProvider>
  )
}
