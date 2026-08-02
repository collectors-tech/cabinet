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

  it('ASSISTANT-WORKSPACE-016/#2011 dispatches Settings Profile Agent Skills with in-app source context', () => {
    bootstrapInventory()
    cy.intercept('POST', '/api/agent/skills/preview', (req) => {
      expect(req.body.profile_id).to.eq('e2e-profile-001')
      expect(req.body.skill_id).to.eq('cabinet.settings.update_profile')
      expect(req.body.source_surface).to.eq('settings.profile.form')
      expect(req.body.source_channel).to.eq('in-app')
      expect(req.body.source_thread_id).to.be.a('string').and.not.eq('')
      expect(req.body.source_message_id).to.eq(
        'assistant-workspace-agent-skill'
      )
      expect(req.body.parameters.settings_profile).to.deep.eq({
        display_currency: 'AUD',
        timezone: 'Australia/Sydney',
        profile_private_note: 'keep this private',
      })
      req.reply({
        statusCode: 200,
        body: {
          skill_id: 'cabinet.settings.update_profile',
          status: 'available',
          safety_level: 'confirm-required',
          allowed: false,
          preview_only: true,
          mutation_applied: false,
          confirmation_required: true,
          blocker: 'confirmation_required',
          source_surface: 'settings.profile.form',
          source_channel: 'in-app',
        },
      })
    }).as('settingsProfileSkillPreview')
    cy.intercept('POST', '/api/agent/skills/apply', (req) => {
      expect(req.body.profile_id).to.eq('e2e-profile-001')
      expect(req.body.skill_id).to.eq('cabinet.settings.update_profile')
      expect(req.body.confirm).to.eq(true)
      expect(req.body.source_surface).to.eq('settings.profile.form')
      expect(req.body.source_channel).to.eq('in-app')
      expect(req.body.parameters.settings_profile).to.deep.eq({
        display_currency: 'AUD',
        timezone: 'Australia/Sydney',
        profile_private_note: 'keep this private',
      })
      req.reply({
        statusCode: 200,
        body: {
          skill_id: 'cabinet.settings.update_profile',
          mutation_applied: true,
          source_surface: 'settings.profile.form',
          source_channel: 'in-app',
          target: {
            operation: 'settings.profile.update',
            settings_persisted: [
              'display_currency',
              'timezone',
              'profile_private_note',
            ],
            external_write_claimed: false,
          },
        },
      })
    }).as('settingsProfileSkillApply')
    openAssistantWorkspace()

    cy.get('[data-testid="shell-assistant-agent-skill-panel"]')
      .scrollIntoView()
      .should('exist')
    cy.get('[data-testid="shell-assistant-agent-skill-select"]').select(
      'cabinet.settings.update_profile',
      { force: true }
    )
    cy.get('[data-testid="shell-assistant-agent-skill-provider"]')
      .clear()
      .type('AUD', { force: true })
    cy.get('[data-testid="shell-assistant-agent-skill-setup-step"]')
      .clear()
      .type('Australia/Sydney', { force: true })
    cy.get('[data-testid="shell-assistant-agent-skill-secret"]')
      .clear()
      .type('keep this private', { force: true })
    cy.get('[data-testid="shell-assistant-agent-skill-preview"]').click({
      force: true,
    })

    cy.wait('@settingsProfileSkillPreview')
    cy.get('[data-testid="shell-assistant-agent-skill-preview-card"]')
      .should('contain', 'cabinet.settings.update_profile')
      .and('contain', 'confirm-required')
      .and('contain', 'confirmation_required')
      .and('not.contain', 'keep this private')

    cy.get('[data-testid="shell-assistant-agent-skill-apply"]').click({
      force: true,
    })
    cy.get('[data-testid="shell-assistant-apply-confirm-summary"]')
      .should('contain', 'cabinet.settings.update_profile')
      .and('contain', 'settings.profile.form')
      .and('not.contain', 'keep this private')
    cy.get('[data-testid="shell-assistant-apply-confirm"]').click()

    cy.wait('@settingsProfileSkillApply')
    cy.get('[data-testid="shell-assistant-agent-skill-result"]')
      .should('contain', 'settings.profile.update')
      .and('contain', 'mutation: true')
      .and('not.contain', 'keep this private')
  })

  it('ASSISTANT-WORKSPACE-016/#2013 dispatches Settings Appearance Agent Skills with in-app source context', () => {
    bootstrapInventory()
    cy.intercept('POST', '/api/agent/skills/preview', (req) => {
      expect(req.body.profile_id).to.eq('e2e-profile-001')
      expect(req.body.skill_id).to.eq('cabinet.settings.update_appearance')
      expect(req.body.source_surface).to.eq('settings.appearance.form')
      expect(req.body.source_channel).to.eq('in-app')
      expect(req.body.source_thread_id).to.be.a('string').and.not.eq('')
      expect(req.body.source_message_id).to.eq(
        'assistant-workspace-agent-skill'
      )
      expect(req.body.parameters).to.deep.eq({
        setting_key: 'appearance.theme',
        setting_scope: 'appearance',
        setting_value: 'dark',
      })
      req.reply({
        statusCode: 200,
        body: {
          skill_id: 'cabinet.settings.update_appearance',
          status: 'available',
          safety_level: 'confirm-required',
          allowed: false,
          preview_only: true,
          mutation_applied: false,
          confirmation_required: true,
          blocker: 'confirmation_required',
          source_surface: 'settings.appearance.form',
          source_channel: 'in-app',
        },
      })
    }).as('settingsAppearanceSkillPreview')
    cy.intercept('POST', '/api/agent/skills/apply', (req) => {
      expect(req.body.profile_id).to.eq('e2e-profile-001')
      expect(req.body.skill_id).to.eq('cabinet.settings.update_appearance')
      expect(req.body.confirm).to.eq(true)
      expect(req.body.source_surface).to.eq('settings.appearance.form')
      expect(req.body.source_channel).to.eq('in-app')
      expect(req.body.parameters).to.deep.eq({
        setting_key: 'appearance.theme',
        setting_scope: 'appearance',
        setting_value: 'dark',
      })
      req.reply({
        statusCode: 200,
        body: {
          skill_id: 'cabinet.settings.update_appearance',
          mutation_applied: true,
          source_surface: 'settings.appearance.form',
          source_channel: 'in-app',
          target: {
            operation: 'settings.appearance.update',
            setting_key: 'appearance.theme',
            settings_persisted: ['appearance.theme'],
            external_write_claimed: false,
          },
        },
      })
    }).as('settingsAppearanceSkillApply')
    openAssistantWorkspace()

    cy.get('[data-testid="shell-assistant-agent-skill-panel"]')
      .scrollIntoView()
      .should('exist')
    cy.get('[data-testid="shell-assistant-agent-skill-select"]').select(
      'cabinet.settings.update_appearance',
      { force: true }
    )
    cy.get('[data-testid="shell-assistant-agent-skill-provider"]')
      .clear()
      .type('appearance.theme', { force: true })
    cy.get('[data-testid="shell-assistant-agent-skill-setup-step"]')
      .clear()
      .type('appearance', { force: true })
    cy.get('[data-testid="shell-assistant-agent-skill-secret"]')
      .clear()
      .type('dark', { force: true })
    cy.get('[data-testid="shell-assistant-agent-skill-preview"]').click({
      force: true,
    })

    cy.wait('@settingsAppearanceSkillPreview')
    cy.get('[data-testid="shell-assistant-agent-skill-preview-card"]')
      .should('contain', 'cabinet.settings.update_appearance')
      .and('contain', 'confirm-required')
      .and('contain', 'confirmation_required')
      .and('not.contain', 'dark')

    cy.get('[data-testid="shell-assistant-agent-skill-apply"]').click({
      force: true,
    })
    cy.get('[data-testid="shell-assistant-apply-confirm-summary"]')
      .should('contain', 'cabinet.settings.update_appearance')
      .and('contain', 'settings.appearance.form')
      .and('not.contain', 'dark')
    cy.get('[data-testid="shell-assistant-apply-confirm"]').click()

    cy.wait('@settingsAppearanceSkillApply')
    cy.get('[data-testid="shell-assistant-agent-skill-result"]')
      .should('contain', 'settings.appearance.update')
      .and('contain', 'mutation: true')
      .and('not.contain', 'dark')
  })

  it('ASSISTANT-WORKSPACE-016/#2017 dispatches Settings Account Agent Skills with in-app source context', () => {
    bootstrapInventory()
    cy.intercept('POST', '/api/agent/skills/preview', (req) => {
      expect(req.body.profile_id).to.eq('e2e-profile-001')
      expect(req.body.skill_id).to.eq('cabinet.settings.update_account')
      expect(req.body.source_surface).to.eq('settings.account.form')
      expect(req.body.source_channel).to.eq('in-app')
      expect(req.body.source_thread_id).to.be.a('string').and.not.eq('')
      expect(req.body.source_message_id).to.eq(
        'assistant-workspace-agent-skill'
      )
      expect(req.body.parameters.settings_account).to.deep.eq({
        account_email: 'agent-account@example.com',
        locale: 'en-AU',
        account_private_note: 'account secret note',
      })
      req.reply({
        statusCode: 200,
        body: {
          skill_id: 'cabinet.settings.update_account',
          status: 'available',
          safety_level: 'confirm-required',
          allowed: false,
          preview_only: true,
          mutation_applied: false,
          confirmation_required: true,
          blocker: 'confirmation_required',
          source_surface: 'settings.account.form',
          source_channel: 'in-app',
        },
      })
    }).as('settingsAccountSkillPreview')
    cy.intercept('POST', '/api/agent/skills/apply', (req) => {
      expect(req.body.profile_id).to.eq('e2e-profile-001')
      expect(req.body.skill_id).to.eq('cabinet.settings.update_account')
      expect(req.body.confirm).to.eq(true)
      expect(req.body.source_surface).to.eq('settings.account.form')
      expect(req.body.source_channel).to.eq('in-app')
      expect(req.body.parameters.settings_account).to.deep.eq({
        account_email: 'agent-account@example.com',
        locale: 'en-AU',
        account_private_note: 'account secret note',
      })
      req.reply({
        statusCode: 200,
        body: {
          skill_id: 'cabinet.settings.update_account',
          mutation_applied: true,
          source_surface: 'settings.account.form',
          source_channel: 'in-app',
          target: {
            operation: 'settings.account.update',
            settings_persisted: [
              'account_email',
              'locale',
              'account_private_note',
            ],
            external_write_claimed: false,
          },
        },
      })
    }).as('settingsAccountSkillApply')
    openAssistantWorkspace()

    cy.get('[data-testid="shell-assistant-agent-skill-panel"]')
      .scrollIntoView()
      .should('exist')
    cy.get('[data-testid="shell-assistant-agent-skill-select"]').select(
      'cabinet.settings.update_account',
      { force: true }
    )
    cy.get('[data-testid="shell-assistant-agent-skill-provider"]')
      .clear()
      .type('agent-account@example.com', { force: true })
    cy.get('[data-testid="shell-assistant-agent-skill-setup-step"]')
      .clear()
      .type('en-AU', { force: true })
    cy.get('[data-testid="shell-assistant-agent-skill-secret"]')
      .clear()
      .type('account secret note', { force: true })
    cy.get('[data-testid="shell-assistant-agent-skill-preview"]').click({
      force: true,
    })

    cy.wait('@settingsAccountSkillPreview')
    cy.get('[data-testid="shell-assistant-agent-skill-preview-card"]')
      .should('contain', 'cabinet.settings.update_account')
      .and('contain', 'confirm-required')
      .and('contain', 'confirmation_required')
      .and('not.contain', 'account secret note')

    cy.get('[data-testid="shell-assistant-agent-skill-apply"]').click({
      force: true,
    })
    cy.get('[data-testid="shell-assistant-apply-confirm-summary"]')
      .should('contain', 'cabinet.settings.update_account')
      .and('contain', 'settings.account.form')
      .and('not.contain', 'account secret note')
    cy.get('[data-testid="shell-assistant-apply-confirm"]').click()

    cy.wait('@settingsAccountSkillApply')
    cy.get('[data-testid="shell-assistant-agent-skill-result"]')
      .should('contain', 'settings.account.update')
      .and('contain', 'mutation: true')
      .and('not.contain', 'account secret note')
  })

  it('ASSISTANT-WORKSPACE-016/#2019 dispatches Settings Storage Agent Skills with in-app source context', () => {
    bootstrapInventory()
    cy.intercept('POST', '/api/agent/skills/preview', (req) => {
      expect(req.body.profile_id).to.eq('e2e-profile-001')
      expect(req.body.skill_id).to.eq('cabinet.storage.configure_backup')
      expect(req.body.source_surface).to.eq('settings.storage.backup')
      expect(req.body.source_channel).to.eq('in-app')
      expect(req.body.source_thread_id).to.be.a('string').and.not.eq('')
      expect(req.body.source_message_id).to.eq(
        'assistant-workspace-agent-skill'
      )
      expect(req.body.parameters.backup_target).to.eq('cabinet-backups')
      expect(req.body.parameters.backup_schedule).to.eq('weekly')
      expect(req.body.parameters.storage_note).to.eq('storage private note')
      req.reply({
        statusCode: 200,
        body: {
          skill_id: 'cabinet.storage.configure_backup',
          status: 'available',
          safety_level: 'confirm-required',
          allowed: false,
          preview_only: true,
          mutation_applied: false,
          confirmation_required: true,
          blocker: 'confirmation_required',
          source_surface: 'settings.storage.backup',
          source_channel: 'in-app',
        },
      })
    }).as('settingsStorageSkillPreview')
    cy.intercept('POST', '/api/agent/skills/apply', (req) => {
      expect(req.body.profile_id).to.eq('e2e-profile-001')
      expect(req.body.skill_id).to.eq('cabinet.storage.configure_backup')
      expect(req.body.confirm).to.eq(true)
      expect(req.body.source_surface).to.eq('settings.storage.backup')
      expect(req.body.source_channel).to.eq('in-app')
      expect(req.body.parameters.backup_target).to.eq('cabinet-backups')
      expect(req.body.parameters.backup_schedule).to.eq('weekly')
      expect(req.body.parameters.storage_note).to.eq('storage private note')
      req.reply({
        statusCode: 200,
        body: {
          skill_id: 'cabinet.storage.configure_backup',
          mutation_applied: true,
          source_surface: 'settings.storage.backup',
          source_channel: 'in-app',
          target: {
            operation: 'storage.backup.configure',
            backup_target_redacted: true,
            external_write_claimed: false,
          },
        },
      })
    }).as('settingsStorageSkillApply')
    openAssistantWorkspace()

    cy.get('[data-testid="shell-assistant-agent-skill-panel"]')
      .scrollIntoView()
      .should('exist')
    cy.get('[data-testid="shell-assistant-agent-skill-select"]').select(
      'cabinet.storage.configure_backup',
      { force: true }
    )
    cy.get('[data-testid="shell-assistant-agent-skill-provider"]')
      .clear()
      .type('cabinet-backups', { force: true })
    cy.get('[data-testid="shell-assistant-agent-skill-setup-step"]')
      .clear()
      .type('weekly', { force: true })
    cy.get('[data-testid="shell-assistant-agent-skill-secret"]')
      .clear()
      .type('storage private note', { force: true })
    cy.get('[data-testid="shell-assistant-agent-skill-preview"]').click({
      force: true,
    })

    cy.wait('@settingsStorageSkillPreview')
    cy.get('[data-testid="shell-assistant-agent-skill-preview-card"]')
      .should('contain', 'cabinet.storage.configure_backup')
      .and('contain', 'confirm-required')
      .and('contain', 'confirmation_required')
      .and('not.contain', 'storage private note')

    cy.get('[data-testid="shell-assistant-agent-skill-apply"]').click({
      force: true,
    })
    cy.get('[data-testid="shell-assistant-apply-confirm-summary"]')
      .should('contain', 'cabinet.storage.configure_backup')
      .and('contain', 'settings.storage.backup')
      .and('not.contain', 'storage private note')
    cy.get('[data-testid="shell-assistant-apply-confirm"]').click()

    cy.wait('@settingsStorageSkillApply')
    cy.get('[data-testid="shell-assistant-agent-skill-result"]')
      .should('contain', 'storage.backup.configure')
      .and('contain', 'mutation: true')
      .and('not.contain', 'storage private note')
      .and('not.contain', 'C:\\')
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

  it('ASSISTANT-WORKSPACE-015/#1707 dispatches Inventory Agent Skills with in-app source context', () => {
    bootstrapInventory()
    cy.intercept('POST', '/api/agent/skills/preview', (req) => {
      expect(req.body.profile_id).to.eq('e2e-profile-001')
      expect(req.body.skill_id).to.eq('cabinet.inventory.create_item')
      expect(req.body.source_surface).to.eq('inventory.quick-create')
      expect(req.body.source_channel).to.eq('in-app')
      expect(req.body.source_thread_id).to.be.a('string').and.not.eq('')
      expect(req.body.source_message_id).to.eq(
        'assistant-workspace-agent-skill'
      )
      expect(req.body.parameters.part_number).to.eq('INV-1707-SP')
      expect(req.body.parameters.title).to.eq(
        'Inventory side panel dispatch proof'
      )
      expect(req.body.parameters.source_url).to.eq(
        'https://example.test/inventory/1707'
      )
      expect(req.body.parameters.notes).to.eq(
        'https://example.test/inventory/1707'
      )
      req.reply({
        statusCode: 200,
        body: {
          skill_id: 'cabinet.inventory.create_item',
          status: 'available',
          safety_level: 'confirm-required',
          allowed: false,
          preview_only: true,
          mutation_applied: false,
          confirmation_required: true,
          blocker: 'confirmation_required',
          source_surface: 'inventory.quick-create',
          source_channel: 'in-app',
        },
      })
    }).as('inventorySkillPreview')
    cy.intercept('POST', '/api/agent/skills/apply', (req) => {
      expect(req.body.profile_id).to.eq('e2e-profile-001')
      expect(req.body.skill_id).to.eq('cabinet.inventory.create_item')
      expect(req.body.confirm).to.eq(true)
      expect(req.body.source_surface).to.eq('inventory.quick-create')
      expect(req.body.source_channel).to.eq('in-app')
      expect(req.body.parameters.part_number).to.eq('INV-1707-SP')
      expect(req.body.parameters.title).to.eq(
        'Inventory side panel dispatch proof'
      )
      expect(req.body.parameters.source_url).to.eq(
        'https://example.test/inventory/1707'
      )
      req.reply({
        statusCode: 200,
        body: {
          skill_id: 'cabinet.inventory.create_item',
          mutation_applied: true,
          source_surface: 'inventory.quick-create',
          source_channel: 'in-app',
          target: {
            operation: 'inventory.item.create',
            item_id: 'item-inventory-1707',
            part_number: 'INV-1707-SP',
            inventory_persisted: true,
            external_write_claimed: false,
          },
        },
      })
    }).as('inventorySkillApply')
    openAssistantWorkspace()

    cy.get('[data-testid="shell-assistant-agent-skill-panel"]')
      .scrollIntoView()
      .should('exist')
    cy.get('[data-testid="shell-assistant-agent-skill-select"]').select(
      'cabinet.inventory.create_item',
      { force: true }
    )
    cy.get('[data-testid="shell-assistant-agent-skill-provider"]')
      .clear()
      .type('INV-1707-SP', { force: true })
    cy.get('[data-testid="shell-assistant-agent-skill-setup-step"]')
      .clear()
      .type('Inventory side panel dispatch proof', { force: true })
    cy.get('[data-testid="shell-assistant-agent-skill-secret"]')
      .clear()
      .type('https://example.test/inventory/1707', { force: true })
    cy.get('[data-testid="shell-assistant-agent-skill-preview"]').click({
      force: true,
    })

    cy.wait('@inventorySkillPreview')
    cy.get('[data-testid="shell-assistant-agent-skill-preview-card"]')
      .should('contain', 'cabinet.inventory.create_item')
      .and('contain', 'confirm-required')
      .and('contain', 'confirmation_required')

    cy.get('[data-testid="shell-assistant-agent-skill-apply"]').click({
      force: true,
    })
    cy.get('[data-testid="shell-assistant-apply-confirm-summary"]')
      .should('contain', 'cabinet.inventory.create_item')
      .and('contain', 'inventory.quick-create')
    cy.get('[data-testid="shell-assistant-apply-confirm"]').click()

    cy.wait('@inventorySkillApply')
    cy.get('[data-testid="shell-assistant-agent-skill-result"]')
      .should('contain', 'inventory.item.create')
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
