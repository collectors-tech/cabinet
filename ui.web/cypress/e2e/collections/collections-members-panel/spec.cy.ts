describe('collections-members-panel', () => {
  const collectionsSettingsKey = 'collections.workspace.v1'

  const inventoryItems = [
    {
      id: 'inventory-item-kobe-rookie',
      part_number: 'KOBE-1996',
      title: '1996 rookie card',
      status: 'active',
      category: 'Trading Card',
      brand: 'Cabinet',
      description: 'Assigned member fixture',
    },
    {
      id: 'inventory-item-pikachu-shadowless',
      part_number: 'PKMN-SHADOWLESS',
      title: 'Shadowless Pikachu',
      status: 'active',
      category: 'Trading Card',
      brand: 'Cabinet',
      description: 'Assigned member fixture',
    },
    {
      id: 'inventory-item-charizard-base',
      part_number: 'PKMN-CHARIZARD',
      title: 'Base Set card',
      status: 'active',
      category: 'Trading Card',
      brand: 'Cabinet',
      description: 'Unassigned member fixture',
    },
  ]

  function mockInventoryItems(
    items: typeof inventoryItems = inventoryItems
  ) {
    cy.intercept('GET', '/api/items', {
      statusCode: 200,
      body: { items },
    }).as('collectionsInventoryItems')
  }

  function signInToCollections() {
    cy.viewport(1512, 967)
    cy.e2eReset()
    cy.e2eSetSetupState('present')
    mockInventoryItems()
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

  function persistedCollectionsSettings(requestBody: unknown) {
    const body = requestBody as { settings?: Record<string, string> }
    return JSON.parse(body.settings?.[collectionsSettingsKey] ?? '{}') as {
      activeCollection?: string
    }
  }

  it('UI-SCREEN-COLLECTIONS-027 reflects selected collection members and empty states', () => {
    signInToCollections()

    cy.get('[data-testid="collections-active-context"]').should(
      'contain.text',
      'All Items'
    )
    cy.get('[data-testid="collections-row-count-all-items"]').should(
      'have.text',
      '3'
    )
    cy.get('[data-testid="collections-members-summary"]').should(
      'contain.text',
      'Showing 3 of 3 items.'
    )
    cy.get('[data-testid="collections-member-row-inventory-item-kobe-rookie"]').should(
      'contain.text',
      '1996 rookie card'
    )
    cy.get(
      '[data-testid="collections-member-current-1996-rookie-card"]'
    ).should('contain.text', 'Currently in Watch List.')
    cy.get('[data-testid="collections-member-row-inventory-item-charizard-base"]').should(
      'contain.text',
      'Base Set card'
    )
    cy.get('[data-testid="collections-member-current-base-set-card"]').should(
      'contain.text',
      'Currently in Unassigned.'
    )
    cy.get('@saveCollectionSettings.all').should('have.length', 0)

    cy.get('[data-testid="collections-row-store-1"]').click()
    cy.wait('@saveCollectionSettings').then(({ request }) => {
      expect(
        persistedCollectionsSettings(request.body).activeCollection
      ).to.equal('Store 1')
    })
    cy.get('[data-testid="collections-active-context"]').should(
      'contain.text',
      'Store 1'
    )
    cy.get('[data-testid="collections-members-summary"]').should(
      'contain.text',
      'Showing 1 of 1 items.'
    )
    cy.get(
      '[data-testid="collections-member-row-inventory-item-pikachu-shadowless"]'
    )
      .should('be.visible')
      .and('contain.text', 'Shadowless Pikachu')
    cy.get(
      '[data-testid="collections-member-current-shadowless-pikachu"]'
    ).should('contain.text', 'Currently in Store 1.')
    cy.get(
      '[data-testid="collections-member-row-inventory-item-kobe-rookie"]'
    ).should('not.exist')

    cy.get('[data-testid="collections-row-overflow"]').click()
    cy.wait('@saveCollectionSettings').then(({ request }) => {
      expect(
        persistedCollectionsSettings(request.body).activeCollection
      ).to.equal('Overflow')
    })
    cy.get('[data-testid="collections-active-context"]').should(
      'contain.text',
      'Overflow'
    )
    cy.get('[data-testid="collections-members-summary"]').should(
      'contain.text',
      'Showing 0 of 0 items.'
    )
    cy.get('[data-testid="collections-members-empty-row"]')
      .should('be.visible')
      .and('contain.text', 'No items are currently assigned to Overflow.')
    cy.get(
      '[data-testid="collections-member-row-inventory-item-pikachu-shadowless"]'
    ).should('not.exist')
  })
})
