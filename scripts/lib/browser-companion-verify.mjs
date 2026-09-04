import { createHash } from 'node:crypto'
import { readFile } from 'node:fs/promises'
import { dirname, join, resolve } from 'node:path'

import { readStoreZip, sha256Buffer } from './browser-companion-package.mjs'

const exactPaths = (values) => [...values].sort().join('\n')
const equalJSON = (left, right) => JSON.stringify(left) === JSON.stringify(right)

export const scanPackagedFiles = (files) => {
  for (const [path, data] of files) {
    if (!path || path.startsWith('/') || path.includes('\\') || path.split('/').includes('..')) throw new Error(`unsafe_package_path_rejected:${path}`)
    if (/\.map$/i.test(path)) throw new Error(`source_map_rejected:${path}`)
    if (/(^|\/)(\.env(?:\.|$)|[^/]+\.(?:pem|key|p12|pfx))$/i.test(path)) throw new Error(`secret_file_rejected:${path}`)
    const source = Buffer.from(data).toString('utf8')
    if (/-----BEGIN (?:RSA |EC |OPENSSH )?PRIVATE KEY-----/.test(source)) throw new Error(`private_key_material_rejected:${path}`)
    if (/\b(?:AKIA[0-9A-Z]{16}|gh[pousr]_[A-Za-z0-9]{30,}|xox[baprs]-[A-Za-z0-9-]{20,})\b/.test(source)) throw new Error(`credential_material_rejected:${path}`)
    if (/\beval\s*\(|\bnew\s+Function\s*\(/.test(source)) throw new Error(`dynamic_code_execution_rejected:${path}`)
    if (path.startsWith('modules/') && /document\.cookie|chrome\.cookies|\batob\s*\(|String\.fromCharCode|XMLHttpRequest|\bfetch\s*\(|\.click\s*\(|\.submit\s*\(/.test(source)) {
      throw new Error(`provider_challenge_or_session_bypass_rejected:${path}`)
    }
  }
  return true
}

const sha256File = async (path) => createHash('sha256').update(await readFile(path)).digest('hex')

export const verifyBrowserCompanionRelease = async (releaseManifestPath, {
  repositoryRoot = resolve('.'),
  expectedSourceCommit,
  previousManifestPath,
} = {}) => {
  const releaseConfig = JSON.parse(await readFile(join(repositoryRoot, 'browser-extension', 'release.json'), 'utf8'))
  const release = JSON.parse(await readFile(releaseManifestPath, 'utf8'))
  const outputDirectory = dirname(releaseManifestPath)
  const expectedVersionName = `${releaseConfig.version_name_prefix}.g${release.source_commit?.slice(0, 12)}`
  if (release.schema_version !== 1 || release.channel !== releaseConfig.channel || release.version !== releaseConfig.version ||
      release.version_name !== expectedVersionName || release.distribution !== releaseConfig.distribution ||
      release.automatic_updates !== false || release.publication_state !== 'private_candidate_not_published') {
    throw new Error('release_manifest_version_or_channel_drift')
  }
  if (!/^[a-f0-9]{40}$/.test(release.source_commit) || (expectedSourceCommit && release.source_commit !== expectedSourceCommit.toLowerCase())) {
    throw new Error('release_manifest_source_commit_mismatch')
  }
  if (!equalJSON(release.protocol_compatibility, releaseConfig.protocol_compatibility)) throw new Error('release_manifest_protocol_drift')
  if (release.immutable_tag !== `${releaseConfig.immutable_tag_prefix}${release.version_name}`) throw new Error('release_manifest_immutable_tag_drift')
  if (!Array.isArray(release.artifacts) || release.artifacts.map((item) => item.target).join(',') !== releaseConfig.targets.join(',')) {
    throw new Error('release_manifest_targets_invalid')
  }
  if (JSON.stringify(release).includes('latest')) throw new Error('mutable_latest_reference_rejected')

  if (previousManifestPath) {
    const previous = JSON.parse(await readFile(previousManifestPath, 'utf8'))
    if (previous.version_name === release.version_name && (previous.source_commit !== release.source_commit ||
        previous.artifacts?.map((item) => item.sha256).join(',') !== release.artifacts.map((item) => item.sha256).join(','))) {
      throw new Error('immutable_version_reused')
    }
  }

  const expectedPaths = exactPaths([...releaseConfig.package_files, 'manifest.json', 'release-channel.json'])
  for (const artifact of release.artifacts) {
    const expectedFilename = `cabinet-browser-companion-${release.version_name}-${artifact.target}.zip`
    if (artifact.filename !== expectedFilename || artifact.sha256_filename !== `${expectedFilename}.sha256`) throw new Error(`artifact_filename_drift:${artifact.target}`)
    const archivePath = join(outputDirectory, artifact.filename)
    const actualHash = await sha256File(archivePath)
    if (actualHash !== artifact.sha256) throw new Error(`artifact_checksum_mismatch:${artifact.target}`)
    const checksumLine = await readFile(join(outputDirectory, artifact.sha256_filename), 'utf8')
    if (checksumLine !== `${actualHash}  ${artifact.filename}\n`) throw new Error(`artifact_checksum_file_invalid:${artifact.target}`)
    const files = await readStoreZip(archivePath)
    if (exactPaths(files.keys()) !== expectedPaths) throw new Error(`artifact_file_allow_list_mismatch:${artifact.target}`)
    scanPackagedFiles(files)
    const recordedFiles = new Map(artifact.files.map((file) => [file.path, file]))
    for (const [path, data] of files) {
      const recorded = recordedFiles.get(path)
      if (!recorded || recorded.size_bytes !== data.length || recorded.sha256 !== sha256Buffer(data)) throw new Error(`artifact_file_manifest_mismatch:${artifact.target}:${path}`)
    }
    if (recordedFiles.size !== files.size) throw new Error(`artifact_file_manifest_count_mismatch:${artifact.target}`)
    const manifest = JSON.parse(files.get('manifest.json').toString('utf8'))
    const channel = JSON.parse(files.get('release-channel.json').toString('utf8'))
    if (manifest.name !== 'Cabinet Browser Companion' || manifest.version !== release.version || manifest.version_name !== release.version_name ||
        !equalJSON(manifest.permissions, releaseConfig.permissions) || !equalJSON(manifest.host_permissions, releaseConfig.host_permissions) ||
        !equalJSON(manifest.optional_host_permissions, releaseConfig.optional_host_permissions) || manifest.host_permissions.includes('https://*/*') ||
        !equalJSON(manifest.content_security_policy, { extension_pages: "script-src 'self'; object-src 'self'" }) ||
        manifest.key || manifest.update_url || manifest.oauth2 || manifest.externally_connectable) {
      throw new Error(`production_manifest_policy_invalid:${artifact.target}`)
    }
    if (channel.target !== artifact.target || channel.source_commit !== release.source_commit || channel.channel !== release.channel ||
        channel.version_name !== release.version_name || channel.automatic_updates !== false ||
        !equalJSON(channel.protocol_compatibility, release.protocol_compatibility)) {
      throw new Error(`release_channel_identity_invalid:${artifact.target}`)
    }
  }
  const expectedReleaseNotes = `cabinet-browser-companion-${release.version_name}-release-notes.md`
  if (release.release_notes_filename !== expectedReleaseNotes) throw new Error('release_notes_filename_invalid')
  await readFile(join(outputDirectory, expectedReleaseNotes), 'utf8')
  await readFile(join(outputDirectory, 'browser-companion-candidate-summary.md'), 'utf8')
  return release
}
