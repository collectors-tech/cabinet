import { useTranslation } from 'react-i18next'
import { Collection } from '@/features/collection'

export function Inventory() {
  const { t } = useTranslation('pages')

  return (
    <Collection
      title={t('inventory.title')}
      description={t('inventory.description')}
      routePath='/_authenticated/inventory/'
    />
  )
}
