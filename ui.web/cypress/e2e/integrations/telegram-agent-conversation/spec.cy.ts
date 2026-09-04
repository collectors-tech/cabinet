describe('integrations/telegram-agent-conversation', () => {
  const profileID = 'e2e-profile-001'
  const peerID = 777001
  const fixtureEnabled =
    Cypress.env('telegramRuntimeFixture') === true ||
    Cypress.env('telegramRuntimeFixture') === 'true'

  beforeEach(() => {
    cy.e2eReset()
    cy.e2eBootstrap({ minimalProfile: true })
  })

  const waitForOffset = (
    wanted: number,
    attempts = 80
  ): Cypress.Chainable<Record<string, unknown>> =>
    cy
      .request(`/api/telegram/connection/status?profile_id=${profileID}`)
      .then((response) => {
        if (Number(response.body.offset) >= wanted) {
          return cy.wrap(response.body)
        }
        if (attempts <= 1) {
          throw new Error(`Telegram fixture did not reach offset ${wanted}`)
        }
        return cy.wait(200).then(() => waitForOffset(wanted, attempts - 1))
      })

  it('TELEGRAM-AGENT-CONVERSATION-002/#2086 rejects forged public HTTP requests even with paired identifiers', () => {
    cy.request('PUT', `/api/profiles/${profileID}/settings`, {
      settings: {
        'telegram.catalog_capture.sender_id': String(peerID),
        'telegram.catalog_capture.chat_id': String(peerID),
      },
    })
    cy.request({
      method: 'POST',
      url: '/api/telegram/agent-text',
      failOnStatusCode: false,
      body: {
        sender_id: String(peerID),
        chat_id: String(peerID),
        chat_type: 'private',
        message_id: '8801',
        text: 'show my inventory',
      },
    }).then((response) => {
      expect(response.status).to.eq(403)
      expect(response.body.error).to.eq('telegram_connector_only')
    })
  })

  ;(fixtureEnabled ? it : it.skip)(
    'TELEGRAM-AGENT-CONVERSATION-001/003/#2086 routes natural getUpdates text through the shared planner and applies its opaque callback once',
    { retries: 0 },
    () => {
      const token = String(
        Cypress.env('telegramBotToken') ?? '123456:test-runtime-token'
      )
      const fixtureURL = String(
        Cypress.env('telegramFixtureControlURL') ?? ''
      ).replace(/\/$/, '')
      expect(fixtureURL).not.to.eq('')

      cy.request('POST', `${fixtureURL}/control/reset`)

      cy.request({
        method: 'POST',
        url: '/api/telegram/connection/test',
        failOnStatusCode: false,
        body: { profile_id: profileID, bot_token: token },
      }).then((tested) => {
        if (tested.status === 409) {
          cy.request('POST', '/api/telegram/connection/resolve-webhook', {
            profile_id: profileID,
          })
        }
      })
      cy.request('POST', '/api/telegram/pairing-codes', {
        profile_id: profileID,
      }).then((pairing) => {
        cy.request('POST', `${fixtureURL}/control/updates`, {
          update_id: 8801,
          sender_id: peerID,
          chat_id: peerID,
          chat_type: 'private',
          text: `/start ${pairing.body.code}`,
        })
      })
      cy.request('POST', `${fixtureURL}/control/release`)
      waitForOffset(8802)

      cy.request('POST', `${fixtureURL}/control/hold`)
      cy.request('POST', `${fixtureURL}/control/updates`, {
        update_id: 8802,
        sender_id: peerID,
        chat_id: peerID,
        chat_type: 'private',
        text: 'add TG-E2E-2086 to inventory',
      })
      cy.request('POST', `${fixtureURL}/control/updates`, {
        update_id: 8802,
        sender_id: peerID,
        chat_id: peerID,
        chat_type: 'private',
        text: 'add TG-E2E-2086 to inventory',
      })
      cy.request('POST', `${fixtureURL}/control/release`)
      waitForOffset(8803)

      cy.request(`/api/chat/threads?profile_id=${profileID}`).then((threads) => {
        const telegramThreads = threads.body.threads.filter(
          (thread: { metadata?: { kind?: string } }) =>
            thread.metadata?.kind === 'telegram_agent_conversation'
        )
        expect(telegramThreads).to.have.length(1)
        const threadID = telegramThreads[0].id as string
        cy.request(
          `/api/chat/messages?profile_id=${profileID}&thread_id=${threadID}`
        ).then((messages) => {
          expect(
            messages.body.messages.filter(
              (message: { role: string }) => message.role === 'user'
            )
          ).to.have.length(1)
        })
        cy.request(
          `/api/chat/workflow-runs?profile_id=${profileID}&thread_id=${threadID}`
        ).then((runs) => {
          const previewRuns = runs.body.runs.filter(
            (run: { result?: { preview_result?: { preview_id?: string } } }) =>
              run.result?.preview_result?.preview_id
          )
          expect(previewRuns).to.have.length(1)
          const previewID = String(
            previewRuns[0]?.result.preview_result.preview_id ?? ''
          )
          expect(previewID).to.match(/^asp_[a-f0-9]{32}$/)
          cy.request('POST', `${fixtureURL}/control/hold`)
          cy.request('POST', `${fixtureURL}/control/updates`, {
            update_id: 8803,
            sender_id: peerID,
            chat_id: peerID,
            chat_type: 'private',
            message_id: 991,
            callback_query_id: 'callback-8803',
            callback_data: `${previewID}:apply`,
          })
          cy.request('POST', `${fixtureURL}/control/updates`, {
            update_id: 8803,
            sender_id: peerID,
            chat_id: peerID,
            chat_type: 'private',
            message_id: 991,
            callback_query_id: 'callback-8803',
            callback_data: `${previewID}:apply`,
          })
          cy.request('POST', `${fixtureURL}/control/release`)
        })
      })
      waitForOffset(8804)
      cy.request('/api/items').then((items) => {
        expect(
          items.body.items.filter(
            (item: { part_number?: string }) =>
              item.part_number === 'TG-E2E-2086'
          )
        ).to.have.length(1)
      })
      cy.request(`${fixtureURL}/control/status`).then((status) => {
        expect(status.body.counts.sendMessage).to.eq(2)
        expect(status.body.counts.answerCallbackQuery).to.eq(1)
        expect(status.body.counts.editMessageText).to.eq(1)
      })
    }
  )
})
