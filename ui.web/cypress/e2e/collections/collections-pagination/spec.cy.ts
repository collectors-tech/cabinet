describe('collections-pagination', () => {
  const collectionsSettingsKey = 'collections.workspace.v1'
  const targetCollection = 'Zed Shelf 12'

  const seededCollections = [
    'All Items',
    'Watch List',
    'Warehouse 1',
    'Store 1',
    'Store 2',
    'Overflow',
    ...Array.from({ length: 12 }, (_, index) =>
      `Zed Shelf ${String(index + 1).padStart(2, '0')}`
    ),
  ]

  const seededItems = [
    {
      id: 'pagination-item-12',
      name: 'Pagination proof card',
      detail: 'Assigned to later-page collection',
      collectionName: targetCollection,
    },
  ]

  function seedPaginatedCollectionsProfile() {
    cy.viewport(1512, 967)
    cy.e2eReset()
    cy.e2eSetSetupState('present')
    cy.intercept('GET', '/api/items', {
      statusCode: 200,
      body: {
        items: [
          {
            id: 'pagination-item-12',
            part_number: 'PAGE-012',
            title: 'Pagination proof card',
            status: 'active',
            category: 'Trading Card',
            brand: 'Cabinet',
            description: 'Assigned to later-page collection',
          },
        ],
      },
    }).as('collectionsInventoryItems')
    cy.intercept('GET', '/api/profiles/*/settings').as(
      'loadCollectionSettings'
    )
    cy.e2eBootstrap().then(({ profile_id, profile_name }) => {
      cy.request('PUT', `/api/profiles/${profile_id}/settings`, {
        settings: {
          [collectionsSettingsKey]: JSON.stringify({
            collections: seededCollections,
            activeCollection: 'All Items',
            items: seededItems,
          }),
        },
      })
        .its('status')
        .should('eq', 200)
      cy.e2eEnsureSignedOut()
      cy.stubLocalServerSession(profile_id)
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

  function persistedCollectionsSettings(requestBody: unknown) {
    const body = requestBody as { settings?: Record<string, string> }
    return JSON.parse(body.settings?.[collectionsSettingsKey] ?? '{}') as {
      activeCollection?: string
    }
  }

  function expectInitialCollectionsPage() {
    cy.get('[data-testid="collections-management-summary"]').should(
      'contain.text',
      'Showing 18 of 18 collections.'
    )
    cy.get('[data-testid="collections-table-pagination"]').should(
      'contain.text',
      'Page 1 of 2'
    )
    cy.get('[data-testid="collections-row-zed-shelf-12"]').should('not.exist')
    cy.get('@saveCollectionSettings.all').should('have.length', 0)
  }

  function navigateToPageTwo() {
    cy.get('[data-testid="collections-table-pagination"]')
      .contains('button', '2')
      .click()
    cy.get('[data-testid="collections-table-pagination"]').should(
      'contain.text',
      'Page 2 of 2'
    )
    cy.get('[data-testid="collections-row-zed-shelf-12"]')
      .scrollIntoView()
      .should('be.visible')
  }

  function selectTargetCollection() {
    navigateToPageTwo()
    cy.get('[data-testid="collections-row-zed-shelf-12"]').click()
    cy.wait('@saveCollectionSettings').then(({ request }) => {
      expect(persistedCollectionsSettings(request.body).activeCollection).to.eq(
        targetCollection
      )
    })
    cy.get('[data-testid="collections-active-context"]').should(
      'contain.text',
      targetCollection
    )
  }

  it('UI-SCREEN-COLLECTIONS-028 reaches page two without passive settings writes', () => {
    seedPaginatedCollectionsProfile()
    expectInitialCollectionsPage()

    navigateToPageTwo()
    cy.get('@saveCollectionSettings.all').should('have.length', 0)
  })

  it('UI-SCREEN-COLLECTIONS-028 keeps a later-page selection on its page and persists active context', () => {
    seedPaginatedCollectionsProfile()

    selectTargetCollection()
    cy.get('[data-testid="collections-table-pagination"]').should(
      'contain.text',
      'Page 2 of 2'
    )
    cy.get('[data-testid="collections-row-zed-shelf-12"]').should(
      'have.attr',
      'data-state',
      'selected'
    )
    cy.get('[data-testid="collections-members-table"]').should('not.exist')
    cy.get('@saveCollectionSettings.all').should('have.length', 1)

    cy.reload()
    cy.wait('@loadCollectionSettings')
    cy.wait('@collectionsInventoryItems')
    cy.get('[data-testid="collections-active-context"]').should(
      'contain.text',
      targetCollection
    )
    navigateToPageTwo()
    cy.get('[data-testid="collections-row-zed-shelf-12"]').should(
      'have.attr',
      'data-state',
      'selected'
    )
    cy.get('@saveCollectionSettings.all').should('have.length', 1)
  })

  it('UI-SCREEN-COLLECTIONS-028 keeps selected context reachable while filtering without passive writes', () => {
    seedPaginatedCollectionsProfile()
    selectTargetCollection()

    cy.get('[data-testid="collections-search-input"]').type(targetCollection)
    cy.get('[data-testid="collections-row-zed-shelf-12"]').should(
      'have.attr',
      'data-state',
      'selected'
    )
    cy.get('[data-testid="collections-management-summary"]').should(
      'contain.text',
      'Showing 1 of 18 collections.'
    )
    cy.get('[data-testid="collections-table-pagination"]').should(
      'contain.text',
      'Page 1 of 1'
    )
    cy.get('@saveCollectionSettings.all').should('have.length', 1)
  })

  it('UI-SCREEN-COLLECTIONS-028 returns a filtered selection to its full-list pagination page', () => {
    seedPaginatedCollectionsProfile()
    selectTargetCollection()

    cy.get('[data-testid="collections-search-input"]').type(targetCollection)
    cy.get('[data-testid="collections-row-zed-shelf-12"]').should(
      'have.attr',
      'data-state',
      'selected'
    )
    cy.get('[data-testid="collections-search-input"]').clear()
    cy.get('[data-testid="collections-management-summary"]').should(
      'contain.text',
      'Showing 18 of 18 collections.'
    )
    cy.get('[data-testid="collections-table-pagination"]').should(
      'contain.text',
      'Page 2 of 2'
    )
    cy.get('[data-testid="collections-row-zed-shelf-12"]').should(
      'have.attr',
      'data-state',
      'selected'
    )
    cy.get('@saveCollectionSettings.all').should('have.length', 1)
  })
})
