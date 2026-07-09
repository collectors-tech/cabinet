describe('collections-members-pagination', () => {
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

  it('UI-SCREEN-COLLECTIONS-029 retires members-table pagination from Collections', () => {
    signInToCollections()

    cy.get('[data-testid="collections-table-pagination"]').should('be.visible')
    cy.get('[data-testid="collections-members-table-pagination"]').should(
      'not.exist'
    )
    cy.get('[data-testid="collections-members-table"]').should('not.exist')
  })
})
