import {
  clearUiGuidance,
  findUiTarget,
  requestUiGuidance,
} from '@/lib/ui-target-registry'

export type ShellCommandStatus =
  | 'queued'
  | 'running'
  | 'success'
  | 'failure'
  | 'skipped'

export type ShellCommandType =
  | 'navigate.open_surface'
  | 'ui.highlight_target'
  | 'ui.scroll_to_target'
  | 'ui.focus_field'
  | 'ui.wait_for_user_action'
  | 'chat.action.preview'
  | 'chat.action.confirm_apply'
  | 'ui.clear_guidance'
  | 'walkthrough.cancel'

export type ShellCommand = {
  id: string
  type: ShellCommandType
  route?: string
  targetId?: string
  title?: string
  instruction?: string
}

export type ShellCommandEvent = {
  id: string
  type: ShellCommandType
  status: ShellCommandStatus
  route?: string
  targetId?: string
  message: string
  timestamp: string
}

export type ShellCommandDispatcher = {
  navigate: (route: string) => Promise<void> | void
  emit: (event: ShellCommandEvent) => void
}

const allowedRoutes = new Set([
  '/dashboard',
  '/inventory',
  '/wishlist',
  '/collections',
  '/media',
  '/discoveries',
  '/scanner',
  '/purchases',
  '/integrations',
  '/chats',
  '/inbox',
  '/settings',
  '/settings/profile',
  '/settings/account',
  '/settings/appearance',
  '/settings/storage',
  '/settings/skills',
  '/settings/display',
])

function commandEvent(
  command: ShellCommand,
  status: ShellCommandStatus,
  message: string
): ShellCommandEvent {
  return {
    id: command.id,
    type: command.type,
    status,
    route: command.route,
    targetId: command.targetId,
    message,
    timestamp: new Date().toISOString(),
  }
}

function isAllowedRoute(route: string) {
  return allowedRoutes.has(route.trim().replace(/\/+$/, '') || '/')
}

export async function dispatchShellCommand(
  command: ShellCommand,
  dispatcher: ShellCommandDispatcher
) {
  dispatcher.emit(commandEvent(command, 'queued', 'Command queued'))
  dispatcher.emit(commandEvent(command, 'running', 'Command running'))

  try {
    switch (command.type) {
      case 'navigate.open_surface': {
        const route = command.route?.trim()
        if (!route || !isAllowedRoute(route)) {
          throw new Error('Route is not in the Cabinet app-control allowlist')
        }
        await dispatcher.navigate(route)
        dispatcher.emit(
          commandEvent(command, 'success', `Opened ${route} without mutation`)
        )
        return
      }
      case 'ui.highlight_target':
      case 'ui.scroll_to_target':
      case 'ui.focus_field': {
        const targetId = command.targetId?.trim()
        if (!targetId || !findUiTarget(targetId)) {
          throw new Error('UI target is not registered for guided walkthroughs')
        }
        requestUiGuidance({
          targetId,
          title: command.title,
          instruction: command.instruction,
        })
        dispatcher.emit(
          commandEvent(command, 'success', `Highlighted ${targetId}`)
        )
        return
      }
      case 'ui.clear_guidance':
      case 'walkthrough.cancel':
        clearUiGuidance()
        dispatcher.emit(commandEvent(command, 'success', 'Guidance cleared'))
        return
      case 'ui.wait_for_user_action':
        dispatcher.emit(
          commandEvent(command, 'success', 'Paused for user checkpoint')
        )
        return
      case 'chat.action.preview':
      case 'chat.action.confirm_apply':
        dispatcher.emit(
          commandEvent(
            command,
            'skipped',
            'Show mode stops before preview, save, or apply'
          )
        )
        return
      default:
        dispatcher.emit(commandEvent(command, 'skipped', 'Unsupported command'))
    }
  } catch (error) {
    dispatcher.emit(
      commandEvent(
        command,
        'failure',
        error instanceof Error ? error.message : 'Command failed'
      )
    )
  }
}
