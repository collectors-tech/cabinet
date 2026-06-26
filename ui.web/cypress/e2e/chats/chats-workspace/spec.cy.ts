describe('chats/chats-workspace', () => {
  function openChats() {
    cy.e2eReset()
    cy.e2eBootstrap({ minimalProfile: true })
    cy.e2eSetSetupState('present')
    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', { path: '/chats/' })
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/chats\/?$/)
  }

  function createThread(title: string) {
    cy.get('[data-testid="chat-new-thread-input"]').clear().type(title)
    cy.get('[data-testid="chat-create-thread-button"]').click()
    cy.contains('[data-testid="chat-thread-item"]', title).click()
    cy.get('[data-testid="chat-thread-title"]').should('contain', title)
  }

  it('CHATS-WORKSPACE-001 renders Cabinet-specific chats semantics instead of placeholder inbox copy', () => {
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
        'Use Assistant for AI-guided help and actions; use Chats for durable conversation threads.'
      )
    cy.get('[data-testid="chat-thread-list"]').should('be.visible')
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

  it('CHATS-WORKSPACE-003 states the Assistant versus Chats boundary explicitly', () => {
    openChats()

    cy.get('[data-testid="chat-workspace-boundary-note"]')
      .should('contain.text', 'Assistant')
      .and('contain.text', 'Chats')
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

    cy.get('[data-testid="chat-tool-card-container"]').should(
      'have.attr',
      'data-visual-priority',
      'secondary'
    )
  })

  it('CHATS-WORKSPACE-005 filters thread rows and keeps new-thread actions route-stable', () => {
    openChats()

    cy.get('[data-testid="chat-empty-workspace-action"]').click()
    cy.get('[data-testid="chat-new-thread-input"]').should('not.be.disabled')
    cy.location('pathname').should('match', /^\/chats\/?$/)

    createThread('E2E Alpha Search Thread')
    cy.get('[data-testid="chat-new-chat-button"]').click()
    cy.get('[data-testid="chat-new-thread-input"]').should('not.be.disabled')
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

  it('CHATS-WORKSPACE-008/#1503 dispatches normal main Chat text to app-control route planning without Inbox noise', () => {
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
    cy.location('pathname').should('match', /^\/chats\/?$/)

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
          expect(serialized).to.include('/media')
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

      cy.get('[data-testid="chat-preview-action-mode"]').select(
        'update_inventory_item'
      )
      cy.get('[data-testid="chat-preview-target-item-id"]')
        .clear()
        .type(itemId)
      cy.get('[data-testid="chat-preview-part-number"]')
        .clear()
        .type('CHAT-UPDATE-001-R1')
      cy.get('[data-testid="chat-preview-title"]')
        .clear()
        .type('Main Chat Updated Title')
      cy.intercept('POST', '/api/chat/actions/preview').as(
        'mainChatUpdatePreview'
      )
      cy.get('[data-testid="chat-preview-action-button"]').click()
      cy.wait('@mainChatUpdatePreview').then(({ request, response }) => {
        expect(request.body.profile_id).to.eq('e2e-profile-001')
        expect(request.body.action).to.eq('update_inventory_item')
        expect(request.body.payload.item_id).to.eq(itemId)
        expect(request.body.payload.part_number).to.eq('CHAT-UPDATE-001-R1')
        expect(request.body.payload.title).to.eq('Main Chat Updated Title')
        expect(response?.statusCode).to.eq(200)
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
