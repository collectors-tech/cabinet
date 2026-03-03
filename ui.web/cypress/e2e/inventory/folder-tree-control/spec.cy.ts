describe('inventory-folder-tree-control', () => {
  function signIn() {
    cy.visit('/sign-in?redirect=%2Finventory%2F')
    cy.get('input[name="email"]').clear().type('e2e-folder-tree@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)
  }

  beforeEach(() => {
    cy.intercept('GET', '/api/items', {
      statusCode: 200,
      body: {
        items: [
          {
            id: 'item-tree-1',
            part_number: 'PN-TREE-1',
            title: 'Tree Item 1',
            status: 'todo',
            category: 'feature',
          },
        ],
      },
    }).as('items')
    signIn()
    cy.wait('@items')
  })

  it('UI-SCREEN-INVENTORY-FOLDER-TREE-001 renders deterministic nested expand/collapse hierarchy', () => {
    cy.get('[data-testid="inventory-folder-tree"][role="tree"]').should('be.visible')
    cy.get('[data-testid="folder-tree-toggle-warehouses"]').as('warehousesToggle')

    cy.get('@warehousesToggle').should('have.attr', 'aria-expanded', 'false').click()
    cy.get('@warehousesToggle').should('have.attr', 'aria-expanded', 'true')
    cy.get('[data-testid="folder-tree-group-warehouses"]').should('be.visible')
    cy.contains('[role="treeitem"]', 'Warehouse 1').should('be.visible')

    cy.get('@warehousesToggle').click()
    cy.get('@warehousesToggle').should('have.attr', 'aria-expanded', 'false')
    cy.get('[data-testid="folder-tree-group-warehouses"]').should('not.exist')
  })

  it('UI-SCREEN-INVENTORY-FOLDER-TREE-002 supports keyboard tree semantics with accessible roles', () => {
    cy.get('[data-testid="inventory-folder-tree"][role="tree"]').focus()
    cy.get('[data-testid="folder-tree-item-watch-list"]').focus().type('{rightarrow}')
    cy.get('[data-testid="folder-tree-item-watch-list"]').should('have.attr', 'aria-selected', 'true')
    cy.get('[data-testid="folder-tree-item-watch-list"]').type('{enter}')
    cy.get('[data-testid="collection-active-context"]').should('contain.text', 'Watch List')
    cy.get('[data-testid="folder-tree-item-watch-list"]').should('have.attr', 'role', 'treeitem')
  })

  it('UI-SCREEN-INVENTORY-FOLDER-TREE-003 selection deterministically updates inventory context', () => {
    cy.get('[data-testid="folder-tree-item-store-2"]').click()
    cy.get('[data-testid="collection-active-context"]').should('contain.text', 'Store 2')

    cy.get('[data-testid="folder-tree-item-all-items"]').click()
    cy.get('[data-testid="collection-active-context"]').should('contain.text', 'All Items')
  })

  it('UI-SCREEN-INVENTORY-FOLDER-TREE-004 remains responsive for deep/large hierarchy interactions', () => {
    cy.get('[data-testid="inventory-folder-tree"]').within(() => {
      cy.get('[role="treeitem"]').should('have.length.greaterThan', 20)
    })

    cy.get('[data-testid="folder-tree-toggle-warehouses"]').click()
    cy.contains('[role="treeitem"]', 'Warehouse 2').click()
    cy.get('[data-testid="collection-active-context"]').should('contain.text', 'Warehouse 2')
  })

  it('UI-SCREEN-INVENTORY-FOLDER-TREE-005 keeps tree scrolling inside pane without growing full page', () => {
    cy.get('[data-testid="inventory-folder-tree"]').as('treeScrollRegion')

    cy.document().then((doc) => {
      const baseline = doc.documentElement.scrollHeight
      cy.wrap(baseline).as('baselineHeight')
    })

    cy.get('[data-testid="folder-tree-toggle-warehouses"]').click()
    cy.get('[data-testid="folder-tree-group-warehouses"]').should('be.visible')

    cy.get('@treeScrollRegion').then(($region) => {
      const node = $region[0]
      const styles = getComputedStyle(node)
      expect(['auto', 'scroll']).to.include(styles.overflowY)
      expect(node.scrollHeight).to.be.greaterThan(node.clientHeight)
    })

    cy.get('@baselineHeight').then((baselineHeight) => {
      cy.document().then((doc) => {
        const current = doc.documentElement.scrollHeight
        expect(current).to.be.lessThan(Number(baselineHeight) + 120)
      })
    })
  })

  it('UI-SCREEN-INVENTORY-FOLDER-TREE-006 provides horizontal overflow access for deep indentation', () => {
    cy.get('[data-testid="folder-tree-toggle-warehouses"]').click()
    cy.get('[data-testid="inventory-folder-tree"]').then(($region) => {
      const node = $region[0]
      const styles = getComputedStyle(node)
      expect(['auto', 'scroll']).to.include(styles.overflowX)
    })
  })
})
