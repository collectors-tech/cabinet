import { createFileRoute } from '@tanstack/react-router'
import { SettingsIntegrations } from '@/features/settings/integrations'

export const Route = createFileRoute('/_authenticated/settings/integrations')({
  component: SettingsIntegrations,
})
