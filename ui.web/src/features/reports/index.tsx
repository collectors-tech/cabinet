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

type PricingStats = {
  min: number
  median: number
  latest: number
}

type ReportsData = {
  profileID: string
  wishlistHits: number
  pricingStats: PricingStats
  trendPoints: number
  sourceCount: number
}

function parseNumber(value: unknown): number {
  if (typeof value === 'number') {
    return value
  }
  if (typeof value === 'string' && value.trim() !== '') {
    const parsed = Number(value)
    if (!Number.isNaN(parsed)) {
      return parsed
    }
  }
  return 0
}

function formatCurrency(value: number) {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    maximumFractionDigits: 2,
  }).format(value)
}

export function Reports() {
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [data, setData] = useState<ReportsData | null>(null)
  const [exportMessage, setExportMessage] = useState<string | null>(null)

  const loadReports = useCallback(async () => {
    setLoading(true)
    setError(null)
    setExportMessage(null)
    try {
      const activeResponse = await fetch('/api/profiles/active')
      if (!activeResponse.ok) {
        throw new Error(`active_profile_${activeResponse.status}`)
      }
      const activePayload = (await activeResponse.json()) as { id?: string }
      const profileID = activePayload.id?.trim() ?? ''
      if (!profileID) {
        throw new Error('active_profile_missing')
      }

      const [wishlistResp, statsResp, trendResp, sourceResp] = await Promise.all([
        fetch(`/api/wishlist/hits?profile_id=${encodeURIComponent(profileID)}`),
        fetch(`/api/pricing/stats?profile_id=${encodeURIComponent(profileID)}`),
        fetch(`/api/pricing/trend?profile_id=${encodeURIComponent(profileID)}`),
        fetch(`/api/pricing/by-source?profile_id=${encodeURIComponent(profileID)}`),
      ])

      if (!wishlistResp.ok) {
        throw new Error(`reports_wishlist_hits_${wishlistResp.status}`)
      }
      if (!statsResp.ok) {
        throw new Error(`reports_pricing_stats_${statsResp.status}`)
      }
      if (!trendResp.ok) {
        throw new Error(`reports_pricing_trend_${trendResp.status}`)
      }
      if (!sourceResp.ok) {
        throw new Error(`reports_pricing_source_${sourceResp.status}`)
      }

      const wishlistPayload = (await wishlistResp.json()) as {
        hits?: Array<unknown>
      }
      const statsPayload = (await statsResp.json()) as {
        min?: number
        median?: number
        latest?: number
      }
      const trendPayload = (await trendResp.json()) as {
        points?: Array<unknown>
      }
      const sourcePayload = (await sourceResp.json()) as {
        sources?: Record<string, unknown>
      }

      setData({
        profileID,
        wishlistHits: wishlistPayload.hits?.length ?? 0,
        pricingStats: {
          min: parseNumber(statsPayload.min),
          median: parseNumber(statsPayload.median),
          latest: parseNumber(statsPayload.latest),
        },
        trendPoints: trendPayload.points?.length ?? 0,
        sourceCount: Object.keys(sourcePayload.sources ?? {}).length,
      })
    } catch (err) {
      setError(err instanceof Error ? err.message : 'reports_load_failed')
      setData(null)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void loadReports()
  }, [loadReports])

  const isEmptyState = useMemo(() => {
    if (!data) {
      return false
    }
    return (
      data.wishlistHits === 0 &&
      data.trendPoints === 0 &&
      data.sourceCount === 0 &&
      data.pricingStats.latest === 0 &&
      data.pricingStats.median === 0 &&
      data.pricingStats.min === 0
    )
  }, [data])

  const exportReport = async () => {
    setExportMessage(null)
    try {
      const response = await fetch('/api/data/export/csv/items')
      if (!response.ok) {
        throw new Error(`reports_export_${response.status}`)
      }
      const csv = await response.text()
      setExportMessage(
        `Export generated (${csv.split('\n').filter((line) => line.trim() !== '').length} lines).`
      )
    } catch (err) {
      setExportMessage(
        err instanceof Error ? err.message : 'reports_export_failed'
      )
    }
  }

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

      <Main className='space-y-6'>
        <div className='flex flex-wrap items-center justify-between gap-3'>
          <div>
            <h1 className='text-2xl font-bold tracking-tight'>Reports</h1>
            <p className='text-muted-foreground'>
              Wishlist and pricing analytics with export-ready snapshots.
            </p>
          </div>
          <div className='flex gap-2'>
            <Button variant='outline' onClick={() => void loadReports()} disabled={loading}>
              {loading ? 'Refreshing...' : 'Refresh Reports'}
            </Button>
            <Button
              data-testid='reports-export-button'
              onClick={() => void exportReport()}
              disabled={loading}
            >
              Export CSV
            </Button>
          </div>
        </div>

        {error ? (
          <Card data-testid='reports-error'>
            <CardHeader>
              <CardTitle>Reports unavailable</CardTitle>
              <CardDescription>{error}</CardDescription>
            </CardHeader>
            <CardContent>
              <Button variant='outline' onClick={() => void loadReports()}>
                Retry
              </Button>
            </CardContent>
          </Card>
        ) : null}

        {exportMessage ? (
          <div className='rounded-md border bg-muted/30 px-3 py-2 text-sm' data-testid='reports-export-message'>
            {exportMessage}
          </div>
        ) : null}

        <div className='grid gap-4 sm:grid-cols-2 lg:grid-cols-4'>
          {loading ? (
            [1, 2, 3, 4].map((slot) => (
              <Card key={slot}>
                <CardHeader className='pb-2'>
                  <CardTitle className='text-sm font-medium'>Loading...</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className='text-2xl font-bold'>--</div>
                </CardContent>
              </Card>
            ))
          ) : (
            <>
              <Card>
                <CardHeader className='pb-2'>
                  <CardTitle className='text-sm font-medium'>Wishlist Hits</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className='text-2xl font-bold'>{data?.wishlistHits ?? 0}</div>
                </CardContent>
              </Card>
              <Card>
                <CardHeader className='pb-2'>
                  <CardTitle className='text-sm font-medium'>Price Min</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className='text-2xl font-bold'>
                    {formatCurrency(data?.pricingStats.min ?? 0)}
                  </div>
                </CardContent>
              </Card>
              <Card>
                <CardHeader className='pb-2'>
                  <CardTitle className='text-sm font-medium'>Price Median</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className='text-2xl font-bold'>
                    {formatCurrency(data?.pricingStats.median ?? 0)}
                  </div>
                </CardContent>
              </Card>
              <Card>
                <CardHeader className='pb-2'>
                  <CardTitle className='text-sm font-medium'>Price Latest</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className='text-2xl font-bold'>
                    {formatCurrency(data?.pricingStats.latest ?? 0)}
                  </div>
                </CardContent>
              </Card>
            </>
          )}
        </div>

        {!loading && !error && isEmptyState ? (
          <Card data-testid='reports-empty-state'>
            <CardHeader>
              <CardTitle>No report data yet</CardTitle>
              <CardDescription>
                Track pricing or add wishlist targets to populate reports.
              </CardDescription>
            </CardHeader>
          </Card>
        ) : null}
      </Main>
    </>
  )
}
