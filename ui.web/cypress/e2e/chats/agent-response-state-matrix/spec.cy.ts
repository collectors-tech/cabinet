type AgentResponseFixture = {
  state: string
  outcome: string
  retryable?: boolean
  expectedAction?: string
}

describe('chats/agent-response-state-matrix', () => {
  const profileID = 'e2e-profile-001'
  const threadID = 'e2e-thread-001'
  const fixtures: AgentResponseFixture[] = [
    { state: 'read_result', outcome: 'success' },
    { state: 'clarification_required', outcome: 'needs_input', expectedAction: 'Provide details' },
    { state: 'setup_required', outcome: 'blocked', expectedAction: 'Open setup' },
    { state: 'authority_blocked', outcome: 'blocked', expectedAction: 'Review authority' },
    { state: 'unsupported', outcome: 'blocked', expectedAction: 'Start a new request' },
    { state: 'provider_unavailable', outcome: 'failed', retryable: true, expectedAction: 'Retry' },
    { state: 'retryable_failure', outcome: 'failed', retryable: true, expectedAction: 'Retry' },
    { state: 'preview_required', outcome: 'preview', expectedAction: 'Apply' },
    { state: 'preview_expired', outcome: 'failed', expectedAction: 'Create a new preview' },
    { state: 'preview_failed', outcome: 'failed', retryable: true, expectedAction: 'Retry' },
    { state: 'preview_stale_target', outcome: 'failed', expectedAction: 'Refresh target' },
    { state: 'cancelled', outcome: 'cancelled', expectedAction: 'Start a new request' },
    { state: 'applied', outcome: 'applied' },
  ]

  function seedState(state: string) {
    return cy.request('POST', '/api/test/chat/agent-response-state', {
      profile_id: profileID,
      thread_id: threadID,
      state,
      original_intent: `bounded intent for ${state}`,
    })
  }

  function assertCard(selector: 'chat-agent-response-state' | 'shell-assistant-agent-response-state', fixture: AgentResponseFixture) {
    const cardSelector = `[data-testid="${selector}"]`
    const actionTestID = selector.replace('-state', '')
    cy.get(cardSelector)
      .should('have.length', 1)
      .and('have.attr', 'data-agent-state', fixture.state)
      .and('have.attr', 'data-agent-outcome', fixture.outcome)
      .and('contain', fixture.state.replace(/_/g, ' '))
      .and('contain', 'Cabinet Inventory Search')
      .and('contain', 'chats.main')
      .and('contain', 'in-app')
    if (fixture.outcome === 'failed' || fixture.outcome === 'blocked') {
      cy.get(cardSelector).should('not.contain', 'Completed').and('not.contain', 'Applied successfully')
      cy.get(cardSelector).find(`[data-testid="${actionTestID}-apply"]`).should('not.exist')
    }
    if (fixture.expectedAction) cy.get(cardSelector).contains(fixture.expectedAction).should('exist')
    if (!fixture.retryable) cy.get(cardSelector).find(`[data-testid="${actionTestID}-retry"]`).should('not.exist')
  }

  beforeEach(() => {
    cy.viewport(1440, 1000)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
  })

  it('AGENT-RESPONSE-STATES-001/#2099 normalizes every server state identically in main and contextual Chat', () => {
    cy.wrap(fixtures).each((fixture) => seedState(fixture.state))
    cy.useBootstrappedProfile(profileID, 'E2E Local', { path: '/chats/' })
    cy.visit(`/chats/?thread_id=${threadID}`)
    fixtures.forEach((fixture, index) => {
      seedState(fixture.state).then(() => {
        if (index > 0) {
          cy.useBootstrappedProfile(profileID, 'E2E Local', { path: '/chats/' })
          cy.visit(`/chats/?thread_id=${threadID}`)
        }
        assertCard('chat-agent-response-state', fixture)
        cy.get('[data-testid="shell-chat-toggle"]').click({ force: true })
        cy.get('[data-testid="shell-assistant-thread-select"]', { timeout: 20000 }).select(threadID, { force: true })
        assertCard('shell-assistant-agent-response-state', fixture)
        cy.get('[data-testid="shell-assistant-close"]').click({ force: true })
      })
    })
  })

  it('AGENT-RESPONSE-STATES-002/#2099 retains only latest server-owned state and retries the bounded original intent on the same profile/thread', () => {
    seedState('preview_required')
    seedState('retryable_failure')
    cy.intercept('POST', '/api/chat/messages').as('retryAgentIntent')
    cy.useBootstrappedProfile(profileID, 'E2E Local', { path: '/chats/' })
    cy.visit(`/chats/?thread_id=${threadID}`)
    cy.get('[data-testid="chat-agent-response-state"]').should('have.length', 1).and('have.attr', 'data-agent-state', 'retryable_failure')
    cy.get('[data-testid="chat-action-preview"]').should('not.exist')
    cy.get('[data-testid="chat-agent-response-retry"]').click()
    cy.wait('@retryAgentIntent').then(({ request }) => {
      expect(request.body.profile_id).to.eq(profileID)
      expect(request.body.thread_id).to.eq(threadID)
      expect(request.body.role).to.eq('user')
      expect(request.body.content).to.eq('bounded intent for retryable_failure')
    })
    seedState('ordinary_response')
    cy.reload()
    cy.get('[data-testid="chat-agent-response-state"]').should('not.exist')
    cy.get('[data-testid="chat-action-preview"]').should('not.exist')
    cy.get('[data-testid="chat-navigation-action"]').should('not.exist')
    cy.get('[data-testid="shell-chat-toggle"]').click({ force: true })
    cy.get('[data-testid="shell-assistant-thread-select"]', { timeout: 20000 }).select(threadID, { force: true })
    cy.get('[data-testid="shell-assistant-agent-response-state"]').should('not.exist')
    cy.get('[data-testid="shell-assistant-action-card"]').should('not.exist')
    cy.get('[data-testid="shell-assistant-navigation-action"]').should('not.exist')
  })

  it('AGENT-RESPONSE-STATES-003 routes assistant setup directly to the OpenAI provider configuration', () => {
    seedState('setup_required')
    cy.useBootstrappedProfile(profileID, 'E2E Local', { path: '/chats/' })
    cy.visit(`/chats/?thread_id=${threadID}`)
    cy.get('[data-testid="chat-agent-response-state"]')
      .should('have.attr', 'data-agent-state', 'setup_required')
      .contains('Open setup')
      .click()
    cy.location('pathname').should('eq', '/integrations')
    cy.get('[data-testid="openai-config-dialog"]', { timeout: 20000 }).should('be.visible')
  })
})
