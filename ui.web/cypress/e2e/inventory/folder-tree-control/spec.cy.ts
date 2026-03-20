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

  it('UI-SCREEN-INVENTORY-FOLDER-TREE-007 allows node-level add-child creation', () => {
    cy.get('[data-testid="folder-tree-add-child-store-1"]').click()
    cy.get('[data-testid="folder-tree-name-input"]').clear().type('Store 1 Child')
    cy.get('[data-testid="folder-tree-create-submit"]').click()
    cy.contains('[role="treeitem"]', 'Store 1 Child').should('be.visible')
  })

  it('UI-SCREEN-INVENTORY-FOLDER-TREE-008 allows explicit root folder creation through a compact plus affordance', () => {
    cy.get('[data-testid="folder-tree-add-root"]')
      .should('have.text', '+')
      .and('have.attr', 'aria-label', 'Add root folder')
      .click()
    cy.get('[data-testid="folder-tree-name-input"]').clear().type('Top Level Added')
    cy.get('[data-testid="folder-tree-create-submit"]').click()
    cy.get('[data-testid="folder-tree-item-top-level-added"]').should('be.visible')
  })

  it('UI-SCREEN-INVENTORY-FOLDER-TREE-009 renders hierarchy connector lines', () => {
    cy.get('[data-testid="folder-tree-toggle-warehouses"]').click()
    cy.get('[data-testid="folder-tree-group-warehouses"]')
      .should('be.visible')
      .and('have.class', 'border-l')
  })

  it('UI-SCREEN-INVENTORY-FOLDER-TREE-010 separates disclosure from selection and clarifies branch vs leaf cues', () => {
    cy.get('[data-testid="folder-tree-item-all-items"]')
      .should('have.attr', 'aria-selected', 'true')
      .and('have.attr', 'data-active', 'true')
      .and('have.attr', 'data-node-kind', 'leaf')

    cy.get('[data-testid="folder-tree-toggle-warehouses"]')
      .should('have.attr', 'data-state', 'collapsed')
      .click()
      .should('have.attr', 'data-state', 'expanded')

    cy.get('[data-testid="collection-active-context"]').should('contain.text', 'All Items')
    cy.get('[data-testid="folder-tree-item-warehouses"]')
      .should('have.attr', 'data-node-kind', 'branch')
      .and('have.attr', 'data-node-expanded', 'true')

    cy.get('[data-testid="folder-tree-item-store-1"]').should('have.attr', 'data-node-kind', 'leaf')
    cy.get('[data-testid="folder-tree-toggle-store-1"]').should('not.exist')

    cy.get('[data-testid="folder-tree-item-store-1"]').click()
    cy.get('[data-testid="collection-active-context"]').should('contain.text', 'Store 1')
    cy.get('[data-testid="folder-tree-item-store-1"]')
      .should('have.attr', 'aria-selected', 'true')
      .and('have.attr', 'data-active', 'true')
    cy.get('[data-testid="folder-tree-item-all-items"]').should('have.attr', 'data-active', 'false')
  })

  it('UI-SCREEN-INVENTORY-FOLDER-TREE-013 renders structured metadata without breaking row readability', () => {
    cy.get('[data-testid="folder-tree-secondary-all-items"]')
      .should('be.visible')
      .and('have.text', 'Entire catalog')
    cy.get('[data-testid="folder-tree-count-all-items"]')
      .should('be.visible')
      .and('have.text', '124')
    cy.get('[data-testid="folder-tree-badge-all-items"]')
      .should('be.visible')
      .and('have.text', 'Live')

    cy.get('[data-testid="folder-tree-item-all-items"]').then(($row) => {
      const rowRect = $row[0].getBoundingClientRect()

      cy.get('[data-testid="collection-folder-all-items"]').then(($label) => {
        const labelRect = $label[0].getBoundingClientRect()

        cy.get('[data-testid="folder-tree-count-all-items"]').then(($count) => {
          const countRect = $count[0].getBoundingClientRect()
          expect(countRect.left, 'count trails label content').to.be.greaterThan(labelRect.right + 8)
          expect(rowRect.right - countRect.right, 'count remains near row end').to.be.lessThan(80)
        })
      })
    })

    cy.get('[data-testid="folder-tree-toggle-warehouses"]').click()
    cy.get('[data-testid="folder-tree-secondary-warehouse-1"]')
      .should('be.visible')
      .and('have.text', 'Pallet zone A')
    cy.get('[data-testid="folder-tree-count-warehouse-1"]').should('have.text', '15')
    cy.get('[data-testid="folder-tree-item-warehouse-1"]')
      .should('have.attr', 'role', 'treeitem')
      .click()
    cy.get('[data-testid="collection-active-context"]').should('contain.text', 'Warehouse 1')
  })

  it('UI-SCREEN-INVENTORY-FOLDER-TREE-014 supports live drag-drop reparenting with visible feedback', () => {
    const dataTransfer = new DataTransfer()

    cy.get('[data-testid="folder-tree-toggle-warehouses"]').click()
    cy.get('[data-testid="folder-tree-group-warehouses"]').should('be.visible')

    cy.get('[data-testid="folder-tree-item-store-1"]')
      .trigger('dragstart', { dataTransfer })
    cy.get('[data-testid="folder-tree-item-warehouses"]')
      .trigger('dragenter', { dataTransfer })
      .trigger('dragover', { dataTransfer })
      .should('have.class', 'bg-primary/20')
      .trigger('drop', { dataTransfer })

    cy.get('[data-testid="folder-tree-item-store-1"]')
      .should('have.attr', 'aria-selected', 'true')
      .and('have.attr', 'data-active', 'true')

    cy.get('[data-testid="folder-tree-group-warehouses"] [data-testid="folder-tree-item-store-1"]')
      .should('exist')
    cy.get('[data-testid="collection-active-context"]').should('contain.text', 'Store 1')
  })
})
