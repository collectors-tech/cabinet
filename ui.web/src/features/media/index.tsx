import {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
  type DragEvent,
} from 'react'
import {
  type ColumnDef,
  flexRender,
  getCoreRowModel,
  getFacetedRowModel,
  getFacetedUniqueValues,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  type SortingState,
  useReactTable,
} from '@tanstack/react-table'
import {
  Archive,
  Download,
  Eye,
  FileImage,
  ImagePlus,
  Link2,
  WandSparkles,
} from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
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
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Textarea } from '@/components/ui/textarea'
import { ConfigDrawer } from '@/components/config-drawer'
import {
  DataTableColumnHeader,
  DataTablePagination,
  DataTableToolbar,
} from '@/components/data-table'
import { LanguageSwitch } from '@/components/language-switch'
import { Header, HeaderTitle } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'

type MediaAsset = {
  id: string
  title: string
  filename: string
  uploaded_at: string
  linkage_state:
    | 'unlinked'
    | 'linked_inventory'
    | 'linked_wishlist'
    | 'linked_both'
  analysis_status: 'ready' | 'pending' | 'not_analyzed'
  source: string
  item_id?: string
  wishlist_id?: string
  thumbnail_url?: string
  download_filename: string
}

type MediaSummary = {
  total: number
  unlinked: number
  linked_inventory: number
  linked_wishlist: number
  linked_both: number
  ready_for_review: number
}

type MediaListResponse = {
  assets?: MediaAsset[]
  summary?: MediaSummary
  filter?: string
}

type DownloadPreview = {
  count: number
  filenames?: string[]
  allowed: boolean
}

type AssignmentPreview = {
  asset_id: string
  target_type: 'inventory' | 'wishlist'
  target_id: string
  current_linkage_state: MediaAsset['linkage_state']
  projected_linkage_state: MediaAsset['linkage_state']
  allowed: boolean
  requires_confirmation: boolean
  blocked_reason?: string
  audit_summary?: string
  applied?: boolean
}

type MediaViewMode = 'cards' | 'rows'

const MEDIA_VIEW_MODE_STORAGE_KEY = 'cabinet.viewMode.media'
const ACCEPTED_IMAGE_TYPES = [
  'image/jpeg',
  'image/png',
  'image/gif',
  'image/webp',
]

const EMPTY_SUMMARY: MediaSummary = {
  total: 0,
  unlinked: 0,
  linked_inventory: 0,
  linked_wishlist: 0,
  linked_both: 0,
  ready_for_review: 0,
}

function linkageLabel(state: MediaAsset['linkage_state']) {
  switch (state) {
    case 'linked_inventory':
      return 'Inventory linked'
    case 'linked_wishlist':
      return 'Wishlist linked'
    case 'linked_both':
      return 'Inventory + Wishlist'
    default:
      return 'Unlinked'
  }
}

function analysisLabel(status: MediaAsset['analysis_status']) {
  switch (status) {
    case 'ready':
      return 'Analysis ready'
    case 'pending':
      return 'Analysis pending'
    default:
      return 'Not analyzed'
  }
}

