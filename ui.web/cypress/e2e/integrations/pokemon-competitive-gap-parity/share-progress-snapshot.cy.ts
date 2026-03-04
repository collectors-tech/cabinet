describe('integrations/pokemon-competitive-gap-parity/share-progress-snapshot', () => {
  beforeEach(() => {
    cy.e2eReset()
    cy.e2eBootstrap()
  })

  it('POKEMON-COMP-008 returns deterministic progress snapshot share payload', () => {
    cy.request('GET', '/api/integrations/pokemon/progress-snapshot?set_id=base-set&total_count=10').then((resp) => {
      expect(resp.status).to.eq(200)
      expect(resp.body).to.include.keys(
        'set_id',
        'owned_count',
        'total_count',
        'completion_percent',
        'share_payload',
        'generated_at'
      )
      expect(resp.body.set_id).to.eq('base-set')
      expect(resp.body.share_payload).to.include.keys('headline', 'summary', 'visibility', 'share_link')
      expect(resp.body.share_payload.visibility).to.eq('private')
    })
  })

  it('POKEMON-COMP-008 rejects missing set_id deterministically', () => {
    cy.request({
      method: 'GET',
      url: '/api/integrations/pokemon/progress-snapshot',
      failOnStatusCode: false,
    }).then((resp) => {
      expect(resp.status).to.eq(400)
      expect(resp.body).to.have.property('error', 'missing_set_id')
    })
  })
})
