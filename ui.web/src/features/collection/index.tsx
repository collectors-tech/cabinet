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
import { Input } from '@/components/ui/input'
import {
  Dialog,
  DialogContent,
  DialogFooter,
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

type InventoryPhoto = {
  id: string
  filename: string
  is_primary?: boolean
}

type AISuggestion = {
  part_number?: string
  brand?: string
  title?: string
  confidence?: number
  requires_confirmation?: boolean
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
  const [activeFolder, setActiveFolder] = useState(folderNames[0])
  const [loading, setLoading] = useState(false)
  const [loadError, setLoadError] = useState<string | null>(null)
  const [inventoryPhotos, setInventoryPhotos] = useState<InventoryPhoto[]>([])
  const [photosLoading, setPhotosLoading] = useState(false)
  const [photosError, setPhotosError] = useState<string | null>(null)
  const [photosBusy, setPhotosBusy] = useState(false)
  const [activeProfileID, setActiveProfileID] = useState('')
  const [aiTitleInput, setAITitleInput] = useState('')
  const [aiPhotoURLInput, setAIPhotoURLInput] = useState('')
  const [aiSuggestion, setAISuggestion] = useState<AISuggestion | null>(null)
  const [aiLoading, setAILoading] = useState(false)
  const [aiError, setAIError] = useState<string | null>(null)
  const [aiApplied, setAIApplied] = useState(false)
  const [aiConfirmOpen, setAIConfirmOpen] = useState(false)
  const [lastAISuggestAction, setLastAISuggestAction] = useState<
    'title' | 'photo' | null
  >(null)
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
  const isInventoryRoute = routePath === '/_authenticated/inventory/'
  const selectedItemID = isInventoryRoute ? tableData[0]?.id?.trim() ?? '' : ''
  const selectedPhoto =
    selectedPhotoIndex === null ? null : inventoryPhotos[selectedPhotoIndex]

  const loadInventoryPhotos = useCallback(async () => {
    if (!isInventoryRoute) {
      setInventoryPhotos([])
      setPhotosError(null)
      setPhotosLoading(false)
      return
    }
    if (!selectedItemID) {
      setInventoryPhotos([])
      setPhotosError(null)
      setPhotosLoading(false)
      return
    }
    setPhotosLoading(true)
    setPhotosError(null)
    try {
      const response = await fetch(
        `/api/items/${encodeURIComponent(selectedItemID)}/photos`
      )
      if (!response.ok) {
        throw new Error(`list_photos_${response.status}`)
      }
      const payload = (await response.json()) as {
        photos?: Array<{
          id?: string
          filename?: string
          is_primary?: boolean
        }>
      }
      const mapped = (payload.photos ?? []).map((photo) => ({
        id: photo.id?.trim() ?? '',
        filename: photo.filename?.trim() || 'photo.jpg',
        is_primary: photo.is_primary ?? false,
      }))
      setInventoryPhotos(mapped.filter((photo) => photo.id !== ''))
    } catch {
      setInventoryPhotos([])
      setPhotosError('Photos could not be loaded for this item. Retry to continue.')
    } finally {
      setPhotosLoading(false)
    }
  }, [isInventoryRoute, selectedItemID])

  useEffect(() => {
    void loadInventoryPhotos()
  }, [loadInventoryPhotos])

  useEffect(() => {
    if (
      selectedPhotoIndex !== null &&
      (selectedPhotoIndex < 0 || selectedPhotoIndex >= inventoryPhotos.length)
    ) {
      setSelectedPhotoIndex(null)
    }
  }, [inventoryPhotos.length, selectedPhotoIndex])

  useEffect(() => {
    if (!isInventoryRoute) {
      setActiveProfileID('')
      return
    }
    let cancelled = false
    const loadActiveProfile = async () => {
      try {
        const response = await fetch('/api/profiles/active')
        if (!response.ok) {
          return
        }
        const payload = (await response.json()) as { id?: string }
        if (!cancelled) {
          setActiveProfileID(payload.id?.trim() ?? '')
        }
      } catch {
        if (!cancelled) {
          setActiveProfileID('')
        }
      }
    }
    void loadActiveProfile()
    return () => {
      cancelled = true
    }
  }, [isInventoryRoute])

  const handlePhotoUpload = useCallback(
    async (file: File | null) => {
      if (!file || !selectedItemID) {
        return
      }
      setPhotosBusy(true)
      setPhotosError(null)
      try {
        const body = new FormData()
        body.append('file', file)
        const response = await fetch(
          `/api/items/${encodeURIComponent(selectedItemID)}/photos`,
          {
            method: 'POST',
            body,
          }
        )
        if (!response.ok) {
          throw new Error(`upload_photo_${response.status}`)
        }
        await loadInventoryPhotos()
      } catch {
        setPhotosError('Photo upload failed. Retry with a supported image file.')
      } finally {
        setPhotosBusy(false)
      }
    },
    [loadInventoryPhotos, selectedItemID]
  )

  const handleSetPrimaryPhoto = useCallback(
    async (photoID: string) => {
      if (!selectedItemID || !photoID) {
        return
      }
      setPhotosBusy(true)
      setPhotosError(null)
      try {
        const response = await fetch(
          `/api/items/${encodeURIComponent(
            selectedItemID
          )}/photos/${encodeURIComponent(photoID)}/primary`,
          {
            method: 'PUT',
          }
        )
        if (!response.ok) {
          throw new Error(`set_primary_${response.status}`)
        }
        await loadInventoryPhotos()
      } catch {
        setPhotosError('Unable to set primary photo right now. Retry this action.')
      } finally {
        setPhotosBusy(false)
      }
    },
    [loadInventoryPhotos, selectedItemID]
  )

  const handleDeletePhoto = useCallback(
    async (photoID: string) => {
      if (!selectedItemID || !photoID) {
        return
      }
      setPhotosBusy(true)
      setPhotosError(null)
      try {
        const response = await fetch(
          `/api/items/${encodeURIComponent(
            selectedItemID
          )}/photos/${encodeURIComponent(photoID)}`,
          {
            method: 'DELETE',
          }
        )
        if (!response.ok) {
          throw new Error(`delete_photo_${response.status}`)
        }
        await loadInventoryPhotos()
      } catch {
        setPhotosError('Unable to delete this photo. Retry this action.')
      } finally {
        setPhotosBusy(false)
      }
    },
    [loadInventoryPhotos, selectedItemID]
  )

  const runAISuggest = useCallback(
    async (mode: 'title' | 'photo') => {
      if (!activeProfileID) {
        setAIError('Active profile is required before AI requests can run.')
        return
      }
      setAILoading(true)
      setAIError(null)
      setLastAISuggestAction(mode)
      setAIApplied(false)
      try {
        const endpoint =
          mode === 'title' ? '/api/ai/suggest/title' : '/api/ai/suggest/photo'
        const payload =
          mode === 'title'
            ? { profile_id: activeProfileID, title: aiTitleInput }
            : { profile_id: activeProfileID, image_url: aiPhotoURLInput }
        const response = await fetch(endpoint, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload),
        })
        if (!response.ok) {
          throw new Error(`ai_suggest_${mode}_${response.status}`)
        }
        const suggestion = (await response.json()) as AISuggestion
        setAISuggestion(suggestion)
      } catch {
        setAISuggestion(null)
        setAIError('AI suggestion failed. Retry the request.')
      } finally {
        setAILoading(false)
      }
    },
    [activeProfileID, aiPhotoURLInput, aiTitleInput]
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
                <>
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
                    <div className='flex flex-wrap items-center gap-2'>
                      <input
                        type='file'
                        accept='image/*'
                        data-testid='inventory-photo-upload-input'
                        onChange={(event) => {
                          const file = event.target.files?.[0] ?? null
                          void handlePhotoUpload(file)
                          event.currentTarget.value = ''
                        }}
                      />
                      {photosBusy ? (
                        <span className='text-xs text-muted-foreground'>Working...</span>
                      ) : null}
                    </div>
                    {photosLoading ? (
                      <div
                        className='rounded-md border p-3 text-sm text-muted-foreground'
                        data-testid='inventory-photos-loading'
                      >
                        Loading photos...
                      </div>
                    ) : null}
                    {photosError ? (
                      <div
                        className='rounded-md border border-destructive/40 bg-destructive/10 p-3 text-sm'
                        data-testid='inventory-photos-error'
                      >
                        <p className='font-medium'>Photos are unavailable.</p>
                        <p className='mt-1 text-muted-foreground'>{photosError}</p>
                        <Button
                          className='mt-3'
                          size='sm'
                          variant='outline'
                          onClick={() => void loadInventoryPhotos()}
                        >
                          Retry
                        </Button>
                      </div>
                    ) : null}
                    {!photosLoading && !photosError && inventoryPhotos.length === 0 ? (
                      <div
                        className='rounded-md border border-dashed p-3 text-sm text-muted-foreground'
                        data-testid='inventory-photos-empty'
                      >
                        No photos yet for the selected item. Upload an image to begin.
                      </div>
                    ) : null}
                    {!photosLoading && !photosError && inventoryPhotos.length > 0 ? (
                      <>
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
                                src={`/api/items/${encodeURIComponent(
                                  selectedItemID
                                )}/photos/${encodeURIComponent(photo.id)}/file?variant=preview`}
                                alt={photo.filename}
                                className='h-32 w-full object-cover'
                              />
                              <div className='p-2 text-sm'>{photo.filename}</div>
                            </button>
                          ))}
                        </div>
                        <div className='space-y-2'>
                          {inventoryPhotos.map((photo) => (
                            <div
                              key={`row-${photo.id}`}
                              className='flex flex-wrap items-center justify-between gap-2 rounded-md border p-2'
                              data-testid='inventory-photo-row'
                            >
                              <div className='flex items-center gap-2 text-sm'>
                                <span>{photo.filename}</span>
                                {photo.is_primary ? (
                                  <span
                                    className='rounded bg-primary/10 px-2 py-0.5 text-xs text-primary'
                                    data-testid='inventory-photo-primary-badge'
                                  >
                                    Primary
                                  </span>
                                ) : null}
                              </div>
                              <div className='flex gap-2'>
                                <Button
                                  size='sm'
                                  variant='outline'
                                  data-testid='inventory-photo-set-primary'
                                  onClick={() => void handleSetPrimaryPhoto(photo.id)}
                                >
                                  Set Primary
                                </Button>
                                <Button
                                  size='sm'
                                  variant='outline'
                                  data-testid='inventory-photo-delete'
                                  onClick={() => void handleDeletePhoto(photo.id)}
                                >
                                  Delete
                                </Button>
                              </div>
                            </div>
                          ))}
                        </div>
                      </>
                    ) : null}
                  </section>
                  <section
                    className='space-y-3 rounded-md border p-4'
                    data-testid='inventory-ai-assist-section'
                  >
                    <div>
                      <h3 className='text-base font-semibold'>AI Assist</h3>
                      <p className='text-sm text-muted-foreground'>
                        Generate title and photo suggestions with explicit confirm-before-apply.
                      </p>
                    </div>
                    <div className='grid grid-cols-1 gap-3 md:grid-cols-2'>
                      <div className='space-y-2'>
                        <Input
                          placeholder='Paste listing title'
                          data-testid='inventory-ai-title-input'
                          value={aiTitleInput}
                          onChange={(event) => setAITitleInput(event.target.value)}
                        />
                        <Button
                          data-testid='inventory-ai-suggest-title'
                          onClick={() => void runAISuggest('title')}
                          disabled={aiLoading || aiTitleInput.trim() === ''}
                        >
                          Suggest from Title
                        </Button>
                      </div>
                      <div className='space-y-2'>
                        <Input
                          placeholder='Paste photo URL'
                          data-testid='inventory-ai-photo-url-input'
                          value={aiPhotoURLInput}
                          onChange={(event) => setAIPhotoURLInput(event.target.value)}
                        />
                        <Button
                          data-testid='inventory-ai-suggest-photo'
                          onClick={() => void runAISuggest('photo')}
                          disabled={aiLoading || aiPhotoURLInput.trim() === ''}
                        >
                          Suggest from Photo
                        </Button>
                      </div>
                    </div>
                    {aiLoading ? (
                      <div
                        className='rounded-md border p-3 text-sm text-muted-foreground'
                        data-testid='inventory-ai-loading'
                      >
                        Running AI suggestion...
                      </div>
                    ) : null}
                    {aiError ? (
                      <div
                        className='rounded-md border border-destructive/40 bg-destructive/10 p-3 text-sm'
                        data-testid='inventory-ai-error'
                      >
                        <p className='font-medium'>AI suggestion failed.</p>
                        <p className='mt-1 text-muted-foreground'>{aiError}</p>
                        <Button
                          className='mt-3'
                          variant='outline'
                          size='sm'
                          data-testid='inventory-ai-retry'
                          onClick={() => {
                            if (lastAISuggestAction) {
                              void runAISuggest(lastAISuggestAction)
                            }
                          }}
                        >
                          Retry
                        </Button>
                      </div>
                    ) : null}
                    {aiSuggestion ? (
                      <div
                        className='rounded-md border p-3 text-sm'
                        data-testid='inventory-ai-suggestion'
                      >
                        <p>
                          <strong>Part:</strong> {aiSuggestion.part_number || 'n/a'}
                        </p>
                        <p>
                          <strong>Brand:</strong> {aiSuggestion.brand || 'n/a'}
                        </p>
                        <p>
                          <strong>Title:</strong> {aiSuggestion.title || 'n/a'}
                        </p>
                        <p>
                          <strong>Confidence:</strong>{' '}
                          {(aiSuggestion.confidence ?? 0).toFixed(2)}
                        </p>
                        <Button
                          className='mt-3'
                          data-testid='inventory-ai-apply'
                          onClick={() => setAIConfirmOpen(true)}
                        >
                          Apply Suggestion
                        </Button>
                      </div>
                    ) : null}
                    {aiApplied ? (
                      <div
                        className='rounded-md border border-emerald-500/50 bg-emerald-500/10 p-3 text-sm'
                        data-testid='inventory-ai-applied-banner'
                      >
                        Suggestion applied to draft item fields.
                      </div>
                    ) : null}
                  </section>
                </>
              ) : null}
            </CardContent>
          </Card>
        </div>
        <Dialog
          open={selectedPhotoIndex !== null && inventoryPhotos.length > 0}
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
              <DialogTitle>{selectedPhoto?.filename ?? 'Photo Viewer'}</DialogTitle>
            </DialogHeader>
            {selectedPhoto ? (
              <img
                src={`/api/items/${encodeURIComponent(
                  selectedItemID
                )}/photos/${encodeURIComponent(selectedPhoto.id)}/file?variant=original`}
                alt={selectedPhoto.filename}
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
        <Dialog open={aiConfirmOpen} onOpenChange={setAIConfirmOpen}>
          <DialogContent data-testid='inventory-ai-confirm-dialog'>
            <DialogHeader>
              <DialogTitle>Confirm AI Apply</DialogTitle>
            </DialogHeader>
            <p className='text-sm text-muted-foreground'>
              AI suggestions require explicit confirmation before apply.
            </p>
            <DialogFooter>
              <Button variant='outline' onClick={() => setAIConfirmOpen(false)}>
                Cancel
              </Button>
              <Button
                onClick={() => {
                  setAIApplied(true)
                  setAIConfirmOpen(false)
                }}
              >
                Confirm Apply
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </Main>
      <TasksDialogs />
    </TasksProvider>
  )
}
