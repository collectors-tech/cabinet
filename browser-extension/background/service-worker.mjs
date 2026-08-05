import { CompanionClient, CompanionProtocolError, companionStorageKeys } from '../src/companion-client.mjs'
import { BrowserStorage, browserPlatform } from '../platform/browser-api.mjs'
import { normaliseCabinetURL } from '../runtime/config.mjs'
import { normaliseCapture } from '../runtime/capture-contract.mjs'
import { DurableQueue } from '../runtime/durable-queue.mjs'
import { classifyReadiness } from '../runtime/readiness.mjs'
import { moduleForURL, normaliseRegistry, permissionOrigins } from '../runtime/module-contract.mjs'
import { RetryGovernor } from '../runtime/retry-governor.mjs'

const stateKey = 'cabinet.companion.host-state.v1'
const deviceKey = 'cabinet.companion.device-id.v1'
const governorKey = 'cabinet.companion.governor.v1'
const defaultCabinetURL = 'http://127.0.0.1:17880/'
const storage = new BrowserStorage()
const queue = new DurableQueue({ storage })
const governor = new RetryGovernor({ snapshot: await storage.get(governorKey) })

await storage.restrictToTrustedContexts()

const savedState = (await storage.get(stateKey)) ?? {}
let savedCabinetURL = defaultCabinetURL
try { savedCabinetURL = normaliseCabinetURL(savedState.cabinet_url ?? defaultCabinetURL) } catch { /* reset unsafe legacy state */ }
const state = {
  cabinet_url: defaultCabinetURL,
  connection: 'disconnected',
  profile_id: '',
  modules: [],
  pending: 0,
	cabinet_pending: 0,
	cabinet_failed: 0,
	cabinet_review: 0,
  last_sync: '',
  error: '',
  sync_paused: false,
  extension_version: browserPlatform.runtime.manifest().version,
  ...savedState,
  cabinet_url: savedCabinetURL,
  sync_paused: savedState.sync_paused === true,
  modules: Array.isArray(savedState.modules) ? savedState.modules : [],
  extension_version: browserPlatform.runtime.manifest().version,
}

const saveState = async () => {
  const pendingJobs = await queue.pending()
  state.pending = pendingJobs.length
  for (const module of state.modules) {
    const moduleJobs = pendingJobs.filter((job) => job.integration_instance_id === module.integration_instance_id)
    module.pending = moduleJobs.length
    module.error = moduleJobs.find((job) => job.last_error)?.last_error ?? module.error ?? ''
  }
  await storage.set(stateKey, state)
  await browserPlatform.badge.text(state.pending ? String(Math.min(state.pending, 99)) : '')
  await browserPlatform.badge.colour(state.error ? '#b42318' : '#3858e9')
}

const deviceID = async () => {
  let id = await storage.get(deviceKey)
  if (!id) {
    id = crypto.randomUUID()
    await storage.set(deviceKey, id)
  }
  return id
}

const client = async () => new CompanionClient({
  baseURL: state.cabinet_url,
  deviceID: await deviceID(),
  storage,
})

const refreshModules = async () => {
	const companionClient = await client()
	const registry = normaliseRegistry(await companionClient.modules())
	let inbox = { captures: [], pending: 0, failed: 0, review: 0 }
	try { inbox = await companionClient.captureInbox() } catch (error) {
	  if (!(error instanceof CompanionProtocolError) || error.status !== 403) throw error
	}
	const captures = Array.isArray(inbox?.captures) ? inbox.captures : []
  const modules = []
  for (const module of registry.modules) {
    const permission = await browserPlatform.permissions.contains(permissionOrigins(module))
    const previous = state.modules.find((item) => item.integration_instance_id === module.integration_instance_id)
    const permissionEnabled = permission && previous?.permission_enabled !== false
    modules.push({
      ...module,
      status: module.configuration.setup_required ? 'setup_needed' : permissionEnabled ? (previous?.status ?? 'browser_required') : 'permission_required',
      permission_enabled: permissionEnabled,
      last_sync: previous?.last_sync ?? '',
      pending: previous?.pending ?? 0,
      error: previous?.error ?? '',
      paused: previous?.paused === true,
      guidance: previous?.guidance ?? '',
      last_checked: previous?.last_checked ?? '',
      last_attempt_at: previous?.last_attempt_at ?? 0,
	  cabinet_pending: captures.filter((capture) => capture.module_id === module.id && ['accepted', 'validating', 'queued', 'processing', 'retryable-failed'].includes(capture.state)).length,
	  cabinet_failed: captures.filter((capture) => capture.module_id === module.id && ['retryable-failed', 'permanently-failed'].includes(capture.state)).length,
	  cabinet_review: captures.filter((capture) => capture.module_id === module.id && capture.state === 'review').length,
    })
  }
  state.connection = 'connected'
  state.profile_id = registry.profile_id
  state.modules = modules
	state.cabinet_pending = Number(inbox?.pending ?? 0)
	state.cabinet_failed = Number(inbox?.failed ?? 0)
	state.cabinet_review = Number(inbox?.review ?? 0)
  state.error = ''
  await saveState()
  return state
}

