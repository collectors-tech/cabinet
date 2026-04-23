import React, { useEffect, useMemo, useState } from 'react'
import { useNavigate } from '@tanstack/react-router'
import {
  ArrowRight,
  Barcode,
  ChevronRight,
  Laptop,
  LoaderCircle,
  Moon,
  PackageSearch,
  ScanSearch,
  Sun,
} from 'lucide-react'
import { useSearch } from '@/context/search-provider'
import { useTheme } from '@/context/theme-provider'
import {
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
} from '@/components/ui/command'
import { sidebarData } from './layout/data/sidebar-data'
import { ScrollArea } from './ui/scroll-area'

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

function normalizeLocalSearchItem(item: Partial<LocalSearchItem>): LocalSearchItem {
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

export function CommandMenu() {
  const navigate = useNavigate()
  const { setTheme } = useTheme()
  const { open, setOpen } = useSearch()
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

  const resetCommandSearch = () => {
    setSearchValue('')
    setLocalResults([])
    setLocalSearchState('idle')
    setBarcodeMatches([])
    setBarcodeState('idle')
  }

  const handleOpenChange = (nextOpen: boolean) => {
    setOpen(nextOpen)
    if (!nextOpen) {
      resetCommandSearch()
    }
  }

  const handleSearchValueChange = (value: string) => {
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

  const runCommand = React.useCallback(
    (command: () => unknown) => {
      setOpen(false)
      command()
    },
    [setOpen]
  )

  useEffect(() => {
    if (!open || trimmedSearch.length < 2) {
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
  }, [open, trimmedSearch])

  useEffect(() => {
    if (!open || !barcodeCandidate) {
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
              .filter((match) => match.id !== '')
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
  }, [barcodeCandidate, open, trimmedSearch])

  const openInventoryFilter = (filter: string) => {
    void navigate({
      to: '/inventory',
      search: { filter } as never,
    })
  }

  const openScannerForBarcode = (barcode: string) => {
    void navigate({
      to: '/scanner',
      search: { barcode } as never,
    })
  }

  return (
    <CommandDialog modal open={open} onOpenChange={handleOpenChange}>
      <CommandInput
        placeholder='Type a command or search...'
        value={searchValue}
        onValueChange={handleSearchValueChange}
        onInput={(event) =>
          handleSearchValueChange(event.currentTarget.value)
        }
      />
      <CommandList>
        <ScrollArea type='hover' className='h-72 pe-1'>
          <CommandEmpty>No results found.</CommandEmpty>
          {trimmedSearch.length >= 2 ? (
            <CommandGroup heading='Local catalog'>
              {localSearchState === 'loading' ? (
                <CommandItem disabled value={`loading-${trimmedSearch}`}>
                  <LoaderCircle className='animate-spin' />
                  Searching local catalog...
                </CommandItem>
              ) : null}
              {localSearchState === 'error' ? (
                <CommandItem disabled value={`search-error-${trimmedSearch}`}>
                  <PackageSearch />
                  Local catalog search is unavailable.
                </CommandItem>
              ) : null}
              {localSearchState === 'success' && localResults.length === 0 ? (
                <CommandItem disabled value={`search-empty-${trimmedSearch}`}>
                  <PackageSearch />
                  No local catalog matches.
                </CommandItem>
              ) : null}
              {localResults.map((item) => {
                const filter = item.part_number || item.title || item.id
                return (
                  <CommandItem
                    key={item.id}
                    value={`${item.part_number} ${item.title} ${item.brand} ${item.category}`}
                    data-testid={`command-local-result-${item.id}`}
                    onSelect={() =>
                      runCommand(() => openInventoryFilter(filter))
                    }
                  >
                    <PackageSearch />
                    <div className='min-w-0'>
                      <p className='truncate font-medium'>{item.title}</p>
                      <p className='truncate text-xs text-muted-foreground'>
                        {item.part_number || item.id} · {item.brand} ·{' '}
                        {item.category}
                      </p>
                    </div>
                  </CommandItem>
                )
              })}
            </CommandGroup>
          ) : null}
          {barcodeCandidate ? (
            <CommandGroup heading='Barcode lookup'>
              {barcodeState === 'loading' ? (
                <CommandItem disabled value={`barcode-loading-${trimmedSearch}`}>
                  <LoaderCircle className='animate-spin' />
                  Looking up barcode {trimmedSearch}...
                </CommandItem>
              ) : null}
              {barcodeState === 'error' ? (
                <CommandItem disabled value={`barcode-error-${trimmedSearch}`}>
                  <Barcode />
                  Barcode lookup is unavailable.
                </CommandItem>
              ) : null}
              {barcodeState === 'success' && barcodeMatches.length > 0
                ? barcodeMatches.map((match) => (
                    <CommandItem
                      key={match.id}
                      value={`${match.barcode} ${match.item_id}`}
                      data-testid={`command-barcode-match-${match.id}`}
                      onSelect={() =>
                        runCommand(() => openInventoryFilter(match.item_id))
                      }
                    >
                      <Barcode />
                      <div className='min-w-0'>
                        <p className='truncate font-medium'>
                          Local barcode match
                        </p>
                        <p className='truncate text-xs text-muted-foreground'>
                          {match.barcode} · {match.item_id}
                        </p>
                      </div>
                    </CommandItem>
                  ))
                : null}
              {barcodeState === 'success' && barcodeMatches.length === 0 ? (
                <>
                  <CommandItem
                    disabled
                    value={`barcode-empty-${trimmedSearch}`}
                    data-testid='command-barcode-empty'
                  >
                    <Barcode />
                    No local barcode match. Open Market Watch to continue.
                  </CommandItem>
                  <CommandItem
                    value={`open-market-watch-${trimmedSearch}`}
                    data-testid='command-barcode-open-scanner'
                    onSelect={() =>
                      runCommand(() => openScannerForBarcode(trimmedSearch))
                    }
                  >
                    <ScanSearch />
                    Open Market Watch for {trimmedSearch}
                  </CommandItem>
                </>
              ) : null}
            </CommandGroup>
          ) : null}
          {trimmedSearch.length >= 2 ? <CommandSeparator /> : null}
          {sidebarData.navGroups.map((group) => (
            <CommandGroup key={group.title} heading={group.title}>
              {group.items.map((navItem, i) => {
                if (navItem.url)
                  return (
                    <CommandItem
                      key={`${navItem.url}-${i}`}
                      value={navItem.title}
                      onSelect={() => {
                        runCommand(() => navigate({ to: navItem.url }))
                      }}
                    >
                      <div className='flex size-4 items-center justify-center'>
                        <ArrowRight className='size-2 text-muted-foreground/80' />
                      </div>
                      {navItem.title}
                    </CommandItem>
                  )

                return navItem.items?.map((subItem, i) => (
                  <CommandItem
                    key={`${navItem.title}-${subItem.url}-${i}`}
                    value={`${navItem.title}-${subItem.url}`}
                    onSelect={() => {
                      runCommand(() => navigate({ to: subItem.url }))
                    }}
                  >
                    <div className='flex size-4 items-center justify-center'>
                      <ArrowRight className='size-2 text-muted-foreground/80' />
                    </div>
                    {navItem.title} <ChevronRight /> {subItem.title}
                  </CommandItem>
                ))
              })}
            </CommandGroup>
          ))}
          <CommandSeparator />
          <CommandGroup heading='Theme'>
            <CommandItem onSelect={() => runCommand(() => setTheme('light'))}>
              <Sun /> <span>Light</span>
            </CommandItem>
            <CommandItem onSelect={() => runCommand(() => setTheme('dark'))}>
              <Moon className='scale-90' />
              <span>Dark</span>
            </CommandItem>
            <CommandItem onSelect={() => runCommand(() => setTheme('system'))}>
              <Laptop />
              <span>System</span>
            </CommandItem>
          </CommandGroup>
        </ScrollArea>
      </CommandList>
    </CommandDialog>
  )
}
