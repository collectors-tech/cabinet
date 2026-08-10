describe('add-integration handoff', () => {
  const bonzaProviderID = 'au-webshop-bonzaslotcars-com-au'
  const frontlineProviderID = 'au-webshop-frontlinehobbies-com-au'
  let sessionToken = ''

  beforeEach(() => {
    cy.viewport(1440, 1000)
    cy.e2eReset()
    cy.e2eBootstrap({ minimalProfile: true }).then((state) => {
      sessionToken = state.session_token
      cy.useBootstrappedProfile(state.profile_id, state.profile_name, {
        path: '/integrations',
      })
    })
  })

  function expectFocusInsideSetupDialog(providerName: string) {
    cy.get('[role="dialog"]:visible')
      .should('have.length', 1)
      .should('be.visible')
      .and('contain.text', providerName)
    cy.focused().should(($focused) => {
      expect($focused.closest('[role="dialog"]')).to.have.length(1)
    })
  }

  function expectPersistedProvider(
    profileID: string,
    providerID: string,
    pollingInterval: string
  ) {
    const providerSlug = providerID.replace(/[^a-z0-9]+/gi, '_').toLowerCase()
    cy.request(`/api/profiles/${profileID}/settings`).then((response) => {
      expect(response.status).to.eq(200)
      expect(response.body.settings).to.include({
        [`integration.${providerSlug}.enabled`]: 'true',
        crawl_interval_minutes: pollingInterval,
      })
    })
    cy.request(`/api/profiles/${profileID}/integration-instances`).then(
      (response) => {
        expect(response.status).to.eq(200)
        const instance = response.body.instances?.find(
          (candidate: { provider_id?: string }) =>
            candidate.provider_id === providerID
        )
        expect(
          response.body.instances?.filter(
            (candidate: { provider_id?: string }) =>
              candidate.provider_id === providerID
          )
        ).to.have.length(1)
        expect(instance, `${providerID} integration instance`).to.include({
          provider_id: providerID,
          enabled: true,
        })
        expect(instance.config).to.include({
          crawl_interval_minutes: pollingInterval,
        })
      }
    )
  }

  it('#2062: hands pointer and keyboard selections into persistent provider setup without configuring on cancel', () => {
    cy.request('/api/profiles/active').then((activeProfileResponse) => {
      const profileID = String(activeProfileResponse.body.id)

      cy.get('[data-testid="integrations-header-add"]').click()
      cy.get('[data-testid="integrations-provider-selector"]')
        .should('be.visible')
        .find('[data-slot="dialog-close"]')
        .click()
      cy.get('[data-testid="integrations-provider-selector"]').should(
        'not.exist'
      )
      cy.get('body').should('not.have.css', 'pointer-events', 'none')
      cy.get('[data-testid="integrations-header-add"]').should('be.focused')
      cy.request(`/api/profiles/${profileID}/settings`).then((response) => {
        expect(response.body.settings).not.to.have.property(
          'integration.au_webshop_bonzaslotcars_com_au.enabled'
        )
        expect(response.body.settings).not.to.have.property(
          'integration.au_webshop_frontlinehobbies_com_au.enabled'
        )
      })
      cy.request(`/api/profiles/${profileID}/integration-instances`).then(
        (response) => {
          expect(response.status).to.eq(200)
          expect(response.body.instances ?? []).to.deep.equal([])
        }
      )

      cy.get('[data-testid="integrations-header-add"]').click()
      cy.get('[data-testid="integrations-provider-selector-search"]').type(
        'frontline'
      )
      cy.get(
        `[data-testid="integrations-provider-selector-option-${frontlineProviderID}"]`
      )
        .scrollIntoView()
        .should('be.visible')
        .focus()
      cy.press(Cypress.Keyboard.Keys.SPACE)
      expectFocusInsideSetupDialog('frontlinehobbies.com.au')
      cy.contains('button', 'Cancel').click()
      cy.get('[role="dialog"]').should('not.exist')
      cy.get('body').should('not.have.css', 'pointer-events', 'none')
      cy.request(`/api/profiles/${profileID}/integration-instances`).then(
        (response) => {
          expect(response.body.instances ?? []).to.deep.equal([])
        }
      )
      cy.request(`/api/profiles/${profileID}/settings`).then((response) => {
        expect(response.body.settings).not.to.have.property(
          'integration.au_webshop_frontlinehobbies_com_au.enabled'
        )
      })

      cy.get('[data-testid="integrations-header-add"]').click()
      cy.get('[data-testid="integrations-provider-selector-search"]').type(
        'bonza'
      )
      cy.get(
        `[data-testid="integrations-provider-selector-option-${bonzaProviderID}"]`
      )
        .scrollIntoView()
        .should('be.visible')
        .click()
      expectFocusInsideSetupDialog('bonzaslotcars.com.au')
      cy.get('[data-testid="provider-schema-field-base_domain"]').should(
        'have.value',
        'bonzaslotcars.com.au'
      )
      cy.get(
        '[data-testid="provider-schema-field-crawl_interval_minutes"]'
      )
        .clear()
        .type('720')
      cy.contains('button', 'Save Integration').click()
      cy.get('[role="dialog"]').should('not.exist')
      cy.get('body').should('not.have.css', 'pointer-events', 'none')
      cy.get(`[data-testid="provider-row-${bonzaProviderID}"]`)
        .should('be.visible')
        .within(() => {
          cy.contains('button', 'Edit').should('be.visible')
        })
      expectPersistedProvider(profileID, bonzaProviderID, '720')

      cy.intercept('PUT', '/api/profiles/*/integration-instances').as(
        'updateBonzaIntegrationInstance'
      )
      cy.get(`[data-testid="provider-open-${bonzaProviderID}"]`).click()
      expectFocusInsideSetupDialog('bonzaslotcars.com.au')
      cy.get(
        '[data-testid="provider-schema-field-crawl_interval_minutes"]'
      ).should('have.value', '720')
      cy.get(
        '[data-testid="provider-schema-field-crawl_interval_minutes"]'
      )
        .clear()
        .type('721')
      cy.contains('button', 'Save Integration').click()
      cy.wait('@updateBonzaIntegrationInstance')
        .its('response.statusCode')
        .should('eq', 200)
      cy.get('[role="dialog"]').should('not.exist')
      cy.get('body').should('not.have.css', 'pointer-events', 'none')
      expectPersistedProvider(profileID, bonzaProviderID, '721')

      cy.get('[data-testid="integrations-header-add"]').click()
      cy.get('[data-testid="integrations-provider-selector-search"]')
        .should('be.focused')
        .type('frontline')
      cy.get(
        `[data-testid="integrations-provider-selector-option-${frontlineProviderID}"]`
      )
        .scrollIntoView()
        .should('be.visible')
        .focus()
      cy.press(Cypress.Keyboard.Keys.ENTER)
      expectFocusInsideSetupDialog('frontlinehobbies.com.au')
      cy.get('[data-testid="provider-schema-field-base_domain"]').should(
        'have.value',
        'frontlinehobbies.com.au'
      )
      cy.get(
        '[data-testid="provider-schema-field-crawl_interval_minutes"]'
      )
        .clear()
        .type('720')
      cy.contains('button', 'Save Integration').click()
      cy.get('[role="dialog"]').should('not.exist')
      cy.get('body').should('not.have.css', 'pointer-events', 'none')
      cy.get(`[data-testid="provider-row-${bonzaProviderID}"]`).should(
        'be.visible'
      )
      cy.get(`[data-testid="provider-row-${frontlineProviderID}"]`)
        .should('be.visible')
        .within(() => {
          cy.contains('button', 'Edit').should('be.visible')
        })
      expectPersistedProvider(profileID, frontlineProviderID, '720')

      cy.request({
        method: 'POST',
        url: '/api/scanner/query-sets',
        body: {
          name: 'Issue 2062 provider handoff',
          keywords: ['AFX'],
          provider_scope: ['bonzaslotcars', 'frontlinehobbies'],
          enabled: true,
        },
      }).its('status').should('eq', 201)
      cy.request('/api/scanner/query-sets').then((response) => {
        expect(response.status).to.eq(200)
        const querySet = response.body.query_sets?.find(
          (candidate: { name?: string }) =>
            candidate.name === 'Issue 2062 provider handoff'
        )
        expect(querySet?.provider_scope).to.have.members([
          'bonzaslotcars',
          'frontlinehobbies',
        ])
      })

      const companionOrigin =
        'chrome-extension://aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa'
      const companionDevice = 'issue-2062-cypress'
      const extensionHeaders = {
        Origin: companionOrigin,
        'X-Cabinet-Companion-Device': companionDevice,
      }
      cy.request({
        method: 'POST',
        url: '/api/companion/pairing/requests',
        headers: extensionHeaders,
        body: {
          device_id: companionDevice,
          device_name: 'Issue 2062 Cypress Chrome',
          protocol_version: '1',
          capabilities: ['modules:read'],
        },
      }).then((pairingResponse) => {
        expect(pairingResponse.status).to.eq(201)
        cy.request({
          method: 'POST',
          url: '/api/companion/pairing/approvals',
          headers: {
            Origin: new URL(String(Cypress.config('baseUrl'))).origin,
            'X-Cabinet-Session': sessionToken,
          },
          body: {
            request_id: pairingResponse.body.request_id,
            profile_id: profileID,
          },
        }).its('status').should('eq', 200)
        cy.request({
          method: 'POST',
          url: '/api/companion/pairing/exchanges',
          headers: extensionHeaders,
          body: {
            request_id: pairingResponse.body.request_id,
            exchange_secret: pairingResponse.body.exchange_secret,
            device_id: companionDevice,
            protocol_version: pairingResponse.body.protocol_version,
          },
        }).then((exchangeResponse) => {
          expect(exchangeResponse.status).to.eq(200)
          cy.request({
            method: 'GET',
            url: '/api/companion/modules',
            headers: {
              ...extensionHeaders,
              Authorization: `Bearer ${exchangeResponse.body.credential}`,
            },
          }).then((modulesResponse) => {
            expect(modulesResponse.status).to.eq(200)
            const moduleIDs = modulesResponse.body.modules.map(
              (module: { id: string }) => module.id
            )
            expect(moduleIDs).to.include.members([
              'bonzaslotcars-search-capture',
              'frontlinehobbies-search-capture',
            ])
          })
        })
      })

      cy.contains('button', 'Cards').click()
      cy.get(`[data-testid="provider-card-${bonzaProviderID}"]`).should(
        'be.visible'
      )
      cy.get(`[data-testid="provider-card-${frontlineProviderID}"]`).should(
        'be.visible'
      )
      cy.get(`[data-testid="provider-open-${frontlineProviderID}"]`).click()
      expectFocusInsideSetupDialog('frontlinehobbies.com.au')
      cy.get(
        '[data-testid="provider-schema-field-crawl_interval_minutes"]'
      ).should('have.value', '720')
    })
  })
})
