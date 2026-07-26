import { DotsHorizontalIcon } from '@radix-ui/react-icons'
import {
  Archive,
  Copy,
  Eye,
  Pencil,
  RotateCcw,
  Trash2,
  type LucideIcon,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuShortcut,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { cn } from '@/lib/utils'

export type RecordActionKind =
  | 'view'
  | 'open'
  | 'edit'
  | 'duplicate'
  | 'archive'
  | 'delete'
  | 'restore'
  | 'permanent-delete'

export type RecordActionDefinition = {
  id: string
  label: string
  kind: RecordActionKind
  available?: boolean
  disabledReason?: string
  onSelect: () => void
}

type RecordActionMenuProps = {
  recordLabel: string
  actions: RecordActionDefinition[]
  className?: string
  triggerClassName?: string
  testId?: string
}

const actionOrder: Record<RecordActionKind, number> = {
  view: 10,
  open: 10,
  edit: 20,
  duplicate: 30,
  archive: 40,
  delete: 40,
  restore: 50,
  'permanent-delete': 60,
}

const actionIcons: Record<RecordActionKind, LucideIcon> = {
  view: Eye,
  open: Eye,
  edit: Pencil,
  duplicate: Copy,
  archive: Archive,
  delete: Trash2,
  restore: RotateCcw,
  'permanent-delete': Trash2,
}

function stopRowEvent(event: {
  stopPropagation: () => void
  preventDefault?: () => void
}) {
  event.stopPropagation()
}

function visibleActions(actions: RecordActionDefinition[]) {
  return actions
    .filter((action) => action.available !== false)
    .map((action, index) => ({ action, index }))
    .sort((left, right) => {
      const orderDelta =
        actionOrder[left.action.kind] - actionOrder[right.action.kind]
      return orderDelta === 0 ? left.index - right.index : orderDelta
    })
}

export function RecordActionMenu({
  recordLabel,
  actions,
  className,
  triggerClassName,
  testId = 'record-action-menu',
}: RecordActionMenuProps) {
  const triggerLabel = `Open actions for ${recordLabel}`
  const menuActions = visibleActions(actions)

  return (
    <DropdownMenu modal={false}>
      <Tooltip>
        <TooltipTrigger asChild>
          <DropdownMenuTrigger asChild>
            <Button
              type='button'
              variant='ghost'
              size='icon'
              className={cn(
                'h-8 w-8 p-0 data-[state=open]:bg-muted',
                triggerClassName
              )}
              data-testid={`${testId}-trigger`}
              aria-label={triggerLabel}
              title={triggerLabel}
              onPointerDown={stopRowEvent}
              onClick={stopRowEvent}
              onDoubleClick={stopRowEvent}
            >
              <DotsHorizontalIcon className='h-4 w-4' aria-hidden='true' />
            </Button>
          </DropdownMenuTrigger>
        </TooltipTrigger>
        <TooltipContent>{triggerLabel}</TooltipContent>
      </Tooltip>
      <DropdownMenuContent
        align='end'
        className={cn('w-56', className)}
        data-testid={`${testId}-content`}
        onPointerDown={stopRowEvent}
        onClick={stopRowEvent}
        onDoubleClick={stopRowEvent}
      >
        {menuActions.map(({ action }) => {
          const Icon = actionIcons[action.kind]
          const disabled = Boolean(action.disabledReason)
          return (
            <DropdownMenuItem
              key={action.id}
              data-testid={`${testId}-item-${action.id}`}
              disabled={disabled}
              variant={
                action.kind === 'delete' ||
                action.kind === 'permanent-delete'
                  ? 'destructive'
                  : 'default'
              }
              onClick={stopRowEvent}
              onSelect={(event) => {
                event.stopPropagation()
                if (disabled) {
                  event.preventDefault()
                  return
                }
                action.onSelect()
              }}
            >
              <span className='min-w-0 flex-1'>
                <span className='block truncate'>{action.label}</span>
                {action.disabledReason ? (
                  <span className='block truncate text-xs text-muted-foreground'>
                    {action.disabledReason}
                  </span>
                ) : null}
              </span>
              <DropdownMenuShortcut>
                <Icon size={16} />
              </DropdownMenuShortcut>
            </DropdownMenuItem>
          )
        })}
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
