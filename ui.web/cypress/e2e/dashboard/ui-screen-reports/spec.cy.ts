describe("UI-SCREEN-REPORTS", () => {
  function signInToReports(profileId: string) {
    cy.stubLocalServerSession(profileId)
    cy.visit("/sign-in?redirect=%2Freports%2F")
    cy.get("body").then(($body) => {
      if ($body.find('input[name="email"]').length > 0) {
        cy.get('input[name="email"]').clear().type("e2e-reports@example.com")
        cy.get('input[name="password"]').clear().type("password123")
        cy.contains("button", "Sign in").click()
        return
      }

      cy.contains("button", "Open local workspace").click()
    })
    cy.location("pathname", { timeout: 15000 }).should("match", /^\/reports\/?$/)
  }

  function reportMetric(label: string) {
    return cy.contains(".grid [data-slot='card-title']", label).parents("[data-slot='card']")
  }

  function stubReports(profileId: string) {
    cy.intercept("GET", "/api/profiles/active", {
      statusCode: 200,
      body: { id: profileId },
    })
    cy.intercept("GET", `/api/wishlist/hits?profile_id=${profileId}`, {
      statusCode: 200,
      body: { hits: [{ id: "h1" }] },
    })
    cy.intercept("GET", `/api/pricing/stats?profile_id=${profileId}`, {
      statusCode: 200,
      body: { min: 12, median: 24, latest: 30 },
    })
    cy.intercept("GET", `/api/pricing/trend?profile_id=${profileId}`, {
      statusCode: 200,
      body: { points: [{ t: "2026-01-01", v: 12 }] },
    })
    cy.intercept("GET", `/api/pricing/by-source?profile_id=${profileId}`, {
      statusCode: 200,
      body: { sources: { ebay: { latest: 30 } } },
    })
  }

  function assertReportsHeaderActionsStayInsideViewport() {
    cy.get('[data-testid="reports-header-title"]').then(($title) => {
      const titleRect = $title[0].getBoundingClientRect()

      cy.get('[data-testid="reports-global-header-actions"]').then(($actions) => {
        const actionsRect = $actions[0].getBoundingClientRect()
        const viewportWidth = Cypress.config("viewportWidth")
        const titleIsCrowded = $title.attr("data-crowded") === "true"

        expect(actionsRect.left).to.be.greaterThan(0)
        expect(actionsRect.right).to.be.lessThan(viewportWidth)
        if (titleIsCrowded) {
          cy.wrap($title)
            .should("have.attr", "aria-hidden", "true")
            .parents('[data-crowded="true"]')
            .should("have.css", "opacity", "0")
          return
        }

        expect(titleRect.right).to.be.lessThan(actionsRect.left - 8)
      })
    })
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

    signInToReports("profile-reports-1")
    cy.wait("@activeProfile")
    cy.wait("@wishlistHits")
    cy.wait("@pricingStats")
    cy.wait("@pricingTrend")
    cy.wait("@pricingSource")

    cy.get('[data-testid="reports-header-title"]').should("exist")
    cy.contains("Wishlist Hits").should("be.visible")
    cy.contains("Price Median").should("be.visible")
    cy.contains("$24.00").should("be.visible")
  })

  it("UI-SCREEN-REPORTS-005 places whole-page actions in the global header without duplicate page title content", () => {
    cy.intercept("GET", "/api/profiles/active", {
      statusCode: 200,
      body: { id: "profile-reports-placement" },
    })
    cy.intercept("GET", "/api/wishlist/hits?profile_id=profile-reports-placement", {
      statusCode: 200,
      body: { hits: [{ id: "h1" }] },
    })
    cy.intercept("GET", "/api/pricing/stats?profile_id=profile-reports-placement", {
      statusCode: 200,
      body: { min: 12, median: 24, latest: 30 },
    })
    cy.intercept("GET", "/api/pricing/trend?profile_id=profile-reports-placement", {
      statusCode: 200,
      body: { points: [{ t: "2026-01-01", v: 12 }] },
    })
    cy.intercept("GET", "/api/pricing/by-source?profile_id=profile-reports-placement", {
      statusCode: 200,
      body: { sources: { ebay: { latest: 30 } } },
    })

    signInToReports("profile-reports-placement")

    cy.get('[data-testid="reports-global-header-actions"]')
      .should("be.visible")
      .within(() => {
        cy.get('[data-testid="reports-refresh-button"]').should("be.visible")
        cy.get('[data-testid="reports-export-button"]').should("be.visible")
      })
    cy.get("main").find("h1").should("not.exist")
    cy.contains("main p", "Wishlist and pricing analytics with export-ready snapshots.").should(
      "not.exist"
    )
  })

  it("UI-SCREEN-REPORTS-006 keeps header actions non-overlapping on desktop and narrow screens", () => {
    stubReports("profile-reports-header-overflow")

    cy.viewport(1280, 720)
    signInToReports("profile-reports-header-overflow")
    assertReportsHeaderActionsStayInsideViewport()

    cy.viewport(390, 720)
    assertReportsHeaderActionsStayInsideViewport()
    cy.get('[data-testid="reports-refresh-button"]').should("be.visible")
    cy.get('[data-testid="reports-export-button"]').should("be.visible")
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

    signInToReports("profile-reports-2")
    cy.get('[data-testid="reports-export-button"]').click()
    cy.wait("@exportCSV")
    cy.get('[data-testid="reports-export-message"]')
      .should("be.visible")
      .and("contain", "Export generated")
  })

  it("UI-SCREEN-REPORTS-004 reports export failures deterministically", () => {
    cy.intercept("GET", "/api/profiles/active", {
      statusCode: 200,
      body: { id: "profile-reports-export-failure" },
    })
    cy.intercept(
      "GET",
      "/api/wishlist/hits?profile_id=profile-reports-export-failure",
      {
        statusCode: 200,
        body: { hits: [{ id: "h1" }] },
      }
    )
    cy.intercept(
      "GET",
      "/api/pricing/stats?profile_id=profile-reports-export-failure",
      {
        statusCode: 200,
        body: { min: 10, median: 20, latest: 30 },
      }
    )
    cy.intercept(
      "GET",
      "/api/pricing/trend?profile_id=profile-reports-export-failure",
      {
        statusCode: 200,
        body: { points: [{ t: "2026-01-01", v: 20 }] },
      }
    )
    cy.intercept(
      "GET",
      "/api/pricing/by-source?profile_id=profile-reports-export-failure",
      {
        statusCode: 200,
        body: { sources: { ebay: { latest: 30 } } },
      }
    )
    cy.intercept("GET", "/api/data/export/csv/items", {
      statusCode: 500,
      body: { error: "csv_export_failed" },
    }).as("exportCSVFailure")

    signInToReports("profile-reports-export-failure")
    cy.get('[data-testid="reports-export-button"]').click()
    cy.wait("@exportCSVFailure")

    cy.location("pathname").should("match", /^\/reports\/?$/)
    cy.get('[data-testid="reports-export-message"]')
      .should("be.visible")
      .and("contain", "reports_export_500")
  })

  it("UI-SCREEN-REPORTS-004 refreshes reports without route transition", () => {
    let wishlistRequestCount = 0
    cy.intercept("GET", "/api/profiles/active", {
      statusCode: 200,
      body: { id: "profile-reports-refresh" },
    }).as("activeProfile")
    cy.intercept(
      "GET",
      "/api/wishlist/hits?profile_id=profile-reports-refresh",
      (req) => {
        wishlistRequestCount += 1
        req.reply({
          statusCode: 200,
          body: {
            hits:
              wishlistRequestCount === 1
                ? [{ id: "initial-hit" }]
                : [{ id: "refreshed-hit-1" }, { id: "refreshed-hit-2" }],
          },
        })
      }
    ).as("wishlistHitsRefresh")
    cy.intercept("GET", "/api/pricing/stats?profile_id=profile-reports-refresh", {
      statusCode: 200,
      body: { min: 8, median: 16, latest: 32 },
    }).as("pricingStats")
    cy.intercept("GET", "/api/pricing/trend?profile_id=profile-reports-refresh", {
      statusCode: 200,
      body: { points: [{ t: "2026-01-01", v: 16 }] },
    }).as("pricingTrend")
    cy.intercept(
      "GET",
      "/api/pricing/by-source?profile_id=profile-reports-refresh",
      {
        statusCode: 200,
        body: { sources: { ebay: { latest: 32 } } },
      }
    ).as("pricingSource")

    signInToReports("profile-reports-refresh")
    cy.wait("@activeProfile")
    cy.wait("@wishlistHitsRefresh")
    cy.wait("@pricingStats")
    cy.wait("@pricingTrend")
    cy.wait("@pricingSource")
    reportMetric("Wishlist Hits").should("contain", "1")

    cy.contains("button", "Refresh Reports").click()
    cy.wait("@wishlistHitsRefresh")
    cy.wait("@pricingStats")
    cy.wait("@pricingTrend")
    cy.wait("@pricingSource")

    cy.location("pathname").should("match", /^\/reports\/?$/)
    reportMetric("Wishlist Hits").should("contain", "2")
    cy.contains("button", "Refresh Reports").should("be.enabled")
  })

  it("UI-SCREEN-REPORTS-004 disables export while reports are unavailable", () => {
    let activeProfileReadCount = 0
    cy.intercept("GET", "/api/profiles/active", (request) => {
      activeProfileReadCount += 1
      request.reply(
        activeProfileReadCount === 1
          ? { statusCode: 200, body: { id: "profile-reports-missing" } }
          : { statusCode: 404, body: { error: "active_profile_404" } }
      )
    }).as("activeProfileMissing")
    cy.intercept("GET", "/api/data/export/csv/items", {
      statusCode: 200,
      body: "id,title\n1,Test\n",
      headers: { "content-type": "text/csv; charset=utf-8" },
    }).as("exportCSV")

    signInToReports("profile-reports-missing")
    cy.wait("@activeProfileMissing")
    cy.get('[data-testid="reports-error"]').should("be.visible")
    cy.contains("active_profile_404").should("be.visible")
    cy.get('[data-testid="reports-export-button"]').should("be.disabled")
    cy.get('@exportCSV.all').should('have.length', 0)
    cy.get('[data-testid="reports-export-message"]').should("not.exist")
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

    signInToReports("profile-reports-3")
    cy.contains("Loading...").should("be.visible")
    cy.wait("@wishlistRetry")
    cy.get('[data-testid="reports-error"]').should("be.visible")
    cy.get('[data-testid="reports-error"]').contains("button", "Retry").click()
    cy.wait("@wishlistRetry")
    cy.get('[data-testid="reports-error"]').should("not.exist")
    cy.get('[data-testid="reports-empty-state"]').should("be.visible")
  })
})
