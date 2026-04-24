describe('inventory barcode modal', () => {
  function signIn() {
    cy.e2eReset()
    cy.e2eSetSetupState('present')
    cy.e2eBootstrap().then(({ profile_id, profile_name }) => {
      cy.useBootstrappedProfile(profile_id, profile_name, { path: '/inventory/' })
    })
  }

  it('opens item-scoped barcodes from page and row actions without inline barcodes section', () => {
    cy.intercept('GET', '/api/items', {
      statusCode: 200,
      body: {
        items: [
          {
            id: 'item-barcode-alpha',
            part_number: 'PN-BARCODE-A',
            title: 'Barcode Alpha',
            status: 'active',
            category: 'Cars',
            brand: 'AFX',
          },
          {
            id: 'item-barcode-bravo',
            part_number: 'PN-BARCODE-B',
            title: 'Barcode Bravo',
            status: 'active',
            category: 'Trains',
            brand: 'Tyco',
          },
        ],
      },
    }).as('items')

    cy.intercept('GET', '/api/barcodes/1234567890123', {
      statusCode: 200,
      body: {
        matches: [
          {
            id: 'barcode-alpha-1',
            item_id: 'item-barcode-alpha',
            barcode: '1234567890123',
            created_at: '2026-04-25T10:00:00Z',
          },
        ],
      },
    }).as('lookupAlpha')

    signIn()
    cy.wait('@items')

    cy.get('[data-testid="inventory-barcodes-section"]').should('not.exist')

    cy.get('[data-testid="inventory-barcodes-action"]').click()
    cy.get('[data-testid="inventory-barcodes-dialog"]')
      .should('be.visible')
      .and('contain', 'Barcode Alpha')
    cy.get('[data-testid="inventory-barcodes-lookup-input"]').type('1234567890123')
    cy.get('[data-testid="inventory-barcodes-lookup-button"]').click()
    cy.wait('@lookupAlpha')
    cy.get('[data-testid="inventory-barcodes-match-row"]')
      .should('have.length', 1)
      .first()
      .should('contain', 'item-barcode-alpha')
    cy.get('[data-testid="inventory-barcodes-dialog-close"]').click()
    cy.get('[data-testid="inventory-barcodes-dialog"]').should('not.exist')

    cy.get(
      '[data-testid="inventory-item-row-item-barcode-bravo"] [data-testid="inventory-row-barcodes-action"]'
    ).click()
    cy.get('[data-testid="inventory-barcodes-dialog"]')
      .should('be.visible')
      .and('contain', 'Barcode Bravo')
    cy.get('[data-testid="collection-selected-item"]').should('contain', 'PN-BARCODE-B')
  })
})
