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
    cy.contains("Needs attention now").should("be.visible")
    cy.contains("Review discoveries").should("be.visible")
    cy.contains("Open pricing drops").should("be.visible")
    cy.contains("Recently added").should("be.visible")
    cy.contains("AFX Camaro").should("be.visible")
    cy.get('[data-testid="dashboard-card-link-open-pricing-drops"]')
      .should("have.attr", "href", "/wishlist")
    cy.get('[data-testid="dashboard-recent-item-afx-camaro"]')
      .should("have.attr", "href", "/inventory")
  })

  it("UI-SCREEN-HOME-009 renders the collector and inventory manager signal hub", () => {
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
          { title: "Low stock watch", value: 1, link: "/discoveries" },
        ],
      },
    }).as("dashboardSignalHub")

    signInToHome()
    cy.wait("@dashboardSignalHub")

    cy.get('[data-testid="dashboard-signal-hub"]').within(() => {
      cy.contains("Signal hub").should("be.visible")
      cy.contains("Collection size").should("be.visible")
      cy.contains("Wishlist hits").should("be.visible")
      cy.contains("Operational alerts").should("be.visible")
      cy.contains("Collection value").should("be.visible")
    })

    cy.get('[data-testid="dashboard-collector-health"]').within(() => {
      cy.contains("Collector health").should("be.visible")
      cy.contains("Recent additions").should("be.visible")
      cy.contains("AFX Camaro").should("be.visible")
      cy.contains("a", "Open inventory")
        .should("have.attr", "href", "/inventory")
    })

    cy.get('[data-testid="dashboard-purchase-pipeline"]').within(() => {
      cy.contains("Purchase pipeline").should("be.visible")
      cy.contains("2 wishlist hits").should("be.visible")
      cy.contains("1 price drop").should("be.visible")
      cy.contains("3 restocks").should("be.visible")
      cy.contains("a", "Open wishlist")
        .should("have.attr", "href", "/wishlist")
    })

    cy.get('[data-testid="dashboard-inventory-readiness"]').within(() => {
      cy.contains("Inventory readiness").should("be.visible")
      cy.contains("240 tracked units").should("be.visible")
      cy.contains("a", "Review media")
        .should("have.attr", "href", "/media")
    })

    cy.get('[data-testid="dashboard-actions-needed"]').within(() => {
      cy.contains("Actions needed").should("be.visible")
      cy.contains("Review discoveries").should("be.visible")
      cy.contains("Open pricing drops").should("be.visible")
      cy.get('[data-testid="dashboard-card-link-open-pricing-drops"]')
        .should("have.attr", "href", "/wishlist")
    })
  })

  it("UI-SCREEN-HOME-001A does not leak raw dashboard translation keys into the live page", () => {
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
    }).as("dashboardLocalized")

    signInToHome()
    cy.wait("@dashboardLocalized")

    cy.contains("dashboard.metrics.inventoryItems").should("not.exist")
    cy.contains("dashboard.metrics.inventoryUnits").should("not.exist")
    cy.contains("dashboard.metrics.wishlistHits").should("not.exist")
    cy.contains("dashboard.metrics.estimatedValue").should("not.exist")
    cy.contains("dashboard.attentionTitle").should("not.exist")
    cy.contains("dashboard.attentionDescription").should("not.exist")
    cy.contains("dashboard.recentlyAdded").should("not.exist")
    cy.contains("dashboard.recentlyAddedDescription").should("not.exist")
    cy.contains("dashboard.openInventory").should("not.exist")
    cy.contains("dashboard.refresh").should("not.exist")
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
    cy.contains("There are no action items right now.").should("be.visible")
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
          new_discoveries: 0,
          wishlist_hits: 0,
          price_drops: 0,
          low_stock_discoveries: 0,
          restocks: 0,
          recently_added: ["Recovery Item"],
          total_items: 1,
          total_instances: 1,
          estimated_value: 99,
          cards: [
            { title: "Recently Added", value: 1, link: "/collections" },
          ],
        },
      })
    }).as("dashboardRetry")

    cy.clearLocalStorage("cabinet.toastHistory.v1")
    signInToHome()
    cy.wait("@dashboardRetry")
    cy.contains("Dashboard unavailable").should("be.visible")
    cy.window()
      .its("localStorage")
      .invoke("getItem", "cabinet.toastHistory.v1")
      .should("be.a", "string")
      .then((raw) => {
        const records = JSON.parse(raw as string) as Array<{
          level: string
          title: string
          summary: string
          source_label: string
          category: string
        }>
        expect(
          records.some(
            (record) =>
              record.level === "error" &&
              record.title === "Dashboard unavailable" &&
              record.summary === "dashboard_fetch_failed_500" &&
              record.source_label === "Home" &&
              record.category === "system"
          )
        ).to.eq(true)
      })
    cy.contains("button", "Retry").click()
    cy.wait("@dashboardRetry")
    cy.contains("Dashboard unavailable").should("not.exist")

    cy.contains("Recently Added")
      .parentsUntil("div.border")
      .parent()
      .within(() => {
        cy.contains("a", "Open").click()
      })

    cy.location("pathname", { timeout: 15000 }).should("match", /^\/collections\/?$/)
  })

  it("UI-SCREEN-HOME-003 routes dashboard action links to live Cabinet destinations", () => {
    cy.intercept("GET", "/api/dashboard", {
      statusCode: 200,
      body: {
        new_discoveries: 0,
        wishlist_hits: 0,
        price_drops: 1,
        low_stock_discoveries: 0,
        restocks: 0,
        recently_added: ["Route Fix Item"],
        total_items: 1,
        total_instances: 1,
        estimated_value: 99,
        cards: [{ title: "Open pricing drops", value: 1, link: "/pricing" }],
      },
    }).as("dashboardRecentRoute")

    signInToHome()
    cy.wait("@dashboardRecentRoute")

    cy.get('[data-testid="dashboard-card-link-open-pricing-drops"]').click()
    cy.location("pathname", { timeout: 15000 }).should("match", /^\/wishlist\/?$/)

    cy.visit("/dashboard")
    cy.get('[data-testid="dashboard-recent-item-route-fix-item"]').click()
    cy.location("pathname", { timeout: 15000 }).should("match", /^\/inventory\/?$/)
  })

  it("UI-SCREEN-HOME-005 refreshes dashboard data without route transition", () => {
    let attempts = 0
    cy.intercept("GET", "/api/dashboard", (req) => {
      attempts += 1
      req.reply({
        statusCode: 200,
        body: {
          new_discoveries: 0,
          wishlist_hits: attempts === 1 ? 1 : 4,
          price_drops: 0,
          low_stock_discoveries: 0,
          restocks: 0,
          recently_added:
            attempts === 1 ? ["Before Refresh"] : ["After Refresh"],
          total_items: attempts === 1 ? 1 : 4,
          total_instances: attempts === 1 ? 1 : 4,
          estimated_value: attempts === 1 ? 100 : 400,
          cards: [],
        },
      })
    }).as("dashboardRefresh")

    signInToHome()
    cy.wait("@dashboardRefresh")
    cy.contains("Before Refresh").should("be.visible")

    cy.contains("button", "Refresh dashboard").click()
    cy.wait("@dashboardRefresh")

    cy.location("pathname", { timeout: 15000 }).should("eq", "/dashboard")
    cy.contains("Before Refresh").should("not.exist")
    cy.contains("After Refresh").should("be.visible")
    cy.contains("$400.00").should("be.visible")
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
