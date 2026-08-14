describe('Telegram local connector real-runtime setup', () => {
  const runtimeFixtureEnabled =
    Cypress.env('telegramRuntimeFixture') === true ||
    Cypress.env('telegramRuntimeFixture') === 'true'

  beforeEach(() => {
    cy.viewport(1440, 1000)
    cy.e2eReset()
    cy.e2eBootstrap({ minimalProfile: true }).then((state) => {
      cy.wrap(state.profile_id).as('telegramProfileID')
      cy.useBootstrappedProfile(state.profile_id, state.profile_name, {
        path: '/integrations',
      })
    })
  })

  const waitForPairing = (
    profileID: string,
    attempts = 60
  ): Cypress.Chainable =>
    cy
      .request(`/api/telegram/connection/status?profile_id=${profileID}`)
      .then((response) => {
        if (response.body.paired === true) {
          return cy.wrap(response.body)
        }
        if (attempts <= 1) {
          throw new Error(
            `Telegram runtime fixture did not complete pairing: ${JSON.stringify(response.body)}`
          )
        }
        return cy.wait(250).then(() => waitForPairing(profileID, attempts - 1))
      })

  ;(runtimeFixtureEnabled ? it : it.skip)(
    '#2085: completes write-only bot validation, explicit webhook resolution, and private pairing against running Cabinet',
    { retries: 0 },
    () => {
      const fixtureToken = String(
        Cypress.env('telegramBotToken') ?? '123456:test-runtime-token'
      )
      const fixtureControlURL = String(
        Cypress.env('telegramFixtureControlURL') ?? ''
      ).replace(/\/$/, '')
      expect(fixtureControlURL, 'controlled Telegram fixture URL').not.to.eq('')

      cy.request('POST', `${fixtureControlURL}/control/reset`)
      cy.request('POST', `${fixtureControlURL}/control/release`)

      cy.get('[data-testid="integrations-header-add"]').click()
      cy.get('[data-testid="integrations-provider-selector-search"]').type(
        'Telegram'
      )
      cy.get(
        '[data-testid="integrations-provider-selector-option-telegram"]'
      ).click()

      cy.get('[data-testid="telegram-local-polling-setup"]')
        .should('contain.text', 'BotFather')
        .and('contain.text', 'No public webhook')
      cy.get('[data-testid="telegram-bot-token"]')
        .scrollIntoView()
        .should('be.visible')
        .focus()
        .should('have.focus')
        .type(fixtureToken, { log: false })
      cy.get('[data-testid="telegram-test-connection"]').click()
      cy.get('[data-testid="telegram-bot-identity"]')
        .should('be.visible')
        .and('contain.text', '@cabinet_fixture_bot')
      cy.get('[data-testid="telegram-bot-token"]').should('have.value', '')

      cy.get('[data-testid="telegram-resolve-webhook"]')
        .should('be.visible')
        .click()
      cy.get('[data-testid="telegram-resolve-webhook"]').should('not.exist')

      cy.get('[data-testid="telegram-create-pairing-code"]').click()
      cy.get('[data-testid="telegram-pairing-code"]')
        .should('be.visible')
        .invoke('text')
        .then((text) => {
          const command = text.trim()
          expect(command).to.match(/^\/start CAB-[A-Z0-9]{4}-[A-Z0-9]{4}$/)
          cy.request({
            method: 'POST',
            url: `${fixtureControlURL}/control/updates`,
            body: {
              update_id: 7001,
              sender_id: 4444,
              chat_id: 4444,
              chat_type: 'private',
              text: command,
            },
          })
        })
      cy.get('[data-testid="telegram-pairing-expiry"]').should(
        'contain.text',
        'expires'
      )

      cy.get('@telegramProfileID').then((profileValue) => {
        const profileID = String(profileValue)
        waitForPairing(profileID).then((status) => {
          expect(status).to.include({
            profile_id: profileID,
            bot_token_present: true,
            bot_validated: true,
            paired: true,
            sender_id: '4444',
            chat_id: '4444',
            offset: 7002,
            credential_returned: false,
            transport: 'long_polling',
            public_listener: false,
          })
          expect(JSON.stringify(status)).not.to.contain(fixtureToken)
        })
      })

      cy.get('[data-testid="telegram-refresh-status"]').click()
      cy.get('[data-testid="telegram-local-polling-setup"]')
        .should('contain.text', 'Paired private chat ending in 4444')
        .and('contain.text', 'Transport: outbound long polling')
    }
  )
})
