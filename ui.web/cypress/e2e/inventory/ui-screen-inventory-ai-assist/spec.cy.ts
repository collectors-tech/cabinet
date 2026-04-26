describe('UI-SCREEN-INVENTORY-QUICK-CREATE', () => {
  function signIn() {
    cy.visit('/sign-in?redirect=%2Finventory%2F')
    cy.get('input[name="email"]').clear().type('e2e-quick-create@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)
  }

  beforeEach(() => {
    cy.intercept('GET', '/api/profiles/active', {
      statusCode: 200,
      body: { id: 'profile-quick-create-1', name: 'Quick Create Profile' },
    }).as('activeProfile')
  })

  it('UI-SCREEN-INVENTORY-QUICK-CREATE-001 creates an item from pasted URL or text on the Collection Browser row', () => {
    const items: Array<{
      id: string
      part_number: string
      title: string
      status: string
      category: string
      brand?: string
      description?: string
    }> = [
      {
        id: 'item-existing-quick',
        part_number: 'PN-EXISTING',
        title: 'Existing Quick Item',
        status: 'active',
        category: 'General',
      },
    ]

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
        description: 'Source URL: https://example.test/listings/afx-22073.jpg',
      })
      expect(req.body.part_number).to.match(/^URL-EXAMPLE-TEST-LISTINGS-AFX-22073/)
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
    }).as('createQuickItem')

    signIn()
    cy.wait('@items')
    cy.wait('@activeProfile')

    cy.get('[data-testid="inventory-ai-assist-section"]').should('not.exist')
    cy.contains('Collection Browser')
      .closest('[data-slot="card"]')
      .within(() => {
        cy.get('[data-testid="inventory-quick-create"]').should('be.visible')
        cy.get('[data-testid="inventory-quick-create-paste"]')
          .should('be.visible')
          .and('contain', 'Paste')
        cy.get('[data-testid="inventory-quick-create-input"]')
          .should('be.visible')
          .type('https://example.test/listings/afx-22073.jpg')
        cy.get('[data-testid="inventory-quick-create-save"]')
          .should('be.visible')
          .and('contain', 'Save')
          .click()
      })

    cy.wait('@createQuickItem')
    cy.wait('@items')
    cy.wait('@urlPhotos')
    cy.contains('example.test afx-22073').should('be.visible')
    cy.get('[data-testid="collection-selected-item"]').should('contain', 'URL-')
    cy.get('[data-testid="inventory-quick-create-input"]').should('have.value', '')
  })

  it('UI-SCREEN-INVENTORY-QUICK-CREATE-002 supports clipboard paste and validates empty saves', () => {
    const items: Array<{
      id: string
      part_number: string
      title: string
      status: string
      category: string
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
        description: 'Rare Aurora slot car rescue lot',
      })
      expect(req.body.part_number).to.match(/^TXT-RARE-AURORA-SLOT-CAR-RESCUE-LOT/)
      const created = {
        id: 'item-text-created',
        part_number: req.body.part_number,
        title: req.body.title,
        status: 'active',
        category: 'General',
      }
      items.push(created)
      req.reply({ statusCode: 201, body: created })
    }).as('createClipboardItem')

    signIn()
    cy.wait('@items')
    cy.wait('@activeProfile')

    cy.get('[data-testid="inventory-quick-create-save"]').click()
    cy.get('[data-testid="inventory-quick-create-error"]')
      .should('be.visible')
      .and('contain', 'Paste a URL or text')

    cy.window().then((win) => {
      const readText = cy.stub().resolves('Rare Aurora slot car rescue lot')
      Object.defineProperty(win.navigator, 'clipboard', {
        value: { readText },
        configurable: true,
      })
    })

    cy.get('[data-testid="inventory-quick-create-paste"]').click()
    cy.get('[data-testid="inventory-quick-create-input"]').should(
      'have.value',
      'Rare Aurora slot car rescue lot'
    )
    cy.get('[data-testid="inventory-quick-create-save"]').click()

    cy.wait('@createClipboardItem')
    cy.wait('@items')
    cy.wait('@textPhotos')
    cy.contains('Rare Aurora slot car rescue lot').should('be.visible')
  })
})
