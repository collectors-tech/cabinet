describe('chats/assistant-workspace-dashboard-summary', () => {
  function bootstrapDashboard() {
    cy.viewport(1400, 900)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/dashboard/',
    })
    cy.location('pathname', { timeout: 15000 }).should(
      'match',
      /^\/dashboard\/?$/
    )
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

  it('AGENT-DASHBOARD-SUMMARY-001/#1942 invokes Dashboard summary from the side-panel Agent Skill dispatcher', () => {
    bootstrapDashboard()
    cy.intercept('POST', '/api/agent/skills/preview').as('agentSkillPreview')
    cy.intercept('POST', '/api/agent/skills/apply').as('agentSkillApply')
    openAssistantWorkspace()

    cy.get('[data-testid="shell-assistant-agent-skill-panel"]')
      .scrollIntoView()
      .should('exist')
    cy.get('[data-testid="shell-assistant-agent-skill-select"]').select(
      'cabinet.dashboard.summarise_activity',
      { force: true }
    )
    cy.get('[data-testid="shell-assistant-agent-skill-provider"]')
      .clear()
      .type('today', { force: true })
    cy.get('[data-testid="shell-assistant-agent-skill-setup-step"]')
      .clear()
      .type('workspace-dashboard', { force: true })
    cy.get('[data-testid="shell-assistant-agent-skill-secret"]')
      .clear()
      .type('side-panel note', { force: true })
    cy.get('[data-testid="shell-assistant-agent-skill-preview"]').click({
      force: true,
    })

    cy.wait('@agentSkillPreview').then(({ request, response }) => {
      expect(request.body.profile_id).to.eq('e2e-profile-001')
      expect(request.body.skill_id).to.eq(
        'cabinet.dashboard.summarise_activity'
      )
      expect(request.body.source_surface).to.eq('dashboard.home')
      expect(request.body.source_channel).to.eq('in-app')
      expect(request.body.parameters.window).to.eq('today')
      expect(request.body.parameters.workspace_id).to.eq('workspace-dashboard')
      expect(response?.statusCode).to.eq(200)
      expect(response?.body.confirmation_required).to.eq(false)
      expect(response?.body.mutation_applied).to.eq(false)
    })
    cy.get('[data-testid="shell-assistant-agent-skill-preview-card"]')
      .should('contain', 'cabinet.dashboard.summarise_activity')
      .and('contain', 'read-only')
      .and('contain', 'ready')

    cy.get('[data-testid="shell-assistant-agent-skill-apply"]').click({
      force: true,
    })
    cy.get('[data-testid="shell-assistant-apply-confirm-dialog"]').should(
      'be.visible'
    )
    cy.get('[data-testid="shell-assistant-apply-confirm-summary"]')
      .should('contain', 'cabinet.dashboard.summarise_activity')
      .and('contain', 'dashboard.home')
    cy.get('[data-testid="shell-assistant-apply-confirm"]').click()

    cy.wait('@agentSkillApply').then(({ request, response }) => {
      expect(request.body.profile_id).to.eq('e2e-profile-001')
      expect(request.body.skill_id).to.eq(
        'cabinet.dashboard.summarise_activity'
      )
      expect(request.body.confirm).to.eq(true)
      expect(request.body.source_surface).to.eq('dashboard.home')
      expect(request.body.source_channel).to.eq('in-app')
      expect(response?.statusCode).to.eq(200)
      expect(response?.body.mutation_applied).to.eq(false)
      expect(response?.body.source_surface).to.eq('dashboard.home')
      expect(response?.body.source_channel).to.eq('in-app')
      expect(response?.body.target.operation).to.eq('dashboard.activity.summary')
      expect(response?.body.target.read_only).to.eq(true)
      expect(response?.body.target.time_window.requested_window).to.eq('today')
      expect(JSON.stringify(response?.body)).not.to.include('side-panel note')
    })
    cy.get('[data-testid="shell-assistant-agent-skill-result"]')
      .should('contain', 'dashboard.activity.summary')
      .and('contain', 'mutation: false')
  })
})
