describe('api-docs', () => {
  it('API-DOCS-001 exposes openapi yaml endpoint with deterministic status and content type', () => {
    cy.request('/api/openapi.yaml').then((response) => {
      expect(response.status).to.eq(200)
      expect(response.headers['content-type']).to.match(/yaml|text\/plain|application\/yaml/)
      expect(response.body).to.contain('openapi:')
    })
  })

  it('API-DOCS-002 serves apidocs route and binds openapi source endpoint', () => {
    cy.visit('/apidocs')
    cy.location('pathname').should('eq', '/apidocs')
    cy.contains('Cabinet API Docs').should('be.visible')
    cy.get('a[href="/api/openapi.yaml"]').should('be.visible')
  })
})
