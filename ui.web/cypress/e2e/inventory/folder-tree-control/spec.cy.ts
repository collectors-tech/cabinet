describe('inventory-folder-tree-control', () => {
  function getFocusableElements(doc: Document) {
    return Array.from(
      doc.querySelectorAll<HTMLElement>(
        'a[href], button, input, select, textarea, [tabindex], [contenteditable="true"]'
      )
    ).filter((element) => {
      if (element.hasAttribute('disabled') || element.getAttribute('aria-hidden') === 'true') {
        return false
      }

      const tabIndexValue = element.getAttribute('tabindex')
      if (tabIndexValue !== null && Number(tabIndexValue) < 0) {
        return false
      }

      const styles = element.ownerDocument.defaultView?.getComputedStyle(element)
      if (!styles || styles.display === 'none' || styles.visibility === 'hidden') {
        return false
      }

      return element.getClientRects().length > 0
    })
  }

  function getNextFocusableElement(doc: Document, current: HTMLElement) {
    const focusableElements = getFocusableElements(doc)
    const currentIndex = focusableElements.indexOf(current)

    expect(currentIndex, `${current.dataset.testid ?? current.tagName} is tabbable`).to.be.greaterThan(
      -1
    )

    return focusableElements[currentIndex + 1] ?? null
  }

  beforeEach(() => {
    cy.visit('about:blank')
    cy.clearCookies()
    cy.clearLocalStorage()
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.intercept('GET', '/api/items', {
      statusCode: 200,
      body: {
        items: [
          {
            id: 'e2e-item-001',
            part_number: 'PN-TREE-1',
            title: 'Tree Item 1',
            status: 'todo',
            category: 'feature',
          },
          {
            id: 'e2e-item-002',
            part_number: 'PN-TREE-2',
            title: 'Tree Item 2',
            status: 'todo',
            category: 'feature',
          },
        ],
      },
    }).as('items')
    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', { path: '/inventory/' })
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
    cy.get('[data-testid="folder-tree-toggle-warehouses"]').click()
    cy.get('[data-testid="folder-tree-item-warehouse-1"]').click()

    cy.document().then((doc) => {
      const rootAddButton = doc.querySelector<HTMLElement>('[data-testid="folder-tree-add-root"]')
      const activeRow = doc.querySelector<HTMLElement>('[data-testid="folder-tree-item-warehouse-1"]')

      expect(rootAddButton, 'root add affordance exists').to.not.equal(null)
      expect(activeRow, 'selected tree row exists').to.not.equal(null)

      const nextAfterRoot = getNextFocusableElement(doc, rootAddButton as HTMLElement)
      expect(nextAfterRoot, 'tab entry lands on the active row').to.equal(activeRow)

      const addChildButton = doc.querySelector<HTMLElement>('[data-testid="folder-tree-add-child-warehouse-1"]')
      const rowActionsButton = doc.querySelector<HTMLElement>('[data-testid="folder-tree-row-actions-warehouse-1"]')

      expect(addChildButton?.getAttribute('tabindex'), 'inline add-child control is removed from normal tab order').to.equal('-1')
      expect(rowActionsButton?.getAttribute('tabindex'), 'row actions trigger is removed from normal tab order').to.equal('-1')

      const nextAfterActiveRow = getNextFocusableElement(doc, activeRow as HTMLElement)
      expect(
        nextAfterActiveRow?.closest('[data-testid="inventory-folder-tree"]') ?? null,
        'tab exits the tree instead of drifting through inline controls'
      ).to.equal(null)
    })

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

  it('UI-SCREEN-INVENTORY-FOLDER-TREE-011 provides structured row actions without stealing selection', () => {
    cy.get('[data-testid="folder-tree-item-all-items"]')
      .should('have.attr', 'data-active', 'true')

    cy.get('[data-testid="folder-tree-row-actions-store-1"]').click()
    cy.get('[data-testid="folder-tree-item-all-items"]')
      .should('have.attr', 'data-active', 'true')
    cy.get('[data-testid="folder-tree-item-store-1"]')
      .should('have.attr', 'data-active', 'false')

    cy.get('[data-testid="folder-tree-row-action-add-child-store-1"]').click()
    cy.contains('[role="dialog"]', 'Add Child Folder').should('be.visible')
    cy.get('[data-testid="folder-tree-item-all-items"]')
      .should('have.attr', 'data-active', 'true')
    cy.get('[data-testid="folder-tree-create-cancel"]').click()

    cy.get('[data-testid="folder-tree-row-actions-store-1"]').click()
    cy.get('[data-testid="folder-tree-row-action-select-store-1"]').click()
    cy.get('[data-testid="folder-tree-item-store-1"]')
      .should('have.attr', 'data-active', 'true')
      .and('have.attr', 'aria-selected', 'true')
    cy.get('[data-testid="collection-active-context"]').should('contain.text', 'Store 1')
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

  it('UI-SCREEN-INVENTORY-FOLDER-TREE-012 restores selected folder and expanded branch context after reload', () => {
    cy.get('[data-testid="folder-tree-toggle-warehouses"]').click()
    cy.get('[data-testid="folder-tree-item-warehouse-2"]').click()
    cy.get('[data-testid="collection-active-context"]').should('contain.text', 'Warehouse 2')

    cy.reload()
    cy.wait('@items')

    cy.get('[data-testid="folder-tree-toggle-warehouses"]')
      .should('have.attr', 'aria-expanded', 'true')
    cy.get('[data-testid="folder-tree-group-warehouses"]').should('be.visible')
    cy.get('[data-testid="folder-tree-item-warehouse-2"]')
      .should('have.attr', 'aria-selected', 'true')
      .and('have.attr', 'data-active', 'true')
    cy.get('[data-testid="collection-active-context"]').should('contain.text', 'Warehouse 2')
  })

  it('UI-SCREEN-INVENTORY-FOLDER-TREE-015 persists created folders and folder property edits after reload', () => {
    cy.get('[data-testid="folder-tree-row-actions-store-1"]').click()
    cy.get('[data-testid="folder-tree-row-action-properties-store-1"]').click()
    cy.get('[data-testid="folder-properties-name-input"]').clear().type('Store 1 Persisted')
    cy.get('[data-testid="folder-properties-category-select"]').select('Warehouse')
    cy.get('[data-testid="folder-properties-secondary-label-input"]').clear().type('Aisle B')
    cy.get('[data-testid="folder-properties-status-badge-input"]').clear().type('Cold')
    cy.get('[data-testid="folder-properties-save"]').click()

    cy.get('[data-testid="collection-folder-store-1"]').should('have.text', 'Store 1 Persisted')
    cy.get('[data-testid="folder-tree-secondary-store-1"]').should('have.text', 'Aisle B')
    cy.get('[data-testid="folder-tree-badge-store-1"]').should('have.text', 'Cold')

    cy.get('[data-testid="folder-tree-add-root"]').click()
    cy.get('[data-testid="folder-tree-name-input"]').clear().type('Refresh Persisted')
    cy.get('[data-testid="folder-tree-create-submit"]').click()
    cy.get('[data-testid="folder-tree-item-refresh-persisted"]').should('be.visible')

    cy.reload()
    cy.wait('@items')

    cy.get('[data-testid="collection-folder-store-1"]').should('have.text', 'Store 1 Persisted')
    cy.get('[data-testid="folder-tree-secondary-store-1"]').should('have.text', 'Aisle B')
    cy.get('[data-testid="folder-tree-badge-store-1"]').should('have.text', 'Cold')
    cy.get('[data-testid="folder-tree-item-refresh-persisted"]').should('be.visible')

    cy.get('[data-testid="folder-tree-row-actions-store-1"]').click()
    cy.get('[data-testid="folder-tree-row-action-properties-store-1"]').click()
    cy.get('[data-testid="folder-properties-name-input"]').should('have.value', 'Store 1 Persisted')
    cy.get('[data-testid="folder-properties-category-select"]').should('have.value', 'Warehouse')
    cy.get('[data-testid="folder-properties-secondary-label-input"]').should('have.value', 'Aisle B')
    cy.get('[data-testid="folder-properties-status-badge-input"]').should('have.value', 'Cold')
  })

  it('UI-SCREEN-INVENTORY-FOLDER-TREE-014 supports child-drop, insertion reordering, root-drop feedback, and practical handle drag coverage', () => {
    const childTransfer = new DataTransfer()

    cy.get('[data-testid="folder-tree-toggle-warehouses"]').click()
    cy.get('[data-testid="folder-tree-group-warehouses"]').should('be.visible')
    cy.get('[data-testid="folder-tree-item-store-1"]')
      .should('have.attr', 'aria-selected', 'false')
      .and('have.attr', 'data-draggable-row', 'true')
    cy.get('[data-testid="folder-tree-drag-handle-store-1"]')
      .should('be.visible')
      .and('have.attr', 'draggable')
      .and('have.attr', 'title', 'Drag Store 1')
      .and('have.css', 'cursor', 'grab')
      .and(($handle) => {
        expect($handle.outerWidth(), 'practical handle width').to.be.gte(28)
        expect($handle.outerHeight(), 'practical handle height').to.be.gte(28)
      })
    cy.get('[data-testid="folder-tree-inline-actions-store-1"]')
      .should('have.css', 'pointer-events', 'none')

    cy.get('[data-testid="folder-tree-drag-handle-store-1"]').trigger('dragstart', { dataTransfer: childTransfer })
    cy.get('[data-testid="folder-tree-item-warehouses"]')
      .trigger('dragenter', { dataTransfer: childTransfer })
      .trigger('dragover', { dataTransfer: childTransfer })
      .should('have.class', 'bg-primary/20')
      .trigger('drop', { dataTransfer: childTransfer })

    cy.get('[data-testid="folder-tree-group-warehouses"] [data-testid="folder-tree-item-store-1"]')
      .should('exist')
    cy.get('[data-testid="collection-active-context"]').should('contain.text', 'Store 1')

    const invalidTransfer = new DataTransfer()
    cy.get('[data-testid="folder-tree-drag-handle-warehouses"]').trigger('dragstart', { dataTransfer: invalidTransfer })
    cy.get('[data-testid="folder-tree-item-store-1"]')
      .trigger('dragenter', { dataTransfer: invalidTransfer })
      .trigger('dragover', { dataTransfer: invalidTransfer })
      .should('have.attr', 'data-invalid-drop-target', 'true')
      .and('not.have.class', 'bg-primary/20')
      .trigger('drop', { dataTransfer: invalidTransfer })

    cy.get('[data-testid="inventory-folder-tree"] > [role="none"] [data-testid="folder-tree-item-warehouses"]')
      .should('exist')
    cy.get('[data-testid="folder-tree-group-warehouses"] [data-testid="folder-tree-item-store-1"]')
      .should('exist')

    const reorderTransfer = new DataTransfer()
    cy.get('[data-testid="folder-tree-drag-handle-store-1"]').trigger('dragstart', { dataTransfer: reorderTransfer })
    cy.get('[data-testid="folder-tree-drop-before-warehouse-1"]')
      .trigger('dragenter', { dataTransfer: reorderTransfer })
      .trigger('dragover', { dataTransfer: reorderTransfer })
      .should('have.class', 'bg-primary/25')
      .trigger('drop', { dataTransfer: reorderTransfer })

    cy.get('[data-testid="folder-tree-group-warehouses"] > [role="none"]').then(($rows) => {
      const labels = [...$rows].map((row) => row.textContent ?? '')
      expect(labels[0]).to.contain('Store 1')
      expect(labels[1]).to.contain('Warehouse 1')
    })

    const rootTransfer = new DataTransfer()
    cy.get('[data-testid="folder-tree-drag-handle-store-1"]').trigger('dragstart', { dataTransfer: rootTransfer })
    cy.get('[data-testid="folder-tree-root-drop-zone"]')
      .trigger('dragenter', { dataTransfer: rootTransfer })
      .trigger('dragover', { dataTransfer: rootTransfer })
      .should('contain.text', 'Drop here to move folder to the root level')
      .and('have.class', 'bg-primary/10')
      .trigger('drop', { dataTransfer: rootTransfer })

    cy.get('[data-testid="inventory-folder-tree"] > [role="none"] [data-testid="folder-tree-item-store-1"]')
      .first()
      .should('be.visible')
    cy.get('[data-testid="folder-tree-group-warehouses"] [data-testid="folder-tree-item-store-1"]')
      .should('not.exist')
  })

  it('UI-SCREEN-INVENTORY-FOLDER-TREE-016 persists inventory item folder assignment after drag drop and reload', () => {
    const transfer = new DataTransfer()

    cy.get('[data-testid="inventory-item-row-e2e-item-001"]')
      .trigger('dragstart', { dataTransfer: transfer })

    cy.get('[data-testid="folder-tree-item-store-1"]')
      .trigger('dragenter', { dataTransfer: transfer })
      .trigger('dragover', { dataTransfer: transfer })
      .trigger('drop', { dataTransfer: transfer })

    cy.get('[data-testid="folder-tree-item-store-1"]').click()
    cy.get('[data-testid="collection-active-context"]').should('contain.text', 'Store 1')
    cy.get('[data-testid="inventory-item-row-e2e-item-001"]').should('be.visible')
    cy.contains('Tree Item 1').should('be.visible')
    cy.contains('Tree Item 2').should('not.exist')

    cy.reload()
    cy.wait('@items')

    cy.get('[data-testid="folder-tree-item-store-1"]').click()
    cy.get('[data-testid="collection-active-context"]').should('contain.text', 'Store 1')
    cy.get('[data-testid="inventory-item-row-e2e-item-001"]').should('be.visible')
    cy.contains('Tree Item 1').should('be.visible')
    cy.contains('Tree Item 2').should('not.exist')
  })

  it('UI-SCREEN-INVENTORY-FOLDER-TREE-017 applies deterministic A/Z sorting to root-level folders while preserving pinned root context', () => {
    cy.get('[data-testid="folder-tree-sort-root-az"]')
      .should('be.visible')
      .and('have.text', 'A/Z')
      .click()

    cy.get('[data-testid="inventory-folder-tree"]').children('[role="none"]').then(($rows) => {
      const labels = [...$rows].map((row) => {
        const treeItem = row.querySelector('[role="treeitem"] [data-testid^="collection-folder-"]')
        return treeItem?.textContent?.trim() ?? ''
      })

      expect(labels[0]).to.equal('All Items')
      expect(labels[1]).to.equal('Archive A')
      expect(labels[2]).to.equal('Archive B')
      expect(labels[3]).to.equal('Archive C')
      expect(labels).to.deep.equal([...labels].sort((left, right) => {
        if (left === 'All Items') return -1
        if (right === 'All Items') return 1
        return left.localeCompare(right)
      }))
    })

    cy.get('[data-testid="folder-tree-toggle-warehouses"]').click()
    cy.get('[data-testid="folder-tree-group-warehouses"] > [role="none"]').then(($rows) => {
      const labels = [...$rows].map((row) => {
        const treeItem = row.querySelector('[role="treeitem"] [data-testid^="collection-folder-"]')
        return treeItem?.textContent?.trim() ?? ''
      })

      expect(labels).to.deep.equal(['Warehouse 1', 'Warehouse 2', 'Warehouse 3'])
    })
  })
})
