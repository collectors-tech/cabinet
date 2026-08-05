(() => {
  if (globalThis.__cabinetProviderBridgeV1) return
  globalThis.__cabinetProviderBridgeV1 = true
  const maxSelectors = 60
  const maxSelectorLength = 256
  let active = true

  globalThis.chrome.runtime.onMessage.addListener((message, _sender, sendResponse) => {
    if (message?.type === 'cabinet:deactivate-readiness') {
      active = false
      sendResponse({ deactivated: true })
      return true
    }
    if (message?.type !== 'cabinet:probe-readiness') return false
    if (!active) {
      sendResponse({ matched: [] })
      return true
    }
    const selectors = Array.isArray(message.selectors) ? message.selectors.slice(0, maxSelectors) : []
    const matched = []
    for (const selector of selectors) {
      if (typeof selector !== 'string' || selector.length === 0 || selector.length > maxSelectorLength) continue
      try {
        if (document.querySelector(selector)) matched.push(selector)
      } catch {
        // Invalid module selectors fail closed and never expose page contents.
      }
    }
    sendResponse({ matched })
    return true
  })
})()
