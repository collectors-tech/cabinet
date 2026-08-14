
describe('chats/agent-skill-durable-preview-api', () => {
  beforeEach(() => {
    cy.e2eReset()
    cy.e2eBootstrap()
  })

  it('AGENT-SKILLS-REGISTRY-013/#2087 confirms an opaque generic Agent Skill preview exactly once', () => {
    const privateNote = 'private-note-2087-cypress'

    cy.request('POST', '/api/agent/skills/preview', {
      profile_id: 'e2e-profile-001',
      skill_id: 'cabinet.wishlist.create_entry',
      source_surface: 'chats.main',
      source_channel: 'in-app',
      source_thread_id: 'thread-2087-cypress',
      source_message_id: 'message-2087-cypress',
      parameters: {
        part_number: 'WISH-2087-CYPRESS',
        title: 'Durable Agent Cypress Wishlist Item',
        brand: 'AFX',
        category: 'Slot Cars',
        notes: privateNote,
      },
    }).then((previewResponse) => {
      expect(previewResponse.status).to.eq(200)
      expect(previewResponse.body.preview_id).to.match(/^asp_[a-f0-9]+$/)
      expect(previewResponse.body.preview_status).to.eq('previewed')
      expect(previewResponse.body.confirmation_required).to.eq(true)
      expect(previewResponse.body.mutation_applied).to.eq(false)
      expect(JSON.stringify(previewResponse.body)).not.to.include(privateNote)

      const previewId = previewResponse.body.preview_id as string
      cy.request('POST', '/api/agent/skills/apply', {
        profile_id: 'e2e-profile-001',
        preview_id: previewId,
        confirm: true,
      }).then((applyResponse) => {
        expect(applyResponse.status).to.eq(200)
        expect(applyResponse.body.preview_id).to.eq(previewId)
        expect(applyResponse.body.preview_status).to.eq('applied')
        expect(applyResponse.body.mutation_applied).to.eq(true)
        expect(applyResponse.body.target.wishlist_entry_id).to.be.a('string')
        expect(JSON.stringify(applyResponse.body)).not.to.include(privateNote)
      })

      cy.request({
        method: 'POST',
        url: '/api/agent/skills/apply',
        failOnStatusCode: false,
        body: {
          profile_id: 'e2e-profile-001',
          preview_id: previewId,
          confirm: true,
        },
      }).then((replayResponse) => {
        expect(replayResponse.status).to.eq(409)
        expect(replayResponse.body.error).to.eq(
          'agent_skill_preview_already_applied'
        )
      })

      cy.request('/api/items?status=wishlist').then((itemsResponse) => {
        expect(itemsResponse.status).to.eq(200)
        const matches = itemsResponse.body.items.filter(
          (item: { part_number?: string }) =>
            item.part_number === 'WISH-2087-CYPRESS'
        )
        expect(matches).to.have.length(1)
      })
    })
  })

  it('AGENT-SKILLS-REGISTRY-013/#2087 cancellation prevents later apply', () => {
    cy.request('POST', '/api/agent/skills/preview', {
      profile_id: 'e2e-profile-001',
      skill_id: 'cabinet.wishlist.create_entry',
      source_surface: 'chats.main',
      source_channel: 'in-app',
      source_thread_id: 'thread-2087-cancel',
      source_message_id: 'message-2087-cancel',
      parameters: {
        part_number: 'WISH-2087-CYPRESS-CANCEL',
        title: 'Cancelled Durable Agent Wishlist Item',
      },
    }).then((previewResponse) => {
      const previewId = previewResponse.body.preview_id as string

      cy.request('POST', '/api/agent/skills/cancel', {
        profile_id: 'e2e-profile-001',
        preview_id: previewId,
      }).then((cancelResponse) => {
        expect(cancelResponse.status).to.eq(200)
        expect(cancelResponse.body.preview_id).to.eq(previewId)
        expect(cancelResponse.body.preview_status).to.eq('cancelled')
        expect(cancelResponse.body.mutation_applied).to.eq(false)
      })

      cy.request({
        method: 'POST',
        url: '/api/agent/skills/apply',
        failOnStatusCode: false,
        body: {
          profile_id: 'e2e-profile-001',
          preview_id: previewId,
          confirm: true,
        },
      }).then((applyResponse) => {
        expect(applyResponse.status).to.eq(409)
        expect(applyResponse.body.error).to.eq(
          'agent_skill_preview_cancelled'
        )
        expect(applyResponse.body.recoverable).to.eq(true)
      })
    })
  })
})
