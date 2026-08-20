import { createHash } from 'node:crypto'
import { readFile } from 'node:fs/promises'
import { basename, dirname, join, resolve } from 'node:path'
import { inflateRawSync } from 'node:zlib'

const sha256File = async (path) => createHash('sha256').update(await readFile(path)).digest('hex')
const safeFilename = (value) => typeof value === 'string' && value === basename(value) && value.length > 0
const safeArchivePath = (value) => typeof value === 'string' && value.length > 0 && !value.startsWith('/') &&
  !value.includes('\\') && !value.split('/').includes('..') && !value.endsWith('/')
const zipCentralSignature = 0x02014b50
const zipEndSignature = 0x06054b50
const zipLocalSignature = 0x04034b50

const readZipEntries = (archive) => {
  let endOffset = -1
  for (let offset = archive.length - 22; offset >= Math.max(0, archive.length - 65_557); offset -= 1) {
    if (archive.readUInt32LE(offset) === zipEndSignature) { endOffset = offset; break }
  }
  if (endOffset < 0) throw new Error('cabinet_zip_end_record_missing')
  const count = archive.readUInt16LE(endOffset + 10)
  let offset = archive.readUInt32LE(endOffset + 16)
  const entries = new Map()
  for (let index = 0; index < count; index += 1) {
    if (offset + 46 > archive.length || archive.readUInt32LE(offset) !== zipCentralSignature) throw new Error('cabinet_zip_central_record_invalid')
    const flags = archive.readUInt16LE(offset + 8)
    const method = archive.readUInt16LE(offset + 10)
    const compressedSize = archive.readUInt32LE(offset + 20)
    const size = archive.readUInt32LE(offset + 24)
    const nameLength = archive.readUInt16LE(offset + 28)
    const extraLength = archive.readUInt16LE(offset + 30)
    const commentLength = archive.readUInt16LE(offset + 32)
    const localOffset = archive.readUInt32LE(offset + 42)
    const name = archive.subarray(offset + 46, offset + 46 + nameLength).toString('utf8')
    if ((flags & 0x0001) !== 0 || ![0, 8].includes(method) || !safeArchivePath(name) || entries.has(name)) {
      throw new Error(`cabinet_zip_entry_policy_invalid:${name}`)
    }
    if (localOffset + 30 > archive.length || archive.readUInt32LE(localOffset) !== zipLocalSignature) throw new Error(`cabinet_zip_local_record_invalid:${name}`)
    const localNameLength = archive.readUInt16LE(localOffset + 26)
    const localExtraLength = archive.readUInt16LE(localOffset + 28)
    const dataOffset = localOffset + 30 + localNameLength + localExtraLength
    const compressed = archive.subarray(dataOffset, dataOffset + compressedSize)
    if (compressed.length !== compressedSize) throw new Error(`cabinet_zip_entry_truncated:${name}`)
    const data = method === 0 ? Buffer.from(compressed) : inflateRawSync(compressed)
    if (data.length !== size) throw new Error(`cabinet_zip_entry_size_invalid:${name}`)
    entries.set(name, data)
    offset += 46 + nameLength + extraLength + commentLength
  }
  return entries
}

