import { useCallback, useMemo, useState } from 'react'
import {
  Inbox,
  FileUp,
  Loader2,
  PackagePlus,
  RefreshCw,
  ShieldCheck,
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

  const updatePackageForm = (field: keyof PackageImportForm, value: string) => {
    setPackageForm((current) => ({ ...current, [field]: value }))
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

  return (
    <>
      <Header>
        <Search />
        <HeaderTitle
          title='Purchase Inbox'
          description='Review captured eBay purchases before linking or converting inventory.'
          icon={Inbox}
          testId='purchase-inbox-header-title'
          iconTestId='purchase-inbox-page-icon'
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
            <h1 className='text-2xl font-bold tracking-tight'>
              Purchase Inbox
            </h1>
            <p className='text-muted-foreground'>
              Captured orders stay review-only until a link or convert action is
              confirmed.
            </p>
          </div>
          <Button
            data-testid='purchase-inbox-load-reviews'
            onClick={() => void loadReviews()}
            disabled={loading}
          >
            {loading ? (
              <Loader2 className='mr-2 h-4 w-4 animate-spin' />
            ) : (
              <RefreshCw className='mr-2 h-4 w-4' />
            )}
            Review captured purchases
          </Button>
        </div>

        <Separator className='my-4' />

        {error ? (
          <section
            className='rounded-md border border-destructive/40 bg-destructive/10 p-4 text-sm'
            data-testid='purchase-inbox-error-state'
          >
            <p className='font-medium'>
              Purchase Inbox could not load reviews.
            </p>
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
              No captured purchases are ready for review.
            </p>
            <p className='mt-1 text-sm text-muted-foreground'>
              Import or capture eBay purchase cards, then prepare review records
              before mutating inventory.
            </p>
          </section>
        ) : null}

        {reviews.length > 0 ? (
          <section
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
                <p className='text-sm text-muted-foreground'>Mutation policy</p>
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
          </section>
        ) : null}

        <Separator className='my-6' />

        <section className='space-y-4' data-testid='forwarder-package-inbox'>
          <div className='flex flex-wrap items-center justify-between gap-3'>
            <div>
              <h2 className='flex items-center gap-2 text-xl font-semibold tracking-tight'>
                <Truck className='h-5 w-5' />
                Forwarder Packages
              </h2>
              <p className='text-sm text-muted-foreground'>
                Import Stackry or freight-forwarder package records before
                matching them to purchases.
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
              Refresh packages
            </Button>
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
                    <Label htmlFor='forwarder-package-provider'>Provider</Label>
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
                    <Label htmlFor='forwarder-package-weight'>Weight g</Label>
                    <Input
                      id='forwarder-package-weight'
                      data-testid='forwarder-package-weight'
                      inputMode='numeric'
                      value={packageForm.weight_grams}
                      onChange={(event) =>
                        updatePackageForm('weight_grams', event.target.value)
                      }
                    />
                  </div>
                </div>
                <div className='grid gap-1.5'>
                  <Label htmlFor='forwarder-package-tracking'>Tracking</Label>
                  <Input
                    id='forwarder-package-tracking'
                    data-testid='forwarder-package-tracking'
                    value={packageForm.tracking_number}
                    onChange={(event) =>
                      updatePackageForm('tracking_number', event.target.value)
                    }
                  />
                </div>
                <div className='grid gap-1.5'>
                  <Label htmlFor='forwarder-package-warehouse'>Warehouse</Label>
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
                  <Label htmlFor='forwarder-package-email'>Package email</Label>
                  <Textarea
                    id='forwarder-package-email'
                    data-testid='forwarder-package-email'
                    className='min-h-32 font-mono text-xs'
                    value={packageEmail}
                    onChange={(event) => setPackageEmail(event.target.value)}
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
                  <p className='font-medium'>Package inbox could not update.</p>
                  <p className='mt-1 text-muted-foreground'>{packageError}</p>
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
                  <p className='font-medium'>No forwarder packages listed.</p>
                  <p className='mt-1 text-sm text-muted-foreground'>
                    Import a package or refresh the current profile inbox.
                  </p>
                </div>
              ) : null}

              {packages.length > 0 ? (
                <div className='space-y-3' data-testid='forwarder-package-list'>
                  {packages.map((pkg) => {
                    const selected = selectedPackageID === pkg.id
                    const rawPayload = pkg.raw_payload
                      ? JSON.stringify(pkg.raw_payload, null, 2)
                      : 'No raw source payload returned for this package.'
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
                      <dl className='mt-3 grid gap-2 text-sm sm:grid-cols-4'>
                        <div>
                          <dt className='text-muted-foreground'>Warehouse</dt>
                          <dd>{pkg.warehouse_location || 'Pending'}</dd>
                        </div>
                        <div>
                          <dt className='text-muted-foreground'>Sender</dt>
                          <dd>{pkg.sender || 'Pending'}</dd>
                        </div>
                        <div>
                          <dt className='text-muted-foreground'>Weight</dt>
                          <dd>{pkg.weight_grams ?? 0} g</dd>
                        </div>
                        <div>
                          <dt className='text-muted-foreground'>Provenance</dt>
                          <dd className='break-all'>{pkg.provenance_key}</dd>
                        </div>
                      </dl>
                      <div className='mt-4 flex justify-end'>
                        <Button
                          type='button'
                          variant='outline'
                          size='sm'
                          data-testid='forwarder-package-detail-toggle'
                          onClick={() =>
                            setSelectedPackageID(selected ? null : pkg.id)
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
                              <dd>{packageDetailValue(pkg.received_at)}</dd>
                            </div>
                            <div>
                              <dt className='text-muted-foreground'>
                                Created
                              </dt>
                              <dd>{packageDetailValue(pkg.created_at)}</dd>
                            </div>
                            <div>
                              <dt className='text-muted-foreground'>
                                Updated
                              </dt>
                              <dd>{packageDetailValue(pkg.updated_at)}</dd>
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
      </Main>

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
