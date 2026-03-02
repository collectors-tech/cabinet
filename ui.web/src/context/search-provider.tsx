import { createContext, useContext, useEffect, useState } from 'react'
import { CommandMenu } from '@/components/command-menu'
import { getShortcutKey } from '@/lib/keyboard-shortcuts'

type SearchContextType = {
  open: boolean
  setOpen: React.Dispatch<React.SetStateAction<boolean>>
}

const SearchContext = createContext<SearchContextType | null>(null)

type SearchProviderProps = {
  children: React.ReactNode
}

export function SearchProvider({ children }: SearchProviderProps) {
  const [open, setOpen] = useState(false)

  useEffect(() => {
    const commandShortcut = getShortcutKey('command-palette')
    const down = (e: KeyboardEvent) => {
      const key = e.key?.toLowerCase()
      const shortcutCode =
        commandShortcut.length === 1
          ? `Key${commandShortcut.toUpperCase()}`
          : commandShortcut
      if (
        (key === commandShortcut || e.code === shortcutCode) &&
        (e.metaKey || e.ctrlKey)
      ) {
        e.preventDefault()
        setOpen((open) => !open)
      }
    }
    document.addEventListener('keydown', down)
    return () => document.removeEventListener('keydown', down)
  }, [])

  return (
    <SearchContext value={{ open, setOpen }}>
      {children}
      <CommandMenu />
    </SearchContext>
  )
}

// eslint-disable-next-line react-refresh/only-export-components
export const useSearch = () => {
  const searchContext = useContext(SearchContext)

  if (!searchContext) {
    throw new Error('useSearch has to be used within SearchProvider')
  }

  return searchContext
}
