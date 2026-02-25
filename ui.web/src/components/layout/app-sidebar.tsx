import { useLayout } from '@/context/layout-provider'
import { useAuthStore } from '@/stores/auth-store'
import { useTranslation } from 'react-i18next'
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarRail,
} from '@/components/ui/sidebar'
// import { AppTitle } from './app-title'
import { sidebarData } from './data/sidebar-data'
import { type NavCollapsible, type NavItem } from './types'
import { NavGroup } from './nav-group'
import { NavUser } from './nav-user'
import { TeamSwitcher } from './team-switcher'

export function AppSidebar() {
  const { collapsible, variant } = useLayout()
  const { t } = useTranslation('nav')
  const authUser = useAuthStore((state) => state.auth.user)
  const sidebarUser = authUser
    ? {
        name: authUser.accountNo || sidebarData.user.name,
        email: authUser.email || sidebarData.user.email,
        avatar: sidebarData.user.avatar,
      }
    : sidebarData.user

  const normalizeNavKey = (title: string) =>
    title.trim().toLowerCase().replace(/\s+/g, '-')

  const translateItem = (item: NavItem): NavItem => {
    if ('items' in item) {
      const collapsible = item as NavCollapsible
      return {
        ...collapsible,
        title: t(`items.${normalizeNavKey(collapsible.title)}`, {
          defaultValue: collapsible.title,
        }),
        items: collapsible.items.map((nested) => ({
          ...nested,
          title: t(`items.${normalizeNavKey(nested.title)}`, {
            defaultValue: nested.title,
          }),
        })),
      }
    }

    return {
      ...item,
      title: t(`items.${normalizeNavKey(item.title)}`, {
        defaultValue: item.title,
      }),
    }
  }

  const translatedNavGroups = sidebarData.navGroups.map((group) => ({
    ...group,
    title: t(`groups.${normalizeNavKey(group.title)}`, { defaultValue: group.title }),
    items: group.items.map(translateItem),
  }))

  return (
    <Sidebar collapsible={collapsible} variant={variant}>
      <SidebarHeader>
        <TeamSwitcher teams={sidebarData.teams} />

        {/* Replace <TeamSwitch /> with the following <AppTitle />
         /* if you want to use the normal app title instead of TeamSwitch dropdown */}
        {/* <AppTitle /> */}
      </SidebarHeader>
      <SidebarContent>
        {translatedNavGroups.map((props) => (
          <NavGroup key={props.title} {...props} />
        ))}
      </SidebarContent>
      <SidebarFooter>
        <NavUser user={sidebarUser} />
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  )
}
