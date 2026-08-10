import helpCenterReadme from '../../../../docs/help-center/README.md?raw'
import gettingStartedGuide from '../../../../docs/help-center/getting-started/login-and-database-setup.md?raw'
import betaDisclosureGuide from '../../../../docs/help-center/reference/cabinet-private-beta-disclosure.md?raw'
import chatsGuide from '../../../../docs/help-center/sections/chats.md?raw'
import collectionsGuide from '../../../../docs/help-center/sections/collections.md?raw'
import integrationsGuide from '../../../../docs/help-center/sections/integrations.md?raw'
import inventoryGuide from '../../../../docs/help-center/sections/inventory.md?raw'
import mediaGuide from '../../../../docs/help-center/sections/media.md?raw'
import settingsGuide from '../../../../docs/help-center/sections/settings.md?raw'
import wishlistGuide from '../../../../docs/help-center/sections/wishlist.md?raw'
import uiElementsGuide from '../../../../docs/help-center/ui-elements.md?raw'

export type HelpCenterArticle = {
  id: string
  title: string
  summary: string
  category: string
  content: string
}

export const helpCenterArticles: HelpCenterArticle[] = [
  {
    id: 'getting-started-login-database-setup',
    title: 'Login and Database Setup',
    summary:
      'Sign in, switch profiles, and confirm you are working in the right Cabinet data context.',
    category: 'Getting Started',
    content: gettingStartedGuide,
  },
  {
    id: 'cabinet-private-beta-disclosure',
    title: 'Cabinet 0.1 Private Beta Disclosure',
    summary:
      'Review supported beta capabilities, limitations, release gates, and recovery pointers.',
    category: 'Getting Started',
    content: betaDisclosureGuide,
  },
  {
    id: 'section-inventory',
    title: 'Inventory',
    summary:
      'Manage owned items, folder browsing, photos, barcodes, and AI assist workflows.',
    category: 'Sections',
    content: inventoryGuide,
  },
  {
    id: 'section-media',
    title: 'Media',
    summary:
      'Review uploaded evidence, unlinked assets, analysis state, and assignment follow-up.',
    category: 'Sections',
    content: mediaGuide,
  },
  {
    id: 'section-wishlist',
    title: 'Wishlist',
    summary:
      'Track wanted items, priorities, and follow-up collection workflows.',
    category: 'Sections',
    content: wishlistGuide,
  },
  {
    id: 'section-collections',
    title: 'Collections',
    summary:
      'Organize grouped item sets and manage collection-scoped browsing context.',
    category: 'Sections',
    content: collectionsGuide,
  },
  {
    id: 'section-integrations',
    title: 'Integrations',
    summary:
      'Connect providers, review diagnostics, and understand supported setup flows.',
    category: 'Sections',
    content: integrationsGuide,
  },
  {
    id: 'section-settings',
    title: 'Settings',
    summary:
      'Adjust profile, appearance, notifications, and other Cabinet configuration surfaces.',
    category: 'Sections',
    content: settingsGuide,
  },
  {
    id: 'section-chats',
    title: 'Chats',
    summary:
      'Understand assistant/chat surfaces and where AI-guided workflows fit into Cabinet.',
    category: 'Sections',
    content: chatsGuide,
  },
  {
    id: 'ui-elements',
    title: 'Generic UI Elements',
    summary:
      'Learn shared Cabinet patterns like New, Create, filters, rows, and toasts.',
    category: 'Reference',
    content: uiElementsGuide,
  },
  {
    id: 'help-center-docs-plan',
    title: 'Help Center Docs Plan',
    summary:
      'See how Help Center content is organized and how section guides are structured.',
    category: 'Reference',
    content: helpCenterReadme,
  },
]
