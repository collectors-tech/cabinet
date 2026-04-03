describe('chats/assistant-workspace', () => {
  function bootstrapInventory() {
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', { path: '/inventory/' })
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)
  }

  it('ASSISTANT-WORKSPACE-001 preserves assistant thread continuity across route and workspace changes', () => {
    bootstrapInventory()
    cy.get('[data-testid="shell-chat-toggle"]').click()
    cy.get('[data-testid="shell-assistant-compose-input"]').type('remember this route context')
    cy.get('[data-testid="shell-assistant-send-button"]').click()
    cy.contains('[data-testid="shell-assistant-message-list"]', 'remember this route context').should(
      'be.visible'
    )
    cy.get('[data-testid="shell-assistant-thread-id"]').invoke('text').as('threadId')

    cy.get('[data-testid="shell-workspace-navigation"]').click()
    cy.get('[data-testid="shell-chat-toggle"]').click()
    cy.get('[data-testid="shell-assistant-message-list"]').contains('remember this route context')
    cy.get('[data-testid="sidebar-nav-link-wishlist"]').click()
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/wishlist\/?$/)
    cy.get('[data-testid="shell-workspace-assistant"]').should('have.attr', 'data-active', 'true')
    cy.get('[data-testid="shell-assistant-message-list"]').contains('remember this route context')
    cy.get('@threadId').then((threadId) => {
      cy.get('[data-testid="shell-assistant-thread-id"]').should('have.text', String(threadId).trim())
    })
  })

  it('ASSISTANT-WORKSPACE-002 sends deterministic route/profile/selection context in assistant message envelopes', () => {
    bootstrapInventory()
    cy.intercept('POST', '/api/chat/messages').as('assistantMessage')
    cy.get('[data-testid="shell-chat-toggle"]').click()
    cy.get('[data-testid="shell-assistant-selection-context"]').should('contain', 'All Items')
    cy.get('[data-testid="shell-assistant-compose-input"]').type('what should I do with this inventory route?')
    cy.get('[data-testid="shell-assistant-send-button"]').click()
    cy.wait('@assistantMessage').then(({ request }) => {
      expect(request.body.profile_id).to.eq('e2e-profile-001')
      expect(request.body.thread_id).to.be.a('string').and.not.be.empty
      expect(request.body.context.route.pathname).to.eq('/inventory/')
      expect(request.body.context.profile.id).to.eq('e2e-profile-001')
      expect(request.body.context.selection.active_workspace_collection).to.eq('All Items')
    })
    cy.contains('[data-testid="shell-assistant-message-list"]', 'what should I do with this inventory route?').should(
      'be.visible'
    )
  })
})
