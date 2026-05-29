import { useMemo, useState } from 'react'
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
  uploadedAt: string
  linkageState: 'unlinked' | 'linked_inventory' | 'linked_wishlist'
  analysisStatus: 'ready' | 'pending' | 'not_analyzed'
  confidence?: string
  source: string
}

const MEDIA_ASSETS: MediaAsset[] = [
  {
    id: 'media-slot-car-front',
    title: 'AFX Mustang front view',
    uploadedAt: '2026-05-25 10:20',
    linkageState: 'unlinked',
    analysisStatus: 'ready',
    confidence: '92%',
    source: 'Scanner intake',
  },
  {
    id: 'media-porsche-box',
    title: 'Porsche 917 box side',
    uploadedAt: '2026-05-24 16:45',
    linkageState: 'linked_inventory',
    analysisStatus: 'pending',
    source: 'Inventory photo',
  },
  {
    id: 'media-wishlist-reference',
    title: 'Wanted chassis reference',
    uploadedAt: '2026-05-23 08:12',
    linkageState: 'linked_wishlist',
    analysisStatus: 'not_analyzed',
    source: 'Wishlist evidence',
  },
]

function linkageLabel(state: MediaAsset['linkageState']) {
  switch (state) {
    case 'linked_inventory':
      return 'Inventory linked'
    case 'linked_wishlist':
      return 'Wishlist linked'
    default:
      return 'Unlinked'
  }
}

function analysisLabel(status: MediaAsset['analysisStatus']) {
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
  const visibleAssets = useMemo(
    () =>
      filter === 'unlinked'
        ? MEDIA_ASSETS.filter((asset) => asset.linkageState === 'unlinked')
        : MEDIA_ASSETS,
    [filter]
  )
  const unlinkedCount = MEDIA_ASSETS.filter(
    (asset) => asset.linkageState === 'unlinked'
  ).length

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
              disabled
              data-testid='media-download-selected-action'
            >
              <Download />
              Download selected
            </Button>
          </div>
        </div>

        <div className='grid gap-4 sm:grid-cols-3'>
          <Card>
            <CardHeader className='pb-2'>
              <CardTitle className='text-sm font-medium'>Assets</CardTitle>
            </CardHeader>
            <CardContent className='text-2xl font-bold'>
              {MEDIA_ASSETS.length}
            </CardContent>
          </Card>
          <Card>
            <CardHeader className='pb-2'>
              <CardTitle className='text-sm font-medium'>Unlinked</CardTitle>
            </CardHeader>
            <CardContent className='text-2xl font-bold'>
              {unlinkedCount}
            </CardContent>
          </Card>
          <Card>
            <CardHeader className='pb-2'>
              <CardTitle className='text-sm font-medium'>
                Ready for review
              </CardTitle>
            </CardHeader>
            <CardContent className='text-2xl font-bold'>
              {
                MEDIA_ASSETS.filter((asset) => asset.analysisStatus === 'ready')
                  .length
              }
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

        {visibleAssets.length === 0 ? (
          <Card data-testid='media-empty-state'>
            <CardHeader>
              <CardTitle>No media assets in this view</CardTitle>
              <CardDescription>
                Change filters or upload evidence when the media ingestion API
                slice is available.
              </CardDescription>
            </CardHeader>
          </Card>
        ) : (
          <div
            className='grid gap-4 md:grid-cols-2 xl:grid-cols-3'
            data-testid='media-card-grid'
          >
            {visibleAssets.map((asset) => (
              <Card key={asset.id} data-testid={`media-card-${asset.id}`}>
                <CardHeader className='space-y-3'>
                  <div className='flex aspect-video items-center justify-center rounded-md border bg-muted'>
                    <FileImage className='h-10 w-10 text-muted-foreground' />
                  </div>
                  <div className='space-y-2'>
                    <div className='flex flex-wrap items-center gap-2'>
                      <Badge
                        variant={
                          asset.linkageState === 'unlinked'
                            ? 'default'
                            : 'secondary'
                        }
                      >
                        {linkageLabel(asset.linkageState)}
                      </Badge>
                      <Badge variant='outline'>
                        {analysisLabel(asset.analysisStatus)}
                      </Badge>
                    </div>
                    <CardTitle className='text-base'>{asset.title}</CardTitle>
                    <CardDescription>{asset.source}</CardDescription>
                  </div>
                </CardHeader>
                <CardContent className='space-y-4'>
                  <dl className='grid grid-cols-2 gap-3 text-sm'>
                    <div>
                      <dt className='text-muted-foreground'>Uploaded</dt>
                      <dd>{asset.uploadedAt}</dd>
                    </div>
                    <div>
                      <dt className='text-muted-foreground'>Confidence</dt>
                      <dd>{asset.confidence ?? 'Pending'}</dd>
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
                      disabled={asset.analysisStatus === 'ready'}
                    >
                      <WandSparkles />
                    </Button>
                    <Button
                      variant='outline'
                      size='icon'
                      aria-label={`Assign ${asset.title}`}
                      data-testid={`media-assign-${asset.id}`}
                      disabled={asset.linkageState !== 'unlinked'}
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
        )}
      </Main>
    </>
  )
}
