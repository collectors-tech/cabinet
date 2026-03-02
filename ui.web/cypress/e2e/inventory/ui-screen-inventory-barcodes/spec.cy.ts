describe('UI-SCREEN-INVENTORY-BARCODES', () => {
  function signIn() {
    cy.visit('/sign-in?redirect=%2Finventory%2F')
    cy.get('input[name="email"]').clear().type('e2e-barcodes@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)
  }

  it('UI-SCREEN-INVENTORY-BARCODES-001 adds barcode and updates local lookup results', () => {
    cy.intercept('GET', '/api/items', {
      statusCode: 200,
      body: {
        items: [
          {
            id: 'item-barcode-1',
            title: 'Barcode Item',
            status: 'todo',
            category: 'feature',
          },
        ],
      },
    }).as('items')
    cy.intercept('POST', '/api/items/item-barcode-1/barcodes', {
      statusCode: 201,
      body: {
        id: 'barcode-record-1',
        item_id: 'item-barcode-1',
        barcode: '9780201379624',
      },
    }).as('addBarcode')
    cy.intercept('GET', '/api/barcodes/9780201379624', {
      statusCode: 200,
      body: {
        matches: [
          {
            id: 'barcode-record-1',
            item_id: 'item-barcode-1',
            barcode: '9780201379624',
            created_at: '2026-03-03T10:00:00Z',
          },
        ],
      },
    }).as('lookupBarcode')

    signIn()
    cy.wait('@items')
    cy.get('[data-testid="inventory-barcodes-section"]').should('be.visible')
    cy.get('[data-testid="inventory-barcodes-add-input"]').type('9780201379624')
    cy.get('[data-testid="inventory-barcodes-add-button"]').click()
    cy.wait('@addBarcode')
    cy.get('[data-testid="inventory-barcodes-lookup-input"]').clear().type('9780201379624')
    cy.get('[data-testid="inventory-barcodes-lookup-button"]').click()
    cy.wait('@lookupBarcode')
    cy.get('[data-testid="inventory-barcodes-add-success"]').should('contain', 'added')
    cy.get('[data-testid="inventory-barcodes-match-row"]')
      .should('have.length', 1)
      .first()
      .should('contain', 'item-barcode-1')
  })

  it('UI-SCREEN-INVENTORY-BARCODES-002 shows external fallback when no local match exists', () => {
    cy.intercept('GET', '/api/items', {
      statusCode: 200,
      body: {
        items: [
          {
            id: 'item-barcode-2',
            title: 'No Match Item',
            status: 'todo',
            category: 'feature',
          },
        ],
      },
    }).as('items')
    cy.intercept('GET', '/api/barcodes/0000000000000', {
      statusCode: 200,
      body: { matches: [] },
    }).as('lookupNoMatch')

    signIn()
    cy.wait('@items')
    cy.get('[data-testid="inventory-barcodes-lookup-input"]').clear().type('0000000000000')
    cy.get('[data-testid="inventory-barcodes-lookup-button"]').click()
    cy.wait('@lookupNoMatch')
    cy.get('[data-testid="inventory-barcodes-lookup-empty"]').should('be.visible')
    cy.get('[data-testid="inventory-barcodes-external-search-link"]')
      .should('have.attr', 'href')
      .and('include', '/api/barcodes/0000000000000/external-search')
  })

  it('UI-SCREEN-INVENTORY-BARCODES-003 renders deterministic loading and error states with retry', () => {
    cy.intercept('GET', '/api/items', {
      statusCode: 200,
      body: {
        items: [
          {
            id: 'item-barcode-3',
            title: 'Error Item',
            status: 'todo',
            category: 'feature',
          },
        ],
      },
    }).as('items')

    let lookupAttempt = 0
    cy.intercept('GET', '/api/barcodes/9999999999999', (req) => {
      lookupAttempt += 1
      if (lookupAttempt === 1) {
        req.reply({
          delayMs: 1200,
          statusCode: 500,
          body: { error: 'failed_to_lookup_barcode' },
        })
        return
      }
      req.reply({
        statusCode: 200,
        body: {
          matches: [
            {
              id: 'barcode-record-2',
              item_id: 'item-barcode-3',
              barcode: '9999999999999',
              created_at: '2026-03-03T12:00:00Z',
            },
          ],
        },
      })
    }).as('lookupRetry')

    signIn()
    cy.wait('@items')
    cy.get('[data-testid="inventory-barcodes-lookup-input"]').clear().type('9999999999999')
    cy.get('[data-testid="inventory-barcodes-lookup-button"]').click()
    cy.get('[data-testid="inventory-barcodes-lookup-loading"]').should('be.visible')
    cy.wait('@lookupRetry')
    cy.get('[data-testid="inventory-barcodes-lookup-error"]').should('be.visible')
    cy.get('[data-testid="inventory-barcodes-lookup-retry"]').click()
    cy.wait('@lookupRetry')
    cy.get('[data-testid="inventory-barcodes-match-row"]').should('have.length', 1)
  })
})
