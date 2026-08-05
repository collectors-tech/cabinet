import assert from 'node:assert/strict'
import { spawn, spawnSync } from 'node:child_process'
import { mkdtemp } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { resolve } from 'node:path'

const extensionRoot = resolve('browser-extension')
const browsers = [
  { name: 'Chrome', commands: ['google-chrome', 'google-chrome-stable'] },
  { name: 'Edge', commands: ['microsoft-edge', 'microsoft-edge-stable'] },
]

const executable = (commands) => commands.find((command) =>
  spawnSync('which', [command], { encoding: 'utf8' }).status === 0
)

const pause = (milliseconds) => new Promise((resolvePause) => setTimeout(resolvePause, milliseconds))

const targets = async (port) => {
  try {
    const response = await fetch(`http://127.0.0.1:${port}/json/list`)
    return response.ok ? response.json() : []
  } catch {
    return []
  }
}

const verify = async (browser, port) => {
  const command = executable(browser.commands)
  assert.ok(command, `${browser.name} is required for the Browser Companion load gate`)
  const xvfb = executable(['xvfb-run'])
  assert.ok(xvfb, 'xvfb-run is required for the normal-mode browser extension load gate')
  const profile = await mkdtemp(`${tmpdir()}/cabinet-${browser.name.toLowerCase()}-`)
  const output = []
  const process = spawn(xvfb, ['-a', command,
    '--no-sandbox',
    '--disable-gpu',
    `--remote-debugging-port=${port}`,
    `--user-data-dir=${profile}`,
    `--disable-extensions-except=${extensionRoot}`,
    `--load-extension=${extensionRoot}`,
    'about:blank',
  ], { stdio: ['ignore', 'pipe', 'pipe'] })
  process.stdout.on('data', (chunk) => output.push(chunk.toString()))
  process.stderr.on('data', (chunk) => output.push(chunk.toString()))

  let extensionTarget
  for (let attempt = 0; attempt < 50; attempt += 1) {
    const available = await targets(port)
    extensionTarget = available.find((target) =>
      target.type === 'service_worker' && /^chrome-extension:\/\/[^/]+\/background\/service-worker\.mjs$/.test(target.url)
    )
    if (extensionTarget) break
    if (process.exitCode !== null) break
    await pause(200)
  }

  process.kill('SIGTERM')
  await Promise.race([
    new Promise((resolveExit) => process.once('exit', resolveExit)),
    pause(2000),
  ])
  assert.ok(extensionTarget, `${browser.name} did not load the MV3 service worker:\n${output.join('').slice(-4000)}`)
  return { browser: browser.name, extension_origin: new URL(extensionTarget.url).origin }
}

const results = []
for (const [index, browser] of browsers.entries()) results.push(await verify(browser, 9330 + index))
assert.deepEqual(results.map((result) => result.browser), ['Chrome', 'Edge'])
for (const result of results) console.log(`${result.browser} loaded Cabinet Browser Companion at ${result.extension_origin}`)