function buildMediaColumns({
  selectedAssetSet,
  onToggleAssetSelection,
  onAssign,
}: {
  selectedAssetSet: Set<string>
  onToggleAssetSelection: (assetID: string, checked: boolean) => void
  onAssign: (asset: MediaAsset) => void
}): ColumnDef<MediaAsset>[] {
  return [
    {
      id: 'select',
      header: () => <span className='sr-only'>Select</span>,
      cell: ({ row }) => (
        <Checkbox
          aria-label={`Select ${row.original.title}`}
          checked={selectedAssetSet.has(row.original.id)}
          data-testid={`media-row-select-${row.original.id}`}
          onCheckedChange={(checked) =>
            onToggleAssetSelection(row.original.id, checked === true)
          }
        />
      ),
      enableSorting: false,
      enableHiding: false,
    },
    {
      accessorKey: 'title',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='Media' />
      ),
      cell: ({ row }) => (
        <div className='flex min-w-0 items-center gap-3'>
          <div className='flex h-11 w-14 shrink-0 items-center justify-center overflow-hidden rounded-md border bg-muted'>
            {row.original.thumbnail_url ? (
              <img
                src={row.original.thumbnail_url}
                alt=''
                className='h-full w-full object-cover'
              />
            ) : (
              <FileImage className='h-5 w-5 text-muted-foreground' />
            )}
          </div>
          <div className='min-w-0 space-y-1'>
            <div className='truncate font-medium'>{row.original.title}</div>
            <div className='truncate text-xs text-muted-foreground'>
              {row.original.filename}
            </div>
          </div>
        </div>
      ),
    },
    {
      accessorKey: 'analysis_status',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='Status' />
      ),
      cell: ({ row }) => (
        <Badge className='max-w-full truncate' variant='outline'>
          {analysisLabel(row.original.analysis_status)}
        </Badge>
      ),
      filterFn: (row, _columnId, value) =>
        analysisLabel(row.original.analysis_status)
          .toLowerCase()
          .includes(String(value).toLowerCase()),
    },
    {
      accessorKey: 'linkage_state',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='Linked' />
      ),
      cell: ({ row }) => (
        <Badge
          className='max-w-full truncate'
          variant={
            row.original.linkage_state === 'unlinked' ? 'default' : 'secondary'
          }
        >
          {linkageLabel(row.original.linkage_state)}
        </Badge>
      ),
      filterFn: (row, _columnId, value) =>
        linkageLabel(row.original.linkage_state)
          .toLowerCase()
          .includes(String(value).toLowerCase()),
    },
    {
      accessorKey: 'uploaded_at',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='Uploaded' />
      ),
      cell: ({ row }) => (
        <span className='block truncate'>
          {row.original.uploaded_at || 'Unknown'}
        </span>
      ),
    },
    {
      accessorKey: 'source',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='Source' />
      ),
      cell: ({ row }) => (
        <span className='block truncate'>{row.original.source}</span>
      ),
    },
    {
      accessorKey: 'download_filename',
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title='Filename' />
      ),
      cell: ({ row }) => (
        <span className='block truncate'>{row.original.download_filename}</span>
      ),
    },
    {
      id: 'actions',
      header: () => <div className='text-right'>Actions</div>,
      cell: ({ row }) => (
        <div className='flex justify-end gap-1'>
          <Button
            variant='outline'
            size='icon'
            className='h-7 w-7'
            aria-label={`Open ${row.original.title}`}
            data-testid={`media-row-open-${row.original.id}`}
          >
            <Eye />
          </Button>
          <Button
            variant='outline'
            size='icon'
            className='h-7 w-7'
            aria-label={`Analyze ${row.original.title}`}
            data-testid={`media-row-analyze-${row.original.id}`}
            disabled={row.original.analysis_status === 'ready'}
          >
            <WandSparkles />
          </Button>
          <Button
            variant='outline'
            size='icon'
            className='h-7 w-7'
            aria-label={`Assign ${row.original.title}`}
            data-testid={`media-row-assign-${row.original.id}`}
            disabled={row.original.linkage_state !== 'unlinked'}
            onClick={() => onAssign(row.original)}
          >
            <Link2 />
          </Button>
          <Button
            variant='outline'
            size='icon'
            className='h-7 w-7'
            aria-label={`Archive ${row.original.title}`}
            data-testid={`media-row-archive-${row.original.id}`}
          >
            <Archive />
          </Button>
        </div>
      ),
      enableSorting: false,
      enableHiding: false,
    },
  ]
}

