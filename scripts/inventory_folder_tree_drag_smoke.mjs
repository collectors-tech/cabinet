import { createRequire } from 'module'
import fs from 'fs'
import path from 'path'
import { fileURLToPath, pathToFileURL } from 'url'

const requireFromHere = createRequire(import.meta.url)

function loadPlaywright() {
  const candidates = [
    `${process.cwd()}/package.json`,
    fileURLToPath(new URL('../package.json', import.meta.url)),
    path.resolve(process.cwd(), '../cabinet/package.json'),
  ]

  try {
    return requireFromHere('playwright')
  } catch {
    for (const candidate of candidates) {
      if (!fs.existsSync(candidate)) {
        continue
      }

      try {
        return createRequire(pathToFileURL(candidate).href)('playwright')
      } catch {
        continue
      }
    }

    throw new Error(
      `Unable to resolve "playwright" from ${candidates.join(', ')}`
    )
  }
}

const { chromium } = loadPlaywright()

const baseURL = process.env.CABINET_BASE_URL || 'http://192.168.1.53:17882'
const email = process.env.CABINET_DEMO_EMAIL || 'demo2.owner@cabinet.local'
const password = process.env.CABINET_DEMO_PASSWORD || 'Demo2Review!2026'
const sourceID = process.env.CABINET_DRAG_SOURCE_ID || 'archive-b'
const targetID = process.env.CABINET_DRAG_TARGET_ID || 'warehouses'
const timeoutMs = Number.parseInt(process.env.CABINET_DRAG_TIMEOUT_MS || '20000', 10)
const folderTreeSettingsKey = 'inventory.folder-tree.v1'

function log(step, detail) {
  const suffix = detail ? ` ${detail}` : ''
  console.error(`[inventory-folder-tree-drag-smoke] ${step}${suffix}`)
}

function requireRect(rect, label) {
  if (!rect) {
    throw new Error(`${label} bounding box was not available`)
  }
  return rect
}

async function withTimeout(label, promise, ms = timeoutMs) {
  let timer
  try {
    return await Promise.race([
      promise,
      new Promise((_, reject) => {
        timer = setTimeout(() => {
          reject(new Error(`${label} exceeded ${ms}ms`))
        }, ms)
      }),
    ])
  } finally {
    clearTimeout(timer)
  }
}

async function signInIfNeeded(page) {
  log('goto', `${baseURL}/inventory`)
  await page.goto(`${baseURL}/inventory`, {
    waitUntil: 'domcontentloaded',
    timeout: timeoutMs,
  })
  await page.waitForTimeout(1500)

  if (!page.url().includes('/sign-in')) {
    log('auth', 'already signed in')
    return
  }

  log('auth', 'signing in')
  await page.fill('input[name="email"]', email)
  await page.fill('input[name="password"]', password)
  await page.getByRole('button', { name: 'Sign in', exact: true }).click()
  await page.waitForURL(/\/inventory\/?$/, { timeout: timeoutMs })
  await page.waitForTimeout(2000)
  log('auth', 'sign-in complete')
}

async function ensureTreeReady(page) {
  log('tree', 'waiting for tree')
  await page.locator('[data-testid="inventory-folder-tree"]').waitFor({
    state: 'visible',
    timeout: timeoutMs,
  })
  log('tree', 'tree visible')
}

async function getActiveProfileID(page) {
  const response = await page.evaluate(async () => {
    const result = await fetch('/api/profiles/active')
    if (!result.ok) {
      throw new Error(`active-profile-${result.status}`)
    }
    return result.json()
  })

  const profileID =
    typeof response?.id === 'string'
      ? response.id
      : typeof response?.profile_id === 'string'
        ? response.profile_id
        : ''

  if (!profileID) {
    throw new Error('active profile id was not available')
  }

  return profileID
}

async function readFolderTreeSnapshot(page, profileID) {
  return page.evaluate(async ({ profileID, folderTreeSettingsKey }) => {
    const response = await fetch(
      `/api/profiles/${encodeURIComponent(profileID)}/settings`
    )
    if (!response.ok) {
      throw new Error(`profile-settings-get-${response.status}`)
    }

    const payload = await response.json()
    return {
      settings: payload.settings ?? {},
      folderTreeValue: payload.settings?.[folderTreeSettingsKey] ?? null,
    }
  }, { profileID, folderTreeSettingsKey })
}

async function restoreFolderTreeSnapshot(page, profileID, settings) {
  await page.evaluate(
    async ({ profileID, settings }) => {
      const response = await fetch(
        `/api/profiles/${encodeURIComponent(profileID)}/settings`,
        {
          method: 'PUT',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ settings }),
        }
      )

      if (!response.ok) {
        throw new Error(`profile-settings-put-${response.status}`)
      }
    },
    { profileID, settings }
  )
}

