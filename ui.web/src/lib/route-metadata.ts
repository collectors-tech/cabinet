import {
  AlertTriangle,
  Bell,
  BrainCircuit,
  ChartColumn,
  CircleHelp,
  CreditCard,
  Database,
  Heart,
  HelpCircle,
  Images,
  Inbox,
  LayoutDashboard,
  ListChecks,
  MessagesSquare,
  Monitor,
  Palette,
  PlugZap,
  ScanSearch,
  Settings,
  Tag,
  Telescope,
  UserCog,
  Users,
  Wrench,
  type LucideIcon,
} from 'lucide-react'

export type RouteNavigationGroup = 'General' | 'Settings' | 'Other' | 'System'

export type RouteMetadata = {
  path: string
  aliases?: string[]
  title: string
  description: string
  icon: LucideIcon
  navigationGroup: RouteNavigationGroup
  documentTitle: string
  testIds: {
    headerTitle: string
    headerIcon: string
    sidebarLink?: string
  }
}

function titleFor(pageTitle: string) {
  return `Cabinet - ${pageTitle}`
}

export const authenticatedRouteMetadata: RouteMetadata[] = [
  {
    path: '/',
    aliases: ['/dashboard'],
    title: 'Home',
    description: 'Review collection health, recent activity, and shortcuts.',
    icon: LayoutDashboard,
    navigationGroup: 'General',
    documentTitle: titleFor('Home'),
    testIds: {
      headerTitle: 'dashboard-header-title',
      headerIcon: 'dashboard-header-icon',
      sidebarLink: 'sidebar-nav-link-dashboard',
    },
  },
  {
    path: '/inventory',
    title: 'Inventory',
    description: 'Manage cataloged items, folders, photos, and item evidence.',
    icon: ListChecks,
    navigationGroup: 'General',
    documentTitle: titleFor('Inventory'),
    testIds: {
      headerTitle: 'inventory-header-title',
      headerIcon: 'inventory-header-icon',
      sidebarLink: 'sidebar-nav-link-inventory',
    },
  },
  {
    path: '/media',
    title: 'Media',
    description: 'Review media assets, attachments, and linked records.',
    icon: Images,
    navigationGroup: 'General',
    documentTitle: titleFor('Media'),
    testIds: {
      headerTitle: 'media-header-title',
      headerIcon: 'media-header-icon',
      sidebarLink: 'sidebar-nav-link-media',
    },
  },
  {
    path: '/collections',
    title: 'Collections',
    description: 'Organize inventory into curated collection groupings.',
    icon: Tag,
    navigationGroup: 'General',
    documentTitle: titleFor('Collections'),
    testIds: {
      headerTitle: 'collections-header-title',
      headerIcon: 'collections-header-icon',
      sidebarLink: 'sidebar-nav-link-collections',
    },
  },
  {
    path: '/wishlist',
    title: 'Wishlist',
    description: 'Track wanted items, pricing, and purchase readiness.',
    icon: Heart,
    navigationGroup: 'General',
    documentTitle: titleFor('Wishlist'),
    testIds: {
      headerTitle: 'wishlist-header-title',
      headerIcon: 'wishlist-header-icon',
      sidebarLink: 'sidebar-nav-link-wishlist',
    },
  },
  {
    path: '/discoveries',
    title: 'Discoveries',
    description: 'Triage new provider findings and handoff decisions.',
    icon: Telescope,
    navigationGroup: 'General',
    documentTitle: titleFor('Discoveries'),
    testIds: {
      headerTitle: 'discoveries-header-title',
      headerIcon: 'discoveries-header-icon',
      sidebarLink: 'sidebar-nav-link-discoveries',
    },
  },
  {
    path: '/scanner',
    title: 'Market Watch',
    description: 'Run saved market searches and review provider results.',
    icon: ScanSearch,
    navigationGroup: 'General',
    documentTitle: titleFor('Market Watch'),
    testIds: {
      headerTitle: 'scanner-header-title',
      headerIcon: 'scanner-header-icon',
      sidebarLink: 'sidebar-nav-link-market-watch',
    },
  },
  {
    path: '/inbox',
    title: 'Notification Inbox',
    description: 'Review captured notifications and inbound work items.',
    icon: Bell,
    navigationGroup: 'General',
    documentTitle: titleFor('Notification Inbox'),
    testIds: {
      headerTitle: 'inbox-header-title',
      headerIcon: 'inbox-header-icon',
      sidebarLink: 'sidebar-nav-link-inbox',
    },
  },
  {
    path: '/purchases',
    title: 'Purchases',
    description: 'Track purchase orders, line items, and reconciliation.',
    icon: Inbox,
    navigationGroup: 'General',
    documentTitle: titleFor('Purchases'),
    testIds: {
      headerTitle: 'purchases-header-title',
      headerIcon: 'purchases-header-icon',
      sidebarLink: 'sidebar-nav-link-purchases',
    },
  },
  {
    path: '/integrations',
    title: 'Integrations',
    description: 'Configure connected providers and integration health.',
    icon: PlugZap,
    navigationGroup: 'General',
    documentTitle: titleFor('Integrations'),
    testIds: {
      headerTitle: 'integrations-header-title',
      headerIcon: 'integrations-header-icon',
      sidebarLink: 'sidebar-nav-link-integrations',
    },
  },
  {
    path: '/chats',
    title: 'Chats',
    description: 'Review assistant conversations and chat attachments.',
    icon: MessagesSquare,
    navigationGroup: 'General',
    documentTitle: titleFor('Chats'),
    testIds: {
      headerTitle: 'chats-header-title',
      headerIcon: 'chats-header-icon',
      sidebarLink: 'sidebar-nav-link-chats',
    },
  },
  {
    path: '/users',
    title: 'Users',
    description: 'Manage user access, invitations, and account state.',
    icon: Users,
    navigationGroup: 'General',
    documentTitle: titleFor('Users'),
    testIds: {
      headerTitle: 'users-header-title',
      headerIcon: 'users-header-icon',
      sidebarLink: 'sidebar-nav-link-users',
    },
  },
  {
    path: '/reports',
    title: 'Reports',
    description: 'Inspect reporting dashboards and collection trends.',
    icon: ChartColumn,
    navigationGroup: 'General',
    documentTitle: titleFor('Reports'),
    testIds: {
      headerTitle: 'reports-header-title',
      headerIcon: 'reports-header-icon',
      sidebarLink: 'sidebar-nav-link-reports',
    },
  },
  {
    path: '/settings',
    aliases: ['/settings/profile'],
    title: 'Profile Settings',
    description: 'Manage active profile identity and collection context.',
    icon: UserCog,
    navigationGroup: 'Settings',
    documentTitle: titleFor('Profile Settings'),
    testIds: {
      headerTitle: 'settings-header-title',
      headerIcon: 'settings-header-icon',
      sidebarLink: 'sidebar-nav-link-profile',
    },
  },
  {
    path: '/settings/account',
    title: 'Account Settings',
    description: 'Manage account details and local access options.',
    icon: Wrench,
    navigationGroup: 'Settings',
    documentTitle: titleFor('Account Settings'),
    testIds: {
      headerTitle: 'settings-header-title',
      headerIcon: 'settings-header-icon',
      sidebarLink: 'sidebar-nav-link-account',
    },
  },
  {
    path: '/settings/appearance',
    title: 'Appearance Settings',
    description: 'Choose theme, density, and display preferences.',
    icon: Palette,
    navigationGroup: 'Settings',
    documentTitle: titleFor('Appearance Settings'),
    testIds: {
      headerTitle: 'settings-header-title',
      headerIcon: 'settings-header-icon',
      sidebarLink: 'sidebar-nav-link-appearance',
    },
  },
  {
    path: '/settings/billing',
    title: 'Billing Settings',
    description: 'Review subscription plan and billing state.',
    icon: CreditCard,
    navigationGroup: 'Settings',
    documentTitle: titleFor('Billing Settings'),
    testIds: {
      headerTitle: 'settings-header-title',
      headerIcon: 'settings-header-icon',
      sidebarLink: 'sidebar-nav-link-billing',
    },
  },
  {
    path: '/settings/categories',
    title: 'Category Settings',
    description: 'Manage inventory categories and condition scales.',
    icon: Tag,
    navigationGroup: 'Settings',
    documentTitle: titleFor('Category Settings'),
    testIds: {
      headerTitle: 'settings-header-title',
      headerIcon: 'settings-header-icon',
      sidebarLink: 'sidebar-nav-link-categories',
    },
  },
  {
    path: '/settings/display',
    title: 'Display Settings',
    description: 'Tune display labels, layout defaults, and screen behavior.',
    icon: Monitor,
    navigationGroup: 'Settings',
    documentTitle: titleFor('Display Settings'),
    testIds: {
      headerTitle: 'settings-header-title',
      headerIcon: 'settings-header-icon',
      sidebarLink: 'sidebar-nav-link-display',
    },
  },
  {
    path: '/settings/integrations',
    title: 'Integration Settings',
    description: 'Manage integration credentials and provider setup.',
    icon: PlugZap,
    navigationGroup: 'Settings',
    documentTitle: titleFor('Integration Settings'),
    testIds: {
      headerTitle: 'settings-header-title',
      headerIcon: 'settings-header-icon',
      sidebarLink: 'sidebar-nav-link-integrations',
    },
  },
  {
    path: '/settings/notifications',
    title: 'Notification Settings',
    description: 'Configure notification inbox and delivery preferences.',
    icon: Bell,
    navigationGroup: 'Settings',
    documentTitle: titleFor('Notification Settings'),
    testIds: {
      headerTitle: 'settings-header-title',
      headerIcon: 'settings-header-icon',
      sidebarLink: 'sidebar-nav-link-notifications',
    },
  },
  {
    path: '/settings/operations',
    title: 'Operations Settings',
    description: 'Control runtime operations, workers, and diagnostics.',
    icon: Settings,
    navigationGroup: 'Settings',
    documentTitle: titleFor('Operations Settings'),
    testIds: {
      headerTitle: 'settings-header-title',
      headerIcon: 'settings-header-icon',
      sidebarLink: 'sidebar-nav-link-operations',
    },
  },
  {
    path: '/settings/skills',
    title: 'Skills Settings',
    description: 'Review assistant skills and execution boundaries.',
    icon: BrainCircuit,
    navigationGroup: 'Settings',
    documentTitle: titleFor('Skills Settings'),
    testIds: {
      headerTitle: 'settings-header-title',
      headerIcon: 'settings-header-icon',
      sidebarLink: 'sidebar-nav-link-skills',
    },
  },
  {
    path: '/settings/storage',
    title: 'Storage Settings',
    description: 'Manage data directories, backups, exports, and restores.',
    icon: Database,
    navigationGroup: 'Settings',
    documentTitle: titleFor('Storage Settings'),
    testIds: {
      headerTitle: 'settings-header-title',
      headerIcon: 'settings-header-icon',
      sidebarLink: 'sidebar-nav-link-storage',
    },
  },
  {
    path: '/help-center',
    title: 'Help Center',
    description: 'Find Cabinet help articles and workflow guidance.',
    icon: HelpCircle,
    navigationGroup: 'Other',
    documentTitle: titleFor('Help Center'),
    testIds: {
      headerTitle: 'help-center-header-title',
      headerIcon: 'help-center-header-icon',
      sidebarLink: 'sidebar-nav-link-help-center',
    },
  },
  {
    path: '/errors/*',
    title: 'Error',
    description: 'Recover from authenticated route errors.',
    icon: AlertTriangle,
    navigationGroup: 'System',
    documentTitle: titleFor('Error'),
    testIds: {
      headerTitle: 'error-header-title',
      headerIcon: 'error-header-icon',
    },
  },
  {
    path: '/404',
    title: 'Not Found',
    description: 'Recover from a missing Cabinet route.',
    icon: CircleHelp,
    navigationGroup: 'System',
    documentTitle: titleFor('Not Found'),
    testIds: {
      headerTitle: 'not-found-header-title',
      headerIcon: 'not-found-header-icon',
    },
  },
]

function normalizePath(pathname: string) {
  const path = pathname.split(/[?#]/)[0]?.trim() || '/'
  if (path === '') {
    return '/'
  }
  if (path.length > 1 && path.endsWith('/')) {
    return path.slice(0, -1)
  }
  return path
}

export function getRouteMetadata(pathname: string) {
  const normalized = normalizePath(pathname)
  return authenticatedRouteMetadata.find((metadata) => {
    if (metadata.path === '/errors/*') {
      return normalized === '/errors' || normalized.startsWith('/errors/')
    }

    return (
      normalized === metadata.path || (metadata.aliases ?? []).includes(normalized)
    )
  })
}
