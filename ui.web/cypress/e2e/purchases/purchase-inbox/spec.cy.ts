describe('purchases/purchase-inbox', () => {
  const groupedPurchaseOrderFixture = () => ({
    order_id: 'EBAY-ORDER-100',
    source: 'ebay',
    seller: 'seller-one',
    tracking: 'TRACK-100',
    status: 'active',
    total_amount: 8.1,
    currency: 'AUD',
    line_item_count: 2,
    received_count: 0,
    unreceived_count: 2,
    created_at: '2026-06-08T10:00:00Z',
    line_items: [
      {
        item_id: 'po-line-1',
        title: 'Accompanying Flute TWM 142',
        quantity: 4,
        amount: 2.4,
        status: 'expected',
        lifecycle_entry_id: 'life-po-1',
        expected_arrival_id: 'arrival-po-1',
      },
      {
        item_id: 'po-line-2',
        title: 'Mystery Pokemon card',
        quantity: 1,
        amount: 5.7,
        status: 'expected',
        lifecycle_entry_id: 'life-po-2',
        expected_arrival_id: 'arrival-po-2',
      },
    ],
  })

  const receivedPurchaseOrderFixture = () => ({
    order_id: 'AMZ-ORDER-200',
    source: 'amazon',
    seller: 'Amazon AU',
    tracking: 'TBA200',
    status: 'received',
    total_amount: 42.5,
    currency: 'AUD',
    line_item_count: 1,
    received_count: 1,
    unreceived_count: 0,
    created_at: '2026-06-09T10:00:00Z',
    line_items: [
      {
        item_id: 'po-line-3',
        title: 'Received order item',
        quantity: 1,
        amount: 42.5,
        status: 'delivered',
        lifecycle_entry_id: 'life-po-3',
        expected_arrival_id: 'arrival-po-3',
      },
    ],
  })

  const secondActivePurchaseOrderFixture = () => ({
    ...groupedPurchaseOrderFixture(),
    order_id: 'EBAY-ORDER-101',
    tracking: 'TRACK-101',
    total_amount: 9.25,
    line_items: [
      {
        item_id: 'po-line-4',
        title: 'Second active purchase item',
        quantity: 1,
        amount: 9.25,
        status: 'expected',
        lifecycle_entry_id: 'life-po-4',
        expected_arrival_id: 'arrival-po-4',
      },
    ],
  })

  // #1487 removed the source-match and captured-review controls from the
  // primary Purchases page. These legacy workflow specs need relocation to the
  // future provenance/detail surface before they can run again.
  const openSourceMatches = () => {
    cy.get('[data-testid="purchases-source-matches-toggle"]').then(
      ($button) => {
        if ($button.attr('aria-expanded') !== 'true') {
          cy.wrap($button).click()
        }
      }
    )
    cy.get('[data-testid="forwarder-package-inbox"]').should('be.visible')
  }

  it('COMMERCE-RECONCILIATION-006 renders first-class Purchases route and add entry point', () => {
    cy.viewport(1400, 900)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.intercept('GET', '/api/commerce/purchase-orders*', {
      statusCode: 200,
      body: { page: 1, page_size: 10, total: 0, total_pages: 0, orders: [] },
    }).as('listPurchaseOrdersEmpty')

    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/purchases',
    })
    cy.wait('@listPurchaseOrdersEmpty')
    cy.location('pathname').should('eq', '/purchases')
    cy.get('[data-testid="purchases-header-title"]').should(
      'contain',
      'Purchases'
    )
    cy.get('[data-testid="purchases-page-icon"]').should('be.visible')
    cy.get('[data-testid="purchases-global-header-actions"]')
      .parents('header')
      .should('exist')
    cy.get('main').find('h1').should('not.exist')
    cy.get('[data-testid="sidebar-nav-link-purchases"]')
      .should('contain', 'Purchases')
      .should('have.attr', 'href', '/purchases')
    cy.get('[data-testid="purchases-table-shell"]')
      .should('be.visible')
      .and('contain', 'Purchase')
      .and('contain', 'Source')
      .and('contain', 'Price')
      .and('contain', 'Purchase date')
      .and('contain', 'Delivery')
      .and('contain', 'Status')
      .and('contain', 'Tracking')
      .and('contain', 'Order link')
      .and('contain', 'Actions')
    cy.get('[data-testid="purchases-table-empty-row"]').should(
      'contain',
      'No persisted purchases loaded'
    )
    cy.get('[data-testid="purchases-table-pagination"]').should(
      'contain',
      '0 persisted orders'
    )
    cy.get('[data-testid="purchase-inbox-empty-state"]').should('not.exist')
    cy.get('[data-testid="purchase-review-tools"]').should('not.exist')
    cy.get('[data-testid="forwarder-package-inbox"]').should('not.exist')
    cy.contains('Purchase Source Matches').should('not.exist')
    cy.get('[data-testid="purchases-add-button"]')
      .should('have.attr', 'aria-label', 'Add purchase')
      .and('have.attr', 'title', 'Add purchase')
      .and('not.contain.text', 'Add')
      .click()
    cy.get('[data-testid="purchases-add-dialog"]').should('be.visible')
    cy.get('[data-testid="purchases-add-tab-new"]').should('contain', 'New')
    cy.get('[data-testid="purchases-add-tab-csv"]').should('contain', 'CSV')
    cy.get('[data-testid="purchases-add-tab-email"]').should('contain', 'Email')
    cy.get('[data-testid="purchases-add-new-title"]').should('be.visible')
    cy.get('[data-testid="purchases-add-csv-input"]').should('not.exist')
    cy.get('[data-testid="purchases-add-tab-csv"]').click()
    cy.get('[data-testid="purchases-add-csv-input"]').should('be.visible')
    cy.get('[data-testid="purchases-add-tab-email"]').click()
    cy.get('[data-testid="purchases-add-email-input"]').should('be.visible')
    cy.contains('button', 'Cancel').click()
    cy.get('[data-testid="purchases-source-matches-toggle"]').should('not.exist')
    cy.get('[data-testid="purchase-inbox-load-reviews"]').should('not.exist')
    cy.get('[data-testid="purchase-review-tools"]').should('not.exist')
    cy.get('[data-testid="forwarder-package-inbox"]').should('not.exist')
    cy.contains('Review source matches').should('not.exist')
    cy.contains('Review captured purchases').should('not.exist')
  })

  it('COMMERCE-RECONCILIATION-013 lists grouped persisted purchase orders with filters search and pagination', () => {
    cy.viewport(1400, 900)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.intercept('GET', '/api/commerce/purchase-orders*', (req) => {
      const url = new URL(req.url)
      const page = url.searchParams.get('page') ?? '1'
      const search = url.searchParams.get('search') ?? ''
      const status = url.searchParams.get('status') ?? ''

      if (search === 'flute') {
        expect(status).to.eq('active')
        req.reply({
          statusCode: 200,
          body: {
            page: 1,
            page_size: 10,
            total: 1,
            total_pages: 1,
            orders: [groupedPurchaseOrderFixture()],
          },
        })
        return
      }

      if (status === 'received') {
        req.reply({
          statusCode: 200,
          body: {
            page: 1,
            page_size: 10,
            total: 1,
            total_pages: 1,
            orders: [receivedPurchaseOrderFixture()],
          },
        })
        return
      }

      req.reply({
        statusCode: 200,
        body:
          page === '2'
            ? {
                page: 2,
                page_size: 10,
                total: 2,
                total_pages: 2,
                orders: [secondActivePurchaseOrderFixture()],
              }
            : {
                page: 1,
                page_size: 10,
                total: 2,
                total_pages: 2,
                orders: [groupedPurchaseOrderFixture()],
              },
      })
    }).as('listPurchaseOrders')

    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/purchases',
    })
    cy.wait('@listPurchaseOrders')
      .its('request.url')
      .should('include', 'status=active')

    cy.get('[data-testid="purchases-table-row"]')
      .should('contain', 'EBAY-ORDER-100')
      .and('contain', 'ebay / seller-one')
      .and('contain', 'AUD 8.10')
      .and('contain', '2 line items')
      .and('contain', 'active')
    cy.get('[data-testid="purchases-line-item-row"]')
      .should('have.length', 2)
      .and('contain', 'Accompanying Flute TWM 142')
      .and('contain', 'Mystery Pokemon card')
      .and('contain', 'Lifecycle life-po-1')
      .and('contain', 'Arrival arrival-po-1')
    cy.get('[data-testid="purchases-table-pagination"]')
      .should('contain', 'Page 1 of 2')
      .and('contain', '2 persisted orders')

    cy.get('[data-testid="purchases-table-search"]').type('flute')
    cy.wait('@listPurchaseOrders')
    cy.wait('@listPurchaseOrders')
    cy.wait('@listPurchaseOrders')
    cy.wait('@listPurchaseOrders')
    cy.wait('@listPurchaseOrders')
      .its('request.url')
      .should('include', 'search=flute')
    cy.get('[data-testid="purchases-table-row"]').should(
      'contain',
      'EBAY-ORDER-100'
    )

    cy.get('[data-testid="purchases-table-search"]').clear()
    cy.wait('@listPurchaseOrders')
    cy.get('[data-testid="purchases-status-filter-received"]').click()
    cy.wait('@listPurchaseOrders')
      .its('request.url')
      .should('include', 'status=received')
    cy.get('[data-testid="purchases-table-row"]')
      .should('contain', 'AMZ-ORDER-200')
      .and('contain', 'received')

    cy.get('[data-testid="purchases-status-filter-active"]').click()
    cy.wait('@listPurchaseOrders')
    cy.get('[data-testid="purchases-page-next"]').click()
    cy.wait('@listPurchaseOrders')
      .its('request.url')
      .should('include', 'page=2')
    cy.get('[data-testid="purchases-table-row"]').should(
      'contain',
      'EBAY-ORDER-101'
    )
    cy.get('[data-testid="purchases-page-previous"]').click()
    cy.wait('@listPurchaseOrders')
      .its('request.url')
      .should('include', 'page=1')
  })

  it('COMMERCE-RECONCILIATION-014 selects order and item detail in split pane', () => {
    cy.viewport(1400, 900)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.intercept('GET', '/api/commerce/purchase-orders*', (req) => {
      const url = new URL(req.url)
      const page = url.searchParams.get('page') ?? '1'
      const status = url.searchParams.get('status') ?? ''

      if (status === 'received') {
        req.reply({
          statusCode: 200,
          body: {
            page: 1,
            page_size: 10,
            total: 1,
            total_pages: 1,
            orders: [receivedPurchaseOrderFixture()],
          },
        })
        return
      }

      req.reply({
        statusCode: 200,
        body:
          page === '2'
            ? {
                page: 2,
                page_size: 10,
                total: 2,
                total_pages: 2,
                orders: [secondActivePurchaseOrderFixture()],
              }
            : {
                page: 1,
                page_size: 10,
                total: 2,
                total_pages: 2,
                orders: [groupedPurchaseOrderFixture()],
              },
      })
    }).as('listPurchaseOrdersForDetail')

    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/purchases',
    })
    cy.wait('@listPurchaseOrdersForDetail')

    cy.get('[data-testid="purchases-split-pane"]').should('be.visible')
    cy.get('[data-testid="purchases-detail-pane"]')
      .should('be.visible')
      .and('contain', 'EBAY-ORDER-100')
      .and('contain', 'seller-one')
      .and('contain', 'AUD 8.10')
      .and('contain', 'TRACK-100')
      .and('contain', '0 received / 2 unreceived')
    cy.get('[data-testid="purchases-order-detail"]')
      .should('contain', 'Receive order')
      .and('contain', 'Reconcile')
      .and('contain', 'Review')
    cy.get('[data-testid="purchases-detail-line-item"]')
      .should('have.length', 2)
      .and('contain', 'Accompanying Flute TWM 142')

    cy.contains(
      '[data-testid="purchases-line-item-row"]',
      'Mystery Pokemon card'
    )
      .find('[data-testid="purchases-line-item-select"]')
      .click()
    cy.get('[data-testid="purchases-detail-pane"]').should(
      'contain',
      'Mystery Pokemon card'
    )
    cy.get('[data-testid="purchases-item-detail"]')
      .should('contain', 'Quantity')
      .and('contain', 'AUD 5.70')
      .and('contain', 'po-line-2')
      .and('contain', 'TRACK-100')
      .and('contain', 'life-po-2')
      .and('contain', 'arrival-po-2')
      .and('contain', 'Receive')
      .and('contain', 'Reconcile')
      .and('contain', 'Review')

    cy.get('[data-testid="purchases-status-filter-received"]').click()
    cy.wait('@listPurchaseOrdersForDetail')
    cy.get('[data-testid="purchases-detail-pane"]')
      .should('contain', 'AMZ-ORDER-200')
      .and('contain', 'Amazon AU')

    cy.get('[data-testid="purchases-status-filter-active"]').click()
    cy.wait('@listPurchaseOrdersForDetail')
    cy.get('[data-testid="purchases-page-next"]').click()
    cy.wait('@listPurchaseOrdersForDetail')
    cy.get('[data-testid="purchases-detail-pane"]').should(
      'contain',
      'EBAY-ORDER-101'
    )
  })

  it.skip('EBAY-PURCHASE-CAPTURE-006 reviews captured purchases before confirmed mutation actions', () => {
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
      path: '/purchases',
    })
    cy.get('[data-testid="purchases-header-title"]').should(
      'contain',
      'Purchases'
    )
    cy.get('[data-testid="purchases-table-shell"]')
      .should('be.visible')
      .and('contain', 'Purchase')
      .and('contain', 'Source')
      .and('contain', 'Tracking')
    cy.get('[data-testid="purchases-add-button"]').click()
    cy.get('[data-testid="purchases-add-dialog"]').should('be.visible')
    cy.get('[data-testid="purchases-add-tab-new"]').should('contain', 'New')
    cy.get('[data-testid="purchases-add-tab-csv"]').should('contain', 'CSV')
    cy.get('[data-testid="purchases-add-tab-email"]').should('contain', 'Email')
    cy.contains('button', 'Cancel').click()
    cy.get('[data-testid="purchase-inbox-empty-state"]').should('not.exist')
    cy.get('[data-testid="purchase-inbox-load-reviews"]').click()
    cy.wait('@preparePurchaseReviews')
    cy.get('[data-testid="purchases-table-row"]')
      .should('contain', 'Accompanying Flute TWM 142')
      .and('contain', 'eBay')
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

  it.skip('COMMERCE-RECONCILIATION-007 filters Purchases rows and marks review state actions', () => {
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
                  seller_username: 'seller-one',
                  purchase_date: '2026-06-08',
                  item_url: 'https://www.ebay.com.au/itm/316046161178',
                },
                suggested_actions: [
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
                  seller_username: 'seller-two',
                },
                missing_fields: ['quantity', 'item_price'],
              },
            ],
          },
        ],
      },
    }).as('preparePurchaseReviews')

    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/purchases',
    })
    cy.get('[data-testid="purchase-inbox-load-reviews"]').click()
    cy.wait('@preparePurchaseReviews')
    cy.get('[data-testid="purchases-filter-result"]').should(
      'contain',
      'Showing 2 of 2 purchases'
    )

    cy.get('[data-testid="purchases-table-search"]').type('Mystery')
    cy.get('[data-testid="purchases-table-row"]')
      .should('have.length', 1)
      .and('contain', 'Mystery purchase')
      .and('contain', 'eBay / seller-two')
    cy.get('[data-testid="purchases-filter-result"]').should(
      'contain',
      'Showing 1 of 2 purchases'
    )

    cy.get('[data-testid="purchases-table-search"]').clear()
    cy.get('[data-testid="purchases-status-filter-ready_to_link_or_convert"]')
      .click()
    cy.get('[data-testid="purchases-table-row"]')
      .should('have.length', 1)
      .and('contain', 'Accompanying Flute TWM 142')
      .and('contain', 'ready to link or convert')
      .and('contain', '2026-06-08')
      .and('contain', 'Pending')
    cy.get('[data-testid="purchases-row-order-link"]')
      .should('have.attr', 'href', 'https://www.ebay.com.au/itm/316046161178')
      .and('contain', 'Open order')
    cy.get('[data-testid="purchases-row-favorite"]')
      .click()
      .should('have.attr', 'aria-pressed', 'true')
    cy.get('[data-testid="purchases-row-arrived"]')
      .click()
      .should('have.attr', 'aria-pressed', 'true')
    cy.get('[data-testid="purchases-table-row"]').should('contain', 'Arrived')
    cy.get('[data-testid="purchases-row-rating"]')
      .click()
      .should('contain', 'Rating 4')

    cy.get('[data-testid="purchases-status-filter-needs_review"]').click()
    cy.get('[data-testid="purchases-table-row"]')
      .should('have.length', 1)
      .and('contain', 'Mystery purchase')
    cy.get('[data-testid="purchases-table-search"]').type('no-match')
    cy.get('[data-testid="purchases-table-filter-empty-row"]').should(
      'contain',
      'No purchases match the current table filters.'
    )
  })

  it('COMMERCE-RECONCILIATION-008/010 creates and persists manual purchase drafts from the add dialog', () => {
    cy.viewport(1400, 900)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.intercept('GET', '/api/commerce/purchase-orders*', {
      statusCode: 200,
      body: { page: 1, page_size: 10, total: 0, total_pages: 0, orders: [] },
    }).as('listPurchaseOrdersEmptyManual')
    cy.intercept('POST', '/api/items', (req) => {
      req.reply({
        statusCode: 201,
        body: {
          id: 'purchase-item-manual-001',
          title: req.body.title,
          part_number: req.body.part_number,
        },
      })
    }).as('createPurchaseDraftItem')
    cy.intercept('POST', '/api/commerce/lifecycle', {
      statusCode: 201,
      body: {
        entry: {
          id: 'life-manual-001',
          state: 'purchase',
          expected_arrival_id: 'arrival-manual-001',
        },
        expected_arrival: {
          id: 'arrival-manual-001',
          lifecycle_entry_id: 'life-manual-001',
          status: 'expected',
        },
      },
    }).as('createPurchaseLifecycle')

    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/purchases',
    })
    cy.wait('@listPurchaseOrdersEmptyManual')
    cy.get('[data-testid="purchases-table-empty-row"]').should(
      'contain',
      'No persisted purchases loaded'
    )
    cy.get('[data-testid="purchases-add-button"]').click()
    cy.get('[data-testid="purchases-add-dialog"]').should('be.visible')
    cy.get('[data-testid="purchases-add-new-save"]').click()
    cy.get('[data-testid="purchases-add-new-error"]').should(
      'contain',
      'Purchase title is required.'
    )

    cy.get('[data-testid="purchases-add-new-title"]').type(
      'Manual Amazon order'
    )
    cy.get('[data-testid="purchases-add-new-source"]').clear().type('Amazon')
    cy.get('[data-testid="purchases-add-new-price"]').type('AU $42.50')
    cy.get('[data-testid="purchases-add-new-tracking"]').type('TBA123456')
    cy.get('[data-testid="purchases-add-new-save"]').click()
    cy.wait('@createPurchaseDraftItem')
      .its('request.body')
      .should('include', {
        title: 'Manual Amazon order',
        brand: 'Amazon',
        category: 'Purchases',
      })
    cy.wait('@createPurchaseLifecycle')
      .its('request.body')
      .should('include', {
        item_id: 'purchase-item-manual-001',
        state: 'purchase',
        source: 'Amazon',
        external_ref: 'TBA123456',
        amount: 42.5,
        currency: 'AUD',
      })
    cy.log('COMMERCE-RECONCILIATION-010 persisted manual draft')
    cy.get('[data-testid="purchases-add-dialog"]').should('not.exist')
    cy.get('[data-testid="purchases-manual-draft-result"]')
      .should(
        'contain',
        'Persisted manual purchase draft for Manual Amazon order'
      )
      .and('contain', 'commerce lifecycle API')
    cy.get('[data-testid="purchases-table-row"]')
      .should('have.length', 1)
      .and('contain', 'Manual Amazon order')
      .and('contain', 'Amazon')
      .and('contain', 'AU $42.50')
      .and('contain', 'manual draft')
      .and('contain', 'TBA123456')
      .and('contain', 'Persisted lifecycle life-man')
    cy.get('[data-testid="purchases-filter-result"]').should(
      'contain',
      'Showing 1 of 1 purchases'
    )

    cy.get('[data-testid="purchases-table-search"]').type('Amazon')
    cy.get('[data-testid="purchases-table-row"]').should(
      'contain',
      'Manual Amazon order'
    )
    cy.get('[data-testid="purchases-status-filter-active"]').click()
    cy.get('[data-testid="purchases-table-row"]')
      .should('have.length', 1)
      .and('contain', 'manual draft')
  })

  it('COMMERCE-RECONCILIATION-009/010 previews confirms and persists CSV and email purchase imports', () => {
    cy.viewport(1400, 900)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    let persistedCount = 0
    cy.intercept('GET', '/api/commerce/purchase-orders*', {
      statusCode: 200,
      body: { page: 1, page_size: 10, total: 0, total_pages: 0, orders: [] },
    }).as('listPurchaseOrdersEmptyImport')
    cy.intercept('POST', '/api/items', (req) => {
      persistedCount += 1
      req.reply({
        statusCode: 201,
        body: {
          id: 'purchase-import-item-' + persistedCount,
          title: req.body.title,
          part_number: req.body.part_number,
        },
      })
    }).as('createPurchaseImportItem')
    cy.intercept('POST', '/api/commerce/lifecycle', (req) => {
      req.reply({
        statusCode: 201,
        body: {
          entry: {
            id: 'life-import-00' + persistedCount,
            state: 'purchase',
            expected_arrival_id: 'arrival-import-00' + persistedCount,
          },
          expected_arrival: {
            id: 'arrival-import-00' + persistedCount,
            lifecycle_entry_id: 'life-import-00' + persistedCount,
            status: 'expected',
          },
        },
      })
    }).as('createPurchaseImportLifecycle')

    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/purchases',
    })
    cy.wait('@listPurchaseOrdersEmptyImport')
    cy.get('[data-testid="purchases-table-empty-row"]').should(
      'contain',
      'No persisted purchases loaded'
    )
    cy.get('[data-testid="purchases-add-button"]').click()
    cy.get('[data-testid="purchases-add-tab-csv"]').click()
    cy.get('[data-testid="purchases-add-csv-input"]')
      .clear()
      .type(
        'source,title,price,currency,purchase_date,seller,channel,tracking,delivery\nAmazon,CSV Pokemon order,42.50,AUD,2026-06-08,Amazon AU,csv,TBA123456,Expected 2026-06-12',
        { delay: 0 }
      )
    cy.get('[data-testid="purchases-add-csv-preview"]').click()
    cy.get('[data-testid="purchases-add-csv-preview-result"]')
      .should('contain', 'Previewing 1 CSV purchase draft')
      .and('contain', 'CSV Pokemon order')
      .and('contain', 'Amazon CSV row 2')
      .and('contain', 'AUD 42.50')
      .and('contain', '2026-06-08')
      .and('contain', 'TBA123456')
    cy.get('[data-testid="purchases-table-row"]').should('not.exist')
    cy.get('[data-testid="purchases-add-csv-confirm"]').click()
    cy.wait('@createPurchaseImportItem')
      .its('request.body')
      .should('include', {
        title: 'CSV Pokemon order',
        brand: 'csv',
        category: 'Purchases',
      })
    cy.wait('@createPurchaseImportLifecycle')
      .its('request.body')
      .should('include', {
        item_id: 'purchase-import-item-1',
        state: 'purchase',
        source: 'Amazon CSV row 2',
        external_ref: 'TBA123456',
        amount: 42.5,
        currency: 'AUD',
      })
    cy.get('[data-testid="purchases-add-dialog"]').should('not.exist')
    cy.get('[data-testid="purchases-manual-draft-result"]')
      .should('contain', 'Confirmed and persisted 1 CSV import draft')
      .and('contain', 'commerce lifecycle API')
    cy.get('[data-testid="purchases-table-row"]')
      .should('have.length', 1)
      .and('contain', 'CSV Pokemon order')
      .and('contain', 'Amazon CSV row 2')
      .and('contain', 'csv import')
      .and('contain', 'TBA123456')
      .and('contain', '2026-06-08')
      .and('contain', 'Expected 2026-06-12')
      .and('contain', 'Persisted lifecycle life-imp')

    cy.get('[data-testid="purchases-add-button"]').click()
    cy.get('[data-testid="purchases-add-tab-email"]').click()
    cy.get('[data-testid="purchases-add-email-input"]')
      .clear()
      .type(
        'Source: eBay\nTitle: Email Flute order\nPrice: AU $2.40\nPurchase Date: 2026-06-08\nSeller: seller-one\nChannel: email\nTracking: 1ZEMAILPURCHASE\nDelivery: Expected 2026-06-14',
        { delay: 0 }
      )
    cy.get('[data-testid="purchases-add-email-preview"]').click()
    cy.get('[data-testid="purchases-add-email-preview-result"]')
      .should('contain', 'Previewing email purchase draft')
      .and('contain', 'Email Flute order')
      .and('contain', 'eBay pasted email text')
      .and('contain', 'AU $2.40')
      .and('contain', 'seller-one')
      .and('contain', '1ZEMAILPURCHASE')
    cy.get('[data-testid="purchases-table-row"]').should('have.length', 1)
    cy.get('[data-testid="purchases-add-email-confirm"]').click()
    cy.wait('@createPurchaseImportItem')
      .its('request.body')
      .should('include', {
        title: 'Email Flute order',
        brand: 'email',
        category: 'Purchases',
      })
    cy.wait('@createPurchaseImportLifecycle')
      .its('request.body')
      .should('include', {
        item_id: 'purchase-import-item-2',
        state: 'purchase',
        source: 'eBay pasted email text',
        external_ref: '1ZEMAILPURCHASE',
        amount: 2.4,
        currency: 'AUD',
      })
    cy.get('[data-testid="purchases-add-dialog"]').should('not.exist')
    cy.get('[data-testid="purchases-table-row"]')
      .should('have.length', 2)
      .and('contain', 'Email Flute order')
      .and('contain', 'email import')
      .and('contain', 'Expected 2026-06-14')
      .and('contain', 'Persisted lifecycle life-imp')
    cy.get('[data-testid="purchases-table-search"]').type('Email Flute')
    cy.get('[data-testid="purchases-table-row"]')
      .should('have.length', 1)
      .and('contain', 'Email Flute order')

    cy.get('[data-testid="purchases-add-button"]').click()
    cy.get('[data-testid="purchases-add-tab-csv"]').click()
    cy.get('[data-testid="purchases-add-csv-input"]').clear().type('title')
    cy.get('[data-testid="purchases-add-csv-preview"]').click()
    cy.get('[data-testid="purchases-add-import-error"]').should(
      'contain',
      'CSV import needs a header row'
    )
  })

  it.skip('COMMERCE-RECONCILIATION-011 shows row purchase metadata and order links', () => {
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
                  item_price: 'AU $2.40',
                  seller_username: 'seller-one',
                  purchase_date: '2026-06-08',
                  item_url: 'https://www.ebay.com.au/itm/316046161178',
                },
              },
            ],
          },
        ],
      },
    }).as('preparePurchaseReviews')

    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/purchases',
    })
    cy.get('[data-testid="purchase-inbox-load-reviews"]').click()
    cy.wait('@preparePurchaseReviews')

    cy.get('[data-testid="purchases-table-row"]')
      .should('have.length', 1)
      .and('contain', 'Accompanying Flute TWM 142')
      .and('contain', '2026-06-08')
      .and('contain', 'Pending')
    cy.get('[data-testid="purchases-row-purchase-date"]').should(
      'contain',
      '2026-06-08'
    )
    cy.get('[data-testid="purchases-row-delivery"]').should(
      'contain',
      'Pending'
    )
    cy.get('[data-testid="purchases-row-order-link"]')
      .should('have.attr', 'target', '_blank')
      .and('have.attr', 'href', 'https://www.ebay.com.au/itm/316046161178')

    cy.get('[data-testid="purchases-table-search"]').type('316046161178')
    cy.get('[data-testid="purchases-table-row"]')
      .should('have.length', 1)
      .and('contain', 'Accompanying Flute TWM 142')
  })

  it.skip('EBAY-PURCHASE-CAPTURE-006 exposes a retryable error state', () => {
    cy.viewport(1400, 900)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.intercept('POST', '/api/integrations/ebay/purchase-inbox/reviews', {
      statusCode: 500,
      body: { error: 'failed' },
    }).as('preparePurchaseReviews')

    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/purchases',
    })
    cy.get('[data-testid="purchase-inbox-load-reviews"]').click()
    cy.wait('@preparePurchaseReviews')
    cy.get('[data-testid="purchase-inbox-error-state"]')
      .should('be.visible')
      .and('contain', 'Purchases could not load reviews.')
  })

  it.skip('EBAY-PURCHASE-CAPTURE-006 exposes loading state while reviews are prepared', () => {
    cy.viewport(1400, 900)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.intercept('POST', '/api/integrations/ebay/purchase-inbox/reviews', (req) => {
      req.reply({
        delay: 500,
        statusCode: 200,
        body: {
          source: 'ebay_purchase_capture',
          profile_id: 'e2e-profile-001',
          reviews: [],
        },
      })
    }).as('preparePurchaseReviews')

    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/purchases',
    })
    cy.get('[data-testid="purchase-inbox-load-reviews"]').click()
    cy.get('[data-testid="purchase-inbox-loading-state"]')
      .should('be.visible')
      .and('contain', 'Preparing purchase review records...')
    cy.wait('@preparePurchaseReviews')
    cy.get('[data-testid="purchase-inbox-loading-state"]').should('not.exist')
    cy.get('[data-testid="purchase-inbox-empty-state"]').should('be.visible')
  })

  it.skip('INTEGRATION-032 imports and lists manual forwarder packages', () => {
    cy.viewport(1400, 900)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.intercept('GET', '/api/forwarding/packages*', {
      statusCode: 200,
      body: {
        packages: [
          {
            id: 'fwdpkg_test_001',
            profile_id: 'e2e-profile-001',
            provider: 'stackry',
            source: 'manual',
            external_package_id: 'STK-PKG-1001',
            shipment_id: 'SHIP-1001',
            tracking_number: '1Z999AA10123456784',
            status: 'received',
            warehouse_location: 'Locker A-12',
            weight_grams: 420,
            provenance_key: 'stackry:manual:STK-PKG-1001',
          },
        ],
        summary: { count: 1 },
      },
    }).as('listForwarderPackages')
    cy.intercept('GET', '/api/forwarding/package-links*', {
      statusCode: 200,
      body: { links: [], summary: { count: 0 } },
    }).as('listForwarderPackageLinks')
    cy.intercept('POST', '/api/forwarding/packages', {
      statusCode: 200,
      body: {
        mode: 'forwarder_package_upsert',
        package: {
          id: 'fwdpkg_test_001',
          profile_id: 'e2e-profile-001',
          provider: 'stackry',
          source: 'manual',
          external_package_id: 'STK-PKG-1001',
          shipment_id: 'SHIP-1001',
          tracking_number: '1Z999AA10123456784',
          status: 'received',
          warehouse_location: 'Locker A-12',
          weight_grams: 420,
          provenance_key: 'stackry:manual:STK-PKG-1001',
        },
      },
    }).as('importForwarderPackage')

    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/purchases',
    })
    openSourceMatches()
    cy.get('[data-testid="forwarder-package-inbox"]').should('be.visible')
    cy.get('[data-testid="forwarder-package-inbox"]')
      .should('contain', 'Purchase Source Matches')
      .and('contain', 'source-backed purchase candidates')
    cy.get('[data-testid="forwarder-package-import"]').click()
    cy.wait('@importForwarderPackage')
      .its('request.body')
      .should('include', {
        profile_id: 'e2e-profile-001',
        provider: 'stackry',
        source: 'manual',
        external_package_id: 'STK-PKG-1001',
        status: 'received',
        weight_grams: 420,
      })
    cy.wait('@listForwarderPackages')
    cy.get('[data-testid="forwarder-package-result"]').should(
      'contain',
      'Imported package STK-PKG-1001'
    )
    cy.get('[data-testid="forwarder-package-list"]')
      .should('contain', 'STK-PKG-1001')
      .and('contain', 'stackry / manual')
      .and('contain', 'Locker A-12')
      .and('contain', 'stackry:manual:STK-PKG-1001')
  })

  it.skip('INTEGRATION-032 shows forwarder package import validation errors', () => {
    cy.viewport(1400, 900)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.intercept('POST', '/api/forwarding/packages', {
      statusCode: 400,
      body: {
        error: 'invalid_forwarder_package',
        message: 'external_package_id is required',
      },
    }).as('importForwarderPackage')

    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/purchases',
    })
    openSourceMatches()
    cy.get('[data-testid="forwarder-package-external-id"]').clear()
    cy.get('[data-testid="forwarder-package-import"]').click()
    cy.wait('@importForwarderPackage')
    cy.get('[data-testid="forwarder-package-error"]')
      .should('be.visible')
      .and('contain', 'external_package_id is required')
  })

  it.skip('INTEGRATION-034 imports CSV forwarder package rows and reports row errors', () => {
    cy.viewport(1400, 900)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.intercept('POST', '/api/forwarding/packages/import-csv', {
      statusCode: 200,
      body: {
        mode: 'forwarder_package_csv_import',
        imported: [
          {
            id: 'fwdpkg_csv_001',
            profile_id: 'e2e-profile-001',
            provider: 'stackry',
            source: 'csv',
            external_package_id: 'STK-CSV-2001',
            shipment_id: 'SHIP-2001',
            tracking_number: '1ZCSV2001',
            status: 'received',
            warehouse_location: 'Locker C-4',
            weight_grams: 520,
            provenance_key: 'stackry:csv:STK-CSV-2001',
          },
        ],
        errors: [{ row: 3, error: 'external_package_id is required' }],
        summary: { imported: 1, errors: 1 },
      },
    }).as('importForwarderPackageCSV')
    cy.intercept('GET', '/api/forwarding/packages*', {
      statusCode: 200,
      body: {
        packages: [
          {
            id: 'fwdpkg_csv_001',
            profile_id: 'e2e-profile-001',
            provider: 'stackry',
            source: 'csv',
            external_package_id: 'STK-CSV-2001',
            shipment_id: 'SHIP-2001',
            tracking_number: '1ZCSV2001',
            status: 'received',
            warehouse_location: 'Locker C-4',
            weight_grams: 520,
            provenance_key: 'stackry:csv:STK-CSV-2001',
          },
        ],
        summary: { count: 1 },
      },
    }).as('listForwarderPackages')
    cy.intercept('GET', '/api/forwarding/package-links*', {
      statusCode: 200,
      body: { links: [], summary: { count: 0 } },
    }).as('listForwarderPackageLinks')

    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/purchases',
    })
    openSourceMatches()
    cy.get('[data-testid="forwarder-package-csv"]').clear()
    cy.get('[data-testid="forwarder-package-csv"]').type(
      'Stackry Package ID,Status,Shipment ID,Tracking Number,Warehouse Location,Weight Grams\nSTK-CSV-2001,received,SHIP-2001,1ZCSV2001,Locker C-4,520\n,received,SHIP-2002,1ZCSV2002,Locker D-8,420',
      { delay: 0 }
    )
    cy.get('[data-testid="forwarder-package-import-csv"]').click()
    cy.wait('@importForwarderPackageCSV')
      .its('request.body')
      .should('include', {
        profile_id: 'e2e-profile-001',
        provider: 'stackry',
      })
    cy.wait('@listForwarderPackages')
    cy.get('[data-testid="forwarder-package-result"]')
      .should('contain', 'Imported 1 CSV package')
      .and('contain', '1 row needs attention')
    cy.get('[data-testid="forwarder-package-csv-errors"]')
      .should('contain', 'Row 3')
      .and('contain', 'external_package_id is required')
    cy.get('[data-testid="forwarder-package-list"]')
      .should('contain', 'STK-CSV-2001')
      .and('contain', 'stackry / csv')
      .and('contain', 'stackry:csv:STK-CSV-2001')
  })

  it.skip('INTEGRATION-036/037 imports email notices and shows package detail provenance', () => {
    cy.viewport(1400, 900)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.intercept('POST', '/api/forwarding/packages/import-email', {
      statusCode: 200,
      body: {
        mode: 'forwarder_package_email_import',
        package: {
          id: 'fwdpkg_email_001',
          profile_id: 'e2e-profile-001',
          provider: 'stackry',
          source: 'email',
          external_package_id: 'STK-EMAIL-3001',
          shipment_id: 'SHIP-3001',
          tracking_number: '1ZEMAIL3001',
          status: 'received',
          received_at: '2026-05-25T05:00:00Z',
          sender: 'Stackry Intake',
          warehouse_location: 'Locker E-5',
          weight_grams: 640,
          provenance_key: 'stackry:email:STK-EMAIL-3001',
          raw_payload: {
            message_id: 'manual-email-notice',
            sender: 'Stackry Intake',
          },
          created_at: '2026-05-25T05:01:00Z',
          updated_at: '2026-05-25T05:02:00Z',
        },
      },
    }).as('importForwarderPackageEmail')
    cy.intercept('GET', '/api/forwarding/packages*', {
      statusCode: 200,
      body: {
        packages: [
          {
            id: 'fwdpkg_email_001',
            profile_id: 'e2e-profile-001',
            provider: 'stackry',
            source: 'email',
            external_package_id: 'STK-EMAIL-3001',
            shipment_id: 'SHIP-3001',
            tracking_number: '1ZEMAIL3001',
            status: 'received',
            received_at: '2026-05-25T05:00:00Z',
            sender: 'Stackry Intake',
            warehouse_location: 'Locker E-5',
            weight_grams: 640,
            provenance_key: 'stackry:email:STK-EMAIL-3001',
            raw_payload: {
              message_id: 'manual-email-notice',
              sender: 'Stackry Intake',
            },
            created_at: '2026-05-25T05:01:00Z',
            updated_at: '2026-05-25T05:02:00Z',
          },
        ],
        summary: { count: 1 },
      },
    }).as('listForwarderPackages')
    cy.intercept('GET', '/api/forwarding/package-links*', {
      statusCode: 200,
      body: { links: [], summary: { count: 0 } },
    }).as('listForwarderPackageLinks')

    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/purchases',
    })
    openSourceMatches()
    cy.get('[data-testid="forwarder-package-email"]').clear()
    cy.get('[data-testid="forwarder-package-email"]').type(
      'Package ID: STK-EMAIL-3001\nStatus: Received\nShipment ID: SHIP-3001\nTracking Number: 1ZEMAIL3001\nWarehouse Location: Locker E-5\nWeight Grams: 640\nSender: Stackry Intake',
      { delay: 0 }
    )
    cy.get('[data-testid="forwarder-package-import-email"]').click()
    cy.wait('@importForwarderPackageEmail')
      .its('request.body')
      .should('include', {
        profile_id: 'e2e-profile-001',
        provider: 'stackry',
        message_id: 'manual-email-notice',
      })
    cy.wait('@listForwarderPackages')
    cy.get('[data-testid="forwarder-package-result"]').should(
      'contain',
      'Imported email package STK-EMAIL-3001'
    )
    cy.get('[data-testid="forwarder-package-list"]')
      .should('contain', 'STK-EMAIL-3001')
      .and('contain', 'stackry / email')
      .and('contain', 'Stackry Intake')
      .and('contain', 'stackry:email:STK-EMAIL-3001')
    cy.get('[data-testid="forwarder-package-detail-toggle"]').click()
    cy.wait('@listForwarderPackageLinks')
    cy.get('[data-testid="forwarder-package-detail"]')
      .should('contain', 'Shipment ID')
      .and('contain', 'SHIP-3001')
      .and('contain', 'Tracking number')
      .and('contain', '1ZEMAIL3001')
      .and('contain', 'Received')
      .and('contain', '2026-05-25T05:00:00Z')
      .and('contain', 'Created')
      .and('contain', '2026-05-25T05:01:00Z')
      .and('contain', 'Updated')
      .and('contain', '2026-05-25T05:02:00Z')
    cy.get('[data-testid="forwarder-package-raw-payload"]')
      .should('contain', 'manual-email-notice')
      .and('contain', 'Stackry Intake')

    cy.intercept('POST', '/api/forwarding/packages/import-email', {
      statusCode: 400,
      body: {
        error: 'invalid_forwarder_package_email',
        message: 'external_package_id is required',
      },
    }).as('importInvalidForwarderPackageEmail')
    cy.get('[data-testid="forwarder-package-email"]').clear()
    cy.get('[data-testid="forwarder-package-email"]').type(
      'Status: Received\nTracking Number: 1ZEMAIL3002',
      { delay: 0 }
    )
    cy.get('[data-testid="forwarder-package-import-email"]').click()
    cy.wait('@importInvalidForwarderPackageEmail')
    cy.get('[data-testid="forwarder-package-error"]')
      .should('be.visible')
      .and('contain', 'external_package_id is required')
  })

  it.skip('INTEGRATION-039 links forwarder packages to purchase arrivals from the inbox UI', () => {
    cy.viewport(1400, 900)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.intercept('GET', '/api/forwarding/packages*', {
      statusCode: 200,
      body: {
        packages: [
          {
            id: 'fwdpkg_link_001',
            profile_id: 'e2e-profile-001',
            provider: 'stackry',
            source: 'email',
            external_package_id: 'STK-LINK-4001',
            shipment_id: 'SHIP-4001',
            tracking_number: '1ZLINK4001',
            status: 'received',
            sender: 'Stackry Intake',
            warehouse_location: 'Locker L-4',
            weight_grams: 720,
            provenance_key: 'stackry:email:STK-LINK-4001',
          },
        ],
        summary: { count: 1 },
      },
    }).as('listForwarderPackages')
    cy.intercept('GET', '/api/forwarding/package-links*', {
      statusCode: 200,
      body: { links: [], summary: { count: 0 } },
    }).as('listForwarderPackageLinks')
    cy.intercept('POST', '/api/forwarding/package-links', {
      statusCode: 200,
      body: {
        mode: 'forwarder_package_reconciliation_link',
        link: {
          id: 'fwdpkg_link_001:item-expected-001:arrival-expected-001',
          profile_id: 'e2e-profile-001',
          package_id: 'fwdpkg_link_001',
          item_id: 'item-expected-001',
          lifecycle_entry_id: 'life-entry-001',
          expected_arrival_id: 'arrival-expected-001',
          source: 'manual_review',
          notes: 'Matched from package inbox review',
        },
      },
    }).as('saveForwarderPackageLink')

    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/purchases',
    })
    openSourceMatches()
    cy.get('[data-testid="forwarder-package-refresh"]').click()
    cy.wait('@listForwarderPackages')
    cy.get('[data-testid="forwarder-package-detail-toggle"]').click()
    cy.wait('@listForwarderPackageLinks')
      .its('request.url')
      .should('include', 'package_id=fwdpkg_link_001')
    cy.get('[data-testid="forwarder-package-link-panel"]')
      .should('be.visible')
      .and('contain', 'No reconciliation link recorded')
    cy.get('[data-testid="forwarder-package-link-save"]').click()
    cy.wait('@saveForwarderPackageLink')
      .its('request.body')
      .should('include', {
        package_id: 'fwdpkg_link_001',
        item_id: 'item-expected-001',
        lifecycle_entry_id: 'life-entry-001',
        expected_arrival_id: 'arrival-expected-001',
        source: 'manual_review',
      })
    cy.get('[data-testid="forwarder-package-link-result"]')
      .should('be.visible')
      .and('contain', 'Confirmed link to item item-expected-001')

    cy.intercept('POST', '/api/forwarding/package-links', {
      statusCode: 400,
      body: {
        error: 'invalid_forwarder_package_link',
        message: 'forwarder package already linked to a different target',
      },
    }).as('rejectAmbiguousForwarderPackageLink')
    cy.get('[data-testid="forwarder-package-link-arrival"]')
      .clear()
      .type('arrival-other-002')
    cy.get('[data-testid="forwarder-package-link-save"]').click()
    cy.wait('@rejectAmbiguousForwarderPackageLink')
    cy.get('[data-testid="forwarder-package-link-error"]')
      .should('be.visible')
      .and('contain', 'already linked to a different target')
  })

  it.skip('INTEGRATION-042 confirms overrides unlinks and shows forwarder package link audit events', () => {
    cy.viewport(1400, 900)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.intercept('GET', '/api/forwarding/packages*', {
      statusCode: 200,
      body: {
        packages: [
          {
            id: 'fwdpkg_audit_001',
            profile_id: 'e2e-profile-001',
            provider: 'stackry',
            source: 'email',
            external_package_id: 'STK-AUDIT-5001',
            shipment_id: 'SHIP-5001',
            tracking_number: '1ZAUDIT5001',
            status: 'received',
            sender: 'Stackry Intake',
            warehouse_location: 'Locker A-5',
            weight_grams: 840,
            provenance_key: 'stackry:email:STK-AUDIT-5001',
          },
        ],
        summary: { count: 1 },
      },
    }).as('listForwarderPackages')

    let linkListRequest = 0
    cy.intercept('GET', '/api/forwarding/package-links*', (req) => {
      linkListRequest += 1
      const states = [
        { links: [], events: [] },
        {
          links: [
            {
              id: 'fwdpkg_audit_001:item-expected-001:arrival-expected-001',
              profile_id: 'e2e-profile-001',
              package_id: 'fwdpkg_audit_001',
              item_id: 'item-expected-001',
              lifecycle_entry_id: 'life-entry-001',
              expected_arrival_id: 'arrival-expected-001',
              source: 'manual_review',
              decision: 'confirmed',
              notes: 'Matched from package inbox review',
              audit_trail: [
                'confirmed from purchase inbox UI: item-expected-001 / arrival-expected-001',
              ],
            },
          ],
          events: [
            {
              id: 'event-confirmed',
              package_id: 'fwdpkg_audit_001',
              action: 'confirmed',
              item_id: 'item-expected-001',
              lifecycle_entry_id: 'life-entry-001',
              expected_arrival_id: 'arrival-expected-001',
              source: 'manual_review',
              notes: 'Matched from package inbox review',
              audit_trail: [
                'confirmed from purchase inbox UI: item-expected-001 / arrival-expected-001',
              ],
              created_at: '2026-05-25T09:00:00Z',
            },
          ],
        },
        {
          links: [
            {
              id: 'fwdpkg_audit_001:item-override-002:arrival-override-002',
              profile_id: 'e2e-profile-001',
              package_id: 'fwdpkg_audit_001',
              item_id: 'item-override-002',
              lifecycle_entry_id: 'life-entry-override-002',
              expected_arrival_id: 'arrival-override-002',
              source: 'manual_override',
              decision: 'overridden',
              notes: 'Override to corrected package target',
              audit_trail: [
                'overridden from purchase inbox UI: item-override-002 / arrival-override-002',
              ],
            },
          ],
          events: [
            {
              id: 'event-overridden',
              package_id: 'fwdpkg_audit_001',
              action: 'overridden',
              item_id: 'item-override-002',
              lifecycle_entry_id: 'life-entry-override-002',
              expected_arrival_id: 'arrival-override-002',
              previous_item_id: 'item-expected-001',
              previous_lifecycle_entry_id: 'life-entry-001',
              previous_expected_arrival_id: 'arrival-expected-001',
              source: 'manual_override',
              notes: 'Override to corrected package target',
              audit_trail: [
                'overridden from purchase inbox UI: item-override-002 / arrival-override-002',
              ],
              created_at: '2026-05-25T09:05:00Z',
            },
            {
              id: 'event-confirmed',
              package_id: 'fwdpkg_audit_001',
              action: 'confirmed',
              item_id: 'item-expected-001',
              lifecycle_entry_id: 'life-entry-001',
              expected_arrival_id: 'arrival-expected-001',
              source: 'manual_review',
              notes: 'Matched from package inbox review',
              audit_trail: [
                'confirmed from purchase inbox UI: item-expected-001 / arrival-expected-001',
              ],
              created_at: '2026-05-25T09:00:00Z',
            },
          ],
        },
        {
          links: [],
          events: [
            {
              id: 'event-unlinked',
              package_id: 'fwdpkg_audit_001',
              action: 'unlinked',
              previous_item_id: 'item-override-002',
              previous_lifecycle_entry_id: 'life-entry-override-002',
              previous_expected_arrival_id: 'arrival-override-002',
              source: 'manual_unlink',
              notes: 'Override to corrected package target',
              audit_trail: [
                'unlinked from purchase inbox UI: item-override-002 / arrival-override-002',
              ],
              created_at: '2026-05-25T09:08:00Z',
            },
            {
              id: 'event-overridden',
              package_id: 'fwdpkg_audit_001',
              action: 'overridden',
              item_id: 'item-override-002',
              lifecycle_entry_id: 'life-entry-override-002',
              expected_arrival_id: 'arrival-override-002',
              previous_item_id: 'item-expected-001',
              previous_lifecycle_entry_id: 'life-entry-001',
              previous_expected_arrival_id: 'arrival-expected-001',
              source: 'manual_override',
              notes: 'Override to corrected package target',
              audit_trail: [
                'overridden from purchase inbox UI: item-override-002 / arrival-override-002',
              ],
              created_at: '2026-05-25T09:05:00Z',
            },
          ],
        },
      ]
      const state = states[Math.min(linkListRequest - 1, states.length - 1)]
      req.reply({
        statusCode: 200,
        body: {
          links: state.links,
          events: state.events,
          summary: {
            count: state.links.length,
            events: state.events.length,
          },
        },
      })
    }).as('listForwarderPackageLinks')

    cy.intercept('POST', '/api/forwarding/package-links', (req) => {
      expect(req.body).to.include({
        package_id: 'fwdpkg_audit_001',
        item_id: 'item-expected-001',
        lifecycle_entry_id: 'life-entry-001',
        expected_arrival_id: 'arrival-expected-001',
        source: 'manual_review',
        decision: 'confirmed',
        override: false,
        actor: 'reviewer',
      })
      expect(req.body.audit_trail).to.deep.equal([
        'confirmed from purchase inbox UI: item-expected-001 / arrival-expected-001',
      ])
      req.reply({
        statusCode: 200,
        body: {
          mode: 'forwarder_package_reconciliation_link',
          link: {
            id: 'fwdpkg_audit_001:item-expected-001:arrival-expected-001',
            package_id: 'fwdpkg_audit_001',
            item_id: 'item-expected-001',
            expected_arrival_id: 'arrival-expected-001',
            source: 'manual_review',
            decision: 'confirmed',
          },
        },
      })
    }).as('confirmForwarderPackageLink')

    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/purchases',
    })
    openSourceMatches()
    cy.get('[data-testid="forwarder-package-refresh"]').click()
    cy.wait('@listForwarderPackages')
    cy.get('[data-testid="forwarder-package-review-summary"]')
      .should('contain', 'Source records')
      .and('contain', '1')
      .and('contain', 'Linked')
      .and('contain', '0')
      .and('contain', 'Unlinked')
      .and('contain', '1')
    cy.get('[data-testid="forwarder-package-review-filter-result"]').should(
      'contain',
      'Showing 1 of 1 packages'
    )
    cy.get('[data-testid="forwarder-package-review-filter-linked"]').click()
    cy.get('[data-testid="forwarder-package-filter-empty"]').should(
      'contain',
      'No source records match this review state'
    )
    cy.get('[data-testid="forwarder-package-review-filter-result"]').should(
      'contain',
      'Showing 0 of 1 packages'
    )
    cy.get('[data-testid="forwarder-package-review-filter-unlinked"]').click()
    cy.get('[data-testid="forwarder-package-row"]')
      .should('have.length', 1)
      .and('contain', 'STK-AUDIT-5001')
    cy.get('[data-testid="forwarder-package-row-evidence"]').should(
      'contain',
      'No loaded reconciliation evidence'
    )
    cy.get('[data-testid="forwarder-package-review-filter-all"]').click()
    cy.get('[data-testid="forwarder-package-detail-toggle"]').click()
    cy.wait('@listForwarderPackageLinks')
    cy.get('[data-testid="forwarder-package-link-events"]')
      .should('be.visible')
      .and('contain', 'No link decisions recorded yet.')

    cy.get('[data-testid="forwarder-package-link-save"]').click()
    cy.wait('@confirmForwarderPackageLink')
    cy.wait('@listForwarderPackageLinks')
    cy.get('[data-testid="forwarder-package-link-result"]')
      .should('be.visible')
      .and('contain', 'Confirmed link to item item-expected-001')
    cy.get('[data-testid="forwarder-package-link-state"]')
      .should('contain', 'confirmed to item item-expected-001')
      .and('contain', 'arrival arrival-expected-001')
      .and('contain', 'Source manual_review')
    cy.get('[data-testid="forwarder-package-review-summary"]')
      .should('contain', 'Linked')
      .and('contain', '1')
      .and('contain', 'Unlinked')
      .and('contain', '0')
      .and('contain', 'Audit events')
      .and('contain', '1')
    cy.get('[data-testid="forwarder-package-review-filter-linked"]').click()
    cy.get('[data-testid="forwarder-package-row"]')
      .should('have.length', 1)
      .and('contain', 'STK-AUDIT-5001')
    cy.get('[data-testid="forwarder-package-row-evidence"]')
      .should('contain', '1 active link')
      .and('contain', '1 audit event')
    cy.get('[data-testid="forwarder-package-review-filter-all"]').click()
    cy.get('[data-testid="forwarder-package-link-audit-trail"]').should(
      'contain',
      'confirmed from purchase inbox UI'
    )
    cy.get('[data-testid="forwarder-package-link-events"]')
      .should('contain', 'confirmed item item-expected-001')
      .and('contain', 'lifecycle life-entry-001')
      .and('contain', '2026-05-25T09:00:00Z')
      .and('contain', 'via manual_review')
    cy.get('[data-testid="forwarder-package-link-event-audit-trail"]').should(
      'contain',
      'confirmed from purchase inbox UI'
    )

    cy.intercept('POST', '/api/forwarding/package-links', (req) => {
      expect(req.body).to.include({
        package_id: 'fwdpkg_audit_001',
        item_id: 'item-override-002',
        lifecycle_entry_id: 'life-entry-override-002',
        expected_arrival_id: 'arrival-override-002',
        source: 'manual_override',
        decision: 'overridden',
        override: true,
        actor: 'reviewer',
      })
      expect(req.body.audit_trail).to.deep.equal([
        'overridden from purchase inbox UI: item-override-002 / arrival-override-002',
      ])
      req.reply({
        statusCode: 200,
        body: {
          mode: 'forwarder_package_reconciliation_link',
          link: {
            id: 'fwdpkg_audit_001:item-override-002:arrival-override-002',
            package_id: 'fwdpkg_audit_001',
            item_id: 'item-override-002',
            expected_arrival_id: 'arrival-override-002',
            source: 'manual_override',
            decision: 'overridden',
          },
        },
      })
    }).as('overrideForwarderPackageLink')

    cy.get('[data-testid="forwarder-package-link-item"]')
      .clear()
      .type('item-override-002')
    cy.get('[data-testid="forwarder-package-link-lifecycle"]')
      .clear()
      .type('life-entry-override-002')
    cy.get('[data-testid="forwarder-package-link-arrival"]')
      .clear()
      .type('arrival-override-002')
    cy.get('[data-testid="forwarder-package-link-notes"]')
      .clear()
      .type('Override to corrected package target')
    cy.get('[data-testid="forwarder-package-link-override"]').click()
    cy.wait('@overrideForwarderPackageLink')
    cy.wait('@listForwarderPackageLinks')
    cy.get('[data-testid="forwarder-package-link-result"]').should(
      'contain',
      'Override linked to item item-override-002'
    )
    cy.get('[data-testid="forwarder-package-link-state"]')
      .should('contain', 'overridden to item item-override-002')
      .and('contain', 'Source manual_override')
    cy.get('[data-testid="forwarder-package-link-events"]')
      .should('contain', 'overridden item item-override-002')
      .and('contain', 'previous item item-expected-001')
      .and('contain', 'previous item item-expected-001 / lifecycle life-entry-001 / arrival arrival-expected-001')
      .and('contain', '2026-05-25T09:05:00Z')

    cy.intercept('DELETE', '/api/forwarding/package-links*', (req) => {
      expect(req.url).to.include('package_id=fwdpkg_audit_001')
      expect(req.body).to.include({
        source: 'manual_unlink',
        actor: 'reviewer',
        notes: 'Override to corrected package target',
      })
      expect(req.body.audit_trail).to.deep.equal([
        'unlinked from purchase inbox UI: item-override-002 / arrival-override-002',
      ])
      req.reply({
        statusCode: 200,
        body: {
          mode: 'forwarder_package_reconciliation_unlink',
          event: {
            id: 'event-unlinked',
            package_id: 'fwdpkg_audit_001',
            action: 'unlinked',
            previous_item_id: 'item-override-002',
            source: 'manual_unlink',
          },
        },
      })
    }).as('unlinkForwarderPackage')

    cy.get('[data-testid="forwarder-package-link-unlink"]').click()
    cy.wait('@unlinkForwarderPackage')
    cy.wait('@listForwarderPackageLinks')
    cy.get('[data-testid="forwarder-package-link-result"]').should(
      'contain',
      'Unlinked package from reconciliation target'
    )
    cy.get('[data-testid="forwarder-package-link-empty"]').should(
      'contain',
      'No reconciliation link recorded'
    )
    cy.get('[data-testid="forwarder-package-review-summary"]')
      .should('contain', 'Linked')
      .and('contain', '0')
      .and('contain', 'Unlinked')
      .and('contain', '1')
      .and('contain', 'Audit events')
      .and('contain', '2')
    cy.get('[data-testid="forwarder-package-review-filter-unlinked"]').click()
    cy.get('[data-testid="forwarder-package-row"]')
      .should('have.length', 1)
      .and('contain', 'STK-AUDIT-5001')
    cy.get('[data-testid="forwarder-package-link-events"]')
      .should('contain', 'unlinked')
      .and('contain', 'previous item item-override-002')
      .and('contain', 'previous item item-override-002 / lifecycle life-entry-override-002 / arrival arrival-override-002')
      .and('contain', 'via manual_unlink')
    cy.get('[data-testid="forwarder-package-link-event-audit-trail"]').should(
      'contain',
      'unlinked from purchase inbox UI'
    )
  })

  it.skip('INTEGRATION-043 shows forwarder package match suggestions and prepares confirmation', () => {
    cy.viewport(1400, 900)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.intercept('GET', '/api/forwarding/packages*', {
      statusCode: 200,
      body: {
        packages: [
          {
            id: 'fwdpkg_suggest_001',
            profile_id: 'e2e-profile-001',
            provider: 'stackry',
            source: 'email',
            external_package_id: 'STK-SUGGEST-6001',
            shipment_id: 'SHIP-6001',
            tracking_number: '1ZSUGGEST6001',
            status: 'received',
            sender: 'slot shop',
            warehouse_location: 'Locker S-6',
            weight_grams: 760,
            provenance_key: 'stackry:email:STK-SUGGEST-6001',
          },
        ],
        summary: { count: 1 },
      },
    }).as('listForwarderPackages')

    cy.intercept('GET', '/api/forwarding/package-links*', {
      statusCode: 200,
      body: { links: [], events: [], summary: { count: 0, events: 0 } },
    }).as('listForwarderPackageLinks')

    let scopedSuggestionRequestSeen = false
    cy.intercept('GET', '/api/forwarding/package-match-suggestions*', (req) => {
      expect(req.query).to.include({ confidence_label: 'high' })
      if (req.query.package_id === 'fwdpkg_suggest_001') {
        scopedSuggestionRequestSeen = true
      }
      req.reply({
        statusCode: 200,
        body: {
          mode: 'forwarder_package_match_suggestions',
          mutable: false,
          confidence_filter: 'high',
          suggestions: [
            {
              id: 'suggestion-fwdpkg-001',
              package_id: 'fwdpkg_suggest_001',
              item_id: 'item-suggested-001',
              lifecycle_entry_id: 'life-suggested-001',
              expected_arrival_id: 'arrival-suggested-001',
              confidence_score: 94,
              confidence_label: 'high',
              signals: [
                {
                  name: 'tracking',
                  score: 40,
                  evidence: '1ZSUGGEST6001 matched purchase notes',
                },
                {
                  name: 'seller',
                  score: 20,
                  evidence: 'slot shop matched package sender',
                },
              ],
              audit_trail: [
                'suggested_match package=fwdpkg_suggest_001 item=item-suggested-001 confidence=high score=94',
              ],
            },
          ],
          summary: {
            count: 1,
            scoped_packages: 1,
            high_confidence: 1,
            medium_confidence: 0,
            low_confidence: 0,
          },
        },
      })
    }).as('listForwarderPackageMatchSuggestions')

    cy.intercept('POST', '/api/forwarding/package-links', (req) => {
      expect(req.body).to.include({
        package_id: 'fwdpkg_suggest_001',
        item_id: 'item-suggested-001',
        lifecycle_entry_id: 'life-suggested-001',
        expected_arrival_id: 'arrival-suggested-001',
        source: 'suggested_match',
        decision: 'confirmed',
        override: false,
        actor: 'reviewer',
      })
      expect(req.body.audit_trail).to.deep.equal([
        'confirmed from purchase inbox UI: item-suggested-001 / arrival-suggested-001',
        'suggested_match package=fwdpkg_suggest_001 item=item-suggested-001 confidence=high score=94',
      ])
      req.reply({
        statusCode: 200,
        body: {
          mode: 'forwarder_package_reconciliation_link',
          link: {
            id: 'fwdpkg_suggest_001:item-suggested-001:arrival-suggested-001',
            package_id: 'fwdpkg_suggest_001',
            item_id: 'item-suggested-001',
            expected_arrival_id: 'arrival-suggested-001',
            source: 'suggested_match',
            decision: 'confirmed',
          },
        },
      })
    }).as('confirmSuggestedForwarderPackageLink')

    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/purchases',
    })
    openSourceMatches()
    cy.get('[data-testid="forwarder-package-refresh"]').click()
    cy.wait('@listForwarderPackages')
    cy.get('[data-testid="forwarder-package-confidence-filter-high"]').click()
    cy.get('[data-testid="forwarder-package-match-suggestions-load"]').click()
    cy.wait('@listForwarderPackageMatchSuggestions')
    cy.get('[data-testid="forwarder-package-suggestion-result"]')
      .should('be.visible')
      .and('contain', 'Found 1 package match suggestion')
      .and('contain', 'high confidence')
    cy.get('[data-testid="forwarder-package-review-summary"]')
      .should('contain', 'Source records')
      .and('contain', 'Suggestions')
      .and('contain', '1')
    cy.get('[data-testid="forwarder-package-suggestion-summary"]')
      .should('contain', 'Candidates')
      .and('contain', '1')
      .and('contain', 'Scoped packages')
      .and('contain', 'High confidence')
      .and('contain', 'Medium confidence')
      .and('contain', 'Low confidence')
      .and('contain', 'Active filter')
      .and('contain', 'high')
      .and('contain', '0')
    cy.get('[data-testid="forwarder-package-review-filter-suggested"]').click()
    cy.get('[data-testid="forwarder-package-row"]')
      .should('have.length', 1)
      .and('contain', 'STK-SUGGEST-6001')
    cy.get('[data-testid="forwarder-package-row-evidence"]').should(
      'contain',
      '1 suggestion'
    )
    cy.get('[data-testid="forwarder-package-detail-toggle"]').click()
    cy.wait('@listForwarderPackageLinks')
    cy.get(
      '[data-testid="forwarder-package-match-suggestions-load-scoped"]'
    ).click()
    cy.wait('@listForwarderPackageMatchSuggestions')
    cy.then(() => {
      expect(scopedSuggestionRequestSeen).to.equal(true)
    })
    cy.get('[data-testid="forwarder-package-suggestion-result"]')
      .should('contain', 'Found 1 package match suggestion')
      .and('contain', 'for package fwdpkg_suggest_001')
      .and('contain', 'high confidence')
    cy.get('[data-testid="forwarder-package-match-suggestions"]')
      .should('be.visible')
      .and('contain', 'Suggested purchase matches')
      .and('contain', 'high match to item item-suggested-001')
      .and('contain', 'Score 94')
      .and('contain', 'arrival arrival-suggested-001')
    cy.get('[data-testid="forwarder-package-match-signals"]')
      .should('contain', 'tracking')
      .and('contain', '1ZSUGGEST6001 matched purchase notes')
    cy.get('[data-testid="forwarder-package-match-audit-trail"]').should(
      'contain',
      'suggested_match package=fwdpkg_suggest_001'
    )

    cy.get('[data-testid="forwarder-package-match-suggestion-use"]').click()
    cy.get('[data-testid="forwarder-package-link-item"]').should(
      'have.value',
      'item-suggested-001'
    )
    cy.get('[data-testid="forwarder-package-link-source"]').should(
      'have.value',
      'suggested_match'
    )
    cy.get('[data-testid="forwarder-package-link-result"]').should(
      'contain',
      'Prepared suggested match for item item-suggested-001'
    )
    cy.get('[data-testid="forwarder-package-link-panel"]')
      .should('contain', 'Match this source evidence')
      .and('contain', 'Confirm link')
      .and('contain', 'Override link')
      .and('contain', 'No reconciliation link recorded for this source record')
    cy.get('[data-testid="forwarder-package-link-save"]').click()
    cy.wait('@confirmSuggestedForwarderPackageLink')
  })

  it.skip('INTEGRATION-059 keeps package rows visible across empty and failed suggestion loads', () => {
    cy.viewport(1400, 900)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.intercept('GET', '/api/forwarding/packages*', {
      statusCode: 200,
      body: {
        packages: [
          {
            id: 'fwdpkg_empty_001',
            profile_id: 'e2e-profile-001',
            provider: 'stackry',
            source: 'manual',
            external_package_id: 'STK-EMPTY-7001',
            shipment_id: 'SHIP-7001',
            tracking_number: '1ZEMPTY7001',
            status: 'received',
            warehouse_location: 'Locker E-7',
            weight_grams: 510,
            provenance_key: 'stackry:manual:STK-EMPTY-7001',
          },
        ],
        summary: { count: 1 },
      },
    }).as('listForwarderPackages')
    cy.intercept('GET', '/api/forwarding/package-links*', {
      statusCode: 200,
      body: { links: [], events: [], summary: { count: 0, events: 0 } },
    }).as('listForwarderPackageLinks')

    let suggestionRequestCount = 0
    cy.intercept('GET', '/api/forwarding/package-match-suggestions*', (req) => {
      suggestionRequestCount += 1
      if (suggestionRequestCount === 1) {
        req.reply({
          statusCode: 200,
          body: {
            mode: 'forwarder_package_match_suggestions',
            mutable: false,
            confidence_filter: 'medium',
            suggestions: [],
            summary: {
              count: 0,
              scoped_packages: 0,
              high_confidence: 0,
              medium_confidence: 0,
              low_confidence: 0,
            },
          },
        })
        return
      }
      if (suggestionRequestCount === 2) {
        req.reply({
          statusCode: 503,
          body: { error: 'suggestions_unavailable' },
        })
        return
      }
      req.reply({
        statusCode: 200,
        body: {
          mode: 'forwarder_package_match_suggestions',
          mutable: false,
          suggestions: [],
          summary: { count: 0, scoped_packages: 0 },
        },
      })
    }).as('loadForwarderPackageSuggestions')

    cy.useBootstrappedProfile('e2e-profile-001', 'E2E Local', {
      path: '/purchases',
    })
    openSourceMatches()
    cy.get('[data-testid="forwarder-package-refresh"]').click()
    cy.wait('@listForwarderPackages')
    cy.get('[data-testid="forwarder-package-row"]')
      .should('have.length', 1)
      .and('contain', 'STK-EMPTY-7001')

    cy.get('[data-testid="forwarder-package-confidence-filter-medium"]').click()
    cy.get('[data-testid="forwarder-package-match-suggestions-load"]').click()
    cy.wait('@loadForwarderPackageSuggestions')
    cy.get('[data-testid="forwarder-package-suggestion-result"]')
      .should('be.visible')
      .and('contain', 'No package match suggestions matched the current inbox')
      .and('contain', 'medium confidence')
      .and('contain', 'Package rows were left unchanged')
    cy.get('[data-testid="forwarder-package-row"]')
      .should('have.length', 1)
      .and('contain', 'STK-EMPTY-7001')

    cy.get('[data-testid="forwarder-package-match-suggestions-load"]').click()
    cy.wait('@loadForwarderPackageSuggestions')
    cy.get('[data-testid="forwarder-package-suggestion-error"]')
      .should('be.visible')
      .and('contain', 'Match suggestions could not load')
      .and('contain', 'forwarder_package_match_suggestions_503')
    cy.get('[data-testid="forwarder-package-row"]')
      .should('have.length', 1)
      .and('contain', 'STK-EMPTY-7001')

    cy.get('[data-testid="forwarder-package-suggestion-retry"]').click()
    cy.wait('@loadForwarderPackageSuggestions')
    cy.get('[data-testid="forwarder-package-suggestion-error"]').should(
      'not.exist'
    )
  })
})
