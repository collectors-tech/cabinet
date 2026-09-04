import { recordNotificationHistory } from '@/lib/toast-history'

type SettingsFeedbackHistoryInput = {
  id: string
  level?: 'success' | 'error' | 'warning' | 'info'
  title: string
  summary: string
  source: string
  sourceLabel: string
}

export function recordSettingsFeedbackHistory({
  id,
  level = 'info',
  title,
  summary,
  source,
  sourceLabel,
}: SettingsFeedbackHistoryInput) {
  recordNotificationHistory({
    id,
    level,
    title,
    summary: `${summary} Source key: ${source}.`,
    source_label: `Settings ${sourceLabel}`,
    category: 'settings',
  })
}
