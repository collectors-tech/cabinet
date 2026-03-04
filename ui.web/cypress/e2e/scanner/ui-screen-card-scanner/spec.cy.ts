describe('scanner/ui-screen-card-scanner', () => {
  function signInToScanner() {
    cy.visit('/sign-in?redirect=%2Fscanner%2F')
    cy.get('input[name="email"]').clear().type('e2e-card-scanner@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/scanner\/?$/)
  }

  beforeEach(() => {
    cy.clearCookies()
    cy.clearLocalStorage()
    cy.intercept('GET', '/api/scanner/query-sets', { statusCode: 200, body: { query_sets: [] } }).as(
      'querySets'
    )
    cy.intercept('GET', '/api/scanner/failures', { statusCode: 200, body: { failures: [] } }).as(
      'failures'
    )
    cy.intercept('GET', '/api/provider/health?provider=ebay', {
      statusCode: 200,
      body: { status: 'ok' },
    }).as('providerHealth')
  })

  it('UI-SCREEN-CARD-SCANNER-006 provides quick-scan action for mobile and desktop with deterministic intake feedback', () => {
    cy.viewport(390, 844)
    signInToScanner()
    cy.wait(['@querySets', '@failures', '@providerHealth'])

    cy.get('[data-testid="card-scanner-quick-scan"]').should('be.visible').click()
    cy.get('[data-testid="card-scanner-quick-scan-status"]')
      .should('be.visible')
      .and('contain', 'Mobile quick capture ready')

    cy.viewport(1280, 800)
    cy.get('[data-testid="card-scanner-quick-scan"]').click()
    cy.get('[data-testid="card-scanner-quick-scan-status"]')
      .should('be.visible')
      .and('contain', 'Desktop quick capture ready')

    cy.get('[data-testid="card-scanner-quick-file-input"]').selectFile(
      'cypress/fixtures/photo-1.jpg',
      { force: true }
    )

    cy.get('[data-testid="card-scanner-queue"]')
      .should('be.visible')
      .and('contain', 'photo-1.jpg')
      .and('contain', 'Queued')
  })

  it('UI-SCREEN-CARD-SCANNER-005 shows recent unlinked scans in cards/table with deterministic newest-first ordering', () => {
    signInToScanner()
    cy.wait(['@querySets', '@failures', '@providerHealth'])

    cy.get('[data-testid="card-scanner-quick-file-input"]').selectFile(
      'cypress/fixtures/photo-1.jpg',
      { force: true }
    )
    cy.wait(100)
    cy.get('[data-testid="card-scanner-quick-file-input"]').selectFile(
      'cypress/fixtures/photo-2.jpg',
      { force: true }
    )

    cy.get('[data-testid="card-scanner-quick-category"]').should('be.visible')
    cy.get('[data-testid="card-scanner-quick-category-view-cards"]').click()
    cy.get('[data-testid="card-scanner-unlinked-cards-list"] [data-testid^=\"card-scanner-unlinked-item-\"]')
      .should('have.length', 2)
      .first()
      .should('contain', 'photo-2.jpg')

    cy.get('[data-testid="card-scanner-mark-linked-photo-2.jpg"]').click()
    cy.get('[data-testid="card-scanner-unlinked-cards-list"] [data-testid^=\"card-scanner-unlinked-item-\"]')
      .should('have.length', 1)
      .first()
      .should('contain', 'photo-1.jpg')

    cy.get('[data-testid="card-scanner-quick-category-view-table"]').click()
    cy.get('[data-testid="card-scanner-unlinked-table"]').within(() => {
      cy.contains('th', 'File').should('be.visible')
      cy.contains('th', 'Queued At').should('be.visible')
      cy.contains('th', 'Status').should('be.visible')
      cy.contains('td', 'photo-1.jpg').should('be.visible')
      cy.contains('photo-2.jpg').should('not.exist')
    })
  })
})
