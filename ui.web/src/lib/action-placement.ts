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

export function getActionPlacementRegion(id: string) {
  return actionPlacementRegions.find((region) => region.id === id)
}

export function isPageActionRegion(action: ActionPlacementDefinition) {
  return action.placement === 'page-header'
}

export function buildActionPlacementChecklist(
  actions: ActionPlacementDefinition[]
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
