describe('chats/assistant-codex-chat-ui', () => {
  function bootstrapInventory() {
    cy.viewport(1400, 900)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/inventory/',
    })
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)
  }

  it('ASSISTANT-CODEX-UI-001 renders a simple chat-first assistant with compact model/context and result links', () => {
    bootstrapInventory()
    cy.intercept('POST', '/api/chat/actions/preview').as('assistantPreview')
    cy.intercept('POST', '/api/chat/actions/apply').as('assistantApply')

    cy.get('[data-testid="shell-chat-toggle"]').click()
    cy.get('[data-testid="shell-assistant-codex-chat"]').should('be.visible')
    cy.get('[data-testid="shell-assistant-thread-id"]').should(($threadId) => {
      expect($threadId.text().trim()).not.to.eq('')
      expect($threadId.text().trim()).not.to.eq('bootstrapping')
    })
    cy.contains('[data-testid="shell-assistant-workspace"]', 'Execution Surface').should('not.exist')
    cy.get('[data-testid="shell-assistant-context-chip"]').should('contain', '/inventory')
    cy.get('[data-testid="shell-assistant-model-chip"]').should('contain', 'openai')
    cy.contains('[data-testid="shell-assistant-message-list"]', 'Ask Cabinet to update records').should('exist')

    cy.get('[data-testid="shell-assistant-compose-input"]').type('create a quick item and give me the link')
    cy.get('[data-testid="shell-assistant-send-button"]').click()
    cy.get('[data-testid="shell-assistant-message-bubble-user"]').should('contain', 'create a quick item and give me the link')
    cy.get('[data-testid="shell-assistant-message-bubble-assistant"]').should('contain', 'Assistant handoff queued in Inbox.')

    cy.get('[data-testid="shell-assistant-preview-part-number"]').clear().type('CODEX-001')
    cy.get('[data-testid="shell-assistant-preview-title"]').clear().type('Codex Style Result Item')
    cy.get('[data-testid="shell-assistant-preview-action"]').click()
    cy.wait('@assistantPreview').its('response.statusCode').should('eq', 200)
    cy.get('[data-testid="shell-assistant-action-card"]').scrollIntoView().should('contain', 'CODEX-001')

    cy.get('[data-testid="shell-assistant-apply-action"]').click()
    cy.get('[data-testid="shell-assistant-apply-confirm"]').click()
    cy.wait('@assistantApply').its('response.statusCode').should('eq', 200)
    cy.get('[data-testid="shell-assistant-apply-result"]').scrollIntoView().should('contain', 'Applied create_item_stub')
    cy.get('[data-testid="shell-assistant-result-link"]')
      .should('contain', 'Open item')
      .and('have.attr', 'href')
      .and('include', '/inventory/?item=')
  })
})
