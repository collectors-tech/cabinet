describe('collections-sorting-no-mutation', () => {
  const collectionsSettingsKey = 'collections.workspace.v1'

  const inventoryItems = [
    {
      id: 'sort-item-alpha-card',
      part_number: 'SORT-001',
      title: 'Alpha Card',
      status: 'active',
      category: 'Trading Card',
      brand: 'Cabinet',
      description: 'First sorted member',
    },
    {
      id: 'sort-item-middle-card',
      part_number: 'SORT-002',
      title: 'Middle Card',
      status: 'active',
      category: 'Trading Card',
      brand: 'Cabinet',
      description: 'Middle sorted member',
    },
    {
      id: 'sort-item-zulu-card',
      part_number: 'SORT-003',
      title: 'Zulu Card',
      status: 'active',
      category: 'Trading Card',
      brand: 'Cabinet',
      description: 'Last sorted member',
    },
  ]

  function rowTestIDs(selector: string) {
    return cy.get(selector).then(($rows) =>
      [...$rows].map((row) => row.getAttribute('data-testid'))
    )
  }

  function sortColumnDescending(tableSelector: string, columnName: string) {
    cy.get(tableSelector).contains('button', columnName).click()
    cy.get('[role="menuitem"]').contains('Desc').click()
  }

  function seedSortableCollectionsProfile() {
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
    cy.e2eBootstrap().then(({ profile_id, profile_name }) => {
      cy.request('PUT', `/api/profiles/${profile_id}/settings`, {
        settings: {
          [collectionsSettingsKey]: JSON.stringify({
            collections: [
              'All Items',
              'Alpha Bin',
              'Middle Shelf',
              'Zulu Vault',
            ],
            activeCollection: 'All Items',
            items: [
              {
                id: 'sort-item-alpha-card',
                name: 'Alpha Card',
                detail: 'First sorted member',
                collectionName: 'Alpha Bin',
              },
              {
                id: 'sort-item-middle-card',
                name: 'Middle Card',
                detail: 'Middle sorted member',
                collectionName: 'Middle Shelf',
              },
              {
                id: 'sort-item-zulu-card',
                name: 'Zulu Card',
                detail: 'Last sorted member',
                collectionName: 'Zulu Vault',
              },
            ],
          }),
        },
      })
        .its('status')
        .should('eq', 200)
      cy.useBootstrappedProfile(profile_id, profile_name, {
        path: '/collections/',
      })
    })
    cy.wait('@loadCollectionSettings')
    cy.wait('@collectionsInventoryItems')
    cy.intercept('PUT', '/api/profiles/e2e-profile-001/settings').as(
      'saveCollectionSettings'
    )
  }

  it('UI-SCREEN-COLLECTIONS-031 sorts collections without passive settings writes', () => {
    seedSortableCollectionsProfile()

    cy.get('[data-testid="collections-active-context"]').should(
      'contain.text',
      'All Items'
    )
    rowTestIDs(
      '[data-testid="collections-shared-table"] tbody tr[data-testid^="collections-row-"]'
    ).should('deep.equal', [
      'collections-row-all-items',
      'collections-row-alpha-bin',
      'collections-row-middle-shelf',
      'collections-row-zulu-vault',
    ])
    cy.get('[data-testid="collections-members-table"]').should('not.exist')
    cy.get('@saveCollectionSettings.all').should('have.length', 0)

    sortColumnDescending('[data-testid="collections-shared-table"]', 'Collection')
    rowTestIDs(
      '[data-testid="collections-shared-table"] tbody tr[data-testid^="collections-row-"]'
    ).should('deep.equal', [
      'collections-row-zulu-vault',
      'collections-row-middle-shelf',
      'collections-row-alpha-bin',
      'collections-row-all-items',
    ])
    cy.get('[data-testid="collections-active-context"]').should(
      'contain.text',
      'All Items'
    )
    cy.get('@saveCollectionSettings.all').should('have.length', 0)

    cy.get('[data-testid="collections-active-context"]').should(
      'contain.text',
      'All Items'
    )
    cy.get('@saveCollectionSettings.all').should('have.length', 0)
  })
})
