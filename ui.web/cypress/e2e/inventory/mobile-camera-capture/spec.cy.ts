describe('MOBILE-CAMERA-CAPTURE', () => {
  function signInWithCameraStub(
    getUserMediaImpl: () => Promise<{ getTracks: () => Array<{ stop: () => void }> }>
  ) {
    cy.visit('/sign-in?redirect=%2Finventory%2F', {
      onBeforeLoad(win) {
        Object.defineProperty(win.navigator, 'mediaDevices', {
          configurable: true,
          value: {
            getUserMedia: getUserMediaImpl,
          },
        })
      },
    })
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

  it('MOBILE-CAMERA-001 supports direct camera capture upload path', () => {
    cy.intercept('POST', '/api/items/item-camera-1/photos', {
      statusCode: 201,
      body: {
        id: 'photo-camera-1',
        item_id: 'item-camera-1',
        original_path: '/media/originals/photo-camera-1.jpg',
        thumbnail_path: '/media/thumbnails/photo-camera-1.jpg',
        preview_path: '/media/previews/photo-camera-1.jpg',
      },
    }).as('uploadPhoto')

    signInWithCameraStub(async () => ({
      getTracks: () => [{ stop: () => {} }],
    }))

    cy.wait('@items')
    cy.wait('@photos')
    cy.wait('@activeProfile')
    cy.get('[data-testid="inventory-camera-take-photo"]').click()
    cy.wait('@uploadPhoto')
    cy.get('[data-testid="inventory-camera-success"]').should('be.visible')
  })

  it('MOBILE-CAMERA-002 shows deterministic fallback when camera access is denied', () => {
    signInWithCameraStub(async () => {
      throw new Error('Permission denied')
    })

    cy.wait('@items')
    cy.wait('@photos')
    cy.wait('@activeProfile')
    cy.get('[data-testid="inventory-camera-take-photo"]').click()
    cy.get('[data-testid="inventory-camera-error"]').should('be.visible')
    cy.get('[data-testid="inventory-photo-upload-input"]').should('be.visible')
  })
})
