import { readFile, writeFile } from 'node:fs/promises'
import { resolve } from 'node:path'

import {
  createOrResumeAcceptanceRun,
  recordAcceptanceResult,
  renderAcceptanceMarkdown,
  validateAcceptanceState,
  writeAcceptanceOutputs,
} from './lib/acceptance-evidence-recorder.mjs'

const help = `Cabinet #1869 packaged-acceptance evidence recorder

Evidence-only commands:
  init      verify an exact candidate and create/resume JSON plus Markdown evidence
  record    record one explicit checklist result and refresh both outputs
  validate  validate a saved JSON evidence file
  render    deterministically render Markdown from validated JSON

init required options:
  --cabinet-manifest <path> --companion-manifest <path> --bundle-manifest <path>
  --candidate-run-id <id> --candidate-artifact <name>
  --os-version <value> --host-profile <value>
  --browser-name <value> --browser-version <value>
  --isolated-profile <value> --data-directory <path>
  --app-version <value> --build-revision <full-sha> --build-date <value> --runtime-port <number> --runtime-pid <number>
  --json <path> --markdown <path>

record required options:
  --json <path> --markdown <path> --row <stable-id> --status <blocked|pass|fail>
  --evidence <non-secret-reference> (repeat for pass/fail)
  --notes <text> --unblock <exact-condition> --operator-confirmed

validate required options: --json <path>
render required options: --json <path> [--markdown <path>]

The recorder cannot operate a browser/provider, alter branches, or create a release.
`

const args = process.argv.slice(2)
const command = args[0]
const values = (name) => {
  const result = []
  for (let index = 1; index < args.length; index += 1) if (args[index] === name && args[index + 1] !== undefined) result.push(args[index + 1])
  return result
}
const value = (name) => values(name).at(-1)
const has = (name) => args.includes(name)
const pathValue = (name) => value(name) ? resolve(value(name)) : undefined
const loadState = async (path) => validateAcceptanceState(JSON.parse(await readFile(path, 'utf8')))

const run = async () => {
  if (!command || command === '--help' || command === '-h' || has('--help')) {
    process.stdout.write(help)
    return
  }
  if (command === 'init') {
    const jsonPath = pathValue('--json')
    const markdownPath = pathValue('--markdown')
    const state = await createOrResumeAcceptanceRun({
      cabinetManifestPath: pathValue('--cabinet-manifest'),
      companionManifestPath: pathValue('--companion-manifest'),
      bundleManifestPath: pathValue('--bundle-manifest'),
      releaseCandidateRunId: value('--candidate-run-id'),
      releaseCandidateArtifactName: value('--candidate-artifact'),
      environment: {
        os_version: value('--os-version'),
        host_profile: value('--host-profile'),
        browser_name: value('--browser-name'),
        browser_version: value('--browser-version'),
        isolated_profile: value('--isolated-profile'),
        isolated_data_directory: value('--data-directory'),
        runtime: {
          app_version: value('--app-version'),
          build_revision: value('--build-revision'),
          build_date: value('--build-date'),
          port: Number(value('--runtime-port')),
          pid: Number(value('--runtime-pid')),
        },
      },
      outputPath: jsonPath,
    })
    await writeAcceptanceOutputs({ state, jsonPath, markdownPath })
    process.stdout.write(`Acceptance evidence ready: candidate=${state.candidate.fingerprint}; overall=${state.overall_result}\n`)
    return
  }
  if (command === 'record') {
    const jsonPath = pathValue('--json')
    const markdownPath = pathValue('--markdown')
    const state = await loadState(jsonPath)
    const updated = await recordAcceptanceResult({
      state,
      rowId: value('--row'),
      status: value('--status'),
      evidenceReferences: values('--evidence'),
      operatorNotes: value('--notes') ?? '',
      unblockCondition: value('--unblock') ?? '',
      operatorConfirmed: has('--operator-confirmed'),
    })
    await writeAcceptanceOutputs({ state: updated, jsonPath, markdownPath })
    process.stdout.write(`Recorded ${value('--row')}: ${value('--status')}; overall=${updated.overall_result}\n`)
    return
  }
  if (command === 'validate') {
    const state = await loadState(pathValue('--json'))
    process.stdout.write(`Valid acceptance evidence: candidate=${state.candidate.fingerprint}; overall=${state.overall_result}\n`)
    return
  }
  if (command === 'render') {
    const state = await loadState(pathValue('--json'))
    const markdown = renderAcceptanceMarkdown(state)
    const output = pathValue('--markdown')
    if (output) await writeFile(output, markdown, { flag: 'w' })
    else process.stdout.write(markdown)
    return
  }
  throw new Error(`acceptance_command_unknown:${command}`)
}

run().catch((error) => {
  process.stderr.write(`Acceptance recorder failed: ${error.message}\n`)
  process.exitCode = 1
})
