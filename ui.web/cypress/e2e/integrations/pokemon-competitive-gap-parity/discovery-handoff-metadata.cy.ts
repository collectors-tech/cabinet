describe('integrations/pokemon-competitive-gap-parity/discovery-handoff-metadata', () => {
  let seededProfileId = ''
  let seededItemId = ''

  beforeEach(() => {
    cy.e2eReset()
    cy.e2eBootstrap().then((state) => {
      seededProfileId = state.profile_id
      seededItemId = state.item_ids[0] ?? ''
      cy.request('PUT', '/api/profiles/active', { profile_id: seededProfileId })
        .its('status')
        .should('eq', 200)
    })
  })

  it('POKEMON-COMP-004 preserves marketplace decision metadata when handing off discovery candidate to wishlist', () => {
    cy.request('POST', '/api/discovery/action', {
      candidate_id: 'e2e-candidate-001',
      type: 'add_to_wishlist',
    })
      .its('status')
      .should('eq', 200)

    cy.request<{ items: Array<{ item_id: string; notes: string }> }>({
      method: 'GET',
      url: '/api/wishlist',
    }).then((resp) => {
      expect(resp.status).to.eq(200)
      const items = Array.isArray(resp.body?.items) ? resp.body.items : []
      const entry = items.find((item) => item.item_id === seededItemId)
      expect(Boolean(entry), 'wishlist entry for matched item').to.equal(true)
      expect(entry?.notes).to.contain('[discovery_metadata]')
      const metadataRaw = (entry?.notes ?? '').split('[discovery_metadata]').pop() ?? '{}'
      const metadata = JSON.parse(metadataRaw)
      expect(metadata).to.include.keys('listing_url', 'seller', 'stock_signal', 'observed_price')
      expect(metadata.listing_url).to.eq('https://example.test/e2e-listing-001')
      expect(metadata.seller).to.eq('e2e-seller')
      expect(metadata.stock_signal).to.eq('in_stock')
      expect(metadata.observed_price).to.eq(44.95)
    })
  })

  it('POKEMON-COMP-004 rejects missing candidate id deterministically', () => {
    cy.request({
      method: 'POST',
      url: '/api/discovery/action',
      failOnStatusCode: false,
      body: { type: 'add_to_wishlist' },
    }).then((resp) => {
      expect(resp.status).to.eq(400)
      expect(resp.body).to.have.property('error', 'failed_to_apply_discovery_action')
    })
  })
})
