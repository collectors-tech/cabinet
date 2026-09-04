describe('chats/assistant-execution-surfaces', () => {
  it('ASSISTANT-EXECUTION-001/002/003/004 uses natural preview, explicit confirmation, and no mutation on cancel', () => {
    cy.viewport(1280, 720)
    cy.e2eReset()
    cy.e2eSetSetupState('present')
    cy.e2eBootstrap()
    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/inventory/',
      shellWorkspace: 'navigation',
    })
    cy.get('[data-testid="shell-chat-toggle"]').click()
    cy.get('[data-testid="shell-assistant-thread-id"]', {
      timeout: 20000,
    }).should('not.contain', 'bootstrapping')

    cy.intercept('POST', '/api/chat/messages').as('naturalPreview')
    cy.get('[data-testid="shell-assistant-compose-input"]').type(
      'create an inventory item EXEC-001 Execution Preview Item'
    )
    cy.get('[data-testid="shell-assistant-send-button"]').click()
    cy.wait('@naturalPreview').then(({ request, response }) => {
      expect(request.body.content).to.eq(
        'create an inventory item EXEC-001 Execution Preview Item'
      )
      expect(response?.statusCode).to.eq(201)
      expect(response?.body.app_control.preview.payload.part_number).to.eq(
        'EXEC-001'
      )
      expect(response?.body.app_control.preview.status).to.eq('previewed')
    })
    cy.get('[data-testid="shell-assistant-action-card"]')
      .should('contain', 'EXEC-001')
      .and('contain', 'previewed')
    cy.get('[data-testid="shell-assistant-permission-guidance"]').should(
      'contain',
      'confirm'
    )
    cy.get('[data-testid="shell-assistant-apply-action"]').click()
    cy.get('[data-testid="shell-assistant-apply-confirm-dialog"]').should(
      'be.visible'
    )
    cy.get('[data-testid="shell-assistant-apply-cancel"]').click()
    cy.get('[data-testid="shell-assistant-apply-confirm-dialog"]').should(
      'not.exist'
    )
    cy.request('/api/items?profile_id=e2e-profile-001').then(({ body }) => {
      expect(JSON.stringify(body)).not.to.include('EXEC-001')
    })
    cy.get('[data-testid="shell-assistant-modal-content"]')
      .find('input[placeholder="Part number"]')
      .should('not.exist')
  })
})
