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
    cy.clearLocalStorage('cabinet.toastHistory.v1')
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
    cy.window().then((win) => {
      win.localStorage.removeItem('cabinet.toastHistory.v1')
      win.dispatchEvent(new Event('cabinet:toast-history'))
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
    cy.clearLocalStorage('cabinet.toastHistory.v1')
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
      .type('Wishlist row')
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

  it('UI-SCREEN-NOTIFICATION-INBOX-008 + #1438 preserves promise toast feedback in Inbox history', () => {
    cy.viewport(1366, 768)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.intercept('GET', '/api/chat/inbox?profile_id=*', {
      statusCode: 200,
      body: { items: [] },
    }).as('loadNotifications')

    cy.visit('/forgot-password')
    cy.get('input[name="email"]').clear().type('e2e-toast-capture@example.com')
    cy.get('[data-testid="forgot-password-submit"]').click()
    cy.contains('Sending email...').should('be.visible')
    cy.contains('Email sent to e2e-toast-capture@example.com', {
      timeout: 5000,
    }).should('be.visible')

    cy.visit('/sign-in?redirect=%2Finbox')
    cy.get('input[name="email"]').clear().type('e2e-inbox-toast@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inbox\/?$/)
    cy.wait('@loadNotifications')

    cy.get('[data-testid="notification-inbox-search"]').type(
      'Email sent to e2e-toast-capture@example.com'
    )
    cy.get('[data-testid="notification-inbox-row"]')
      .should('have.length.at.least', 1)
      .first()
      .should('contain', 'Email sent to e2e-toast-capture@example.com')
      .and('contain', 'Toast History')
      .and('contain', 'system')
    cy.get('[data-testid="notification-inbox-row"]').first().click()
    cy.get('[data-testid="notification-inbox-detail-pane"]')
      .should('contain', 'Email sent to e2e-toast-capture@example.com')
      .and(
        'contain',
        'Promise toast settled and was preserved in Inbox history.'
      )
      .and('contain', 'Toast History')
  })

  it('UI-SCREEN-NOTIFICATION-INBOX-008 + #1438 preserves confirmation dialog warnings in Inbox history', () => {
    cy.viewport(1366, 768)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.intercept('GET', '/api/chat/inbox?profile_id=*', {
      statusCode: 200,
      body: { items: [] },
    }).as('loadNotifications')

    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/dashboard/',
    })
    cy.window().then((win) => {
      win.localStorage.removeItem('cabinet.toastHistory.v1')
      win.dispatchEvent(new Event('cabinet:toast-history'))
    })

    cy.get('[data-testid="profile-dropdown-trigger"]:visible').first().click()
    cy.contains('[data-slot="dropdown-menu-item"]', 'Sign out').click()
    cy.get('[role="alertdialog"]')
      .should('be.visible')
      .and('contain', 'Sign out')
      .and('contain', 'Are you sure you want to sign out?')
    cy.contains('button', 'Cancel').click()

    cy.visit('/inbox/')
    cy.wait('@loadNotifications')
    cy.get('[data-testid="notification-inbox-search"]').type('Sign out')
    cy.get('[data-testid="notification-inbox-row"]')
      .should('have.length', 1)
      .and('contain', 'Sign out')
      .and('contain', 'Dialog History')
      .and('contain', 'system')
    cy.get('[data-testid="notification-inbox-row"]').first().click()
    cy.get('[data-testid="notification-inbox-detail-pane"]')
      .should('contain', 'Sign out')
      .and('contain', 'Are you sure you want to sign out?')
      .and('contain', 'Dialog History')
  })

  it('UI-SCREEN-NOTIFICATION-INBOX-008 + #1438 preserves settings status banners in Inbox history', () => {
    cy.viewport(1366, 768)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.intercept('GET', '/api/runtime', {
      statusCode: 200,
      body: {
        app_version: 'rev-inline-status',
        build_date: '2026-06-25T02:00:00Z',
        bind_mode: 'loopback',
        runtime_host: '127.0.0.1',
        runtime_port: 17882,
        update_channel: 'stable',
        update_public_key_configured: true,
      },
    }).as('runtimeInfo')
    cy.intercept('GET', '/api/runtime/recovery', {
      statusCode: 200,
      body: {
        recovery_required: false,
      },
    }).as('runtimeRecovery')
    cy.intercept('GET', '/api/logs/export', {
      statusCode: 200,
      body: '2026-06-25T02:00:00Z runtime log line',
    }).as('exportLogs')
    cy.intercept('POST', '/api/chat/inbox', (req) => {
      const record = req.body.records.find(
        (candidate: Record<string, unknown>) =>
          candidate.local_history_id ===
          'settings-operations-logs-export-success'
      )
      expect(record).to.deep.include({
        level: 'success',
        title: 'Exported runtime logs successfully.',
        summary: 'Diagnostics logs status from Settings Operations.',
        source_label: 'Settings Operations',
        category: 'system',
      })
      expect(record.created_at).to.match(/\d{4}-\d{2}-\d{2}T/)
      req.reply({ statusCode: 201, body: { items: [] } })
    }).as('syncInlineStatusHistory')
    cy.intercept('GET', '/api/chat/inbox?profile_id=*', {
      statusCode: 200,
      body: { items: [] },
    }).as('loadNotifications')

    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/settings/operations/',
    })
    cy.wait('@runtimeInfo')
    cy.wait('@runtimeRecovery')
    cy.window().then((win) => {
      win.localStorage.removeItem('cabinet.toastHistory.v1')
      win.dispatchEvent(new Event('cabinet:toast-history'))
    })

    cy.get('[data-testid="settings-operations-export-logs"]').click()
    cy.wait('@exportLogs')
    cy.get('[data-testid="settings-operations-logs-status"]').should(
      'contain',
      'Exported runtime logs successfully.'
    )

    cy.visit('/inbox/')
    cy.wait('@syncInlineStatusHistory')
    cy.wait('@loadNotifications')
  })

  it('UI-SCREEN-NOTIFICATION-INBOX-008 + #1438 preserves settings notification save failures in Inbox history', () => {
    cy.viewport(1366, 768)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.intercept('GET', '/api/profiles/active', {
      statusCode: 200,
      body: { id: 'e2e-profile-001', name: 'E2E Local' },
    }).as('activeProfile')
    cy.intercept('GET', '/api/profiles/e2e-profile-001/settings', {
      statusCode: 200,
      body: {
        settings: {
          'notifications.type': 'mentions',
          'notifications.mobile': 'false',
          'notifications.communication_emails': 'false',
          'notifications.social_emails': 'true',
          'notifications.marketing_emails': 'false',
          'notifications.security_emails': 'true',
        },
      },
    }).as('notificationSettings')
    cy.intercept('PUT', '/api/profiles/e2e-profile-001/settings', {
      statusCode: 503,
      body: { error: 'notification_settings_save_unavailable' },
    }).as('saveNotificationsFailure')
    cy.intercept('POST', '/api/chat/inbox', (req) => {
      const record = req.body.records.find(
        (candidate: Record<string, unknown>) =>
          candidate.local_history_id ===
          'settings-notifications-save-failed'
      )
      expect(record).to.deep.include({
        level: 'error',
        title: 'profile_settings_save_503',
        summary: 'Settings Notifications save failure preserved for review.',
        source_label: 'Settings Notifications',
        category: 'settings',
      })
      expect(record.created_at).to.match(/\d{4}-\d{2}-\d{2}T/)
      req.reply({ statusCode: 201, body: { items: [] } })
    }).as('syncNotificationSettingsFailureHistory')
    cy.intercept('GET', '/api/chat/inbox?profile_id=*', {
      statusCode: 200,
      body: { items: [] },
    }).as('loadNotifications')

    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/settings/notifications/',
    })
    cy.wait('@activeProfile')
    cy.wait('@notificationSettings')
    cy.window().then((win) => {
      win.localStorage.removeItem('cabinet.toastHistory.v1')
      win.dispatchEvent(new Event('cabinet:toast-history'))
    })

    cy.contains('label', 'Nothing').click()
    cy.contains('button', 'Update notifications').click()
    cy.wait('@saveNotificationsFailure')
    cy.contains('profile_settings_save_503').should('be.visible')

    cy.visit('/inbox/')
    cy.wait('@syncNotificationSettingsFailureHistory')
    cy.wait('@loadNotifications')
  })

  it('UI-SCREEN-NOTIFICATION-INBOX-008 + #1438 preserves storage maintenance status banners in Inbox history', () => {
    cy.viewport(1366, 768)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.intercept('GET', '/api/profiles/active', {
      statusCode: 200,
      body: { id: 'e2e-profile-001', name: 'E2E Local' },
    }).as('activeProfile')
    cy.intercept('GET', '/api/profiles/*/storage', {
      statusCode: 200,
      body: {
        db_path: 'C:/cabinet/e2e/cabinet.db',
        media_dir: 'C:/cabinet/e2e/media',
      },
    }).as('storageInfo')
    cy.intercept('GET', '/api/backup/list', {
      statusCode: 200,
      body: { backups: [] },
    }).as('backupList')
    cy.intercept('POST', '/api/data/reindex', {
      statusCode: 200,
      body: { ok: true },
    }).as('reindexSearch')
    cy.intercept('POST', '/api/chat/inbox', (req) => {
      const record = req.body.records.find(
        (candidate: Record<string, unknown>) =>
          candidate.local_history_id === 'settings-storage-reindex-success'
      )
      expect(record).to.deep.include({
        level: 'success',
        title: 'Search reindex completed successfully.',
        summary: 'Maintenance status from Settings Storage.',
        source_label: 'Settings Storage',
        category: 'system',
      })
      expect(record.created_at).to.match(/\d{4}-\d{2}-\d{2}T/)
      req.reply({ statusCode: 201, body: { items: [] } })
    }).as('syncStorageStatusHistory')
    cy.intercept('GET', '/api/chat/inbox?profile_id=*', {
      statusCode: 200,
      body: { items: [] },
    }).as('loadNotifications')

    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/settings/storage/',
    })
    cy.wait('@activeProfile')
    cy.wait('@storageInfo')
    cy.wait('@backupList')
    cy.window().then((win) => {
      win.localStorage.removeItem('cabinet.toastHistory.v1')
      win.dispatchEvent(new Event('cabinet:toast-history'))
    })

    cy.contains('button', 'Reindex Search').click()
    cy.wait('@reindexSearch')
    cy.get('[data-testid="settings-storage-action-status"]').should(
      'contain',
      'Search reindex completed successfully.'
    )

    cy.visit('/inbox/')
    cy.wait('@syncStorageStatusHistory')
    cy.wait('@loadNotifications')
  })

  it('UI-SCREEN-NOTIFICATION-INBOX-008 + #1438 preserves settings taxonomy status banners in Inbox history', () => {
    cy.viewport(1366, 768)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.intercept('GET', '/api/profiles/active', {
      statusCode: 200,
      body: { id: 'e2e-profile-001', name: 'E2E Local' },
    }).as('activeProfile')
    cy.intercept('GET', '/api/profiles/e2e-profile-001/settings', {
      statusCode: 200,
      body: {
        settings: {
          inventory_category_options: 'Cards',
          inventory_packaging_grades: 'Sealed',
          inventory_item_type_condition_scales:
            '[{"item_type":"Card","conditions":["Near Mint","Played"]}]',
        },
      },
    }).as('categorySettings')
    cy.intercept('PUT', '/api/profiles/e2e-profile-001/settings', (req) => {
      req.reply({
        statusCode: 200,
        body: { settings: req.body.settings },
      })
    }).as('saveCategorySettings')
    cy.intercept('POST', '/api/chat/inbox', (req) => {
      const record = req.body.records.find(
        (candidate: Record<string, unknown>) =>
          candidate.local_history_id ===
          'settings-categories-taxonomy-save-success'
      )
      expect(record).to.deep.include({
        level: 'success',
        title: 'Saved categories, packaging grades, and item type condition scales.',
        summary: 'Taxonomy settings status from Settings Categories.',
        source_label: 'Settings Categories',
        category: 'system',
      })
      expect(record.created_at).to.match(/\d{4}-\d{2}-\d{2}T/)
      req.reply({ statusCode: 201, body: { items: [] } })
    }).as('syncCategoriesStatusHistory')
    cy.intercept('GET', '/api/chat/inbox?profile_id=*', {
      statusCode: 200,
      body: { items: [] },
    }).as('loadNotifications')

    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/settings/categories/',
    })
    cy.wait('@activeProfile')
    cy.wait('@categorySettings')
    cy.window().then((win) => {
      win.localStorage.removeItem('cabinet.toastHistory.v1')
      win.dispatchEvent(new Event('cabinet:toast-history'))
    })

    cy.get('[data-testid="settings-categories-new"]').type('Garage Kit')
    cy.get('[data-testid="settings-categories-add"]').click()
    cy.get('[data-testid="settings-categories-save"]').click()
    cy.wait('@saveCategorySettings')
    cy.get('[data-testid="settings-categories-status"]').should(
      'contain',
      'Saved categories, packaging grades, and item type condition scales.'
    )

    cy.visit('/inbox/')
    cy.wait('@syncCategoriesStatusHistory')
    cy.wait('@loadNotifications')
  })

  it('UI-SCREEN-NOTIFICATION-INBOX-008 + #1438 preserves Collections workspace status toasts in Inbox history', () => {
    cy.viewport(1366, 768)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.intercept('GET', '/api/items', {
      statusCode: 200,
      body: { items: [] },
    }).as('collectionsInventoryItems')
    cy.intercept('GET', '/api/profiles/*/settings').as(
      'loadCollectionSettings'
    )
    cy.intercept('PUT', '/api/profiles/e2e-profile-001/settings', (req) => {
      req.reply({
        statusCode: 200,
        body: { settings: req.body.settings },
      })
    }).as('saveCollectionSettings')
    cy.intercept('POST', '/api/chat/inbox', (req) => {
      const record = req.body.records.find(
        (candidate: Record<string, unknown>) =>
          candidate.local_history_id === 'collections-rename-success'
      )
      expect(record).to.deep.include({
        level: 'success',
        title: 'Watch List renamed to Watch List Inbox.',
        summary: 'Collections workspace status from Collections.',
        source_label: 'Collections',
        category: 'notification',
      })
      expect(record.created_at).to.match(/\d{4}-\d{2}-\d{2}T/)
      req.reply({ statusCode: 201, body: { items: [] } })
    }).as('syncCollectionsStatusHistory')
    cy.intercept('GET', '/api/chat/inbox?profile_id=*', {
      statusCode: 200,
      body: { items: [] },
    }).as('loadNotifications')

    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/collections/',
    })
    cy.wait('@loadCollectionSettings')
    cy.wait('@collectionsInventoryItems')
    cy.window().then((win) => {
      win.localStorage.removeItem('cabinet.toastHistory.v1')
      win.dispatchEvent(new Event('cabinet:toast-history'))
    })

    cy.get('[data-testid="collections-row-edit-watch-list"]')
      .scrollIntoView()
      .click({ force: true })
    cy.get('[data-testid="collections-edit-input"]')
      .clear()
      .type('Watch List Inbox')
    cy.get('[data-testid="collections-edit-submit"]').click()
    cy.wait('@saveCollectionSettings')
    cy.contains('Watch List renamed to Watch List Inbox.').should('be.visible')

    cy.visit('/inbox/')
    cy.wait('@syncCollectionsStatusHistory')
    cy.wait('@loadNotifications')
  })

  it('UI-SCREEN-NOTIFICATION-INBOX-008 + #1438 preserves Integrations provider health status in Inbox history', () => {
    cy.viewport(1366, 768)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.intercept('GET', '/api/profiles/active', {
      statusCode: 200,
      body: { id: 'e2e-profile-001', name: 'E2E Local' },
    }).as('activeProfile')
    cy.intercept('GET', '/api/providers/registry', {
      statusCode: 200,
      body: {
        providers: [
          {
            provider_id: 'ebay',
            display_name: 'eBay',
            base_domain: 'ebay.com',
            integration_mode: 'official_api',
            auth_mode: 'api_key',
            state: 'ready',
            has_token: true,
            setup_instructions: 'Configure eBay token and marketplace.',
            capabilities: {
              search: true,
              stock_observation: false,
              pricing: true,
              health: true,
            },
            health: { status: 'unknown', last_checked_at: null },
            last_run: { status: 'never', finished_at: null },
          },
        ],
      },
    }).as('registry')
    cy.intercept('GET', '/api/profiles/*/settings', {
      statusCode: 200,
      body: { settings: { 'integration.ebay.enabled': 'true' } },
    }).as('settings')
    cy.intercept('GET', '/api/provider/health?provider=ebay', {
      statusCode: 200,
      body: {
        provider: 'ebay',
        status: 'ok',
        state: 'ready',
        message: 'eBay credentials are ready.',
        updated_at: '2026-06-25T08:00:00Z',
      },
    }).as('providerHealth')
    cy.intercept('POST', '/api/chat/inbox', (req) => {
      const record = req.body.records.find(
        (candidate: Record<string, unknown>) =>
          candidate.local_history_id ===
          'integrations-provider-health-ebay-ok'
      )
      expect(record).to.deep.include({
        level: 'success',
        title: 'Validated eBay health: ok.',
        summary: 'Provider health validation status from Integrations.',
        source_label: 'Integrations',
        category: 'system',
      })
      expect(record.created_at).to.match(/\d{4}-\d{2}-\d{2}T/)
      req.reply({ statusCode: 201, body: { items: [] } })
    }).as('syncIntegrationsHealthHistory')
    cy.intercept('GET', '/api/chat/inbox?profile_id=*', {
      statusCode: 200,
      body: { items: [] },
    }).as('loadNotifications')

    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/integrations/',
    })
    cy.wait('@activeProfile')
    cy.wait('@registry')
    cy.wait('@settings')
    cy.window().then((win) => {
      win.localStorage.removeItem('cabinet.toastHistory.v1')
      win.dispatchEvent(new Event('cabinet:toast-history'))
    })

    cy.get('[data-testid="provider-open-ebay"]').click()
    cy.contains('button', 'Validate').click()
    cy.wait('@providerHealth')
    cy.contains('Validated eBay health: ok.').scrollIntoView().should('be.visible')

    cy.visit('/inbox/')
    cy.wait('@syncIntegrationsHealthHistory')
    cy.wait('@loadNotifications')
  })

  it('UI-SCREEN-NOTIFICATION-INBOX-008 + #1438 syncs local history into durable server Inbox without duplicates', () => {
    cy.viewport(1366, 768)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.intercept('POST', '/api/chat/inbox', (req) => {
      expect(String(req.body.profile_id)).to.match(/\S/)
      expect(req.body.records).to.deep.include({
        local_history_id: 'local-sync-proof-1',
        level: 'warning',
        title: 'Settings save warning',
        summary: 'Banner warning preserved for review.',
        source_label: 'Settings Banner',
        category: 'system',
        created_at: '2026-06-22T10:00:00Z',
      })
      req.reply({
        statusCode: 201,
        body: {
          items: [
            {
              id: 'server-history-1',
              status: 'read',
              source: 'notification_history',
              title: 'Settings save warning',
              summary: 'Banner warning preserved for review.',
              created_at: '2026-06-22T10:00:00Z',
              updated_at: '2026-06-22T10:00:00Z',
              metadata: {
                category: 'system',
                detail: 'Banner warning preserved for review.',
                local_history_id: 'local-sync-proof-1',
                source_label: 'Settings Banner',
              },
            },
          ],
        },
      })
    }).as('syncLocalHistory')
    cy.intercept('GET', '/api/chat/inbox?profile_id=*', {
      statusCode: 200,
      body: {
        items: [
          {
            id: 'server-history-1',
            status: 'read',
            source: 'notification_history',
            title: 'Settings save warning',
            summary: 'Banner warning preserved for review.',
            created_at: '2026-06-22T10:00:00Z',
            updated_at: '2026-06-22T10:00:00Z',
            metadata: {
              category: 'system',
              detail: 'Banner warning preserved for review.',
              local_history_id: 'local-sync-proof-1',
              source_label: 'Settings Banner',
            },
          },
        ],
      },
    }).as('loadNotifications')

    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/inbox/',
    })
    cy.window().then((win) => {
      win.localStorage.setItem(
        'cabinet.toastHistory.v1',
        JSON.stringify([
          {
            id: 'local-sync-proof-1',
            level: 'warning',
            title: 'Settings save warning',
            summary: 'Banner warning preserved for review.',
            source_label: 'Settings Banner',
            category: 'system',
            created_at: '2026-06-22T10:00:00Z',
          },
        ])
      )
      win.dispatchEvent(new Event('cabinet:toast-history'))
    })
    cy.reload()
    cy.wait('@syncLocalHistory')
    cy.wait('@loadNotifications')

    cy.get('[data-testid="notification-inbox-search"]').type(
      'Settings save warning'
    )
    cy.get('[data-testid="notification-inbox-row"]')
      .should('have.length', 1)
      .and('contain', 'Settings save warning')
      .and('contain', 'Settings Banner')
    cy.get('[data-testid="notification-inbox-row"]').first().click()
    cy.get('[data-testid="notification-inbox-detail-pane"]')
      .should('contain', 'Settings save warning')
      .and('contain', 'Banner warning preserved for review.')
      .and('contain', 'Settings Banner')
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
    cy.window().then((win) => {
      win.localStorage.removeItem('cabinet.toastHistory.v1')
      win.dispatchEvent(new Event('cabinet:toast-history'))
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
