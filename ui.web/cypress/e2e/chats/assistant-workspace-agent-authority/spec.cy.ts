describe('chats/assistant-workspace-agent-authority', () => {
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

  function seedAssistantThread() {
    cy.request('POST', '/api/chat/threads', {
      profile_id: 'e2e-profile-001',
      title: 'Assistant Authority #1932',
      metadata: {
        model: 'gpt-4o-mini',
        provider: 'openai',
        thread_semantics: 'assistant_workspace_session',
      },
    }).then(({ body }) => {
      const threadID = String(body.id)
      cy.window().then((win) => {
        win.localStorage.setItem(
          'cabinet.assistant.workspace.thread.e2e-profile-001',
          threadID
        )
        win.localStorage.setItem(
          'cabinet.assistant.workspace.provider.e2e-profile-001',
          'openai'
        )
        win.localStorage.setItem(
          'cabinet.assistant.workspace.model.e2e-profile-001',
          'gpt-4o-mini'
        )
      })
    })
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

  it('AGENT-AUTHORITY-005/#1932 shows read-only policy blockers for side-panel Agent Skill mutations', () => {
    bootstrapInventory()
    seedAssistantThread()
    cy.request('PUT', '/api/profiles/e2e-profile-001/settings', {
      settings: {
        'agent.authority.external_write_approved': 'false',
        'agent.authority.mode': 'read_only',
      },
    })
    cy.intercept('POST', '/api/agent/skills/preview').as('agentSkillPreview')
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
      .type('INV-1932-RO', { force: true })
    cy.get('[data-testid="shell-assistant-agent-skill-setup-step"]')
      .clear()
      .type('Read-only side panel blocked item', { force: true })
    cy.get('[data-testid="shell-assistant-agent-skill-preview"]').click({
      force: true,
    })

    cy.wait('@agentSkillPreview').then(({ request, response }) => {
      expect(request.body.profile_id).to.eq('e2e-profile-001')
      expect(request.body.skill_id).to.eq('cabinet.inventory.create_item')
      expect(request.body.source_channel).to.eq('in-app')
      expect(request.body.parameters.part_number).to.eq('INV-1932-RO')
      expect(request.body.parameters.title).to.eq(
        'Read-only side panel blocked item'
      )
      expect(response?.statusCode).to.eq(409)
      expect(response?.body.error).to.eq('agent_authority_read_only')
      expect(response?.body.authority.entry_point).to.eq('direct-api')
    })

    cy.get('[data-testid="shell-assistant-error"]')
      .should('contain', 'agent_authority_read_only')
      .and('contain', 'direct-api')
    cy.get('[data-testid="shell-assistant-permission-guidance"]').should(
      'contain',
      'Switch the profile Agent authority mode'
    )
    cy.contains('[data-testid="inventory-list-item"]', 'INV-1932-RO').should(
      'not.exist'
    )
  })

  it('AGENT-AUTHORITY-007/#1932 blocks side-panel Agent Skill apply after policy changes to read-only', () => {
    bootstrapInventory()
    seedAssistantThread()
    cy.request('PUT', '/api/profiles/e2e-profile-001/settings', {
      settings: {
        'agent.authority.external_write_approved': 'false',
        'agent.authority.mode': 'ask_before_local_changes',
      },
    })
    cy.intercept('POST', '/api/agent/skills/preview').as('agentSkillPreview')
    cy.intercept('POST', '/api/agent/skills/apply').as('agentSkillApply')
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
      .type('INV-1932-BYPASS', { force: true })
    cy.get('[data-testid="shell-assistant-agent-skill-setup-step"]')
      .clear()
      .type('Side panel late apply bypass item', { force: true })
    cy.get('[data-testid="shell-assistant-agent-skill-preview"]').click({
      force: true,
    })

    cy.wait('@agentSkillPreview').then(({ response }) => {
      expect(response?.statusCode).to.eq(200)
      expect(response?.body.skill_id).to.eq('cabinet.inventory.create_item')
      expect(response?.body.confirmation_required).to.eq(true)
      expect(response?.body.blocker).to.eq('confirmation_required')
    })
    cy.get('[data-testid="shell-assistant-agent-skill-preview-card"]')
      .should('contain', 'cabinet.inventory.create_item')
      .and('contain', 'confirmation_required')

    cy.request('PUT', '/api/profiles/e2e-profile-001/settings', {
      settings: {
        'agent.authority.external_write_approved': 'false',
        'agent.authority.mode': 'read_only',
      },
    })
    cy.get('[data-testid="shell-assistant-agent-skill-apply"]').click({
      force: true,
    })
    cy.get('[data-testid="shell-assistant-apply-confirm-dialog"]').should(
      'be.visible'
    )
    cy.get('[data-testid="shell-assistant-apply-confirm"]').click({
      force: true,
    })

    cy.wait('@agentSkillApply').then(({ request, response }) => {
      expect(request.body.profile_id).to.eq('e2e-profile-001')
      expect(request.body.skill_id).to.eq('cabinet.inventory.create_item')
      expect(request.body.parameters.part_number).to.eq('INV-1932-BYPASS')
      expect(response?.statusCode).to.eq(409)
      expect(response?.body.error).to.eq('agent_authority_read_only')
      expect(response?.body.authority.entry_point).to.eq('direct-api')
    })
    cy.get('[data-testid="shell-assistant-error"]')
      .should('contain', 'agent_authority_read_only')
      .and('contain', 'direct-api')
    cy.get('[data-testid="shell-assistant-permission-guidance"]').should(
      'contain',
      'Switch the profile Agent authority mode'
    )
    cy.contains(
      '[data-testid="inventory-list-item"]',
      'INV-1932-BYPASS'
    ).should('not.exist')
  })

  it('AGENT-CONTEXT-003/#1714 sends selected inventory row context with side-panel Agent Skill preview', () => {
    cy.intercept('GET', '/api/items*', {
      statusCode: 200,
      body: {
        items: [
          {
            id: 'item-agent-context-1',
            part_number: 'CTX-1714-ROW',
            title: 'Agent Context Row',
            status: 'todo',
            condition: 'new',
            category: 'feature',
            item_type: 'car',
            packaging_grade_type: 'carded',
            brand: 'AFX',
            priority: 'medium',
            description: '',
            notes: '',
            tags: [],
            source_urls: [],
          },
        ],
      },
    }).as('itemsForAgentContext')
    bootstrapInventory()
    seedAssistantThread()
    cy.wait('@itemsForAgentContext')
    cy.get('[data-testid="inventory-item-row-item-agent-context-1"]')
      .should('be.visible')
      .click()
    cy.intercept('POST', '/api/agent/skills/preview').as('agentSkillPreview')
    openAssistantWorkspace()

    cy.get('[data-testid="shell-assistant-selected-record-context"]').should(
      'contain',
      'inventory_item:CTX-1714-ROW'
    )
    cy.get('[data-testid="shell-assistant-agent-skill-panel"]')
      .scrollIntoView()
      .should('exist')
    cy.get('[data-testid="shell-assistant-agent-skill-select"]').select(
      'cabinet.inventory.update_item',
      { force: true }
    )
    cy.get('[data-testid="shell-assistant-agent-skill-provider"]')
      .clear()
      .type('item-agent-context-1', { force: true })
    cy.get('[data-testid="shell-assistant-agent-skill-setup-step"]')
      .clear()
      .type('Agent Context Row renamed', { force: true })
    cy.get('[data-testid="shell-assistant-agent-skill-preview"]').click({
      force: true,
    })

    cy.wait('@agentSkillPreview').then(({ request, response }) => {
      expect(request.body.profile_id).to.eq('e2e-profile-001')
      expect(request.body.skill_id).to.eq('cabinet.inventory.update_item')
      expect(request.body.agent_context.route_id).to.match(/^\/inventory\/?$/)
      expect(request.body.agent_context.surface_id).to.eq(
        'inventory.item.detail'
      )
      expect(request.body.agent_context.source_channel).to.eq('in-app')
      expect(request.body.agent_context.thread_id)
        .to.be.a('string')
        .and.not.eq('')
      expect(request.body.agent_context.selected_record).to.deep.eq({
        type: 'inventory_item',
        id: 'item-agent-context-1',
      })
      expect(request.body.parameters.item_id).to.eq('item-agent-context-1')
      expect(response?.statusCode).to.eq(200)
      expect(response?.body.source_channel).to.eq('in-app')
      expect(response?.body.source_thread_id).to.eq(
        request.body.agent_context.thread_id
      )
    })
    cy.get('[data-testid="shell-assistant-agent-skill-preview-card"]')
      .should('contain', 'cabinet.inventory.update_item')
      .and('contain', 'confirmation_required')
  })
})
