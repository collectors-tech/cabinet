describe('profile-context-recovery', () => {
  function visibleByTestId(testId: string) {
    return cy.get(`[data-testid="${testId}"]`).first()
  }

  function signInTo(path: string) {
    cy.visit(`/sign-in?redirect=${encodeURIComponent(path)}`)
    cy.get('input[name="email"]').clear().type('e2e-profile-recovery@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
  }

  beforeEach(() => {
    cy.clearCookies()
    cy.clearLocalStorage()
  })

  it('PROFILES-003 surfaces active profile load failure and recovers on retry', () => {
    let activeProfileCalls = 0

    cy.intercept('GET', '/api/profiles', {
      statusCode: 200,
      body: {
        profiles: [
          { id: 'primary-db', name: 'Primary DB' },
          { id: 'showcase-db', name: 'Showcase DB' },
        ],
      },
    }).as('profiles')
    cy.intercept('GET', '/api/profiles/active', (req) => {
      activeProfileCalls += 1
      if (activeProfileCalls === 1) {
        req.reply({
          statusCode: 503,
          body: { error: 'active profile unavailable' },
        })
        return
      }

      req.reply({
        statusCode: 200,
        body: { id: 'showcase-db', name: 'Showcase DB' },
      })
    }).as('activeProfile')

    signInTo('/inventory/')
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)
    cy.wait('@activeProfile')

    visibleByTestId('active-profile-status')
      .should('be.visible')
      .and('contain', 'Profile unavailable')

    visibleByTestId('team-switcher-trigger').click()
    visibleByTestId('team-switcher-profile-error')
      .should('be.visible')
      .and('contain', 'Retry loading databases')
    visibleByTestId('team-switcher-retry-profiles').click()
    cy.wait('@activeProfile')

    visibleByTestId('active-profile-name').should('contain', 'Showcase DB')
    visibleByTestId('active-profile-status').should('have.text', 'Showcase sample data')
  })
})
