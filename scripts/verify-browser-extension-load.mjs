import assert from 'node:assert/strict'
import { spawn, spawnSync } from 'node:child_process'
import { existsSync } from 'node:fs'
import { mkdtemp, readFile } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { resolve } from 'node:path'

const browsers = [
  {
    name: 'Chrome',
    commands: ['google-chrome', 'google-chrome-stable'],
    executableVariable: 'CABINET_CHROME_BIN',
    windowsPaths: [
      resolve(process.env.PROGRAMFILES ?? 'C:\\Program Files', 'Google', 'Chrome', 'Application', 'chrome.exe'),
      resolve(
        process.env['PROGRAMFILES(X86)'] ?? 'C:\\Program Files (x86)',
        'Google',
        'Chrome',
        'Application',
        'chrome.exe',
      ),
      resolve(process.env.LOCALAPPDATA ?? '', 'Google', 'Chrome', 'Application', 'chrome.exe'),
    ],
    rootVariable: 'CABINET_EXTENSION_CHROME_ROOT',
    startupTimeoutMilliseconds: 30_000,
    attempts: 1,
  },
  {
    name: 'Edge',
    commands: ['microsoft-edge', 'microsoft-edge-stable'],
    executableVariable: 'CABINET_EDGE_BIN',
    windowsPaths: [
      resolve(
        process.env['PROGRAMFILES(X86)'] ?? 'C:\\Program Files (x86)',
        'Microsoft',
        'Edge',
        'Application',
        'msedge.exe',
      ),
      resolve(process.env.PROGRAMFILES ?? 'C:\\Program Files', 'Microsoft', 'Edge', 'Application', 'msedge.exe'),
      resolve(process.env.LOCALAPPDATA ?? '', 'Microsoft', 'Edge', 'Application', 'msedge.exe'),
    ],
    rootVariable: 'CABINET_EXTENSION_EDGE_ROOT',
    windowsArguments: ['--edge-skip-compat-layer-relaunch'],
    startupTimeoutMilliseconds: 90_000,
    attempts: 2,
  },
]

const executable = (browser) => {
  const override = process.env[browser.executableVariable]
  if (override && existsSync(resolve(override))) return resolve(override)
  if (process.platform === 'win32') return browser.windowsPaths.find((candidate) => existsSync(candidate))
  return browser.commands.find((command) => spawnSync('which', [command], { encoding: 'utf8' }).status === 0)
}

const pause = (milliseconds) => new Promise((resolvePause) => setTimeout(resolvePause, milliseconds))

let nextCommandID = 0

const cdpCommand = (browserProcess, method, params, timeoutMilliseconds = 30_000) => new Promise((resolveCommand, rejectCommand) => {
  const input = browserProcess.stdio[3]
  const output = browserProcess.stdio[4]
  const commandID = ++nextCommandID
  let buffered = ''
  const timeout = setTimeout(() => {
    output.removeListener('data', onData)
    rejectCommand(new Error(`cdp_timeout:${method}`))
  }, timeoutMilliseconds)
  const onData = (chunk) => {
    buffered += chunk.toString()
    const messages = buffered.split('\0')
    buffered = messages.pop() ?? ''
    for (const message of messages.filter(Boolean)) {
      const response = JSON.parse(message)
      if (response.id !== commandID) continue
      clearTimeout(timeout)
      output.removeListener('data', onData)
      if (response.error) rejectCommand(new Error(`cdp_${method}:${response.error.message}`))
      else resolveCommand(response.result)
    }
  }
  output.on('data', onData)
  input.write(`${JSON.stringify({ id: commandID, method, params })}\0`)
})

const tryVerify = async (browser, command, extensionRoot, manifest) => {
  const profile = await mkdtemp(`${tmpdir()}/cabinet-${browser.name.toLowerCase()}-`)
  const output = []
  const browserProcess = spawn(command, [
    ...(process.platform === 'win32' ? (browser.windowsArguments ?? []) : []),
    '--headless=new',
    '--no-sandbox',
    '--disable-gpu',
    '--disable-dev-shm-usage',
    '--disable-background-networking',
    '--disable-component-update',
    '--no-first-run',
    '--no-default-browser-check',
    '--enable-unsafe-extension-debugging',
    '--remote-debugging-pipe',
    `--user-data-dir=${profile}`,
    'about:blank',
  ], { stdio: ['ignore', 'pipe', 'pipe', 'pipe', 'pipe'] })
  browserProcess.stdout.on('data', (chunk) => output.push(chunk.toString()))
  browserProcess.stderr.on('data', (chunk) => output.push(chunk.toString()))

  let loaded
  let installedExtension
  try {
    await cdpCommand(browserProcess, 'Browser.getVersion', {}, browser.startupTimeoutMilliseconds)
    loaded = await cdpCommand(browserProcess, 'Extensions.loadUnpacked', { path: extensionRoot })
    assert.match(loaded?.id ?? '', /^[a-p]{32}$/, `${browser.name} did not return a valid extension ID`)

    const installed = await cdpCommand(browserProcess, 'Extensions.getExtensions', {})
    installedExtension = installed.extensions?.find((extension) => extension.id === loaded.id)
    assert.equal(installedExtension?.enabled, true, `${browser.name} did not enable the unpacked extension`)
    assert.equal(installedExtension?.name, manifest.name, `${browser.name} loaded the wrong extension`)
    assert.equal(installedExtension?.version, manifest.version, `${browser.name} loaded the wrong extension version`)
    assert.equal(resolve(installedExtension?.path ?? ''), extensionRoot, `${browser.name} loaded the extension from the wrong path`)
  } catch (error) {
    throw new Error(`${error.message}\n${output.join('').slice(-4000)}`)
  } finally {
    browserProcess.kill('SIGTERM')
    await Promise.race([
      new Promise((resolveExit) => browserProcess.once('exit', resolveExit)),
      pause(2000),
    ])
  }

  assert.ok(installedExtension, `${browser.name} did not report the unpacked extension after loading`)
  return {
    result: { browser: browser.name, extension_origin: `chrome-extension://${loaded.id}` },
    output: output.join(''),
  }
}

const verify = async (browser) => {
  const command = executable(browser)
  assert.ok(command, `${browser.name} is required for the Browser Companion load gate`)
  const extensionRoot = resolve(process.env[browser.rootVariable] ?? 'browser-extension')
  const manifest = JSON.parse(await readFile(resolve(extensionRoot, 'manifest.json'), 'utf8'))

  const errors = []
  for (let attempt = 1; attempt <= browser.attempts; attempt += 1) {
    try {
      const { result } = await tryVerify(browser, command, extensionRoot, manifest)
      return result
    } catch (error) {
      errors.push(error)
      if (attempt < browser.attempts) await pause(2000)
    }
  }

  const summary = errors.map((error, index) => `attempt ${index + 1}: ${error.message}`).join('\n\n')
  throw new Error(`${browser.name} DevTools load failed after ${browser.attempts} attempt(s):\n${summary.slice(-4000)}`)
}

const results = []
for (const browser of browsers) results.push(await verify(browser))
assert.deepEqual(results.map((result) => result.browser), ['Chrome', 'Edge'])
for (const result of results) console.log(`${result.browser} loaded Cabinet Browser Companion at ${result.extension_origin}`)
