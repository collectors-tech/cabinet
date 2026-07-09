describe('collections-filter-empty-states', () => {
  const inventoryItems = [
    {
      id: 'inventory-item-kobe-rookie',
      part_number: 'KOBE-1996',
      title: '1996 Topps Kobe Bryant rookie',
      status: 'active',
      category: 'Trading Card',
      brand: 'Topps',
      description: 'PSA candidate, Lakers lot',
    },
    {
      id: 'inventory-item-charizard-base',
      part_number: 'PKMN-CHARIZARD',
      title: 'Base Set Charizard',
      status: 'active',
      category: 'Trading Card',
      brand: 'Pokemon',
      description: 'High-value vault candidate',
    },
  ]

  function signInToCollections() {
    cy.viewport(1512, 967)
    cy.e2eReset()
    cy.e2eSetSetupState('present')
    cy.intercept('GET', '/api/items', {
      statusCode: 200,
      body: { items: inventoryItems },
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

  it('UI-SCREEN-COLLECTIONS-024 shows deterministic zero-result filter states without saving settings', () => {
    signInToCollections()

    cy.get('[data-testid="collections-active-context"]').should(
      'contain.text',
      'All Items'
    )
    cy.get('[data-testid="collections-management-summary"]').should(
      'contain.text',
      'Showing 6 of 6 collections.'
    )
    cy.get('[data-testid="collections-search-input"]').type(
      'no matching collection row'
    )
    cy.get('[data-testid="collections-management-summary"]').should(
      'contain.text',
      'Showing 0 of 6 collections.'
    )
    cy.get('[data-testid="collections-row-all-items"]').should('not.exist')
    cy.contains('No collections match the current filter.').should(
      'be.visible'
    )
    cy.get('[data-testid="collections-active-context"]').should(
      'contain.text',
      'All Items'
    )
    cy.get('@saveCollectionSettings.all').should('have.length', 0)

    cy.get('[data-testid="collections-search-input"]').clear()
    cy.get('[data-testid="collections-row-all-items"]').should('be.visible')
    cy.get('[data-testid="collections-members-search-input"]').should('not.exist')
    cy.get('[data-testid="collections-members-table"]').should('not.exist')
    cy.contains('No collection members match the current filter.').should(
      'not.exist'
    )
    cy.get('[data-testid="collections-active-context"]').should(
      'contain.text',
      'All Items'
    )
    cy.get('@saveCollectionSettings.all').should('have.length', 0)
  })
})
