describe('workspace header action lanes', () => {
  const inventoryItems = [
    {
      id: 'item-header-alpha',
      part_number: 'PN-HEAD-A',
      title: 'Header Alpha',
      status: 'active',
      category: 'Cars',
      brand: 'AFX',
      priority: 'medium',
      description: 'Header action coverage item',
    },
  ]

  function bootstrap(path: string) {
    cy.e2eReset()
    cy.e2eSetSetupState('present')
    cy.e2eBootstrap().then(({ profile_id, profile_name }) => {
      cy.useBootstrappedProfile(profile_id, profile_name, { path })
    })
  }

  function assertWorkspaceStartsBelowHeader(
    headerTestID: string,
    workspaceTestID: string
  ) {
    cy.get(`[data-testid="${headerTestID}"]`).then(($header) => {
      const headerBottom = $header[0].getBoundingClientRect().bottom
      cy.get(`[data-testid="${workspaceTestID}"]`).then(($workspace) => {
        const workspaceTop = $workspace[0].getBoundingClientRect().top
        expect(workspaceTop - headerBottom).to.be.within(0, 32)
      })
    })
  }

  it('keeps Inventory actions in the global shell header without duplicate workspace chrome', () => {
    cy.viewport(1280, 800)
    cy.intercept('GET', '/api/items', {
      statusCode: 200,
      body: { items: inventoryItems },
    }).as('items')

    bootstrap('/inventory/')
    cy.wait('@items')

    cy.get('[data-testid="inventory-shell-header"]').should('be.visible')
    cy.get('[data-testid="inventory-page-header"]').should('not.exist')
    cy.get('[data-testid="inventory-global-header-actions"]')
      .should('be.visible')
      .within(() => {
        cy.contains('Inventory').should('be.visible')
        cy.get('[data-testid="inventory-header-context"]').should(
          'contain.text',
          'All Items'
        )
        cy.get('[data-testid="inventory-new-action"]').should('be.visible')
        cy.get('[data-testid="inventory-create-menu-trigger"]').should(
          'be.visible'
        )
      })
    cy.get('[data-testid="inventory-header-action-separator"]').should('be.visible')
    cy.contains('Browse, organize, and update the items you already own.').should(
      'not.exist'
    )
    cy.get('[data-testid="collection-context-label"]').should('not.exist')
    assertWorkspaceStartsBelowHeader('inventory-shell-header', 'inventory-workspace')
  })

  it('keeps Wishlist actions in the same global shell header pattern', () => {
    cy.viewport(1280, 800)
    cy.intercept('GET', '/api/wishlist', {
      statusCode: 200,
      body: {
        items: [
          {
            id: 'wish-header-alpha',
            item_id: 'item-header-wishlist',
            priority: 'high',
            below_target_now: true,
          },
        ],
      },
    }).as('wishlistItems')
    cy.intercept('GET', '/api/items?status=wishlist', {
      statusCode: 200,
      body: {
        items: [
          {
            id: 'item-header-wishlist',
            title: 'Wishlist Header Alpha',
            part_number: 'WISH-HEAD-A',
            status: 'wishlist',
            category: 'Cards',
            priority: 'high',
          },
        ],
      },
    }).as('wishlistCatalog')
    cy.intercept('GET', '/api/profiles/*/settings').as('profileSettings')

    bootstrap('/wishlist/')
    cy.wait('@wishlistItems')
    cy.wait('@wishlistCatalog')
    cy.wait('@profileSettings')

    cy.get('[data-testid="wishlist-shell-header"]').should('be.visible')
    cy.get('[data-testid="wishlist-page-header"]').should('not.exist')
    cy.get('[data-testid="wishlist-global-header-actions"]')
      .should('be.visible')
      .within(() => {
        cy.contains('Wishlist').should('be.visible')
        cy.get('[data-testid="wishlist-new-action"]').should('be.visible')
        cy.get('[data-testid="wishlist-create-menu-trigger"]').should(
          'be.visible'
        )
      })
    cy.get('[data-testid="wishlist-header-action-separator"]').should('be.visible')
    cy.contains(
      'Track wanted items, target prices, and planning decisions before they become owned inventory.'
    ).should('not.exist')
    assertWorkspaceStartsBelowHeader(
      'wishlist-shell-header',
      'wishlist-workspace'
    )
  })

  it('keeps Collections actions in the same global shell header pattern', () => {
    cy.viewport(1280, 800)
    bootstrap('/collections/')

    cy.get('[data-testid="collections-shell-header"]').should('be.visible')
    cy.get('[data-testid="collections-page-header"]').should('not.exist')
    cy.get('[data-testid="collections-global-header-actions"]')
      .should('be.visible')
      .within(() => {
        cy.contains('Collections').should('be.visible')
        cy.get('[data-testid="collections-page-icon"]').should('be.visible')
        cy.get('[data-testid="collections-new-action"]').should('be.visible')
      })
    cy.get('[data-testid="collections-header-action-separator"]').should('be.visible')
    cy.contains(
      'Manage collection rows and item placement from the shared Cabinet table surface.'
    ).should('not.exist')
    assertWorkspaceStartsBelowHeader(
      'collections-shell-header',
      'collections-workspace'
    )
  })
})
