import { useCallback, useEffect, useState } from 'react'
import { Telescope } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { ConfigDrawer } from '@/components/config-drawer'
import { LanguageSwitch } from '@/components/language-switch'
import { Header, HeaderTitle } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'

type DiscoveryItem = {
  candidate_id: string
  title: string
  price?: number
  currency?: string
  url?: string
  last_seen: string
  first_seen?: string
  stock_state: string
  stock_count: number
  source?: string
  provider?: string
  source_provider?: string
  source_result_url?: string
  source_result_link?: string
  source_result_id?: string
  listing_id?: string
  triage_status?: string
  destination_status?: string
  confidence?: number
  needs_review?: boolean
  review_signal?: string
  seller_label?: string
  source_label?: string
}

type DiscoveryActionType =
  | 'ignore'
  | 'add_to_wishlist'
  | 'track_price'
  | 'create_item'

export function Discover() {
  const [query, setQuery] = useState('')
  const [priceMax, setPriceMax] = useState('')
  const [dateFrom, setDateFrom] = useState('')
  const [items, setItems] = useState<DiscoveryItem[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [actionStatus, setActionStatus] = useState<string | null>(null)

  const loadItems = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const params = new URLSearchParams()
      if (query.trim()) {
        params.set('q', query.trim())
      }
      if (priceMax.trim()) {
        params.set('price_max', priceMax.trim())
      }
      if (dateFrom.trim()) {
        params.set('date_from', dateFrom.trim())
      }
      const suffix = params.toString() ? `?${params.toString()}` : ''
      const response = await fetch(`/api/discovery/not-in-collection${suffix}`)
      if (!response.ok) {
        throw new Error(`discover_list_${response.status}`)
      }
      const payload = (await response.json()) as { items?: DiscoveryItem[] }
      setItems(payload.items ?? [])
    } catch (err) {
      const message =
        err instanceof Error ? err.message : 'discover_list_failed'
      setError(message)
      setItems([])
    } finally {
      setLoading(false)
    }
  }, [dateFrom, priceMax, query])

  useEffect(() => {
    void loadItems()
  }, [loadItems])

  const applyAction = async (
    candidateID: string,
    type: DiscoveryActionType
  ) => {
    setActionStatus(null)
    const response = await fetch('/api/discovery/action', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        candidate_id: candidateID,
        type,
        payload: {},
      }),
    })
    if (!response.ok) {
      setActionStatus(`discover_action_${response.status}`)
      return
    }
    setActionStatus(`${type}:${candidateID}`)
    await loadItems()
  }

  const formatMoney = (item: DiscoveryItem) => {
    if (typeof item.price !== 'number') {
      return 'Price pending'
    }
    const currency = item.currency?.trim() || 'USD'
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency,
    }).format(item.price)
  }

  const formatDate = (value?: string) => {
    if (!value) {
      return 'Recency pending'
    }
    const parsed = new Date(value)
    if (Number.isNaN(parsed.getTime())) {
      return value
    }
    return parsed.toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    })
  }

  const sourceLabel = (item: DiscoveryItem) =>
    item.source_provider?.trim() ||
    item.provider?.trim() ||
    item.source?.trim() ||
    item.source_label?.trim() ||
    'Source pending'

  const sourceResultURL = (item: DiscoveryItem) =>
    item.source_result_url?.trim() ||
    item.source_result_link?.trim() ||
    item.url?.trim() ||
    ''

  const reviewSignal = (item: DiscoveryItem) => {
    if (item.review_signal?.trim()) {
      return item.review_signal
    }
    if (item.needs_review) {
      return 'Needs review'
    }
    if (typeof item.confidence === 'number') {
      return `Confidence ${Math.round(item.confidence * 100)}%`
    }
    return 'Review ready'
  }

  return (
    <>
      <Header fixed>
        <Search />
        <HeaderTitle
          title='Discoveries'
          description='Not-in-collection candidates with triage actions.'
          icon={Telescope}
          testId='discoveries-header-title'
          iconTestId='discoveries-page-icon'
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

      <Main className='space-y-4'>
        <div>
          <h1 className='text-2xl font-bold tracking-tight'>Discoveries</h1>
          <p className='text-muted-foreground'>
            Pending found-item triage for candidates Cabinet found outside your
            owned Inventory, wanted Wishlist records, and Market Watch query
            history.
          </p>
        </div>

        <section
          className='rounded-md border border-dashed p-3 text-sm'
          data-testid='discover-candidate-inbox-purpose'
        >
          <p className='font-medium'>Candidate inbox</p>
          <p className='text-muted-foreground'>
            Review found items, inspect provenance, then decide whether each
            candidate belongs on Wishlist, needs Inventory or Purchase
            follow-up, or should be ignored or archived.
          </p>
        </section>

        <section
          className='rounded-md border border-dashed p-3 text-sm'
          data-testid='discover-market-watch-handoff'
        >
          <p className='font-medium'>Need provider query controls?</p>
          <p className='text-muted-foreground'>
            Query creation and run execution are available in Market Watch.
          </p>
          <Button
            variant='outline'
            className='mt-2'
            data-testid='discover-open-market-watch'
            onClick={() => {
              const suffix = query.trim()
                ? `?from=discoveries&q=${encodeURIComponent(query.trim())}`
                : '?from=discoveries'
              window.location.assign(`/scanner/${suffix}`)
            }}
          >
            Open Market Watch
          </Button>
        </section>

        <section className='grid gap-2 md:grid-cols-4'>
          <Input
            data-testid='discover-filter-query'
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder='Filter title'
          />
          <Input
            data-testid='discover-filter-price'
            value={priceMax}
            onChange={(event) => setPriceMax(event.target.value)}
            placeholder='Max price'
          />
          <Input
            data-testid='discover-filter-date'
            value={dateFrom}
            onChange={(event) => setDateFrom(event.target.value)}
            placeholder='Date from (YYYY-MM-DD)'
          />
          <Button
            data-testid='discover-apply-filters'
            onClick={() => void loadItems()}
          >
            Apply Filters
          </Button>
        </section>

        {error ? (
          <div
            data-testid='discover-error-state'
            className='rounded-md border border-destructive/50 bg-destructive/10 p-3 text-sm'
          >
            <p className='font-medium'>Discoveries are unavailable.</p>
            <p className='text-muted-foreground'>{error}</p>
            <Button
              className='mt-2'
              size='sm'
              variant='outline'
              onClick={() => void loadItems()}
            >
              Retry
            </Button>
          </div>
        ) : null}

        {actionStatus ? (
          <p data-testid='discover-action-status' className='text-sm'>
            {actionStatus}
          </p>
        ) : null}

        <div data-testid='discover-list' className='rounded-md border'>
          {loading ? (
            <p className='p-4 text-sm text-muted-foreground'>
              Loading discoveries...
            </p>
          ) : items.length === 0 ? (
            <div className='space-y-1 p-4 text-sm'>
              <p className='font-medium'>No pending found-item candidates.</p>
              <p className='text-muted-foreground'>
                Discoveries stays empty until provider runs or imports find
                items that need triage separate from Inventory, Wishlist, and
                Market Watch query history.
              </p>
            </div>
          ) : (
            <div className='divide-y'>
              {items.map((item) => (
                <div
                  key={item.candidate_id}
                  className='space-y-3 p-3'
                  data-testid={`discover-candidate-row-${item.candidate_id}`}
                >
                  <div className='flex items-start justify-between gap-3'>
                    <div className='space-y-1'>
                      <p className='font-medium'>{item.title}</p>
                      <p className='text-xs text-muted-foreground'>
                        {item.candidate_id} | {formatMoney(item)} | Stock{' '}
                        {item.stock_count}
                      </p>
                    </div>
                    {sourceResultURL(item) ? (
                      <a
                        href={sourceResultURL(item)}
                        className='text-sm underline'
                        target='_blank'
                        rel='noreferrer'
                        data-testid={`discover-source-result-${item.candidate_id}`}
                      >
                        Review source result
                      </a>
                    ) : null}
                  </div>
                  <dl
                    className='grid gap-2 text-xs text-muted-foreground sm:grid-cols-2 lg:grid-cols-4'
                    data-testid={`discover-provenance-${item.candidate_id}`}
                  >
                    <div>
                      <dt className='font-medium text-foreground'>Source</dt>
                      <dd>{sourceLabel(item)}</dd>
                    </div>
                    <div>
                      <dt className='font-medium text-foreground'>Source ID</dt>
                      <dd>
                        {item.source_result_id?.trim() ||
                          item.listing_id?.trim() ||
                          item.candidate_id}
                      </dd>
                    </div>
                    <div>
                      <dt className='font-medium text-foreground'>Recency</dt>
                      <dd>
                        First seen {formatDate(item.first_seen)}; last seen{' '}
                        {formatDate(item.last_seen)}
                      </dd>
                    </div>
                    <div>
                      <dt className='font-medium text-foreground'>Status</dt>
                      <dd>
                        {item.triage_status?.trim() ||
                          item.destination_status?.trim() ||
                          item.stock_state}
                        {' - '}
                        {reviewSignal(item)}
                      </dd>
                    </div>
                    <div>
                      <dt className='font-medium text-foreground'>
                        Seller/source
                      </dt>
                      <dd>
                        {item.seller_label?.trim() ||
                          item.source_label?.trim() ||
                          'Seller pending'}
                      </dd>
                    </div>
                  </dl>
                  <div className='flex flex-wrap gap-2'>
                    {sourceResultURL(item) ? (
                      <Button size='sm' variant='outline' asChild>
                        <a
                          href={sourceResultURL(item)}
                          target='_blank'
                          rel='noreferrer'
                          data-testid={`discover-action-review-source-${item.candidate_id}`}
                        >
                          Review Source
                        </a>
                      </Button>
                    ) : null}
                    <Button
                      size='sm'
                      variant='outline'
                      data-testid={`discover-action-ignore-${item.candidate_id}`}
                      onClick={() =>
                        void applyAction(item.candidate_id, 'ignore')
                      }
                    >
                      Ignore / Archive
                    </Button>
                    <Button
                      size='sm'
                      variant='outline'
                      data-testid={`discover-action-wishlist-${item.candidate_id}`}
                      onClick={() =>
                        void applyAction(item.candidate_id, 'add_to_wishlist')
                      }
                    >
                      Promote to Wishlist
                    </Button>
                    <Button
                      size='sm'
                      variant='outline'
                      data-testid={`discover-action-track-${item.candidate_id}`}
                      onClick={() =>
                        void applyAction(item.candidate_id, 'track_price')
                      }
                    >
                      Purchase Follow-up
                    </Button>
                    <Button
                      size='sm'
                      data-testid={`discover-action-create-${item.candidate_id}`}
                      onClick={() =>
                        void applyAction(item.candidate_id, 'create_item')
                      }
                    >
                      Inventory Handoff
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </Main>
    </>
  )
}
