import {
  type ChangeEvent,
  type MouseEvent,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from 'react'
import { getRouteApi } from '@tanstack/react-router'
import {
  ArrowDownAZ,
  ArrowUpAZ,
  PlugZap,
  SlidersHorizontal,
  Store,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
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
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { ConfigDrawer } from '@/components/config-drawer'
import { LanguageSwitch } from '@/components/language-switch'
import { Header, HeaderTitle } from '@/components/layout/header'
import { Main } from '@/components/layout/main'
import { ProfileDropdown } from '@/components/profile-dropdown'
import { Search } from '@/components/search'
import { ThemeSwitch } from '@/components/theme-switch'

const route = getRouteApi('/_authenticated/integrations/')

type AppType = 'all' | 'connected' | 'notConnected'
type ViewMode = 'rows' | 'cards'

type ProviderRecord = {
  provider_id: string
  display_name: string
  base_domain: string
  api_family?: string
  api_support_profile?: string
  integration_mode: 'official_api' | 'web_ingestion' | 'program_api' | string
  auth_mode: 'none' | 'oauth' | 'api_key' | 'hybrid' | string
  state: 'ready' | 'degraded' | 'disabled' | string
  has_token?: boolean
  setup_instructions?: string
  capabilities: {
    search: boolean
    stock_observation: boolean
    pricing: boolean
    health: boolean
  }
  health?: {
    status: 'ok' | 'degraded' | 'down' | 'unknown' | string
    last_checked_at?: string | null
    message?: string
  }
  last_run?: {
    status: 'idle' | 'running' | 'success' | 'failed' | 'never' | string
    finished_at?: string | null
  }
}

type IntegrationForm = {
  baseURL: string
  token: string
  marketplace: string
  itemsPerPage: string
}

type ProfileOption = {
  id: string
  name: string
}

type AppsProps = {
  title?: string
  description?: string
}

const appText = new Map<AppType, string>([
  ['all', 'All Integrations'],
  ['connected', 'Connected'],
  ['notConnected', 'Not Connected'],
])

function providerSettingsKeys(providerID: string) {
  if (providerID === 'ebay') {
    return {
      baseURLKey: 'ebay_base_url',
      tokenKey: 'ebay_bearer_token',
      marketplaceKey: 'ebay_marketplace',
      enabledKey: 'integration.ebay.enabled',
      itemsPerPageKey: 'integration.ebay.items_per_page',
    }
  }
  const slug = providerID.replace(/[^a-z0-9]+/gi, '_').toLowerCase()
  return {
    baseURLKey: `integration.${slug}.base_url`,
    tokenKey: `integration.${slug}.token`,
    marketplaceKey: `integration.${slug}.marketplace`,
    enabledKey: `integration.${slug}.enabled`,
    itemsPerPageKey: `integration.${slug}.items_per_page`,
  }
}

function isConnected(
  provider: ProviderRecord,
  settings: Record<string, string>
) {
  if (provider.auth_mode === 'none') {
    return true
  }
  const keys = providerSettingsKeys(provider.provider_id)
  if (provider.has_token) {
    return true
  }
  return settings[keys.enabledKey] === 'true'
}

function bootstrapErrorMessage(error: unknown) {
  if (!(error instanceof Error)) {
    return 'Failed to bootstrap integrations workspace.'
  }
  if (error.message.startsWith('active_profile_')) {
    return 'Failed to load active profile. Select or create a profile, then retry.'
  }
  if (error.message.startsWith('providers_registry_')) {
    return 'Failed to load provider registry. Retry to reconnect to runtime.'
  }
  if (error.message.startsWith('profile_settings_')) {
    return 'Failed to load profile integration settings. Retry and check runtime logs.'
  }
  return 'Failed to bootstrap integrations workspace.'
}

async function loadProfilesForRecovery() {
  const profilesResp = await fetch('/api/profiles')
  if (!profilesResp.ok) {
    throw new Error(`profiles_${profilesResp.status}`)
  }
  const payload = (await profilesResp.json()) as {
    profiles?: Array<{ id?: string; name?: string }>
  }
  return (payload.profiles ?? [])
    .map((profile) => ({
      id: profile.id?.trim() ?? '',
      name: profile.name?.trim() ?? '',
    }))
    .filter((profile): profile is ProfileOption =>
      Boolean(profile.id && profile.name)
    )
}

export function Apps({
  title = 'Integrations',
  description = 'Configure provider credentials, health checks, and sync actions.',
}: AppsProps = {}) {
  const {
    filter = '',
    type = 'all',
    sort: initSort = 'asc',
    view,
  } = route.useSearch()
  const navigate = route.useNavigate()

  const [sort, setSort] = useState(initSort)
  const [appType, setAppType] = useState(type)
  const [searchTerm, setSearchTerm] = useState(filter)
  const viewMode: ViewMode = view ?? 'cards'
  const [activeProfileId, setActiveProfileId] = useState('')
  const [settings, setSettings] = useState<Record<string, string>>({})
  const [providers, setProviders] = useState<ProviderRecord[]>([])
  const [loading, setLoading] = useState(true)
  const [bootstrapError, setBootstrapError] = useState<string | null>(null)
  const [availableProfiles, setAvailableProfiles] = useState<ProfileOption[]>(
    []
  )
  const [profileRecoveryLoading, setProfileRecoveryLoading] = useState(false)
  const [createProfileName, setCreateProfileName] = useState('')
  const [profileRecoveryError, setProfileRecoveryError] = useState<
    string | null
  >(null)
  const [editingProviderID, setEditingProviderID] = useState<string | null>(
    null
  )
  const [saving, setSaving] = useState(false)
  const [validating, setValidating] = useState(false)
  const [actionMessage, setActionMessage] = useState<string | null>(null)
  const [saveError, setSaveError] = useState<string | null>(null)
  const [replaceToken, setReplaceToken] = useState(false)
  const [rowDetailsProviderID, setRowDetailsProviderID] = useState<
    string | null
  >(null)
  const [rowDetailsOpen, setRowDetailsOpen] = useState(false)
  const [rowEditProviderID, setRowEditProviderID] = useState<string | null>(
    null
  )
  const [rowEditOpen, setRowEditOpen] = useState(false)
  const clickTimerRef = useRef<number | null>(null)
  const [form, setForm] = useState<IntegrationForm>({
    baseURL: '',
    token: '',
    marketplace: '',
    itemsPerPage: '',
  })

  const loadBootstrap = useCallback(async () => {
    setLoading(true)
    setBootstrapError(null)
    setActionMessage(null)
    setProfileRecoveryError(null)
    try {
      const activeResp = await fetch('/api/profiles/active')
      if (!activeResp.ok) {
        throw new Error(`active_profile_${activeResp.status}`)
      }
      const active = (await activeResp.json()) as { id?: string }
      const profileID = (active.id ?? '').trim()
      if (!profileID) {
        throw new Error('active_profile_missing')
      }
      setActiveProfileId(profileID)
      setAvailableProfiles([])

      const registryResp = await fetch('/api/providers/registry')
      if (!registryResp.ok) {
        throw new Error(`providers_registry_${registryResp.status}`)
      }
      const registryPayload = (await registryResp.json()) as {
        providers?: ProviderRecord[]
      }
      setProviders(registryPayload.providers ?? [])

      const settingsResp = await fetch(`/api/profiles/${profileID}/settings`)
      if (!settingsResp.ok) {
        throw new Error(`profile_settings_${settingsResp.status}`)
      }
      const settingsPayload = (await settingsResp.json()) as {
        settings?: Record<string, string>
      }
      setSettings(settingsPayload.settings ?? {})
    } catch (error) {
      if (
        error instanceof Error &&
        error.message.startsWith('active_profile_')
      ) {
        try {
          setAvailableProfiles(await loadProfilesForRecovery())
        } catch {
          setAvailableProfiles([])
        }
      }
      setBootstrapError(bootstrapErrorMessage(error))
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void loadBootstrap()
  }, [loadBootstrap])

  const providerByID = useMemo(
    () =>
      new Map(providers.map((provider) => [provider.provider_id, provider])),
    [providers]
  )
  const editingProvider =
    editingProviderID !== null
      ? (providerByID.get(editingProviderID) ?? null)
      : null
  const rowDetailsProvider =
    rowDetailsProviderID !== null
      ? (providerByID.get(rowDetailsProviderID) ?? null)
      : null
  const rowEditProvider =
    rowEditProviderID !== null
      ? (providerByID.get(rowEditProviderID) ?? null)
      : null

  const sortedProviders = useMemo(() => {
    const list = [...providers]
    list.sort((a, b) =>
      sort === 'asc'
        ? a.display_name.localeCompare(b.display_name)
        : b.display_name.localeCompare(a.display_name)
    )
    return list
  }, [providers, sort])

  const filteredProviders = sortedProviders
    .filter((provider) =>
      appType === 'connected'
        ? isConnected(provider, settings)
        : appType === 'notConnected'
          ? !isConnected(provider, settings)
          : true
    )
    .filter((provider) =>
      provider.display_name.toLowerCase().includes(searchTerm.toLowerCase())
    )

  const handleSearch = (e: ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value
    setSearchTerm(value)
    navigate({
      search: (prev) => ({
        ...prev,
        filter: value || undefined,
      }),
    })
  }

  const recoverWithProfile = useCallback(
    async (profileID: string) => {
      setProfileRecoveryLoading(true)
      setProfileRecoveryError(null)
      try {
        const response = await fetch('/api/profiles/active', {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ profile_id: profileID }),
        })
        if (!response.ok) {
          throw new Error(`set_active_profile_${response.status}`)
        }
        await loadBootstrap()
      } catch (error) {
        setProfileRecoveryError(
          error instanceof Error
            ? 'Could not activate the selected profile. Retry or create a new profile.'
            : 'Could not activate the selected profile.'
        )
      } finally {
        setProfileRecoveryLoading(false)
      }
    },
    [loadBootstrap]
  )

  const createProfileInline = useCallback(async () => {
    const trimmedName = createProfileName.trim()
    if (!trimmedName) {
      setProfileRecoveryError('Enter a profile name before creating one.')
      return
    }

    setProfileRecoveryLoading(true)
    setProfileRecoveryError(null)
    try {
      const createResp = await fetch('/api/profiles', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: trimmedName }),
      })
      if (!createResp.ok) {
        throw new Error(`create_profile_${createResp.status}`)
      }
      const created = (await createResp.json()) as {
        id?: string
        name?: string
      }
      const createdProfileID = created.id?.trim() ?? ''
      if (!createdProfileID) {
        throw new Error('created_profile_missing')
      }
      setCreateProfileName('')
      await recoverWithProfile(createdProfileID)
    } catch (error) {
      setProfileRecoveryError(
        error instanceof Error
          ? 'Could not create a profile inline. Retry and check runtime state.'
          : 'Could not create a profile inline.'
      )
    } finally {
      setProfileRecoveryLoading(false)
    }
  }, [createProfileName, recoverWithProfile])

  const handleTypeChange = (value: AppType) => {
    setAppType(value)
    navigate({
      search: (prev) => ({
        ...prev,
        type: value === 'all' ? undefined : value,
      }),
    })
  }

  const handleSortChange = (nextSort: 'asc' | 'desc') => {
    setSort(nextSort)
    navigate({ search: (prev) => ({ ...prev, sort: nextSort }) })
  }

  const handleViewChange = (nextView: ViewMode) => {
    navigate({
      search: (prev) => ({
        ...prev,
        view: nextView === 'cards' ? undefined : nextView,
      }),
    })
  }

  const openIntegration = (provider: ProviderRecord) => {
    setActionMessage(null)
    setSaveError(null)
    setEditingProviderID(provider.provider_id)
    const keys = providerSettingsKeys(provider.provider_id)
    setForm({
      baseURL:
        settings[keys.baseURLKey] ??
        (provider.base_domain ? `https://${provider.base_domain}` : ''),
      token: '',
      marketplace: settings[keys.marketplaceKey] ?? 'AU',
      itemsPerPage: settings[keys.itemsPerPageKey] ?? '24',
    })
    setReplaceToken(!provider.has_token)
  }

  useEffect(() => {
    return () => {
      if (clickTimerRef.current !== null) {
        window.clearTimeout(clickTimerRef.current)
      }
    }
  }, [])

  const updateSelectedProviderContext = (providerID: string) => {
    if (typeof window !== 'undefined') {
      const url = new URL(window.location.href)
      url.searchParams.set('selected', providerID)
      window.history.replaceState({}, '', url.toString())
    }
  }

  const isInteractiveTarget = (target: EventTarget | null) => {
    if (!(target instanceof HTMLElement)) {
      return false
    }
    return Boolean(
      target.closest('button, a, input, select, textarea, [role="button"]')
    )
  }

  const handleRowClick = (
    provider: ProviderRecord,
    event: MouseEvent<HTMLElement>
  ) => {
    if (isInteractiveTarget(event.target)) {
      return
    }
    if (clickTimerRef.current !== null) {
      window.clearTimeout(clickTimerRef.current)
    }
    clickTimerRef.current = window.setTimeout(() => {
      setRowDetailsProviderID(provider.provider_id)
      setRowEditProviderID(null)
      setRowEditOpen(false)
      setRowDetailsOpen(true)
      updateSelectedProviderContext(provider.provider_id)
    }, 180)
  }

  const handleRowDoubleClick = (
    provider: ProviderRecord,
    event: MouseEvent<HTMLElement>
  ) => {
    if (isInteractiveTarget(event.target)) {
      return
    }
    if (clickTimerRef.current !== null) {
      window.clearTimeout(clickTimerRef.current)
      clickTimerRef.current = null
    }
    setRowDetailsProviderID(null)
    setRowDetailsOpen(false)
    setRowEditProviderID(provider.provider_id)
    setRowEditOpen(true)
    updateSelectedProviderContext(provider.provider_id)
  }

  const closeIntegration = () => {
    setEditingProviderID(null)
    setSaveError(null)
    setReplaceToken(false)
    setForm({
      baseURL: '',
      token: '',
      marketplace: '',
      itemsPerPage: '',
    })
  }

  const saveIntegration = async () => {
    if (!activeProfileId || !editingProvider) {
      return
    }
    setSaving(true)
    setSaveError(null)
    setActionMessage(null)
    try {
      const keys = providerSettingsKeys(editingProvider.provider_id)
      const trimmedToken = form.token.trim()
      if (
        editingProvider.auth_mode !== 'none' &&
        !editingProvider.has_token &&
        trimmedToken === ''
      ) {
        throw new Error('token_required_for_provider')
      }

      const next: Record<string, string> = {
        [keys.baseURLKey]: form.baseURL.trim(),
        [keys.marketplaceKey]: form.marketplace.trim(),
        [keys.itemsPerPageKey]: form.itemsPerPage.trim(),
        [keys.enabledKey]:
          editingProvider.auth_mode === 'none' ||
          editingProvider.has_token ||
          trimmedToken !== ''
            ? 'true'
            : 'false',
      }

      if (replaceToken && trimmedToken !== '') {
        next[keys.tokenKey] = trimmedToken
      }

      const response = await fetch(
        `/api/profiles/${activeProfileId}/settings`,
        {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ settings: next }),
        }
      )
      if (!response.ok) {
        throw new Error(`save_failed_${response.status}`)
      }
      const payload = (await response.json()) as {
        settings?: Record<string, string>
      }
      setSettings(payload.settings ?? {})
      setProviders((prev) =>
        prev.map((provider) =>
          provider.provider_id === editingProvider.provider_id
            ? {
                ...provider,
                has_token: provider.has_token || trimmedToken !== '',
              }
            : provider
        )
      )
      setActionMessage('Provider configuration saved.')
      closeIntegration()
    } catch (err) {
      if (
        err instanceof Error &&
        err.message === 'token_required_for_provider'
      ) {
        setSaveError('Token is required for this provider.')
      } else {
        setSaveError(err instanceof Error ? err.message : 'save_failed')
      }
    } finally {
      setSaving(false)
    }
  }

  const validateProvider = async () => {
    if (!editingProvider) {
      return
    }
    setValidating(true)
    setSaveError(null)
    try {
      const response = await fetch(
        `/api/provider/health?provider=${encodeURIComponent(editingProvider.provider_id)}`
      )
      if (!response.ok) {
        throw new Error(`validate_failed_${response.status}`)
      }
      const payload = (await response.json()) as {
        status?: string
        message?: string
        updated_at?: string
      }
      setProviders((prev) =>
        prev.map((provider) =>
          provider.provider_id === editingProvider.provider_id
            ? {
                ...provider,
                health: {
                  status: payload.status ?? 'unknown',
                  message: payload.message,
                  last_checked_at: payload.updated_at ?? null,
                },
                last_run: {
                  status:
                    payload.status === 'ok'
                      ? 'success'
                      : payload.status === 'unknown'
                        ? 'never'
                        : 'failed',
                  finished_at: payload.updated_at ?? null,
                },
              }
            : provider
        )
      )
      setActionMessage(`Validated ${editingProvider.display_name} health.`)
    } catch (error) {
      setSaveError(error instanceof Error ? error.message : 'validate_failed')
    } finally {
      setValidating(false)
    }
  }

  return (
    <>
      <Header>
        <Search />
        <HeaderTitle
          title={title}
          description={description}
          icon={PlugZap}
          testId='integrations-header-title'
          iconTestId='integrations-page-icon'
        />
        <div
          className='ms-auto flex items-center gap-4'
          data-header-title-avoid='true'
        >
          <LanguageSwitch />
          <ThemeSwitch />
          <ConfigDrawer />
          <ProfileDropdown />
        </div>
      </Header>

      <Main fixed>
        <div>
          <h1 className='text-2xl font-bold tracking-tight'>{title}</h1>
          <p className='text-muted-foreground'>{description}</p>
        </div>

        {bootstrapError ? (
          <div
            className='rounded-md border border-destructive/40 bg-destructive/10 p-4 text-sm'
            data-testid='integrations-bootstrap-error'
          >
            <p className='font-medium'>Integrations bootstrap failed</p>
            <p className='mt-1 text-muted-foreground'>{bootstrapError}</p>
            <div className='mt-3 flex flex-wrap gap-2'>
              <Button
                size='sm'
                variant='outline'
                onClick={() => void loadBootstrap()}
              >
                Retry
              </Button>
            </div>

            {bootstrapError.includes('active profile') ? (
              <div
                className='mt-4 rounded-md border border-border/60 bg-background/70 p-3'
                data-testid='integrations-profile-recovery'
              >
                <p className='font-medium'>
                  Recover profile context in this route
                </p>
                <p className='mt-1 text-muted-foreground'>
                  Pick an existing profile or create one here, then reload the
                  provider catalog in place.
                </p>

                {availableProfiles.length ? (
                  <div className='mt-3 flex flex-wrap gap-2'>
                    {availableProfiles.map((profile) => (
                      <Button
                        key={profile.id}
                        type='button'
                        size='sm'
                        variant='secondary'
                        data-testid={`integrations-recovery-profile-${profile.id}`}
                        disabled={profileRecoveryLoading}
                        onClick={() => void recoverWithProfile(profile.id)}
                      >
                        Use {profile.name}
                      </Button>
                    ))}
                  </div>
                ) : (
                  <p
                    className='mt-3 text-muted-foreground'
                    data-testid='integrations-recovery-no-profiles'
                  >
                    No selectable profiles were found. Create one below.
                  </p>
                )}

                <div className='mt-3 flex flex-col gap-2 sm:flex-row'>
                  <Input
                    value={createProfileName}
                    onChange={(event) =>
                      setCreateProfileName(event.target.value)
                    }
                    placeholder='New profile name'
                    data-testid='integrations-recovery-create-input'
                    disabled={profileRecoveryLoading}
                  />
                  <Button
                    type='button'
                    size='sm'
                    data-testid='integrations-recovery-create-submit'
                    disabled={profileRecoveryLoading}
                    onClick={() => void createProfileInline()}
                  >
                    Create profile
                  </Button>
                </div>

                {profileRecoveryError ? (
                  <p
                    className='mt-2 text-destructive'
                    data-testid='integrations-recovery-error'
                  >
                    {profileRecoveryError}
                  </p>
                ) : null}
              </div>
            ) : null}
          </div>
        ) : null}

        {actionMessage ? (
          <div className='rounded-md border bg-muted/30 px-3 py-2 text-sm'>
            {actionMessage}
          </div>
        ) : null}

        <div className='my-4 flex items-end justify-between sm:my-0 sm:items-center'>
          <div className='flex flex-col gap-4 sm:my-4 sm:flex-row'>
            <Input
              placeholder='Filter providers...'
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

        <div className='flex items-center justify-end gap-2 pt-4'>
          <Button
            size='sm'
            variant={viewMode === 'rows' ? 'default' : 'outline'}
            onClick={() => handleViewChange('rows')}
            aria-pressed={viewMode === 'rows'}
            aria-label='Switch to rows view'
          >
            Rows
          </Button>
          <Button
            size='sm'
            variant={viewMode === 'cards' ? 'default' : 'outline'}
            onClick={() => handleViewChange('cards')}
            aria-pressed={viewMode === 'cards'}
            aria-label='Switch to cards view'
          >
            Cards
          </Button>
        </div>

        {loading ? (
          <div className='rounded-md border p-6 text-sm text-muted-foreground'>
            Loading providers...
          </div>
        ) : null}

        {!loading && viewMode === 'rows' ? (
          <div className='overflow-hidden rounded-md border'>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Provider</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Health</TableHead>
                  <TableHead>Description</TableHead>
                  <TableHead className='text-right'>Action</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredProviders.length ? (
                  filteredProviders.map((provider) => (
                    <TableRow
                      key={provider.provider_id}
                      onClick={(event) => handleRowClick(provider, event)}
                      onDoubleClick={(event) =>
                        handleRowDoubleClick(provider, event)
                      }
                    >
                      <TableCell>
                        <div className='flex items-center gap-2'>
                          <div className='flex size-8 items-center justify-center rounded-md bg-muted p-1.5'>
                            <Store className='size-4' />
                          </div>
                          <span className='font-medium'>
                            {provider.display_name}
                          </span>
                        </div>
                      </TableCell>
                      <TableCell>
                        {isConnected(provider, settings)
                          ? 'Connected'
                          : 'Not Connected'}
                      </TableCell>
                      <TableCell>
                        {provider.health?.status ?? 'unknown'}
                      </TableCell>
                      <TableCell className='text-muted-foreground'>
                        {provider.integration_mode}
                      </TableCell>
                      <TableCell className='text-right'>
                        <Button
                          variant='outline'
                          size='sm'
                          onClick={() => openIntegration(provider)}
                        >
                          {isConnected(provider, settings) ? 'Edit' : 'Connect'}
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))
                ) : (
                  <TableRow>
                    <TableCell colSpan={5} className='h-24 text-center'>
                      No integrations match current filters.
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </div>
        ) : null}

        {!loading && viewMode === 'cards' ? (
          <ul className='faded-bottom no-scrollbar grid gap-4 overflow-auto pb-16 md:grid-cols-2 lg:grid-cols-3'>
            {filteredProviders.map((provider) => (
              <li
                key={provider.provider_id}
                className='rounded-lg border p-4 hover:shadow-md'
                data-testid={`provider-card-${provider.provider_id}`}
              >
                <div className='mb-6 flex items-start justify-between gap-3'>
                  <div className='space-y-1'>
                    <h2 className='font-semibold'>{provider.display_name}</h2>
                    <p className='text-xs text-muted-foreground'>
                      {provider.base_domain}
                    </p>
                    <p
                      className='text-xs text-muted-foreground'
                      data-testid={`provider-api-family-${provider.provider_id}`}
                    >
                      API Family: {provider.api_family ?? 'custom'}
                    </p>
                  </div>
                  <Button
                    variant='outline'
                    size='sm'
                    data-testid={`provider-open-${provider.provider_id}`}
                    onClick={() => openIntegration(provider)}
                  >
                    {isConnected(provider, settings) ? 'Edit' : 'Connect'}
                  </Button>
                </div>
                <div className='space-y-1 text-xs text-muted-foreground'>
                  <p>Status: {provider.state}</p>
                  <p>Health: {provider.health?.status ?? 'unknown'}</p>
                  <p>Last run: {provider.last_run?.status ?? 'never'}</p>
                </div>
                <div className='mt-3 flex flex-wrap gap-1 text-[11px] text-muted-foreground'>
                  {provider.capabilities.search ? (
                    <span className='rounded bg-muted px-2 py-0.5'>Search</span>
                  ) : null}
                  {provider.capabilities.stock_observation ? (
                    <span className='rounded bg-muted px-2 py-0.5'>Stock</span>
                  ) : null}
                  {provider.capabilities.pricing ? (
                    <span className='rounded bg-muted px-2 py-0.5'>
                      Pricing
                    </span>
                  ) : null}
                </div>
              </li>
            ))}
          </ul>
        ) : null}
      </Main>

      <Dialog
        open={editingProvider !== null}
        onOpenChange={(open) => {
          if (!open) {
            closeIntegration()
          }
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {editingProvider?.display_name ?? 'Integration'}
            </DialogTitle>
            <DialogDescription>
              Manage provider credentials, validation, and setup controls.
            </DialogDescription>
          </DialogHeader>
          {editingProvider ? (
            <div className='space-y-4'>
              <div className='rounded-md border bg-muted/20 p-3 text-xs'>
                <p>Mode: {editingProvider.integration_mode}</p>
                <p data-testid='provider-detail-api-family'>
                  API Family: {editingProvider.api_family ?? 'custom'}
                </p>
                <p data-testid='provider-detail-api-support-profile'>
                  Support Profile:{' '}
                  {editingProvider.api_support_profile ?? 'unknown'}
                </p>
                <p>Health: {editingProvider.health?.status ?? 'unknown'}</p>
                <p>Last run: {editingProvider.last_run?.status ?? 'never'}</p>
                <p>
                  Last checked:{' '}
                  {editingProvider.health?.last_checked_at ??
                    editingProvider.last_run?.finished_at ??
                    'n/a'}
                </p>
              </div>

              <div className='rounded-md border p-3 text-xs text-muted-foreground'>
                {editingProvider.setup_instructions ??
                  'Enter provider details, validate health, and save configuration.'}
              </div>

              <Input
                placeholder='Base URL'
                value={form.baseURL}
                onChange={(e) =>
                  setForm((prev) => ({ ...prev, baseURL: e.target.value }))
                }
              />
              <Input
                placeholder='Marketplace / Region'
                value={form.marketplace}
                onChange={(e) =>
                  setForm((prev) => ({ ...prev, marketplace: e.target.value }))
                }
              />
              <Input
                type='number'
                min='1'
                placeholder='Items per page'
                value={form.itemsPerPage}
                onChange={(e) =>
                  setForm((prev) => ({ ...prev, itemsPerPage: e.target.value }))
                }
              />

              {editingProvider.auth_mode !== 'none' ? (
                <div className='space-y-2'>
                  {editingProvider.has_token && !replaceToken ? (
                    <div className='rounded-md border bg-muted/20 p-3 text-xs'>
                      <p>Token on file.</p>
                      <p className='text-muted-foreground'>
                        Existing token is hidden. Use replace-token to update
                        it.
                      </p>
                      <Button
                        size='sm'
                        variant='outline'
                        className='mt-2'
                        data-testid='replace-token'
                        onClick={() => setReplaceToken(true)}
                      >
                        Replace Token
                      </Button>
                    </div>
                  ) : (
                    <Input
                      type='password'
                      data-testid='provider-token-input'
                      placeholder='New token / API key'
                      value={form.token}
                      onChange={(e) =>
                        setForm((prev) => ({ ...prev, token: e.target.value }))
                      }
                    />
                  )}
                </div>
              ) : null}

              {saveError ? (
                <p className='text-sm text-destructive'>{saveError}</p>
              ) : null}

              <div className='flex flex-wrap justify-end gap-2'>
                <Button
                  variant='outline'
                  onClick={() => void validateProvider()}
                  disabled={validating}
                >
                  {validating ? 'Validating...' : 'Validate'}
                </Button>
                <div className='flex max-w-xs flex-col items-end gap-1'>
                  <Button
                    variant='outline'
                    disabled
                    aria-describedby='provider-sync-unavailable'
                  >
                    Sync
                  </Button>
                  <p
                    id='provider-sync-unavailable'
                    className='text-right text-xs text-muted-foreground'
                  >
                    Sync runs from Market Watch query sets.
                  </p>
                </div>
                <Button
                  variant='outline'
                  onClick={closeIntegration}
                  disabled={saving}
                >
                  Cancel
                </Button>
                <Button
                  onClick={() => void saveIntegration()}
                  disabled={saving}
                >
                  {saving ? 'Saving...' : 'Save Integration'}
                </Button>
              </div>
            </div>
          ) : null}
        </DialogContent>
      </Dialog>
      <Dialog open={rowDetailsOpen} onOpenChange={setRowDetailsOpen}>
        <DialogContent data-testid='integrations-row-details-modal'>
          <DialogHeader>
            <DialogTitle>Integration Details</DialogTitle>
            <DialogDescription>
              Opened from row single-click interaction.
            </DialogDescription>
          </DialogHeader>
          {rowDetailsProvider ? (
            <div className='space-y-1 text-sm'>
              <p>
                <strong>Provider:</strong> {rowDetailsProvider.display_name}
              </p>
              <p>
                <strong>State:</strong> {rowDetailsProvider.state}
              </p>
              <p>
                <strong>Health:</strong>{' '}
                {rowDetailsProvider.health?.status ?? 'unknown'}
              </p>
            </div>
          ) : null}
        </DialogContent>
      </Dialog>
      <Dialog open={rowEditOpen} onOpenChange={setRowEditOpen}>
        <DialogContent data-testid='integrations-row-edit-modal'>
          <DialogHeader>
            <DialogTitle>Edit Integration</DialogTitle>
            <DialogDescription>
              Opened from row double-click interaction.
            </DialogDescription>
          </DialogHeader>
          {rowEditProvider ? (
            <div className='space-y-1 text-sm'>
              <p>
                <strong>Provider:</strong> {rowEditProvider.display_name}
              </p>
              <p>
                <strong>Mode:</strong> {rowEditProvider.integration_mode}
              </p>
            </div>
          ) : null}
        </DialogContent>
      </Dialog>
    </>
  )
}
