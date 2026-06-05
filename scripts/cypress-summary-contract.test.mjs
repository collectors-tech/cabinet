import test from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { dirname, resolve } from 'node:path'

const repoRoot = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const runnerSource = readFileSync(resolve(repoRoot, 'cypress.ps1'), 'utf8')

test('cypress.ps1 writes lane isolation metadata to the run summary', () => {
  const expectedFields = [
    'runtime_port',
    'runtime_data_dir',
    'runtime_profile',
    'runtime_instance_name',
    'source_commit',
  ]

  for (const field of expectedFields) {
    assert.match(
      runnerSource,
      new RegExp(`\\b${field}\\s*=`),
      `missing summary field ${field}`
    )
  }

  const expectedArguments = [
    '-runtimePort $runtimePort',
    '-runtimeDataDir $e2eDataDir',
    '-runtimeProfile $e2eProfile',
    '-runtimeInstanceName $e2eInstanceName',
    '-sourceCommit $sourceCommit',
  ]

  for (const argument of expectedArguments) {
    assert.ok(
      runnerSource.includes(argument),
      `missing Write-RunSummary argument: ${argument}`
    )
  }
})

test('cypress.ps1 logs lane isolation metadata before validation starts', () => {
  assert.match(
    runnerSource,
    /Write-Step "Lane isolation: port=\$runtimePort data_dir=\$e2eDataDir profile=\$e2eProfile instance=\$e2eInstanceName commit=\$sourceCommit"/
  )
})
