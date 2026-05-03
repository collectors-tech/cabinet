describe('chats/assistant-execution-surfaces', () => {
  function bootstrapInventory() {
    cy.viewport(1400, 900)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', { path: '/inventory/' })
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)
  }

  it('ASSISTANT-EXECUTION-001/002/003/004 renders preview-before-apply with confirm and explicit permission guidance', () => {
    bootstrapInventory()
    cy.intercept('POST', '/api/chat/actions/preview').as('assistantPreview')
    cy.intercept('POST', '/api/chat/actions/apply').as('assistantApply')

    cy.get('[data-testid="shell-chat-toggle"]').click()
    cy.get('[data-testid="shell-assistant-thread-id"]').should(($threadId) => {
      expect($threadId.text().trim()).not.to.eq('')
      expect($threadId.text().trim()).not.to.eq('bootstrapping')
    })
    cy.get('[data-testid="shell-assistant-permission-boundary"]').should('contain', 'preview-first, confirm-required')
    cy.get('[data-testid="shell-assistant-execution-state"]').should('contain', 'idle')

    cy.get('[data-testid="shell-assistant-preview-part-number"]').clear().type('EXEC-001')
    cy.get('[data-testid="shell-assistant-preview-title"]').clear().type('Execution Preview Item')
    cy.get('[data-testid="shell-assistant-preview-action"]').click()

    cy.wait('@assistantPreview').then(({ request, response }) => {
      expect(request.body.action).to.eq('create_item_stub')
      expect(request.body.payload.part_number).to.eq('EXEC-001')
      expect(response?.statusCode).to.eq(200)
    })

    cy.get('[data-testid="shell-assistant-action-preview"]').should('contain', 'create_item_stub')
    cy.get('[data-testid="shell-assistant-action-preview"]').should('contain', 'EXEC-001')
    cy.get('[data-testid="shell-assistant-execution-state"]').should('contain', 'running')
    cy.get('[data-testid="shell-assistant-apply-action"]').click()
    cy.get('[data-testid="shell-assistant-apply-confirm-dialog"]').should('be.visible')
    cy.get('[data-testid="shell-assistant-apply-confirm-summary"]').should('contain', 'EXEC-001')
    cy.get('[data-testid="shell-assistant-apply-cancel"]').click()
    cy.get('[data-testid="shell-assistant-apply-confirm-dialog"]').should('not.exist')

    cy.get('[data-testid="shell-assistant-apply-action"]').click()
    cy.get('[data-testid="shell-assistant-apply-confirm"]').click()
    cy.wait('@assistantApply').then(({ request, response }) => {
      expect(request.body.confirm).to.eq(true)
      expect(response?.statusCode).to.eq(200)
    })

    cy.get('[data-testid="shell-assistant-execution-state"]').should('contain', 'success')
    cy.get('[data-testid="shell-assistant-apply-result"]').should('contain', 'Applied create_item_stub')
    cy.get('[data-testid="shell-assistant-permission-guidance"]').should('contain', 'preview-only until you explicitly confirm apply')
  })
})
