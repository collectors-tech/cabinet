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

  function signInWithCameraStub(
    getUserMediaImpl: (
      constraints: MediaStreamConstraints
    ) => Promise<{
      getTracks: () => Array<{ kind: string; stop: () => void }>
      getVideoTracks: () => Array<{ kind: string; stop: () => void }>
    }>,
    installImageCapture = true
  ) {
    cy.visit('/sign-in?redirect=%2Finventory%2F', {
      onBeforeLoad(win) {
        const cameraConstraintCalls: MediaStreamConstraints[] = []
        Object.defineProperty(win, '__cameraConstraintCalls', {
          configurable: true,
          value: cameraConstraintCalls,
        })
        Object.defineProperty(win.navigator, 'mediaDevices', {
          configurable: true,
          value: {
            getUserMedia: (constraints: MediaStreamConstraints) => {
              cameraConstraintCalls.push(constraints)
              return getUserMediaImpl(constraints)
            },
          },
        })
        if (installImageCapture) {
          Object.defineProperty(win, 'ImageCapture', {
            configurable: true,
            value: class {
              takePhoto() {
                return Promise.resolve(
                  new win.Blob(['real-camera-frame'], { type: 'image/jpeg' })
                )
              }
            },
          })
        }
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

  function openPhotosModal() {
    cy.get('[data-testid="inventory-photos-action"]').click()
    cy.get('[data-testid="inventory-photos-dialog"]').should('be.visible')
  }

  it('MOBILE-CAMERA-001 supports direct camera capture upload path', () => {
    cy.intercept('POST', '/api/items/item-camera-1/photos', (req) => {
      const contentType = String(req.headers['content-type'] ?? '')
      expect(contentType).to.include('multipart/form-data')

      const bodyText = uploadBodyText(req.body)
      expect(bodyText).to.match(
        /name="file"; filename="camera-capture-\d+\.jpg"/
      )
      expect(bodyText).to.include('Content-Type: image/jpeg')
      expect(bodyText).to.include('real-camera-frame')

      req.reply({
        statusCode: 201,
        body: {
          id: 'photo-camera-1',
          item_id: 'item-camera-1',
          original_path: '/media/originals/photo-camera-1.jpg',
          thumbnail_path: '/media/thumbnails/photo-camera-1.jpg',
          preview_path: '/media/previews/photo-camera-1.jpg',
        },
      })
    }).as('uploadPhoto')

    const videoTrack = { kind: 'video', stop: () => {} }
    signInWithCameraStub(async () => ({
      getTracks: () => [videoTrack],
      getVideoTracks: () => [videoTrack],
    }))

    cy.wait('@items')
    cy.wait('@photos')
    cy.wait('@activeProfile')
    openPhotosModal()
    cy.get('[data-testid="inventory-camera-take-photo"]').click()
    cy.wait('@uploadPhoto')
    cy.window()
      .its('__cameraConstraintCalls.0.video.facingMode.ideal')
      .should('equal', 'environment')
    cy.get('[data-testid="inventory-camera-success"]').should('be.visible')
    cy.get('[data-testid="inventory-photo-upload-input"]')
      .should('have.attr', 'accept', 'image/*')
      .and('have.attr', 'capture', 'environment')
  })

  it('MOBILE-CAMERA-002 shows deterministic fallback when camera access is denied', () => {
    signInWithCameraStub(
      async () => {
        throw new Error('Permission denied')
      },
      false
    )

    cy.wait('@items')
    cy.wait('@photos')
    cy.wait('@activeProfile')
    openPhotosModal()
    cy.get('[data-testid="inventory-camera-take-photo"]').click()
    cy.get('[data-testid="inventory-camera-error"]').should('be.visible')
    cy.get('[data-testid="inventory-photo-upload-input"]').should('be.visible')
  })
})
