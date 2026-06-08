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
  { prefix: '/inbox', title: 'Purchases' },
  { prefix: '/settings', title: 'Settings' },
  { prefix: '/help-center', title: 'Help Center' },
  { prefix: '/users', title: 'Users' },
]

export function getDocumentTitle(pathname: string) {
  const normalized = pathname === '' ? '/' : pathname
  if (normalized === '/' || normalized === '/dashboard/') {
    return `${PRODUCT_NAME} - Home`
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
