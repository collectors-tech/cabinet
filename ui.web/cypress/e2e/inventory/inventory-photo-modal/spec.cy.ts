describe('inventory photo modal', () => {
  function signIn() {
    cy.e2eReset()
    cy.e2eSetSetupState('present')
    cy.e2eBootstrap().then(({ profile_id, profile_name }) => {
      cy.useBootstrappedProfile(profile_id, profile_name, { path: '/inventory/' })
    })
  }

  it('opens item-scoped photos from row actions without inline photos section', () => {
    cy.intercept('GET', '/api/items', {
      statusCode: 200,
      body: {
        items: [
          {
            id: 'item-photo-alpha',
            part_number: 'PN-PHOTO-A',
            title: 'Photo Alpha',
            status: 'active',
            category: 'Cars',
            brand: 'AFX',
          },
          {
            id: 'item-photo-bravo',
            part_number: 'PN-PHOTO-B',
            title: 'Photo Bravo',
            status: 'active',
            category: 'Trains',
            brand: 'Tyco',
          },
        ],
      },
    }).as('items')

    cy.intercept('GET', '/api/items/item-photo-alpha/photos', {
      statusCode: 200,
      body: {
        photos: [{ id: 'alpha-photo-1', filename: 'alpha-one.jpg', is_primary: true }],
      },
    }).as('alphaPhotos')

    cy.intercept('GET', '/api/items/item-photo-bravo/photos', {
      statusCode: 200,
      body: {
        photos: [{ id: 'bravo-photo-1', filename: 'bravo-one.jpg', is_primary: true }],
      },
    }).as('bravoPhotos')

    signIn()
    cy.wait('@items')
    cy.wait('@alphaPhotos')

    cy.get('[data-testid="inventory-photos-section"]').should('not.exist')

    cy.get(
      '[data-testid="inventory-item-row-item-photo-alpha"] [data-testid="inventory-row-photos-action"]'
    ).click()
    cy.get('[data-testid="inventory-photos-dialog"]')
      .should('be.visible')
      .and('contain', 'Photo Alpha')
    cy.get('[data-testid="inventory-photos-dialog"] [data-slot="dialog-close"]').should(
      'have.length',
      1
    )
    cy.get('[data-testid="inventory-photos-dialog-close"]').should(
      'have.attr',
      'data-slot',
      'dialog-close'
    )
    cy.contains('[data-testid="inventory-photo-row"]', 'alpha-one.jpg').should('be.visible')
    cy.get('[data-testid="inventory-photos-dialog-close"]').click()
    cy.get('[data-testid="inventory-photos-dialog"]').should('not.exist')

    cy.get(
      '[data-testid="inventory-item-row-item-photo-bravo"] [data-testid="inventory-row-photos-action"]'
    ).click()
    cy.wait('@bravoPhotos')
    cy.get('[data-testid="inventory-photos-dialog"]')
      .should('be.visible')
      .and('contain', 'Photo Bravo')
    cy.contains('[data-testid="inventory-photo-row"]', 'bravo-one.jpg').should('be.visible')
    cy.get('[data-testid="collection-selected-item"]').should('contain', 'PN-PHOTO-B')
  })
})
