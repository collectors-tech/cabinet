import { useTranslation } from 'react-i18next'
import { ContentSection } from '../components/content-section'
import { AppearanceForm } from './appearance-form'

export function SettingsAppearance() {
  const { t } = useTranslation('pages')

  return (
    <ContentSection
      title={t('settings.appearance.title')}
      desc={t('settings.appearance.description')}
    >
      <AppearanceForm />
    </ContentSection>
  )
}
