import { spawn } from 'node:child_process'
import { pathToFileURL } from 'node:url'

export function sanitizeCypressEnv(baseEnv = process.env) {
  const nextEnv = { ...baseEnv }
  delete nextEnv.ELECTRON_RUN_AS_NODE
  return nextEnv
}

export function buildCypressInvocation(
  cliArgs,
  {
    platform = process.platform,
    baseEnv = process.env,
  } = {}
) {
  const windows = platform === 'win32'
  return {
    command: windows ? 'cmd.exe' : 'npx',
    args: windows
      ? ['/d', '/s', '/c', 'npx', 'cypress', ...cliArgs]
      : ['cypress', ...cliArgs],
    env: sanitizeCypressEnv(baseEnv),
    clearedElectronRunAsNode:
      typeof baseEnv.ELECTRON_RUN_AS_NODE !== 'undefined',
  }
}

async function main() {
  const cliArgs = process.argv.slice(2)
  const invocation = buildCypressInvocation(cliArgs)

  if (invocation.clearedElectronRunAsNode) {
    console.error(
      '[run-cypress] Clearing ELECTRON_RUN_AS_NODE for Cypress compatibility.'
    )
  }

  const child = spawn(invocation.command, invocation.args, {
    stdio: 'inherit',
    env: invocation.env,
    shell: false,
  })

  child.on('error', (error) => {
    console.error(`[run-cypress] Failed to start Cypress: ${String(error)}`)
    process.exitCode = 1
  })

  child.on('exit', (code, signal) => {
    if (signal) {
      process.kill(process.pid, signal)
      return
    }
    process.exitCode = code ?? 1
  })
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) {
  await main()
}
