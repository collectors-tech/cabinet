
describe('chats/cabinet-agent-continuity', () => {
  function openInventoryWithAgent(key: 'enter' | 'space' = 'enter') {
    cy.viewport(1280, 720)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/inventory/',
      shellWorkspace: 'navigation',
    })
    cy.location('pathname', { timeout: 15000 }).should(
      'match',
      /^\/inventory\/?$/
    )
    cy.get('[data-testid="shell-chat-toggle"]')
      .should('have.attr', 'aria-label', 'Open Cabinet Agent')
      .focus()
      .should('be.focused')
    cy.press(
      key === 'space'
        ? Cypress.Keyboard.Keys.SPACE
        : Cypress.Keyboard.Keys.ENTER
    )
    cy.get('[data-testid="shell-assistant-modal-content"]', {
      timeout: 20000,
    }).should('be.visible')
    cy.get('[data-testid="shell-assistant-thread-id"]', {
      timeout: 20000,
    }).should(($threadID) => {
      expect($threadID.text().trim()).not.to.eq('')
      expect($threadID.text().trim()).not.to.eq('bootstrapping')
    })
  }

  it('AGENT-CENTRAL-001 keeps one governed thread and pending preview across contextual and full Agent', () => {
    openInventoryWithAgent()
    cy.intercept('POST', '/api/chat/messages').as('agentMessage')

    cy.get('[data-testid="shell-assistant-compose-input"]').type(
      'create an inventory item AGENT-2081 Unified Agent'
    )
    cy.get('[data-testid="shell-assistant-send-button"]').click()
    cy.wait('@agentMessage').its('response.statusCode').should('eq', 201)
    cy.get('[data-testid="shell-assistant-message-list"]').should(
      'contain',
      'create an inventory item AGENT-2081 Unified Agent'
    )
    cy.get('[data-testid="shell-assistant-action-preview"]')
      .should('contain', 'create_inventory_item')
      .and('contain', 'AGENT-2081')
    cy.get('[data-testid="shell-assistant-thread-id"]')
      .invoke('text')
      .then((threadID) => {
        cy.get('[data-testid="shell-assistant-open-full-agent"]')
          .should('have.attr', 'aria-label', 'Open this thread in full Cabinet Agent')
          .and('be.enabled')
          .click()

        cy.location('pathname', { timeout: 15000 }).should(
          'match',
          /^\/chats\/?$/
        )
        cy.location('search').should(
          'contain',
          `thread_id=${encodeURIComponent(threadID.trim())}`
        )
        cy.get('[data-testid="chat-thread-title"]').should(
          'contain',
          'Assistant Workspace'
        )
        cy.get('[data-testid="chat-message-list"]').should(
          'contain',
          'create an inventory item AGENT-2081 Unified Agent'
        )
        cy.get('[data-testid="chat-action-preview"]')
          .should('contain', 'create_inventory_item')
          .and('contain', 'AGENT-2081')

        cy.get('[data-testid="shell-workspace-assistant"]')
          .should('be.visible')
          .and('be.enabled')
          .click()
        cy.get('[data-testid="shell-assistant-thread-id"]').should(
          'have.text',
          threadID.trim()
        )
        cy.get('[data-testid="shell-assistant-message-list"]').should(
          'contain',
          'create an inventory item AGENT-2081 Unified Agent'
        )
        cy.get('[data-testid="shell-assistant-action-preview"]')
          .should('contain', 'create_inventory_item')
          .and('contain', 'AGENT-2081')
      })
  })

  it('AGENT-CENTRAL-002 explains governed capabilities in contextual and full Agent', () => {
    openInventoryWithAgent('space')
    cy.intercept('POST', '/api/chat/messages').as('capabilityMessage')

    cy.get('[data-testid="shell-assistant-compose-input"]').type(
      'what can Cabinet Agent do here?'
    )
    cy.get('[data-testid="shell-assistant-send-button"]').click()
    cy.wait('@capabilityMessage').then(({ response }) => {
      expect(response?.statusCode).to.eq(201)
      expect(response?.body.agent_capabilities.mode).to.eq(
        'capability_explanation'
      )
      expect(response?.body.agent_capabilities.summary.total).to.be.greaterThan(
        0
      )
      expect(
        response?.body.agent_capabilities.thread_message.context.agent_response
          .state
      ).to.eq('read_result')
      expect(
        response?.body.agent_capabilities.thread_message.context.agent_response
          .outcome
      ).to.eq('success')
    })

    cy.get('[data-testid="shell-assistant-agent-response-state"]')
      .scrollIntoView()
      .should('be.visible')
      .and('have.attr', 'data-agent-state', 'read_result')
      .and('have.attr', 'data-agent-outcome', 'success')
      .and('contain', 'Cabinet Agent capabilities')
      .and('contain', 'Cabinet Agent can explain available skills')

    cy.get('[data-testid="shell-assistant-open-full-agent"]')
      .should('be.enabled')
      .focus()
      .should('be.focused')
    cy.press(Cypress.Keyboard.Keys.SPACE)
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/chats\/?$/)
    cy.get('[data-testid="chat-agent-response-state"]')
      .scrollIntoView()
      .should('be.visible')
      .and('have.attr', 'data-agent-state', 'read_result')
      .and('have.attr', 'data-agent-outcome', 'success')
      .and('contain', 'Cabinet Agent capabilities')
      .and('contain', 'Cabinet Agent can explain available skills')
  })

  it('AGENT-CENTRAL-003 renders governed setup guidance without claiming an action ran', () => {
    openInventoryWithAgent()
    cy.intercept('POST', '/api/chat/messages').as('plannerMessage')

    cy.get('[data-testid="shell-assistant-compose-input"]').type(
      'search my inventory for E2E'
    )
    cy.get('[data-testid="shell-assistant-send-button"]').click()
    cy.wait('@plannerMessage').then(({ response }) => {
      expect(response?.statusCode).to.eq(201)
      expect(response?.body.agent_planner.mode).to.eq('provider_planner')
      expect(response?.body.agent_planner.recoverable).to.eq(true)
      expect(response?.body.agent_planner.error.code).to.eq(
        'assistant_provider_unhealthy_provider'
      )
      expect(response?.body.agent_planner.preview_result).to.eq(undefined)
      expect(response?.body.agent_planner.execution_result).to.eq(undefined)
      expect(
        response?.body.agent_planner.thread_message.context.agent_response.state
      ).to.eq('setup_required')
      expect(
        response?.body.agent_planner.thread_message.context.agent_response.outcome
      ).to.eq('blocked')
      expect(
        response?.body.agent_planner.thread_message.context.agent_response
          .retryable
      ).to.eq(false)
      expect(
        response?.body.agent_planner.thread_message.context.agent_response
          .next_action
      ).to.deep.eq({
        kind: 'open_setup',
        label: 'Open setup',
        route: '/settings/integrations',
      })
    })

    cy.get('[data-testid="shell-assistant-agent-response-state"]')
      .scrollIntoView()
      .should('be.visible')
      .and('have.attr', 'data-agent-state', 'setup_required')
      .and('have.attr', 'data-agent-outcome', 'blocked')
      .and('contain', 'Setup required')
      .and('contain', 'Open setup')
    cy.get('[data-testid="shell-assistant-action-preview"]').should('not.exist')

    cy.get('[data-testid="shell-assistant-open-full-agent"]').click()
    cy.get('[data-testid="chat-agent-response-state"]')
      .scrollIntoView()
      .should('be.visible')
      .and('have.attr', 'data-agent-state', 'setup_required')
      .and('have.attr', 'data-agent-outcome', 'blocked')
      .and('contain', 'Setup required')
      .and('contain', 'Open setup')
    cy.get('[data-testid="chat-action-preview"]').should('not.exist')
  })
})
