import { useEffect, useMemo, useState } from 'react'
import { useNavigate } from '@tanstack/react-router'
import {
  ArrowRight,
  Barcode,
  LoaderCircle,
  PackageSearch,
  ScanSearch,
  SearchIcon,
} from 'lucide-react'
import { Input } from '@/components/ui/input'

type LocalSearchItem = {
  id: string
  part_number: string
  title: string
  brand: string
  category: string
}

type BarcodeMatch = {
  id: string
  item_id: string
  barcode: string
}

function normalizeLocalSearchItem(
  item: Partial<LocalSearchItem>
): LocalSearchItem {
  return {
    id: item.id?.trim() ?? '',
    part_number: item.part_number?.trim() ?? '',
    title: item.title?.trim() || 'Untitled item',
    brand: item.brand?.trim() || 'Unknown brand',
    category: item.category?.trim() || 'General',
  }
}

function isBarcodeCandidate(value: string) {
  return /^[0-9]{6,}$/.test(value.trim())
}

export function SearchWorkspacePanel() {
  const navigate = useNavigate()
  const [searchValue, setSearchValue] = useState('')
  const [localResults, setLocalResults] = useState<LocalSearchItem[]>([])
  const [localSearchState, setLocalSearchState] = useState<
    'idle' | 'loading' | 'success' | 'error'
  >('idle')
  const [barcodeMatches, setBarcodeMatches] = useState<BarcodeMatch[]>([])
  const [barcodeState, setBarcodeState] = useState<
    'idle' | 'loading' | 'success' | 'error'
  >('idle')
  const trimmedSearch = searchValue.trim()
  const barcodeCandidate = useMemo(
    () => isBarcodeCandidate(trimmedSearch),
    [trimmedSearch]
  )

  const handleSearchChange = (value: string) => {
    setSearchValue(value)
    const nextValue = value.trim()
    if (nextValue.length < 2) {
      setLocalResults([])
      setLocalSearchState('idle')
    }
    if (!isBarcodeCandidate(nextValue)) {
      setBarcodeMatches([])
      setBarcodeState('idle')
    }
  }

  useEffect(() => {
    if (trimmedSearch.length < 2) {
      return
    }

    const controller = new AbortController()
    const timeoutID = window.setTimeout(() => {
      setLocalSearchState('loading')
      fetch(
        `/api/search/items?text=${encodeURIComponent(trimmedSearch)}&limit=5`,
        { signal: controller.signal }
      )
        .then(async (response) => {
          if (!response.ok) {
            throw new Error(`search_items_${response.status}`)
          }
          const payload = (await response.json()) as {
            items?: Array<Partial<LocalSearchItem>>
          }
          setLocalResults(
            (payload.items ?? [])
              .map(normalizeLocalSearchItem)
              .filter((item) => item.id !== '')
          )
          setLocalSearchState('success')
        })
        .catch((error: unknown) => {
          if (error instanceof DOMException && error.name === 'AbortError') {
            return
          }
          setLocalResults([])
          setLocalSearchState('error')
        })
    }, 150)

    return () => {
      window.clearTimeout(timeoutID)
      controller.abort()
    }
  }, [trimmedSearch])

  useEffect(() => {
    if (!barcodeCandidate) {
      return
    }

    const controller = new AbortController()
    const timeoutID = window.setTimeout(() => {
      setBarcodeState('loading')
      fetch(`/api/barcodes/${encodeURIComponent(trimmedSearch)}`, {
        signal: controller.signal,
      })
        .then(async (response) => {
          if (!response.ok) {
            throw new Error(`barcode_lookup_${response.status}`)
          }
          const payload = (await response.json()) as {
            matches?: Array<Partial<BarcodeMatch>>
          }
          setBarcodeMatches(
            (payload.matches ?? [])
              .map((match) => ({
                id: match.id?.trim() ?? '',
                item_id: match.item_id?.trim() ?? '',
                barcode: match.barcode?.trim() ?? trimmedSearch,
              }))
              .filter((match) => match.id !== '' && match.item_id !== '')
          )
          setBarcodeState('success')
        })
        .catch((error: unknown) => {
          if (error instanceof DOMException && error.name === 'AbortError') {
            return
          }
          setBarcodeMatches([])
          setBarcodeState('error')
        })
    }, 150)

    return () => {
      window.clearTimeout(timeoutID)
      controller.abort()
    }
  }, [barcodeCandidate, trimmedSearch])

  const openInventoryFilter = (filter: string) => {
    void navigate({
      to: '/inventory',
      search: { filter } as never,
    })
  }

  const openMarketWatch = () => {
    void navigate({
      to: '/scanner',
      search: { barcode: trimmedSearch } as never,
    })
  }

  return (
    <aside
      className='flex min-h-full flex-col gap-3 px-2 py-2'
      data-testid='shell-search-workspace'
    >
      <div className='space-y-1'>
        <h2 className='text-sm font-semibold text-sidebar-foreground'>
          Search
        </h2>
        <p className='text-xs leading-5 text-muted-foreground'>
          Find inventory, scan barcode evidence, or jump to a workspace.
        </p>
      </div>
      <div className='relative'>
        <SearchIcon
          aria-hidden
          className='pointer-events-none absolute top-1/2 left-2.5 h-4 w-4 -translate-y-1/2 text-muted-foreground'
        />
        <Input
          autoFocus
          aria-label='Search Cabinet workspace'
          className='h-9 rounded-md border-sidebar-border bg-sidebar-accent/30 pl-8 text-sm'
          data-testid='shell-search-workspace-input'
          onChange={(event) => handleSearchChange(event.target.value)}
          placeholder='Search Cabinet'
          value={searchValue}
        />
      </div>
      <div className='min-h-0 flex-1 space-y-3 overflow-y-auto pr-1'>
        <section
          className='rounded-md border border-sidebar-border bg-sidebar-accent/20'
          data-testid='shell-search-local-results'
        >
          <div className='flex items-center gap-2 border-b border-sidebar-border px-2 py-2 text-xs font-medium text-muted-foreground'>
            <PackageSearch className='h-4 w-4' aria-hidden />
            <span>Local catalog</span>
            {localSearchState === 'loading' ? (
              <LoaderCircle
                className='ml-auto h-3.5 w-3.5 animate-spin'
                aria-hidden
              />
            ) : null}
          </div>
          {trimmedSearch.length < 2 ? (
            <p className='px-2 py-3 text-xs leading-5 text-muted-foreground'>
              Enter at least two characters.
            </p>
          ) : localSearchState === 'error' ? (
            <p className='px-2 py-3 text-xs leading-5 text-destructive'>
              Local catalog search is unavailable.
            </p>
          ) : localSearchState === 'success' && localResults.length === 0 ? (
            <p className='px-2 py-3 text-xs leading-5 text-muted-foreground'>
              No local inventory matches.
            </p>
          ) : (
            <div className='divide-y divide-sidebar-border/70'>
              {localResults.map((item) => (
                <button
                  key={item.id}
                  type='button'
                  className='flex w-full items-center gap-2 px-2 py-2 text-left hover:bg-sidebar-accent focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none'
                  data-testid='shell-search-local-result'
                  onClick={() =>
                    openInventoryFilter(item.part_number || item.title)
                  }
                >
                  <span className='min-w-0 flex-1'>
                    <span className='block truncate text-xs font-medium text-sidebar-foreground'>
                      {item.title}
                    </span>
                    <span className='block truncate text-[11px] text-muted-foreground'>
                      {[item.part_number, item.brand, item.category]
                        .filter(Boolean)
                        .join(' - ')}
                    </span>
                  </span>
                  <ArrowRight className='h-3.5 w-3.5 text-muted-foreground' />
                </button>
              ))}
            </div>
          )}
        </section>
        <section
          className='rounded-md border border-sidebar-border bg-sidebar-accent/20'
          data-testid='shell-search-barcode-results'
        >
          <div className='flex items-center gap-2 border-b border-sidebar-border px-2 py-2 text-xs font-medium text-muted-foreground'>
            <Barcode className='h-4 w-4' aria-hidden />
            <span>Barcode lookup</span>
            {barcodeState === 'loading' ? (
              <LoaderCircle
                className='ml-auto h-3.5 w-3.5 animate-spin'
                aria-hidden
              />
            ) : null}
          </div>
          {!barcodeCandidate ? (
            <p className='px-2 py-3 text-xs leading-5 text-muted-foreground'>
              Enter a numeric barcode to check matches.
            </p>
          ) : barcodeState === 'error' ? (
            <p className='px-2 py-3 text-xs leading-5 text-destructive'>
              Barcode lookup is unavailable.
            </p>
          ) : barcodeState === 'success' && barcodeMatches.length === 0 ? (
            <button
              type='button'
              className='flex w-full items-center gap-2 px-2 py-3 text-left text-xs text-sidebar-foreground hover:bg-sidebar-accent focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none'
              data-testid='shell-search-open-market-watch'
              onClick={openMarketWatch}
            >
              <ScanSearch className='h-4 w-4 text-muted-foreground' />
              <span className='min-w-0 flex-1'>
                Open Market Watch for {trimmedSearch}
              </span>
              <ArrowRight className='h-3.5 w-3.5 text-muted-foreground' />
            </button>
          ) : (
            <div className='divide-y divide-sidebar-border/70'>
              {barcodeMatches.map((match) => (
                <button
                  key={match.id}
                  type='button'
                  className='flex w-full items-center gap-2 px-2 py-2 text-left hover:bg-sidebar-accent focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none'
                  data-testid='shell-search-barcode-result'
                  onClick={() => openInventoryFilter(match.item_id)}
                >
                  <span className='min-w-0 flex-1'>
                    <span className='block truncate text-xs font-medium text-sidebar-foreground'>
                      {match.barcode}
                    </span>
                    <span className='block truncate text-[11px] text-muted-foreground'>
                      Inventory item {match.item_id}
                    </span>
                  </span>
                  <ArrowRight className='h-3.5 w-3.5 text-muted-foreground' />
                </button>
              ))}
            </div>
          )}
        </section>
      </div>
    </aside>
  )
}