const reconnect = async () => {
  try {
    await (await client()).reconnect()
    return await refreshModules()
  } catch (error) {
    state.connection = error instanceof CompanionProtocolError && error.status === 401 ? 'pairing_required' : 'disconnected'
    state.error = error.code ?? 'cabinet_unavailable'
    for (const module of state.modules) module.status = 'disconnected'
    await saveState()
    return state
  }
}

const openOrFocus = async (url, patterns = [`${new URL(url).origin}/*`]) => {
  const matches = await browserPlatform.tabs.query({ url: patterns })
  if (matches[0]?.id !== undefined) {
    await browserPlatform.tabs.update(matches[0].id, { active: true })
    return matches[0]
  }
  return browserPlatform.tabs.create(url)
}

const selectorsFor = (module) => [
  ...module.browser.readiness.challenge,
  ...module.browser.readiness.logged_out,
  ...module.browser.readiness.ready,
]

const checkModule = async (integrationInstanceID) => {
  const module = state.modules.find((item) => item.integration_instance_id === integrationInstanceID)
  if (!module) throw new Error('module_not_found')
  if (module.configuration.setup_required) {
    module.status = 'setup_needed'
    await saveState()
    return module
  }
  if (!module.permission_enabled || !await browserPlatform.permissions.contains(permissionOrigins(module))) {
    module.status = 'permission_required'
    await saveState()
    return module
  }
  const tabs = await browserPlatform.tabs.query({ url: module.browser.url_patterns })
  const tab = tabs.find((candidate) => candidate.id !== undefined)
  if (!tab) {
    module.status = 'browser_required'
    await saveState()
    return module
  }
  await browserPlatform.scripting.execute({ target: { tabId: tab.id }, files: ['content/provider-bridge.js'] })
  const response = await browserPlatform.tabs.sendMessage(tab.id, {
    type: 'cabinet:probe-readiness',
    selectors: selectorsFor(module),
  })
  const readiness = classifyReadiness(module.browser.readiness, response?.matched)
  module.status = readiness.state
  module.guidance = readiness.guidance
  module.last_checked = new Date().toISOString()
  await saveState()
  return module
}

const processQueue = async () => {
  if (state.sync_paused) return
  const job = await queue.claim()
  if (!job) return
  const module = state.modules.find((item) => item.integration_instance_id === job.integration_instance_id)
  if (module?.paused) {
    await queue.fail(job.id, 'module_sync_paused', 60_000)
    return saveState()
  }
  if (!governor.canRun(job.module_id)) {
    await queue.fail(job.id, 'provider_circuit_open', 60_000)
    return saveState()
  }
  try {
    if (job.kind !== 'capture') throw new Error('queue_kind_unsupported')
	const accepted = await (await client()).submitCapture(job.payload)
	if (accepted?.committed !== true || !['completed', 'partial', 'review'].includes(accepted.state)) {
	  throw new Error('capture_not_committed')
	}
    governor.success(job.module_id)
    await storage.set(governorKey, governor.snapshot())
    await queue.complete(job.id)
    state.last_sync = new Date().toISOString()
    if (module) {
      module.last_sync = state.last_sync
      module.status = job.partial ? 'partial' : 'ready'
      module.guidance = job.partial ? 'Cabinet accepted a partial observation; review missing fields.' : 'Cabinet accepted the observation.'
      module.error = ''
    }
    state.error = ''
  } catch (error) {
    const retry = governor.failure(job.module_id)
    await storage.set(governorKey, governor.snapshot())
    await queue.fail(job.id, error.code ?? 'sync_failed', retry.delay_ms)
    state.error = error.code ?? 'sync_failed'
    if (module) module.error = state.error
  }
  await saveState()
}

