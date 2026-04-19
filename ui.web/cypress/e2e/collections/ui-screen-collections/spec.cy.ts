describe('ui-screen-collections', () => {
  function signInToCollections() {
    cy.visit('/sign-in?redirect=%2Fcollections%2F')
    cy.get('input[name="email"]').clear().type('e2e-collections@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/collections\/?$/)
  }

  it('UI-SCREEN-COLLECTIONS-001 renders shared collections management table', () => {
    signInToCollections()

    cy.get('[data-testid="collections-section"]').should('be.visible')
    cy.get('[data-testid="collections-shared-table"]').should('be.visible')
    cy.get('[data-testid="collections-new-action"]').should('be.visible')
    cy.get('[data-testid="collections-row-all-items"]').should('be.visible')
  })

  it('UI-SCREEN-COLLECTIONS-002 selects a row and persists active context across refresh', () => {
    signInToCollections()

    cy.get('[data-testid="collections-row-watch-list"]').click()
    cy.get('[data-testid="collections-selected-name"]').should('contain.text', 'Watch List')
    cy.get('[data-testid="collections-active-context-message"]').should(
      'contain.text',
      'Active collection is Watch List'
    )

    cy.reload()
    cy.get('[data-testid="collections-selected-name"]').should('contain.text', 'Watch List')
    cy.get('[data-testid="collections-active-context-persistence"]').should(
      'contain.text',
      'Persists for this signed-in profile'
    )
  })

  it('UI-SCREEN-COLLECTIONS-003 creates a collection from the table workflow and persists after refresh', () => {
    signInToCollections()

    cy.get('[data-testid="collections-new-action"]').click()
    cy.get('[data-testid="collections-create-input"]').type('Collections Alpha')
    cy.get('[data-testid="collections-create-submit"]').click()

    cy.contains('Collections Alpha created and set as the active collection.').should('be.visible')
    cy.get('[data-testid="collections-row-collections-alpha"]').should('be.visible')
    cy.reload()
    cy.get('[data-testid="collections-row-collections-alpha"]').should('be.visible')
    cy.get('[data-testid="collections-selected-name"]').should('contain.text', 'Collections Alpha')
  })

  it('UI-SCREEN-COLLECTIONS-004 renames a collection from the row workflow and persists after refresh', () => {
    signInToCollections()

    cy.get('[data-testid="collections-row-store-2"]').click()
    cy.get('[data-testid="collections-row-edit-store-2"]').click()
    cy.get('[data-testid="collections-edit-input"]').clear().type('Store 2 Prime')
    cy.get('[data-testid="collections-edit-submit"]').click()

    cy.contains('Store 2 renamed to Store 2 Prime.').should('be.visible')
    cy.get('[data-testid="collections-row-store-2-prime"]').should('be.visible')
    cy.reload()
    cy.get('[data-testid="collections-row-store-2-prime"]').should('be.visible')
  })

  it('UI-SCREEN-COLLECTIONS-005 deletes a collection and releases assigned items', () => {
    signInToCollections()

    cy.get('[data-testid="collections-row-store-1"]').click()
    cy.get('[data-testid="collections-member-shadowless-pikachu"]').should('be.visible')
    cy.get('[data-testid="collections-row-delete-store-1"]').click()
    cy.get('[data-testid="collections-delete-submit"]').click()

    cy.contains('Store 1 removed from workspace collections.').should('be.visible')
    cy.get('[data-testid="collections-row-store-1"]').should('not.exist')

    cy.get('[data-testid="collections-row-watch-list"]').click()
    cy.get('[data-testid="collections-assignment-select"]').click()
    cy.contains('Shadowless Pikachu (Unassigned)').should('be.visible')
  })

  it('UI-SCREEN-COLLECTIONS-006 filters collections within the shared table surface', () => {
    signInToCollections()

    cy.get('[data-testid="collections-management-summary"]').should('contain.text', 'Showing 6 of 6 collections.')
    cy.get('[data-testid="collections-search-input"]').type('watch')
    cy.get('[data-testid="collections-row-watch-list"]').should('be.visible')
    cy.get('[data-testid="collections-row-all-items"]').should('not.exist')
    cy.get('[data-testid="collections-management-summary"]').should('contain.text', 'Showing 1 of 6 collections.')
  })

  it('UI-SCREEN-COLLECTIONS-007 assigns an item into the selected collection and persists after refresh', () => {
    signInToCollections()

    cy.get('[data-testid="collections-row-warehouse-1"]').click()
    cy.get('[data-testid="collections-assignment-select"]').click()
    cy.contains('Base Set Charizard (Unassigned)').click()
    cy.get('[data-testid="collections-assignment-submit"]').click()

    cy.contains('Base Set Charizard assigned to Warehouse 1.').should('be.visible')
    cy.get('[data-testid="collections-member-base-set-charizard"]').should('be.visible')

    cy.reload()
    cy.get('[data-testid="collections-selected-name"]').should('contain.text', 'Warehouse 1')
    cy.get('[data-testid="collections-member-base-set-charizard"]').should('be.visible')
  })

  it('UI-SCREEN-COLLECTIONS-008 moves an assigned item between collections and persists after refresh', () => {
    signInToCollections()

    cy.get('[data-testid="collections-row-watch-list"]').click()
    cy.get('[data-testid="collections-member-1996-topps-kobe-bryant-rookie"]').should('be.visible')
    cy.get('[data-testid="collections-move-target-1996-topps-kobe-bryant-rookie"]').click()
    cy.contains('[role="option"]', 'Warehouse 1').click()
    cy.get('[data-testid="collections-move-submit-1996-topps-kobe-bryant-rookie"]').click()

    cy.contains('1996 Topps Kobe Bryant rookie moved to Warehouse 1.').should('be.visible')
    cy.get('[data-testid="collections-member-1996-topps-kobe-bryant-rookie"]').should('not.exist')

    cy.get('[data-testid="collections-row-warehouse-1"]').click()
    cy.get('[data-testid="collections-member-1996-topps-kobe-bryant-rookie"]').should('be.visible')
    cy.reload()
    cy.get('[data-testid="collections-member-1996-topps-kobe-bryant-rookie"]').should('be.visible')
  })

  it('UI-SCREEN-COLLECTIONS-009 retains tag iconography for collections route identity', () => {
    signInToCollections()

    cy.get('[data-testid="sidebar-nav-link-collections"]').should('be.visible')
    cy.get('[data-testid="collections-page-icon"]').should('be.visible')
    cy.get('[data-testid="sidebar-nav-link-collections"] svg').should('have.attr', 'data-lucide', 'tag')
    cy.get('[data-testid="collections-page-icon"]').should('have.attr', 'data-lucide', 'tag')
  })
})
