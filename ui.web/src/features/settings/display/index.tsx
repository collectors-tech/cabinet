import { useTranslation } from 'react-i18next'
import { ContentSection } from '../components/content-section'
import { DisplayForm } from './display-form'

export function SettingsDisplay() {
  const { t } = useTranslation('pages')

  return (
    <ContentSection
      title={t('settings.display.title')}
      desc={t('settings.display.description')}
    >
      <DisplayForm />
    </ContentSection>
  )
}
