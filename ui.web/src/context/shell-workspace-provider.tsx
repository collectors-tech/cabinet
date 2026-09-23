import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import {
  ShellWorkspaceContext,
  type ShellWorkspace,
  type ShellWorkspaceContextValue,
  normalizeWorkspace,
  shellWorkspaceStorageKey,
} from './shell-workspace-context'

type ShellWorkspaceProviderProps = {
  children: React.ReactNode
}

export function ShellWorkspaceProvider({
  children,
}: ShellWorkspaceProviderProps) {
  const [activeProfileId, setActiveProfileId] = useState('local')
  const [activeWorkspace, setActiveWorkspaceState] =
    useState<ShellWorkspace>('navigation')
  const activeWorkspaceRef = useRef<ShellWorkspace>('navigation')
  const userSelectedWorkspaceRef = useRef<ShellWorkspace | null>(null)

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
        const savedWorkspace = (() => {
          try {
            return normalizeWorkspace(
              window.localStorage.getItem(
                shellWorkspaceStorageKey(nextProfileId)
              )
            )
          } catch {
            return 'navigation'
          }
        })()
        const userSelectedWorkspace = userSelectedWorkspaceRef.current
        setActiveProfileId(nextProfileId)
        if (userSelectedWorkspace) {
          activeWorkspaceRef.current = userSelectedWorkspace
          setActiveWorkspaceState(userSelectedWorkspace)
          try {
            window.localStorage.setItem(
              shellWorkspaceStorageKey(nextProfileId),
              userSelectedWorkspace
            )
          } catch {
            // Ignore storage failures and keep the user's in-memory selection.
          }
          return
        }
        activeWorkspaceRef.current = savedWorkspace
        setActiveWorkspaceState(savedWorkspace)
      } catch {
        if (!cancelled) {
          setActiveProfileId('local')
          activeWorkspaceRef.current = 'navigation'
          setActiveWorkspaceState('navigation')
        }
      }
    }

    void loadActiveProfile()
    return () => {
      cancelled = true
    }
  }, [])

  const setActiveWorkspace = useCallback(
    (workspace: ShellWorkspace) => {
      activeWorkspaceRef.current = workspace
      userSelectedWorkspaceRef.current = workspace
      setActiveWorkspaceState(workspace)
      try {
        window.localStorage.setItem(
          shellWorkspaceStorageKey(activeProfileId),
          workspace
        )
      } catch {
        // Ignore storage failures and keep in-memory state.
      }
    },
    [activeProfileId]
  )

  const toggleAssistantWorkspace = useCallback(() => {
    setActiveWorkspace(
      activeWorkspaceRef.current === 'assistant' ? 'navigation' : 'assistant'
    )
  }, [setActiveWorkspace])

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
