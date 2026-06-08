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

describe('chats/notification-inbox', () => {
  function bootInbox(items: StubItem[] = baseItems) {
    cy.viewport(1400, 900)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.intercept('GET', '/api/chat/inbox?profile_id=*', {
      statusCode: 200,
      body: { items },
    }).as('loadNotifications')
    cy.intercept('PATCH', '/api/chat/inbox/*', (req) => {
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
})
