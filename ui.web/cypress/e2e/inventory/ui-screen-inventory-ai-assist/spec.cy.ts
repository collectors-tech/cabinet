describe('UI-SCREEN-INVENTORY-PASTE-CREATE', () => {
  function signIn() {
    cy.visit('/sign-in?redirect=%2Finventory%2F')
    cy.get('input[name="email"]').clear().type('e2e-paste-create@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)
  }

  function openCreateDialog() {
    return cy
      .get('[data-testid="inventory-item-editor-dialog"]')
      .should('be.visible')
      .find('[data-testid="inventory-item-create-dialog"]')
      .should('exist')
  }

  beforeEach(() => {
    cy.intercept('GET', '/api/profiles/active', {
      statusCode: 200,
      body: { id: 'profile-paste-create-1', name: 'Paste Create Profile' },
    }).as('activeProfile')
  })

  it('UI-SCREEN-INVENTORY-PASTE-CREATE-001 keeps Inventory title visible and header actions compact', () => {
    cy.intercept('GET', '/api/items', {
      statusCode: 200,
      body: {
        items: [
          {
            id: 'item-header-compact',
            part_number: 'PN-HEADER',
            title: 'Header Compact Item',
            status: 'active',
            category: 'General',
          },
        ],
      },
    }).as('items')

    signIn()
    cy.wait('@items')
    cy.wait('@activeProfile')

    cy.get('[data-testid="inventory-header-title"]')
      .should('be.visible')
      .and('contain', 'Inventory')
    cy.get('[data-testid="inventory-header-title"]').then(($title) => {
      const titleRect = $title[0].getBoundingClientRect()
      cy.contains('button', 'Search').then(($search) => {
        const searchRect = $search[0].getBoundingClientRect()
        expect(titleRect.left).to.be.greaterThan(searchRect.right)
      })
      cy.get('[data-testid="inventory-global-header-actions"]').then(($actions) => {
        const actionsRect = $actions[0].getBoundingClientRect()
        expect(titleRect.right).to.be.lessThan(actionsRect.left)
      })
    })

    cy.get('[data-testid="inventory-ai-assist-section"]').should('not.exist')
    cy.get('[data-testid="inventory-quick-create"]').should('not.exist')
    cy.contains(/Folders:\s*\d+/)
      .closest('[data-slot="card"]')
      .within(() => {
        cy.get('[data-testid="inventory-quick-create-input"]').should('not.exist')
      })

    cy.get('[data-testid="inventory-paste-action"]')
      .should('be.visible')
      .and('have.attr', 'aria-label', 'Paste URL or text into a new item')
      .and('not.contain', 'Paste')
    cy.get('[data-testid="inventory-barcodes-action"]').and('not.contain', 'Barcodes')
    cy.get('[data-testid="inventory-photos-action"]').and('not.contain', 'Photos')
    cy.get('[data-testid="inventory-new-action"]').and('not.contain', 'New')
    cy.get('[data-testid="inventory-create-menu-trigger"]').and('not.contain', 'Create')
  })

  it('UI-SCREEN-INVENTORY-PASTE-CREATE-002 opens create modal from header paste and processes clipboard URL into editable fields', () => {
    const items: Array<{
      id: string
      part_number: string
      title: string
      status: string
      category: string
      brand?: string
      description?: string
    }> = []

    cy.intercept('GET', '/api/items', (req) => {
      req.reply({ statusCode: 200, body: { items } })
    }).as('items')
    cy.intercept('GET', '/api/items/item-url-created/photos', {
      statusCode: 200,
      body: { photos: [] },
    }).as('urlPhotos')
    cy.intercept('POST', '/api/items', (req) => {
      expect(req.body).to.include({
        title: 'example.test afx-22073',
        brand: 'Unknown',
        category: 'General',
      })
      expect(req.body.part_number).to.match(/^URL-EXAMPLE-TEST-LISTINGS-AFX-22073/)
      expect(req.body.description).to.contain(
        'Source link: https://example.test/listings/afx-22073.jpg'
      )
      expect(req.body.description).to.contain('Creation history:')
      expect(req.body.description).to.contain(
        '- Pasted URL: https://example.test/listings/afx-22073.jpg'
      )
      const created = {
        id: 'item-url-created',
        part_number: req.body.part_number,
        title: req.body.title,
        status: 'active',
        category: 'General',
        brand: req.body.brand,
        description: req.body.description,
      }
      items.unshift(created)
      req.reply({ statusCode: 201, body: created })
    }).as('createPastedItem')

    signIn()
    cy.wait('@items')
    cy.wait('@activeProfile')

    cy.window().then((win) => {
      const readText = cy
        .stub()
        .resolves('https://example.test/listings/afx-22073.jpg')
      Object.defineProperty(win.navigator, 'clipboard', {
        value: { readText },
        configurable: true,
      })
    })

    cy.get('[data-testid="inventory-paste-action"]').click()
    openCreateDialog()
    cy.get('[data-testid="inventory-create-paste-input"]').should(
      'have.value',
      'https://example.test/listings/afx-22073.jpg'
    )
    cy.get('[data-testid="inventory-item-title"]').should(
      'have.value',
      'example.test afx-22073'
    )
    cy.get('[data-testid="inventory-item-description"]')
      .should('contain.value', 'Source link:')
      .and('contain.value', 'Creation history:')

    cy.get('[data-testid="inventory-item-create-submit"]').click()
    cy.wait('@createPastedItem')
    cy.wait('@items')
    cy.wait('@urlPhotos')
    cy.contains('example.test afx-22073').should('be.visible')
  })

  it('UI-SCREEN-INVENTORY-PASTE-CREATE-004 opens paste modal when clipboard is unavailable', () => {
    cy.intercept('GET', '/api/items', {
      statusCode: 200,
      body: { items: [] },
    }).as('items')

    signIn()
    cy.wait('@items')
    cy.wait('@activeProfile')

    cy.window().then((win) => {
      Object.defineProperty(win.navigator, 'clipboard', {
        value: undefined,
        configurable: true,
      })
    })

    cy.get('[data-testid="inventory-paste-action"]').click()
    openCreateDialog()
    cy.get('[data-testid="inventory-create-paste-input"]')
      .should('be.visible')
      .and('be.focused')
      .and('have.value', '')
    cy.get('[data-testid="inventory-create-paste-error"]').should('not.exist')
    cy.get('[data-testid="inventory-create-paste-success"]')
      .should('be.visible')
      .and('contain', 'Paste into the field manually')
  })

  it('UI-SCREEN-INVENTORY-PASTE-CREATE-003 processes additional prompts into creation history before save', () => {
    const items: Array<{
      id: string
      part_number: string
      title: string
      status: string
      category: string
      description?: string
    }> = []

    cy.intercept('GET', '/api/items', (req) => {
      req.reply({ statusCode: 200, body: { items } })
    }).as('items')
    cy.intercept('GET', '/api/items/item-text-created/photos', {
      statusCode: 200,
      body: { photos: [] },
    }).as('textPhotos')
    cy.intercept('POST', '/api/items', (req) => {
      expect(req.body).to.include({
        title: 'Rare Aurora slot car rescue lot',
        brand: 'Unknown',
        category: 'General',
      })
      expect(req.body.description).to.contain(
        '- Pasted text: Rare Aurora slot car rescue lot'
      )
      expect(req.body.description).to.contain(
        '- Prompt: Include cracked windscreen note'
      )
      const created = {
        id: 'item-text-created',
        part_number: req.body.part_number,
        title: req.body.title,
        status: 'active',
        category: 'General',
        description: req.body.description,
      }
      items.push(created)
      req.reply({ statusCode: 201, body: created })
    }).as('createTextItem')

    signIn()
    cy.wait('@items')
    cy.wait('@activeProfile')

    cy.get('[data-testid="inventory-new-action"]').click()
    cy.get('[data-testid="inventory-create-paste-process"]').click()
    cy.get('[data-testid="inventory-create-paste-error"]')
      .should('be.visible')
      .and('contain', 'Paste a URL or text')

    cy.get('[data-testid="inventory-create-paste-input"]').type(
      'Rare Aurora slot car rescue lot'
    )
    cy.get('[data-testid="inventory-create-paste-process"]').click()
    cy.get('[data-testid="inventory-item-title"]').should(
      'have.value',
      'Rare Aurora slot car rescue lot'
    )
    cy.get('[data-testid="inventory-create-paste-input"]')
      .clear()
      .type('Include cracked windscreen note')
    cy.get('[data-testid="inventory-create-paste-process"]').click()
    cy.get('[data-testid="inventory-create-source-history"]')
      .should('contain', 'Pasted text: Rare Aurora slot car rescue lot')
      .and('contain', 'Prompt: Include cracked windscreen note')

    cy.get('[data-testid="inventory-item-create-submit"]').click()
    cy.wait('@createTextItem')
    cy.wait('@items')
    cy.wait('@textPhotos')
    cy.contains('Rare Aurora slot car rescue lot').should('be.visible')
  })
})
