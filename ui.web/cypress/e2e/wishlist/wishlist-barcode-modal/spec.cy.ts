describe('wishlist-barcode-modal', () => {
  function openWishlist() {
    cy.e2eReset()
    cy.e2eSetSetupState('present')
    cy.e2eBootstrap().then(({ profile_id, profile_name }) => {
      cy.useBootstrappedProfile(profile_id, profile_name, {
        path: '/wishlist/',
      })
    })
  }

  it('accepts a barcode and creates a wishlist item with barcode attached', () => {
    let wishlistEntries: Array<Record<string, unknown>> = []
    let wishlistItems: Array<Record<string, unknown>> = []

    cy.intercept('GET', '/api/wishlist', (req) => {
      req.reply({ statusCode: 200, body: { items: wishlistEntries } })
    }).as('wishlistEntries')
    cy.intercept('GET', '/api/items?status=wishlist', (req) => {
      req.reply({ statusCode: 200, body: { items: wishlistItems } })
    }).as('wishlistItems')
    cy.intercept('GET', '/api/pricing/stats?item_id=*', {
      statusCode: 200,
      body: {},
    }).as('priceStats')
    cy.intercept('GET', '/api/pricing/trend?item_id=*', {
      statusCode: 200,
      body: { points: [] },
    }).as('priceTrend')
    cy.intercept('GET', '/api/pricing/history?item_id=*', {
      statusCode: 200,
      body: { history: [] },
    }).as('priceHistory')
    cy.intercept('POST', '/api/items', (req) => {
      expect(req.body.title).to.eq('Barcode 9312345678901')
      wishlistItems = [
        ...wishlistItems,
        {
          id: 'barcode-item-1',
          title: req.body.title,
          part_number: '9312345678901',
          category: 'Barcode',
          priority: 'medium',
          status: 'wishlist',
        },
      ]
      req.reply({ statusCode: 201, body: { id: 'barcode-item-1' } })
    }).as('createWishlistItem')
    cy.intercept('POST', '/api/wishlist', (req) => {
      expect(req.body.item_id).to.eq('barcode-item-1')
      expect(req.body.notes).to.contain('Created from barcode 9312345678901')
      wishlistEntries = [
        ...wishlistEntries,
        {
          id: 'barcode-entry-1',
          item_id: 'barcode-item-1',
          priority: req.body.priority,
          notes: req.body.notes,
        },
      ]
      req.reply({ statusCode: 201, body: { id: 'barcode-entry-1' } })
    }).as('createWishlistEntry')
    cy.intercept('POST', '/api/items/barcode-item-1/barcodes', (req) => {
      expect(req.body.barcode).to.eq('9312345678901')
      req.reply({
        statusCode: 201,
        body: {
          id: 'barcode-record-1',
          item_id: 'barcode-item-1',
          barcode: '9312345678901',
        },
      })
    }).as('attachBarcode')

    openWishlist()
    cy.wait('@wishlistEntries')
    cy.wait('@wishlistItems')

    cy.window().then((win) => {
      ;(win as unknown as { cabinetWishlistBarcodeTestValue: string })
        .cabinetWishlistBarcodeTestValue = '9312345678901'
    })

    cy.get('[data-testid="wishlist-barcode-action"]').click()
    cy.get('[data-testid="wishlist-barcode-dialog"]').should('be.visible')
    cy.get('[data-testid="wishlist-barcode-scan"]').click()
    cy.get('[data-testid="wishlist-barcode-input"]').should(
      'have.value',
      '9312345678901'
    )
    cy.get('[data-testid="wishlist-barcode-title"]').should(
      'have.value',
      'Barcode 9312345678901'
    )

    cy.get('[data-testid="wishlist-barcode-save"]').click()
    cy.wait('@createWishlistItem')
    cy.wait('@createWishlistEntry')
    cy.wait('@attachBarcode')
    cy.wait('@wishlistEntries')
    cy.wait('@wishlistItems')

    cy.contains('Barcode 9312345678901').should('be.visible')
  })
})
