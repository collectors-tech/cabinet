describe('chats/assistant-workspace-agent-skills', () => {
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

  it('ASSISTANT-WORKSPACE-011/#1711 previews and applies Integrations Agent Skills from the side panel without secret echo', () => {
    bootstrapInventory()
    cy.intercept('POST', '/api/agent/skills/preview').as('agentSkillPreview')
    cy.intercept('POST', '/api/agent/skills/apply').as('agentSkillApply')
    openAssistantWorkspace()

    cy.get('[data-testid="shell-assistant-agent-skill-panel"]')
      .scrollIntoView()
      .should('exist')
    cy.get('[data-testid="shell-assistant-agent-skill-select"]').select(
      'cabinet.integrations.configure_provider',
      { force: true }
    )
    cy.get('[data-testid="shell-assistant-agent-skill-provider"]')
      .clear()
      .type('ebay', { force: true })
    cy.get('[data-testid="shell-assistant-agent-skill-setup-step"]')
      .clear()
      .type('oauth', { force: true })
    cy.get('[data-testid="shell-assistant-agent-skill-secret"]')
      .clear()
      .type('agent-secret-1711', { force: true })
    cy.get('[data-testid="shell-assistant-agent-skill-preview"]').click({
      force: true,
    })

    cy.wait('@agentSkillPreview').then(({ request, response }) => {
      expect(request.body.profile_id).to.eq('e2e-profile-001')
      expect(request.body.skill_id).to.eq(
        'cabinet.integrations.configure_provider'
      )
      expect(request.body.source_surface).to.eq(
        'settings.integrations.provider.card'
      )
      expect(request.body.source_channel).to.eq('in-app')
      expect(request.body.parameters.provider_id).to.eq('ebay')
      expect(request.body.parameters.provider_secret).to.eq('agent-secret-1711')
      expect(response?.statusCode).to.eq(200)
      expect(response?.body.confirmation_required).to.eq(true)
      expect(JSON.stringify(response?.body)).not.to.include('agent-secret-1711')
    })
    cy.get('[data-testid="shell-assistant-agent-skill-preview-card"]')
      .should('contain', 'cabinet.integrations.configure_provider')
      .and('contain', 'confirm-required')
      .and('not.contain', 'agent-secret-1711')

    cy.get('[data-testid="shell-assistant-agent-skill-apply"]').click({
      force: true,
    })
    cy.get('[data-testid="shell-assistant-apply-confirm-dialog"]').should(
      'be.visible'
    )
    cy.get('[data-testid="shell-assistant-apply-confirm-summary"]')
      .should('contain', 'cabinet.integrations.configure_provider')
      .and('contain', 'settings.integrations.provider.card')
      .and('not.contain', 'agent-secret-1711')
    cy.get('[data-testid="shell-assistant-apply-confirm"]').click()

    cy.wait('@agentSkillApply').then(({ request, response }) => {
      expect(request.body.profile_id).to.eq('e2e-profile-001')
      expect(request.body.skill_id).to.eq(
        'cabinet.integrations.configure_provider'
      )
      expect(request.body.confirm).to.eq(true)
      expect(request.body.source_surface).to.eq(
        'settings.integrations.provider.card'
      )
      expect(request.body.source_channel).to.eq('in-app')
      expect(response?.statusCode).to.eq(200)
      expect(response?.body.mutation_applied).to.eq(true)
      expect(response?.body.source_surface).to.eq(
        'settings.integrations.provider.card'
      )
      expect(response?.body.source_channel).to.eq('in-app')
      expect(response?.body.target.operation).to.eq(
        'integrations.provider.configure'
      )
      expect(response?.body.target.secret_redacted).to.eq(true)
      expect(response?.body.target.external_write_claimed).to.eq(false)
      expect(JSON.stringify(response?.body)).not.to.include('agent-secret-1711')
    })
    cy.get('[data-testid="shell-assistant-agent-skill-result"]')
      .should('contain', 'integrations.provider.configure')
      .and('contain', 'mutation: true')
      .and('contain', 'secret redacted: true')
      .and('not.contain', 'agent-secret-1711')
  })

  it('ASSISTANT-WORKSPACE-012/#1710 dispatches Market Watch Agent Skills with in-app source context', () => {
    bootstrapInventory()
    cy.intercept('POST', '/api/agent/skills/preview', (req) => {
      expect(req.body.profile_id).to.eq('e2e-profile-001')
      expect(req.body.skill_id).to.eq('cabinet.market_watch.run_watch')
      expect(req.body.source_surface).to.eq('market_watch.saved_watch.row')
      expect(req.body.source_channel).to.eq('in-app')
      expect(req.body.source_thread_id).to.be.a('string').and.not.eq('')
      expect(req.body.source_message_id).to.eq(
        'assistant-workspace-agent-skill'
      )
      expect(req.body.parameters.provider_id).to.eq('ebay')
      expect(req.body.parameters.watch_id).to.eq('watch-1710')
      req.reply({
        statusCode: 200,
        body: {
          skill_id: 'cabinet.market_watch.run_watch',
          status: 'available',
          safety_level: 'confirm-required',
          allowed: false,
          preview_only: true,
          mutation_applied: false,
          confirmation_required: true,
          blocker: 'confirmation_required',
          source_surface: 'market_watch.saved_watch.row',
          source_channel: 'in-app',
        },
      })
    }).as('marketWatchSkillPreview')
    cy.intercept('POST', '/api/agent/skills/apply', (req) => {
      expect(req.body.profile_id).to.eq('e2e-profile-001')
      expect(req.body.skill_id).to.eq('cabinet.market_watch.run_watch')
      expect(req.body.confirm).to.eq(true)
      expect(req.body.source_surface).to.eq('market_watch.saved_watch.row')
      expect(req.body.source_channel).to.eq('in-app')
      expect(req.body.parameters.provider_id).to.eq('ebay')
      expect(req.body.parameters.watch_id).to.eq('watch-1710')
      req.reply({
        statusCode: 200,
        body: {
          skill_id: 'cabinet.market_watch.run_watch',
          mutation_applied: true,
          source_surface: 'market_watch.saved_watch.row',
          source_channel: 'in-app',
          target: {
            operation: 'market_watch.watch.run',
            provider_id: 'ebay',
            external_write_claimed: false,
          },
        },
      })
    }).as('marketWatchSkillApply')
    openAssistantWorkspace()

    cy.get('[data-testid="shell-assistant-agent-skill-panel"]')
      .scrollIntoView()
      .should('exist')
    cy.get('[data-testid="shell-assistant-agent-skill-select"]').select(
      'cabinet.market_watch.run_watch',
      { force: true }
    )
    cy.get('[data-testid="shell-assistant-agent-skill-provider"]')
      .clear()
      .type('ebay', { force: true })
    cy.get('[data-testid="shell-assistant-agent-skill-setup-step"]')
      .clear()
      .type('watch-1710', { force: true })
    cy.get('[data-testid="shell-assistant-agent-skill-preview"]').click({
      force: true,
    })

    cy.wait('@marketWatchSkillPreview')
    cy.get('[data-testid="shell-assistant-agent-skill-preview-card"]')
      .should('contain', 'cabinet.market_watch.run_watch')
      .and('contain', 'confirm-required')
      .and('contain', 'confirmation_required')

    cy.get('[data-testid="shell-assistant-agent-skill-apply"]').click({
      force: true,
    })
    cy.get('[data-testid="shell-assistant-apply-confirm-summary"]')
      .should('contain', 'cabinet.market_watch.run_watch')
      .and('contain', 'market_watch.saved_watch.row')
    cy.get('[data-testid="shell-assistant-apply-confirm"]').click()

    cy.wait('@marketWatchSkillApply')
    cy.get('[data-testid="shell-assistant-agent-skill-result"]')
      .should('contain', 'market_watch.watch.run')
      .and('contain', 'mutation: true')
  })
})
