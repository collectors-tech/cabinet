import { createContext, useContext } from 'react'

export type ShellWorkspace = 'navigation' | 'search' | 'assistant' | 'inbox'

export type ShellWorkspaceContextValue = {
  activeWorkspace: ShellWorkspace
  setActiveWorkspace: (workspace: ShellWorkspace) => void
  toggleAssistantWorkspace: () => void
  activeProfileId: string
}

export const ShellWorkspaceContext =
  createContext<ShellWorkspaceContextValue | null>(null)

export function normalizeWorkspace(
  value: string | null | undefined
): ShellWorkspace {
  switch (value) {
    case 'search':
    case 'assistant':
    case 'inbox':
      return value
    default:
      return 'navigation'
  }
}

export function shellWorkspaceStorageKey(profileId: string) {
  return `cabinet.shell.workspace.active.${profileId || 'local'}`
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
