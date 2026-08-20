describe('chats/ui-screen-chat-copilot', () => {
  function signInToChats() {
    cy.e2eReset()
    cy.e2eBootstrap({ minimalProfile: true }).then((bootstrap) => {
      cy.request('PUT', '/api/profiles/active', { profile_id: bootstrap.profile_id })
        .its('status')
        .should('eq', 200)
      cy.visit('/sign-in?redirect=%2Fchats%2F', {
        onBeforeLoad(win) {
          win.localStorage.setItem(`cabinet.workspace.${bootstrap.profile_id}`, '1')
        },
      })
      cy.contains('button', 'Open local workspace').click()
      cy.get('body').then(($body) => {
        const profileButton = `Use ${bootstrap.profile_name}`
        if ($body.text().includes(profileButton)) {
          cy.contains('button', profileButton).click()
        }
      })
    })
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/chats\/?$/)
  }

  function openInventory() {
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/inventory/',
      shellWorkspace: 'navigation',
    })
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)
  }

  function openChats() {
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', { path: '/chats/' })
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/chats\/?$/)
  }

  function openChatsWithAssistantDefaults(provider: string, model: string) {
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.request('PUT', '/api/profiles/e2e-profile-001/settings', {
      settings: {
        assistant_default_provider: provider,
        assistant_default_model: model,
      },
    })
      .its('status')
      .should('eq', 200)
    cy.e2eSetSetupState('present')
    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', { path: '/chats/' })
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/chats\/?$/)
  }

  function openInbox() {
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', { path: '/inbox/' })
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inbox\/?$/)
  }

  function createThread(title: string) {
    cy.get('[data-testid="chat-new-chat-button"]')
      .should('be.visible')
      .and('not.be.disabled')
      .click()
    cy.get('[data-testid="chat-new-thread-dialog"]').should('be.visible')
    cy.get('[data-testid="chat-new-thread-input"]').clear().type(title)
    cy.get('[data-testid="chat-create-thread-button"]').click()
    cy.get('[data-testid="chat-thread-title"]').should('contain', title)
  }

  function openActionPreview(
    threadTitle: string,
    action: string,
    payload: Record<string, string>
  ) {
    cy.request('/api/chat/threads?profile_id=e2e-profile-001')
      .its('body.threads')
      .then((threads: Array<{ id: string; title: string }>) => {
        const thread = threads.find((candidate) => candidate.title === threadTitle)
        expect(thread?.id, `thread ${threadTitle}`).to.be.a('string').and.not.eq('')
        cy.request('POST', '/api/chat/actions/preview', {
          profile_id: 'e2e-profile-001',
          thread_id: thread?.id,
          action,
          payload,
        }).then(({ body, status }) => {
          expect(status).to.eq(200)
          cy.visit(
            `/chats/?thread_id=${encodeURIComponent(String(thread?.id))}&preview_id=${encodeURIComponent(String(body.id))}`
          )
        })
      })
    cy.get('[data-testid="chat-action-preview"]', { timeout: 20000 })
      .scrollIntoView()
      .should('be.visible')
  }

  it('CHAT-COPILOT-001 toggles assistant workspace from the header without route context loss', () => {
    openInventory()
    cy.get('[data-testid="active-profile-name"]').should(
      'contain',
      'E2E Local'
    )
    cy.get('[data-testid="shell-workspace-navigation"]')
      .should('have.attr', 'data-active', 'true')
    cy.location('pathname').then((initialPathname) => {
      cy.get('[data-testid="shell-chat-toggle"]')
        .invoke('attr', 'aria-label')
        .should('eq', 'Open Cabinet Agent')
      cy.get('[data-testid="shell-chat-toggle"]').click({ force: true })
      cy.get('[data-testid="shell-assistant-workspace"]').should('exist')
      cy.get('[data-testid="shell-workspace-assistant"]')
        .should('have.attr', 'data-active', 'true')
      cy.get('[data-testid="shell-assistant-route-context"]')
        .invoke('text')
        .should('contain', '/inventory')
      cy.location('pathname').should('eq', initialPathname)

      cy.get('[data-testid="shell-chat-toggle"]')
        .invoke('attr', 'aria-label')
        .should('eq', 'Close Cabinet Agent')
      cy.get('[data-testid="shell-chat-toggle"]').click({ force: true })
      cy.get('[data-testid="shell-assistant-workspace"]').should('not.exist')
      cy.get('[data-testid="shell-workspace-navigation"]')
        .should('have.attr', 'data-active', 'true')
      cy.location('pathname').should('eq', initialPathname)
    })
  })

  it('UI-SCREEN-CHAT-COPILOT-002 disables thread creation while chat service is unavailable', () => {
    cy.intercept('GET', '/api/profiles/active', {
      statusCode: 404,
      body: { error: 'active_profile_404' },
    }).as('activeProfileMissing')
    cy.intercept('POST', '/api/chat/threads', {
      statusCode: 200,
      body: {
        id: 'should-not-create',
        profile_id: 'missing-profile',
        title: 'Should not create',
        created_at: '2026-03-14T00:00:00Z',
        updated_at: '2026-03-14T00:00:00Z',
      },
    }).as('createThread')

    signInToChats()
    cy.wait('@activeProfileMissing')
    cy.get('[data-testid="chat-bootstrap-error"]').should('be.visible')
    cy.contains('active_profile_404').should('be.visible')
    cy.get('[data-testid="chat-empty-workspace-action"]').should('be.disabled')
    cy.get('[data-testid="chat-new-thread-dialog"]').should('not.exist')
    cy.get('@createThread.all').should('have.length', 0)
  })

  it('UI-SCREEN-CHAT-COPILOT-007 keeps Preview Action gated until thread context exists', () => {
    openChats()
    createThread('E2E Empty Thread Preview Gate')

    cy.contains('No messages in this thread yet.').should('be.visible')
    cy.get('[data-testid="chat-preview-action-button"]').should('not.exist')
    cy.get('[data-testid="chat-action-preview"]').should('not.exist')
  })

  it('UI-SCREEN-CHAT-COPILOT-017 renders full Chats route through assistant-ui primitives while preserving Cabinet APIs', () => {
    openChatsWithAssistantDefaults('anthropic', 'claude-3-7-sonnet')
    createThread('E2E Chats Assistant UI Thread')
    cy.intercept('POST', '/api/chat/messages').as('chatMessage')

    cy.get('[data-testid="chat-message-list"]')
      .should('have.attr', 'data-message-count', '0')
    cy.get('[data-testid="chat-assistant-ui-composer-primitive"]').should(
      'be.visible'
    )
    cy.get('[data-testid="chat-composer-attachment-button"]').should(
      'be.visible'
    )
    cy.get('[data-testid="chat-preview-action-button"]').should('not.exist')
    cy.get('[data-testid="chat-tool-card-container"]').should('not.exist')

    cy.get('[data-testid="chat-compose-input"]').type(
      'route assistant ui composer send'
    )
    cy.get('[data-testid="chat-send-button"]').click()
    cy.wait('@chatMessage').then(({ request }) => {
      expect(request.body.profile_id).to.eq('e2e-profile-001')
      expect(String(request.body.thread_id).trim()).not.to.eq('')
      expect(request.body.role).to.eq('user')
      expect(request.body.content).to.eq('route assistant ui composer send')
      expect(request.body.context.route.pathname).to.eq('/chats/')
      expect(request.body.context.profile.id).to.eq('e2e-profile-001')
      expect(request.body.context.assistant.provider).to.eq('anthropic')
      expect(request.body.context.assistant.model).to.eq('claude-3-7-sonnet')
    })
    cy.get('[data-testid="chat-message-list"]')
      .should('contain', 'route assistant ui composer send')
      .invoke('attr', 'data-message-count')
      .then((count) => {
        expect(Number(count)).to.be.greaterThan(0)
      })
    cy.get('[data-testid="chat-assistant-ui-message-primitive"]').should(
      'have.length.at.least',
      1
    )
  })

  it('CHATS-WORKSPACE-011/#2313 renders an ordinary selected-provider reply without deterministic fallback', () => {
    openChatsWithAssistantDefaults('fake', 'cabinet-e2e-direct')
    createThread('E2E Direct Provider Conversation')
    cy.intercept('POST', '/api/chat/messages').as('directProviderMessage')

    cy.get('[data-testid="chat-compose-input"]').type(
      'Tell me something helpful about my Cabinet workspace.'
    )
    cy.get('[data-testid="chat-send-button"]').click()

    cy.wait('@directProviderMessage').then(({ response }) => {
      expect(response?.statusCode).to.eq(201)
      expect(response?.body.assistant_response.mode).to.eq('provider')
      expect(response?.body.assistant_response.provider).to.eq('fake')
      expect(response?.body.assistant_response.model).to.eq(
        'cabinet-e2e-direct'
      )
      expect(response?.body.assistant_response.thread_message.content).to.eq(
        'E2E direct provider response'
      )
      expect(
        response?.body.assistant_response.provider_trace
          .cabinet_tool_authority
      ).to.eq('none')
    })

    cy.get('[data-testid="chat-message-list"]')
      .should('contain', 'E2E direct provider response')
      .and('not.contain', 'I can help with Cabinet inventory')
    cy.get('[data-testid="chat-tool-card-container"]').should('not.exist')
  })

  it('CHATS-WORKSPACE-014/#2329 keeps literal response text containing an Agent action word in provider Chat', () => {
    openChatsWithAssistantDefaults('fake', 'cabinet-e2e-direct')
    createThread('E2E Literal Response Routing')
    cy.intercept('POST', '/api/chat/messages').as('literalProviderMessage')

    cy.get('[data-testid="chat-compose-input"]').type(
      'Reply with exactly: CABINET_BROWSER_AUTH_AFTER_RESTORE_OK'
    )
    cy.get('[data-testid="chat-send-button"]').click()

    cy.wait('@literalProviderMessage').then(({ response }) => {
      expect(response?.statusCode).to.eq(201)
      expect(response?.body.agent_planner).to.eq(undefined)
      expect(response?.body.assistant_response.mode).to.eq('provider')
      expect(response?.body.assistant_response.provider).to.eq('fake')
      expect(response?.body.assistant_response.thread_message.content).to.eq(
        'E2E direct provider response'
      )
    })

    cy.get('[data-testid="chat-message-list"]')
      .should('contain', 'E2E direct provider response')
      .and('not.contain', 'Cabinet does not support that Agent request')
    cy.get('[data-testid="chat-tool-card-container"]').should('not.exist')
  })

  it('UI-SCREEN-CHAT-COPILOT-019 renders assistant-ui dark shell structure', () => {
    cy.viewport(1400, 900)
    openChatsWithAssistantDefaults('anthropic', 'claude-3-7-sonnet')
    createThread('E2E Dark Shell Thread')

    cy.get('[data-testid="chat-layout"]').should('be.visible')
    cy.get('[data-testid="chat-conversation-rail"]')
      .should('be.visible')
      .and('contain', 'Cabinet Agent')
    cy.get('[data-testid="chat-new-thread-action"]')
      .should('be.visible')
      .and('contain', 'New Thread')
    cy.get('[data-testid="chat-main-topbar"]').should('be.visible')
    cy.get('[data-testid="chat-new-chat-button"]')
      .should('be.visible')
      .and('contain', 'New Chat')
    cy.get('[data-testid="chat-share-export-button"]')
      .should('be.visible')
      .and('have.attr', 'aria-label', 'Share or export chat')
    cy.get('[data-testid="chat-main-canvas"]').should('be.visible')
    cy.get('[data-testid="chat-empty-thread-state"]')
      .should('be.visible')
      .and('contain', 'How can I help you today?')
      .and('contain', 'No messages in this thread yet.')
    cy.get('[data-testid="chat-composer-shell"]').should('be.visible')
    cy.get('[data-testid="chat-compose-input"]').should(
      'have.attr',
      'placeholder',
      'Send a message... (@ to mention, / for commands)'
    )
    cy.get('[data-testid="chat-composer-attachment-button"]')
      .should('be.visible')
      .and('contain', 'Attach')
    cy.get('[data-testid="chat-model-selector-row"]')
      .should('contain', 'anthropic')
      .and('contain', 'claude-3-7-sonnet')
    cy.get('[data-testid="chat-voice-control"]')
      .should('be.visible')
      .and('be.disabled')
    cy.get('[data-testid="chat-prompt-chips"]').within(() => {
      for (const chip of ['Weather', 'Code', 'Write', 'Analyze', 'Brainstorm']) {
        cy.contains('button', chip).should('be.visible')
      }
    })
    cy.contains('p', 'Attachments').should('not.exist')
    cy.contains('p', 'Action Preview').should('not.exist')
    cy.get('[data-testid="chat-tool-card-container"]').should('not.exist')
    cy.get('[data-testid="chat-upload-attachment-button"]').should('not.exist')
  })

  it('UI-SCREEN-CHAT-COPILOT-008 supports confirm-before-apply for inventory and wishlist mutations', () => {
    openChats()
    const threadTitle = 'E2E Copilot CRUD Thread'
    createThread(threadTitle)

    cy.get('[data-testid="chat-compose-input"]').clear().type('Please create this item')
    cy.get('[data-testid="chat-send-button"]').click()
    cy.get('[data-testid="chat-message-list"]').should('contain', 'Please create this item')

    openActionPreview(threadTitle, 'create_inventory_item', {
      part_number: 'CP-007-INV',
      title: 'Copilot Inventory Create',
    })
    cy.get('[data-testid="chat-action-preview"]').should('contain', 'create_inventory_item')
    cy.get('[data-testid="chat-apply-action-button"]').click()
    cy.get('[data-testid="chat-apply-confirm-dialog"]').should('be.visible')
    cy.get('[data-testid="chat-apply-confirm-summary"]').should('contain', 'create_inventory_item')
    cy.get('[data-testid="chat-apply-confirm-submit"]').click()
    cy.get('[data-testid="chat-action-apply-result"]').should('contain', 'create_inventory_item')
    cy.request('/api/items?profile_id=e2e-profile-001').then((response) => {
      expect(response.status).to.eq(200)
      const items = response.body.items as Array<{
        id?: string
        part_number?: string
        title?: string
      }>
      const created = items.find((item) => item.part_number === 'CP-007-INV')
      expect(created, 'created inventory item').to.not.equal(undefined)

      openActionPreview(threadTitle, 'update_inventory_item', {
        item_id: created?.id ?? '',
        part_number: 'CP-007-INV-UPD',
        title: 'Copilot Inventory Updated',
      })
      cy.get('[data-testid="chat-action-preview"]')
        .should('contain', 'update_inventory_item')
        .and('contain', 'CP-007-INV-UPD')
      cy.get('[data-testid="chat-apply-action-button"]').click()
      cy.get('[data-testid="chat-apply-confirm-summary"]')
        .should('contain', 'Apply update_inventory_item')
        .and('contain', 'part_number=CP-007-INV-UPD')
        .and('contain', 'title=Copilot Inventory Updated')
      cy.get('[data-testid="chat-apply-confirm-submit"]').click()
      cy.get('[data-testid="chat-action-apply-result"]')
        .should('contain', 'update_inventory_item')
        .and('contain', 'part_number=CP-007-INV-UPD')
        .and('contain', 'title=Copilot Inventory Updated')
      cy.get('[data-testid="chat-message-list"]')
        .should('contain', 'Applied update_inventory_item')
        .and('contain', 'part_number=CP-007-INV-UPD')
        .and('contain', 'title=Copilot Inventory Updated')
    })

    openActionPreview(threadTitle, 'create_wishlist_entry', {
      part_number: 'CP-007-WISH',
      title: 'Copilot Wishlist Create',
    })
    cy.get('[data-testid="chat-action-preview"]').should('contain', 'create_wishlist_entry')
    cy.get('[data-testid="chat-apply-action-button"]').click()
    cy.get('[data-testid="chat-apply-confirm-dialog"]').should('be.visible')
    cy.get('[data-testid="chat-apply-confirm-submit"]').click()
    cy.get('[data-testid="chat-action-apply-result"]').should('contain', 'create_wishlist_entry')
    cy.request('/api/chat/threads?profile_id=e2e-profile-001').then((threadsResponse) => {
      expect(threadsResponse.status).to.eq(200)
      const threads = threadsResponse.body.threads as Array<{
        id: string
        title: string
      }>
      const thread = threads.find((item) => item.title === threadTitle)
      expect(thread, 'created chat thread').to.not.equal(undefined)
      cy.request(
        `/api/chat/messages?profile_id=e2e-profile-001&thread_id=${thread?.id}`
      ).then((messagesResponse) => {
        expect(messagesResponse.status).to.eq(200)
        const messages = messagesResponse.body.messages as Array<{
          content?: string
          role?: string
        }>
        expect(
          messages.some(
            (message) =>
              message.role === 'assistant' &&
              String(message.content ?? '').includes(
                'Applied create_wishlist_entry to wishlist'
              ) &&
              String(message.content ?? '').includes('for item')
          ),
          'wishlist assistant outcome links entry and item'
        ).to.eq(true)
      })
    })
    cy.request('/api/items?status=wishlist').then((response) => {
      expect(response.status).to.eq(200)
      const items = response.body.items as Array<{
        part_number?: string
        status?: string
        title?: string
      }>
      const created = items.find((item) => item.part_number === 'CP-007-WISH')
      expect(created, 'created wishlist-backed item').to.not.equal(undefined)
      expect(created?.status).to.eq('wishlist')
      expect(created?.title).to.eq('Copilot Wishlist Create')
    })
  })

  it('UI-SCREEN-CHAT-COPILOT-012 reflects assistant provider defaults in chat action previews', () => {
    openChatsWithAssistantDefaults('anthropic', 'claude-3-7-sonnet')
    createThread('E2E Copilot Provider Defaults Thread')

    cy.get('[data-testid="chat-model-selector-row"]')
      .should('contain', 'anthropic')
      .and('contain', 'claude-3-7-sonnet')
    cy.get('[data-testid="chat-compose-input"]').clear().type('Draft this with the active assistant defaults')
    cy.get('[data-testid="chat-send-button"]').click()
    cy.get('[data-testid="chat-message-list"]').should(
      'contain',
      'Draft this with the active assistant defaults'
    )

    openActionPreview('E2E Copilot Provider Defaults Thread', 'create_inventory_item', {
      part_number: 'CP-012-PROVIDER',
      title: 'Provider Default Preview',
      assistant_provider: 'anthropic',
      assistant_model: 'claude-3-7-sonnet',
    })
    cy.get('[data-testid="chat-action-preview"]').should(
      'contain',
      'anthropic / claude-3-7-sonnet'
    )
    cy.get('[data-testid="chat-apply-action-button"]').click()
    cy.get('[data-testid="chat-apply-confirm-summary"]').should(
      'contain',
      'assistant=anthropic/claude-3-7-sonnet'
    )
  })

  it('UI-SCREEN-CHAT-COPILOT-013 previews structured collection assignment targets before apply', () => {
    openChatsWithAssistantDefaults('openai', 'gpt-4.1-mini')
    createThread('E2E Copilot Collection Preview Thread')

    cy.get('[data-testid="chat-compose-input"]')
      .clear()
      .type('Assign this item to the retail collection')
    cy.get('[data-testid="chat-send-button"]').click()
    cy.get('[data-testid="chat-message-list"]').should(
      'contain',
      'Assign this item to the retail collection'
    )

    openActionPreview(
      'E2E Copilot Collection Preview Thread',
      'assign_collection_item',
      {
        item_id: 'e2e-item-001',
        part_number: 'CP-013-COLLECT',
        title: 'E2E Starter Car',
        collection_name: 'Store 1',
        assistant_provider: 'openai',
        assistant_model: 'gpt-4.1-mini',
      }
    )

    cy.get('[data-testid="chat-action-preview"]')
      .should('contain', 'assign_collection_item')
      .and('contain', 'e2e-item-001')
      .and('contain', 'Store 1')
      .and('contain', 'openai / gpt-4.1-mini')
    cy.get('[data-testid="chat-apply-action-button"]').click()
    cy.get('[data-testid="chat-apply-confirm-summary"]')
      .should('contain', 'Apply assign_collection_item')
      .and('contain', 'target=e2e-item-001')
      .and('contain', 'collection=Store 1')
      .and('contain', 'assistant=openai/gpt-4.1-mini')
    cy.get('[data-testid="chat-apply-confirm-submit"]').click()
    cy.get('[data-testid="chat-action-apply-result"]')
      .should('contain', 'Applied assign_collection_item')
      .and('contain', 'collection Store 1')
      .and('contain', 'e2e-item-001')
    cy.get('[data-testid="chat-message-list"]').should(
      'contain',
      'Applied assign_collection_item to e2e-item-001 in Store 1.'
    )

    cy.visit('/collections')
    cy.get('[data-testid="collections-active-context"]').should(
      'contain',
      'Store 1'
    )
    cy.get('[data-testid="collections-row-count-store-1"]').should(
      'have.text',
      '1'
    )
  })

  it('UI-SCREEN-CHAT-COPILOT-011 cancels preview apply without mutating inventory and records history', () => {
    openChats()
    createThread('E2E Copilot Cancel Apply Thread')

    cy.get('[data-testid="chat-compose-input"]').clear().type('Draft this item only')
    cy.get('[data-testid="chat-send-button"]').click()
    cy.get('[data-testid="chat-message-list"]').should('contain', 'Draft this item only')

    openActionPreview('E2E Copilot Cancel Apply Thread', 'update_inventory_item', {
      item_id: 'e2e-item-001',
      part_number: 'CP-011-CANCEL',
      title: 'Copilot Cancel Preview',
    })
    cy.get('[data-testid="chat-action-preview"]')
      .should('contain', 'update_inventory_item')
      .and('contain', 'pending')
    cy.get('[data-testid="chat-cancel-action-button"]').click()

    cy.get('[data-testid="chat-apply-confirm-dialog"]').should('not.exist')
    cy.get('[data-testid="chat-action-preview"]')
      .should('contain', 'update_inventory_item')
      .and('contain', 'cancelled')
    cy.get('[data-testid="chat-apply-action-button"]').should('be.disabled')
    cy.get('[data-testid="chat-action-apply-notice"]').should(
      'contain',
      'Action apply canceled; no mutation applied.'
    )
    cy.get('[data-testid="chat-action-apply-result"]').should('not.exist')
    cy.get('[data-testid="chat-message-list"]')
      .should('contain', 'Canceled update_inventory_item')
      .and('contain', 'no mutation applied')
      .and('not.contain', 'Applied update_inventory_item')
    cy.request('/api/items?profile_id=e2e-profile-001').then((response) => {
      expect(response.status).to.eq(200)
      const items = response.body.items as Array<{
        id?: string
        part_number?: string
        title?: string
      }>
      const target = items.find((item) => item.id === 'e2e-item-001')
      expect(target, 'existing target item').to.not.equal(undefined)
      expect(target?.part_number).to.not.eq('CP-011-CANCEL')
      expect(target?.title).to.not.eq('Copilot Cancel Preview')
    })
  })

  it('UI-SCREEN-CHAT-COPILOT-014 keeps failed update apply pending without false history', () => {
    openChats()
    createThread('E2E Copilot Failed Update Thread')

    cy.get('[data-testid="chat-compose-input"]')
      .clear()
      .type('Try to update a missing inventory item')
    cy.get('[data-testid="chat-send-button"]').click()
    cy.get('[data-testid="chat-message-list"]').should(
      'contain',
      'Try to update a missing inventory item'
    )

    cy.request('/api/items?profile_id=e2e-profile-001').then((response) => {
      expect(response.status).to.eq(200)
      const items = response.body.items as Array<{
        id?: string
        part_number?: string
      }>
      expect(
        items.some(
          (item) =>
            item.id === 'missing-chat-update-target' ||
            item.part_number === 'CP-014-MISSING'
        )
      ).to.eq(false)
    })

    openActionPreview('E2E Copilot Failed Update Thread', 'update_inventory_item', {
      item_id: 'missing-chat-update-target',
      part_number: 'CP-014-MISSING',
      title: 'Missing Update Target',
    })
    cy.get('[data-testid="chat-action-preview"]')
      .should('contain', 'update_inventory_item')
      .and('contain', 'missing-chat-update-target')
      .and('contain', 'pending')

    cy.get('[data-testid="chat-apply-action-button"]').click()
    cy.get('[data-testid="chat-apply-confirm-summary"]')
      .should('contain', 'Apply update_inventory_item')
      .and('contain', 'target=missing-chat-update-target')
    cy.get('[data-testid="chat-apply-confirm-submit"]').click()

    cy.get('[data-testid="chat-apply-confirm-dialog"]').should('not.exist')
    cy.get('[data-testid="chat-action-preview"]')
      .should('contain', 'update_inventory_item')
      .and('contain', 'pending')
    cy.get('[data-testid="chat-action-apply-notice"]').should(
      'contain',
      'Action apply failed; preview remains pending.'
    )
    cy.get('[data-testid="chat-action-apply-result"]').should('not.exist')
    cy.get('[data-testid="chat-message-list"]').should(
      'not.contain',
      'Applied update_inventory_item'
    )
    cy.request('/api/items?profile_id=e2e-profile-001').then((response) => {
      expect(response.status).to.eq(200)
      const items = response.body.items as Array<{
        id?: string
        part_number?: string
      }>
      expect(
        items.some(
          (item) =>
            item.id === 'missing-chat-update-target' ||
            item.part_number === 'CP-014-MISSING'
        )
      ).to.eq(false)
    })
  })

  it('UI-SCREEN-CHAT-COPILOT-015 clears pending action state when thread context changes', () => {
    openChats()
    createThread('E2E Copilot Thread Context A')

    cy.get('[data-testid="chat-compose-input"]')
      .clear()
      .type('Draft a preview that must stay scoped to this thread')
    cy.get('[data-testid="chat-send-button"]').click()
    cy.get('[data-testid="chat-message-list"]').should(
      'contain',
      'Draft a preview that must stay scoped to this thread'
    )

    openActionPreview('E2E Copilot Thread Context A', 'create_inventory_item', {
      part_number: 'CP-015-STALE',
      title: 'Thread Scoped Preview',
    })
    cy.get('[data-testid="chat-action-preview"]')
      .should('contain', 'create_inventory_item')
      .and('contain', 'CP-015-STALE')
    cy.get('[data-testid="chat-cancel-action-button"]').click()
    cy.get('[data-testid="chat-action-apply-notice"]').should(
      'contain',
      'Action apply canceled; no mutation applied.'
    )

    createThread('E2E Copilot Thread Context B')
    cy.get('[data-testid="chat-thread-title"]').should(
      'contain',
      'E2E Copilot Thread Context B'
    )
    cy.get('[data-testid="chat-action-preview"]').should('not.exist')
    cy.get('[data-testid="chat-action-apply-notice"]').should('not.exist')
    cy.get('[data-testid="chat-action-apply-result"]').should('not.exist')
    cy.get('[data-testid="chat-apply-action-button"]').should('be.disabled')
  })

  it('UI-SCREEN-CHAT-COPILOT-016 restores pending action preview after route return and reload', () => {
    openChats()
    createThread('E2E Copilot Route Return Thread')

    cy.get('[data-testid="chat-compose-input"]')
      .clear()
      .type('Keep this pending preview available while I check another route')
    cy.get('[data-testid="chat-send-button"]').click()
    cy.get('[data-testid="chat-message-list"]').should(
      'contain',
      'Keep this pending preview available while I check another route'
    )

    openActionPreview('E2E Copilot Route Return Thread', 'create_inventory_item', {
      part_number: 'CP-016-RETURN',
      title: 'Route Return Preview',
    })
    cy.get('[data-testid="chat-action-preview"]')
      .should('contain', 'create_inventory_item')
      .and('contain', 'CP-016-RETURN')
      .and('contain', 'pending')

    cy.visit('/inventory')
    cy.location('pathname', { timeout: 15000 }).should(
      'match',
      /^\/inventory\/?$/
    )
    cy.visit('/chats')
    cy.get('[data-testid="chat-thread-title"]').should(
      'contain',
      'E2E Copilot Route Return Thread'
    )
    cy.get('[data-testid="chat-action-preview"]')
      .should('contain', 'create_inventory_item')
      .and('contain', 'CP-016-RETURN')
      .and('contain', 'pending')

    cy.reload()
    cy.get('[data-testid="chat-thread-title"]').should(
      'contain',
      'E2E Copilot Route Return Thread'
    )
    cy.get('[data-testid="chat-action-preview"]')
      .should('contain', 'create_inventory_item')
      .and('contain', 'CP-016-RETURN')
      .and('contain', 'pending')
    cy.get('[data-testid="chat-apply-action-button"]').should('not.be.disabled')
  })

  it('UI-SCREEN-CHAT-COPILOT-009 supports mobile image attachment and confirm-before-apply flow once message context exists', () => {
    cy.viewport(390, 844)
    openChats()
    createThread('E2E Mobile Copilot Thread')
    cy.intercept('POST', '/api/chat/attachments').as('chatAttachment')
    cy.intercept('POST', '/api/chat/messages').as('chatMessage')

    let removedAttachmentId = ''
    let keptAttachmentId = ''

    cy.get('[data-testid="chat-attachment-input"]').selectFile(
      {
        contents: Cypress.Buffer.from('remove before send'),
        fileName: 'mobile-remove-me.txt',
        mimeType: 'text/plain',
      },
      { force: true }
    )
    cy.wait('@chatAttachment').then(({ response }) => {
      expect(response?.statusCode).to.eq(201)
      expect(response?.body.filename).to.eq('mobile-remove-me.txt')
      removedAttachmentId = String(response?.body.id)
    })
    cy.get('[data-testid="chat-attachment-list"]').should(
      'contain',
      'mobile-remove-me.txt'
    )
    cy.get('[data-testid="chat-remove-attachment-button"]').click({
      force: true,
    })
    cy.get('[data-testid="chat-attachment-list"]').should('not.exist')

    cy.get('[data-testid="chat-attachment-input"]').selectFile(
      'public/images/favicon.png',
      { force: true }
    )
    cy.wait('@chatAttachment').then(({ response }) => {
      expect(response?.statusCode).to.eq(201)
      expect(response?.body.filename).to.eq('favicon.png')
      keptAttachmentId = String(response?.body.id)
      expect(keptAttachmentId).not.to.eq(removedAttachmentId)
    })
    cy.get('[data-testid="chat-attachment-list"]').should('contain', 'favicon.png')
    cy.get('[data-testid="chat-compose-input"]')
      .clear()
      .type('Use the attached photo to create an item')
    cy.get('[data-testid="chat-send-button"]').click()
    cy.wait('@chatMessage').then(({ request, response }) => {
      expect(request.body.attachment_ids).to.deep.eq([keptAttachmentId])
      expect(request.body.attachment_ids).not.to.include(removedAttachmentId)
      expect(response?.statusCode).to.eq(201)
    })
    cy.get('[data-testid="chat-message-list"]').should(
      'contain',
      'Use the attached photo to create an item'
    )

    openActionPreview('E2E Mobile Copilot Thread', 'create_inventory_item', {
      part_number: 'CP-008-MOBILE',
      title: 'Mobile Image Suggestion',
    })
    cy.get('[data-testid="chat-action-preview"]').should('be.visible')
    cy.get('[data-testid="chat-apply-action-button"]').click()
    cy.get('[data-testid="chat-apply-confirm-dialog"]').should('be.visible')
    cy.get('[data-testid="chat-apply-confirm-summary"]').should('contain', 'CP-008-MOBILE')
    cy.get('[data-testid="chat-apply-confirm-submit"]').click()
    cy.get('[data-testid="chat-action-apply-result"]').should('contain', 'CP-008-MOBILE')
  })

  it('UI-SCREEN-CHAT-COPILOT-009 keeps top-level /inbox reachable as a communications surface', () => {
    openInbox()
    cy.get('[data-testid="notification-inbox-page"]').should('be.visible')
    cy.get('[data-testid="notification-inbox-header-title"]').should(
      'contain',
      'Notification Inbox'
    )
    cy.get('[data-testid="purchase-inbox-load-reviews"]').should('not.exist')
    cy.contains('404').should('not.exist')
    cy.contains('Oops! Page Not Found!').should('not.exist')
  })
})
