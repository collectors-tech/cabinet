describe('chats/chat-new-thread-dialog', { retries: 0 }, () => {
  it('CHATS-WORKSPACE-005/#2307 creates and selects a Chat when the conversation rail is hidden', () => {
    cy.viewport(1000, 660)
    cy.e2eReset()
    cy.e2eBootstrap({ minimalProfile: true })
    cy.e2eSetSetupState('present')
    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/chats/',
    })
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/chats\/?$/)
    cy.get('[data-testid="chat-conversation-rail"]').should('not.be.visible')

    cy.intercept('POST', '/api/chat/threads').as('createThread')
    cy.get(
      '[data-testid="chat-new-chat-button"], [data-testid="chat-empty-workspace-action"]'
    )
      .filter(':visible')
      .first()
      .should('be.visible')
      .and('be.enabled')
      .click()
    cy.get('[data-testid="chat-new-thread-dialog"]')
      .should('be.visible')
      .and('contain', 'messages, context, attachments, and governed actions')
    cy.get('[data-testid="chat-new-thread-input"]')
      .should('be.focused')
      .type('Compact release planning')
    cy.get('[data-testid="chat-create-thread-button"]')
      .should('be.enabled')
      .click()

    cy.wait('@createThread').then(({ request, response }) => {
      expect(request.body).to.deep.equal({
        profile_id: 'e2e-profile-001',
        title: 'Compact release planning',
      })
      expect(response?.statusCode).to.equal(201)
    })
    cy.get('[data-testid="chat-new-thread-dialog"]').should('not.exist')
    cy.get('[data-testid="chat-thread-title"]')
      .should('be.visible')
      .and('contain', 'Compact release planning')
    cy.location('pathname').should('match', /^\/chats\/?$/)
  })
})
