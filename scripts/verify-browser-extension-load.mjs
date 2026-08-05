import assert from 'node:assert/strict'
import { spawn, spawnSync } from 'node:child_process'
import { mkdtemp, readFile } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { resolve } from 'node:path'

const browsers = [
  { name: 'Chrome', commands: ['google-chrome', 'google-chrome-stable'], rootVariable: 'CABINET_EXTENSION_CHROME_ROOT' },
  { name: 'Edge', commands: ['microsoft-edge', 'microsoft-edge-stable'], rootVariable: 'CABINET_EXTENSION_EDGE_ROOT' },
]

const executable = (commands) => commands.find((command) =>
  spawnSync('which', [command], { encoding: 'utf8' }).status === 0
)

const pause = (milliseconds) => new Promise((resolvePause) => setTimeout(resolvePause, milliseconds))

let nextCommandID = 0

const cdpCommand = (browserProcess, method, params) => new Promise((resolveCommand, rejectCommand) => {
  const input = browserProcess.stdio[3]
  const output = browserProcess.stdio[4]
  const commandID = ++nextCommandID
  let buffered = ''
  const timeout = setTimeout(() => {
    output.removeListener('data', onData)
    rejectCommand(new Error(`cdp_timeout:${method}`))
  }, 30_000)
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

const verify = async (browser) => {
  const command = executable(browser.commands)
  assert.ok(command, `${browser.name} is required for the Browser Companion load gate`)
  const extensionRoot = resolve(process.env[browser.rootVariable] ?? 'browser-extension')
  const manifest = JSON.parse(await readFile(resolve(extensionRoot, 'manifest.json'), 'utf8'))
  const profile = await mkdtemp(`${tmpdir()}/cabinet-${browser.name.toLowerCase()}-`)
  const output = []
  const process = spawn(command, [
    '--headless=new',
    '--no-sandbox',
    '--disable-gpu',
    '--disable-dev-shm-usage',
    '--no-first-run',
    '--enable-unsafe-extension-debugging',
    '--remote-debugging-pipe',
    `--user-data-dir=${profile}`,
    'about:blank',
  ], { stdio: ['ignore', 'pipe', 'pipe', 'pipe', 'pipe'] })
  process.stdout.on('data', (chunk) => output.push(chunk.toString()))
  process.stderr.on('data', (chunk) => output.push(chunk.toString()))

  let loaded
  let installedExtension
  try {
    await cdpCommand(process, 'Browser.getVersion', {})
    loaded = await cdpCommand(process, 'Extensions.loadUnpacked', { path: extensionRoot })
    assert.match(loaded?.id ?? '', /^[a-p]{32}$/, `${browser.name} did not return a valid extension ID`)

    const installed = await cdpCommand(process, 'Extensions.getExtensions', {})
    installedExtension = installed.extensions?.find((extension) => extension.id === loaded.id)
    assert.equal(installedExtension?.enabled, true, `${browser.name} did not enable the unpacked extension`)
    assert.equal(installedExtension?.name, manifest.name, `${browser.name} loaded the wrong extension`)
    assert.equal(installedExtension?.version, manifest.version, `${browser.name} loaded the wrong extension version`)
    assert.equal(resolve(installedExtension?.path ?? ''), extensionRoot, `${browser.name} loaded the extension from the wrong path`)
  } catch (error) {
    throw new Error(`${browser.name} DevTools load failed: ${error.message}\n${output.join('').slice(-4000)}`)
  } finally {
    process.kill('SIGTERM')
    await Promise.race([
      new Promise((resolveExit) => process.once('exit', resolveExit)),
      pause(2000),
    ])
  }

  assert.ok(installedExtension, `${browser.name} did not report the unpacked extension after loading`)
  return { browser: browser.name, extension_origin: `chrome-extension://${loaded.id}` }
}

const results = []
for (const browser of browsers) results.push(await verify(browser))
assert.deepEqual(results.map((result) => result.browser), ['Chrome', 'Edge'])
for (const result of results) console.log(`${result.browser} loaded Cabinet Browser Companion at ${result.extension_origin}`)
