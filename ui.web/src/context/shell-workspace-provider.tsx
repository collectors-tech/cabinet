import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from 'react'

export type ShellWorkspace = 'navigation' | 'assistant' | 'inbox'

type ShellWorkspaceContextValue = {
  activeWorkspace: ShellWorkspace
  setActiveWorkspace: (workspace: ShellWorkspace) => void
  toggleAssistantWorkspace: () => void
  activeProfileId: string
}

const ShellWorkspaceContext = createContext<ShellWorkspaceContextValue | null>(
  null
)

function normalizeWorkspace(value: string | null | undefined): ShellWorkspace {
  switch (value) {
    case 'assistant':
    case 'inbox':
      return value
    default:
      return 'navigation'
  }
}

function storageKey(profileId: string) {
  return `cabinet.shell.workspace.active.${profileId || 'local'}`
}

type ShellWorkspaceProviderProps = {
  children: React.ReactNode
}

export function ShellWorkspaceProvider({
  children,
}: ShellWorkspaceProviderProps) {
  const [activeProfileId, setActiveProfileId] = useState('local')
  const [activeWorkspace, setActiveWorkspaceState] =
    useState<ShellWorkspace>('navigation')

  useEffect(() => {
    let cancelled = false

    async function loadActiveProfile() {
      try {
        const response = await fetch('/api/profiles/active')
        if (!response.ok) {
          return
        }
        const payload = (await response.json()) as { id?: string }
        if (cancelled) {
          return
        }
        const nextProfileId = payload.id?.trim() || 'local'
        setActiveProfileId(nextProfileId)
      } catch {
        if (!cancelled) {
          setActiveProfileId('local')
        }
      }
    }

    void loadActiveProfile()
    return () => {
      cancelled = true
    }
  }, [])

  useEffect(() => {
    try {
      const saved = window.localStorage.getItem(storageKey(activeProfileId))
      setActiveWorkspaceState(normalizeWorkspace(saved))
    } catch {
      setActiveWorkspaceState('navigation')
    }
  }, [activeProfileId])

  const setActiveWorkspace = useCallback(
    (workspace: ShellWorkspace) => {
      setActiveWorkspaceState(workspace)
      try {
        window.localStorage.setItem(storageKey(activeProfileId), workspace)
      } catch {
        // Ignore storage failures and keep in-memory state.
      }
    },
    [activeProfileId]
  )

  const toggleAssistantWorkspace = useCallback(() => {
    setActiveWorkspace(
      activeWorkspace === 'assistant' ? 'navigation' : 'assistant'
    )
  }, [activeWorkspace, setActiveWorkspace])

  const contextValue = useMemo<ShellWorkspaceContextValue>(
    () => ({
      activeWorkspace,
      setActiveWorkspace,
      toggleAssistantWorkspace,
      activeProfileId,
    }),
    [
      activeProfileId,
      activeWorkspace,
      setActiveWorkspace,
      toggleAssistantWorkspace,
    ]
  )

  return (
    <ShellWorkspaceContext.Provider value={contextValue}>
      {children}
    </ShellWorkspaceContext.Provider>
  )
}

export function useShellWorkspace() {
  const context = useContext(ShellWorkspaceContext)
  if (!context) {
    throw new Error(
      'useShellWorkspace must be used within a ShellWorkspaceProvider'
    )
  }
  return context
}
