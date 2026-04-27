describe('wishlist-visible-data', () => {
  const collectionsSettingsKey = 'collections.workspace.v1'

  function openWishlistWithStaleCollectionContext() {
    cy.e2eReset()
    cy.e2eSetSetupState('present')
    cy.e2eBootstrap().then(({ profile_id, profile_name }) => {
      cy.request('PUT', `/api/profiles/${profile_id}/settings`, {
        settings: {
          [collectionsSettingsKey]: JSON.stringify({
            collections: ['All Items', 'Overflow'],
            activeCollection: 'Overflow',
            items: [],
          }),
        },
      }).its('status').should('eq', 200)

      cy.useBootstrappedProfile(profile_id, profile_name, {
        path: '/wishlist/',
      })
    })
  }

  it('shows seeded wishlist rows even after another screen leaves an empty collection active', () => {
    cy.intercept('GET', '/api/wishlist').as('wishlistEntries')
    cy.intercept('GET', '/api/items?status=wishlist').as('wishlistItems')
    cy.intercept('GET', '/api/profiles/*/settings').as('profileSettings')

    openWishlistWithStaleCollectionContext()

    cy.wait('@wishlistEntries')
    cy.wait('@wishlistItems')
    cy.wait('@profileSettings')

    cy.get('[data-testid="wishlist-table-collection-selected"]').should(
      'contain',
      'All Items'
    )
    cy.contains('Wishlist Sample Grail Chase').should('be.visible')
    cy.contains('Wishlist Sample Price Drop Watch').should('be.visible')
    cy.contains('Wishlist Sample Steady Watch').should('be.visible')
    cy.contains('No results.').should('not.exist')
  })

  it('does not flash generic task seed rows while wishlist API data is loading', () => {
    cy.intercept('GET', '/api/wishlist', {
      delay: 3000,
      statusCode: 200,
      body: {
        items: [
          {
            id: 'wish-visible-1',
            item_id: 'wish-item-visible-1',
            priority: 'high',
            below_target_now: false,
          },
        ],
      },
    }).as('slowWishlistEntries')
    cy.intercept('GET', '/api/items?status=wishlist', {
      delay: 3000,
      statusCode: 200,
      body: {
        items: [
          {
            id: 'wish-item-visible-1',
            title: 'Visible Wishlist API Row',
            part_number: 'WISH-VISIBLE-1',
            status: 'wishlist',
            category: 'Cards',
            priority: 'high',
          },
        ],
      },
    }).as('slowWishlistItems')

    cy.e2eReset()
    cy.e2eSetSetupState('present')
    cy.e2eBootstrap().then(({ profile_id, profile_name }) => {
      cy.useBootstrappedProfile(profile_id, profile_name, {
        path: '/wishlist/',
      })
    })

    cy.get('body').then(($body) => {
      expect($body.text()).not.to.match(/TASK-\d+/)
    })

    cy.wait('@slowWishlistEntries')
    cy.wait('@slowWishlistItems')
    cy.contains('Visible Wishlist API Row').should('be.visible')
    cy.contains(/TASK-\d+/).should('not.exist')
  })
})
