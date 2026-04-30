import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import {
  Camera,
  ClipboardPaste,
  FolderPlus,
  Heart,
  Plus,
  Upload,
} from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Separator } from '@/components/ui/separator'
import { Textarea } from '@/components/ui/textarea'
import { ConfigDrawer } from '@/components/config-drawer'
import { LanguageSwitch } from '@/components/language-switch'
import { Header, HeaderTitle } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { useWorkspaceCollections } from '@/features/collections/use-workspace-collections'
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

type WishlistItemPayload = {
  id?: string
  title?: string
  part_number?: string
  category?: string
  priority?: string
}

type WishlistPriceTrend = 'up' | 'steady' | 'down' | 'unknown'

type WishlistPriceHistorySnapshot = {
  snapshot_date?: string
  source?: string
  latest_price?: number
  stock_count?: number
}

type WishlistPricingSummary = {
  marketPrice?: number
  priceTrend: WishlistPriceTrend
  priceHistory: number[]
  priceHistoryDates: string[]
  priceSampleCount: number
  priceFirstDate?: string
  priceLatestDate?: string
  priceSources: string[]
  priceStockCount?: number
}

type WishlistEntryPayload = {
  id?: string
  item_id?: string
  priority?: string
  notes?: string
  below_target_now?: boolean
  target_price?: number
  highlight_hit?: boolean
  owned?: boolean
  price_paid?: number
  purchase_url?: string
  purchase_date?: string
  purchase_condition?: string
  quantity?: number
  needed_quantity?: number
}

