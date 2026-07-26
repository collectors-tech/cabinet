describe('chats/assistant-workspace-agent-context-route', () => {
  function clearSelectedRecordContext() {
    cy.window().then((win) => {
      win.localStorage.removeItem(
        'cabinet.agent.selected_record.e2e-profile-001'
      )
      win.localStorage.removeItem('cabinet.agent.selected_record.local')
    })
  }

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
    clearSelectedRecordContext()
  }

  function seedAssistantThread() {
    cy.request('POST', '/api/chat/threads', {
      profile_id: 'e2e-profile-001',
      title: 'Assistant Context Route #1714',
      metadata: {
        model: 'gpt-4o-mini',
        provider: 'openai',
        thread_semantics: 'assistant_workspace_session',
      },
    }).then(({ body }) => {
      const threadID = String(body.id)
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
  }

  function openAssistantWorkspace() {
    cy.get('[data-testid="active-profile-status"]', { timeout: 20000 }).should(
      'not.contain',
      'Loading profiles'
    )
    cy.get('[data-testid="shell-chat-toggle"]').click({ force: true })
    cy.get('[data-testid="shell-assistant-modal-content"]', {
      timeout: 20000,
    }).should('be.visible')
    cy.get('[data-testid="shell-assistant-thread-id"]', {
      timeout: 30000,
    }).should(($threadId) => {
      expect($threadId.text().trim()).not.to.eq('')
      expect($threadId.text().trim()).not.to.eq('bootstrapping')
    })
  }

  it('AGENT-CONTEXT-004/#1714 preserves side-panel Agent context across governed route changes', () => {
    bootstrapInventory()
    seedAssistantThread()
    cy.intercept('POST', '/api/chat/messages').as('assistantSidePanelPlanner')
    openAssistantWorkspace()

    cy.get('[data-testid="shell-assistant-compose-input"]').type('open media')
    cy.get('[data-testid="shell-assistant-send-button"]').click()

    let threadId = ''
    cy.wait('@assistantSidePanelPlanner').then(({ request, response }) => {
      expect(response?.statusCode).to.eq(201)
      expect(request.body.profile_id).to.eq('e2e-profile-001')
      expect(String(request.body.thread_id).trim()).not.to.eq('')
      expect(request.body.content).to.eq('open media')
      expect(request.body.agent_context.profile_id).to.eq('e2e-profile-001')
      expect(request.body.agent_context.route_id).to.match(/^\/inventory\/?$/)
      expect(request.body.agent_context.surface_id).to.eq(
        'inventory.item.detail'
      )
      expect(request.body.agent_context.source_channel).to.eq('in-app')
      expect(response?.body.app_control.capability_id).to.eq(
        'navigate.open_surface'
      )
      expect(response?.body.app_control.route).to.eq('/media')
      threadId = String(request.body.thread_id)
      expect(request.body.agent_context.thread_id).to.eq(threadId)
    })

    cy.get('[data-testid="shell-assistant-navigation-action-open"]').click({
      force: true,
    })
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/media\/?$/)
    cy.get('[data-testid="shell-assistant-modal-content"]').should(
      'be.visible'
    )
    cy.then(() => {
      cy.get('[data-testid="shell-assistant-thread-id"]').should(
        'have.text',
        threadId
      )
    })

    cy.intercept('POST', '/api/chat/messages').as(
      'assistantAfterRouteChange'
    )
    cy.get('[data-testid="shell-assistant-compose-input"]').type(
      'continue from the media route'
    )
    cy.get('[data-testid="shell-assistant-send-button"]').click()
    cy.wait('@assistantAfterRouteChange').then(({ request, response }) => {
      expect(response?.statusCode).to.eq(201)
      expect(request.body.thread_id).to.eq(threadId)
      expect(request.body.agent_context.profile_id).to.eq('e2e-profile-001')
      expect(request.body.agent_context.route_id).to.match(/^\/media\/?$/)
      expect(request.body.agent_context.surface_id).to.eq('chats.side-panel')
      expect(request.body.agent_context.source_channel).to.eq('in-app')
      expect(request.body.agent_context.permission_state).to.eq(
        'ask_before_local_changes'
      )
      expect(request.body.agent_context.setup_state).to.eq('ready')
    })
  })
})
