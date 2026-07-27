export const actionPlacementRegionOrder = [
  'page-header',
  'table-toolbar',
  'bulk-toolbar',
  'row-menu',
  'dialog-footer',
  'shell-utility',
] as const

export type ActionPlacementRegionID =
  (typeof actionPlacementRegionOrder)[number]

export type ActionPlacementScope =
  | 'whole-page'
  | 'table-or-list'
  | 'selected-records'
  | 'single-record'
  | 'active-dialog'
  | 'shell'

export type ActionPlacementRegion = {
  id: ActionPlacementRegionID
  label: string
  scope: ActionPlacementScope
  owns: string
}

export type ActionPlacementKind =
  | 'create'
  | 'import'
  | 'export'
  | 'refresh'
  | 'run'
  | 'filter'
  | 'sort'
  | 'view'
  | 'bulk'
  | 'record'
  | 'dialog-cancel'
  | 'dialog-confirm'
  | 'shell'

export type ActionPlacementDefinition = {
  id: string
  label: string
  placement: ActionPlacementRegionID
  kind: ActionPlacementKind
  priority?: 'primary' | 'secondary'
}

export type PageHeaderActionLayoutOptions = {
  viewport: 'wide' | 'narrow'
  maxVisibleActions: number
}

export type PageHeaderActionLayout = {
  visibleActions: ActionPlacementDefinition[]
  overflowActions: ActionPlacementDefinition[]
  shellUtilityActions: ActionPlacementDefinition[]
  overflowMenu?: {
    ariaLabel: 'More page actions'
    testId: 'page-header-action-overflow'
    keyboardReachable: true
  }
}

export type ActionBoundaryPlacement = {
  placement: ActionPlacementRegionID
  includedInPageOverflow: boolean
  ownsGlobalChrome: boolean
}

export type ShellActionBoundary = {
  shellUtilityActionIds: string[]
  pageActionIds: string[]
  nonPageActionIds: string[]
  shellUtilityPlacement: ActionBoundaryPlacement
  pageActionPlacement: ActionBoundaryPlacement
}

export type RouteActionRegionContract = {
  route: string
  surface: string
  pageActionRegionTestId?: string
  wholePageActionIds: string[]
}

export const actionPlacementRegions: ActionPlacementRegion[] = [
  {
    id: 'page-header',
    label: 'Page header',
    scope: 'whole-page',
    owns: 'Whole-page create, add, import, export, refresh, run, invite, backup, and restore actions.',
  },
  {
    id: 'table-toolbar',
    label: 'Table/list toolbar',
    scope: 'table-or-list',
    owns: 'Search, filters, sort, view switches, and table-scoped refresh controls.',
  },
  {
    id: 'bulk-toolbar',
    label: 'Bulk action toolbar',
    scope: 'selected-records',
    owns: 'Actions that operate on the current table or list selection.',
  },
  {
    id: 'row-menu',
    label: 'Row action menu',
    scope: 'single-record',
    owns: 'Single-record view, edit, run, validate, archive, restore, delete, and handoff operations.',
  },
  {
    id: 'dialog-footer',
    label: 'Dialog footer',
    scope: 'active-dialog',
    owns: 'Cancel, confirm, apply, save, and destructive confirmation controls for the active dialog.',
  },
  {
    id: 'shell-utility',
    label: 'Shell utility',
    scope: 'shell',
    owns: 'Global shell utilities that are not page actions.',
  },
]

export const shellUtilityActionIds = [
  'language-switch',
  'theme-switch',
  'configuration',
  'profile-menu',
  'sidebar-toggle',
] as const

export const routeActionRegionContracts: RouteActionRegionContract[] = [
  {
    route: '/dashboard',
    surface: 'Dashboard',
    wholePageActionIds: [],
  },
  {
    route: '/inventory',
    surface: 'Inventory',
    pageActionRegionTestId: 'inventory-global-header-actions',
    wholePageActionIds: [
      'inventory-new-item',
      'inventory-import',
      'inventory-export',
    ],
  },
  {
    route: '/collections',
    surface: 'Collections',
    pageActionRegionTestId: 'collections-global-header-actions',
    wholePageActionIds: ['collections-new'],
  },
  {
    route: '/wishlist',
    surface: 'Wishlist',
    pageActionRegionTestId: 'wishlist-global-header-actions',
    wholePageActionIds: ['wishlist-add', 'wishlist-import', 'wishlist-export'],
  },
  {
    route: '/media',
    surface: 'Media',
    pageActionRegionTestId: 'media-global-header-actions',
    wholePageActionIds: ['media-upload'],
  },
  {
    route: '/purchases',
    surface: 'Purchases',
    pageActionRegionTestId: 'purchases-global-header-actions',
    wholePageActionIds: ['purchases-new', 'purchases-export'],
  },
  {
    route: '/integrations',
    surface: 'Integrations',
    wholePageActionIds: [],
  },
  {
    route: '/chats',
    surface: 'Chats',
    wholePageActionIds: [],
  },
  {
    route: '/inbox',
    surface: 'Inbox',
    wholePageActionIds: [],
  },
  {
    route: '/discoveries',
    surface: 'Discoveries',
    wholePageActionIds: [],
  },
  {
    route: '/reports',
    surface: 'Reports',
    pageActionRegionTestId: 'reports-global-header-actions',
    wholePageActionIds: ['reports-refresh', 'reports-export'],
  },
  {
    route: '/scanner',
    surface: 'Market Watch',
    pageActionRegionTestId: 'market-watch-global-header-actions',
    wholePageActionIds: ['market-watch-create', 'market-watch-run'],
  },
  {
    route: '/settings/profile',
    surface: 'Settings Profile',
    wholePageActionIds: [],
  },
  {
    route: '/settings/account',
    surface: 'Settings Account',
    wholePageActionIds: [],
  },
  {
    route: '/settings/appearance',
    surface: 'Settings Appearance',
    wholePageActionIds: [],
  },
  {
    route: '/settings/display',
    surface: 'Settings Display',
    wholePageActionIds: [],
  },
  {
    route: '/settings/billing',
    surface: 'Settings Billing',
    wholePageActionIds: [],
  },
  {
    route: '/settings/categories',
    surface: 'Settings Categories',
    wholePageActionIds: [],
  },
  {
    route: '/settings/integrations',
    surface: 'Settings Integrations',
    wholePageActionIds: [],
  },
  {
    route: '/settings/notifications',
    surface: 'Settings Notifications',
    wholePageActionIds: [],
  },
  {
    route: '/settings/operations',
    surface: 'Settings Operations',
    wholePageActionIds: [],
  },
  {
    route: '/settings/skills',
    surface: 'Settings Skills',
    wholePageActionIds: [],
  },
  {
    route: '/settings/storage',
    surface: 'Settings Storage',
    wholePageActionIds: [],
  },
  {
    route: '/users',
    surface: 'Users',
    pageActionRegionTestId: 'users-global-header-actions',
    wholePageActionIds: ['users-add'],
  },
  {
    route: '/help-center',
    surface: 'Help Center',
    wholePageActionIds: [],
  },
]

