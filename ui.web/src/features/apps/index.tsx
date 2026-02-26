import { type ChangeEvent, useEffect, useMemo, useState } from 'react'
import { getRouteApi } from '@tanstack/react-router'
import { SlidersHorizontal, ArrowUpAZ, ArrowDownAZ } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Separator } from '@/components/ui/separator'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { ConfigDrawer } from '@/components/config-drawer'
import { Header } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'
import { apps } from './data/apps'

const route = getRouteApi('/_authenticated/integrations/')

type AppType = 'all' | 'connected' | 'notConnected'
type IntegrationForm = {
  baseURL: string
  token: string
  marketplace: string
}

const appText = new Map<AppType, string>([
  ['all', 'All Integrations'],
  ['connected', 'Connected'],
  ['notConnected', 'Not Connected'],
])

type AppsProps = {
  title?: string
  description?: string
}

export function Apps({
  title = 'App Integrations',
  description = "Here's a list of your apps for the integration!",
}: AppsProps = {}) {
  const {
    filter = '',
    type = 'all',
    sort: initSort = 'asc',
  } = route.useSearch()
  const navigate = route.useNavigate()

  const [sort, setSort] = useState(initSort)
  const [appType, setAppType] = useState(type)
  const [searchTerm, setSearchTerm] = useState(filter)
  const [activeProfileId, setActiveProfileId] = useState('')
  const [settings, setSettings] = useState<Record<string, string>>({})
  const [editingIntegration, setEditingIntegration] = useState<string | null>(null)
  const [saving, setSaving] = useState(false)
  const [saveError, setSaveError] = useState<string | null>(null)
  const [form, setForm] = useState<IntegrationForm>({
    baseURL: '',
    token: '',
    marketplace: '',
  })

  useEffect(() => {
    const load = async () => {
      try {
        const activeResp = await fetch('/api/profiles/active')
        if (!activeResp.ok) {
          return
        }
        const active = (await activeResp.json()) as { id?: string }
        const profileID = active.id ?? ''
        if (!profileID) {
          return
        }
        setActiveProfileId(profileID)
        const settingsResp = await fetch(`/api/profiles/${profileID}/settings`)
        if (!settingsResp.ok) {
          return
        }
        const payload = (await settingsResp.json()) as {
          settings?: Record<string, string>
        }
        setSettings(payload.settings ?? {})
      } catch {
        // Keep integrations available even if settings load fails.
      }
    }
    void load()
  }, [])

  const resolvedApps = useMemo(() => {
    const ebayConnected =
      Boolean(settings['ebay_bearer_token']) || settings['integration.ebay.enabled'] === 'true'
    return apps.map((app) =>
      app.name === 'eBay'
        ? {
            ...app,
            connected: ebayConnected,
          }
        : app
    )
  }, [settings])

  const filteredApps = resolvedApps
    .sort((a, b) =>
      sort === 'asc'
        ? a.name.localeCompare(b.name)
        : b.name.localeCompare(a.name)
    )
    .filter((app) =>
      appType === 'connected'
        ? app.connected
        : appType === 'notConnected'
          ? !app.connected
          : true
    )
    .filter((app) => app.name.toLowerCase().includes(searchTerm.toLowerCase()))

  const handleSearch = (e: ChangeEvent<HTMLInputElement>) => {
    setSearchTerm(e.target.value)
    navigate({
      search: (prev) => ({
        ...prev,
        filter: e.target.value || undefined,
      }),
    })
  }

  const handleTypeChange = (value: AppType) => {
    setAppType(value)
    navigate({
      search: (prev) => ({
        ...prev,
        type: value === 'all' ? undefined : value,
      }),
    })
  }

  const handleSortChange = (sort: 'asc' | 'desc') => {
    setSort(sort)
    navigate({ search: (prev) => ({ ...prev, sort }) })
  }

  const openIntegration = (name: string) => {
    setSaveError(null)
    setEditingIntegration(name)
    if (name === 'eBay') {
      setForm({
        baseURL: settings['ebay_base_url'] ?? '',
        token: settings['ebay_bearer_token'] ?? '',
        marketplace: settings['ebay_marketplace'] ?? 'EBAY-US',
      })
      return
    }
    const slug = name.toLowerCase().replace(/[^a-z0-9]+/g, '_')
    setForm({
      baseURL: settings[`integration.${slug}.base_url`] ?? '',
      token: settings[`integration.${slug}.token`] ?? '',
      marketplace: settings[`integration.${slug}.marketplace`] ?? '',
    })
  }

  const saveIntegration = async () => {
    if (!activeProfileId || !editingIntegration) {
      return
    }
    setSaving(true)
    setSaveError(null)
    try {
      const slug = editingIntegration.toLowerCase().replace(/[^a-z0-9]+/g, '_')
      const next: Record<string, string> = {
        ...settings,
      }
      if (editingIntegration === 'eBay') {
        next['ebay_base_url'] = form.baseURL
        next['ebay_bearer_token'] = form.token
        next['ebay_marketplace'] = form.marketplace
        next['integration.ebay.enabled'] = String(Boolean(form.token))
      } else {
        next[`integration.${slug}.base_url`] = form.baseURL
        next[`integration.${slug}.token`] = form.token
        next[`integration.${slug}.marketplace`] = form.marketplace
        next[`integration.${slug}.enabled`] = String(Boolean(form.token))
      }
      const response = await fetch(`/api/profiles/${activeProfileId}/settings`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ settings: next }),
      })
      if (!response.ok) {
        throw new Error(`save_failed_${response.status}`)
      }
      const payload = (await response.json()) as { settings?: Record<string, string> }
      setSettings(payload.settings ?? next)
      setEditingIntegration(null)
    } catch (err) {
      setSaveError(err instanceof Error ? err.message : 'save_failed')
    } finally {
      setSaving(false)
    }
  }

  return (
    <>
      {/* ===== Top Heading ===== */}
      <Header>
        <Search />
        <div className='ms-auto flex items-center gap-4'>
          <ThemeSwitch />
          <ConfigDrawer />
          <ProfileDropdown />
        </div>
      </Header>

      {/* ===== Content ===== */}
      <Main fixed>
        <div>
          <h1 className='text-2xl font-bold tracking-tight'>{title}</h1>
          <p className='text-muted-foreground'>{description}</p>
        </div>
        <div className='my-4 flex items-end justify-between sm:my-0 sm:items-center'>
          <div className='flex flex-col gap-4 sm:my-4 sm:flex-row'>
            <Input
              placeholder='Filter apps...'
              className='h-9 w-40 lg:w-[250px]'
              value={searchTerm}
              onChange={handleSearch}
            />
            <Select value={appType} onValueChange={handleTypeChange}>
              <SelectTrigger className='w-36'>
                <SelectValue>{appText.get(appType)}</SelectValue>
              </SelectTrigger>
              <SelectContent>
              <SelectItem value='all'>All Integrations</SelectItem>
                <SelectItem value='connected'>Connected</SelectItem>
                <SelectItem value='notConnected'>Not Connected</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <Select value={sort} onValueChange={handleSortChange}>
            <SelectTrigger className='w-16'>
              <SelectValue>
                <SlidersHorizontal size={18} />
              </SelectValue>
            </SelectTrigger>
            <SelectContent align='end'>
              <SelectItem value='asc'>
                <div className='flex items-center gap-4'>
                  <ArrowUpAZ size={16} />
                  <span>Ascending</span>
                </div>
              </SelectItem>
              <SelectItem value='desc'>
                <div className='flex items-center gap-4'>
                  <ArrowDownAZ size={16} />
                  <span>Descending</span>
                </div>
              </SelectItem>
            </SelectContent>
          </Select>
        </div>
        <Separator className='shadow-sm' />
        <ul className='faded-bottom no-scrollbar grid gap-4 overflow-auto pt-4 pb-16 md:grid-cols-2 lg:grid-cols-3'>
          {filteredApps.map((app) => (
            <li
              key={app.name}
              className='rounded-lg border p-4 hover:shadow-md'
            >
              <div className='mb-8 flex items-center justify-between'>
                <div
                  className={`flex size-10 items-center justify-center rounded-lg bg-muted p-2`}
                >
                  {app.logo}
                </div>
                <Button
                  variant='outline'
                  size='sm'
                  className={`${app.connected ? 'border border-blue-300 bg-blue-50 hover:bg-blue-100 dark:border-blue-700 dark:bg-blue-950 dark:hover:bg-blue-900' : ''}`}
                  onClick={() => openIntegration(app.name)}
                >
                  {app.connected ? 'Edit' : 'Connect'}
                </Button>
              </div>
              <div>
                <h2 className='mb-1 font-semibold'>{app.name}</h2>
                <p className='line-clamp-2 text-gray-500'>{app.desc}</p>
              </div>
            </li>
          ))}
        </ul>
      </Main>

      <Dialog
        open={editingIntegration !== null}
        onOpenChange={(open) => {
          if (!open) {
            setEditingIntegration(null)
          }
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editingIntegration ?? 'Integration'}</DialogTitle>
            <DialogDescription>
              Update provider credentials and connection details.
            </DialogDescription>
          </DialogHeader>
          <div className='space-y-3'>
            <Input
              placeholder='Base URL'
              value={form.baseURL}
              onChange={(e) =>
                setForm((prev) => ({ ...prev, baseURL: e.target.value }))
              }
            />
            <Input
              placeholder='Token / API key'
              value={form.token}
              onChange={(e) =>
                setForm((prev) => ({ ...prev, token: e.target.value }))
              }
            />
            <Input
              placeholder='Marketplace / Region'
              value={form.marketplace}
              onChange={(e) =>
                setForm((prev) => ({ ...prev, marketplace: e.target.value }))
              }
            />
            {saveError ? (
              <p className='text-sm text-destructive'>{saveError}</p>
            ) : null}
            <div className='flex justify-end gap-2'>
              <Button
                variant='outline'
                onClick={() => setEditingIntegration(null)}
                disabled={saving}
              >
                Cancel
              </Button>
              <Button onClick={() => void saveIntegration()} disabled={saving}>
                {saving ? 'Saving...' : 'Save Integration'}
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </>
  )
}
