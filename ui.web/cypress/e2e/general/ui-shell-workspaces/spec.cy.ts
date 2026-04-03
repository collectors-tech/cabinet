describe('general/ui-shell-workspaces', () => {
  function openInventory() {
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', { path: '/inventory/' })
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)
  }

  it('UI-SHELL-WORKSPACES-001 switches Navigation, Assistant, and Inbox in the left workspace region', () => {
    openInventory()
    cy.get('[data-testid="shell-workspace-navigation"]').should('have.attr', 'data-active', 'true')
    cy.get('[data-testid="sidebar-nav-link-inventory"]').should('be.visible')

    cy.get('[data-testid="shell-workspace-assistant"]').click()
    cy.get('[data-testid="shell-workspace-assistant"]').should('have.attr', 'data-active', 'true')
    cy.get('[data-testid="shell-assistant-workspace"]').should('be.visible')
    cy.get('[data-testid="shell-assistant-compose-input"]').should('be.visible')

    cy.get('[data-testid="shell-workspace-inbox"]').click()
    cy.get('[data-testid="shell-workspace-inbox"]').should('have.attr', 'data-active', 'true')
    cy.get('[data-testid="shell-inbox-workspace"]').should('be.visible')
    cy.contains('[data-testid="shell-inbox-workspace"]', 'Notifications and asynchronous assistant outcomes').should('be.visible')

    cy.get('[data-testid="shell-workspace-navigation"]').click()
    cy.get('[data-testid="shell-workspace-navigation"]').should('have.attr', 'data-active', 'true')
    cy.get('[data-testid="sidebar-nav-link-inventory"]').should('be.visible')
  })

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

  it('UI-SHELL-WORKSPACES-004 keeps Assistant, Inbox, and /chats semantics distinct', () => {
    openInventory()
    cy.get('[data-testid="shell-workspace-assistant"]').click()
    cy.contains('[data-testid="shell-assistant-workspace"]', 'Persistent route-aware helper workspace for guided actions.').should('be.visible')

    cy.get('[data-testid="shell-workspace-inbox"]').click()
    cy.contains('[data-testid="shell-inbox-workspace"]', 'Notifications and asynchronous assistant outcomes will surface here.').should('be.visible')
    cy.contains('[data-testid="shell-inbox-workspace"]', 'Assistant Thread').should('not.exist')

    cy.visit('/chats')
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/chats\/?$/)
    cy.contains('Persistent profile-scoped conversation threads backed by Cabinet runtime.').should('be.visible')
    cy.contains('Use Assistant for AI-guided help and actions; use Chats for durable conversation threads.').should('be.visible')
    cy.get('[data-testid="shell-workspace-inbox"]').should('have.attr', 'data-active', 'true')
  })
})
