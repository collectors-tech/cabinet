import { useCallback, useEffect, useState } from 'react'
import {
  Archive,
  ArrowUpDown,
  ExternalLink,
  Heart,
  PackagePlus,
  RotateCcw,
  ShoppingCart,
  Telescope,
} from 'lucide-react'
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
  query_set_id?: string
  query_name?: string
  triage_status?: string
  destination_status?: string
  confidence?: number
  needs_review?: boolean
  review_signal?: string
  seller_label?: string
  source_label?: string
  match_type?: string
  match_reason?: string
  wishlist_id?: string
  wishlist_item_id?: string
  target_price?: number
  observed_price_baseline?: number
  market_price_baseline?: number
  price_delta_amount?: number
  price_delta_percent?: number
  deal_score?: number
  source_trust_status?: string
  thumbnail_url?: string
  destination_link?: string
  availability?: string
}

type DiscoveryActionType =
  | 'ignore'
  | 'add_to_wishlist'
  | 'track_price'
  | 'create_item'
  | 'review'
  | 'archive'

type DiscoveryTab =
  | 'all'
  | 'wishlist'
  | 'deals'
  | 'market_watch'
  | 'stores'
  | 'other'
  | 'archived'

type DiscoverySort = 'ranked' | 'deal' | 'last_seen'

