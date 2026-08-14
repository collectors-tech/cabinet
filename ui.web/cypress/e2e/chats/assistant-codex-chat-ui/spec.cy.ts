describe('chats/assistant-codex-chat-ui', () => {
  it('ASSISTANT-CODEX-UI-001 keeps a compact conversation-first Agent with context and result links', () => {
    cy.viewport(1280, 720)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.e2eEnsureSignedOut()
    cy.stubLocalServerSession('e2e-profile-001')
    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/inventory/',
      shellWorkspace: 'navigation',
    })
    cy.get('[data-testid="shell-chat-toggle"]').click()
    cy.get('[data-testid="shell-assistant-modal-content"]').should('be.visible')
    cy.get('[data-testid="shell-assistant-panel-title"]').should(
      'contain',
      'Cabinet Agent'
    )
    cy.get('[data-testid="shell-assistant-provider-select"]').should(
      'have.value',
      'openai'
    )
    cy.get('[data-testid="shell-assistant-selected-collection"]').should(
      'contain',
      'All Items'
    )

    cy.get('[data-testid="shell-assistant-compose-input"]').type(
      'create an inventory item CODEX-001 Conversation First Result'
    )
    cy.get('[data-testid="shell-assistant-send-button"]').click()
    cy.get('[data-testid="shell-assistant-action-card"]')
      .should('contain', 'CODEX-001')
      .and('contain', 'Conversation First Result')
    cy.get('[data-testid="shell-assistant-modal-content"]').should(
      'not.contain.text',
      'Agent Skill'
    )
    cy.get('[data-testid="shell-assistant-ui-composer-primitive"]').should(
      'be.visible'
    )
  })
})
