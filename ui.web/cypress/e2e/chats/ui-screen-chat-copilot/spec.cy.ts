describe('chats/ui-screen-chat-copilot', () => {
  function openInventory() {
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', { path: '/inventory/' })
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)
  }

  function openChats() {
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', { path: '/chats/' })
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/chats\/?$/)
  }

  function createThread(title: string) {
    cy.get('[data-testid="chat-new-thread-input"]').clear().type(title)
    cy.get('[data-testid="chat-create-thread-button"]').click()
    cy.get('[data-testid="chat-thread-item"]').contains(title).click()
    cy.get('[data-testid="chat-thread-title"]').should('contain', title)
  }

  it('CHAT-COPILOT-001 toggles chat rail from header without route context loss', () => {
    openInventory()
    cy.location('pathname').then((initialPathname) => {
      cy.get('[data-testid="shell-chat-toggle"]')
        .invoke('attr', 'aria-label')
        .should('match', /open.*chat/i)
      cy.get('[data-testid="shell-chat-toggle"]').click()
      cy.get('[data-testid="shell-chat-rail"]').should('be.visible')
      cy.location('pathname').should('eq', initialPathname)

      cy.get('[data-testid="shell-chat-toggle"]')
        .invoke('attr', 'aria-label')
        .should('match', /close.*chat/i)
      cy.get('[data-testid="shell-chat-toggle"]').click()
      cy.get('[data-testid="shell-chat-rail"]').should('not.exist')
      cy.location('pathname').should('eq', initialPathname)
    })
  })

  it('UI-SCREEN-CHAT-COPILOT-007 supports confirm-before-apply for inventory and wishlist mutations', () => {
    openChats()
    createThread('E2E Copilot CRUD Thread')

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

  it('UI-SCREEN-CHAT-COPILOT-008 supports mobile image attachment and confirm-before-apply flow', () => {
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
})
