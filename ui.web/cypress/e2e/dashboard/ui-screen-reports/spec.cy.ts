describe("UI-SCREEN-REPORTS", () => {
  function signInToReports() {
    cy.visit("/sign-in?redirect=%2Freports%2F")
    cy.get('input[name="email"]').clear().type("e2e-reports@example.com")
    cy.get('input[name="password"]').clear().type("password123")
    cy.contains("button", "Sign in").click()
    cy.location("pathname", { timeout: 15000 }).should("match", /^\/reports\/?$/)
  }

  it("UI-SCREEN-REPORTS-001 renders wishlist and pricing summary metrics", () => {
    cy.intercept("GET", "/api/profiles/active", {
      statusCode: 200,
      body: { id: "profile-reports-1" },
    }).as("activeProfile")
    cy.intercept("GET", "/api/wishlist/hits?profile_id=profile-reports-1", {
      statusCode: 200,
      body: { hits: [{ id: "h1" }, { id: "h2" }] },
    }).as("wishlistHits")
    cy.intercept("GET", "/api/pricing/stats?profile_id=profile-reports-1", {
      statusCode: 200,
      body: { min: 12, median: 24, latest: 30 },
    }).as("pricingStats")
    cy.intercept("GET", "/api/pricing/trend?profile_id=profile-reports-1", {
      statusCode: 200,
      body: { points: [{ t: "2026-01-01", v: 12 }] },
    }).as("pricingTrend")
    cy.intercept("GET", "/api/pricing/by-source?profile_id=profile-reports-1", {
      statusCode: 200,
      body: { sources: { ebay: { latest: 30 } } },
    }).as("pricingSource")

    signInToReports()
    cy.wait("@activeProfile")
    cy.wait("@wishlistHits")
    cy.wait("@pricingStats")
    cy.wait("@pricingTrend")
    cy.wait("@pricingSource")

    cy.contains("Reports").should("be.visible")
    cy.contains("Wishlist Hits").should("be.visible")
    cy.contains("Price Median").should("be.visible")
    cy.contains("$24.00").should("be.visible")
  })

  it("UI-SCREEN-REPORTS-002 exports report output deterministically", () => {
    cy.intercept("GET", "/api/profiles/active", {
      statusCode: 200,
      body: { id: "profile-reports-2" },
    })
    cy.intercept("GET", "/api/wishlist/hits?profile_id=profile-reports-2", {
      statusCode: 200,
      body: { hits: [] },
    })
    cy.intercept("GET", "/api/pricing/stats?profile_id=profile-reports-2", {
      statusCode: 200,
      body: { min: 0, median: 0, latest: 0 },
    })
    cy.intercept("GET", "/api/pricing/trend?profile_id=profile-reports-2", {
      statusCode: 200,
      body: { points: [] },
    })
    cy.intercept("GET", "/api/pricing/by-source?profile_id=profile-reports-2", {
      statusCode: 200,
      body: { sources: {} },
    })
    cy.intercept("GET", "/api/data/export/csv/items", {
      statusCode: 200,
      body: "id,title\n1,Test\n",
      headers: { "content-type": "text/csv; charset=utf-8" },
    }).as("exportCSV")

    signInToReports()
    cy.get('[data-testid="reports-export-button"]').click()
    cy.wait("@exportCSV")
    cy.get('[data-testid="reports-export-message"]')
      .should("be.visible")
      .and("contain", "Export generated")
  })

  it("UI-SCREEN-REPORTS-003 handles loading/empty/error states deterministically", () => {
    let attempts = 0
    cy.intercept("GET", "/api/profiles/active", {
      statusCode: 200,
      body: { id: "profile-reports-3" },
    })
    cy.intercept("GET", "/api/wishlist/hits?profile_id=profile-reports-3", (req) => {
      attempts += 1
      if (attempts === 1) {
        req.reply({ statusCode: 500, body: { error: "wishlist_failed" } })
        return
      }
      req.reply({ statusCode: 200, body: { hits: [] } })
    }).as("wishlistRetry")
    cy.intercept("GET", "/api/pricing/stats?profile_id=profile-reports-3", {
      delay: 1000,
      statusCode: 200,
      body: { min: 0, median: 0, latest: 0 },
    })
    cy.intercept("GET", "/api/pricing/trend?profile_id=profile-reports-3", {
      statusCode: 200,
      body: { points: [] },
    })
    cy.intercept("GET", "/api/pricing/by-source?profile_id=profile-reports-3", {
      statusCode: 200,
      body: { sources: {} },
    })

    signInToReports()
    cy.contains("Loading...").should("be.visible")
    cy.wait("@wishlistRetry")
    cy.get('[data-testid="reports-error"]').should("be.visible")
    cy.contains("button", "Retry").click()
    cy.wait("@wishlistRetry")
    cy.get('[data-testid="reports-error"]').should("not.exist")
    cy.get('[data-testid="reports-empty-state"]').should("be.visible")
  })
})
