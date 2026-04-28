describe('chats/assistant-inbox-handoff', () => {
  function bootstrapInventory() {
    cy.viewport(1400, 900)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', { path: '/inventory/' })
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)
  }

  it('ASSISTANT-INBOX-001/002/003 persists assistant handoff items and reopens linked threads from Inbox', () => {
    bootstrapInventory()
    cy.intercept('POST', '/api/chat/messages').as('assistantMessage')
    cy.get('[data-testid="shell-chat-toggle"]').click()
    cy.get('[data-testid="shell-assistant-compose-input"]').type('check this inventory route asynchronously')
    cy.get('[data-testid="shell-assistant-send-button"]').click()
    cy.wait('@assistantMessage').its('response.statusCode').should('eq', 201)
    cy.contains('[data-testid="shell-assistant-message-list"]', 'Assistant handoff queued in Inbox.').should('exist')

    cy.get('[data-testid="shell-workspace-inbox"]').click()
    cy.get('[data-testid="shell-inbox-workspace"]').should('be.visible')
    cy.get('[data-testid="shell-inbox-notification-card"]').first().scrollIntoView().within(() => {
      cy.contains('Assistant handoff queued').should('be.visible')
      cy.get('[data-testid="shell-inbox-item-status"]').should('contain', 'unread')
      cy.contains('check this inventory route asynchronously').should('be.visible')
      cy.get('[data-testid="shell-inbox-open-assistant"]').click()
    })

    cy.get('[data-testid="shell-workspace-assistant"]').should('have.attr', 'data-active', 'true')
    cy.contains('[data-testid="shell-assistant-message-list"]', 'check this inventory route asynchronously').should('exist')
    cy.contains('[data-testid="shell-assistant-message-list"]', 'Assistant handoff queued in Inbox.').should('exist')
  })
})
