describe('chats/assistant-workspace-agent-skills', () => {
  function openAgent() {
    cy.viewport(1400, 900)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/inventory/',
      shellWorkspace: 'navigation',
    })
    cy.get('[data-testid="shell-chat-toggle"]').click()
    cy.get('[data-testid="shell-assistant-thread-id"]', {
      timeout: 20000,
    }).should(($threadID) => {
      expect($threadID.text().trim()).not.to.eq('')
      expect($threadID.text().trim()).not.to.eq('bootstrapping')
    })
  }

  it('ASSISTANT-WORKSPACE-011/012/013/014/015/016 routes every product domain from natural conversation with governed context', { retries: 0 }, () => {
    openAgent()

    const requests = [
      ['summarize my dashboard today', 'dashboard'],
      ['review integration provider setup', 'acquisition'],
      ['run my eBay saved market watch', 'acquisition'],
      ['create a purchase order for this item', 'acquisition'],
      ['find inventory item AFX-22020', 'inventory'],
      ['review unlinked media', 'media'],
      ['find eBay discoveries for slot cars', 'acquisition'],
      ['find wishlist entries for AFX', 'wishlist'],
      ['create a collection called Touring Cars', 'collections'],
      ['review profile settings and backup storage', 'admin'],
    ] as const

    requests.forEach(([intent, expectedDomain], index) => {
      const alias = `naturalDomain${index}`
      cy.intercept('POST', '/api/chat/messages', (request) => {
        if (request.body.content !== intent) return
        request.body.context.assistant.provider = 'anthropic'
        request.body.context.assistant.model = 'contract-only'
        request.continue()
      }).as(alias)

      cy.get('[data-testid="shell-assistant-compose-input"]').clear().type(intent)
      cy.get('[data-testid="shell-assistant-send-button"]').click()
      cy.wait(`@${alias}`).then(({ request, response }) => {
        expect(request.body.profile_id).to.eq('e2e-profile-001')
        expect(request.body.agent_context.profile_id).to.eq('e2e-profile-001')
        expect(request.body.agent_context.thread_id).to.be.a('string').and.not.eq('')
        expect(request.body.agent_context.route_id).to.match(/^\/inventory\/?$/)
        expect(request.body.agent_context.surface_id).to.eq(
          'inventory.item.detail'
        )
        expect(request.body.agent_context.selected_record).to.eq(undefined)
        expect(request.body.agent_context.source_channel).to.eq('in-app')
        expect(request.body.content).to.eq(intent)
        expect(request.body.agent_context.intent_text).to.eq(undefined)
        expect(response?.statusCode).to.eq(201)
        expect(response?.body.agent_planner.intent_domain).to.eq(expectedDomain)
        expect(response?.body.agent_planner.source_surface).to.eq(
          'inventory.item.detail'
        )
        expect(response?.body.agent_planner.error.code).to.eq(
          'assistant_provider_adapter_unavailable'
        )
        expect(response?.body.agent_planner.preview_result).to.eq(undefined)
        expect(response?.body.agent_planner.execution_result).to.eq(undefined)
        const agentResponse =
          response?.body.agent_planner.thread_message.context.agent_response
        expect(agentResponse.state).to.eq('provider_unavailable')
        expect(agentResponse.outcome).to.eq('failed')
        expect(agentResponse.retryable).to.eq(true)
        expect(agentResponse.original_intent).to.eq(intent)
        expect(agentResponse.source).to.deep.eq({
          surface: 'inventory.item.detail',
          channel: 'in-app',
        })
        expect(agentResponse.next_action).to.deep.eq({
          kind: 'retry',
          label: 'Retry',
        })
        expect(agentResponse.skill).to.deep.eq({ id: '', name: '' })
      })
    })

    cy.get('[data-testid="shell-assistant-modal-content"]').should(
      'not.contain.text',
      'Agent Skill'
    )
    cy.get('[data-testid="shell-assistant-agent-response-state"]')
      .should('have.attr', 'data-agent-state', 'provider_unavailable')
      .and('have.attr', 'data-agent-outcome', 'failed')
      .and('contain', 'Provider unavailable')
      .and('contain', 'Source: inventory.item.detail / in-app')
      .and('contain', 'Retry')
      .and('not.contain', 'cabinet.')
  })
})
