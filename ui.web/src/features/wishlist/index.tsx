import { useTranslation } from 'react-i18next'
import { Tasks } from '@/features/tasks'

export function Wishlist() {
  const { t } = useTranslation('pages')

  return (
    <Tasks
      title={t('wishlist.title')}
      description={t('wishlist.description')}
      routePath='/_authenticated/wishlist/'
    />
  )
}
