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
  Download,
  PlugZap,
  Plus,
  SearchCheck,
  SlidersHorizontal,
  Store,
} from 'lucide-react'
import { recordNotificationHistory } from '@/lib/toast-history'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
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
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
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
  provider_category?: string
  provider_type?: string
  api_family?: string
  api_support_profile?: string
  config_schema_ref?: string
  workflow_refs?: string[]
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
    assistant?: boolean
    image_help?: boolean
    content_generation?: boolean
    media_capture?: boolean
    text_capture?: boolean
  }
  active_auth_method?: 'api_key' | 'browser_auth' | string
  auth_methods?: {
    api_key?: {
      state?: string
      connected?: boolean
      credential_present?: boolean
    }
    browser_auth?: {
      state?: string
      connected?: boolean
      credential_present?: boolean
      setup_message?: string
    }
    sender_chat?: {
      state?: string
      connected?: boolean
      credential_present?: boolean
      setup_message?: string
    }
    bot_token?: {
      state?: string
      connected?: boolean
      credential_present?: boolean
      setup_message?: string
    }
    webhook?: {
      state?: string
      connected?: boolean
      setup_message?: string
    }
  }
  model_options?: string[]
  health?: {
    status: 'ok' | 'degraded' | 'down' | 'unknown' | string
    state?: 'ready' | 'degraded' | 'disabled' | string
    last_checked_at?: string | null
    message?: string
    last_error?: string | null
    retry_after_seconds?: number | null
    next_action?: string | null
  }
  setup_status?: {
    auth_mode?: string
    marketplace?: string
    token_state?: string
    validation_status?: string
    health_state?: string
    next_action?: string
    base_url_set?: boolean
    sender_chat_state?: string
    bot_token_state?: string
    webhook_state?: string
    runtime_proof?: string
  }
  last_run?: {
    status: 'idle' | 'running' | 'success' | 'failed' | 'never' | string
    finished_at?: string | null
  }
  seller_operations?: SellerOperationStatus[]
  setup_schema?: IntegrationSetupSchema
}

type IntegrationSetupSchema = {
  schema_ref: string
  persistence_scope?: string
  submit_target?: string
  secret_target?: string
  validate_action?: string
  fields?: IntegrationSetupField[]
}

type IntegrationSetupField = {
  key: string
  label: string
  type:
    | 'text'
    | 'secret'
    | 'url'
    | 'number'
    | 'select'
    | 'multiselect'
    | 'checkbox'
    | 'textarea'
    | 'file'
    | 'oauth-connect'
    | 'browser-auth-status'
    | string
  required?: boolean
  write_only?: boolean
  read_only?: boolean
  persistence?: string
  secret_key?: string
  default?: string | number | boolean | null
  placeholder?: string
  helper_text?: string
  options?: Array<{ value: string; label: string }>
  validation_rules?: string[]
  condition?: Record<string, string>
  documentation_url?: string
}

const formatEbaySetupNextAction = (nextAction?: string | null) => {
  switch (nextAction) {
    case 'run_ebay_query_sets_from_market_watch':
    case 'run_market_watch_query_sets':
      return 'Ready for Market Watch runs'
    case 'check_provider_health_and_credentials':
      return 'Check provider health and credentials'
    case 'review_provider_credentials_and_health':
      return 'Review credentials and provider health'
    case 'retry_after_backoff':
      return 'Wait for retry backoff before running eBay query sets'
    case 'save_ebay_credentials_and_marketplace':
      return 'Save credentials and marketplace, then validate health'
    default:
      return 'Save credentials, then validate health'
  }
}

const formatEbaySetupBaseURLState = (baseURLSet?: boolean) =>
  baseURLSet
    ? 'Base URL override configured'
    : 'Using default eBay Browse API base URL'

const providerManifestActions = (provider: ProviderRecord) =>
  [
    ...(provider.workflow_refs ?? []),
    ...Object.entries(provider.capabilities ?? {})
      .filter(([, enabled]) => enabled)
      .map(([capability]) => `capability:${capability}`),
  ].filter(Boolean)

type IntegrationForm = {
  baseURL: string
  token: string
  marketplace: string
  itemsPerPage: string
  openAiModel: string
  openAiTestPrompt: string
  buyerInterestPayload: string
  landedCostPayload: string
  listingLifecycleItemId: string
  listingLifecycleDraftId: string
  listingLifecycleListingId: string
  listingLifecycleTitle: string
}

type TelegramConnectionStatus = {
  profile_id: string
  status: string
  state: string
  code: string
  message: string
  next_action: string
  transport: 'long_polling'
  public_listener: false
  bot_token_present: boolean
  credential_returned: false
  bot_username?: string
  sender_id?: string
  chat_id?: string
  paired: boolean
  paused: boolean
  webhook_conflict: boolean
  offset: number
  last_success_at?: string
  last_update_id?: string
  last_error_code?: string
  retry_after_seconds?: number
}

type TelegramConnectionTest = {
  status: string
  code: string
  message: string
  next_action: string
  webhook_conflict: boolean
  bot_token_present: boolean
  credential_returned: false
  bot?: { username?: string }
}

type TelegramPairingCode = {
  profile_id: string
  code: string
  expires_at: string
}

type BuyerInterestSyncResult = {
  provider?: string
  mode?: string
  total?: number
  counts?: Record<string, number>
  mappings?: Array<{
    title?: string
    listing_id?: string
    interest_state?: string
    destination?: string
    write_back_allowed?: boolean
    write_back_blocker?: string
  }>
  results?: Array<{
    title?: string
    listing_id?: string
    interest_state?: string
    destination?: string
    persisted_id?: string
    item_id?: string
    candidate_id?: string
    write_back_allowed?: boolean
    write_back_blocker?: string
  }>
}

type SellerOperationStatus = {
  operation: string
  capability: string
  read_available: boolean
  write_available: boolean
  confirmation_required: boolean
  blocker?: string
}

type SellerOperationPreviewResult = {
  provider?: string
  mode?: string
  preview?: {
    operation?: string
    action?: string
    capability?: string
    read_available?: boolean
    write_available?: boolean
    confirmation_required?: boolean
    confirmed?: boolean
    allowed?: boolean
    remote_write?: boolean
    blocker?: string
  }
}

type SellerOperationExecuteResult = {
  provider?: string
  mode?: string
  execution?: {
    operation?: string
    action?: string
    capability?: string
    read_available?: boolean
    write_available?: boolean
    confirmation_required?: boolean
    confirmed?: boolean
    allowed?: boolean
    remote_write?: boolean
    blocker?: string
    executed?: boolean
    local_only?: boolean
    status?: string
    result?: {
      operation?: string
      source?: string
      records?: Array<{
        id?: string
        kind?: string
        status?: string
        title?: string
      }>
      summary?: Record<string, number>
    }
  }
}

type SellerListingLifecycleCommand =
  | 'draft'
  | 'publish'
  | 'revise'
  | 'end'
  | 'relist'

type SellerListingLifecyclePreviewResult = {
  provider?: string
  mode?: string
  preview?: {
    command?: string
    capability?: string
    confirmed?: boolean
    allowed?: boolean
    local_only?: boolean
    remote_write?: boolean
    confirmation_required?: boolean
    blocker?: string
  }
}

type SellerListingLifecycleExecuteResult = {
  provider?: string
  mode?: string
  execution?: {
    command?: string
    capability?: string
    confirmed?: boolean
    allowed?: boolean
    local_only?: boolean
    remote_write?: boolean
    confirmation_required?: boolean
    blocker?: string
    executed?: boolean
    status?: string
    response?: {
      provider?: string
      command?: string
      draft_id?: string
      listing_id?: string
      status?: string
    }
  }
}

type LandedCostPlanResult = {
  provider?: string
  mode?: string
  mutable?: boolean
  allocation?: {
    total_direct_cents?: number
    total_shared_cents?: number
    total_landed_cents?: number
    items?: Array<{
      item_id?: string
      direct_cost_cents?: number
      allocated_cost_cents?: number
      landed_cost_cents?: number
      component_allocations?: Array<{
        component_id?: string
        label?: string
        method?: string
        amount_cents?: number
        provenance?: string
      }>
      allocation_provenance_id?: string[]
    }>
  }
  consolidation?: {
    item_ids?: string[]
    estimated_value_cents?: number
    estimated_fee_cents?: number
    estimated_total_cents?: number
    threshold_state?: string
    warnings?: string[]
    mutable?: boolean
  }
}

type ProfileOption = {
  id: string
  name: string
}

type CompanionPairingRequest = {
  request_id: string
  pairing_code: string
  device_id: string
  device_name: string
  origin: string
  protocol_version: string
  capabilities: string[]
  status: 'pending' | 'approved' | string
  expires_at: string
  created_at: string
}

type CompanionSession = {
  id: string
  cabinet_instance_id: string
  profile_id: string
  device_id: string
  device_name: string
  origin: string
  protocol_version: string
  capabilities: string[]
  created_at: string
  expires_at: string
  last_used_at?: string
  revoked_at?: string
}

type AppsProps = {
  title?: string
  description?: string
}

const integrationTablePageSize = 10

const appText = new Map<AppType, string>([
  ['all', 'All Integrations'],
  ['connected', 'Connected'],
  ['notConnected', 'Not Connected'],
])

const defaultBuyerInterestPayload = JSON.stringify(
  {
    source_account: 'buyer@example.test',
    items: [
      {
        listing_id: 'v1|watch|0',
        title: 'Watched eBay listing',
        url: 'https://www.ebay.com/itm/watch',
        state: 'watched',
      },
      {
        listing_id: 'v1|cart|0',
        title: 'Cart eBay listing',
        url: 'https://www.ebay.com/itm/cart',
        state: 'cart_like',
      },
    ],
  },
  null,
  2
)

const defaultLandedCostPayload = JSON.stringify(
  {
    items: [
      {
        id: 'card-b',
        purchase_cents: 30000,
        domestic_shipping_cents: 1500,
        tax_cents: 3000,
        weight_grams: 300,
      },
      {
        id: 'card-a',
        purchase_cents: 10000,
        domestic_shipping_cents: 500,
        tax_cents: 1000,
        weight_grams: 100,
      },
    ],
    components: [
      {
        id: 'intl',
        label: 'International shipping',
        amount_cents: 8000,
        allocation_method: 'weight',
        provenance: 'forwarder-shipment:SHIP-1',
      },
      {
        id: 'handling',
        label: 'Handling',
        amount_cents: 1200,
        allocation_method: 'equal',
        provenance: 'forwarder-invoice:INV-1',
      },
    ],
    consolidation: {
      shipment_fee_cents: 2500,
      destination_limit_cents: 60000,
      warning_buffer_cents: 1500,
    },
  },
  null,
  2
)

const listingLifecycleCommands: SellerListingLifecycleCommand[] = [
  'draft',
  'publish',
  'revise',
  'end',
  'relist',
]

function sellerListingLifecycleLabel(command: string) {
  const labels: Record<string, string> = {
    draft: 'Create draft',
    publish: 'Publish',
    revise: 'Revise',
    end: 'End',
    relist: 'Relist',
  }
  return labels[command] ?? command
}

function formatCents(value?: number) {
  return '$' + ((value ?? 0) / 100).toFixed(2)
}

function sellerOperationLabel(operation: string) {
  const labels: Record<string, string> = {
    messages: 'Messages',
    notifications: 'Notifications',
    sold_orders: 'Sold orders',
    fulfillment: 'Fulfilment',
    offers: 'Offers',
  }
  return labels[operation] ?? operation
}

function sellerOperationState(status: SellerOperationStatus) {
  if (status.write_available) {
    return status.confirmation_required ? 'Write confirmed' : 'Writable'
  }
  if (status.read_available) {
    return 'Read-only'
  }
  return 'Blocked'
}

