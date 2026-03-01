import * as React from 'react'
import { ChevronsUpDown, Plus } from 'lucide-react'
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

type TeamSwitcherProps = {
  teams: {
    name: string
    logo: React.ElementType
    plan: string
  }[]
}

export function TeamSwitcher({ teams }: TeamSwitcherProps) {
  const { t } = useTranslation('common')
  const { isMobile } = useSidebar()
  const [availableWorkspaces, setAvailableWorkspaces] = React.useState(teams)
  const [activeTeam, setActiveTeam] = React.useState(teams[0])
  const [loading, setLoading] = React.useState(false)

  React.useEffect(() => {
    let cancelled = false
    async function loadProfiles() {
      setLoading(true)
      try {
        const [profilesResp, activeResp] = await Promise.all([
          fetch('/api/profiles'),
          fetch('/api/profiles/active'),
        ])
        if (!profilesResp.ok || !activeResp.ok) {
          return
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
              plan: 'Profile DB',
            }
          })
          .filter((workspace): workspace is { id: string; name: string; logo: React.ElementType; plan: string } => Boolean(workspace))

        if (!profileWorkspaces.length || cancelled) {
          return
        }

        setAvailableWorkspaces(
          profileWorkspaces.map((workspace) => ({
            name: workspace.name,
            logo: workspace.logo,
            plan: workspace.plan,
          }))
        )

        const selected =
          profileWorkspaces.find((workspace) => workspace.id === activePayload.id) ??
          profileWorkspaces[0]

        setActiveTeam({
          name: selected.name,
          logo: selected.logo,
          plan: selected.plan,
        })
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
  }, [teams])

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
    const selectedWorkspace = availableWorkspaces.find((workspace) => workspace.name === targetName)
    if (selectedWorkspace) {
      setActiveTeam(selectedWorkspace)
    }
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
                <activeTeam.logo className='size-4' />
              </div>
              <div className='grid flex-1 text-start text-sm leading-tight'>
                <span className='truncate font-semibold' data-testid='active-profile-name'>
                  {activeTeam.name}
                </span>
                <span className='truncate text-xs'>
                  {loading ? 'Loading profiles...' : activeTeam.plan}
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
                data-testid={`team-option-${team.name.toLowerCase().replace(/[^a-z0-9]+/g, '-')}`}
                onClick={() => {
                  void switchProfile(team.name)
                }}
                className='gap-2 p-2'
              >
                <div className='flex size-6 items-center justify-center rounded-sm border'>
                  <team.logo className='size-4 shrink-0' />
                </div>
                {team.name}
                <DropdownMenuShortcut>⌘{index + 1}</DropdownMenuShortcut>
              </DropdownMenuItem>
            ))}
            <DropdownMenuSeparator />
            <DropdownMenuItem className='gap-2 p-2'>
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
