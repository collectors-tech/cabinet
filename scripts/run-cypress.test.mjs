import test from 'node:test'
import assert from 'node:assert/strict'
import {
  buildCypressInvocation,
  sanitizeCypressEnv,
} from './run-cypress.mjs'

test('sanitizeCypressEnv removes ELECTRON_RUN_AS_NODE and keeps other values', () => {
  const env = sanitizeCypressEnv({
    ELECTRON_RUN_AS_NODE: '1',
    PATH: 'C:\\tools',
    CYPRESS_CACHE_FOLDER: 'C:\\cache',
  })

  assert.equal(env.ELECTRON_RUN_AS_NODE, undefined)
  assert.equal(env.PATH, 'C:\\tools')
  assert.equal(env.CYPRESS_CACHE_FOLDER, 'C:\\cache')
})

test('buildCypressInvocation uses cmd.exe + npx on Windows and preserves argv passthrough', () => {
  const invocation = buildCypressInvocation(
    ['run', '--browser', 'electron', '--spec', 'cypress/e2e/foo.cy.ts'],
    {
      platform: 'win32',
      baseEnv: {
        ELECTRON_RUN_AS_NODE: '1',
        PATH: 'C:\\tools',
      },
    }
  )

  assert.equal(invocation.command, 'cmd.exe')
  assert.deepEqual(invocation.args, [
    '/d',
    '/s',
    '/c',
    'npx',
    'cypress',
    'run',
    '--browser',
    'electron',
    '--spec',
    'cypress/e2e/foo.cy.ts',
  ])
  assert.equal(invocation.env.ELECTRON_RUN_AS_NODE, undefined)
  assert.equal(invocation.env.PATH, 'C:\\tools')
  assert.equal(invocation.clearedElectronRunAsNode, true)
})

test('buildCypressInvocation uses npx outside Windows and reports when nothing was cleared', () => {
  const invocation = buildCypressInvocation(['open'], {
    platform: 'linux',
    baseEnv: {
      PATH: '/usr/bin',
    },
  })

  assert.equal(invocation.command, 'npx')
  assert.deepEqual(invocation.args, ['cypress', 'open'])
  assert.equal(invocation.env.PATH, '/usr/bin')
  assert.equal(invocation.clearedElectronRunAsNode, false)
})
