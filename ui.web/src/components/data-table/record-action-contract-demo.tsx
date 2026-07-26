import { useRef, useState } from 'react'
import { Pencil } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { RecordActionMenu } from './record-action-menu'
import { RecordDestructiveConfirmDialog } from './record-destructive-confirm-dialog'
import { RecordEditDialog } from './record-edit-dialog'
import type { RecordDestructiveAction } from './record-destructive-confirm-dialog'

const demoRecord = {
  id: 'demo-charizard',
  name: 'Charizard Base Set',
  category: 'Trading card',
  state: 'Active',
}

export function RecordActionContractDemo() {
  const menuReturnRef = useRef<HTMLButtonElement | null>(null)
  const [editOpen, setEditOpen] = useState(false)
  const [destroyOpen, setDestroyOpen] = useState(false)
  const [destructiveAction, setDestructiveAction] =
    useState<RecordDestructiveAction>('delete')
  const [name, setName] = useState(demoRecord.name)

  return (
    <section
      className='space-y-4 rounded-md border p-4'
      aria-label='Record action contract demo'
    >
      <div className='space-y-1'>
        <h2 className='text-base font-semibold'>Record action contract demo</h2>
        <p className='text-sm text-muted-foreground'>
          Shared table-row menu, edit dialog, and destructive confirmation
          fixtures for page migrations.
        </p>
      </div>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Name</TableHead>
            <TableHead>Category</TableHead>
            <TableHead>Status</TableHead>
            <TableHead className='w-10'>
              <span className='sr-only'>Actions</span>
            </TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow>
            <TableCell>{demoRecord.name}</TableCell>
            <TableCell>{demoRecord.category}</TableCell>
            <TableCell>{demoRecord.state}</TableCell>
            <TableCell className='text-end'>
              <RecordActionMenu
                recordLabel={demoRecord.name}
                actions={[
                  {
                    id: 'open',
                    label: 'Open details',
                    kind: 'open',
                    onSelect: () => undefined,
                  },
                  {
                    id: 'edit',
                    label: 'Edit',
                    kind: 'edit',
                    onSelect: () => setEditOpen(true),
                  },
                  {
                    id: 'duplicate',
                    label: 'Duplicate',
                    kind: 'duplicate',
                    disabledReason: 'Requires write permission',
                    onSelect: () => undefined,
                  },
                  {
                    id: 'archive',
                    label: 'Archive',
                    kind: 'archive',
                    onSelect: () => {
                      setDestructiveAction('archive')
                      setDestroyOpen(true)
                    },
                  },
                  {
                    id: 'delete',
                    label: 'Delete',
                    kind: 'delete',
                    onSelect: () => {
                      setDestructiveAction('delete')
                      setDestroyOpen(true)
                    },
                  },
                  {
                    id: 'permanent-delete',
                    label: 'Permanent delete',
                    kind: 'permanent-delete',
                    onSelect: () => {
                      setDestructiveAction('permanent-delete')
                      setDestroyOpen(true)
                    },
                  },
                ]}
                ref={menuReturnRef}
              />
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
      <RecordEditDialog
        open={editOpen}
        onOpenChange={setEditOpen}
        returnFocusRef={menuReturnRef}
        title={`Edit ${demoRecord.name}`}
        description='Server and field errors remain actionable.'
        icon={Pencil}
        isDirty={name !== demoRecord.name}
        onSubmit={() => setEditOpen(false)}
      >
        <div className='space-y-2'>
          <Label htmlFor='record-action-demo-name'>Name</Label>
          <Input
            id='record-action-demo-name'
            value={name}
            onChange={(event) => setName(event.target.value)}
            autoFocus
          />
        </div>
      </RecordEditDialog>
      <RecordDestructiveConfirmDialog
        open={destroyOpen}
        onOpenChange={setDestroyOpen}
        returnFocusRef={menuReturnRef}
        action={destructiveAction}
        recordLabel={demoRecord.name}
        consequence={
          destructiveAction === 'permanent-delete'
            ? 'This cannot be undone.'
            : 'This table keeps record-level destructive choices separate from bulk actions.'
        }
        onConfirm={() => setDestroyOpen(false)}
      />
    </section>
  )
}
