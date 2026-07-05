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

  it('ASSISTANT-WORKSPACE-012/#1710 dispatches Purchases Agent Skills with in-app source context', () => {
    bootstrapInventory()
    cy.intercept('POST', '/api/agent/skills/preview', (req) => {
      expect(req.body.profile_id).to.eq('e2e-profile-001')
      expect(req.body.skill_id).to.eq('cabinet.purchases.create_order')
      expect(req.body.source_surface).to.eq('purchases.inbox.capture')
      expect(req.body.source_channel).to.eq('in-app')
      expect(req.body.source_thread_id).to.be.a('string').and.not.eq('')
      expect(req.body.source_message_id).to.eq(
        'assistant-workspace-agent-skill'
      )
      expect(req.body.parameters.purchase_source).to.eq('ebay')
      expect(req.body.parameters.source).to.eq('ebay')
      expect(req.body.parameters.item_id).to.eq('item-1710')
      expect(req.body.parameters.tracking_number).to.eq(
        'https://example.test/orders/1710'
      )
      expect(req.body.parameters.source_url).to.eq(
        'https://example.test/orders/1710'
      )
      req.reply({
        statusCode: 200,
        body: {
          skill_id: 'cabinet.purchases.create_order',
          status: 'available',
          safety_level: 'confirm-required',
          allowed: false,
          preview_only: true,
          mutation_applied: false,
          confirmation_required: true,
          blocker: 'confirmation_required',
          source_surface: 'purchases.inbox.capture',
          source_channel: 'in-app',
        },
      })
    }).as('purchasesSkillPreview')
    cy.intercept('POST', '/api/agent/skills/apply', (req) => {
      expect(req.body.profile_id).to.eq('e2e-profile-001')
      expect(req.body.skill_id).to.eq('cabinet.purchases.create_order')
      expect(req.body.confirm).to.eq(true)
      expect(req.body.source_surface).to.eq('purchases.inbox.capture')
      expect(req.body.source_channel).to.eq('in-app')
      expect(req.body.parameters.purchase_source).to.eq('ebay')
      expect(req.body.parameters.source).to.eq('ebay')
      expect(req.body.parameters.item_id).to.eq('item-1710')
      expect(req.body.parameters.source_url).to.eq(
        'https://example.test/orders/1710'
      )
      req.reply({
        statusCode: 200,
        body: {
          skill_id: 'cabinet.purchases.create_order',
          mutation_applied: true,
          source_surface: 'purchases.inbox.capture',
          source_channel: 'in-app',
          target: {
            operation: 'purchases.order.create',
            purchase_persisted: true,
            external_write_claimed: false,
          },
        },
      })
    }).as('purchasesSkillApply')
    openAssistantWorkspace()

    cy.get('[data-testid="shell-assistant-agent-skill-panel"]')
      .scrollIntoView()
      .should('exist')
    cy.get('[data-testid="shell-assistant-agent-skill-select"]').select(
      'cabinet.purchases.create_order',
      { force: true }
    )
    cy.get('[data-testid="shell-assistant-agent-skill-provider"]')
      .clear()
      .type('ebay', { force: true })
    cy.get('[data-testid="shell-assistant-agent-skill-setup-step"]')
      .clear()
      .type('item-1710', { force: true })
    cy.get('[data-testid="shell-assistant-agent-skill-secret"]')
      .clear()
      .type('https://example.test/orders/1710', { force: true })
    cy.get('[data-testid="shell-assistant-agent-skill-preview"]').click({
      force: true,
    })

    cy.wait('@purchasesSkillPreview')
    cy.get('[data-testid="shell-assistant-agent-skill-preview-card"]')
      .should('contain', 'cabinet.purchases.create_order')
      .and('contain', 'confirm-required')
      .and('contain', 'confirmation_required')

    cy.get('[data-testid="shell-assistant-agent-skill-apply"]').click({
      force: true,
    })
    cy.get('[data-testid="shell-assistant-apply-confirm-summary"]')
      .should('contain', 'cabinet.purchases.create_order')
      .and('contain', 'purchases.inbox.capture')
    cy.get('[data-testid="shell-assistant-apply-confirm"]').click()

    cy.wait('@purchasesSkillApply')
    cy.get('[data-testid="shell-assistant-agent-skill-result"]')
      .should('contain', 'purchases.order.create')
      .and('contain', 'mutation: true')
  })

  it('ASSISTANT-WORKSPACE-013/#1709 dispatches Media Agent Skills with in-app source context', () => {
    bootstrapInventory()
    cy.intercept('POST', '/api/agent/skills/preview', (req) => {
      expect(req.body.profile_id).to.eq('e2e-profile-001')
      expect(req.body.skill_id).to.eq('cabinet.media.attach_to_item')
      expect(req.body.source_surface).to.eq('media.workspace.assignment')
      expect(req.body.source_channel).to.eq('in-app')
      expect(req.body.source_thread_id).to.be.a('string').and.not.eq('')
      expect(req.body.source_message_id).to.eq(
        'assistant-workspace-agent-skill'
      )
      expect(req.body.parameters.media_id).to.eq('media-1709')
      expect(req.body.parameters.item_id).to.eq('item-1709')
      expect(req.body.parameters.target_item).to.eq('item-1709')
      expect(req.body.parameters.notes).to.eq('preserve source note')
      req.reply({
        statusCode: 200,
        body: {
          skill_id: 'cabinet.media.attach_to_item',
          status: 'available',
          safety_level: 'confirm-required',
          allowed: false,
          preview_only: true,
          mutation_applied: false,
          confirmation_required: true,
          blocker: 'confirmation_required',
          source_surface: 'media.workspace.assignment',
          source_channel: 'in-app',
        },
      })
    }).as('mediaSkillPreview')
    cy.intercept('POST', '/api/agent/skills/apply', (req) => {
      expect(req.body.profile_id).to.eq('e2e-profile-001')
      expect(req.body.skill_id).to.eq('cabinet.media.attach_to_item')
      expect(req.body.confirm).to.eq(true)
      expect(req.body.source_surface).to.eq('media.workspace.assignment')
      expect(req.body.source_channel).to.eq('in-app')
      expect(req.body.parameters.media_id).to.eq('media-1709')
      expect(req.body.parameters.item_id).to.eq('item-1709')
      expect(req.body.parameters.notes).to.eq('preserve source note')
      req.reply({
        statusCode: 200,
        body: {
          skill_id: 'cabinet.media.attach_to_item',
          mutation_applied: true,
          source_surface: 'media.workspace.assignment',
          source_channel: 'in-app',
          target: {
            operation: 'media.attach_to_item',
            media_id: 'media-1709',
            item_id: 'item-1709',
            attachment_persisted: true,
            provenance_preserved: true,
            external_write_claimed: false,
          },
        },
      })
    }).as('mediaSkillApply')
    openAssistantWorkspace()

    cy.get('[data-testid="shell-assistant-agent-skill-panel"]')
      .scrollIntoView()
      .should('exist')
    cy.get('[data-testid="shell-assistant-agent-skill-select"]').select(
      'cabinet.media.attach_to_item',
      { force: true }
    )
    cy.get('[data-testid="shell-assistant-agent-skill-provider"]')
      .clear()
      .type('media-1709', { force: true })
    cy.get('[data-testid="shell-assistant-agent-skill-setup-step"]')
      .clear()
      .type('item-1709', { force: true })
    cy.get('[data-testid="shell-assistant-agent-skill-secret"]')
      .clear()
      .type('preserve source note', { force: true })
    cy.get('[data-testid="shell-assistant-agent-skill-preview"]').click({
      force: true,
    })

    cy.wait('@mediaSkillPreview')
    cy.get('[data-testid="shell-assistant-agent-skill-preview-card"]')
      .should('contain', 'cabinet.media.attach_to_item')
      .and('contain', 'confirm-required')
      .and('contain', 'confirmation_required')

    cy.get('[data-testid="shell-assistant-agent-skill-apply"]').click({
      force: true,
    })
    cy.get('[data-testid="shell-assistant-apply-confirm-summary"]')
      .should('contain', 'cabinet.media.attach_to_item')
      .and('contain', 'media.workspace.assignment')
    cy.get('[data-testid="shell-assistant-apply-confirm"]').click()

    cy.wait('@mediaSkillApply')
    cy.get('[data-testid="shell-assistant-agent-skill-result"]')
      .should('contain', 'media.attach_to_item')
      .and('contain', 'mutation: true')
  })

  it('ASSISTANT-WORKSPACE-013/#1709 dispatches Discoveries Agent Skills with in-app source context', () => {
    bootstrapInventory()
    cy.intercept('POST', '/api/agent/skills/preview', (req) => {
      expect(req.body.profile_id).to.eq('e2e-profile-001')
      expect(req.body.skill_id).to.eq('cabinet.discoveries.send_to_wishlist')
      expect(req.body.source_surface).to.eq('discoveries.result.card')
      expect(req.body.source_channel).to.eq('in-app')
      expect(req.body.source_thread_id).to.be.a('string').and.not.eq('')
      expect(req.body.source_message_id).to.eq(
        'assistant-workspace-agent-skill'
      )
      expect(req.body.parameters.provider_id).to.eq('ebay')
      expect(req.body.parameters.result_id).to.eq('disc-1709')
      expect(req.body.parameters.candidate_id).to.eq('disc-1709')
      expect(req.body.parameters.notes).to.eq('send to wishlist')
      req.reply({
        statusCode: 200,
        body: {
          skill_id: 'cabinet.discoveries.send_to_wishlist',
          status: 'available',
          safety_level: 'confirm-required',
          allowed: false,
          preview_only: true,
          mutation_applied: false,
          confirmation_required: true,
          blocker: 'confirmation_required',
          source_surface: 'discoveries.result.card',
          source_channel: 'in-app',
        },
      })
    }).as('discoveriesSkillPreview')
    cy.intercept('POST', '/api/agent/skills/apply', (req) => {
      expect(req.body.profile_id).to.eq('e2e-profile-001')
      expect(req.body.skill_id).to.eq('cabinet.discoveries.send_to_wishlist')
      expect(req.body.confirm).to.eq(true)
      expect(req.body.source_surface).to.eq('discoveries.result.card')
      expect(req.body.source_channel).to.eq('in-app')
      expect(req.body.parameters.provider_id).to.eq('ebay')
      expect(req.body.parameters.result_id).to.eq('disc-1709')
      expect(req.body.parameters.notes).to.eq('send to wishlist')
      req.reply({
        statusCode: 200,
        body: {
          skill_id: 'cabinet.discoveries.send_to_wishlist',
          mutation_applied: true,
          source_surface: 'discoveries.result.card',
          source_channel: 'in-app',
          target: {
            operation: 'discoveries.send_to_wishlist',
            provider_id: 'ebay',
            result_id: 'disc-1709',
            discovery_persisted: true,
            provenance_preserved: true,
            external_write_claimed: false,
          },
        },
      })
    }).as('discoveriesSkillApply')
    openAssistantWorkspace()

    cy.get('[data-testid="shell-assistant-agent-skill-panel"]')
      .scrollIntoView()
      .should('exist')
    cy.get('[data-testid="shell-assistant-agent-skill-select"]').select(
      'cabinet.discoveries.send_to_wishlist',
      { force: true }
    )
    cy.get('[data-testid="shell-assistant-agent-skill-provider"]')
      .clear()
      .type('ebay', { force: true })
    cy.get('[data-testid="shell-assistant-agent-skill-setup-step"]')
      .clear()
      .type('disc-1709', { force: true })
    cy.get('[data-testid="shell-assistant-agent-skill-secret"]')
      .clear()
      .type('send to wishlist', { force: true })
    cy.get('[data-testid="shell-assistant-agent-skill-preview"]').click({
      force: true,
    })

    cy.wait('@discoveriesSkillPreview')
    cy.get('[data-testid="shell-assistant-agent-skill-preview-card"]')
      .should('contain', 'cabinet.discoveries.send_to_wishlist')
      .and('contain', 'confirm-required')
      .and('contain', 'confirmation_required')

    cy.get('[data-testid="shell-assistant-agent-skill-apply"]').click({
      force: true,
    })
    cy.get('[data-testid="shell-assistant-apply-confirm-summary"]')
      .should('contain', 'cabinet.discoveries.send_to_wishlist')
      .and('contain', 'discoveries.result.card')
    cy.get('[data-testid="shell-assistant-apply-confirm"]').click()

    cy.wait('@discoveriesSkillApply')
    cy.get('[data-testid="shell-assistant-agent-skill-result"]')
      .should('contain', 'discoveries.send_to_wishlist')
      .and('contain', 'mutation: true')
  })

  it('ASSISTANT-WORKSPACE-014/#1708 dispatches Wishlist Agent Skills with in-app source context', () => {
    bootstrapInventory()
    cy.intercept('POST', '/api/agent/skills/preview', (req) => {
      expect(req.body.profile_id).to.eq('e2e-profile-001')
      expect(req.body.skill_id).to.eq('cabinet.wishlist.create_entry')
      expect(req.body.source_surface).to.eq('wishlist.intent.capture')
      expect(req.body.source_channel).to.eq('in-app')
      expect(req.body.source_thread_id).to.be.a('string').and.not.eq('')
      expect(req.body.source_message_id).to.eq(
        'assistant-workspace-agent-skill'
      )
      expect(req.body.parameters.title).to.eq('AFX Mega G+ wishlist')
      expect(req.body.parameters.target_price).to.eq('45')
      expect(req.body.parameters.source_url).to.eq(
        'https://example.test/wishlist/1708'
      )
      expect(req.body.parameters.notes).to.eq(
        'https://example.test/wishlist/1708'
      )
      req.reply({
        statusCode: 200,
        body: {
          skill_id: 'cabinet.wishlist.create_entry',
          status: 'available',
          safety_level: 'confirm-required',
          allowed: false,
          preview_only: true,
          mutation_applied: false,
          confirmation_required: true,
          blocker: 'confirmation_required',
          source_surface: 'wishlist.intent.capture',
          source_channel: 'in-app',
        },
      })
    }).as('wishlistSkillPreview')
    cy.intercept('POST', '/api/agent/skills/apply', (req) => {
      expect(req.body.profile_id).to.eq('e2e-profile-001')
      expect(req.body.skill_id).to.eq('cabinet.wishlist.create_entry')
      expect(req.body.confirm).to.eq(true)
      expect(req.body.source_surface).to.eq('wishlist.intent.capture')
      expect(req.body.source_channel).to.eq('in-app')
      expect(req.body.parameters.title).to.eq('AFX Mega G+ wishlist')
      expect(req.body.parameters.target_price).to.eq('45')
      expect(req.body.parameters.source_url).to.eq(
        'https://example.test/wishlist/1708'
      )
      req.reply({
        statusCode: 200,
        body: {
          skill_id: 'cabinet.wishlist.create_entry',
          mutation_applied: true,
          source_surface: 'wishlist.intent.capture',
          source_channel: 'in-app',
          target: {
            operation: 'wishlist.entry.create',
            wishlist_entry_id: 'wish-1708',
            provenance_preserved: true,
            external_write_claimed: false,
          },
        },
      })
    }).as('wishlistSkillApply')
    openAssistantWorkspace()

    cy.get('[data-testid="shell-assistant-agent-skill-panel"]')
      .scrollIntoView()
      .should('exist')
    cy.get('[data-testid="shell-assistant-agent-skill-select"]').select(
      'cabinet.wishlist.create_entry',
      { force: true }
    )
    cy.get('[data-testid="shell-assistant-agent-skill-provider"]')
      .clear()
      .type('AFX Mega G+ wishlist', { force: true })
    cy.get('[data-testid="shell-assistant-agent-skill-setup-step"]')
      .clear()
      .type('45', { force: true })
    cy.get('[data-testid="shell-assistant-agent-skill-secret"]')
      .clear()
      .type('https://example.test/wishlist/1708', { force: true })
    cy.get('[data-testid="shell-assistant-agent-skill-preview"]').click({
      force: true,
    })

    cy.wait('@wishlistSkillPreview')
    cy.get('[data-testid="shell-assistant-agent-skill-preview-card"]')
      .should('contain', 'cabinet.wishlist.create_entry')
      .and('contain', 'confirm-required')
      .and('contain', 'confirmation_required')

    cy.get('[data-testid="shell-assistant-agent-skill-apply"]').click({
      force: true,
    })
    cy.get('[data-testid="shell-assistant-apply-confirm-summary"]')
      .should('contain', 'cabinet.wishlist.create_entry')
      .and('contain', 'wishlist.intent.capture')
    cy.get('[data-testid="shell-assistant-apply-confirm"]').click()

    cy.wait('@wishlistSkillApply')
    cy.get('[data-testid="shell-assistant-agent-skill-result"]')
      .should('contain', 'wishlist.entry.create')
      .and('contain', 'mutation: true')
  })

  it('ASSISTANT-WORKSPACE-014/#1708 dispatches Collections Agent Skills with in-app source context', () => {
    bootstrapInventory()
    cy.intercept('POST', '/api/agent/skills/preview', (req) => {
      expect(req.body.profile_id).to.eq('e2e-profile-001')
      expect(req.body.skill_id).to.eq('cabinet.collections.assign_item')
      expect(req.body.source_surface).to.eq('collections.workspace.assignment')
      expect(req.body.source_channel).to.eq('in-app')
      expect(req.body.source_thread_id).to.be.a('string').and.not.eq('')
      expect(req.body.source_message_id).to.eq(
        'assistant-workspace-agent-skill'
      )
      expect(req.body.parameters.collection_id).to.eq('collection-1708')
      expect(req.body.parameters.collection_name).to.eq('collection-1708')
      expect(req.body.parameters.item_id).to.eq('item-1708')
      expect(req.body.parameters.notes).to.eq('assign from side panel')
      req.reply({
        statusCode: 200,
        body: {
          skill_id: 'cabinet.collections.assign_item',
          status: 'available',
          safety_level: 'confirm-required',
          allowed: false,
          preview_only: true,
          mutation_applied: false,
          confirmation_required: true,
          blocker: 'confirmation_required',
          source_surface: 'collections.workspace.assignment',
          source_channel: 'in-app',
        },
      })
    }).as('collectionsSkillPreview')
    cy.intercept('POST', '/api/agent/skills/apply', (req) => {
      expect(req.body.profile_id).to.eq('e2e-profile-001')
      expect(req.body.skill_id).to.eq('cabinet.collections.assign_item')
      expect(req.body.confirm).to.eq(true)
      expect(req.body.source_surface).to.eq('collections.workspace.assignment')
      expect(req.body.source_channel).to.eq('in-app')
      expect(req.body.parameters.collection_id).to.eq('collection-1708')
      expect(req.body.parameters.item_id).to.eq('item-1708')
      expect(req.body.parameters.notes).to.eq('assign from side panel')
      req.reply({
        statusCode: 200,
        body: {
          skill_id: 'cabinet.collections.assign_item',
          mutation_applied: true,
          source_surface: 'collections.workspace.assignment',
          source_channel: 'in-app',
          target: {
            operation: 'collections.item.assign',
            collection_id: 'collection-1708',
            item_id: 'item-1708',
            assignment_persisted: true,
            external_write_claimed: false,
          },
        },
      })
    }).as('collectionsSkillApply')
    openAssistantWorkspace()

    cy.get('[data-testid="shell-assistant-agent-skill-panel"]')
      .scrollIntoView()
      .should('exist')
    cy.get('[data-testid="shell-assistant-agent-skill-select"]').select(
      'cabinet.collections.assign_item',
      { force: true }
    )
    cy.get('[data-testid="shell-assistant-agent-skill-provider"]')
      .clear()
      .type('collection-1708', { force: true })
    cy.get('[data-testid="shell-assistant-agent-skill-setup-step"]')
      .clear()
      .type('item-1708', { force: true })
    cy.get('[data-testid="shell-assistant-agent-skill-secret"]')
      .clear()
      .type('assign from side panel', { force: true })
    cy.get('[data-testid="shell-assistant-agent-skill-preview"]').click({
      force: true,
    })

    cy.wait('@collectionsSkillPreview')
    cy.get('[data-testid="shell-assistant-agent-skill-preview-card"]')
      .should('contain', 'cabinet.collections.assign_item')
      .and('contain', 'confirm-required')
      .and('contain', 'confirmation_required')

    cy.get('[data-testid="shell-assistant-agent-skill-apply"]').click({
      force: true,
    })
    cy.get('[data-testid="shell-assistant-apply-confirm-summary"]')
      .should('contain', 'cabinet.collections.assign_item')
      .and('contain', 'collections.workspace.assignment')
    cy.get('[data-testid="shell-assistant-apply-confirm"]').click()

    cy.wait('@collectionsSkillApply')
    cy.get('[data-testid="shell-assistant-agent-skill-result"]')
      .should('contain', 'collections.item.assign')
      .and('contain', 'mutation: true')
  })
})
