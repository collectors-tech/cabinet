describe('ui-global-search-command', () => {
  const commandInputSelector = 'input[placeholder="Type a command or search..."]'

  function signInToHome() {
    cy.visit('/sign-in?redirect=%2F')
    cy.get('input[name="email"]').clear().type('e2e-command@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
    cy.location('pathname', { timeout: 15000 }).should('eq', '/dashboard')
    cy.contains('button', /search/i).should('be.visible')
  }

  function openCommandPaletteWithShortcut() {
    cy.get('body').click(0, 0).type('{ctrl}k')
  }

  function openCommandPaletteFromSearchButton() {
    cy.contains('button', /search/i).first().click()
  }

  beforeEach(() => {
    cy.clearCookies()
    cy.clearLocalStorage()
    signInToHome()
  })

  it('UI-GLOBAL-SEARCH-COMMAND-001 opens command palette with Ctrl/Cmd+K and focuses input', () => {
    openCommandPaletteWithShortcut()
    cy.get(commandInputSelector).should('be.visible').and('be.focused')
  })

  it('UI-GLOBAL-SEARCH-COMMAND-002 executes navigation command and closes palette', () => {
    openCommandPaletteFromSearchButton()
    cy.get(commandInputSelector).type('Inventory')
    cy.contains('[cmdk-item]', 'Inventory').click()

    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)
    cy.get(commandInputSelector).should('not.exist')
  })

  it('UI-GLOBAL-SEARCH-COMMAND-003 executes theme action command and closes palette', () => {
    openCommandPaletteFromSearchButton()
    cy.get(commandInputSelector).type('Dark')
    cy.contains('[cmdk-item]', 'Dark').click()

    cy.get('html').should('have.class', 'dark')
    cy.get(commandInputSelector).should('not.exist')
  })

  it('UI-GLOBAL-SEARCH-COMMAND-004 surfaces local catalog results and opens filtered Inventory', () => {
    cy.intercept('GET', '/api/search/items*', (req) => {
      expect(req.query.text).to.eq('AFX')
      req.reply({
        statusCode: 200,
        body: {
          items: [
            {
              id: 'item-search-1',
              part_number: 'AFX-123',
              title: 'AFX Mega G+ Camaro',
              brand: 'AFX',
              category: 'Slot Cars',
            },
          ],
        },
      })
    }).as('localSearch')

    openCommandPaletteFromSearchButton()
    cy.get(commandInputSelector).type('AFX')
    cy.wait('@localSearch')

    cy.get('[data-testid="command-local-result-item-search-1"]')
      .should('contain', 'AFX Mega G+ Camaro')
      .and('contain', 'AFX-123')
      .click()

    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)
    cy.location('search').should('contain', 'filter=AFX-123')
  })

  it('UI-GLOBAL-SEARCH-COMMAND-005 makes unresolved barcode lookup actionable from Market Watch', () => {
    cy.intercept('GET', '/api/search/items*', {
      statusCode: 200,
      body: { items: [] },
    }).as('localSearch')
    cy.intercept('GET', '/api/barcodes/0000000000000', {
      statusCode: 200,
      body: { matches: [] },
    }).as('barcodeLookup')

    openCommandPaletteFromSearchButton()
    cy.get(commandInputSelector).type('0000000000000')
    cy.wait(['@localSearch', '@barcodeLookup'])

    cy.get('[data-testid="command-barcode-empty"]')
      .should('contain', 'No local barcode match')
      .and('contain', 'Open Market Watch')
    cy.get('[data-testid="command-barcode-open-scanner"]').click()

    cy.location('pathname', { timeout: 15000 }).should('match', /^\/scanner\/?$/)
    cy.location('search').should('contain', 'barcode=0000000000000')
    cy.get('[data-testid="scanner-new-query-name"]').should(
      'have.value',
      'Barcode 0000000000000'
    )
    cy.get('[data-testid="scanner-new-query-keywords"]').should(
      'have.value',
      '0000000000000'
    )
  })
})
