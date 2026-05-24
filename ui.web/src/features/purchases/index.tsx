import { useCallback, useMemo, useState } from 'react'
import { Inbox, Loader2, RefreshCw, ShieldCheck } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Separator } from '@/components/ui/separator'
import { Header, HeaderTitle } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ConfigDrawer } from '@/components/config-drawer'
import { LanguageSwitch } from '@/components/language-switch'
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

function actionTone(action: PurchaseInboxAction) {
  if (action.requires_confirmation) {
    return 'border-amber-300 bg-amber-50 text-amber-950 dark:border-amber-900 dark:bg-amber-950/30 dark:text-amber-100'
  }
  return 'border-border bg-muted/40 text-muted-foreground'
}

function labelForStatus(status: string) {
  return status.split('_').join(' ')
}

export function Purchases() {
  const [reviews, setReviews] = useState<PurchaseInboxReview[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [confirmedAction, setConfirmedAction] = useState<string | null>(null)
  const [pendingAction, setPendingAction] = useState<PurchaseInboxAction | null>(
    null
  )

  const readyItemCount = useMemo(
    () =>
      reviews.reduce(
        (count, review) =>
          count +
          review.items.filter((item) => item.status === 'ready_to_link_or_convert')
            .length,
        0
      ),
    [reviews]
  )

  const loadReviews = useCallback(async () => {
    setLoading(true)
    setError(null)
    setConfirmedAction(null)
    try {
      const response = await fetch('/api/integrations/ebay/purchase-inbox/reviews', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ cards: sampleCards }),
      })
      if (!response.ok) {
        throw new Error('purchase_inbox_reviews_' + response.status)
      }
      const payload = (await response.json()) as ReviewResponse
      setReviews(payload.reviews ?? [])
    } catch (err) {
      setReviews([])
      setError(
        err instanceof Error
          ? err.message
          : 'purchase_inbox_reviews_failed'
      )
    } finally {
      setLoading(false)
    }
  }, [])

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
            <h1 className='text-2xl font-bold tracking-tight'>Purchase Inbox</h1>
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
            <p className='font-medium'>Purchase Inbox could not load reviews.</p>
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
            <p className='font-medium'>No captured purchases are ready for review.</p>
            <p className='mt-1 text-sm text-muted-foreground'>
              Import or capture eBay purchase cards, then prepare review records
              before mutating inventory.
            </p>
          </section>
        ) : null}

        {reviews.length > 0 ? (
          <section className='space-y-4' data-testid='purchase-inbox-ready-state'>
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
