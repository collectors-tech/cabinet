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

export function Dashboard() {
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
      { title: 'Inventory Items', value: `${summary.total_items}` },
      { title: 'Inventory Units', value: `${summary.total_instances}` },
      { title: 'Wishlist Hits', value: `${summary.wishlist_hits}` },
      { title: 'Estimated Value', value: formatCurrency(summary.estimated_value) },
    ]
  }, [summary])

  return (
    <>
      <Header fixed>
        <Search />
        <div className='ms-auto flex items-center space-x-4'>
          <ThemeSwitch />
          <ConfigDrawer />
          <ProfileDropdown />
        </div>
      </Header>

      <Main className='space-y-6'>
        <div className='flex flex-wrap items-center justify-between gap-3'>
          <div>
            <h1 className='text-2xl font-bold tracking-tight'>Home</h1>
            <p className='text-muted-foreground'>
              What needs action now in your collection.
            </p>
          </div>
          <Button onClick={() => void loadDashboard()} disabled={loading}>
            {loading ? 'Refreshing...' : 'Refresh Dashboard'}
          </Button>
        </div>

        {error ? (
          <Card>
            <CardHeader>
              <CardTitle>Dashboard unavailable</CardTitle>
              <CardDescription>{error}</CardDescription>
            </CardHeader>
            <CardContent>
              <Button variant='outline' onClick={() => void loadDashboard()}>
                Retry
              </Button>
            </CardContent>
          </Card>
        ) : null}

        <div className='grid gap-4 sm:grid-cols-2 lg:grid-cols-4'>
          {loading
            ? [1, 2, 3, 4].map((slot) => (
                <Card key={slot}>
                  <CardHeader className='pb-2'>
                    <CardTitle className='text-sm font-medium'>
                      Loading...
                    </CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className='text-2xl font-bold'>--</div>
                  </CardContent>
                </Card>
              ))
            : metricCards.map((card) => (
                <Card key={card.title}>
                  <CardHeader className='pb-2'>
                    <CardTitle className='text-sm font-medium'>
                      {card.title}
                    </CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className='text-2xl font-bold'>{card.value}</div>
                  </CardContent>
                </Card>
              ))}
        </div>

        <div className='grid grid-cols-1 gap-4 lg:grid-cols-7'>
          <Card className='col-span-1 lg:col-span-4'>
            <CardHeader>
              <CardTitle>What needs attention now</CardTitle>
              <CardDescription>
                New discoveries, watch-list hits, and price movement.
              </CardDescription>
            </CardHeader>
            <CardContent className='space-y-3'>
              {loading ? (
                <p className='text-sm text-muted-foreground'>
                  Loading activity signals...
                </p>
              ) : summary?.cards?.length ? (
                summary.cards.map((card) => (
                  <div
                    key={card.title}
                    className='flex items-center justify-between rounded-md border p-3'
                  >
                    <div>
                      <div className='text-sm font-medium'>{card.title}</div>
                      <div className='text-xs text-muted-foreground'>
                        {card.value} pending
                      </div>
                    </div>
                    <Button asChild variant='outline' size='sm'>
                      <a href={card.link}>Open</a>
                    </Button>
                  </div>
                ))
              ) : (
                <p className='text-sm text-muted-foreground'>
                  No action items right now.
                </p>
              )}
            </CardContent>
          </Card>

          <Card className='col-span-1 lg:col-span-3'>
            <CardHeader>
              <CardTitle>Recently added</CardTitle>
              <CardDescription>Latest items added to inventory.</CardDescription>
            </CardHeader>
            <CardContent>
              {loading ? (
                <p className='text-sm text-muted-foreground'>Loading items...</p>
              ) : summary?.recently_added?.length ? (
                <ul className='space-y-2'>
                  {summary.recently_added.map((item) => (
                    <li
                      key={item}
                      className='truncate rounded-md border px-3 py-2 text-sm'
                    >
                      {item}
                    </li>
                  ))}
                </ul>
              ) : (
                <p className='text-sm text-muted-foreground'>
                  No recently added items yet.
                </p>
              )}
            </CardContent>
          </Card>
        </div>
      </Main>
    </>
  )
}
