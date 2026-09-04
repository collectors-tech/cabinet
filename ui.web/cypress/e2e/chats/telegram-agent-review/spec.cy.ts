describe('chats/telegram-agent-review', () => {
  it('TELEGRAM-AGENT-REVIEW-001 opens external review thread and preview from URL state', () => {
    cy.viewport(1400, 900)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/chats',
    })

    cy.intercept('GET', '/api/chat/threads?profile_id=*', {
      statusCode: 200,
      body: {
        threads: [
          {
            id: 'thread-general',
            profile_id: 'e2e-profile-001',
            title: 'General chat',
            created_at: '2026-07-08T00:00:00Z',
            updated_at: '2026-07-08T00:00:00Z',
          },
          {
            id: 'thread-telegram-agent',
            profile_id: 'e2e-profile-001',
            title: 'Telegram Agent: inventory create review',
            created_at: '2026-07-08T00:05:00Z',
            updated_at: '2026-07-08T00:05:00Z',
          },
        ],
      },
    }).as('loadThreads')
    cy.intercept(
      'GET',
      '/api/chat/messages?profile_id=*&thread_id=thread-telegram-agent',
      {
        statusCode: 200,
        body: {
          messages: [
            {
              id: 'message-telegram-agent',
              profile_id: 'e2e-profile-001',
              thread_id: 'thread-telegram-agent',
              role: 'user',
              content: 'Create inventory item TG-1705 from Telegram',
              created_at: '2026-07-08T00:05:00Z',
            },
          ],
        },
      }
    ).as('loadMessages')
    cy.intercept(
      'GET',
      '/api/chat/workflow-runs?profile_id=*&thread_id=thread-telegram-agent',
      {
        statusCode: 200,
        body: {
          runs: [
            {
              id: 'workflow-telegram-agent-preview',
              workflow_id: 'telegram-agent-skill-text',
              capability_id: 'cabinet.inventory.create_item',
              source_channel: 'telegram',
              source_thread_id: 'thread-telegram-agent',
              source_message_id: 'telegram-message-1705',
              status: 'waiting_confirmation',
              confirmation_state: 'preview_required',
              result: {
                preview_id: 'preview-telegram-agent',
                review_url:
                  '/chats?profile_id=e2e-profile-001&thread_id=thread-telegram-agent&preview_id=preview-telegram-agent',
              },
              created_at: '2026-07-08T00:05:00Z',
              updated_at: '2026-07-08T00:05:00Z',
            },
          ],
        },
      }
    ).as('loadWorkflowRuns')
    cy.intercept(
      'GET',
      '/api/chat/actions/preview?profile_id=e2e-profile-001&preview_id=preview-telegram-agent',
      {
        statusCode: 200,
        body: {
          id: 'preview-telegram-agent',
          profile_id: 'e2e-profile-001',
          thread_id: 'thread-telegram-agent',
          action: 'create_inventory_item',
          payload: {
            part_number: 'TG-1705',
            title: 'Telegram Agent Truck',
            assistant_provider: 'telegram',
            assistant_model: 'cabinet-agent',
          },
          status: 'previewed',
          created_at: '2026-07-08T00:05:00Z',
        },
      }
    ).as('loadActionPreview')

    cy.visit(
      '/chats?profile_id=e2e-profile-001&thread_id=thread-telegram-agent&preview_id=preview-telegram-agent'
    )

    cy.location('pathname', { timeout: 15000 }).should('match', /^\/chats\/?$/)
    cy.wait('@loadThreads')
    cy.wait('@loadMessages')
    cy.wait('@loadWorkflowRuns')
    cy.wait('@loadActionPreview')

    cy.get('[data-testid="chat-thread-title"]').should(
      'contain',
      'Telegram Agent: inventory create review'
    )
    cy.contains(
      '[data-testid="chat-message-list"]',
      'Create inventory item TG-1705 from Telegram'
    ).should('be.visible')
    cy.get('[data-testid="chat-action-preview"]')
      .should('contain', 'preview-telegram-agent')
      .and('contain', 'create_inventory_item')
      .and('contain', 'part_number=TG-1705')
      .and('contain', 'title=Telegram Agent Truck')
    cy.get('[data-testid="chat-apply-action-button"]').should('be.enabled')
    cy.get('[data-testid="chat-action-timeline"]')
      .should('contain', 'cabinet.inventory.create_item')
      .and('contain', 'telegram-agent-skill-text / preview_required')
      .and('contain', 'preview preview-telegram-agent')
  })
})
