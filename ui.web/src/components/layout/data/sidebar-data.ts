import { createElement } from 'react'
import {
  LayoutDashboard,
  Telescope,
  ScanSearch,
  Monitor,
  ListChecks,
  Heart,
  PlugZap,
  HelpCircle,
  Inbox,
  Images,
  Bell,
  Palette,
  Settings,
  Wrench,
  UserCog,
  Users,
  MessagesSquare,
  Database,
  ChartColumn,
  Tag,
  type LucideProps,
} from 'lucide-react'
import { type SidebarData } from '../types'

function CollectionsTagIcon(props: LucideProps) {
  return createElement(Tag as any, {
    ...props,
    'data-lucide': 'tag',
  })
}

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
      testIdKey: 'general',
      items: [
        {
          title: 'Dashboard',
          testIdKey: 'dashboard',
          url: '/dashboard',
          icon: LayoutDashboard,
        },
        {
          title: 'Inventory',
          testIdKey: 'inventory',
          url: '/inventory',
          icon: ListChecks,
        },
        {
          title: 'Media',
          testIdKey: 'media',
          url: '/media',
          icon: Images,
        },
        {
          title: 'Collections',
          testIdKey: 'collections',
          url: '/collections',
          icon: CollectionsTagIcon,
        },
        {
          title: 'Wishlist',
          testIdKey: 'wishlist',
          url: '/wishlist',
          icon: Heart,
        },
        {
          title: 'Discoveries',
          testIdKey: 'discoveries',
          url: '/discoveries',
          icon: Telescope,
        },
        {
          title: 'Market Watch',
          testIdKey: 'market-watch',
          url: '/scanner',
          icon: ScanSearch,
        },
        {
          title: 'Purchase Inbox',
          testIdKey: 'purchase-inbox',
          url: '/inbox',
          icon: Inbox,
        },
        {
          title: 'Integrations',
          testIdKey: 'integrations',
          url: '/integrations',
          icon: PlugZap,
        },
        {
          title: 'Chats',
          testIdKey: 'chats',
          url: '/chats',
          badge: '3',
          icon: MessagesSquare,
        },
        {
          title: 'Users',
          testIdKey: 'users',
          url: '/users',
          icon: Users,
        },
        {
          title: 'Reports',
          testIdKey: 'reports',
          url: '/reports',
          icon: ChartColumn,
        },
      ],
    },
    {
      title: 'Other',
      testIdKey: 'other',
      items: [
        {
          title: 'Settings',
          testIdKey: 'settings',
          icon: Settings,
          items: [
            {
              title: 'Profile',
              testIdKey: 'profile',
              url: '/settings/profile',
              icon: UserCog,
            },
            {
              title: 'Account',
              testIdKey: 'account',
              url: '/settings/account',
              icon: Wrench,
            },
            {
              title: 'Appearance',
              testIdKey: 'appearance',
              url: '/settings/appearance',
              icon: Palette,
            },
            {
              title: 'Notifications',
              testIdKey: 'notifications',
              url: '/settings/notifications',
              icon: Bell,
            },
            {
              title: 'Display',
              testIdKey: 'display',
              url: '/settings/display',
              icon: Monitor,
            },
            {
              title: 'Storage',
              testIdKey: 'storage',
              url: '/settings/storage',
              icon: Database,
            },
          ],
        },
        {
          title: 'Storage',
          testIdKey: 'storage-shortcut',
          url: '/settings/storage',
          icon: Database,
        },
        {
          title: 'Help Center',
          testIdKey: 'help-center',
          url: '/help-center',
          icon: HelpCircle,
        },
      ],
    },
  ],
}
