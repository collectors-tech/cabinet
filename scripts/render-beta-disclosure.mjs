import { readFile } from 'node:fs/promises'
import { resolve } from 'node:path'

const option = (name) => {
  const index = process.argv.indexOf(name)
  return index >= 0 ? process.argv[index + 1] : undefined
}

const statusLabel = (status) => status.replaceAll('_', ' ')

export const loadBetaDisclosure = async (sourcePath = resolve('release/cabinet-beta-disclosure.json')) => {
  const disclosure = JSON.parse(await readFile(sourcePath, 'utf8'))
  if (disclosure.schema_version !== 1 || disclosure.release_channel !== 'private-beta') {
    throw new Error('beta_disclosure_identity_invalid')
  }
  if (!Array.isArray(disclosure.statements) || disclosure.statements.length === 0) {
    throw new Error('beta_disclosure_statements_missing')
  }
  const ids = new Set()
  for (const statement of disclosure.statements) {
    if (!statement.id || ids.has(statement.id) || !statement.user_facing || statement.channel !== disclosure.release_channel) {
      throw new Error(`beta_disclosure_statement_invalid:${statement.id ?? 'missing'}`)
    }
    ids.add(statement.id)
  }
  return disclosure
}

export const renderBetaDisclosureMarkdown = (disclosure, { format = 'help-center' } = {}) => {
  const lines = [
    `# ${disclosure.generated_heading}`,
    '',
    `Release channel: ${disclosure.release_channel}`,
    `Release version: ${disclosure.release_version}`,
    '',
  ]

  if (format === 'help-center') {
    lines.push('This article is generated from the governed release disclosure source so the product UI and release notes stay aligned.', '')
  }

  lines.push('## Capability and limitation statements', '')
  for (const statement of disclosure.statements.filter((item) => format === 'release-notes' ? item.release_note : item.ui)) {
    lines.push(`- ${statement.title} (${statusLabel(statement.status)}): ${statement.user_facing}`)
  }

  lines.push('', '## Support and recovery pointers', '')
  for (const pointer of disclosure.support_pointers ?? []) {
    lines.push(`- ${pointer}`)
  }

  return `${lines.join('\n')}\n`
}

if (import.meta.url === `file:///${process.argv[1]?.replaceAll('\\', '/')}`) {
  const source = resolve(option('--source') ?? 'release/cabinet-beta-disclosure.json')
  const format = option('--format') ?? 'help-center'
  const disclosure = await loadBetaDisclosure(source)
  process.stdout.write(renderBetaDisclosureMarkdown(disclosure, { format }))
}
