const secretKey = /(secret|token|password|cookie|api[_-]?key|authorization|credential)/i
const moduleID = /^[a-z0-9][a-z0-9._-]{0,127}$/

const requiredString = (value, field) => {
  if (typeof value !== 'string' || !value.trim()) throw new Error(`module_${field}_required`)
  return value.trim()
}

const safeOrigin = (pattern) => {
  const value = requiredString(pattern, 'origin')
  if (!value.startsWith('https://') || !value.endsWith('/*')) {
    throw new Error('module_origin_must_use_exact_https_host')
  }
  const parsed = new URL(value.slice(0, -1))
  if (parsed.hostname.includes('*') || parsed.username || parsed.password || parsed.pathname !== '/' || parsed.search || parsed.hash) {
    throw new Error('module_origin_must_use_exact_https_host')
  }
  return value
}

const assertNoSecretKeys = (value, path = 'module') => {
  if (!value || typeof value !== 'object') return
  for (const [key, child] of Object.entries(value)) {
    if (secretKey.test(key)) throw new Error(`module_secret-like_key_rejected:${path}.${key}`)
    assertNoSecretKeys(child, `${path}.${key}`)
  }
}

const selectorList = (value) => {
  if (!Array.isArray(value)) return []
  return [...new Set(value.filter((item) => typeof item === 'string' && item.length > 0 && item.length <= 256))].slice(0, 20)
}

const nameList = (value, field, maximum = 64) => {
  if (!Array.isArray(value) || value.length === 0 || value.length > maximum) throw new Error(`module_${field}_required`)
  const result = value.map((item) => requiredString(item, field))
  if (result.some((item) => item.length > 128)) throw new Error(`module_${field}_invalid`)
  return Object.freeze([...new Set(result)])
}

const urlPattern = (value, origins) => {
  const pattern = requiredString(value, 'url_pattern')
  if (!pattern.endsWith('*') || (pattern.match(/\*/g) ?? []).length !== 1) throw new Error('module_url_pattern_invalid')
  const prefix = new URL(pattern.slice(0, -1))
  if (prefix.protocol !== 'https:' || prefix.username || prefix.password || !origins.some((origin) => prefix.origin === new URL(origin.slice(0, -2)).origin)) {
    throw new Error('module_url_pattern_outside_origins')
  }
  return pattern
}

const captureSchemas = (schemas) => {
  if (!Array.isArray(schemas) || schemas.length === 0 || schemas.length > 20) throw new Error('module_capture_schemas_required')
  return Object.freeze(schemas.map((schema) => Object.freeze({
    payload_type: requiredString(schema?.payload_type, 'payload_type'),
    fields: nameList(schema?.fields, 'capture_fields'),
    media_fields: Array.isArray(schema?.media_fields) && schema.media_fields.length > 0
      ? nameList(schema.media_fields, 'media_fields', 32)
      : Object.freeze([]),
  })))
}

const moduleConfiguration = (input, captureScript) => {
  if (!input || typeof input !== 'object') throw new Error('module_configuration_required')
  if (!['manual_user_present', 'browser_open_scheduled'].includes(input.capture_mode)) throw new Error('module_capture_mode_invalid')
  const helpURL = requiredString(input.help_url, 'help_url')
  if (helpURL.length < 2 || !helpURL.startsWith('/') || helpURL.startsWith('//')) throw new Error('module_help_url_invalid')
  const rateLimit = Number(input.rate_limit_per_minute)
  if (!Number.isInteger(rateLimit) || rateLimit < 1 || rateLimit > 60) throw new Error('module_rate_limit_invalid')
  if (input.sync_available === true && !captureScript) throw new Error('module_sync_script_required')
  const reviewDestination = requiredString(input.review_destination, 'review_destination')
  if (!/^[a-z0-9][a-z0-9-]*$/.test(reviewDestination)) throw new Error('module_review_destination_invalid')
  return Object.freeze({
    capture_mode: input.capture_mode,
    item_fields: nameList(input.item_fields, 'item_fields'),
    media_policy: requiredString(input.media_policy, 'media_policy'),
    review_destination: reviewDestination,
    rate_limit_per_minute: rateLimit,
    help_url: helpURL,
    setup_required: input.setup_required === true,
    sync_available: input.sync_available === true,
  })
}

