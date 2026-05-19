describe('general/ui-performance', () => {
  function signInToInventory() {
    cy.visit('/sign-in?redirect=%2Finventory%2F')
    cy.get('input[name="email"]').clear().type('e2e-performance@example.com')
    cy.get('input[name="password"]').clear().type('password123')
    cy.contains('button', 'Sign in').click()
    cy.location('pathname', { timeout: 15000 }).should('match', /^\/inventory\/?$/)
  }

  it('UI-PERFORMANCE-001 enforces measurable S2 interaction thresholds', () => {
    cy.request('POST', '/api/test/scale/bootstrap', { profile: 'S2', seed: 4242 }).then((bootstrapResp) => {
      expect(bootstrapResp.status).to.eq(200)
      const querySetID = bootstrapResp.body?.query_set_id as string
      expect(String(querySetID).trim()).not.to.equal('')

      const itemDurations: number[] = []
      const searchDurations: number[] = []
      const scannerDurations: number[] = []

      for (let i = 0; i < 5; i += 1) {
        cy.request('GET', '/api/items?status=active').then((resp) => {
          expect(resp.status).to.eq(200)
          itemDurations.push(resp.duration)
        })

        cy.request('GET', '/api/search/items?q=scale&limit=20').then((resp) => {
          expect(resp.status).to.eq(200)
          searchDurations.push(resp.duration)
        })

        cy.request('GET', `/api/scanner/candidates?query_set_id=${encodeURIComponent(querySetID)}`).then((resp) => {
          expect(resp.status).to.eq(200)
          scannerDurations.push(resp.duration)
        })
      }

      cy.then(() => {
        const median = (values: number[]) => {
          const sorted = [...values].sort((a, b) => a - b)
          return sorted[Math.floor(sorted.length / 2)]
        }

        expect(median(itemDurations)).to.be.lessThan(1500)
        expect(median(searchDurations)).to.be.lessThan(1500)
        expect(median(scannerDurations)).to.be.lessThan(1500)
      })
    })
  })

  it('UI-PERFORMANCE-002 keeps UI responsive during delayed large-list loading transition', () => {
    const bulkItems = Array.from({ length: 900 }, (_, index) => ({
      id: `perf-item-${index + 1}`,
      part_number: `PERF-${index + 1}`,
      title: `Performance Item ${index + 1}`,
      status: 'todo',
      category: 'feature',
    }))

    cy.intercept('GET', '/api/items*', (req) => {
      req.reply((res) => {
        res.delay = 1200
        res.send({
          statusCode: 200,
          body: { items: bulkItems },
        })
      })
    }).as('itemsDelayed')

    signInToInventory()
    cy.get('[data-testid="inventory-loading"]').should('be.visible')
    cy.wait('@itemsDelayed')
    cy.get('[data-testid="inventory-loading"]').should('not.exist')
    cy.get('[data-testid="inventory-table-surface"]').should('exist')
    cy.get('table[data-slot="table"]').should('exist')
    cy.contains('Performance Item 1').should('exist')
    cy.contains('500').should('not.exist')
  })
})