export const verifyCabinetReleasePackage = async (manifestPath, {
  repositoryRoot = resolve('.'),
  expectedSourceCommit,
} = {}) => {
  const canonical = JSON.parse(await readFile(join(repositoryRoot, 'release', 'cabinet-beta-version.json'), 'utf8'))
  const release = JSON.parse(await readFile(manifestPath, 'utf8'))
  const outputDirectory = dirname(manifestPath)

  if (release.schema_version !== 1 || release.product !== 'Cabinet' || release.channel !== canonical.channel ||
      release.version !== canonical.version || release.publication_state !== 'private_candidate_not_published') {
    throw new Error('cabinet_release_manifest_identity_invalid')
  }
  if (!/^[a-f0-9]{40}$/.test(release.source_commit) ||
      (expectedSourceCommit && release.source_commit !== expectedSourceCommit.toLowerCase())) {
    throw new Error('cabinet_release_source_commit_mismatch')
  }
  if (!/^\d{4}-\d{2}-\d{2}T/.test(release.build_date ?? '')) throw new Error('cabinet_release_build_date_invalid')

  const expectedFilename = `cabinet-${release.version}-windows-amd64-portable.zip`
  const expectedChecksumFilename = `${expectedFilename}.sha256`
  const expectedNotesFilename = `cabinet-${release.version}-release-notes.md`
  const artifact = release.artifact
  if (artifact?.target !== 'windows-amd64' || artifact?.kind !== 'portable_zip' || artifact?.filename !== expectedFilename ||
      artifact?.sha256_filename !== expectedChecksumFilename || !safeFilename(artifact.filename) || !safeFilename(artifact.sha256_filename) ||
      release.release_notes_filename !== expectedNotesFilename || !safeFilename(release.release_notes_filename)) {
    throw new Error('cabinet_release_artifact_identity_invalid')
  }

  const archivePath = join(outputDirectory, expectedFilename)
  const archive = await readFile(archivePath)
  const actualHash = createHash('sha256').update(archive).digest('hex')
  if (actualHash !== artifact.sha256 || archive.length !== artifact.size_bytes) throw new Error('cabinet_artifact_checksum_mismatch')
  const checksum = await readFile(join(outputDirectory, expectedChecksumFilename), 'utf8')
  if (checksum !== `${actualHash}  ${expectedFilename}\n`) throw new Error('cabinet_checksum_file_invalid')
  const notes = await readFile(join(outputDirectory, expectedNotesFilename), 'utf8')
  if (!notes.includes(release.version) || !notes.includes(release.source_commit) || !/portable package/i.test(notes) || !/not an installer/i.test(notes)) {
    throw new Error('cabinet_release_notes_identity_invalid')
  }
  await verifyCabinetReleaseDisclosure(notes, repositoryRoot)

  if (!Array.isArray(release.package_files) || release.package_files.length < 4) throw new Error('cabinet_package_file_inventory_invalid')
  const paths = new Set()
  const recordedFiles = new Map()
  for (const file of release.package_files) {
    if (!safeArchivePath(file.path) || paths.has(file.path) || !Number.isInteger(file.size_bytes) || file.size_bytes < 1 || !/^[a-f0-9]{64}$/.test(file.sha256)) {
      throw new Error('cabinet_package_file_inventory_invalid')
    }
    paths.add(file.path)
    recordedFiles.set(file.path, file)
  }
  for (const required of ['cabinet.exe', 'cabinet-mcp.exe', 'README.md', 'WINDOWS-PORTABLE-BETA.md']) {
    if (!paths.has(required)) throw new Error(`cabinet_required_package_file_missing:${required}`)
  }
  const archivedFiles = readZipEntries(archive)
  if ([...archivedFiles.keys()].sort().join('\n') !== [...paths].sort().join('\n')) throw new Error('cabinet_zip_file_inventory_mismatch')
  const portableGuide = archivedFiles.get('WINDOWS-PORTABLE-BETA.md').toString('utf8')
  if (!portableGuide.includes(`Cabinet \`${release.version}\``) ||
      !portableGuide.includes(`\`${expectedFilename}\``) ||
      /\{\{CABINET_[A-Z_]+\}\}/.test(portableGuide)) {
    throw new Error('cabinet_portable_guide_version_mismatch')
  }
  for (const [path, data] of archivedFiles) {
    const recorded = recordedFiles.get(path)
    if (recorded.size_bytes !== data.length || recorded.sha256 !== createHash('sha256').update(data).digest('hex')) {
      throw new Error(`cabinet_zip_file_hash_mismatch:${path}`)
    }
  }

  // Keep an independently computed helper call in this module for auditability.
  if (await sha256File(archivePath) !== actualHash) throw new Error('cabinet_artifact_checksum_race')
  return release
}

export const verifyCabinetReleaseDisclosure = async (notes, repositoryRoot = resolve('.')) => {
  const disclosure = JSON.parse(await readFile(join(repositoryRoot, 'release', 'cabinet-beta-disclosure.json'), 'utf8'))
  if (disclosure.schema_version !== 1 || disclosure.release_channel !== 'private-beta') {
    throw new Error('cabinet_release_disclosure_identity_invalid')
  }
  for (const statement of disclosure.statements ?? []) {
    if (statement.release_note !== true) continue
    if (!notes.includes(statement.title) || !notes.includes(statement.user_facing)) {
      throw new Error(`cabinet_release_disclosure_missing:${statement.id}`)
    }
  }
  for (const pointer of disclosure.support_pointers ?? []) {
    if (!notes.includes(pointer)) throw new Error('cabinet_release_disclosure_support_pointer_missing')
  }
}
