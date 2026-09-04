describe('chats/assistant-inbox-agent-context', () => {
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

  it('AGENT-UNIVERSAL-CHANNELS-001/#1987 preserves Inbox review notification context in the Agent envelope', () => {
    bootstrapInventory()

    cy.request('POST', '/api/chat/threads', {
      profile_id: 'e2e-profile-001',
      title: 'Inbox review Agent context #1987',
      metadata: {
        provider: 'openai',
        model: 'gpt-4o-mini',
        thread_semantics: 'assistant_workspace_session',
      },
    }).then(({ body }) => {
      const threadID = String(body.id)
      cy.intercept('GET', '/api/chat/inbox?profile_id=*', {
        statusCode: 200,
        body: {
          items: [
            {
              id: 'notice-agent-context-1987',
              status: 'unread',
              source: 'assistant_handoff',
              thread_id: threadID,
              title: 'Inbox Agent context review',
              summary: 'Continue this governed Inbox review in Agent.',
              created_at: '2026-07-31T10:45:00Z',
              updated_at: '2026-07-31T10:45:00Z',
              metadata: {
                category: 'assistant',
                source_label: 'Inbox review',
                source_surface: 'inbox.notification.card',
                source_channel: 'in-app',
                source_thread_id: 'inbox-source-thread-1987',
                source_message_id: 'inbox-source-message-1987',
                review_item_id: 'notice-agent-context-1987',
                assistant: {
                  provider: 'openai',
                  model: 'gpt-4o-mini',
                },
              },
            },
          ],
        },
      }).as('loadInboxContext')
    })

    cy.get('[data-testid="shell-workspace-bell"]').click()
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inbox\/?$/)
    cy.wait('@loadInboxContext')
    cy.contains(
      '[data-testid="notification-inbox-detail-pane"]',
      'Inbox Agent context review'
    ).should('be.visible')
    cy.get('[data-testid="notification-inbox-open-agent"]').click()

    cy.get('[data-testid="shell-workspace-assistant"]').should(
      'have.attr',
      'data-active',
      'true'
    )
    cy.get('[data-testid="shell-assistant-thread-id"]')
      .should(($threadId) => {
        expect($threadId.text().trim()).not.to.eq('bootstrapping')
      })
      .then(($threadId) => {
        const activeThreadID = $threadId.text().trim()
        expect(activeThreadID).not.to.eq('')
        cy.intercept('POST', '/api/chat/messages').as(
          'assistantFromInboxReview'
        )
        cy.get('[data-testid="shell-assistant-compose-input"]').type(
          'continue the inbox review'
        )
        cy.get('[data-testid="shell-assistant-send-button"]').click()
        cy.wait('@assistantFromInboxReview').then(({ request, response }) => {
          expect(response?.statusCode).to.eq(201)
          expect(request.body.profile_id).to.eq('e2e-profile-001')
          expect(request.body.thread_id).to.eq(activeThreadID)
          expect(request.body.agent_context.profile_id).to.eq('e2e-profile-001')
          expect(request.body.agent_context.thread_id).to.eq(activeThreadID)
          expect(request.body.agent_context.route_id).to.match(
            /^\/inventory\/?$/
          )
          expect(request.body.agent_context.surface_id).to.eq(
            'inbox.notification.card'
          )
          expect(request.body.agent_context.source_channel).to.eq('in-app')
          expect(request.body.agent_context.selected_notification).to.deep.eq({
            id: 'notice-agent-context-1987',
            source: 'assistant_handoff',
          })
          expect(request.body.agent_context.source_thread_id).to.eq(
            'inbox-source-thread-1987'
          )
          expect(request.body.agent_context.source_message_id).to.eq(
            'inbox-source-message-1987'
          )
          expect(request.body.agent_context.permission_state).to.eq(
            'ask_before_local_changes'
          )
          expect(request.body.agent_context.setup_state).to.eq('ready')
        })
      })
  })
})
