describe('chats/assistant-workspace', () => {
  function bootstrapInventory() {
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', { path: '/inventory/' })
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)
  }

  it('ASSISTANT-WORKSPACE-001 preserves assistant thread continuity across route and workspace changes', () => {
    bootstrapInventory()
    cy.get('[data-testid="shell-chat-toggle"]').click()
    cy.get('[data-testid="shell-assistant-compose-input"]').type('remember this route context')
    cy.get('[data-testid="shell-assistant-send-button"]').click()
    cy.contains('[data-testid="shell-assistant-message-list"]', 'remember this route context').should(
      'be.visible'
    )
    cy.get('[data-testid="shell-assistant-thread-id"]').invoke('text').as('threadId')

    cy.get('[data-testid="shell-workspace-navigation"]').click()
    cy.get('[data-testid="shell-chat-toggle"]').click()
    cy.get('[data-testid="shell-assistant-message-list"]').contains('remember this route context')
    cy.get('[data-testid="sidebar-nav-link-wishlist"]').click()
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
    cy.get('[data-testid="shell-chat-toggle"]').click()
    cy.get('[data-testid="shell-assistant-selection-context"]').should('contain', 'All Items')
    cy.get('[data-testid="shell-assistant-compose-input"]').type('what should I do with this inventory route?')
    cy.get('[data-testid="shell-assistant-send-button"]').click()
    cy.wait('@assistantMessage').then(({ request }) => {
      expect(request.body.profile_id).to.eq('e2e-profile-001')
      expect(String(request.body.thread_id).trim()).not.to.equal('')
      expect(request.body.context.route.pathname).to.eq('/inventory/')
      expect(request.body.context.profile.id).to.eq('e2e-profile-001')
      expect(request.body.context.selection.active_workspace_collection).to.eq('All Items')
      expect(request.body.context.assistant.provider).to.eq('openai')
      expect(request.body.context.assistant.model).to.eq('gpt-4o-mini')
    })
    cy.contains('[data-testid="shell-assistant-message-list"]', 'what should I do with this inventory route?').should(
      'be.visible'
    )
  })

  it('ASSISTANT-WORKSPACE-003 changes provider/model with deterministic forked-thread semantics', () => {
    bootstrapInventory()
    cy.intercept('POST', '/api/chat/threads').as('assistantThreadCreate')
    cy.get('[data-testid="shell-chat-toggle"]').click()
    cy.get('[data-testid="shell-assistant-thread-id"]').invoke('text').as('originalThreadId')
    cy.get('[data-testid="shell-assistant-thread-provider"]').should('contain', 'openai')
    cy.get('[data-testid="shell-assistant-thread-model"]').should('contain', 'gpt-4o-mini')

    cy.get('[data-testid="shell-assistant-provider-select"]').select('anthropic')
    cy.wait('@assistantThreadCreate').then(({ request }) => {
      expect(request.body.metadata.provider).to.eq('anthropic')
      expect(request.body.metadata.model).to.eq('claude-3-5-haiku')
      expect(request.body.metadata.thread_semantics).to.eq('fork_on_provider_model_change')
      cy.get('@originalThreadId').then((originalThreadId) => {
        expect(request.body.metadata.forked_from_thread_id).to.eq(String(originalThreadId).trim())
      })
    })

    cy.get('[data-testid="shell-assistant-thread-provider"]').should('contain', 'anthropic')
    cy.get('[data-testid="shell-assistant-thread-model"]').should('contain', 'claude-3-5-haiku')
    cy.get('[data-testid="shell-assistant-thread-semantics"]').should('contain', 'fork a new assistant thread')
    cy.get('@originalThreadId').then((originalThreadId) => {
      cy.get('[data-testid="shell-assistant-thread-id"]').should(($next) => {
        expect($next.text().trim()).not.to.eq(String(originalThreadId).trim())
      })
    })
  })

  it('ASSISTANT-WORKSPACE-004 applies explicit reset boundaries for manual new-thread and active profile changes', () => {
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

        cy.get('[data-testid="shell-chat-toggle"]').click()
        cy.get('[data-testid="shell-assistant-compose-input"]').type('primary profile message')
        cy.get('[data-testid="shell-assistant-send-button"]').click()
        cy.contains('[data-testid="shell-assistant-message-list"]', 'primary profile message').should('be.visible')
        cy.get('[data-testid="shell-assistant-thread-id"]').invoke('text').as('primaryThreadId')

        cy.intercept('POST', '/api/chat/threads').as('assistantResetThread')
        cy.get('[data-testid="shell-assistant-new-thread"]').click()
        cy.wait('@assistantResetThread').then(({ request }) => {
          expect(request.body.metadata.thread_semantics).to.eq('manual_new_thread')
        })
        cy.get('@primaryThreadId').then((primaryThreadId) => {
          cy.get('[data-testid="shell-assistant-thread-id"]').should(($next) => {
            expect($next.text().trim()).not.to.eq(String(primaryThreadId).trim())
          })
        })
        cy.contains('[data-testid="shell-assistant-message-list"]', 'primary profile message').should('not.exist')

        cy.request('PUT', '/api/profiles/active', { profile_id: showcaseID }).its('status').should('eq', 200)
        cy.reload()
        cy.get('[data-testid="active-profile-name"]', { timeout: 20000 }).should('contain', 'Showcase DB')
        cy.get('[data-testid="shell-chat-toggle"]').click()
        cy.get('[data-testid="shell-assistant-profile-scope"]').should('eq', showcaseID)
        cy.contains('[data-testid="shell-assistant-message-list"]', 'primary profile message').should('not.exist')
        cy.get('[data-testid="shell-assistant-thread-id"]').should(($next) => {
          expect($next.text().trim()).not.to.eq('')
        })
      })
    })
  })
})
