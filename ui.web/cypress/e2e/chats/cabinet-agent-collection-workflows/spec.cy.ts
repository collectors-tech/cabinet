
describe('chats/cabinet-agent-collection-workflows', () => {
  function openChats() {
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
      .type('E2E Collection Agent Contract')
    cy.get('[data-testid="chat-create-thread-button"]').click()
    cy.contains(
      '[data-testid="chat-thread-item"]',
      'E2E Collection Agent Contract'
    ).click()
  }

  it('AGENT-COLLECTION-WORKFLOWS-001/#2082 routes collection-domain natural language through the shared governed planner contract', () => {
    openChats()

    const requests = [
      {
        intent: 'find wishlist entries for AFX slot cars',
        expectedSurface: 'wishlist',
      },
      {
        intent: 'create a collection called Australian Touring Cars',
        expectedSurface: 'collections',
      },
      {
        intent: 'attach this media to the selected inventory item',
        expectedSurface: 'media',
      },
    ]

    requests.forEach(({ intent, expectedSurface }, index) => {
      const alias = `collectionPlanner${index}`
      cy.intercept('POST', '/api/chat/messages', (req) => {
        if (req.body.content !== intent) return

        // Use the deterministic unavailable adapter so this acceptance proves
        // routing and context without a network call or test credential.
        req.body.context.assistant.provider = 'anthropic'
        req.body.context.assistant.model = 'contract-only'
        req.continue()
      }).as(alias)

      cy.get('[data-testid="chat-compose-input"]').clear().type(intent)
      cy.get('[data-testid="chat-send-button"]').click()

      cy.wait(`@${alias}`).then(({ request, response }) => {
        expect(request.body.profile_id).to.eq('e2e-profile-001')
        expect(request.body.content).to.eq(intent)
        expect(request.body.agent_context.profile_id).to.eq(
          'e2e-profile-001'
        )
        expect(request.body.agent_context.surface_id).to.eq('chats.main')
        expect(request.body.agent_context.source_channel).to.eq('in-app')
        expect(request.body.agent_context.intent_text).to.eq(intent)
        expect(response?.statusCode).to.eq(201)
        expect(response?.body.agent_planner.mode).to.eq('provider_planner')
        expect(response?.body.agent_planner.source_surface).to.eq('chats.main')
        expect(response?.body.agent_planner.source_channel).to.eq('in-app')
        expect(response?.body.agent_planner.recoverable).to.eq(true)
        expect(response?.body.agent_planner.error.code).to.eq(
          'assistant_provider_adapter_unavailable'
        )
        expect(response?.body.agent_planner.intent_domain).to.eq(
          expectedSurface
        )
      })
    })
  })
})
