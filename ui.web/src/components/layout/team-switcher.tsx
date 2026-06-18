import * as React from 'react'
import { AlertTriangle, ChevronsUpDown, Plus, RefreshCw } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuShortcut,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from '@/components/ui/sidebar'
import { DatabaseProfileIcon } from '@/assets/database-profile-icon'

type TeamSwitcherProps = {
  teams: {
    name: string
    logo: React.ElementType
    plan: string
  }[]
}

function profilePlanLabel(name: string) {
  return /showcase|sample|demo/i.test(name)
    ? 'Showcase sample data'
    : 'Database'
}

function profileTestIdSlug(name: string) {
  return name.toLowerCase().replace(/[^a-z0-9]+/g, '-')
}

export function TeamSwitcher({ teams }: TeamSwitcherProps) {
  const { t } = useTranslation('common')
  const { isMobile } = useSidebar()
  const [availableWorkspaces, setAvailableWorkspaces] = React.useState(teams)
  const [activeTeam, setActiveTeam] = React.useState(teams[0])
  const [loading, setLoading] = React.useState(false)
  const [loadError, setLoadError] = React.useState<string | null>(null)
  const [reloadKey, setReloadKey] = React.useState(0)

  React.useEffect(() => {
    let cancelled = false
    async function loadProfiles() {
      setLoading(true)
      setLoadError(null)
      try {
        const [profilesResp, activeResp] = await Promise.all([
          fetch('/api/profiles'),
          fetch('/api/profiles/active'),
        ])
        if (!profilesResp.ok || !activeResp.ok) {
          throw new Error('profile-load-failed')
        }

        const profilesPayload = (await profilesResp.json()) as {
          profiles?: Array<{ id?: string; name?: string }>
        }
        const activePayload = (await activeResp.json()) as {
          id?: string
          name?: string
        }

        const profileWorkspaces = (profilesPayload.profiles ?? [])
          .map((profile, index) => {
            const id = profile.id?.trim()
            const name = profile.name?.trim()
            if (!id || !name) {
              return null
            }
            return {
              id,
              name,
              logo: teams[index]?.logo ?? teams[0]?.logo,
              plan: profilePlanLabel(name),
            }
          })
          .filter(
            (
              workspace
            ): workspace is {
              id: string
              name: string
              logo: React.ElementType
              plan: string
            } => Boolean(workspace)
          )

        if (!profileWorkspaces.length || cancelled) {
          throw new Error('profile-empty')
        }

        setAvailableWorkspaces(
          profileWorkspaces.map((workspace) => ({
            name: workspace.name,
            logo: workspace.logo,
            plan: workspace.plan,
          }))
        )

        const selected =
          profileWorkspaces.find(
            (workspace) => workspace.id === activePayload.id
          ) ?? profileWorkspaces[0]

        setActiveTeam({
          name: selected.name,
          logo: selected.logo,
          plan: selected.plan,
        })
      } catch {
        if (!cancelled) {
          setLoadError('Profile unavailable. Retry loading databases.')
        }
      } finally {
        if (!cancelled) {
          setLoading(false)
        }
      }
    }

    void loadProfiles()
    return () => {
      cancelled = true
    }
  }, [teams, reloadKey])

  const switchProfile = async (targetName: string) => {
    const profileResp = await fetch('/api/profiles')
    if (!profileResp.ok) {
      return
    }
    const payload = (await profileResp.json()) as {
      profiles?: Array<{ id?: string; name?: string }>
    }
    const target = (payload.profiles ?? []).find(
      (profile) => profile.name?.trim() === targetName
    )
    const profileID = target?.id?.trim()
    if (!profileID) {
      return
    }
    const setActiveResp = await fetch('/api/profiles/active', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ profile_id: profileID }),
    })
    if (!setActiveResp.ok) {
      return
    }
    const selectedWorkspace = availableWorkspaces.find(
      (workspace) => workspace.name === targetName
    )
    if (selectedWorkspace) {
      setActiveTeam(selectedWorkspace)
    }
    if (typeof window !== 'undefined') {
      window.location.reload()
    }
  }

  const createProfileFromSwitcher = async () => {
    if (typeof window === 'undefined') {
      return
    }

    const profileName = window.prompt('New database name')?.trim()
    if (!profileName) {
      return
    }

    setLoadError(null)
    const createResp = await fetch('/api/profiles', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name: profileName }),
    })
    if (!createResp.ok) {
      setLoadError('Profile unavailable. Retry loading databases.')
      return
    }

    const created = (await createResp.json()) as { id?: string; name?: string }
    const profileID = created.id?.trim()
    if (!profileID) {
      setLoadError('Profile unavailable. Retry loading databases.')
      return
    }

    const setActiveResp = await fetch('/api/profiles/active', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ profile_id: profileID }),
    })
    if (!setActiveResp.ok) {
      setLoadError('Profile unavailable. Retry loading databases.')
      return
    }

    const selectedWorkspace = {
      name: created.name?.trim() || profileName,
      logo: teams[availableWorkspaces.length]?.logo ?? teams[0]?.logo,
      plan: profilePlanLabel(created.name?.trim() || profileName),
    }
    setAvailableWorkspaces((workspaces) => [...workspaces, selectedWorkspace])
    setActiveTeam(selectedWorkspace)
    window.location.reload()
  }

  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <SidebarMenuButton
              data-testid='team-switcher-trigger'
              size='lg'
              className='data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground'
            >
              <div className='flex aspect-square size-8 items-center justify-center rounded-lg bg-sidebar-primary text-sidebar-primary-foreground'>
                <DatabaseProfileIcon
                  className='size-8'
                  label={`${activeTeam.name} database profile`}
                  testId='active-profile-db-icon'
                  variant='dark'
                />
              </div>
              <div className='grid flex-1 text-start text-sm leading-tight'>
                <span
                  className='truncate font-semibold'
                  data-testid='active-profile-name'
                >
                  {activeTeam.name}
                </span>
                <span className='truncate text-xs'>
                  <span data-testid='active-profile-status'>
                    {loadError ??
                      (loading ? 'Loading profiles...' : activeTeam.plan)}
                  </span>
                </span>
              </div>
              <ChevronsUpDown className='ms-auto' />
            </SidebarMenuButton>
          </DropdownMenuTrigger>
          <DropdownMenuContent
            className='w-(--radix-dropdown-menu-trigger-width) min-w-56 rounded-lg'
            align='start'
            side={isMobile ? 'bottom' : 'right'}
            sideOffset={4}
          >
            <DropdownMenuLabel className='text-xs text-muted-foreground'>
              {t('common:workspace.label')}
            </DropdownMenuLabel>
            {availableWorkspaces.map((team, index) => (
              <DropdownMenuItem
                key={team.name}
                data-testid={`team-option-${profileTestIdSlug(team.name)}`}
                onClick={() => {
                  void switchProfile(team.name)
                }}
                className='gap-2 p-2'
              >
                <div className='flex size-6 items-center justify-center rounded-sm border'>
                  <DatabaseProfileIcon
                    className='size-6 shrink-0'
                    label={`${team.name} database profile`}
                    testId={`team-option-${profileTestIdSlug(team.name)}-icon`}
                    variant='theme'
                  />
                </div>
                <div className='grid flex-1 text-start leading-tight'>
                  <span>{team.name}</span>
                  <span
                    className='text-xs text-muted-foreground'
                    data-testid={`team-option-${team.name.toLowerCase().replace(/[^a-z0-9]+/g, '-')}-plan`}
                  >
                    {team.plan}
                  </span>
                </div>
                <DropdownMenuShortcut>⌘{index + 1}</DropdownMenuShortcut>
              </DropdownMenuItem>
            ))}
            {loadError ? (
              <>
                <DropdownMenuSeparator />
                <DropdownMenuItem
                  className='gap-2 p-2 text-destructive focus:text-destructive'
                  data-testid='team-switcher-profile-error'
                  onSelect={(event) => {
                    event.preventDefault()
                  }}
                >
                  <AlertTriangle className='size-4 shrink-0' />
                  <span className='text-xs'>{loadError}</span>
                </DropdownMenuItem>
                <DropdownMenuItem
                  className='gap-2 p-2'
                  data-testid='team-switcher-retry-profiles'
                  onClick={() => setReloadKey((key) => key + 1)}
                >
                  <RefreshCw className='size-4' />
                  <div className='font-medium'>Retry profiles</div>
                </DropdownMenuItem>
              </>
            ) : null}
            <DropdownMenuSeparator />
            <DropdownMenuItem
              className='gap-2 p-2'
              data-testid='team-switcher-add-profile'
              onClick={() => {
                void createProfileFromSwitcher()
              }}
            >
              <div className='flex size-6 items-center justify-center rounded-md border bg-background'>
                <Plus className='size-4' />
              </div>
              <div className='font-medium text-muted-foreground'>
                {t('common:workspace.add')}
              </div>
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </SidebarMenuItem>
    </SidebarMenu>
  )
}
