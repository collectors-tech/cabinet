import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { Heart } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Input } from '@/components/ui/input'
import { Separator } from '@/components/ui/separator'
import { ConfigDrawer } from '@/components/config-drawer'
import { LanguageSwitch } from '@/components/language-switch'
import { Header, HeaderTitle } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import {
  collectionKey,
  useWorkspaceCollections,
} from '@/features/collections/use-workspace-collections'
import { TasksDialogs, type TasksDialogType } from './components/tasks-dialogs'
import { type WishlistEntryDraft } from './components/tasks-mutate-drawer'
import { TasksTable } from './components/tasks-table'
import { type Task } from './data/schema'
import { tasks } from './data/tasks'

type TasksProps = {
  title?: string
  description?: string
  routePath?: '/_authenticated/inventory/' | '/_authenticated/wishlist/'
}

const WISHLIST_PLANNING_FOCUS_STORAGE_KEY = 'cabinet.wishlistPlanningFocus'

type WishlistPlanningFocus =
  | 'all'
  | 'high-priority'
  | 'below-target'
  | 'watchlist'

const wishlistPlanningFocusConfig: Array<{
  id: WishlistPlanningFocus
  label: string
  description: string
}> = [
  {
    id: 'all',
    label: 'All planned',
    description: 'Everything currently being tracked.',
  },
  {
    id: 'high-priority',
    label: 'High priority',
    description: 'High and critical items that need active attention.',
  },
  {
    id: 'below-target',
    label: 'Below target',
    description: 'Items currently priced at or below target.',
  },
  {
    id: 'watchlist',
    label: 'Steady watch',
    description: 'Items to monitor without immediate urgency.',
  },
]

function normalizeWishlistPlanningFocus(
  value: string | null | undefined
): WishlistPlanningFocus {
  switch (value) {
    case 'high-priority':
    case 'below-target':
    case 'watchlist':
      return value
    default:
      return 'all'
  }
}

function matchesWishlistPlanningFocus(
  task: Task,
  focus: WishlistPlanningFocus
): boolean {
  const priority = task.priority.trim().toLowerCase()
  const isBelowTarget = Boolean(task.belowTargetNow)
  switch (focus) {
    case 'high-priority':
      return priority === 'high' || priority === 'critical'
    case 'below-target':
      return isBelowTarget
    case 'watchlist':
      return !isBelowTarget && priority !== 'high' && priority !== 'critical'
    default:
      return true
  }
}

type WishlistItemPayload = {
  id?: string
  title?: string
  part_number?: string
  category?: string
  priority?: string
}

type WishlistPriceTrend = 'up' | 'steady' | 'down' | 'unknown'

type WishlistPricingSummary = {
  marketPrice?: number
  priceTrend: WishlistPriceTrend
}

type WishlistEntryPayload = {
  id?: string
  item_id?: string
  priority?: string
  notes?: string
  below_target_now?: boolean
  target_price?: number
  highlight_hit?: boolean
}

function normalizeWishlistPriority(raw: string) {
  const trimmed = raw.trim().toLowerCase()
  return trimmed || 'medium'
}

function normalizeTargetPrice(raw: string) {
  if (raw.trim() === '') {
    return 0
  }
  const parsed = Number(raw)
  if (Number.isNaN(parsed) || parsed < 0) {
    throw new Error('invalid_target_price')
  }
  return parsed
}

function inferWishlistPriceTrend(
  points: Array<{ latest?: number }> | undefined
): WishlistPriceTrend {
  const values = (points ?? [])
    .map((point) => point.latest)
    .filter((value): value is number => typeof value === 'number')

  if (values.length < 2) {
    return 'unknown'
  }

  const previous = values[values.length - 2] ?? 0
  const latest = values[values.length - 1] ?? 0
  const difference = latest - previous
  if (Math.abs(difference) < 0.01) {
    return 'steady'
  }
  return difference > 0 ? 'up' : 'down'
}

