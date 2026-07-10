import { useCallback, useEffect, useMemo, useState } from 'react'
import { RefreshCw, Upload } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { ContentSection } from '../components/content-section'

type EffectiveSession = {
  plan?: string
  features?: string[]
}

type LicenseStatus = {
  state?: string
  tier?: string
  features?: string[]
  expires_at?: string
}

type ActiveProfile = {
  id?: string
}

type SignedLicenseFile = {
  payload_base64?: string
  signature_base64?: string
}

function planLabel(plan?: string) {
  const normalized = (plan || 'free').trim().toLowerCase()
  if (normalized === 'plus') return 'Plus'
  if (normalized === 'pro') return 'Pro'
  return 'Free'
}

function planSummary(plan?: string) {
  switch ((plan || 'free').trim().toLowerCase()) {
    case 'plus':
      return 'Plus includes unlimited inventory items, Market Watch, scanner automation, and export access.'
    case 'pro':
      return 'Pro includes unlimited inventory items, Market Watch, Discoveries, Assistant, and export access.'
    default:
      return 'Free includes 250 inventory items and export access.'
  }
}

function licenseSummary(status: LicenseStatus | null) {
  if (!status || status.state === 'free') {
    return 'No signed license imported'
  }
  if (status.state === 'valid') {
    return `${planLabel(status.tier)} license active`
  }
  if (status.state === 'expired') {
    return 'Signed license expired'
  }
  return 'Signed license needs review'
}

function parseLicense(raw: string): SignedLicenseFile | null {
  try {
    const parsed = JSON.parse(raw) as unknown
    const root =
      typeof parsed === 'object' && parsed !== null
        ? (parsed as Record<string, unknown>)
        : null
    const nested =
      root && typeof root.license === 'object' && root.license !== null
        ? (root.license as Record<string, unknown>)
        : null
    const candidate = nested ?? root
    const payloadBase64 = candidate?.payload_base64
    const signatureBase64 = candidate?.signature_base64
    if (
      typeof payloadBase64 === 'string' &&
      typeof signatureBase64 === 'string' &&
      payloadBase64.trim() &&
      signatureBase64.trim()
    ) {
      return {
        payload_base64: payloadBase64,
        signature_base64: signatureBase64,
      }
    }
  } catch {
    return null
  }
  return null
}

export function SettingsBilling() {
  const { t } = useTranslation('pages')
  const [profileID, setProfileID] = useState('')
  const [session, setSession] = useState<EffectiveSession | null>(null)
  const [license, setLicense] = useState<LicenseStatus | null>(null)
  const [loading, setLoading] = useState(true)
  const [licenseInput, setLicenseInput] = useState('')
  const [importing, setImporting] = useState(false)
  const [status, setStatus] = useState<string | null>(null)

  const signedLicense = useMemo(
    () => parseLicense(licenseInput),
    [licenseInput]
  )

  const loadBillingState = useCallback(async () => {
    setLoading(true)
    setStatus(null)
    try {
      const [sessionResp, activeResp] = await Promise.all([
        fetch('/api/auth/cloud/session/effective'),
        fetch('/api/profiles/active'),
      ])
      if (sessionResp.ok) {
        setSession((await sessionResp.json()) as EffectiveSession)
      }
      let nextProfileID = ''
      if (activeResp.ok) {
        const active = (await activeResp.json()) as ActiveProfile
        nextProfileID = active.id?.trim() ?? ''
        setProfileID(nextProfileID)
      }
      const licenseResp = await fetch(
        `/api/license/status?profile_id=${encodeURIComponent(nextProfileID)}`
      )
      if (licenseResp.ok) {
        setLicense((await licenseResp.json()) as LicenseStatus)
      }
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void loadBillingState()
  }, [loadBillingState])

  const importLicense = useCallback(async () => {
    if (!signedLicense || !profileID) return
    setImporting(true)
    setStatus(null)
    try {
      const resp = await fetch('/api/license/import', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ profile_id: profileID, license: signedLicense }),
      })
      if (!resp.ok) {
        throw new Error('license_import_failed')
      }
      setLicenseInput('')
      setStatus('Signed license imported')
      await loadBillingState()
    } catch {
      setStatus('Signed license could not be imported')
    } finally {
      setImporting(false)
    }
  }, [loadBillingState, profileID, signedLicense])

  const currentPlan = planLabel(session?.plan)
  const currentPlanSummary = planSummary(session?.plan)

  return (
    <ContentSection
      title={t('settings.billing.title')}
      desc={t('settings.billing.description')}
    >
      <div className='space-y-4 text-sm'>
        <div
          className='rounded-md border p-3'
          data-testid='billing-plan-card'
          aria-busy={loading}
        >
          <div className='flex flex-wrap items-center justify-between gap-2'>
            <p className='font-medium'>Plan</p>
            <Badge variant='outline'>Current plan</Badge>
          </div>
          <p className='mt-2 text-base font-semibold'>{currentPlan}</p>
          <p className='text-muted-foreground'>
            Source: Cloud entitlement state
          </p>
          <p className='mt-2 text-muted-foreground'>{currentPlanSummary}</p>
          <div className='mt-3 flex flex-wrap gap-2 text-xs text-muted-foreground'>
            {(session?.features ?? ['collection_core']).map((feature) => (
              <span key={feature} className='rounded-md border px-2 py-1'>
                {feature.replace(/_/g, ' ')}
              </span>
            ))}
          </div>
        </div>
        <div
          className='rounded-md border p-3'
          data-testid='billing-license-card'
        >
          <div className='flex flex-wrap items-center justify-between gap-2'>
            <p className='font-medium'>License Status</p>
            <Badge variant='outline'>Founding license</Badge>
          </div>
          <p className='mt-2 text-base font-semibold'>
            {licenseSummary(license)}
          </p>
          <p className='text-muted-foreground'>
            Paste a signed founding license payload below.
          </p>
          {license?.expires_at ? (
            <p className='mt-2 text-muted-foreground'>
              Expires: {license.expires_at}
            </p>
          ) : null}
          <Textarea
            className='mt-3 min-h-24 font-mono text-xs'
            data-testid='founding-license-import'
            value={licenseInput}
            onChange={(event) => setLicenseInput(event.target.value)}
            placeholder='{"payload_base64":"...","signature_base64":"..."}'
            aria-label='Signed founding license JSON'
          />
          <div className='mt-3 flex flex-wrap gap-2'>
            <Button
              variant='outline'
              size='sm'
              onClick={() => void importLicense()}
              disabled={!signedLicense || !profileID || importing}
            >
              <Upload />
              Import signed license
            </Button>
            <Button
              variant='ghost'
              size='sm'
              onClick={() => void loadBillingState()}
              disabled={loading}
            >
              <RefreshCw />
              Refresh
            </Button>
          </div>
          {status ? (
            <p className='mt-2 text-muted-foreground'>{status}</p>
          ) : null}
        </div>
        <Button variant='outline' size='sm' disabled>
          Open Billing Portal (Coming soon)
        </Button>
      </div>
    </ContentSection>
  )
}
