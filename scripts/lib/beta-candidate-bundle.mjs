import { readFile, writeFile } from 'node:fs/promises'

export const createCandidateBundle = async ({
  cabinetManifestPath,
  companionManifestPath,
  outputPath,
  expectedSourceCommit,
}) => {
  if (!/^[a-f0-9]{40}$/.test(expectedSourceCommit ?? '')) throw new Error('candidate_bundle_expected_commit_invalid')
  const cabinet = JSON.parse(await readFile(cabinetManifestPath, 'utf8'))
  const companion = JSON.parse(await readFile(companionManifestPath, 'utf8'))
  if (cabinet.source_commit !== expectedSourceCommit || companion.source_commit !== expectedSourceCommit) {
    throw new Error('candidate_bundle_source_commit_mismatch')
  }
  const controls = {
    'private-beta': { publicationState: 'private_candidate_not_published', product: 'Cabinet 0.1 private beta candidate' },
    ga: { publicationState: 'ga_candidate_not_published', product: 'Cabinet 1.0 GA candidate' },
  }[cabinet.channel]
  if (!controls || companion.channel !== cabinet.channel ||
      cabinet.publication_state !== controls.publicationState || companion.publication_state !== controls.publicationState) {
    throw new Error('candidate_bundle_publication_boundary_invalid')
  }
  const bundle = {
    schema_version: 1,
    product: controls.product,
    channel: cabinet.channel,
    source_commit: expectedSourceCommit,
    publication_state: controls.publicationState,
    components: [
      {
        product: cabinet.product,
        version: cabinet.version,
        manifest_filename: 'cabinet-release-manifest.json',
        release_notes_filename: cabinet.release_notes_filename,
        artifacts: [cabinet.artifact],
        sbom: cabinet.sbom,
      },
      {
        product: companion.product,
        version: companion.version_name,
        manifest_filename: 'browser-companion-release-manifest.json',
        release_notes_filename: companion.release_notes_filename,
        protocol_compatibility: companion.protocol_compatibility,
        artifacts: companion.artifacts.map(({ target, filename, sha256_filename, sha256 }) => ({ target, filename, sha256_filename, sha256 })),
      },
    ],
  }
  await writeFile(outputPath, `${JSON.stringify(bundle, null, 2)}\n`, { flag: 'wx' })
  return bundle
}

export const createBetaCandidateBundle = createCandidateBundle
