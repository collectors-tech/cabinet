describe('chats/local-agent-session-bridge', () => {
  it('LOCAL-AUTH-008/#2093 carries a memory-only server session to protected Agent requests', () => {
    cy.viewport(1400, 900)
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.request('PUT', '/api/profiles/active', {
      profile_id: 'e2e-profile-001',
    })

    let sessionToken = ''
    cy.intercept('POST', '/api/auth/local/session', (request) => {
      expect(request.body).to.deep.eq({ profile_id: 'e2e-profile-001' })
      request.continue((response) => {
        expect(response.statusCode).to.eq(200)
        sessionToken = String(response.body.session_token || '')
        expect(sessionToken).to.have.length.greaterThan(40)
      })
    }).as('localServerSession')
    cy.intercept('POST', '/api/chat/messages').as('naturalAgentMessage')

    cy.visit('/sign-in?redirect=%2Finventory%2F')
    cy.contains('button', 'Open local workspace').click()
    cy.wait('@localServerSession')
    cy.location('pathname', { timeout: 15000 }).should(
      'match',
      /^\/inventory\/?$/
    )

    cy.get('[data-testid="active-profile-status"]', { timeout: 20000 }).should(
      'not.contain',
      'Loading profiles'
    )
    cy.get('[data-testid="shell-chat-toggle"]').scrollIntoView().click()
    cy.get('[data-testid="shell-assistant-thread-id"]', {
      timeout: 30000,
    }).should(($threadId) => {
      expect($threadId.text().trim()).not.to.eq('')
      expect($threadId.text().trim()).not.to.eq('bootstrapping')
    })
    const prompt = 'Find the local owner in Cabinet users'
    cy.get('[data-testid="shell-assistant-compose-input"]')
      .scrollIntoView()
      .type(prompt)
    cy.get('[data-testid="shell-assistant-send-button"]').click()

    cy.wait('@naturalAgentMessage').then(({ request, response }) => {
      expect(sessionToken).not.to.eq('')
      expect(request.headers['x-cabinet-session']).to.eq(sessionToken)
      expect(request.body.content).to.eq(prompt)
      expect(request.body.agent_context?.admin_session).not.to.eq(sessionToken)
      expect(response?.statusCode).to.be.oneOf([200, 201])
      expect(JSON.stringify(response?.body)).not.to.contain(sessionToken)
    })

    cy.location('href').should(($href) => {
      expect($href).not.to.contain(sessionToken)
    })
    cy.get('html').should(($html) => {
      expect($html.text()).not.to.contain(sessionToken)
    })
    cy.getAllCookies().then((cookies) => {
      expect(JSON.stringify(cookies)).not.to.contain(sessionToken)
    })
    cy.window().then((window) => {
      expect(JSON.stringify(window.localStorage)).not.to.contain(sessionToken)
      expect(JSON.stringify(window.sessionStorage)).not.to.contain(sessionToken)
    })
  })

  it('LOCAL-AUTH-008/#2093 revokes the server session before UI sign-out navigation', () => {
    cy.e2eReset()
    cy.e2eBootstrap()
    cy.e2eSetSetupState('present')
    cy.request('PUT', '/api/profiles/active', {
      profile_id: 'e2e-profile-001',
    })

    let sessionToken = ''
    cy.intercept('POST', '/api/auth/local/session', (request) => {
      request.continue((response) => {
        sessionToken = String(response.body.session_token || '')
      })
    }).as('localServerSession')
    cy.intercept('POST', '/api/auth/session/lock').as('serverSessionLock')

    cy.visit('/sign-in?redirect=%2Fdashboard')
    cy.contains('button', 'Open local workspace').click()
    cy.wait('@localServerSession')
    cy.location('pathname').should('eq', '/dashboard')
    cy.get('[data-testid="profile-dropdown-trigger"]:visible')
      .first()
      .scrollIntoView()
      .click()
    cy.contains('[data-slot="dropdown-menu-item"]', 'Sign out').click()
    cy.contains('[role="alertdialog"] button', 'Sign out').click()

    cy.wait('@serverSessionLock').then(({ request, response }) => {
      expect(sessionToken).to.have.length.greaterThan(40)
      expect(request.headers['x-cabinet-session']).to.eq(sessionToken)
      expect(JSON.stringify(request.body)).not.to.contain(sessionToken)
      expect(JSON.stringify(response?.body)).not.to.contain(sessionToken)
    })
    cy.location('pathname').should('eq', '/sign-in')
  })
})
