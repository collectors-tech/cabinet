describe('chats/assistant-workspace', () => {
  function bootstrapInventory() {
    cy.viewport(1400, 900)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', { path: '/inventory/' })
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)
  }

  function openAssistantWorkspace() {
    cy.intercept('POST', '/api/chat/threads').as('assistantBootstrapThread')
    cy.get('[data-testid="shell-chat-toggle"]').click()
    cy.wait('@assistantBootstrapThread', { timeout: 20000 })
    cy.get('[data-testid="shell-assistant-thread-id"]', {
      timeout: 20000,
    }).should(($threadId) => {
      expect($threadId.text().trim()).not.to.eq('')
      expect($threadId.text().trim()).not.to.eq('bootstrapping')
    })
  }

  it('ASSISTANT-WORKSPACE-001 preserves assistant thread continuity across route and workspace changes', () => {
    bootstrapInventory()
    openAssistantWorkspace()
    cy.get('[data-testid="shell-assistant-compose-input"]').type('remember this route context')
    cy.get('[data-testid="shell-assistant-send-button"]').click()
    cy.contains('[data-testid="shell-assistant-message-list"]', 'remember this route context').should('exist')
    cy.get('[data-testid="shell-assistant-thread-id"]').invoke('text').as('threadId')

    cy.visit('/wishlist')
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/wishlist\/?$/)
    cy.get('[data-testid="shell-workspace-assistant"]').should('have.attr', 'data-active', 'true')
    cy.get('[data-testid="shell-assistant-message-list"]').contains('remember this route context')
    cy.get('@threadId').then((threadId) => {
      cy.get('[data-testid="shell-assistant-thread-id"]').should('have.text', String(threadId).trim())
    })
  })

  it('ASSISTANT-WORKSPACE-002 sends deterministic route/profile/selection context in assistant message envelopes', () => {
    bootstrapInventory()
    cy.intercept('POST', '/api/chat/messages').as('assistantMessage')
    openAssistantWorkspace()
    cy.get('[data-testid="shell-assistant-selection-context"]').should('contain', 'All Items')
    cy.get('[data-testid="shell-assistant-compose-input"]').type('what should I do with this inventory route?')
    cy.get('[data-testid="shell-assistant-send-button"]').click()
    cy.wait('@assistantMessage').then(({ request }) => {
      expect(request.body.profile_id).to.eq('e2e-profile-001')
      expect(String(request.body.thread_id).trim()).not.to.equal('')
      expect(request.body.context.route.pathname).to.match(/^\/inventory\/?$/)
      expect(request.body.context.profile.id).to.eq('e2e-profile-001')
      expect(request.body.context.selection.active_workspace_collection).to.eq('All Items')
      expect(request.body.context.assistant.provider).to.eq('openai')
      expect(request.body.context.assistant.model).to.eq('gpt-4o-mini')
    })
    cy.contains('[data-testid="shell-assistant-message-list"]', 'what should I do with this inventory route?').should('exist')
  })

  it('ASSISTANT-WORKSPACE-003 changes provider/model with deterministic forked-thread semantics', () => {
    bootstrapInventory()
    let originalThreadId = ''
    openAssistantWorkspace()
    cy.get('[data-testid="shell-assistant-thread-id"]')
      .invoke('text')
      .then((threadId) => {
        originalThreadId = String(threadId).trim()
      })
    cy.get('[data-testid="shell-assistant-thread-provider"]').should('contain', 'openai')
    cy.get('[data-testid="shell-assistant-thread-model"]').should('contain', 'gpt-4o-mini')

    cy.intercept('POST', '/api/chat/threads').as('assistantThreadCreate')
    cy.get('[data-testid="shell-assistant-provider-select"]').select('anthropic')
    cy.wait('@assistantThreadCreate').then(({ request }) => {
      expect(request.body.metadata.provider).to.eq('anthropic')
      expect(request.body.metadata.model).to.eq('claude-3-5-haiku')
      expect(request.body.metadata.thread_semantics).to.eq('fork_on_provider_model_change')
      expect(request.body.metadata.forked_from_thread_id).to.eq(originalThreadId)
    })

    cy.get('[data-testid="shell-assistant-thread-provider"]').should('contain', 'anthropic')
    cy.get('[data-testid="shell-assistant-thread-model"]').should('contain', 'claude-3-5-haiku')
    cy.get('[data-testid="shell-assistant-thread-semantics"]').should('contain', 'fork a new assistant thread')
    cy.get('[data-testid="shell-assistant-thread-id"]').should(($next) => {
      expect($next.text().trim()).not.to.eq(originalThreadId)
    })
  })

  it('ASSISTANT-WORKSPACE-004 applies explicit reset boundaries for manual new-thread and active profile changes', () => {
    let primaryThreadId = ''
    cy.request('POST', '/api/test/reset', {})
    cy.request('POST', '/api/profiles', { name: 'Primary DB' }).then((primaryResp) => {
      expect(primaryResp.status).to.eq(201)
      const primaryID = primaryResp.body.id as string

      cy.request('POST', '/api/profiles', { name: 'Showcase DB' }).then((showcaseResp) => {
        expect(showcaseResp.status).to.eq(201)
        const showcaseID = showcaseResp.body.id as string

        cy.request('PUT', '/api/profiles/active', { profile_id: primaryID }).its('status').should('eq', 200)
        cy.visit('/sign-in?redirect=%2Finventory%2F', {
          onBeforeLoad(win) {
            win.localStorage.setItem(`cabinet.workspace.${primaryID}`, '1')
          },
        })
        cy.get('input[name="email"]').clear().type('e2e-login-session@example.com')
        cy.get('input[name="password"]').clear().type('password123')
        cy.contains('button', 'Sign in').click()
        cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)

        openAssistantWorkspace()
        cy.get('[data-testid="shell-assistant-compose-input"]').type('primary profile message')
        cy.get('[data-testid="shell-assistant-send-button"]').click()
        cy.contains('[data-testid="shell-assistant-message-list"]', 'primary profile message').should('exist')
        cy.get('[data-testid="shell-assistant-thread-id"]')
          .invoke('text')
          .then((threadId) => {
            primaryThreadId = String(threadId).trim()
          })

        cy.intercept('POST', '/api/chat/threads').as('assistantResetThread')
        cy.get('[data-testid="shell-assistant-new-thread"]').click()
        cy.wait('@assistantResetThread').then(({ request }) => {
          expect(request.body.metadata.thread_semantics).to.eq('manual_new_thread')
        })
        cy.get('[data-testid="shell-assistant-thread-id"]').should(($next) => {
          expect($next.text().trim()).not.to.eq(primaryThreadId)
        })
        cy.contains('[data-testid="shell-assistant-message-list"]', 'primary profile message').should('not.exist')

        cy.request('PUT', '/api/profiles/active', { profile_id: showcaseID }).its('status').should('eq', 200)
        cy.reload()
        cy.get('[data-testid="active-profile-name"]', { timeout: 20000 }).should('contain', 'Showcase DB')
        cy.get('[data-testid="shell-chat-toggle"]').click()
        cy.get('[data-testid="shell-assistant-profile-scope"]').should('have.text', showcaseID)
        cy.contains('[data-testid="shell-assistant-message-list"]', 'primary profile message').should('not.exist')
        cy.get('[data-testid="shell-assistant-thread-id"]').should(($next) => {
          expect($next.text().trim()).not.to.eq('')
        })
      })
    })
  })

  it('ASSISTANT-WORKSPACE-006 selects chats, creates a new chat, and exposes a layout navigation action', () => {
    bootstrapInventory()
    openAssistantWorkspace()
    cy.get('[data-testid="shell-assistant-thread-select"]').should('be.visible')
    cy.get('[data-testid="shell-assistant-thread-id"]').should(($threadId) => {
      expect($threadId.text().trim()).not.to.eq('')
      expect($threadId.text().trim()).not.to.eq('bootstrapping')
    })
    cy.get('[data-testid="shell-assistant-thread-id"]')
      .invoke('text')
      .then((initialThreadId) => {
        const firstThreadId = String(initialThreadId).trim()
        expect(firstThreadId).not.to.eq('')
        expect(firstThreadId).not.to.eq('bootstrapping')

        cy.get('[data-testid="shell-assistant-new-thread"]').click()
        cy.get('[data-testid="shell-assistant-thread-id"]').should(($next) => {
          const secondThreadId = $next.text().trim()
          expect(secondThreadId).not.to.eq('')
          expect(secondThreadId).not.to.eq('bootstrapping')
          expect(secondThreadId).not.to.eq(firstThreadId)
        })
        cy.get('[data-testid="shell-assistant-thread-select"] option').should(
          'have.length.at.least',
          2
        )

        cy.get('[data-testid="shell-assistant-compose-input"]').type(
          'show me a config for layout'
        )
        cy.get('[data-testid="shell-assistant-send-button"]').click()
        cy.get('[data-testid="shell-assistant-navigation-action"]')
          .should('be.visible')
          .and('contain', 'Open layout settings')
        cy.location('pathname').should('match', /^\/inventory\/?$/)

        cy.get('[data-testid="shell-assistant-thread-select"]').select(
          firstThreadId
        )
        cy.get('[data-testid="shell-assistant-thread-id"]').should(
          'have.text',
          firstThreadId
        )
        cy.get('[data-testid="shell-assistant-navigation-action"]').should(
          'not.exist'
        )

        cy.get('[data-testid="shell-assistant-new-thread"]').click()
        cy.get('[data-testid="shell-assistant-thread-id"]').should(($next) => {
          const nextThreadId = $next.text().trim()
          expect(nextThreadId).not.to.eq('')
          expect(nextThreadId).not.to.eq('bootstrapping')
          expect(nextThreadId).not.to.eq(firstThreadId)
        })
        cy.get('[data-testid="shell-assistant-compose-input"]').type(
          'show me a config for layout'
        )
        cy.get('[data-testid="shell-assistant-send-button"]').click()
        cy.get('[data-testid="shell-assistant-navigation-action-open"]').click()
        cy.location('pathname', { timeout: 15000 }).should(
          'match',
          /^\/settings\/display\/?$/
        )
      })
  })

  it('ASSISTANT-WORKSPACE-005 renders Cabinet assistant-ui adapter primitives while preserving context envelopes and manual action confirmation', () => {
    bootstrapInventory()
    cy.intercept('POST', '/api/chat/messages').as('assistantMessage')
    openAssistantWorkspace()
    cy.get('[data-testid="shell-assistant-modal-anchor"]').should('exist')
    cy.get('[data-testid="shell-assistant-modal-trigger"]').should('be.visible')
    cy.get('[data-testid="shell-assistant-modal-content"]').should('be.visible')
    cy.get('[data-testid="shell-assistant-ui-adapter"]').should('exist')
    cy.get('[data-testid="shell-assistant-ui-composer-primitive"]').should(
      'be.visible'
    )

    cy.get('[data-testid="shell-assistant-compose-input"]').type(
      'adapter should preserve cabinet context'
    )
    cy.get('[data-testid="shell-assistant-send-button"]').click()
    cy.wait('@assistantMessage').then(({ request }) => {
      expect(request.body.content).to.eq('adapter should preserve cabinet context')
      expect(request.body.profile_id).to.eq('e2e-profile-001')
      expect(String(request.body.thread_id).trim()).not.to.equal('')
      expect(request.body.context.route.pathname).to.match(/^\/inventory\/?$/)
      expect(request.body.context.profile.id).to.eq('e2e-profile-001')
      expect(request.body.context.selection.active_workspace_collection).to.eq(
        'All Items'
      )
      expect(request.body.context.assistant.provider).to.eq('openai')
      expect(request.body.context.assistant.model).to.eq('gpt-4o-mini')
    })
    cy.contains(
      '[data-testid="shell-assistant-message-list"]',
      'adapter should preserve cabinet context'
    ).should('exist')
    cy.get('[data-testid="shell-assistant-ui-message-primitive"]').should(
      'have.length.at.least',
      1
    )

    cy.get('[data-testid="shell-assistant-preview-part-number"]')
      .clear()
      .type('ADAPT-1133')
    cy.get('[data-testid="shell-assistant-preview-title"]')
      .clear()
      .type('Adapter Guarded Item')
    cy.get('[data-testid="shell-assistant-preview-action"]').click()
    cy.get('[data-testid="shell-assistant-action-card"]').should(
      'contain',
      'ADAPT-1133'
    )
    cy.get('[data-testid="shell-assistant-apply-action"]').click()
    cy.get('[data-testid="shell-assistant-apply-confirm-dialog"]').should(
      'be.visible'
    )
    cy.get('[data-testid="shell-assistant-apply-confirm-summary"]').should(
      'contain',
      'ADAPT-1133'
    )
    cy.get('[data-testid="shell-assistant-apply-cancel"]').click()
    cy.get('[data-testid="shell-assistant-apply-confirm-dialog"]').should(
      'not.exist'
    )
    cy.get('[data-testid="shell-assistant-action-card"]').should(
      'contain',
      'ADAPT-1133'
    )
  })

  it('ASSISTANT-WORKSPACE-007 exposes governed action results with persistence proof', () => {
    bootstrapInventory()
    cy.intercept('POST', '/api/chat/actions/preview').as('assistantPreview')
    cy.intercept('POST', '/api/chat/actions/apply').as('assistantApply')

    openAssistantWorkspace()
    cy.get('[data-testid="shell-assistant-execution-panel"]').should('exist')
    cy.get('[data-testid="shell-assistant-execution-state"]').should(
      'contain',
      'idle'
    )
    cy.get('[data-testid="shell-assistant-apply-action"]').should(
      'be.disabled'
    )

    cy.get('[data-testid="shell-assistant-preview-part-number"]')
      .clear()
      .type('WS-1083')
    cy.get('[data-testid="shell-assistant-preview-title"]')
      .clear()
      .type('Workspace Execution Proof')
    cy.get('[data-testid="shell-assistant-preview-action"]').click()

    cy.wait('@assistantPreview').then(({ request, response }) => {
      expect(request.body.profile_id).to.eq('e2e-profile-001')
      expect(request.body.action).to.eq('create_item_stub')
      expect(request.body.payload.part_number).to.eq('WS-1083')
      expect(request.body.payload.title).to.eq('Workspace Execution Proof')
      expect(response?.statusCode).to.eq(200)
    })
    cy.get('[data-testid="shell-assistant-action-preview"]')
      .should('contain', 'create_item_stub')
      .and('contain', 'WS-1083')
      .and('contain', 'Workspace Execution Proof')
    cy.request('/api/items?profile_id=e2e-profile-001')
      .its('body')
      .should((items) => {
        expect(JSON.stringify(items)).not.to.include('WS-1083')
      })

    cy.get('[data-testid="shell-assistant-apply-action"]').should(
      'not.be.disabled'
    )
    cy.get('[data-testid="shell-assistant-apply-action"]').click()
    cy.get('[data-testid="shell-assistant-apply-confirm-dialog"]').should(
      'be.visible'
    )
    cy.get('[data-testid="shell-assistant-apply-confirm-summary"]').should(
      'contain',
      'WS-1083'
    )
    cy.get('[data-testid="shell-assistant-apply-confirm"]').click()

    cy.wait('@assistantApply').then(({ request, response }) => {
      expect(request.body.profile_id).to.eq('e2e-profile-001')
      expect(request.body.confirm).to.eq(true)
      expect(response?.statusCode).to.eq(200)
      expect(response?.body.applied).to.eq(true)
      expect(String(response?.body.item_id).trim()).not.to.eq('')
    })
    cy.get('[data-testid="shell-assistant-execution-state"]').should(
      'contain',
      'success'
    )
    cy.get('[data-testid="shell-assistant-apply-result"]')
      .should('contain', 'Applied create_item_stub')
      .and('contain', 'to ')
    cy.get('[data-testid="shell-assistant-result-link"]')
      .should('contain', 'Open item')
      .and('have.attr', 'href')
      .and('include', '/inventory/?item=')
    cy.request('/api/items?profile_id=e2e-profile-001')
      .its('body')
      .should((items) => {
        expect(JSON.stringify(items)).to.include('WS-1083')
        expect(JSON.stringify(items)).to.include('Workspace Execution Proof')
      })
  })
})
