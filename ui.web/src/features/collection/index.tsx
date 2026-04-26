import {
  type ChangeEvent,
  type DragEvent as ReactDragEvent,
  type KeyboardEvent,
  type PointerEvent as ReactPointerEvent,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from 'react'
import {
  Barcode,
  ChevronRight,
  Ellipsis,
  GripVertical,
  Images,
  ListChecks,
  Plus,
} from 'lucide-react'
import { createPortal } from 'react-dom'
import { cn } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
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
import { Input } from '@/components/ui/input'
import { Separator } from '@/components/ui/separator'
import {
  Sheet,
  SheetClose,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'
import { ConfigDrawer } from '@/components/config-drawer'
import { LanguageSwitch } from '@/components/language-switch'
import { Header, HeaderTitle } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import {
  type WorkspaceCollectionItem,
  useWorkspaceCollections,
} from '@/features/collections/use-workspace-collections'
import {
  TasksDialogs,
  type TasksDialogType,
} from '@/features/tasks/components/tasks-dialogs'
import { TasksProvider } from '@/features/tasks/components/tasks-provider'
import { TasksTable } from '@/features/tasks/components/tasks-table'
import { type Task } from '@/features/tasks/data/schema'
import { tasks } from '@/features/tasks/data/tasks'

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
  category?: string
  itemCount?: number
  secondaryLabel?: string
  statusBadge?: string
  children?: FolderNode[]
}

type ProfileSettingsPayload = {
  settings?: Record<string, string>
}

type FolderDropTarget =
  | { kind: 'child'; nodeID: string }
  | { kind: 'before'; nodeID: string }
  | { kind: 'after'; nodeID: string }
  | { kind: 'root' }

function folderDropTargetsEqual(
  left: FolderDropTarget | null,
  right: FolderDropTarget | null
) {
  if (left === right) {
    return true
  }
  if (!left || !right) {
    return false
  }
  if (left.kind !== right.kind) {
    return false
  }
  if (left.kind === 'root' && right.kind === 'root') {
    return true
  }
  if (left.kind === 'root' || right.kind === 'root') {
    return false
  }
  return left.nodeID === right.nodeID
}

function findFolderNodeByID(
  nodes: FolderNode[],
  id: string
): FolderNode | null {
  for (const node of nodes) {
    if (node.id === id) {
      return node
    }
    const childMatch = findFolderNodeByID(node.children ?? [], id)
    if (childMatch) {
      return childMatch
    }
  }
  return null
}

const inventoryTreeStorageKey = 'cabinet.inventory.tree-state'
const inventoryWorkspaceSettingsStorageKeyPrefix =
  'cabinet.inventory.workspace-settings.v1.'
const inventoryFolderTreeSettingsKey = 'inventory.folder-tree.v1'
const inventoryItemFolderAssignmentsStorageKey =
  'cabinet.inventory.item-folder-assignments.v1'
const inventoryItemDragDataType = 'application/x-cabinet-inventory-item-id'
const folderDragDataType = 'application/x-cabinet-folder-id'

const folderCategoryOptions = [
  'Catalog',
  'Watch',
  'Wishlist',
  'Store',
  'Warehouse',
  'Archive',
  'General',
]

const initialFolderTree: FolderNode[] = [
  {
    id: 'all-items',
    name: 'All Items',
    category: 'Catalog',
    itemCount: 124,
    secondaryLabel: 'Entire catalog',
    statusBadge: 'Live',
  },
  {
    id: 'watch-list',
    name: 'Watch List',
    category: 'Watch',
    itemCount: 12,
    secondaryLabel: 'Needs review',
    statusBadge: 'Watch',
  },
  {
    id: 'wishlist-focus',
    name: 'Wishlist Focus',
    category: 'Wishlist',
    itemCount: 7,
    secondaryLabel: 'Priority targets',
    statusBadge: 'Goal',
  },
  {
    id: 'store-1',
    name: 'Store 1',
    category: 'Store',
    itemCount: 18,
    secondaryLabel: 'Aisle A',
  },
  {
    id: 'store-2',
    name: 'Store 2',
    category: 'Store',
    itemCount: 14,
    secondaryLabel: 'Aisle B',
  },
  {
    id: 'store-3',
    name: 'Store 3',
    category: 'Store',
    itemCount: 11,
    secondaryLabel: 'Aisle C',
  },
  {
    id: 'store-4',
    name: 'Store 4',
    category: 'Store',
    itemCount: 9,
    secondaryLabel: 'Overflow shelf',
  },
  {
    id: 'store-5',
    name: 'Store 5',
    category: 'Store',
    itemCount: 8,
    secondaryLabel: 'Back room',
  },
  {
    id: 'store-6',
    name: 'Store 6',
    category: 'Store',
    itemCount: 10,
    secondaryLabel: 'Bin 6',
  },
  {
    id: 'store-7',
    name: 'Store 7',
    category: 'Store',
    itemCount: 6,
    secondaryLabel: 'Bin 7',
  },
  {
    id: 'store-8',
    name: 'Store 8',
    category: 'Store',
    itemCount: 5,
    secondaryLabel: 'Bin 8',
  },
  {
    id: 'store-9',
    name: 'Store 9',
    category: 'Store',
    itemCount: 4,
    secondaryLabel: 'Bin 9',
  },
  {
    id: 'store-10',
    name: 'Store 10',
    category: 'Store',
    itemCount: 3,
    secondaryLabel: 'Bin 10',
  },
  {
    id: 'warehouses',
    name: 'Warehouses',
    category: 'Warehouse',
    itemCount: 39,
    secondaryLabel: 'Bulk storage',
    statusBadge: 'Cold',
    children: [
      {
        id: 'warehouse-1',
        name: 'Warehouse 1',
        category: 'Warehouse',
        itemCount: 15,
        secondaryLabel: 'Pallet zone A',
      },
      {
        id: 'warehouse-2',
        name: 'Warehouse 2',
        category: 'Warehouse',
        itemCount: 13,
        secondaryLabel: 'Pallet zone B',
      },
      {
        id: 'warehouse-3',
        name: 'Warehouse 3',
        category: 'Warehouse',
        itemCount: 11,
        secondaryLabel: 'Pallet zone C',
      },
    ],
  },
  {
    id: 'archive-a',
    name: 'Archive A',
    category: 'Archive',
    itemCount: 2,
    secondaryLabel: 'Retired stock',
  },
  {
    id: 'archive-b',
    name: 'Archive B',
    category: 'Archive',
    itemCount: 2,
    secondaryLabel: 'Retired stock',
  },
  {
    id: 'archive-c',
    name: 'Archive C',
    category: 'Archive',
    itemCount: 1,
    secondaryLabel: 'Retired stock',
  },
  {
    id: 'archive-d',
    name: 'Archive D',
    category: 'Archive',
    itemCount: 1,
    secondaryLabel: 'Retired stock',
  },
  {
    id: 'archive-e',
    name: 'Archive E',
    category: 'Archive',
    itemCount: 1,
    secondaryLabel: 'Retired stock',
  },
  {
    id: 'archive-f',
    name: 'Archive F',
    category: 'Archive',
    itemCount: 1,
    secondaryLabel: 'Retired stock',
  },
  {
    id: 'archive-g',
    name: 'Archive G',
    category: 'Archive',
    itemCount: 1,
    secondaryLabel: 'Retired stock',
  },
  {
    id: 'archive-h',
    name: 'Archive H',
    category: 'Archive',
    itemCount: 1,
    secondaryLabel: 'Retired stock',
  },
  {
    id: 'archive-i',
    name: 'Archive I',
    category: 'Archive',
    itemCount: 1,
    secondaryLabel: 'Retired stock',
  },
  {
    id: 'archive-j',
    name: 'Archive J',
    category: 'Archive',
    itemCount: 1,
    secondaryLabel: 'Retired stock',
  },
  {
    id: 'archive-k',
    name: 'Archive K',
    category: 'Archive',
    itemCount: 1,
    secondaryLabel: 'Retired stock',
  },
  {
    id: 'archive-l',
    name: 'Archive L',
    category: 'Archive',
    itemCount: 1,
    secondaryLabel: 'Retired stock',
  },
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

function buildWorkspaceCollectionFilterOptions(
  collections: string[]
): Array<FolderNode & { level: number }> {
  return collections.map((name) => ({
    id: slugifyFolderName(name) || 'collection',
    name,
    level: 0,
  }))
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

function removeFolderNode(
  nodes: FolderNode[],
  targetID: string
): { removed: FolderNode | null; next: FolderNode[] } {
  let removed: FolderNode | null = null
  const next = nodes.flatMap((node) => {
    if (node.id === targetID) {
      removed = node
      return []
    }
    if (node.children?.length) {
      const result = removeFolderNode(node.children, targetID)
      if (result.removed) {
        removed = result.removed
        return [{ ...node, children: result.next }]
      }
    }
    return [node]
  })
  return { removed, next }
}

function folderNodeContainsID(node: FolderNode, targetID: string): boolean {
  if (node.id === targetID) {
    return true
  }
  return (node.children ?? []).some((child) =>
    folderNodeContainsID(child, targetID)
  )
}

function moveFolderNode(
  nodes: FolderNode[],
  draggedID: string,
  targetID: string
): FolderNode[] {
  if (draggedID === targetID) {
    return nodes
  }
  const { removed, next } = removeFolderNode(nodes, draggedID)
  if (!removed) {
    return nodes
  }
  if (folderNodeContainsID(removed, targetID)) {
    return nodes
  }
  return addChildFolder(next, targetID, removed)
}

function isInvalidFolderDropTarget(
  nodes: FolderNode[],
  draggedID: string,
  targetID: string
): boolean {
  if (draggedID === targetID) {
    return true
  }

  const draggedNode = findFolderNodeByID(nodes, draggedID)
  if (!draggedNode) {
    return true
  }

  return folderNodeContainsID(draggedNode, targetID)
}

function insertFolderNodeRelative(
  nodes: FolderNode[],
  targetID: string,
  nodeToInsert: FolderNode,
  position: 'before' | 'after'
): FolderNode[] {
  return nodes.flatMap((node) => {
    if (node.id === targetID) {
      return position === 'before' ? [nodeToInsert, node] : [node, nodeToInsert]
    }
    if (node.children?.length) {
      return [
        {
          ...node,
          children: insertFolderNodeRelative(
            node.children,
            targetID,
            nodeToInsert,
            position
          ),
        },
      ]
    }
    return [node]
  })
}

function moveFolderNodeRelative(
  nodes: FolderNode[],
  draggedID: string,
  targetID: string,
  position: 'before' | 'after'
): FolderNode[] {
  if (draggedID === targetID) {
    return nodes
  }
  const { removed, next } = removeFolderNode(nodes, draggedID)
  if (!removed) {
    return nodes
  }
  if (folderNodeContainsID(removed, targetID)) {
    return nodes
  }
  return insertFolderNodeRelative(next, targetID, removed, position)
}

function moveFolderNodeToRoot(
  nodes: FolderNode[],
  draggedID: string
): FolderNode[] {
  const { removed, next } = removeFolderNode(nodes, draggedID)
  if (!removed) {
    return nodes
  }
  return [...next, removed]
}

function updateFolderNodeByID(
  nodes: FolderNode[],
  targetID: string,
  updater: (node: FolderNode) => FolderNode
): FolderNode[] {
  return nodes.map((node) => {
    if (node.id === targetID) {
      return updater(node)
    }
    if (node.children?.length) {
      return {
        ...node,
        children: updateFolderNodeByID(node.children, targetID, updater),
      }
    }
    return node
  })
}

function sortRootFolderNodesAlphabetically(nodes: FolderNode[]): FolderNode[] {
  if (nodes.length <= 1) {
    return nodes
  }

  const pinnedRoot = nodes.find((node) => node.id === 'all-items') ?? null
  const sortable = nodes.filter((node) => node.id !== 'all-items')
  const sorted = [...sortable].sort((left, right) =>
    left.name.localeCompare(right.name, undefined, { sensitivity: 'base' })
  )

  return pinnedRoot ? [pinnedRoot, ...sorted] : sorted
}

function inventoryWorkspaceSettingsStorageKey(profileID: string): string {
  return `${inventoryWorkspaceSettingsStorageKeyPrefix}${profileID.trim()}`
}

function folderTreeContainsName(
  nodes: FolderNode[],
  targetName: string
): boolean {
  return nodes.some(
    (node) =>
      node.name === targetName ||
      (node.children?.length
        ? folderTreeContainsName(node.children, targetName)
        : false)
  )
}

function sanitizeFolderNodes(value: unknown): FolderNode[] {
  if (!Array.isArray(value)) {
    return initialFolderTree
  }

  const walk = (nodes: unknown[]): FolderNode[] =>
    nodes.flatMap((entry) => {
      if (!entry || typeof entry !== 'object') {
        return []
      }

      const candidate = entry as {
        id?: unknown
        name?: unknown
        category?: unknown
        itemCount?: unknown
        secondaryLabel?: unknown
        statusBadge?: unknown
        children?: unknown
      }

      const id = typeof candidate.id === 'string' ? candidate.id.trim() : ''
      const name =
        typeof candidate.name === 'string' ? candidate.name.trim() : ''
      if (id === '' || name === '') {
        return []
      }

      const node: FolderNode = { id, name }
      if (
        typeof candidate.category === 'string' &&
        candidate.category.trim() !== ''
      ) {
        node.category = candidate.category.trim()
      }
      if (
        typeof candidate.itemCount === 'number' &&
        Number.isFinite(candidate.itemCount)
      ) {
        node.itemCount = candidate.itemCount
      }
      if (
        typeof candidate.secondaryLabel === 'string' &&
        candidate.secondaryLabel.trim() !== ''
      ) {
        node.secondaryLabel = candidate.secondaryLabel.trim()
      }
      if (
        typeof candidate.statusBadge === 'string' &&
        candidate.statusBadge.trim() !== ''
      ) {
        node.statusBadge = candidate.statusBadge.trim()
      }

      const children = Array.isArray(candidate.children)
        ? walk(candidate.children)
        : []
      if (children.length > 0) {
        node.children = children
      }

      return [node]
    })

  const sanitized = walk(value)
  return findFolderNodeByID(sanitized, 'all-items')
    ? sanitized
    : initialFolderTree
}

function parsePersistedWorkspaceSnapshot(
  value: string | null
): FolderNode[] | null {
  if (!value) {
    return null
  }

  try {
    const parsed = JSON.parse(value) as { folderTree?: unknown } | unknown
    if (Array.isArray(parsed)) {
      return sanitizeFolderNodes(parsed)
    }
    if (parsed && typeof parsed === 'object' && 'folderTree' in parsed) {
      return sanitizeFolderNodes(parsed.folderTree)
    }
  } catch {
    return null
  }

  return null
}

function loadPersistedWorkspaceSnapshot(
  profileID: string
): FolderNode[] | null {
  if (typeof window === 'undefined') {
    return null
  }

  const normalizedProfileID = profileID.trim()
  if (normalizedProfileID === '') {
    return null
  }

  return parsePersistedWorkspaceSnapshot(
    window.localStorage.getItem(
      inventoryWorkspaceSettingsStorageKey(normalizedProfileID)
    )
  )
}

async function loadProfileWorkspaceSnapshot(
  profileID: string
): Promise<FolderNode[] | null> {
  const normalizedProfileID = profileID.trim()
  if (normalizedProfileID === '') {
    return null
  }

  const response = await fetch(
    `/api/profiles/${encodeURIComponent(profileID)}/settings`
  )
  if (!response.ok) {
    return null
  }

  const payload = (await response.json()) as ProfileSettingsPayload
  return parsePersistedWorkspaceSnapshot(
    payload.settings?.[inventoryFolderTreeSettingsKey] ?? null
  )
}

function savePersistedWorkspaceSnapshot(
  profileID: string,
  folderTree: FolderNode[]
) {
  if (typeof window === 'undefined') {
    return
  }

  const normalizedProfileID = profileID.trim()
  if (normalizedProfileID === '') {
    return
  }

  window.localStorage.setItem(
    inventoryWorkspaceSettingsStorageKey(normalizedProfileID),
    JSON.stringify({ folderTree })
  )
}

async function saveProfileWorkspaceSnapshot(
  profileID: string,
  folderTree: FolderNode[]
): Promise<void> {
  const normalizedProfileID = profileID.trim()
  if (normalizedProfileID === '') {
    return
  }

  const response = await fetch(
    `/api/profiles/${encodeURIComponent(profileID)}/settings`
  )
  if (!response.ok) {
    return
  }

  const payload = (await response.json()) as ProfileSettingsPayload
  const nextTreeValue = JSON.stringify(folderTree)
  if (payload.settings?.[inventoryFolderTreeSettingsKey] === nextTreeValue) {
    return
  }

  await fetch(`/api/profiles/${encodeURIComponent(profileID)}/settings`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      settings: {
        ...(payload.settings ?? {}),
        [inventoryFolderTreeSettingsKey]: nextTreeValue,
      },
    }),
  })
}

