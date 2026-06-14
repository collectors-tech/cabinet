import { useCallback, useMemo, useState } from 'react'
import {
  Inbox,
  CheckCircle2,
  FileUp,
  Loader2,
  PackagePlus,
  Plus,
  RefreshCw,
  ShieldCheck,
  ShoppingCart,
  Star,
  Truck,
} from 'lucide-react'
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
import { Label } from '@/components/ui/label'
import { Separator } from '@/components/ui/separator'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Textarea } from '@/components/ui/textarea'
import { ConfigDrawer } from '@/components/config-drawer'
import { LanguageSwitch } from '@/components/language-switch'
import { Header, HeaderTitle } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'

type PurchaseCard = {
  order_id?: string
  listing_id?: string
  transaction_id?: string
  listing_title?: string
  purchased_identity?: string
  quantity?: number
  item_price?: string
  item_url?: string
  purchase_date?: string
  seller_username?: string
  order_total?: string
  currency?: string
}

type PurchaseInboxAction = {
  id: string
  label: string
  scope: string
  target_key: string
  requires_confirmation: boolean
}

type PurchaseInboxItem = {
  item: PurchaseCard
  status: string
  missing_fields?: string[]
  suggested_actions?: PurchaseInboxAction[]
}

type PurchaseInboxReview = {
  status: string
  order: {
    order_id?: string
    seller_usernames?: string[]
    order_total?: string
    currency?: string
  }
  items: PurchaseInboxItem[]
}

type PurchaseTableStatusFilter =
  | 'all'
  | 'ready_to_link_or_convert'
  | 'needs_review'
  | 'manual_draft'
  | 'csv_import'
  | 'email_import'

type PurchaseTableRow = {
  key: string
  title: string
  source: string
  price: string
  purchaseDate: string
  delivery: string
  status: string
  tracking: string
  orderLink: string
  persistence: string
  actionCount: number
  searchText: string
}

type ManualPurchaseForm = {
  title: string
  source: string
  price: string
  tracking: string
}

type ManualPurchaseDraft = ManualPurchaseForm & {
  key: string
  persistence?: PurchaseDraftPersistence
}

type PurchaseDraftPersistence = {
  itemID: string
  lifecycleEntryID: string
  expectedArrivalID: string
}

type PurchaseImportDraft = {
  key: string
  mode: 'csv' | 'email'
  provenance: string
  title: string
  price: string
  currency: string
  purchaseDate: string
  seller: string
  channel: string
  tracking: string
  delivery: string
  persistence?: PurchaseDraftPersistence
}

type CreatedItemResponse = {
  id?: string
}

type CommerceLifecycleResponse = {
  entry?: {
    id?: string
    expected_arrival_id?: string
  }
  expected_arrival?: {
    id?: string
  } | null
}

type ReviewResponse = {
  source?: string
  profile_id?: string
  reviews?: PurchaseInboxReview[]
}

type ForwarderPackage = {
  id: string
  profile_id: string
  provider: string
  source: string
  external_package_id: string
  shipment_id?: string
  tracking_number?: string
  status: string
  received_at?: string
  sender?: string
  warehouse_location?: string
  weight_grams?: number
  provenance_key: string
  raw_payload?: Record<string, unknown>
  created_at?: string
  updated_at?: string
}

type ForwarderPackageListResponse = {
  packages?: ForwarderPackage[]
  summary?: {
    count?: number
  }
}

type PackageImportForm = {
  profile_id: string
  provider: string
  source: string
  external_package_id: string
  status: string
  shipment_id: string
  tracking_number: string
  sender: string
  warehouse_location: string
  weight_grams: string
}

type ForwarderPackageCSVError = {
  row?: number
  error: string
}

type ForwarderPackageCSVImportResponse = {
  imported?: ForwarderPackage[]
  errors?: ForwarderPackageCSVError[]
  summary?: {
    imported?: number
    errors?: number
  }
}

type ForwarderPackageLink = {
  id: string
  profile_id: string
  package_id: string
  item_id: string
  lifecycle_entry_id?: string
  expected_arrival_id?: string
  source: string
  decision?: string
  notes?: string
  audit_trail?: string[]
  created_at?: string
  updated_at?: string
}

type ForwarderPackageLinkEvent = {
  id: string
  package_id: string
  link_id?: string
  action: string
  item_id?: string
  lifecycle_entry_id?: string
  expected_arrival_id?: string
  previous_item_id?: string
  previous_lifecycle_entry_id?: string
  previous_expected_arrival_id?: string
  source: string
  notes?: string
  audit_trail?: string[]
  created_at?: string
}

type ForwarderPackageLinkForm = {
  item_id: string
  lifecycle_entry_id: string
  expected_arrival_id: string
  source: string
  notes: string
  audit_trail?: string[]
}

type ForwarderPackageMatchSignal = {
  name?: string
  score?: number
  evidence?: string
}

type ForwarderPackageMatchSuggestion = {
  id?: string
  package_id: string
  item_id: string
  lifecycle_entry_id?: string
  expected_arrival_id?: string
  confidence_score?: number
  confidence_label?: string
  signals?: ForwarderPackageMatchSignal[]
  explanations?: string[]
  audit_trail?: string[]
}

type ForwarderPackageMatchSuggestionResponse = {
  confidence_filter?: string
  suggestions?: ForwarderPackageMatchSuggestion[]
  summary?: {
    count?: number
    high_confidence?: number
    medium_confidence?: number
    low_confidence?: number
    scoped_packages?: number
  }
}

type ForwarderPackageMatchSuggestionSummary = NonNullable<
  ForwarderPackageMatchSuggestionResponse['summary']
>

type ForwarderPackageReviewFilter = 'all' | 'linked' | 'unlinked' | 'suggested'

type ForwarderPackageSuggestionConfidenceFilter =
  | 'all'
  | 'high'
  | 'medium'
  | 'low'

const sampleCards: PurchaseCard[] = [
  {
    order_id: '20-14595-70928',
    listing_id: '316046161178',
    transaction_id: '10080684936020',
    listing_title: 'Accompanying Flute listing',
    purchased_identity: 'Accompanying Flute TWM 142 (142/167)',
    quantity: 4,
    item_price: 'AU $2.40',
    seller_username: 'seller-one',
    order_total: 'AU $8.10',
    currency: 'AUD',
  },
  {
    order_id: '20-14595-70928',
    listing_id: '316046161179',
    listing_title: 'Mystery purchase',
    seller_username: 'seller-one',
  },
]

const defaultPackageImport: PackageImportForm = {
  profile_id: 'e2e-profile-001',
  provider: 'stackry',
  source: 'manual',
  external_package_id: 'STK-PKG-1001',
  status: 'received',
  shipment_id: 'SHIP-1001',
  tracking_number: '1Z999AA10123456784',
  sender: 'Stackry warehouse intake',
  warehouse_location: 'Locker A-12',
  weight_grams: '420',
}

const defaultPackageCSV = [
  'Stackry Package ID,Status,Shipment ID,Tracking Number,Warehouse Location,Weight Grams',
  'STK-CSV-2001,received,SHIP-2001,1ZCSV2001,Locker C-4,520',
].join('\n')

const defaultPackageEmail = [
  'Package ID: STK-EMAIL-3001',
  'Status: Received',
  'Shipment ID: SHIP-3001',
  'Tracking Number: 1ZEMAIL3001',
  'Warehouse Location: Locker E-5',
  'Weight Grams: 640',
  'Sender: Stackry Intake',
].join('\n')

const defaultPackageLinkForm: ForwarderPackageLinkForm = {
  item_id: 'item-expected-001',
  lifecycle_entry_id: 'life-entry-001',
  expected_arrival_id: 'arrival-expected-001',
  source: 'manual_review',
  notes: 'Matched from package inbox review',
}

const defaultManualPurchaseForm: ManualPurchaseForm = {
  title: '',
  source: 'manual',
  price: '',
  tracking: '',
}

const defaultPurchaseCSVImport = [
  'source,title,price,currency,purchase_date,seller,channel,tracking,delivery',
  'Amazon,Manual Amazon order,42.50,AUD,2026-06-08,Amazon AU,email,TBA123456,Expected 2026-06-12',
].join('\n')

const defaultPurchaseEmailImport = [
  'Source: eBay',
  'Title: Accompanying Flute TWM 142',
  'Price: AU $2.40',
  'Purchase Date: 2026-06-08',
  'Seller: seller-one',
  'Channel: email',
  'Tracking: 1ZEMAILPURCHASE',
  'Delivery: Expected 2026-06-14',
].join('\n')

function actionTone(action: PurchaseInboxAction) {
  if (action.requires_confirmation) {
    return 'border-amber-300 bg-amber-50 text-amber-950 dark:border-amber-900 dark:bg-amber-950/30 dark:text-amber-100'
  }
  return 'border-border bg-muted/40 text-muted-foreground'
}

function labelForStatus(status: string) {
  return status.split('_').join(' ')
}

function packageDetailValue(value?: string | number) {
  if (value === undefined || value === null || value === '') {
    return 'Pending'
  }
  return value
}

function persistenceLabel(persistence?: PurchaseDraftPersistence) {
  if (!persistence) {
    return 'Not persisted'
  }
  return (
    'Persisted lifecycle ' +
    persistence.lifecycleEntryID.slice(0, 8) +
    ' / arrival ' +
    persistence.expectedArrivalID.slice(0, 8)
  )
}

function parsePurchaseAmount(price: string) {
  const match = price.replace(/,/g, '').match(/-?\d+(?:\.\d+)?/)
  return match ? Number(match[0]) : 0
}

