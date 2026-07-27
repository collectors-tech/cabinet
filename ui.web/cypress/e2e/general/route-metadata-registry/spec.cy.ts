import {
  authenticatedRouteMetadata,
  getRouteMetadata,
} from '../../../../src/lib/route-metadata'
import { getDocumentTitle } from '../../../../src/lib/document-title'
import {
  buildSearchNavigationResults,
  getRouteHeaderTitleProps,
} from '../../../../src/lib/route-navigation'

const expectedRoutes = [
  { path: '/', title: 'Home', documentTitle: 'Cabinet - Home' },
  { path: '/dashboard', title: 'Home', documentTitle: 'Cabinet - Home' },
  {
    path: '/inventory',
    title: 'Inventory',
    documentTitle: 'Cabinet - Inventory',
  },
  { path: '/media', title: 'Media', documentTitle: 'Cabinet - Media' },
  {
    path: '/collections',
    title: 'Collections',
    documentTitle: 'Cabinet - Collections',
  },
  { path: '/wishlist', title: 'Wishlist', documentTitle: 'Cabinet - Wishlist' },
  {
    path: '/discoveries',
    title: 'Discoveries',
    documentTitle: 'Cabinet - Discoveries',
  },
  {
    path: '/scanner',
    title: 'Market Watch',
    documentTitle: 'Cabinet - Market Watch',
  },
  {
    path: '/inbox',
    title: 'Notification Inbox',
    documentTitle: 'Cabinet - Notification Inbox',
  },
  {
    path: '/purchases',
    title: 'Purchases',
    documentTitle: 'Cabinet - Purchases',
  },
  {
    path: '/integrations',
    title: 'Integrations',
    documentTitle: 'Cabinet - Integrations',
  },
  { path: '/chats', title: 'Chats', documentTitle: 'Cabinet - Chats' },
  { path: '/users', title: 'Users', documentTitle: 'Cabinet - Users' },
  { path: '/reports', title: 'Reports', documentTitle: 'Cabinet - Reports' },
  {
    path: '/settings',
    title: 'Profile Settings',
    documentTitle: 'Cabinet - Profile Settings',
  },
  {
    path: '/settings/profile',
    title: 'Profile Settings',
    documentTitle: 'Cabinet - Profile Settings',
  },
  {
    path: '/settings/account',
    title: 'Account Settings',
    documentTitle: 'Cabinet - Account Settings',
  },
  {
    path: '/settings/appearance',
    title: 'Appearance Settings',
    documentTitle: 'Cabinet - Appearance Settings',
  },
  {
    path: '/settings/billing',
    title: 'Billing Settings',
    documentTitle: 'Cabinet - Billing Settings',
  },
  {
    path: '/settings/categories',
    title: 'Category Settings',
    documentTitle: 'Cabinet - Category Settings',
  },
  {
    path: '/settings/display',
    title: 'Display Settings',
    documentTitle: 'Cabinet - Display Settings',
  },
  {
    path: '/settings/integrations',
    title: 'Integration Settings',
    documentTitle: 'Cabinet - Integration Settings',
  },
  {
    path: '/settings/notifications',
    title: 'Notification Settings',
    documentTitle: 'Cabinet - Notification Settings',
  },
  {
    path: '/settings/operations',
    title: 'Operations Settings',
    documentTitle: 'Cabinet - Operations Settings',
  },
  {
    path: '/settings/skills',
    title: 'Skills Settings',
    documentTitle: 'Cabinet - Skills Settings',
  },
  {
    path: '/settings/storage',
    title: 'Storage Settings',
    documentTitle: 'Cabinet - Storage Settings',
  },
  {
    path: '/help-center',
    title: 'Help Center',
    documentTitle: 'Cabinet - Help Center',
  },
  { path: '/errors/not-found', title: 'Error', documentTitle: 'Cabinet - Error' },
  { path: '/404', title: 'Not Found', documentTitle: 'Cabinet - Not Found' },
] as const

describe('route metadata registry', () => {
  it('UI-ROUTE-METADATA-001 covers every authenticated route with canonical metadata', () => {
    const canonicalPaths = new Set(
      authenticatedRouteMetadata.map((metadata) => metadata.path)
    )

    expect(canonicalPaths.size, 'duplicate canonical paths').to.eq(
      authenticatedRouteMetadata.length
    )

    expectedRoutes.forEach(({ path, title, documentTitle }) => {
      const metadata = getRouteMetadata(path)

      expect(metadata, path).to.include({
        title,
        documentTitle,
      })
      expect(metadata?.description, `${path} description`).to.be.a('string').and
        .not.be.empty
      expect(metadata?.icon, `${path} icon`).to.exist
      expect(metadata?.navigationGroup, `${path} navigation group`).to.be.oneOf([
        'General',
        'Settings',
        'Other',
        'System',
      ])
      expect(metadata?.testIds.headerTitle, `${path} header title id`).to.match(
        /^[a-z0-9-]+-header-title$/
      )
      expect(metadata?.testIds.headerIcon, `${path} header icon id`).to.match(
        /^[a-z0-9-]+-header-icon$/
      )
    })
  })

  it('UI-ROUTE-METADATA-002 resolves document titles from canonical route metadata', () => {
    expectedRoutes.forEach(({ path, documentTitle }) => {
      expect(getDocumentTitle(path), `${path} document title`).to.eq(
        documentTitle
      )
    })
  })

  it('UI-ROUTE-METADATA-003 builds search navigation from canonical route metadata', () => {
    const navigationResults = buildSearchNavigationResults()
    const byPath = new Map(
      navigationResults.map((result) => [result.path, result])
    )

    expect(byPath.get('/scanner')).to.include({
      title: 'Market Watch',
      group: 'General',
      path: '/scanner',
    })
    expect(byPath.get('/purchases')).to.include({
      title: 'Purchases',
      group: 'General',
      path: '/purchases',
    })
    expect(byPath.get('/settings/profile')).to.include({
      title: 'Profile Settings',
      group: 'Settings',
      path: '/settings/profile',
    })
    expect(byPath.has('/errors/*'), 'system error route excluded').to.eq(false)
    expect(byPath.has('/404'), 'system not-found route excluded').to.eq(false)
  })

  it('UI-ROUTE-METADATA-004 exposes HeaderTitle props from canonical route metadata', () => {
    const marketWatchHeader = getRouteHeaderTitleProps('/scanner')

    expect(marketWatchHeader).to.include({
      title: 'Market Watch',
      description: 'Run saved market searches and review provider results.',
    })
    expect(marketWatchHeader?.icon).to.equal(getRouteMetadata('/scanner')?.icon)
  })
})