export function Media() {
  const [filter, setFilter] = useState<'all' | 'unlinked'>('all')
  const [viewMode, setViewMode] = useState<MediaViewMode>(() => {
    if (typeof window === 'undefined') return 'rows'
    const saved = window.localStorage.getItem(MEDIA_VIEW_MODE_STORAGE_KEY)
    return saved === 'cards' ? 'cards' : 'rows'
  })
  const [sorting, setSorting] = useState<SortingState>([
    { id: 'uploaded_at', desc: true },
  ])
  const [globalFilter, setGlobalFilter] = useState('')
  const [assets, setAssets] = useState<MediaAsset[]>([])
  const [summary, setSummary] = useState<MediaSummary>(EMPTY_SUMMARY)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [selectedAssetIds, setSelectedAssetIds] = useState<string[]>([])
  const [downloadPreview, setDownloadPreview] =
    useState<DownloadPreview | null>(null)
  const [downloadError, setDownloadError] = useState<string | null>(null)
  const [downloadLoading, setDownloadLoading] = useState(false)
  const [assignmentAsset, setAssignmentAsset] = useState<MediaAsset | null>(
    null
  )
  const [assignmentTargetType, setAssignmentTargetType] = useState<
    'inventory' | 'wishlist'
  >('wishlist')
  const [assignmentTargetID, setAssignmentTargetID] = useState('')
  const [assignmentPreview, setAssignmentPreview] =
    useState<AssignmentPreview | null>(null)
  const [assignmentError, setAssignmentError] = useState<string | null>(null)
  const [assignmentLoading, setAssignmentLoading] = useState(false)
  const [assignmentSuccess, setAssignmentSuccess] = useState<string | null>(
    null
  )
  const [addMediaOpen, setAddMediaOpen] = useState(false)
  const [addMediaFile, setAddMediaFile] = useState<File | null>(null)
  const [addMediaTitle, setAddMediaTitle] = useState('')
  const [addMediaSource, setAddMediaSource] = useState('')
  const [addMediaNotes, setAddMediaNotes] = useState('')
  const [addMediaError, setAddMediaError] = useState<string | null>(null)
  const [addMediaSaving, setAddMediaSaving] = useState(false)
  const [isPageDragOver, setIsPageDragOver] = useState(false)
  const fileInputRef = useRef<HTMLInputElement | null>(null)

  const loadAssets = useCallback(async () => {
    setLoading(true)
    setError(null)
    setDownloadPreview(null)
    setDownloadError(null)
    try {
      const suffix = filter === 'unlinked' ? '?filter=unlinked' : ''
      const response = await fetch(`/api/media/assets${suffix}`)
      if (!response.ok) {
        throw new Error(`media_assets_${response.status}`)
      }
      const payload = (await response.json()) as MediaListResponse
      setAssets(payload.assets ?? [])
      setSummary(payload.summary ?? EMPTY_SUMMARY)
      setSelectedAssetIds((current) =>
        current.filter((id) =>
          (payload.assets ?? []).some((asset) => asset.id === id)
        )
      )
    } catch (err) {
      setAssets([])
      setSummary(EMPTY_SUMMARY)
      setSelectedAssetIds([])
      setError(err instanceof Error ? err.message : 'media_assets_failed')
    } finally {
      setLoading(false)
    }
  }, [filter])

  useEffect(() => {
    void loadAssets()
  }, [loadAssets])

  useEffect(() => {
    window.localStorage.setItem(MEDIA_VIEW_MODE_STORAGE_KEY, viewMode)
  }, [viewMode])

  const selectedAssetSet = useMemo(
    () => new Set(selectedAssetIds),
    [selectedAssetIds]
  )

  const toggleAssetSelection = useCallback(
    (assetID: string, checked: boolean) => {
      setDownloadPreview(null)
      setDownloadError(null)
      setSelectedAssetIds((current) => {
        if (checked) {
          return current.includes(assetID) ? current : [...current, assetID]
        }
        return current.filter((id) => id !== assetID)
      })
    },
    []
  )

  const resetAddMedia = () => {
    setAddMediaOpen(false)
    setAddMediaFile(null)
    setAddMediaTitle('')
    setAddMediaSource('')
    setAddMediaNotes('')
    setAddMediaError(null)
    setAddMediaSaving(false)
    setIsPageDragOver(false)
    if (fileInputRef.current) {
      fileInputRef.current.value = ''
    }
  }

  const validateImageFile = (file: File) => {
    if (ACCEPTED_IMAGE_TYPES.includes(file.type)) return null
    if (/\.(jpe?g|png|gif|webp)$/i.test(file.name)) return null
    return 'Unsupported file type. Add a JPG, PNG, GIF, or WebP image.'
  }

  const stageAddMediaFile = (file: File, source: 'page-drop' | 'dialog') => {
    const validationError = validateImageFile(file)
    setAddMediaOpen(true)
    setIsPageDragOver(false)
    if (validationError) {
      setAddMediaFile(null)
      setAddMediaError(validationError)
      return
    }
    setAddMediaFile(file)
    setAddMediaError(null)
    if (source === 'page-drop' && addMediaTitle.trim() === '') {
      setAddMediaTitle(file.name.replace(/\.[^.]+$/, ''))
    }
  }

  const handlePageDragOver = (event: DragEvent<HTMLElement>) => {
    if (!event.dataTransfer.types.includes('Files')) return
    event.preventDefault()
    event.dataTransfer.dropEffect = 'copy'
    setIsPageDragOver(true)
  }

  const handlePageDrop = (event: DragEvent<HTMLElement>) => {
    if (!event.dataTransfer.types.includes('Files')) return
    event.preventDefault()
    const file = event.dataTransfer.files.item(0)
    if (file) {
      stageAddMediaFile(file, 'page-drop')
    }
  }

  const saveAddMedia = async () => {
    if (!addMediaFile) {
      setAddMediaError('Choose an image before saving.')
      return
    }
    const validationError = validateImageFile(addMediaFile)
    if (validationError) {
      setAddMediaError(validationError)
      return
    }
    setAddMediaSaving(true)
    setAddMediaError(null)
    const formData = new FormData()
    formData.append('file', addMediaFile)
    formData.append('title', addMediaTitle.trim())
    formData.append('source', addMediaSource.trim())
    formData.append('notes', addMediaNotes.trim())
    try {
      const response = await fetch('/api/media/assets', {
        method: 'POST',
        body: formData,
      })
      if (!response.ok) {
        throw new Error(`media_asset_save_${response.status}`)
      }
      resetAddMedia()
      setAssignmentSuccess('Media asset added to the unlinked review queue.')
      await loadAssets()
    } catch (err) {
      setAddMediaError(
        err instanceof Error ? err.message : 'media_asset_save_failed'
      )
    } finally {
      setAddMediaSaving(false)
    }
  }

  const previewDownload = async () => {
    setDownloadLoading(true)
    setDownloadPreview(null)
    setDownloadError(null)
    try {
      const response = await fetch('/api/media/downloads/preview', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          asset_ids: selectedAssetIds,
          filter,
        }),
      })
      if (!response.ok) {
        throw new Error(`media_download_preview_${response.status}`)
      }
      const payload = (await response.json()) as DownloadPreview
      setDownloadPreview(payload)
    } catch (err) {
      setDownloadError(
        err instanceof Error ? err.message : 'media_download_preview_failed'
      )
    } finally {
      setDownloadLoading(false)
    }
  }

  const resetAssignment = () => {
    setAssignmentAsset(null)
    setAssignmentTargetType('wishlist')
    setAssignmentTargetID('')
    setAssignmentPreview(null)
    setAssignmentError(null)
    setAssignmentLoading(false)
  }

  const openAssignment = useCallback((asset: MediaAsset) => {
    setAssignmentAsset(asset)
    setAssignmentTargetType('wishlist')
    setAssignmentTargetID(asset.wishlist_id ?? asset.item_id ?? '')
    setAssignmentPreview(null)
    setAssignmentError(null)
    setAssignmentSuccess(null)
  }, [])

  const columns = useMemo(
    () =>
      buildMediaColumns({
        selectedAssetSet,
        onToggleAssetSelection: toggleAssetSelection,
        onAssign: openAssignment,
      }),
    [openAssignment, selectedAssetSet, toggleAssetSelection]
  )

  const table = useReactTable({
    data: assets,
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
        row.original.title,
        row.original.filename,
        row.original.download_filename,
        row.original.source,
        row.original.uploaded_at,
        analysisLabel(row.original.analysis_status),
        linkageLabel(row.original.linkage_state),
        row.original.item_id ?? '',
        row.original.wishlist_id ?? '',
      ].some((value) => value.toLowerCase().includes(searchValue))
    },
    getCoreRowModel: getCoreRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getFacetedRowModel: getFacetedRowModel(),
    getFacetedUniqueValues: getFacetedUniqueValues(),
  })

  const filteredAssetCount = table.getFilteredRowModel().rows.length

  const previewAssignment = async () => {
    if (!assignmentAsset) return
    setAssignmentLoading(true)
    setAssignmentPreview(null)
    setAssignmentError(null)
    try {
      const response = await fetch('/api/media/assignments/preview', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          asset_id: assignmentAsset.id,
          target_type: assignmentTargetType,
          target_id: assignmentTargetID.trim(),
        }),
      })
      if (!response.ok) {
        throw new Error(`media_assignment_preview_${response.status}`)
      }
      const payload = (await response.json()) as AssignmentPreview
      setAssignmentPreview(payload)
    } catch (err) {
      setAssignmentError(
        err instanceof Error ? err.message : 'media_assignment_preview_failed'
      )
    } finally {
      setAssignmentLoading(false)
    }
  }

  const confirmAssignment = async () => {
    if (!assignmentAsset || !assignmentPreview?.allowed) return
    setAssignmentLoading(true)
    setAssignmentError(null)
    try {
      const response = await fetch('/api/media/assignments', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          asset_id: assignmentAsset.id,
          target_type: assignmentPreview.target_type,
          target_id: assignmentPreview.target_id,
        }),
      })
      if (!response.ok) {
        throw new Error(`media_assignment_${response.status}`)
      }
      const payload = (await response.json()) as AssignmentPreview
      setAssignmentSuccess(
        payload.audit_summary ??
          `Assigned ${assignmentAsset.title} to ${payload.target_type} target ${payload.target_id}.`
      )
      resetAssignment()
      await loadAssets()
    } catch (err) {
      setAssignmentError(
        err instanceof Error ? err.message : 'media_assignment_failed'
      )
    } finally {
      setAssignmentLoading(false)
    }
  }

  return (
    <>
      <Header fixed>
        <Search />
        <HeaderTitle
          title='Media'
          description='Review uploaded evidence, linkage state, and assignment readiness.'
          icon={FileImage}
          testId='media-header-title'
          iconTestId='media-page-icon'
        />
        <div
          className='ms-auto flex items-center space-x-4'
          data-header-title-avoid='true'
        >
          <LanguageSwitch />
          <ThemeSwitch />
          <ConfigDrawer />
          <ProfileDropdown />
        </div>
      </Header>

      <Main
        className='relative space-y-6'
        data-testid='media-workspace'
        onDragEnter={handlePageDragOver}
        onDragOver={handlePageDragOver}
        onDragLeave={() => setIsPageDragOver(false)}
        onDrop={handlePageDrop}
      >
        {isPageDragOver ? (
          <div
            className='pointer-events-none absolute inset-0 z-20 flex items-center justify-center rounded-md border-2 border-dashed border-primary bg-background/80 text-sm font-medium shadow-sm'
            data-testid='media-page-drop-feedback'
          >
            Drop image to add media
          </div>
        ) : null}
        <div className='flex flex-wrap items-center justify-between gap-3'>
          <div>
            <h1 className='text-2xl font-bold tracking-tight'>Media</h1>
            <p className='text-muted-foreground'>
              Card-first review for uploaded photos, unlinked evidence, and
              assignment follow-up.
            </p>
          </div>
          <div className='flex flex-wrap gap-2'>
            <Button
              type='button'
              size='icon'
              aria-label='Add media'
              data-testid='media-upload-action'
              onClick={() => {
                setAddMediaOpen(true)
                setAddMediaError(null)
              }}
            >
              <ImagePlus />
            </Button>
            <Button
              variant='outline'
              disabled={selectedAssetIds.length === 0 || downloadLoading}
              data-testid='media-download-selected-action'
              onClick={() => void previewDownload()}
            >
              <Download />
              {downloadLoading ? 'Previewing...' : 'Download selected'}
            </Button>
          </div>
        </div>

        <div className='grid gap-4 sm:grid-cols-3'>
          <Card>
            <CardHeader className='pb-2'>
              <CardTitle className='text-sm font-medium'>Assets</CardTitle>
            </CardHeader>
            <CardContent className='text-2xl font-bold'>
              {loading ? '--' : summary.total}
            </CardContent>
          </Card>
          <Card>
            <CardHeader className='pb-2'>
              <CardTitle className='text-sm font-medium'>Unlinked</CardTitle>
            </CardHeader>
            <CardContent className='text-2xl font-bold'>
              {loading ? '--' : summary.unlinked}
            </CardContent>
          </Card>
          <Card>
            <CardHeader className='pb-2'>
              <CardTitle className='text-sm font-medium'>
                Ready for review
              </CardTitle>
            </CardHeader>
            <CardContent className='text-2xl font-bold'>
              {loading ? '--' : summary.ready_for_review}
            </CardContent>
          </Card>
        </div>

        <Tabs
          value={filter}
          onValueChange={(value) =>
            setFilter(value === 'unlinked' ? 'unlinked' : 'all')
          }
        >
          <TabsList aria-label='Media filters'>
            <TabsTrigger value='all' data-testid='media-filter-all'>
              All
            </TabsTrigger>
            <TabsTrigger value='unlinked' data-testid='media-filter-unlinked'>
              Unlinked
            </TabsTrigger>
          </TabsList>
        </Tabs>

        <div className='flex flex-wrap items-center justify-between gap-3'>
          <div className='text-sm text-muted-foreground'>
            {loading
              ? 'Loading view'
              : `${assets.length} asset${assets.length === 1 ? '' : 's'} in ${filter === 'unlinked' ? 'Unlinked' : 'All'} view`}
          </div>
          <div className='flex items-center gap-2' aria-label='Media view mode'>
            <Button
              type='button'
              variant={viewMode === 'cards' ? 'default' : 'outline'}
              aria-pressed={viewMode === 'cards'}
              data-testid='media-view-mode-cards'
              onClick={() => setViewMode('cards')}
            >
              Cards
            </Button>
            <Button
              type='button'
              variant={viewMode === 'rows' ? 'default' : 'outline'}
              aria-pressed={viewMode === 'rows'}
              data-testid='media-view-mode-rows'
              onClick={() => setViewMode('rows')}
            >
              Rows
            </Button>
          </div>
        </div>

        {error ? (
          <Card data-testid='media-error-state'>
            <CardHeader>
              <CardTitle>Media assets are unavailable</CardTitle>
              <CardDescription>{error}</CardDescription>
            </CardHeader>
            <CardContent>
              <Button
                variant='outline'
                data-testid='media-retry-action'
                onClick={() => void loadAssets()}
              >
                Retry
              </Button>
            </CardContent>
          </Card>
        ) : null}

        {downloadPreview ? (
          <Card data-testid='media-download-preview'>
            <CardHeader>
              <CardTitle>Download preview</CardTitle>
              <CardDescription>
                {downloadPreview.count} file
                {downloadPreview.count === 1 ? '' : 's'} ready with
                human-readable filenames.
              </CardDescription>
            </CardHeader>
            <CardContent className='text-sm text-muted-foreground'>
              {(downloadPreview.filenames ?? []).join(', ')}
            </CardContent>
          </Card>
        ) : null}

        {downloadError ? (
          <Card data-testid='media-download-error'>
            <CardHeader>
              <CardTitle>Download preview failed</CardTitle>
              <CardDescription>{downloadError}</CardDescription>
            </CardHeader>
          </Card>
        ) : null}

        {assignmentSuccess ? (
          <Card data-testid='media-assignment-success'>
            <CardHeader>
              <CardTitle>Assignment saved</CardTitle>
              <CardDescription>{assignmentSuccess}</CardDescription>
            </CardHeader>
          </Card>
        ) : null}

        {loading ? (
          <Card data-testid='media-loading-state'>
            <CardHeader>
              <CardTitle>Loading media assets...</CardTitle>
              <CardDescription>
                Retrieving profile-scoped media evidence from Cabinet.
              </CardDescription>
            </CardHeader>
          </Card>
        ) : !error && assets.length === 0 ? (
          <Card data-testid='media-empty-state'>
            <CardHeader>
              <CardTitle>No media assets in this view</CardTitle>
              <CardDescription>
                Change filters or upload evidence when new media is available.
              </CardDescription>
            </CardHeader>
          </Card>
        ) : !error && viewMode === 'cards' ? (
          <div
            className='grid gap-3 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-6'
            data-testid='media-card-grid'
          >
            {assets.map((asset) => (
              <Card
                key={asset.id}
                className='overflow-hidden'
                data-testid={`media-card-${asset.id}`}
              >
                <CardHeader className='space-y-2 p-2.5'>
                  <div className='flex aspect-video items-center justify-center overflow-hidden rounded-md border bg-muted'>
                    {asset.thumbnail_url ? (
                      <img
                        src={asset.thumbnail_url}
                        alt=''
                        className='h-full w-full object-cover'
                      />
                    ) : (
                      <FileImage className='h-7 w-7 text-muted-foreground' />
                    )}
                  </div>
                  <div className='space-y-1.5'>
                    <div className='flex items-start justify-between gap-2'>
                      <div className='flex min-w-0 flex-wrap items-center gap-1.5'>
                        <Badge
                          className='max-w-full truncate text-[11px]'
                          variant={
                            asset.linkage_state === 'unlinked'
                              ? 'default'
                              : 'secondary'
                          }
                        >
                          {linkageLabel(asset.linkage_state)}
                        </Badge>
                        <Badge
                          className='max-w-full truncate text-[11px]'
                          variant='outline'
                        >
                          {analysisLabel(asset.analysis_status)}
                        </Badge>
                      </div>
                      <Checkbox
                        aria-label={`Select ${asset.title}`}
                        checked={selectedAssetSet.has(asset.id)}
                        data-testid={`media-select-${asset.id}`}
                        onCheckedChange={(checked) =>
                          toggleAssetSelection(asset.id, checked === true)
                        }
                      />
                    </div>
                    <CardTitle className='line-clamp-2 text-sm leading-snug'>
                      {asset.title}
                    </CardTitle>
                    <CardDescription className='truncate text-xs'>
                      {asset.source}
                    </CardDescription>
                  </div>
                </CardHeader>
                <CardContent className='space-y-2.5 p-2.5 pt-0'>
                  <dl className='grid grid-cols-2 gap-2 text-[11px]'>
                    <div className='min-w-0'>
                      <dt className='text-muted-foreground'>Uploaded</dt>
                      <dd className='truncate'>
                        {asset.uploaded_at || 'Unknown'}
                      </dd>
                    </div>
                    <div className='min-w-0'>
                      <dt className='text-muted-foreground'>Filename</dt>
                      <dd className='truncate'>{asset.download_filename}</dd>
                    </div>
                  </dl>
                  <div className='grid grid-cols-4 gap-1'>
                    <Button
                      variant='outline'
                      size='icon'
                      className='h-7 w-7'
                      aria-label={`Open ${asset.title}`}
                      data-testid={`media-open-${asset.id}`}
                    >
                      <Eye />
                    </Button>
                    <Button
                      variant='outline'
                      size='icon'
                      className='h-7 w-7'
                      aria-label={`Analyze ${asset.title}`}
                      data-testid={`media-analyze-${asset.id}`}
                      disabled={asset.analysis_status === 'ready'}
                    >
                      <WandSparkles />
                    </Button>
                    <Button
                      variant='outline'
                      size='icon'
                      className='h-7 w-7'
                      aria-label={`Assign ${asset.title}`}
                      data-testid={`media-assign-${asset.id}`}
                      disabled={asset.linkage_state !== 'unlinked'}
                      onClick={() => openAssignment(asset)}
                    >
                      <Link2 />
                    </Button>
                    <Button
                      variant='outline'
                      size='icon'
                      className='h-7 w-7'
                      aria-label={`Archive ${asset.title}`}
                      data-testid={`media-archive-${asset.id}`}
                    >
                      <Archive />
                    </Button>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        ) : !error ? (
          <div className='space-y-3' data-testid='media-table-section'>
            <div data-testid='media-table-toolbar'>
              <DataTableToolbar
                table={table}
                searchPlaceholder='Filter media...'
                searchInputTestId='media-table-search-input'
              />
            </div>
            <div
              className='flex flex-wrap items-center gap-2 text-sm text-muted-foreground'
              data-testid='media-table-summary'
            >
              <span>
                Showing {filteredAssetCount} of {assets.length} media assets.
              </span>
              <span>
                {filter === 'unlinked' ? 'Unlinked media' : 'All media'}
              </span>
            </div>
            <div
              className='overflow-auto rounded-md border'
              data-table-surface='true'
              data-testid='media-shared-table'
            >
              <Table
                className='w-full min-w-[920px] table-fixed text-xs'
                data-testid='media-row-table'
              >
                <TableHeader>
                  {table.getHeaderGroups().map((headerGroup) => (
                    <TableRow key={headerGroup.id}>
                      {headerGroup.headers.map((header) => (
                        <TableHead
                          key={header.id}
                          className={
                            header.column.id === 'select'
                              ? 'w-12'
                              : header.column.id === 'actions'
                                ? 'w-36 text-end'
                                : 'max-w-0 truncate'
                          }
                        >
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
                  {table.getRowModel().rows.length ? (
                    table.getRowModel().rows.map((row) => (
                      <TableRow
                        key={row.id}
                        data-testid={`media-row-${row.original.id}`}
                      >
                        {row.getVisibleCells().map((cell) => (
                          <TableCell
                            key={cell.id}
                            className={
                              cell.column.id === 'select'
                                ? 'w-12'
                                : cell.column.id === 'actions'
                                  ? 'w-36'
                                  : 'max-w-0 truncate'
                            }
                          >
                            {flexRender(
                              cell.column.columnDef.cell,
                              cell.getContext()
                            )}
                          </TableCell>
                        ))}
                      </TableRow>
                    ))
                  ) : (
                    <TableRow data-testid='media-table-empty-row'>
                      <TableCell
                        colSpan={columns.length}
                        className='h-24 text-center'
                      >
                        No media assets match the current table filter.
                      </TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
            </div>
            <div data-testid='media-table-pagination'>
              <DataTablePagination table={table} />
            </div>
          </div>
        ) : null}

        <Dialog
          open={addMediaOpen}
          onOpenChange={(open) => {
            if (open) {
              setAddMediaOpen(true)
              setAddMediaError(null)
              return
            }
            resetAddMedia()
          }}
        >
          <DialogContent data-testid='media-add-dialog'>
            <DialogHeader>
              <DialogTitle>Add media</DialogTitle>
            </DialogHeader>
            <div className='space-y-4'>
              <button
                type='button'
                className='flex w-full flex-col items-center justify-center gap-2 rounded-md border border-dashed p-6 text-center text-sm text-muted-foreground hover:bg-muted/50'
                data-testid='media-add-dropzone'
                onClick={() => fileInputRef.current?.click()}
                onDragOver={(event) => {
                  event.preventDefault()
                  event.dataTransfer.dropEffect = 'copy'
                }}
                onDrop={(event) => {
                  event.preventDefault()
                  const file = event.dataTransfer.files.item(0)
                  if (file) {
                    stageAddMediaFile(file, 'dialog')
                  }
                }}
              >
                <ImagePlus className='h-6 w-6' />
                <span>
                  {addMediaFile
                    ? addMediaFile.name
                    : 'Drop image or choose file'}
                </span>
              </button>
              <input
                ref={fileInputRef}
                type='file'
                accept='image/jpeg,image/png,image/gif,image/webp'
                className='hidden'
                data-testid='media-add-file-input'
                onChange={(event) => {
                  const file = event.target.files?.item(0)
                  if (file) {
                    stageAddMediaFile(file, 'dialog')
                  }
                }}
              />
              <div className='grid gap-2'>
                <Label htmlFor='media-add-title'>Title</Label>
                <Input
                  id='media-add-title'
                  value={addMediaTitle}
                  data-testid='media-add-title'
                  onChange={(event) => setAddMediaTitle(event.target.value)}
                />
              </div>
              <div className='grid gap-2'>
                <Label htmlFor='media-add-source'>Source</Label>
                <Input
                  id='media-add-source'
                  value={addMediaSource}
                  data-testid='media-add-source'
                  onChange={(event) => setAddMediaSource(event.target.value)}
                />
              </div>
              <div className='grid gap-2'>
                <Label htmlFor='media-add-notes'>Notes</Label>
                <Textarea
                  id='media-add-notes'
                  value={addMediaNotes}
                  data-testid='media-add-notes'
                  onChange={(event) => setAddMediaNotes(event.target.value)}
                />
              </div>
              {addMediaError ? (
                <div
                  className='rounded-md border border-destructive/40 p-3 text-sm text-destructive'
                  data-testid='media-add-error'
                >
                  {addMediaError}
                </div>
              ) : null}
            </div>
            <DialogFooter>
              <Button
                variant='outline'
                onClick={() => resetAddMedia()}
                disabled={addMediaSaving}
              >
                Cancel
              </Button>
              <Button
                onClick={() => void saveAddMedia()}
                disabled={addMediaSaving || addMediaFile === null}
                data-testid='media-add-save-action'
              >
                {addMediaSaving ? 'Saving...' : 'Save media'}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        <Dialog
          open={assignmentAsset !== null}
          onOpenChange={(open) => {
            if (!open) resetAssignment()
          }}
        >
          <DialogContent data-testid='media-assignment-dialog'>
            <DialogHeader>
              <DialogTitle>Assign media</DialogTitle>
            </DialogHeader>
            <div className='space-y-4'>
              <div className='text-sm text-muted-foreground'>
                {assignmentAsset?.title}
              </div>
              <div className='grid gap-2'>
                <Label htmlFor='media-assignment-target-type'>
                  Target type
                </Label>
                <Select
                  value={assignmentTargetType}
                  onValueChange={(value) => {
                    setAssignmentTargetType(
                      value === 'inventory' ? 'inventory' : 'wishlist'
                    )
                    setAssignmentPreview(null)
                    setAssignmentError(null)
                  }}
                >
                  <SelectTrigger
                    id='media-assignment-target-type'
                    data-testid='media-assignment-target-type'
                  >
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value='wishlist'>Wishlist</SelectItem>
                    <SelectItem value='inventory'>Inventory</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className='grid gap-2'>
                <Label htmlFor='media-assignment-target-id'>Target ID</Label>
                <Input
                  id='media-assignment-target-id'
                  value={assignmentTargetID}
                  data-testid='media-assignment-target-id'
                  onChange={(event) => {
                    setAssignmentTargetID(event.target.value)
                    setAssignmentPreview(null)
                    setAssignmentError(null)
                  }}
                />
              </div>
              {assignmentPreview ? (
                <div
                  className='rounded-md border p-3 text-sm'
                  data-testid='media-assignment-preview'
                >
                  <div>
                    {linkageLabel(assignmentPreview.current_linkage_state)} to{' '}
                    {linkageLabel(assignmentPreview.projected_linkage_state)}
                  </div>
                  <div className='mt-1 text-muted-foreground'>
                    {assignmentPreview.audit_summary}
                  </div>
                </div>
              ) : null}
              {assignmentError ? (
                <div
                  className='rounded-md border border-destructive/40 p-3 text-sm text-destructive'
                  data-testid='media-assignment-error'
                >
                  {assignmentError}
                </div>
              ) : null}
            </div>
            <DialogFooter>
              <Button
                variant='outline'
                onClick={() => resetAssignment()}
                disabled={assignmentLoading}
              >
                Cancel
              </Button>
              <Button
                variant='outline'
                onClick={() => void previewAssignment()}
                disabled={assignmentLoading || assignmentTargetID.trim() === ''}
                data-testid='media-assignment-preview-action'
              >
                {assignmentLoading ? 'Working...' : 'Preview'}
              </Button>
              <Button
                onClick={() => void confirmAssignment()}
                disabled={
                  assignmentLoading ||
                  !assignmentPreview?.allowed ||
                  assignmentPreview.requires_confirmation !== true
                }
                data-testid='media-assignment-confirm-action'
              >
                Confirm assignment
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </Main>
    </>
  )
}
