describe('chats/ui-screen-chat-copilot', () => {
  function signInToChats() {
    cy.visit('/sign-in?redirect=%2Fchats%2F')
    cy.get('input[name="email"]').clear().type('e2e-chat@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/chats\/?$/)
  }

  beforeEach(() => {
    cy.clearCookies()
    cy.clearLocalStorage()
    signInToChats()
  })

  it('UI-SCREEN-CHAT-COPILOT-001 loads threads/messages from API and persists send+reload flow', () => {
    const threadTitle = `E2E Thread ${Date.now()}`
    const messageText = `E2E message ${Date.now()}`

    cy.get('[data-testid="chat-create-thread-button"]').should('be.disabled')
    cy.get('[data-testid="chat-new-thread-input"]').should('not.be.disabled')
    cy.get('[data-testid="chat-new-thread-input"]').type(threadTitle)
    cy.get('[data-testid="chat-create-thread-button"]').should('be.enabled')
    cy.get('[data-testid="chat-create-thread-button"]').click()
    cy.get('[data-testid="chat-thread-list"]').contains(threadTitle).click()

    cy.get('[data-testid="chat-compose-input"]').type(messageText)
    cy.get('[data-testid="chat-send-button"]').click()
    cy.get('[data-testid="chat-message-list"]').contains(messageText).should('be.visible')

    cy.reload()
    cy.get('[data-testid="chat-thread-list"]').contains(threadTitle).click()
    cy.get('[data-testid="chat-message-list"]').contains(messageText).should('be.visible')
  })

  it('UI-SCREEN-CHAT-COPILOT-003 shows deterministic error state with retry when threads bootstrap fails', () => {
    cy.intercept('GET', '/api/chat/threads*', {
      statusCode: 500,
      body: { error: 'failed_to_list_chat_threads' },
    }).as('chatThreadsFailure')

    cy.visit('/chats')
    cy.wait('@chatThreadsFailure')
    cy.get('[data-testid="chat-bootstrap-error"]').should('be.visible')
    cy.contains('button', 'Retry').should('be.visible')
  })
})
