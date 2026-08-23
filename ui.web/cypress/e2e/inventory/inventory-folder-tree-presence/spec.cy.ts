describe('inventory-folder-tree-presence', () => {
  function openInventory() {
    cy.e2eReset()
    cy.e2eSetSetupState('present')
    cy.e2eBootstrap().then(({ profile_id, profile_name }) => {
      cy.useBootstrappedProfile(profile_id, profile_name, {
        path: '/inventory/',
      })
    })
  }

  it('keeps the inventory folder tree visible and wired to context filtering', () => {
    cy.intercept('GET', '/api/items', {
      statusCode: 200,
      body: {
        items: [
          {
            id: 'item-tree-visible-1',
            part_number: 'PN-TREE-VISIBLE-1',
            title: 'Store One Tree Item',
            status: 'active',
            category: 'Cars',
          },
          {
            id: 'item-tree-visible-2',
            part_number: 'PN-TREE-VISIBLE-2',
            title: 'Watch List Tree Item',
            status: 'active',
            category: 'Cars',
          },
        ],
      },
    }).as('itemsForTree')

    openInventory()
    cy.wait('@itemsForTree')
    cy.window().then((win) => {
      win.localStorage.setItem(
        'cabinet.inventory.item-folder-assignments.v1',
        JSON.stringify({
          'item-tree-visible-1': 'Store 1',
          'item-tree-visible-2': 'Watch List',
        })
      )
    })
    cy.reload()
    cy.wait('@itemsForTree')

    cy.get('[data-testid="inventory-folder-tree"][role="tree"]').should(
      'be.visible'
    )
    cy.get('[data-testid="folder-tree-item-all-items"]').should('be.visible')
    cy.get('[data-testid="folder-tree-item-store-1"]').should('be.visible')
    cy.get('[data-testid="collection-folder-all-items"]').should(
      'contain.text',
      'All Items'
    )

    cy.get('[data-testid="folder-tree-item-store-1"]').click()
    cy.get('[data-testid="collection-active-context"]').should(
      'contain.text',
      'Store 1'
    )
    cy.get('[data-testid="inventory-item-row-item-tree-visible-1"]')
      .scrollIntoView()
      .should('be.visible')
      .within(() => {
        cy.contains('Store One Tree Item').should('be.visible')
      })
    cy.contains('Watch List Tree Item').should('not.exist')
  })
})
