import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import test from 'node:test'

const read = (path) => readFile(new URL(`../${path}`, import.meta.url), 'utf8')

test('explicit package workflows create build and CycloneDX SBOM attestations', async () => {
  for (const path of ['.github/workflows/beta-release-candidate.yml', '.github/workflows/release-installers.yml']) {
    const workflow = await read(path)
    assert.match(workflow, /id-token:\s*write/)
    assert.match(workflow, /attestations:\s*write/)
    assert.match(workflow, /actions\/attest@[0-9a-f]{40}/)
    assert.match(workflow, /subject-path:\s*dist\/cabinet\/cabinet-\*-windows-amd64-portable\.zip/)
    assert.match(workflow, /sbom-path:\s*dist\/cabinet\/cabinet-\*-sbom\.cdx\.json/)
    assert.match(workflow, /cabinet-\*-sbom\.cdx\.json/)
  }
})

test('ordinary validation workflows never mint attestations', async () => {
  for (const path of ['.github/workflows/develop-quality-gate.yml', '.github/workflows/main-gate.yml']) {
    const workflow = await read(path)
    assert.doesNotMatch(workflow, /actions\/attest@|attestations:\s*write|id-token:\s*write/)
  }
})

test('approved publication verifies both attestations and publishes the SBOM', async () => {
  const workflow = await read('.github/workflows/publish-beta-prerelease.yml')
  assert.match(workflow, /gh attestation verify.+cabinet-\*-windows-amd64-portable\.zip/s)
  assert.match(workflow, /--predicate-type https:\/\/cyclonedx\.org\/bom/)
  assert.match(workflow, /candidate\/dist\/cabinet\/cabinet-\*-sbom\.cdx\.json/)
  assert.match(workflow, /verify-cabinet-release-package\.mjs/)
  assert.match(workflow, /#1864 approval/)
})

test('package and candidate manifests keep the SBOM in the verified artifact chain', async () => {
  const packageScript = await read('scripts/package-installers.ps1')
  const bundle = await read('scripts/lib/beta-candidate-bundle.mjs')
  const verifier = await read('scripts/lib/cabinet-release-verify.mjs')
  assert.match(packageScript, /create-cabinet-sbom\.mjs/)
  assert.match(packageScript, /CABINET-SBOM\.cdx\.json/)
  assert.match(packageScript, /predicate_type/)
  assert.match(bundle, /sbom:\s*cabinet\.sbom/)
  assert.match(verifier, /verifyCabinetSBOM/)
  assert.match(verifier, /cabinet_sbom_embedded_bytes_mismatch/)
})

test('portable guidance explains local and hosted verification without overstating the evidence', async () => {
  const guidance = await read('openspec/migration/windows-portable-beta.md')
  assert.match(guidance, /CABINET-SBOM\.cdx\.json/)
  assert.match(guidance, /Get-FileHash/)
  assert.match(guidance, /gh attestation verify/)
  assert.match(guidance, /--predicate-type https:\/\/cyclonedx\.org\/bom/)
  assert.match(guidance, /does not prove.+free of vulnerabilities/is)
  assert.match(guidance, /does not replace.+release approval/is)
})

test('private package workflow prepares embedded UI before running the Go suite', async () => {
  const workflow = await read('.github/workflows/release-installers.yml')
  const uiBuild = workflow.indexOf('scripts/build-ui-static.ps1')
  const goTests = workflow.indexOf('go test ./...')
  assert.notEqual(uiBuild, -1)
  assert.notEqual(goTests, -1)
  assert.ok(uiBuild < goTests, 'the static UI must exist before Go resolves internal/ui embed patterns')
})
