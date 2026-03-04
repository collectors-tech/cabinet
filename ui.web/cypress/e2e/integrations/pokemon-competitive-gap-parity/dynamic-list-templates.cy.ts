describe('integrations/pokemon-competitive-gap-parity/dynamic-list-templates', () => {
  beforeEach(() => {
    cy.e2eReset()
    cy.e2eBootstrap()
  })

  it('POKEMON-COMP-006 lists reusable templates with deterministic contract fields', () => {
    cy.request('GET', '/api/integrations/pokemon/list-templates').then((resp) => {
      expect(resp.status).to.eq(200)
      const templates = Array.isArray(resp.body?.templates) ? resp.body.templates : []
      expect(templates.length).to.be.greaterThan(0)
      const ids = templates.map((template) => template.id)
      expect(ids).to.include.members(['wishlist', 'trade_binder', 'watchlist'])
      for (const template of templates) {
        expect(template).to.include.keys(
          'id',
          'label',
          'default_fields',
          'default_filters',
          'sort_order'
        )
      }
    })
  })

  it('POKEMON-COMP-006 applies trade_binder template with grading-centric defaults', () => {
    cy.request('POST', '/api/integrations/pokemon/list-templates/apply', {
      template_id: 'trade_binder',
      list_name: 'Trade Binder',
    }).then((resp) => {
      expect(resp.status).to.eq(201)
      expect(resp.body).to.include.keys(
        'list_id',
        'list_name',
        'template_id',
        'default_fields',
        'default_filters',
        'sort_order'
      )
      expect(resp.body.template_id).to.eq('trade_binder')
      expect(resp.body.default_fields).to.include.members([
        'grader',
        'grade_numeric',
        'collector_classification',
      ])
      expect(resp.body.default_filters).to.have.property('status')
    })
  })

  it('POKEMON-COMP-006 rejects unknown template ids deterministically', () => {
    cy.request({
      method: 'POST',
      url: '/api/integrations/pokemon/list-templates/apply',
      failOnStatusCode: false,
      body: { template_id: 'unknown-template' },
    }).then((resp) => {
      expect(resp.status).to.eq(400)
      expect(resp.body).to.have.property('error', 'invalid_template_id')
    })
  })
})
