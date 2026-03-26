describe("UI-SCREEN-HOME", () => {
  function signInToHome(redirectPath = "/dashboard") {
    cy.request("POST", "/api/test/reset", {})
    cy.request("POST", "/api/profiles", { name: "E2E Local" }).then((createResp) => {
      expect(createResp.status).to.eq(201)
      const profileId = createResp.body.id as string
      cy.request("PUT", "/api/profiles/active", { profile_id: profileId }).its("status").should("eq", 200)
    })

    cy.visit(`/sign-in?redirect=${encodeURIComponent(redirectPath)}`)
    cy.get('input[name="email"]').clear().type("e2e-home@example.com")
    cy.get('input[name="password"]').clear().type("password123")
    cy.contains("button", "Sign in").click()
    cy.location("pathname", { timeout: 15000 }).should("eq", "/dashboard")
  }

  it("UI-SCREEN-HOME-001 renders actionable priority cards with direct actions", () => {
    cy.intercept("GET", "/api/dashboard", {
      statusCode: 200,
      body: {
        new_discoveries: 4,
        wishlist_hits: 2,
        price_drops: 1,
        low_stock_discoveries: 1,
        restocks: 3,
        recently_added: ["AFX Camaro", "Mega G+ Set"],
        total_items: 200,
        total_instances: 240,
        estimated_value: 12345.67,
        cards: [
          { title: "Review discoveries", value: 4, link: "/discoveries" },
          { title: "Open pricing drops", value: 1, link: "/pricing" },
        ],
      },
    }).as("dashboardSuccess")

    signInToHome()
    cy.wait("@dashboardSuccess")

    cy.contains("Home").should("be.visible")
    cy.contains("What needs action now in your collection.").should("be.visible")
    cy.contains("What needs attention now").should("be.visible")
    cy.contains("Review discoveries").should("be.visible")
    cy.contains("Open pricing drops").should("be.visible")
    cy.contains("Recently added").should("be.visible")
    cy.contains("AFX Camaro").should("be.visible")
  })

  it("UI-SCREEN-HOME-002 renders deterministic loading and empty states", () => {
    cy.intercept("GET", "/api/dashboard", (req) => {
      req.reply({
        delay: 1200,
        statusCode: 200,
        body: {
          new_discoveries: 0,
          wishlist_hits: 0,
          price_drops: 0,
          low_stock_discoveries: 0,
          restocks: 0,
          recently_added: [],
          total_items: 0,
          total_instances: 0,
          estimated_value: 0,
          cards: [],
        },
      })
    }).as("dashboardEmpty")

    signInToHome()
    cy.contains("Loading...").should("be.visible")
    cy.wait("@dashboardEmpty")
    cy.contains("No action items right now.").should("be.visible")
    cy.contains("No recently added items yet.").should("be.visible")
  })

  it("UI-SCREEN-HOME-003 handles fetch error and supports retry + quick action routing", () => {
    let attempts = 0
    cy.intercept("GET", "/api/dashboard", (req) => {
      attempts += 1
      if (attempts === 1) {
        req.reply({
          statusCode: 500,
          body: { error: "failed_to_load_dashboard" },
        })
        return
      }
      req.reply({
        statusCode: 200,
        body: {
          new_discoveries: 1,
          wishlist_hits: 0,
          price_drops: 0,
          low_stock_discoveries: 0,
          restocks: 0,
          recently_added: ["Recovery Item"],
          total_items: 1,
          total_instances: 1,
          estimated_value: 99,
          cards: [{ title: "Review discoveries", value: 1, link: "/discoveries" }],
        },
      })
    }).as("dashboardRetry")

    signInToHome()
    cy.wait("@dashboardRetry")
    cy.contains("Dashboard unavailable").should("be.visible")
    cy.contains("button", "Retry").click()
    cy.wait("@dashboardRetry")
    cy.contains("Dashboard unavailable").should("not.exist")

    cy.contains("Review discoveries")
      .parentsUntil("div.border")
      .parent()
      .within(() => {
        cy.contains("a", "Open").click()
      })

    cy.location("pathname", { timeout: 15000 }).should("match", /^\/discoveries\/?$/)
  })

  it("UI-SCREEN-HOME-007 resolves canonical /dashboard route, root redirect, and nav target stability", () => {
    cy.intercept("GET", "/api/dashboard", {
      statusCode: 200,
      body: {
        new_discoveries: 2,
        wishlist_hits: 1,
        price_drops: 0,
        low_stock_discoveries: 0,
        restocks: 0,
        recently_added: ["Canonical Route Item"],
        total_items: 2,
        total_instances: 2,
        estimated_value: 200,
        cards: [{ title: "Review discoveries", value: 2, link: "/discoveries" }],
      },
    }).as("dashboardCanonical")

    signInToHome("/")
    cy.wait("@dashboardCanonical")
    cy.location("pathname", { timeout: 15000 }).should("eq", "/dashboard")
    cy.get('[data-testid="sidebar-nav-link-dashboard"]').should(
      "have.attr",
      "href",
      "/dashboard"
    )

    cy.visit("/")
    cy.location("pathname", { timeout: 15000 }).should("eq", "/dashboard")

    cy.visit("/inventory")
    cy.location("pathname", { timeout: 15000 }).should("match", /^\/inventory\/?$/)
    cy.get('[data-testid="active-profile-name"]', { timeout: 15000 }).should("contain", "E2E Local")
    cy.get('[data-testid="sidebar-nav-link-dashboard"]').click()
    cy.location("pathname", { timeout: 15000 }).should("eq", "/dashboard")

    cy.reload()
    cy.location("pathname", { timeout: 15000 }).should("eq", "/dashboard")
  })
})
