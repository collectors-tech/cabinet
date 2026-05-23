describe('chats/ui-screen-chat-copilot', () => {
  function signInToChats() {
    cy.visit('/sign-in?redirect=%2Fchats%2F')
    cy.get('input[name="email"]').clear().type('e2e-chats@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
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
    cy.get('[data-testid="chat-new-thread-input"]').clear().type(title)
    cy.get('[data-testid="chat-create-thread-button"]').click()
    cy.get('[data-testid="chat-thread-item"]').contains(title).click()
    cy.get('[data-testid="chat-thread-title"]').should('contain', title)
  }

  it('CHAT-COPILOT-001 toggles assistant workspace from the header without route context loss', () => {
    openInventory()
    cy.get('[data-testid="active-profile-name"]').should('be.visible')
    cy.get('[data-testid="shell-workspace-navigation"]')
      .should('have.attr', 'data-active', 'true')
    cy.location('pathname').then((initialPathname) => {
      cy.get('[data-testid="shell-chat-toggle"]')
        .invoke('attr', 'aria-label')
        .should('match', /open.*assistant workspace/i)
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
        .should('match', /return to navigation workspace/i)
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
    cy.get('[data-testid="chat-new-thread-input"]').should('be.disabled')
    cy.get('[data-testid="chat-create-thread-button"]').should('be.disabled')
    cy.get('@createThread.all').should('have.length', 0)
  })

  it('UI-SCREEN-CHAT-COPILOT-007 keeps Preview Action gated until thread context exists', () => {
    openChats()
    createThread('E2E Empty Thread Preview Gate')

    cy.contains('No messages in this thread yet.').should('be.visible')
    cy.get('[data-testid="chat-preview-action-button"]').should('be.disabled')
    cy.get('[data-testid="chat-action-preview"]').should('not.exist')
  })

  it('UI-SCREEN-CHAT-COPILOT-008 supports confirm-before-apply for inventory and wishlist mutations', () => {
    openChats()
    createThread('E2E Copilot CRUD Thread')

    cy.get('[data-testid="chat-compose-input"]').clear().type('Please create this item')
    cy.get('[data-testid="chat-send-button"]').click()
    cy.get('[data-testid="chat-message-list"]').should('contain', 'Please create this item')

    cy.get('[data-testid="chat-preview-action-mode"]').select('create_inventory_item')
    cy.get('[data-testid="chat-preview-part-number"]').clear().type('CP-007-INV')
    cy.get('[data-testid="chat-preview-title"]').clear().type('Copilot Inventory Create')
    cy.get('[data-testid="chat-preview-action-button"]').click()
    cy.get('[data-testid="chat-action-preview"]').should('contain', 'create_inventory_item')
    cy.get('[data-testid="chat-apply-action-button"]').click()
    cy.get('[data-testid="chat-apply-confirm-dialog"]').should('be.visible')
    cy.get('[data-testid="chat-apply-confirm-summary"]').should('contain', 'create_inventory_item')
    cy.get('[data-testid="chat-apply-confirm-submit"]').click()
    cy.get('[data-testid="chat-action-apply-result"]').should('contain', 'create_inventory_item')

    cy.get('[data-testid="chat-preview-action-mode"]').select('create_wishlist_entry')
    cy.get('[data-testid="chat-preview-part-number"]').clear().type('CP-007-WISH')
    cy.get('[data-testid="chat-preview-title"]').clear().type('Copilot Wishlist Create')
    cy.get('[data-testid="chat-preview-action-button"]').click()
    cy.get('[data-testid="chat-action-preview"]').should('contain', 'create_wishlist_entry')
    cy.get('[data-testid="chat-apply-action-button"]').click()
    cy.get('[data-testid="chat-apply-confirm-dialog"]').should('be.visible')
    cy.get('[data-testid="chat-apply-confirm-submit"]').click()
    cy.get('[data-testid="chat-action-apply-result"]').should('contain', 'create_wishlist_entry')
  })

  it('UI-SCREEN-CHAT-COPILOT-012 reflects assistant provider defaults in chat action previews', () => {
    openChatsWithAssistantDefaults('anthropic', 'claude-3-7-sonnet')
    createThread('E2E Copilot Provider Defaults Thread')

    cy.get('[data-testid="chat-assistant-defaults"]').should(
      'contain',
      'anthropic / claude-3-7-sonnet'
    )
    cy.get('[data-testid="chat-compose-input"]').clear().type('Draft this with the active assistant defaults')
    cy.get('[data-testid="chat-send-button"]').click()
    cy.get('[data-testid="chat-message-list"]').should(
      'contain',
      'Draft this with the active assistant defaults'
    )

    cy.get('[data-testid="chat-preview-action-mode"]').select('create_inventory_item')
    cy.get('[data-testid="chat-preview-part-number"]').clear().type('CP-012-PROVIDER')
    cy.get('[data-testid="chat-preview-title"]').clear().type('Provider Default Preview')
    cy.get('[data-testid="chat-preview-action-button"]').click()
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

    cy.get('[data-testid="chat-preview-action-mode"]').select(
      'assign_collection_item'
    )
    cy.get('[data-testid="chat-preview-target-item-id"]')
      .clear()
      .type('inventory-item-pikachu-shadowless')
    cy.get('[data-testid="chat-preview-part-number"]')
      .clear()
      .type('CP-013-COLLECT')
    cy.get('[data-testid="chat-preview-title"]')
      .clear()
      .type('Shadowless Pikachu')
    cy.get('[data-testid="chat-preview-collection-name"]')
      .clear()
      .type('Store 1')
    cy.get('[data-testid="chat-preview-action-button"]').click()

    cy.get('[data-testid="chat-action-preview"]')
      .should('contain', 'assign_collection_item')
      .and('contain', 'inventory-item-pikachu-shadowless')
      .and('contain', 'Store 1')
      .and('contain', 'openai / gpt-4.1-mini')
    cy.get('[data-testid="chat-apply-action-button"]').click()
    cy.get('[data-testid="chat-apply-confirm-summary"]')
      .should('contain', 'Apply assign_collection_item')
      .and('contain', 'target=inventory-item-pikachu-shadowless')
      .and('contain', 'collection=Store 1')
      .and('contain', 'assistant=openai/gpt-4.1-mini')
  })

  it('UI-SCREEN-CHAT-COPILOT-011 cancels preview apply without mutating the pending action', () => {
    openChats()
    createThread('E2E Copilot Cancel Apply Thread')

    cy.get('[data-testid="chat-compose-input"]').clear().type('Draft this item only')
    cy.get('[data-testid="chat-send-button"]').click()
    cy.get('[data-testid="chat-message-list"]').should('contain', 'Draft this item only')

    cy.get('[data-testid="chat-preview-action-mode"]').select('create_inventory_item')
    cy.get('[data-testid="chat-preview-part-number"]').clear().type('CP-011-CANCEL')
    cy.get('[data-testid="chat-preview-title"]').clear().type('Copilot Cancel Preview')
    cy.get('[data-testid="chat-preview-action-button"]').click()
    cy.get('[data-testid="chat-action-preview"]').should('contain', 'create_inventory_item')
    cy.get('[data-testid="chat-apply-action-button"]').click()
    cy.get('[data-testid="chat-apply-confirm-dialog"]').should('be.visible')
    cy.get('[data-testid="chat-apply-confirm-summary"]').should('contain', 'CP-011-CANCEL')
    cy.get('[data-testid="chat-apply-confirm-cancel"]').click()

    cy.get('[data-testid="chat-apply-confirm-dialog"]').should('not.exist')
    cy.get('[data-testid="chat-action-preview"]').should('contain', 'create_inventory_item')
    cy.get('[data-testid="chat-action-apply-notice"]').should(
      'contain',
      'Action apply canceled; preview remains pending.'
    )
    cy.get('[data-testid="chat-action-apply-result"]').should('not.exist')
  })

  it('UI-SCREEN-CHAT-COPILOT-009 supports mobile image attachment and confirm-before-apply flow once message context exists', () => {
    cy.viewport(390, 844)
    openChats()
    createThread('E2E Mobile Copilot Thread')

    cy.get('[data-testid="chat-attachment-input"]').selectFile(
      {
        contents: Cypress.Buffer.from('fake-image-data'),
        fileName: 'mobile-chat-photo.jpg',
        mimeType: 'image/jpeg',
      },
      { force: true }
    )
    cy.get('[data-testid="chat-upload-attachment-button"]').click()
    cy.get('[data-testid="chat-attachment-list"]').should('contain', 'mobile-chat-photo.jpg')
    cy.get('[data-testid="chat-compose-input"]').clear().type('Use the attached photo to create an item')
    cy.get('[data-testid="chat-send-button"]').click()
    cy.get('[data-testid="chat-message-list"]').should(
      'contain',
      'Use the attached photo to create an item'
    )

    cy.get('[data-testid="chat-preview-action-mode"]').select('create_inventory_item')
    cy.get('[data-testid="chat-preview-part-number"]').clear().type('CP-008-MOBILE')
    cy.get('[data-testid="chat-preview-title"]').clear().type('Mobile Image Suggestion')
    cy.get('[data-testid="chat-preview-action-button"]').click()
    cy.get('[data-testid="chat-action-preview"]').should('be.visible')
    cy.get('[data-testid="chat-apply-action-button"]').click()
    cy.get('[data-testid="chat-apply-confirm-dialog"]').should('be.visible')
    cy.get('[data-testid="chat-apply-confirm-summary"]').should('contain', 'CP-008-MOBILE')
    cy.get('[data-testid="chat-apply-confirm-submit"]').click()
    cy.get('[data-testid="chat-action-apply-result"]').should('contain', 'CP-008-MOBILE')
  })

  it('UI-SCREEN-CHAT-COPILOT-010 keeps top-level /inbox reachable as a communications surface', () => {
    openInbox()
    cy.get('[data-testid="purchase-inbox-load-reviews"]').should('be.visible')
    cy.contains('404').should('not.exist')
    cy.contains('Oops! Page Not Found!').should('not.exist')
  })
})
