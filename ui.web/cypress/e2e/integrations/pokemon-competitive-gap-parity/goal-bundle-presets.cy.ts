describe('integrations/pokemon-competitive-gap-parity/goal-bundle-presets', () => {
  beforeEach(() => {
    cy.e2eReset()
    cy.e2eBootstrap()
  })

  it('POKEMON-COMP-010 returns deterministic goal bundle catalog', () => {
    cy.request('GET', '/api/integrations/pokemon/goal-bundles').then((resp) => {
      expect(resp.status).to.eq(200)
      const bundles = Array.isArray(resp.body?.bundles) ? resp.body.bundles : []
      expect(bundles.length).to.be.greaterThan(0)
      const ids = bundles.map((bundle) => bundle.id)
      expect(ids).to.include.members([
        'finish-master-set',
        'optimize-trade-binder',
        'price-drop-watch',
      ])
      for (const bundle of bundles) {
        expect(bundle).to.include.keys('id', 'label', 'filters', 'actions', 'shortcut')
      }
    })
  })

  it('POKEMON-COMP-010 applies goal bundle with deterministic workspace payload', () => {
    cy.request('POST', '/api/integrations/pokemon/goal-bundles/apply', {
      bundle_id: 'finish-master-set',
      workspace_name: 'Master Set Focus',
    }).then((resp) => {
      expect(resp.status).to.eq(201)
      expect(resp.body).to.include.keys(
        'workspace_id',
        'workspace_name',
        'bundle_id',
        'filters',
        'actions',
        'shortcut'
      )
      expect(resp.body.bundle_id).to.eq('finish-master-set')
      expect(resp.body.workspace_name).to.eq('Master Set Focus')
    })
  })

  it('POKEMON-COMP-010 rejects unknown bundle ids deterministically', () => {
    cy.request({
      method: 'POST',
      url: '/api/integrations/pokemon/goal-bundles/apply',
      failOnStatusCode: false,
      body: { bundle_id: 'unknown-bundle' },
    }).then((resp) => {
      expect(resp.status).to.eq(400)
      expect(resp.body).to.have.property('error', 'invalid_bundle_id')
    })
  })
})
