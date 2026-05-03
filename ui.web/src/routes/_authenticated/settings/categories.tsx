import { createFileRoute } from '@tanstack/react-router'
import { SettingsCategories } from '@/features/settings/categories'

export const Route = createFileRoute('/_authenticated/settings/categories')({
  component: SettingsCategories,
})
