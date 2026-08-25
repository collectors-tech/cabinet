describe('scanner/ui-screen-card-scanner', () => {
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
    cy.get('[data-testid="card-scanner-unlinked-cards-list"] [data-testid^="card-scanner-unlinked-item-"]')
      .should('have.length', 2)
      .first()
      .should('contain', 'photo-2.jpg')

    cy.get('[data-testid="card-scanner-mark-linked-photo-2.jpg"]').click()
    cy.get('[data-testid="card-scanner-unlinked-cards-list"] [data-testid^="card-scanner-unlinked-item-"]')
      .should('have.length', 1)
      .first()
      .should('contain', 'photo-1.jpg')

    cy.get('[data-testid="card-scanner-quick-category-view-table"]').click()
    cy.get('[data-testid="card-scanner-unlinked-table"]').within(() => {
      cy.contains('th', 'File').should('be.visible')
      cy.contains('th', 'Grading').should('be.visible')
      cy.contains('th', 'Queued At').should('be.visible')
      cy.contains('th', 'Status').should('be.visible')
      cy.contains('td', 'photo-1.jpg').should('be.visible')
      cy.contains('photo-2.jpg').should('not.exist')
    })
  })

  it('UI-SCREEN-CARD-SCANNER-012 queues manual-entry scans for review before writes', () => {
    signInToScanner()
    cy.wait(['@querySets', '@failures', '@providerHealth'])

    cy.get('[data-testid="card-scanner-manual-entry-queue"]').click()
    cy.get('[data-testid="card-scanner-quick-scan-status"]')
      .should('be.visible')
      .and('contain', 'Manual entry needs a card title before queueing.')
    cy.get('[data-testid="card-scanner-queue"]').should('contain', 'No quick-scan items queued.')

    cy.get('[data-testid="card-scanner-manual-entry-title"]').type('Pikachu Promo 001')
    cy.get('[data-testid="card-scanner-manual-entry-queue"]').click()

    cy.get('[data-testid="card-scanner-quick-scan-status"]')
      .should('be.visible')
      .and('contain', 'Manual entry queued for review: Pikachu Promo 001')
    cy.get('[data-testid="card-scanner-queue"]')
      .should('contain', 'Pikachu Promo 001')
      .and('contain', 'Queued')
      .and('not.contain', 'Linked')
    cy.get('[data-testid="card-scanner-manual-entry-title"]').should('have.value', '')
    cy.get('[data-testid^="card-scanner-suggestion-"]')
      .should('be.visible')
      .and('contain', 'pikachu promo 001 (primary match)')
  })

  it('UI-SCREEN-CARD-SCANNER-011 preserves grading context in candidate review before writes', () => {
    cy.intercept('POST', '/api/scanner/recognition-review/apply', (req) => {
      const candidates = req.body.candidates as Array<{
        item_type?: string
        condition_estimate?: string
        grading_status?: string
      }>
      expect(req.body).to.include({ confirmed: false, target: 'inventory' })
      expect(candidates).to.have.length(3)
      candidates.forEach((candidate) => {
        expect(candidate).to.include({
          item_type: 'Trading Card',
          condition_estimate: 'Near Mint (NM)',
          grading_status: 'ungraded',
        })
      })
      req.reply({
        statusCode: 409,
        body: {
          error: 'scanner_review_confirmation_required',
          confirmation_state: 'required',
          review: {
            top_candidate: req.body.candidates[0],
            alternates: req.body.candidates.slice(1),
            selected_candidate: req.body.candidates[0],
            confidence_label: 'high',
            requires_manual_review: false,
            confirm_before_create: true,
            target: 'inventory',
            media_evidence: { media_id: req.body.candidates[0].media_id },
            provenance: ['quick-scan-upload|ui-upload-preview'],
            manual_override_applied: false,
          },
        },
      })
    }).as('scannerApplyGradingReview')

    signInToScanner()
    cy.wait(['@querySets', '@failures', '@providerHealth'])

    cy.get('[data-testid="card-scanner-quick-file-input"]').selectFile(
      'cypress/fixtures/photo-1.jpg',
      { force: true }
    )

    cy.get('[data-testid^="card-scanner-confidence-"]')
      .should('be.visible')
      .and('contain', 'Confidence:')
    cy.get('[data-testid^="card-scanner-grading-"]')
      .should('be.visible')
      .and('contain', 'Trading Card')
      .and('contain', 'Near Mint (NM)')
      .and('contain', 'ungraded')
    cy.get('[data-testid^="card-scanner-suggestion-"]')
      .should('be.visible')
      .and('contain', 'photo-1 (primary match)')

    cy.get('[data-testid^="card-scanner-review-apply-"]').click()
    cy.wait('@scannerApplyGradingReview')
    cy.get('[data-testid^="card-scanner-review-summary-"]')
      .should('be.visible')
      .and('contain', 'high confidence')
      .and('contain', 'confirm-before-create required')
    cy.get('[data-testid="card-scanner-queue"]').should('contain', 'Queued')
    cy.get('[data-testid="card-scanner-queue"]').should('not.contain', 'Linked')
  })

  it('UI-SCREEN-CARD-SCANNER-009 reviews and confirms scanner apply through the API before marking linked', () => {
    cy.intercept('POST', '/api/scanner/recognition-review/apply', (req) => {
      const selected = req.body.candidates.find((candidate: { override_id?: string }) =>
        Boolean(candidate.override_id)
      ) ?? req.body.candidates[0]
      const review = {
        top_candidate: req.body.candidates[0],
        alternates: req.body.candidates.slice(1),
        selected_candidate: selected,
        confidence_label: selected.confidence >= 0.85 ? 'high' : 'medium',
        requires_manual_review: true,
        confirm_before_create: true,
        target: req.body.target,
        media_evidence: { media_id: selected.media_id, media_url: selected.media_url },
        provenance: ['quick-scan-upload|ui-upload-preview'],
        manual_override_applied: Boolean(selected.override_id),
      }
      if (!req.body.confirmed) {
        req.reply({
          statusCode: 409,
          body: {
            error: 'scanner_review_confirmation_required',
            confirmation_state: 'required',
            review,
          },
        })
        return
      }
      req.reply({
        statusCode: 201,
        body: {
          confirmation_state: 'confirmed',
          target: req.body.target,
          review,
          item: {
            id: 'item-scan-1',
            title: selected.title,
            part_number: selected.id,
            status: req.body.target,
          },
          wishlist_entry: {
            id: 'wish-scan-1',
            item_id: 'item-scan-1',
            owned: false,
          },
        },
      })
    }).as('scannerApply')
    cy.intercept('GET', '/api/items?status=wishlist', {
      statusCode: 200,
      body: { items: [{ id: 'item-scan-1', title: 'photo-1 (alt: foil variant)' }] },
    }).as('wishlistReload')

    signInToScanner()
    cy.wait(['@querySets', '@failures', '@providerHealth'])

    cy.get('[data-testid="card-scanner-quick-file-input"]').selectFile(
      'cypress/fixtures/photo-1.jpg',
      { force: true }
    )
    cy.get('[data-testid^="card-scanner-apply-target-"]').select('wishlist')
    cy.get('[data-testid^="card-scanner-override-"]').click()
    cy.get('[data-testid^="card-scanner-review-apply-"]').click()

    cy.wait('@scannerApply')
      .its('request.body')
      .should('include', { confirmed: false, target: 'wishlist' })
    cy.get('[data-testid^="card-scanner-review-summary-"]')
      .should('be.visible')
      .and('contain', 'medium confidence')
      .and('contain', 'confirm-before-create required')

    cy.get('[data-testid^="card-scanner-confirm-apply-"]').click()
    cy.wait('@scannerApply')
      .its('request.body')
      .should('include', { confirmed: true, target: 'wishlist' })
    cy.wait('@wishlistReload')
    cy.get('[data-testid="card-scanner-queue"]').should('contain', 'Linked')
    cy.get('[data-testid^="card-scanner-apply-result-"]')
      .should('be.visible')
      .and('contain', 'Created wishlist item')
  })

  it('UI-SCREEN-CARD-SCANNER-010 keeps failed reads in manual review without linking the scan', () => {
    cy.intercept('POST', '/api/scanner/recognition-review/apply', {
      statusCode: 422,
      body: {
        error: 'recognition_low_confidence_manual_review_required',
      },
    }).as('scannerApplyFailedRead')

    signInToScanner()
    cy.wait(['@querySets', '@failures', '@providerHealth'])

    cy.get('[data-testid="card-scanner-quick-file-input"]').selectFile(
      'cypress/fixtures/photo-1.jpg',
      { force: true }
    )
    cy.get('[data-testid^="card-scanner-override-"]').click()
    cy.get('[data-testid^="card-scanner-override-flag-"]')
      .should('be.visible')
      .and('contain', 'Manual override active')

    cy.get('[data-testid^="card-scanner-review-apply-"]').click()
    cy.wait('@scannerApplyFailedRead')
      .its('request.body')
      .should((body) => {
        expect(body).to.include({ confirmed: false, target: 'inventory' })
        expect(body.candidates).to.have.length(3)
        expect(body.candidates.some((candidate: { override_id?: string }) => candidate.override_id)).to.eq(
          true
        )
      })

    cy.get('[data-testid="card-scanner-quick-scan-status"]')
      .should('be.visible')
      .and(
        'contain',
        'Review preview failed: recognition_low_confidence_manual_review_required'
      )
    cy.get('[data-testid="card-scanner-queue"]').should('contain', 'Queued')
    cy.get('[data-testid="card-scanner-queue"]').should('not.contain', 'Linked')
    cy.get('[data-testid^="card-scanner-apply-result-"]').should('not.exist')
  })
})
