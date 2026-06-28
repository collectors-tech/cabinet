export type UiTargetType =
  | 'nav'
  | 'button'
  | 'field'
  | 'table-row'
  | 'panel'
  | 'status'

export type UiTarget = {
  id: string
  route: string
  label: string
  description: string
  selector: string
  targetType: UiTargetType
  visibilityRequirement: string
  scrollBehaviour: ScrollLogicalPosition
  safeActions: string[]
  fallbackInstruction: string
  calloutPlacement: 'top' | 'right' | 'bottom' | 'left'
}

export type UiGuidanceRequest = {
  targetId: string
  title?: string
  instruction?: string
}

export const uiGuidanceEventName = 'cabinet:ui-guidance'
export const uiGuidanceClearEventName = 'cabinet:ui-guidance-clear'

export const uiTargets: UiTarget[] = [
  {
    id: 'inventory.surface',
    route: '/inventory',
    label: 'Inventory surface',
    description: 'The main Inventory workspace used to find and edit items.',
    selector: '[data-testid="inventory-page"]',
    targetType: 'panel',
    visibilityRequirement: 'Authenticated Inventory route is open.',
    scrollBehaviour: 'start',
    safeActions: ['highlight', 'scroll'],
    fallbackInstruction: 'Open Inventory from the sidebar, then retry.',
    calloutPlacement: 'bottom',
  },
  {
    id: 'inventory.item.row',
    route: '/inventory',
    label: 'Inventory item row',
    description:
      'A selectable inventory table row that opens item detail or edit actions.',
    selector:
      '[data-testid^="inventory-item-row-"], [data-testid="task-row-actions-trigger"]',
    targetType: 'table-row',
    visibilityRequirement: 'Inventory contains at least one visible item row.',
    scrollBehaviour: 'center',
    safeActions: ['highlight', 'scroll'],
    fallbackInstruction:
      'Create or import an inventory item before starting this walkthrough.',
    calloutPlacement: 'right',
  },
  {
    id: 'inventory.item.actions',
    route: '/inventory',
    label: 'Item actions menu',
    description:
      'The row actions menu that exposes edit, delete, and related item commands.',
    selector: '[data-testid="task-row-actions-trigger"]',
    targetType: 'button',
    visibilityRequirement: 'An inventory item row is visible.',
    scrollBehaviour: 'center',
    safeActions: ['highlight', 'scroll', 'focus'],
    fallbackInstruction: 'Find the item row, then open its actions menu.',
    calloutPlacement: 'left',
  },
  {
    id: 'inventory.item.editor',
    route: '/inventory',
    label: 'Item editor panel',
    description: 'The edit panel for changing inventory item fields.',
    selector: '[data-testid="inventory-edit-panel"]',
    targetType: 'panel',
    visibilityRequirement: 'The inventory item editor is open.',
    scrollBehaviour: 'center',
    safeActions: ['highlight', 'scroll'],
    fallbackInstruction: 'Open an inventory item row action and choose Edit.',
    calloutPlacement: 'left',
  },
  {
    id: 'inventory.item.title',
    route: '/inventory',
    label: 'Item title field',
    description: 'The editable title field for the selected inventory item.',
    selector: '[data-testid="inventory-edit-title"]',
    targetType: 'field',
    visibilityRequirement: 'The inventory item editor is open.',
    scrollBehaviour: 'center',
    safeActions: ['highlight', 'scroll', 'focus'],
    fallbackInstruction: 'Open the inventory editor, then use the Title field.',
    calloutPlacement: 'left',
  },
  {
    id: 'inventory.item.status',
    route: '/inventory',
    label: 'Item status field',
    description:
      'The editable status selector for the selected inventory item.',
    selector: '[data-testid="inventory-edit-status"]',
    targetType: 'field',
    visibilityRequirement: 'The inventory item editor is open.',
    scrollBehaviour: 'center',
    safeActions: ['highlight', 'scroll', 'focus'],
    fallbackInstruction:
      'Open the inventory editor, then use the Status field.',
    calloutPlacement: 'left',
  },
  {
    id: 'inventory.item.category',
    route: '/inventory',
    label: 'Item category field',
    description: 'The editable category field for the selected inventory item.',
    selector: '[data-testid="inventory-edit-category"]',
    targetType: 'field',
    visibilityRequirement: 'The inventory item editor is open.',
    scrollBehaviour: 'center',
    safeActions: ['highlight', 'scroll', 'focus'],
    fallbackInstruction:
      'Open the inventory editor, then use the Category field.',
    calloutPlacement: 'left',
  },
  {
    id: 'inventory.item.save',
    route: '/inventory',
    label: 'Save item changes',
    description:
      'The save button that commits confirmed inventory editor changes.',
    selector: '[data-testid="inventory-edit-save"]',
    targetType: 'button',
    visibilityRequirement: 'The inventory item editor is open and valid.',
    scrollBehaviour: 'center',
    safeActions: ['highlight', 'scroll'],
    fallbackInstruction:
      'Review the editor fields, then save after confirmation.',
    calloutPlacement: 'top',
  },
  {
    id: 'inventory.item.cancel',
    route: '/inventory',
    label: 'Cancel item editing',
    description:
      'The cancel control that exits the inventory editor without applying changes.',
    selector: '[data-testid="inventory-edit-cancel"]',
    targetType: 'button',
    visibilityRequirement: 'The inventory item editor is open.',
    scrollBehaviour: 'center',
    safeActions: ['highlight', 'scroll'],
    fallbackInstruction: 'Close the editor to cancel this editing step.',
    calloutPlacement: 'top',
  },
]

export function findUiTarget(targetId: string) {
  return uiTargets.find((target) => target.id === targetId) ?? null
}

export function uiTargetsForRoute(pathname: string) {
  return uiTargets.filter((target) => pathname.startsWith(target.route))
}

export function requestUiGuidance(detail: UiGuidanceRequest) {
  window.dispatchEvent(new CustomEvent(uiGuidanceEventName, { detail }))
}

export function clearUiGuidance() {
  window.dispatchEvent(new CustomEvent(uiGuidanceClearEventName))
}
