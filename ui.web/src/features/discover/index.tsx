import { useCallback, useEffect, useState } from 'react'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { Search } from '@/components/search'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { LanguageSwitch } from '@/components/language-switch'
import { ThemeSwitch } from '@/components/theme-switch'
import { ConfigDrawer } from '@/components/config-drawer'
import { ProfileDropdown } from '@/components/profile-dropdown'

type DiscoveryItem = {
  candidate_id: string
  title: string
  price: number
  url: string
  last_seen: string
  stock_state: string
  stock_count: number
}

type DiscoveryActionType = 'ignore' | 'add_to_wishlist' | 'track_price' | 'create_item'

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
      const message = err instanceof Error ? err.message : 'discover_list_failed'
      setError(message)
      setItems([])
    } finally {
      setLoading(false)
    }
  }, [dateFrom, priceMax, query])

  useEffect(() => {
    void loadItems()
  }, [loadItems])

  const applyAction = async (candidateID: string, type: DiscoveryActionType) => {
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

      <Main className='space-y-4'>
        <div>
          <h1 className='text-2xl font-bold tracking-tight'>Discoveries</h1>
          <p className='text-muted-foreground'>
            Not-in-collection candidates with triage actions.
          </p>
        </div>

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
          <Button data-testid='discover-apply-filters' onClick={() => void loadItems()}>
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
            <p className='p-4 text-sm text-muted-foreground'>Loading discoveries...</p>
          ) : items.length === 0 ? (
            <p className='p-4 text-sm text-muted-foreground'>
              No discovery candidates match current filters.
            </p>
          ) : (
            <div className='divide-y'>
              {items.map((item) => (
                <div key={item.candidate_id} className='space-y-2 p-3'>
                  <div className='flex items-start justify-between gap-3'>
                    <div>
                      <p className='font-medium'>{item.title}</p>
                      <p className='text-xs text-muted-foreground'>
                        {item.candidate_id} | ${item.price.toFixed(2)} | Stock {item.stock_count}
                      </p>
                    </div>
                    <a href={item.url} className='text-sm underline' target='_blank' rel='noreferrer'>
                      Open Listing
                    </a>
                  </div>
                  <div className='flex flex-wrap gap-2'>
                    <Button
                      size='sm'
                      variant='outline'
                      data-testid={`discover-action-ignore-${item.candidate_id}`}
                      onClick={() => void applyAction(item.candidate_id, 'ignore')}
                    >
                      Ignore
                    </Button>
                    <Button
                      size='sm'
                      variant='outline'
                      data-testid={`discover-action-wishlist-${item.candidate_id}`}
                      onClick={() => void applyAction(item.candidate_id, 'add_to_wishlist')}
                    >
                      Add to Wishlist
                    </Button>
                    <Button
                      size='sm'
                      variant='outline'
                      data-testid={`discover-action-track-${item.candidate_id}`}
                      onClick={() => void applyAction(item.candidate_id, 'track_price')}
                    >
                      Track Price
                    </Button>
                    <Button
                      size='sm'
                      data-testid={`discover-action-create-${item.candidate_id}`}
                      onClick={() => void applyAction(item.candidate_id, 'create_item')}
                    >
                      Create Item
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
