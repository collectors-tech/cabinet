
describe('chats/assistant-acquisition-workflows', () => {
  function openMainChat() {
    cy.viewport(1400, 900)
    cy.e2eReset()
    cy.e2eBootstrap({ minimalProfile: true })
    cy.e2eSetSetupState('present')
    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/chats/',
    })
    cy.location('pathname', { timeout: 15000 }).should(
      'match',
      /^\/chats\/?$/
    )

    cy.get('[data-testid="chat-new-thread-input"]')
      .clear()
      .type('E2E acquisition Agent')
    cy.get('[data-testid="chat-create-thread-button"]').click()
    cy.contains(
      '[data-testid="chat-thread-item"]',
      'E2E acquisition Agent'
    ).click()
    cy.get('[data-testid="chat-thread-title"]').should(
      'contain',
      'E2E acquisition Agent'
    )
  }

  it('AGENT-ACQUISITION-001/#2083 routes a discovery search through the configured provider and fails safely when setup is missing', () => {
    openMainChat()
    cy.intercept('POST', '/api/chat/messages').as('agentMessage')

    cy.get('[data-testid="chat-compose-input"]').type(
      'Find eBay discoveries for AFX slot cars'
    )
    cy.get('[data-testid="chat-send-button"]').click()

    cy.wait('@agentMessage').then(({ request, response }) => {
      expect(request.body.content).to.eq(
        'Find eBay discoveries for AFX slot cars'
      )
      expect(request.body.agent_context.surface_id).to.eq('chats.main')
      expect(response?.statusCode).to.eq(201)
      expect(response?.body.app_control).to.eq(undefined)
      expect(response?.body.assistant_handoff).to.eq(undefined)
      expect(response?.body.agent_planner.mode).to.eq('provider_planner')
      expect(response?.body.agent_planner.provider).to.eq('openai')
      expect(response?.body.agent_planner.recoverable).to.eq(true)
      expect(response?.body.agent_planner.error.code).to.eq(
        'assistant_provider_unhealthy_provider'
      )
      expect(response?.body.agent_planner.setup_next_action).to.eq(
        'configure_openai_provider'
      )
      expect(response?.body.agent_planner.next_action).to.eq(
        'Configure and test the OpenAI API key in Integrations, then retry this request.'
      )
      expect(response?.body.agent_planner.preview_result).to.eq(undefined)
      expect(response?.body.agent_planner.execution_result).to.eq(undefined)
      const agentResponse =
        response?.body.agent_planner.thread_message.context.agent_response
      expect(agentResponse.state).to.eq('setup_required')
      expect(agentResponse.outcome).to.eq('blocked')
      expect(agentResponse.retryable).to.eq(false)
      expect(agentResponse.original_intent).to.eq(
        'Find eBay discoveries for AFX slot cars'
      )
      expect(agentResponse.next_action).to.deep.eq({
        kind: 'open_setup',
        label: 'Open setup',
        route: '/settings/integrations',
      })
      expect(JSON.stringify(response?.body)).not.to.include('mutation_applied":true')
    })

    cy.get('[data-testid="chat-message-list"]')
      .should('contain', 'Find eBay discoveries for AFX slot cars')
      .and(
        'contain',
        'The assistant provider needs setup before Cabinet can plan that request. No action was completed.'
      )
    cy.get('[data-testid="chat-agent-response-state"]')
      .should('have.attr', 'data-agent-state', 'setup_required')
      .and('have.attr', 'data-agent-outcome', 'blocked')
      .and('contain', 'Setup required')
      .and('not.contain', 'cabinet.discoveries.search')
      .and('not.contain', 'sk-test-secret')
    cy.contains(
      '[data-testid="chat-agent-response-state"] button',
      'Open setup'
    )
      .should('be.visible')
      .click()
    cy.location('pathname').should('eq', '/settings/integrations')
  })
})
