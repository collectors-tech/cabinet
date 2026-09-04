const browserAPI = globalThis.browser ?? globalThis.chrome

export const assertBrowserAPI = () => {
  if (!browserAPI?.runtime || !browserAPI?.storage?.local) {
    throw new Error('browser_api_unavailable')
  }
  return browserAPI
}

const invoke = (owner, method, ...args) => {
  const fn = owner?.[method]
  if (typeof fn !== 'function') throw new Error(`browser_api_missing:${method}`)
  try {
    const result = fn.call(owner, ...args)
    if (result && typeof result.then === 'function') return result
    return Promise.resolve(result)
  } catch (error) {
    return Promise.reject(error)
  }
}

export class BrowserStorage {
  constructor(area = assertBrowserAPI().storage.local) {
    this.area = area
  }

  async get(key) {
    const result = await invoke(this.area, 'get', key)
    return result?.[key]
  }

  async set(key, value) {
    await invoke(this.area, 'set', { [key]: value })
  }

  async delete(key) {
    await invoke(this.area, 'remove', key)
  }

  async restrictToTrustedContexts() {
    if (typeof this.area.setAccessLevel === 'function') {
      await invoke(this.area, 'setAccessLevel', { accessLevel: 'TRUSTED_CONTEXTS' })
    }
  }
}

export const browserPlatform = Object.freeze({
  alarms: {
    create: (name, options) => invoke(assertBrowserAPI().alarms, 'create', name, options),
    onAlarm: (listener) => assertBrowserAPI().alarms.onAlarm.addListener(listener),
  },
  badge: {
    text: (text) => invoke(assertBrowserAPI().action, 'setBadgeText', { text }),
    colour: (color) => invoke(assertBrowserAPI().action, 'setBadgeBackgroundColor', { color }),
  },
  permissions: {
    contains: (origins) => invoke(assertBrowserAPI().permissions, 'contains', { origins }),
    request: (origins) => invoke(assertBrowserAPI().permissions, 'request', { origins }),
    remove: (origins) => invoke(assertBrowserAPI().permissions, 'remove', { origins }),
  },
  runtime: {
    manifest: () => assertBrowserAPI().runtime.getManifest(),
    origin: () => assertBrowserAPI().runtime.getURL('').replace(/\/$/, ''),
    onInstalled: (listener) => assertBrowserAPI().runtime.onInstalled.addListener(listener),
    onMessage: (listener) => assertBrowserAPI().runtime.onMessage.addListener(listener),
    onStartup: (listener) => assertBrowserAPI().runtime.onStartup.addListener(listener),
    sendMessage: (message) => invoke(assertBrowserAPI().runtime, 'sendMessage', message),
  },
  scripting: {
    execute: (details) => invoke(assertBrowserAPI().scripting, 'executeScript', details),
  },
  tabs: {
    create: (url) => invoke(assertBrowserAPI().tabs, 'create', { url }),
    query: (query) => invoke(assertBrowserAPI().tabs, 'query', query),
    sendMessage: (tabId, message) => invoke(assertBrowserAPI().tabs, 'sendMessage', tabId, message),
    update: (tabId, update) => invoke(assertBrowserAPI().tabs, 'update', tabId, update),
  },
})