export function getActionPlacementRegion(id: string) {
  return actionPlacementRegions.find((region) => region.id === id)
}

export function getRouteActionRegionContract(route: string) {
  return routeActionRegionContracts.find(
    (contract) => contract.route === route
  )
}

export function isPageActionRegion(action: ActionPlacementDefinition) {
  return action.placement === 'page-header'
}

export function buildActionPlacementChecklist(
  actions: readonly ActionPlacementDefinition[]
) {
  return {
    pageActions: actions.filter((action) => action.placement === 'page-header'),
    toolbarActions: actions.filter(
      (action) => action.placement === 'table-toolbar'
    ),
    bulkActions: actions.filter((action) => action.placement === 'bulk-toolbar'),
    rowActions: actions.filter((action) => action.placement === 'row-menu'),
    dialogFooterActions: actions.filter(
      (action) => action.placement === 'dialog-footer'
    ),
    shellUtilityActions: actions.filter(
      (action) => action.placement === 'shell-utility'
    ),
  }
}

export function buildPageHeaderActionLayout(
  actions: readonly ActionPlacementDefinition[],
  options: PageHeaderActionLayoutOptions
): PageHeaderActionLayout {
  const pageActions = actions.filter(isPageActionRegion)
  const shellUtilityActions = actions.filter(
    (action) => action.placement === 'shell-utility'
  )
  const primaryActions = pageActions.filter(
    (action) => action.priority === 'primary'
  )
  const secondaryActions = pageActions.filter(
    (action) => action.priority !== 'primary'
  )
  const maxVisibleActions = Math.max(options.maxVisibleActions, 1)

  if (options.viewport === 'wide') {
    return {
      visibleActions: pageActions.slice(0, maxVisibleActions),
      overflowActions: pageActions.slice(maxVisibleActions),
      shellUtilityActions,
      overflowMenu:
        pageActions.length > maxVisibleActions
          ? {
              ariaLabel: 'More page actions',
              testId: 'page-header-action-overflow',
              keyboardReachable: true,
            }
          : undefined,
    }
  }

  const visibleActions = [
    ...primaryActions,
    ...secondaryActions.slice(0, Math.max(maxVisibleActions - primaryActions.length, 0)),
  ].slice(0, maxVisibleActions)
  const visibleIDs = new Set(visibleActions.map((action) => action.id))
  const overflowActions = pageActions.filter((action) => !visibleIDs.has(action.id))

  return {
    visibleActions,
    overflowActions,
    shellUtilityActions,
    overflowMenu:
      overflowActions.length > 0
        ? {
            ariaLabel: 'More page actions',
            testId: 'page-header-action-overflow',
            keyboardReachable: true,
          }
        : undefined,
  }
}

export function buildShellActionBoundary(
  actions: readonly ActionPlacementDefinition[]
): ShellActionBoundary {
  const shellUtilityActions = actions.filter(
    (action) => action.placement === 'shell-utility'
  )
  const pageActions = actions.filter(isPageActionRegion)
  const pageActionIDs = new Set(pageActions.map((action) => action.id))

  return {
    shellUtilityActionIds: shellUtilityActions.map((action) => action.id),
    pageActionIds: pageActions.map((action) => action.id),
    nonPageActionIds: actions
      .filter(
        (action) =>
          action.placement !== 'shell-utility' && !pageActionIDs.has(action.id)
      )
      .map((action) => action.id),
    shellUtilityPlacement: {
      placement: 'shell-utility',
      includedInPageOverflow: false,
      ownsGlobalChrome: true,
    },
    pageActionPlacement: {
      placement: 'page-header',
      includedInPageOverflow: true,
      ownsGlobalChrome: false,
    },
  }
}
