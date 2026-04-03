describe('general/ui-shell-workspaces', () => {
  function openInventory() {
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', { path: '/inventory/' })
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)
  }

  it('UI-SHELL-WORKSPACES-002 activates Assistant workspace from header without route loss', () => {
    openInventory()
    cy.location('pathname').then((initialPathname) => {
      cy.get('[data-testid="shell-workspace-assistant"]').should('have.attr', 'data-active', 'false')
      cy.get('[data-testid="shell-chat-toggle"]').click()
      cy.get('[data-testid="shell-workspace-assistant"]').should('have.attr', 'data-active', 'true')
      cy.get('[data-testid="shell-assistant-workspace"]').should('be.visible')
      cy.get('[data-testid="shell-assistant-route-context"]').should('contain', '/inventory')
      cy.location('pathname').should('eq', initialPathname)
      cy.get('[data-testid="shell-chat-toggle"]').click()
      cy.get('[data-testid="shell-workspace-navigation"]').should('have.attr', 'data-active', 'true')
      cy.get('[data-testid="shell-assistant-workspace"]').should('not.exist')
      cy.location('pathname').should('eq', initialPathname)
    })
  })

  it('UI-SHELL-WORKSPACES-003 preserves Assistant workspace across authenticated route changes', () => {
    openInventory()
    cy.get('[data-testid="shell-chat-toggle"]').click()
    cy.get('[data-testid="shell-workspace-assistant"]').should('have.attr', 'data-active', 'true')
    cy.get('[data-testid="shell-assistant-route-context"]').should('contain', '/inventory')

    cy.get('[data-testid="sidebar-nav-link-wishlist"]').click()
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/wishlist\/?$/)
    cy.get('[data-testid="shell-workspace-assistant"]').should('have.attr', 'data-active', 'true')
    cy.get('[data-testid="shell-assistant-workspace"]').should('be.visible')
    cy.get('[data-testid="shell-assistant-route-context"]').should('contain', '/wishlist')
  })
})
