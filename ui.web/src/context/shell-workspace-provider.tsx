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
        setActiveWorkspaceState(savedWorkspace)
      } catch {
        if (!cancelled) {
          setActiveProfileId('local')
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
