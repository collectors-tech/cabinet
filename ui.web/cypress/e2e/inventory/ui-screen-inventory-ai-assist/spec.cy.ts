describe('UI-SCREEN-INVENTORY-AI-ASSIST', () => {
  function signIn() {
    cy.visit('/sign-in?redirect=%2Finventory%2F')
    cy.get('input[name="email"]').clear().type('e2e-ai-assist@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)
  }

  beforeEach(() => {
    cy.intercept('GET', '/api/items', {
      statusCode: 200,
      body: {
        items: [{ id: 'item-ai-1', title: 'AI Item', status: 'todo', category: 'feature' }],
      },
    }).as('items')
    cy.intercept('GET', '/api/items/item-ai-1/photos', {
      statusCode: 200,
      body: { photos: [] },
    }).as('photos')
    cy.intercept('GET', '/api/profiles/active', {
      statusCode: 200,
      body: { id: 'profile-ai-1', name: 'AI Profile' },
    }).as('activeProfile')
  })

  it('UI-SCREEN-INVENTORY-AI-ASSIST-001 supports title and photo suggestion workflows', () => {
    cy.intercept('POST', '/api/ai/suggest/title', {
      statusCode: 200,
      body: {
        part_number: 'AFX-22073',
        brand: 'AFX',
        title: '1970 Chevrolet Camaro Wildfire',
        confidence: 0.97,
        requires_confirmation: true,
      },
    }).as('suggestTitle')
    cy.intercept('POST', '/api/ai/suggest/photo', {
      statusCode: 200,
      body: {
        part_number: 'AFX-22073',
        brand: 'AFX',
        title: 'Photo Match Camaro Wildfire',
        confidence: 0.93,
        requires_confirmation: true,
      },
    }).as('suggestPhoto')

    signIn()
    cy.wait('@items')
    cy.wait('@photos')
    cy.wait('@activeProfile')

    cy.get('[data-testid="inventory-ai-assist-section"]').should('be.visible')
    cy.get('[data-testid="inventory-ai-title-input"]').type(
      '22073 AFX Mega G Plus 1970 Chevrolet Camaro Wildfire'
    )
    cy.get('[data-testid="inventory-ai-suggest-title"]').click()
    cy.wait('@suggestTitle')
    cy.get('[data-testid="inventory-ai-suggestion"]').should('contain', 'AFX-22073')

    cy.get('[data-testid="inventory-ai-photo-url-input"]').type(
      'https://example.test/images/afx-22073.jpg'
    )
    cy.get('[data-testid="inventory-ai-suggest-photo"]').click()
    cy.wait('@suggestPhoto')
    cy.get('[data-testid="inventory-ai-suggestion"]').should(
      'contain',
      'Photo Match Camaro Wildfire'
    )
  })

  it('UI-SCREEN-INVENTORY-AI-ASSIST-002 enforces confirm-before-apply', () => {
    cy.intercept('POST', '/api/ai/suggest/title', {
      statusCode: 200,
      body: {
        part_number: 'AFX-777',
        brand: 'AFX',
        title: 'Guarded Apply Suggestion',
        confidence: 0.88,
        requires_confirmation: true,
      },
    }).as('suggestTitle')

    signIn()
    cy.wait('@items')
    cy.wait('@photos')
    cy.wait('@activeProfile')

    cy.get('[data-testid="inventory-ai-title-input"]').type('AFX guarded apply sample')
    cy.get('[data-testid="inventory-ai-suggest-title"]').click()
    cy.wait('@suggestTitle')

    cy.get('[data-testid="inventory-ai-apply"]').click()
    cy.get('[data-testid="inventory-ai-confirm-dialog"]').should('be.visible')
    cy.contains('button', 'Cancel').click()
    cy.get('[data-testid="inventory-ai-applied-banner"]').should('not.exist')

    cy.get('[data-testid="inventory-ai-apply"]').click()
    cy.get('[data-testid="inventory-ai-confirm-dialog"]').should('be.visible')
    cy.contains('button', 'Confirm Apply').click()
    cy.get('[data-testid="inventory-ai-applied-banner"]').should('be.visible')
  })

  it('UI-SCREEN-INVENTORY-AI-ASSIST-003 supports deterministic loading and error retry states', () => {
    let attempt = 0
    cy.intercept('POST', '/api/ai/suggest/title', (req) => {
      attempt += 1
      if (attempt === 1) {
        req.reply({ statusCode: 500, body: { error: 'failed_ai_suggest_title' }, delay: 200 })
        return
      }
      req.reply({
        statusCode: 200,
        body: {
          part_number: 'AFX-RET-1',
          brand: 'AFX',
          title: 'Recovered Suggestion',
          confidence: 0.9,
          requires_confirmation: true,
        },
      })
    }).as('suggestTitle')

    signIn()
    cy.wait('@items')
    cy.wait('@photos')
    cy.wait('@activeProfile')

    cy.get('[data-testid="inventory-ai-title-input"]').type('retry suggestion input')
    cy.get('[data-testid="inventory-ai-suggest-title"]').click()
    cy.get('[data-testid="inventory-ai-loading"]').should('be.visible')
    cy.wait('@suggestTitle')
    cy.get('[data-testid="inventory-ai-error"]').should('be.visible')
    cy.get('[data-testid="inventory-ai-retry"]').click()
    cy.wait('@suggestTitle')
    cy.get('[data-testid="inventory-ai-error"]').should('not.exist')
    cy.get('[data-testid="inventory-ai-suggestion"]').should(
      'contain',
      'Recovered Suggestion'
    )
  })
})
