import { getRouteMetadata } from './route-metadata'

const PRODUCT_NAME = 'Cabinet'

type TitleRule = {
  prefix: string
  title: string
}

const TITLE_RULES: TitleRule[] = [
  { prefix: '/dashboard', title: 'Home' },
  { prefix: '/inventory', title: 'Inventory' },
  { prefix: '/media', title: 'Media' },
  { prefix: '/collections', title: 'Collections' },
  { prefix: '/wishlist', title: 'Wishlist' },
  { prefix: '/discoveries', title: 'Discoveries' },
  { prefix: '/scanner', title: 'Scanner' },
  { prefix: '/pricing', title: 'Pricing' },
  { prefix: '/reports', title: 'Reports' },
  { prefix: '/integrations', title: 'Integrations' },
  { prefix: '/chats', title: 'Chats' },
  { prefix: '/inbox', title: 'Notification Inbox' },
  { prefix: '/settings/skills', title: 'Skills Settings' },
  { prefix: '/settings', title: 'Settings' },
  { prefix: '/help-center', title: 'Help Center' },
  { prefix: '/users', title: 'Users' },
  { prefix: '/errors', title: 'Error' },
  { prefix: '/404', title: 'Not Found' },
]

export function getDocumentTitle(pathname: string) {
  const normalized = pathname === '' ? '/' : pathname
  const routeMetadata = getRouteMetadata(normalized)

  if (routeMetadata) {
    return routeMetadata.documentTitle
  }

  const matched = TITLE_RULES.find(
    (rule) =>
      normalized === rule.prefix || normalized.startsWith(`${rule.prefix}/`)
  )

  if (!matched) {
    return PRODUCT_NAME
  }

  return `${PRODUCT_NAME} - ${matched.title}`
}
