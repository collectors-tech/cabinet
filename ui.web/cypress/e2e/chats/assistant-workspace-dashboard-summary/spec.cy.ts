describe('chats/assistant-workspace-dashboard-summary', () => {
  it('AGENT-DASHBOARD-SUMMARY-001/#1942 requests a Dashboard summary through natural conversation', () => {
    cy.viewport(1400, 900)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/dashboard/',
      shellWorkspace: 'navigation',
    })
    cy.get('[data-testid="shell-chat-toggle"]').click()
    cy.get('[data-testid="shell-assistant-thread-id"]', {
      timeout: 20000,
    }).should('not.contain', 'bootstrapping')

    const intent = 'summarize my dashboard activity today'
    cy.intercept('POST', '/api/chat/messages', (request) => {
      request.body.context.assistant.provider = 'anthropic'
      request.body.context.assistant.model = 'contract-only'
      request.continue()
    }).as('dashboardSummary')
    cy.get('[data-testid="shell-assistant-compose-input"]').type(intent)
    cy.get('[data-testid="shell-assistant-send-button"]').click()

    cy.wait('@dashboardSummary').then(({ request, response }) => {
      expect(request.body.content).to.eq(intent)
      expect(request.body.agent_context.route_id).to.match(/^\/dashboard\/?$/)
      expect(request.body.agent_context.source_channel).to.eq('in-app')
      expect(response?.statusCode).to.eq(201)
      expect(response?.body.agent_planner.intent_domain).to.eq('dashboard')
      expect(response?.body.agent_planner.preview_result).to.eq(undefined)
      expect(response?.body.agent_planner.execution_result).to.eq(undefined)
      const agentResponse =
        response?.body.agent_planner.thread_message.context.agent_response
      expect(agentResponse.state).to.eq('provider_unavailable')
      expect(agentResponse.outcome).to.eq('failed')
      expect(agentResponse.retryable).to.eq(true)
      expect(agentResponse.original_intent).to.eq(intent)
      expect(agentResponse.next_action).to.deep.eq({
        kind: 'retry',
        label: 'Retry',
      })
    })
    cy.get('[data-testid="shell-assistant-agent-response-state"]')
      .should('have.attr', 'data-agent-state', 'provider_unavailable')
      .and('have.attr', 'data-agent-outcome', 'failed')
      .and('contain', 'Provider unavailable')
      .and('contain', 'Retry')
      .and('not.contain', 'cabinet.dashboard.summarise_activity')
  })
})
