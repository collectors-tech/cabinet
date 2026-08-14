describe('chats/cabinet-agent-production-surface', () => {
  it('ASSISTANT-WORKSPACE-017/#2092 keeps production Chat conversation-first', () => {
    cy.request('POST', '/api/runtime/setup-complete', {
      instance_name: 'Conversation First Production Proof',
      profile_key: 'conversation-first-production',
      auth_mode: 'local',
      storage_mode: 'exe_local',
      storage_data_dir: '',
      portable_mode: false,
      runtime_port_mode: 'auto',
      runtime_fixed_port: 0,
      feature_scanner: true,
      feature_providers: true,
      feature_chat: true,
      bootstrap_workspace: 'Local Workspace',
      bootstrap_database_ref: 'Primary DB',
    })
    cy.request('POST', '/api/profiles', {
      name: 'Conversation First Production Profile',
    }).then((response) => {
      expect(response.status).to.eq(201)
      const profileID = String(response.body.id || '')
      expect(profileID).not.to.eq('')
      cy.request('PUT', '/api/profiles/active', { profile_id: profileID })
        .its('status')
        .should('eq', 200)

      cy.visit('/sign-in?redirect=%2Finventory%2F', {
        onBeforeLoad(window) {
          window.localStorage.setItem(`cabinet.workspace.${profileID}`, '1')
        },
      })
      cy.contains('button', 'Open local workspace').click()
      cy.location('pathname', { timeout: 15000 }).should(
        'match',
        /^\/inventory\/?$/
      )
      cy.get('[data-testid="active-profile-status"]', {
        timeout: 20000,
      }).should('not.contain', 'Loading profiles')
      cy.get('[data-testid="shell-chat-toggle"]').click()
      cy.get('[data-testid="shell-assistant-modal-content"]', {
        timeout: 20000,
      }).should('be.visible')
      cy.get('[data-testid="shell-assistant-ui-composer-primitive"]').should(
        'be.visible'
      )
      cy.get('[data-testid="shell-assistant-action-timeline"]').should('exist')

      cy.get('[data-testid="shell-assistant-modal-content"]')
        .find('input[placeholder="Part number"]')
        .should('not.exist')
      cy.get('[data-testid="shell-assistant-modal-content"]')
        .find('input[placeholder="Item title"]')
        .should('not.exist')
      cy.get('[data-testid="shell-assistant-modal-content"]')
        .should('not.contain.text', 'Preview skill')
        .and('not.contain.text', 'Agent Skill')
    })
  })
})
