import { useTranslation } from 'react-i18next'
import { Apps } from '@/features/apps'

export function Integrations() {
  const { t } = useTranslation('pages')

  return (
    <Apps
      title={t('integrations.title')}
      description={t('integrations.description')}
    />
  )
}