function loadInventoryTreeState() {
  if (typeof window === 'undefined') {
    return {
      activeFolder: 'All Items',
      expandedNodeIDs: new Set<string>(),
    }
  }

  try {
    const raw = window.localStorage.getItem(inventoryTreeStorageKey)
    if (!raw) {
      return {
        activeFolder: 'All Items',
        expandedNodeIDs: new Set<string>(),
      }
    }

    const parsed = JSON.parse(raw) as {
      activeFolder?: string
      expandedNodeIDs?: string[]
    }

    return {
      activeFolder: parsed.activeFolder?.trim() || 'All Items',
      expandedNodeIDs: new Set(parsed.expandedNodeIDs ?? []),
    }
  } catch {
    return {
      activeFolder: 'All Items',
      expandedNodeIDs: new Set<string>(),
    }
  }
}

function loadInventoryItemFolderAssignments(): Record<string, string> {
  if (typeof window === 'undefined') {
    return {}
  }

  try {
    const raw = window.localStorage.getItem(
      inventoryItemFolderAssignmentsStorageKey
    )
    if (!raw) {
      return {}
    }

    const parsed = JSON.parse(raw) as Record<string, unknown>
    return Object.fromEntries(
      Object.entries(parsed)
        .map(([itemID, folderName]) => [
          itemID.trim(),
          typeof folderName === 'string' ? folderName.trim() : '',
        ])
        .filter(([itemID, folderName]) => itemID !== '' && folderName !== '')
    )
  } catch {
    return {}
  }
}

function readInventoryItemDragID(dataTransfer: DataTransfer): string {
  return (
    dataTransfer.getData(inventoryItemDragDataType) ||
    dataTransfer.getData('text/plain')
  ).trim()
}

function readFolderDragID(dataTransfer: DataTransfer): string {
  return dataTransfer.getData(folderDragDataType).trim()
}

function canMoveFolderToTarget(
  nodes: FolderNode[],
  draggedID: string,
  target: FolderDropTarget | null
): boolean {
  if (draggedID === '' || draggedID === 'all-items' || !target) {
    return false
  }
  if (target.kind === 'root') {
    return true
  }
  if (draggedID === target.nodeID) {
    return false
  }
  return !isInvalidFolderDropTarget(nodes, draggedID, target.nodeID)
}

function resolveFolderDragPreviewPosition(clientX: number, clientY: number) {
  const offset = 18
  const fallback = {
    x: clientX + offset,
    y: clientY + offset,
  }
  if (typeof window === 'undefined') {
    return fallback
  }

  const margin = 12
  const previewWidth = 256
  const previewHeight = 112
  return {
    x: Math.max(
      margin,
      Math.min(clientX + offset, window.innerWidth - previewWidth - margin)
    ),
    y: Math.max(
      margin,
      Math.min(clientY + offset, window.innerHeight - previewHeight - margin)
    ),
  }
}

