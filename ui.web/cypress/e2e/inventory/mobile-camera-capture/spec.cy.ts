describe('MOBILE-CAMERA-CAPTURE', () => {
  function uploadBodyText(body: unknown): string {
    if (typeof body === 'string') {
      return body
    }
    if (body instanceof ArrayBuffer) {
      return new TextDecoder().decode(body)
    }
    if (ArrayBuffer.isView(body)) {
      return new TextDecoder().decode(body)
    }
    return JSON.stringify(body)
  }

  function signInToInventory() {
    cy.visit('/sign-in?redirect=%2Finventory%2F')
    cy.get('input[name="email"]').clear().type('e2e-camera@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)
  }

  beforeEach(() => {
    cy.intercept('GET', '/api/profiles/active', {
      statusCode: 200,
      body: { id: 'profile-camera-1', name: 'Camera Profile' },
    }).as('activeProfile')
    cy.intercept('GET', '/api/items', {
      statusCode: 200,
      body: {
        items: [{ id: 'item-camera-1', title: 'Camera Item', status: 'todo', category: 'feature' }],
      },
    }).as('items')
    cy.intercept('GET', '/api/items/item-camera-1/photos', {
      statusCode: 200,
      body: { photos: [] },
    }).as('photos')
  })

  function openPhotosModal() {
    cy.wait('@items')
    cy.wait('@photos')
    cy.wait('@activeProfile')
    cy.get(
      '[data-testid="inventory-item-row-item-camera-1"] [data-testid="inventory-row-photos-action"]'
    ).click()
    cy.get('[data-testid="inventory-photos-dialog"]').should('be.visible')
  }

  it('MOBILE-CAMERA-001 uses a native capture input for Take Photo uploads', () => {
    cy.intercept('POST', '/api/items/item-camera-1/photos', (req) => {
      const contentType = String(req.headers['content-type'] ?? '')
      expect(contentType).to.include('multipart/form-data')

      const bodyText = uploadBodyText(req.body)
      expect(bodyText).to.include(
        'name="file"; filename="mobile-camera-photo.jpg"'
      )
      expect(bodyText).to.include('Content-Type: image/jpeg')
      expect(bodyText).to.include('native-camera-frame')

      req.reply({
        statusCode: 201,
        body: {
          id: 'photo-camera-1',
          item_id: 'item-camera-1',
          filename: 'mobile-camera-photo.jpg',
          is_primary: true,
        },
      })
    }).as('uploadPhoto')

    signInToInventory()
    openPhotosModal()

    cy.get('[data-testid="inventory-camera-take-photo"]')
      .should('have.attr', 'aria-controls', 'inventory-photo-capture-input')
      .and('not.be.disabled')
    cy.get('[data-testid="inventory-photo-capture-input"]')
      .should('have.attr', 'accept', 'image/*')
      .and('have.attr', 'capture', 'environment')
      .selectFile({
        contents: Cypress.Buffer.from('native-camera-frame'),
        fileName: 'mobile-camera-photo.jpg',
        mimeType: 'image/jpeg',
      }, { force: true })

    cy.wait('@uploadPhoto')
  })

  it('MOBILE-CAMERA-002 keeps desktop upload available without camera capture', () => {
    cy.intercept('POST', '/api/items/item-camera-1/photos', {
      statusCode: 201,
      body: {
        id: 'photo-upload-1',
        item_id: 'item-camera-1',
        filename: 'desktop-upload.jpg',
        is_primary: true,
      },
    }).as('uploadPhoto')

    signInToInventory()
    openPhotosModal()

    cy.get('[data-testid="inventory-photo-upload-input"]')
      .should('have.attr', 'accept', 'image/*')
      .should('not.have.attr', 'capture')
    cy.get('[data-testid="inventory-photo-upload-input"]')
      .selectFile({
        contents: Cypress.Buffer.from('desktop-upload-frame'),
        fileName: 'desktop-upload.jpg',
        mimeType: 'image/jpeg',
      }, { force: true })

    cy.wait('@uploadPhoto')
  })
})
