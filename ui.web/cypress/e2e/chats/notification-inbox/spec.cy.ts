type StubItem = {
  id: string
  status: string
  source: string
  title: string
  summary: string
  created_at: string
  updated_at: string
  metadata?: Record<string, unknown>
}

const baseItems: StubItem[] = [
  {
    id: 'notice-assistant-price',
    status: 'unread',
    source: 'assistant_handoff',
    title: 'Assistant price review is ready',
    summary: 'Three watched cards crossed the review threshold.',
    created_at: '2026-06-08T08:00:00Z',
    updated_at: '2026-06-08T08:00:00Z',
    metadata: {
      category: 'assistant',
      source_label: 'Assistant Workflow',
      detail: 'Assistant prepared a review packet with three price movements.',
      item: {
        id: 'card-001',
        title: 'Showcase Seed One',
        part_number: 'SHW-100',
        href: '/inventory/?item=card-001',
      },
    },
  },
  {
    id: 'notice-system-runtime',
    status: 'read',
    source: 'runtime_system',
    title: 'Demo runtime recycled',
    summary: 'demo2 restarted from develop after merge validation.',
    created_at: '2026-06-08T07:30:00Z',
    updated_at: '2026-06-08T07:30:00Z',
    metadata: {
      category: 'system',
      source_label: 'Runtime',
      detail: 'Runtime process completed health and /api/runtime verification.',
    },
  },
  {
    id: 'notice-import-ready',
    status: 'unread',
    source: 'purchase_import',
    title: 'Forwarder import needs review',
    summary: 'Two packages need purchase matching decisions.',
    created_at: '2026-06-08T06:00:00Z',
    updated_at: '2026-06-08T06:00:00Z',
    metadata: {
      category: 'notification',
      source_label: 'Purchases',
      detail: 'Forwarder package rows are ready for review.',
    },
  },
]

const denseInboxItems: StubItem[] = [
  ...baseItems,
  {
    id: 'notice-mention-review',
    status: 'unread',
    source: 'assistant_mention',
    title: 'Mention review requested',
    summary: 'A workspace mention needs follow-up.',
    created_at: '2026-06-08T05:00:00Z',
    updated_at: '2026-06-08T05:00:00Z',
    metadata: {
      category: 'mention',
      source_label: 'Mentions',
      detail: 'A teammate mentioned this workspace review.',
    },
  },
  {
    id: 'notice-system-warning',
    status: 'read',
    source: 'runtime_system',
    title: 'Background sync warning',
    summary: 'A background process reported a retryable warning.',
    created_at: '2026-06-08T04:00:00Z',
    updated_at: '2026-06-08T04:00:00Z',
    metadata: {
      category: 'system',
      source_label: 'Runtime',
      detail: 'The sync retried and completed on the second attempt.',
    },
  },
  {
    id: 'notice-import-failed',
    status: 'unread',
    source: 'purchase_import',
    title: 'Import failed validation',
    summary: 'One row needs correction before it can be imported.',
    created_at: '2026-06-08T03:00:00Z',
    updated_at: '2026-06-08T03:00:00Z',
    metadata: {
      category: 'notification',
      source_label: 'Imports',
      detail: 'The CSV row is missing a required identifier.',
    },
  },
  {
    id: 'notice-hidden-deploy',
    status: 'archived',
    source: 'runtime_system',
    title: 'Hidden deploy note',
    summary: 'Previously cleared runtime note.',
    created_at: '2026-06-08T02:00:00Z',
    updated_at: '2026-06-08T02:00:00Z',
    metadata: {
      category: 'system',
      source_label: 'Runtime',
      detail: 'This record stays recoverable after clear/archive.',
    },
  },
]

