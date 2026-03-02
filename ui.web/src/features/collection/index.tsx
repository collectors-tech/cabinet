import { useCallback, useEffect, useMemo, useState } from 'react'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { ConfigDrawer } from '@/components/config-drawer'
import { Header } from '@/components/layout/header'
import { LanguageSwitch } from '@/components/language-switch'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
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

const inventoryPhotos = [
  {
    id: 'inventory-photo-1',
    title: 'AFX Camaro Wildfire',
    src: 'data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" width="1280" height="720"><defs><linearGradient id="g" x1="0" x2="1"><stop offset="0" stop-color="%23152a4a"/><stop offset="1" stop-color="%23305d8a"/></linearGradient></defs><rect width="100%" height="100%" fill="url(%23g)"/><text x="50%" y="50%" fill="%23f8fafc" font-size="52" text-anchor="middle" font-family="Arial">AFX Camaro Wildfire</text></svg>',
  },
  {
    id: 'inventory-photo-2',
    title: 'Mega-G Plus Porsche 962',
    src: 'data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" width="1280" height="720"><defs><linearGradient id="g" x1="0" x2="1"><stop offset="0" stop-color="%23212b44"/><stop offset="1" stop-color="%234f46e5"/></linearGradient></defs><rect width="100%" height="100%" fill="url(%23g)"/><text x="50%" y="50%" fill="%23f8fafc" font-size="52" text-anchor="middle" font-family="Arial">Mega-G Plus Porsche 962</text></svg>',
  },
  {
    id: 'inventory-photo-3',
    title: 'Ford GT40 Gulf Livery',
    src: 'data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" width="1280" height="720"><defs><linearGradient id="g" x1="0" x2="1"><stop offset="0" stop-color="%230f172a"/><stop offset="1" stop-color="%230ea5e9"/></linearGradient></defs><rect width="100%" height="100%" fill="url(%23g)"/><text x="50%" y="50%" fill="%23f8fafc" font-size="52" text-anchor="middle" font-family="Arial">Ford GT40 Gulf Livery</text></svg>',
  },
]

export function Collection({
  title = 'Collection',
  description = 'Command your inventory and move from folders to item actions quickly.',
  routePath,
}: CollectionWorkspaceProps) {
  const [tableData, setTableData] = useState<Task[]>(tasks)
  const [activeFolder, setActiveFolder] = useState(folderNames[0])
  const [loading, setLoading] = useState(false)
  const [loadError, setLoadError] = useState<string | null>(null)
  const [selectedPhotoIndex, setSelectedPhotoIndex] = useState<number | null>(
    null
  )

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
      activeContext: activeFolder,
    }),
    [activeFolder, tableData.length]
  )
  const selectedPhoto =
    selectedPhotoIndex === null ? null : inventoryPhotos[selectedPhotoIndex]
  const isInventoryRoute = routePath === '/_authenticated/inventory/'

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
        <div className='flex flex-wrap items-end justify-between gap-3'>
          <div className='space-y-2'>
            <h1 className='text-2xl font-bold tracking-tight'>{title}</h1>
            <p className='text-muted-foreground'>{description}</p>
            <p
              className='text-xs text-muted-foreground'
              data-testid='collection-context-label'
            >
              Collection context: {activeFolder}
            </p>
          </div>
          <div className='flex gap-2'>
            <Button>Add Item</Button>
            <Button variant='outline'>Add Folder</Button>
          </div>
        </div>

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
                  data-testid={`collection-folder-${folder
                    .trim()
                    .toLowerCase()
                    .replace(/\s+/g, '-')}`}
                  className='w-full justify-start'
                  variant={folder === activeFolder ? 'default' : 'outline'}
                  onClick={() => setActiveFolder(folder)}
                >
                  {folder}
                </Button>
              ))}
            </CardContent>
          </Card>

          <Card className='lg:col-span-9'>
            <CardHeader>
              <CardTitle>Collection Browser</CardTitle>
            </CardHeader>
            <CardContent className='space-y-4'>
              <p className='text-sm text-muted-foreground'>
                Folders: <strong>{summary.folders}</strong>{' '}
                <span className='mx-2'>Items: <strong>{summary.items}</strong></span>
                Active Brand: <strong>{summary.activeBrand}</strong>{' '}
                <span className='mx-2'>
                  Active Category: <strong>{summary.activeCategory}</strong>
                </span>
                <span className='mx-2'>
                  Active Context:{' '}
                  <strong data-testid='collection-active-context'>
                    {summary.activeContext}
                  </strong>
                </span>
              </p>
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
              {isInventoryRoute ? (
                <section
                  className='space-y-3 rounded-md border p-4'
                  data-testid='inventory-photos-section'
                >
                  <div>
                    <h3 className='text-base font-semibold'>Photos</h3>
                    <p className='text-sm text-muted-foreground'>
                      Review item media and inspect photos in fullscreen mode.
                    </p>
                  </div>
                  <div className='grid grid-cols-1 gap-3 md:grid-cols-3'>
                    {inventoryPhotos.map((photo, index) => (
                      <button
                        key={photo.id}
                        type='button'
                        className='overflow-hidden rounded-md border text-left transition hover:border-primary/60'
                        data-testid='inventory-photo-thumb'
                        onClick={() => setSelectedPhotoIndex(index)}
                      >
                        <img
                          src={photo.src}
                          alt={photo.title}
                          className='h-32 w-full object-cover'
                        />
                        <div className='p-2 text-sm'>{photo.title}</div>
                      </button>
                    ))}
                  </div>
                </section>
              ) : null}
            </CardContent>
          </Card>
        </div>
        <Dialog
          open={selectedPhotoIndex !== null}
          onOpenChange={(open) => {
            if (!open) {
              setSelectedPhotoIndex(null)
            }
          }}
        >
          <DialogContent
            className='max-w-4xl'
            data-testid='inventory-photo-fullscreen'
          >
            <DialogHeader>
              <DialogTitle>{selectedPhoto?.title ?? 'Photo Viewer'}</DialogTitle>
            </DialogHeader>
            {selectedPhoto ? (
              <img
                src={selectedPhoto.src}
                alt={selectedPhoto.title}
                className='max-h-[70vh] w-full rounded-md object-contain'
              />
            ) : null}
            <div className='flex justify-end gap-2'>
              <Button
                type='button'
                variant='outline'
                data-testid='inventory-photo-prev'
                onClick={() => {
                  if (selectedPhotoIndex === null) {
                    return
                  }
                  const nextIndex =
                    (selectedPhotoIndex - 1 + inventoryPhotos.length) %
                    inventoryPhotos.length
                  setSelectedPhotoIndex(nextIndex)
                }}
              >
                Previous
              </Button>
              <Button
                type='button'
                variant='outline'
                data-testid='inventory-photo-next'
                onClick={() => {
                  if (selectedPhotoIndex === null) {
                    return
                  }
                  const nextIndex =
                    (selectedPhotoIndex + 1) % inventoryPhotos.length
                  setSelectedPhotoIndex(nextIndex)
                }}
              >
                Next
              </Button>
              <Button
                type='button'
                data-testid='inventory-photo-fullscreen-close'
                onClick={() => setSelectedPhotoIndex(null)}
              >
                Close
              </Button>
            </div>
          </DialogContent>
        </Dialog>
      </Main>
      <TasksDialogs />
    </TasksProvider>
  )
}
