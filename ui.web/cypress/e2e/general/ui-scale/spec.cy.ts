describe('general/ui-scale', () => {
  beforeEach(() => {
    cy.e2eReset()
  })

  it('UI-SCALE-001 reproduces deterministic S2 benchmark profile output for same seed', () => {
    cy.request('POST', '/api/test/scale/bootstrap', {
      profile: 'S2',
      seed: 4242,
    }).then((first) => {
      expect(first.status).to.eq(200)
      const firstHash = first.body?.dataset_hash
      expect(firstHash).to.be.a('string').and.not.empty

      cy.e2eReset()
      cy.request('POST', '/api/test/scale/bootstrap', {
        profile: 'S2',
        seed: 4242,
      }).then((second) => {
        expect(second.status).to.eq(200)
        expect(second.body?.dataset_hash).to.eq(firstHash)
        expect(second.body?.counts).to.deep.eq(first.body?.counts)
      })
    })
  })

  it('UI-SCALE-002 keeps table-heavy API workflows operational under S3 profile', () => {
    cy.request('POST', '/api/test/scale/bootstrap', {
      profile: 'S3',
      seed: 5001,
    }).then((bootstrapResp) => {
      expect(bootstrapResp.status).to.eq(200)
      const querySetID = bootstrapResp.body?.query_set_id
      expect(querySetID).to.be.a('string').and.not.empty

      cy.request('GET', '/api/items?status=active').then((itemsResp) => {
        expect(itemsResp.status).to.eq(200)
        expect(itemsResp.body?.items).to.be.an('array').and.not.empty
      })

      cy.request('GET', '/api/search/items?q=scale&limit=25').then((searchResp) => {
        expect(searchResp.status).to.eq(200)
        expect(searchResp.body?.items).to.be.an('array')
      })

      cy.request('GET', `/api/scanner/candidates?query_set_id=${encodeURIComponent(querySetID as string)}`).then((candResp) => {
        expect(candResp.status).to.eq(200)
        expect(candResp.body?.candidates).to.be.an('array')
      })
    })
  })

  it('UI-SCALE-003 sustains repeated high-volume action loops without unrecoverable stalls', () => {
    cy.request('POST', '/api/test/scale/bootstrap', {
      profile: 'S3',
      seed: 7007,
    }).then((bootstrapResp) => {
      expect(bootstrapResp.status).to.eq(200)
      const querySetID = bootstrapResp.body?.query_set_id
      expect(querySetID).to.be.a('string').and.not.empty

      for (let idx = 0; idx < 12; idx += 1) {
        cy.request('GET', '/api/search/items?q=scale&limit=20').its('status').should('eq', 200)
        cy.request('GET', `/api/scanner/candidates?query_set_id=${encodeURIComponent(querySetID as string)}`).its('status').should('eq', 200)
      }
    })
  })
})
