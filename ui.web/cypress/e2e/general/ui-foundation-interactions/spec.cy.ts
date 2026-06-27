describe('ui-foundation-interactions', () => {
  function signInWithRedirect(path: string, email: string) {
    cy.visit(`/sign-in?redirect=${encodeURIComponent(path)}`)
    cy.get('input[name="email"]').clear().type(email)
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
    const expected = path.endsWith('/') ? path.slice(0, -1) : path
    cy.location('pathname', { timeout: 15000 }).should(
      'match',
      new RegExp(`^${expected}/?$`)
    )
  }

  it('UI-FOUNDATION-INTERACTIONS-003 opens details on single click and edit on double-click for inventory rows', () => {
    cy.intercept('GET', '/api/items', {
      statusCode: 200,
      body: {
        items: [
          {
            id: 'item-interactions-1',
            part_number: 'PN-INTERACT-1',
            title: 'Interaction Inventory Row',
            status: 'todo',
            category: 'feature',
          },
        ],
      },
    }).as('items')
    signInWithRedirect('/inventory/', 'e2e-interactions@example.com')
    cy.wait('@items')

    cy.get('button[aria-label="Switch to rows view"]').click()
    cy.get('tbody tr').eq(0).find('td').eq(1).click()
    cy.get('[data-testid="row-details-modal"]').should('be.visible')
    cy.location('search').should('contain', 'selected=')
    cy.get('body').type('{esc}')
    cy.get('[data-testid="row-details-modal"]').should('not.exist')

    cy.get('tbody tr').eq(0).find('td').eq(1).dblclick()
    cy.get('[data-testid="row-edit-modal"]').should('be.visible')

    cy.contains('Deleting tasks...').should('not.exist')
    cy.contains('Error').should('not.exist')
  })

  it('UI-FOUNDATION-INTERACTIONS-001 keeps row-click detail behavior and thumbnail lightbox behavior distinct', () => {
    cy.intercept('GET', '/api/items', {
      statusCode: 200,
      body: {
        items: [
          {
            id: 'item-interactions-1',
            part_number: 'PN-INTERACT-1',
            title: 'Interaction Inventory Row',
            status: 'todo',
            category: 'feature',
          },
        ],
      },
    }).as('items')
    cy.intercept('GET', '/api/items/item-interactions-1/photos', {
      statusCode: 200,
      body: {
        photos: [
          { id: 'photo-1', filename: 'photo-1.jpg', is_primary: true },
          { id: 'photo-2', filename: 'photo-2.jpg', is_primary: false },
        ],
      },
    }).as('photos')
    cy.intercept('GET', '/api/items/item-interactions-1/photos/*/file*', {
      statusCode: 200,
      headers: { 'content-type': 'image/jpeg' },
      body: 'test-image',
    })

    signInWithRedirect('/inventory/', 'e2e-interactions-001@example.com')
    cy.wait('@items')
    cy.wait('@photos')

    cy.get('button[aria-label="Switch to rows view"]').click()
    cy.get('tbody tr').eq(0).find('td').eq(1).click()
    cy.get('[data-testid="row-details-modal"]').should('be.visible')
    cy.get('body').type('{esc}')

    cy.get('[data-testid="inventory-photo-thumb"]').first().click()
    cy.get('[data-testid="inventory-photo-fullscreen"]').should('be.visible')
    cy.get('[data-testid="inventory-photo-next"]').click()
    cy.get('[data-testid="inventory-photo-prev"]').click()
    cy.get('[data-testid="inventory-photo-fullscreen-close"]').click()
    cy.get('[data-testid="inventory-photo-fullscreen"]').should('not.exist')
  })

  it('UI-FOUNDATION-INTERACTIONS-002 uses explicit checkbox selection for bulk mode', () => {
    cy.intercept('GET', '/api/users*', {
      statusCode: 200,
      body: {
        users: [
          {
            id: 'user-interactions-2',
            firstName: 'Bulk',
            lastName: 'Mode',
            username: 'bulk.mode',
            email: 'bulk.mode@example.com',
            phoneCode: '+61',
            phoneNumber: '123456789',
            role: 'view',
            status: 'active',
          },
        ],
      },
    }).as('users')

    signInWithRedirect('/users/', 'e2e-interactions-002@example.com')
    cy.wait('@users')

    cy.get('tbody tr').first().find('[role="checkbox"]').first().click({ force: true })
    cy.get('[role="toolbar"]').should('be.visible')
    cy.contains('[role="toolbar"]', 'selected').should('be.visible')

    cy.get('tbody tr').eq(0).find('td').eq(1).click()
    cy.get('[role="toolbar"]').should('be.visible')
  })

  it('UI-FOUNDATION-INTERACTIONS-003 keeps selection context stable for users and integrations rows', () => {
    cy.intercept('GET', '/api/users*', {
      statusCode: 200,
      body: {
        users: [
          {
            id: 'user-interactions-1',
            firstName: 'Row',
            lastName: 'User',
            username: 'row.user',
            email: 'row.user@example.com',
            phoneCode: '+61',
            phoneNumber: '000000000',
            role: 'view',
            status: 'active',
          },
        ],
      },
    }).as('users')
    signInWithRedirect('/users/', 'e2e-users@example.com')
    cy.wait('@users')
    cy.get('tbody tr').eq(0).find('td').eq(1).click()
    cy.get('tbody tr').eq(0).should('have.attr', 'data-state', 'selected')
    cy.get('[data-testid="users-row-details-modal"]').should('not.exist')
    cy.location('search').should('contain', 'selected=')
    cy.get('[data-testid="users-view-selected-action"]').click()
    cy.get('[data-testid="users-row-details-modal"]').should('be.visible')
    cy.get('body').type('{esc}')
    cy.get('tbody tr').eq(0).find('td').eq(1).dblclick()
    cy.get('[data-testid="users-row-details-modal"]').should('be.visible')
    cy.get('body').type('{esc}')
    cy.get('[data-testid="users-row-details-modal"]').should('not.exist')

    cy.intercept('GET', '/api/profiles/active', {
      statusCode: 200,
      body: { id: 'profile-interactions-1', name: 'Interactions' },
    }).as('activeProfile')
    cy.intercept('GET', '/api/providers/registry', {
      statusCode: 200,
      body: {
        providers: [
          {
            provider_id: 'ebay',
            display_name: 'eBay',
            base_domain: 'ebay.com',
            integration_mode: 'official_api',
            auth_mode: 'api_key',
            state: 'ready',
            has_token: true,
            capabilities: {
              search: true,
              stock_observation: false,
              pricing: true,
              health: true,
            },
            health: { status: 'ok' },
            last_run: { status: 'success' },
          },
        ],
      },
    }).as('registry')
    cy.intercept('GET', '/api/profiles/*/settings', {
      statusCode: 200,
      body: { settings: { 'integration.ebay.enabled': 'true' } },
    }).as('settings')
    signInWithRedirect('/integrations/', 'e2e-inventory@example.com')
    cy.wait('@activeProfile')
    cy.wait('@registry')
    cy.wait('@settings')
    cy.contains('button', 'Rows').click()
    cy.get('tbody tr').eq(0).find('td').eq(0).click()
    cy.get('[data-testid="integrations-row-details-modal"]').should('be.visible')
    cy.get('body').type('{esc}')
    cy.get('[data-testid="integrations-row-details-modal"]').should('not.exist')
    cy.get('tbody tr').eq(0).find('td').eq(0).dblclick()
    cy.get('[data-testid="integrations-row-edit-modal"]').should('be.visible')
    cy.contains('Provider configuration saved.').should('not.exist')
  })
})
