import { browserPlatform } from '../platform/browser-api.mjs'

const elements = {
  announcement: document.querySelector('#announcement'),
  cabinetDetail: document.querySelector('#cabinet-detail'),
  cabinetURL: document.querySelector('#cabinet-url'),
  connect: document.querySelector('#connect'),
  connection: document.querySelector('#connection'),
  empty: document.querySelector('#empty-state'),
  error: document.querySelector('#error'),
  list: document.querySelector('#module-list'),
  openCabinet: document.querySelector('#open-cabinet'),
  refresh: document.querySelector('#refresh'),
  settingsForm: document.querySelector('#settings-form'),
  syncToggle: document.querySelector('#sync-toggle'),
  template: document.querySelector('#module-row'),
}

const labels = {
  action_required: 'Action required',
  browser_required: 'Browser required',
  disconnected: 'Disconnected',
  logged_out: 'Login required',
  partial: 'Partial',
  permission_required: 'Site access required',
  ready: 'Ready to sync',
  setup_needed: 'Setup needed',
  syncing: 'Syncing',
  unsupported: 'Page not supported',
}

const send = async (type, detail = {}) => {
  const response = await browserPlatform.runtime.sendMessage({ type, ...detail })
  if (!response?.ok) throw new Error(response?.error ?? 'companion_request_failed')
  return response.result
}

const announce = (message) => { elements.announcement.textContent = message }
const showError = (error) => {
  elements.error.hidden = false
  elements.error.textContent = String(error?.message ?? error)
}
const clearError = () => { elements.error.hidden = true; elements.error.textContent = '' }

const act = async (button, action, detail) => {
  clearError()
  button.disabled = true
  try {
    const result = await send(action, detail)
    announce('Browser Companion updated.')
    if (action === 'host:open-cabinet' || action === 'host:open-module') return
    await render(result?.modules ? result : await send('host:get-state'))
  } catch (error) {
    showError(error)
  } finally {
    button.disabled = false
  }
}

const moduleRow = (module) => {
  const fragment = elements.template.content.cloneNode(true)
  const row = fragment.querySelector('li')
  row.dataset.integrationInstanceId = module.integration_instance_id
  row.querySelector('[data-module-icon]').textContent = module.display.name.trim().slice(0, 1).toLocaleUpperCase()
  row.querySelector('[data-module-name]').textContent = module.display.name
  row.querySelector('[data-module-status]').textContent = labels[module.status] ?? 'Not checked'
  row.querySelector('[data-module-detail]').textContent = module.guidance ?? 'Open the provider to check this browser session.'
  row.querySelector('[data-module-meta]').textContent = [
    'Enabled',
    `${module.pending ?? 0} pending`,
	(module.cabinet_pending ?? 0) ? `${module.cabinet_pending} Cabinet pending` : '',
	(module.cabinet_failed ?? 0) ? `${module.cabinet_failed} failed` : '',
	(module.cabinet_review ?? 0) ? `${module.cabinet_review} to review` : '',
    module.last_sync ? `last sync ${new Date(module.last_sync).toLocaleString()}` : 'not yet synced',
    module.error ? `error: ${module.error}` : '',
  ].filter(Boolean).join(' · ')
  const detail = { integration_instance_id: module.integration_instance_id }
  const open = row.querySelector('[data-action="open"]')
  open.textContent = module.status === 'logged_out' ? 'Sign in' : module.status === 'action_required' ? 'Continue in browser' : 'Open'
  open.addEventListener('click', (event) => act(event.currentTarget, 'host:open-module', detail))
  const check = row.querySelector('[data-action="check"]')
  check.hidden = module.permission_enabled !== true || module.configuration.setup_required
  check.textContent = ['logged_out', 'action_required', 'unsupported'].includes(module.status) ? 'Retry detection' : 'Check session'
  check.addEventListener('click', (event) => act(event.currentTarget, 'host:check-module', detail))
  const permission = row.querySelector('[data-action="permission"]')
  const hasPermission = module.permission_enabled === true
  permission.textContent = hasPermission ? 'Remove site access' : 'Allow site access'
  permission.addEventListener('click', (event) => act(
    event.currentTarget,
    hasPermission ? 'host:remove-permission' : 'host:grant-permission',
    detail
  ))
  const setup = row.querySelector('[data-action="setup"]')
  setup.hidden = !module.configuration.setup_required
  setup.addEventListener('click', (event) => act(event.currentTarget, 'host:setup-module', detail))
  const sync = row.querySelector('[data-action="sync"]')
  sync.hidden = !(module.configuration.sync_available && module.status === 'ready' && !module.paused)
  sync.addEventListener('click', (event) => act(event.currentTarget, 'host:sync-module', detail))
  const pause = row.querySelector('[data-action="pause"]')
  pause.hidden = !module.configuration.sync_available
  pause.textContent = module.paused ? 'Resume' : 'Pause'
  pause.addEventListener('click', (event) => act(event.currentTarget, 'host:set-module-paused', { ...detail, paused: !module.paused }))
  const review = row.querySelector('[data-action="review"]')
	review.hidden = !module.last_sync && !(module.cabinet_review > 0 || module.cabinet_failed > 0)
  review.addEventListener('click', (event) => act(event.currentTarget, 'host:review-module', detail))
  row.querySelector('[data-action="help"]').addEventListener('click', (event) => act(event.currentTarget, 'host:help-module', detail))
  return fragment
}

const render = async (state) => {
  elements.connection.textContent = state.connection.replaceAll('_', ' ')
  elements.cabinetDetail.textContent = state.pairing_code
    ? `Pairing code ${state.pairing_code}. Compare and approve it in Cabinet.`
    : state.profile_id
	? `Version ${state.extension_version}. Connected to profile ${state.profile_id}. ${state.pending ?? 0} extension jobs pending; ${state.cabinet_pending ?? 0} Cabinet jobs pending; ${state.cabinet_failed ?? 0} failed; ${state.cabinet_review ?? 0} to review.`
    : `Version ${state.extension_version}. Connect to load enabled integrations.`
  elements.cabinetURL.value = state.cabinet_url ?? ''
  elements.syncToggle.textContent = state.sync_paused ? 'Resume sync' : 'Pause sync'
  elements.connect.textContent = state.connection === 'approval_required' ? 'Finish pairing' : 'Connect'
  elements.list.replaceChildren()
  const modules = Array.isArray(state.modules) ? state.modules : []
  elements.empty.hidden = modules.length > 0
  for (const module of modules) elements.list.append(moduleRow(module))
}

elements.openCabinet.addEventListener('click', (event) => act(event.currentTarget, 'host:open-cabinet'))
elements.refresh.addEventListener('click', (event) => act(event.currentTarget, 'host:refresh-modules'))
elements.syncToggle.addEventListener('click', (event) => act(event.currentTarget, 'host:set-sync-paused', { paused: event.currentTarget.textContent === 'Pause sync' }))
elements.settingsForm.addEventListener('submit', (event) => {
  event.preventDefault()
  act(event.submitter, 'host:update-config', { cabinet_url: elements.cabinetURL.value })
})
elements.connect.addEventListener('click', async (event) => {
  const state = await send('host:get-state')
  const action = state.connection === 'approval_required' ? 'host:exchange-pairing' :
    state.connection === 'pairing_required' ? 'host:start-pairing' : 'host:reconnect'
  await act(event.currentTarget, action)
})

render(await send('host:get-state')).catch(showError)
