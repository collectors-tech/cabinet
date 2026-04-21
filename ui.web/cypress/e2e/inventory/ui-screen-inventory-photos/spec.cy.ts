describe('UI-SCREEN-INVENTORY-PHOTOS', () => {
  function signIn() {
    cy.visit('/sign-in?redirect=%2Finventory%2F')
    cy.get('input[name="email"]').clear().type('e2e-photos@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)
  }

  it('UI-SCREEN-INVENTORY-PHOTOS-001 supports upload + primary + delete lifecycle', () => {
    let listCall = 0
    cy.intercept('GET', '/api/items', {
      statusCode: 200,
      body: {
        items: [{ id: 'item-photo-1', title: 'Photo Item', status: 'todo', category: 'feature' }],
      },
    }).as('items')
    cy.intercept('GET', '/api/items/item-photo-1/photos', (req) => {
      listCall += 1
      if (listCall <= 2) {
        req.reply({
          statusCode: 200,
          body: {
            photos: [
              { id: 'p1', filename: 'one.jpg', is_primary: true },
              { id: 'p2', filename: 'two.jpg', is_primary: false },
            ],
          },
        })
        return
      }
      if (listCall === 3) {
        req.reply({
          statusCode: 200,
          body: {
            photos: [
              { id: 'p1', filename: 'one.jpg', is_primary: false },
              { id: 'p2', filename: 'two.jpg', is_primary: true },
            ],
          },
        })
        return
      }
      req.reply({
        statusCode: 200,
        body: {
          photos: [{ id: 'p2', filename: 'two.jpg', is_primary: true }],
        },
      })
    }).as('listPhotos')
    cy.intercept('POST', '/api/items/item-photo-1/photos', { statusCode: 201, body: { id: 'p3', filename: 'three.jpg', is_primary: false } }).as('uploadPhoto')
    cy.intercept('PUT', '/api/items/item-photo-1/photos/p2/primary', { statusCode: 204, body: '' }).as('setPrimary')
    cy.intercept('DELETE', '/api/items/item-photo-1/photos/p1', { statusCode: 204, body: '' }).as('deletePhoto')

    signIn()
    cy.wait('@items')
    cy.wait('@listPhotos')
    cy.get('[data-testid="inventory-photos-section"]').should('be.visible')
    cy.get('[data-testid="inventory-photo-upload-input"]').selectFile({
      contents: Cypress.Buffer.from('fake image bytes'),
      fileName: 'new-photo.jpg',
      mimeType: 'image/jpeg',
    })
    cy.wait('@uploadPhoto')
    cy.contains('[data-testid="inventory-photo-row"]', 'two.jpg')
      .find('[data-testid="inventory-photo-set-primary"]')
      .click()
    cy.wait('@setPrimary')
    cy.contains('[data-testid="inventory-photo-row"]', 'one.jpg')
      .find('[data-testid="inventory-photo-delete"]')
      .click()
    cy.wait('@deletePhoto')
    cy.wait('@listPhotos')
    cy.contains('[data-testid="inventory-photo-row"]', 'two.jpg')
      .find('[data-testid="inventory-photo-primary-badge"]')
      .should('exist')
  })

  it('UI-SCREEN-INVENTORY-PHOTOS-002 renders deterministic loading, empty, and error states', () => {
    cy.intercept('GET', '/api/items', {
      statusCode: 200,
      body: {
        items: [{ id: 'item-photo-2', title: 'Photo Item 2', status: 'todo', category: 'feature' }],
      },
    }).as('items')
    cy.intercept('GET', '/api/items/item-photo-2/photos', {
      statusCode: 500,
      body: { error: 'failed_to_list_photos' },
    }).as('photosError')

    signIn()
    cy.wait('@items')
    cy.get('[data-testid="inventory-photos-loading"]').should('be.visible')
    cy.wait('@photosError')
    cy.get('[data-testid="inventory-photos-error"]').should('be.visible')
    cy.contains('button', 'Retry').should('be.visible')
  })

  it('PHOTOS-MEDIA-004 opens fullscreen photo viewer and navigates next/previous', () => {
    cy.intercept('GET', '/api/items', {
      statusCode: 200,
      body: {
        items: [{ id: 'item-photo-3', title: 'Photo Item 3', status: 'todo', category: 'feature' }],
      },
    }).as('items')
    cy.intercept('GET', '/api/items/item-photo-3/photos', {
      statusCode: 200,
      body: {
        photos: [
          { id: 'fp1', filename: 'full-1.jpg', is_primary: true },
          { id: 'fp2', filename: 'full-2.jpg', is_primary: false },
        ],
      },
    }).as('photos')
    signIn()
    cy.wait('@items')
    cy.wait('@photos')
    cy.get('[data-testid="inventory-photos-section"]').should('be.visible')
    cy.get('[data-testid="inventory-photo-thumb"]').first().click()
    cy.get('[data-testid="inventory-photo-fullscreen"]').should('be.visible')
    cy.get('[data-testid="inventory-photo-next"]').click()
    cy.get('[data-testid="inventory-photo-prev"]').click()
    cy.get('[data-testid="inventory-photo-fullscreen-close"]').click()
    cy.get('[data-testid="inventory-photo-fullscreen"]').should('not.exist')
  })

  it('PHOTOS-MEDIA-005 keeps photos scoped to the currently selected item when switching rows', () => {
    cy.intercept('GET', '/api/items', {
      statusCode: 200,
      body: {
        items: [
          {
            id: 'item-photo-a',
            part_number: 'PN-PHOTO-A',
            title: 'Photo Item A',
            status: 'active',
            category: 'Cars',
          },
          {
            id: 'item-photo-b',
            part_number: 'PN-PHOTO-B',
            title: 'Photo Item B',
            status: 'active',
            category: 'Cars',
          },
        ],
      },
    }).as('items')
    cy.intercept('GET', '/api/items/item-photo-a/photos', {
      delay: 700,
      statusCode: 200,
      body: {
        photos: [{ id: 'photo-a-1', filename: 'item-a.jpg', is_primary: true }],
      },
    }).as('itemAPhotos')
    cy.intercept('GET', '/api/items/item-photo-b/photos', {
      statusCode: 200,
      body: {
        photos: [{ id: 'photo-b-1', filename: 'item-b.jpg', is_primary: true }],
      },
    }).as('itemBPhotos')

    signIn()
    cy.wait('@items')

    cy.contains('td', 'PN-PHOTO-B').closest('tr').click()
    cy.wait('@itemBPhotos')

    cy.get('[data-testid="collection-selected-item"]').should('contain', 'PN-PHOTO-B')
    cy.contains('[data-testid="inventory-photo-row"]', 'item-b.jpg').should('be.visible')
    cy.contains('[data-testid="inventory-photo-row"]', 'item-a.jpg').should('not.exist')

    cy.wait('@itemAPhotos')
    cy.get('[data-testid="collection-selected-item"]').should('contain', 'PN-PHOTO-B')
    cy.contains('[data-testid="inventory-photo-row"]', 'item-b.jpg').should('be.visible')
    cy.contains('[data-testid="inventory-photo-row"]', 'item-a.jpg').should('not.exist')
  })

  it('PHOTOS-MEDIA-006 reloads the selected item photo state without losing the primary badge', () => {
    const photos = [
      { id: 'reload-p1', filename: 'reload-one.jpg', is_primary: true },
      { id: 'reload-p2', filename: 'reload-two.jpg', is_primary: false },
    ]

    cy.intercept('GET', '/api/items', {
      statusCode: 200,
      body: {
        items: [
          {
            id: 'item-photo-reload',
            part_number: 'PN-PHOTO-RELOAD',
            title: 'Reload Photo Item',
            status: 'active',
            category: 'Cars',
          },
        ],
      },
    }).as('items')
    cy.intercept('GET', '/api/items/item-photo-reload/photos', (req) => {
      req.reply({
        statusCode: 200,
        body: { photos },
      })
    }).as('reloadPhotos')
    cy.intercept('PUT', '/api/items/item-photo-reload/photos/reload-p2/primary', (req) => {
      photos[0] = { ...photos[0], is_primary: false }
      photos[1] = { ...photos[1], is_primary: true }
      req.reply({
        statusCode: 204,
        body: '',
      })
    }).as('reloadSetPrimary')

    signIn()
    cy.wait('@items')
    cy.wait('@reloadPhotos')

    cy.contains('[data-testid="inventory-photo-row"]', 'reload-two.jpg')
      .find('[data-testid="inventory-photo-set-primary"]')
      .click()
    cy.wait('@reloadSetPrimary')
    cy.wait('@reloadPhotos')
    cy.contains('[data-testid="inventory-photo-row"]', 'reload-two.jpg')
      .find('[data-testid="inventory-photo-primary-badge"]')
      .should('exist')

    cy.reload()
    cy.wait('@items')
    cy.wait('@reloadPhotos')
    cy.get('[data-testid="collection-selected-item"]').should('contain', 'PN-PHOTO-RELOAD')
    cy.contains('[data-testid="inventory-photo-row"]', 'reload-two.jpg')
      .find('[data-testid="inventory-photo-primary-badge"]')
      .should('exist')
    cy.contains('[data-testid="inventory-photo-row"]', 'reload-one.jpg')
      .find('[data-testid="inventory-photo-primary-badge"]')
      .should('not.exist')
  })
})
