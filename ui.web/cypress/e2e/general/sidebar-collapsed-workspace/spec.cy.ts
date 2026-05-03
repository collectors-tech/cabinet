describe('sidebar-collapsed-workspace', () => {
  function signInTo(path: string) {
    cy.visit(`/sign-in?redirect=${encodeURIComponent(path)}`)
    cy.get('input[name="email"]').clear().type('e2e-shell-nav@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
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

    cy.get('[data-slot="sidebar-trigger"]').first().click()
    cy.get('[data-slot="sidebar"]')
      .first()
      .should('have.attr', 'data-state', 'collapsed')

    cy.get('[data-testid="sidebar-runtime-meta"]').should('not.exist')
    cy.get('[data-testid="shell-workspace-label"]')
      .should('be.visible')
      .and('have.text', 'Work')
    cy.get('[data-testid="shell-workspace-switcher"]').should(
      'not.contain',
      'Workspace'
    )
    cy.get('[data-testid="shell-workspace-menu-trigger"]').should('be.visible')
    cy.get('[data-testid="shell-workspace-navigation"]').should('not.exist')
    cy.get('[data-testid="shell-workspace-assistant"]').should('not.exist')
    cy.get('[data-testid="shell-workspace-inbox"]').should('not.exist')

    cy.get('[data-testid="shell-workspace-menu-trigger"]').click()
    cy.get('[data-testid="shell-workspace-menu-navigation"]').should(
      'be.visible'
    )
    cy.get('[data-testid="shell-workspace-menu-assistant"]').should(
      'be.visible'
    )
    cy.get('[data-testid="shell-workspace-menu-inbox"]').should('be.visible')
  })
})