function purchasePartNumber(title: string, key: string) {
  const slug = title
    .toUpperCase()
    .replace(/[^A-Z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
    .slice(0, 24)
  return 'PUR-' + (slug || 'DRAFT') + '-' + key.slice(-6).toUpperCase()
}

function valueFromImportRow(
  row: Record<string, string>,
  keys: string[],
  fallback = ''
) {
  const match = keys.find((key) => row[key]?.trim())
  return match ? row[match].trim() : fallback
}

function parsePurchaseCSVImport(csv: string): PurchaseImportDraft[] {
  const lines = csv
    .split(/\r?\n/)
    .map((line) => line.trim())
    .filter(Boolean)
  if (lines.length < 2) {
    return []
  }

  const headers = lines[0].split(',').map((header) =>
    header
      .trim()
      .toLowerCase()
      .replace(/[\s-]+/g, '_')
  )

  return lines.slice(1).flatMap((line, index) => {
    const values = line.split(',').map((value) => value.trim())
    const row = headers.reduce<Record<string, string>>((current, header, i) => {
      current[header] = values[i] ?? ''
      return current
    }, {})
    const title = valueFromImportRow(row, ['title', 'item', 'purchase'])
    if (!title) {
      return []
    }
    const source = valueFromImportRow(row, ['source', 'provider'], 'CSV')
    const currency = valueFromImportRow(row, ['currency'], 'AUD')
    const price = valueFromImportRow(row, ['price', 'amount', 'total'])

    return [
      {
        key:
          'csv-import-preview-' +
          index +
          '-' +
          title.toLowerCase().replace(/[^a-z0-9]+/g, '-'),
        mode: 'csv' as const,
        provenance: source + ' CSV row ' + (index + 2),
        title,
        price: price ? currency + ' ' + price.replace(/^[A-Z]{3}\s+/, '') : '-',
        currency,
        purchaseDate: valueFromImportRow(row, ['purchase_date', 'date']),
        seller: valueFromImportRow(row, ['seller', 'merchant', 'source_name']),
        channel: valueFromImportRow(row, ['channel'], 'csv'),
        tracking: valueFromImportRow(row, ['tracking', 'tracking_number']),
        delivery: valueFromImportRow(row, ['delivery', 'delivery_status']),
      },
    ]
  })
}

function parsePurchaseEmailImport(message: string): PurchaseImportDraft[] {
  const fields = message
    .split(/\r?\n/)
    .reduce<Record<string, string>>((current, line) => {
      const [label, ...rest] = line.split(':')
      if (!label || rest.length === 0) {
        return current
      }
      current[
        label
          .trim()
          .toLowerCase()
          .replace(/[\s-]+/g, '_')
      ] = rest.join(':').trim()
      return current
    }, {})
  const title = valueFromImportRow(fields, ['title', 'item', 'purchase'])
  if (!title) {
    return []
  }
  const source = valueFromImportRow(fields, ['source', 'provider'], 'Email')
  const price = valueFromImportRow(fields, ['price', 'amount', 'total'], '-')

  return [
    {
      key:
        'email-import-preview-' +
        title.toLowerCase().replace(/[^a-z0-9]+/g, '-'),
      mode: 'email',
      provenance: source + ' pasted email text',
      title,
      price,
      currency: valueFromImportRow(fields, ['currency'], 'AUD'),
      purchaseDate: valueFromImportRow(fields, ['purchase_date', 'date']),
      seller: valueFromImportRow(fields, ['seller', 'merchant', 'source_name']),
      channel: valueFromImportRow(fields, ['channel'], 'email'),
      tracking: valueFromImportRow(fields, ['tracking', 'tracking_number']),
      delivery: valueFromImportRow(fields, ['delivery', 'delivery_status']),
    },
  ]
}

function purchaseRowKey(review: PurchaseInboxReview, item: PurchaseInboxItem) {
  return (
    item.item.transaction_id ??
    item.item.listing_id ??
    review.order.order_id ??
    item.item.listing_title ??
    item.item.purchased_identity ??
    'purchase-row'
  )
}

export function Purchases() {
  const [reviews, setReviews] = useState<PurchaseInboxReview[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [packages, setPackages] = useState<ForwarderPackage[]>([])
  const [packagesLoading, setPackagesLoading] = useState(false)
  const [packageError, setPackageError] = useState<string | null>(null)
  const [packageResult, setPackageResult] = useState<string | null>(null)
  const [packageCSV, setPackageCSV] = useState(defaultPackageCSV)
  const [packageEmail, setPackageEmail] = useState(defaultPackageEmail)
  const [packageCSVErrors, setPackageCSVErrors] = useState<
    ForwarderPackageCSVError[]
  >([])
  const [packageLinks, setPackageLinks] = useState<
    Record<string, ForwarderPackageLink[]>
  >({})
  const [packageLinkEvents, setPackageLinkEvents] = useState<
    Record<string, ForwarderPackageLinkEvent[]>
  >({})
  const [packageLinkForms, setPackageLinkForms] = useState<
    Record<string, ForwarderPackageLinkForm>
  >({})
  const [packageLinkErrors, setPackageLinkErrors] = useState<
    Record<string, string | null>
  >({})
  const [packageLinkResults, setPackageLinkResults] = useState<
    Record<string, string | null>
  >({})
  const [packageSuggestions, setPackageSuggestions] = useState<
    Record<string, ForwarderPackageMatchSuggestion[]>
  >({})
  const [packageSuggestionError, setPackageSuggestionError] = useState<
    string | null
  >(null)
  const [packageSuggestionResult, setPackageSuggestionResult] = useState<
    string | null
  >(null)
  const [addDialogOpen, setAddDialogOpen] = useState(false)
  const [capturedReviewsOpen, setCapturedReviewsOpen] = useState(false)
  const [sourceMatchesOpen, setSourceMatchesOpen] = useState(false)
  const [addDialogTab, setAddDialogTab] = useState<'new' | 'csv' | 'email'>(
    'new'
  )
  const [manualPurchaseForm, setManualPurchaseForm] =
    useState<ManualPurchaseForm>(defaultManualPurchaseForm)
  const [manualPurchaseDrafts, setManualPurchaseDrafts] = useState<
    ManualPurchaseDraft[]
  >([])
  const [importPurchaseDrafts, setImportPurchaseDrafts] = useState<
    PurchaseImportDraft[]
  >([])
  const [purchaseCSVImport, setPurchaseCSVImport] = useState(
    defaultPurchaseCSVImport
  )
  const [purchaseEmailImport, setPurchaseEmailImport] = useState(
    defaultPurchaseEmailImport
  )
  const [purchaseCSVPreview, setPurchaseCSVPreview] = useState<
    PurchaseImportDraft[]
  >([])
  const [purchaseEmailPreview, setPurchaseEmailPreview] = useState<
    PurchaseImportDraft[]
  >([])
  const [purchaseImportError, setPurchaseImportError] = useState<string | null>(
    null
  )
  const [manualPurchaseError, setManualPurchaseError] = useState<string | null>(
    null
  )
  const [manualPurchaseResult, setManualPurchaseResult] = useState<
    string | null
  >(null)
  const [purchaseSearch, setPurchaseSearch] = useState('')
  const [purchaseStatusFilter, setPurchaseStatusFilter] =
    useState<PurchaseTableStatusFilter>('all')
  const [favoritePurchaseKeys, setFavoritePurchaseKeys] = useState<
    Record<string, boolean>
  >({})
  const [arrivedPurchaseKeys, setArrivedPurchaseKeys] = useState<
    Record<string, boolean>
  >({})
  const [purchaseRatings, setPurchaseRatings] = useState<
    Record<string, number>
  >({})
  const [packageSuggestionSummary, setPackageSuggestionSummary] =
    useState<ForwarderPackageMatchSuggestionSummary | null>(null)
  const [
    packageSuggestionConfidenceFilter,
    setPackageSuggestionConfidenceFilter,
  ] = useState<ForwarderPackageSuggestionConfidenceFilter>('all')
  const [packageSuggestionActiveFilter, setPackageSuggestionActiveFilter] =
    useState<string | null>(null)
  const [packageReviewFilter, setPackageReviewFilter] =
    useState<ForwarderPackageReviewFilter>('all')
  const [selectedPackageID, setSelectedPackageID] = useState<string | null>(
    null
  )
  const [packageForm, setPackageForm] =
    useState<PackageImportForm>(defaultPackageImport)
  const [confirmedAction, setConfirmedAction] = useState<string | null>(null)
  const [pendingAction, setPendingAction] =
    useState<PurchaseInboxAction | null>(null)

  const readyItemCount = useMemo(
    () =>
      reviews.reduce(
        (count, review) =>
          count +
          review.items.filter(
            (item) => item.status === 'ready_to_link_or_convert'
          ).length,
        0
      ),
    [reviews]
  )

  const packageReviewSummary = useMemo(() => {
    const linkedPackageIDs = new Set<string>()
    let auditEventCount = 0
    let suggestionCount = 0

    Object.entries(packageLinks).forEach(([packageID, links]) => {
      if ((links ?? []).length > 0) {
        linkedPackageIDs.add(packageID)
      }
    })
    Object.values(packageLinkEvents).forEach((events) => {
      auditEventCount += (events ?? []).length
    })
    Object.values(packageSuggestions).forEach((suggestions) => {
      suggestionCount += (suggestions ?? []).length
    })

    return {
      packageCount: packages.length,
      linkedCount: linkedPackageIDs.size,
      unlinkedCount: Math.max(packages.length - linkedPackageIDs.size, 0),
      auditEventCount,
      suggestionCount,
    }
  }, [packageLinkEvents, packageLinks, packageSuggestions, packages.length])

  const filteredPackages = useMemo(
    () =>
      packages.filter((pkg) => {
        const hasLink = (packageLinks[pkg.id] ?? []).length > 0
        const hasSuggestion = (packageSuggestions[pkg.id] ?? []).length > 0

        if (packageReviewFilter === 'linked') {
          return hasLink
        }
        if (packageReviewFilter === 'unlinked') {
          return !hasLink
        }
        if (packageReviewFilter === 'suggested') {
          return hasSuggestion
        }
        return true
      }),
    [packageLinks, packageReviewFilter, packageSuggestions, packages]
  )

  const purchaseRows = useMemo<PurchaseTableRow[]>(() => {
    const capturedRows = reviews.flatMap((review) =>
      review.items.map((item) => {
        const source = item.item.seller_username
          ? 'eBay / ' + item.item.seller_username
          : 'eBay'
        const title =
          item.item.purchased_identity ??
          item.item.listing_title ??
          'Untitled purchase item'
        const price = item.item.item_price ?? review.order.order_total ?? '-'
        const status = item.status
        const tracking = 'Pending'
        const purchaseDate = item.item.purchase_date ?? 'Pending'
        const delivery = 'Pending'
        const orderLink = item.item.item_url ?? ''
        const key = purchaseRowKey(review, item)

        return {
          key,
          title,
          source,
          price,
          purchaseDate,
          delivery,
          status,
          tracking,
          orderLink,
          persistence: 'Captured review only',
          actionCount: (item.suggested_actions ?? []).length,
          searchText: [
            title,
            source,
            price,
            purchaseDate,
            delivery,
            labelForStatus(status),
            tracking,
            orderLink,
            review.order.order_id,
          ]
            .filter(Boolean)
            .join(' ')
            .toLowerCase(),
        }
      })
    )
    const manualRows = manualPurchaseDrafts.map((draft) => ({
      key: draft.key,
      title: draft.title,
      source: draft.source,
      price: draft.price || '-',
      purchaseDate: 'Pending',
      delivery: 'Pending',
      status: 'manual_draft',
      tracking: draft.tracking || 'Pending',
      orderLink: '',
      persistence: persistenceLabel(draft.persistence),
      actionCount: 1,
      searchText: [
        draft.title,
        draft.source,
        draft.price,
        'date pending',
        'delivery pending',
        'manual draft',
        draft.tracking,
        persistenceLabel(draft.persistence),
      ]
        .filter(Boolean)
        .join(' ')
        .toLowerCase(),
    }))
    const importedRows = importPurchaseDrafts.map((draft) => ({
      key: draft.key,
      title: draft.title,
      source: draft.provenance,
      price: draft.price || '-',
      purchaseDate: draft.purchaseDate || 'Pending',
      delivery: draft.delivery || 'Pending',
      status: draft.mode === 'csv' ? 'csv_import' : 'email_import',
      tracking: draft.tracking || draft.delivery || 'Pending',
      orderLink: '',
      persistence: persistenceLabel(draft.persistence),
      actionCount: 1,
      searchText: [
        draft.title,
        draft.provenance,
        draft.price,
        draft.currency,
        draft.purchaseDate,
        draft.delivery,
        draft.seller,
        draft.channel,
        draft.tracking,
        draft.delivery,
        draft.mode + ' import',
        persistenceLabel(draft.persistence),
      ]
        .filter(Boolean)
        .join(' ')
        .toLowerCase(),
    }))

    return [...importedRows, ...manualRows, ...capturedRows]
  }, [importPurchaseDrafts, manualPurchaseDrafts, reviews])

  const filteredPurchaseRows = useMemo(() => {
    const query = purchaseSearch.trim().toLowerCase()

    return purchaseRows.filter((row) => {
      if (
        purchaseStatusFilter !== 'all' &&
        row.status !== purchaseStatusFilter
      ) {
        return false
      }
      if (query && !row.searchText.includes(query)) {
        return false
      }
      return true
    })
  }, [purchaseRows, purchaseSearch, purchaseStatusFilter])

  const loadReviews = useCallback(async () => {
    setLoading(true)
    setError(null)
    setConfirmedAction(null)
    try {
      const response = await fetch(
        '/api/integrations/ebay/purchase-inbox/reviews',
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ cards: sampleCards }),
        }
      )
      if (!response.ok) {
        throw new Error('purchase_inbox_reviews_' + response.status)
      }
      const payload = (await response.json()) as ReviewResponse
      setReviews(payload.reviews ?? [])
    } catch (err) {
      setReviews([])
      setError(
        err instanceof Error ? err.message : 'purchase_inbox_reviews_failed'
      )
    } finally {
      setLoading(false)
    }
  }, [])

  const loadPackages = useCallback(
    async (profileId = packageForm.profile_id) => {
      setPackagesLoading(true)
      setPackageError(null)
      try {
        const params = new URLSearchParams()
        if (profileId.trim()) {
          params.set('profile_id', profileId.trim())
        }
        const response = await fetch(
          '/api/forwarding/packages?' + params.toString()
        )
        if (!response.ok) {
          throw new Error('forwarder_packages_' + response.status)
        }
        const payload = (await response.json()) as ForwarderPackageListResponse
        setPackages(payload.packages ?? [])
      } catch (err) {
        setPackages([])
        setPackageError(
          err instanceof Error ? err.message : 'forwarder_packages_failed'
        )
      } finally {
        setPackagesLoading(false)
      }
    },
    [packageForm.profile_id]
  )

  const loadPackageLinks = useCallback(async (packageID: string) => {
    try {
      setPackageLinkErrors((current) => ({ ...current, [packageID]: null }))
      const params = new URLSearchParams({ package_id: packageID })
      const response = await fetch(
        '/api/forwarding/package-links?' + params.toString()
      )
      if (!response.ok) {
        throw new Error('forwarder_package_links_' + response.status)
      }
      const payload = (await response.json()) as {
        links?: ForwarderPackageLink[]
        events?: ForwarderPackageLinkEvent[]
      }
      setPackageLinks((current) => ({
        ...current,
        [packageID]: payload.links ?? [],
      }))
      setPackageLinkEvents((current) => ({
        ...current,
        [packageID]: payload.events ?? [],
      }))
    } catch (err) {
      setPackageLinks((current) => ({ ...current, [packageID]: [] }))
      setPackageLinkEvents((current) => ({ ...current, [packageID]: [] }))
      setPackageLinkErrors((current) => ({
        ...current,
        [packageID]:
          err instanceof Error ? err.message : 'forwarder_package_links_failed',
      }))
    }
  }, [])

  const loadPackageSuggestions = useCallback(
    async (packageID?: string) => {
      setPackagesLoading(true)
      setPackageError(null)
      setPackageSuggestionError(null)
      setPackageSuggestionResult(null)
      setPackageSuggestionSummary(null)
      try {
        const params = new URLSearchParams()
        if (packageForm.profile_id.trim()) {
          params.set('profile_id', packageForm.profile_id.trim())
        }
        if (packageSuggestionConfidenceFilter !== 'all') {
          params.set('confidence_label', packageSuggestionConfidenceFilter)
        }
        if (packageID?.trim()) {
          params.set('package_id', packageID.trim())
        }
        const response = await fetch(
          '/api/forwarding/package-match-suggestions?' + params.toString()
        )
        if (!response.ok) {
          throw new Error(
            'forwarder_package_match_suggestions_' + response.status
          )
        }
        const payload =
          (await response.json()) as ForwarderPackageMatchSuggestionResponse
        const grouped = (payload.suggestions ?? []).reduce<
          Record<string, ForwarderPackageMatchSuggestion[]>
        >((current, suggestion) => {
          current[suggestion.package_id] = [
            ...(current[suggestion.package_id] ?? []),
            suggestion,
          ]
          return current
        }, {})
        setPackageSuggestions(grouped)
        setPackageSuggestionSummary(payload.summary ?? null)
        setPackageSuggestionActiveFilter(payload.confidence_filter ?? null)
        const count = payload.summary?.count ?? payload.suggestions?.length ?? 0
        if (count === 0) {
          setPackageSuggestionResult(
            'No package match suggestions matched' +
              (packageID ? ' package ' + packageID : ' the current inbox') +
              (payload.confidence_filter
                ? ' for ' + payload.confidence_filter + ' confidence'
                : '') +
              '. Package rows were left unchanged.'
          )
          return
        }
        setPackageSuggestionResult(
          'Found ' +
            count +
            ' package match suggestion' +
            (count === 1 ? '' : 's') +
            (packageID ? ' for package ' + packageID : '') +
            (payload.confidence_filter
              ? ' for ' + payload.confidence_filter + ' confidence'
              : '')
        )
      } catch (err) {
        setPackageSuggestions({})
        setPackageSuggestionSummary(null)
        setPackageSuggestionActiveFilter(null)
        setPackageSuggestionError(
          err instanceof Error
            ? err.message
            : 'forwarder_package_match_suggestions_failed'
        )
      } finally {
        setPackagesLoading(false)
      }
    },
    [packageForm.profile_id, packageSuggestionConfidenceFilter]
  )

  const updatePackageForm = (field: keyof PackageImportForm, value: string) => {
    setPackageForm((current) => ({ ...current, [field]: value }))
  }

  const updateManualPurchaseForm = (
    field: keyof ManualPurchaseForm,
    value: string
  ) => {
    setManualPurchaseForm((current) => ({ ...current, [field]: value }))
    setManualPurchaseError(null)
    setManualPurchaseResult(null)
  }

  const persistPurchaseDraft = async (draft: {
    key: string
    title: string
    source: string
    price: string
    currency?: string
    tracking?: string
    provenance?: string
  }) => {
    const itemResponse = await fetch('/api/items', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        part_number: purchasePartNumber(draft.title, draft.key),
        title: draft.title,
        brand: draft.source || draft.provenance || 'Purchase',
        category: 'Purchases',
        notes:
          'Created from Purchases add/import flow. Source: ' +
          (draft.provenance || draft.source || 'manual') +
          (draft.tracking ? '. Tracking: ' + draft.tracking : ''),
      }),
    })
    if (!itemResponse.ok) {
      throw new Error('purchase_item_create_' + itemResponse.status)
    }
    const item = (await itemResponse.json()) as CreatedItemResponse
    if (!item.id) {
      throw new Error('purchase_item_missing_id')
    }

    const lifecycleResponse = await fetch('/api/commerce/lifecycle', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        item_id: item.id,
        state: 'purchase',
        source: draft.provenance || draft.source || 'manual',
        external_ref: draft.tracking || draft.key,
        quantity: 1,
        amount: parsePurchaseAmount(draft.price),
        currency: draft.currency || 'AUD',
        notes: 'Purchases UI persisted draft ' + draft.key,
      }),
    })
    if (!lifecycleResponse.ok) {
      throw new Error('purchase_lifecycle_create_' + lifecycleResponse.status)
    }
    const lifecycle =
      (await lifecycleResponse.json()) as CommerceLifecycleResponse
    const lifecycleEntryID = lifecycle.entry?.id
    const expectedArrivalID =
      lifecycle.expected_arrival?.id ?? lifecycle.entry?.expected_arrival_id
    if (!lifecycleEntryID || !expectedArrivalID) {
      throw new Error('purchase_lifecycle_missing_persistence_ids')
    }
    return {
      itemID: item.id,
      lifecycleEntryID,
      expectedArrivalID,
    }
  }

  const previewPurchaseCSVImport = () => {
    const preview = parsePurchaseCSVImport(purchaseCSVImport)
    setPurchaseCSVPreview(preview)
    setPurchaseImportError(
      preview.length === 0
        ? 'CSV import needs a header row and at least one purchase title.'
        : null
    )
    setManualPurchaseResult(null)
  }

  const previewPurchaseEmailImport = () => {
    const preview = parsePurchaseEmailImport(purchaseEmailImport)
    setPurchaseEmailPreview(preview)
    setPurchaseImportError(
      preview.length === 0
        ? 'Email import needs labeled purchase text with a Title field.'
        : null
    )
    setManualPurchaseResult(null)
  }

  const confirmPurchaseImportDrafts = async (mode: 'csv' | 'email') => {
    const preview = mode === 'csv' ? purchaseCSVPreview : purchaseEmailPreview
    if (preview.length === 0) {
      setPurchaseImportError(
        mode === 'csv'
          ? 'Preview CSV purchases before confirming.'
          : 'Preview email purchases before confirming.'
      )
      return
    }
    try {
      const persistedPreview = await Promise.all(
        preview.map(async (draft) => ({
          ...draft,
          persistence: await persistPurchaseDraft({
            key: draft.key,
            title: draft.title,
            source: draft.channel,
            price: draft.price,
            currency: draft.currency,
            tracking: draft.tracking || draft.delivery,
            provenance: draft.provenance,
          }),
        }))
      )
      setImportPurchaseDrafts((current) => [...persistedPreview, ...current])
      setManualPurchaseResult(
        'Confirmed and persisted ' +
          preview.length +
          ' ' +
          mode.toUpperCase() +
          ' import draft' +
          (preview.length === 1 ? '' : 's') +
          ' through the commerce lifecycle API.'
      )
    } catch (err) {
      setPurchaseImportError(
        err instanceof Error ? err.message : 'purchase_import_persist_failed'
      )
      setManualPurchaseResult(null)
      return
    }
    setPurchaseImportError(null)
    if (mode === 'csv') {
      setPurchaseCSVPreview([])
    } else {
      setPurchaseEmailPreview([])
    }
    setPurchaseStatusFilter('all')
    setAddDialogOpen(false)
  }

  const saveManualPurchaseDraft = async () => {
    const title = manualPurchaseForm.title.trim()
    const source = manualPurchaseForm.source.trim() || 'manual'
    const price = manualPurchaseForm.price.trim()
    const tracking = manualPurchaseForm.tracking.trim()

    if (!title) {
      setManualPurchaseError('Purchase title is required.')
      setManualPurchaseResult(null)
      return
    }

    const draft: ManualPurchaseDraft = {
      key:
        'manual-purchase-' +
        Date.now().toString(36) +
        '-' +
        title.toLowerCase().replace(/[^a-z0-9]+/g, '-'),
      title,
      source,
      price,
      tracking,
    }
    try {
      draft.persistence = await persistPurchaseDraft({
        key: draft.key,
        title,
        source,
        price,
        tracking,
      })
    } catch (err) {
      setManualPurchaseError(
        err instanceof Error ? err.message : 'purchase_draft_persist_failed'
      )
      setManualPurchaseResult(null)
      return
    }
    setManualPurchaseDrafts((current) => [draft, ...current])
    setManualPurchaseForm(defaultManualPurchaseForm)
    setManualPurchaseError(null)
    setManualPurchaseResult(
      'Persisted manual purchase draft for ' +
        title +
        ' through the commerce lifecycle API.'
    )
    setPurchaseStatusFilter('all')
    setAddDialogOpen(false)
  }

  const packageLinkFormFor = (packageID: string) =>
    packageLinkForms[packageID] ?? defaultPackageLinkForm

  const updatePackageLinkForm = (
    packageID: string,
    field: keyof ForwarderPackageLinkForm,
    value: string
  ) => {
    setPackageLinkForms((current) => ({
      ...current,
      [packageID]: { ...packageLinkFormFor(packageID), [field]: value },
    }))
  }

  const applyPackageSuggestion = (
    packageID: string,
    suggestion: ForwarderPackageMatchSuggestion
  ) => {
    setPackageLinkForms((current) => ({
      ...current,
      [packageID]: {
        item_id: suggestion.item_id,
        lifecycle_entry_id: suggestion.lifecycle_entry_id ?? '',
        expected_arrival_id: suggestion.expected_arrival_id ?? '',
        source: 'suggested_match',
        notes:
          (suggestion.confidence_label ?? 'suggested') +
          ' confidence package match suggestion',
        audit_trail: suggestion.audit_trail ?? [],
      },
    }))
    setPackageLinkResults((current) => ({
      ...current,
      [packageID]:
        'Prepared suggested match for item ' +
        suggestion.item_id +
        (suggestion.expected_arrival_id
          ? ' / arrival ' + suggestion.expected_arrival_id
          : ''),
    }))
  }

  const selectPackage = async (packageID: string, selected: boolean) => {
    if (selected) {
      setSelectedPackageID(null)
      return
    }
    setSelectedPackageID(packageID)
    if (!packageLinkForms[packageID]) {
      setPackageLinkForms((current) => ({
        ...current,
        [packageID]: defaultPackageLinkForm,
      }))
    }
    await loadPackageLinks(packageID)
  }

  const linkPackage = async (
    pkg: ForwarderPackage,
    decision: 'confirmed' | 'overridden' = 'confirmed'
  ) => {
    const form = packageLinkFormFor(pkg.id)
    setPackageLinkErrors((current) => ({ ...current, [pkg.id]: null }))
    setPackageLinkResults((current) => ({ ...current, [pkg.id]: null }))
    const override = decision === 'overridden'
    const source = override ? 'manual_override' : form.source
    const auditTrail = [
      decision +
        ' from purchase inbox UI: ' +
        form.item_id +
        ' / ' +
        form.expected_arrival_id,
      ...(source === 'suggested_match' ? (form.audit_trail ?? []) : []),
    ]
    try {
      const response = await fetch('/api/forwarding/package-links', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          package_id: pkg.id,
          item_id: form.item_id,
          lifecycle_entry_id: form.lifecycle_entry_id,
          expected_arrival_id: form.expected_arrival_id,
          source,
          decision,
          notes: form.notes,
          override,
          actor: 'reviewer',
          audit_trail: auditTrail,
        }),
      })
      const payload = await response.json().catch(() => null)
      if (!response.ok) {
        const message =
          payload && typeof payload === 'object' && 'message' in payload
            ? String(payload.message)
            : 'forwarder_package_link_' + response.status
        throw new Error(message)
      }
      const result = payload as { link?: ForwarderPackageLink }
      const link = result.link
      setPackageLinkResults((current) => ({
        ...current,
        [pkg.id]: link
          ? (decision === 'overridden' ? 'Override linked' : 'Confirmed link') +
            ' to item ' +
            link.item_id
          : decision === 'overridden'
            ? 'Override linked package'
            : 'Confirmed package link',
      }))
      await loadPackageLinks(pkg.id)
    } catch (err) {
      setPackageLinkErrors((current) => ({
        ...current,
        [pkg.id]:
          err instanceof Error ? err.message : 'forwarder_package_link_failed',
      }))
    }
  }

  const unlinkPackage = async (pkg: ForwarderPackage) => {
    const form = packageLinkFormFor(pkg.id)
    setPackageLinkErrors((current) => ({ ...current, [pkg.id]: null }))
    setPackageLinkResults((current) => ({ ...current, [pkg.id]: null }))
    try {
      const params = new URLSearchParams({ package_id: pkg.id })
      const response = await fetch(
        '/api/forwarding/package-links?' + params.toString(),
        {
          method: 'DELETE',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            source: 'manual_unlink',
            actor: 'reviewer',
            notes: form.notes || 'Unlinked from purchase inbox review',
            audit_trail: [
              'unlinked from purchase inbox UI: ' +
                form.item_id +
                ' / ' +
                form.expected_arrival_id,
            ],
          }),
        }
      )
      const payload = await response.json().catch(() => null)
      if (!response.ok) {
        const message =
          payload && typeof payload === 'object' && 'message' in payload
            ? String(payload.message)
            : 'forwarder_package_unlink_' + response.status
        throw new Error(message)
      }
      setPackageLinkResults((current) => ({
        ...current,
        [pkg.id]: 'Unlinked package from reconciliation target',
      }))
      await loadPackageLinks(pkg.id)
    } catch (err) {
      setPackageLinkErrors((current) => ({
        ...current,
        [pkg.id]:
          err instanceof Error
            ? err.message
            : 'forwarder_package_unlink_failed',
      }))
    }
  }

  const importPackage = async () => {
    setPackagesLoading(true)
    setPackageError(null)
    setPackageResult(null)
    try {
      const response = await fetch('/api/forwarding/packages', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          profile_id: packageForm.profile_id,
          provider: packageForm.provider,
          source: packageForm.source,
          external_package_id: packageForm.external_package_id,
          status: packageForm.status,
          shipment_id: packageForm.shipment_id,
          tracking_number: packageForm.tracking_number,
          sender: packageForm.sender,
          warehouse_location: packageForm.warehouse_location,
          weight_grams: Number(packageForm.weight_grams || 0),
          raw_payload: {
            manual_import: true,
            ui_surface: 'purchase_inbox_forwarder_packages',
          },
        }),
      })
      if (!response.ok) {
        const payload = await response.json().catch(() => null)
        const message =
          payload && typeof payload === 'object' && 'message' in payload
            ? String(payload.message)
            : 'forwarder_package_import_' + response.status
        throw new Error(message)
      }
      const payload = (await response.json()) as { package?: ForwarderPackage }
      const saved = payload.package
      setPackageResult(
        saved
          ? 'Imported package ' + saved.external_package_id
          : 'Imported package'
      )
      await loadPackages(packageForm.profile_id)
    } catch (err) {
      setPackageError(
        err instanceof Error ? err.message : 'forwarder_package_import_failed'
      )
    } finally {
      setPackagesLoading(false)
    }
  }

  const importPackageCSV = async () => {
    setPackagesLoading(true)
    setPackageError(null)
    setPackageResult(null)
    setPackageCSVErrors([])
    try {
      const response = await fetch('/api/forwarding/packages/import-csv', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          profile_id: packageForm.profile_id,
          provider: packageForm.provider,
          csv: packageCSV,
        }),
      })
      const payload = await response.json().catch(() => null)
      if (!response.ok) {
        const message =
          payload && typeof payload === 'object' && 'message' in payload
            ? String(payload.message)
            : 'forwarder_package_csv_import_' + response.status
        throw new Error(message)
      }
      const result = payload as ForwarderPackageCSVImportResponse
      const imported = result.summary?.imported ?? result.imported?.length ?? 0
      const errors = result.errors ?? []
      setPackageCSVErrors(errors)
      setPackageResult(
        'Imported ' +
          imported +
          ' CSV package' +
          (imported === 1 ? '' : 's') +
          (errors.length > 0
            ? '; ' +
              errors.length +
              ' row' +
              (errors.length === 1 ? '' : 's') +
              ' needs attention'
            : '')
      )
      await loadPackages(packageForm.profile_id)
    } catch (err) {
      setPackageError(
        err instanceof Error ? err.message : 'forwarder_package_csv_failed'
      )
    } finally {
      setPackagesLoading(false)
    }
  }

  const importPackageEmail = async () => {
    setPackagesLoading(true)
    setPackageError(null)
    setPackageResult(null)
    setPackageCSVErrors([])
    try {
      const response = await fetch('/api/forwarding/packages/import-email', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          profile_id: packageForm.profile_id,
          provider: packageForm.provider,
          message_id: 'manual-email-notice',
          body: packageEmail,
        }),
      })
      const payload = await response.json().catch(() => null)
      if (!response.ok) {
        const message =
          payload && typeof payload === 'object' && 'message' in payload
            ? String(payload.message)
            : 'forwarder_package_email_import_' + response.status
        throw new Error(message)
      }
      const result = payload as { package?: ForwarderPackage }
      setPackageResult(
        'Imported email package ' +
          (result.package?.external_package_id ?? 'notice')
      )
      await loadPackages(packageForm.profile_id)
    } catch (err) {
      setPackageError(
        err instanceof Error ? err.message : 'forwarder_package_email_failed'
      )
    } finally {
      setPackagesLoading(false)
    }
  }

  const requestAction = (action: PurchaseInboxAction) => {
    if (!action.requires_confirmation) {
      setConfirmedAction(action.label + ': ' + action.target_key)
      return
    }
    setPendingAction(action)
  }

  const confirmPendingAction = () => {
    if (!pendingAction) {
      return
    }
    setConfirmedAction(pendingAction.label + ': ' + pendingAction.target_key)
    setPendingAction(null)
  }

  const openCapturedReviews = () => {
    setCapturedReviewsOpen(true)
    void loadReviews()
  }

  return (
    <>
      <Header>
        <Search />
        <HeaderTitle
          title='Purchases'
          description='Review, import, and reconcile purchases across channels.'
          icon={ShoppingCart}
          testId='purchases-header-title'
          iconTestId='purchases-page-icon'
        />
        <div
          className='ms-auto flex items-center space-x-4'
          data-header-title-avoid='true'
        >
          <LanguageSwitch />
          <ThemeSwitch />
          <ConfigDrawer />
          <ProfileDropdown />
        </div>
      </Header>

      <Main>
        <div className='flex flex-wrap items-center justify-between gap-3'>
          <div>
            <h1 className='text-2xl font-bold tracking-tight'>Purchases</h1>
            <p className='text-muted-foreground'>
              Track purchases from eBay, Amazon, CSV, email, desktop imports,
              and manual entries in one review-first workspace.
            </p>
          </div>
          <div className='flex flex-wrap items-center gap-2'>
            <Button
              data-testid='purchases-add-button'
              onClick={() => setAddDialogOpen(true)}
              aria-label='Add purchase'
            >
              <Plus className='mr-2 h-4 w-4' />
              Add
            </Button>
            <Button
              variant='outline'
              data-testid='purchases-source-matches-toggle'
              onClick={() => setSourceMatchesOpen((current) => !current)}
              aria-expanded={sourceMatchesOpen}
            >
              <Truck className='mr-2 h-4 w-4' />
              Source matches
            </Button>
            <Button
              variant='outline'
              data-testid='purchase-inbox-load-reviews'
              onClick={openCapturedReviews}
              disabled={loading}
              aria-expanded={capturedReviewsOpen}
            >
              {loading ? (
                <Loader2 className='mr-2 h-4 w-4 animate-spin' />
              ) : (
                <RefreshCw className='mr-2 h-4 w-4' />
              )}
              Load captured reviews
            </Button>
          </div>
        </div>
        {manualPurchaseResult ? (
          <div
            className='mt-3 rounded-md border bg-muted/30 p-3 text-sm'
            data-testid='purchases-manual-draft-result'
          >
            {manualPurchaseResult}
          </div>
        ) : null}

        <div
          className='mt-4 overflow-x-auto rounded-md border'
          data-testid='purchases-table-shell'
        >
          <div
            className='flex flex-wrap items-center gap-2 border-b p-3'
            data-testid='purchases-table-filters'
          >
            <Input
              className='h-9 w-full sm:w-72'
              data-testid='purchases-table-search'
              placeholder='Search purchases, sources, statuses'
              value={purchaseSearch}
              onChange={(event) => setPurchaseSearch(event.target.value)}
            />
            {(
              [
                ['all', 'All'],
                ['ready_to_link_or_convert', 'Ready'],
                ['needs_review', 'Needs review'],
                ['manual_draft', 'Manual draft'],
                ['csv_import', 'CSV import'],
                ['email_import', 'Email import'],
              ] satisfies Array<[PurchaseTableStatusFilter, string]>
            ).map(([value, label]) => (
              <Button
                key={value}
                type='button'
                size='sm'
                variant={purchaseStatusFilter === value ? 'default' : 'outline'}
                data-testid={'purchases-status-filter-' + value}
                onClick={() => setPurchaseStatusFilter(value)}
              >
                {label}
              </Button>
            ))}
            <span
              className='text-sm text-muted-foreground'
              data-testid='purchases-filter-result'
            >
              Showing {filteredPurchaseRows.length} of {purchaseRows.length}{' '}
              purchases
            </span>
          </div>
          <Table className='min-w-[88rem] table-fixed'>
            <TableHeader>
              <TableRow>
                <TableHead className='w-[18rem]'>Purchase</TableHead>
                <TableHead className='w-[12rem]'>Source</TableHead>
                <TableHead className='w-[10rem]'>Price</TableHead>
                <TableHead className='w-[9rem]'>Purchase date</TableHead>
                <TableHead className='w-[12rem]'>Delivery</TableHead>
                <TableHead className='w-[12rem]'>Status</TableHead>
                <TableHead className='w-[14rem]'>Tracking</TableHead>
                <TableHead className='w-[9rem]'>Order link</TableHead>
                <TableHead className='w-[10rem] text-right'>Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredPurchaseRows.map((row) => (
                <TableRow key={row.key} data-testid='purchases-table-row'>
                  <TableCell className='truncate font-medium'>
                    <div className='min-w-0'>
                      <p className='truncate'>{row.title}</p>
                      <p
                        className='truncate text-xs font-normal text-muted-foreground'
                        data-testid='purchases-row-persistence'
                      >
                        {row.persistence}
                      </p>
                    </div>
                  </TableCell>
                  <TableCell className='truncate'>{row.source}</TableCell>
                  <TableCell>{row.price}</TableCell>
                  <TableCell data-testid='purchases-row-purchase-date'>
                    {row.purchaseDate}
                  </TableCell>
                  <TableCell
                    className='text-muted-foreground'
                    data-testid='purchases-row-delivery'
                  >
                    {row.delivery}
                  </TableCell>
                  <TableCell>{labelForStatus(row.status)}</TableCell>
                  <TableCell className='text-muted-foreground'>
                    {arrivedPurchaseKeys[row.key] ? 'Arrived' : row.tracking}
                  </TableCell>
                  <TableCell>
                    {row.orderLink ? (
                      <a
                        className='text-sm font-medium text-primary underline-offset-4 hover:underline'
                        data-testid='purchases-row-order-link'
                        href={row.orderLink}
                        target='_blank'
                        rel='noreferrer'
                      >
                        Open order
                      </a>
                    ) : (
                      <span
                        className='text-sm text-muted-foreground'
                        data-testid='purchases-row-order-link-empty'
                      >
                        Pending
                      </span>
                    )}
                  </TableCell>
                  <TableCell>
                    <div className='flex flex-wrap justify-end gap-1'>
                      <Button
                        type='button'
                        size='sm'
                        variant={
                          favoritePurchaseKeys[row.key] ? 'default' : 'outline'
                        }
                        aria-pressed={favoritePurchaseKeys[row.key] ?? false}
                        data-testid='purchases-row-favorite'
                        onClick={() =>
                          setFavoritePurchaseKeys((current) => ({
                            ...current,
                            [row.key]: !current[row.key],
                          }))
                        }
                      >
                        <Star className='mr-1 h-3.5 w-3.5' />
                        Favorite
                      </Button>
                      <Button
                        type='button'
                        size='sm'
                        variant={
                          arrivedPurchaseKeys[row.key] ? 'default' : 'outline'
                        }
                        aria-pressed={arrivedPurchaseKeys[row.key] ?? false}
                        data-testid='purchases-row-arrived'
                        onClick={() =>
                          setArrivedPurchaseKeys((current) => ({
                            ...current,
                            [row.key]: !current[row.key],
                          }))
                        }
                      >
                        <CheckCircle2 className='mr-1 h-3.5 w-3.5' />
                        Arrived
                      </Button>
                      <Button
                        type='button'
                        size='sm'
                        variant={
                          purchaseRatings[row.key] === 4 ? 'default' : 'outline'
                        }
                        data-testid='purchases-row-rating'
                        onClick={() =>
                          setPurchaseRatings((current) => ({
                            ...current,
                            [row.key]: current[row.key] === 4 ? 0 : 4,
                          }))
                        }
                      >
                        Rating {purchaseRatings[row.key] || '-'}
                      </Button>
                      <span
                        className='self-center text-xs text-muted-foreground'
                        data-testid='purchases-row-action-count'
                      >
                        {row.actionCount} action
                        {row.actionCount === 1 ? '' : 's'}
                      </span>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
              {purchaseRows.length === 0 ? (
                <TableRow data-testid='purchases-table-empty-row'>
                  <TableCell
                    colSpan={9}
                    className='h-20 text-center text-sm text-muted-foreground'
                  >
                    No purchases loaded. Add a purchase or review captured
                    purchases to populate the table.
                  </TableCell>
                </TableRow>
              ) : null}
              {purchaseRows.length > 0 && filteredPurchaseRows.length === 0 ? (
                <TableRow data-testid='purchases-table-filter-empty-row'>
                  <TableCell
                    colSpan={9}
                    className='h-20 text-center text-sm text-muted-foreground'
                  >
                    No purchases match the current table filters.
                  </TableCell>
                </TableRow>
              ) : null}
            </TableBody>
          </Table>
        </div>

        {capturedReviewsOpen ? (
          <section
            className='mt-4 space-y-4'
            data-testid='purchase-review-tools'
          >
            <div className='flex flex-wrap items-center justify-between gap-3 rounded-md border bg-muted/20 p-4'>
              <div>
                <h2 className='text-lg font-semibold tracking-tight'>
                  Captured purchase reviews
                </h2>
                <p className='text-sm text-muted-foreground'>
                  Review captured eBay order records only when you need to
                  reconcile source captures into the Purchases table.
                </p>
              </div>
              <Button
                type='button'
                variant='ghost'
                size='sm'
                data-testid='purchase-review-tools-hide'
                onClick={() => setCapturedReviewsOpen(false)}
              >
                Hide
              </Button>
            </div>

            {error ? (
              <section
                className='rounded-md border border-destructive/40 bg-destructive/10 p-4 text-sm'
                data-testid='purchase-inbox-error-state'
              >
                <p className='font-medium'>Purchases could not load reviews.</p>
                <p className='mt-1 text-muted-foreground'>{error}</p>
              </section>
            ) : null}

            {loading ? (
              <section
                className='rounded-md border border-dashed p-6 text-sm text-muted-foreground'
                data-testid='purchase-inbox-loading-state'
              >
                Preparing purchase review records...
              </section>
            ) : null}

            {!loading && !error && reviews.length === 0 ? (
              <section
                className='rounded-md border border-dashed p-6'
                data-testid='purchase-inbox-empty-state'
              >
                <p className='font-medium'>
                  No captured purchase reviews are loaded.
                </p>
                <p className='mt-1 text-sm text-muted-foreground'>
                  Captured purchase review is a secondary reconciliation flow.
                  The Purchases table and Add dialog remain the primary
                  workflow.
                </p>
              </section>
            ) : null}

            {reviews.length > 0 ? (
              <div
                className='space-y-4'
                data-testid='purchase-inbox-ready-state'
              >
                <div className='grid gap-3 md:grid-cols-3'>
                  <div className='rounded-md border p-3'>
                    <p className='text-sm text-muted-foreground'>Orders</p>
                    <p className='text-2xl font-semibold'>{reviews.length}</p>
                  </div>
                  <div className='rounded-md border p-3'>
                    <p className='text-sm text-muted-foreground'>Ready items</p>
                    <p className='text-2xl font-semibold'>{readyItemCount}</p>
                  </div>
                  <div className='rounded-md border p-3'>
                    <p className='text-sm text-muted-foreground'>
                      Mutation policy
                    </p>
                    <p className='flex items-center gap-2 text-sm font-medium'>
                      <ShieldCheck className='h-4 w-4' />
                      Confirmation required
                    </p>
                  </div>
                </div>

                {confirmedAction ? (
                  <p
                    className='rounded-md border bg-muted/30 p-3 text-sm'
                    data-testid='purchase-inbox-action-result'
                  >
                    Queued after confirmation: {confirmedAction}
                  </p>
                ) : null}

                {reviews.map((review) => (
                  <article
                    key={review.order.order_id ?? 'orderless'}
                    className='rounded-md border p-4'
                    data-testid='purchase-inbox-order-card'
                  >
                    <div className='flex flex-wrap items-start justify-between gap-3'>
                      <div>
                        <h2 className='font-semibold'>
                          Order {review.order.order_id ?? 'without order id'}
                        </h2>
                        <p className='text-sm text-muted-foreground'>
                          {(review.order.seller_usernames ?? []).join(', ') ||
                            'Unknown seller'}{' '}
                          · {review.order.order_total ?? 'Total pending'}{' '}
                          {review.order.currency ?? ''}
                        </p>
                      </div>
                      <span className='rounded-md border px-2 py-1 text-xs font-medium'>
                        {labelForStatus(review.status)}
                      </span>
                    </div>

                    <div className='mt-4 space-y-3'>
                      {review.items.map((item) => (
                        <div
                          key={
                            item.item.transaction_id ??
                            item.item.listing_id ??
                            item.item.listing_title
                          }
                          className='rounded-md border bg-muted/20 p-3'
                          data-testid='purchase-inbox-item-row'
                        >
                          <div className='flex flex-wrap items-start justify-between gap-3'>
                            <div>
                              <p className='font-medium'>
                                {item.item.purchased_identity ??
                                  item.item.listing_title ??
                                  'Untitled purchase item'}
                              </p>
                              <p className='text-sm text-muted-foreground'>
                                Qty {item.item.quantity ?? 'pending'} ·{' '}
                                {item.item.item_price ?? 'price pending'}
                              </p>
                              {(item.missing_fields ?? []).length > 0 ? (
                                <p
                                  className='mt-1 text-xs text-amber-700 dark:text-amber-300'
                                  data-testid='purchase-inbox-missing-fields'
                                >
                                  Missing: {item.missing_fields?.join(', ')}
                                </p>
                              ) : null}
                            </div>
                            <span className='rounded-md border px-2 py-1 text-xs'>
                              {labelForStatus(item.status)}
                            </span>
                          </div>
                          <div className='mt-3 flex flex-wrap gap-2'>
                            {(item.suggested_actions ?? []).map((action) => (
                              <Button
                                key={action.id}
                                variant='outline'
                                size='sm'
                                className={actionTone(action)}
                                data-testid='purchase-inbox-suggested-action'
                                onClick={() => requestAction(action)}
                              >
                                {action.label}
                              </Button>
                            ))}
                          </div>
                        </div>
                      ))}
                    </div>
                  </article>
                ))}
              </div>
            ) : null}
          </section>
        ) : null}

        {sourceMatchesOpen ? (
          <>
            <Separator className='my-6' />

            <section
              className='space-y-4'
              data-testid='forwarder-package-inbox'
            >
              <div className='flex flex-wrap items-center justify-between gap-3'>
                <div>
                  <h2 className='flex items-center gap-2 text-xl font-semibold tracking-tight'>
                    <Truck className='h-5 w-5' />
                    Purchase Source Matches
                  </h2>
                  <p className='text-sm text-muted-foreground'>
                    Import Stackry or freight-forwarder evidence as
                    source-backed purchase candidates before confirming matches.
                  </p>
                </div>
                <Button
                  variant='outline'
                  data-testid='forwarder-package-refresh'
                  onClick={() => void loadPackages()}
                  disabled={packagesLoading}
                >
                  {packagesLoading ? (
                    <Loader2 className='mr-2 h-4 w-4 animate-spin' />
                  ) : (
                    <RefreshCw className='mr-2 h-4 w-4' />
                  )}
                  Refresh source records
                </Button>
                <Button
                  variant='secondary'
                  data-testid='forwarder-package-match-suggestions-load'
                  onClick={() => void loadPackageSuggestions()}
                  disabled={packagesLoading}
                >
                  {packagesLoading ? (
                    <Loader2 className='mr-2 h-4 w-4 animate-spin' />
                  ) : (
                    <ShieldCheck className='mr-2 h-4 w-4' />
                  )}
                  Find matches
                </Button>
              </div>

              <div
                className='flex flex-wrap items-center gap-2'
                data-testid='forwarder-package-confidence-filter'
              >
                <span className='text-sm text-muted-foreground'>
                  Suggestion confidence
                </span>
                {(
                  [
                    ['all', 'All'],
                    ['high', 'High'],
                    ['medium', 'Medium'],
                    ['low', 'Low'],
                  ] satisfies Array<
                    [ForwarderPackageSuggestionConfidenceFilter, string]
                  >
                ).map(([value, label]) => (
                  <Button
                    key={value}
                    type='button'
                    variant={
                      packageSuggestionConfidenceFilter === value
                        ? 'default'
                        : 'outline'
                    }
                    size='sm'
                    data-testid={'forwarder-package-confidence-filter-' + value}
                    onClick={() => setPackageSuggestionConfidenceFilter(value)}
                  >
                    {label}
                  </Button>
                ))}
              </div>

              <dl
                className='grid gap-3 rounded-md border bg-muted/20 p-4 text-sm sm:grid-cols-5'
                data-testid='forwarder-package-review-summary'
              >
                <div>
                  <dt className='text-muted-foreground'>Source records</dt>
                  <dd className='text-lg font-semibold'>
                    {packageReviewSummary.packageCount}
                  </dd>
                </div>
                <div>
                  <dt className='text-muted-foreground'>Linked</dt>
                  <dd className='text-lg font-semibold'>
                    {packageReviewSummary.linkedCount}
                  </dd>
                </div>
                <div>
                  <dt className='text-muted-foreground'>Unlinked</dt>
                  <dd className='text-lg font-semibold'>
                    {packageReviewSummary.unlinkedCount}
                  </dd>
                </div>
                <div>
                  <dt className='text-muted-foreground'>Audit events</dt>
                  <dd className='text-lg font-semibold'>
                    {packageReviewSummary.auditEventCount}
                  </dd>
                </div>
                <div>
                  <dt className='text-muted-foreground'>Suggestions</dt>
                  <dd className='text-lg font-semibold'>
                    {packageReviewSummary.suggestionCount}
                  </dd>
                </div>
              </dl>

              {packageSuggestionSummary ? (
                <dl
                  className='grid gap-3 rounded-md border bg-muted/20 p-4 text-sm sm:grid-cols-5'
                  data-testid='forwarder-package-suggestion-summary'
                >
                  <div>
                    <dt className='text-muted-foreground'>Candidates</dt>
                    <dd className='text-lg font-semibold'>
                      {packageSuggestionSummary.count ?? 0}
                    </dd>
                  </div>
                  <div>
                    <dt className='text-muted-foreground'>Scoped packages</dt>
                    <dd className='text-lg font-semibold'>
                      {packageSuggestionSummary.scoped_packages ?? 0}
                    </dd>
                  </div>
                  <div>
                    <dt className='text-muted-foreground'>High confidence</dt>
                    <dd className='text-lg font-semibold'>
                      {packageSuggestionSummary.high_confidence ?? 0}
                    </dd>
                  </div>
                  <div>
                    <dt className='text-muted-foreground'>Medium confidence</dt>
                    <dd className='text-lg font-semibold'>
                      {packageSuggestionSummary.medium_confidence ?? 0}
                    </dd>
                  </div>
                  <div>
                    <dt className='text-muted-foreground'>Low confidence</dt>
                    <dd className='text-lg font-semibold'>
                      {packageSuggestionSummary.low_confidence ?? 0}
                    </dd>
                  </div>
                  <div className='sm:col-span-5'>
                    <dt className='text-muted-foreground'>Active filter</dt>
                    <dd className='text-lg font-semibold'>
                      {packageSuggestionActiveFilter ?? 'all'}
                    </dd>
                  </div>
                </dl>
              ) : null}

              <div
                className='flex flex-wrap items-center gap-2'
                data-testid='forwarder-package-review-filter'
              >
                {(
                  [
                    ['all', 'All'],
                    ['linked', 'Linked'],
                    ['unlinked', 'Unlinked'],
                    ['suggested', 'Suggestions'],
                  ] satisfies Array<[ForwarderPackageReviewFilter, string]>
                ).map(([value, label]) => (
                  <Button
                    key={value}
                    type='button'
                    variant={
                      packageReviewFilter === value ? 'default' : 'outline'
                    }
                    size='sm'
                    data-testid={'forwarder-package-review-filter-' + value}
                    onClick={() => setPackageReviewFilter(value)}
                  >
                    {label}
                  </Button>
                ))}
                <span
                  className='text-sm text-muted-foreground'
                  data-testid='forwarder-package-review-filter-result'
                >
                  Showing {filteredPackages.length} of {packages.length}{' '}
                  packages
                </span>
              </div>

              <div className='grid gap-4 lg:grid-cols-[minmax(280px,360px)_1fr]'>
                <div className='rounded-md border p-4'>
                  <div className='mb-3 flex items-center gap-2'>
                    <PackagePlus className='h-4 w-4' />
                    <h3 className='font-medium'>Manual import</h3>
                  </div>
                  <div className='grid gap-3'>
                    <div className='grid gap-1.5'>
                      <Label htmlFor='forwarder-package-profile'>Profile</Label>
                      <Input
                        id='forwarder-package-profile'
                        data-testid='forwarder-package-profile'
                        value={packageForm.profile_id}
                        onChange={(event) =>
                          updatePackageForm('profile_id', event.target.value)
                        }
                      />
                    </div>
                    <div className='grid grid-cols-2 gap-3'>
                      <div className='grid gap-1.5'>
                        <Label htmlFor='forwarder-package-provider'>
                          Provider
                        </Label>
                        <Input
                          id='forwarder-package-provider'
                          data-testid='forwarder-package-provider'
                          value={packageForm.provider}
                          onChange={(event) =>
                            updatePackageForm('provider', event.target.value)
                          }
                        />
                      </div>
                      <div className='grid gap-1.5'>
                        <Label htmlFor='forwarder-package-source'>Source</Label>
                        <Input
                          id='forwarder-package-source'
                          data-testid='forwarder-package-source'
                          value={packageForm.source}
                          onChange={(event) =>
                            updatePackageForm('source', event.target.value)
                          }
                        />
                      </div>
                    </div>
                    <div className='grid gap-1.5'>
                      <Label htmlFor='forwarder-package-external-id'>
                        Package ID
                      </Label>
                      <Input
                        id='forwarder-package-external-id'
                        data-testid='forwarder-package-external-id'
                        value={packageForm.external_package_id}
                        onChange={(event) =>
                          updatePackageForm(
                            'external_package_id',
                            event.target.value
                          )
                        }
                      />
                    </div>
                    <div className='grid grid-cols-2 gap-3'>
                      <div className='grid gap-1.5'>
                        <Label htmlFor='forwarder-package-status'>Status</Label>
                        <Input
                          id='forwarder-package-status'
                          data-testid='forwarder-package-status'
                          value={packageForm.status}
                          onChange={(event) =>
                            updatePackageForm('status', event.target.value)
                          }
                        />
                      </div>
                      <div className='grid gap-1.5'>
                        <Label htmlFor='forwarder-package-weight'>
                          Weight g
                        </Label>
                        <Input
                          id='forwarder-package-weight'
                          data-testid='forwarder-package-weight'
                          inputMode='numeric'
                          value={packageForm.weight_grams}
                          onChange={(event) =>
                            updatePackageForm(
                              'weight_grams',
                              event.target.value
                            )
                          }
                        />
                      </div>
                    </div>
                    <div className='grid gap-1.5'>
                      <Label htmlFor='forwarder-package-tracking'>
                        Tracking
                      </Label>
                      <Input
                        id='forwarder-package-tracking'
                        data-testid='forwarder-package-tracking'
                        value={packageForm.tracking_number}
                        onChange={(event) =>
                          updatePackageForm(
                            'tracking_number',
                            event.target.value
                          )
                        }
                      />
                    </div>
                    <div className='grid gap-1.5'>
                      <Label htmlFor='forwarder-package-warehouse'>
                        Warehouse
                      </Label>
                      <Input
                        id='forwarder-package-warehouse'
                        data-testid='forwarder-package-warehouse'
                        value={packageForm.warehouse_location}
                        onChange={(event) =>
                          updatePackageForm(
                            'warehouse_location',
                            event.target.value
                          )
                        }
                      />
                    </div>
                    <Button
                      data-testid='forwarder-package-import'
                      onClick={() => void importPackage()}
                      disabled={packagesLoading}
                    >
                      {packagesLoading ? (
                        <Loader2 className='mr-2 h-4 w-4 animate-spin' />
                      ) : (
                        <PackagePlus className='mr-2 h-4 w-4' />
                      )}
                      Import package
                    </Button>
                  </div>
                </div>

                <div className='rounded-md border p-4'>
                  <div className='mb-3 flex items-center gap-2'>
                    <FileUp className='h-4 w-4' />
                    <h3 className='font-medium'>CSV import</h3>
                  </div>
                  <div className='grid gap-3'>
                    <div className='grid gap-1.5'>
                      <Label htmlFor='forwarder-package-csv'>Package CSV</Label>
                      <Textarea
                        id='forwarder-package-csv'
                        data-testid='forwarder-package-csv'
                        className='min-h-32 font-mono text-xs'
                        value={packageCSV}
                        onChange={(event) => setPackageCSV(event.target.value)}
                      />
                    </div>
                    <Button
                      variant='secondary'
                      data-testid='forwarder-package-import-csv'
                      onClick={() => void importPackageCSV()}
                      disabled={packagesLoading}
                    >
                      {packagesLoading ? (
                        <Loader2 className='mr-2 h-4 w-4 animate-spin' />
                      ) : (
                        <FileUp className='mr-2 h-4 w-4' />
                      )}
                      Import CSV
                    </Button>
                  </div>
                </div>

                <div className='rounded-md border p-4'>
                  <div className='mb-3 flex items-center gap-2'>
                    <Inbox className='h-4 w-4' />
                    <h3 className='font-medium'>Email import</h3>
                  </div>
                  <div className='grid gap-3'>
                    <div className='grid gap-1.5'>
                      <Label htmlFor='forwarder-package-email'>
                        Package email
                      </Label>
                      <Textarea
                        id='forwarder-package-email'
                        data-testid='forwarder-package-email'
                        className='min-h-32 font-mono text-xs'
                        value={packageEmail}
                        onChange={(event) =>
                          setPackageEmail(event.target.value)
                        }
                      />
                    </div>
                    <Button
                      variant='secondary'
                      data-testid='forwarder-package-import-email'
                      onClick={() => void importPackageEmail()}
                      disabled={packagesLoading}
                    >
                      {packagesLoading ? (
                        <Loader2 className='mr-2 h-4 w-4 animate-spin' />
                      ) : (
                        <Inbox className='mr-2 h-4 w-4' />
                      )}
                      Import Email
                    </Button>
                  </div>
                </div>

                <div className='space-y-3'>
                  {packageError ? (
                    <div
                      className='rounded-md border border-destructive/40 bg-destructive/10 p-4 text-sm'
                      data-testid='forwarder-package-error'
                    >
                      <p className='font-medium'>
                        Package inbox could not update.
                      </p>
                      <p className='mt-1 text-muted-foreground'>
                        {packageError}
                      </p>
                    </div>
                  ) : null}

                  {packageResult ? (
                    <div
                      className='rounded-md border bg-muted/30 p-3 text-sm'
                      data-testid='forwarder-package-result'
                    >
                      {packageResult}
                    </div>
                  ) : null}

                  {packageSuggestionResult ? (
                    <div
                      className='rounded-md border bg-muted/30 p-3 text-sm'
                      data-testid='forwarder-package-suggestion-result'
                    >
                      {packageSuggestionResult}
                    </div>
                  ) : null}

                  {packageSuggestionError ? (
                    <div
                      className='rounded-md border border-destructive/40 bg-destructive/10 p-4 text-sm'
                      data-testid='forwarder-package-suggestion-error'
                    >
                      <p className='font-medium'>
                        Match suggestions could not load.
                      </p>
                      <p className='mt-1 text-muted-foreground'>
                        {packageSuggestionError}
                      </p>
                      <Button
                        type='button'
                        size='sm'
                        variant='outline'
                        className='mt-3'
                        data-testid='forwarder-package-suggestion-retry'
                        onClick={() => void loadPackageSuggestions()}
                        disabled={packagesLoading}
                      >
                        Retry suggestions
                      </Button>
                    </div>
                  ) : null}

                  {packageCSVErrors.length > 0 ? (
                    <div
                      className='rounded-md border border-amber-300 bg-amber-50 p-4 text-sm text-amber-950 dark:border-amber-900 dark:bg-amber-950/30 dark:text-amber-100'
                      data-testid='forwarder-package-csv-errors'
                    >
                      <p className='font-medium'>CSV rows need attention.</p>
                      <ul className='mt-2 space-y-1'>
                        {packageCSVErrors.map((rowError, index) => (
                          <li key={index}>
                            Row {rowError.row ?? 'unknown'}: {rowError.error}
                          </li>
                        ))}
                      </ul>
                    </div>
                  ) : null}

                  {packages.length === 0 && !packagesLoading ? (
                    <div
                      className='rounded-md border border-dashed p-6'
                      data-testid='forwarder-package-empty'
                    >
                      <p className='font-medium'>
                        No purchase source records listed.
                      </p>
                      <p className='mt-1 text-sm text-muted-foreground'>
                        Import forwarder evidence or refresh the current profile
                        inbox.
                      </p>
                    </div>
                  ) : null}

                  {packages.length > 0 && filteredPackages.length === 0 ? (
                    <div
                      className='rounded-md border border-dashed p-6'
                      data-testid='forwarder-package-filter-empty'
                    >
                      <p className='font-medium'>
                        No source records match this review state.
                      </p>
                      <p className='mt-1 text-sm text-muted-foreground'>
                        Switch filters or refresh link and suggestion evidence.
                      </p>
                    </div>
                  ) : null}

                  {packages.length > 0 ? (
                    <div
                      className='space-y-3'
                      data-testid='forwarder-package-list'
                    >
                      {filteredPackages.map((pkg) => {
                        const selected = selectedPackageID === pkg.id
                        const rawPayload = pkg.raw_payload
                          ? JSON.stringify(pkg.raw_payload, null, 2)
                          : 'No raw source payload returned for this package.'
                        const linkForm = packageLinkFormFor(pkg.id)
                        const links = packageLinks[pkg.id] ?? []
                        const events = packageLinkEvents[pkg.id] ?? []
                        const suggestions = packageSuggestions[pkg.id] ?? []
                        const evidenceBadges = [
                          {
                            label:
                              links.length === 1
                                ? '1 active link'
                                : links.length + ' active links',
                            visible: links.length > 0,
                          },
                          {
                            label:
                              suggestions.length === 1
                                ? '1 suggestion'
                                : suggestions.length + ' suggestions',
                            visible: suggestions.length > 0,
                          },
                          {
                            label:
                              events.length === 1
                                ? '1 audit event'
                                : events.length + ' audit events',
                            visible: events.length > 0,
                          },
                        ].filter((badge) => badge.visible)
                        return (
                          <article
                            key={pkg.id}
                            className='rounded-md border p-4'
                            data-testid='forwarder-package-row'
                          >
                            <div className='flex flex-wrap items-start justify-between gap-3'>
                              <div>
                                <h3 className='font-semibold'>
                                  {pkg.external_package_id}
                                </h3>
                                <p className='text-sm text-muted-foreground'>
                                  {pkg.provider} / {pkg.source} ·{' '}
                                  {pkg.tracking_number || 'tracking pending'}
                                </p>
                              </div>
                              <span className='rounded-md border px-2 py-1 text-xs font-medium'>
                                {labelForStatus(pkg.status)}
                              </span>
                            </div>
                            <div
                              className='mt-3 flex flex-wrap gap-2 text-xs'
                              data-testid='forwarder-package-row-evidence'
                            >
                              {evidenceBadges.length > 0 ? (
                                evidenceBadges.map((badge) => (
                                  <span
                                    key={badge.label}
                                    className='rounded-md border bg-muted/30 px-2 py-1 font-medium'
                                  >
                                    {badge.label}
                                  </span>
                                ))
                              ) : (
                                <span className='text-muted-foreground'>
                                  No loaded reconciliation evidence
                                </span>
                              )}
                            </div>
                            <dl className='mt-3 grid gap-2 text-sm sm:grid-cols-4'>
                              <div>
                                <dt className='text-muted-foreground'>
                                  Warehouse
                                </dt>
                                <dd>{pkg.warehouse_location || 'Pending'}</dd>
                              </div>
                              <div>
                                <dt className='text-muted-foreground'>
                                  Sender
                                </dt>
                                <dd>{pkg.sender || 'Pending'}</dd>
                              </div>
                              <div>
                                <dt className='text-muted-foreground'>
                                  Weight
                                </dt>
                                <dd>{pkg.weight_grams ?? 0} g</dd>
                              </div>
                              <div>
                                <dt className='text-muted-foreground'>
                                  Provenance
                                </dt>
                                <dd className='break-all'>
                                  {pkg.provenance_key}
                                </dd>
                              </div>
                            </dl>
                            <div className='mt-4 flex justify-end'>
                              <Button
                                type='button'
                                variant='outline'
                                size='sm'
                                data-testid='forwarder-package-detail-toggle'
                                onClick={() =>
                                  void selectPackage(pkg.id, selected)
                                }
                              >
                                {selected ? 'Hide details' : 'View details'}
                              </Button>
                            </div>
                            {selected ? (
                              <div
                                className='mt-4 space-y-4 rounded-md border bg-muted/20 p-4'
                                data-testid='forwarder-package-detail'
                              >
                                <dl className='grid gap-3 text-sm md:grid-cols-3'>
                                  <div>
                                    <dt className='text-muted-foreground'>
                                      Package ID
                                    </dt>
                                    <dd className='break-all'>
                                      {pkg.external_package_id}
                                    </dd>
                                  </div>
                                  <div>
                                    <dt className='text-muted-foreground'>
                                      Shipment ID
                                    </dt>
                                    <dd className='break-all'>
                                      {packageDetailValue(pkg.shipment_id)}
                                    </dd>
                                  </div>
                                  <div>
                                    <dt className='text-muted-foreground'>
                                      Tracking number
                                    </dt>
                                    <dd className='break-all'>
                                      {packageDetailValue(pkg.tracking_number)}
                                    </dd>
                                  </div>
                                  <div>
                                    <dt className='text-muted-foreground'>
                                      Received
                                    </dt>
                                    <dd>
                                      {packageDetailValue(pkg.received_at)}
                                    </dd>
                                  </div>
                                  <div>
                                    <dt className='text-muted-foreground'>
                                      Created
                                    </dt>
                                    <dd>
                                      {packageDetailValue(pkg.created_at)}
                                    </dd>
                                  </div>
                                  <div>
                                    <dt className='text-muted-foreground'>
                                      Updated
                                    </dt>
                                    <dd>
                                      {packageDetailValue(pkg.updated_at)}
                                    </dd>
                                  </div>
                                </dl>
                                <div className='space-y-2'>
                                  <p className='text-sm font-medium'>
                                    Source provenance
                                  </p>
                                  <pre
                                    className='max-h-48 overflow-auto rounded-md border bg-background p-3 text-xs'
                                    data-testid='forwarder-package-raw-payload'
                                  >
                                    {rawPayload}
                                  </pre>
                                </div>
                                <div
                                  className='space-y-3 rounded-md border bg-background p-4'
                                  data-testid='forwarder-package-link-panel'
                                >
                                  {suggestions.length > 0 ? (
                                    <div
                                      className='space-y-2 rounded-md border bg-muted/20 p-3 text-sm'
                                      data-testid='forwarder-package-match-suggestions'
                                    >
                                      <p className='font-medium'>
                                        Suggested purchase matches
                                      </p>
                                      {suggestions.map((suggestion, index) => (
                                        <div
                                          key={suggestion.id ?? index}
                                          className='rounded-md border bg-background p-3'
                                          data-testid='forwarder-package-match-suggestion'
                                        >
                                          <div className='flex flex-wrap items-start justify-between gap-3'>
                                            <div>
                                              <p className='font-medium'>
                                                {labelForStatus(
                                                  suggestion.confidence_label ??
                                                    'suggested'
                                                )}{' '}
                                                match to item{' '}
                                                {suggestion.item_id}
                                              </p>
                                              <p className='text-xs text-muted-foreground'>
                                                Score{' '}
                                                {suggestion.confidence_score ??
                                                  0}
                                                {suggestion.expected_arrival_id
                                                  ? ' · arrival ' +
                                                    suggestion.expected_arrival_id
                                                  : ''}
                                              </p>
                                            </div>
                                            <Button
                                              type='button'
                                              size='sm'
                                              variant='outline'
                                              data-testid='forwarder-package-match-suggestion-use'
                                              onClick={() =>
                                                applyPackageSuggestion(
                                                  pkg.id,
                                                  suggestion
                                                )
                                              }
                                            >
                                              Use suggestion
                                            </Button>
                                          </div>
                                          {(suggestion.signals ?? []).length >
                                          0 ? (
                                            <ul
                                              className='mt-2 list-disc space-y-1 ps-5 text-xs text-muted-foreground'
                                              data-testid='forwarder-package-match-signals'
                                            >
                                              {(suggestion.signals ?? []).map(
                                                (signal, signalIndex) => (
                                                  <li key={signalIndex}>
                                                    {signal.name ?? 'signal'}:{' '}
                                                    {signal.evidence ??
                                                      'matched'}{' '}
                                                    ({signal.score ?? 0})
                                                  </li>
                                                )
                                              )}
                                            </ul>
                                          ) : null}
                                          {(suggestion.audit_trail ?? [])
                                            .length > 0 ? (
                                            <ul
                                              className='mt-2 list-disc space-y-1 ps-5 text-xs text-muted-foreground'
                                              data-testid='forwarder-package-match-audit-trail'
                                            >
                                              {(
                                                suggestion.audit_trail ?? []
                                              ).map((entry, auditIndex) => (
                                                <li key={auditIndex}>
                                                  {entry}
                                                </li>
                                              ))}
                                            </ul>
                                          ) : null}
                                        </div>
                                      ))}
                                    </div>
                                  ) : null}
                                  <div>
                                    <p className='text-sm font-medium'>
                                      Reconciliation link
                                    </p>
                                    <p className='text-xs text-muted-foreground'>
                                      Match this source evidence to the reviewed
                                      inventory item and expected arrival
                                      target.
                                    </p>
                                  </div>
                                  <div className='flex flex-wrap justify-end gap-2'>
                                    <Button
                                      type='button'
                                      size='sm'
                                      variant='outline'
                                      data-testid='forwarder-package-match-suggestions-load-scoped'
                                      onClick={() =>
                                        void loadPackageSuggestions(pkg.id)
                                      }
                                      disabled={packagesLoading}
                                    >
                                      {packagesLoading ? (
                                        <Loader2 className='mr-2 h-4 w-4 animate-spin' />
                                      ) : (
                                        <ShieldCheck className='mr-2 h-4 w-4' />
                                      )}
                                      Find matches for this source record
                                    </Button>
                                  </div>
                                  {links.length > 0 ? (
                                    <div
                                      className='rounded-md border bg-muted/30 p-3 text-sm'
                                      data-testid='forwarder-package-link-state'
                                    >
                                      {links.map((link) => (
                                        <div
                                          key={link.id}
                                          className='space-y-1'
                                        >
                                          <p>
                                            {labelForStatus(
                                              link.decision || 'confirmed'
                                            )}{' '}
                                            to item {link.item_id}
                                            {link.expected_arrival_id
                                              ? ' / arrival ' +
                                                link.expected_arrival_id
                                              : ''}
                                          </p>
                                          <p className='text-xs text-muted-foreground'>
                                            Source {link.source}
                                            {link.notes
                                              ? ' · ' + link.notes
                                              : ''}
                                          </p>
                                          {link.audit_trail &&
                                          link.audit_trail.length > 0 ? (
                                            <ul
                                              className='list-disc space-y-1 ps-5 text-xs text-muted-foreground'
                                              data-testid='forwarder-package-link-audit-trail'
                                            >
                                              {link.audit_trail.map(
                                                (entry, index) => (
                                                  <li key={index}>{entry}</li>
                                                )
                                              )}
                                            </ul>
                                          ) : null}
                                        </div>
                                      ))}
                                    </div>
                                  ) : (
                                    <div
                                      className='rounded-md border border-dashed p-3 text-sm text-muted-foreground'
                                      data-testid='forwarder-package-link-empty'
                                    >
                                      No reconciliation link recorded for this
                                      source record.
                                    </div>
                                  )}
                                  <div className='grid gap-3 md:grid-cols-2'>
                                    <div className='grid gap-1.5'>
                                      <Label
                                        htmlFor={
                                          'forwarder-link-item-' + pkg.id
                                        }
                                      >
                                        Item ID
                                      </Label>
                                      <Input
                                        id={'forwarder-link-item-' + pkg.id}
                                        data-testid='forwarder-package-link-item'
                                        value={linkForm.item_id}
                                        onChange={(event) =>
                                          updatePackageLinkForm(
                                            pkg.id,
                                            'item_id',
                                            event.target.value
                                          )
                                        }
                                      />
                                    </div>
                                    <div className='grid gap-1.5'>
                                      <Label
                                        htmlFor={
                                          'forwarder-link-arrival-' + pkg.id
                                        }
                                      >
                                        Expected arrival ID
                                      </Label>
                                      <Input
                                        id={'forwarder-link-arrival-' + pkg.id}
                                        data-testid='forwarder-package-link-arrival'
                                        value={linkForm.expected_arrival_id}
                                        onChange={(event) =>
                                          updatePackageLinkForm(
                                            pkg.id,
                                            'expected_arrival_id',
                                            event.target.value
                                          )
                                        }
                                      />
                                    </div>
                                    <div className='grid gap-1.5'>
                                      <Label
                                        htmlFor={
                                          'forwarder-link-lifecycle-' + pkg.id
                                        }
                                      >
                                        Lifecycle entry ID
                                      </Label>
                                      <Input
                                        id={
                                          'forwarder-link-lifecycle-' + pkg.id
                                        }
                                        data-testid='forwarder-package-link-lifecycle'
                                        value={linkForm.lifecycle_entry_id}
                                        onChange={(event) =>
                                          updatePackageLinkForm(
                                            pkg.id,
                                            'lifecycle_entry_id',
                                            event.target.value
                                          )
                                        }
                                      />
                                    </div>
                                    <div className='grid gap-1.5'>
                                      <Label
                                        htmlFor={
                                          'forwarder-link-source-' + pkg.id
                                        }
                                      >
                                        Source
                                      </Label>
                                      <Input
                                        id={'forwarder-link-source-' + pkg.id}
                                        data-testid='forwarder-package-link-source'
                                        value={linkForm.source}
                                        onChange={(event) =>
                                          updatePackageLinkForm(
                                            pkg.id,
                                            'source',
                                            event.target.value
                                          )
                                        }
                                      />
                                    </div>
                                  </div>
                                  <div className='grid gap-1.5'>
                                    <Label
                                      htmlFor={'forwarder-link-notes-' + pkg.id}
                                    >
                                      Notes
                                    </Label>
                                    <Textarea
                                      id={'forwarder-link-notes-' + pkg.id}
                                      data-testid='forwarder-package-link-notes'
                                      value={linkForm.notes}
                                      onChange={(event) =>
                                        updatePackageLinkForm(
                                          pkg.id,
                                          'notes',
                                          event.target.value
                                        )
                                      }
                                    />
                                  </div>
                                  <div className='flex flex-wrap justify-end gap-2'>
                                    <Button
                                      type='button'
                                      size='sm'
                                      data-testid='forwarder-package-link-save'
                                      onClick={() => void linkPackage(pkg)}
                                    >
                                      Confirm link
                                    </Button>
                                    <Button
                                      type='button'
                                      size='sm'
                                      variant='secondary'
                                      data-testid='forwarder-package-link-override'
                                      onClick={() =>
                                        void linkPackage(pkg, 'overridden')
                                      }
                                    >
                                      Override link
                                    </Button>
                                    <Button
                                      type='button'
                                      size='sm'
                                      variant='outline'
                                      data-testid='forwarder-package-link-unlink'
                                      onClick={() => void unlinkPackage(pkg)}
                                      disabled={links.length === 0}
                                    >
                                      Unlink package
                                    </Button>
                                  </div>
                                  <div
                                    className='rounded-md border bg-muted/20 p-3 text-sm'
                                    data-testid='forwarder-package-link-events'
                                  >
                                    <p className='font-medium'>
                                      Decision audit
                                    </p>
                                    {events.length > 0 ? (
                                      <ul className='mt-2 space-y-2'>
                                        {events.map((event) => (
                                          <li
                                            key={event.id}
                                            className='space-y-1'
                                          >
                                            <p>
                                              <span className='font-medium'>
                                                {labelForStatus(event.action)}
                                              </span>
                                              {event.item_id
                                                ? ' item ' + event.item_id
                                                : ''}
                                              {event.lifecycle_entry_id
                                                ? ' / lifecycle ' +
                                                  event.lifecycle_entry_id
                                                : ''}
                                              {event.expected_arrival_id
                                                ? ' / arrival ' +
                                                  event.expected_arrival_id
                                                : ''}
                                              {event.previous_item_id ||
                                              event.previous_lifecycle_entry_id ||
                                              event.previous_expected_arrival_id ? (
                                                <span className='text-muted-foreground'>
                                                  {' '}
                                                  (previous
                                                  {event.previous_item_id
                                                    ? ' item ' +
                                                      event.previous_item_id
                                                    : ''}
                                                  {event.previous_lifecycle_entry_id
                                                    ? ' / lifecycle ' +
                                                      event.previous_lifecycle_entry_id
                                                    : ''}
                                                  {event.previous_expected_arrival_id
                                                    ? ' / arrival ' +
                                                      event.previous_expected_arrival_id
                                                    : ''}
                                                  )
                                                </span>
                                              ) : null}
                                            </p>
                                            <p className='text-xs text-muted-foreground'>
                                              via {event.source}
                                              {event.created_at
                                                ? ' · ' + event.created_at
                                                : ''}
                                              {event.notes
                                                ? ' · ' + event.notes
                                                : ''}
                                            </p>
                                            {event.audit_trail &&
                                            event.audit_trail.length > 0 ? (
                                              <ul
                                                className='list-disc space-y-1 ps-5 text-xs text-muted-foreground'
                                                data-testid='forwarder-package-link-event-audit-trail'
                                              >
                                                {event.audit_trail.map(
                                                  (entry, index) => (
                                                    <li key={index}>{entry}</li>
                                                  )
                                                )}
                                              </ul>
                                            ) : null}
                                          </li>
                                        ))}
                                      </ul>
                                    ) : (
                                      <p className='mt-1 text-muted-foreground'>
                                        No link decisions recorded yet.
                                      </p>
                                    )}
                                  </div>
                                  {packageLinkResults[pkg.id] ? (
                                    <div
                                      className='rounded-md border bg-muted/30 p-3 text-sm'
                                      data-testid='forwarder-package-link-result'
                                    >
                                      {packageLinkResults[pkg.id]}
                                    </div>
                                  ) : null}
                                  {packageLinkErrors[pkg.id] ? (
                                    <div
                                      className='rounded-md border border-destructive/40 bg-destructive/10 p-3 text-sm'
                                      data-testid='forwarder-package-link-error'
                                    >
                                      {packageLinkErrors[pkg.id]}
                                    </div>
                                  ) : null}
                                </div>
                              </div>
                            ) : null}
                          </article>
                        )
                      })}
                    </div>
                  ) : null}
                </div>
              </div>
            </section>
          </>
        ) : null}
      </Main>

      <Dialog
        open={addDialogOpen}
        onOpenChange={(open) => {
          setAddDialogOpen(open)
          if (open) {
            setPurchaseImportError(null)
          }
        }}
      >
        <DialogContent data-testid='purchases-add-dialog'>
          <DialogHeader>
            <DialogTitle>New purchase</DialogTitle>
            <DialogDescription>
              Create or import a purchase draft before confirming persistence.
            </DialogDescription>
          </DialogHeader>
          <Tabs
            value={addDialogTab}
            onValueChange={(value) =>
              setAddDialogTab(value as 'new' | 'csv' | 'email')
            }
            data-testid='purchases-add-tabs'
          >
            <TabsList aria-label='Purchase creation modes'>
              <TabsTrigger value='new' data-testid='purchases-add-tab-new'>
                New
              </TabsTrigger>
              <TabsTrigger value='csv' data-testid='purchases-add-tab-csv'>
                CSV
              </TabsTrigger>
              <TabsTrigger value='email' data-testid='purchases-add-tab-email'>
                Email
              </TabsTrigger>
            </TabsList>
            <TabsContent value='new' className='space-y-3 pt-4'>
              <div className='grid gap-1.5'>
                <Label htmlFor='purchase-manual-title'>Purchase title</Label>
                <Input
                  id='purchase-manual-title'
                  data-testid='purchases-add-new-title'
                  placeholder='Item name or order title'
                  value={manualPurchaseForm.title}
                  onChange={(event) =>
                    updateManualPurchaseForm('title', event.target.value)
                  }
                />
              </div>
              <div className='grid gap-3 sm:grid-cols-3'>
                <div className='grid gap-1.5'>
                  <Label htmlFor='purchase-manual-source'>Source</Label>
                  <Input
                    id='purchase-manual-source'
                    data-testid='purchases-add-new-source'
                    value={manualPurchaseForm.source}
                    onChange={(event) =>
                      updateManualPurchaseForm('source', event.target.value)
                    }
                  />
                </div>
                <div className='grid gap-1.5'>
                  <Label htmlFor='purchase-manual-price'>Price</Label>
                  <Input
                    id='purchase-manual-price'
                    data-testid='purchases-add-new-price'
                    placeholder='AU $0.00'
                    value={manualPurchaseForm.price}
                    onChange={(event) =>
                      updateManualPurchaseForm('price', event.target.value)
                    }
                  />
                </div>
                <div className='grid gap-1.5'>
                  <Label htmlFor='purchase-manual-tracking'>Tracking</Label>
                  <Input
                    id='purchase-manual-tracking'
                    data-testid='purchases-add-new-tracking'
                    placeholder='Carrier or tracking number'
                    value={manualPurchaseForm.tracking}
                    onChange={(event) =>
                      updateManualPurchaseForm('tracking', event.target.value)
                    }
                  />
                </div>
              </div>
              {manualPurchaseError ? (
                <p
                  className='text-sm text-destructive'
                  data-testid='purchases-add-new-error'
                >
                  {manualPurchaseError}
                </p>
              ) : null}
            </TabsContent>
            <TabsContent value='csv' className='space-y-3 pt-4'>
              <Label htmlFor='purchase-csv-import'>CSV import</Label>
              <Textarea
                id='purchase-csv-import'
                data-testid='purchases-add-csv-input'
                placeholder='Paste CSV rows with source, title, price, date, and tracking fields'
                value={purchaseCSVImport}
                onChange={(event) => {
                  setPurchaseCSVImport(event.target.value)
                  setPurchaseCSVPreview([])
                  setPurchaseImportError(null)
                }}
              />
              <Button
                type='button'
                variant='outline'
                data-testid='purchases-add-csv-preview'
                onClick={previewPurchaseCSVImport}
              >
                Preview CSV
              </Button>
              {purchaseCSVPreview.length > 0 ? (
                <div
                  className='rounded-md border bg-muted/20 p-3 text-sm'
                  data-testid='purchases-add-csv-preview-result'
                >
                  <p className='font-medium'>
                    Previewing {purchaseCSVPreview.length} CSV purchase draft
                    {purchaseCSVPreview.length === 1 ? '' : 's'}
                  </p>
                  <div className='mt-2 space-y-2'>
                    {purchaseCSVPreview.map((draft) => (
                      <div
                        key={draft.key}
                        className='rounded-md border bg-background p-2'
                        data-testid='purchases-import-preview-row'
                      >
                        <p className='font-medium'>{draft.title}</p>
                        <p className='text-muted-foreground'>
                          {draft.provenance} · {draft.price} ·{' '}
                          {draft.purchaseDate || 'date pending'}
                        </p>
                        <p className='text-muted-foreground'>
                          Seller {draft.seller || 'pending'} · Channel{' '}
                          {draft.channel} · Tracking{' '}
                          {draft.tracking || 'pending'} · Delivery{' '}
                          {draft.delivery || 'pending'}
                        </p>
                      </div>
                    ))}
                  </div>
                </div>
              ) : null}
            </TabsContent>
            <TabsContent value='email' className='space-y-3 pt-4'>
              <Label htmlFor='purchase-email-import'>Email or order text</Label>
              <Textarea
                id='purchase-email-import'
                data-testid='purchases-add-email-input'
                placeholder='Paste order confirmation text for preview parsing'
                value={purchaseEmailImport}
                onChange={(event) => {
                  setPurchaseEmailImport(event.target.value)
                  setPurchaseEmailPreview([])
                  setPurchaseImportError(null)
                }}
              />
              <Button
                type='button'
                variant='outline'
                data-testid='purchases-add-email-preview'
                onClick={previewPurchaseEmailImport}
              >
                Preview email
              </Button>
              {purchaseEmailPreview.length > 0 ? (
                <div
                  className='rounded-md border bg-muted/20 p-3 text-sm'
                  data-testid='purchases-add-email-preview-result'
                >
                  <p className='font-medium'>Previewing email purchase draft</p>
                  {purchaseEmailPreview.map((draft) => (
                    <div
                      key={draft.key}
                      className='mt-2 rounded-md border bg-background p-2'
                      data-testid='purchases-import-preview-row'
                    >
                      <p className='font-medium'>{draft.title}</p>
                      <p className='text-muted-foreground'>
                        {draft.provenance} · {draft.price} ·{' '}
                        {draft.purchaseDate || 'date pending'}
                      </p>
                      <p className='text-muted-foreground'>
                        Seller {draft.seller || 'pending'} · Channel{' '}
                        {draft.channel} · Tracking {draft.tracking || 'pending'}{' '}
                        · Delivery {draft.delivery || 'pending'}
                      </p>
                    </div>
                  ))}
                </div>
              ) : null}
            </TabsContent>
          </Tabs>
          {purchaseImportError ? (
            <p
              className='text-sm text-destructive'
              data-testid='purchases-add-import-error'
            >
              {purchaseImportError}
            </p>
          ) : null}
          <DialogFooter>
            <Button variant='outline' onClick={() => setAddDialogOpen(false)}>
              Cancel
            </Button>
            {addDialogTab === 'new' ? (
              <Button
                type='button'
                data-testid='purchases-add-new-save'
                onClick={saveManualPurchaseDraft}
              >
                Save draft
              </Button>
            ) : null}
            {addDialogTab === 'csv' ? (
              <Button
                type='button'
                data-testid='purchases-add-csv-confirm'
                onClick={() => confirmPurchaseImportDrafts('csv')}
              >
                Confirm CSV drafts
              </Button>
            ) : null}
            {addDialogTab === 'email' ? (
              <Button
                type='button'
                data-testid='purchases-add-email-confirm'
                onClick={() => confirmPurchaseImportDrafts('email')}
              >
                Confirm email draft
              </Button>
            ) : null}
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog
        open={pendingAction !== null}
        onOpenChange={(open) => {
          if (!open) {
            setPendingAction(null)
          }
        }}
      >
        <DialogContent data-testid='purchase-inbox-confirm-dialog'>
          <DialogHeader>
            <DialogTitle>Confirm purchase action</DialogTitle>
            <DialogDescription>
              Cabinet will not link or convert this purchase until you confirm
              the selected action.
            </DialogDescription>
          </DialogHeader>
          <p className='text-sm'>
            {pendingAction?.label} for {pendingAction?.target_key}
          </p>
          <DialogFooter>
            <Button variant='outline' onClick={() => setPendingAction(null)}>
              Cancel
            </Button>
            <Button
              data-testid='purchase-inbox-confirm-action'
              onClick={confirmPendingAction}
            >
              Confirm
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
