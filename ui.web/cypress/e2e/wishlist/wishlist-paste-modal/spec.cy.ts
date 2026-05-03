describe('wishlist-paste-modal', () => {
  function openWishlist() {
    cy.e2eReset()
    cy.e2eSetSetupState('present')
    cy.e2eBootstrap().then(({ profile_id, profile_name }) => {
      cy.useBootstrappedProfile(profile_id, profile_name, {
        path: '/wishlist/',
      })
    })
  }

  it('opens pasted content in a modal and creates a wishlist entry', () => {
    const pastedContent = [
      'https://example.test/listings/slot-car-22',
      'Boxed Aurora Thunderjet Mustang',
      'Seller says it has light track wear.',
    ].join('\n')
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
      expect(req.body.title).to.eq('Boxed Aurora Thunderjet Mustang')
      wishlistItems = [
        ...wishlistItems,
        {
          id: 'paste-item-1',
          title: req.body.title,
          part_number: '',
          category: '',
          priority: 'medium',
          status: 'wishlist',
        },
      ]
      req.reply({ statusCode: 201, body: { id: 'paste-item-1' } })
    }).as('createWishlistItem')
    cy.intercept('POST', '/api/wishlist', (req) => {
      expect(req.body.item_id).to.eq('paste-item-1')
      expect(req.body.priority).to.eq('medium')
      expect(req.body.notes).to.eq(pastedContent)
      expect(req.body.purchase_url).to.eq(
        'https://example.test/listings/slot-car-22'
      )
      wishlistEntries = [
        ...wishlistEntries,
        {
          id: 'paste-entry-1',
          item_id: 'paste-item-1',
          priority: req.body.priority,
          notes: req.body.notes,
          target_price: req.body.target_price,
          purchase_url: req.body.purchase_url,
        },
      ]
      req.reply({ statusCode: 201, body: { id: 'paste-entry-1' } })
    }).as('createWishlistEntry')

    openWishlist()
    cy.wait('@wishlistEntries')
    cy.wait('@wishlistItems')

    cy.window().then((win) => {
      Object.defineProperty(win.navigator, 'clipboard', {
        configurable: true,
        value: {
          readText: cy.stub().resolves(pastedContent),
        },
      })
    })

    cy.get('[data-testid="wishlist-paste-action"]').click()
    cy.get('[data-testid="wishlist-paste-dialog"]').should('be.visible')
    cy.get('[data-testid="wishlist-paste-content"]').should(
      'have.value',
      pastedContent
    )
    cy.get('[data-testid="wishlist-paste-title"]').should(
      'have.value',
      'Boxed Aurora Thunderjet Mustang'
    )

    cy.get('[data-testid="wishlist-paste-save"]').click()
    cy.wait('@createWishlistItem')
    cy.wait('@createWishlistEntry')
    cy.wait('@wishlistEntries')
    cy.wait('@wishlistItems')

    cy.contains('Boxed Aurora Thunderjet Mustang').should('be.visible')
  })
})
