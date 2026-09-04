describe('inventory-folder-tree-persistence', () => {
  const folderTreeSettingsKey = 'inventory.folder-tree.v2'
  const itemAssignmentsSettingsKey = 'inventory.folder-item-assignments.v1'
  let holdProfileSettingsReads = false
  let releaseProfileSettingsReads: Array<() => void> = []

  function findNode(
    nodes: Array<{
      id?: string
      name?: string
      category?: string
      secondaryLabel?: string
      statusBadge?: string
      children?: unknown
    }>,
    id: string
  ): {
    id?: string
    name?: string
    category?: string
    secondaryLabel?: string
    statusBadge?: string
    children?: unknown
  } | null {
    for (const node of nodes) {
      if (node.id === id) {
        return node
      }

      if (Array.isArray(node.children)) {
        const childMatch = findNode(
          node.children as Array<{
            id?: string
            name?: string
            category?: string
            secondaryLabel?: string
            statusBadge?: string
            children?: unknown
          }>,
          id
        )
        if (childMatch) {
          return childMatch
        }
      }
    }

    return null
  }

  function expectPersistedFolderTreeAfterMove(attempt = 1): Cypress.Chainable<void> {
    return cy.request('/api/profiles/e2e-profile-001/settings').then((response) => {
      const settings = (response.body.settings ?? {}) as Record<string, string>
      const persistedTree = JSON.parse(settings[folderTreeSettingsKey] ?? '[]') as Array<{
        id?: string
        name?: string
        category?: string
        secondaryLabel?: string
        statusBadge?: string
        children?: unknown
      }>

      const storeOne = findNode(persistedTree, 'store-1')
      const warehouses = findNode(persistedTree, 'warehouses') as {
        children?: Array<{ id?: string }>
      } | null
      const storeMovedToWarehouses =
        Array.isArray(warehouses?.children) &&
        warehouses.children.some((child) => child.id === 'store-1')

      if (!storeMovedToWarehouses && attempt < 12) {
        return cy.wait(250).then(() => expectPersistedFolderTreeAfterMove(attempt + 1))
      }

      expect(storeOne?.name).to.equal('Store 1 Persisted')
      expect(storeOne?.category).to.equal('Warehouse')
      expect(storeOne?.secondaryLabel).to.equal('Aisle B')
      expect(storeOne?.statusBadge).to.equal('Cold')
      expect(findNode(persistedTree, 'refresh-persisted')).to.not.equal(null)
      expect(warehouses?.children).to.satisfy(
        (children?: Array<{ id?: string }>) =>
          Array.isArray(children) && children.some((child) => child.id === 'store-1')
      )
    })
  }

  beforeEach(() => {
    holdProfileSettingsReads = Cypress.currentTest.title.includes(
      'UI-SCREEN-INVENTORY-FOLDER-TREE-016'
    )
    releaseProfileSettingsReads = []

    cy.visit('about:blank')
    cy.clearCookies()
    cy.clearLocalStorage()
    cy.e2eReset()
    cy.e2eBootstrap()
    if (holdProfileSettingsReads) {
      cy.request('PUT', '/api/profiles/e2e-profile-001/settings', {
        settings: {
          [itemAssignmentsSettingsKey]: JSON.stringify({
            'stale-item': 'Store 2',
          }),
        },
      })
    }
    cy.intercept('GET', '/api/profiles/e2e-profile-001/settings', (request) => {
      if (!holdProfileSettingsReads) {
        request.continue()
        return
      }

      return new Cypress.Promise<void>((resolve) => {
        releaseProfileSettingsReads.push(() => {
          request.continue()
          resolve()
        })
      })
    })
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
    cy.intercept('PUT', '/api/profiles/e2e-profile-001/settings', (request) => {
      const settings = request.body?.settings as Record<string, unknown> | undefined
      if (typeof settings?.[itemAssignmentsSettingsKey] === 'string') {
        request.alias = 'saveItemAssignments'
      }
    }).as('saveSettings')
    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', { path: '/inventory/' })
    cy.wait('@items')
  })

  it('UI-SCREEN-INVENTORY-FOLDER-TREE-015 persists folder edits, new folders, and moved folders to profile settings and across refresh', () => {
    cy.get('[data-testid="folder-tree-row-actions-store-1"]').click()
    cy.get('[data-testid="folder-tree-row-action-properties-store-1"]').click()
    cy.get('[data-testid="folder-properties-name-input"]').clear().type('Store 1 Persisted')
    cy.get('[data-testid="folder-properties-category-select"]').select('Warehouse')
    cy.get('[data-testid="folder-properties-secondary-label-input"]').clear().type('Aisle B')
    cy.get('[data-testid="folder-properties-status-badge-input"]').clear().type('Cold')
    cy.get('[data-testid="folder-properties-save"]').click()
    cy.wait('@saveSettings')

    cy.get('[data-testid="collection-folder-store-1"]').should('have.text', 'Store 1 Persisted')

    cy.get('[data-testid="folder-tree-add-root"]').click()
    cy.get('[data-testid="folder-tree-name-input"]').clear().type('Refresh Persisted')
    cy.get('[data-testid="folder-tree-create-submit"]').click()
    cy.wait('@saveSettings')
    cy.get('[data-testid="folder-tree-item-refresh-persisted"]')
      .scrollIntoView()
      .should('be.visible')

    const moveTransfer = new DataTransfer()
    cy.get('[data-testid="folder-tree-drag-handle-store-1"]').trigger('dragstart', {
      dataTransfer: moveTransfer,
    })
    cy.get('[data-testid="folder-tree-item-warehouses"]')
      .trigger('dragenter', { dataTransfer: moveTransfer })
      .trigger('dragover', { dataTransfer: moveTransfer })
      .trigger('drop', { dataTransfer: moveTransfer })
    cy.wait('@saveSettings')

    cy.get('[data-testid="folder-tree-group-warehouses"] [data-testid="folder-tree-item-store-1"]')
      .should('exist')

    expectPersistedFolderTreeAfterMove()

    cy.reload()
    cy.wait('@items')

    cy.get('[data-testid="collection-folder-store-1"]').should('have.text', 'Store 1 Persisted')
    cy.get('[data-testid="folder-tree-secondary-store-1"]').should('have.text', 'Aisle B')
    cy.get('[data-testid="folder-tree-badge-store-1"]').should('have.text', 'Cold')
    cy.get('[data-testid="folder-tree-item-refresh-persisted"]')
      .scrollIntoView()
      .should('be.visible')
    cy.get('[data-testid="folder-tree-group-warehouses"] [data-testid="folder-tree-item-store-1"]')
      .should('exist')
  })

  it('UI-SCREEN-INVENTORY-FOLDER-TREE-016 persists inventory item folder assignment to profile settings and across refresh', () => {
    const transfer = new DataTransfer()

    cy.get('[data-testid="inventory-item-row-e2e-item-001"]').trigger('dragstart', {
      dataTransfer: transfer,
    })

    cy.get('[data-testid="folder-tree-item-store-1"]')
      .trigger('dragenter', { dataTransfer: transfer })
      .trigger('dragover', { dataTransfer: transfer })
      .trigger('drop', { dataTransfer: transfer })

    cy.then(() => {
      expect(
        releaseProfileSettingsReads.length,
        'profile settings hydration reads are held'
      ).to.be.greaterThan(0)
      holdProfileSettingsReads = false
      releaseProfileSettingsReads.splice(0).forEach((release) => release())
    })

    cy.wait('@saveItemAssignments').then((interception) => {
      const settings = interception.request.body?.settings as
        | Record<string, string>
        | undefined
      const savedAssignments = JSON.parse(
        settings?.[itemAssignmentsSettingsKey] ?? '{}'
      ) as Record<string, string>

      expect(savedAssignments['e2e-item-001'], 'assignment save payload').to.equal('Store 1')
    })

    cy.request('/api/profiles/e2e-profile-001/settings').then((response) => {
      const settings = (response.body.settings ?? {}) as Record<string, string>
      const assignments = JSON.parse(settings[itemAssignmentsSettingsKey] ?? '{}') as Record<
        string,
        string
      >

      expect(assignments['e2e-item-001']).to.equal('Store 1')
    })

    cy.get('[data-testid="folder-tree-item-store-1"]').click()
    cy.get('[data-testid="collection-active-context"]').should('contain.text', 'Store 1')
    cy.get('[data-testid="inventory-item-row-e2e-item-001"]')
      .scrollIntoView()
      .should('be.visible')
    cy.contains('Tree Item 1').should('be.visible')
    cy.contains('Tree Item 2').should('not.exist')

    cy.reload()
    cy.wait('@items')

    cy.request('/api/profiles/e2e-profile-001/settings').then((response) => {
      const settings = (response.body.settings ?? {}) as Record<string, string>
      const assignments = JSON.parse(settings[itemAssignmentsSettingsKey] ?? '{}') as Record<
        string,
        string
      >

      expect(assignments['e2e-item-001']).to.equal('Store 1')
    })

    cy.get('[data-testid="folder-tree-item-store-1"]').click()
    cy.get('[data-testid="collection-active-context"]').should('contain.text', 'Store 1')
    cy.get('[data-testid="inventory-item-row-e2e-item-001"]')
      .scrollIntoView()
      .should('be.visible')
    cy.contains('Tree Item 1').should('be.visible')
    cy.contains('Tree Item 2').should('not.exist')
  })
})
