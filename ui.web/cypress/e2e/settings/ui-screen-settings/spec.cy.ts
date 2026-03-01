describe('ui-screen-settings', () => {
  function signInToSettings() {
    cy.visit('/sign-in?redirect=%2Fsettings%2F')
    cy.get('input[name="email"]').clear().type('e2e-settings@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/settings\/?$/)
  }

  beforeEach(() => {
    cy.clearCookies()
    cy.clearLocalStorage()
    signInToSettings()
  })

  it('UI-SCREEN-SETTINGS-001 resolves direct section URLs to the matching settings section', () => {
    const routes: Array<{ path: string; title: string }> = [
      { path: '/settings/', title: 'Profile' },
      { path: '/settings/account', title: 'Account' },
      { path: '/settings/appearance', title: 'Appearance' },
      { path: '/settings/notifications', title: 'Notifications' },
      { path: '/settings/display', title: 'Display' },
      { path: '/settings/storage', title: 'Storage' },
    ]

    routes.forEach(({ path, title }) => {
      cy.visit(path)
      cy.location('pathname').should('match', new RegExp(`^${path.replace(/\/$/, '')}\\/?$`))
      cy.contains('h3', title).should('be.visible')
    })
  })

  it('UI-SCREEN-SETTINGS-002 renders canonical settings section labels with stable route links', () => {
    const sections = [
      ['Profile', '/settings'],
      ['Account', '/settings/account'],
      ['Appearance', '/settings/appearance'],
      ['Notifications', '/settings/notifications'],
      ['Display', '/settings/display'],
      ['Storage', '/settings/storage'],
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
})
