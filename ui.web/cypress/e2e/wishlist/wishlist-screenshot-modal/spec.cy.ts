describe('wishlist-screenshot-modal', () => {
  function openWishlist() {
    cy.e2eReset()
    cy.e2eSetSetupState('present')
    cy.e2eBootstrap().then(({ profile_id, profile_name }) => {
      cy.useBootstrappedProfile(profile_id, profile_name, {
        path: '/wishlist/',
      })
    })
  }

  it('captures a screenshot preview and creates a wishlist item with photo', () => {
    const screenshotDataUrl =
      'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAFgwJ/lz3+6QAAAABJRU5ErkJggg=='
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
      expect(req.body.title).to.eq('Wishlist screenshot')
      wishlistItems = [
        ...wishlistItems,
        {
          id: 'screenshot-item-1',
          title: req.body.title,
          part_number: '',
          category: 'Screenshot',
          priority: 'medium',
          status: 'wishlist',
        },
      ]
      req.reply({ statusCode: 201, body: { id: 'screenshot-item-1' } })
    }).as('createWishlistItem')
    cy.intercept('POST', '/api/wishlist', (req) => {
      expect(req.body.item_id).to.eq('screenshot-item-1')
      expect(req.body.notes).to.contain('Created from screenshot')
      wishlistEntries = [
        ...wishlistEntries,
        {
          id: 'screenshot-entry-1',
          item_id: 'screenshot-item-1',
          priority: req.body.priority,
          notes: req.body.notes,
        },
      ]
      req.reply({ statusCode: 201, body: { id: 'screenshot-entry-1' } })
    }).as('createWishlistEntry')
    cy.intercept('POST', '/api/items/screenshot-item-1/photos', (req) => {
      expect(req.headers['content-type']).to.contain('multipart/form-data')
      req.reply({
        statusCode: 201,
        body: {
          id: 'screenshot-photo-1',
          item_id: 'screenshot-item-1',
          filename: 'wishlist-screenshot.png',
        },
      })
    }).as('uploadScreenshotPhoto')

    openWishlist()
    cy.wait('@wishlistEntries')
    cy.wait('@wishlistItems')

    cy.window().then((win) => {
      ;(win as unknown as { cabinetWishlistScreenshotTestDataUrl: string })
        .cabinetWishlistScreenshotTestDataUrl = screenshotDataUrl
    })

    cy.get('[data-testid="wishlist-screenshot-action"]').click()
    cy.get('[data-testid="wishlist-screenshot-dialog"]').should('be.visible')
    cy.get('[data-testid="wishlist-screenshot-capture"]').click()
    cy.get('[data-testid="wishlist-screenshot-preview"]')
      .should('be.visible')
      .and('have.attr', 'src', screenshotDataUrl)
    cy.get('[data-testid="wishlist-screenshot-title"]').should(
      'have.value',
      'Wishlist screenshot'
    )

    cy.get('[data-testid="wishlist-screenshot-save"]').click()
    cy.wait('@createWishlistItem')
    cy.wait('@createWishlistEntry')
    cy.wait('@uploadScreenshotPhoto')
    cy.wait('@wishlistEntries')
    cy.wait('@wishlistItems')

    cy.contains('Wishlist screenshot').should('be.visible')
  })

  it('creates a wishlist item with photo when an image is dropped onto wishlist', () => {
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
      expect(req.body.title).to.eq('wishlist-drop')
      expect(req.body).to.include({
        category: 'Image',
        priority: 'medium',
      })
      wishlistItems = [
        ...wishlistItems,
        {
          id: 'wishlist-drop-item-1',
          title: req.body.title,
          part_number: '',
          category: 'Image',
          priority: 'medium',
          status: 'wishlist',
        },
      ]
      req.reply({ statusCode: 201, body: { id: 'wishlist-drop-item-1' } })
    }).as('createDroppedWishlistItem')
    cy.intercept('POST', '/api/wishlist', (req) => {
      expect(req.body.item_id).to.eq('wishlist-drop-item-1')
      expect(req.body.notes).to.contain('Created from dropped image')
      wishlistEntries = [
        ...wishlistEntries,
        {
          id: 'wishlist-drop-entry-1',
          item_id: 'wishlist-drop-item-1',
          priority: req.body.priority,
          notes: req.body.notes,
        },
      ]
      req.reply({ statusCode: 201, body: { id: 'wishlist-drop-entry-1' } })
    }).as('createDroppedWishlistEntry')
    cy.intercept('POST', '/api/items/wishlist-drop-item-1/photos', (req) => {
      expect(req.headers['content-type']).to.contain('multipart/form-data')
      req.reply({
        statusCode: 201,
        body: {
          id: 'wishlist-drop-photo-1',
          item_id: 'wishlist-drop-item-1',
          filename: 'wishlist-drop.png',
        },
      })
    }).as('uploadDroppedWishlistPhoto')

    openWishlist()
    cy.wait('@wishlistEntries')
    cy.wait('@wishlistItems')

    cy.window().then((win) => {
      const dataTransfer = new win.DataTransfer()
      dataTransfer.items.add(
        new win.File(['fake-image'], 'wishlist-drop.png', {
          type: 'image/png',
        })
      )

      cy.get('[data-testid="wishlist-image-drop-zone"]')
        .trigger('dragenter', { dataTransfer, force: true })
        .trigger('dragover', { dataTransfer, force: true })
        .trigger('drop', { dataTransfer, force: true })
    })

    cy.wait('@createDroppedWishlistItem')
    cy.wait('@createDroppedWishlistEntry')
    cy.wait('@uploadDroppedWishlistPhoto')
    cy.wait('@wishlistEntries')
    cy.wait('@wishlistItems')
    cy.contains('wishlist-drop').should('be.visible')
    cy.get('[data-testid="wishlist-image-drop-status"]').should(
      'contain',
      'wishlist-drop'
    )
  })
})