const normaliseModule = (input) => {
  assertNoSecretKeys(input)
  const id = requiredString(input.id, 'id')
  if (!moduleID.test(id)) throw new Error('module_id_invalid')
  if (input.passive_only !== true) throw new Error('module_must_be_passive')
  const browser = input.browser
  if (!browser || typeof browser !== 'object') throw new Error('module_browser_contract_required')
  const origins = [...new Set((browser.origins ?? []).map(safeOrigin))]
  if (origins.length === 0 || origins.length > 10) throw new Error('module_origins_required')
  const startURL = new URL(requiredString(browser.start_url, 'start_url'))
  if (startURL.protocol !== 'https:' || !origins.some((origin) => startURL.href.startsWith(origin.slice(0, -1)))) {
    throw new Error('module_start_url_outside_origins')
  }
  const patterns = [...new Set((browser.url_patterns ?? []).map((pattern) => urlPattern(pattern, origins)))]
  if (patterns.length === 0 || patterns.length > 20) throw new Error('module_url_patterns_required')
  const captureScript = browser.capture_script === undefined || browser.capture_script === '' ? '' : requiredString(browser.capture_script, 'capture_script')
  if (captureScript && (!captureScript.startsWith('modules/') || !captureScript.endsWith('.js') || captureScript.includes('..'))) {
    throw new Error('module_capture_script_invalid')
  }
  return Object.freeze({
    id,
    module_version: requiredString(input.module_version, 'version'),
    site: requiredString(input.site, 'site'),
    provider_id: requiredString(input.provider_id, 'provider_id'),
    integration_instance_id: requiredString(input.integration_instance_id, 'integration_instance_id'),
    actions: Object.freeze([...(input.actions ?? [])].filter((item) => typeof item === 'string').slice(0, 20)),
    passive_only: true,
    capture_schemas: captureSchemas(input.capture_schemas),
    workflows: nameList(input.workflows, 'workflows', 20),
    redaction_rules: nameList(input.redaction_rules, 'redaction_rules', 20),
    fixture_version: requiredString(input.fixture_version, 'fixture_version'),
    display: Object.freeze({ name: requiredString(input.display?.name ?? input.site, 'display_name') }),
    browser: Object.freeze({
      start_url: startURL.href,
      origins: Object.freeze(origins),
      url_patterns: Object.freeze(patterns),
      capture_script: captureScript,
      readiness: Object.freeze({
        ready: Object.freeze(selectorList(browser.readiness?.ready)),
        logged_out: Object.freeze(selectorList(browser.readiness?.logged_out)),
        challenge: Object.freeze(selectorList(browser.readiness?.challenge)),
      }),
    }),
    configuration: moduleConfiguration(input.configuration, captureScript),
    safe_config: Object.freeze({ ...(input.safe_config ?? {}) }),
    status: 'permission_required',
  })
}

export const normaliseRegistry = (registry) => {
  if (registry?.protocol_version !== '1') throw new Error('registry_protocol_unsupported')
  return Object.freeze({
    protocol_version: '1',
    profile_id: typeof registry.profile_id === 'string' ? registry.profile_id : '',
    modules: Object.freeze((registry.modules ?? []).map(normaliseModule)),
  })
}

export const permissionOrigins = (module) => [...module.browser.origins]

export const moduleForURL = (modules, rawURL) => {
  let url
  try { url = new URL(rawURL) } catch { return undefined }
  return modules.find((module) => module.browser.url_patterns.some((pattern) => url.href.startsWith(pattern.slice(0, -1))))
}
