import { useCallback, useEffect, useMemo, useState } from 'react'
import {
  AlertTriangle,
  Boxes,
  Heart,
  LayoutDashboard,
  PackageCheck,
  SearchCheck,
  TrendingDown,
} from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { ConfigDrawer } from '@/components/config-drawer'
import { LanguageSwitch } from '@/components/language-switch'
import { Header, HeaderTitle } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'

type DashboardCard = {
  title: string
  value: number
  link: string
}

type DashboardSummary = {
  new_discoveries: number
  wishlist_hits: number
  price_drops: number
  low_stock_discoveries: number
  restocks: number
  recently_added: string[]
  total_items: number
  total_instances: number
  estimated_value: number
  cards: DashboardCard[]
}

function formatCurrency(value: number) {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    maximumFractionDigits: 2,
  }).format(value)
}

function pluralize(value: number, singular: string, plural = `${singular}s`) {
  return `${value} ${value === 1 ? singular : plural}`
}

function normalizeDashboardLink(link: string) {
  if (link === '/pricing') {
    return '/wishlist'
  }
  return link
}

export function Dashboard() {
  const { t } = useTranslation('pages')
  const [summary, setSummary] = useState<DashboardSummary | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const loadDashboard = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const response = await fetch('/api/dashboard')
      if (!response.ok) {
        throw new Error(`dashboard_fetch_failed_${response.status}`)
      }
      const payload = (await response.json()) as DashboardSummary
      setSummary(payload)
    } catch (err) {
      const message =
        err instanceof Error ? err.message : 'dashboard_fetch_failed'
      setError(message)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void loadDashboard()
  }, [loadDashboard])

  const metricCards = useMemo(() => {
    if (!summary) {
      return []
    }
    return [
      {
        title: 'Collection size',
        value: `${summary.total_items}`,
        detail: pluralize(summary.total_instances, 'tracked unit'),
        icon: Boxes,
      },
      {
        title: 'Wishlist hits',
        value: `${summary.wishlist_hits}`,
        detail: pluralize(summary.price_drops, 'price drop'),
        icon: Heart,
      },
      {
        title: 'Operational alerts',
        value: `${summary.new_discoveries + summary.low_stock_discoveries + summary.restocks}`,
        detail: `${pluralize(summary.new_discoveries, 'discovery', 'discoveries')} ready`,
        icon: AlertTriangle,
      },
      {
        title: 'Collection value',
        value: formatCurrency(summary.estimated_value),
        detail: 'Inventory estimate',
        icon: TrendingDown,
      },
    ]
  }, [summary])

  const attentionTotal = summary
    ? summary.new_discoveries +
      summary.low_stock_discoveries +
      summary.wishlist_hits +
      summary.price_drops +
      summary.restocks
    : 0

  return (
    <>
      <Header fixed>
        <Search />
        <HeaderTitle
          title={t('dashboard.title')}
          description={t('dashboard.subtitle')}
          icon={LayoutDashboard}
          testId='dashboard-header-title'
          iconTestId='dashboard-page-icon'
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

      <Main className='space-y-6'>
        <div className='flex flex-wrap items-start justify-between gap-3'>
          <div className='max-w-3xl'>
            <div className='mb-2 flex flex-wrap items-center gap-2'>
              <Badge variant='secondary'>Signal hub</Badge>
              {summary && !loading ? (
                <span className='text-sm text-muted-foreground'>
                  {pluralize(attentionTotal, 'signal')} requiring review
                </span>
              ) : null}
            </div>
            <h1 className='text-2xl font-bold tracking-tight'>
              {t('dashboard.title')}
            </h1>
            <p className='text-muted-foreground'>{t('dashboard.subtitle')}</p>
          </div>
          <Button onClick={() => void loadDashboard()} disabled={loading}>
            {loading ? t('dashboard.refreshing') : t('dashboard.refresh')}
          </Button>
        </div>

        {error ? (
          <Card>
            <CardHeader>
              <CardTitle>{t('dashboard.unavailable')}</CardTitle>
              <CardDescription>{error}</CardDescription>
            </CardHeader>
            <CardContent>
              <Button variant='outline' onClick={() => void loadDashboard()}>
                {t('common:actions.retry')}
              </Button>
            </CardContent>
          </Card>
        ) : null}

        <section
          className='rounded-lg border bg-card p-4 shadow-sm'
          data-testid='dashboard-signal-hub'
        >
          <div className='mb-4 flex flex-wrap items-center justify-between gap-2'>
            <div>
              <h2 className='text-base font-semibold'>Signal hub</h2>
              <p className='text-sm text-muted-foreground'>
                Collector and inventory manager state at a glance.
              </p>
            </div>
            <Button asChild variant='outline' size='sm'>
              <a href='/reports'>Open reports</a>
            </Button>
          </div>
          <div className='grid gap-3 sm:grid-cols-2 xl:grid-cols-4'>
            {loading
              ? [1, 2, 3, 4].map((slot) => (
                  <div key={slot} className='rounded-md border p-4'>
                    <div className='mb-3 text-sm font-medium'>
                      {t('common:status.loading')}
                    </div>
                    <div className='text-2xl font-bold'>--</div>
                  </div>
                ))
              : metricCards.map((card) => (
                  <div
                    key={card.title}
                    className='rounded-md border bg-background p-4'
                  >
                    <div className='mb-4 flex items-center justify-between gap-3'>
                      <div className='text-sm font-medium'>{card.title}</div>
                      <card.icon className='size-4 text-muted-foreground' />
                    </div>
                    <div className='text-2xl font-bold'>{card.value}</div>
                    <div className='mt-1 text-xs text-muted-foreground'>
                      {card.detail}
                    </div>
                  </div>
                ))}
          </div>
        </section>

        <div className='grid grid-cols-1 gap-4 xl:grid-cols-12'>
          <Card
            className='xl:col-span-5'
            data-testid='dashboard-collector-health'
          >
            <CardHeader>
              <CardTitle>Collector health</CardTitle>
              <CardDescription>
                Collection growth, recency, and discovery momentum.
              </CardDescription>
            </CardHeader>
            <CardContent className='space-y-4'>
              <div className='grid gap-3 sm:grid-cols-2'>
                <div className='rounded-md border p-3'>
                  <div className='flex items-center gap-2 text-sm font-medium'>
                    <PackageCheck className='size-4 text-muted-foreground' />
                    Collection growth
                  </div>
                  <div className='mt-2 text-2xl font-bold'>
                    {loading ? '--' : (summary?.total_items ?? 0)}
                  </div>
                  <p className='text-xs text-muted-foreground'>
                    Cataloged items in the active profile
                  </p>
                </div>
                <div className='rounded-md border p-3'>
                  <div className='flex items-center gap-2 text-sm font-medium'>
                    <SearchCheck className='size-4 text-muted-foreground' />
                    Market signals
                  </div>
                  <div className='mt-2 text-2xl font-bold'>
                    {loading ? '--' : (summary?.new_discoveries ?? 0)}
                  </div>
                  <p className='text-xs text-muted-foreground'>
                    New discoveries outside the collection
                  </p>
                </div>
              </div>
              <div>
                <div className='mb-2 flex items-center justify-between gap-3'>
                  <div>
                    <h3 className='text-sm font-medium'>Recent additions</h3>
                    <p className='text-xs text-muted-foreground'>
                      {t('dashboard.recentlyAdded')}
                    </p>
                  </div>
                  <Button asChild variant='outline' size='sm'>
                    <a href='/inventory'>{t('dashboard.openInventory')}</a>
                  </Button>
                </div>
                {loading ? (
                  <p className='text-sm text-muted-foreground'>
                    {t('dashboard.loadingItems')}
                  </p>
                ) : summary?.recently_added?.length ? (
                  <ul className='space-y-2'>
                    {summary.recently_added.map((item) => (
                      <li key={item}>
                        <a
                          href='/inventory'
                          data-testid={`dashboard-recent-item-${item.toLowerCase().replace(/[^a-z0-9]+/g, '-')}`}
                          className='flex items-center justify-between rounded-md border px-3 py-2 text-sm transition-colors hover:bg-muted'
                        >
                          <span className='truncate'>{item}</span>
                          <span className='text-xs text-muted-foreground'>
                            {t('dashboard.openInventory')}
                          </span>
                        </a>
                      </li>
                    ))}
                  </ul>
                ) : (
                  <p className='rounded-md border border-dashed p-3 text-sm text-muted-foreground'>
                    {t('dashboard.noRecentlyAdded')}
                  </p>
                )}
              </div>
            </CardContent>
          </Card>

          <Card
            className='xl:col-span-4'
            data-testid='dashboard-purchase-pipeline'
          >
            <CardHeader>
              <CardTitle>Purchase pipeline</CardTitle>
              <CardDescription>
                Wishlist, price, and stock movement ready for review.
              </CardDescription>
            </CardHeader>
            <CardContent className='space-y-3'>
              <div className='rounded-md border p-3'>
                <div className='font-medium'>
                  {loading
                    ? '--'
                    : pluralize(summary?.wishlist_hits ?? 0, 'wishlist hit')}
                </div>
                <p className='text-sm text-muted-foreground'>
                  Wanted items with marketplace matches
                </p>
              </div>
              <div className='rounded-md border p-3'>
                <div className='font-medium'>
                  {loading
                    ? '--'
                    : pluralize(summary?.price_drops ?? 0, 'price drop')}
                </div>
                <p className='text-sm text-muted-foreground'>
                  Watched items trending below prior snapshots
                </p>
              </div>
              <div className='rounded-md border p-3'>
                <div className='font-medium'>
                  {loading
                    ? '--'
                    : pluralize(summary?.restocks ?? 0, 'restock')}
                </div>
                <p className='text-sm text-muted-foreground'>
                  Previously unavailable items that came back
                </p>
              </div>
              <div className='flex flex-wrap gap-2 pt-1'>
                <Button asChild variant='outline' size='sm'>
                  <a href='/wishlist'>Open wishlist</a>
                </Button>
                <Button asChild variant='outline' size='sm'>
                  <a href='/purchases'>Open purchases</a>
                </Button>
              </div>
            </CardContent>
          </Card>

          <Card
            className='xl:col-span-3'
            data-testid='dashboard-inventory-readiness'
          >
            <CardHeader>
              <CardTitle>Inventory readiness</CardTitle>
              <CardDescription>
                Operational coverage for records, media, and valuation.
              </CardDescription>
            </CardHeader>
            <CardContent className='space-y-3'>
              <div className='rounded-md border p-3'>
                <div className='font-medium'>
                  {loading
                    ? '--'
                    : pluralize(summary?.total_instances ?? 0, 'tracked unit')}
                </div>
                <p className='text-sm text-muted-foreground'>
                  Units represented in inventory
                </p>
              </div>
              <div className='rounded-md border p-3'>
                <div className='font-medium'>
                  {loading
                    ? '--'
                    : formatCurrency(summary?.estimated_value ?? 0)}
                </div>
                <p className='text-sm text-muted-foreground'>
                  Estimated collection value
                </p>
              </div>
              <Button asChild variant='outline' size='sm'>
                <a href='/media'>Review media</a>
              </Button>
            </CardContent>
          </Card>

          <Card
            className='xl:col-span-12'
            data-testid='dashboard-actions-needed'
          >
            <CardHeader>
              <CardTitle>Actions needed</CardTitle>
              <CardDescription>
                {t('dashboard.attentionTitle')}:{' '}
                {t('dashboard.attentionDescription')}
              </CardDescription>
            </CardHeader>
            <CardContent className='grid gap-3 md:grid-cols-2 xl:grid-cols-3'>
              {loading ? (
                <p className='text-sm text-muted-foreground'>
                  {t('dashboard.loadingActivity')}
                </p>
              ) : summary?.cards?.length ? (
                summary.cards.map((card) => (
                  <div
                    key={card.title}
                    className='flex items-center justify-between gap-3 rounded-md border p-3'
                  >
                    <div>
                      <div className='text-sm font-medium'>{card.title}</div>
                      <div className='text-xs text-muted-foreground'>
                        {card.value} {t('dashboard.pending')}
                      </div>
                    </div>
                    <Button asChild variant='outline' size='sm'>
                      <a
                        href={normalizeDashboardLink(card.link)}
                        data-testid={`dashboard-card-link-${card.title.toLowerCase().replace(/[^a-z0-9]+/g, '-')}`}
                      >
                        {t('dashboard.open')}
                      </a>
                    </Button>
                  </div>
                ))
              ) : (
                <p className='rounded-md border border-dashed p-3 text-sm text-muted-foreground'>
                  {t('dashboard.noActionItems')}
                </p>
              )}
            </CardContent>
          </Card>
        </div>
      </Main>
    </>
  )
}