async function handleHitTest(page, handleSelector) {
  return page.evaluate((selector) => {
    const handle = document.querySelector(selector)
    if (!(handle instanceof HTMLElement)) {
      return { ok: false, reason: 'handle-missing' }
    }

    const rect = handle.getBoundingClientRect()
    const hit = document.elementFromPoint(
      rect.left + rect.width / 2,
      rect.top + rect.height / 2
    )

    return {
      ok: Boolean(hit?.closest(selector)),
      reason: hit ? null : 'hit-target-missing',
      rect: {
        left: rect.left,
        right: rect.right,
        top: rect.top,
        bottom: rect.bottom,
        width: rect.width,
        height: rect.height,
      },
      hit: hit
        ? {
            tag: hit.tagName,
            testid: hit.getAttribute('data-testid'),
            title: hit.getAttribute('title'),
            closestHandleTitle: hit.closest('[title]')?.getAttribute('title') ?? null,
          }
        : null,
    }
  }, handleSelector)
}

async function countNestedTarget(page, targetID, sourceID) {
  return page
    .locator(
      `[data-testid="folder-tree-group-${targetID}"] [data-testid="folder-tree-item-${sourceID}"]`
    )
    .count()
}

async function dragFolderIntoTarget(page, sourceID, targetID) {
  log('drag', `prepare ${sourceID} -> ${targetID}`)
  const handleSelector = `[data-testid="folder-tree-drag-handle-${sourceID}"]`
  const targetSelector = `[data-testid="folder-tree-item-${targetID}"]`
  const rootHandle = page.locator(handleSelector)
  const target = page.locator(targetSelector)

  await rootHandle.waitFor({ state: 'visible', timeout: timeoutMs })
  await target.waitFor({ state: 'visible', timeout: timeoutMs })

  const before = await countNestedTarget(page, targetID, sourceID)
  const hitTest = await handleHitTest(page, handleSelector)
  if (!hitTest.ok) {
    throw new Error(
      `drag handle for ${sourceID} is not pointer-hit-testable: ${JSON.stringify(hitTest)}`
    )
  }

  const handleBox = requireRect(await rootHandle.boundingBox(), `handle ${sourceID}`)
  const targetBox = requireRect(await target.boundingBox(), `target ${targetID}`)
  log(
    'drag',
    `boxes handle=(${handleBox.x.toFixed(1)},${handleBox.y.toFixed(1)}) target=(${targetBox.x.toFixed(1)},${targetBox.y.toFixed(1)})`
  )

  await withTimeout(
    `drag step ${sourceID} -> ${targetID}`,
    rootHandle.dragTo(target, {
      timeout: timeoutMs,
      sourcePosition: {
        x: handleBox.width / 2,
        y: handleBox.height / 2,
      },
      targetPosition: {
        x: Math.max(16, Math.min(targetBox.width * 0.35, targetBox.width - 16)),
        y: targetBox.height / 2,
      },
    })
  )

  await page.waitForTimeout(1200)
  const after = await countNestedTarget(page, targetID, sourceID)
  log('drag', `counts before=${before} after=${after}`)

  if (after <= before) {
    const activeContext =
      (await page
        .locator('[data-testid="collection-active-context"]')
        .textContent()
        .catch(() => null)) ?? '<missing>'
    throw new Error(
      [
        `folder drag did not move ${sourceID} under ${targetID}`,
        `before nested count: ${before}`,
        `after nested count: ${after}`,
        `active context: ${activeContext.trim()}`,
        `handle hit-test: ${JSON.stringify(hitTest)}`,
      ].join('\n')
    )
  }

  return { before, after, hitTest }
}

const browser = await chromium.launch({
  headless: process.env.CABINET_HEADLESS !== 'false',
})

let page
let profileID = ''
let snapshot = null

try {
  log('browser', 'launching page')
  page = await browser.newPage({ viewport: { width: 1440, height: 1200 } })

  await signInIfNeeded(page)
  await ensureTreeReady(page)
  log('profile', 'loading active profile')
  profileID = await getActiveProfileID(page)
  snapshot = await readFolderTreeSnapshot(page, profileID)
  log('profile', `snapshot loaded for ${profileID}`)

  const result = await dragFolderIntoTarget(page, sourceID, targetID)
  console.log(
    JSON.stringify(
      {
        ok: true,
        baseURL,
        profileID,
        sourceID,
        targetID,
        before: result.before,
        after: result.after,
        hitTest: result.hitTest,
      },
      null,
      2
    )
  )
} finally {
  if (page && profileID && snapshot) {
    log('profile', `restoring snapshot for ${profileID}`)
    await restoreFolderTreeSnapshot(page, profileID, snapshot.settings).catch(
      (error) => {
        console.error(
          `failed to restore folder tree snapshot for ${profileID}: ${String(error)}`
        )
      }
    )
  }
  log('browser', 'closing')
  await browser.close()
}
