describe('ui-screen-settings', () => {
  function signInToSettings() {
    cy.visit('/sign-in?redirect=%2Fsettings%2F')
    cy.get('input[name="email"]').clear().type('e2e-settings@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/settings\/profile\/?$/)
  }

  beforeEach(() => {
    cy.clearCookies()
    cy.clearLocalStorage()
    signInToSettings()
  })

  it('UI-SCREEN-SETTINGS-001 resolves direct section URLs to the matching settings section', () => {
    const routes: Array<{ path: string; title: string }> = [
      { path: '/settings/profile', title: 'Profile' },
      { path: '/settings/account', title: 'Account' },
      { path: '/settings/appearance', title: 'Appearance' },
      { path: '/settings/notifications', title: 'Notifications' },
      { path: '/settings/display', title: 'Display' },
      { path: '/settings/storage', title: 'Storage' },
      { path: '/settings/categories', title: 'Categories' },
      { path: '/settings/operations', title: 'Operations' },
      { path: '/settings/billing', title: 'Billing' },
    ]

    routes.forEach(({ path, title }) => {
      cy.visit(path)
      cy.location('pathname').should('match', new RegExp(`^${path.replace(/\/$/, '')}\\/?$`))
      cy.contains('h3', title).should('be.visible')
    })
  })

  it('UI-SCREEN-SETTINGS-002 renders canonical settings section labels with stable route links', () => {
    const sections = [
      ['Profile', '/settings/profile'],
      ['Account', '/settings/account'],
      ['Appearance', '/settings/appearance'],
      ['Notifications', '/settings/notifications'],
      ['Display', '/settings/display'],
      ['Storage', '/settings/storage'],
      ['Categories', '/settings/categories'],
      ['Operations', '/settings/operations'],
      ['Billing', '/settings/billing'],
    ] as const

    sections.forEach(([label, href]) => {
      cy.get(`aside a[href="${href}"]`)
        .contains(label)
        .should('exist')
        .and('have.attr', 'href', href)
    })
  })

  it('UI-SCREEN-SETTINGS-003 keeps settings route active and shows actionable section error state when section data fails', () => {
    cy.intercept('GET', '/api/profiles/active', {
      statusCode: 500,
      body: { error: 'failed_to_get_active_profile' },
    }).as('activeProfileFailure')

    cy.visit('/settings/storage')
    cy.wait('@activeProfileFailure')

    cy.location('pathname').should('match', /^\/settings\/storage\/?$/)
    cy.contains('h3', 'Storage').should('be.visible')
    cy.contains('Storage information is unavailable right now.').should(
      'be.visible'
    )
    cy.contains('button', 'Retry').should('be.visible')
  })

  it('UI-SCREEN-SETTINGS-004 exposes Storage route in primary navigation rail', () => {
    cy.get('[data-testid="sidebar-nav-group-other"]')
      .find('a[href="/settings/storage"]')
      .contains('Storage')
      .scrollIntoView()
      .should('exist')
      .click({ force: true })

    cy.location('pathname').should('match', /^\/settings\/storage\/?$/)
  })

  it('UI-SCREEN-SETTINGS-005 blocks editable settings controls when active profile is unavailable', () => {
    cy.clearCookies()
    cy.clearLocalStorage()
    cy.intercept('GET', '/api/profiles/active', {
      statusCode: 404,
      body: { error: 'active_profile_404' },
    }).as('activeProfileMissing')

    cy.visit('/sign-in?redirect=%2Fsettings%2F')
    cy.get('input[name="email"]').clear().type('e2e-settings@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/settings\/profile\/?$/)
    cy.wait('@activeProfileMissing')
    cy.get('[data-testid="settings-profile-context-blocked"]').should(
      'be.visible'
    )
    cy.contains('a', 'Create or Select Profile').should('be.visible')
    cy.contains('button', 'Update profile').should('not.exist')

    cy.visit('/settings/notifications')
    cy.wait('@activeProfileMissing')
    cy.get('[data-testid="settings-profile-context-blocked"]').should(
      'be.visible'
    )
    cy.contains('button', 'Update notifications').should('not.exist')

    cy.visit('/settings/storage')
    cy.wait('@activeProfileMissing')
    cy.contains('a', 'Create or Select Profile').should('be.visible')
  })

  it('UI-SCREEN-SETTINGS-STORAGE-004 keeps last-known paths visible during degraded storage state', () => {
    let storageAttempt = 0
    cy.intercept('GET', '/api/profiles/active', {
      statusCode: 200,
      body: { id: 'default' },
    }).as('activeProfile')
    cy.intercept('GET', '/api/profiles/*/storage', (req) => {
      storageAttempt += 1
      if (storageAttempt === 1) {
        req.reply(200, {
          db_path: 'C:/cabinet/profiles/default/cabinet.db',
          media_dir: 'C:/cabinet/profiles/default/media',
        })
        return
      }
      if (storageAttempt === 2) {
        req.reply(503, { error: 'storage_unavailable' })
        return
      }
      req.reply(200, {
        db_path: 'C:/cabinet/profiles/default/cabinet.db',
        media_dir: 'C:/cabinet/profiles/default/media',
      })
    }).as('storageInfo')

    cy.visit('/settings/storage')
    cy.wait('@activeProfile')
    cy.wait('@storageInfo')
    cy.contains('cabinet.db').should('be.visible')
    cy.contains('/default/media').should('be.visible')

    cy.visit('/settings/storage')
    cy.wait('@activeProfile')
    cy.wait('@storageInfo')
    cy.contains('Storage information is unavailable right now.').should(
      'be.visible'
    )
    cy.contains('cabinet.db').should('be.visible')
    cy.contains('/default/media').should('be.visible')
    cy.contains('Diagnostics actions are unavailable while storage info is degraded.')
      .should('be.visible')
  })

  it('UI-SCREEN-SETTINGS-STORAGE-005 retries storage fetch and recovers without route reload', () => {
    let attempt = 0
    cy.intercept('GET', '/api/profiles/active', {
      statusCode: 200,
      body: { id: 'default' },
    }).as('activeProfile')
    cy.intercept('GET', '/api/profiles/*/storage', (req) => {
      attempt += 1
      if (attempt === 1) {
        req.reply(503, { error: 'storage_unavailable' })
        return
      }
      req.reply(200, {
        db_path: 'C:/cabinet/profiles/default/recovered.db',
        media_dir: 'C:/cabinet/profiles/default/recovered-media',
      })
    }).as('storageInfo')

    cy.visit('/settings/storage')
    cy.wait('@activeProfile')
    cy.wait('@storageInfo')
    cy.contains('Storage information is unavailable right now.').should(
      'be.visible'
    )
    cy.location('pathname').should('match', /^\/settings\/storage\/?$/)

    cy.get('[data-testid="settings-storage-retry"]').click()
    cy.wait('@storageInfo')
    cy.location('pathname').should('match', /^\/settings\/storage\/?$/)
    cy.contains('Storage information is unavailable right now.').should(
      'not.exist'
    )
    cy.contains('C:/cabinet/profiles/default/recovered.db').should('be.visible')
    cy.contains('C:/cabinet/profiles/default/recovered-media').should(
      'be.visible'
    )
  })

  it('UI-SCREEN-SETTINGS-006 resolves Operations section route and active nav state', () => {
    cy.visit('/settings/operations')
    cy.location('pathname').should('match', /^\/settings\/operations\/?$/)
    cy.contains('h3', 'Operations').should('be.visible')
    cy.get('aside a[href="/settings/operations"]').should(
      'have.attr',
      'aria-current',
      'page'
    )
  })

  it('UI-SCREEN-SETTINGS-007 resolves Billing section route and active nav state', () => {
    cy.visit('/settings/billing')
    cy.location('pathname').should('match', /^\/settings\/billing\/?$/)
    cy.contains('h3', 'Billing').should('be.visible')
    cy.get('aside a[href="/settings/billing"]').should(
      'have.attr',
      'aria-current',
      'page'
    )
  })

  it('UI-SCREEN-SETTINGS-008 supports keyboard navigation across settings workflow sections', () => {
    const keyboardRoutes: Array<{ label: string; path: string }> = [
      { label: 'Notifications', path: '/settings/notifications' },
      { label: 'Categories', path: '/settings/categories' },
      { label: 'Operations', path: '/settings/operations' },
      { label: 'Billing', path: '/settings/billing' },
    ]

    keyboardRoutes.forEach(({ label, path }) => {
      cy.get(`aside a[href="${path}"]`)
        .focus()
        .should('be.focused')
        .click()

      cy.location('pathname').should(
        'match',
        new RegExp(`^${path.replace(/\/$/, '')}\\/?$`)
      )
      cy.contains('h3', label).should('be.visible')
      cy.get(`aside a[href="${path}"]`).should(
        'have.attr',
        'aria-current',
        'page'
      )
    })
  })
})