async function loadWishlistPricingSummary(
  itemID: string
): Promise<WishlistPricingSummary> {
  try {
    const [statsResponse, trendResponse] = await Promise.all([
      fetch(`/api/pricing/stats?item_id=${encodeURIComponent(itemID)}`),
      fetch(`/api/pricing/trend?item_id=${encodeURIComponent(itemID)}`),
    ])

    let marketPrice: number | undefined
    if (statsResponse.ok) {
      const statsPayload = (await statsResponse.json()) as { latest?: number }
      marketPrice =
        typeof statsPayload.latest === 'number' && statsPayload.latest > 0
          ? statsPayload.latest
          : undefined
    }

    let priceTrend: WishlistPriceTrend = 'unknown'
    if (trendResponse.ok) {
      const trendPayload = (await trendResponse.json()) as {
        points?: Array<{ latest?: number }>
      }
      priceTrend = inferWishlistPriceTrend(trendPayload.points)
    }

    return { marketPrice, priceTrend }
  } catch {
    return { priceTrend: 'unknown' }
  }
}

function buildWishlistCsv(tasksToExport: Task[]) {
  const escapeCell = (value: string | number | undefined) => {
    const text = String(value ?? '')
    if (/[",\n]/.test(text)) {
      return `"${text.split('"').join('""')}"`
    }
    return text
  }

  return [
    [
      'title',
      'part_number',
      'category',
      'priority',
      'notes',
      'target_price',
    ].join(','),
    ...tasksToExport.map((task) =>
      [
        escapeCell(task.title),
        escapeCell(task.partNumber),
        escapeCell(task.label),
        escapeCell(task.priority),
        escapeCell(task.notes),
        escapeCell(task.targetPrice ?? ''),
      ].join(',')
    ),
  ].join('\n')
}

function TasksHeaderActions({
  isWishlistRoute,
  onOpenCollectionCreate,
  onCreate,
  onImport,
}: {
  isWishlistRoute: boolean
  onOpenCollectionCreate: () => void
  onCreate: () => void
  onImport: () => void
}) {
  return (
    <div
      className='flex flex-wrap items-center justify-end gap-2'
      data-testid='wishlist-header-actions'
    >
      <Button
        type='button'
        data-testid='wishlist-new-action'
        onClick={onCreate}
      >
        New
      </Button>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button
            type='button'
            variant='outline'
            data-testid='wishlist-create-menu-trigger'
          >
            Create
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align='end'>
          <DropdownMenuItem
            data-testid='wishlist-create-menu-entry'
            onClick={onCreate}
          >
            New Wishlist Entry
          </DropdownMenuItem>
          <DropdownMenuItem
            data-testid='wishlist-create-menu-import'
            onClick={onImport}
          >
            Import Wishlist
          </DropdownMenuItem>
          {isWishlistRoute ? (
            <DropdownMenuItem
              data-testid='wishlist-create-menu-collection'
              onClick={onOpenCollectionCreate}
            >
              New Collection
            </DropdownMenuItem>
          ) : null}
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  )
}

export function Tasks({
  title = 'Tasks',
  description = '',
  routePath = '/_authenticated/inventory/',
}: TasksProps) {
  const {
    workspaceCollections,
    activeWorkspaceCollection,
    collectionItems,
    setActiveWorkspaceCollection,
    addCollection,
  } = useWorkspaceCollections()
  const [inlineCollectionInputOpen, setInlineCollectionInputOpen] =
    useState(false)
  const [inlineCollectionName, setInlineCollectionName] = useState('')
  const [
    inlineCollectionValidationMessage,
    setInlineCollectionValidationMessage,
  ] = useState('')
  const isWishlistRoute = routePath === '/_authenticated/wishlist/'
  const [tableData, setTableData] = useState<Task[]>(() =>
    isWishlistRoute ? [] : tasks
  )
  const tableDataRef = useRef(tableData)
  const [dialogOpen, setDialogOpen] = useState<TasksDialogType | null>(null)
  const [currentDialogRow, setCurrentDialogRow] = useState<Task | null>(null)
  const [wishlistActionItemID, setWishlistActionItemID] = useState<
    string | null
  >(null)
  const [isWishlistMutating, setIsWishlistMutating] = useState(false)
  const [wishlistPlanningFocus, setWishlistPlanningFocus] =
    useState<WishlistPlanningFocus>(() => {
      if (typeof window === 'undefined') {
        return 'all'
      }
      return normalizeWishlistPlanningFocus(
        window.localStorage.getItem(WISHLIST_PLANNING_FOCUS_STORAGE_KEY)
      )
    })
  const wishlistCollectionNormalizedRef = useRef(false)
  const loadWishlistData = useCallback(async () => {
    const [wishlistResponse, itemsResponse] = await Promise.all([
      fetch('/api/wishlist'),
      fetch('/api/items?status=wishlist'),
    ])
    if (!wishlistResponse.ok || !itemsResponse.ok) {
      throw new Error('wishlist_bootstrap_failed')
    }
    const wishlistPayload = (await wishlistResponse.json()) as {
      items?: WishlistEntryPayload[]
    }
    const itemsPayload = (await itemsResponse.json()) as {
      items?: WishlistItemPayload[]
    }
    const wishlistByItemID = new Map<string, WishlistEntryPayload>()
    ;(wishlistPayload.items ?? []).forEach((entry) => {
      const itemID = entry.item_id?.trim()
      if (!itemID) {
        return
      }
      wishlistByItemID.set(itemID, entry)
    })
    return Promise.all(
      (itemsPayload.items ?? []).map(async (item, index) => {
        const itemID = item.id?.trim() || `wishlist-item-${index + 1}`
        const wishlistEntry = wishlistByItemID.get(itemID)
        const pricingSummary = await loadWishlistPricingSummary(itemID)
        return {
          id: itemID,
          itemID: itemID,
          wishlistEntryID: wishlistEntry?.id?.trim(),
          title:
            item.title?.trim() ||
            item.part_number?.trim() ||
            `Wishlist item ${index + 1}`,
          partNumber: item.part_number?.trim(),
          status: wishlistEntry?.below_target_now ? 'discovered' : 'wishlist',
          label: item.category?.trim() || 'collection',
          priority:
            wishlistEntry?.priority?.trim() ||
            item.priority?.trim() ||
            'medium',
          notes: wishlistEntry?.notes?.trim(),
          belowTargetNow: Boolean(wishlistEntry?.below_target_now),
          targetPrice:
            typeof wishlistEntry?.target_price === 'number'
              ? wishlistEntry.target_price
              : undefined,
          marketPrice: pricingSummary.marketPrice,
          priceTrend: pricingSummary.priceTrend,
          highlightHit: Boolean(wishlistEntry?.highlight_hit),
        } satisfies Task
      })
    )
  }, [])

  useEffect(() => {
    tableDataRef.current = tableData
  }, [tableData])

  useEffect(() => {
    if (!isWishlistRoute) {
      setTableData(tasks)
      return
    }

    let cancelled = false
    setTableData([])
    void loadWishlistData()
      .then((mapped) => {
        if (!cancelled) {
          setTableData(mapped)
        }
      })
      .catch(() => {
        if (!cancelled) {
          setTableData([])
        }
      })

    return () => {
      cancelled = true
    }
  }, [isWishlistRoute, loadWishlistData])

  const handleWishlistMarkOwned = useCallback(
    async (task: Task) => {
      const wishlistEntryID = task.wishlistEntryID?.trim()
      if (!wishlistEntryID) {
        toast.error('Wishlist entry is missing transition metadata.')
        return
      }
      setWishlistActionItemID(task.id)
      try {
        const response = await fetch('/api/wishlist/convert-owned', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ id: wishlistEntryID }),
        })
        if (!response.ok) {
          throw new Error('failed_to_move_wishlist_item_to_owned')
        }
        const mapped = await loadWishlistData()
        setTableData(mapped)
        toast.success(`${task.title} moved to inventory.`)
      } catch {
        toast.error('Move to inventory failed. Try again.')
      } finally {
        setWishlistActionItemID(null)
      }
    },
    [loadWishlistData]
  )

  const refreshWishlistTable = useCallback(async () => {
    const mapped = await loadWishlistData()
    setTableData(mapped)
  }, [loadWishlistData])

  const saveWishlistDraft = useCallback(
    async (draft: WishlistEntryDraft, currentRow?: Task) => {
      const itemResponse = currentRow?.itemID
        ? await fetch(`/api/items/${encodeURIComponent(currentRow.itemID)}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
              title: draft.title,
              part_number: draft.partNumber,
              category: draft.category,
            }),
          })
        : await fetch('/api/items', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
              title: draft.title,
              part_number: draft.partNumber,
              category: draft.category,
              priority: normalizeWishlistPriority(draft.priority),
            }),
          })

      if (!itemResponse.ok) {
        throw new Error('wishlist_item_save_failed')
      }

      const savedItem = (await itemResponse.json()) as { id?: string }
      const itemID = currentRow?.itemID ?? savedItem.id?.trim()
      if (!itemID) {
        throw new Error('wishlist_item_id_missing')
      }

      const wishlistResponse = await fetch('/api/wishlist', {
        method: currentRow?.wishlistEntryID ? 'PUT' : 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          id: currentRow?.wishlistEntryID ?? undefined,
          item_id: itemID,
          priority: normalizeWishlistPriority(draft.priority),
          notes: draft.notes,
          target_price: normalizeTargetPrice(draft.targetPrice),
          highlight_hit: currentRow?.highlightHit ?? false,
        }),
      })

      if (!wishlistResponse.ok) {
        throw new Error('wishlist_entry_save_failed')
      }
    },
    []
  )

  const handleWishlistSubmit = useCallback(
    async (draft: WishlistEntryDraft, currentRow?: Task) => {
      setIsWishlistMutating(true)
      try {
        await saveWishlistDraft(draft, currentRow)
        await refreshWishlistTable()
        toast.success(
          currentRow
            ? `${draft.title} updated.`
            : `${draft.title} added to wishlist.`
        )
      } catch (error) {
        if (
          error instanceof Error &&
          error.message === 'invalid_target_price'
        ) {
          toast.error('Target price must be a positive number.')
        } else {
          toast.error('Wishlist save failed. Try again.')
        }
        throw error
      } finally {
        setIsWishlistMutating(false)
      }
    },
    [refreshWishlistTable, saveWishlistDraft]
  )

  const handleWishlistInlineUpdate = useCallback(
    async (
      task: Task,
      changes: { targetPrice?: number; priority?: string }
    ) => {
      const wishlistEntryID = task.wishlistEntryID?.trim()
      if (!wishlistEntryID) {
        toast.error('Wishlist entry is missing update metadata.')
        return
      }

      const currentTask =
        tableDataRef.current.find((candidate) => candidate.id === task.id) ??
        task
      const nextPriority = normalizeWishlistPriority(
        changes.priority ?? currentTask.priority
      )
      let nextTargetPrice = changes.targetPrice ?? currentTask.targetPrice ?? 0

      if (changes.targetPrice === undefined) {
        try {
          const latestWishlistResponse = await fetch('/api/wishlist')
          if (latestWishlistResponse.ok) {
            const latestWishlistPayload =
              (await latestWishlistResponse.json()) as {
                items?: WishlistEntryPayload[]
              }
            const latestEntry = (latestWishlistPayload.items ?? []).find(
              (entry) => entry.id?.trim() === wishlistEntryID
            )
            if (typeof latestEntry?.target_price === 'number') {
              nextTargetPrice = latestEntry.target_price
            }
          }
        } catch {
          // Keep the local row value if the latest entry cannot be read.
        }
      }

      const nextTableData = tableDataRef.current.map((candidate) =>
        candidate.id === task.id
          ? {
              ...candidate,
              priority: nextPriority,
              targetPrice: nextTargetPrice,
            }
          : candidate
      )
      tableDataRef.current = nextTableData
      setTableData((previous) =>
        previous.map((candidate) =>
          candidate.id === task.id
            ? {
                ...candidate,
                priority: nextPriority,
                targetPrice: nextTargetPrice,
              }
            : candidate
        )
      )
      setIsWishlistMutating(true)
      try {
        const requestBody: Record<string, unknown> = {
          id: wishlistEntryID,
          item_id: task.itemID,
          priority: nextPriority,
        }
        if (changes.targetPrice !== undefined) {
          requestBody.target_price = nextTargetPrice
        }

        const response = await fetch('/api/wishlist', {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(requestBody),
        })
        if (!response.ok) {
          throw new Error('wishlist_inline_update_failed')
        }
        await refreshWishlistTable()
        toast.success('Wishlist row updated.')
      } catch {
        toast.error('Wishlist row update failed. Try again.')
        throw new Error('wishlist_inline_update_failed')
      } finally {
        setIsWishlistMutating(false)
      }
    },
    [refreshWishlistTable]
  )

  const handleWishlistDelete = useCallback(
    async (task: Task) => {
      const wishlistEntryID = task.wishlistEntryID?.trim()
      if (!wishlistEntryID) {
        toast.error('Wishlist entry is missing delete metadata.')
        return
      }

      setIsWishlistMutating(true)
      try {
        const response = await fetch(
          `/api/wishlist?id=${encodeURIComponent(wishlistEntryID)}`,
          { method: 'DELETE' }
        )
        if (!response.ok) {
          throw new Error('wishlist_delete_failed')
        }
        await refreshWishlistTable()
        toast.success(`${task.title} removed from wishlist.`)
      } catch {
        toast.error('Wishlist delete failed. Try again.')
        throw new Error('wishlist_delete_failed')
      } finally {
        setIsWishlistMutating(false)
      }
    },
    [refreshWishlistTable]
  )

  const handleWishlistImport = useCallback(
    async (entries: WishlistEntryDraft[]) => {
      setIsWishlistMutating(true)
      try {
        for (const entry of entries) {
          await saveWishlistDraft(entry)
        }
        await refreshWishlistTable()
        toast.success(
          `Imported ${entries.length} wishlist entr${entries.length === 1 ? 'y' : 'ies'}.`
        )
      } catch (error) {
        if (
          error instanceof Error &&
          error.message === 'invalid_target_price'
        ) {
          toast.error('Target price must be a positive number.')
        } else {
          toast.error('Wishlist import failed. Try again.')
        }
        throw error
      } finally {
        setIsWishlistMutating(false)
      }
    },
    [refreshWishlistTable, saveWishlistDraft]
  )

  const handleWishlistBulkPriorityChange = useCallback(
    async (selectedTasks: Task[], priority: string) => {
      setIsWishlistMutating(true)
      try {
        for (const task of selectedTasks) {
          const wishlistEntryID = task.wishlistEntryID?.trim()
          if (!wishlistEntryID) {
            continue
          }
          const response = await fetch('/api/wishlist', {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
              id: wishlistEntryID,
              item_id: task.itemID,
              priority: normalizeWishlistPriority(priority),
              notes: task.notes ?? '',
              target_price: task.targetPrice ?? 0,
              highlight_hit: task.highlightHit ?? false,
            }),
          })
          if (!response.ok) {
            throw new Error('wishlist_bulk_priority_failed')
          }
        }
        await refreshWishlistTable()
        toast.success(
          `Updated priority for ${selectedTasks.length} wishlist entr${selectedTasks.length === 1 ? 'y' : 'ies'}.`
        )
      } catch {
        toast.error('Bulk priority update failed. Try again.')
        throw new Error('wishlist_bulk_priority_failed')
      } finally {
        setIsWishlistMutating(false)
      }
    },
    [refreshWishlistTable]
  )

  const handleWishlistBulkStatusChange = useCallback(
    async (selectedTasks: Task[], status: string) => {
      const belowTargetNow = status === 'discovered'
      setIsWishlistMutating(true)
      try {
        for (const task of selectedTasks) {
          const wishlistEntryID = task.wishlistEntryID?.trim()
          if (!wishlistEntryID) {
            continue
          }
          const response = await fetch('/api/wishlist', {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
              id: wishlistEntryID,
              item_id: task.itemID,
              priority: normalizeWishlistPriority(task.priority),
              notes: task.notes ?? '',
              target_price: task.targetPrice ?? 0,
              highlight_hit: task.highlightHit ?? false,
              below_target_now: belowTargetNow,
            }),
          })
          if (!response.ok) {
            throw new Error('wishlist_bulk_status_failed')
          }
        }
        await refreshWishlistTable()
        toast.success(
          `Updated watch status for ${selectedTasks.length} wishlist entr${selectedTasks.length === 1 ? 'y' : 'ies'}.`
        )
      } catch {
        toast.error('Bulk watch status update failed. Try again.')
        throw new Error('wishlist_bulk_status_failed')
      } finally {
        setIsWishlistMutating(false)
      }
    },
    [refreshWishlistTable]
  )

  const handleWishlistBulkDelete = useCallback(
    async (selectedTasks: Task[]) => {
      setIsWishlistMutating(true)
      try {
        for (const task of selectedTasks) {
          const wishlistEntryID = task.wishlistEntryID?.trim()
          if (!wishlistEntryID) {
            continue
          }
          const response = await fetch(
            `/api/wishlist?id=${encodeURIComponent(wishlistEntryID)}`,
            { method: 'DELETE' }
          )
          if (!response.ok) {
            throw new Error('wishlist_bulk_delete_failed')
          }
        }
        await refreshWishlistTable()
        toast.success(
          `Deleted ${selectedTasks.length} wishlist entr${selectedTasks.length === 1 ? 'y' : 'ies'}.`
        )
      } catch {
        toast.error('Bulk delete failed. Try again.')
        throw new Error('wishlist_bulk_delete_failed')
      } finally {
        setIsWishlistMutating(false)
      }
    },
    [refreshWishlistTable]
  )

  const handleWishlistExport = useCallback((selectedTasks: Task[]) => {
    const csv = buildWishlistCsv(selectedTasks)
    const blob = new Blob([csv], { type: 'text/csv;charset=utf-8' })
    const url = window.URL.createObjectURL(blob)
    const anchor = document.createElement('a')
    anchor.href = url
    anchor.download = 'wishlist-export.csv'
    document.body.appendChild(anchor)
    anchor.click()
    anchor.remove()
    window.URL.revokeObjectURL(url)
    toast.success(
      `Exported ${selectedTasks.length} wishlist entr${selectedTasks.length === 1 ? 'y' : 'ies'}.`
    )
  }, [])

  useEffect(() => {
    if (!isWishlistRoute || typeof window === 'undefined') {
      return
    }
    window.localStorage.setItem(
      WISHLIST_PLANNING_FOCUS_STORAGE_KEY,
      wishlistPlanningFocus
    )
  }, [isWishlistRoute, wishlistPlanningFocus])

  const wishlistPlanningSummary = useMemo(() => {
    if (!isWishlistRoute) {
      return []
    }
    return wishlistPlanningFocusConfig.map((focus) => ({
      ...focus,
      count: tableData.filter((task) =>
        matchesWishlistPlanningFocus(task, focus.id)
      ).length,
    }))
  }, [isWishlistRoute, tableData])

  const displayedData = useMemo(() => {
    if (!isWishlistRoute) {
      return tableData
    }

    const focusedData = tableData.filter((task) =>
      matchesWishlistPlanningFocus(task, wishlistPlanningFocus)
    )
    if (activeWorkspaceCollection === 'All Items') {
      return focusedData
    }

    const selectedCollectionItemIDs = new Set(
      collectionItems
        .filter((item) => item.collectionName === activeWorkspaceCollection)
        .map((item) => item.id)
    )

    return focusedData.filter((task) =>
      selectedCollectionItemIDs.has(task.itemID ?? task.id)
    )
  }, [
    activeWorkspaceCollection,
    collectionItems,
    isWishlistRoute,
    tableData,
    wishlistPlanningFocus,
  ])

  useEffect(() => {
    if (
      !isWishlistRoute ||
      wishlistCollectionNormalizedRef.current ||
      activeWorkspaceCollection === 'All Items' ||
      tableData.length === 0
    ) {
      return
    }

    wishlistCollectionNormalizedRef.current = true
    const wishlistItemIDs = new Set(
      tableData.map((task) => task.itemID ?? task.id)
    )
    const collectionHasWishlistRows = collectionItems.some(
      (item) =>
        item.collectionName === activeWorkspaceCollection &&
        wishlistItemIDs.has(item.id)
    )

    if (!collectionHasWishlistRows) {
      void setActiveWorkspaceCollection('All Items')
    }
  }, [
    activeWorkspaceCollection,
    collectionItems,
    isWishlistRoute,
    setActiveWorkspaceCollection,
    tableData,
  ])

  const wishlistCollectionFilter = isWishlistRoute ? (
    <div
      className='flex flex-wrap items-center gap-2'
      data-testid='wishlist-table-collection-filter'
    >
      <select
        className='h-8 min-w-[10rem] rounded-md border bg-background px-2 text-sm'
        data-testid='wishlist-table-collection-select'
        value={activeWorkspaceCollection}
        onChange={(event) => {
          void setActiveWorkspaceCollection(event.target.value)
        }}
      >
        {workspaceCollections.map((collection) => (
          <option
            key={collection}
            value={collection}
            data-testid={`wishlist-table-collection-option-${collectionKey(collection)}`}
          >
            {collection}
          </option>
        ))}
      </select>
      <span
        className='text-xs text-muted-foreground'
        data-testid='wishlist-table-collection-selected'
      >
        {activeWorkspaceCollection}
      </span>
      <Button
        type='button'
        variant='outline'
        className='h-8 px-2 text-xs'
        data-testid='wishlist-table-add-collection'
        onClick={() => {
          setInlineCollectionValidationMessage('')
          setInlineCollectionInputOpen((open) => !open)
        }}
      >
        + New Collection
      </Button>
      {inlineCollectionInputOpen ? (
        <div
          className='flex flex-wrap items-center gap-2'
          data-testid='wishlist-table-new-collection-form'
        >
          <Input
            className='h-8 w-[12rem]'
            data-testid='wishlist-table-new-collection-name'
            placeholder='Collection name'
            aria-invalid={inlineCollectionValidationMessage ? 'true' : 'false'}
            value={inlineCollectionName}
            onChange={(event) => {
              setInlineCollectionName(event.target.value)
              if (inlineCollectionValidationMessage) {
                setInlineCollectionValidationMessage('')
              }
            }}
          />
          <Button
            type='button'
            className='h-8 px-3'
            data-testid='wishlist-table-new-collection-save'
            onClick={async () => {
              const created = await addCollection(inlineCollectionName)
              if (!created) {
                setInlineCollectionValidationMessage(
                  'Collection name is required.'
                )
                return
              }
              setInlineCollectionValidationMessage('')
              setInlineCollectionName('')
              setInlineCollectionInputOpen(false)
            }}
          >
            Save
          </Button>
          {inlineCollectionValidationMessage ? (
            <span
              className='text-sm text-destructive'
              data-testid='wishlist-table-new-collection-validation'
              role='alert'
            >
              {inlineCollectionValidationMessage}
            </span>
          ) : null}
        </div>
      ) : null}
    </div>
  ) : undefined

  return (
    <>
      <Header
        fixed
        data-testid={isWishlistRoute ? 'wishlist-shell-header' : undefined}
      >
        <Search />
        {isWishlistRoute ? (
          <HeaderTitle
            title={title}
            description={description}
            icon={Heart}
            testId='wishlist-header-title'
            iconTestId='wishlist-page-icon'
          />
        ) : null}
        <div
          className='ms-auto flex min-w-0 items-center gap-3'
          data-header-title-avoid='true'
        >
          {isWishlistRoute ? (
            <>
              <div
                className='flex min-w-0 flex-wrap items-center justify-end gap-2'
                data-testid='wishlist-global-header-actions'
              >
                <TasksHeaderActions
                  isWishlistRoute={isWishlistRoute}
                  onOpenCollectionCreate={() => {
                    setInlineCollectionValidationMessage('')
                    setInlineCollectionInputOpen(true)
                  }}
                  onCreate={() => {
                    setCurrentDialogRow(null)
                    setDialogOpen('create')
                  }}
                  onImport={() => {
                    setCurrentDialogRow(null)
                    setDialogOpen('import')
                  }}
                />
              </div>
              <Separator
                orientation='vertical'
                className='h-6'
                data-testid='wishlist-header-action-separator'
              />
            </>
          ) : null}
          <div className='flex items-center space-x-4'>
            <LanguageSwitch />
            <ThemeSwitch />
            <ConfigDrawer />
            <ProfileDropdown />
          </div>
        </div>
      </Header>

      <Main className='flex flex-1 flex-col gap-3 sm:gap-4'>
        <div
          className='flex flex-1 flex-col gap-4 sm:gap-6'
          data-testid={isWishlistRoute ? 'wishlist-workspace' : undefined}
        >
          {isWishlistRoute ? (
            <div
              className='grid gap-3 md:grid-cols-2 xl:grid-cols-4'
              data-testid='wishlist-planning-summary'
            >
              {wishlistPlanningSummary.map((focus) => (
                <button
                  key={focus.id}
                  type='button'
                  data-testid={`wishlist-planning-focus-${focus.id}`}
                  className={`rounded-lg border p-4 text-left transition-colors ${
                    wishlistPlanningFocus === focus.id
                      ? 'border-primary bg-primary/5'
                      : 'hover:bg-accent/40'
                  }`}
                  aria-pressed={wishlistPlanningFocus === focus.id}
                  onClick={() => setWishlistPlanningFocus(focus.id)}
                >
                  <div className='flex items-start justify-between gap-3'>
                    <div>
                      <p className='text-sm font-medium'>{focus.label}</p>
                      <p className='mt-1 text-xs text-muted-foreground'>
                        {focus.description}
                      </p>
                    </div>
                    <span className='text-2xl font-semibold'>
                      {focus.count}
                    </span>
                  </div>
                </button>
              ))}
            </div>
          ) : null}
          <TasksTable
            data={displayedData}
            routePath={routePath}
            customFilters={wishlistCollectionFilter}
            onEditRow={(task) => {
              setCurrentDialogRow(task)
              setDialogOpen('update')
            }}
            onDeleteRow={(task) => {
              setCurrentDialogRow(task)
              setDialogOpen('delete')
            }}
            onWishlistMarkOwned={
              isWishlistRoute ? handleWishlistMarkOwned : undefined
            }
            onWishlistBulkStatusChange={
              isWishlistRoute ? handleWishlistBulkStatusChange : undefined
            }
            onWishlistBulkPriorityChange={
              isWishlistRoute ? handleWishlistBulkPriorityChange : undefined
            }
            onWishlistBulkDelete={
              isWishlistRoute ? handleWishlistBulkDelete : undefined
            }
            onWishlistExport={
              isWishlistRoute ? handleWishlistExport : undefined
            }
            onWishlistInlineUpdate={
              isWishlistRoute ? handleWishlistInlineUpdate : undefined
            }
            wishlistActionItemID={wishlistActionItemID}
            isWishlistMutating={isWishlistMutating}
          />
        </div>
      </Main>

      <TasksDialogs
        routePath={routePath}
        open={dialogOpen}
        setOpen={setDialogOpen}
        currentRow={currentDialogRow}
        setCurrentRow={setCurrentDialogRow}
        onWishlistSubmit={isWishlistRoute ? handleWishlistSubmit : undefined}
        onWishlistDelete={isWishlistRoute ? handleWishlistDelete : undefined}
        onWishlistImport={isWishlistRoute ? handleWishlistImport : undefined}
        isWishlistMutating={isWishlistMutating}
      />
    </>
  )
}
