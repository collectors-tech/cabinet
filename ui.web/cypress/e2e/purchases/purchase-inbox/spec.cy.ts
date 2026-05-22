describe('purchases/purchase-inbox', () => {
  it('EBAY-PURCHASE-CAPTURE-006 reviews captured purchases before confirmed mutation actions', () => {
    cy.viewport(1400, 900)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.intercept('POST', '/api/integrations/ebay/purchase-inbox/reviews', {
      statusCode: 200,
      body: {
        source: 'ebay_purchase_capture',
        profile_id: 'e2e-profile-001',
        reviews: [
          {
            status: 'needs_review',
            order: {
              order_id: '20-14595-70928',
              seller_usernames: ['seller-one'],
              order_total: 'AU $8.10',
              currency: 'AUD',
            },
            items: [
              {
                status: 'ready_to_link_or_convert',
                item: {
                  transaction_id: '10080684936020',
                  listing_title: 'Accompanying Flute listing',
                  purchased_identity: 'Accompanying Flute TWM 142',
                  quantity: 4,
                  item_price: 'AU $2.40',
                },
                suggested_actions: [
                  {
                    id: 'link_existing_inventory_item',
                    label: 'Link existing inventory item',
                    scope: 'item',
                    target_key: '10080684936020',
                    requires_confirmation: true,
                  },
                  {
                    id: 'convert_to_inventory_item',
                    label: 'Convert to inventory item',
                    scope: 'item',
                    target_key: '10080684936020',
                    requires_confirmation: true,
                  },
                ],
              },
              {
                status: 'needs_review',
                item: {
                  listing_id: '316046161179',
                  listing_title: 'Mystery purchase',
                },
                missing_fields: ['quantity', 'item_price'],
                suggested_actions: [
                  {
                    id: 'complete_purchase_item_fields',
                    label: 'Complete missing purchase item fields',
                    scope: 'item',
                    target_key: '316046161179:Mystery purchase',
                    requires_confirmation: false,
                  },
                ],
              },
            ],
          },
        ],
      },
    }).as('preparePurchaseReviews')

    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/inbox',
    })
    cy.get('[data-testid="purchase-inbox-empty-state"]').should('be.visible')
    cy.get('[data-testid="purchase-inbox-load-reviews"]').click()
    cy.wait('@preparePurchaseReviews')
    cy.get('[data-testid="purchase-inbox-ready-state"]')
      .should('contain', '20-14595-70928')
      .and('contain', 'Accompanying Flute TWM 142')
      .and('contain', 'Missing: quantity, item_price')
    cy.contains(
      '[data-testid="purchase-inbox-suggested-action"]',
      'Convert to inventory item'
    ).click()
    cy.get('[data-testid="purchase-inbox-confirm-dialog"]').should(
      'be.visible'
    )
    cy.get('[data-testid="purchase-inbox-confirm-action"]').click()
    cy.get('[data-testid="purchase-inbox-action-result"]').should(
      'contain',
      'Queued after confirmation: Convert to inventory item: 10080684936020'
    )
  })

  it('EBAY-PURCHASE-CAPTURE-006 exposes a retryable error state', () => {
    cy.viewport(1400, 900)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.intercept('POST', '/api/integrations/ebay/purchase-inbox/reviews', {
      statusCode: 500,
      body: { error: 'failed' },
    }).as('preparePurchaseReviews')

    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/inbox',
    })
    cy.get('[data-testid="purchase-inbox-load-reviews"]').click()
    cy.wait('@preparePurchaseReviews')
    cy.get('[data-testid="purchase-inbox-error-state"]')
      .should('be.visible')
      .and('contain', 'Purchase Inbox could not load reviews.')
  })
})