type WishlistInlineChanges = {
  targetPrice?: number
  priority?: string
  owned?: boolean
  pricePaid?: number
  purchaseUrl?: string
  purchaseDate?: string
  purchaseCondition?: string
  quantity?: number
  neededQuantity?: number
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

function normalizeOptionalMoney(raw: string) {
  if (raw.trim() === '') {
    return 0
  }
  const parsed = Number(raw)
  if (Number.isNaN(parsed) || parsed < 0) {
    throw new Error('invalid_money')
  }
  return parsed
}

function normalizeOptionalWholeNumber(raw: string, fallback: number) {
  if (raw.trim() === '') {
    return fallback
  }
  const parsed = Number(raw)
  if (!Number.isInteger(parsed) || parsed < 0) {
    throw new Error('invalid_quantity')
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
    const [statsResponse, trendResponse, historyResponse] = await Promise.all([
      fetch(`/api/pricing/stats?item_id=${encodeURIComponent(itemID)}`),
      fetch(`/api/pricing/trend?item_id=${encodeURIComponent(itemID)}`),
      fetch(`/api/pricing/history?item_id=${encodeURIComponent(itemID)}`),
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
    let priceHistory: number[] = []
    let priceHistoryDates: string[] = []
    if (trendResponse.ok) {
      const trendPayload = (await trendResponse.json()) as {
        points?: Array<{ date?: string; latest?: number }>
      }
      priceHistory = (trendPayload.points ?? [])
        .map((point) => point.latest)
        .filter((value): value is number => typeof value === 'number')
      priceHistoryDates = (trendPayload.points ?? [])
        .map((point) => point.date?.trim())
        .filter((value): value is string => Boolean(value))
      priceTrend = inferWishlistPriceTrend(trendPayload.points)
    }

    let history: WishlistPriceHistorySnapshot[] = []
    if (historyResponse.ok) {
      const historyPayload = (await historyResponse.json()) as {
        history?: WishlistPriceHistorySnapshot[]
      }
      history = historyPayload.history ?? []
    }

    const sourceSet = new Set<string>()
    let priceStockCount = 0
    history.forEach((snapshot) => {
      const source = snapshot.source?.trim()
      if (source) {
        sourceSet.add(source)
      }
      if (typeof snapshot.stock_count === 'number') {
        priceStockCount += snapshot.stock_count
      }
    })

    const historyDates = history
      .map((snapshot) => snapshot.snapshot_date?.trim())
      .filter((value): value is string => Boolean(value))
    const allDates = [...priceHistoryDates, ...historyDates].sort()
    const latestHistoryPrice = [...history]
      .reverse()
      .find(
        (snapshot) =>
          typeof snapshot.latest_price === 'number' && snapshot.latest_price > 0
      )?.latest_price
    if (!marketPrice && latestHistoryPrice) {
      marketPrice = latestHistoryPrice
    }

    return {
      marketPrice,
      priceTrend,
      priceHistory,
      priceHistoryDates,
      priceSampleCount: Math.max(priceHistory.length, history.length),
      priceFirstDate: allDates[0],
      priceLatestDate: allDates[allDates.length - 1],
      priceSources: Array.from(sourceSet).sort(),
      priceStockCount: priceStockCount > 0 ? priceStockCount : undefined,
    }
  } catch {
    return {
      priceTrend: 'unknown',
      priceHistory: [],
      priceHistoryDates: [],
      priceSampleCount: 0,
      priceSources: [],
    }
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

function extractFirstUrl(text: string) {
  return text.match(/https?:\/\/\S+/i)?.[0] ?? ''
}

function inferWishlistTitleFromPaste(text: string) {
  const lines = text
    .split(/\r?\n/)
    .map((line) => line.trim())
    .filter(Boolean)
  const descriptiveLine = lines.find((line) => !/^https?:\/\//i.test(line))
  if (descriptiveLine) {
    return descriptiveLine
  }

  const firstUrl = extractFirstUrl(text)
  if (firstUrl) {
    try {
      const url = new URL(firstUrl)
      return url.hostname.replace(/^www\./i, '')
    } catch {
      return firstUrl
    }
  }

  return text.trim().slice(0, 80)
}

async function captureWishlistScreenshotDataUrl() {
  const testDataUrl = (
    window as unknown as { cabinetWishlistScreenshotTestDataUrl?: string }
  ).cabinetWishlistScreenshotTestDataUrl
  if (testDataUrl) {
    return testDataUrl
  }

  const mediaDevices = navigator.mediaDevices as
    | (MediaDevices & {
        getDisplayMedia?: (constraints?: MediaStreamConstraints) => Promise<MediaStream>
      })
    | undefined
  if (typeof mediaDevices?.getDisplayMedia !== 'function') {
    throw new Error('screenshot_capture_unavailable')
  }

  const stream = await mediaDevices.getDisplayMedia({ video: true })
  try {
    const video = document.createElement('video')
    video.srcObject = stream
    video.muted = true
    video.playsInline = true
    await video.play()
    await new Promise((resolve) => window.setTimeout(resolve, 100))

    const width = video.videoWidth || 1280
    const height = video.videoHeight || 720
    const canvas = document.createElement('canvas')
    canvas.width = width
    canvas.height = height
    const context = canvas.getContext('2d')
    if (!context) {
      throw new Error('screenshot_canvas_unavailable')
    }
    context.drawImage(video, 0, 0, width, height)
    return canvas.toDataURL('image/png')
  } finally {
    stream.getTracks().forEach((track) => track.stop())
  }
}

function dataUrlToFile(dataUrl: string, filename: string) {
  const [metadata = '', data = ''] = dataUrl.split(',')
  const mimeType = metadata.match(/data:(.*?);base64/)?.[1] ?? 'image/png'
  const binary = window.atob(data)
  const bytes = new Uint8Array(binary.length)
  for (let index = 0; index < binary.length; index += 1) {
    bytes[index] = binary.charCodeAt(index)
  }
  return new File([bytes], filename, { type: mimeType })
}

function TasksHeaderActions({
  onPaste,
  onScreenshot,
  onOpenCollectionCreate,
  onCreate,
  onImport,
}: {
  onPaste: () => void
  onScreenshot: () => void
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
        variant='outline'
        size='icon'
        data-testid='wishlist-paste-action'
        aria-label='Paste wishlist item'
        title='Paste wishlist item'
        onClick={onPaste}
      >
        <ClipboardPaste className='h-4 w-4' aria-hidden='true' />
      </Button>
      <Button
        type='button'
        variant='outline'
        size='icon'
        data-testid='wishlist-screenshot-action'
        aria-label='Add wishlist screenshot'
        title='Add wishlist screenshot'
        onClick={onScreenshot}
      >
        <Camera className='h-4 w-4' aria-hidden='true' />
      </Button>
      <Button
        type='button'
        size='icon'
        data-testid='wishlist-new-action'
        aria-label='New wishlist item'
        title='New wishlist item'
        onClick={onCreate}
      >
        <Plus className='h-4 w-4' aria-hidden='true' />
      </Button>
      <Button
        type='button'
        variant='outline'
        size='icon'
        data-testid='wishlist-create-collection-action'
        aria-label='Create collection'
        title='Create collection'
        onClick={onOpenCollectionCreate}
      >
        <FolderPlus className='h-4 w-4' aria-hidden='true' />
      </Button>
      <Button
        type='button'
        variant='outline'
        size='icon'
        data-testid='wishlist-import-action'
        aria-label='Import wishlist entries'
        title='Import wishlist entries'
        onClick={onImport}
      >
        <Upload className='h-4 w-4' aria-hidden='true' />
      </Button>
    </div>
  )
}

export function Tasks({
  title = 'Tasks',
  description = '',
  routePath = '/_authenticated/inventory/',
}: TasksProps) {
  const { workspaceCollections, collectionItems, addCollection } =
    useWorkspaceCollections()
  const [collectionCreateOpen, setCollectionCreateOpen] = useState(false)
  const [collectionCreateName, setCollectionCreateName] = useState('')
  const [
    collectionCreateValidationMessage,
    setCollectionCreateValidationMessage,
  ] = useState('')
  const isWishlistRoute = routePath === '/_authenticated/wishlist/'
  const [tableData, setTableData] = useState<Task[]>(() =>
    isWishlistRoute ? [] : tasks
  )
  const tableDataRef = useRef(tableData)
  const [dialogOpen, setDialogOpen] = useState<TasksDialogType | null>(null)
  const [currentDialogRow, setCurrentDialogRow] = useState<Task | null>(null)
  const [dialogNavigationRows, setDialogNavigationRows] = useState<Task[]>([])
  const [isWishlistMutating, setIsWishlistMutating] = useState(false)
  const [wishlistPasteOpen, setWishlistPasteOpen] = useState(false)
  const [wishlistPasteContent, setWishlistPasteContent] = useState('')
  const [wishlistPasteTitle, setWishlistPasteTitle] = useState('')
  const [wishlistScreenshotOpen, setWishlistScreenshotOpen] = useState(false)
  const [wishlistScreenshotDataUrl, setWishlistScreenshotDataUrl] = useState('')
  const [wishlistScreenshotTitle, setWishlistScreenshotTitle] = useState(
    'Wishlist screenshot'
  )
  const [wishlistScreenshotError, setWishlistScreenshotError] = useState('')
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
          priceHistory: pricingSummary.priceHistory,
          priceHistoryDates: pricingSummary.priceHistoryDates,
          priceSampleCount: pricingSummary.priceSampleCount,
          priceFirstDate: pricingSummary.priceFirstDate,
          priceLatestDate: pricingSummary.priceLatestDate,
          priceSources: pricingSummary.priceSources,
          priceStockCount: pricingSummary.priceStockCount,
          highlightHit: Boolean(wishlistEntry?.highlight_hit),
          owned: Boolean(wishlistEntry?.owned),
          pricePaid:
            typeof wishlistEntry?.price_paid === 'number'
              ? wishlistEntry.price_paid
              : undefined,
          purchaseUrl: wishlistEntry?.purchase_url?.trim(),
          purchaseDate: wishlistEntry?.purchase_date?.trim(),
          purchaseCondition: wishlistEntry?.purchase_condition?.trim(),
          quantity:
            typeof wishlistEntry?.quantity === 'number'
              ? wishlistEntry.quantity
              : 0,
          neededQuantity:
            typeof wishlistEntry?.needed_quantity === 'number'
              ? wishlistEntry.needed_quantity
              : 1,
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
          owned: draft.owned,
          price_paid: normalizeOptionalMoney(draft.pricePaid),
          purchase_url: draft.purchaseUrl,
          purchase_date: draft.purchaseDate,
          purchase_condition: draft.purchaseCondition,
          quantity: normalizeOptionalWholeNumber(draft.quantity, 0),
          needed_quantity: normalizeOptionalWholeNumber(
            draft.neededQuantity,
            1
          ),
          highlight_hit: currentRow?.highlightHit ?? false,
        }),
      })

      if (!wishlistResponse.ok) {
        throw new Error('wishlist_entry_save_failed')
      }

      return itemID
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
        } else if (
          error instanceof Error &&
          error.message === 'invalid_money'
        ) {
          toast.error('Money fields must be positive numbers.')
        } else if (
          error instanceof Error &&
          error.message === 'invalid_quantity'
        ) {
          toast.error('Quantity fields must be whole numbers.')
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
    async (task: Task, changes: WishlistInlineChanges) => {
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
      const nextOwned = changes.owned ?? currentTask.owned ?? false
      const nextPricePaid = changes.pricePaid ?? currentTask.pricePaid ?? 0
      const nextPurchaseUrl =
        changes.purchaseUrl ?? currentTask.purchaseUrl ?? ''
      const nextPurchaseDate =
        changes.purchaseDate ?? currentTask.purchaseDate ?? ''
      const nextPurchaseCondition =
        changes.purchaseCondition ?? currentTask.purchaseCondition ?? ''
      const nextQuantity = changes.quantity ?? currentTask.quantity ?? 0
      const nextNeededQuantity =
        changes.neededQuantity ?? currentTask.neededQuantity ?? 1

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
              owned: nextOwned,
              pricePaid: nextPricePaid,
              purchaseUrl: nextPurchaseUrl,
              purchaseDate: nextPurchaseDate,
              purchaseCondition: nextPurchaseCondition,
              quantity: nextQuantity,
              neededQuantity: nextNeededQuantity,
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
                owned: nextOwned,
                pricePaid: nextPricePaid,
                purchaseUrl: nextPurchaseUrl,
                purchaseDate: nextPurchaseDate,
                purchaseCondition: nextPurchaseCondition,
                quantity: nextQuantity,
                neededQuantity: nextNeededQuantity,
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
        if (changes.owned !== undefined) {
          requestBody.owned = nextOwned
        }
        if (changes.pricePaid !== undefined) {
          requestBody.price_paid = nextPricePaid
        }
        if (changes.purchaseUrl !== undefined) {
          requestBody.purchase_url = nextPurchaseUrl
        }
        if (changes.purchaseDate !== undefined) {
          requestBody.purchase_date = nextPurchaseDate
        }
        if (changes.purchaseCondition !== undefined) {
          requestBody.purchase_condition = nextPurchaseCondition
        }
        if (changes.quantity !== undefined) {
          requestBody.quantity = nextQuantity
        }
        if (changes.neededQuantity !== undefined) {
          requestBody.needed_quantity = nextNeededQuantity
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

  const openWishlistPasteDialog = useCallback(async () => {
    let clipboardText = ''
    try {
      clipboardText = (await navigator.clipboard?.readText?.()) ?? ''
    } catch {
      clipboardText = ''
    }
    setWishlistPasteContent(clipboardText)
    setWishlistPasteTitle(inferWishlistTitleFromPaste(clipboardText))
    setWishlistPasteOpen(true)
  }, [])

  const handleWishlistPasteContentChange = useCallback((value: string) => {
    setWishlistPasteContent(value)
    setWishlistPasteTitle(inferWishlistTitleFromPaste(value))
  }, [])

  const handleWishlistPasteOpenChange = useCallback((open: boolean) => {
    setWishlistPasteOpen(open)
    if (!open) {
      setWishlistPasteContent('')
      setWishlistPasteTitle('')
    }
  }, [])

  const handleWishlistPasteSave = useCallback(async () => {
    const content = wishlistPasteContent.trim()
    const title = wishlistPasteTitle.trim() || inferWishlistTitleFromPaste(content)
    if (!content || !title) {
      toast.error('Paste text or a URL before adding to wishlist.')
      return
    }

    await handleWishlistSubmit({
      title,
      partNumber: '',
      category: '',
      priority: 'medium',
      notes: content,
      targetPrice: '',
      owned: false,
      pricePaid: '',
      purchaseUrl: extractFirstUrl(content),
      purchaseDate: '',
      purchaseCondition: '',
      quantity: '0',
      neededQuantity: '1',
    })
    handleWishlistPasteOpenChange(false)
  }, [
    handleWishlistPasteOpenChange,
    handleWishlistSubmit,
    wishlistPasteContent,
    wishlistPasteTitle,
  ])

  const handleWishlistScreenshotOpenChange = useCallback((open: boolean) => {
    setWishlistScreenshotOpen(open)
    if (!open) {
      setWishlistScreenshotDataUrl('')
      setWishlistScreenshotTitle('Wishlist screenshot')
      setWishlistScreenshotError('')
    }
  }, [])

  const handleWishlistScreenshotCapture = useCallback(async () => {
    setWishlistScreenshotError('')
    try {
      const dataUrl = await captureWishlistScreenshotDataUrl()
      setWishlistScreenshotDataUrl(dataUrl)
    } catch {
      setWishlistScreenshotError(
        'Screenshot capture is unavailable or was cancelled. Try again from a supported browser.'
      )
    }
  }, [])

  const handleWishlistScreenshotSave = useCallback(async () => {
    const title = wishlistScreenshotTitle.trim() || 'Wishlist screenshot'
    if (!wishlistScreenshotDataUrl) {
      setWishlistScreenshotError('Capture a screenshot before saving.')
      return
    }

    setIsWishlistMutating(true)
    try {
      const itemID = await saveWishlistDraft({
        title,
        partNumber: '',
        category: 'Screenshot',
        priority: 'medium',
        notes: 'Created from screenshot.',
        targetPrice: '',
        owned: false,
        pricePaid: '',
        purchaseUrl: '',
        purchaseDate: '',
        purchaseCondition: '',
        quantity: '0',
        neededQuantity: '1',
      })

      const photoBody = new FormData()
      photoBody.append(
        'file',
        dataUrlToFile(wishlistScreenshotDataUrl, 'wishlist-screenshot.png')
      )
      const photoResponse = await fetch(
        `/api/items/${encodeURIComponent(itemID)}/photos`,
        {
          method: 'POST',
          body: photoBody,
        }
      )
      if (!photoResponse.ok) {
        throw new Error('wishlist_screenshot_photo_failed')
      }

      await refreshWishlistTable()
      toast.success(`${title} added to wishlist with screenshot.`)
      handleWishlistScreenshotOpenChange(false)
    } catch {
      setWishlistScreenshotError(
        'Screenshot wishlist save failed. Try again.'
      )
      toast.error('Screenshot wishlist save failed. Try again.')
    } finally {
      setIsWishlistMutating(false)
    }
  }, [
    handleWishlistScreenshotOpenChange,
    refreshWishlistTable,
    saveWishlistDraft,
    wishlistScreenshotDataUrl,
    wishlistScreenshotTitle,
  ])

  useEffect(() => {
    if (!isWishlistRoute || typeof window === 'undefined') {
      return
    }
    window.localStorage.removeItem('cabinet.wishlistPlanningFocus')
  }, [isWishlistRoute])

  const displayedData = useMemo(() => {
    if (!isWishlistRoute) {
      return tableData
    }

    const collectionByItemID = new Map(
      collectionItems.map((item) => [item.id, item.collectionName])
    )
    return tableData.map((task) => ({
      ...task,
      collectionName:
        collectionByItemID.get(task.itemID ?? task.id) ?? 'Unassigned',
    }))
  }, [collectionItems, isWishlistRoute, tableData])

  const handleCreateWishlistCollection = useCallback(async () => {
    const created = await addCollection(collectionCreateName)
    if (!created) {
      setCollectionCreateValidationMessage('Collection name is required.')
      return
    }
    setCollectionCreateValidationMessage('')
    setCollectionCreateName('')
    setCollectionCreateOpen(false)
  }, [addCollection, collectionCreateName])

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
                  onPaste={() => {
                    void openWishlistPasteDialog()
                  }}
                  onScreenshot={() => {
                    setWishlistScreenshotOpen(true)
                  }}
                  onOpenCollectionCreate={() => {
                    setCollectionCreateValidationMessage('')
                    setCollectionCreateOpen(true)
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

      {isWishlistRoute ? (
        <Dialog
          open={wishlistPasteOpen}
          onOpenChange={handleWishlistPasteOpenChange}
        >
          <DialogContent
            className='sm:max-w-2xl'
            data-testid='wishlist-paste-dialog'
          >
            <DialogHeader>
              <DialogTitle>Add from Paste</DialogTitle>
              <DialogDescription>
                Paste a listing URL or notes and Cabinet will create a wishlist
                entry while keeping the original text on the record.
              </DialogDescription>
            </DialogHeader>
            <div className='grid gap-4'>
              <label className='grid gap-2 text-sm font-medium'>
                Paste URL or text
                <Textarea
                  data-testid='wishlist-paste-content'
                  className='min-h-36'
                  value={wishlistPasteContent}
                  placeholder='Paste a marketplace link, listing title, or notes...'
                  onChange={(event) =>
                    handleWishlistPasteContentChange(event.target.value)
                  }
                />
              </label>
              <label className='grid gap-2 text-sm font-medium'>
                Wishlist title
                <Input
                  data-testid='wishlist-paste-title'
                  value={wishlistPasteTitle}
                  placeholder='Title inferred from pasted content'
                  onChange={(event) =>
                    setWishlistPasteTitle(event.target.value)
                  }
                />
              </label>
            </div>
            <DialogFooter>
              <Button
                type='button'
                variant='outline'
                disabled={isWishlistMutating}
                onClick={() => handleWishlistPasteOpenChange(false)}
              >
                Cancel
              </Button>
              <Button
                type='button'
                data-testid='wishlist-paste-save'
                disabled={
                  isWishlistMutating ||
                  !wishlistPasteContent.trim() ||
                  !wishlistPasteTitle.trim()
                }
                onClick={() => {
                  void handleWishlistPasteSave()
                }}
              >
                {isWishlistMutating ? 'Adding...' : 'Add to wishlist'}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      ) : null}

      {isWishlistRoute ? (
        <Dialog
          open={wishlistScreenshotOpen}
          onOpenChange={handleWishlistScreenshotOpenChange}
        >
          <DialogContent
            className='sm:max-w-2xl'
            data-testid='wishlist-screenshot-dialog'
          >
            <DialogHeader>
              <DialogTitle>Add Screenshot</DialogTitle>
              <DialogDescription>
                Capture a screen or window and save it as a new wishlist item
                with the image attached.
              </DialogDescription>
            </DialogHeader>
            <div className='grid gap-4'>
              <div className='rounded-lg border border-dashed p-3'>
                {wishlistScreenshotDataUrl ? (
                  <img
                    src={wishlistScreenshotDataUrl}
                    alt='Captured wishlist screenshot preview'
                    data-testid='wishlist-screenshot-preview'
                    className='max-h-72 w-full rounded-md object-contain'
                  />
                ) : (
                  <div className='flex min-h-44 items-center justify-center rounded-md bg-muted/30 text-sm text-muted-foreground'>
                    No screenshot captured yet.
                  </div>
                )}
              </div>
              <div className='flex flex-wrap gap-2'>
                <Button
                  type='button'
                  variant='outline'
                  data-testid='wishlist-screenshot-capture'
                  disabled={isWishlistMutating}
                  onClick={() => {
                    void handleWishlistScreenshotCapture()
                  }}
                >
                  Capture screenshot
                </Button>
              </div>
              <label className='grid gap-2 text-sm font-medium'>
                Wishlist title
                <Input
                  data-testid='wishlist-screenshot-title'
                  value={wishlistScreenshotTitle}
                  placeholder='Wishlist screenshot'
                  onChange={(event) =>
                    setWishlistScreenshotTitle(event.target.value)
                  }
                />
              </label>
              {wishlistScreenshotError ? (
                <p
                  className='rounded-md border border-destructive/40 bg-destructive/10 p-3 text-sm text-destructive'
                  data-testid='wishlist-screenshot-error'
                >
                  {wishlistScreenshotError}
                </p>
              ) : null}
            </div>
            <DialogFooter>
              <Button
                type='button'
                variant='outline'
                disabled={isWishlistMutating}
                onClick={() => handleWishlistScreenshotOpenChange(false)}
              >
                Cancel
              </Button>
              <Button
                type='button'
                data-testid='wishlist-screenshot-save'
                disabled={
                  isWishlistMutating ||
                  !wishlistScreenshotDataUrl ||
                  !wishlistScreenshotTitle.trim()
                }
                onClick={() => {
                  void handleWishlistScreenshotSave()
                }}
              >
                {isWishlistMutating ? 'Adding...' : 'Add to wishlist'}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      ) : null}

      <Main className='flex flex-1 flex-col gap-3 sm:gap-4'>
        <div
          className='flex flex-1 flex-col gap-4 sm:gap-6'
          data-testid={isWishlistRoute ? 'wishlist-workspace' : undefined}
        >
          <TasksTable
            data={displayedData}
            routePath={routePath}
            onEditRow={(task, navigationRows) => {
              setDialogNavigationRows(navigationRows ?? displayedData)
              setCurrentDialogRow(task)
              setDialogOpen('update')
            }}
            onDeleteRow={(task) => {
              setDialogNavigationRows([])
              setCurrentDialogRow(task)
              setDialogOpen('delete')
            }}
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
            isWishlistMutating={isWishlistMutating}
            wishlistCollectionOptions={
              isWishlistRoute ? workspaceCollections : undefined
            }
          />
        </div>
      </Main>

      {isWishlistRoute ? (
        <Dialog
          open={collectionCreateOpen}
          onOpenChange={(open) => {
            setCollectionCreateOpen(open)
            if (!open) {
              setCollectionCreateName('')
              setCollectionCreateValidationMessage('')
            }
          }}
        >
          <DialogContent data-testid='wishlist-create-collection-dialog'>
            <DialogHeader>
              <DialogTitle>Create collection</DialogTitle>
              <DialogDescription>
                Add a reusable collection filter option for Wishlist and
                Collections.
              </DialogDescription>
            </DialogHeader>
            <div className='space-y-2'>
              <Input
                data-testid='wishlist-create-collection-name'
                placeholder='Collection name'
                aria-invalid={
                  collectionCreateValidationMessage ? 'true' : 'false'
                }
                value={collectionCreateName}
                onChange={(event) => {
                  setCollectionCreateName(event.target.value)
                  if (collectionCreateValidationMessage) {
                    setCollectionCreateValidationMessage('')
                  }
                }}
                onKeyDown={(event) => {
                  if (event.key === 'Enter') {
                    event.preventDefault()
                    void handleCreateWishlistCollection()
                  }
                }}
              />
              {collectionCreateValidationMessage ? (
                <p
                  className='text-sm text-destructive'
                  data-testid='wishlist-create-collection-validation'
                  role='alert'
                >
                  {collectionCreateValidationMessage}
                </p>
              ) : null}
            </div>
            <DialogFooter>
              <Button
                type='button'
                variant='outline'
                onClick={() => setCollectionCreateOpen(false)}
              >
                Cancel
              </Button>
              <Button
                type='button'
                data-testid='wishlist-create-collection-save'
                onClick={() => void handleCreateWishlistCollection()}
              >
                Save
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      ) : null}

      <TasksDialogs
        routePath={routePath}
        open={dialogOpen}
        setOpen={setDialogOpen}
        currentRow={currentDialogRow}
        setCurrentRow={setCurrentDialogRow}
        navigationRows={dialogNavigationRows}
        onWishlistSubmit={isWishlistRoute ? handleWishlistSubmit : undefined}
        onWishlistDelete={isWishlistRoute ? handleWishlistDelete : undefined}
        onWishlistImport={isWishlistRoute ? handleWishlistImport : undefined}
        isWishlistMutating={isWishlistMutating}
      />
    </>
  )
}
