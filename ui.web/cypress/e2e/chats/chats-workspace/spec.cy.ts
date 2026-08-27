describe('chats/chats-workspace', () => {
  function openChats() {
    cy.viewport(1400, 900)
    cy.e2eReset()
    cy.e2eBootstrap({ minimalProfile: true })
    cy.e2eSetSetupState('present')
    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', { path: '/chats/' })
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/chats\/?$/)
  }

  function createThread(title: string) {
    cy.get('body').then(($body) => {
      if (
        $body.find('[data-testid="chat-new-thread-dialog"]:visible').length === 0
      ) {
        cy.get(
          '[data-testid="chat-new-chat-button"], [data-testid="chat-empty-workspace-action"], [data-testid="chat-new-thread-action"]'
        )
          .filter(':visible')
          .first()
          .click()
      }
    })
    cy.get('[data-testid="chat-new-thread-dialog"]').should('be.visible')
    cy.get('[data-testid="chat-new-thread-input"]').clear().type(title)
    cy.get('[data-testid="chat-create-thread-button"]').click()
    cy.get('[data-testid="chat-new-thread-dialog"]').should('not.exist')
    cy.get('[data-testid="chat-thread-title"]').should('contain', title)
  }

  it('CHATS-WORKSPACE-001 renders Cabinet-specific chats semantics instead of placeholder inbox copy', () => {
    cy.intercept('GET', '/api/chat/threads?profile_id=*', {
      statusCode: 200,
      delay: 15000,
      body: { threads: [] },
    }).as('delayedEmptyThreads')

    openChats()

    cy.contains('h1', 'Chats').should('be.visible')
    cy.get('[data-testid="chat-workspace-description"]')
      .should(
        'contain.text',
        'Persistent profile-scoped conversation threads backed by Cabinet runtime.'
      )
    cy.get('[data-testid="chat-workspace-boundary-note"]')
      .should(
        'contain.text',
        'Cabinet Agent keeps the same governed conversation, context, and action reviews in this full workspace and the contextual panel.'
      )
    cy.get('[data-testid="chat-thread-list"]').should('be.visible')
    cy.wait('@delayedEmptyThreads')
    cy.contains('No chat threads yet.').should('be.visible')
    cy.contains(/inbox template|stock inbox|placeholder/i).should('not.exist')
  })

  it('CHATS-WORKSPACE-002 preserves the active thread after send', () => {
    openChats()
    createThread('E2E Workspace Thread Preservation')

    cy.get('[data-testid="chat-compose-input"]').type('Hello persistent chats workspace')
    cy.get('[data-testid="chat-send-button"]').click()

    cy.location('pathname').should('match', /^\/chats\/?$/)
    cy.get('[data-testid="chat-thread-title"]').should('contain', 'E2E Workspace Thread Preservation')
    cy.get('[data-testid="chat-message-list"]').should('contain', 'Hello persistent chats workspace')
    cy.contains('[data-testid="chat-thread-item"]', 'E2E Workspace Thread Preservation')
      .should('have.class', 'border-cyan-400/60')
  })

  it('CHATS-WORKSPACE-003 states the unified contextual and full Agent boundary explicitly', () => {
    openChats()

    cy.get('[data-testid="chat-workspace-boundary-note"]')
      .should('contain.text', 'Cabinet Agent')
      .and('contain.text', 'same governed conversation')
      .and('contain.text', 'contextual panel')
    cy.get('[data-testid="shell-chat-toggle"]').should('be.visible')
    cy.location('pathname').should('match', /^\/chats\/?$/)
  })

  it('CHATS-WORKSPACE-004 renders original two-pane chats layout parity', () => {
    openChats()

    cy.get('[data-testid="chat-layout"]').should('be.visible')
    cy.get('[data-testid="chat-conversation-rail"]').should('be.visible')
    cy.get('[data-testid="chat-conversation-search"]')
      .should('be.visible')
      .and('have.attr', 'placeholder', 'Search messages')
    cy.get('[data-testid="chat-empty-workspace-state"]')
      .should('be.visible')
      .and('contain.text', 'How can I help you today?')
      .and('contain.text', 'Choose an existing thread or create a new one to continue a durable Cabinet conversation.')
    cy.get('[data-testid="chat-empty-workspace-action"]')
      .should('be.visible')
      .and('contain.text', 'Start a conversation')

    createThread('E2E Visual Parity Thread')

    cy.contains('[data-testid="chat-thread-item"]', 'E2E Visual Parity Thread')
      .should('be.visible')
      .within(() => {
        cy.get('[data-testid="chat-thread-avatar"]').should('be.visible')
        cy.get('[data-testid="chat-thread-preview"]').should(
          'contain.text',
          'No messages yet'
        )
      })
  })

  it('CHATS-WORKSPACE-006 renders the assistant-ui example visual contract', () => {
    openChats()
    createThread('E2E Assistant UI Visual Contract')

    cy.get('[data-testid="chat-layout"]')
      .should('have.attr', 'data-visual-contract', 'assistant-ui-example')
      .then(($layout) => {
        const layoutRect = $layout[0].getBoundingClientRect()
        expect(layoutRect.height, 'layout height').to.be.greaterThan(560)
      })

    cy.get('[data-testid="chat-conversation-rail"]').then(($rail) => {
      const railRect = $rail[0].getBoundingClientRect()
      expect(railRect.width, 'compact rail width').to.be.within(260, 330)
    })

    cy.get('[data-testid="chat-main-surface"]').then(($surface) => {
      const surfaceRect = $surface[0].getBoundingClientRect()
      cy.get('[data-testid="chat-conversation-rail"]').then(($rail) => {
        const railRect = $rail[0].getBoundingClientRect()
        expect(surfaceRect.width, 'main surface wider than rail').to.be.greaterThan(
          railRect.width
        )
      })
    })

    cy.get('[data-testid="chat-main-canvas"]').then(($canvas) => {
      const canvasRect = $canvas[0].getBoundingClientRect()
      expect(canvasRect.height, 'dominant message canvas height').to.be.greaterThan(330)
    })

    cy.get('[data-testid="chat-composer-shell"]')
      .should('have.attr', 'data-position', 'bottom-center')
      .then(($composer) => {
        const composerRect = $composer[0].getBoundingClientRect()
        cy.get('[data-testid="chat-main-surface"]').then(($surface) => {
          const surfaceRect = $surface[0].getBoundingClientRect()
          const composerCenter = composerRect.left + composerRect.width / 2
          const surfaceCenter = surfaceRect.left + surfaceRect.width / 2
          expect(Math.abs(composerCenter - surfaceCenter), 'composer centered').to.be.lessThan(24)
          expect(
            surfaceRect.bottom - composerRect.bottom,
            'composer docked near bottom'
          ).to.be.lessThan(40)
        })
      })

    cy.contains('p', 'Attachments').should('not.exist')
    cy.contains('p', 'Action Preview').should('not.exist')
    cy.get('[data-testid="chat-tool-card-container"]').should('not.exist')
    cy.get('[data-testid="chat-upload-attachment-button"]').should('not.exist')
  })

  it('CHATS-WORKSPACE-005 filters thread rows and keeps new-thread actions route-stable', () => {
    openChats()

    cy.get('[data-testid="chat-empty-workspace-action"]').click()
    cy.get('[data-testid="chat-new-thread-dialog"]').should('be.visible')
    cy.get('[data-testid="chat-new-thread-input"]').should('be.enabled')
    cy.location('pathname').should('match', /^\/chats\/?$/)

    createThread('E2E Alpha Search Thread')
    cy.get('[data-testid="chat-new-chat-button"]').click()
    cy.get('[data-testid="chat-new-thread-dialog"]').should('be.visible')
    cy.get('[data-testid="chat-new-thread-input"]').should('be.enabled')
    createThread('E2E Beta Search Thread')

    cy.get('[data-testid="chat-conversation-search"]').clear().type('Alpha')
    cy.contains('[data-testid="chat-thread-item"]', 'E2E Alpha Search Thread')
      .should('be.visible')
    cy.contains('[data-testid="chat-thread-item"]', 'E2E Beta Search Thread')
      .should('not.exist')

    cy.get('[data-testid="chat-conversation-search"]').clear()
    cy.contains('[data-testid="chat-thread-item"]', 'E2E Alpha Search Thread')
      .should('be.visible')
    cy.contains('[data-testid="chat-thread-item"]', 'E2E Beta Search Thread')
      .should('be.visible')

    cy.request('/api/chat/threads?profile_id=e2e-profile-001').then((response) => {
      expect(response.status).to.eq(200)
      const titles = (response.body.threads as Array<{ title?: string }>).map(
        (thread) => thread.title
      )
      expect(titles).to.include('E2E Alpha Search Thread')
      expect(titles).to.include('E2E Beta Search Thread')
    })
  })

  it('AGENT-UNIVERSAL-CHANNELS-001/#1979 preserves main Chat Agent context during route planning', () => {
    openChats()
    createThread('E2E Main Chat Route Planner')

    cy.intercept('POST', '/api/chat/messages').as('mainChatPlannerMessage')
    cy.get('[data-testid="chat-compose-input"]').type('open media')
    cy.get('[data-testid="chat-send-button"]').click()

    let threadId = ''
    cy.wait('@mainChatPlannerMessage').then(({ request, response }) => {
      expect(request.body.profile_id).to.eq('e2e-profile-001')
      expect(String(request.body.thread_id).trim()).not.to.eq('')
      expect(request.body.content).to.eq('open media')
      expect(request.body.agent_context.profile_id).to.eq('e2e-profile-001')
      expect(request.body.agent_context.thread_id).to.eq(
        request.body.thread_id
      )
      expect(request.body.agent_context.route_id).to.eq('/chats/')
      expect(request.body.agent_context.surface_id).to.eq('chats.main')
      expect(request.body.agent_context.source_channel).to.eq('in-app')
      expect(request.body.agent_context.intent_text).to.eq('open media')
      expect(request.body.agent_context.permission_state).to.eq(
        'ask_before_local_changes'
      )
      expect(request.body.agent_context.setup_state).to.eq('ready')
      expect(request.body.agent_context.selected_record).to.eq(undefined)
      expect(request.body.context.route.pathname).to.eq('/chats/')
      expect(request.body.context.profile.id).to.eq('e2e-profile-001')
      expect(request.body.context.assistant.provider).to.eq('openai')
      expect(request.body.context.assistant.model).to.eq('gpt-4o-mini')
      expect(response?.statusCode).to.eq(201)
      expect(response?.body.assistant_handoff).to.eq(undefined)
      expect(response?.body.app_control.capability_id).to.eq(
        'navigate.open_surface'
      )
      expect(response?.body.app_control.route).to.eq('/media')
      expect(response?.body.app_control.workflow_run.workflow_id).to.eq(
        'chat.app_control.dispatch'
      )
      expect(response?.body.app_control.workflow_run.confirmation_state).to.eq(
        'not_required'
      )
      threadId = String(request.body.thread_id)
    })

    cy.get('[data-testid="chat-message-list"]')
      .should('contain', 'open media')
      .and('contain', 'I can open Media from this thread')
    cy.get('[data-testid="chat-navigation-action"]')
      .should('be.visible')
      .and('contain', 'Open Media')
    cy.get('[data-testid="chat-navigation-reason"]').should(
      'contain',
      'read-only navigation action'
    )
    cy.get('[data-testid="chat-action-timeline"]')
      .should('contain', 'Action Timeline')
      .and('contain', 'navigate.open_surface')
      .and('contain', 'completed')
      .and('contain', '/media')
    cy.reload()
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/chats\/?$/)
    cy.get('[data-testid="chat-action-timeline-run"]')
      .should('contain', 'navigate.open_surface')
      .and('contain', '/media')
    cy.location('pathname').should('match', /^\/chats\/?$/)
    cy.get('[data-testid="chat-navigation-action-open"]').click()
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/media\/?$/)

    cy.then(() => {
      cy.request(
        `/api/chat/workflow-runs?profile_id=e2e-profile-001&thread_id=${encodeURIComponent(
          threadId
        )}`
      )
        .its('body')
        .should((payload) => {
          const serialized = JSON.stringify(payload)
          expect(serialized).to.include('chat.app_control.dispatch')
          expect(serialized).to.include('navigate.open_surface')
          expect(serialized).to.include('chats.main')
          expect(serialized).to.include('ask_before_local_changes')
          expect(serialized).to.include('/media')
        })
      cy.request('/api/chat/inbox?profile_id=e2e-profile-001')
        .its('body')
        .should((payload) => {
          expect(JSON.stringify(payload)).not.to.include(threadId)
        })
    })
  })

  it('CHATS-WORKSPACE-010/#1508 shows provider setup-needed guidance for unavailable provider-backed main Chat actions', () => {
    openChats()
    createThread('E2E Main Chat Provider Setup Needed')

    cy.intercept('POST', '/api/chat/messages').as('mainChatProviderSetup')
    cy.get('[data-testid="chat-compose-input"]').type(
      'generate listing content for this item'
    )
    cy.get('[data-testid="chat-send-button"]').click()

    cy.wait('@mainChatProviderSetup').then(({ request, response }) => {
      expect(request.body.profile_id).to.eq('e2e-profile-001')
      expect(request.body.content).to.eq(
        'generate listing content for this item'
      )
      expect(response?.statusCode).to.eq(201)
      expect(response?.body.app_control.capability_id).to.eq(
        'content_generate'
      )
      expect(response?.body.app_control.setup_needed).to.eq(true)
      expect(response?.body.app_control.workflow_run.confirmation_state).to.eq(
        'not_required'
      )
    })

    cy.get('[data-testid="chat-message-list"]')
      .should('contain', 'generate listing content for this item')
      .and('contain', 'needs provider setup')
    cy.get('[data-testid="chat-setup-needed-guidance"]')
      .should('be.visible')
      .and('contain', 'Provider setup is needed')
    cy.get('[data-testid="chat-navigation-action"]').should('not.exist')
    cy.location('pathname').should('match', /^\/chats\/?$/)
  })

  it('CHATS-WORKSPACE-009/#1503 dispatches normal main Chat text to a pending create-item preview without mutating early', () => {
    openChats()
    createThread('E2E Main Chat Preview Planner')

    const command =
      'create an inventory item CHAT-PLAN-1503 Planner Preview Coupe'
    cy.intercept('POST', '/api/chat/messages').as('mainChatPreviewPlanner')
    cy.get('[data-testid="chat-compose-input"]').type(command)
    cy.get('[data-testid="chat-send-button"]').click()

    let threadId = ''
    let previewId = ''
    cy.wait('@mainChatPreviewPlanner').then(({ request, response }) => {
      expect(request.body.profile_id).to.eq('e2e-profile-001')
      expect(String(request.body.thread_id).trim()).not.to.eq('')
      expect(request.body.content).to.eq(command)
      expect(request.body.context.route.pathname).to.eq('/chats/')
      expect(response?.statusCode).to.eq(201)
      expect(response?.body.assistant_handoff).to.eq(undefined)
      expect(response?.body.app_control.capability_id).to.eq(
        'inventory.item.create'
      )
      expect(response?.body.app_control.preview.action).to.eq(
        'create_inventory_item'
      )
      expect(response?.body.app_control.preview.payload.part_number).to.eq(
        'CHAT-PLAN-1503'
      )
      expect(response?.body.app_control.preview.payload.title).to.eq(
        'Planner Preview Coupe'
      )
      expect(response?.body.app_control.workflow_run.workflow_id).to.eq(
        'chat.app_control.dispatch'
      )
      expect(response?.body.app_control.workflow_run.confirmation_state).to.eq(
        'pending'
      )
      expect(response?.body.app_control.workflow_run.result.preview_id).to.eq(
        response?.body.app_control.preview.id
      )
      expect(
        response?.body.app_control.workflow_run.result.confirmation_required
      ).to.eq(true)
      threadId = String(request.body.thread_id)
      previewId = String(response?.body.app_control.preview.id)
    })

    cy.get('[data-testid="chat-message-list"]')
      .should('contain', command)
      .and('contain', 'I prepared a preview to create CHAT-PLAN-1503')
    cy.location('pathname').should('match', /^\/chats\/?$/)

    cy.then(() => {
      expect(previewId).not.to.eq('')
      cy.request(
        `/api/chat/workflow-runs?profile_id=e2e-profile-001&thread_id=${encodeURIComponent(
          threadId
        )}`
      )
        .its('body')
        .should((payload) => {
          const serialized = JSON.stringify(payload)
          expect(serialized).to.include('chat.app_control.dispatch')
          expect(serialized).to.include('inventory.item.create')
          expect(serialized).to.include(previewId)
          expect(serialized).to.include('confirmation_required')
        })
      cy.request('/api/items?profile_id=e2e-profile-001')
        .its('body')
        .should((items) => {
          const serialized = JSON.stringify(items)
          expect(serialized).not.to.include('CHAT-PLAN-1503')
          expect(serialized).not.to.include('Planner Preview Coupe')
        })
      cy.request('/api/chat/inbox?profile_id=e2e-profile-001')
        .its('body')
        .should((payload) => {
          expect(JSON.stringify(payload)).not.to.include(threadId)
        })
    })
  })

  it('CHATS-WORKSPACE-007/#1508 previews and confirms a main Chat inventory update without mutating early', () => {
    openChats()
    createThread('E2E Main Chat Item Update')

    cy.request('POST', '/api/items', {
      part_number: 'CHAT-UPDATE-001',
      title: 'Main Chat Original Title',
      brand: 'AFX',
      category: 'Slot Cars',
    }).then((createResponse) => {
      expect(createResponse.status).to.eq(201)
      const itemId = String(createResponse.body.id)
      expect(itemId).not.to.eq('')

      cy.get('[data-testid="chat-compose-input"]').type(
        'rename the open item through governed chat preview'
      )
      cy.get('[data-testid="chat-send-button"]').click()
      cy.get('[data-testid="chat-message-list"]').should(
        'contain',
        'rename the open item through governed chat preview'
      )

      cy.request('/api/chat/threads?profile_id=e2e-profile-001')
        .its('body.threads')
        .then((threads: Array<{ id: string; title?: string }>) => {
          const thread = threads.find(
            (candidate) => candidate.title === 'E2E Main Chat Item Update'
          )
          expect(thread?.id, 'created chat thread id').to.be.a('string').and.not.eq('')

          cy.request('POST', '/api/chat/actions/preview', {
            profile_id: 'e2e-profile-001',
            thread_id: thread?.id,
            action: 'update_inventory_item',
            payload: {
              item_id: itemId,
              part_number: 'CHAT-UPDATE-001-R1',
              title: 'Main Chat Updated Title',
              assistant_provider: 'openai',
              assistant_model: 'gpt-4o-mini',
            },
          }).then((previewResponse) => {
            expect(previewResponse.status).to.eq(200)
            const previewId = String(previewResponse.body.id)
            cy.visit(`/chats/?thread_id=${thread?.id}&preview_id=${previewId}`)
          })
        })

      cy.get('[data-testid="chat-action-preview"]')
        .should('contain', 'update_inventory_item')
        .and('contain', `target=${itemId}`)
        .and('contain', 'part_number=CHAT-UPDATE-001-R1')
        .and('contain', 'title=Main Chat Updated Title')

      cy.request('/api/items?profile_id=e2e-profile-001')
        .its('body')
        .should((items) => {
          const serialized = JSON.stringify(items)
          expect(serialized).to.include('Main Chat Original Title')
          expect(serialized).not.to.include('Main Chat Updated Title')
          expect(serialized).not.to.include('CHAT-UPDATE-001-R1')
        })

      cy.intercept('POST', '/api/chat/actions/apply').as('mainChatUpdateApply')
      cy.get('[data-testid="chat-apply-action-button"]').click()
      cy.get('[data-testid="chat-apply-confirm-dialog"]').should('be.visible')
      cy.get('[data-testid="chat-apply-confirm-summary"]')
        .should('contain', 'update_inventory_item')
        .and('contain', itemId)
        .and('contain', 'Main Chat Updated Title')
      cy.get('[data-testid="chat-apply-confirm-submit"]').click()
      cy.wait('@mainChatUpdateApply').then(({ request, response }) => {
        expect(request.body.profile_id).to.eq('e2e-profile-001')
        expect(request.body.confirm).to.eq(true)
        expect(response?.statusCode).to.eq(200)
        expect(response?.body.applied).to.eq(true)
        expect(response?.body.action).to.eq('update_inventory_item')
        expect(response?.body.item_id).to.eq(itemId)
        expect(response?.body.part_number).to.eq('CHAT-UPDATE-001-R1')
        expect(response?.body.title).to.eq('Main Chat Updated Title')
      })
      cy.get('[data-testid="chat-action-apply-result"]')
        .should('contain', 'Applied update_inventory_item')
        .and('contain', itemId)
        .and('contain', 'part_number=CHAT-UPDATE-001-R1')
        .and('contain', 'title=Main Chat Updated Title')
      cy.get('[data-testid="chat-action-timeline"]')
        .should('contain', 'Action Timeline')
        .and('contain', 'update_inventory_item')
        .and('contain', 'confirmed')
      cy.get('[data-testid="chat-action-timeline-run"]')
        .last()
        .should('have.attr', 'data-workflow-status', 'completed')
      cy.get('[data-testid="chat-message-list"]')
        .should('contain', 'Applied update_inventory_item')
        .and('contain', 'Main Chat Updated Title')

      cy.request('/api/items?profile_id=e2e-profile-001')
        .its('body')
        .should((items) => {
          const serialized = JSON.stringify(items)
          expect(serialized).to.include('CHAT-UPDATE-001-R1')
          expect(serialized).to.include('Main Chat Updated Title')
          expect(serialized).not.to.include('Main Chat Original Title')
        })
    })
  })
})
