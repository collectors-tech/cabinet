describe('sidebar-collapsed-workspace', () => {
  function signInTo(path: string) {
    cy.e2eReset()
    cy.e2eBootstrap().then((bootstrap) => {
      cy.e2eSetSetupState('present')
      cy.e2eEnsureSignedOut()
      cy.stubLocalServerSession(bootstrap.profile_id)
      cy.useBootstrappedProfile(bootstrap.profile_id, bootstrap.profile_name, {
        path,
      })
    })
  }

  beforeEach(() => {
    cy.clearCookies()
    cy.clearLocalStorage()
  })

  it('keeps collapsed sidebar workspace and footer compact', () => {
    cy.intercept('GET', '/api/runtime', {
      statusCode: 200,
      body: {
        app_version: '2.3.0-e2e',
        build_date: '2026-03-02T00:00:00Z',
      },
    }).as('runtimeMeta')

    signInTo('/inventory/')
    cy.location('pathname', { timeout: 15000 }).should(
      'match',
      /^\/inventory\/?$/
    )
    cy.wait('@runtimeMeta')

    cy.get('[data-slot="sidebar"]')
      .first()
      .then(($sidebar) => {
        if ($sidebar.attr('data-state') !== 'collapsed') {
          cy.get('[data-slot="sidebar-trigger"]').first().click()
        }
      })
    cy.get('[data-slot="sidebar"]')
      .first()
      .should('have.attr', 'data-state', 'collapsed')

    cy.get('[data-testid="sidebar-runtime-meta"]').should('not.exist')
    cy.get('[data-testid="shell-workspace-label"]').should('not.exist')
    cy.get('[data-testid="shell-workspace-switcher"]').should(
      'not.contain',
      'Workspace'
    )
    cy.get('[data-testid="shell-workspace-menu-trigger"]')
      .should('be.visible')
      .and('have.attr', 'aria-label', 'Open workspace menu')
    cy.get('[data-testid="shell-workspace-icon-rail"]').should('be.visible')
    cy.get('[data-testid="shell-workspace-navigation"]')
      .should('be.visible')
      .and('have.attr', 'aria-label', 'Navigation workspace')
    cy.get('[data-testid="shell-workspace-search"]')
      .should('be.visible')
      .and('have.attr', 'aria-label', 'Search workspace')
    cy.get('[data-testid="shell-workspace-assistant"]')
      .should('be.visible')
      .and('have.attr', 'aria-label', 'Chat workspace')
    cy.get('[data-testid="shell-workspace-bell"]')
      .should('be.visible')
      .and('have.attr', 'aria-label', 'Open notification inbox')
    cy.get('[data-testid="shell-workspace-inbox"]').should('not.exist')
    cy.get('[data-testid="shell-workspace-switcher"]').should(
      'not.contain',
      'Search'
    )
    cy.get('[data-testid="shell-workspace-switcher"]').should(
      'not.contain',
      'Chat'
    )
    cy.get('[data-testid="shell-workspace-switcher"]').should(
      'not.contain',
      'Inbox'
    )

    cy.get('[data-testid="shell-workspace-menu-trigger"]').click()
    cy.get('[data-testid="shell-workspace-menu-customise-nav"]').should(
      'be.visible'
    )
    cy.get('[data-testid="shell-workspace-menu-settings"]').should('be.visible')
  })
})
