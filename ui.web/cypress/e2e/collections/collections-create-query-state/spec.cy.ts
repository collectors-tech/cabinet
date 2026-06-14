describe('collections-create-query-state', () => {
  const collectionsSettingsKey = 'collections.workspace.v1'

  function signInToCollectionsCreateQuery() {
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
    cy.visit('/collections/?create=collection')
  }

  it('opens create collection from direct route query state and persists the result', () => {
    signInToCollectionsCreateQuery()
    cy.intercept('PUT', '/api/profiles/e2e-profile-001/settings').as(
      'saveCollectionSettings'
    )

    cy.get('[data-testid="collections-create-dialog"]').should('be.visible')
    cy.location('pathname', { timeout: 15000 }).should(
      'match',
      /^\/collections\/?$/
    )
    cy.location('search').should('not.contain', 'create=collection')

    cy.get('[data-testid="collections-create-input"]').type('Route Query Shelf')
    cy.get('[data-testid="collections-create-submit"]').click()
    cy.wait('@saveCollectionSettings').then(({ request }) => {
      const settings = (request.body.settings ?? {}) as Record<string, string>
      const persisted = JSON.parse(settings[collectionsSettingsKey] ?? '{}') as {
        collections?: string[]
        activeCollection?: string
      }

      expect(persisted.collections).to.include('Route Query Shelf')
      expect(persisted.activeCollection).to.equal('Route Query Shelf')
    })

    cy.get('[data-testid="collections-create-dialog"]').should('not.exist')
    cy.get('[data-testid="collections-row-route-query-shelf"]')
      .scrollIntoView()
      .should('be.visible')
    cy.get('[data-testid="collections-active-context"]').should(
      'contain.text',
      'Route Query Shelf'
    )

    cy.reload()
    cy.get('[data-testid="collections-create-dialog"]').should('not.exist')
    cy.get('[data-testid="collections-row-route-query-shelf"]')
      .scrollIntoView()
      .should('be.visible')
  })
})
