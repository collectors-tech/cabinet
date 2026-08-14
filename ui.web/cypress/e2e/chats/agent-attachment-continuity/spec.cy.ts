describe('chats/agent-attachment-continuity', () => {
  function bootstrapInventory() {
    cy.viewport(1400, 900)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/inventory/',
    })
    cy.location('pathname', { timeout: 15000 }).should(
      'match',
      /^\/inventory\/?$/
    )
  }

  function openContextualAgent() {
    cy.get('[data-testid="active-profile-status"]', {
      timeout: 20000,
    }).should('not.contain', 'Loading profiles')
    cy.get('body').then(($body) => {
      const visibleModal = $body
        .find('[data-testid="shell-assistant-modal-content"]')
        .filter(':visible')
      if (visibleModal.length > 0) {
        cy.wrap(visibleModal).should('have.length', 1).and('be.visible')
        return
      }

      cy.get('[data-testid="shell-chat-toggle"]')
        .filter(':visible')
        .should('have.length', 1)
        .and('have.attr', 'aria-label', 'Open Cabinet Agent')
        .and('be.enabled')
        .click()
    })
    cy.get('[data-testid="shell-assistant-modal-content"]', {
      timeout: 20000,
    }).should('be.visible')
    cy.get('[data-testid="shell-assistant-thread-id"]', {
      timeout: 20000,
    }).should('not.have.text', 'bootstrapping')
  }

  it('AGENT-ATTACHMENTS-002 persists one scoped attachment across contextual/full/reload handoff', () => {
    bootstrapInventory()
    cy.intercept('POST', '/api/chat/attachments').as('attachmentUpload')
    cy.intercept('POST', '/api/chat/messages').as('attachmentMessage')
    cy.request('POST', '/api/chat/threads', {
      profile_id: 'e2e-profile-001',
      title: 'Agent Attachment Continuity',
      metadata: {
        provider: 'openai',
        model: 'gpt-4o-mini',
        thread_semantics: 'assistant_workspace_session',
      },
    }).then(({ body }) => {
      const threadID = String(body.id)
      cy.wrap(threadID).as('attachmentThreadId')
      cy.window().then((win) => {
        win.localStorage.setItem(
          'cabinet.assistant.workspace.thread.e2e-profile-001',
          threadID
        )
        win.localStorage.setItem(
          'cabinet.assistant.workspace.provider.e2e-profile-001',
          'openai'
        )
        win.localStorage.setItem(
          'cabinet.assistant.workspace.model.e2e-profile-001',
          'gpt-4o-mini'
        )
      })
    })
    openContextualAgent()
    cy.get('[data-testid="shell-assistant-attachment-input"]').selectFile(
      {
        contents: Cypress.Buffer.from('agent attachment continuity proof'),
        fileName: 'agent-continuity-proof.txt',
        mimeType: 'text/plain',
      },
      { force: true }
    )
    cy.get('[data-testid="shell-assistant-attachment-upload"]').click()
    cy.wait('@attachmentUpload').then(({ response }) => {
      expect(response?.statusCode).to.eq(201)
      cy.wrap(String(response?.body.id)).as('attachmentId')
    })
    cy.get('[data-testid="shell-assistant-attachment-list"]').should(
      'contain',
      'agent-continuity-proof.txt'
    )
    cy.get('[data-testid="shell-assistant-compose-input"]').type(
      'remember this exact attachment'
    )
    cy.get('[data-testid="shell-assistant-send-button"]').click()
    cy.wait('@attachmentMessage').its('response.statusCode').should('eq', 201)
    cy.get('[data-testid="shell-assistant-message-attachment"]')
      .should('have.length', 1)
      .and('contain', 'agent-continuity-proof.txt')
      .and('contain', 'Uploaded in Cabinet')

    cy.get('@attachmentThreadId').then((threadID) => {
      cy.visit(`/chats?thread_id=${encodeURIComponent(String(threadID))}`)
    })
    cy.get('[data-testid="chat-message-attachment"]', { timeout: 20000 })
      .should('have.length', 1)
      .and('contain', 'agent-continuity-proof.txt')
    cy.reload()
    cy.get('[data-testid="chat-message-attachment"]', { timeout: 20000 })
      .should('have.length', 1)
      .and('contain', 'agent-continuity-proof.txt')

    cy.visit('/inventory/')
    openContextualAgent()
    cy.get('[data-testid="shell-assistant-message-attachment"]', {
      timeout: 20000,
    })
      .should('have.length', 1)
      .and('contain', 'agent-continuity-proof.txt')
  })

  it('AGENT-ATTACHMENTS-003 clears staged attachments at thread and provider/model fork boundaries', () => {
    bootstrapInventory()
    cy.intercept('POST', '/api/chat/attachments').as('scopedUpload')
    openContextualAgent()
    cy.get('[data-testid="shell-assistant-attachment-input"]').selectFile(
      {
        contents: Cypress.Buffer.from('must remain in original scope'),
        fileName: 'scope-boundary.txt',
        mimeType: 'text/plain',
      },
      { force: true }
    )
    cy.get('[data-testid="shell-assistant-attachment-upload"]').click()
    cy.wait('@scopedUpload').its('response.statusCode').should('eq', 201)
    cy.get('[data-testid="shell-assistant-attachment-list"]').should(
      'contain',
      'scope-boundary.txt'
    )
    cy.get('[data-testid="shell-assistant-new-thread"]').click()
    cy.get('[data-testid="shell-assistant-attachment-list"]').should(
      'not.exist'
    )

    cy.get('[data-testid="shell-assistant-attachment-input"]').selectFile(
      {
        contents: Cypress.Buffer.from('must not cross model fork'),
        fileName: 'fork-boundary.txt',
        mimeType: 'text/plain',
      },
      { force: true }
    )
    cy.get('[data-testid="shell-assistant-attachment-upload"]').click()
    cy.wait('@scopedUpload').its('response.statusCode').should('eq', 201)
    cy.get('[data-testid="shell-assistant-attachment-list"]').should(
      'contain',
      'fork-boundary.txt'
    )
    cy.get('[data-testid="shell-assistant-model-select"]').select(
      'gpt-4.1-mini'
    )
    cy.get('[data-testid="shell-assistant-attachment-list"]').should(
      'not.exist'
    )
  })
})
