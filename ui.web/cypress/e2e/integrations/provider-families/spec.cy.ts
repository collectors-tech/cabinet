describe('provider-families', () => {
  it('PROVIDER-FAMILY-005 + UC-PF-05: detects provider family from URL evidence markers', () => {
    cy.request('POST', '/api/providers/family-detect', {
      provider_url: 'https://example-boost.test',
      html: `
        <script src="https://services.mybcapps.com/bc-sf-filter/search"></script>
        <script>window.__BOOST__ = true</script>
      `,
    }).then((response) => {
      expect(response.status).to.eq(200)
      expect(response.body.proposed_api_family).to.eq('boost_shopify')
      expect(response.body.confidence).to.be.greaterThan(0)
      expect((response.body.evidence as unknown[]).length).to.be.greaterThan(0)
    })
  })

  it('PROVIDER-FAMILY-005 + UC-PF-06: override persists and is reflected in provider registry', () => {
    cy.request('POST', '/api/providers/family-override', {
      provider_domain: 'frontlinehobbies.com.au',
      api_family: 'doofinder',
    }).then((overrideResponse) => {
      expect(overrideResponse.status).to.eq(200)
      expect(overrideResponse.body.saved).to.eq(true)
    })

    cy.request('GET', '/api/providers/registry').then((registryResponse) => {
      expect(registryResponse.status).to.eq(200)
      const provider = (registryResponse.body.providers as Array<Record<string, unknown>>).find(
        (entry) => String(entry.base_domain || '') === 'frontlinehobbies.com.au'
      )
      expect(Boolean(provider)).to.equal(true)
      expect(String(provider!.api_family || '')).to.eq('doofinder')
    })
  })
})

