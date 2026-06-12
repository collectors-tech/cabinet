describe('tasks-workflow-dialogs', () => {
  type WishlistEntry = {
    id: string
    item_id: string
    priority: string
    notes?: string
    target_price?: number
    quantity?: number
    needed_quantity?: number
  }

  type WishlistItem = {
    id: string
    title: string
    part_number: string
    status: string
    category: string
    priority: string
  }

  function signInToWishlistWithState(
    wishlistEntries: WishlistEntry[],
    wishlistItems: WishlistItem[]
  ) {
    cy.intercept('GET', '/api/wishlist', (req) => {
      req.reply({ statusCode: 200, body: { items: wishlistEntries } })
    }).as('wishlistItems')
    cy.intercept('GET', '/api/items?status=wishlist', (req) => {
      req.reply({ statusCode: 200, body: { items: wishlistItems } })
    }).as('catalogItems')
    cy.intercept('GET', '/api/inventory/grading/enums', {
      statusCode: 200,
      body: {
        condition_grades: ['Mint', 'Used'],
        packaging_grades: ['Sealed', 'Loose'],
        item_type_condition_scales: [
          { item_type: 'Trading Cards', conditions: ['Mint', 'Used'] },
          { item_type: 'Slot Cars', conditions: ['Mint', 'Used'] },
        ],
      },
    }).as('gradingEnums')
    cy.intercept('GET', '/api/pricing/stats?item_id=*', {
      statusCode: 200,
      body: { latest: 0 },
    }).as('priceStats')
    cy.intercept('GET', '/api/pricing/trend?item_id=*', {
      statusCode: 200,
      body: { points: [] },
    }).as('priceTrend')
    cy.intercept('GET', '/api/pricing/history?item_id=*', {
      statusCode: 200,
      body: { history: [] },
    }).as('priceHistory')
    cy.intercept('GET', '/api/profiles/*/settings').as('profileSettings')

    cy.e2eReset()
    cy.e2eSetSetupState('present')
    cy.e2eBootstrap().then(({ profile_id, profile_name }) => {
      cy.useBootstrappedProfile(profile_id, profile_name, {
        path: '/wishlist/',
      })
    })
    cy.wait('@wishlistItems')
    cy.wait('@catalogItems')
    cy.wait('@profileSettings')
  }

  it('QA-1094-001 keeps create drawer validation visible for missing and invalid fields', () => {
    const wishlistEntries: WishlistEntry[] = []
    const wishlistItems: WishlistItem[] = []
    signInToWishlistWithState(wishlistEntries, wishlistItems)

    cy.get('[data-testid="wishlist-new-action"]').click()
    cy.get('[data-testid="wishlist-create-panel"]').should('be.visible')
    cy.contains('button', 'Save changes').click()
    cy.contains('Title is required.').should('be.visible')
    cy.get('[data-testid="wishlist-create-panel"]').should('be.visible')

    cy.get('input[name="title"]').type('Validation Draft')
    cy.get('input[name="targetPrice"]').clear().type('-1')
    cy.get('input[name="quantity"]').clear().type('1.5')
    cy.contains('button', 'Save changes').click()
    cy.contains('Target price must be a positive number.').should(
      'be.visible'
    )
    cy.contains('Quantity must be a whole number.').should('be.visible')
    cy.get('[data-testid="wishlist-create-panel"]').should('be.visible')
  })

  it('QA-1094-002 imports wishlist CSV rows through the task import dialog and refreshes the table', () => {
    const wishlistEntries: WishlistEntry[] = []
    const wishlistItems: WishlistItem[] = []
    let createdItemCount = 0

    cy.intercept('POST', '/api/items', (req) => {
      createdItemCount += 1
      const title = String(req.body.title)
      const created: WishlistItem = {
        id: `item-import-${createdItemCount}`,
        title,
        part_number: String(req.body.part_number),
        status: 'wishlist',
        category: String(req.body.category || 'General'),
        priority: String(req.body.priority || 'medium'),
      }
      wishlistItems.push(created)
      req.reply({ statusCode: 201, body: created })
    }).as('createWishlistItem')
    cy.intercept('POST', '/api/wishlist', (req) => {
      const created: WishlistEntry = {
        id: `wish-import-${wishlistEntries.length + 1}`,
        item_id: String(req.body.item_id),
        priority: String(req.body.priority || 'medium'),
        notes: String(req.body.notes || ''),
        target_price: Number(req.body.target_price || 0),
        quantity: Number(req.body.quantity || 0),
        needed_quantity: Number(req.body.needed_quantity || 1),
      }
      wishlistEntries.push(created)
      req.reply({ statusCode: 201, body: created })
    }).as('createWishlistEntry')

    signInToWishlistWithState(wishlistEntries, wishlistItems)

    cy.get('[data-testid="wishlist-import-action"]').click()
    cy.contains('[role="dialog"]', 'Import Wishlist Entries').should(
      'be.visible'
    )
    cy.get('input[type="file"]').selectFile(
      {
        contents: Cypress.Buffer.from(
          [
            'title,part_number,category,priority,notes,target_price',
            'Imported Alpha,IMP-001,Trading Cards,high,Watch the convention price,45.50',
            'Imported Beta,IMP-002,Slot Cars,low,Wait for boxed listing,70',
          ].join('\n')
        ),
        fileName: 'wishlist-import.csv',
        mimeType: 'text/csv',
      },
      { force: true }
    )
    cy.contains('[role="dialog"] button', 'Import').click()

    cy.wait('@createWishlistItem')
      .its('request.body')
      .should('include', {
        title: 'Imported Alpha',
        part_number: 'IMP-001',
        category: 'Trading Cards',
        priority: 'high',
      })
    cy.wait('@createWishlistEntry')
      .its('request.body')
      .should('include', {
        item_id: 'item-import-1',
        priority: 'high',
        notes: 'Watch the convention price',
        target_price: 45.5,
      })
    cy.wait('@createWishlistItem')
      .its('request.body')
      .should('include', {
        title: 'Imported Beta',
        part_number: 'IMP-002',
        category: 'Slot Cars',
        priority: 'low',
      })
    cy.wait('@createWishlistEntry')
      .its('request.body')
      .should('include', {
        item_id: 'item-import-2',
        priority: 'low',
        notes: 'Wait for boxed listing',
        target_price: 70,
      })
    cy.wait('@wishlistItems')
    cy.wait('@catalogItems')

    cy.contains('[role="dialog"]', 'Import Wishlist Entries').should(
      'not.exist'
    )
    cy.contains('Imported Alpha').should('be.visible')
    cy.contains('Imported Beta').should('be.visible')
    cy.contains('Watch the convention price').should('be.visible')
    cy.contains('Wait for boxed listing').should('be.visible')
  })
})
