describe('collections-members-pagination', () => {
  const collectionsSettingsKey = 'collections.workspace.v1'
  const selectedCollection = 'Store 1'
  const laterPageMemberID = 'members-page-item-12'
  const laterPageMemberName = 'Members page proof 12'

  const memberItems = Array.from({ length: 12 }, (_, index) => {
    const ordinal = index + 1
    return {
      id: `members-page-item-${String(ordinal).padStart(2, '0')}`,
      name: `Members page proof ${String(ordinal).padStart(2, '0')}`,
      detail: `Selected collection member ${String(ordinal).padStart(2, '0')}`,
      collectionName: selectedCollection,
    }
  })

  function seedSelectedCollectionWithPaginatedMembers() {
    cy.viewport(1512, 967)
    cy.e2eReset()
    cy.e2eSetSetupState('present')
    cy.intercept('GET', '/api/items', {
      statusCode: 200,
      body: {
        items: memberItems.map((item) => ({
          id: item.id,
          part_number: item.id.toUpperCase(),
          title: item.name,
          status: 'active',
          category: 'Trading Card',
          brand: 'Cabinet',
          description: item.detail,
        })),
      },
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
              'Watch List',
              'Warehouse 1',
              selectedCollection,
              'Store 2',
              'Overflow',
            ],
            activeCollection: selectedCollection,
            items: memberItems,
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

  it('UI-SCREEN-COLLECTIONS-029 preserves paginated collection members without passive settings writes', () => {
    seedSelectedCollectionWithPaginatedMembers()

    cy.get('[data-testid="collections-active-context"]').should(
      'contain.text',
      selectedCollection
    )
    cy.get('[data-testid="collections-members-summary"]').should(
      'contain.text',
      'Showing 12 of 12 items.'
    )
    cy.get('[data-testid="collections-members-table-pagination"]').should(
      'contain.text',
      'Page 1 of 2'
    )
    cy.get(`[data-testid="collections-member-row-${laterPageMemberID}"]`).should(
      'not.exist'
    )
    cy.get('@saveCollectionSettings.all').should('have.length', 0)

    cy.get('[data-testid="collections-members-table-pagination"]')
      .contains('button', '2')
      .click()
    cy.get(`[data-testid="collections-member-row-${laterPageMemberID}"]`)
      .scrollIntoView()
      .should('be.visible')
      .and('contain.text', laterPageMemberName)
      .and('contain.text', `Currently in ${selectedCollection}.`)
    cy.get('[data-testid="collections-active-context"]').should(
      'contain.text',
      selectedCollection
    )
    cy.get('@saveCollectionSettings.all').should('have.length', 0)

    cy.reload()
    cy.wait('@loadCollectionSettings')
    cy.wait('@collectionsInventoryItems')
    cy.get('[data-testid="collections-active-context"]').should(
      'contain.text',
      selectedCollection
    )
    cy.get('[data-testid="collections-members-table-pagination"]')
      .contains('button', '2')
      .click()
    cy.get(`[data-testid="collections-member-row-${laterPageMemberID}"]`).should(
      'be.visible'
    )

    cy.get('[data-testid="collections-members-search-input"]').type(
      laterPageMemberName
    )
    cy.get(`[data-testid="collections-member-row-${laterPageMemberID}"]`)
      .should('be.visible')
      .and('contain.text', laterPageMemberName)
    cy.get('[data-testid="collections-members-summary"]').should(
      'contain.text',
      'Showing 1 of 12 items.'
    )
    cy.get('[data-testid="collections-active-context"]').should(
      'contain.text',
      selectedCollection
    )
    cy.get('@saveCollectionSettings.all').should('have.length', 0)
  })
})
