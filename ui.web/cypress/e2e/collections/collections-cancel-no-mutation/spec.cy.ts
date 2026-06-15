describe('collections-cancel-no-mutation', () => {
  function signInToCollections() {
    cy.viewport(1512, 967)
    cy.e2eReset()
    cy.e2eSetSetupState('present')
    cy.intercept('GET', '/api/items', {
      statusCode: 200,
      body: { items: [] },
    }).as('collectionsInventoryItems')
    cy.intercept('GET', '/api/profiles/*/settings').as(
      'loadCollectionSettings'
    )
    cy.intercept('PUT', '/api/profiles/e2e-profile-001/settings').as(
      'saveCollectionSettings'
    )
    cy.e2eBootstrap().then(({ profile_id, profile_name }) => {
      cy.useBootstrappedProfile(profile_id, profile_name, {
        path: '/collections/',
      })
    })
    cy.wait('@loadCollectionSettings')
    cy.wait('@collectionsInventoryItems')
  }

  it('UI-SCREEN-COLLECTIONS-025 cancels create edit and delete workflows without saving settings', () => {
    signInToCollections()

    cy.get('[data-testid="collections-active-context"]').should(
      'contain.text',
      'All Items'
    )
    cy.get('@saveCollectionSettings.all').should('have.length', 0)

    cy.get('[data-testid="collections-new-action"]').click()
    cy.get('[data-testid="collections-create-dialog"]').should('be.visible')
    cy.get('[data-testid="collections-create-input"]').type(
      'Cancelled Create Shelf'
    )
    cy.get('[data-testid="collections-create-dialog"]').within(() => {
      cy.contains('button', 'Cancel').click()
    })
    cy.get('[data-testid="collections-create-dialog"]').should('not.exist')
    cy.get('[data-testid="collections-row-cancelled-create-shelf"]').should(
      'not.exist'
    )
    cy.get('[data-testid="collections-active-context"]').should(
      'contain.text',
      'All Items'
    )
    cy.get('@saveCollectionSettings.all').should('have.length', 0)

    cy.get('[data-testid="collections-row-store-1"]').click()
    cy.wait('@saveCollectionSettings')
    cy.get('[data-testid="collections-active-context"]').should(
      'contain.text',
      'Store 1'
    )
    cy.get('@saveCollectionSettings.all').should('have.length', 1)
    cy.get('[data-testid="collections-row-edit-store-1"]')
      .scrollIntoView()
      .click({ force: true })
    cy.get('[data-testid="collections-edit-panel"]')
      .should('be.visible')
      .should('have.attr', 'data-side', 'right')
    cy.get('[data-testid="collections-edit-input"]')
      .should('have.value', 'Store 1')
      .clear()
      .type('Cancelled Rename Shelf')
    cy.get('[data-testid="collections-edit-panel"]').within(() => {
      cy.contains('button', 'Cancel').click()
    })
    cy.get('[data-testid="collections-edit-panel"]').should('not.exist')
    cy.get('[data-testid="collections-row-name-store-1"]').should(
      'contain.text',
      'Store 1'
    )
    cy.get('[data-testid="collections-row-cancelled-rename-shelf"]').should(
      'not.exist'
    )
    cy.get('[data-testid="collections-active-context"]').should(
      'contain.text',
      'Store 1'
    )
    cy.get('@saveCollectionSettings.all').should('have.length', 1)

    cy.get('[data-testid="collections-row-delete-store-1"]')
      .scrollIntoView()
      .click({ force: true })
    cy.get('[data-testid="collections-delete-dialog"]').should('be.visible')
    cy.get('[data-testid="collections-delete-dialog"]').within(() => {
      cy.contains('button', 'Cancel').click()
    })
    cy.get('[data-testid="collections-delete-dialog"]').should('not.exist')
    cy.get('[data-testid="collections-row-store-1"]').should('be.visible')
    cy.get('[data-testid="collections-active-context"]').should(
      'contain.text',
      'Store 1'
    )
    cy.get('@saveCollectionSettings.all').should('have.length', 1)
  })
})