export function Discover() {
  const [query, setQuery] = useState('')
  const [priceMax, setPriceMax] = useState('')
  const [dateFrom, setDateFrom] = useState('')
  const [activeTab, setActiveTab] = useState<DiscoveryTab>('all')
  const [sortMode, setSortMode] = useState<DiscoverySort>('ranked')
  const [items, setItems] = useState<DiscoveryItem[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [actionStatus, setActionStatus] = useState<string | null>(null)

  const loadItems = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const params = new URLSearchParams()
      params.set('include_archived', 'true')
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
    return formatMoneyValue(item.price, item.currency)
  }

  const formatMoneyValue = (value?: number, currency = 'USD') => {
    if (typeof value !== 'number') {
      return 'Price pending'
    }
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: currency?.trim() || 'USD',
    }).format(value)
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

  const statusLabel = (item: DiscoveryItem) =>
    item.triage_status?.trim() ||
    item.destination_status?.trim() ||
    item.stock_state ||
    'new'

  const isArchived = (item: DiscoveryItem) => {
    const status = statusLabel(item).toLowerCase()
    return status.includes('ignored') || status.includes('archived')
  }

  const isPromotedStatus = (item: DiscoveryItem) => {
    const status = statusLabel(item).toLowerCase()
    return (
      Boolean(item.destination_link) ||
      status.includes('wishlisted') ||
      status.includes('purchase_candidate') ||
      status.includes('inventory_candidate')
    )
  }

  const isWishlistMatch = (item: DiscoveryItem) =>
    item.match_type === 'wishlist_match' ||
    Boolean(item.wishlist_id || item.wishlist_item_id) ||
    item.match_reason?.toLowerCase().includes('wishlist')

  const isMarketWatch = (item: DiscoveryItem) =>
    item.match_type === 'market_watch_result' ||
    Boolean(item.query_set_id || item.query_name) ||
    sourceLabel(item).toLowerCase().includes('market')

  const isStoreOrProvider = (item: DiscoveryItem) => {
    const type = item.match_type ?? ''
    const label = sourceLabel(item).toLowerCase()
    return (
      type === 'store_stock' ||
      type === 'provider_search' ||
      (!isMarketWatch(item) &&
        !isWishlistMatch(item) &&
        (label.includes('store') || label.includes('provider')))
    )
  }

  const dealDeltaPercent = (item: DiscoveryItem) => {
    if (typeof item.price_delta_percent === 'number') {
      return item.price_delta_percent
    }
    const baseline = item.target_price || item.market_price_baseline
    if (
      typeof item.price === 'number' &&
      typeof baseline === 'number' &&
      baseline > 0
    ) {
      return Math.round(((baseline - item.price) / baseline) * 100)
    }
    return undefined
  }

  const isGreatDeal = (item: DiscoveryItem) => {
    const delta = dealDeltaPercent(item)
    return (
      (typeof item.deal_score === 'number' && item.deal_score >= 70) ||
      (typeof delta === 'number' && delta > 0) ||
      (typeof item.target_price === 'number' &&
        typeof item.price === 'number' &&
        item.price <= item.target_price)
    )
  }

  const matchReason = (item: DiscoveryItem) => {
    if (item.match_reason?.trim()) {
      return item.match_reason.trim()
    }
    if (isWishlistMatch(item)) {
      return 'Wishlist match'
    }
    if (isGreatDeal(item)) {
      return 'Below target'
    }
    if (isMarketWatch(item)) {
      return 'New Market Watch result'
    }
    if (isStoreOrProvider(item)) {
      return 'Store stock found'
    }
    return 'Found candidate'
  }

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

  const dealScore = (item: DiscoveryItem) =>
    (isArchived(item) ? -1000 : 0) +
    (isWishlistMatch(item) ? 500 : 0) +
    (isGreatDeal(item) ? 300 : 0) +
    (item.deal_score ?? 0) +
    (item.needs_review ? 10 : 0)

  const lastSeenTime = (item: DiscoveryItem) => {
    const parsed = new Date(item.last_seen)
    return Number.isNaN(parsed.getTime()) ? 0 : parsed.getTime()
  }

  const sortedItems = [...items].sort((a, b) => {
    if (sortMode === 'deal') {
      return dealScore(b) - dealScore(a)
    }
    if (sortMode === 'last_seen') {
      return lastSeenTime(b) - lastSeenTime(a)
    }
    return dealScore(b) - dealScore(a)
  })

  const visibleItems = sortedItems.filter((item) => {
    switch (activeTab) {
      case 'wishlist':
        return isWishlistMatch(item) && !isArchived(item)
      case 'deals':
        return isGreatDeal(item) && !isArchived(item)
      case 'market_watch':
        return isMarketWatch(item) && !isArchived(item)
      case 'stores':
        return isStoreOrProvider(item) && !isArchived(item)
      case 'other':
        return (
          (item.match_type === 'public_binder_match' ||
            item.match_type === 'peer_session_match') &&
          !isArchived(item)
        )
      case 'archived':
        return isArchived(item)
      default:
        return !isArchived(item)
    }
  })

  const summary = {
    deals: items.filter((item) => isGreatDeal(item) && !isArchived(item))
      .length,
    wishlist: items.filter((item) => isWishlistMatch(item) && !isArchived(item))
      .length,
    newFinds: items.filter((item) => statusLabel(item).toLowerCase() === 'new')
      .length,
    marketWatch: items.filter(
      (item) => isMarketWatch(item) && !isArchived(item)
    ).length,
    attention: items.filter(
      (item) =>
        !isArchived(item) &&
        (item.needs_review ||
          item.source_trust_status?.toLowerCase().includes('attention') ||
          item.source_trust_status?.toLowerCase().includes('auth') ||
          item.stock_state?.toLowerCase().includes('unknown'))
    ).length,
  }

  const tabCounts: Record<DiscoveryTab, number> = {
    all: items.filter((item) => !isArchived(item)).length,
    wishlist: summary.wishlist,
    deals: summary.deals,
    market_watch: summary.marketWatch,
    stores: items.filter((item) => isStoreOrProvider(item) && !isArchived(item))
      .length,
    other: items.filter(
      (item) =>
        (item.match_type === 'public_binder_match' ||
          item.match_type === 'peer_session_match') &&
        !isArchived(item)
    ).length,
    archived: items.filter(isArchived).length,
  }

  const tabs: Array<{ id: DiscoveryTab; label: string }> = [
    { id: 'all', label: 'All discoveries' },
    { id: 'wishlist', label: 'Wishlist matches' },
    { id: 'deals', label: 'Great prices' },
    { id: 'market_watch', label: 'Market Watch' },
    { id: 'stores', label: 'Stores/providers' },
    { id: 'other', label: 'Other inventories' },
    { id: 'archived', label: 'Ignored / archived' },
  ]

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
            Found deals and source outputs worth reviewing from wishlist, Market
            Watch, stores, providers, and shared public discovery surfaces.
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
          className='grid gap-3 md:grid-cols-5'
          data-testid='discover-dashboard-summary'
        >
          {[
            ['Best deals found', summary.deals],
            ['Wishlist matches', summary.wishlist],
            ['New since last visit', summary.newFinds],
            ['Market Watch review', summary.marketWatch],
            ['Provider attention', summary.attention],
          ].map(([label, value]) => (
            <div
              key={label}
              className='rounded-md border bg-card p-3'
              data-testid={`discover-summary-${String(label)
                .toLowerCase()
                .replace(/\s+/g, '-')}`}
            >
              <p className='text-xs font-medium text-muted-foreground'>
                {label}
              </p>
              <p className='text-2xl font-semibold'>{value}</p>
            </div>
          ))}
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

        <section
          className='flex flex-wrap gap-2'
          aria-label='Discovery source filters'
          data-testid='discover-source-tabs'
        >
          {tabs.map((tab) => (
            <Button
              key={tab.id}
              type='button'
              size='sm'
              variant={activeTab === tab.id ? 'default' : 'outline'}
              data-testid={`discover-filter-tab-${tab.id}`}
              onClick={() => setActiveTab(tab.id)}
            >
              {tab.label} ({tabCounts[tab.id]})
            </Button>
          ))}
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

        <div
          data-testid='discover-list'
          className='overflow-hidden rounded-md border'
        >
          {loading ? (
            <p className='p-4 text-sm text-muted-foreground'>
              Loading discoveries... Checking provider signals.
            </p>
          ) : items.length === 0 ? (
            <div className='space-y-1 p-4 text-sm'>
              <p className='font-medium'>No pending found-item candidates.</p>
              <p className='text-muted-foreground'>
                Discoveries stays empty until provider runs or imports find
                items that need triage separate from Inventory, Wishlist, and
                Market Watch query history.
              </p>
              <p className='text-muted-foreground'>
                Check provider health, run Market Watch searches, or add
                wishlist targets to teach Cabinet which finds matter.
              </p>
            </div>
          ) : visibleItems.length === 0 ? (
            <div className='space-y-1 p-4 text-sm'>
              <p className='font-medium'>No discoveries match this filter.</p>
              <p className='text-muted-foreground'>
                Try All discoveries, Great prices, or Market Watch to review the
                current source outputs.
              </p>
            </div>
          ) : (
            <div className='overflow-x-auto'>
              <table className='w-full min-w-[980px] text-sm'>
                <thead className='bg-muted/50 text-left text-xs text-muted-foreground'>
                  <tr>
                    <th className='px-3 py-2 font-medium'>Discovery</th>
                    <th className='px-3 py-2 font-medium'>
                      <Button
                        type='button'
                        size='sm'
                        variant='ghost'
                        className='h-auto px-0 text-xs font-medium text-muted-foreground hover:bg-transparent'
                        data-testid='discover-sort-deal'
                        aria-pressed={sortMode === 'deal'}
                        onClick={() => setSortMode('deal')}
                      >
                        Deal
                        <ArrowUpDown className='ml-1 h-3 w-3' />
                      </Button>
                    </th>
                    <th className='px-3 py-2 font-medium'>Source</th>
                    <th className='px-3 py-2 font-medium'>Stock</th>
                    <th className='px-3 py-2 font-medium'>
                      <Button
                        type='button'
                        size='sm'
                        variant='ghost'
                        className='h-auto px-0 text-xs font-medium text-muted-foreground hover:bg-transparent'
                        data-testid='discover-sort-last-seen'
                        aria-pressed={sortMode === 'last_seen'}
                        onClick={() => setSortMode('last_seen')}
                      >
                        Seen
                        <ArrowUpDown className='ml-1 h-3 w-3' />
                      </Button>
                    </th>
                    <th className='px-3 py-2 font-medium'>Status</th>
                    <th className='px-3 py-2 font-medium'>Actions</th>
                  </tr>
                </thead>
                <tbody className='divide-y'>
                  {visibleItems.map((item) => {
                    const delta = dealDeltaPercent(item)
                    const alreadyPromoted = isPromotedStatus(item)
                    const archived = isArchived(item)
                    const canPromoteWishlist =
                      !archived && !alreadyPromoted && !isWishlistMatch(item)
                    const canPurchaseFollowUp = !archived && !alreadyPromoted
                    const canInventoryHandoff =
                      !archived && !alreadyPromoted && !isWishlistMatch(item)
                    return (
                      <tr
                        key={item.candidate_id}
                        data-testid={`discover-candidate-row-${item.candidate_id}`}
                        className={isWishlistMatch(item) ? 'bg-primary/5' : ''}
                      >
                        <td className='px-3 py-3 align-top'>
                          <div className='flex gap-3'>
                            {item.thumbnail_url ? (
                              <img
                                src={item.thumbnail_url}
                                alt=''
                                className='h-12 w-12 rounded object-cover'
                              />
                            ) : (
                              <div className='flex h-12 w-12 items-center justify-center rounded bg-muted text-muted-foreground'>
                                <Telescope className='h-5 w-5' />
                              </div>
                            )}
                            <div className='space-y-1'>
                              <p className='font-medium'>{item.title}</p>
                              <p className='text-xs text-muted-foreground'>
                                {matchReason(item)}
                              </p>
                              <p className='text-xs text-muted-foreground'>
                                {item.candidate_id}
                              </p>
                            </div>
                          </div>
                        </td>
                        <td className='px-3 py-3 align-top'>
                          <div className='space-y-1'>
                            <p className='font-medium'>{formatMoney(item)}</p>
                            <p className='text-xs text-muted-foreground'>
                              Target{' '}
                              {formatMoneyValue(
                                item.target_price,
                                item.currency
                              )}
                            </p>
                            <p className='text-xs text-muted-foreground'>
                              Baseline{' '}
                              {formatMoneyValue(
                                item.market_price_baseline ??
                                  item.observed_price_baseline,
                                item.currency
                              )}
                            </p>
                            {typeof delta === 'number' ? (
                              <p className='text-xs font-medium text-emerald-600'>
                                {delta}% saving
                              </p>
                            ) : null}
                          </div>
                        </td>
                        <td
                          className='px-3 py-3 align-top'
                          data-testid={`discover-provenance-${item.candidate_id}`}
                        >
                          <div className='space-y-1'>
                            <p>{sourceLabel(item)}</p>
                            <p className='text-xs text-muted-foreground'>
                              {item.query_name?.trim() ||
                                item.source_trust_status ||
                                'Source ready'}
                            </p>
                            <p className='text-xs text-muted-foreground'>
                              {item.source_result_id?.trim() ||
                                item.listing_id?.trim() ||
                                item.candidate_id}
                            </p>
                            <p className='text-xs text-muted-foreground'>
                              {item.seller_label?.trim() ||
                                item.source_label?.trim() ||
                                'Seller pending'}
                            </p>
                          </div>
                        </td>
                        <td className='px-3 py-3 align-top'>
                          <p>{item.availability || item.stock_state}</p>
                          <p className='text-xs text-muted-foreground'>
                            Stock {item.stock_count}
                          </p>
                        </td>
                        <td className='px-3 py-3 align-top text-xs text-muted-foreground'>
                          <p>First seen {formatDate(item.first_seen)}</p>
                          <p>Last seen {formatDate(item.last_seen)}</p>
                        </td>
                        <td className='px-3 py-3 align-top'>
                          <p>{statusLabel(item)}</p>
                          <p className='text-xs text-muted-foreground'>
                            {reviewSignal(item)}
                          </p>
                        </td>
                        <td className='px-3 py-3 align-top'>
                          {alreadyPromoted && item.destination_link ? (
                            <Button size='sm' variant='outline' asChild>
                              <a
                                href={item.destination_link}
                                data-testid={`discover-action-open-destination-${item.candidate_id}`}
                              >
                                Open destination
                              </a>
                            </Button>
                          ) : (
                            <div className='flex flex-wrap gap-2'>
                              {sourceResultURL(item) ? (
                                <Button size='icon' variant='outline' asChild>
                                  <a
                                    href={sourceResultURL(item)}
                                    target='_blank'
                                    rel='noreferrer'
                                    aria-label='Review source result'
                                    title='Review source result'
                                    data-testid={`discover-action-review-source-${item.candidate_id}`}
                                  >
                                    <ExternalLink className='h-4 w-4' />
                                  </a>
                                </Button>
                              ) : null}
                              {archived ? (
                                <Button
                                  size='icon'
                                  variant='outline'
                                  aria-label='Restore for review'
                                  title='Restore for review'
                                  data-testid={`discover-action-restore-${item.candidate_id}`}
                                  onClick={() =>
                                    void applyAction(
                                      item.candidate_id,
                                      'review'
                                    )
                                  }
                                >
                                  <RotateCcw className='h-4 w-4' />
                                </Button>
                              ) : null}
                              {!archived ? (
                                <Button
                                  size='icon'
                                  variant='outline'
                                  aria-label='Ignore or archive'
                                  title='Ignore or archive'
                                  data-testid={`discover-action-ignore-${item.candidate_id}`}
                                  onClick={() =>
                                    void applyAction(
                                      item.candidate_id,
                                      'ignore'
                                    )
                                  }
                                >
                                  <Archive className='h-4 w-4' />
                                </Button>
                              ) : null}
                              {canPromoteWishlist ? (
                                <Button
                                  size='icon'
                                  variant='outline'
                                  aria-label='Promote to Wishlist'
                                  title='Promote to Wishlist'
                                  data-testid={`discover-action-wishlist-${item.candidate_id}`}
                                  onClick={() =>
                                    void applyAction(
                                      item.candidate_id,
                                      'add_to_wishlist'
                                    )
                                  }
                                >
                                  <Heart className='h-4 w-4' />
                                </Button>
                              ) : null}
                              {canPurchaseFollowUp ? (
                                <Button
                                  size='icon'
                                  variant='outline'
                                  aria-label='Purchase follow-up'
                                  title='Purchase follow-up'
                                  data-testid={`discover-action-track-${item.candidate_id}`}
                                  onClick={() =>
                                    void applyAction(
                                      item.candidate_id,
                                      'track_price'
                                    )
                                  }
                                >
                                  <ShoppingCart className='h-4 w-4' />
                                </Button>
                              ) : null}
                              {canInventoryHandoff ? (
                                <Button
                                  size='icon'
                                  aria-label='Inventory handoff'
                                  title='Inventory handoff'
                                  data-testid={`discover-action-create-${item.candidate_id}`}
                                  onClick={() =>
                                    void applyAction(
                                      item.candidate_id,
                                      'create_item'
                                    )
                                  }
                                >
                                  <PackagePlus className='h-4 w-4' />
                                </Button>
                              ) : null}
                            </div>
                          )}
                        </td>
                      </tr>
                    )
                  })}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </Main>
    </>
  )
}
