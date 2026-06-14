describe('collections-protected-all-items', () => {
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
    cy.e2eBootstrap().then(({ profile_id, profile_name }) => {
      cy.useBootstrappedProfile(profile_id, profile_name, {
        path: '/collections/',
      })
    })
    cy.wait('@loadCollectionSettings')
    cy.wait('@collectionsInventoryItems')
  }

  it('keeps All Items protected from row rename and delete actions', () => {
    signInToCollections()
    cy.intercept('PUT', '/api/profiles/e2e-profile-001/settings').as(
      'saveCollectionSettings'
    )

    cy.get('[data-testid="collections-row-all-items"]').should(
      'have.attr',
      'data-state',
      'selected'
    )
    cy.get('[data-testid="collections-active-context"]').should(
      'contain.text',
      'All Items'
    )

    cy.get('[data-testid="collections-row-edit-all-items"]').click()
    cy.get('[data-testid="collections-edit-panel"]').should('be.visible')
    cy.get('[data-testid="collections-edit-input"]')
      .should('have.value', 'All Items')
      .clear()
      .type('Everything Else')
    cy.get('[data-testid="collections-edit-submit"]').click({ force: true })
    cy.get('@saveCollectionSettings.all').should('have.length', 0)
    cy.get('[data-testid="collections-edit-panel"]').should('be.visible')
    cy.get('[data-testid="collections-row-all-items"]').should('be.visible')
    cy.get('[data-testid="collections-row-everything-else"]').should(
      'not.exist'
    )
    cy.get('[data-testid="collections-edit-panel"]').within(() => {
      cy.contains('button', 'Cancel').click()
    })

    cy.get('[data-testid="collections-row-delete-all-items"]').click()
    cy.get('[data-testid="collections-delete-dialog"]').should('be.visible')
    cy.get('[data-testid="collections-delete-submit"]').click()
    cy.get('@saveCollectionSettings.all').should('have.length', 0)
    cy.get('[data-testid="collections-delete-dialog"]').should('be.visible')
    cy.get('[data-testid="collections-row-all-items"]').should('be.visible')
    cy.get('[data-testid="collections-active-context"]').should(
      'contain.text',
      'All Items'
    )
  })
})
