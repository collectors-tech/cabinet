type InventoryFixtureItem = {
  id: string
  part_number: string
  title: string
  status: string
  category: string
  brand: string
  priority: string
  description: string
  notes: string
  source_urls: string[]
}

describe('inventory editor scroll footer', () => {
  const items: InventoryFixtureItem[] = [
    {
      id: 'item-scroll-alpha',
      part_number: 'CAB-SHOW-022',
      title: 'Constrained Editor Scroll Alpha',
      status: 'active',
      category: 'Cars',
      brand: 'AFX',
      priority: 'medium',
      description:
        'A long display description used to exercise the responsive editor body layout.',
      notes:
        'Private notes that should remain inside the independently scrolling editor body.',
      source_urls: ['https://example.test/source-a', 'https://example.test/source-b'],
    },
    {
      id: 'item-scroll-bravo',
      part_number: 'CAB-SHOW-023',
      title: 'Constrained Editor Scroll Bravo',
      status: 'used',
      category: 'Trains',
      brand: 'Tyco',
      priority: 'medium',
      description: 'Second fixture item for adjacent editor navigation.',
      notes: 'Secondary item notes.',
      source_urls: [],
    },
  ]

  function signIn() {
    cy.viewport(768, 640)
    cy.e2eReset()
    cy.e2eSetSetupState('present')
    cy.intercept('GET', '/api/items', {
      statusCode: 200,
      body: { items },
    }).as('items')
    cy.intercept('GET', '/api/items/item-scroll-alpha/photos', {
      statusCode: 200,
      body: {
        photos: [
          {
            id: 'scroll-alpha-photo',
            filename: 'scroll-alpha.jpg',
            is_primary: true,
          },
        ],
      },
    }).as('alphaPhotos')
    cy.intercept('GET', '/api/items/item-scroll-bravo/photos', {
      statusCode: 200,
      body: { photos: [] },
    }).as('bravoPhotos')
    cy.e2eBootstrap().then(({ profile_id, profile_name }) => {
      cy.useBootstrappedProfile(profile_id, profile_name, { path: '/inventory/' })
    })
    cy.wait('@items')
  }

  it('UI-SCREEN-INVENTORY-ITEMS-013 keeps editor panel footer visible while the body scrolls', () => {
    signIn()

    cy.get('[data-testid="inventory-item-row-item-scroll-alpha"]')
      .scrollIntoView()
      .click({ force: true })
    cy.wait('@alphaPhotos')

    cy.get('[data-testid="inventory-item-editor-panel"]')
      .should('be.visible')
      .and('contain', 'Edit Item')
    cy.get('[data-testid="inventory-item-edit-panel"]').then(($body) => {
      expect($body[0].scrollHeight, 'editor body scrolls').to.be.greaterThan(
        $body[0].clientHeight
      )
    })

    cy.get('[data-testid="inventory-item-editor-previous"]').should('be.visible')
    cy.get('[data-testid="inventory-item-editor-next"]').should('be.visible')
    cy.get('[data-testid="inventory-item-editor-cancel"]').should('be.visible')
    cy.get('[data-testid="inventory-item-save"]').should('be.visible')

    cy.get('[data-testid="inventory-item-edit-panel"]').scrollTo('bottom')
    cy.get('[data-testid="inventory-item-edit-panel"]').then(($body) => {
      expect($body[0].scrollTop, 'editor body scrolled').to.be.greaterThan(0)
    })
    cy.get('[data-testid="inventory-item-save"]').should('be.visible')
    cy.get('[data-testid="inventory-item-editor-cancel"]')
      .should('be.visible')
      .click()
    cy.get('[data-testid="inventory-item-editor-panel"]').should('not.exist')
  })
})