describe('chats/notification-inbox', () => {
  function bootInbox(
    items: StubItem[] = baseItems,
    options: { failUpdates?: boolean } = {}
  ) {
    cy.viewport(1400, 900)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.intercept('GET', '/api/chat/inbox?profile_id=*', {
      statusCode: 200,
      body: { items },
    }).as('loadNotifications')
    cy.intercept('PATCH', '/api/chat/inbox/*', (req) => {
      if (options.failUpdates) {
        req.reply({
          statusCode: 500,
          body: { error: 'failed' },
        })
        return
      }
      const id = String(req.url).split('/api/chat/inbox/')[1]
      const item = items.find((candidate) => candidate.id === id)
      req.reply({
        statusCode: 200,
        body: {
          ...(item ?? items[0]),
          id,
          status: req.body.status,
          updated_at: '2026-06-08T09:00:00Z',
        },
      })
    }).as('updateNotification')
    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/inbox/',
    })
    cy.wait('@loadNotifications')
  }

  it('UI-SCREEN-NOTIFICATION-INBOX-001 renders first-class notification inbox at /inbox', () => {
    bootInbox()

    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inbox\/?$/)
    cy.title().should('eq', 'Cabinet - Notification Inbox')
    cy.get('[data-testid="notification-inbox-page"]').should('be.visible')
    cy.get('[data-testid="notification-inbox-header-title"]').should(
      'contain',
      'Notification Inbox'
    )
    cy.get('[data-testid="purchase-inbox-load-reviews"]').should('not.exist')
  })

  it('UI-SCREEN-NOTIFICATION-INBOX-007 + #1426 exposes Inbox as primary nav with bell navigation, search, split detail, and toast history', () => {
    cy.viewport(1366, 768)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.intercept('GET', '/api/chat/inbox?profile_id=*', {
      statusCode: 200,
      body: { items: baseItems },
    }).as('loadNotifications')

    cy.visit('/sign-in?redirect=%2Fdashboard')
    cy.get('input[name="email"]').clear().type('e2e-inbox-nav@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.window().then((win) => {
      win.localStorage.setItem(
        'cabinet.toastHistory.v1',
        JSON.stringify([
          {
            id: 'toast-history-1',
            level: 'success',
            title: 'Wishlist row updated.',
            summary: 'Toast feedback persisted for later review.',
            created_at: '2026-06-08T09:30:00Z',
          },
        ])
      )
    })
    cy.contains('button', 'Sign in').click()
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/dashboard\/?$/)

    cy.get('[data-testid="sidebar-nav-link-inbox"]')
      .should('be.visible')
      .and('contain', 'Inbox')
      .click()
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inbox\/?$/)
    cy.wait('@loadNotifications')
    cy.get('[data-testid="notification-inbox-filter-all"]').should(
      'have.attr',
      'data-state',
      'active'
    )

    cy.visit('/dashboard')
    cy.get('[data-testid="shell-workspace-bell"]').click()
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inbox\/?$/)
    cy.get('[data-testid="notification-inbox-page"]').should(
      'have.attr',
      'data-layout',
      'dense-two-pane'
    )
    cy.get('[data-testid="notification-inbox-list-pane"]').should('be.visible')
    cy.get('[data-testid="notification-inbox-detail-pane"]').should('be.visible')
    cy.get('[data-testid="notification-inbox-search"]')
      .should('be.visible')
      .type('toast')
    cy.get('[data-testid="notification-inbox-row"]')
      .should('have.length', 1)
      .and('contain', 'Wishlist row updated.')
    cy.get('[data-testid="notification-inbox-detail-pane"]')
      .should('contain', 'Wishlist row updated.')
      .and('contain', 'Toast feedback persisted for later review.')
    cy.get('[data-testid="notification-inbox-list-pane"]').then(($pane) => {
      const pane = $pane[0]
      expect(pane.clientHeight, 'list pane is height constrained').to.be.lessThan(
        620
      )
      expect(
        pane.scrollHeight,
        'list pane owns vertical scrolling'
      ).to.be.greaterThan(pane.clientHeight)
    })
  })

  it('UI-SCREEN-NOTIFICATION-INBOX-002 filters categories and shows contextual empty states', () => {
    bootInbox()

    cy.get('[data-testid="notification-inbox-row"]').should('have.length', 3)
    cy.get('[data-testid="notification-inbox-filter-unread"]').click()
    cy.get('[data-testid="notification-inbox-row"]').should('have.length', 2)
    cy.get('[data-testid="notification-inbox-filter-assistant"]').click()
    cy.get('[data-testid="notification-inbox-row"]')
      .should('have.length', 1)
      .and('contain', 'Assistant price review is ready')
    cy.get('[data-testid="notification-inbox-filter-system"]').click()
    cy.get('[data-testid="notification-inbox-row"]')
      .should('have.length', 1)
      .and('contain', 'Demo runtime recycled')

    cy.intercept('GET', '/api/chat/inbox?profile_id=*', {
      statusCode: 200,
      body: { items: [] },
    }).as('loadEmptyNotifications')
    cy.get('[data-testid="notification-inbox-refresh"]').click()
    cy.wait('@loadEmptyNotifications')
    cy.get('[data-testid="notification-inbox-empty-state"]')
      .should('be.visible')
      .and('contain', 'No system or runtime notices are waiting.')
  })

  it('UI-SCREEN-NOTIFICATION-INBOX-007 + #1438 keeps dense Inbox actions icon-only and clearing recoverable', () => {
    bootInbox(denseInboxItems)

    cy.get('[data-testid="notification-inbox-page"]').should(
      'have.attr',
      'data-layout',
      'dense-two-pane'
    )
    cy.get('[data-testid="notification-inbox-filter-all"]')
      .should('contain', 'All')
      .and('contain', '6')
    cy.get('[data-testid="notification-inbox-filter-unread"]').should(
      'contain',
      '4'
    )
    ;[
      'notification-inbox-refresh',
      'notification-inbox-bulk-read',
      'notification-inbox-clear-visible',
      'notification-inbox-show-hidden',
    ].forEach((testId) => {
      cy.get(`[data-testid="${testId}"]`)
        .should('have.attr', 'aria-label')
        .and('not.be.empty')
      cy.get(`[data-testid="${testId}"]`).invoke('text').should('match', /^\s*$/)
    })

    cy.get('[data-testid="notification-inbox-row"]').should('have.length', 5)
    cy.get('[data-testid="notification-inbox-total-count"]').should(
      'contain',
      '6 total messages'
    )
    cy.get('[data-testid="notification-inbox-pagination"]').should(
      'contain',
      'Page 1 of 2'
    )
    cy.get('[data-testid="notification-inbox-next-page"]').click()
    cy.get('[data-testid="notification-inbox-row"]')
      .should('have.length', 1)
      .and('contain', 'Background sync warning')
    cy.get('[data-testid="notification-inbox-prev-page"]').click()

    cy.contains(
      '[data-testid="notification-inbox-row"]',
      'Assistant price review is ready'
    ).click()
    cy.get('[data-testid="notification-inbox-detail-empty"]').should('not.exist')
    cy.get('[data-testid="notification-inbox-detail-pane"]')
      .should('contain', 'Assistant price review is ready')
      .and('contain', 'Mark handled')
      .and('contain', 'Delete')

    cy.get('[data-testid="notification-inbox-clear-visible"]').click()
    cy.wait([
      '@updateNotification',
      '@updateNotification',
      '@updateNotification',
      '@updateNotification',
      '@updateNotification',
      '@updateNotification',
    ])
    cy.get('[data-testid="notification-inbox-empty-state"]').should(
      'be.visible'
    )
    cy.get('[data-testid="notification-inbox-total-count"]').should(
      'contain',
      '0 total messages'
    )
    cy.get('[data-testid="notification-inbox-show-hidden"]').click()
    cy.get('[data-testid="notification-inbox-total-count"]').should(
      'contain',
      '7 total messages'
    )
    cy.get('[data-testid="notification-inbox-row"]').should('have.length', 5)
    cy.get('[data-testid="notification-inbox-search"]')
      .clear()
      .type('Hidden deploy note')
    cy.get('[data-testid="notification-inbox-list"]').should(
      'contain',
      'Hidden deploy note'
    )
  })

  it('UI-SCREEN-NOTIFICATION-INBOX-003 expands detail and preserves linked target navigation', () => {
    bootInbox()

    cy.contains(
      '[data-testid="notification-inbox-row"]',
      'Assistant price review is ready'
    ).within(() => {
      cy.get('[data-testid="notification-inbox-row-source"]').should(
        'contain',
        'Assistant Workflow'
      )
      cy.get('[data-testid="notification-inbox-row-link"]')
        .should('contain', 'SHW-100 - Showcase Seed One')
        .and('have.attr', 'href', '/inventory/?item=card-001')
      cy.get('[data-testid="notification-inbox-row-expand"]').click()
      cy.get('[data-testid="notification-inbox-row-detail"]')
        .should('be.visible')
        .and(
          'contain',
          'Assistant prepared a review packet with three price movements.'
        )
    })
  })

  it('UI-SCREEN-NOTIFICATION-INBOX-004 supports row and bulk triage actions', () => {
    bootInbox()

    cy.contains(
      '[data-testid="notification-inbox-row"]',
      'Assistant price review is ready'
    )
      .find('[data-testid="notification-inbox-row-read"]')
      .click()
    cy.wait('@updateNotification').its('request.body').should('deep.include', {
      profile_id: 'e2e-profile-001',
      status: 'read',
    })
    cy.contains(
      '[data-testid="notification-inbox-row"]',
      'Assistant price review is ready'
    )
      .find('[data-testid="notification-inbox-row-status"]')
      .should('contain', 'read')

    cy.get('[data-testid="notification-inbox-select-all"]').click()
    cy.get('[data-testid="notification-inbox-bulk-read"]').click()
    cy.wait('@updateNotification')
    cy.get('[data-testid="notification-inbox-select-all"]').click()
    cy.get('[data-testid="notification-inbox-bulk-archive"]').click()
    cy.wait('@updateNotification')
    cy.get('[data-testid="notification-inbox-row"]').should('have.length.lessThan', 3)
  })

  it('UI-SCREEN-NOTIFICATION-INBOX-005 shows retryable API error state', () => {
    cy.viewport(1400, 900)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.intercept('GET', '/api/chat/inbox?profile_id=*', {
      statusCode: 500,
      body: { error: 'failed' },
    }).as('loadNotificationsFailure')
    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/inbox/',
    })
    cy.wait('@loadNotificationsFailure')
    cy.get('[data-testid="notification-inbox-error-state"]')
      .should('be.visible')
      .and('contain', 'notification_inbox_load_failed')
    cy.wait(500)
    cy.get('[data-testid="notification-inbox-retry"]').should('be.visible')

    cy.intercept('GET', '/api/chat/inbox?profile_id=*', {
      statusCode: 200,
      body: { items: [] },
    }).as('retryNotifications')
    cy.get('[data-testid="notification-inbox-retry"]').click({ force: true })
    cy.wait('@retryNotifications')
    cy.get('[data-testid="notification-inbox-empty-state"]').should('be.visible')
  })

  it('UI-SCREEN-NOTIFICATION-INBOX-006 keeps rows retryable when triage updates fail', () => {
    bootInbox(baseItems, { failUpdates: true })

    cy.contains(
      '[data-testid="notification-inbox-row"]',
      'Assistant price review is ready'
    )
      .as('assistantRow')
      .find('[data-testid="notification-inbox-row-read"]')
      .click()
    cy.wait('@updateNotification').its('request.body').should('deep.include', {
      profile_id: 'e2e-profile-001',
      status: 'read',
    })
    cy.get('[data-testid="notification-inbox-error-state"]')
      .should('be.visible')
      .and('contain', 'notification_inbox_update_failed')
    cy.get('@assistantRow')
      .should('be.visible')
      .find('[data-testid="notification-inbox-row-status"]')
      .should('contain', 'unread')
    cy.get('@assistantRow')
      .find('[data-testid="notification-inbox-row-read"]')
      .should('be.enabled')
  })
})