function capabilityLabels(provider: ProviderRecord) {
  return [
    provider.capabilities.search ? 'Search' : null,
    provider.capabilities.stock_observation ? 'Stock' : null,
    provider.capabilities.pricing ? 'Pricing' : null,
    provider.capabilities.health ? 'Health' : null,
    provider.capabilities.assistant ? 'Assistant' : null,
    provider.capabilities.image_help ? 'Image help' : null,
    provider.capabilities.media_capture ? 'Media capture' : null,
    provider.capabilities.text_capture ? 'Text capture' : null,
  ].filter((label): label is string => Boolean(label))
}

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
  if (providerID === 'telegram') {
    return {
      baseURLKey: 'telegram.polling.transport',
      tokenKey: 'telegram_bot_token',
      marketplaceKey: 'telegram.bot_username',
      enabledKey: 'integration.telegram.enabled',
      itemsPerPageKey: 'telegram.polling.state',
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

function providerSettingAliases(providerID: string) {
  const keys = providerSettingsKeys(providerID)
  return {
    enabledKeys: [keys.enabledKey, `integration.${providerID}.enabled`],
    baseURLKeys: [keys.baseURLKey, `integration.${providerID}.base_url`],
    tokenKeys: [keys.tokenKey, `integration.${providerID}.token`],
    marketplaceKeys: [
      keys.marketplaceKey,
      `integration.${providerID}.marketplace`,
    ],
    itemsPerPageKeys: [
      keys.itemsPerPageKey,
      `integration.${providerID}.items_per_page`,
    ],
  }
}

function notificationHistoryID(value: string) {
  return value.replace(/[^a-z0-9]+/gi, '-').toLowerCase()
}

function providerDefaultSetupSchema(
  provider: ProviderRecord
): IntegrationSetupSchema | undefined {
  if (provider.setup_schema?.fields?.length) {
    return provider.setup_schema
  }
  if (
    provider.provider_id !== 'ebay' &&
    provider.integration_mode !== 'web_ingestion'
  ) {
    return provider.setup_schema
  }
  const keys = providerSettingsKeys(provider.provider_id)
  return {
    schema_ref: `cabinet.provider.${provider.provider_id}.default`,
    persistence_scope: 'profile_settings',
    fields: [
      {
        key: keys.baseURLKey,
        label: 'Base URL',
        type: 'url',
        placeholder: provider.base_domain
          ? `https://${provider.base_domain}`
          : 'https://provider.example',
      },
      {
        key: keys.marketplaceKey,
        label: 'Marketplace / Region',
        type: 'text',
        default: provider.provider_id === 'ebay' ? 'EBAY_AU' : '',
      },
      {
        key: keys.itemsPerPageKey,
        label: 'Items per page',
        type: 'number',
        default: '50',
      },
      ...(provider.auth_mode === 'none'
        ? []
        : [
            {
              key: keys.tokenKey,
              label: 'New token / API key',
              type: 'secret',
              write_only: true,
              persistence: 'profile_secrets',
              secret_key: keys.tokenKey,
            } satisfies IntegrationSetupField,
          ]),
    ],
  }
}

function withDefaultSetupSchema(provider: ProviderRecord): ProviderRecord {
  const setupSchema = providerDefaultSetupSchema(provider)
  return setupSchema === provider.setup_schema
    ? provider
    : { ...provider, setup_schema: setupSchema }
}

function setupFieldID(field: IntegrationSetupField) {
  switch (field.key) {
    case 'ebay_base_url':
      return 'provider-base-url'
    case 'ebay_marketplace':
      return 'provider-marketplace'
    case 'integration.ebay.items_per_page':
      return 'provider-items-per-page'
    case 'ebay_bearer_token':
      return 'provider-token'
    default:
      return `provider-schema-${notificationHistoryID(field.key)}`
  }
}

function recordIntegrationsStatusHistory({
  id,
  level = 'success',
  title,
  summary = 'Operational status from Integrations.',
}: {
  id: string
  level?: 'success' | 'error' | 'warning'
  title: string
  summary?: string
}) {
  recordNotificationHistory({
    id,
    level,
    title,
    summary,
    source_label: 'Integrations',
    category: 'system',
  })
}

function isConnected(
  provider: ProviderRecord,
  settings: Record<string, string>
) {
  if (provider.provider_id === 'openai') {
    return (
      settings['openai.active_auth_method'] === 'api_key' ||
      settings['openai.active_auth_method'] === 'browser_auth' ||
      provider.active_auth_method === 'api_key' ||
      provider.active_auth_method === 'browser_auth'
    )
  }
  if (provider.auth_methods?.sender_chat?.connected) {
    return true
  }
  if (provider.auth_mode === 'none') {
    return true
  }
  if (provider.has_token) {
    return true
  }
  const aliases = providerSettingAliases(provider.provider_id)
  return aliases.enabledKeys.some((key) => settings[key] === 'true')
}

function isConfiguredIntegration(
  provider: ProviderRecord,
  settings: Record<string, string>
) {
  if (provider.has_token || provider.active_auth_method) {
    return true
  }
  if (
    provider.auth_methods?.api_key?.connected ||
    provider.auth_methods?.browser_auth?.connected ||
    provider.auth_methods?.sender_chat?.connected
  ) {
    return true
  }
  const aliases = providerSettingAliases(provider.provider_id)
  return [
    ...aliases.enabledKeys,
    ...aliases.baseURLKeys,
    ...aliases.tokenKeys,
    ...aliases.marketplaceKeys,
    ...aliases.itemsPerPageKeys,
  ].some((key) => Object.prototype.hasOwnProperty.call(settings, key))
}

function fieldDefaultValue(field: IntegrationSetupField) {
  if (field.default === undefined || field.default === null) {
    return ''
  }
  return String(field.default)
}

function providerManifestFieldValue(
  provider: ProviderRecord,
  field: IntegrationSetupField
) {
  const value = provider[field.key as keyof ProviderRecord]
  if (
    typeof value === 'string' ||
    typeof value === 'number' ||
    typeof value === 'boolean'
  ) {
    return String(value)
  }
  return fieldDefaultValue(field)
}

function setupFieldInitialValue(
  provider: ProviderRecord,
  field: IntegrationSetupField,
  settings: Record<string, string>
) {
  if (field.persistence === 'provider_manifest') {
    return providerManifestFieldValue(provider, field)
  }
  return settings[field.key] ?? fieldDefaultValue(field)
}

function setupFieldVisible(
  field: IntegrationSetupField,
  values: Record<string, string>
) {
  if (!field.condition) {
    return true
  }
  return Object.entries(field.condition).every(
    ([key, expected]) => values[key] === expected
  )
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
    provider: requestedProviderID = '',
  } = route.useSearch()
  const navigate = route.useNavigate()

  const [sort, setSort] = useState(initSort)
  const [appType, setAppType] = useState(type)
  const [searchTerm, setSearchTerm] = useState(filter)
  const viewMode: ViewMode = view ?? 'rows'
  const [tablePage, setTablePage] = useState(1)
  const [activeProfileId, setActiveProfileId] = useState('')
  const [settings, setSettings] = useState<Record<string, string>>({})
  const [providers, setProviders] = useState<ProviderRecord[]>([])
  const [loading, setLoading] = useState(true)
  const [bootstrapError, setBootstrapError] = useState<string | null>(null)
  const [companionPairingRequests, setCompanionPairingRequests] = useState<
    CompanionPairingRequest[]
  >([])
  const [companionSessions, setCompanionSessions] = useState<
    CompanionSession[]
  >([])
  const [companionSecurityError, setCompanionSecurityError] = useState<
    string | null
  >(null)
  const [companionWorking, setCompanionWorking] = useState<string | null>(null)
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
  const [providerSelectorOpen, setProviderSelectorOpen] = useState(false)
  const [pendingProviderSetup, setPendingProviderSetup] =
    useState<ProviderRecord | null>(null)
  const [providerCatalogSearchTerm, setProviderCatalogSearchTerm] = useState('')
  const [providerCatalogCategory, setProviderCatalogCategory] = useState('all')
  const [saving, setSaving] = useState(false)
  const [validating, setValidating] = useState(false)
  const [actionMessage, setActionMessage] = useState<string | null>(null)
  const [saveError, setSaveError] = useState<string | null>(null)
  const [tokenFieldError, setTokenFieldError] = useState<string | null>(null)
  const [replaceToken, setReplaceToken] = useState(false)
  const [telegramStatus, setTelegramStatus] =
    useState<TelegramConnectionStatus | null>(null)
  const [telegramPairing, setTelegramPairing] =
    useState<TelegramPairingCode | null>(null)
  const [telegramBotUsername, setTelegramBotUsername] = useState('')
  const [telegramWorking, setTelegramWorking] = useState<string | null>(null)
  const [rowDetailsProviderID, setRowDetailsProviderID] = useState<
    string | null
  >(null)
  const [rowDetailsOpen, setRowDetailsOpen] = useState(false)
  const [rowEditProviderID, setRowEditProviderID] = useState<string | null>(
    null
  )
  const [rowEditOpen, setRowEditOpen] = useState(false)
  const clickTimerRef = useRef<number | null>(null)
  const addIntegrationButtonRef = useRef<HTMLButtonElement | null>(null)
  const providerSelectorWasOpenedRef = useRef(false)
  const providerSetupHandoffFrameRef = useRef<number | null>(null)
  const deepLinkedProviderRef = useRef('')
  const tokenInputRef = useRef<HTMLInputElement | null>(null)
  const [form, setForm] = useState<IntegrationForm>({
    baseURL: '',
    token: '',
    marketplace: '',
    itemsPerPage: '',
    openAiModel: 'gpt-4o-mini',
    openAiTestPrompt:
      'Write one sentence confirming OpenAI is connected to Cabinet.',
    buyerInterestPayload: defaultBuyerInterestPayload,
    landedCostPayload: defaultLandedCostPayload,
    listingLifecycleItemId: 'item-local-1',
    listingLifecycleDraftId: 'draft-local-1',
    listingLifecycleListingId: 'listing-live-1',
    listingLifecycleTitle: 'Cabinet eBay listing draft',
  })
  const [setupSchemaValues, setSetupSchemaValues] = useState<
    Record<string, string>
  >({})
  const [buyerInterestWorking, setBuyerInterestWorking] = useState(false)
  const [buyerInterestError, setBuyerInterestError] = useState<string | null>(
    null
  )
  const [buyerInterestResult, setBuyerInterestResult] =
    useState<BuyerInterestSyncResult | null>(null)
  const [sellerOperationWorking, setSellerOperationWorking] = useState<
    string | null
  >(null)
  const [sellerOperationError, setSellerOperationError] = useState<
    string | null
  >(null)
  const [sellerOperationResult, setSellerOperationResult] =
    useState<SellerOperationPreviewResult | null>(null)
  const [sellerOperationExecution, setSellerOperationExecution] =
    useState<SellerOperationExecuteResult | null>(null)
  const [listingLifecycleWorking, setListingLifecycleWorking] = useState<
    string | null
  >(null)
  const [listingLifecycleError, setListingLifecycleError] = useState<
    string | null
  >(null)
  const [listingLifecycleResult, setListingLifecycleResult] =
    useState<SellerListingLifecyclePreviewResult | null>(null)
  const [listingLifecycleExecution, setListingLifecycleExecution] =
    useState<SellerListingLifecycleExecuteResult | null>(null)
  const [landedCostWorking, setLandedCostWorking] = useState(false)
  const [landedCostError, setLandedCostError] = useState<string | null>(null)
  const [landedCostResult, setLandedCostResult] =
    useState<LandedCostPlanResult | null>(null)

  const loadCompanionSecurity = useCallback(async () => {
    setCompanionSecurityError(null)
    try {
      const [pairingResponse, sessionsResponse] = await Promise.all([
        fetch('/api/companion/pairing/requests'),
        fetch('/api/companion/sessions'),
      ])
      if (!pairingResponse.ok || !sessionsResponse.ok) {
        throw new Error('companion_security_load_failed')
      }
      const pairingPayload = (await pairingResponse.json()) as {
        requests?: CompanionPairingRequest[]
      }
      const sessionsPayload = (await sessionsResponse.json()) as {
        sessions?: CompanionSession[]
      }
      setCompanionPairingRequests(pairingPayload.requests ?? [])
      setCompanionSessions(sessionsPayload.sessions ?? [])
    } catch {
      setCompanionPairingRequests([])
      setCompanionSessions([])
      setCompanionSecurityError(
        'Browser Companion security status is unavailable. Unlock Cabinet and retry.'
      )
    }
  }, [])

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
      setProviders(
        (registryPayload.providers ?? []).map(withDefaultSetupSchema)
      )

      const settingsResp = await fetch(`/api/profiles/${profileID}/settings`)
      if (!settingsResp.ok) {
        throw new Error(`profile_settings_${settingsResp.status}`)
      }
      const settingsPayload = (await settingsResp.json()) as {
        settings?: Record<string, string>
      }
      setSettings(settingsPayload.settings ?? {})
      await loadCompanionSecurity()
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
  }, [loadCompanionSecurity])

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

  const configuredProviders = sortedProviders.filter((provider) =>
    isConfiguredIntegration(provider, settings)
  )
  const addableProviders = sortedProviders.filter(
    (provider) => !isConfiguredIntegration(provider, settings)
  )
  const filteredProviders = configuredProviders
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
  const tablePageCount = Math.max(
    1,
    Math.ceil(filteredProviders.length / integrationTablePageSize)
  )
  const tableStart = (tablePage - 1) * integrationTablePageSize
  const paginatedProviders = filteredProviders.slice(
    tableStart,
    tableStart + integrationTablePageSize
  )
  const tableRangeStart = filteredProviders.length ? tableStart + 1 : 0
  const tableRangeEnd = Math.min(
    filteredProviders.length,
    tableStart + paginatedProviders.length
  )
  const providerCatalogProviders = addableProviders.length
    ? addableProviders
    : sortedProviders
  const providerCatalogCategories = useMemo(() => {
    const categories = new Set<string>()
    providerCatalogProviders.forEach((provider) => {
      const category =
        provider.provider_category?.trim() || provider.provider_type?.trim()
      if (category) {
        categories.add(category)
      }
    })
    return [...categories].sort((a, b) => a.localeCompare(b))
  }, [providerCatalogProviders])
  const filteredProviderCatalogProviders = useMemo(() => {
    const search = providerCatalogSearchTerm.trim().toLowerCase()
    return providerCatalogProviders
      .filter((provider) => {
        if (providerCatalogCategory === 'all') {
          return true
        }
        return (
          provider.provider_category === providerCatalogCategory ||
          provider.provider_type === providerCatalogCategory
        )
      })
      .filter((provider) => {
        if (!search) {
          return true
        }
        const capabilityTerms = Object.entries(provider.capabilities ?? {})
          .filter(([, enabled]) => enabled)
          .map(([capability]) => capability)
        return [
          provider.display_name,
          provider.base_domain,
          provider.provider_category,
          provider.provider_type,
          provider.auth_mode,
          provider.state,
          provider.health?.status,
          provider.health?.state,
          provider.setup_instructions,
          provider.setup_status?.next_action,
          ...(provider.workflow_refs ?? []),
          ...capabilityTerms,
        ]
          .filter(Boolean)
          .join(' ')
          .toLowerCase()
          .includes(search)
      })
  }, [
    providerCatalogCategory,
    providerCatalogProviders,
    providerCatalogSearchTerm,
  ])
  const providerCatalogStatusClassName = (provider: ProviderRecord) => {
    const status = (
      provider.health?.state ??
      provider.health?.status ??
      provider.state
    ).toLowerCase()
    if (status.includes('disabled') || status.includes('deprecated')) {
      return 'border border-destructive/40 bg-destructive/10 text-destructive'
    }
    if (status.includes('degraded') || status.includes('repair')) {
      return 'border border-amber-500/40 bg-amber-500/10 text-amber-700 dark:text-amber-300'
    }
    if (status.includes('beta')) {
      return 'border border-sky-500/40 bg-sky-500/10 text-sky-700 dark:text-sky-300'
    }
    return 'bg-muted text-muted-foreground'
  }

  useEffect(() => {
    setTablePage((page) => Math.min(page, tablePageCount))
  }, [tablePageCount])

  useEffect(() => {
    if (providerSelectorOpen) {
      setProviderCatalogSearchTerm('')
      setProviderCatalogCategory('all')
    }
  }, [providerSelectorOpen])

  const handleSearch = (e: ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value
    setSearchTerm(value)
    setTablePage(1)
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
    setTablePage(1)
    navigate({
      search: (prev) => ({
        ...prev,
        type: value === 'all' ? undefined : value,
      }),
    })
  }

  const handleSortChange = (nextSort: 'asc' | 'desc') => {
    setSort(nextSort)
    setTablePage(1)
    navigate({ search: (prev) => ({ ...prev, sort: nextSort }) })
  }

  const handleViewChange = (nextView: ViewMode) => {
    navigate({
      search: (prev) => ({
        ...prev,
        view: nextView === 'rows' ? undefined : nextView,
      }),
    })
  }

  const approveCompanionPairing = async (request: CompanionPairingRequest) => {
    setCompanionWorking(request.request_id)
    setCompanionSecurityError(null)
    try {
      const response = await fetch('/api/companion/pairing/approvals', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          request_id: request.request_id,
          profile_id: activeProfileId,
          capabilities: request.capabilities,
        }),
      })
      if (!response.ok) {
        throw new Error(`companion_pairing_approval_${response.status}`)
      }
      setActionMessage(`Approved Browser Companion ${request.device_name}.`)
      await loadCompanionSecurity()
    } catch {
      setCompanionSecurityError(
        'Pairing approval failed. Confirm the request is still open and Cabinet is unlocked.'
      )
    } finally {
      setCompanionWorking(null)
    }
  }

  const rejectCompanionPairing = async (request: CompanionPairingRequest) => {
    setCompanionWorking(request.request_id)
    setCompanionSecurityError(null)
    try {
      const response = await fetch(
        `/api/companion/pairing/requests?id=${encodeURIComponent(request.request_id)}`,
        { method: 'DELETE' }
      )
      if (!response.ok) {
        throw new Error(`companion_pairing_reject_${response.status}`)
      }
      setActionMessage(`Rejected Browser Companion ${request.device_name}.`)
      await loadCompanionSecurity()
    } catch {
      setCompanionSecurityError('Could not reject the pairing request.')
    } finally {
      setCompanionWorking(null)
    }
  }

  const revokeCompanionSessions = async (sessionID?: string) => {
    const workingID = sessionID ?? 'all'
    setCompanionWorking(workingID)
    setCompanionSecurityError(null)
    try {
      const query = sessionID
        ? `id=${encodeURIComponent(sessionID)}`
        : 'all=true'
      const response = await fetch(`/api/companion/sessions?${query}`, {
        method: 'DELETE',
      })
      if (!response.ok) {
        throw new Error(`companion_session_revoke_${response.status}`)
      }
      setActionMessage(
        sessionID
          ? 'Revoked the selected Browser Companion session.'
          : 'Revoked all Browser Companion sessions for this profile.'
      )
      await loadCompanionSecurity()
    } catch {
      setCompanionSecurityError('Could not revoke Browser Companion access.')
    } finally {
      setCompanionWorking(null)
    }
  }

  const openIntegration = useCallback(
    (provider: ProviderRecord) => {
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
        openAiModel: settings['assistant_default_model'] ?? 'gpt-4o-mini',
        openAiTestPrompt:
          'Write one sentence confirming OpenAI is connected to Cabinet.',
        buyerInterestPayload: defaultBuyerInterestPayload,
        landedCostPayload: defaultLandedCostPayload,
        listingLifecycleItemId: 'item-local-1',
        listingLifecycleDraftId: 'draft-local-1',
        listingLifecycleListingId: 'listing-live-1',
        listingLifecycleTitle: 'Cabinet eBay listing draft',
      })
      setSetupSchemaValues(
        Object.fromEntries(
          (provider.setup_schema?.fields ?? []).map((field) => [
            field.key,
            setupFieldInitialValue(provider, field, settings),
          ])
        )
      )
      setBuyerInterestError(null)
      setBuyerInterestResult(null)
      setSellerOperationError(null)
      setSellerOperationResult(null)
      setSellerOperationExecution(null)
      setSellerOperationWorking(null)
      setListingLifecycleError(null)
      setListingLifecycleResult(null)
      setListingLifecycleExecution(null)
      setListingLifecycleWorking(null)
      setLandedCostError(null)
      setLandedCostResult(null)
      setLandedCostWorking(false)
      setReplaceToken(
        !(
          provider.has_token ||
          provider.auth_methods?.bot_token?.credential_present
        )
      )
    },
    [settings]
  )

  useEffect(() => {
    const providerID = requestedProviderID.trim()
    if (!providerID) {
      deepLinkedProviderRef.current = ''
      return
    }
    if (loading || deepLinkedProviderRef.current === providerID) {
      return
    }
    const requestedProvider = providerByID.get(providerID)
    if (!requestedProvider) {
      return
    }
    deepLinkedProviderRef.current = providerID
    openIntegration(requestedProvider)
    void navigate({
      search: (previous) => ({ ...previous, provider: undefined }),
      replace: true,
    })
  }, [loading, navigate, openIntegration, providerByID, requestedProviderID])

  useEffect(() => {
    if (providerSelectorOpen || !providerSelectorWasOpenedRef.current) {
      return
    }
    providerSelectorWasOpenedRef.current = false
    if (providerSetupHandoffFrameRef.current !== null) {
      window.cancelAnimationFrame(providerSetupHandoffFrameRef.current)
    }
    providerSetupHandoffFrameRef.current = window.requestAnimationFrame(() => {
      providerSetupHandoffFrameRef.current = null
      if (pendingProviderSetup) {
        openIntegration(pendingProviderSetup)
        setPendingProviderSetup(null)
        return
      }
      addIntegrationButtonRef.current?.focus()
    })

    return () => {
      if (providerSetupHandoffFrameRef.current !== null) {
        window.cancelAnimationFrame(providerSetupHandoffFrameRef.current)
        providerSetupHandoffFrameRef.current = null
      }
    }
  }, [openIntegration, pendingProviderSetup, providerSelectorOpen])

  const selectProviderForSetup = (provider: ProviderRecord) => {
    setPendingProviderSetup(provider)
    setProviderSelectorOpen(false)
  }

  useEffect(() => {
    return () => {
      if (clickTimerRef.current !== null) {
        window.clearTimeout(clickTimerRef.current)
      }
      if (providerSetupHandoffFrameRef.current !== null) {
        window.cancelAnimationFrame(providerSetupHandoffFrameRef.current)
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
    setTokenFieldError(null)
    setReplaceToken(false)
    setForm({
      baseURL: '',
      token: '',
      marketplace: '',
      itemsPerPage: '',
      openAiModel: 'gpt-4o-mini',
      openAiTestPrompt:
        'Write one sentence confirming OpenAI is connected to Cabinet.',
      buyerInterestPayload: defaultBuyerInterestPayload,
      landedCostPayload: defaultLandedCostPayload,
      listingLifecycleItemId: 'item-local-1',
      listingLifecycleDraftId: 'draft-local-1',
      listingLifecycleListingId: 'listing-live-1',
      listingLifecycleTitle: 'Cabinet eBay listing draft',
    })
    setSetupSchemaValues({})
    setBuyerInterestError(null)
    setBuyerInterestResult(null)
    setSellerOperationError(null)
    setSellerOperationResult(null)
    setSellerOperationExecution(null)
    setSellerOperationWorking(null)
    setListingLifecycleError(null)
    setListingLifecycleResult(null)
    setListingLifecycleExecution(null)
    setListingLifecycleWorking(null)
    setLandedCostError(null)
    setLandedCostResult(null)
    setLandedCostWorking(false)
    setTelegramPairing(null)
    setTelegramWorking(null)
  }

  const loadTelegramStatus = useCallback(async () => {
    if (!activeProfileId) return
    try {
      const response = await fetch(
        `/api/telegram/connection/status?profile_id=${encodeURIComponent(activeProfileId)}`
      )
      if (!response.ok) throw new Error(`telegram_status_${response.status}`)
      const payload = (await response.json()) as TelegramConnectionStatus
      setTelegramStatus(payload)
      setTelegramBotUsername(payload.bot_username ?? '')
    } catch {
      setTelegramStatus(null)
    }
  }, [activeProfileId])

  useEffect(() => {
    if (editingProviderID === 'telegram') {
      void loadTelegramStatus()
    }
  }, [editingProviderID, loadTelegramStatus])

  const testTelegramConnection = async () => {
    if (!activeProfileId) return
    const token = form.token.trim()
    if (!token && !telegramStatus?.bot_token_present) {
      showTokenFieldError('Paste the BotFather token before testing Telegram.')
      return
    }
    setTelegramWorking('test')
    setSaveError(null)
    try {
      const response = await fetch('/api/telegram/connection/test', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ profile_id: activeProfileId, bot_token: token }),
      })
      const payload = (await response.json()) as TelegramConnectionTest
      setForm((previous) => ({ ...previous, token: '' }))
      setTelegramBotUsername(payload.bot?.username ?? telegramBotUsername)
      setActionMessage(payload.message)
      if (!response.ok && response.status !== 409) {
        throw new Error(payload.message || `telegram_test_${response.status}`)
      }
      await loadTelegramStatus()
    } catch (error) {
      setSaveError(
        error instanceof Error ? error.message : 'Telegram test failed.'
      )
    } finally {
      setTelegramWorking(null)
    }
  }

  const resolveTelegramWebhook = async () => {
    if (!activeProfileId) return
    setTelegramWorking('webhook')
    setSaveError(null)
    try {
      const response = await fetch('/api/telegram/connection/resolve-webhook', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ profile_id: activeProfileId }),
      })
      const payload = (await response.json()) as TelegramConnectionTest
      if (!response.ok) {
        throw new Error(
          payload.message || `telegram_webhook_${response.status}`
        )
      }
      setActionMessage(payload.message)
      await loadTelegramStatus()
    } catch (error) {
      setSaveError(
        error instanceof Error ? error.message : 'Webhook removal failed.'
      )
    } finally {
      setTelegramWorking(null)
    }
  }

  const createTelegramPairingCode = async () => {
    if (!activeProfileId) return
    setTelegramWorking('pair')
    setSaveError(null)
    try {
      const response = await fetch('/api/telegram/pairing-codes', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ profile_id: activeProfileId }),
      })
      if (!response.ok) {
        throw new Error(`telegram_pairing_${response.status}`)
      }
      setTelegramPairing((await response.json()) as TelegramPairingCode)
      setActionMessage(
        'Pairing code created. Send the shown /start command to your bot in a private Telegram chat.'
      )
    } catch (error) {
      setSaveError(
        error instanceof Error ? error.message : 'Pairing code failed.'
      )
    } finally {
      setTelegramWorking(null)
    }
  }

  const setTelegramPaused = async (paused: boolean) => {
    if (!activeProfileId) return
    setTelegramWorking(paused ? 'pause' : 'resume')
    try {
      const response = await fetch('/api/telegram/connection/pause', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ profile_id: activeProfileId, paused }),
      })
      if (!response.ok) throw new Error(`telegram_pause_${response.status}`)
      setTelegramStatus((await response.json()) as TelegramConnectionStatus)
    } catch (error) {
      setSaveError(error instanceof Error ? error.message : 'Pause failed.')
    } finally {
      setTelegramWorking(null)
    }
  }

  const disconnectTelegram = async () => {
    if (!activeProfileId) return
    setTelegramWorking('disconnect')
    try {
      const response = await fetch('/api/telegram/connection/disconnect', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ profile_id: activeProfileId }),
      })
      if (!response.ok) {
        throw new Error(`telegram_disconnect_${response.status}`)
      }
      setTelegramStatus((await response.json()) as TelegramConnectionStatus)
      setTelegramPairing(null)
      setTelegramBotUsername('')
      setReplaceToken(true)
      setActionMessage('Telegram disconnected and its bot token removed.')
    } catch (error) {
      setSaveError(
        error instanceof Error ? error.message : 'Disconnect failed.'
      )
    } finally {
      setTelegramWorking(null)
    }
  }

  useEffect(() => {
    if (tokenFieldError) {
      tokenInputRef.current?.focus()
    }
  }, [tokenFieldError])

  const showTokenFieldError = (message: string) => {
    setTokenFieldError(message)
    recordIntegrationsStatusHistory({
      id: `integrations-token-field-${notificationHistoryID(message)}`,
      level: 'error',
      title: message,
      summary: 'Provider credential field validation status from Integrations.',
    })
    window.setTimeout(() => tokenInputRef.current?.focus(), 0)
  }

  const upsertProviderIntegrationInstance = async (
    profileID: string,
    provider: ProviderRecord,
    config: Record<string, string>
  ) => {
    const instancesResponse = await fetch(
      `/api/profiles/${profileID}/integration-instances`
    )
    if (!instancesResponse.ok) {
      throw new Error(`integration_instances_list_${instancesResponse.status}`)
    }
    const instancesPayload = (await instancesResponse.json()) as {
      instances?: Array<{ id: string; provider_id: string }>
    }
    const existingInstance = instancesPayload.instances?.find(
      (instance) => instance.provider_id === provider.provider_id
    )
    const instanceResponse = await fetch(
      `/api/profiles/${profileID}/integration-instances`,
      {
        method: existingInstance ? 'PUT' : 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          ...(existingInstance
            ? { id: existingInstance.id }
            : { auth_state: 'configured', health_state: 'unknown' }),
          provider_id: provider.provider_id,
          display_name: provider.display_name,
          enabled: true,
          config,
        }),
      }
    )
    if (!instanceResponse.ok) {
      throw new Error(
        `integration_instance_save_failed_${instanceResponse.status}`
      )
    }
  }

  const saveIntegration = async () => {
    if (!activeProfileId || !editingProvider) {
      return
    }
    setSaving(true)
    setSaveError(null)
    setTokenFieldError(null)
    setActionMessage(null)
    try {
      const keys = providerSettingsKeys(editingProvider.provider_id)
      const trimmedToken = form.token.trim()
      if (editingProvider.provider_id === 'openai') {
        if (
          trimmedToken === '' &&
          !editingProvider.has_token &&
          settings['openai.active_auth_method'] !== 'browser_auth'
        ) {
          showTokenFieldError('OpenAI API key is required before connecting.')
          return
        }
        const openAiSettings: Record<string, string> = {
          openai_base_url: form.baseURL.trim(),
          openai_active_auth_method:
            trimmedToken !== ''
              ? 'api_key'
              : (settings['openai_active_auth_method'] ??
                settings['openai.active_auth_method'] ??
                ''),
          'openai.active_auth_method':
            trimmedToken !== ''
              ? 'api_key'
              : (settings['openai.active_auth_method'] ?? ''),
          assistant_default_provider: 'openai',
          assistant_default_model: form.openAiModel.trim() || 'gpt-4o-mini',
          'integration.openai.enabled':
            trimmedToken !== '' ||
            settings['openai.active_auth_method'] === 'browser_auth'
              ? 'true'
              : 'false',
        }
        const settingsResponse = await fetch(
          `/api/profiles/${activeProfileId}/settings`,
          {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ settings: openAiSettings }),
          }
        )
        if (!settingsResponse.ok) {
          throw new Error(`save_failed_${settingsResponse.status}`)
        }
        if (trimmedToken !== '') {
          const secretResponse = await fetch(
            `/api/profiles/${activeProfileId}/secrets`,
            {
              method: 'PUT',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify({
                key: 'openai_api_key',
                value: trimmedToken,
              }),
            }
          )
          if (!secretResponse.ok) {
            throw new Error(`secret_save_failed_${secretResponse.status}`)
          }
        }
        const payload = (await settingsResponse.json()) as {
          settings?: Record<string, string>
        }
        setSettings(payload.settings ?? {})
        setProviders((prev) =>
          prev.map((provider) =>
            provider.provider_id === 'openai'
              ? {
                  ...provider,
                  has_token: provider.has_token || trimmedToken !== '',
                  active_auth_method:
                    trimmedToken !== ''
                      ? 'api_key'
                      : provider.active_auth_method,
                }
              : provider
          )
        )
        const message = 'OpenAI configuration saved.'
        setActionMessage(message)
        recordIntegrationsStatusHistory({
          id: 'integrations-openai-save-success',
          title: message,
          summary: 'Provider configuration save status from Integrations.',
        })
        closeIntegration()
        return
      }
      if (editingProvider.setup_schema?.fields?.length) {
        const settingsPayload: Record<string, string> = {}
        const visibleFields = editingProvider.setup_schema.fields.filter(
          (field) => setupFieldVisible(field, setupSchemaValues)
        )
        const missingField = visibleFields.find((field) => {
          if (!field.required || field.read_only) {
            return false
          }
          if (field.write_only && editingProvider.has_token) {
            return false
          }
          return (setupSchemaValues[field.key] ?? '').trim() === ''
        })
        if (missingField) {
          throw new Error(`${missingField.label} is required.`)
        }
        for (const field of visibleFields) {
          if (field.read_only || field.persistence === 'provider_manifest') {
            continue
          }
          const value = (setupSchemaValues[field.key] ?? '').trim()
          if (field.write_only || field.persistence === 'profile_secrets') {
            if (value === '') {
              continue
            }
            const secretResponse = await fetch(
              `/api/profiles/${activeProfileId}/secrets`,
              {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                  key: field.secret_key ?? field.key,
                  value,
                }),
              }
            )
            if (!secretResponse.ok) {
              throw new Error(`secret_save_failed_${secretResponse.status}`)
            }
            continue
          }
          settingsPayload[field.key] = value
        }
        const instanceConfig = { ...settingsPayload }
        settingsPayload[
          providerSettingsKeys(editingProvider.provider_id).enabledKey
        ] = 'true'
        const settingsResponse = await fetch(
          `/api/profiles/${activeProfileId}/settings`,
          {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ settings: settingsPayload }),
          }
        )
        if (!settingsResponse.ok) {
          throw new Error(`save_failed_${settingsResponse.status}`)
        }
        await upsertProviderIntegrationInstance(
          activeProfileId,
          editingProvider,
          instanceConfig
        )
        const payload = (await settingsResponse.json()) as {
          settings?: Record<string, string>
        }
        setSettings(payload.settings ?? {})
        setProviders((prev) =>
          prev.map((provider) =>
            provider.provider_id === editingProvider.provider_id
              ? {
                  ...provider,
                  has_token:
                    provider.has_token ||
                    visibleFields.some(
                      (field) =>
                        (field.write_only ||
                          field.persistence === 'profile_secrets') &&
                        (setupSchemaValues[field.key] ?? '').trim() !== ''
                    ),
                }
              : provider
          )
        )
        const message = 'Provider schema configuration saved.'
        setActionMessage(message)
        recordIntegrationsStatusHistory({
          id: `integrations-provider-schema-save-${notificationHistoryID(editingProvider.provider_id)}`,
          title: message,
          summary:
            'Schema-driven provider configuration save status from Integrations.',
        })
        closeIntegration()
        return
      }
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
      const message = 'Provider configuration saved.'
      setActionMessage(message)
      recordIntegrationsStatusHistory({
        id: `integrations-provider-save-${notificationHistoryID(editingProvider.provider_id)}`,
        title: message,
        summary: 'Provider configuration save status from Integrations.',
      })
      closeIntegration()
    } catch (err) {
      const message =
        err instanceof Error && err.message === 'token_required_for_provider'
          ? 'Token is required for this provider.'
          : err instanceof Error
            ? err.message
            : 'save_failed'
      if (
        err instanceof Error &&
        err.message === 'token_required_for_provider'
      ) {
        setSaveError(message)
      } else {
        setSaveError(message)
      }
      recordIntegrationsStatusHistory({
        id: editingProvider
          ? `integrations-provider-save-${notificationHistoryID(editingProvider.provider_id)}-failed`
          : 'integrations-provider-save-failed',
        level: 'error',
        title: message,
        summary: 'Provider configuration save status from Integrations.',
      })
    } finally {
      setSaving(false)
    }
  }

  const validateProvider = async () => {
    if (!editingProvider) {
      return
    }
    if (editingProvider.provider_id === 'telegram') {
      await testTelegramConnection()
      return
    }
    if (
      editingProvider.provider_id === 'openai' &&
      form.token.trim() === '' &&
      !editingProvider.has_token &&
      settings['openai.active_auth_method'] !== 'browser_auth'
    ) {
      setSaveError(null)
      showTokenFieldError('OpenAI API key is required before validating.')
      return
    }
    setValidating(true)
    setSaveError(null)
    setTokenFieldError(null)
    try {
      const response = await fetch(
        `/api/provider/health?provider=${encodeURIComponent(editingProvider.provider_id)}`
      )
      if (!response.ok) {
        throw new Error(`validate_failed_${response.status}`)
      }
      const payload = (await response.json()) as {
        status?: string
        state?: string
        message?: string
        last_error?: string | null
        retry_after_seconds?: number | null
        next_action?: string | null
        updated_at?: string
      }
      const checkedAt = payload.updated_at ?? new Date().toISOString()
      const healthStatus = payload.status ?? 'unknown'
      const readinessState = payload.state ?? healthStatus
      const nextProvider: ProviderRecord = {
        ...editingProvider,
        state: readinessState,
        health: {
          status: healthStatus,
          state: readinessState,
          message: payload.message,
          last_error: payload.last_error,
          retry_after_seconds: payload.retry_after_seconds,
          next_action: payload.next_action,
          last_checked_at: checkedAt,
        },
        last_run: {
          status:
            healthStatus === 'ok'
              ? 'success'
              : healthStatus === 'unknown'
                ? 'never'
                : 'failed',
          finished_at: checkedAt,
        },
      }
      setProviders((prev) =>
        prev.map((provider) =>
          provider.provider_id === editingProvider.provider_id
            ? {
                ...provider,
                state: nextProvider.state,
                health: nextProvider.health,
                last_run: nextProvider.last_run,
              }
            : provider
        )
      )
      const message = `Validated ${editingProvider.display_name} health: ${healthStatus}.`
      setActionMessage(message)
      recordNotificationHistory({
        id: `integrations-provider-health-${notificationHistoryID(editingProvider.provider_id)}-${notificationHistoryID(healthStatus)}`,
        level: healthStatus === 'ok' ? 'success' : 'warning',
        title: message,
        summary: 'Provider health validation status from Integrations.',
        source_label: 'Integrations',
        category: 'system',
      })
    } catch (error) {
      const message = error instanceof Error ? error.message : 'validate_failed'
      setSaveError(message)
      recordNotificationHistory({
        id: `integrations-provider-health-${notificationHistoryID(editingProvider.provider_id)}-failed`,
        level: 'error',
        title: message,
        summary: 'Provider health validation status from Integrations.',
        source_label: 'Integrations',
        category: 'system',
      })
    } finally {
      setValidating(false)
    }
  }

  const disconnectOpenAIApiKey = async () => {
    if (
      !activeProfileId ||
      !editingProvider ||
      editingProvider.provider_id !== 'openai'
    ) {
      return
    }
    setSaving(true)
    setSaveError(null)
    setTokenFieldError(null)
    setActionMessage(null)
    try {
      const deleteResponse = await fetch(
        `/api/profiles/${activeProfileId}/secrets?key=openai_api_key`,
        { method: 'DELETE' }
      )
      if (!deleteResponse.ok && deleteResponse.status !== 404) {
        throw new Error(`secret_delete_failed_${deleteResponse.status}`)
      }
      const currentBrowserState =
        settings['openai.browser_auth_state'] ??
        editingProvider.auth_methods?.browser_auth?.state ??
        ''
      const browserRemainsActive =
        settings['openai.active_auth_method'] === 'browser_auth' ||
        editingProvider.active_auth_method === 'browser_auth'
      const nextSettings: Record<string, string> = {
        openai_active_auth_method: browserRemainsActive ? 'browser_auth' : '',
        'openai.active_auth_method': browserRemainsActive ? 'browser_auth' : '',
        'integration.openai.enabled': browserRemainsActive ? 'true' : 'false',
      }
      if (currentBrowserState) {
        nextSettings['openai.browser_auth_state'] = currentBrowserState
      }
      const settingsResponse = await fetch(
        `/api/profiles/${activeProfileId}/settings`,
        {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ settings: nextSettings }),
        }
      )
      if (!settingsResponse.ok) {
        throw new Error(`save_failed_${settingsResponse.status}`)
      }
      const payload = (await settingsResponse.json()) as {
        settings?: Record<string, string>
      }
      setSettings(payload.settings ?? nextSettings)
      setProviders((prev) =>
        prev.map((provider) =>
          provider.provider_id === 'openai'
            ? {
                ...provider,
                has_token: false,
                active_auth_method: browserRemainsActive ? 'browser_auth' : '',
                auth_methods: {
                  ...provider.auth_methods,
                  api_key: {
                    ...(provider.auth_methods?.api_key ?? {}),
                    state: 'setup_needed',
                    connected: false,
                    credential_present: false,
                  },
                },
              }
            : provider
        )
      )
      setReplaceToken(true)
      setForm((prev) => ({ ...prev, token: '' }))
      const message = 'OpenAI API key disconnected.'
      setActionMessage(message)
      recordIntegrationsStatusHistory({
        id: 'integrations-openai-api-key-disconnect-success',
        title: message,
        summary: 'Provider credential disconnect status from Integrations.',
      })
    } catch (error) {
      const message =
        error instanceof Error ? error.message : 'disconnect_failed'
      setSaveError(message)
      recordIntegrationsStatusHistory({
        id: 'integrations-openai-api-key-disconnect-failed',
        level: 'error',
        title: message,
        summary: 'Provider credential disconnect status from Integrations.',
      })
    } finally {
      setSaving(false)
    }
  }

  const runBuyerInterestSync = async (mode: 'preview' | 'import') => {
    if (!editingProvider || editingProvider.provider_id !== 'ebay') {
      return
    }
    setBuyerInterestWorking(true)
    setBuyerInterestError(null)
    setBuyerInterestResult(null)
    try {
      const parsed = JSON.parse(form.buyerInterestPayload) as unknown
      const response = await fetch(
        `/api/providers/ebay/buyer-interest/${mode}`,
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(parsed),
        }
      )
      if (!response.ok) {
        throw new Error(`buyer_interest_${mode}_failed_${response.status}`)
      }
      const payload = (await response.json()) as BuyerInterestSyncResult
      setBuyerInterestResult(payload)
      const message =
        mode === 'preview'
          ? 'Buyer-interest preview mapped without remote write-back.'
          : 'Buyer-interest import persisted local Wishlist and Discovery state.'
      setActionMessage(message)
      recordIntegrationsStatusHistory({
        id: `integrations-buyer-interest-${mode}-success`,
        title: message,
        summary: 'Buyer-interest action status from Integrations.',
      })
    } catch (error) {
      const message =
        error instanceof SyntaxError
          ? 'Buyer-interest payload must be valid JSON.'
          : error instanceof Error
            ? error.message
            : 'buyer_interest_sync_failed'
      setBuyerInterestError(message)
      recordIntegrationsStatusHistory({
        id: `integrations-buyer-interest-${mode}-failed`,
        level: 'error',
        title: message,
        summary: 'Buyer-interest action status from Integrations.',
      })
    } finally {
      setBuyerInterestWorking(false)
    }
  }

  const previewSellerOperation = async (
    status: SellerOperationStatus,
    confirmed: boolean
  ) => {
    if (!editingProvider || editingProvider.provider_id !== 'ebay') {
      return
    }
    const action =
      status.read_available && !status.write_available ? 'sync' : 'fulfill'
    const workingKey = `${status.operation}-${confirmed ? 'confirmed' : 'preview'}`
    setSellerOperationWorking(workingKey)
    setSellerOperationError(null)
    setSellerOperationResult(null)
    setSellerOperationExecution(null)
    try {
      const response = await fetch(
        '/api/providers/ebay/seller-operations/preview',
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            operation: status.operation,
            capability: status.capability,
            action,
            confirmed,
            reference_id: `${status.operation}-local-preview`,
          }),
        }
      )
      if (!response.ok) {
        throw new Error(`seller_operation_preview_failed_${response.status}`)
      }
      const payload = (await response.json()) as SellerOperationPreviewResult
      setSellerOperationResult(payload)
      const message = payload.preview?.remote_write
        ? 'Seller operation preview is allowed only after explicit confirmation.'
        : 'Seller operation preview completed without remote write.'
      setActionMessage(message)
      recordIntegrationsStatusHistory({
        id: `integrations-seller-operation-${notificationHistoryID(status.operation)}-${confirmed ? 'confirmed-' : ''}preview-success`,
        title: message,
        summary: 'Seller operation preview status from Integrations.',
      })
    } catch (error) {
      const message =
        error instanceof Error
          ? error.message
          : 'seller_operation_preview_failed'
      setSellerOperationError(message)
      recordIntegrationsStatusHistory({
        id: `integrations-seller-operation-${notificationHistoryID(status.operation)}-${confirmed ? 'confirmed-' : ''}preview-failed`,
        level: 'error',
        title: message,
        summary: 'Seller operation preview status from Integrations.',
      })
    } finally {
      setSellerOperationWorking(null)
    }
  }

  const executeSellerOperation = async (
    status: SellerOperationStatus,
    confirmed: boolean
  ) => {
    if (!editingProvider || editingProvider.provider_id !== 'ebay') {
      return
    }
    const action =
      status.read_available && !status.write_available ? 'sync' : 'fulfill'
    const workingKey = `${status.operation}-${confirmed ? 'confirmed-execute' : 'execute'}`
    setSellerOperationWorking(workingKey)
    setSellerOperationError(null)
    setSellerOperationResult(null)
    setSellerOperationExecution(null)
    try {
      const response = await fetch(
        '/api/providers/ebay/seller-operations/execute',
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            operation: status.operation,
            capability: status.capability,
            action,
            confirmed,
            reference_id: `${status.operation}-local-execute`,
          }),
        }
      )
      const payload = (await response.json()) as SellerOperationExecuteResult
      setSellerOperationExecution(payload)
      if (!response.ok) {
        throw new Error(
          payload.execution?.blocker ??
            `seller_operation_execute_failed_${response.status}`
        )
      }
      const message = payload.execution?.local_only
        ? 'Seller operation read-only sync completed locally without remote write.'
        : 'Seller operation execute request returned without local completion.'
      setActionMessage(message)
      recordIntegrationsStatusHistory({
        id: `integrations-seller-operation-${notificationHistoryID(status.operation)}-${confirmed ? 'confirmed-' : ''}execute-success`,
        title: message,
        summary: 'Seller operation execute status from Integrations.',
      })
    } catch (error) {
      const message =
        error instanceof Error
          ? error.message
          : 'seller_operation_execute_failed'
      setSellerOperationError(message)
      recordIntegrationsStatusHistory({
        id: `integrations-seller-operation-${notificationHistoryID(status.operation)}-${confirmed ? 'confirmed-' : ''}execute-failed`,
        level: 'error',
        title: message,
        summary: 'Seller operation execute status from Integrations.',
      })
    } finally {
      setSellerOperationWorking(null)
    }
  }

  const listingLifecycleRequest = (
    command: SellerListingLifecycleCommand,
    confirmed: boolean
  ) => ({
    command,
    capability: command === 'draft' ? 'draft_only' : 'confirmed_api',
    confirmed,
    item_id: form.listingLifecycleItemId,
    draft_id: form.listingLifecycleDraftId,
    listing_id: form.listingLifecycleListingId,
    title: form.listingLifecycleTitle,
  })

  const previewListingLifecycle = async (
    command: SellerListingLifecycleCommand,
    confirmed: boolean
  ) => {
    if (!editingProvider || editingProvider.provider_id !== 'ebay') {
      return
    }
    const workingKey = `${command}-${confirmed ? 'confirmed-preview' : 'preview'}`
    setListingLifecycleWorking(workingKey)
    setListingLifecycleError(null)
    setListingLifecycleResult(null)
    setListingLifecycleExecution(null)
    try {
      const response = await fetch(
        '/api/providers/ebay/listing-lifecycle/preview',
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(listingLifecycleRequest(command, confirmed)),
        }
      )
      const payload =
        (await response.json()) as SellerListingLifecyclePreviewResult
      setListingLifecycleResult(payload)
      if (!response.ok) {
        throw new Error(
          payload.preview?.blocker ??
            `listing_lifecycle_preview_failed_${response.status}`
        )
      }
      const message = payload.preview?.remote_write
        ? 'Listing lifecycle preview requires explicit confirmation before any eBay write.'
        : 'Listing lifecycle preview completed without remote write.'
      setActionMessage(message)
      recordIntegrationsStatusHistory({
        id: `integrations-listing-lifecycle-${notificationHistoryID(command)}-${confirmed ? 'confirmed-' : ''}preview-success`,
        title: message,
        summary: 'Listing lifecycle preview status from Integrations.',
      })
    } catch (error) {
      const message =
        error instanceof Error
          ? error.message
          : 'listing_lifecycle_preview_failed'
      setListingLifecycleError(message)
      recordIntegrationsStatusHistory({
        id: `integrations-listing-lifecycle-${notificationHistoryID(command)}-${confirmed ? 'confirmed-' : ''}preview-failed`,
        level: 'error',
        title: message,
        summary: 'Listing lifecycle preview status from Integrations.',
      })
    } finally {
      setListingLifecycleWorking(null)
    }
  }

  const executeListingLifecycle = async (
    command: SellerListingLifecycleCommand,
    confirmed: boolean
  ) => {
    if (!editingProvider || editingProvider.provider_id !== 'ebay') {
      return
    }
    const workingKey = `${command}-${confirmed ? 'confirmed-execute' : 'execute'}`
    setListingLifecycleWorking(workingKey)
    setListingLifecycleError(null)
    setListingLifecycleResult(null)
    setListingLifecycleExecution(null)
    try {
      const response = await fetch(
        '/api/providers/ebay/listing-lifecycle/execute',
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(listingLifecycleRequest(command, confirmed)),
        }
      )
      const payload =
        (await response.json()) as SellerListingLifecycleExecuteResult
      setListingLifecycleExecution(payload)
      if (!response.ok) {
        throw new Error(
          payload.execution?.blocker ??
            `listing_lifecycle_execute_failed_${response.status}`
        )
      }
      const message = payload.execution?.local_only
        ? 'Listing draft was created locally without eBay remote write.'
        : 'Listing lifecycle execute request returned without remote completion.'
      setActionMessage(message)
      recordIntegrationsStatusHistory({
        id: `integrations-listing-lifecycle-${notificationHistoryID(command)}-${confirmed ? 'confirmed-' : ''}execute-success`,
        title: message,
        summary: 'Listing lifecycle execute status from Integrations.',
      })
    } catch (error) {
      const message =
        error instanceof Error
          ? error.message
          : 'listing_lifecycle_execute_failed'
      setListingLifecycleError(message)
      recordIntegrationsStatusHistory({
        id: `integrations-listing-lifecycle-${notificationHistoryID(command)}-${confirmed ? 'confirmed-' : ''}execute-failed`,
        level: 'error',
        title: message,
        summary: 'Listing lifecycle execute status from Integrations.',
      })
    } finally {
      setListingLifecycleWorking(null)
    }
  }

  const previewLandedCostPlan = async () => {
    if (!editingProvider || editingProvider.provider_id !== 'ebay') {
      return
    }
    setLandedCostWorking(true)
    setLandedCostError(null)
    setLandedCostResult(null)
    try {
      const parsed = JSON.parse(form.landedCostPayload) as unknown
      const response = await fetch('/api/commerce/landed-cost/plan', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(parsed),
      })
      const payload = (await response.json()) as LandedCostPlanResult
      setLandedCostResult(payload)
      if (!response.ok) {
        throw new Error('landed_cost_plan_failed_' + response.status)
      }
      const message =
        'Landed-cost plan previewed without mutating inventory or shipment state.'
      setActionMessage(message)
      recordIntegrationsStatusHistory({
        id: 'integrations-landed-cost-preview-success',
        title: message,
        summary: 'Landed-cost preview status from Integrations.',
      })
    } catch (error) {
      const message =
        error instanceof SyntaxError
          ? 'Landed-cost payload must be valid JSON.'
          : error instanceof Error
            ? error.message
            : 'landed_cost_plan_failed'
      setLandedCostError(message)
      recordIntegrationsStatusHistory({
        id: 'integrations-landed-cost-preview-failed',
        level: 'error',
        title: message,
        summary: 'Landed-cost preview status from Integrations.',
      })
    } finally {
      setLandedCostWorking(false)
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
          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                ref={addIntegrationButtonRef}
                type='button'
                size='icon'
                data-testid='integrations-header-add'
                aria-label='Add integration'
                title='Add integration'
                disabled={!sortedProviders.length}
                onClick={() => {
                  setPendingProviderSetup(null)
                  providerSelectorWasOpenedRef.current = true
                  setProviderSelectorOpen(true)
                }}
              >
                <Plus className='size-4' aria-hidden='true' />
              </Button>
            </TooltipTrigger>
            <TooltipContent>Add integration</TooltipContent>
          </Tooltip>
          <LanguageSwitch />
          <ThemeSwitch />
          <ConfigDrawer />
          <ProfileDropdown />
        </div>
      </Header>

      <Main fixed className='min-h-0 gap-3'>
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

        {!bootstrapError && activeProfileId ? (
          <section
            className='max-h-36 shrink-0 overflow-y-auto rounded-md border bg-muted/20 p-4 text-sm'
            data-testid='browser-companion-security'
            aria-labelledby='browser-companion-security-title'
          >
            <div className='flex flex-wrap items-start justify-between gap-3'>
              <div>
                <h2
                  id='browser-companion-security-title'
                  className='font-semibold'
                >
                  Browser Companion access
                </h2>
                <p className='mt-1 text-muted-foreground'>
                  Approve only an extension and pairing code you recognise.
                  Credentials stay hidden and can be revoked here at any time.
                </p>
              </div>
              <div className='flex gap-2'>
                <Button
                  type='button'
                  size='sm'
                  variant='outline'
                  disabled={companionWorking !== null}
                  onClick={() => void loadCompanionSecurity()}
                >
                  Refresh
                </Button>
                {companionSessions.some((session) => !session.revoked_at) ? (
                  <Button
                    type='button'
                    size='sm'
                    variant='destructive'
                    disabled={companionWorking !== null}
                    onClick={() => void revokeCompanionSessions()}
                  >
                    Revoke all
                  </Button>
                ) : null}
              </div>
            </div>

            {companionSecurityError ? (
              <p
                className='mt-3 text-destructive'
                role='alert'
                data-testid='browser-companion-security-error'
              >
                {companionSecurityError}
              </p>
            ) : null}

            <div className='mt-4 grid gap-3 lg:grid-cols-2'>
              <div>
                <h3 className='font-medium'>Pairing requests</h3>
                {companionPairingRequests.length ? (
                  <div className='mt-2 grid gap-2'>
                    {companionPairingRequests.map((request) => (
                      <article
                        key={request.request_id}
                        className='rounded-md border bg-background p-3'
                        data-testid={`browser-companion-pairing-${request.request_id}`}
                      >
                        <div className='flex flex-wrap justify-between gap-2'>
                          <div>
                            <p className='font-medium'>{request.device_name}</p>
                            <p className='text-muted-foreground'>
                              Code {request.pairing_code} · {request.origin}
                            </p>
                            <p className='mt-1 text-xs text-muted-foreground'>
                              {request.capabilities.join(', ')} · expires{' '}
                              {new Date(request.expires_at).toLocaleString()}
                            </p>
                          </div>
                          {request.status === 'pending' ? (
                            <div className='flex gap-2'>
                              <Button
                                type='button'
                                size='sm'
                                disabled={companionWorking !== null}
                                onClick={() =>
                                  void approveCompanionPairing(request)
                                }
                              >
                                Approve
                              </Button>
                              <Button
                                type='button'
                                size='sm'
                                variant='outline'
                                disabled={companionWorking !== null}
                                onClick={() =>
                                  void rejectCompanionPairing(request)
                                }
                              >
                                Reject
                              </Button>
                            </div>
                          ) : (
                            <span className='text-xs font-medium text-muted-foreground'>
                              Approved; waiting for extension
                            </span>
                          )}
                        </div>
                      </article>
                    ))}
                  </div>
                ) : (
                  <p className='mt-2 text-muted-foreground'>
                    No active pairing requests.
                  </p>
                )}
              </div>

              <div>
                <h3 className='font-medium'>Paired extensions</h3>
                {companionSessions.filter((session) => !session.revoked_at)
                  .length ? (
                  <div className='mt-2 grid gap-2'>
                    {companionSessions
                      .filter((session) => !session.revoked_at)
                      .map((session) => (
                        <article
                          key={session.id}
                          className='flex flex-wrap items-start justify-between gap-2 rounded-md border bg-background p-3'
                          data-testid={`browser-companion-session-${session.id}`}
                        >
                          <div>
                            <p className='font-medium'>{session.device_name}</p>
                            <p className='text-muted-foreground'>
                              {session.origin} · protocol{' '}
                              {session.protocol_version}
                            </p>
                            <p className='mt-1 text-xs text-muted-foreground'>
                              Last used{' '}
                              {session.last_used_at
                                ? new Date(
                                    session.last_used_at
                                  ).toLocaleString()
                                : 'never'}{' '}
                              · expires{' '}
                              {new Date(session.expires_at).toLocaleString()}
                            </p>
                          </div>
                          <Button
                            type='button'
                            size='sm'
                            variant='outline'
                            disabled={companionWorking !== null}
                            onClick={() =>
                              void revokeCompanionSessions(session.id)
                            }
                          >
                            Revoke
                          </Button>
                        </article>
                      ))}
                  </div>
                ) : (
                  <p className='mt-2 text-muted-foreground'>
                    No paired extensions for this profile.
                  </p>
                )}
              </div>
            </div>
          </section>
        ) : null}

        <div className='flex items-end justify-between gap-3 sm:items-center'>
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

        <div className='flex items-center justify-end gap-2'>
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
          <div
            className='flex min-h-0 flex-1 flex-col gap-3'
            data-testid='integrations-table-surface'
          >
            <div
              className='min-h-0 flex-1 overflow-auto rounded-md border'
              data-testid='integrations-table-scroll-body'
            >
              <Table className='min-w-[72rem] table-fixed'>
                <TableHeader className='sticky top-0 z-10 bg-background'>
                  <TableRow>
                    <TableHead className='w-[18rem]'>Provider</TableHead>
                    <TableHead className='w-[14rem]'>Category / Type</TableHead>
                    <TableHead className='w-[11rem]'>Connection</TableHead>
                    <TableHead>Actions</TableHead>
                    <TableHead className='w-[14rem]'>
                      Health / Last run
                    </TableHead>
                    <TableHead className='w-[8rem] text-right'>
                      Row actions
                    </TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {paginatedProviders.length ? (
                    paginatedProviders.map((provider) => {
                      const capabilities = capabilityLabels(provider)
                      return (
                        <TableRow
                          key={provider.provider_id}
                          data-testid={`provider-row-${provider.provider_id}`}
                          onClick={(event) => handleRowClick(provider, event)}
                          onDoubleClick={(event) =>
                            handleRowDoubleClick(provider, event)
                          }
                        >
                          <TableCell className='align-top'>
                            <div className='flex min-w-0 items-center gap-2'>
                              <div className='flex size-8 shrink-0 items-center justify-center rounded-md bg-muted p-1.5'>
                                <Store className='size-4' />
                              </div>
                              <div className='min-w-0'>
                                <p className='truncate font-medium'>
                                  {provider.display_name}
                                </p>
                                <p className='truncate text-xs text-muted-foreground'>
                                  {provider.base_domain || provider.provider_id}
                                </p>
                              </div>
                            </div>
                          </TableCell>
                          <TableCell className='align-top text-sm'>
                            <p className='truncate'>
                              {provider.integration_mode}
                            </p>
                            <p className='truncate text-xs text-muted-foreground'>
                              API Family: {provider.api_family ?? 'custom'}
                            </p>
                          </TableCell>
                          <TableCell className='align-top text-sm'>
                            {isConnected(provider, settings)
                              ? 'Connected'
                              : 'Not Connected'}
                            <p className='text-xs text-muted-foreground'>
                              Auth: {provider.auth_mode}
                            </p>
                          </TableCell>
                          <TableCell className='align-top text-sm text-muted-foreground'>
                            {capabilities.length
                              ? capabilities.join(', ')
                              : 'None'}
                          </TableCell>
                          <TableCell className='align-top text-sm'>
                            <p>
                              Health: {provider.health?.status ?? 'unknown'}
                            </p>
                            <p className='text-xs text-muted-foreground'>
                              Last run: {provider.last_run?.status ?? 'never'}
                            </p>
                          </TableCell>
                          <TableCell className='text-right align-top'>
                            <Button
                              variant='outline'
                              size='sm'
                              data-testid={`provider-open-${provider.provider_id}`}
                              onClick={(event) => {
                                event.stopPropagation()
                                openIntegration(provider)
                              }}
                            >
                              {isConnected(provider, settings)
                                ? 'Edit'
                                : 'Connect'}
                            </Button>
                          </TableCell>
                        </TableRow>
                      )
                    })
                  ) : (
                    <TableRow>
                      <TableCell colSpan={6} className='h-24 text-center'>
                        No configured integrations match current filters.
                      </TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
            </div>
            <div
              className='flex flex-col gap-2 text-sm text-muted-foreground sm:flex-row sm:items-center sm:justify-between'
              data-testid='integrations-table-pagination'
            >
              <p>
                Showing {tableRangeStart}-{tableRangeEnd} of{' '}
                {filteredProviders.length} integrations
              </p>
              <div className='flex items-center gap-2'>
                <Button
                  type='button'
                  size='sm'
                  variant='outline'
                  disabled={tablePage === 1}
                  data-testid='integrations-table-prev-page'
                  onClick={() => setTablePage((page) => Math.max(1, page - 1))}
                >
                  Previous
                </Button>
                <span data-testid='integrations-table-page-status'>
                  Page {tablePage} of {tablePageCount}
                </span>
                <Button
                  type='button'
                  size='sm'
                  variant='outline'
                  disabled={tablePage >= tablePageCount}
                  data-testid='integrations-table-next-page'
                  onClick={() =>
                    setTablePage((page) => Math.min(tablePageCount, page + 1))
                  }
                >
                  Next
                </Button>
              </div>
            </div>
          </div>
        ) : null}

        {!loading && viewMode === 'cards' ? (
          <ul className='faded-bottom no-scrollbar grid gap-4 overflow-auto pb-16 md:grid-cols-2 lg:grid-cols-3'>
            {filteredProviders.length > 0 ? (
              filteredProviders.map((provider) => (
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
                      <span className='rounded bg-muted px-2 py-0.5'>
                        Search
                      </span>
                    ) : null}
                    {provider.capabilities.stock_observation ? (
                      <span className='rounded bg-muted px-2 py-0.5'>
                        Stock
                      </span>
                    ) : null}
                    {provider.capabilities.pricing ? (
                      <span className='rounded bg-muted px-2 py-0.5'>
                        Pricing
                      </span>
                    ) : null}
                    {provider.capabilities.assistant ? (
                      <span className='rounded bg-muted px-2 py-0.5'>
                        Assistant
                      </span>
                    ) : null}
                    {provider.capabilities.image_help ? (
                      <span className='rounded bg-muted px-2 py-0.5'>
                        Image help
                      </span>
                    ) : null}
                    {provider.capabilities.media_capture ? (
                      <span className='rounded bg-muted px-2 py-0.5'>
                        Media capture
                      </span>
                    ) : null}
                    {provider.capabilities.text_capture ? (
                      <span className='rounded bg-muted px-2 py-0.5'>
                        Text capture
                      </span>
                    ) : null}
                  </div>
                </li>
              ))
            ) : (
              <li
                className='rounded-lg border border-dashed p-6 text-center text-sm text-muted-foreground md:col-span-2 lg:col-span-3'
                data-testid='integrations-cards-empty-state'
              >
                No configured integrations match current filters.
              </li>
            )}
          </ul>
        ) : null}
      </Main>

      {providerSelectorOpen ? (
        <Dialog open onOpenChange={setProviderSelectorOpen}>
          <DialogContent
            className='max-h-[90vh] overflow-hidden'
            data-testid='integrations-provider-selector'
          >
            <DialogHeader>
              <DialogTitle>Add Integration</DialogTitle>
              <DialogDescription>
                Search registry providers and review setup requirements before
                opening provider-specific setup.
              </DialogDescription>
            </DialogHeader>
            <div className='grid min-h-0 min-w-0 gap-3'>
              <div className='grid gap-2 sm:grid-cols-[1fr_12rem]'>
                <div className='relative'>
                  <SearchCheck className='pointer-events-none absolute top-2.5 left-2.5 size-4 text-muted-foreground' />
                  <Input
                    className='pl-8'
                    data-testid='integrations-provider-selector-search'
                    placeholder='Search name, domain, capability, auth, or status'
                    value={providerCatalogSearchTerm}
                    onChange={(event) =>
                      setProviderCatalogSearchTerm(event.target.value)
                    }
                  />
                </div>
                <Select
                  value={providerCatalogCategory}
                  onValueChange={setProviderCatalogCategory}
                >
                  <SelectTrigger data-testid='integrations-provider-selector-category'>
                    <SelectValue placeholder='Category' />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value='all'>All categories</SelectItem>
                    {providerCatalogCategories.map((category) => (
                      <SelectItem key={category} value={category}>
                        {category}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className='grid max-h-[58vh] min-h-0 min-w-0 gap-2 overflow-x-hidden overflow-y-auto pr-1'>
                {filteredProviderCatalogProviders.length ? (
                  filteredProviderCatalogProviders.map((provider) => (
                    <Button
                      key={provider.provider_id}
                      type='button'
                      variant='outline'
                      className='h-auto w-full min-w-0 justify-start overflow-hidden px-3 py-3 text-left'
                      data-testid={`integrations-provider-selector-option-${provider.provider_id}`}
                      onClick={() => selectProviderForSetup(provider)}
                      onKeyDown={(event) => {
                        if (event.key === 'Enter' || event.key === ' ') {
                          event.preventDefault()
                          selectProviderForSetup(provider)
                        }
                      }}
                    >
                      <span className='flex min-w-0 flex-1 flex-col items-start gap-2'>
                        <span className='flex w-full min-w-0 flex-wrap items-start justify-between gap-2'>
                          <span className='min-w-0'>
                            <span className='block truncate font-medium'>
                              {provider.display_name}
                            </span>
                            <span className='block truncate text-xs text-muted-foreground'>
                              {provider.base_domain || provider.provider_id}
                            </span>
                          </span>
                          <span
                            className={`rounded px-2 py-0.5 text-xs ${providerCatalogStatusClassName(provider)}`}
                            data-testid={`integrations-provider-selector-status-${provider.provider_id}`}
                          >
                            {provider.health?.state ??
                              provider.health?.status ??
                              provider.state}
                          </span>
                        </span>
                        <span className='flex flex-wrap gap-1 text-xs text-muted-foreground'>
                          <span className='rounded bg-muted px-2 py-0.5'>
                            {provider.provider_category ??
                              provider.provider_type ??
                              'uncategorized'}
                          </span>
                          <span className='rounded bg-muted px-2 py-0.5'>
                            {provider.auth_mode}
                          </span>
                          <span className='rounded bg-muted px-2 py-0.5'>
                            {provider.integration_mode}
                          </span>
                        </span>
                        <span className='line-clamp-2 w-full min-w-0 text-left text-xs break-words text-muted-foreground'>
                          {provider.setup_instructions ??
                            'Review provider setup requirements before saving.'}
                        </span>
                      </span>
                    </Button>
                  ))
                ) : (
                  <div
                    className='rounded-md border border-dashed p-4 text-sm text-muted-foreground'
                    data-testid='integrations-provider-selector-empty'
                  >
                    No registry providers match current Add Integration filters.
                  </div>
                )}
              </div>
            </div>
          </DialogContent>
        </Dialog>
      ) : null}

      <Dialog
        open={editingProvider !== null}
        onOpenChange={(open) => {
          if (!open) {
            closeIntegration()
          }
        }}
      >
        <DialogContent className='top-4 max-h-[90vh] translate-y-0 scroll-py-12 overflow-y-auto sm:max-w-2xl'>
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
                <p data-testid='provider-detail-category'>
                  Category:{' '}
                  {editingProvider.provider_category ?? 'uncategorized'}
                  {' / '}
                  Type: {editingProvider.provider_type ?? 'unknown'}
                </p>
                <p data-testid='provider-detail-api-family'>
                  API Family: {editingProvider.api_family ?? 'custom'}
                </p>
                <p data-testid='provider-detail-api-support-profile'>
                  Support Profile:{' '}
                  {editingProvider.api_support_profile ?? 'unknown'}
                </p>
                <p data-testid='provider-detail-config-schema'>
                  Config Schema:{' '}
                  {editingProvider.config_schema_ref ?? 'provider-specific'}
                </p>
                <p data-testid='provider-detail-workflows'>
                  Workflows:{' '}
                  {editingProvider.workflow_refs?.length
                    ? editingProvider.workflow_refs.join(', ')
                    : 'none'}
                </p>
                <p data-testid='provider-detail-manifest-actions'>
                  Manifest Actions:{' '}
                  {providerManifestActions(editingProvider).join(', ') ||
                    'none'}
                </p>
                <p>
                  Readiness:{' '}
                  {editingProvider.health?.state ?? editingProvider.state}
                </p>
                <p>Health: {editingProvider.health?.status ?? 'unknown'}</p>
                {editingProvider.health?.message ? (
                  <p>Message: {editingProvider.health.message}</p>
                ) : null}
                {editingProvider.health?.last_error ? (
                  <p>Last error: {editingProvider.health.last_error}</p>
                ) : null}
                {typeof editingProvider.health?.retry_after_seconds ===
                'number' ? (
                  <p>
                    Retry after: {editingProvider.health.retry_after_seconds}{' '}
                    seconds
                  </p>
                ) : null}
                {editingProvider.health?.next_action ? (
                  <p>Next action: {editingProvider.health.next_action}</p>
                ) : null}
                <p>Last run: {editingProvider.last_run?.status ?? 'never'}</p>
                <p>
                  Last checked:{' '}
                  {editingProvider.health?.last_checked_at ??
                    editingProvider.last_run?.finished_at ??
                    'n/a'}
                </p>
                {actionMessage ? <p>{actionMessage}</p> : null}
              </div>

              <div className='rounded-md border p-3 text-xs text-muted-foreground'>
                {editingProvider.setup_instructions ??
                  'Enter provider details, validate health, and save configuration.'}
              </div>

              {editingProvider.provider_id === 'telegram' ? (
                <section
                  className='rounded-md border p-3 text-xs'
                  data-testid='telegram-capture-status-panel'
                >
                  <div className='flex flex-wrap items-start justify-between gap-3'>
                    <div>
                      <p className='font-medium'>
                        Telegram Cabinet Agent channel
                      </p>
                      <p className='text-muted-foreground'>
                        A validated bot and exact private-chat pairing gate
                        Agent intake before preview and confirmation.
                      </p>
                    </div>
                    <span
                      className='rounded bg-muted px-2 py-1 text-muted-foreground'
                      data-testid='telegram-capture-next-action'
                    >
                      {editingProvider.health?.next_action ??
                        editingProvider.setup_status?.next_action ??
                        (editingProvider.auth_methods?.sender_chat?.connected
                          ? 'talk_to_cabinet'
                          : 'test_connection')}
                    </span>
                  </div>
                  <div className='mt-3 grid gap-2 sm:grid-cols-2'>
                    <p data-testid='telegram-capture-auth-mode'>
                      Auth method: write-only bot token + private pairing
                    </p>
                    <p data-testid='telegram-capture-sender-chat-state'>
                      Private chat state:{' '}
                      {editingProvider.auth_methods?.sender_chat?.state ??
                        (editingProvider.auth_methods?.sender_chat?.connected
                          ? 'paired'
                          : 'pairing_required')}
                    </p>
                    <p data-testid='telegram-capture-api-family'>
                      API family:{' '}
                      {editingProvider.api_family ?? 'messaging_channel'}
                    </p>
                    <p data-testid='telegram-capture-support-profile'>
                      Support profile:{' '}
                      {editingProvider.api_support_profile ??
                        'bot_long_polling_private_pairing_v1'}
                    </p>
                    <p data-testid='telegram-capture-bot-token-state'>
                      Bot token state:{' '}
                      {editingProvider.auth_methods?.bot_token?.state ??
                        editingProvider.setup_status?.bot_token_state ??
                        'setup_needed'}
                    </p>
                    <p data-testid='telegram-capture-polling-state'>
                      Transport: outbound long polling / public listener: no
                    </p>
                    <p data-testid='telegram-capture-runtime-proof-state'>
                      Runtime proof:{' '}
                      {editingProvider.setup_status?.runtime_proof ??
                        editingProvider.health?.state ??
                        'pending_live_channel_check'}
                    </p>
                    <p data-testid='telegram-capture-health-message'>
                      Diagnostics:{' '}
                      {editingProvider.health?.message ??
                        'Validate Telegram setup before live channel intake.'}
                    </p>
                  </div>
                </section>
              ) : null}

              {editingProvider.provider_id === 'openai' ? (
                <div className='space-y-4' data-testid='openai-config-dialog'>
                  <section
                    className='rounded-md border p-3'
                    data-testid='openai-browser-auth-section'
                  >
                    <div className='flex items-center justify-between gap-3'>
                      <div>
                        <p className='font-medium'>Browser Auth</p>
                        <p className='text-xs text-muted-foreground'>
                          Use a verified OpenAI/Codex browser-auth artifact when
                          available.
                        </p>
                      </div>
                      <span
                        className='rounded bg-muted px-2 py-1 text-xs'
                        data-testid='openai-browser-auth-status'
                      >
                        {settings['openai.browser_auth_state'] ??
                          editingProvider.auth_methods?.browser_auth?.state ??
                          'setup_needed'}
                      </span>
                    </div>
                    <p
                      className='mt-2 text-xs text-muted-foreground'
                      data-testid='openai-browser-auth-setup-needed'
                    >
                      Browser Auth is setup-needed until Cabinet can verify an
                      acquired callback/artifact. Navigation alone never marks
                      OpenAI connected.
                    </p>
                    <div className='mt-3 flex flex-wrap gap-2'>
                      <Button
                        type='button'
                        size='sm'
                        disabled
                        data-testid='openai-browser-auth-connect'
                      >
                        Only available on this PC
                      </Button>
                      <Button
                        type='button'
                        size='sm'
                        variant='outline'
                        data-testid='openai-browser-auth-disconnect'
                        onClick={() => {
                          setSettings((prev) => ({
                            ...prev,
                            'openai.browser_auth_state': 'disconnected',
                            'openai.active_auth_method':
                              prev['openai.active_auth_method'] ===
                              'browser_auth'
                                ? ''
                                : prev['openai.active_auth_method'],
                          }))
                        }}
                      >
                        Disconnect
                      </Button>
                    </div>
                  </section>

                  <section
                    className='rounded-md border p-3'
                    data-testid='openai-api-key-section'
                  >
                    <div className='flex items-center justify-between gap-3'>
                      <div>
                        <p className='font-medium'>API key</p>
                        <p className='text-xs text-muted-foreground'>
                          Validate and save an OpenAI API key for Cabinet
                          assistant, image, and content workflows.
                        </p>
                      </div>
                      <span
                        className='rounded bg-muted px-2 py-1 text-xs'
                        data-testid='openai-api-key-status'
                      >
                        {editingProvider.has_token ||
                        settings['openai.active_auth_method'] === 'api_key'
                          ? 'connected'
                          : 'setup_needed'}
                      </span>
                    </div>
                    {editingProvider.has_token && !replaceToken ? (
                      <div className='mt-3 rounded-md border bg-muted/20 p-3 text-xs'>
                        <p>API key on file.</p>
                        <p className='text-muted-foreground'>
                          Existing key is hidden. Replace it to update the
                          active API-key method.
                        </p>
                        <Button
                          size='sm'
                          variant='outline'
                          className='mt-2'
                          data-testid='replace-token'
                          onClick={() => setReplaceToken(true)}
                        >
                          Replace API key
                        </Button>
                      </div>
                    ) : (
                      <div className='mt-3 space-y-2'>
                        <Label htmlFor='provider-token'>OpenAI API key</Label>
                        <Input
                          ref={tokenInputRef}
                          id='provider-token'
                          type='password'
                          data-testid='provider-token-input'
                          placeholder='OpenAI API key'
                          value={form.token}
                          aria-invalid={tokenFieldError ? 'true' : undefined}
                          aria-describedby={
                            tokenFieldError ? 'provider-token-error' : undefined
                          }
                          onChange={(e) => {
                            setForm((prev) => ({
                              ...prev,
                              token: e.target.value,
                            }))
                            if (tokenFieldError) {
                              setTokenFieldError(null)
                            }
                          }}
                        />
                        {tokenFieldError ? (
                          <p
                            id='provider-token-error'
                            className='text-xs text-destructive'
                          >
                            {tokenFieldError}
                          </p>
                        ) : null}
                      </div>
                    )}
                    <div className='mt-3 flex flex-wrap gap-2'>
                      <Button
                        type='button'
                        size='sm'
                        variant='outline'
                        onClick={() => void validateProvider()}
                        disabled={validating}
                        data-testid='openai-api-key-validate'
                      >
                        {validating ? 'Validating...' : 'Validate'}
                      </Button>
                      <Button
                        type='button'
                        size='sm'
                        data-testid='openai-api-key-connect'
                        onClick={() => void saveIntegration()}
                        disabled={saving}
                      >
                        Connect
                      </Button>
                      <Button
                        type='button'
                        size='sm'
                        variant='outline'
                        data-testid='openai-api-key-disconnect'
                        onClick={() => void disconnectOpenAIApiKey()}
                        disabled={saving}
                      >
                        Disconnect
                      </Button>
                    </div>
                  </section>

                  <section
                    className='rounded-md border p-3'
                    data-testid='openai-test-section'
                  >
                    <p className='font-medium'>Test OpenAI</p>
                    <p className='text-xs text-muted-foreground'>
                      Uses the active connected method. If no method is
                      verified, Cabinet shows setup-needed instead of pretending
                      readiness.
                    </p>
                    <div className='mt-3 grid gap-2 sm:grid-cols-2'>
                      <div className='space-y-2'>
                        <Label htmlFor='openai-test-model'>Test model</Label>
                        <Select
                          value={form.openAiModel}
                          onValueChange={(value) =>
                            setForm((prev) => ({
                              ...prev,
                              openAiModel: value,
                            }))
                          }
                        >
                          <SelectTrigger
                            id='openai-test-model'
                            data-testid='openai-test-model'
                          >
                            <SelectValue />
                          </SelectTrigger>
                          <SelectContent>
                            {(
                              editingProvider.model_options ?? [
                                'gpt-4o-mini',
                                'gpt-4.1-mini',
                              ]
                            ).map((model) => (
                              <SelectItem key={model} value={model}>
                                {model}
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                      </div>
                      <div className='space-y-2'>
                        <Label htmlFor='openai-active-method'>
                          Active method
                        </Label>
                        <Input
                          id='openai-active-method'
                          data-testid='openai-active-method'
                          readOnly
                          value={
                            settings['openai.active_auth_method'] ===
                            'browser_auth'
                              ? 'Browser Auth'
                              : settings['openai.active_auth_method'] ===
                                    'api_key' || editingProvider.has_token
                                ? 'API key'
                                : 'None connected'
                          }
                        />
                      </div>
                    </div>
                    <Label className='mt-3 block' htmlFor='openai-test-prompt'>
                      Test prompt
                    </Label>
                    <Input
                      className='mt-2'
                      id='openai-test-prompt'
                      data-testid='openai-test-prompt'
                      aria-label='OpenAI test prompt'
                      value={form.openAiTestPrompt}
                      onChange={(e) =>
                        setForm((prev) => ({
                          ...prev,
                          openAiTestPrompt: e.target.value,
                        }))
                      }
                    />
                    <Button
                      type='button'
                      size='sm'
                      variant='outline'
                      className='mt-3'
                      data-testid='openai-test-run'
                      disabled={!isConnected(editingProvider, settings)}
                      onClick={() =>
                        setActionMessage(
                          'OpenAI test requires a verified active method before Cabinet runs provider calls.'
                        )
                      }
                    >
                      Test
                    </Button>
                  </section>
                </div>
              ) : editingProvider.provider_id === 'telegram' ? (
                <section
                  className='space-y-4 rounded-md border p-3'
                  data-testid='telegram-local-polling-setup'
                >
                  <div>
                    <p className='font-medium'>Connect Cabinet to Telegram</p>
                    <ol className='mt-2 list-decimal space-y-1 pl-5 text-xs text-muted-foreground'>
                      <li>
                        Open @BotFather in Telegram and create a bot with
                        /newbot.
                      </li>
                      <li>
                        Paste the bot token below and test the connection.
                      </li>
                      <li>
                        Create a short-lived pairing code, then send its /start
                        command to your bot in a private chat.
                      </li>
                    </ol>
                    <p className='mt-2 text-xs text-muted-foreground'>
                      Outbound long polling only. No public webhook or listener
                      is opened by Cabinet.
                    </p>
                  </div>

                  <div className='space-y-2'>
                    {telegramStatus?.bot_token_present && !replaceToken ? (
                      <div className='rounded-md border bg-muted/20 p-3 text-xs'>
                        <p>Bot token on file.</p>
                        <p className='text-muted-foreground'>
                          The existing token is write-only and cannot be read
                          back from Cabinet.
                        </p>
                        <Button
                          type='button'
                          size='sm'
                          variant='outline'
                          className='mt-2'
                          data-testid='telegram-replace-token'
                          onClick={() => setReplaceToken(true)}
                        >
                          Rotate bot token
                        </Button>
                      </div>
                    ) : (
                      <>
                        <Label htmlFor='telegram-bot-token'>Bot token</Label>
                        <Input
                          id='telegram-bot-token'
                          type='password'
                          autoComplete='new-password'
                          data-testid='telegram-bot-token'
                          placeholder='Paste BotFather token'
                          value={form.token}
                          onChange={(event) =>
                            setForm((previous) => ({
                              ...previous,
                              token: event.target.value,
                            }))
                          }
                        />
                      </>
                    )}
                    <div className='flex flex-wrap gap-2'>
                      <Button
                        type='button'
                        variant='outline'
                        data-testid='telegram-test-connection'
                        disabled={telegramWorking !== null}
                        onClick={() => void testTelegramConnection()}
                      >
                        {telegramWorking === 'test'
                          ? 'Testing...'
                          : 'Test connection'}
                      </Button>
                      {telegramStatus?.webhook_conflict ? (
                        <Button
                          type='button'
                          variant='outline'
                          data-testid='telegram-resolve-webhook'
                          disabled={telegramWorking !== null}
                          onClick={() => void resolveTelegramWebhook()}
                        >
                          Remove existing webhook
                        </Button>
                      ) : null}
                    </div>
                    {telegramBotUsername ? (
                      <p
                        className='text-xs text-muted-foreground'
                        data-testid='telegram-bot-identity'
                      >
                        Validated bot: @{telegramBotUsername}
                      </p>
                    ) : null}
                  </div>

                  <div className='rounded-md bg-muted/30 p-3 text-xs'>
                    <p className='font-medium'>Private-chat pairing</p>
                    <p className='mt-1 text-muted-foreground'>
                      Cabinet learns the exact numeric user and private-chat IDs
                      from Telegram. You do not enter or guess them manually.
                    </p>
                    <Button
                      type='button'
                      size='sm'
                      variant='outline'
                      className='mt-3'
                      data-testid='telegram-create-pairing-code'
                      disabled={
                        telegramWorking !== null ||
                        !telegramStatus?.bot_token_present ||
                        telegramStatus.webhook_conflict
                      }
                      onClick={() => void createTelegramPairingCode()}
                    >
                      {telegramStatus?.paired
                        ? 'Re-pair chat'
                        : 'Create pairing code'}
                    </Button>
                    {telegramPairing ? (
                      <div className='mt-3 rounded border bg-background p-3'>
                        <p data-testid='telegram-pairing-code'>
                          /start {telegramPairing.code}
                        </p>
                        <p
                          className='mt-1 text-muted-foreground'
                          data-testid='telegram-pairing-expiry'
                        >
                          Single-use code expires{' '}
                          {new Date(
                            telegramPairing.expires_at
                          ).toLocaleTimeString()}
                          .
                        </p>
                      </div>
                    ) : null}
                    {telegramStatus?.paired ? (
                      <p className='mt-3 text-muted-foreground'>
                        Paired private chat ending in{' '}
                        {telegramStatus.chat_id?.slice(-4)}.
                      </p>
                    ) : null}
                  </div>

                  <div className='grid gap-2 text-xs sm:grid-cols-2'>
                    <p>State: {telegramStatus?.state ?? 'not connected'}</p>
                    <p>Transport: outbound long polling</p>
                    <p>
                      Last success: {telegramStatus?.last_success_at ?? 'never'}
                    </p>
                    <p>
                      Last update: {telegramStatus?.last_update_id ?? 'none'}
                    </p>
                    {telegramStatus?.last_error_code ? (
                      <p className='sm:col-span-2'>
                        Last error: {telegramStatus.last_error_code}
                      </p>
                    ) : null}
                  </div>

                  <div className='flex flex-wrap gap-2'>
                    {telegramStatus?.paired ? (
                      <Button
                        type='button'
                        size='sm'
                        variant='outline'
                        data-testid='telegram-pause-toggle'
                        disabled={telegramWorking !== null}
                        onClick={() =>
                          void setTelegramPaused(!telegramStatus.paused)
                        }
                      >
                        {telegramStatus.paused ? 'Resume' : 'Pause'}
                      </Button>
                    ) : null}
                    <Button
                      type='button'
                      size='sm'
                      variant='outline'
                      data-testid='telegram-refresh-status'
                      disabled={telegramWorking !== null}
                      onClick={() => void loadTelegramStatus()}
                    >
                      Refresh status
                    </Button>
                    {telegramStatus?.bot_token_present ? (
                      <Button
                        type='button'
                        size='sm'
                        variant='destructive'
                        data-testid='telegram-disconnect'
                        disabled={telegramWorking !== null}
                        onClick={() => void disconnectTelegram()}
                      >
                        Disconnect
                      </Button>
                    ) : null}
                  </div>
                </section>
              ) : editingProvider.setup_schema?.fields?.length ? (
                <>
                  <section
                    className='grid gap-3 rounded-md border p-3 sm:grid-cols-2'
                    data-testid='integration-schema-form'
                  >
                    {(editingProvider.setup_schema.fields ?? [])
                      .filter((field) =>
                        setupFieldVisible(field, setupSchemaValues)
                      )
                      .map((field) => {
                        const fieldID = setupFieldID(field)
                        const fieldValue = setupSchemaValues[field.key] ?? ''
                        const updateField = (value: string) =>
                          setSetupSchemaValues((prev) => ({
                            ...prev,
                            [field.key]: value,
                          }))
                        if (
                          field.write_only &&
                          editingProvider.has_token &&
                          !replaceToken
                        ) {
                          return (
                            <div
                              key={field.key}
                              className='space-y-2 sm:col-span-2'
                            >
                              <p className='text-sm font-medium'>
                                {field.label}
                              </p>
                              <div className='rounded-md border bg-muted/20 p-3 text-xs'>
                                <p>Credential on file.</p>
                                <p className='text-muted-foreground'>
                                  Existing value is hidden. Replace it to update
                                  this provider secret.
                                </p>
                                <Button
                                  size='sm'
                                  variant='outline'
                                  className='mt-2'
                                  data-testid='replace-token'
                                  onClick={() => setReplaceToken(true)}
                                >
                                  Replace credential
                                </Button>
                              </div>
                            </div>
                          )
                        }
                        return (
                          <div
                            key={field.key}
                            className={
                              field.type === 'textarea' ||
                              field.type === 'secret' ||
                              field.type === 'browser-auth-status' ||
                              field.type === 'oauth-connect'
                                ? 'space-y-2 sm:col-span-2'
                                : 'space-y-2'
                            }
                          >
                            <Label htmlFor={fieldID}>
                              {field.label}
                              {field.required ? ' *' : ''}
                            </Label>
                            {field.type === 'select' &&
                            field.options?.length ? (
                              <Select
                                value={fieldValue}
                                onValueChange={updateField}
                                disabled={field.read_only}
                              >
                                <SelectTrigger
                                  id={fieldID}
                                  data-testid={`provider-schema-field-${field.key}`}
                                >
                                  <SelectValue
                                    placeholder={field.placeholder}
                                  />
                                </SelectTrigger>
                                <SelectContent>
                                  {field.options.map((option) => (
                                    <SelectItem
                                      key={option.value}
                                      value={option.value}
                                    >
                                      {option.label}
                                    </SelectItem>
                                  ))}
                                </SelectContent>
                              </Select>
                            ) : field.type === 'textarea' ? (
                              <textarea
                                id={fieldID}
                                className='min-h-24 w-full rounded-md border bg-background p-2 text-sm'
                                data-testid={`provider-schema-field-${field.key}`}
                                placeholder={field.placeholder}
                                value={fieldValue}
                                readOnly={field.read_only}
                                onChange={(e) => updateField(e.target.value)}
                              />
                            ) : field.type === 'checkbox' ? (
                              <input
                                id={fieldID}
                                type='checkbox'
                                data-testid={`provider-schema-field-${field.key}`}
                                checked={fieldValue === 'true'}
                                disabled={field.read_only}
                                onChange={(e) =>
                                  updateField(
                                    e.target.checked ? 'true' : 'false'
                                  )
                                }
                              />
                            ) : field.type === 'oauth-connect' ||
                              field.type === 'browser-auth-status' ? (
                              <Input
                                id={fieldID}
                                data-testid={`provider-schema-field-${field.key}`}
                                readOnly
                                value={fieldValue || 'setup_needed'}
                              />
                            ) : (
                              <Input
                                id={fieldID}
                                type={
                                  field.type === 'secret'
                                    ? 'password'
                                    : field.type === 'number'
                                      ? 'number'
                                      : field.type === 'url'
                                        ? 'url'
                                        : 'text'
                                }
                                data-testid={`provider-schema-field-${field.key}`}
                                placeholder={field.placeholder}
                                value={fieldValue}
                                readOnly={field.read_only}
                                onChange={(e) => updateField(e.target.value)}
                              />
                            )}
                            {field.helper_text ? (
                              <p className='text-xs text-muted-foreground'>
                                {field.helper_text}
                              </p>
                            ) : null}
                          </div>
                        )
                      })}
                  </section>
                  {editingProvider.provider_id === 'ebay' ? (
                    <section
                      className='rounded-md border p-3 text-xs'
                      data-testid='ebay-setup-status-panel'
                    >
                      {(() => {
                        const setupStatus = editingProvider.setup_status
                        const tokenState =
                          editingProvider.has_token && !replaceToken
                            ? 'stored token on file'
                            : form.token.trim()
                              ? 'new token pending save'
                              : setupStatus?.token_state === 'stored'
                                ? 'stored token on file'
                                : 'token required'
                        const setupNextAction =
                          setupStatus?.next_action ??
                          editingProvider.health?.next_action
                        const validationStatus =
                          setupStatus?.validation_status ??
                          editingProvider.health?.status ??
                          'unknown'
                        const healthState =
                          setupStatus?.health_state ??
                          editingProvider.health?.state ??
                          editingProvider.state
                        return (
                          <>
                            <div className='flex flex-wrap items-start justify-between gap-3'>
                              <div>
                                <p className='font-medium'>eBay setup status</p>
                                <p className='text-muted-foreground'>
                                  Verify credentials, marketplace, token state,
                                  and provider health before running eBay query
                                  sets.
                                </p>
                              </div>
                              <span
                                className='rounded bg-muted px-2 py-1 text-muted-foreground'
                                data-testid='ebay-setup-next-action'
                              >
                                {formatEbaySetupNextAction(setupNextAction)}
                              </span>
                            </div>
                            <div className='mt-3 grid gap-2 sm:grid-cols-2'>
                              <p data-testid='ebay-setup-auth-mode'>
                                Auth mode:{' '}
                                {setupStatus?.auth_mode ??
                                  editingProvider.auth_mode}
                              </p>
                              <p data-testid='ebay-setup-marketplace'>
                                Marketplace / Region:{' '}
                                {setupStatus?.marketplace ||
                                  form.marketplace ||
                                  'unset'}
                              </p>
                              <p data-testid='ebay-setup-token-state'>
                                Token state: {tokenState}
                              </p>
                              <p data-testid='ebay-setup-health-state'>
                                Validation status: {validationStatus}
                              </p>
                              <p data-testid='ebay-setup-readiness-state'>
                                Health state: {healthState}
                              </p>
                              <p data-testid='ebay-setup-base-url-override'>
                                {formatEbaySetupBaseURLState(
                                  setupStatus?.base_url_set
                                )}
                              </p>
                            </div>
                          </>
                        )
                      })()}
                    </section>
                  ) : null}
                  {editingProvider.provider_id === 'ebay' ? (
                    <div className='space-y-3'>
                      <section
                        className='rounded-md border p-3'
                        data-testid='ebay-seller-operations-panel'
                      >
                        <div className='flex flex-wrap items-start justify-between gap-3'>
                          <div>
                            <p className='font-medium'>Seller Operations</p>
                            <p className='text-xs text-muted-foreground'>
                              Messages, notifications, sold orders, fulfilment,
                              and offers stay blocked until exact eBay API
                              support is verified.
                            </p>
                          </div>
                          <span
                            className='rounded bg-muted px-2 py-1 text-xs text-muted-foreground'
                            data-testid='ebay-seller-operations-safe-mode'
                          >
                            External writes require confirmation
                          </span>
                        </div>
                        <div className='mt-3 grid gap-2 sm:grid-cols-2'>
                          {(editingProvider.seller_operations ?? []).map(
                            (status) => (
                              <div
                                key={status.operation}
                                className='rounded-md bg-muted/40 p-3 text-xs'
                                data-testid={`ebay-seller-operation-${status.operation}`}
                              >
                                <div className='flex items-center justify-between gap-2'>
                                  <span className='font-medium'>
                                    {sellerOperationLabel(status.operation)}
                                  </span>
                                  <span>{sellerOperationState(status)}</span>
                                </div>
                                <p className='mt-1 text-muted-foreground'>
                                  Read: {status.read_available ? 'yes' : 'no'} /
                                  Write: {status.write_available ? 'yes' : 'no'}
                                </p>
                                <p className='mt-1 text-muted-foreground'>
                                  {status.blocker ??
                                    (status.confirmation_required
                                      ? 'confirmation_required'
                                      : 'available')}
                                </p>
                                <div
                                  className='mt-3 flex flex-wrap gap-2'
                                  data-testid={`ebay-seller-operation-preview-controls-${status.operation}`}
                                >
                                  <Button
                                    type='button'
                                    size='sm'
                                    variant='outline'
                                    data-testid={`ebay-seller-operation-preview-${status.operation}`}
                                    disabled={
                                      sellerOperationWorking !== null ||
                                      (!status.read_available &&
                                        !status.write_available)
                                    }
                                    onClick={() =>
                                      void previewSellerOperation(status, false)
                                    }
                                  >
                                    Preview
                                  </Button>
                                  <Button
                                    type='button'
                                    size='sm'
                                    variant='outline'
                                    data-testid={`ebay-seller-operation-confirm-${status.operation}`}
                                    disabled={
                                      sellerOperationWorking !== null ||
                                      !status.write_available ||
                                      !status.confirmation_required
                                    }
                                    onClick={() =>
                                      void previewSellerOperation(status, true)
                                    }
                                  >
                                    Confirm Preview
                                  </Button>
                                  <Button
                                    type='button'
                                    size='sm'
                                    variant='outline'
                                    data-testid={`ebay-seller-operation-execute-${status.operation}`}
                                    disabled={
                                      sellerOperationWorking !== null ||
                                      !status.read_available ||
                                      status.write_available
                                    }
                                    onClick={() =>
                                      void executeSellerOperation(status, false)
                                    }
                                  >
                                    Sync
                                  </Button>
                                  <Button
                                    type='button'
                                    size='sm'
                                    variant='outline'
                                    data-testid={`ebay-seller-operation-confirm-execute-${status.operation}`}
                                    disabled={
                                      sellerOperationWorking !== null ||
                                      !status.write_available ||
                                      !status.confirmation_required
                                    }
                                    onClick={() =>
                                      void executeSellerOperation(status, true)
                                    }
                                  >
                                    Confirm Execute
                                  </Button>
                                </div>
                              </div>
                            )
                          )}
                        </div>
                        {sellerOperationError ? (
                          <p
                            className='mt-3 text-xs text-destructive'
                            data-testid='ebay-seller-operation-preview-error'
                          >
                            {sellerOperationError}
                          </p>
                        ) : null}
                        {sellerOperationResult?.preview ? (
                          <div
                            className='mt-3 rounded-md bg-muted/40 p-3 text-xs'
                            data-testid='ebay-seller-operation-preview-result'
                          >
                            <p className='font-medium'>
                              Preview:{' '}
                              {sellerOperationLabel(
                                sellerOperationResult.preview.operation ?? ''
                              )}
                            </p>
                            <p className='mt-1 text-muted-foreground'>
                              Allowed:{' '}
                              {sellerOperationResult.preview.allowed
                                ? 'yes'
                                : 'no'}{' '}
                              / Remote write:{' '}
                              {sellerOperationResult.preview.remote_write
                                ? 'yes'
                                : 'no'}
                            </p>
                            <p className='mt-1 text-muted-foreground'>
                              {sellerOperationResult.preview.blocker ??
                                'No blocker'}
                            </p>
                          </div>
                        ) : null}
                        {sellerOperationExecution?.execution ? (
                          <div
                            className='mt-3 rounded-md bg-muted/40 p-3 text-xs'
                            data-testid='ebay-seller-operation-execute-result'
                          >
                            <p className='font-medium'>
                              Execute:{' '}
                              {sellerOperationLabel(
                                sellerOperationExecution.execution.operation ??
                                  ''
                              )}
                            </p>
                            <p className='mt-1 text-muted-foreground'>
                              Executed:{' '}
                              {sellerOperationExecution.execution.executed
                                ? 'yes'
                                : 'no'}{' '}
                              / Local only:{' '}
                              {sellerOperationExecution.execution.local_only
                                ? 'yes'
                                : 'no'}
                            </p>
                            <p className='mt-1 text-muted-foreground'>
                              {sellerOperationExecution.execution.status ??
                                sellerOperationExecution.execution.blocker ??
                                'No status'}
                            </p>
                            {sellerOperationExecution.execution.result ? (
                              <div
                                className='mt-3 space-y-2'
                                data-testid='ebay-seller-operation-read-result'
                              >
                                <p className='font-medium'>
                                  Read result:{' '}
                                  {sellerOperationExecution.execution.result
                                    .source ?? 'local_read_model'}
                                </p>
                                <div
                                  className='grid gap-2 sm:grid-cols-2'
                                  data-testid='ebay-seller-operation-read-records'
                                >
                                  {(
                                    sellerOperationExecution.execution.result
                                      .records ?? []
                                  ).map((record) => (
                                    <div
                                      key={record.id ?? record.title}
                                      className='rounded border bg-background/60 p-2'
                                    >
                                      <p className='font-medium'>
                                        {record.title ?? record.id}
                                      </p>
                                      <p className='text-muted-foreground'>
                                        {record.kind ?? 'seller_operation'} /{' '}
                                        {record.status ?? 'unknown'}
                                      </p>
                                    </div>
                                  ))}
                                </div>
                              </div>
                            ) : null}
                          </div>
                        ) : null}
                      </section>
                      <section
                        className='rounded-md border p-3'
                        data-testid='ebay-listing-lifecycle-panel'
                      >
                        <div className='flex flex-wrap items-start justify-between gap-3'>
                          <div>
                            <p className='font-medium'>Listing Lifecycle</p>
                            <p className='text-xs text-muted-foreground'>
                              Create Cabinet-local drafts and preview publish,
                              revise, end, or relist commands before any eBay
                              write is confirmed.
                            </p>
                            <p className='text-xs text-muted-foreground'>
                              Confirmed remote writes remain blocked with
                              ebay_listing_lifecycle_adapter_required until the
                              eBay lifecycle adapter is configured.
                            </p>
                          </div>
                          <span
                            className='rounded bg-muted px-2 py-1 text-xs text-muted-foreground'
                            data-testid='ebay-listing-lifecycle-safe-mode'
                          >
                            Publish, revise, end, and relist require
                            confirmation
                          </span>
                        </div>
                        <div className='mt-3 grid gap-2 md:grid-cols-2'>
                          <div className='space-y-2'>
                            <Label htmlFor='ebay-listing-lifecycle-item-id'>
                              Item ID
                            </Label>
                            <Input
                              id='ebay-listing-lifecycle-item-id'
                              data-testid='ebay-listing-lifecycle-item-id'
                              value={form.listingLifecycleItemId}
                              onChange={(e) =>
                                setForm((prev) => ({
                                  ...prev,
                                  listingLifecycleItemId: e.target.value,
                                }))
                              }
                            />
                          </div>
                          <div className='space-y-2'>
                            <Label htmlFor='ebay-listing-lifecycle-title'>
                              Draft title
                            </Label>
                            <Input
                              id='ebay-listing-lifecycle-title'
                              data-testid='ebay-listing-lifecycle-title'
                              value={form.listingLifecycleTitle}
                              onChange={(e) =>
                                setForm((prev) => ({
                                  ...prev,
                                  listingLifecycleTitle: e.target.value,
                                }))
                              }
                            />
                          </div>
                          <div className='space-y-2'>
                            <Label htmlFor='ebay-listing-lifecycle-draft-id'>
                              Draft ID
                            </Label>
                            <Input
                              id='ebay-listing-lifecycle-draft-id'
                              data-testid='ebay-listing-lifecycle-draft-id'
                              value={form.listingLifecycleDraftId}
                              onChange={(e) =>
                                setForm((prev) => ({
                                  ...prev,
                                  listingLifecycleDraftId: e.target.value,
                                }))
                              }
                            />
                          </div>
                          <div className='space-y-2'>
                            <Label htmlFor='ebay-listing-lifecycle-listing-id'>
                              Listing ID
                            </Label>
                            <Input
                              id='ebay-listing-lifecycle-listing-id'
                              data-testid='ebay-listing-lifecycle-listing-id'
                              value={form.listingLifecycleListingId}
                              onChange={(e) =>
                                setForm((prev) => ({
                                  ...prev,
                                  listingLifecycleListingId: e.target.value,
                                }))
                              }
                            />
                          </div>
                        </div>
                        <div className='mt-3 grid gap-2 sm:grid-cols-2 lg:grid-cols-5'>
                          {listingLifecycleCommands.map((command) => (
                            <div
                              key={command}
                              className='rounded-md bg-muted/40 p-3 text-xs'
                              data-testid={`ebay-listing-lifecycle-${command}`}
                            >
                              <p className='font-medium'>
                                {sellerListingLifecycleLabel(command)}
                              </p>
                              <p className='mt-1 text-muted-foreground'>
                                {command === 'draft'
                                  ? 'Local draft only'
                                  : 'Confirmed eBay write gate'}
                              </p>
                              <div className='mt-3 flex flex-wrap gap-2'>
                                <Button
                                  type='button'
                                  size='sm'
                                  variant='outline'
                                  data-testid={`ebay-listing-lifecycle-preview-${command}`}
                                  disabled={listingLifecycleWorking !== null}
                                  onClick={() =>
                                    void previewListingLifecycle(command, false)
                                  }
                                >
                                  Preview
                                </Button>
                                <Button
                                  type='button'
                                  size='sm'
                                  variant='outline'
                                  data-testid={`ebay-listing-lifecycle-confirm-preview-${command}`}
                                  disabled={
                                    listingLifecycleWorking !== null ||
                                    command === 'draft'
                                  }
                                  onClick={() =>
                                    void previewListingLifecycle(command, true)
                                  }
                                >
                                  Confirm Preview
                                </Button>
                                <Button
                                  type='button'
                                  size='sm'
                                  variant='outline'
                                  data-testid={`ebay-listing-lifecycle-execute-${command}`}
                                  disabled={
                                    listingLifecycleWorking !== null ||
                                    command !== 'draft'
                                  }
                                  onClick={() =>
                                    void executeListingLifecycle(command, false)
                                  }
                                >
                                  Execute
                                </Button>
                                <Button
                                  type='button'
                                  size='sm'
                                  variant='outline'
                                  data-testid={`ebay-listing-lifecycle-confirm-execute-${command}`}
                                  disabled={
                                    listingLifecycleWorking !== null ||
                                    command === 'draft'
                                  }
                                  onClick={() =>
                                    void executeListingLifecycle(command, true)
                                  }
                                >
                                  Confirm Execute
                                </Button>
                              </div>
                            </div>
                          ))}
                        </div>
                        {listingLifecycleError ? (
                          <p
                            className='mt-3 text-xs text-destructive'
                            data-testid='ebay-listing-lifecycle-error'
                          >
                            {listingLifecycleError}
                          </p>
                        ) : null}
                        {listingLifecycleResult?.preview ? (
                          <div
                            className='mt-3 rounded-md bg-muted/40 p-3 text-xs'
                            data-testid='ebay-listing-lifecycle-preview-result'
                          >
                            <p className='font-medium'>
                              Preview:{' '}
                              {sellerListingLifecycleLabel(
                                listingLifecycleResult.preview.command ?? ''
                              )}
                            </p>
                            <p className='mt-1 text-muted-foreground'>
                              Allowed:{' '}
                              {listingLifecycleResult.preview.allowed
                                ? 'yes'
                                : 'no'}{' '}
                              / Local only:{' '}
                              {listingLifecycleResult.preview.local_only
                                ? 'yes'
                                : 'no'}{' '}
                              / Remote write:{' '}
                              {listingLifecycleResult.preview.remote_write
                                ? 'yes'
                                : 'no'}
                            </p>
                            <p className='mt-1 text-muted-foreground'>
                              {listingLifecycleResult.preview.blocker ??
                                'No blocker'}
                            </p>
                          </div>
                        ) : null}
                        {listingLifecycleExecution?.execution ? (
                          <div
                            className='mt-3 rounded-md bg-muted/40 p-3 text-xs'
                            data-testid='ebay-listing-lifecycle-execute-result'
                          >
                            <p className='font-medium'>
                              Execute:{' '}
                              {sellerListingLifecycleLabel(
                                listingLifecycleExecution.execution.command ??
                                  ''
                              )}
                            </p>
                            <p className='mt-1 text-muted-foreground'>
                              Executed:{' '}
                              {listingLifecycleExecution.execution.executed
                                ? 'yes'
                                : 'no'}{' '}
                              / Local only:{' '}
                              {listingLifecycleExecution.execution.local_only
                                ? 'yes'
                                : 'no'}{' '}
                              / Remote write:{' '}
                              {listingLifecycleExecution.execution.remote_write
                                ? 'yes'
                                : 'no'}
                            </p>
                            <p className='mt-1 text-muted-foreground'>
                              {listingLifecycleExecution.execution.response
                                ?.draft_id ??
                                listingLifecycleExecution.execution.status ??
                                listingLifecycleExecution.execution.blocker ??
                                'No status'}
                            </p>
                          </div>
                        ) : null}
                      </section>
                      <details
                        className='rounded-md border p-3'
                        data-testid='ebay-landed-cost-planner-panel'
                      >
                        <summary className='cursor-pointer text-sm font-medium'>
                          Landed Cost Planner
                        </summary>
                        <div className='mt-3 space-y-3'>
                          <div className='flex flex-wrap items-start justify-between gap-3'>
                            <div>
                              <p className='font-medium'>Landed Cost Planner</p>
                              <p className='text-xs text-muted-foreground'>
                                Preview allocation, provenance, and
                                consolidation thresholds from the commerce
                                planning API.
                              </p>
                            </div>
                            <span
                              className='rounded bg-muted px-2 py-1 text-xs text-muted-foreground'
                              data-testid='ebay-landed-cost-mutation-status'
                            >
                              Preview only / no mutation
                            </span>
                          </div>
                          <Label
                            className='block'
                            htmlFor='ebay-landed-cost-payload'
                          >
                            Planning payload
                          </Label>
                          <textarea
                            id='ebay-landed-cost-payload'
                            className='min-h-40 w-full rounded-md border bg-background p-2 font-mono text-xs'
                            data-testid='ebay-landed-cost-payload'
                            value={form.landedCostPayload}
                            onChange={(e) =>
                              setForm((prev) => ({
                                ...prev,
                                landedCostPayload: e.target.value,
                              }))
                            }
                          />
                          <Button
                            type='button'
                            size='sm'
                            variant='outline'
                            data-testid='ebay-landed-cost-preview'
                            disabled={landedCostWorking}
                            onClick={() => void previewLandedCostPlan()}
                          >
                            {landedCostWorking
                              ? 'Previewing...'
                              : 'Preview Plan'}
                          </Button>
                          {landedCostError ? (
                            <p
                              className='text-xs text-destructive'
                              data-testid='ebay-landed-cost-error'
                            >
                              {landedCostError}
                            </p>
                          ) : null}
                          {landedCostResult?.allocation &&
                          landedCostResult.consolidation ? (
                            <div
                              className='rounded-md bg-muted/40 p-3 text-xs'
                              data-testid='ebay-landed-cost-result'
                            >
                              <p className='font-medium'>
                                Mode: {landedCostResult.mode} / Mutable:{' '}
                                {landedCostResult.mutable ? 'yes' : 'no'}
                              </p>
                              <p className='mt-1 text-muted-foreground'>
                                Direct:{' '}
                                {formatCents(
                                  landedCostResult.allocation.total_direct_cents
                                )}{' '}
                                / Shared:{' '}
                                {formatCents(
                                  landedCostResult.allocation.total_shared_cents
                                )}{' '}
                                / Landed:{' '}
                                {formatCents(
                                  landedCostResult.allocation.total_landed_cents
                                )}
                              </p>
                              <div className='mt-3 grid gap-2 sm:grid-cols-2'>
                                {(landedCostResult.allocation.items ?? []).map(
                                  (item) => (
                                    <div
                                      key={item.item_id}
                                      className='rounded bg-background/70 p-2'
                                    >
                                      <p className='font-medium'>
                                        {item.item_id}: landed{' '}
                                        {formatCents(item.landed_cost_cents)}
                                      </p>
                                      <p className='mt-1 text-muted-foreground'>
                                        Allocated:{' '}
                                        {formatCents(item.allocated_cost_cents)}{' '}
                                        / Direct:{' '}
                                        {formatCents(item.direct_cost_cents)}
                                      </p>
                                      <p className='mt-1 text-muted-foreground'>
                                        Provenance:{' '}
                                        {(
                                          item.allocation_provenance_id ?? []
                                        ).join(', ') || 'none'}
                                      </p>
                                    </div>
                                  )
                                )}
                              </div>
                              <p className='mt-3 text-muted-foreground'>
                                Consolidation:{' '}
                                {landedCostResult.consolidation
                                  .threshold_state ?? 'unknown'}{' '}
                                / Total:{' '}
                                {formatCents(
                                  landedCostResult.consolidation
                                    .estimated_total_cents
                                )}{' '}
                                / Items:{' '}
                                {(
                                  landedCostResult.consolidation.item_ids ?? []
                                ).join(', ')}
                              </p>
                              {(landedCostResult.consolidation.warnings ?? [])
                                .length > 0 ? (
                                <p className='mt-1 text-muted-foreground'>
                                  Warnings:{' '}
                                  {landedCostResult.consolidation.warnings?.join(
                                    ', '
                                  )}
                                </p>
                              ) : null}
                            </div>
                          ) : null}
                        </div>
                      </details>
                      <details
                        className='rounded-md border p-3'
                        data-testid='ebay-buyer-interest-sync-panel'
                      >
                        <summary className='cursor-pointer text-sm font-medium'>
                          Buyer Interest Sync
                        </summary>
                        <div className='mt-3 space-y-3'>
                          <div className='flex flex-wrap items-start justify-between gap-3'>
                            <div>
                              <p className='font-medium'>Buyer Interest Sync</p>
                              <p className='text-xs text-muted-foreground'>
                                Preview or import watched, saved, liked, and
                                cart-like listings into local Wishlist and
                                Discoveries.
                              </p>
                            </div>
                            <span
                              className='rounded bg-muted px-2 py-1 text-xs text-muted-foreground'
                              data-testid='ebay-buyer-interest-writeback-status'
                            >
                              Write-back blocked until eBay capability is
                              verified
                            </span>
                          </div>
                          <Label
                            className='mt-3 block'
                            htmlFor='ebay-buyer-interest-payload'
                          >
                            Sync payload
                          </Label>
                          <textarea
                            id='ebay-buyer-interest-payload'
                            className='mt-2 min-h-32 w-full rounded-md border bg-background p-2 font-mono text-xs'
                            data-testid='ebay-buyer-interest-payload'
                            value={form.buyerInterestPayload}
                            onChange={(e) =>
                              setForm((prev) => ({
                                ...prev,
                                buyerInterestPayload: e.target.value,
                              }))
                            }
                          />
                          <div className='mt-3 flex flex-wrap gap-2'>
                            <Button
                              type='button'
                              size='sm'
                              variant='outline'
                              data-testid='ebay-buyer-interest-preview'
                              disabled={buyerInterestWorking}
                              onClick={() =>
                                void runBuyerInterestSync('preview')
                              }
                            >
                              <SearchCheck className='mr-2 size-4' />
                              Preview
                            </Button>
                            <Button
                              type='button'
                              size='sm'
                              data-testid='ebay-buyer-interest-import'
                              disabled={buyerInterestWorking}
                              onClick={() =>
                                void runBuyerInterestSync('import')
                              }
                            >
                              <Download className='mr-2 size-4' />
                              Import
                            </Button>
                          </div>
                          {buyerInterestError ? (
                            <p className='mt-2 text-xs text-destructive'>
                              {buyerInterestError}
                            </p>
                          ) : null}
                          {buyerInterestResult ? (
                            <div
                              className='mt-3 rounded-md bg-muted/40 p-3 text-xs'
                              data-testid='ebay-buyer-interest-result'
                            >
                              <p>
                                Mode: {buyerInterestResult.mode ?? 'unknown'} /
                                Total: {buyerInterestResult.total ?? 0}
                              </p>
                              <p>
                                Wishlist:{' '}
                                {buyerInterestResult.counts?.wishlist ?? 0} /
                                Discoveries:{' '}
                                {buyerInterestResult.counts?.discovery ?? 0}
                              </p>
                              <ul className='mt-2 space-y-1'>
                                {(
                                  buyerInterestResult.results ??
                                  buyerInterestResult.mappings ??
                                  []
                                ).map((entry, index) => (
                                  <li
                                    key={`${entry.listing_id ?? 'entry'}-${index}`}
                                  >
                                    {entry.title ??
                                      entry.listing_id ??
                                      'Untitled'}
                                    {' -> '}
                                    {entry.destination ?? 'unknown'}; write-back{' '}
                                    {entry.write_back_allowed
                                      ? 'allowed'
                                      : (entry.write_back_blocker ?? 'blocked')}
                                  </li>
                                ))}
                              </ul>
                            </div>
                          ) : null}
                        </div>
                      </details>
                    </div>
                  ) : null}
                </>
              ) : null}

              {saveError ? (
                <p className='text-sm text-destructive'>{saveError}</p>
              ) : null}

              <div className='flex flex-wrap justify-end gap-2'>
                {editingProvider.provider_id !== 'openai' ? (
                  <>
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
                  </>
                ) : null}
                <Button
                  variant='outline'
                  onClick={closeIntegration}
                  disabled={saving}
                >
                  Cancel
                </Button>
                <Button
                  onClick={() =>
                    editingProvider.provider_id === 'telegram'
                      ? closeIntegration()
                      : void saveIntegration()
                  }
                  disabled={saving}
                >
                  {saving
                    ? 'Saving...'
                    : editingProvider.provider_id === 'openai'
                      ? 'Save OpenAI'
                      : editingProvider.provider_id === 'telegram'
                        ? 'Done'
                        : 'Save Integration'}
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
              <p data-testid='integrations-row-details-provider-contract'>
                <strong>Registry contract:</strong>{' '}
                {rowDetailsProvider.provider_category ?? 'uncategorized'}
                {' / '}
                {rowDetailsProvider.provider_type ?? 'unknown'}
                {' / '}
                {rowDetailsProvider.config_schema_ref ?? 'provider-specific'}
              </p>
              <p data-testid='integrations-row-details-provider-actions'>
                <strong>Actions:</strong>{' '}
                {providerManifestActions(rowDetailsProvider).join(', ') ||
                  'none'}
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
