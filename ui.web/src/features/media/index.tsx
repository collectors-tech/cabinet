import { useCallback, useEffect, useMemo, useState } from 'react'
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
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { ConfigDrawer } from '@/components/config-drawer'
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

export function Media() {
  const [filter, setFilter] = useState<'all' | 'unlinked'>('all')
  const [assets, setAssets] = useState<MediaAsset[]>([])
  const [summary, setSummary] = useState<MediaSummary>(EMPTY_SUMMARY)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [selectedAssetIds, setSelectedAssetIds] = useState<string[]>([])
  const [downloadPreview, setDownloadPreview] = useState<DownloadPreview | null>(
    null
  )
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
  const [assignmentSuccess, setAssignmentSuccess] = useState<string | null>(null)

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

  const selectedAssetSet = useMemo(
    () => new Set(selectedAssetIds),
    [selectedAssetIds]
  )

  const toggleAssetSelection = (assetID: string, checked: boolean) => {
    setDownloadPreview(null)
    setDownloadError(null)
    setSelectedAssetIds((current) => {
      if (checked) {
        return current.includes(assetID) ? current : [...current, assetID]
      }
      return current.filter((id) => id !== assetID)
    })
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

  const openAssignment = (asset: MediaAsset) => {
    setAssignmentAsset(asset)
    setAssignmentTargetType('wishlist')
    setAssignmentTargetID(asset.wishlist_id ?? asset.item_id ?? '')
    setAssignmentPreview(null)
    setAssignmentError(null)
    setAssignmentSuccess(null)
  }

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

      <Main className='space-y-6' data-testid='media-workspace'>
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
              variant='outline'
              disabled
              data-testid='media-upload-action'
            >
              <ImagePlus />
              Upload
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
        ) : !error ? (
          <div
            className='grid gap-4 md:grid-cols-2 xl:grid-cols-3'
            data-testid='media-card-grid'
          >
            {assets.map((asset) => (
              <Card key={asset.id} data-testid={`media-card-${asset.id}`}>
                <CardHeader className='space-y-3'>
                  <div className='flex aspect-video items-center justify-center overflow-hidden rounded-md border bg-muted'>
                    {asset.thumbnail_url ? (
                      <img
                        src={asset.thumbnail_url}
                        alt=''
                        className='h-full w-full object-cover'
                      />
                    ) : (
                      <FileImage className='h-10 w-10 text-muted-foreground' />
                    )}
                  </div>
                  <div className='space-y-2'>
                    <div className='flex flex-wrap items-center justify-between gap-2'>
                      <div className='flex flex-wrap items-center gap-2'>
                        <Badge
                          variant={
                            asset.linkage_state === 'unlinked'
                              ? 'default'
                              : 'secondary'
                          }
                        >
                          {linkageLabel(asset.linkage_state)}
                        </Badge>
                        <Badge variant='outline'>
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
                    <CardTitle className='text-base'>{asset.title}</CardTitle>
                    <CardDescription>{asset.source}</CardDescription>
                  </div>
                </CardHeader>
                <CardContent className='space-y-4'>
                  <dl className='grid grid-cols-2 gap-3 text-sm'>
                    <div>
                      <dt className='text-muted-foreground'>Uploaded</dt>
                      <dd>{asset.uploaded_at || 'Unknown'}</dd>
                    </div>
                    <div>
                      <dt className='text-muted-foreground'>Filename</dt>
                      <dd className='truncate'>{asset.download_filename}</dd>
                    </div>
                  </dl>
                  <div className='grid grid-cols-4 gap-2'>
                    <Button
                      variant='outline'
                      size='icon'
                      aria-label={`Open ${asset.title}`}
                      data-testid={`media-open-${asset.id}`}
                    >
                      <Eye />
                    </Button>
                    <Button
                      variant='outline'
                      size='icon'
                      aria-label={`Analyze ${asset.title}`}
                      data-testid={`media-analyze-${asset.id}`}
                      disabled={asset.analysis_status === 'ready'}
                    >
                      <WandSparkles />
                    </Button>
                    <Button
                      variant='outline'
                      size='icon'
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
        ) : null}

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
