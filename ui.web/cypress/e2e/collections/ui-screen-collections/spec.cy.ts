describe('ui-screen-collections', () => {
  function signInToCollections() {
    cy.visit('/sign-in?redirect=%2Fcollections%2F')
    cy.get('input[name="email"]').clear().type('e2e-collections@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/collections\/?$/)
  }

  beforeEach(() => {
    cy.visit('/sign-in')
    cy.window().then((win) => {
      win.localStorage.clear()
    })
  })

  it('renders collections as shared management table', () => {
    signInToCollections()

    cy.get('[data-testid="collections-shared-table"]').should('be.visible')
    cy.contains('Collection management table').should('be.visible')
    cy.contains('th', 'Collection').should('be.visible')
    cy.contains('th', 'Items').should('be.visible')
    cy.contains('th', 'Scope').should('be.visible')
    cy.contains('th', 'Status').should('be.visible')
    cy.get('[data-testid="collections-row-all-items"]').should('be.visible')
    cy.get('[data-testid="collections-new-action"]').should('be.visible')
  })

  it('selects a collection row and updates management context', () => {
    signInToCollections()

    cy.get('[data-testid="collections-row-store-1"]').click()
    cy.get('[data-testid="collections-selected-name"]').should('contain', 'Store 1')
    cy.get('[data-testid="collections-selected-count"]').should('not.be.empty')
    cy.get('[data-testid="collections-row-store-1"]').should('have.attr', 'data-state', 'selected')
  })

  it('creates a collection from the table workflow and persists after refresh', () => {
    signInToCollections()

    cy.get('[data-testid="collections-new-action"]').click()
    cy.get('[data-testid="collections-create-input"]').type('Cabinet Alpha')
    cy.get('[data-testid="collections-create-submit"]').click()

    cy.get('[data-testid="collections-row-cabinet-alpha"]').should('be.visible')
    cy.get('[data-testid="collections-selected-name"]').should('contain', 'Cabinet Alpha')

    cy.reload()
    cy.get('[data-testid="collections-row-cabinet-alpha"]').should('be.visible')
    cy.get('[data-testid="collections-selected-name"]').should('contain', 'Cabinet Alpha')
  })

  it('renames a collection from the row workflow and persists after refresh', () => {
    signInToCollections()

    cy.get('[data-testid="collections-row-store-2"]').click()
    cy.get('[data-testid="collections-row-edit-store-2"]').click()
    cy.get('[data-testid="collections-edit-input"]').clear().type('Store 2 Updated')
    cy.get('[data-testid="collections-edit-submit"]').click()

    cy.get('[data-testid="collections-row-store-2-updated"]').should('be.visible')
    cy.get('[data-testid="collections-selected-name"]').should('contain', 'Store 2 Updated')

    cy.reload()
    cy.get('[data-testid="collections-row-store-2-updated"]').should('be.visible')
    cy.get('[data-testid="collections-selected-name"]').should('contain', 'Store 2 Updated')
  })

  it('deletes a collection from the row workflow', () => {
    signInToCollections()

    cy.get('[data-testid="collections-row-warehouse-1"]').click()
    cy.get('[data-testid="collections-selected-delete"]').click()
    cy.get('[data-testid="collections-delete-submit"]').click()

    cy.get('[data-testid="collections-row-warehouse-1"]').should('not.exist')
  })

  it('filters collections within the shared table surface', () => {
    signInToCollections()

    cy.get('input[placeholder="Filter collections..."]').type('watch')
    cy.get('[data-testid="collections-row-watch-list"]').should('be.visible')
    cy.get('[data-testid="collections-row-all-items"]').should('not.exist')
    cy.get('[data-testid="collections-filtered-count"]').should('contain', '1')
  })
})
