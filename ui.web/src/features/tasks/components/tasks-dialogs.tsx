import { useEffect, useRef } from 'react'
import { showSubmittedData } from '@/lib/show-submitted-data'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { type Task } from '../data/schema'
import { TasksImportDialog } from './tasks-import-dialog'
import {
  TasksMutateDrawer,
  type WishlistEntryDraft,
} from './tasks-mutate-drawer'

export type TasksDialogType = 'create' | 'update' | 'delete' | 'import'

type TasksDialogsProps = {
  routePath: '/_authenticated/inventory/' | '/_authenticated/wishlist/'
  open: TasksDialogType | null
  setOpen: (open: TasksDialogType | null) => void
  currentRow: Task | null
  setCurrentRow: (task: Task | null) => void
  navigationRows?: Task[]
  onWishlistSubmit?: (
    draft: WishlistEntryDraft,
    currentRow?: Task
  ) => Promise<void>
  onWishlistDelete?: (currentRow: Task) => Promise<void>
  onWishlistImport?: (entries: WishlistEntryDraft[]) => Promise<void>
  isWishlistMutating?: boolean
  wishlistItemTypeOptions?: string[]
  wishlistPackagingGradeOptions?: string[]
  wishlistConditionOptions?: string[]
}

export function TasksDialogs({
  routePath,
  open,
  setOpen,
  currentRow,
  setCurrentRow,
  navigationRows = [],
  onWishlistSubmit,
  onWishlistDelete,
  onWishlistImport,
  isWishlistMutating = false,
  wishlistItemTypeOptions = [],
  wishlistPackagingGradeOptions = [],
  wishlistConditionOptions = [],
}: TasksDialogsProps) {
  const isWishlistRoute = routePath === '/_authenticated/wishlist/'
  const clearCurrentRowTimerRef = useRef<number | null>(null)

  const cancelPendingClearCurrentRow = () => {
    if (clearCurrentRowTimerRef.current !== null) {
      window.clearTimeout(clearCurrentRowTimerRef.current)
      clearCurrentRowTimerRef.current = null
    }
  }

  const clearCurrentRow = () => {
    cancelPendingClearCurrentRow()
    clearCurrentRowTimerRef.current = window.setTimeout(() => {
      setCurrentRow(null)
      clearCurrentRowTimerRef.current = null
    }, 500)
  }

  const currentRowIndex = currentRow
    ? navigationRows.findIndex((task) => task.id === currentRow.id)
    : -1
  const canNavigatePrevious = currentRowIndex > 0
  const canNavigateNext =
    currentRowIndex >= 0 && currentRowIndex < navigationRows.length - 1

  const navigateCurrentRow = (offset: number) => {
    if (currentRowIndex < 0) {
      return
    }
    const nextRow = navigationRows[currentRowIndex + offset]
    if (!nextRow) {
      return
    }
    cancelPendingClearCurrentRow()
    setCurrentRow(nextRow)
    setOpen('update')
  }

  useEffect(() => {
    if (open !== null) {
      cancelPendingClearCurrentRow()
    }
  }, [open])

  useEffect(() => cancelPendingClearCurrentRow, [])

  return (
    <>
      <TasksMutateDrawer
        key='task-create'
        open={open === 'create'}
        onOpenChange={(isOpen) => {
          if (isOpen) {
            cancelPendingClearCurrentRow()
          }
          setOpen(isOpen ? 'create' : null)
        }}
        routePath={routePath}
        onWishlistSubmit={onWishlistSubmit}
        isLoading={isWishlistMutating}
        wishlistItemTypeOptions={wishlistItemTypeOptions}
        wishlistPackagingGradeOptions={wishlistPackagingGradeOptions}
        wishlistConditionOptions={wishlistConditionOptions}
      />

      <TasksImportDialog
        key='tasks-import'
        open={open === 'import'}
        onOpenChange={(isOpen) => {
          if (isOpen) {
            cancelPendingClearCurrentRow()
          }
          setOpen(isOpen ? 'import' : null)
        }}
        routePath={routePath}
        onWishlistImport={onWishlistImport}
        isLoading={isWishlistMutating}
      />

      {currentRow && (
        <>
          <TasksMutateDrawer
            key={
              isWishlistRoute
                ? 'wishlist-update-panel'
                : `task-update-${currentRow.id}`
            }
            open={open === 'update'}
            onOpenChange={(isOpen) => {
              if (isOpen) {
                cancelPendingClearCurrentRow()
              }
              setOpen(isOpen ? 'update' : null)
              if (!isOpen) {
                clearCurrentRow()
              }
            }}
            currentRow={currentRow}
            routePath={routePath}
            onWishlistSubmit={onWishlistSubmit}
            isLoading={isWishlistMutating}
            canNavigatePrevious={canNavigatePrevious}
            canNavigateNext={canNavigateNext}
            onNavigatePrevious={() => navigateCurrentRow(-1)}
            onNavigateNext={() => navigateCurrentRow(1)}
            wishlistItemTypeOptions={wishlistItemTypeOptions}
            wishlistPackagingGradeOptions={wishlistPackagingGradeOptions}
            wishlistConditionOptions={wishlistConditionOptions}
          />

          <ConfirmDialog
            key='task-delete'
            destructive
            open={open === 'delete'}
            onOpenChange={(isOpen) => {
              if (isOpen) {
                cancelPendingClearCurrentRow()
              }
              setOpen(isOpen ? 'delete' : null)
              if (!isOpen) {
                clearCurrentRow()
              }
            }}
            handleConfirm={() => {
              if (isWishlistRoute && onWishlistDelete) {
                void onWishlistDelete(currentRow).then(() => {
                  setOpen(null)
                  clearCurrentRow()
                })
                return
              }
              setOpen(null)
              clearCurrentRow()
              showSubmittedData(
                currentRow,
                'The following task has been deleted:'
              )
            }}
            isLoading={isWishlistMutating}
            className='max-w-md'
            contentTestId='task-delete-dialog'
            cancelTestId='task-delete-cancel'
            confirmTestId='task-delete-confirm'
            title={
              isWishlistRoute
                ? `Delete this wishlist entry: ${currentRow.title} ?`
                : `Delete this task: ${currentRow.id} ?`
            }
            desc={
              isWishlistRoute ? (
                <>
                  You are about to delete the wishlist entry for{' '}
                  <strong>{currentRow.title}</strong>. <br />
                  This action cannot be undone.
                </>
              ) : (
                <>
                  You are about to delete a task with the ID{' '}
                  <strong>{currentRow.id}</strong>. <br />
                  This action cannot be undone.
                </>
              )
            }
            confirmText='Delete'
          />
        </>
      )}
    </>
  )
}
