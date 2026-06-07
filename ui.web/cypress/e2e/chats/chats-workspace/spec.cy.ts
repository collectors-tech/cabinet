describe('chats/chats-workspace', () => {
  function openChats() {
    cy.e2eReset()
    cy.e2eBootstrap({ minimalProfile: true })
    cy.e2eSetSetupState('present')
    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', { path: '/chats/' })
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/chats\/?$/)
  }

  function createThread(title: string) {
    cy.get('[data-testid="chat-new-thread-input"]').clear().type(title)
    cy.get('[data-testid="chat-create-thread-button"]').click()
    cy.contains('[data-testid="chat-thread-item"]', title).click()
    cy.get('[data-testid="chat-thread-title"]').should('contain', title)
  }

  it('CHATS-WORKSPACE-001 renders Cabinet-specific chats semantics instead of placeholder inbox copy', () => {
    openChats()

    cy.contains('h1', 'Chats').should('be.visible')
    cy.get('[data-testid="chat-workspace-description"]')
      .should('be.visible')
      .and('contain.text', 'Persistent profile-scoped conversation threads backed by Cabinet runtime.')
    cy.get('[data-testid="chat-workspace-boundary-note"]')
      .should('be.visible')
      .and('contain.text', 'Use Assistant for AI-guided help and actions; use Chats for durable conversation threads.')
    cy.get('[data-testid="chat-thread-list"]').should('be.visible')
    cy.contains('No chat threads yet.').should('be.visible')
    cy.contains(/inbox template|stock inbox|placeholder/i).should('not.exist')
  })

  it('CHATS-WORKSPACE-002 preserves the active thread after send', () => {
    openChats()
    createThread('E2E Workspace Thread Preservation')

    cy.get('[data-testid="chat-compose-input"]').type('Hello persistent chats workspace')
    cy.get('[data-testid="chat-send-button"]').click()

    cy.location('pathname').should('match', /^\/chats\/?$/)
    cy.get('[data-testid="chat-thread-title"]').should('contain', 'E2E Workspace Thread Preservation')
    cy.get('[data-testid="chat-message-list"]').should('contain', 'Hello persistent chats workspace')
    cy.contains('[data-testid="chat-thread-item"]', 'E2E Workspace Thread Preservation')
      .should('have.class', 'border-primary')
  })

  it('CHATS-WORKSPACE-003 states the Assistant versus Chats boundary explicitly', () => {
    openChats()

    cy.get('[data-testid="chat-workspace-boundary-note"]')
      .should('contain.text', 'Assistant')
      .and('contain.text', 'Chats')
    cy.get('[data-testid="shell-chat-toggle"]').should('be.visible')
    cy.location('pathname').should('match', /^\/chats\/?$/)
  })

  it('CHATS-WORKSPACE-004 renders original two-pane chats layout parity', () => {
    openChats()

    cy.get('[data-testid="chat-layout"]').should('be.visible')
    cy.get('[data-testid="chat-conversation-rail"]').should('be.visible')
    cy.get('[data-testid="chat-conversation-search"]')
      .should('be.visible')
      .and('have.attr', 'placeholder', 'Search messages')
    cy.get('[data-testid="chat-empty-workspace-state"]')
      .should('be.visible')
      .and('contain.text', 'Select a conversation')
      .and('contain.text', 'Choose an existing thread or create a new one to continue a durable Cabinet conversation.')
    cy.get('[data-testid="chat-empty-workspace-action"]')
      .should('be.visible')
      .and('contain.text', 'Start a conversation')

    createThread('E2E Visual Parity Thread')

    cy.contains('[data-testid="chat-thread-item"]', 'E2E Visual Parity Thread')
      .should('be.visible')
      .within(() => {
        cy.get('[data-testid="chat-thread-avatar"]').should('be.visible')
        cy.get('[data-testid="chat-thread-preview"]').should(
          'contain.text',
          'No messages yet'
        )
      })
  })
})
