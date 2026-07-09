import { useCallback, useEffect, useMemo, useState } from 'react'
import {
  AlertTriangle,
  CheckCircle2,
  FileArchive,
  Info,
  RefreshCw,
} from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'
import { ContentSection } from '../components/content-section'

type ActiveProfileResponse = {
  id?: string
}

type PermissionDeclaration = {
  local_read?: boolean
  local_write?: boolean
  external_read?: boolean
  external_write?: boolean
  secret_access?: boolean
  destructive?: boolean
  requires_confirm?: boolean
}

type AgentSkill = {
  id: string
  version?: string
  display_name?: string
  description?: string
  category?: string
  source?: string
  status?: string
  safety_level?: string
  required_context?: string[]
  required_providers?: string[]
  capabilities?: string[]
  guided_workflows?: string[]
  ui_targets?: string[]
  integration_workflows?: string[]
  permissions?: PermissionDeclaration
  audit_behavior?: string
  provenance?: string
  built_in?: boolean
  removable?: boolean
  enabled?: boolean
  executable?: boolean
  validation_warnings?: string[]
  validation_errors?: string[]
}

type SkillsResponse = {
  skills?: AgentSkill[]
}

type ImportResult = {
  state?: string
  skill?: AgentSkill
  warnings?: string[]
  errors?: string[]
}

type ImportResponse = ImportResult & {
  result?: ImportResult
}

const missingProfileCopy =
  'Select or create a profile before managing imported skills.'

