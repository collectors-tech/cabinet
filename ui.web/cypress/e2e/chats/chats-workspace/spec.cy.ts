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
})
