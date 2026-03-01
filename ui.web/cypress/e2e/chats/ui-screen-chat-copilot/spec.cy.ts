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
    cy.get('[data-testid="chat-message-list"]').contains(messageText).should('exist')

    cy.reload()
    cy.get('[data-testid="chat-thread-list"]').contains(threadTitle).click()
    cy.get('[data-testid="chat-message-list"]').contains(messageText).should('exist')
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

  it('UI-SCREEN-CHAT-COPILOT-002 supports attachment upload and preview/apply mutation flow', () => {
    const threadTitle = `E2E Action Thread ${Date.now()}`

    cy.get('[data-testid="chat-new-thread-input"]').type(threadTitle)
    cy.get('[data-testid="chat-create-thread-button"]').click()
    cy.get('[data-testid="chat-thread-list"]').contains(threadTitle).click()

    cy.intercept('POST', '/api/chat/attachments', {
      statusCode: 201,
      body: {
        id: 'att-e2e-1',
        profile_id: 'profile-e2e',
        thread_id: 'thread-e2e',
        filename: 'sample.txt',
        mime_type: 'text/plain',
        size_bytes: 12,
        path: '/tmp/sample.txt',
        created_at: '2026-03-02T00:00:00Z',
      },
    }).as('uploadAttachment')

    cy.intercept('POST', '/api/chat/actions/preview', {
      statusCode: 200,
      body: {
        id: 'preview-e2e-1',
        profile_id: 'profile-e2e',
        thread_id: 'thread-e2e',
        action: 'create_item_stub',
        status: 'previewed',
        created_at: '2026-03-02T00:00:00Z',
      },
    }).as('previewAction')

    cy.intercept('POST', '/api/chat/actions/apply', {
      statusCode: 200,
      body: {
        applied: true,
        action: 'create_item_stub',
        item_id: 'item-e2e-1',
        preview_id: 'preview-e2e-1',
      },
    }).as('applyAction')

    cy.get('[data-testid="chat-attachment-input"]').selectFile({
      contents: Cypress.Buffer.from('attachment-e2e'),
      fileName: 'sample.txt',
      mimeType: 'text/plain',
      lastModified: Date.now(),
    })
    cy.get('[data-testid="chat-upload-attachment-button"]').click()
    cy.wait('@uploadAttachment')
    cy.get('[data-testid="chat-attachment-list"]').contains('sample.txt').should('be.visible')

    cy.get('[data-testid="chat-preview-part-number"]').clear().type('CHAT-E2E-001')
    cy.get('[data-testid="chat-preview-title"]').clear().type('E2E Preview Item')
    cy.get('[data-testid="chat-preview-action-button"]').click()
    cy.wait('@previewAction')
    cy.get('[data-testid="chat-action-preview"]').should('contain', 'previewed')

    cy.get('[data-testid="chat-apply-action-button"]').click()
    cy.wait('@applyAction')
    cy.get('[data-testid="chat-action-apply-result"]').should('contain', 'item-e2e-1')
  })
})
