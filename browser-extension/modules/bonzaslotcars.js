(() => {
  if (globalThis.__cabinetBonzaCaptureV1) return
  globalThis.__cabinetBonzaCaptureV1 = true

  const moduleID = 'bonzaslotcars-search-capture'
  const productCardSelector = 'li.product, article.product, .products .product, .wc-block-grid__product, [data-product_id], [data-product-id], [itemtype="https://schema.org/Product"], [itemtype="http://schema.org/Product"]'
  const challengeSelectors = ['[data-cabinet-provider-state="challenge"]']
  const loggedOutSelectors = ['form.woocommerce-form-login input[type=password]', 'form[action*="my-account"] input[type=password]']
  const maximumItems = 200
  const maximumScripts = 80

  const firstNode = (root, selectors) => {
    for (const selector of selectors) {
      let node
      try { node = root.querySelector(selector) } catch { continue }
      if (node) return node
    }
    return null
  }

  const text = (node) => String(node?.textContent ?? '').replace(/\s+/g, ' ').trim()
  const attribute = (node, name) => String(node?.getAttribute?.(name) ?? '').trim()
  const bounded = (value, limit = 2048) => String(value ?? '').replace(/[\u0000-\u001f\u007f]/g, '').trim().slice(0, limit)

  const safeURL = (raw, base) => {
    let parsed
    try { parsed = new URL(String(raw ?? ''), base) } catch { return '' }
    const hosts = new Set(['bonzaslotcars.com.au', 'www.bonzaslotcars.com.au'])
    if (parsed.protocol !== 'https:' || parsed.username || parsed.password || !hosts.has(parsed.hostname)) return ''
    parsed.hostname = 'www.bonzaslotcars.com.au'
    parsed.search = ''
    parsed.hash = ''
    return parsed.href
  }

  const parsePrice = (value) => {
    const match = String(value ?? '').replace(/,/g, '').match(/(?:AUD\s*)?\$?\s*(\d+(?:\.\d{1,2})?)/i)
    if (!match) return undefined
    const result = Number(match[1])
    return Number.isFinite(result) && result >= 0 ? result : undefined
  }

  const productItem = (card, pageURL) => {
    if (typeof card?.getClientRects === 'function' && card.getClientRects().length === 0) return undefined
    const link = firstNode(card, [
      'a.woocommerce-LoopProduct-link', 'a.wc-block-grid__product-link', 'a[href*="/product/"]', 'a[href]',
    ])
    const titleNode = firstNode(card, [
      '.woocommerce-loop-product__title', '.wc-block-grid__product-title', '[itemprop="name"]', 'h2', 'h3',
    ])
    const priceNode = firstNode(card, [
      '[data-price-amount]', '.price ins .woocommerce-Price-amount', 'ins .woocommerce-Price-amount', '[itemprop="price"]',
      '.price .woocommerce-Price-amount', '.wc-block-grid__product-price', '.price',
    ])
    const imageNode = firstNode(card, [
      'img.attachment-woocommerce_thumbnail', '.wc-block-grid__product-image img', 'img[itemprop="image"]', 'img',
    ])
    const stockNode = firstNode(card, [
      '.stock.out-of-stock', '.out-of-stock', '.stock.low-stock', '.stock.available', '.stock.in-stock', '.stock',
    ])
    const itemURL = safeURL(link?.href || attribute(link, 'href'), pageURL)
    const title = bounded(text(titleNode) || text(link))
    const price = parsePrice(attribute(priceNode, 'data-price-amount') || attribute(priceNode, 'content') || text(priceNode))
    const imageURL = safeURL(imageNode?.currentSrc || imageNode?.src || attribute(imageNode, 'data-src') || attribute(imageNode, 'src'), pageURL)
    if (!itemURL || !title || price === undefined) return undefined
    const stockText = text(stockNode).toLowerCase()
    const stockState = /sold\s*out|out\s*of\s*stock|unavailable/.test(stockText)
      ? 'out_of_stock'
      : /low\s*stock|only\s+\d+/.test(stockText)
        ? 'low_stock'
        : /in\s*stock|available/.test(stockText)
          ? 'available'
          : 'unknown'
    const stockMatch = stockText.match(/(?:only\s*)?(\d+)\s*(?:in stock|available|left)/)
    const listingID = bounded(
      attribute(card, 'data-product_id') || attribute(card, 'data-product-id') || attribute(card, 'data-product-sku') ||
      attribute(firstNode(card, ['[data-product_id]', '[data-product-id]', '[data-product-sku]', '[itemprop="sku"]']), 'data-product_id') ||
      attribute(firstNode(card, ['[data-product-id]', '[data-product-sku]', '[itemprop="sku"]']), 'data-product-id') ||
      new URL(itemURL).pathname.replace(/^\/+|\/+$/g, ''),
      256,
    )
    if (!listingID) return undefined
    const variationID = bounded(attribute(card, 'data-variation_id') || attribute(card, 'data-variation-id'), 256)
    return {
      listing_id: listingID,
      ...(variationID ? { variation_id: variationID } : {}),
      title,
      price,
      currency: 'AUD',
      url: itemURL,
      image_url: imageURL,
      seller: 'Bonza Slot Cars',
      stock_state: stockState,
      stock_count: stockMatch ? Number(stockMatch[1]) : stockState === 'out_of_stock' ? 0 : -1,
      field_confidence: { title: 0.98, price: 0.96, stock_state: stockState === 'unknown' ? 0.6 : 0.9 },
    }
  }

  const pageNumber = (pageURL, documentRef) => {
    const fromURL = Number(pageURL.searchParams.get('product-page') || pageURL.searchParams.get('paged') || pageURL.searchParams.get('page'))
    if (Number.isInteger(fromURL) && fromURL > 0) return fromURL
    const current = Number(text(firstNode(documentRef, ['.page-numbers.current', '[aria-current="page"]'])))
    return Number.isInteger(current) && current > 0 ? current : 1
  }

  const paginationState = (documentRef, pageURL) => {
    const current = pageNumber(pageURL, documentRef)
    const pageNodes = documentRef.querySelectorAll?.('.page-numbers, [data-page-number]') ?? []
    let total = current
    for (const node of pageNodes) {
      const candidate = Number(attribute(node, 'data-page-number') || text(node))
      if (Number.isInteger(candidate) && candidate > total) total = candidate
    }
    const hasNext = Boolean(firstNode(documentRef, ['a.next.page-numbers', 'a[rel="next"]', '[data-action="load-more"]']))
    return { current, total, has_next: hasNext }
  }

  const hasSucuriMarker = (documentRef) => [...(documentRef.querySelectorAll?.('script') ?? [])]
    .slice(0, maximumScripts)
    .some((script) => String(script?.textContent ?? '').includes('sucuri_cloudproxy_js'))

  const findState = (documentRef) => {
    if (hasSucuriMarker(documentRef) || challengeSelectors.some((selector) => documentRef.querySelector(selector))) return 'challenge'
    if (loggedOutSelectors.some((selector) => documentRef.querySelector(selector))) return 'logged_out'
    const cards = [...(documentRef.querySelectorAll?.(productCardSelector) ?? [])].slice(0, maximumItems)
    return cards.length > 0 ? 'ready' : 'unsupported'
  }

  const capture = (documentRef, pageURL, explicitQuery = '', paginationOverride) => {
    const state = findState(documentRef)
    if (state === 'challenge') return { state, error: 'bonza_challenge_action_required' }
    if (state === 'logged_out') return { state, error: 'bonza_login_required' }
    if (state !== 'ready') return { state: 'unsupported', error: 'bonza_selector_drift' }

    const items = [...documentRef.querySelectorAll(productCardSelector)].slice(0, maximumItems)
      .map((card) => productItem(card, pageURL)).filter(Boolean)
    if (items.length === 0) return { state: 'unsupported', error: 'bonza_selector_drift' }
    const pagination = paginationOverride ?? paginationState(documentRef, pageURL)
    const current = Math.max(1, Number(pagination.current) || 1)
    const total = Math.max(current, Number(pagination.total) || current)
    const complete = pagination.has_next !== true && current >= total
    const query = bounded(explicitQuery || pageURL.searchParams.get('s') || pageURL.searchParams.get('q') || documentRef.title || 'Bonza search', 512)
    const rangeStart = (current - 1) * items.length
    return {
      state: 'ready',
      passive: true,
      payload_type: 'search_results',
      page_complete: complete,
      confidence_score: 0.94,
      data: {
        query,
        page: current,
        page_size: items.length,
        range_start: rangeStart,
        range_end: rangeStart + items.length - 1,
        total_pages: total,
        complete,
        items,
      },
    }
  }

  if (typeof globalThis.__cabinetBonzaCaptureTestHooks === 'function') {
    globalThis.__cabinetBonzaCaptureTestHooks({ capture, parsePrice, productItem, safeURL })
  }

  globalThis.chrome.runtime.onMessage.addListener((message, _sender, sendResponse) => {
    if (message?.type !== 'cabinet:capture' || message.module_id !== moduleID) return false
    sendResponse(capture(document, new URL(location.href)))
    return true
  })
})()