export function Collection({
  title = 'Collection',
  description = '',
  routePath,
}: CollectionWorkspaceProps) {
  const [tableData, setTableData] = useState<Task[]>(tasks)
  const [inventoryItems, setInventoryItems] = useState<InventoryItem[]>([])
  const [folderTree, setFolderTree] = useState<FolderNode[]>(initialFolderTree)
  const [activeFolder, setActiveFolder] = useState(
    () => loadInventoryTreeState().activeFolder
  )
  const [expandedNodeIDs, setExpandedNodeIDs] = useState<Set<string>>(
    () => loadInventoryTreeState().expandedNodeIDs
  )
  const [itemFolderAssignments, setItemFolderAssignments] = useState<
    Record<string, string>
  >(() => loadInventoryItemFolderAssignments())
  const [folderCreateOpen, setFolderCreateOpen] = useState(false)
  const [folderCreateParentID, setFolderCreateParentID] = useState<
    string | null
  >(null)
  const [folderCreateName, setFolderCreateName] = useState('')
  const [collectionBrowserOpen, setCollectionBrowserOpen] = useState(false)
  const [draggedFolderID, setDraggedFolderID] = useState<string | null>(null)
  const [dragTarget, setDragTarget] = useState<FolderDropTarget | null>(null)
  const [dragPreviewPosition, setDragPreviewPosition] = useState<{
    x: number
    y: number
  } | null>(null)
  const folderPointerDragRef = useRef<{
    pointerID: number
    startX: number
    startY: number
    moved: boolean
    lastTarget: FolderDropTarget | null
  } | null>(null)
  const [folderPropertiesOpen, setFolderPropertiesOpen] = useState(false)
  const [folderPropertiesID, setFolderPropertiesID] = useState<string | null>(
    null
  )
  const [folderPropertiesName, setFolderPropertiesName] = useState('')
  const [folderPropertiesCategory, setFolderPropertiesCategory] =
    useState('General')
  const [folderPropertiesSecondaryLabel, setFolderPropertiesSecondaryLabel] =
    useState('')
  const [folderPropertiesStatusBadge, setFolderPropertiesStatusBadge] =
    useState('')
  const treeItemRefs = useRef<Record<string, HTMLButtonElement | null>>({})
  const photoCaptureInputRef = useRef<HTMLInputElement | null>(null)
  const photoUploadInputRef = useRef<HTMLInputElement | null>(null)
  const [loading, setLoading] = useState(false)
  const [loadError, setLoadError] = useState<string | null>(null)
  const [inventoryPhotos, setInventoryPhotos] = useState<InventoryPhoto[]>([])
  const [photosDialogOpen, setPhotosDialogOpen] = useState(false)
  const [barcodesDialogOpen, setBarcodesDialogOpen] = useState(false)
  const [photosLoading, setPhotosLoading] = useState(false)
  const [photosError, setPhotosError] = useState<string | null>(null)
  const [photosBusy, setPhotosBusy] = useState(false)
  const [photoRebuildError, setPhotoRebuildError] = useState<string | null>(
    null
  )
  const [photoRebuildSuccess, setPhotoRebuildSuccess] = useState<string | null>(
    null
  )
  const [cameraError, setCameraError] = useState<string | null>(null)
  const [cameraSuccess, setCameraSuccess] = useState<string | null>(null)
  const [activeProfileID, setActiveProfileID] = useState('')
  const [workspaceSnapshotReady, setWorkspaceSnapshotReady] = useState(false)
  const [selectedItemID, setSelectedItemID] = useState('')
  const [selectedItemLabel, setSelectedItemLabel] = useState('')
  const [itemEditorMode, setItemEditorMode] = useState<'create' | 'edit'>(
    'edit'
  )
  const [itemEditorSurface, setItemEditorSurface] = useState<
    'dialog' | 'panel'
  >('dialog')
  const [itemEditorOpen, setItemEditorOpen] = useState(false)
  const [itemDraft, setItemDraft] = useState<InventoryItemDraft>(
    emptyInventoryItemDraft
  )
  const [itemSaveBusy, setItemSaveBusy] = useState(false)
  const [itemSaveError, setItemSaveError] = useState<string | null>(null)
  const [itemSaveSuccess, setItemSaveSuccess] = useState<string | null>(null)
  const [assignCollectionOpen, setAssignCollectionOpen] = useState(false)
  const [assignCollectionItem, setAssignCollectionItem] =
    useState<InventoryItem | null>(null)
  const [assignCollectionName, setAssignCollectionName] = useState('')
  const [assignCollectionError, setAssignCollectionError] = useState<
    string | null
  >(null)
  const [tasksDialogOpen, setTasksDialogOpen] =
    useState<TasksDialogType | null>(null)
  const [tasksDialogRow, setTasksDialogRow] = useState<Task | null>(null)
  const [barcodeAddInput, setBarcodeAddInput] = useState('')
  const [barcodeLookupInput, setBarcodeLookupInput] = useState('')
  const [barcodeAddBusy, setBarcodeAddBusy] = useState(false)
  const [barcodeLookupBusy, setBarcodeLookupBusy] = useState(false)
  const [barcodeAddError, setBarcodeAddError] = useState<string | null>(null)
  const [barcodeAddSuccess, setBarcodeAddSuccess] = useState<string | null>(
    null
  )
  const [barcodeLookupError, setBarcodeLookupError] = useState<string | null>(
    null
  )
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
  const inventoryPhotoRequestIDRef = useRef(0)
  const selectedItemIDRef = useRef('')
  const {
    workspaceCollections,
    activeWorkspaceCollection,
    setActiveWorkspaceCollection,
    addCollection,
    assignWorkspaceItemToCollection,
  } = useWorkspaceCollections()
  const assignableWorkspaceCollections = useMemo(
    () =>
      workspaceCollections.filter((collection) => collection !== 'All Items'),
    [workspaceCollections]
  )
  const activeWorkspaceCollectionRef = useRef(activeWorkspaceCollection)

  const selectInventoryItem = useCallback((item: InventoryItem | null) => {
    setSelectedItemID(item?.id ?? '')
    setSelectedItemLabel(item ? item.part_number || item.id : '')
  }, [])

  useEffect(() => {
    selectedItemIDRef.current = selectedItemID
  }, [selectedItemID])

  const startCreateItem = useCallback(() => {
    setItemEditorMode('create')
    setItemEditorSurface('dialog')
    setItemDraft(emptyInventoryItemDraft())
    setItemSaveError(null)
    setItemSaveSuccess(null)
    selectInventoryItem(null)
    setItemEditorOpen(true)
  }, [selectInventoryItem])

  const openInventoryItemEditor = useCallback(
    (item: InventoryItem | null) => {
      if (!item) {
        return
      }
      setItemEditorMode('edit')
      setItemEditorSurface('panel')
      setItemDraft(inventoryItemToDraft(item))
      setItemSaveError(null)
      setItemSaveSuccess(null)
      selectInventoryItem(item)
      setItemEditorOpen(true)
    },
    [selectInventoryItem]
  )

  const openInventoryPhotosForItem = useCallback(
    (item: InventoryItem | null) => {
      setCameraError(null)
      setCameraSuccess(null)
      setPhotoRebuildError(null)
      setPhotoRebuildSuccess(null)
      setPhotosError(null)
      if (!item) {
        selectInventoryItem(null)
        setPhotosError('Select an inventory item before managing photos.')
        setPhotosDialogOpen(true)
        return
      }
      selectInventoryItem(item)
      setPhotosDialogOpen(true)
    },
    [selectInventoryItem]
  )

  const openInventoryBarcodesForItem = useCallback(
    (item: InventoryItem | null) => {
      setBarcodeAddError(null)
      setBarcodeAddSuccess(null)
      setBarcodeLookupError(null)
      setBarcodeLookupCompleted(false)
      setBarcodeMatches([])
      if (!item) {
        selectInventoryItem(null)
        setBarcodeAddError('Select an inventory item before managing barcodes.')
        setBarcodesDialogOpen(true)
        return
      }
      selectInventoryItem(item)
      setBarcodesDialogOpen(true)
    },
    [selectInventoryItem]
  )

  const openInventoryAssignCollectionForItem = useCallback(
    (item: InventoryItem | null) => {
      setAssignCollectionError(null)
      if (!item) {
        selectInventoryItem(null)
        setAssignCollectionItem(null)
        setAssignCollectionName(assignableWorkspaceCollections[0] ?? '')
        setAssignCollectionError(
          'Select an inventory item before assigning a collection.'
        )
        setAssignCollectionOpen(true)
        return
      }

      const currentAssignment = itemFolderAssignments[item.id]
      const defaultCollection =
        currentAssignment && currentAssignment !== 'All Items'
          ? currentAssignment
          : (assignableWorkspaceCollections[0] ?? '')
      selectInventoryItem(item)
      setAssignCollectionItem(item)
      setAssignCollectionName(defaultCollection)
      setAssignCollectionOpen(true)
    },
    [assignableWorkspaceCollections, itemFolderAssignments, selectInventoryItem]
  )

  const handleAssignInventoryItemToCollection = useCallback(async () => {
    if (!assignCollectionItem || !assignCollectionName) {
      setAssignCollectionError('Choose an item and collection before saving.')
      return
    }

    const workspaceItem: WorkspaceCollectionItem = {
      id: assignCollectionItem.id,
      name: assignCollectionItem.title || assignCollectionItem.part_number,
      detail:
        [assignCollectionItem.brand, assignCollectionItem.category]
          .filter(Boolean)
          .join(' - ') || 'Inventory item',
      collectionName: assignCollectionName,
    }
    const updated = await assignWorkspaceItemToCollection(
      workspaceItem,
      assignCollectionName
    )
    if (!updated) {
      setAssignCollectionError('Collection assignment could not be saved.')
      return
    }

    setItemFolderAssignments((current) => ({
      ...current,
      [assignCollectionItem.id]: assignCollectionName,
    }))
    selectInventoryItem(assignCollectionItem)
    setAssignCollectionOpen(false)
    setAssignCollectionError(null)
  }, [
    assignCollectionItem,
    assignCollectionName,
    assignWorkspaceItemToCollection,
    selectInventoryItem,
  ])

  const selectInventoryFolder = useCallback(
    (nextFolder: string) => {
      const safeFolder = workspaceCollections.includes(nextFolder)
        ? nextFolder
        : 'All Items'
      setActiveFolder(safeFolder)
      setCollectionBrowserOpen(false)
      void setActiveWorkspaceCollection(safeFolder)
    },
    [setActiveWorkspaceCollection, workspaceCollections]
  )

  const resolveInventoryItemFromTask = useCallback(
    (task: Task) =>
      inventoryItems.find((item) => item.id === task.itemID) ??
      inventoryItems.find((item) => item.part_number === task.id) ??
      inventoryItems.find((item) => item.part_number === task.partNumber) ??
      inventoryItems.find((item) => item.title === task.title) ??
      null,
    [inventoryItems]
  )

  const openFolderProperties = useCallback((node: FolderNode) => {
    setFolderPropertiesID(node.id)
    setFolderPropertiesName(node.name)
    setFolderPropertiesCategory(node.category?.trim() || 'General')
    setFolderPropertiesSecondaryLabel(node.secondaryLabel?.trim() || '')
    setFolderPropertiesStatusBadge(node.statusBadge?.trim() || '')
    setFolderPropertiesOpen(true)
  }, [])

  const loadInventoryItems = useCallback(
    async (preferredSelectedItemID?: string) => {
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
              (preferredSelectedItemID?.trim() ||
                selectedItemIDRef.current.trim())
          ) ??
          items[0] ??
          null
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
    },
    [routePath, selectInventoryItem]
  )

  useEffect(() => {
    void loadInventoryItems()
  }, [loadInventoryItems])

  useEffect(() => {
    if (typeof window === 'undefined') {
      return
    }

    window.localStorage.setItem(
      inventoryTreeStorageKey,
      JSON.stringify({
        activeFolder,
        expandedNodeIDs: Array.from(expandedNodeIDs),
      })
    )
  }, [activeFolder, expandedNodeIDs])

  useEffect(() => {
    if (typeof window === 'undefined') {
      return
    }

    window.localStorage.setItem(
      inventoryItemFolderAssignmentsStorageKey,
      JSON.stringify(itemFolderAssignments)
    )
  }, [itemFolderAssignments])

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

  const moveDraggedFolder = useCallback(
    (draggedID: string, target: FolderDropTarget | null) => {
      const draggedNode = findFolderNodeByID(folderTree, draggedID)
      if (!draggedNode || draggedID === 'all-items' || !target) {
        return
      }

      if (target.kind === 'child') {
        if (draggedID === target.nodeID) {
          return
        }
        setFolderTree((previous) =>
          moveFolderNode(previous, draggedID, target.nodeID)
        )
        setExpandedNodeIDs((previous) => {
          const next = new Set(previous)
          next.add(target.nodeID)
          return next
        })
      } else if (target.kind === 'before' || target.kind === 'after') {
        if (draggedID === target.nodeID) {
          return
        }
        setFolderTree((previous) =>
          moveFolderNodeRelative(
            previous,
            draggedID,
            target.nodeID,
            target.kind
          )
        )
      } else {
        setFolderTree((previous) => moveFolderNodeToRoot(previous, draggedID))
      }

      setActiveFolder(draggedNode.name)
    },
    [folderTree]
  )

  const draggedFolderNode = useMemo(
    () =>
      draggedFolderID ? findFolderNodeByID(folderTree, draggedFolderID) : null,
    [draggedFolderID, folderTree]
  )

  const folderDropHint = useMemo(() => {
    if (!dragTarget || !draggedFolderNode) {
      return ''
    }
    if (dragTarget.kind === 'root') {
      return `Move ${draggedFolderNode.name} to the root level`
    }
    const targetNode = findFolderNodeByID(folderTree, dragTarget.nodeID)
    const targetName = targetNode?.name ?? 'this folder'
    if (dragTarget.kind === 'child') {
      return `Move into ${targetName}`
    }
    if (dragTarget.kind === 'before') {
      return `Place before ${targetName}`
    }
    return `Place after ${targetName}`
  }, [dragTarget, draggedFolderNode, folderTree])

  const resolvePointerFolderDropTarget = useCallback(
    (
      clientX: number,
      clientY: number,
      draggedID: string
    ): FolderDropTarget | null => {
      if (typeof document === 'undefined') {
        return null
      }

      const hit = document.elementFromPoint(clientX, clientY)
      if (!(hit instanceof HTMLElement)) {
        return null
      }

      if (hit.closest('[data-testid="folder-tree-root-drop-zone"]')) {
        return { kind: 'root' }
      }

      const explicitDropZone = hit.closest<HTMLElement>(
        '[data-folder-drop-zone-kind][data-folder-row-id]'
      )
      const explicitDropKind =
        explicitDropZone?.dataset.folderDropZoneKind === 'after'
          ? 'after'
          : explicitDropZone?.dataset.folderDropZoneKind === 'before'
            ? 'before'
            : null
      const explicitNodeID = explicitDropZone?.dataset.folderRowId?.trim() ?? ''
      if (
        explicitDropKind &&
        explicitNodeID &&
        explicitNodeID !== draggedID &&
        !isInvalidFolderDropTarget(folderTree, draggedID, explicitNodeID)
      ) {
        return { kind: explicitDropKind, nodeID: explicitNodeID }
      }

      const rowShell = hit.closest<HTMLElement>('[data-folder-row-id]')
      const nodeID = rowShell?.dataset.folderRowId?.trim() ?? ''
      if (!rowShell || nodeID === '' || nodeID === draggedID) {
        return null
      }

      if (nodeID === 'all-items') {
        return { kind: 'root' }
      }

      if (isInvalidFolderDropTarget(folderTree, draggedID, nodeID)) {
        return null
      }

      const rect = rowShell.getBoundingClientRect()
      const edgeThreshold = Math.max(10, Math.min(18, rect.height * 0.25))
      if (clientY <= rect.top + edgeThreshold) {
        return { kind: 'before', nodeID }
      }
      if (clientY >= rect.bottom - edgeThreshold) {
        return { kind: 'after', nodeID }
      }
      return { kind: 'child', nodeID }
    },
    [folderTree]
  )

  const startFolderPointerDrag = useCallback(
    (nodeID: string, event: ReactPointerEvent<HTMLButtonElement>) => {
      if (nodeID === 'all-items' || event.button !== 0) {
        return
      }

      const target = event.target as HTMLElement | null
      if (target?.closest('[data-drag-disabled="true"]')) {
        return
      }

      event.preventDefault()
      event.stopPropagation()
      folderPointerDragRef.current = {
        pointerID: event.pointerId,
        startX: event.clientX,
        startY: event.clientY,
        moved: false,
        lastTarget: null,
      }
      setDragPreviewPosition(
        resolveFolderDragPreviewPosition(event.clientX, event.clientY)
      )
      setDragTarget(null)
      setDraggedFolderID(nodeID)
    },
    []
  )

  useEffect(() => {
    const pointerDrag = folderPointerDragRef.current
    if (!draggedFolderID || !pointerDrag) {
      return
    }

    const previousUserSelect = document.body.style.userSelect
    const previousCursor = document.body.style.cursor
    document.body.style.userSelect = 'none'
    document.body.style.cursor = 'grabbing'

    const updateDragTarget = (clientX: number, clientY: number) => {
      setDragPreviewPosition(resolveFolderDragPreviewPosition(clientX, clientY))
      const nextTarget = resolvePointerFolderDropTarget(
        clientX,
        clientY,
        draggedFolderID
      )
      pointerDrag.lastTarget = nextTarget
      setDragTarget((previous) =>
        folderDropTargetsEqual(previous, nextTarget) ? previous : nextTarget
      )
    }

    const handlePointerMove = (event: PointerEvent) => {
      if (event.pointerId !== pointerDrag.pointerID) {
        return
      }

      const deltaX = event.clientX - pointerDrag.startX
      const deltaY = event.clientY - pointerDrag.startY
      if (!pointerDrag.moved && Math.hypot(deltaX, deltaY) < 6) {
        return
      }

      pointerDrag.moved = true
      updateDragTarget(event.clientX, event.clientY)
    }

    const clearFolderPointerDrag = () => {
      folderPointerDragRef.current = null
      setDragPreviewPosition(null)
      setDraggedFolderID(null)
      setDragTarget(null)
    }

    const handlePointerUp = (event: PointerEvent) => {
      if (event.pointerId !== pointerDrag.pointerID) {
        return
      }

      if (!pointerDrag.moved) {
        clearFolderPointerDrag()
        return
      }

      const target =
        pointerDrag.lastTarget ??
        resolvePointerFolderDropTarget(
          event.clientX,
          event.clientY,
          draggedFolderID
        )
      if (target) {
        moveDraggedFolder(draggedFolderID, target)
      }
      clearFolderPointerDrag()
    }

    const handlePointerCancel = (event: PointerEvent) => {
      if (event.pointerId !== pointerDrag.pointerID) {
        return
      }
      clearFolderPointerDrag()
    }

    const handleKeyDown = (event: globalThis.KeyboardEvent) => {
      if (event.key === 'Escape') {
        clearFolderPointerDrag()
      }
    }

    window.addEventListener('pointermove', handlePointerMove)
    window.addEventListener('pointerup', handlePointerUp)
    window.addEventListener('pointercancel', handlePointerCancel)
    window.addEventListener('keydown', handleKeyDown)

    return () => {
      document.body.style.userSelect = previousUserSelect
      document.body.style.cursor = previousCursor
      window.removeEventListener('pointermove', handlePointerMove)
      window.removeEventListener('pointerup', handlePointerUp)
      window.removeEventListener('pointercancel', handlePointerCancel)
      window.removeEventListener('keydown', handleKeyDown)
    }
  }, [draggedFolderID, moveDraggedFolder, resolvePointerFolderDropTarget])

  const handleInventoryItemFolderDragOver = useCallback(
    (node: FolderNode, event: ReactDragEvent<HTMLElement>) => {
      if (readFolderDragID(event.dataTransfer)) {
        return
      }
      if (node.id === 'all-items') {
        return
      }

      const itemID = readInventoryItemDragID(event.dataTransfer)
      if (!itemID) {
        return
      }

      event.preventDefault()
      event.dataTransfer.dropEffect = 'move'
    },
    []
  )

  const handleInventoryItemFolderDrop = useCallback(
    (node: FolderNode, event: ReactDragEvent<HTMLElement>) => {
      if (readFolderDragID(event.dataTransfer)) {
        return
      }
      if (node.id === 'all-items') {
        return
      }

      const itemID = readInventoryItemDragID(event.dataTransfer)
      if (!itemID) {
        return
      }

      event.preventDefault()
      event.stopPropagation()
      setItemFolderAssignments((previous) => ({
        ...previous,
        [itemID]: node.name,
      }))
      setActiveFolder(node.name)
    },
    []
  )

  const clearFolderHTMLDrag = useCallback(() => {
    setDragPreviewPosition(null)
    setDraggedFolderID(null)
    setDragTarget(null)
  }, [])

  const startFolderHTMLDrag = useCallback(
    (node: FolderNode, event: ReactDragEvent<HTMLButtonElement>) => {
      if (node.id === 'all-items') {
        event.preventDefault()
        return
      }

      event.stopPropagation()
      event.dataTransfer.effectAllowed = 'move'
      event.dataTransfer.setData(folderDragDataType, node.id)
      setDraggedFolderID(node.id)
      setDragTarget(null)
      setDragPreviewPosition(
        resolveFolderDragPreviewPosition(event.clientX, event.clientY)
      )
    },
    []
  )

  const handleFolderHTMLDragOver = useCallback(
    (target: FolderDropTarget, event: ReactDragEvent<HTMLElement>) => {
      const folderID = readFolderDragID(event.dataTransfer)
      if (!folderID) {
        return false
      }

      event.preventDefault()
      event.stopPropagation()
      setDraggedFolderID(folderID)
      setDragPreviewPosition(
        resolveFolderDragPreviewPosition(event.clientX, event.clientY)
      )

      if (!canMoveFolderToTarget(folderTree, folderID, target)) {
        event.dataTransfer.dropEffect = 'none'
        setDragTarget(null)
        return true
      }

      event.dataTransfer.dropEffect = 'move'
      setDragTarget((previous) =>
        folderDropTargetsEqual(previous, target) ? previous : target
      )
      return true
    },
    [folderTree]
  )

  const handleFolderHTMLDrop = useCallback(
    (target: FolderDropTarget, event: ReactDragEvent<HTMLElement>) => {
      const folderID = readFolderDragID(event.dataTransfer)
      if (!folderID) {
        return false
      }

      event.preventDefault()
      event.stopPropagation()
      if (canMoveFolderToTarget(folderTree, folderID, target)) {
        moveDraggedFolder(folderID, target)
      }
      clearFolderHTMLDrag()
      return true
    },
    [clearFolderHTMLDrag, folderTree, moveDraggedFolder]
  )

  const renderFolderTree = useCallback(
    (nodes: FolderNode[], level = 1) =>
      nodes.map((node) => {
        const hasChildren = Boolean(node.children?.length)
        const expanded = hasChildren && expandedNodeIDs.has(node.id)
        const isActive = activeFolder === node.name
        const isChildDropTarget =
          dragTarget?.kind === 'child' && dragTarget.nodeID === node.id
        const isBeforeDropTarget =
          dragTarget?.kind === 'before' && dragTarget.nodeID === node.id
        const isAfterDropTarget =
          dragTarget?.kind === 'after' && dragTarget.nodeID === node.id
        const hasInvalidFolderDropTarget = draggedFolderID
          ? isInvalidFolderDropTarget(folderTree, draggedFolderID, node.id)
          : false
        return (
          <div key={node.id} role='none' className='relative'>
            {draggedFolderID ? (
              <div
                role='presentation'
                data-testid={`folder-tree-drop-before-${node.id}`}
                data-folder-drop-zone-kind='before'
                data-folder-row-id={node.id}
                data-invalid-drop-target={
                  hasInvalidFolderDropTarget ? 'true' : 'false'
                }
                onDragEnter={(event) =>
                  handleFolderHTMLDragOver(
                    { kind: 'before', nodeID: node.id },
                    event
                  )
                }
                onDragOver={(event) =>
                  handleFolderHTMLDragOver(
                    { kind: 'before', nodeID: node.id },
                    event
                  )
                }
                onDrop={(event) =>
                  handleFolderHTMLDrop(
                    { kind: 'before', nodeID: node.id },
                    event
                  )
                }
                className={cn(
                  'mx-2 mb-1 h-2 rounded-full border border-dashed transition-colors',
                  hasInvalidFolderDropTarget
                    ? 'border-destructive/60 bg-destructive/10'
                    : isBeforeDropTarget
                      ? 'border-primary bg-primary/25'
                      : 'border-border/40 bg-muted/20'
                )}
              />
            ) : null}
            <div
              data-testid={`folder-tree-row-shell-${node.id}`}
              data-folder-row-id={node.id}
              className='group relative flex items-center gap-1'
              style={{ paddingInlineStart: `${(level - 1) * 1}rem` }}
            >
              {hasChildren ? (
                <Button
                  type='button'
                  variant='ghost'
                  size='icon'
                  className='h-6 w-6 shrink-0 rounded-sm text-muted-foreground/70 transition-colors hover:bg-transparent hover:text-foreground'
                  data-testid={`folder-tree-toggle-${node.id}`}
                  data-drag-disabled='true'
                  tabIndex={-1}
                  aria-label={`Toggle ${node.name}`}
                  aria-expanded={expanded ? 'true' : 'false'}
                  data-state={expanded ? 'expanded' : 'collapsed'}
                  onClick={() => toggleNodeExpanded(node.id)}
                >
                  <ChevronRight
                    className={cn(
                      'size-4 transition-transform duration-200',
                      expanded && 'rotate-90'
                    )}
                  />
                </Button>
              ) : (
                <span
                  aria-hidden='true'
                  data-testid={`folder-tree-connector-${node.id}`}
                  className='inline-flex h-6 w-6 shrink-0 items-center justify-center text-muted-foreground/30'
                >
                  <span className='h-6 border-l border-border/70' />
                </span>
              )}
              <button
                ref={(element) => {
                  treeItemRefs.current[node.id] = element
                }}
                type='button'
                role='treeitem'
                tabIndex={isActive ? 0 : -1}
                aria-level={level}
                aria-selected={isActive ? 'true' : 'false'}
                aria-expanded={
                  hasChildren ? (expanded ? 'true' : 'false') : undefined
                }
                data-testid={`folder-tree-item-${node.id}`}
                data-active={isActive ? 'true' : 'false'}
                data-node-kind={hasChildren ? 'branch' : 'leaf'}
                data-node-expanded={
                  hasChildren ? (expanded ? 'true' : 'false') : undefined
                }
                data-draggable-row={node.id !== 'all-items' ? 'true' : 'false'}
                data-invalid-drop-target={
                  hasInvalidFolderDropTarget ? 'true' : 'false'
                }
                className={cn(
                  'relative flex min-w-0 flex-1 items-start rounded-md bg-transparent px-3 py-2 text-left text-sm focus-visible:ring-2 focus-visible:ring-ring/70 focus-visible:ring-offset-1 focus-visible:outline-none',
                  isChildDropTarget && 'bg-primary/20'
                )}
                onClick={() => selectInventoryFolder(node.name)}
                onKeyDown={(event) => handleTreeItemKeyDown(node, event)}
                onDragEnter={(event) => {
                  if (
                    handleFolderHTMLDragOver(
                      { kind: 'child', nodeID: node.id },
                      event
                    )
                  ) {
                    return
                  }
                  handleInventoryItemFolderDragOver(node, event)
                }}
                onDragOver={(event) => {
                  if (
                    handleFolderHTMLDragOver(
                      { kind: 'child', nodeID: node.id },
                      event
                    )
                  ) {
                    return
                  }
                  handleInventoryItemFolderDragOver(node, event)
                }}
                onDrop={(event) => {
                  if (
                    handleFolderHTMLDrop(
                      { kind: 'child', nodeID: node.id },
                      event
                    )
                  ) {
                    return
                  }
                  handleInventoryItemFolderDrop(node, event)
                }}
              >
                <span
                  className={cn(
                    'pointer-events-none absolute inset-0 rounded-md border border-transparent transition-colors',
                    isActive
                      ? 'border-white/15 bg-white/10 shadow-sm'
                      : 'group-hover:border-border/60 group-hover:bg-accent/70',
                    hasInvalidFolderDropTarget &&
                      'border-destructive/60 bg-destructive/10',
                    isChildDropTarget && 'border-primary bg-primary/20'
                  )}
                  aria-hidden='true'
                />
                <span
                  className={cn(
                    'pointer-events-none absolute inset-y-1 left-0 w-1 rounded-full transition-opacity',
                    isActive ? 'bg-white/85 opacity-100' : 'opacity-0'
                  )}
                  aria-hidden='true'
                />
                <span
                  className={cn(
                    'relative flex min-w-0 flex-1 items-start',
                    isActive
                      ? 'font-semibold text-accent-foreground'
                      : 'text-foreground/90'
                  )}
                >
                  <span className='min-w-0 flex-1'>
                    <span
                      data-testid={`collection-folder-${node.id}`}
                      className='block truncate'
                    >
                      {node.name}
                    </span>
                    {node.secondaryLabel ? (
                      <span
                        data-testid={`folder-tree-secondary-${node.id}`}
                        className='block truncate text-xs font-normal text-muted-foreground'
                      >
                        {node.secondaryLabel}
                      </span>
                    ) : null}
                  </span>
                </span>
              </button>
              <div
                data-testid={`folder-tree-trailing-${node.id}`}
                className={cn(
                  'relative flex shrink-0 items-center gap-2 pe-1',
                  hasInvalidFolderDropTarget && 'text-foreground/70',
                  isChildDropTarget && 'text-primary'
                )}
              >
                <span className='flex shrink-0 items-center gap-2'>
                  {typeof node.itemCount === 'number' ? (
                    <span
                      data-testid={`folder-tree-count-${node.id}`}
                      className='text-xs font-medium text-muted-foreground tabular-nums'
                    >
                      {node.itemCount}
                    </span>
                  ) : null}
                  {node.statusBadge ? (
                    <Badge
                      data-testid={`folder-tree-badge-${node.id}`}
                      variant='secondary'
                      className='rounded-full px-2 py-0.5 text-[11px] font-semibold'
                    >
                      {node.statusBadge}
                    </Badge>
                  ) : null}
                </span>
                <div
                  data-testid={`folder-tree-inline-actions-${node.id}`}
                  data-drag-disabled='true'
                  className={cn(
                    'flex w-14 shrink-0 items-center justify-end gap-1 transition-opacity',
                    'pointer-events-auto opacity-100'
                  )}
                >
                  <Button
                    type='button'
                    variant='ghost'
                    size='icon'
                    tabIndex={-1}
                    className='h-6 w-6 rounded-sm text-muted-foreground/70 hover:bg-transparent hover:text-foreground'
                    data-testid={`folder-tree-add-child-${node.id}`}
                    data-drag-disabled='true'
                    aria-label={`Add child folder under ${node.name}`}
                    onClick={(event) => {
                      event.stopPropagation()
                      setFolderCreateParentID(node.id)
                      setFolderCreateName('')
                      setCollectionBrowserOpen(false)
                      setFolderCreateOpen(true)
                    }}
                  >
                    <Plus className='size-4' />
                  </Button>
                  <DropdownMenu modal={false}>
                    <DropdownMenuTrigger asChild>
                      <Button
                        type='button'
                        variant='ghost'
                        size='icon'
                        tabIndex={-1}
                        className='h-6 w-6 rounded-sm text-muted-foreground/70 hover:bg-transparent hover:text-foreground'
                        data-testid={`folder-tree-row-actions-${node.id}`}
                        data-drag-disabled='true'
                        aria-label={`Open folder actions for ${node.name}`}
                        onClick={(event) => event.stopPropagation()}
                      >
                        <Ellipsis className='size-4' />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent
                      align='end'
                      onClick={(event) => event.stopPropagation()}
                    >
                      <DropdownMenuItem
                        data-testid={`folder-tree-row-action-select-${node.id}`}
                        onClick={(event) => {
                          event.stopPropagation()
                          selectInventoryFolder(node.name)
                        }}
                      >
                        Select folder
                      </DropdownMenuItem>
                      <DropdownMenuItem
                        data-testid={`folder-tree-row-action-properties-${node.id}`}
                        onClick={(event) => {
                          event.stopPropagation()
                          setCollectionBrowserOpen(false)
                          openFolderProperties(node)
                        }}
                      >
                        Properties
                      </DropdownMenuItem>
                      <DropdownMenuItem
                        data-testid={`folder-tree-row-action-add-child-${node.id}`}
                        onClick={(event) => {
                          event.stopPropagation()
                          setFolderCreateParentID(node.id)
                          setFolderCreateName('')
                          setCollectionBrowserOpen(false)
                          setFolderCreateOpen(true)
                        }}
                      >
                        Add child folder
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </div>
                <button
                  type='button'
                  tabIndex={-1}
                  draggable={node.id !== 'all-items'}
                  data-testid={`folder-tree-drag-handle-${node.id}`}
                  aria-label={`Drag ${node.name}`}
                  title={`Drag ${node.name}`}
                  className={cn(
                    'inline-flex h-8 w-8 shrink-0 cursor-grab items-center justify-center rounded-sm text-muted-foreground/70 transition-colors group-hover:text-foreground active:cursor-grabbing',
                    draggedFolderID === node.id && 'text-foreground'
                  )}
                  onPointerDown={(event) =>
                    startFolderPointerDrag(node.id, event)
                  }
                  onDragStart={(event) => startFolderHTMLDrag(node, event)}
                  onDragEnd={clearFolderHTMLDrag}
                  onMouseDown={(event) => event.stopPropagation()}
                  onClick={(event) => event.preventDefault()}
                >
                  <GripVertical className='size-4' />
                </button>
              </div>
            </div>
            {draggedFolderID ? (
              <div
                role='presentation'
                data-testid={`folder-tree-drop-after-${node.id}`}
                data-folder-drop-zone-kind='after'
                data-folder-row-id={node.id}
                data-invalid-drop-target={
                  hasInvalidFolderDropTarget ? 'true' : 'false'
                }
                onDragEnter={(event) =>
                  handleFolderHTMLDragOver(
                    { kind: 'after', nodeID: node.id },
                    event
                  )
                }
                onDragOver={(event) =>
                  handleFolderHTMLDragOver(
                    { kind: 'after', nodeID: node.id },
                    event
                  )
                }
                onDrop={(event) =>
                  handleFolderHTMLDrop(
                    { kind: 'after', nodeID: node.id },
                    event
                  )
                }
                className={cn(
                  'mx-2 mt-1 h-2 rounded-full border border-dashed transition-colors',
                  hasInvalidFolderDropTarget
                    ? 'border-destructive/60 bg-destructive/10'
                    : isAfterDropTarget
                      ? 'border-primary bg-primary/25'
                      : 'border-border/40 bg-muted/20'
                )}
              />
            ) : null}
            {hasChildren && expanded ? (
              <div
                role='group'
                data-testid={`folder-tree-group-${node.id}`}
                className='ml-4 border-l pl-1'
              >
                {renderFolderTree(node.children ?? [], level + 1)}
              </div>
            ) : null}
          </div>
        )
      }),
    [
      activeFolder,
      clearFolderHTMLDrag,
      dragTarget,
      draggedFolderID,
      expandedNodeIDs,
      handleFolderHTMLDragOver,
      handleFolderHTMLDrop,
      handleInventoryItemFolderDragOver,
      handleInventoryItemFolderDrop,
      handleTreeItemKeyDown,
      folderTree,
      openFolderProperties,
      selectInventoryFolder,
      startFolderHTMLDrag,
      startFolderPointerDrag,
      toggleNodeExpanded,
    ]
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
  const folderFilterOptions = useMemo(
    () => buildWorkspaceCollectionFilterOptions(workspaceCollections),
    [workspaceCollections]
  )
  const activeFolderIsAvailable = folderFilterOptions.some(
    (folder) => folder.name === activeFolder
  )
  const activeFolderFilterValue = activeFolderIsAvailable
    ? activeFolder
    : 'All Items'
  const isInventoryRoute = routePath === '/_authenticated/inventory/'
  const selectedInventoryItem = useMemo(
    () => inventoryItems.find((item) => item.id === selectedItemID) ?? null,
    [inventoryItems, selectedItemID]
  )
  const visibleTableData = useMemo(() => {
    if (!isInventoryRoute || activeFolderFilterValue === 'All Items') {
      return tableData
    }

    return tableData.filter((row) => {
      const itemID = row.itemID ?? row.id
      return itemFolderAssignments[itemID] === activeFolderFilterValue
    })
  }, [
    activeFolderFilterValue,
    isInventoryRoute,
    itemFolderAssignments,
    tableData,
  ])
  const selectedFolderHasNoItems =
    isInventoryRoute &&
    activeFolderFilterValue !== 'All Items' &&
    visibleTableData.length === 0
  const selectedItemContext = selectedItemLabel || selectedItemID || 'None'
  const selectedVisibleInventoryIndex = selectedItemID
    ? visibleTableData.findIndex(
        (row) => (row.itemID ?? row.id) === selectedItemID
      )
    : -1
  const canNavigateToPreviousInventoryItem = selectedVisibleInventoryIndex > 0
  const canNavigateToNextInventoryItem =
    selectedVisibleInventoryIndex >= 0 &&
    selectedVisibleInventoryIndex < visibleTableData.length - 1
  const openAdjacentInventoryItem = useCallback(
    (offset: number) => {
      if (selectedVisibleInventoryIndex < 0) {
        return
      }
      const nextRow = visibleTableData[selectedVisibleInventoryIndex + offset]
      if (!nextRow) {
        return
      }
      const nextID = nextRow.itemID ?? nextRow.id
      const nextItem =
        inventoryItems.find((item) => item.id === nextID) ??
        inventoryItems.find((item) => item.part_number === nextRow.id) ??
        null
      openInventoryItemEditor(nextItem)
    },
    [
      inventoryItems,
      openInventoryItemEditor,
      selectedVisibleInventoryIndex,
      visibleTableData,
    ]
  )
  const selectedPhoto =
    selectedPhotoIndex === null ? null : inventoryPhotos[selectedPhotoIndex]
  useEffect(() => {
    const previous = activeWorkspaceCollectionRef.current
    activeWorkspaceCollectionRef.current = activeWorkspaceCollection
    if (
      activeWorkspaceCollection &&
      (activeWorkspaceCollection !== 'All Items' ||
        previous !== activeWorkspaceCollection)
    ) {
      setActiveFolder(activeWorkspaceCollection)
    }
  }, [activeWorkspaceCollection])

  const loadInventoryPhotos = useCallback(async () => {
    if (!isInventoryRoute) {
      inventoryPhotoRequestIDRef.current += 1
      setInventoryPhotos([])
      setPhotosError(null)
      setPhotosLoading(false)
      return
    }
    if (!selectedItemID) {
      inventoryPhotoRequestIDRef.current += 1
      setInventoryPhotos([])
      setPhotosError(null)
      setPhotosLoading(false)
      return
    }
    const requestID = inventoryPhotoRequestIDRef.current + 1
    inventoryPhotoRequestIDRef.current = requestID
    setPhotosLoading(true)
    setPhotosError(null)
    setInventoryPhotos([])
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
      if (inventoryPhotoRequestIDRef.current !== requestID) {
        return
      }
      setInventoryPhotos(mapped.filter((photo) => photo.id !== ''))
    } catch {
      if (inventoryPhotoRequestIDRef.current !== requestID) {
        return
      }
      setInventoryPhotos([])
      setPhotosError(
        'Photos could not be loaded for this item. Retry to continue.'
      )
    } finally {
      if (inventoryPhotoRequestIDRef.current === requestID) {
        setPhotosLoading(false)
      }
    }
  }, [isInventoryRoute, selectedItemID])

  useEffect(() => {
    void loadInventoryPhotos()
  }, [loadInventoryPhotos])

  useEffect(() => {
    setSelectedPhotoIndex(null)
  }, [selectedItemID])

  useEffect(() => {
    setPhotoRebuildError(null)
    setPhotoRebuildSuccess(null)
  }, [selectedItemID])

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
      setWorkspaceSnapshotReady(false)
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
    if (!activeProfileID) {
      setWorkspaceSnapshotReady(false)
      return
    }

    let cancelled = false
    setWorkspaceSnapshotReady(false)

    const hydrateWorkspaceSnapshot = async () => {
      try {
        const remoteTree = await loadProfileWorkspaceSnapshot(activeProfileID)
        const localTree = loadPersistedWorkspaceSnapshot(activeProfileID)
        const nextTree = remoteTree ?? localTree ?? initialFolderTree

        if (cancelled) {
          return
        }

        setFolderTree(nextTree)
        setActiveFolder((previous) =>
          folderTreeContainsName(nextTree, previous) ? previous : 'All Items'
        )
        savePersistedWorkspaceSnapshot(activeProfileID, nextTree)
      } catch {
        if (cancelled) {
          return
        }

        const localTree = loadPersistedWorkspaceSnapshot(activeProfileID)
        const nextTree = localTree ?? initialFolderTree
        setFolderTree(nextTree)
        setActiveFolder((previous) =>
          folderTreeContainsName(nextTree, previous) ? previous : 'All Items'
        )
      } finally {
        if (!cancelled) {
          setWorkspaceSnapshotReady(true)
        }
      }
    }

    void hydrateWorkspaceSnapshot()
    return () => {
      cancelled = true
    }
  }, [activeProfileID])

  useEffect(() => {
    if (!activeProfileID || !workspaceSnapshotReady) {
      return
    }
    savePersistedWorkspaceSnapshot(activeProfileID, folderTree)
    void saveProfileWorkspaceSnapshot(activeProfileID, folderTree)
  }, [activeProfileID, folderTree, workspaceSnapshotReady])

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
    const wasCreateMode = itemEditorMode === 'create'
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
        wasCreateMode
          ? 'Item created and selected for follow-up media attach.'
          : 'Item changes saved and reloaded from the API.'
      )
      await loadInventoryItems(savedID)
      setItemEditorOpen(false)
    } catch {
      setItemSaveError('Inventory save failed. Review the fields and retry.')
    } finally {
      setItemSaveBusy(false)
    }
  }, [itemDraft, itemEditorMode, loadInventoryItems, selectedItemID])

  const handlePhotoUpload = useCallback(
    async (file: File | null) => {
      if (!file || !selectedItemID) {
        return false
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
        return true
      } catch {
        setPhotosError(
          'Photo upload failed. Retry with a supported image file.'
        )
        return false
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
        setPhotosError(
          'Unable to set primary photo right now. Retry this action.'
        )
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

  const handleReorderPhotos = useCallback(
    async (photoID: string, direction: 'up' | 'down') => {
      if (!selectedItemID || !photoID) {
        return
      }
      const currentIndex = inventoryPhotos.findIndex(
        (photo) => photo.id === photoID
      )
      if (currentIndex < 0) {
        return
      }
      const targetIndex =
        direction === 'up' ? currentIndex - 1 : currentIndex + 1
      if (targetIndex < 0 || targetIndex >= inventoryPhotos.length) {
        return
      }

      const reordered = [...inventoryPhotos]
      const [moved] = reordered.splice(currentIndex, 1)
      reordered.splice(targetIndex, 0, moved)

      setPhotosBusy(true)
      setPhotosError(null)
      try {
        const response = await fetch(
          `/api/items/${encodeURIComponent(selectedItemID)}/photos/reorder`,
          {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
              photo_ids: reordered.map((photo) => photo.id),
            }),
          }
        )
        if (!response.ok) {
          throw new Error(`reorder_photos_${response.status}`)
        }
        await loadInventoryPhotos()
      } catch {
        setPhotosError('Unable to reorder photos right now. Retry this action.')
      } finally {
        setPhotosBusy(false)
      }
    },
    [inventoryPhotos, loadInventoryPhotos, selectedItemID]
  )

  const handleRebuildPhotos = useCallback(async () => {
    if (!selectedItemID) {
      return
    }
    setPhotosBusy(true)
    setPhotosError(null)
    setPhotoRebuildError(null)
    setPhotoRebuildSuccess(null)
    try {
      const response = await fetch(
        `/api/items/${encodeURIComponent(selectedItemID)}/photos-rebuild`,
        {
          method: 'POST',
        }
      )
      if (!response.ok) {
        throw new Error(`rebuild_photos_${response.status}`)
      }
      setPhotoRebuildSuccess('Photo previews rebuilt successfully.')
      await loadInventoryPhotos()
    } catch {
      setPhotoRebuildError(
        'Unable to rebuild photo previews right now. Retry this action.'
      )
    } finally {
      setPhotosBusy(false)
    }
  }, [loadInventoryPhotos, selectedItemID])

  const handlePhotoInputChange = useCallback(
    async (
      event: ChangeEvent<HTMLInputElement>,
      source: 'capture' | 'upload'
    ) => {
      const file = event.target.files?.[0] ?? null
      setCameraError(null)
      setCameraSuccess(null)
      const uploaded = await handlePhotoUpload(file)
      if (uploaded && source === 'capture') {
        setCameraSuccess('Photo uploaded successfully.')
      }
      event.currentTarget.value = ''
    },
    [handlePhotoUpload]
  )

  const handleTakePhoto = useCallback(() => {
    setCameraError(null)
    setCameraSuccess(null)
    if (!selectedItemID) {
      setCameraError('Select an inventory item before taking a photo.')
      return
    }
    if (!photoCaptureInputRef.current) {
      setCameraError(
        'Camera picker is unavailable right now. Use Upload File instead.'
      )
      return
    }
    photoCaptureInputRef.current.click()
  }, [selectedItemID])

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
      const response = await fetch(
        `/api/barcodes/${encodeURIComponent(barcode)}`
      )
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

  const folderDragOverlay =
    typeof document !== 'undefined' && draggedFolderNode && dragPreviewPosition
      ? createPortal(
          <>
            <div
              data-testid='folder-tree-drag-preview'
              className='pointer-events-none fixed z-[70] w-64 rounded-lg border border-border/80 bg-background/95 p-3 shadow-2xl backdrop-blur'
              style={{
                left: `${dragPreviewPosition.x}px`,
                top: `${dragPreviewPosition.y}px`,
              }}
            >
              <div className='flex items-start gap-3'>
                <span className='mt-0.5 inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-md bg-accent/80 text-foreground/80'>
                  <GripVertical className='size-4' />
                </span>
                <div className='min-w-0 flex-1'>
                  <p className='truncate text-sm font-semibold text-foreground'>
                    {draggedFolderNode.name}
                  </p>
                  {draggedFolderNode.secondaryLabel ? (
                    <p className='truncate text-xs text-muted-foreground'>
                      {draggedFolderNode.secondaryLabel}
                    </p>
                  ) : null}
                </div>
                <div className='flex shrink-0 items-center gap-2 text-xs'>
                  {typeof draggedFolderNode.itemCount === 'number' ? (
                    <span className='font-medium text-muted-foreground tabular-nums'>
                      {draggedFolderNode.itemCount}
                    </span>
                  ) : null}
                  {draggedFolderNode.statusBadge ? (
                    <Badge
                      variant='secondary'
                      className='rounded-full px-2 py-0.5 text-[11px] font-semibold'
                    >
                      {draggedFolderNode.statusBadge}
                    </Badge>
                  ) : null}
                </div>
              </div>
            </div>
            {folderDropHint ? (
              <div
                data-testid='folder-tree-drop-hint'
                className='pointer-events-none fixed z-[71] rounded-full border border-primary/30 bg-primary/14 px-3 py-1 text-xs font-medium text-primary shadow-lg backdrop-blur'
                style={{
                  left: `${dragPreviewPosition.x}px`,
                  top: `${dragPreviewPosition.y + 74}px`,
                }}
              >
                {folderDropHint}
              </div>
            ) : null}
          </>,
          document.body
        )
      : null
  const inventoryCollectionFilter = isInventoryRoute ? (
    <div
      className='flex flex-wrap items-center gap-2'
      data-testid='inventory-collection-filter'
    >
      <select
        className='h-8 min-w-[11rem] rounded-md border bg-background px-2 text-sm'
        data-testid='inventory-collection-filter-select'
        value={activeFolderFilterValue}
        onChange={(event) => selectInventoryFolder(event.target.value)}
      >
        {folderFilterOptions.map((folder) => (
          <option key={folder.id} value={folder.name}>
            {'\u00a0\u00a0'.repeat(folder.level)}
            {folder.name}
          </option>
        ))}
      </select>
      <span
        className='text-xs text-muted-foreground'
        data-testid='inventory-collection-filter-selected'
      >
        {activeFolderFilterValue}
      </span>
      <DropdownMenu
        modal={false}
        open={collectionBrowserOpen}
        onOpenChange={setCollectionBrowserOpen}
      >
        <DropdownMenuTrigger asChild>
          <Button
            type='button'
            variant='outline'
            className='h-8 px-2 text-xs'
            data-testid='inventory-collection-browser-trigger'
          >
            Browse
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align='start' className='w-[min(32rem,92vw)] p-2'>
          <div
            role='tree'
            aria-label='Inventory folder filter'
            className='max-h-[22rem] overflow-auto rounded-md border p-2'
            data-testid='inventory-folder-browser-menu'
          >
            <div className='w-max min-w-full space-y-2'>
              {draggedFolderID ? (
                <div
                  role='presentation'
                  data-testid='folder-tree-root-drop-zone'
                  onDragEnter={(event) =>
                    handleFolderHTMLDragOver({ kind: 'root' }, event)
                  }
                  onDragOver={(event) =>
                    handleFolderHTMLDragOver({ kind: 'root' }, event)
                  }
                  onDrop={(event) =>
                    handleFolderHTMLDrop({ kind: 'root' }, event)
                  }
                  className='rounded-md border border-dashed border-primary/40 bg-primary/10 px-3 py-2 text-xs text-primary'
                >
                  Drop here to move folder to the root level
                </div>
              ) : null}
              {renderFolderTree(folderTree)}
            </div>
          </div>
        </DropdownMenuContent>
      </DropdownMenu>
      <Button
        type='button'
        variant='outline'
        className='h-8 px-2 text-xs'
        data-testid='inventory-collection-add-root'
        onClick={() => {
          setFolderCreateParentID(null)
          setFolderCreateName('')
          setFolderCreateOpen(true)
        }}
      >
        + New Collection
      </Button>
      {folderDragOverlay}
    </div>
  ) : undefined

  return (
    <TasksProvider>
      <Header
        fixed
        className='h-auto min-h-16'
        data-testid='inventory-shell-header'
      >
        <Search className='hidden min-w-32 sm:flex' />
        <HeaderTitle
          title={title}
          description={description}
          icon={ListChecks}
          testId='inventory-header-title'
          iconTestId='inventory-page-icon'
        />
        <div className='ms-auto flex min-w-0 shrink-0 items-center justify-end gap-2 sm:gap-3'>
          <div
            className='flex min-w-0 shrink-0 flex-nowrap items-center justify-end gap-1.5 sm:gap-2'
            data-testid='inventory-global-header-actions'
          >
            {isInventoryRoute ? (
              <>
                <Button
                  type='button'
                  variant='outline'
                  className='h-8 w-8 px-0 text-xs sm:h-9 sm:w-9 2xl:w-auto 2xl:px-4 2xl:text-sm'
                  data-testid='inventory-barcodes-action'
                  aria-label='Manage barcodes'
                  title='Barcodes'
                  onClick={() =>
                    openInventoryBarcodesForItem(
                      selectedInventoryItem ?? inventoryItems[0] ?? null
                    )
                  }
                >
                  <Barcode className='size-4 2xl:mr-2' aria-hidden />
                  <span className='hidden 2xl:inline'>Barcodes</span>
                </Button>
                <Button
                  type='button'
                  variant='outline'
                  className='h-8 w-8 px-0 text-xs sm:h-9 sm:w-9 2xl:w-auto 2xl:px-4 2xl:text-sm'
                  data-testid='inventory-photos-action'
                  aria-label='Manage photos'
                  title='Photos'
                  onClick={() =>
                    openInventoryPhotosForItem(
                      selectedInventoryItem ?? inventoryItems[0] ?? null
                    )
                  }
                >
                  <Images className='size-4 2xl:mr-2' aria-hidden />
                  <span className='hidden 2xl:inline'>Photos</span>
                </Button>
              </>
            ) : null}
            <Button
              type='button'
              className='h-8 w-8 px-0 text-xs sm:h-9 sm:w-9 2xl:w-auto 2xl:px-4 2xl:text-sm'
              data-testid='inventory-new-action'
              aria-label='New item'
              title='New'
              onClick={startCreateItem}
            >
              <Plus className='size-4 2xl:mr-2' aria-hidden />
              <span className='hidden 2xl:inline'>New</span>
            </Button>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button
                  type='button'
                  variant='outline'
                  className='h-8 w-8 px-0 text-xs sm:h-9 sm:w-9 2xl:w-auto 2xl:px-4 2xl:text-sm'
                  data-testid='inventory-create-menu-trigger'
                  aria-label='Create menu'
                  title='Create'
                >
                  <Ellipsis className='size-4 2xl:mr-2' aria-hidden />
                  <span className='hidden 2xl:inline'>Create</span>
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
                  onClick={() => {
                    setFolderCreateParentID(null)
                    setFolderCreateName('')
                    setFolderCreateOpen(true)
                  }}
                >
                  New Collection
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
          <Separator
            orientation='vertical'
            className='hidden h-6 md:block'
            data-testid='inventory-header-action-separator'
          />
          <div className='hidden items-center space-x-4 md:flex'>
            <LanguageSwitch />
            <ThemeSwitch />
            <ConfigDrawer />
            <ProfileDropdown />
          </div>
        </div>
      </Header>

      <Main className='space-y-3'>
        <div
          className='grid grid-cols-1 gap-4'
          data-testid='inventory-workspace'
        >
          <Card className='hidden'>
            <CardHeader>
              <CardTitle>Folders</CardTitle>
              <CardDescription>
                Browse folders before drilling into results.
              </CardDescription>
            </CardHeader>
            <CardContent
              data-testid='folder-tree-card-content'
              className='flex min-h-0 flex-1 flex-col gap-2'
            >
              <div
                data-testid='folder-tree-toolbar'
                className='flex justify-end gap-2'
              >
                <Button
                  type='button'
                  variant='outline'
                  size='sm'
                  className='min-w-11 font-semibold'
                  data-testid='folder-tree-sort-root-az'
                  aria-label='Sort root folders A to Z'
                  onClick={() => {
                    setFolderTree((previous) =>
                      sortRootFolderNodesAlphabetically(previous)
                    )
                  }}
                >
                  A/Z
                </Button>
                <Button
                  type='button'
                  variant='outline'
                  size='icon'
                  data-testid='folder-tree-add-root'
                  aria-label='Add root folder'
                  onClick={() => {
                    setFolderCreateParentID(null)
                    setFolderCreateName('')
                    setFolderCreateOpen(true)
                  }}
                >
                  +
                </Button>
              </div>
              <div
                role='tree'
                aria-label='Inventory folders'
                data-testid='inventory-folder-tree-legacy'
                className='max-h-[42rem] min-h-[26rem] flex-1 overflow-x-auto overflow-y-auto rounded-md border p-2'
              >
                <div
                  className='w-max min-w-full space-y-2'
                  data-testid='inventory-folder-tree-scroll-region'
                >
                  {draggedFolderID ? (
                    <div
                      role='presentation'
                      data-testid='folder-tree-root-drop-zone'
                      onDragEnter={(event) =>
                        handleFolderHTMLDragOver({ kind: 'root' }, event)
                      }
                      onDragOver={(event) =>
                        handleFolderHTMLDragOver({ kind: 'root' }, event)
                      }
                      onDrop={(event) =>
                        handleFolderHTMLDrop({ kind: 'root' }, event)
                      }
                      className='rounded-md border border-dashed border-primary/40 bg-primary/10 px-3 py-2 text-xs text-primary'
                    >
                      Drop here to move folder to the root level
                    </div>
                  ) : null}
                  {renderFolderTree(folderTree)}
                </div>
              </div>
              {folderDragOverlay}
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
                      {folderCreateParentID
                        ? 'Add Child Folder'
                        : 'Add Root Folder'}
                    </DialogTitle>
                  </DialogHeader>
                  <Input
                    data-testid='folder-tree-name-input'
                    placeholder='Folder name'
                    value={folderCreateName}
                    onChange={(event) =>
                      setFolderCreateName(event.target.value)
                    }
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
                      onClick={async () => {
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
                          return addChildFolder(
                            previous,
                            folderCreateParentID,
                            newNode
                          )
                        })
                        if (folderCreateParentID) {
                          setExpandedNodeIDs((previous) => {
                            const next = new Set(previous)
                            next.add(folderCreateParentID)
                            return next
                          })
                        }
                        const created = await addCollection(name)
                        if (!created && workspaceCollections.includes(name)) {
                          await setActiveWorkspaceCollection(name)
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
              <Sheet
                open={folderPropertiesOpen}
                onOpenChange={(open) => {
                  setFolderPropertiesOpen(open)
                  if (!open) {
                    setFolderPropertiesID(null)
                    setFolderPropertiesName('')
                    setFolderPropertiesCategory('General')
                    setFolderPropertiesSecondaryLabel('')
                    setFolderPropertiesStatusBadge('')
                  }
                }}
              >
                <SheetContent className='flex flex-col'>
                  <SheetHeader className='text-start'>
                    <SheetTitle>Folder Properties</SheetTitle>
                    <SheetDescription>
                      Update the selected folder title, category, secondary
                      label, and status badge.
                    </SheetDescription>
                  </SheetHeader>
                  <div className='grid gap-4 py-4'>
                    <div className='grid gap-2'>
                      <label
                        className='text-sm font-medium'
                        htmlFor='folder-properties-name'
                      >
                        Title
                      </label>
                      <Input
                        id='folder-properties-name'
                        data-testid='folder-properties-name-input'
                        value={folderPropertiesName}
                        onChange={(event) =>
                          setFolderPropertiesName(event.target.value)
                        }
                      />
                    </div>
                    <div className='grid gap-2'>
                      <label
                        className='text-sm font-medium'
                        htmlFor='folder-properties-category'
                      >
                        Category
                      </label>
                      <select
                        id='folder-properties-category'
                        data-testid='folder-properties-category-select'
                        className='h-9 rounded-md border bg-background px-2 text-sm'
                        value={folderPropertiesCategory}
                        onChange={(event) =>
                          setFolderPropertiesCategory(event.target.value)
                        }
                      >
                        {folderCategoryOptions.map((category) => (
                          <option key={category} value={category}>
                            {category}
                          </option>
                        ))}
                      </select>
                    </div>
                    <div className='grid gap-2'>
                      <label
                        className='text-sm font-medium'
                        htmlFor='folder-properties-secondary-label'
                      >
                        Secondary label
                      </label>
                      <Input
                        id='folder-properties-secondary-label'
                        data-testid='folder-properties-secondary-label-input'
                        placeholder='Aisle B'
                        value={folderPropertiesSecondaryLabel}
                        onChange={(event) =>
                          setFolderPropertiesSecondaryLabel(event.target.value)
                        }
                      />
                    </div>
                    <div className='grid gap-2'>
                      <label
                        className='text-sm font-medium'
                        htmlFor='folder-properties-status-badge'
                      >
                        Status badge
                      </label>
                      <Input
                        id='folder-properties-status-badge'
                        data-testid='folder-properties-status-badge-input'
                        placeholder='Cold'
                        value={folderPropertiesStatusBadge}
                        onChange={(event) =>
                          setFolderPropertiesStatusBadge(event.target.value)
                        }
                      />
                    </div>
                  </div>
                  <SheetFooter className='gap-2'>
                    <Button
                      type='button'
                      variant='outline'
                      data-testid='folder-properties-cancel'
                      onClick={() => setFolderPropertiesOpen(false)}
                    >
                      Cancel
                    </Button>
                    <Button
                      type='button'
                      data-testid='folder-properties-save'
                      onClick={() => {
                        const name = folderPropertiesName.trim()
                        if (!folderPropertiesID || name === '') {
                          return
                        }
                        setFolderTree((previous) =>
                          updateFolderNodeByID(
                            previous,
                            folderPropertiesID,
                            (node) => ({
                              ...node,
                              name,
                              category:
                                folderPropertiesCategory.trim() || 'General',
                              secondaryLabel:
                                folderPropertiesSecondaryLabel.trim() ||
                                undefined,
                              statusBadge:
                                folderPropertiesStatusBadge.trim() || undefined,
                            })
                          )
                        )
                        setActiveFolder(name)
                        setFolderPropertiesOpen(false)
                      }}
                    >
                      Save Properties
                    </Button>
                  </SheetFooter>
                </SheetContent>
              </Sheet>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Collection Browser</CardTitle>
            </CardHeader>
            <CardContent className='space-y-4'>
              <p className='text-sm text-muted-foreground'>
                Folders: <strong>{summary.folders}</strong>{' '}
                <span className='mx-2'>
                  Items: <strong>{summary.items}</strong>
                </span>
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
              {selectedFolderHasNoItems ? (
                <div
                  className='rounded-md border border-dashed bg-muted/30 p-4 text-sm text-muted-foreground'
                  data-testid='inventory-selected-folder-empty'
                >
                  No items are assigned to {activeFolderFilterValue} yet. Use
                  the row collection action to place inventory here.
                </div>
              ) : null}
              <TasksTable
                data={visibleTableData}
                routePath={routePath}
                currentRecordID={selectedItemID}
                customFilters={inventoryCollectionFilter}
                onRecordFocus={(itemID, recordID) => {
                  const matchedItem =
                    inventoryItems.find((item) => item.id === itemID) ??
                    inventoryItems.find(
                      (item) => item.part_number === recordID
                    ) ??
                    null
                  setItemEditorMode('edit')
                  setItemSaveError(null)
                  setItemSaveSuccess(null)
                  selectInventoryItem(matchedItem)
                }}
                onEditRow={(task) => {
                  const matchedItem = resolveInventoryItemFromTask(task)
                  openInventoryItemEditor(matchedItem)
                }}
                onPhotoRow={(task) => {
                  const matchedItem = resolveInventoryItemFromTask(task)
                  openInventoryPhotosForItem(matchedItem)
                }}
                onBarcodeRow={(task) => {
                  const matchedItem = resolveInventoryItemFromTask(task)
                  openInventoryBarcodesForItem(matchedItem)
                }}
                onAssignCollectionRow={(task) => {
                  const matchedItem = resolveInventoryItemFromTask(task)
                  openInventoryAssignCollectionForItem(matchedItem)
                }}
              />
              {isInventoryRoute ? (
                <>
                  <Dialog
                    open={assignCollectionOpen}
                    onOpenChange={(open) => {
                      setAssignCollectionOpen(open)
                      if (!open) {
                        setAssignCollectionError(null)
                      }
                    }}
                  >
                    <DialogContent data-testid='inventory-assign-collection-dialog'>
                      <div className='space-y-4'>
                        <DialogHeader>
                          <DialogTitle>Assign to Collection</DialogTitle>
                        </DialogHeader>
                        <p className='text-sm text-muted-foreground'>
                          {assignCollectionItem
                            ? `Choose a collection for ${assignCollectionItem.title}.`
                            : 'Choose an inventory item before assigning.'}
                        </p>
                        <div className='grid gap-2'>
                          <label
                            className='text-sm font-medium'
                            htmlFor='inventory-assign-collection-select'
                          >
                            Collection
                          </label>
                          <select
                            id='inventory-assign-collection-select'
                            data-testid='inventory-assign-collection-select'
                            className='h-9 rounded-md border bg-background px-2 text-sm'
                            value={assignCollectionName}
                            onChange={(event) =>
                              setAssignCollectionName(event.target.value)
                            }
                          >
                            {assignableWorkspaceCollections.map(
                              (collection) => (
                                <option key={collection} value={collection}>
                                  {collection}
                                </option>
                              )
                            )}
                          </select>
                        </div>
                        {assignCollectionError ? (
                          <div
                            className='rounded-md border border-destructive/40 bg-destructive/10 p-3 text-sm'
                            data-testid='inventory-assign-collection-error'
                          >
                            {assignCollectionError}
                          </div>
                        ) : null}
                        <DialogFooter>
                          <Button
                            type='button'
                            variant='outline'
                            onClick={() => setAssignCollectionOpen(false)}
                          >
                            Cancel
                          </Button>
                          <Button
                            type='button'
                            data-testid='inventory-assign-collection-submit'
                            disabled={
                              !assignCollectionItem || !assignCollectionName
                            }
                            onClick={() =>
                              void handleAssignInventoryItemToCollection()
                            }
                          >
                            Assign
                          </Button>
                        </DialogFooter>
                      </div>
                    </DialogContent>
                  </Dialog>
                  <Dialog
                    open={itemEditorOpen && itemEditorSurface === 'dialog'}
                    onOpenChange={(open) => {
                      setItemEditorOpen(open)
                      if (!open) {
                        setItemSaveError(null)
                      }
                    }}
                  >
                    <DialogContent data-testid='inventory-item-editor-dialog'>
                      <div
                        className='space-y-4'
                        data-testid={
                          itemEditorMode === 'create'
                            ? 'inventory-item-create-dialog'
                            : 'inventory-item-edit-dialog'
                        }
                      >
                        <DialogHeader>
                          <DialogTitle>
                            {itemEditorMode === 'create'
                              ? 'Create Item'
                              : 'Edit Item'}
                          </DialogTitle>
                        </DialogHeader>
                        <p
                          className='text-sm text-muted-foreground'
                          data-testid='inventory-item-editor-mode'
                        >
                          {itemEditorMode === 'create'
                            ? 'Creating new item draft.'
                            : `Editing selected item: ${selectedItemContext}`}
                        </p>
                        <div className='grid grid-cols-1 gap-3 md:grid-cols-2'>
                          <div className='space-y-2'>
                            <label
                              className='text-sm font-medium'
                              htmlFor='inventory-item-part-number'
                            >
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
                            <label
                              className='text-sm font-medium'
                              htmlFor='inventory-item-title'
                            >
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
                            <label
                              className='text-sm font-medium'
                              htmlFor='inventory-item-brand'
                            >
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
                            <label
                              className='text-sm font-medium'
                              htmlFor='inventory-item-category'
                            >
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
                          <label
                            className='text-sm font-medium'
                            htmlFor='inventory-item-description'
                          >
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
                        <DialogFooter className='flex flex-wrap gap-2 sm:justify-between'>
                          {itemEditorMode === 'edit' ? (
                            <div className='flex flex-1 gap-2'>
                              <Button
                                type='button'
                                variant='outline'
                                data-testid='inventory-item-editor-previous'
                                disabled={!canNavigateToPreviousInventoryItem}
                                onClick={() => openAdjacentInventoryItem(-1)}
                              >
                                Previous
                              </Button>
                              <Button
                                type='button'
                                variant='outline'
                                data-testid='inventory-item-editor-next'
                                disabled={!canNavigateToNextInventoryItem}
                                onClick={() => openAdjacentInventoryItem(1)}
                              >
                                Next
                              </Button>
                            </div>
                          ) : null}
                          <div className='flex gap-2'>
                            <Button
                              type='button'
                              variant='outline'
                              data-testid='inventory-item-editor-cancel'
                              onClick={() => setItemEditorOpen(false)}
                            >
                              <span
                                data-testid={
                                  itemEditorMode === 'create'
                                    ? 'inventory-item-create-cancel'
                                    : undefined
                                }
                              >
                                Cancel
                              </span>
                            </Button>
                            <Button
                              type='button'
                              data-testid='inventory-item-save'
                              disabled={itemSaveBusy}
                              onClick={() => void handleSaveItem()}
                            >
                              <span
                                data-testid={
                                  itemEditorMode === 'create'
                                    ? 'inventory-item-create-submit'
                                    : undefined
                                }
                              >
                                {itemEditorMode === 'create'
                                  ? 'Create Item'
                                  : 'Save Changes'}
                              </span>
                            </Button>
                          </div>
                        </DialogFooter>
                      </div>
                    </DialogContent>
                  </Dialog>
                  <Sheet
                    open={itemEditorOpen && itemEditorSurface === 'panel'}
                    onOpenChange={(open) => {
                      setItemEditorOpen(open)
                      if (!open) {
                        setItemSaveError(null)
                      }
                    }}
                  >
                    <SheetContent
                      side='right'
                      className='w-[min(28rem,92vw)] overflow-y-auto sm:max-w-md'
                      data-testid='inventory-item-editor-panel'
                    >
                      <SheetHeader className='text-start'>
                        <SheetTitle>Edit Item</SheetTitle>
                        <SheetDescription data-testid='inventory-item-editor-mode'>
                          Editing selected item: {selectedItemContext}
                        </SheetDescription>
                      </SheetHeader>
                      <div
                        className='space-y-4 px-4'
                        data-testid='inventory-item-edit-panel'
                      >
                        <div className='grid grid-cols-1 gap-3'>
                          <div className='space-y-2'>
                            <label
                              className='text-sm font-medium'
                              htmlFor='inventory-panel-item-part-number'
                            >
                              Part Number
                            </label>
                            <Input
                              id='inventory-panel-item-part-number'
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
                            <label
                              className='text-sm font-medium'
                              htmlFor='inventory-panel-item-title'
                            >
                              Title
                            </label>
                            <Input
                              id='inventory-panel-item-title'
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
                            <label
                              className='text-sm font-medium'
                              htmlFor='inventory-panel-item-brand'
                            >
                              Brand
                            </label>
                            <Input
                              id='inventory-panel-item-brand'
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
                            <label
                              className='text-sm font-medium'
                              htmlFor='inventory-panel-item-category'
                            >
                              Category
                            </label>
                            <Input
                              id='inventory-panel-item-category'
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
                          <label
                            className='text-sm font-medium'
                            htmlFor='inventory-panel-item-description'
                          >
                            Description
                          </label>
                          <Input
                            id='inventory-panel-item-description'
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
                      </div>
                      <SheetFooter className='gap-2 sm:flex-row sm:justify-between'>
                        <div className='flex gap-2'>
                          <Button
                            type='button'
                            variant='outline'
                            data-testid='inventory-item-editor-previous'
                            disabled={!canNavigateToPreviousInventoryItem}
                            onClick={() => openAdjacentInventoryItem(-1)}
                          >
                            Previous
                          </Button>
                          <Button
                            type='button'
                            variant='outline'
                            data-testid='inventory-item-editor-next'
                            disabled={!canNavigateToNextInventoryItem}
                            onClick={() => openAdjacentInventoryItem(1)}
                          >
                            Next
                          </Button>
                        </div>
                        <div className='flex gap-2'>
                          <SheetClose asChild>
                            <Button
                              type='button'
                              variant='outline'
                              data-testid='inventory-item-editor-cancel'
                            >
                              Cancel
                            </Button>
                          </SheetClose>
                          <Button
                            type='button'
                            data-testid='inventory-item-save'
                            disabled={itemSaveBusy}
                            onClick={() => void handleSaveItem()}
                          >
                            Save Changes
                          </Button>
                        </div>
                      </SheetFooter>
                    </SheetContent>
                  </Sheet>
                  <Dialog
                    open={photosDialogOpen}
                    onOpenChange={setPhotosDialogOpen}
                  >
                    <DialogContent
                      className='max-h-[86vh] overflow-y-auto sm:max-w-4xl'
                      closeButtonTestId='inventory-photos-dialog-close'
                      data-testid='inventory-photos-dialog'
                    >
                      <section
                        className='space-y-3'
                        data-testid='inventory-photos-panel'
                      >
                        <div className='flex flex-wrap items-start justify-between gap-3 pr-8'>
                          <div>
                            <DialogHeader>
                              <DialogTitle>Photos</DialogTitle>
                            </DialogHeader>
                            <p
                              className='text-sm text-muted-foreground'
                              data-testid='inventory-photos-item-context'
                            >
                              {selectedInventoryItem
                                ? `Managing photos for ${selectedInventoryItem.title}`
                                : 'Select an inventory item before managing photos.'}
                            </p>
                          </div>
                        </div>
                        <div className='sr-only'>
                          Review item media and inspect photos in fullscreen
                          mode.
                        </div>
                        <div className='flex flex-wrap items-center gap-2'>
                          <input
                            ref={photoCaptureInputRef}
                            id='inventory-photo-capture-input'
                            type='file'
                            accept='image/*'
                            capture='environment'
                            className='sr-only'
                            data-testid='inventory-photo-capture-input'
                            disabled={!selectedItemID}
                            onChange={(event) => {
                              void handlePhotoInputChange(event, 'capture')
                            }}
                          />
                          <Button
                            type='button'
                            variant='outline'
                            aria-controls='inventory-photo-capture-input'
                            data-testid='inventory-camera-take-photo'
                            onClick={handleTakePhoto}
                            disabled={!selectedItemID}
                          >
                            Take Photo
                          </Button>
                          <input
                            ref={photoUploadInputRef}
                            type='file'
                            accept='image/*'
                            data-testid='inventory-photo-upload-input'
                            disabled={!selectedItemID}
                            onChange={(event) => {
                              void handlePhotoInputChange(event, 'upload')
                            }}
                          />
                          <Button
                            type='button'
                            variant='outline'
                            data-testid='inventory-photo-rebuild'
                            onClick={() => void handleRebuildPhotos()}
                            disabled={photosBusy || selectedItemID === ''}
                          >
                            Rebuild Photos
                          </Button>
                          {photosBusy ? (
                            <span className='text-xs text-muted-foreground'>
                              Working...
                            </span>
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
                        {photoRebuildError ? (
                          <div
                            className='rounded-md border border-destructive/40 bg-destructive/10 p-3 text-sm'
                            data-testid='inventory-photo-rebuild-error'
                          >
                            <p className='font-medium'>Photo rebuild failed.</p>
                            <p className='mt-1 text-muted-foreground'>
                              {photoRebuildError}
                            </p>
                            <Button
                              className='mt-3'
                              size='sm'
                              variant='outline'
                              data-testid='inventory-photo-rebuild-retry'
                              onClick={() => void handleRebuildPhotos()}
                            >
                              Retry
                            </Button>
                          </div>
                        ) : null}
                        {photoRebuildSuccess ? (
                          <div
                            className='rounded-md border border-emerald-500/40 bg-emerald-500/10 p-2 text-sm'
                            data-testid='inventory-photo-rebuild-success'
                          >
                            {photoRebuildSuccess}
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
                            <p className='font-medium'>
                              Photos are unavailable.
                            </p>
                            <p className='mt-1 text-muted-foreground'>
                              {photosError}
                            </p>
                            <Button
                              className='mt-3'
                              size='sm'
                              variant='outline'
                              onClick={() => void loadInventoryPhotos()}
                              disabled={!selectedItemID}
                            >
                              Retry
                            </Button>
                          </div>
                        ) : null}
                        {!photosLoading &&
                        !photosError &&
                        inventoryPhotos.length === 0 ? (
                          <div
                            className='rounded-md border border-dashed p-3 text-sm text-muted-foreground'
                            data-testid='inventory-photos-empty'
                          >
                            No photos yet for the selected item. Upload an image
                            to begin.
                          </div>
                        ) : null}
                        {!photosLoading &&
                        !photosError &&
                        inventoryPhotos.length > 0 ? (
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
                                  <div className='p-2 text-sm'>
                                    {photo.filename}
                                  </div>
                                </button>
                              ))}
                            </div>
                            <div className='space-y-2'>
                              {inventoryPhotos.map((photo, index) => (
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
                                      data-testid='inventory-photo-move-up'
                                      onClick={() =>
                                        void handleReorderPhotos(photo.id, 'up')
                                      }
                                      disabled={photosBusy || index === 0}
                                    >
                                      Move Up
                                    </Button>
                                    <Button
                                      size='sm'
                                      variant='outline'
                                      data-testid='inventory-photo-move-down'
                                      onClick={() =>
                                        void handleReorderPhotos(
                                          photo.id,
                                          'down'
                                        )
                                      }
                                      disabled={
                                        photosBusy ||
                                        index === inventoryPhotos.length - 1
                                      }
                                    >
                                      Move Down
                                    </Button>
                                    <Button
                                      size='sm'
                                      variant='outline'
                                      data-testid='inventory-photo-set-primary'
                                      onClick={() =>
                                        void handleSetPrimaryPhoto(photo.id)
                                      }
                                    >
                                      Set Primary
                                    </Button>
                                    <Button
                                      size='sm'
                                      variant='outline'
                                      data-testid='inventory-photo-delete'
                                      onClick={() =>
                                        void handleDeletePhoto(photo.id)
                                      }
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
                    </DialogContent>
                  </Dialog>
                  <Dialog
                    open={barcodesDialogOpen}
                    onOpenChange={setBarcodesDialogOpen}
                  >
                    <DialogContent
                      className='max-h-[86vh] overflow-y-auto sm:max-w-4xl'
                      closeButtonTestId='inventory-barcodes-dialog-close'
                      data-testid='inventory-barcodes-dialog'
                    >
                      <section
                        className='space-y-3'
                        data-testid='inventory-barcodes-panel'
                      >
                        <div className='flex flex-wrap items-start justify-between gap-3 pr-8'>
                          <div>
                            <DialogHeader>
                              <DialogTitle>Barcodes</DialogTitle>
                            </DialogHeader>
                            <p
                              className='text-sm text-muted-foreground'
                              data-testid='inventory-barcodes-item-context'
                            >
                              {selectedInventoryItem
                                ? `Managing barcodes for ${selectedInventoryItem.title}`
                                : 'Select an inventory item before managing barcodes.'}
                            </p>
                          </div>
                        </div>
                        <div className='sr-only'>
                          Add barcodes to the selected item, run local lookup,
                          and continue with external fallback when there is no
                          local match.
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
                              onClick={() =>
                                void lookupBarcode(barcodeLookupInput)
                              }
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
                            <p className='font-medium'>
                              Barcode lookup failed.
                            </p>
                            <p className='mt-1 text-muted-foreground'>
                              {barcodeLookupError}
                            </p>
                            <Button
                              className='mt-3'
                              size='sm'
                              variant='outline'
                              data-testid='inventory-barcodes-lookup-retry'
                              onClick={() =>
                                void lookupBarcode(lastLookupBarcode)
                              }
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
                            <p className='font-medium'>
                              No local barcode match.
                            </p>
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
                    </DialogContent>
                  </Dialog>
                  <section
                    className='space-y-3 rounded-md border p-4'
                    data-testid='inventory-ai-assist-section'
                  >
                    <div>
                      <h3 className='text-base font-semibold'>AI Assist</h3>
                      <p className='text-sm text-muted-foreground'>
                        Generate title and photo suggestions with explicit
                        confirm-before-apply.
                      </p>
                    </div>
                    <div className='grid grid-cols-1 gap-3 md:grid-cols-2'>
                      <div className='space-y-2'>
                        <Input
                          placeholder='Paste listing title'
                          data-testid='inventory-ai-title-input'
                          value={aiTitleInput}
                          onChange={(event) =>
                            setAITitleInput(event.target.value)
                          }
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
                          onChange={(event) =>
                            setAIPhotoURLInput(event.target.value)
                          }
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
                          <strong>Part:</strong>{' '}
                          {aiSuggestion.part_number || 'n/a'}
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
              <DialogTitle>
                {selectedPhoto?.filename ?? 'Photo Viewer'}
              </DialogTitle>
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
      <TasksDialogs
        routePath='/_authenticated/inventory/'
        open={tasksDialogOpen}
        setOpen={setTasksDialogOpen}
        currentRow={tasksDialogRow}
        setCurrentRow={setTasksDialogRow}
      />
    </TasksProvider>
  )
}