browserPlatform.runtime.onMessage((message, _sender, sendResponse) => {
  const run = async () => {
    switch (message?.type) {
      case 'host:get-state': return state
      case 'host:reconnect': return reconnect()
      case 'host:update-config': {
        const cabinetURL = normaliseCabinetURL(message.cabinet_url)
        try { await (await client()).revoke() } catch { await storage.delete(companionStorageKeys.credential) }
        await storage.delete(companionStorageKeys.pendingPairing)
        state.cabinet_url = cabinetURL
        state.connection = 'disconnected'
        state.profile_id = ''
        state.modules = []
        state.error = ''
        delete state.pairing_code
        await saveState()
        return state
      }
      case 'host:set-sync-paused': {
        state.sync_paused = message.paused === true
        await saveState()
        if (!state.sync_paused) await processQueue()
        return state
      }
      case 'host:start-pairing': {
        const receipt = await (await client()).startPairing(message.device_name ?? 'Chrome or Edge', [
          'modules:read', 'captures:submit', 'media:submit', 'session:manage',
        ])
        state.connection = 'approval_required'
        state.pairing_code = receipt.pairing_code
        await saveState()
        return state
      }
      case 'host:exchange-pairing': {
        await (await client()).exchangePairing()
        delete state.pairing_code
        return refreshModules()
      }
      case 'host:open-cabinet': return openOrFocus(new URL('/integrations/', state.cabinet_url).href)
      case 'host:open-module': {
        const module = state.modules.find((item) => item.integration_instance_id === message.integration_instance_id)
        if (!module) throw new Error('module_not_found')
        return openOrFocus(module.browser.start_url, module.browser.url_patterns)
      }
      case 'host:setup-module': {
        const module = state.modules.find((item) => item.integration_instance_id === message.integration_instance_id)
        if (!module) throw new Error('module_not_found')
        return openOrFocus(new URL(`/integrations/?instance=${encodeURIComponent(module.integration_instance_id)}`, state.cabinet_url).href)
      }
      case 'host:review-module': {
        const module = state.modules.find((item) => item.integration_instance_id === message.integration_instance_id)
        if (!module) throw new Error('module_not_found')
        return openOrFocus(new URL(`/${module.configuration.review_destination}`, state.cabinet_url).href)
      }
      case 'host:help-module': {
        const module = state.modules.find((item) => item.integration_instance_id === message.integration_instance_id)
        if (!module) throw new Error('module_not_found')
        return openOrFocus(new URL(module.configuration.help_url, state.cabinet_url).href)
      }
      case 'host:grant-permission': {
        const module = state.modules.find((item) => item.integration_instance_id === message.integration_instance_id)
        if (!module) throw new Error('module_not_found')
        const granted = await browserPlatform.permissions.request(permissionOrigins(module))
        module.permission_enabled = granted
        module.status = module.configuration.setup_required ? 'setup_needed' : granted ? 'browser_required' : 'permission_required'
        await saveState()
        return module
      }
      case 'host:remove-permission': {
        const module = state.modules.find((item) => item.integration_instance_id === message.integration_instance_id)
        if (!module) throw new Error('module_not_found')
        const tabs = await browserPlatform.tabs.query({ url: permissionOrigins(module) })
        for (const tab of tabs) {
          if (tab.id === undefined) continue
          try { await browserPlatform.tabs.sendMessage(tab.id, { type: 'cabinet:deactivate-readiness' }) } catch { /* bridge was not loaded */ }
        }
        module.permission_enabled = false
        const removableOrigins = permissionOrigins(module).filter((origin) => !state.modules.some((other) =>
          other.integration_instance_id !== module.integration_instance_id && other.permission_enabled && permissionOrigins(other).includes(origin)
        ))
        if (removableOrigins.length > 0) await browserPlatform.permissions.remove(removableOrigins)
        module.status = 'permission_required'
        await saveState()
        return module
      }
      case 'host:check-module': return checkModule(message.integration_instance_id)
      case 'host:set-module-paused': {
        const module = state.modules.find((item) => item.integration_instance_id === message.integration_instance_id)
        if (!module) throw new Error('module_not_found')
        module.paused = message.paused === true
        await saveState()
        if (!module.paused) await processQueue()
        return module
      }
      case 'host:sync-module': {
        const module = state.modules.find((item) => item.integration_instance_id === message.integration_instance_id)
        if (!module || !module.configuration.sync_available || !module.browser.capture_script || module.paused) {
          throw new Error('module_sync_unavailable')
        }
        const now = Date.now()
        const minimumInterval = Math.ceil(60_000 / module.configuration.rate_limit_per_minute)
        if (module.last_attempt_at && now - module.last_attempt_at < minimumInterval) throw new Error('module_rate_limited')
        module.last_attempt_at = now
        const readiness = await checkModule(module.integration_instance_id)
        if (readiness.status !== 'ready') throw new Error('module_session_not_ready')
        const tabs = await browserPlatform.tabs.query({ url: module.browser.url_patterns })
        const tab = tabs.find((candidate) => candidate.id !== undefined && candidate.url)
        if (!tab) throw new Error('module_browser_required')
        await browserPlatform.scripting.execute({ target: { tabId: tab.id }, files: [module.browser.capture_script] })
        const observation = await browserPlatform.tabs.sendMessage(tab.id, { type: 'cabinet:capture', module_id: module.id })
		const jobID = crypto.randomUUID()
		const payload = await normaliseCapture(module, observation, tab.url, state.profile_id, jobID)
		const partial = payload.page_complete !== true
        await queue.enqueue({
		  id: jobID,
          integration_instance_id: module.integration_instance_id,
          module_id: module.id,
          kind: 'capture',
          partial,
          payload,
        })
        module.status = 'syncing'
        await saveState()
        await processQueue()
        return state
      }
      case 'host:refresh-modules': return refreshModules()
      default: throw new Error('host_message_unsupported')
    }
  }
  run().then((result) => sendResponse({ ok: true, result })).catch((error) => {
    sendResponse({ ok: false, error: String(error.code ?? error.message ?? error).slice(0, 160) })
  })
  return true
})

browserPlatform.runtime.onStartup(reconnect)
browserPlatform.runtime.onInstalled(() => reconnect())
browserPlatform.alarms.onAlarm((alarm) => {
  if (alarm.name === 'cabinet-companion-queue') processQueue()
})
await browserPlatform.alarms.create('cabinet-companion-queue', { periodInMinutes: 1 })
await reconnect()

export { moduleForURL }
