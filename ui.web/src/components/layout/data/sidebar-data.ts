import {
  LayoutDashboard,
  Telescope,
  ScanSearch,
  Monitor,
  ListChecks,
  Heart,
  PlugZap,
  HelpCircle,
  Bell,
  Palette,
  Settings,
  Wrench,
  UserCog,
  Users,
  MessagesSquare,
  Database,
  ChartColumn,
} from 'lucide-react'
import { type SidebarData } from '../types'

export const sidebarData: SidebarData = {
  user: {
    name: 'Local Admin',
    email: 'admin@local',
    avatar: '/avatars/01.png',
  },
  teams: [
    {
      name: 'Local Workspace',
      logo: Database,
      plan: 'Primary DB',
    },
  ],
  navGroups: [
    {
      title: 'General',
      items: [
        {
          title: 'Dashboard',
          url: '/dashboard',
          icon: LayoutDashboard,
        },
        {
          title: 'Inventory',
          url: '/inventory',
          icon: ListChecks,
        },
        {
          title: 'Collections',
          url: '/collections',
          icon: ListChecks,
        },
        {
          title: 'Wishlist',
          url: '/wishlist',
          icon: Heart,
        },
        {
          title: 'Discoveries',
          url: '/discoveries',
          icon: Telescope,
        },
        {
          title: 'Market Watch',
          url: '/scanner',
          icon: ScanSearch,
        },
        {
          title: 'Integrations',
          url: '/integrations',
          icon: PlugZap,
        },
        {
          title: 'Chats',
          url: '/chats',
          badge: '3',
          icon: MessagesSquare,
        },
        {
          title: 'Users',
          url: '/users',
          icon: Users,
        },
        {
          title: 'Reports',
          url: '/reports',
          icon: ChartColumn,
        },
      ],
    },
    {
      title: 'Other',
      items: [
        {
          title: 'Settings',
          icon: Settings,
          items: [
            {
              title: 'Profile',
              url: '/settings/profile',
              icon: UserCog,
            },
            {
              title: 'Account',
              url: '/settings/account',
              icon: Wrench,
            },
            {
              title: 'Appearance',
              url: '/settings/appearance',
              icon: Palette,
            },
            {
              title: 'Notifications',
              url: '/settings/notifications',
              icon: Bell,
            },
            {
              title: 'Display',
              url: '/settings/display',
              icon: Monitor,
            },
            {
              title: 'Storage',
              url: '/settings/storage',
              icon: Database,
            },
          ],
        },
        {
          title: 'Storage',
          url: '/settings/storage',
          icon: Database,
        },
        {
          title: 'Help Center',
          url: '/help-center',
          icon: HelpCircle,
        },
      ],
    },
  ],
}
