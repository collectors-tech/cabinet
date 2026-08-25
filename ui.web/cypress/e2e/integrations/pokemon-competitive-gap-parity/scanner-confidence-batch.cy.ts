describe('integrations/pokemon-competitive-gap-parity/scanner-confidence-batch', () => {
  function signInToScanner() {
    cy.e2eSetSetupState('present')
    cy.e2eBootstrap().then(({ profile_id, profile_name }) => {
      cy.e2eEnsureSignedOut()
      cy.stubLocalServerSession(profile_id)
      cy.useBootstrappedProfile(profile_id, profile_name, { path: '/scanner/' })
    })
    cy.get('[data-testid="market-watch-capture-reveal"]').click()
  }

  beforeEach(() => {
    cy.e2eReset()
    cy.clearCookies()
    cy.clearLocalStorage()
  })

  it('POKEMON-COMP-001 supports confidence-first batch capture with manual override and explicit apply confirmation', () => {
    cy.intercept('GET', '/api/scanner/query-sets', {
      statusCode: 200,
      body: { query_sets: [] },
    }).as('querySets')
    cy.intercept('GET', '/api/scanner/failures', { statusCode: 200, body: { failures: [] } }).as(
      'failures'
    )
    cy.intercept('GET', '/api/provider/health?provider=ebay', {
      statusCode: 200,
      body: { status: 'ok' },
    }).as('providerHealth')

    signInToScanner()
    cy.wait(['@querySets', '@failures', '@providerHealth'])

    cy.get('[data-testid="card-scanner-quick-file-input"]').selectFile(
      {
        contents: Cypress.Buffer.from('pokemon-scan-001'),
        fileName: 'charizard-1999.png',
        mimeType: 'image/png',
      },
      { force: true }
    )

    cy.get('[data-testid="card-scanner-unlinked-cards-list"]').should('be.visible')
    cy.get('[data-testid^="card-scanner-confidence-"]').first().should('contain', 'Confidence:')
    cy.get('[data-testid^="card-scanner-suggestion-"]')
      .first()
      .invoke('text')
      .then((initialSuggestion) => {
        cy.get('[data-testid^="card-scanner-override-"]').first().click()
        cy.get('[data-testid^="card-scanner-override-flag-"]')
          .first()
          .should('contain', 'Manual override active')
        cy.get('[data-testid^="card-scanner-suggestion-"]')
          .first()
          .invoke('text')
          .should((nextSuggestion) => {
            expect(nextSuggestion).not.to.equal(initialSuggestion)
          })
      })

    cy.get('[data-testid^="card-scanner-review-apply-"]').first().click()
    cy.get('[data-testid^="card-scanner-apply-confirmation-"]').first().should('be.visible')
    cy.get('[data-testid="card-scanner-queue"]').should('contain', 'Queued')

    cy.get('[data-testid^="card-scanner-confirm-apply-"]').first().click()
    cy.get('[data-testid="card-scanner-quick-category"]').should('not.contain', 'charizard-1999.png')
    cy.get('[data-testid="card-scanner-queue"]').should('contain', 'Linked')
    cy.get('[data-testid="card-scanner-quick-scan-status"]').should(
      'contain',
      'Inventory mutation applied after explicit confirmation.'
    )
  })
})
