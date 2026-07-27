import { useEffect } from 'react'
import { Outlet, useLocation } from '@tanstack/react-router'
import {
  Monitor,
  Bell,
  Palette,
  Wrench,
  UserCog,
  Cog,
  CreditCard,
  HardDrive,
  PlugZap,
  Settings as SettingsIcon,
  Sparkles,
  Tags,
} from 'lucide-react'
import { ConfigDrawer } from '@/components/config-drawer'
import { LanguageSwitch } from '@/components/language-switch'
import { Header, HeaderTitle } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { getRouteMetadata } from '@/lib/route-metadata'
import { SidebarNav } from './components/sidebar-nav'

const sidebarNavItems = [
  {
    title: 'Profile',
    href: '/settings/profile',
    icon: <UserCog size={18} />,
  },
  {
    title: 'Account',
    href: '/settings/account',
    icon: <Wrench size={18} />,
  },
  {
    title: 'Appearance',
    href: '/settings/appearance',
    icon: <Palette size={18} />,
  },
  {
    title: 'Notifications',
    href: '/settings/notifications',
    icon: <Bell size={18} />,
  },
  {
    title: 'Display',
    href: '/settings/display',
    icon: <Monitor size={18} />,
  },
  {
    title: 'Storage',
    href: '/settings/storage',
    icon: <HardDrive size={18} />,
  },
  {
    title: 'Integrations',
    href: '/settings/integrations',
    icon: <PlugZap size={18} />,
  },
  {
    title: 'Skills',
    href: '/settings/skills',
    icon: <Sparkles size={18} />,
  },
  {
    title: 'Categories',
    href: '/settings/categories',
    icon: <Tags size={18} />,
  },
  {
    title: 'Operations',
    href: '/settings/operations',
    icon: <Cog size={18} />,
  },
  {
    title: 'Billing',
    href: '/settings/billing',
    icon: <CreditCard size={18} />,
  },
]

export function Settings() {
  const { pathname } = useLocation()
  const routeMetadata = getRouteMetadata(pathname)

  useEffect(() => {
    if (routeMetadata) {
      document.title = routeMetadata.documentTitle
    }
  }, [routeMetadata])

  return (
    <>
      {/* ===== Top Heading ===== */}
      <Header>
        <Search />
        <HeaderTitle
          title={routeMetadata?.title ?? 'Settings'}
          description={
            routeMetadata?.description ??
            'Manage account, appearance, storage, and operations preferences.'
          }
          icon={routeMetadata?.icon ?? SettingsIcon}
          testId={routeMetadata?.testIds.headerTitle ?? 'settings-header-title'}
          iconTestId={routeMetadata?.testIds.headerIcon ?? 'settings-header-icon'}
        />
        <div
          className='ms-auto flex items-center space-x-4'
          data-header-title-avoid='true'
        >
          <LanguageSwitch />
          <ThemeSwitch />
          <ConfigDrawer />
          <ProfileDropdown />
        </div>
      </Header>

      <Main fixed>
        <div className='flex flex-1 flex-col space-y-2 overflow-hidden md:space-y-2 lg:flex-row lg:space-y-0 lg:space-x-12'>
          <aside className='top-0 lg:sticky lg:w-1/5'>
            <SidebarNav items={sidebarNavItems} />
          </aside>
          <div className='flex w-full overflow-y-hidden p-1'>
            <Outlet />
          </div>
        </div>
      </Main>
    </>
  )
}
