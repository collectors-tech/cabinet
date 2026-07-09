import { createFileRoute } from '@tanstack/react-router'
import { SettingsSkills } from '@/features/settings/skills'

export const Route = createFileRoute('/_authenticated/settings/skills')({
  component: SettingsSkills,
})
