import {
  type KeyboardEvent,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from 'react'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { ConfigDrawer } from '@/components/config-drawer'
import { Header } from '@/components/layout/header'
import { LanguageSwitch } from '@/components/language-switch'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { Input } from '@/components/ui/input'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { TasksTable } from '@/features/tasks/components/tasks-table'
import { TasksDialogs } from '@/features/tasks/components/tasks-dialogs'
import { TasksProvider } from '@/features/tasks/components/tasks-provider'
import { tasks } from '@/features/tasks/data/tasks'
import { type Task } from '@/features/tasks/data/schema'
import {
  collectionKey,
  useWorkspaceCollections,
} from '@/features/collections/use-workspace-collections'

type CollectionWorkspaceProps = {
  title?: string
  description?: string
  routePath: '/_authenticated/inventory/' | '/_authenticated/wishlist/'
}

type InventoryPhoto = {
  id: string
  filename: string
  is_primary?: boolean
}

type BarcodeMatch = {
  id: string
  item_id: string
  barcode: string
  created_at?: string
}

type InventoryItem = {
  id: string
  part_number: string
  title: string
  status: string
  category: string
  brand: string
  priority: string
  description: string
}

type InventoryItemDraft = {
  part_number: string
  title: string
  brand: string
  category: string
  description: string
}

type AISuggestion = {
  part_number?: string
  brand?: string
  title?: string
  confidence?: number
  requires_confirmation?: boolean
}

type FolderNode = {
  id: string
  name: string
  children?: FolderNode[]
}

const initialFolderTree: FolderNode[] = [
  { id: 'all-items', name: 'All Items' },
  { id: 'watch-list', name: 'Watch List' },
  { id: 'wishlist-focus', name: 'Wishlist Focus' },
  { id: 'store-1', name: 'Store 1' },
  { id: 'store-2', name: 'Store 2' },
  { id: 'store-3', name: 'Store 3' },
  { id: 'store-4', name: 'Store 4' },
  { id: 'store-5', name: 'Store 5' },
  { id: 'store-6', name: 'Store 6' },
  { id: 'store-7', name: 'Store 7' },
  { id: 'store-8', name: 'Store 8' },
  { id: 'store-9', name: 'Store 9' },
  { id: 'store-10', name: 'Store 10' },
  {
    id: 'warehouses',
    name: 'Warehouses',
    children: [
      { id: 'warehouse-1', name: 'Warehouse 1' },
      { id: 'warehouse-2', name: 'Warehouse 2' },
      { id: 'warehouse-3', name: 'Warehouse 3' },
    ],
  },
  { id: 'archive-a', name: 'Archive A' },
  { id: 'archive-b', name: 'Archive B' },
  { id: 'archive-c', name: 'Archive C' },
  { id: 'archive-d', name: 'Archive D' },
  { id: 'archive-e', name: 'Archive E' },
  { id: 'archive-f', name: 'Archive F' },
  { id: 'archive-g', name: 'Archive G' },
  { id: 'archive-h', name: 'Archive H' },
  { id: 'archive-i', name: 'Archive I' },
  { id: 'archive-j', name: 'Archive J' },
  { id: 'archive-k', name: 'Archive K' },
  { id: 'archive-l', name: 'Archive L' },
]

function emptyInventoryItemDraft(): InventoryItemDraft {
  return {
    part_number: '',
    title: '',
    brand: '',
    category: '',
    description: '',
  }
}

function inventoryItemToDraft(item: InventoryItem): InventoryItemDraft {
  return {
    part_number: item.part_number,
    title: item.title,
    brand: item.brand,
    category: item.category,
    description: item.description,
  }
}

function inventoryItemToTask(item: InventoryItem): Task {
  return {
    id: item.part_number || item.id,
    itemID: item.id,
    title: item.title || 'Untitled item',
    status: item.status || 'todo',
    label: item.category || 'feature',
    priority: item.priority || 'medium',
  }
}

function countFolderNodes(nodes: FolderNode[]): number {
  return nodes.reduce((count, node) => {
    const childCount = node.children ? countFolderNodes(node.children) : 0
    return count + 1 + childCount
  }, 0)
}

function slugifyFolderName(value: string): string {
  return value
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
}

function buildUniqueFolderID(name: string, nodes: FolderNode[]): string {
  const slug = slugifyFolderName(name) || 'folder'
  const existingIDs = new Set<string>()
  const walk = (items: FolderNode[]) => {
    items.forEach((item) => {
      existingIDs.add(item.id)
      if (item.children?.length) {
        walk(item.children)
      }
    })
  }
  walk(nodes)
  if (!existingIDs.has(slug)) {
    return slug
  }
  let counter = 2
  while (existingIDs.has(`${slug}-${counter}`)) {
    counter += 1
  }
  return `${slug}-${counter}`
}

function addChildFolder(
  nodes: FolderNode[],
  parentID: string,
  child: FolderNode
): FolderNode[] {
  return nodes.map((node) => {
    if (node.id === parentID) {
      return {
        ...node,
        children: [...(node.children ?? []), child],
      }
    }
    if (node.children?.length) {
      return {
        ...node,
        children: addChildFolder(node.children, parentID, child),
      }
    }
    return node
  })
}

export function Collection({
  title = 'Collection',
  description = 'Command your inventory and move from folders to item actions quickly.',
  routePath,
}: CollectionWorkspaceProps) {
  const [tableData, setTableData] = useState<Task[]>(tasks)
  const [inventoryItems, setInventoryItems] = useState<InventoryItem[]>([])
  const [folderTree, setFolderTree] = useState<FolderNode[]>(initialFolderTree)
  const [activeFolder, setActiveFolder] = useState('All Items')
  const [inlineCollectionInputOpen, setInlineCollectionInputOpen] = useState(false)
  const [inlineCollectionName, setInlineCollectionName] = useState('')
  const [expandedNodeIDs, setExpandedNodeIDs] = useState<Set<string>>(
    () => new Set()
  )
  const [folderCreateOpen, setFolderCreateOpen] = useState(false)
  const [folderCreateParentID, setFolderCreateParentID] = useState<string | null>(
    null
  )
  const [folderCreateName, setFolderCreateName] = useState('')
  const treeItemRefs = useRef<Record<string, HTMLButtonElement | null>>({})
  const [loading, setLoading] = useState(false)
  const [loadError, setLoadError] = useState<string | null>(null)
  const [inventoryPhotos, setInventoryPhotos] = useState<InventoryPhoto[]>([])
  const [photosLoading, setPhotosLoading] = useState(false)
  const [photosError, setPhotosError] = useState<string | null>(null)
  const [photosBusy, setPhotosBusy] = useState(false)
  const [cameraError, setCameraError] = useState<string | null>(null)
  const [cameraSuccess, setCameraSuccess] = useState<string | null>(null)
  const [activeProfileID, setActiveProfileID] = useState('')
  const [selectedItemID, setSelectedItemID] = useState('')
  const [selectedItemLabel, setSelectedItemLabel] = useState('')
  const [itemEditorMode, setItemEditorMode] = useState<'create' | 'edit'>('edit')
  const [itemDraft, setItemDraft] = useState<InventoryItemDraft>(
    emptyInventoryItemDraft
  )
  const [itemSaveBusy, setItemSaveBusy] = useState(false)
  const [itemSaveError, setItemSaveError] = useState<string | null>(null)
  const [itemSaveSuccess, setItemSaveSuccess] = useState<string | null>(null)
  const [barcodeAddInput, setBarcodeAddInput] = useState('')
  const [barcodeLookupInput, setBarcodeLookupInput] = useState('')
  const [barcodeAddBusy, setBarcodeAddBusy] = useState(false)
  const [barcodeLookupBusy, setBarcodeLookupBusy] = useState(false)
  const [barcodeAddError, setBarcodeAddError] = useState<string | null>(null)
  const [barcodeAddSuccess, setBarcodeAddSuccess] = useState<string | null>(null)
  const [barcodeLookupError, setBarcodeLookupError] = useState<string | null>(null)
  const [barcodeMatches, setBarcodeMatches] = useState<BarcodeMatch[]>([])
  const [barcodeLookupCompleted, setBarcodeLookupCompleted] = useState(false)
  const [lastLookupBarcode, setLastLookupBarcode] = useState('')
  const [aiTitleInput, setAITitleInput] = useState('')
  const [aiPhotoURLInput, setAIPhotoURLInput] = useState('')
  const [aiSuggestion, setAISuggestion] = useState<AISuggestion | null>(null)
  const [aiLoading, setAILoading] = useState(false)
  const [aiError, setAIError] = useState<string | null>(null)
  const [aiApplied, setAIApplied] = useState(false)
  const [aiConfirmOpen, setAIConfirmOpen] = useState(false)
  const [lastAISuggestAction, setLastAISuggestAction] = useState<
    'title' | 'photo' | null
  >(null)
  const [selectedPhotoIndex, setSelectedPhotoIndex] = useState<number | null>(
    null
  )
  const aiApplyTriggerRef = useRef<HTMLButtonElement | null>(null)
  const {
    workspaceCollections,
    activeWorkspaceCollection,
    setActiveWorkspaceCollection,
    addCollection,
  } = useWorkspaceCollections()

  const selectInventoryItem = useCallback((item: InventoryItem | null) => {
    setSelectedItemID(item?.id ?? '')
    setSelectedItemLabel(item ? item.part_number || item.id : '')
  }, [])

  const startCreateItem = useCallback(() => {
    setItemEditorMode('create')
    setItemDraft(emptyInventoryItemDraft())
    setItemSaveError(null)
    setItemSaveSuccess(null)
    selectInventoryItem(null)
  }, [selectInventoryItem])

  const loadInventoryItems = useCallback(async (preferredSelectedItemID?: string) => {
    if (routePath !== '/_authenticated/inventory/') {
      setInventoryItems([])
      setTableData(tasks)
      setLoadError(null)
      selectInventoryItem(null)
      return
    }
    setLoading(true)
    setLoadError(null)
    try {
      const response = await fetch('/api/items')
      if (!response.ok) {
        throw new Error(`items_api_${response.status}`)
      }
      const payload = (await response.json()) as {
        items?: Array<{
          id?: string
          part_number?: string
          title?: string
          status?: string
          category?: string
          brand?: string
          priority?: string
          description?: string
        }>
      }
      const items = (payload.items ?? [])
        .map((item) => ({
          id: item.id?.trim() ?? '',
          part_number: item.part_number?.trim() ?? '',
          title: item.title?.trim() || 'Untitled item',
          status: item.status?.trim() || 'active',
          category: item.category?.trim() || 'General',
          brand: item.brand?.trim() || 'Unknown',
          priority: item.priority?.trim() || 'medium',
          description: item.description?.trim() ?? '',
        }))
        .filter((item) => item.id !== '')
      const mapped: Task[] = items.map(inventoryItemToTask)
      const targetItem =
        items.find(
          (item) =>
            item.id ===
            (preferredSelectedItemID?.trim() || selectedItemID.trim())
        ) ?? items[0] ?? null
      setInventoryItems(items)
      setTableData(mapped)
      selectInventoryItem(targetItem)
    } catch {
      setLoadError(
        'Inventory failed to load. Retry and confirm runtime API availability.'
      )
      setInventoryItems([])
      setTableData([])
      selectInventoryItem(null)
    } finally {
      setLoading(false)
    }
  }, [routePath, selectInventoryItem, selectedItemID])

  useEffect(() => {
    void loadInventoryItems()
  }, [loadInventoryItems])

  const visibleTreeNodes = useMemo(() => {
    const nodes: Array<{ id: string; name: string; hasChildren: boolean }> = []
    const walk = (treeNodes: FolderNode[]) => {
      treeNodes.forEach((node) => {
        nodes.push({
          id: node.id,
          name: node.name,
          hasChildren: Boolean(node.children?.length),
        })
        if (node.children?.length && expandedNodeIDs.has(node.id)) {
          walk(node.children)
        }
      })
    }
    walk(folderTree)
    return nodes
  }, [expandedNodeIDs, folderTree])

  const focusTreeItemByOffset = useCallback(
    (currentID: string, offset: number) => {
      const currentIndex = visibleTreeNodes.findIndex(
        (node) => node.id === currentID
      )
      if (currentIndex < 0) {
        return
      }
      const nextIndex = currentIndex + offset
      if (nextIndex < 0 || nextIndex >= visibleTreeNodes.length) {
        return
      }
      const nextID = visibleTreeNodes[nextIndex]?.id
      if (!nextID) {
        return
      }
      treeItemRefs.current[nextID]?.focus()
    },
    [visibleTreeNodes]
  )

  const toggleNodeExpanded = useCallback((nodeID: string) => {
    setExpandedNodeIDs((previous) => {
      const next = new Set(previous)
      if (next.has(nodeID)) {
        next.delete(nodeID)
      } else {
        next.add(nodeID)
      }
      return next
    })
  }, [])

  const handleTreeItemKeyDown = useCallback(
    (node: FolderNode, event: KeyboardEvent<HTMLButtonElement>) => {
      if (event.key === 'ArrowRight') {
        event.preventDefault()
        if (node.children?.length && !expandedNodeIDs.has(node.id)) {
          toggleNodeExpanded(node.id)
        }
        setActiveFolder(node.name)
        return
      }
      if (event.key === 'ArrowLeft') {
        event.preventDefault()
        if (node.children?.length && expandedNodeIDs.has(node.id)) {
          toggleNodeExpanded(node.id)
        }
        return
      }
      if (event.key === 'ArrowDown') {
        event.preventDefault()
        focusTreeItemByOffset(node.id, 1)
        return
      }
      if (event.key === 'ArrowUp') {
        event.preventDefault()
        focusTreeItemByOffset(node.id, -1)
        return
      }
      if (event.key === 'Enter' || event.key === ' ') {
        event.preventDefault()
        setActiveFolder(node.name)
      }
    },
    [expandedNodeIDs, focusTreeItemByOffset, toggleNodeExpanded]
  )

  const renderFolderTree = useCallback(
    (nodes: FolderNode[], level = 1) =>
      nodes.map((node) => {
        const hasChildren = Boolean(node.children?.length)
        const expanded = hasChildren && expandedNodeIDs.has(node.id)
        return (
          <div key={node.id} role='none'>
            <div
              className='flex items-center gap-2'
              style={{ paddingInlineStart: `${(level - 1) * 0.75}rem` }}
            >
              <span
                aria-hidden='true'
                data-testid={`folder-tree-connector-${node.id}`}
                className='inline-flex h-7 w-2 shrink-0 border-l border-border/70'
              />
              {hasChildren ? (
                <Button
                  type='button'
                  variant='ghost'
                  size='icon'
                  className='h-7 w-7'
                  data-testid={`folder-tree-toggle-${node.id}`}
                  aria-label={`Toggle ${node.name}`}
                  aria-expanded={expanded ? 'true' : 'false'}
                  onClick={() => toggleNodeExpanded(node.id)}
                >
                  {expanded ? '−' : '+'}
                </Button>
              ) : (
                <span className='inline-flex h-7 w-7 shrink-0' />
              )}
              <button
                ref={(element) => {
                  treeItemRefs.current[node.id] = element
                }}
                type='button'
                role='treeitem'
                tabIndex={activeFolder === node.name ? 0 : -1}
                aria-level={level}
                aria-selected={activeFolder === node.name ? 'true' : 'false'}
                aria-expanded={hasChildren ? (expanded ? 'true' : 'false') : undefined}
                data-testid={`folder-tree-item-${node.id}`}
                className={`w-full rounded-md px-3 py-2 text-left text-sm ${activeFolder === node.name ? 'bg-primary text-primary-foreground' : 'bg-muted/30 hover:bg-muted/60'}`}
                onClick={() => setActiveFolder(node.name)}
                onKeyDown={(event) => handleTreeItemKeyDown(node, event)}
              >
                <span data-testid={`collection-folder-${node.id}`}>{node.name}</span>
              </button>
              <Button
                type='button'
                variant='ghost'
                size='icon'
                className='h-7 w-7 shrink-0 text-muted-foreground hover:text-foreground'
                data-testid={`folder-tree-add-child-${node.id}`}
                aria-label={`Add child folder under ${node.name}`}
                onClick={() => {
                  setFolderCreateParentID(node.id)
                  setFolderCreateName('')
                  setFolderCreateOpen(true)
                }}
              >
                +
              </Button>
            </div>
            {hasChildren && expanded ? (
              <div
                role='group'
                data-testid={`folder-tree-group-${node.id}`}
                className='space-y-2'
              >
                {renderFolderTree(node.children ?? [], level + 1)}
              </div>
            ) : null}
          </div>
        )
      }),
    [activeFolder, expandedNodeIDs, handleTreeItemKeyDown, toggleNodeExpanded]
  )

  const summary = useMemo(
    () => ({
      folders: countFolderNodes(folderTree),
      items: tableData.length,
      activeBrand: 'All',
      activeCategory: 'All',
      activeContext: activeFolder,
    }),
    [activeFolder, folderTree, tableData.length]
  )
  const isInventoryRoute = routePath === '/_authenticated/inventory/'
  const selectedInventoryItem = useMemo(
    () => inventoryItems.find((item) => item.id === selectedItemID) ?? null,
    [inventoryItems, selectedItemID]
  )
  const selectedItemContext = selectedItemLabel || selectedItemID || 'None'
  const selectedPhoto =
    selectedPhotoIndex === null ? null : inventoryPhotos[selectedPhotoIndex]

  useEffect(() => {
    if (activeWorkspaceCollection) {
      setActiveFolder(activeWorkspaceCollection)
    }
  }, [activeWorkspaceCollection])

  const loadInventoryPhotos = useCallback(async () => {
    if (!isInventoryRoute) {
      setInventoryPhotos([])
      setPhotosError(null)
      setPhotosLoading(false)
      return
    }
    if (!selectedItemID) {
      setInventoryPhotos([])
      setPhotosError(null)
      setPhotosLoading(false)
      return
    }
    setPhotosLoading(true)
    setPhotosError(null)
    try {
      const response = await fetch(
        `/api/items/${encodeURIComponent(selectedItemID)}/photos`
      )
      if (!response.ok) {
        throw new Error(`list_photos_${response.status}`)
      }
      const payload = (await response.json()) as {
        photos?: Array<{
          id?: string
          filename?: string
          is_primary?: boolean
        }>
      }
      const mapped = (payload.photos ?? []).map((photo) => ({
        id: photo.id?.trim() ?? '',
        filename: photo.filename?.trim() || 'photo.jpg',
        is_primary: photo.is_primary ?? false,
      }))
      setInventoryPhotos(mapped.filter((photo) => photo.id !== ''))
    } catch {
      setInventoryPhotos([])
      setPhotosError('Photos could not be loaded for this item. Retry to continue.')
    } finally {
      setPhotosLoading(false)
    }
  }, [isInventoryRoute, selectedItemID])

  useEffect(() => {
    void loadInventoryPhotos()
  }, [loadInventoryPhotos])

  useEffect(() => {
    if (
      selectedPhotoIndex !== null &&
      (selectedPhotoIndex < 0 || selectedPhotoIndex >= inventoryPhotos.length)
    ) {
      setSelectedPhotoIndex(null)
    }
  }, [inventoryPhotos.length, selectedPhotoIndex])

  useEffect(() => {
    if (!isInventoryRoute) {
      setActiveProfileID('')
      return
    }
    let cancelled = false
    const loadActiveProfile = async () => {
      try {
        const response = await fetch('/api/profiles/active')
        if (!response.ok) {
          return
        }
        const payload = (await response.json()) as { id?: string }
        if (!cancelled) {
          setActiveProfileID(payload.id?.trim() ?? '')
        }
      } catch {
        if (!cancelled) {
          setActiveProfileID('')
        }
      }
    }
    void loadActiveProfile()
    return () => {
      cancelled = true
    }
  }, [isInventoryRoute])

  useEffect(() => {
    if (itemEditorMode !== 'edit') {
      return
    }
    if (!selectedInventoryItem) {
      setItemDraft(emptyInventoryItemDraft())
      return
    }
    setItemDraft(inventoryItemToDraft(selectedInventoryItem))
  }, [itemEditorMode, selectedInventoryItem])

  const handleSaveItem = useCallback(async () => {
    const payload = {
      part_number: itemDraft.part_number.trim(),
      title: itemDraft.title.trim(),
      brand: itemDraft.brand.trim(),
      category: itemDraft.category.trim(),
      description: itemDraft.description.trim(),
    }
    if (payload.part_number === '' || payload.title === '') {
      setItemSaveError('Part number and title are required before saving.')
      setItemSaveSuccess(null)
      return
    }
    if (itemEditorMode === 'edit' && !selectedItemID) {
      setItemSaveError('Select an existing inventory item before editing.')
      setItemSaveSuccess(null)
      return
    }
    setItemSaveBusy(true)
    setItemSaveError(null)
    setItemSaveSuccess(null)
    try {
      const response = await fetch(
        itemEditorMode === 'create'
          ? '/api/items'
          : `/api/items/${encodeURIComponent(selectedItemID)}`,
        {
          method: itemEditorMode === 'create' ? 'POST' : 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload),
        }
      )
      if (!response.ok) {
        throw new Error(`save_item_${response.status}`)
      }
      const saved = (await response.json()) as { id?: string }
      const savedID = saved.id?.trim() ?? selectedItemID
      setItemEditorMode('edit')
      setItemSaveSuccess(
        itemEditorMode === 'create'
          ? 'Item created and selected for follow-up media attach.'
          : 'Item changes saved and reloaded from the API.'
      )
      await loadInventoryItems(savedID)
    } catch {
      setItemSaveError('Inventory save failed. Review the fields and retry.')
    } finally {
      setItemSaveBusy(false)
    }
  }, [itemDraft, itemEditorMode, loadInventoryItems, selectedItemID])

  const handlePhotoUpload = useCallback(
    async (file: File | null) => {
      if (!file || !selectedItemID) {
        return
      }
      setPhotosBusy(true)
      setPhotosError(null)
      try {
        const body = new FormData()
        body.append('file', file)
        const response = await fetch(
          `/api/items/${encodeURIComponent(selectedItemID)}/photos`,
          {
            method: 'POST',
            body,
          }
        )
        if (!response.ok) {
          throw new Error(`upload_photo_${response.status}`)
        }
        await loadInventoryPhotos()
      } catch {
        setPhotosError('Photo upload failed. Retry with a supported image file.')
      } finally {
        setPhotosBusy(false)
      }
    },
    [loadInventoryPhotos, selectedItemID]
  )

  const handleSetPrimaryPhoto = useCallback(
    async (photoID: string) => {
      if (!selectedItemID || !photoID) {
        return
      }
      setPhotosBusy(true)
      setPhotosError(null)
      try {
        const response = await fetch(
          `/api/items/${encodeURIComponent(
            selectedItemID
          )}/photos/${encodeURIComponent(photoID)}/primary`,
          {
            method: 'PUT',
          }
        )
        if (!response.ok) {
          throw new Error(`set_primary_${response.status}`)
        }
        await loadInventoryPhotos()
      } catch {
        setPhotosError('Unable to set primary photo right now. Retry this action.')
      } finally {
        setPhotosBusy(false)
      }
    },
    [loadInventoryPhotos, selectedItemID]
  )

  const handleDeletePhoto = useCallback(
    async (photoID: string) => {
      if (!selectedItemID || !photoID) {
        return
      }
      setPhotosBusy(true)
      setPhotosError(null)
      try {
        const response = await fetch(
          `/api/items/${encodeURIComponent(
            selectedItemID
          )}/photos/${encodeURIComponent(photoID)}`,
          {
            method: 'DELETE',
          }
        )
        if (!response.ok) {
          throw new Error(`delete_photo_${response.status}`)
        }
        await loadInventoryPhotos()
      } catch {
        setPhotosError('Unable to delete this photo. Retry this action.')
      } finally {
        setPhotosBusy(false)
      }
    },
    [loadInventoryPhotos, selectedItemID]
  )

  const handleTakePhoto = useCallback(async () => {
    setCameraError(null)
    setCameraSuccess(null)
    if (
      typeof navigator === 'undefined' ||
      !navigator.mediaDevices ||
      typeof navigator.mediaDevices.getUserMedia !== 'function'
    ) {
      setCameraError(
        'Camera capture is unavailable on this device. Use Upload File instead.'
      )
      return
    }
    let stream: MediaStream | null = null
    try {
      stream = await navigator.mediaDevices.getUserMedia({ video: true })
      const file = new File(['camera-capture'], 'camera-capture.jpg', {
        type: 'image/jpeg',
      })
      await handlePhotoUpload(file)
      setCameraSuccess('Camera capture uploaded successfully.')
    } catch {
      setCameraError(
        'Camera permission was denied or unavailable. Use Upload File instead.'
      )
    } finally {
      if (stream) {
        stream.getTracks().forEach((track) => {
          track.stop()
        })
      }
    }
  }, [handlePhotoUpload])

  const lookupBarcode = useCallback(async (rawBarcode: string) => {
    const barcode = rawBarcode.trim()
    if (barcode === '') {
      setBarcodeLookupError('Enter a barcode before lookup.')
      setBarcodeLookupCompleted(false)
      setBarcodeMatches([])
      return
    }
    setBarcodeLookupBusy(true)
    setBarcodeLookupError(null)
    setBarcodeLookupCompleted(false)
    setBarcodeMatches([])
    setLastLookupBarcode(barcode)
    try {
      const response = await fetch(`/api/barcodes/${encodeURIComponent(barcode)}`)
      if (!response.ok) {
        throw new Error(`lookup_barcode_${response.status}`)
      }
      const payload = (await response.json()) as {
        matches?: Array<{
          id?: string
          item_id?: string
          barcode?: string
          created_at?: string
        }>
      }
      const matches = (payload.matches ?? [])
        .map((match) => ({
          id: match.id?.trim() ?? '',
          item_id: match.item_id?.trim() ?? '',
          barcode: match.barcode?.trim() ?? '',
          created_at: match.created_at,
        }))
        .filter((match) => match.id !== '')
      setBarcodeMatches(matches)
      setBarcodeLookupCompleted(true)
    } catch {
      setBarcodeLookupError(
        'Barcode lookup failed. Retry or use external search for this code.'
      )
      setBarcodeLookupCompleted(false)
      setBarcodeMatches([])
    } finally {
      setBarcodeLookupBusy(false)
    }
  }, [])

  const handleAddBarcode = useCallback(async () => {
    const barcode = barcodeAddInput.trim()
    if (!selectedItemID) {
      setBarcodeAddError('No inventory item is selected for barcode attach.')
      return
    }
    if (barcode === '') {
      setBarcodeAddError('Enter a barcode before add.')
      return
    }
    setBarcodeAddBusy(true)
    setBarcodeAddError(null)
    setBarcodeAddSuccess(null)
    try {
      const response = await fetch(
        `/api/items/${encodeURIComponent(selectedItemID)}/barcodes`,
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ barcode }),
        }
      )
      if (!response.ok) {
        throw new Error(`add_barcode_${response.status}`)
      }
      setBarcodeAddSuccess(`Barcode ${barcode} added.`)
      setBarcodeAddInput('')
      if (lastLookupBarcode === barcode) {
        await lookupBarcode(barcode)
      }
    } catch {
      setBarcodeAddError('Unable to add barcode right now. Retry this action.')
    } finally {
      setBarcodeAddBusy(false)
    }
  }, [barcodeAddInput, lastLookupBarcode, lookupBarcode, selectedItemID])

  const externalSearchHref = useMemo(() => {
    if (!lastLookupBarcode) {
      return ''
    }
    return `/api/barcodes/${encodeURIComponent(
      lastLookupBarcode
    )}/external-search?source=ebay&region=AU`
  }, [lastLookupBarcode])

  const runAISuggest = useCallback(
    async (mode: 'title' | 'photo') => {
      if (!activeProfileID) {
        setAIError('Active profile is required before AI requests can run.')
        return
      }
      setAILoading(true)
      setAIError(null)
      setLastAISuggestAction(mode)
      setAIApplied(false)
      try {
        const endpoint =
          mode === 'title' ? '/api/ai/suggest/title' : '/api/ai/suggest/photo'
        const payload =
          mode === 'title'
            ? { profile_id: activeProfileID, title: aiTitleInput }
            : { profile_id: activeProfileID, image_url: aiPhotoURLInput }
        const response = await fetch(endpoint, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload),
        })
        if (!response.ok) {
          throw new Error(`ai_suggest_${mode}_${response.status}`)
        }
        const suggestion = (await response.json()) as AISuggestion
        setAISuggestion(suggestion)
      } catch {
        setAISuggestion(null)
        setAIError('AI suggestion failed. Retry the request.')
      } finally {
        setAILoading(false)
      }
    },
    [activeProfileID, aiPhotoURLInput, aiTitleInput]
  )

  return (
    <TasksProvider>
      <Header fixed>
        <Search />
        <div className='ms-auto flex items-center space-x-4'>
          <LanguageSwitch />
          <ThemeSwitch />
          <ConfigDrawer />
          <ProfileDropdown />
        </div>
      </Header>

      <Main className='space-y-4'>
        <div className='flex flex-wrap items-end justify-between gap-3'>
          <div className='space-y-2'>
            <h1 className='text-2xl font-bold tracking-tight'>{title}</h1>
            <p className='text-muted-foreground'>{description}</p>
            <p
              className='text-xs text-muted-foreground'
              data-testid='collection-context-label'
            >
              Collection context: {activeFolder}
            </p>
          </div>
          <div className='flex gap-2'>
            <Button
              type='button'
              data-testid='inventory-new-action'
              onClick={startCreateItem}
            >
              New
            </Button>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button
                  type='button'
                  variant='outline'
                  data-testid='inventory-create-menu-trigger'
                >
                  Create
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align='end'>
                <DropdownMenuItem
                  data-testid='inventory-create-menu-item'
                  onClick={startCreateItem}
                >
                  New Item
                </DropdownMenuItem>
                <DropdownMenuItem
                  data-testid='inventory-create-menu-folder'
                  onClick={() => setInlineCollectionInputOpen(true)}
                >
                  New Collection
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>

        <div className='grid grid-cols-1 gap-4 lg:grid-cols-12'>
          <Card className='lg:col-span-3 min-h-0'>
            <CardHeader>
              <CardTitle>Folders</CardTitle>
              <CardDescription>
                Browse folders before drilling into results.
              </CardDescription>
            </CardHeader>
            <CardContent className='space-y-2 min-h-0'>
              <div className='flex justify-end'>
                <Button
                  type='button'
                  variant='outline'
                  size='sm'
                  data-testid='folder-tree-add-root'
                  onClick={() => {
                    setFolderCreateParentID(null)
                    setFolderCreateName('')
                    setFolderCreateOpen(true)
                  }}
                >
                  Add Root Folder
                </Button>
              </div>
              <div
                role='tree'
                tabIndex={0}
                aria-label='Inventory folders'
                data-testid='inventory-folder-tree'
                className='h-[26rem] max-h-[26rem] overflow-x-auto overflow-y-auto rounded-md border p-2'
              >
                <div
                  className='min-w-full w-max space-y-2'
                  data-testid='inventory-folder-tree-scroll-region'
                >
                  {renderFolderTree(folderTree)}
                </div>
              </div>
              <Dialog
                open={folderCreateOpen}
                onOpenChange={(open) => {
                  setFolderCreateOpen(open)
                  if (!open) {
                    setFolderCreateName('')
                    setFolderCreateParentID(null)
                  }
                }}
              >
                <DialogContent>
                  <DialogHeader>
                    <DialogTitle>
                      {folderCreateParentID ? 'Add Child Folder' : 'Add Root Folder'}
                    </DialogTitle>
                  </DialogHeader>
                  <Input
                    data-testid='folder-tree-name-input'
                    placeholder='Folder name'
                    value={folderCreateName}
                    onChange={(event) => setFolderCreateName(event.target.value)}
                  />
                  <DialogFooter>
                    <Button
                      type='button'
                      variant='outline'
                      data-testid='folder-tree-create-cancel'
                      onClick={() => {
                        setFolderCreateOpen(false)
                        setFolderCreateName('')
                        setFolderCreateParentID(null)
                      }}
                    >
                      Cancel
                    </Button>
                    <Button
                      type='button'
                      data-testid='folder-tree-create-submit'
                      onClick={() => {
                        const name = folderCreateName.trim()
                        if (name === '') {
                          return
                        }
                        setFolderTree((previous) => {
                          const newNode: FolderNode = {
                            id: buildUniqueFolderID(name, previous),
                            name,
                          }
                          if (!folderCreateParentID) {
                            return [...previous, newNode]
                          }
                          return addChildFolder(previous, folderCreateParentID, newNode)
                        })
                        if (folderCreateParentID) {
                          setExpandedNodeIDs((previous) => {
                            const next = new Set(previous)
                            next.add(folderCreateParentID)
                            return next
                          })
                        }
                        setActiveFolder(name)
                        setFolderCreateOpen(false)
                        setFolderCreateName('')
                        setFolderCreateParentID(null)
                      }}
                    >
                      Create
                    </Button>
                  </DialogFooter>
                </DialogContent>
              </Dialog>
            </CardContent>
          </Card>

          <Card className='lg:col-span-9'>
            <CardHeader>
              <CardTitle>Collection Browser</CardTitle>
            </CardHeader>
            <CardContent className='space-y-4'>
              <p className='text-sm text-muted-foreground'>
                Folders: <strong>{summary.folders}</strong>{' '}
                <span className='mx-2'>Items: <strong>{summary.items}</strong></span>
                Active Brand: <strong>{summary.activeBrand}</strong>{' '}
                <span className='mx-2'>
                  Active Category: <strong>{summary.activeCategory}</strong>
                </span>
                <span className='mx-2'>
                  Active Context:{' '}
                  <strong data-testid='collection-active-context'>
                    {summary.activeContext}
                  </strong>
                </span>
                <span className='mx-2'>
                  Selected Item:{' '}
                  <strong data-testid='collection-selected-item'>
                    {selectedItemContext}
                  </strong>
                </span>
              </p>
              <div className='rounded-md border p-3' data-testid='collection-inline-picker'>
                <div className='grid gap-2 md:grid-cols-[minmax(0,1fr)_auto_auto] md:items-center'>
                  <select
                    className='h-9 rounded-md border bg-background px-2 text-sm'
                    value={activeWorkspaceCollection}
                    onChange={(event) => {
                      const selected = event.target.value
                      setActiveWorkspaceCollection(selected)
                      setActiveFolder(selected)
                    }}
                  >
                    {workspaceCollections.map((collection) => (
                      <option
                        key={collection}
                        value={collection}
                        data-testid={`collection-inline-picker-option-${collectionKey(collection)}`}
                      >
                        {collection}
                      </option>
                    ))}
                  </select>
                  <span
                    className='text-sm text-muted-foreground'
                    data-testid='collection-inline-picker-selected'
                  >
                    {activeWorkspaceCollection}
                  </span>
                  <Button
                    type='button'
                    variant='outline'
                    data-testid='collection-inline-add-new'
                    onClick={() => setInlineCollectionInputOpen((open) => !open)}
                  >
                    + New Collection
                  </Button>
                </div>
                {inlineCollectionInputOpen ? (
                  <div className='mt-2 flex gap-2'>
                    <Input
                      data-testid='collection-inline-new-name'
                      placeholder='Collection name'
                      value={inlineCollectionName}
                      onChange={(event) => setInlineCollectionName(event.target.value)}
                    />
                    <Button
                      type='button'
                      data-testid='collection-inline-save'
                      onClick={() => {
                        const created = addCollection(inlineCollectionName)
                        if (created) {
                          setActiveFolder(created)
                        }
                        setInlineCollectionName('')
                        setInlineCollectionInputOpen(false)
                      }}
                    >
                      Save
                    </Button>
                  </div>
                ) : null}
              </div>
              {loading ? (
                <div
                  className='rounded-md border p-6 text-sm text-muted-foreground'
                  data-testid='inventory-loading'
                >
                  Loading inventory...
                </div>
              ) : null}
              {loadError ? (
                <div
                  className='rounded-md border border-destructive/40 bg-destructive/10 p-4 text-sm'
                  data-testid='inventory-load-error'
                >
                  <p className='font-medium'>Inventory load failed</p>
                  <p className='mt-1 text-muted-foreground'>{loadError}</p>
                  <Button
                    className='mt-3'
                    variant='outline'
                    size='sm'
                    onClick={() => void loadInventoryItems()}
                  >
                    Retry
                  </Button>
                </div>
              ) : null}
              <TasksTable
                data={tableData}
                routePath={routePath}
                currentRecordID={selectedItemID}
                onRecordFocus={(itemID, recordID) => {
                  const matchedItem =
                    inventoryItems.find((item) => item.id === itemID) ??
                    inventoryItems.find((item) => item.part_number === recordID) ??
                    null
                  setItemEditorMode('edit')
                  setItemSaveError(null)
                  setItemSaveSuccess(null)
                  selectInventoryItem(matchedItem)
                }}
              />
              {isInventoryRoute ? (
                <>
                  <section
                    className='space-y-3 rounded-md border p-4'
                    data-testid='inventory-item-editor'
                  >
                    <div className='flex flex-wrap items-start justify-between gap-3'>
                      <div>
                        <h3 className='text-base font-semibold'>Item Save</h3>
                        <p className='text-sm text-muted-foreground'>
                          Create a new inventory item or edit the selected persisted item.
                        </p>
                      </div>
                      <div className='flex gap-2'>
                        <Button
                          type='button'
                          variant={itemEditorMode === 'create' ? 'default' : 'outline'}
                          data-testid='inventory-item-create-mode'
                          onClick={startCreateItem}
                        >
                          Create New
                        </Button>
                        <Button
                          type='button'
                          variant='outline'
                          data-testid='inventory-item-edit-selected'
                          disabled={!selectedInventoryItem}
                          onClick={() => {
                            if (!selectedInventoryItem) {
                              return
                            }
                            setItemEditorMode('edit')
                            setItemDraft(inventoryItemToDraft(selectedInventoryItem))
                            setItemSaveError(null)
                            setItemSaveSuccess(null)
                          }}
                        >
                          Edit Selected
                        </Button>
                      </div>
                    </div>
                    <div className='grid grid-cols-1 gap-3 md:grid-cols-2'>
                      <div className='space-y-2'>
                        <label className='text-sm font-medium' htmlFor='inventory-item-part-number'>
                          Part Number
                        </label>
                        <Input
                          id='inventory-item-part-number'
                          data-testid='inventory-item-part-number'
                          value={itemDraft.part_number}
                          onChange={(event) =>
                            setItemDraft((current) => ({
                              ...current,
                              part_number: event.target.value,
                            }))
                          }
                        />
                      </div>
                      <div className='space-y-2'>
                        <label className='text-sm font-medium' htmlFor='inventory-item-title'>
                          Title
                        </label>
                        <Input
                          id='inventory-item-title'
                          data-testid='inventory-item-title'
                          value={itemDraft.title}
                          onChange={(event) =>
                            setItemDraft((current) => ({
                              ...current,
                              title: event.target.value,
                            }))
                          }
                        />
                      </div>
                      <div className='space-y-2'>
                        <label className='text-sm font-medium' htmlFor='inventory-item-brand'>
                          Brand
                        </label>
                        <Input
                          id='inventory-item-brand'
                          data-testid='inventory-item-brand'
                          value={itemDraft.brand}
                          onChange={(event) =>
                            setItemDraft((current) => ({
                              ...current,
                              brand: event.target.value,
                            }))
                          }
                        />
                      </div>
                      <div className='space-y-2'>
                        <label className='text-sm font-medium' htmlFor='inventory-item-category'>
                          Category
                        </label>
                        <Input
                          id='inventory-item-category'
                          data-testid='inventory-item-category'
                          value={itemDraft.category}
                          onChange={(event) =>
                            setItemDraft((current) => ({
                              ...current,
                              category: event.target.value,
                            }))
                          }
                        />
                      </div>
                    </div>
                    <div className='space-y-2'>
                      <label className='text-sm font-medium' htmlFor='inventory-item-description'>
                        Description
                      </label>
                      <Input
                        id='inventory-item-description'
                        data-testid='inventory-item-description'
                        value={itemDraft.description}
                        onChange={(event) =>
                          setItemDraft((current) => ({
                            ...current,
                            description: event.target.value,
                          }))
                        }
                      />
                    </div>
                    <div className='flex flex-wrap items-center gap-2'>
                      <Button
                        type='button'
                        data-testid='inventory-item-save'
                        disabled={itemSaveBusy}
                        onClick={() => void handleSaveItem()}
                      >
                        {itemEditorMode === 'create' ? 'Create Item' : 'Save Changes'}
                      </Button>
                      {itemEditorMode === 'create' ? (
                        <span
                          className='text-sm text-muted-foreground'
                          data-testid='inventory-item-editor-mode'
                        >
                          Creating new item draft.
                        </span>
                      ) : (
                        <span
                          className='text-sm text-muted-foreground'
                          data-testid='inventory-item-editor-mode'
                        >
                          Editing selected item: {selectedItemContext}
                        </span>
                      )}
                    </div>
                    {itemSaveError ? (
                      <div
                        className='rounded-md border border-destructive/40 bg-destructive/10 p-3 text-sm'
                        data-testid='inventory-item-save-error'
                      >
                        {itemSaveError}
                      </div>
                    ) : null}
                    {itemSaveSuccess ? (
                      <div
                        className='rounded-md border border-emerald-500/40 bg-emerald-500/10 p-3 text-sm'
                        data-testid='inventory-item-save-success'
                      >
                        {itemSaveSuccess}
                      </div>
                    ) : null}
                  </section>
                  <section
                    className='space-y-3 rounded-md border p-4'
                    data-testid='inventory-photos-section'
                  >
                    <div>
                      <h3 className='text-base font-semibold'>Photos</h3>
                      <p className='text-sm text-muted-foreground'>
                        Review item media and inspect photos in fullscreen mode.
                      </p>
                    </div>
                    <div className='flex flex-wrap items-center gap-2'>
                      <Button
                        type='button'
                        variant='outline'
                        data-testid='inventory-camera-take-photo'
                        onClick={() => void handleTakePhoto()}
                      >
                        Take Photo
                      </Button>
                      <input
                        type='file'
                        accept='image/*'
                        data-testid='inventory-photo-upload-input'
                        onChange={(event) => {
                          const file = event.target.files?.[0] ?? null
                          void handlePhotoUpload(file)
                          event.currentTarget.value = ''
                        }}
                      />
                      {photosBusy ? (
                        <span className='text-xs text-muted-foreground'>Working...</span>
                      ) : null}
                    </div>
                    {cameraError ? (
                      <div
                        className='rounded-md border border-destructive/40 bg-destructive/10 p-2 text-sm'
                        data-testid='inventory-camera-error'
                      >
                        {cameraError}
                      </div>
                    ) : null}
                    {cameraSuccess ? (
                      <div
                        className='rounded-md border border-emerald-500/40 bg-emerald-500/10 p-2 text-sm'
                        data-testid='inventory-camera-success'
                      >
                        {cameraSuccess}
                      </div>
                    ) : null}
                    {photosLoading ? (
                      <div
                        className='rounded-md border p-3 text-sm text-muted-foreground'
                        data-testid='inventory-photos-loading'
                      >
                        Loading photos...
                      </div>
                    ) : null}
                    {photosError ? (
                      <div
                        className='rounded-md border border-destructive/40 bg-destructive/10 p-3 text-sm'
                        data-testid='inventory-photos-error'
                      >
                        <p className='font-medium'>Photos are unavailable.</p>
                        <p className='mt-1 text-muted-foreground'>{photosError}</p>
                        <Button
                          className='mt-3'
                          size='sm'
                          variant='outline'
                          onClick={() => void loadInventoryPhotos()}
                        >
                          Retry
                        </Button>
                      </div>
                    ) : null}
                    {!photosLoading && !photosError && inventoryPhotos.length === 0 ? (
                      <div
                        className='rounded-md border border-dashed p-3 text-sm text-muted-foreground'
                        data-testid='inventory-photos-empty'
                      >
                        No photos yet for the selected item. Upload an image to begin.
                      </div>
                    ) : null}
                    {!photosLoading && !photosError && inventoryPhotos.length > 0 ? (
                      <>
                        <div className='grid grid-cols-1 gap-3 md:grid-cols-3'>
                          {inventoryPhotos.map((photo, index) => (
                            <button
                              key={photo.id}
                              type='button'
                              className='overflow-hidden rounded-md border text-left transition hover:border-primary/60'
                              data-testid='inventory-photo-thumb'
                              onClick={() => setSelectedPhotoIndex(index)}
                            >
                              <img
                                src={`/api/items/${encodeURIComponent(
                                  selectedItemID
                                )}/photos/${encodeURIComponent(photo.id)}/file?variant=preview`}
                                alt={photo.filename}
                                className='h-32 w-full object-cover'
                              />
                              <div className='p-2 text-sm'>{photo.filename}</div>
                            </button>
                          ))}
                        </div>
                        <div className='space-y-2'>
                          {inventoryPhotos.map((photo) => (
                            <div
                              key={`row-${photo.id}`}
                              className='flex flex-wrap items-center justify-between gap-2 rounded-md border p-2'
                              data-testid='inventory-photo-row'
                            >
                              <div className='flex items-center gap-2 text-sm'>
                                <span>{photo.filename}</span>
                                {photo.is_primary ? (
                                  <span
                                    className='rounded bg-primary/10 px-2 py-0.5 text-xs text-primary'
                                    data-testid='inventory-photo-primary-badge'
                                  >
                                    Primary
                                  </span>
                                ) : null}
                              </div>
                              <div className='flex gap-2'>
                                <Button
                                  size='sm'
                                  variant='outline'
                                  data-testid='inventory-photo-set-primary'
                                  onClick={() => void handleSetPrimaryPhoto(photo.id)}
                                >
                                  Set Primary
                                </Button>
                                <Button
                                  size='sm'
                                  variant='outline'
                                  data-testid='inventory-photo-delete'
                                  onClick={() => void handleDeletePhoto(photo.id)}
                                >
                                  Delete
                                </Button>
                              </div>
                            </div>
                          ))}
                        </div>
                      </>
                    ) : null}
                  </section>
                  <section
                    className='space-y-3 rounded-md border p-4'
                    data-testid='inventory-barcodes-section'
                  >
                    <div>
                      <h3 className='text-base font-semibold'>Barcodes</h3>
                      <p className='text-sm text-muted-foreground'>
                        Add barcodes to the selected item, run local lookup, and
                        continue with external fallback when there is no local
                        match.
                      </p>
                    </div>
                    <div className='grid grid-cols-1 gap-3 md:grid-cols-2'>
                      <div className='space-y-2'>
                        <Input
                          placeholder='Enter barcode to attach'
                          data-testid='inventory-barcodes-add-input'
                          value={barcodeAddInput}
                          onChange={(event) =>
                            setBarcodeAddInput(event.target.value)
                          }
                        />
                        <Button
                          data-testid='inventory-barcodes-add-button'
                          onClick={() => void handleAddBarcode()}
                          disabled={barcodeAddBusy || selectedItemID === ''}
                        >
                          Add Barcode
                        </Button>
                        {barcodeAddError ? (
                          <div
                            className='rounded-md border border-destructive/40 bg-destructive/10 p-2 text-sm'
                            data-testid='inventory-barcodes-add-error'
                          >
                            {barcodeAddError}
                          </div>
                        ) : null}
                        {barcodeAddSuccess ? (
                          <div
                            className='rounded-md border border-emerald-500/40 bg-emerald-500/10 p-2 text-sm'
                            data-testid='inventory-barcodes-add-success'
                          >
                            {barcodeAddSuccess}
                          </div>
                        ) : null}
                      </div>
                      <div className='space-y-2'>
                        <Input
                          placeholder='Lookup barcode'
                          data-testid='inventory-barcodes-lookup-input'
                          value={barcodeLookupInput}
                          onChange={(event) =>
                            setBarcodeLookupInput(event.target.value)
                          }
                        />
                        <Button
                          data-testid='inventory-barcodes-lookup-button'
                          variant='outline'
                          onClick={() => void lookupBarcode(barcodeLookupInput)}
                          disabled={barcodeLookupBusy}
                        >
                          Lookup Barcode
                        </Button>
                      </div>
                    </div>
                    {barcodeLookupBusy ? (
                      <div
                        className='rounded-md border p-3 text-sm text-muted-foreground'
                        data-testid='inventory-barcodes-lookup-loading'
                      >
                        Looking up barcode...
                      </div>
                    ) : null}
                    {barcodeLookupError ? (
                      <div
                        className='rounded-md border border-destructive/40 bg-destructive/10 p-3 text-sm'
                        data-testid='inventory-barcodes-lookup-error'
                      >
                        <p className='font-medium'>Barcode lookup failed.</p>
                        <p className='mt-1 text-muted-foreground'>
                          {barcodeLookupError}
                        </p>
                        <Button
                          className='mt-3'
                          size='sm'
                          variant='outline'
                          data-testid='inventory-barcodes-lookup-retry'
                          onClick={() => void lookupBarcode(lastLookupBarcode)}
                        >
                          Retry
                        </Button>
                      </div>
                    ) : null}
                    {barcodeLookupCompleted &&
                    !barcodeLookupBusy &&
                    !barcodeLookupError &&
                    barcodeMatches.length === 0 ? (
                      <div
                        className='rounded-md border border-dashed p-3 text-sm'
                        data-testid='inventory-barcodes-lookup-empty'
                      >
                        <p className='font-medium'>No local barcode match.</p>
                        <p className='mt-1 text-muted-foreground'>
                          Continue with an external provider search.
                        </p>
                        {externalSearchHref ? (
                          <a
                            className='mt-2 inline-flex text-sm text-primary underline'
                            href={externalSearchHref}
                            target='_blank'
                            rel='noreferrer'
                            data-testid='inventory-barcodes-external-search-link'
                          >
                            Search eBay
                          </a>
                        ) : null}
                      </div>
                    ) : null}
                    {barcodeMatches.length > 0 ? (
                      <div className='space-y-2'>
                        {barcodeMatches.map((match) => (
                          <div
                            key={match.id}
                            className='rounded-md border p-2 text-sm'
                            data-testid='inventory-barcodes-match-row'
                          >
                            <p>
                              <strong>Barcode:</strong> {match.barcode}
                            </p>
                            <p>
                              <strong>Item:</strong> {match.item_id}
                            </p>
                          </div>
                        ))}
                      </div>
                    ) : null}
                  </section>
                  <section
                    className='space-y-3 rounded-md border p-4'
                    data-testid='inventory-ai-assist-section'
                  >
                    <div>
                      <h3 className='text-base font-semibold'>AI Assist</h3>
                      <p className='text-sm text-muted-foreground'>
                        Generate title and photo suggestions with explicit confirm-before-apply.
                      </p>
                    </div>
                    <div className='grid grid-cols-1 gap-3 md:grid-cols-2'>
                      <div className='space-y-2'>
                        <Input
                          placeholder='Paste listing title'
                          data-testid='inventory-ai-title-input'
                          value={aiTitleInput}
                          onChange={(event) => setAITitleInput(event.target.value)}
                        />
                        <Button
                          data-testid='inventory-ai-suggest-title'
                          onClick={() => void runAISuggest('title')}
                          disabled={aiLoading || aiTitleInput.trim() === ''}
                        >
                          Suggest from Title
                        </Button>
                      </div>
                      <div className='space-y-2'>
                        <Input
                          placeholder='Paste photo URL'
                          data-testid='inventory-ai-photo-url-input'
                          value={aiPhotoURLInput}
                          onChange={(event) => setAIPhotoURLInput(event.target.value)}
                        />
                        <Button
                          data-testid='inventory-ai-suggest-photo'
                          onClick={() => void runAISuggest('photo')}
                          disabled={aiLoading || aiPhotoURLInput.trim() === ''}
                        >
                          Suggest from Photo
                        </Button>
                      </div>
                    </div>
                    {aiLoading ? (
                      <div
                        className='rounded-md border p-3 text-sm text-muted-foreground'
                        data-testid='inventory-ai-loading'
                      >
                        Running AI suggestion...
                      </div>
                    ) : null}
                    {aiError ? (
                      <div
                        className='rounded-md border border-destructive/40 bg-destructive/10 p-3 text-sm'
                        data-testid='inventory-ai-error'
                      >
                        <p className='font-medium'>AI suggestion failed.</p>
                        <p className='mt-1 text-muted-foreground'>{aiError}</p>
                        <Button
                          className='mt-3'
                          variant='outline'
                          size='sm'
                          data-testid='inventory-ai-retry'
                          onClick={() => {
                            if (lastAISuggestAction) {
                              void runAISuggest(lastAISuggestAction)
                            }
                          }}
                        >
                          Retry
                        </Button>
                      </div>
                    ) : null}
                    {aiSuggestion ? (
                      <div
                        className='rounded-md border p-3 text-sm'
                        data-testid='inventory-ai-suggestion'
                      >
                        <p>
                          <strong>Part:</strong> {aiSuggestion.part_number || 'n/a'}
                        </p>
                        <p>
                          <strong>Brand:</strong> {aiSuggestion.brand || 'n/a'}
                        </p>
                        <p>
                          <strong>Title:</strong> {aiSuggestion.title || 'n/a'}
                        </p>
                        <p>
                          <strong>Confidence:</strong>{' '}
                          {(aiSuggestion.confidence ?? 0).toFixed(2)}
                        </p>
                        <Button
                          className='mt-3'
                          data-testid='inventory-ai-apply'
                          ref={aiApplyTriggerRef}
                          onClick={() => setAIConfirmOpen(true)}
                        >
                          Apply Suggestion
                        </Button>
                      </div>
                    ) : null}
                    {aiApplied ? (
                      <div
                        className='rounded-md border border-emerald-500/50 bg-emerald-500/10 p-3 text-sm'
                        data-testid='inventory-ai-applied-banner'
                      >
                        Suggestion applied to draft item fields.
                      </div>
                    ) : null}
                  </section>
                </>
              ) : null}
            </CardContent>
          </Card>
        </div>
        <Dialog
          open={selectedPhotoIndex !== null && inventoryPhotos.length > 0}
          onOpenChange={(open) => {
            if (!open) {
              setSelectedPhotoIndex(null)
            }
          }}
        >
          <DialogContent
            className='max-w-4xl'
            data-testid='inventory-photo-fullscreen'
          >
            <DialogHeader>
              <DialogTitle>{selectedPhoto?.filename ?? 'Photo Viewer'}</DialogTitle>
            </DialogHeader>
            {selectedPhoto ? (
              <img
                src={`/api/items/${encodeURIComponent(
                  selectedItemID
                )}/photos/${encodeURIComponent(selectedPhoto.id)}/file?variant=original`}
                alt={selectedPhoto.filename}
                className='max-h-[70vh] w-full rounded-md object-contain'
              />
            ) : null}
            <div className='flex justify-end gap-2'>
              <Button
                type='button'
                variant='outline'
                data-testid='inventory-photo-prev'
                onClick={() => {
                  if (selectedPhotoIndex === null) {
                    return
                  }
                  const nextIndex =
                    (selectedPhotoIndex - 1 + inventoryPhotos.length) %
                    inventoryPhotos.length
                  setSelectedPhotoIndex(nextIndex)
                }}
              >
                Previous
              </Button>
              <Button
                type='button'
                variant='outline'
                data-testid='inventory-photo-next'
                onClick={() => {
                  if (selectedPhotoIndex === null) {
                    return
                  }
                  const nextIndex =
                    (selectedPhotoIndex + 1) % inventoryPhotos.length
                  setSelectedPhotoIndex(nextIndex)
                }}
              >
                Next
              </Button>
              <Button
                type='button'
                data-testid='inventory-photo-fullscreen-close'
                onClick={() => setSelectedPhotoIndex(null)}
              >
                Close
              </Button>
            </div>
          </DialogContent>
        </Dialog>
        <Dialog
          open={aiConfirmOpen}
          onOpenChange={(open) => {
            setAIConfirmOpen(open)
            if (!open) {
              requestAnimationFrame(() => {
                aiApplyTriggerRef.current?.focus()
              })
            }
          }}
        >
          <DialogContent data-testid='inventory-ai-confirm-dialog'>
            <DialogHeader>
              <DialogTitle>Confirm AI Apply</DialogTitle>
            </DialogHeader>
            <p className='text-sm text-muted-foreground'>
              AI suggestions require explicit confirmation before apply.
            </p>
            <DialogFooter>
              <Button variant='outline' onClick={() => setAIConfirmOpen(false)}>
                Cancel
              </Button>
              <Button
                onClick={() => {
                  setItemDraft((current) => ({
                    ...current,
                    part_number:
                      aiSuggestion?.part_number?.trim() || current.part_number,
                    brand: aiSuggestion?.brand?.trim() || current.brand,
                    title: aiSuggestion?.title?.trim() || current.title,
                  }))
                  setAIApplied(true)
                  setAIConfirmOpen(false)
                }}
              >
                Confirm Apply
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </Main>
      <TasksDialogs />
    </TasksProvider>
  )
}
