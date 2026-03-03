import { createFileRoute } from '@tanstack/react-router'
import { SettingsOperations } from '@/features/settings/operations'

export const Route = createFileRoute('/_authenticated/settings/operations')({
  component: SettingsOperations,
})
