describe('profile-context-recovery', () => {
  function visibleByTestId(testId: string) {
    return cy.get(`[data-testid="${testId}"]`).first()
  }

  function openTeamSwitcher() {
    cy.get('body').then(($body) => {
      if ($body.find('[data-slot="dropdown-menu-content"]:visible').length > 0) {
        cy.press(Cypress.Keyboard.Keys.ESC)
        cy.get('[data-slot="dropdown-menu-content"]:visible').should(
          'not.exist'
        )
      }
    })
    visibleByTestId('team-switcher-trigger').should('be.visible').click()
    cy.get('[data-slot="dropdown-menu-content"]:visible').should('be.visible')
  }

  function openLocalWorkspace(path: string, profileID: string) {
    cy.stubLocalServerSession(profileID)
    cy.visit(`/sign-in?redirect=${encodeURIComponent(path)}`)
    cy.get('[data-testid="local-device-auth-boundary"]').should('be.visible')
    cy.get('[data-testid="open-local-workspace"]').should('be.enabled').click()
    cy.wait('@activeProfile')
      .its('response.statusCode')
      .should('eq', 200)
    cy.wait('@localServerSession')
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
      if (activeProfileCalls === 2) {
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

    openLocalWorkspace('/inventory/', 'showcase-db')
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)
    cy.wait('@activeProfile')

    openTeamSwitcher()
    visibleByTestId('team-switcher-profile-error')
      .should('be.visible')
      .and('contain', 'Retry loading databases')
    visibleByTestId('team-switcher-retry-profiles').click()
    cy.wait('@activeProfile')

    openTeamSwitcher()
    visibleByTestId('team-option-showcase-db')
      .should('be.visible')
      .and('contain', 'Showcase DB')
    visibleByTestId('team-option-showcase-db-plan')
      .should('be.visible')
      .and('have.text', 'Showcase sample data')
  })

  it('PROFILES-003 recovers a blocked active profile state from the switcher retry', () => {
    let activeProfileCalls = 0

    cy.intercept('GET', '/api/profiles', {
      statusCode: 200,
      body: {
        profiles: [
          { id: 'blocked-db', name: 'Blocked DB' },
          { id: 'working-db', name: 'Working DB' },
        ],
      },
    }).as('profiles')
    cy.intercept('GET', '/api/profiles/active', (req) => {
      activeProfileCalls += 1
      if (activeProfileCalls === 2) {
        req.reply({
          statusCode: 403,
          body: { error: 'active_profile_blocked' },
        })
        return
      }

      req.reply({
        statusCode: 200,
        body: { id: 'working-db', name: 'Working DB' },
      })
    }).as('activeProfile')

    openLocalWorkspace('/inventory/', 'working-db')
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)
    cy.wait('@activeProfile')

    openTeamSwitcher()
    visibleByTestId('team-switcher-profile-error')
      .should('be.visible')
      .and('contain', 'Retry loading databases')
    visibleByTestId('team-switcher-retry-profiles').click()
    cy.wait('@activeProfile')

    openTeamSwitcher()
    visibleByTestId('team-option-working-db')
      .should('be.visible')
      .and('contain', 'Working DB')
    visibleByTestId('team-option-working-db-plan')
      .should('be.visible')
      .and('have.text', 'Database')
  })
})
