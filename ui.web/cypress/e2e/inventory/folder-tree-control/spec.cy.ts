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

  function resolvePointerCoordinates(
    element: Element,
    mode: 'child' | 'before' | 'after' | 'center' = 'center'
  ) {
    const rect = element.getBoundingClientRect()
    const centerX = rect.left + Math.min(rect.width / 2, 96)
    const edgeInset = Math.max(8, Math.min(16, rect.height * 0.2))

    if (mode === 'before') {
      return { x: centerX, y: rect.top + edgeInset }
    }
    if (mode === 'after') {
      return { x: centerX, y: rect.bottom - edgeInset }
    }
    return { x: centerX, y: rect.top + rect.height / 2 }
  }

  function dispatchPointerDown(win: Window, selector: string, pointerId: number) {
    const element = win.document.querySelector<HTMLElement>(selector)
    expect(element, `${selector} exists`).to.not.equal(null)
    const { x, y } = resolvePointerCoordinates(element as HTMLElement)
    element?.dispatchEvent(
      new win.PointerEvent('pointerdown', {
        bubbles: true,
        cancelable: true,
        button: 0,
        buttons: 1,
        pointerId,
        pointerType: 'mouse',
        isPrimary: true,
        clientX: x,
        clientY: y,
      })
    )
  }

  function dispatchPointerMove(
    win: Window,
    selector: string,
    mode: 'child' | 'before' | 'after' | 'center',
    pointerId: number
  ) {
    const element = win.document.querySelector<HTMLElement>(selector)
    expect(element, `${selector} exists`).to.not.equal(null)
    const { x, y } = resolvePointerCoordinates(element as HTMLElement, mode)
    win.dispatchEvent(
      new win.PointerEvent('pointermove', {
        bubbles: true,
        cancelable: true,
        button: 0,
        buttons: 1,
        pointerId,
        pointerType: 'mouse',
        isPrimary: true,
        clientX: x,
        clientY: y,
      })
    )
  }

  function dispatchPointerUp(
    win: Window,
    selector: string,
    mode: 'child' | 'before' | 'after' | 'center',
    pointerId: number
  ) {
    const element = win.document.querySelector<HTMLElement>(selector)
    expect(element, `${selector} exists`).to.not.equal(null)
    const { x, y } = resolvePointerCoordinates(element as HTMLElement, mode)
    win.dispatchEvent(
      new win.PointerEvent('pointerup', {
        bubbles: true,
        cancelable: true,
        button: 0,
        buttons: 0,
        pointerId,
        pointerType: 'mouse',
        isPrimary: true,
        clientX: x,
        clientY: y,
      })
    )
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
    cy.get('[data-testid="folder-tree-item-top-level-added"]')
      .scrollIntoView()
      .should('be.visible')
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
      .scrollIntoView()
      .should('have.attr', 'aria-expanded', 'true')
    cy.get('[data-testid="folder-tree-group-warehouses"]')
      .scrollIntoView()
      .should('be.visible')
    cy.get('[data-testid="folder-tree-item-warehouse-2"]')
      .scrollIntoView()
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
    cy.get('[data-testid="folder-tree-item-refresh-persisted"]')
      .scrollIntoView()
      .should('be.visible')

    cy.reload()
    cy.wait('@items')

    cy.get('[data-testid="collection-folder-store-1"]').should('have.text', 'Store 1 Persisted')
    cy.get('[data-testid="folder-tree-secondary-store-1"]').should('have.text', 'Aisle B')
    cy.get('[data-testid="folder-tree-badge-store-1"]').should('have.text', 'Cold')
    cy.get('[data-testid="folder-tree-item-refresh-persisted"]')
      .scrollIntoView()
      .should('be.visible')

    cy.get('[data-testid="folder-tree-row-actions-store-1"]').click()
    cy.get('[data-testid="folder-tree-row-action-properties-store-1"]').click()
    cy.get('[data-testid="folder-properties-name-input"]').should('have.value', 'Store 1 Persisted')
    cy.get('[data-testid="folder-properties-category-select"]').should('have.value', 'Warehouse')
    cy.get('[data-testid="folder-properties-secondary-label-input"]').should('have.value', 'Aisle B')
    cy.get('[data-testid="folder-properties-status-badge-input"]').should('have.value', 'Cold')
  })

  it('UI-SCREEN-INVENTORY-FOLDER-TREE-014 supports child-drop, insertion reordering, root-drop feedback, and practical handle drag coverage', () => {
    cy.get('[data-testid="folder-tree-toggle-warehouses"]').click()
    cy.get('[data-testid="folder-tree-group-warehouses"]').should('be.visible')
    cy.get('[data-testid="folder-tree-item-store-1"]')
      .scrollIntoView()
      .should('have.attr', 'aria-selected', 'false')
      .and('have.attr', 'data-draggable-row', 'true')
    cy.get('[data-testid="folder-tree-drag-handle-store-1"]')
      .scrollIntoView()
      .should('be.visible')
      .and('have.attr', 'draggable')
    cy.get('[data-testid="folder-tree-drag-handle-store-1"]')
      .should('have.attr', 'title', 'Drag Store 1')
      .and('have.css', 'cursor', 'grab')
      .and(($handle) => {
        expect($handle.outerWidth(), 'practical handle width').to.be.gte(28)
        expect($handle.outerHeight(), 'practical handle height').to.be.gte(28)
      })
    cy.get('[data-testid="folder-tree-inline-actions-store-1"]')
      .should('have.css', 'pointer-events', 'auto')

    cy.window().then((win) => {
      dispatchPointerDown(win, '[data-testid="folder-tree-drag-handle-store-1"]', 11)
    })
    cy.get('[data-testid="folder-tree-drag-preview"]')
      .should(($preview) => {
        const rect = $preview[0].getBoundingClientRect()
        expect(rect.width, 'drag preview width').to.be.greaterThan(120)
        expect(rect.height, 'drag preview height').to.be.greaterThan(40)
        expect(rect.left, 'drag preview left edge').to.be.gte(0)
        expect(rect.top, 'drag preview top edge').to.be.gte(0)
        expect(rect.right, 'drag preview right edge').to.be.lte(
          $preview[0].ownerDocument.defaultView?.innerWidth ?? rect.right
        )
        expect(rect.bottom, 'drag preview bottom edge').to.be.lte(
          $preview[0].ownerDocument.defaultView?.innerHeight ?? rect.bottom
        )
      })
      .and('contain.text', 'Store 1')
      .and('contain.text', 'Aisle A')
    cy.get('[data-testid="folder-tree-root-drop-zone"]').should('be.visible')
    cy.get('[data-testid="folder-tree-row-shell-warehouses"]').then(($row) => {
      $row[0].scrollIntoView({ block: 'center' })
    })
    cy.window().then((win) => {
      dispatchPointerMove(win, '[data-testid="folder-tree-row-shell-warehouses"]', 'child', 11)
    })
    cy.get('[data-testid="folder-tree-item-warehouses"]').should('have.class', 'bg-primary/20')
    cy.get('[data-testid="folder-tree-drop-hint"]')
      .should('exist')
      .and('contain.text', 'Move into Warehouses')
    cy.window().then((win) => {
      dispatchPointerUp(win, '[data-testid="folder-tree-row-shell-warehouses"]', 'child', 11)
    })
    cy.get('[data-testid="folder-tree-drag-preview"]').should('not.exist')
    cy.get('[data-testid="folder-tree-root-drop-zone"]').should('not.exist')
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
      .and('have.class', 'border')
      .click()

    cy.get('[data-testid="inventory-folder-tree-scroll-region"]').children('[role="none"]').then(($rows) => {
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

  it('UI-SCREEN-INVENTORY-FOLDER-TREE-018 keeps active-row controls in separate lanes and preserves a practical drag handle hit area', () => {
    cy.get('[data-testid="folder-tree-item-watch-list"]').click()

    cy.get('[data-testid="folder-tree-inline-actions-watch-list"]')
      .should('have.css', 'pointer-events', 'auto')

    cy.get('[data-testid="folder-tree-badge-watch-list"]').then(($badge) => {
      const badgeRect = $badge[0].getBoundingClientRect()

      cy.get('[data-testid="folder-tree-inline-actions-watch-list"]').then(($actions) => {
        const actionsRect = $actions[0].getBoundingClientRect()
        expect(actionsRect.left, 'actions should not overlap badge').to.be.greaterThan(
          badgeRect.right + 4
        )

        cy.get('[data-testid="folder-tree-drag-handle-watch-list"]').then(($handle) => {
          const handleRect = $handle[0].getBoundingClientRect()
          expect(handleRect.left, 'handle should not overlap actions').to.be.greaterThan(
            actionsRect.right + 4
          )
          expect(handleRect.width, 'practical handle width').to.be.gte(32)
          expect(handleRect.height, 'practical handle height').to.be.gte(32)

          cy.get('[data-testid="inventory-folder-tree"]').then(($tree) => {
            const treeRect = $tree[0].getBoundingClientRect()
            expect(handleRect.right, 'handle should stay inside the tree viewport').to.be.lessThan(
              treeRect.right - 4
            )

            cy.document().then((doc) => {
              const hit = doc.elementFromPoint(
                handleRect.left + handleRect.width / 2,
                handleRect.top + handleRect.height / 2
              ) as HTMLElement | null

              expect(
                hit?.closest('[data-testid="folder-tree-drag-handle-watch-list"]') ?? null,
                'handle center should be pointer-hit-testable'
              ).to.not.equal(null)
            })
          })
        })
      })
    })
  })

  it('UI-SCREEN-INVENTORY-FOLDER-TREE-019 stretches the folder tree to fill the card content height', () => {
    cy.get('[data-testid="folder-tree-card-content"]').should('be.visible')
    cy.get('[data-testid="folder-tree-toolbar"]').should('be.visible')
    cy.get('[data-testid="inventory-folder-tree"]').should('be.visible')

    cy.get('[data-testid="folder-tree-card-content"]').then(($content) => {
      const contentRect = $content[0].getBoundingClientRect()

      cy.get('[data-testid="folder-tree-toolbar"]').then(($toolbar) => {
        const toolbarRect = $toolbar[0].getBoundingClientRect()

        cy.get('[data-testid="inventory-folder-tree"]').then(($tree) => {
          const treeRect = $tree[0].getBoundingClientRect()
          const availableHeight =
            contentRect.height - (toolbarRect.bottom - contentRect.top) - 8

          expect(
            Math.abs(contentRect.bottom - treeRect.bottom),
            'tree reaches the card bottom'
          ).to.be.lte(6)
          expect(
            treeRect.height,
            'tree uses the remaining card height'
          ).to.be.gte(availableHeight - 6)
        })
      })
    })
  })
})