export function SettingsSkills() {
  const [profileID, setProfileID] = useState('')
  const [skills, setSkills] = useState<AgentSkill[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [selectedSkillID, setSelectedSkillID] = useState<string | null>(null)
  const [detailOpen, setDetailOpen] = useState(false)
  const [importOpen, setImportOpen] = useState(false)
  const [importSource, setImportSource] = useState('')
  const [importPending, setImportPending] = useState(false)
  const [importResult, setImportResult] = useState<ImportResult | null>(null)
  const [importError, setImportError] = useState<string | null>(null)
  const [statePendingSkillID, setStatePendingSkillID] = useState<string | null>(
    null
  )
  const [stateError, setStateError] = useState<string | null>(null)

  const loadSkills = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const activeProfileResp = await fetch('/api/profiles/active')
      if (!activeProfileResp.ok) {
        throw new Error('active_profile_unavailable')
      }
      const activeProfile =
        (await activeProfileResp.json()) as ActiveProfileResponse
      const nextProfileID = activeProfile.id?.trim()
      if (!nextProfileID) {
        throw new Error('active_profile_unavailable')
      }
      setProfileID(nextProfileID)

      const skillsResp = await fetch(
        `/api/agent/skills?profile_id=${encodeURIComponent(nextProfileID)}`
      )
      if (!skillsResp.ok) {
        throw new Error('skills_unavailable')
      }
      const payload = (await skillsResp.json()) as SkillsResponse
      setSkills((payload.skills ?? []).filter((skill) => skill.id?.trim()))
    } catch (err) {
      setSkills([])
      setError(
        err instanceof Error && err.message === 'active_profile_unavailable'
          ? missingProfileCopy
          : 'Agent Skills are unavailable right now.'
      )
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void loadSkills()
  }, [loadSkills])

  const selectedSkill = useMemo(
    () => skills.find((skill) => skill.id === selectedSkillID) ?? null,
    [selectedSkillID, skills]
  )

  const summary = useMemo(() => {
    const imported = skills.filter((skill) => skill.source !== 'built-in')
    const needsSetup = skills.filter((skill) =>
      ['requires-implementation', 'preview-only', 'disabled'].includes(
        normalized(skill.status)
      )
    )
    const blocked = skills.filter((skill) =>
      ['blocked', 'invalid'].includes(normalized(skill.status))
    )
    return {
      installed: imported.length,
      enabled: skills.filter((skill) => skill.enabled || skill.built_in).length,
      needsSetup: needsSetup.length,
      blocked: blocked.length,
    }
  }, [skills])

  async function runImport() {
    const sourcePath = importSource.trim()
    if (!profileID || !sourcePath) {
      setImportError(
        profileID
          ? 'Choose a local skill archive or development folder path.'
          : missingProfileCopy
      )
      return
    }
    setImportPending(true)
    setImportError(null)
    setImportResult(null)
    try {
      const response = await fetch('/api/agent/skills/import', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          profile_id: profileID,
          path: sourcePath,
        }),
      })
      const payload = (await response.json()) as ImportResponse
      const result = payload.result ?? payload
      if (!response.ok) {
        const message = result.errors?.[0] || 'Skill import failed.'
        throw new Error(message)
      }
      setImportResult(result ?? null)
      await loadSkills()
    } catch (err) {
      setImportError(
        err instanceof Error ? err.message : 'Skill import failed.'
      )
    } finally {
      setImportPending(false)
    }
  }

  async function updateSkillEnabled(skill: AgentSkill, enabled: boolean) {
    if (!profileID) {
      setStateError(missingProfileCopy)
      return
    }
    if (!canToggleSkill(skill)) {
      setStateError('Built-in skills stay enabled and cannot be changed here.')
      return
    }
    const needsStrongConfirmation =
      enabled &&
      ['external-write', 'destructive'].includes(normalized(skill.safety_level))
    if (
      needsStrongConfirmation &&
      !globalThis.confirm(
        `Enable ${skill.display_name || skill.id}? This skill declares ${labelize(skill.safety_level)} safety and may affect external or destructive workflows.`
      )
    ) {
      return
    }
    setStatePendingSkillID(skill.id)
    setStateError(null)
    try {
      const response = await fetch('/api/agent/skills/state', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          profile_id: profileID,
          skill_id: skill.id,
          enabled,
          confirm: needsStrongConfirmation,
        }),
      })
      if (!response.ok) {
        const payload = (await response.json().catch(() => null)) as
          | { error?: string }
          | null
        throw new Error(payload?.error || 'skill_state_update_failed')
      }
      await loadSkills()
    } catch (err) {
      setStateError(
        err instanceof Error
          ? labelize(err.message) || 'Skill state update failed.'
          : 'Skill state update failed.'
      )
    } finally {
      setStatePendingSkillID(null)
    }
  }

  return (
    <ContentSection
      title='Skills'
      desc='Manage local Cabinet Agent Skills, imports, safety metadata, and setup state.'
      contentClassName='lg:max-w-none'
    >
      <div className='space-y-4 text-sm' data-testid='settings-skills-page'>
        <div className='flex flex-wrap items-center justify-between gap-3'>
          <div className='max-w-2xl text-muted-foreground'>
            Local built-in and imported skills are shown here. Marketplace
            browsing, ratings, payments, and remote installs are not available.
          </div>
          <div className='flex flex-wrap gap-2'>
            <Button
              variant='outline'
              size='sm'
              data-testid='settings-skills-refresh'
              onClick={() => {
                void loadSkills()
              }}
            >
              <RefreshCw className='mr-2 h-4 w-4' />
              Refresh built-ins
            </Button>
            <Button
              size='sm'
              data-testid='settings-skills-import-open'
              onClick={() => {
                setImportOpen(true)
              }}
            >
              <FileArchive className='mr-2 h-4 w-4' />
              Import skill
            </Button>
          </div>
        </div>

        {error ? (
          <div
            className='rounded-md border border-destructive/50 bg-destructive/10 p-3 text-destructive'
            data-testid='settings-skills-error'
          >
            <p className='font-medium'>{error}</p>
          </div>
        ) : null}
        {stateError ? (
          <div
            className='rounded-md border border-destructive/50 bg-destructive/10 p-3 text-destructive'
            data-testid='settings-skills-state-error'
          >
            <p className='font-medium'>{stateError}</p>
          </div>
        ) : null}

        <div className='grid gap-3 sm:grid-cols-2 xl:grid-cols-4'>
          <SummaryTile
            label='Installed skills'
            value={summary.installed}
            testId='settings-skills-summary-installed'
          />
          <SummaryTile
            label='Enabled skills'
            value={summary.enabled}
            testId='settings-skills-summary-enabled'
          />
          <SummaryTile
            label='Needs setup'
            value={summary.needsSetup}
            testId='settings-skills-summary-setup'
          />
          <SummaryTile
            label='Blocked or invalid'
            value={summary.blocked}
            testId='settings-skills-summary-blocked'
          />
        </div>

        <div className='overflow-x-auto rounded-md border'>
          <table
            className='w-full min-w-[980px] table-fixed text-left text-sm'
            data-testid='settings-skills-table'
          >
            <thead className='bg-muted/40 text-xs text-muted-foreground uppercase'>
              <tr>
                <th className='w-[22%] px-3 py-2 font-medium'>Name</th>
                <th className='w-[12%] px-3 py-2 font-medium'>Category</th>
                <th className='w-[10%] px-3 py-2 font-medium'>Source</th>
                <th className='w-[12%] px-3 py-2 font-medium'>Status</th>
                <th className='w-[14%] px-3 py-2 font-medium'>Safety</th>
                <th className='w-[16%] px-3 py-2 font-medium'>Required setup</th>
                <th className='w-[6%] px-3 py-2 font-medium'>Version</th>
                <th className='w-[8%] px-3 py-2 text-right font-medium'>
                  Actions
                </th>
              </tr>
            </thead>
            <tbody>
              {loading ? (
                <tr>
                  <td className='px-3 py-4 text-muted-foreground' colSpan={8}>
                    Loading Agent Skills...
                  </td>
                </tr>
              ) : null}
              {!loading && skills.length === 0 ? (
                <tr>
                  <td className='px-3 py-4 text-muted-foreground' colSpan={8}>
                    No Agent Skills are available for this profile.
                  </td>
                </tr>
              ) : null}
              {skills.map((skill) => (
                <tr
                  key={skill.id}
                  className='border-t align-top'
                  data-testid='settings-skills-row'
                >
                  <td className='px-3 py-2'>
                    <p className='font-medium text-foreground'>
                      {skill.display_name || skill.id}
                    </p>
                    <p className='mt-1 break-all text-xs text-muted-foreground'>
                      {skill.id}
                    </p>
                  </td>
                  <td className='px-3 py-2 text-muted-foreground'>
                    {skill.category || 'Uncategorised'}
                  </td>
                  <td className='px-3 py-2'>
                    <Badge variant='outline'>{sourceLabel(skill)}</Badge>
                  </td>
                  <td className='px-3 py-2'>
                    <StatusBadge status={skill.status} />
                  </td>
                  <td className='px-3 py-2 text-muted-foreground'>
                    {labelize(skill.safety_level)}
                  </td>
                  <td className='px-3 py-2 text-muted-foreground'>
                    {formatList(skill.required_context)}
                  </td>
                  <td className='px-3 py-2 text-muted-foreground'>
                    {skill.version || 'n/a'}
                  </td>
                  <td className='px-3 py-2'>
                    <div className='flex justify-end gap-2'>
                      {canToggleSkill(skill) ? (
                        <Button
                          variant='outline'
                          size='sm'
                          data-testid={`settings-skills-toggle-${skill.id}`}
                          disabled={statePendingSkillID === skill.id}
                          onClick={() => {
                            void updateSkillEnabled(skill, !skill.enabled)
                          }}
                        >
                          {skill.enabled ? 'Disable' : 'Enable'}
                        </Button>
                      ) : null}
                    <Button
                      variant='outline'
                      size='sm'
                      aria-label={`Open ${skill.display_name || skill.id} details`}
                      data-testid={`settings-skills-detail-${skill.id}`}
                      onClick={() => {
                        setSelectedSkillID(skill.id)
                        setDetailOpen(true)
                      }}
                    >
                      <Info className='h-4 w-4' />
                    </Button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        <Sheet open={detailOpen} onOpenChange={setDetailOpen}>
          <SheetContent
            className='w-full overflow-y-auto sm:max-w-xl'
            data-testid='settings-skills-detail-panel'
          >
            <SheetHeader>
              <SheetTitle>
                {selectedSkill?.display_name || selectedSkill?.id || 'Skill'}
              </SheetTitle>
              <SheetDescription>{selectedSkill?.id}</SheetDescription>
            </SheetHeader>
            {selectedSkill ? (
              <SkillDetail
                skill={selectedSkill}
                pending={statePendingSkillID === selectedSkill.id}
                onToggle={(enabled) => {
                  void updateSkillEnabled(selectedSkill, enabled)
                }}
              />
            ) : null}
          </SheetContent>
        </Sheet>

        <Sheet open={importOpen} onOpenChange={setImportOpen}>
          <SheetContent
            className='w-full overflow-y-auto sm:max-w-xl'
            data-testid='settings-skills-import-panel'
          >
            <SheetHeader>
              <SheetTitle>Import skill</SheetTitle>
              <SheetDescription>
                Import a local .cabinet-skill.zip archive or development
                folder. Marketplace browsing is not available.
              </SheetDescription>
            </SheetHeader>
            <div className='space-y-4 px-4 pb-6'>
              <div className='space-y-2'>
                <Label htmlFor='settings-skills-import-source'>
                  Local archive or folder path
                </Label>
                <Input
                  id='settings-skills-import-source'
                  data-testid='settings-skills-import-source'
                  placeholder='C:/cabinet/skills/my-skill.cabinet-skill.zip'
                  value={importSource}
                  onChange={(event) => {
                    setImportSource(event.target.value)
                  }}
                />
              </div>
              <Button
                data-testid='settings-skills-import-submit'
                disabled={importPending}
                onClick={() => {
                  void runImport()
                }}
              >
                {importPending ? 'Validating...' : 'Validate and install'}
              </Button>
              {importError ? (
                <div
                  className='rounded-md border border-destructive/50 bg-destructive/10 p-3 text-destructive'
                  data-testid='settings-skills-import-error'
                >
                  <div className='flex gap-2'>
                    <AlertTriangle className='mt-0.5 h-4 w-4' />
                    <p>{importError}</p>
                  </div>
                </div>
              ) : null}
              {importResult ? (
                <div
                  className='space-y-2 rounded-md border p-3'
                  data-testid='settings-skills-import-result'
                >
                  <div className='flex items-start gap-2'>
                    <CheckCircle2 className='mt-0.5 h-4 w-4 text-emerald-600' />
                    <div>
                      <p className='font-medium'>
                        {labelize(importResult.state) || 'Import complete'}
                      </p>
                      <p className='text-muted-foreground'>
                        {importResult.skill?.display_name ||
                          importResult.skill?.id ||
                          'Imported skill'}{' '}
                        was added through local import.
                      </p>
                    </div>
                  </div>
                  <IssueList
                    title='Warnings'
                    values={importResult.warnings}
                    testId='settings-skills-import-warnings'
                  />
                  <IssueList
                    title='Errors'
                    values={importResult.errors}
                    testId='settings-skills-import-errors'
                  />
                </div>
              ) : null}
            </div>
          </SheetContent>
        </Sheet>
      </div>
    </ContentSection>
  )
}

function SummaryTile({
  label,
  value,
  testId,
}: {
  label: string
  value: number
  testId: string
}) {
  return (
    <div className='rounded-md border p-3' data-testid={testId}>
      <p className='text-xs text-muted-foreground'>{label}</p>
      <p className='mt-1 text-2xl font-semibold'>{value}</p>
    </div>
  )
}

function SkillDetail({
  skill,
  pending,
  onToggle,
}: {
  skill: AgentSkill
  pending: boolean
  onToggle: (enabled: boolean) => void
}) {
  return (
    <div className='space-y-4 px-4 pb-6 text-sm'>
      <p className='text-muted-foreground'>
        {skill.description || 'No description supplied.'}
      </p>
      <div className='grid gap-3 sm:grid-cols-2'>
        <DetailItem label='Source' value={sourceLabel(skill)} />
        <DetailItem label='Status' value={labelize(skill.status)} />
        <DetailItem label='Safety level' value={labelize(skill.safety_level)} />
        <DetailItem
          label='Action Timeline'
          value={skill.audit_behavior || 'Not declared'}
        />
      </div>
      <DetailItem
        label='Permissions'
        value={formatPermissions(skill.permissions)}
      />
      <DetailItem
        label='Required context'
        value={formatList(skill.required_context)}
      />
      <DetailItem
        label='Required integrations'
        value={formatList(skill.required_providers)}
      />
      <DetailItem label='Capabilities' value={formatList(skill.capabilities)} />
      <DetailItem
        label='Guided workflows'
        value={formatList(skill.guided_workflows)}
      />
      <DetailItem label='UI targets' value={formatList(skill.ui_targets)} />
      <DetailItem
        label='Integration workflows'
        value={formatList(skill.integration_workflows)}
      />
      <DetailItem
        label='Source provenance'
        value={skill.provenance || 'Cabinet built-in registry'}
      />
      <IssueList
        title='Validation warnings'
        values={skill.validation_warnings}
        testId='settings-skills-detail-warnings'
      />
      <IssueList
        title='Validation errors'
        values={skill.validation_errors}
        testId='settings-skills-detail-errors'
      />
      <div className='rounded-md border p-3'>
        <p className='font-medium'>Enable / disable</p>
        <p className='mt-1 text-muted-foreground'>
          {skill.built_in
            ? 'Built-in skills stay enabled and cannot be removed by local imports.'
            : skill.enabled
              ? 'This imported skill is enabled for the active profile.'
              : 'This imported skill is disabled for the active profile.'}
        </p>
        {canToggleSkill(skill) ? (
          <Button
            className='mt-3'
            size='sm'
            variant={skill.enabled ? 'outline' : 'default'}
            data-testid={`settings-skills-detail-toggle-${skill.id}`}
            disabled={pending}
            onClick={() => {
              onToggle(!skill.enabled)
            }}
          >
            {skill.enabled ? 'Disable skill' : 'Enable skill'}
          </Button>
        ) : null}
      </div>
    </div>
  )
}

function DetailItem({ label, value }: { label: string; value: string }) {
  return (
    <div className='rounded-md border p-3'>
      <p className='text-xs text-muted-foreground'>{label}</p>
      <p className='mt-1 break-words font-medium'>{value || 'None'}</p>
    </div>
  )
}

function IssueList({
  title,
  values,
  testId,
}: {
  title: string
  values?: string[]
  testId: string
}) {
  if (!values?.length) {
    return null
  }
  return (
    <div className='rounded-md border p-3' data-testid={testId}>
      <p className='font-medium'>{title}</p>
      <ul className='mt-2 list-disc space-y-1 ps-5 text-muted-foreground'>
        {values.map((value) => (
          <li key={value}>{value}</li>
        ))}
      </ul>
    </div>
  )
}

function StatusBadge({ status }: { status?: string }) {
  const value = normalized(status)
  const destructive = value === 'invalid' || value === 'blocked'
  const outline =
    value === 'disabled' || value === 'requires-implementation' || !value
  return (
    <Badge
      variant={destructive ? 'destructive' : outline ? 'outline' : 'secondary'}
    >
      {labelize(status) || 'Unknown'}
    </Badge>
  )
}

function sourceLabel(skill: AgentSkill) {
  if (skill.built_in || skill.source === 'built-in') {
    return 'Built-in'
  }
  return labelize(skill.source) || 'Imported'
}

function canToggleSkill(skill: AgentSkill) {
  return !skill.built_in && skill.source !== 'built-in'
}

function formatList(values?: string[]) {
  const cleaned = (values ?? []).map((value) => value.trim()).filter(Boolean)
  return cleaned.length ? cleaned.join(', ') : 'None'
}

function formatPermissions(permissions?: PermissionDeclaration) {
  if (!permissions) {
    return 'None declared'
  }
  const labels = [
    permissions.local_read ? 'local read' : '',
    permissions.local_write ? 'local write' : '',
    permissions.external_read ? 'external read' : '',
    permissions.external_write ? 'external write' : '',
    permissions.secret_access ? 'secret access' : '',
    permissions.destructive ? 'destructive' : '',
    permissions.requires_confirm ? 'confirmation required' : '',
  ].filter(Boolean)
  return labels.length ? labels.join(', ') : 'None declared'
}

function normalized(value?: string) {
  return value?.trim().toLowerCase() ?? ''
}

function labelize(value?: string) {
  return normalized(value)
    .split(/[-_]/)
    .filter(Boolean)
    .map((part) => part[0]?.toUpperCase() + part.slice(1))
    .join(' ')
}
