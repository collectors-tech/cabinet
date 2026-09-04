import { useMemo, useState } from 'react'
import { useNavigate } from '@tanstack/react-router'
import { ArrowRight } from 'lucide-react'
import { buildSearchNavigationResults } from '@/lib/route-navigation'
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from '@/components/ui/command'

export function SearchWorkspacePanel() {
  const navigate = useNavigate()
  const [searchValue, setSearchValue] = useState('')
  const navigationResults = useMemo(() => buildSearchNavigationResults(), [])

  return (
    <aside
      className='flex min-h-full w-full min-w-0 flex-col bg-slate-950 px-2 py-2 text-slate-100'
      data-testid='shell-search-workspace'
    >
      <Command
        className='min-h-0 w-full min-w-0 flex-1 rounded-none bg-transparent text-slate-100 [&_[cmdk-input-wrapper]]:h-10 [&_[cmdk-input-wrapper]]:w-full [&_[cmdk-input-wrapper]]:min-w-0 [&_[cmdk-input-wrapper]]:rounded-md [&_[cmdk-input-wrapper]]:border [&_[cmdk-input-wrapper]]:border-slate-700 [&_[cmdk-input-wrapper]]:bg-slate-900/80 [&_[cmdk-input-wrapper]]:px-2 [&_[cmdk-input-wrapper]]:text-slate-100 [&_[cmdk-input]]:h-9 [&_[cmdk-input]]:min-w-0 [&_[cmdk-input]]:flex-1 [&_[cmdk-input]]:text-sm [&_[cmdk-input]]:placeholder:text-slate-500 [&_[cmdk-list]]:max-h-none [&_[cmdk-list]]:flex-1 [&_[cmdk-list]]:overflow-y-auto [&_[cmdk-item]]:px-2 [&_[cmdk-item]]:py-2 [&_[cmdk-item]]:text-slate-100 [&_[cmdk-item][data-selected=true]]:bg-slate-800 [&_[cmdk-item][data-selected=true]]:text-white'
        value={searchValue}
        onValueChange={setSearchValue}
      >
        <CommandInput
          autoFocus
          aria-label='Search Cabinet navigation'
          data-testid='shell-search-workspace-input'
          placeholder='Search nav, settings, help...'
        />
        <CommandList
          className='mt-2 min-h-0 flex-1 pr-1'
          data-testid='shell-search-nav-results'
        >
          <CommandEmpty className='py-6 text-center text-xs text-slate-400'>
            No navigation results.
          </CommandEmpty>
          <CommandGroup heading='Navigation'>
            {navigationResults.map((result) => (
              <CommandItem
                key={result.id}
                value={result.value}
                data-testid='shell-search-nav-result'
                onSelect={() => {
                  void navigate({ to: result.path })
                }}
              >
                <div className='flex size-4 items-center justify-center'>
                  <ArrowRight
                    className='size-2 text-slate-500'
                    aria-hidden
                  />
                </div>
                <div className='min-w-0'>
                  <p
                    className='truncate text-sm font-semibold text-white'
                    data-testid='shell-search-nav-result-title'
                  >
                    {result.title}
                  </p>
                  <p
                    className='truncate text-xs text-slate-400'
                    data-testid='shell-search-nav-result-meta'
                  >
                    {result.group} · {result.path}
                  </p>
                </div>
              </CommandItem>
            ))}
          </CommandGroup>
        </CommandList>
      </Command>
    </aside>
  )
}
