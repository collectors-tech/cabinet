describe("wishlist-pricing-columns", () => {
  function expectHeaderVisible(title: string) {
    cy.contains("th", title).scrollIntoView().should("be.visible");
  }

  function openWishlist() {
    cy.e2eReset();
    cy.e2eSetSetupState("present");
    cy.e2eBootstrap().then(({ profile_id, profile_name }) => {
      cy.useBootstrappedProfile(profile_id, profile_name, {
        path: "/wishlist/",
      });
    });
  }

  it("shows buying columns and persists inline cost and priority edits", () => {
    let wishlistEntries = [
      {
        id: "wish-price-1",
        item_id: "item-price-1",
        priority: "medium",
        below_target_now: false,
        notes: "Watch clean boxed examples",
        target_price: 25,
      },
    ];

    cy.intercept("GET", "/api/wishlist", (req) => {
      req.reply({ statusCode: 200, body: { items: wishlistEntries } });
    }).as("wishlistItems");
    cy.intercept("GET", "/api/items?status=wishlist", {
      statusCode: 200,
      body: {
        items: [
          {
            id: "item-price-1",
            title: "Wishlist Pricing Candidate",
            part_number: "WPC-001",
            status: "wishlist",
            category: "Cards",
            priority: "medium",
          },
        ],
      },
    }).as("catalogItems");
    cy.intercept("GET", "/api/pricing/stats?item_id=item-price-1", {
      statusCode: 200,
      body: { min: 18, median: 21, latest: 22.5 },
    }).as("priceStats");
    cy.intercept("GET", "/api/pricing/trend?item_id=item-price-1", {
      statusCode: 200,
      body: {
        points: [
          { date: "2026-04-15", latest: 32 },
          { date: "2026-04-16", latest: 31 },
          { date: "2026-04-17", latest: 30 },
          { date: "2026-04-18", latest: 29 },
          { date: "2026-04-19", latest: 28 },
          { date: "2026-04-20", latest: 27 },
          { date: "2026-04-21", latest: 26 },
          { date: "2026-04-22", latest: 25.5 },
          { date: "2026-04-23", latest: 25 },
          { date: "2026-04-24", latest: 24.5 },
          { date: "2026-04-25", latest: 24 },
          { date: "2026-04-26", latest: 22.5 },
        ],
      },
    }).as("priceTrend");
    cy.intercept("GET", "/api/pricing/history?item_id=item-price-1", {
      statusCode: 200,
      body: {
        history: [
          {
            snapshot_date: "2026-04-25",
            source: "ebay",
            latest_price: 24,
            stock_count: 2,
          },
          {
            snapshot_date: "2026-04-26",
            source: "ebay",
            latest_price: 22.5,
            stock_count: 3,
          },
        ],
      },
    }).as("priceHistory");
    cy.intercept("PUT", "/api/wishlist", (req) => {
      expect(req.body.id).to.eq("wish-price-1");
      wishlistEntries = wishlistEntries.map((entry) =>
        entry.id === req.body.id
          ? {
              ...entry,
              priority: req.body.priority ?? entry.priority,
              target_price: req.body.target_price ?? entry.target_price,
              notes: req.body.notes ?? entry.notes,
              highlight_hit: req.body.highlight_hit ?? entry.highlight_hit,
              below_target_now:
                req.body.below_target_now ?? entry.below_target_now,
            }
          : entry
      );
      req.reply({ statusCode: 204, body: "" });
    }).as("updateWishlistEntry");

    openWishlist();

    cy.wait("@wishlistItems");
    cy.wait("@catalogItems");
    cy.wait("@priceStats");
    cy.wait("@priceTrend");
    cy.wait("@priceHistory");

    cy.get('button[aria-label="Switch to rows view"]').click();
    cy.contains("th", "Item ID").should("not.exist");
    expectHeaderVisible("Market Price");
    expectHeaderVisible("Price Graph");
    expectHeaderVisible("Cost");
    expectHeaderVisible("Priority");
    cy.contains("th", "Target Priority").should("not.exist");

    cy.contains("Wishlist Pricing Candidate")
      .scrollIntoView()
      .should("be.visible");
    cy.contains("$22.50").should("be.visible");
    cy.get('[data-testid="wishlist-price-trend-item-price-1"]')
      .scrollIntoView()
      .should("have.attr", "aria-label", "Price trending down")
      .and("be.visible");
    cy.get('[data-testid="wishlist-price-sparkline-item-price-1"]')
      .should("be.visible")
      .should("have.attr", "data-slot", "chart")
      .find(".recharts-line-curve")
      .should("have.attr", "d")
      .and("not.be.empty");
    cy.get('[data-testid="wishlist-price-sparkline-item-price-1"]').focus();
    cy.get('[data-testid="wishlist-price-points-item-price-1"]')
      .should("be.visible")
      .and("contain", "Latest 10")
      .and("contain", "2026-04-17 $30.00")
      .and("contain", "2026-04-26 $22.50")
      .and("not.contain", "2026-04-16");

    cy.get('[data-testid="wishlist-cost-input-item-price-1"]')
      .scrollIntoView()
      .clear()
      .type("19.75{enter}");
    cy.wait("@updateWishlistEntry")
      .its("request.body")
      .should("include", {
        id: "wish-price-1",
        target_price: 19.75,
      });

    cy.get('[data-testid="wishlist-priority-select-item-price-1"]')
      .scrollIntoView()
      .then(($select) => {
        const select = $select[0] as HTMLSelectElement;
        select.value = "high";
        select.dispatchEvent(new Event("change", { bubbles: true }));
      });
    cy.wait("@updateWishlistEntry")
      .its("request.body")
      .should("include", {
        id: "wish-price-1",
        priority: "high",
      });

    cy.reload();
    cy.wait("@wishlistItems");
    cy.wait("@catalogItems");
    cy.get('[data-testid="wishlist-cost-input-item-price-1"]')
      .scrollIntoView()
      .should("have.value", "19.75");
    cy.get('[data-testid="wishlist-priority-select-item-price-1"]').should(
      "have.value",
      "high"
    );
  });

  it("tracks owned state, purchase details, quantity, and needs inline", () => {
    const today = new Date().toISOString().slice(0, 10);
    let wishlistEntries = [
      {
        id: "wish-price-1",
        item_id: "item-price-1",
        priority: "medium",
        below_target_now: false,
        notes: "Watch clean boxed examples",
        target_price: 25,
        owned: false,
        price_paid: 0,
        purchase_url: "",
        purchase_date: "",
        purchase_condition: "",
        quantity: 0,
        needed_quantity: 1,
      },
    ];

    cy.intercept("GET", "/api/wishlist", (req) => {
      req.reply({ statusCode: 200, body: { items: wishlistEntries } });
    }).as("wishlistItems");
    cy.intercept("GET", "/api/items?status=wishlist", {
      statusCode: 200,
      body: {
        items: [
          {
            id: "item-price-1",
            title: "Wishlist Pricing Candidate",
            part_number: "WPC-001",
            status: "wishlist",
            category: "Cards",
            priority: "medium",
          },
        ],
      },
    }).as("catalogItems");
    cy.intercept("GET", "/api/pricing/stats?item_id=item-price-1", {
      statusCode: 200,
      body: { min: 18, median: 21, latest: 22.5 },
    }).as("priceStats");
    cy.intercept("GET", "/api/pricing/trend?item_id=item-price-1", {
      statusCode: 200,
      body: {
        points: [
          { date: "2026-04-25", latest: 22.5 },
          { date: "2026-04-26", latest: 22.5 },
        ],
      },
    }).as("priceTrend");
    cy.intercept("PUT", "/api/wishlist", (req) => {
      expect(req.body.id).to.eq("wish-price-1");
      wishlistEntries = wishlistEntries.map((entry) =>
        entry.id === req.body.id ? { ...entry, ...req.body } : entry
      );
      req.reply({ statusCode: 204, body: "" });
    }).as("updateWishlistEntry");

    openWishlist();

    cy.wait("@wishlistItems");
    cy.wait("@catalogItems");
    cy.wait("@priceStats");
    cy.wait("@priceTrend");

    cy.get('button[aria-label="Switch to rows view"]').click();
    expectHeaderVisible("Owned");
    expectHeaderVisible("Price Paid");
    expectHeaderVisible("Qty");
    expectHeaderVisible("Needs");

    cy.get('[data-testid="wishlist-owned-checkbox-item-price-1"]').should(
      "not.exist"
    );
    cy.get('[data-testid="wishlist-owned-tick-item-price-1"]').should(
      "not.exist"
    );
    cy.get('[data-testid="wishlist-purchase-open-item-price-1"]')
      .scrollIntoView()
      .should("be.visible");

    cy.get('[data-testid="wishlist-purchase-open-item-price-1"]')
      .click({ force: true });
    cy.get('[data-testid="wishlist-purchase-dialog"]').should("be.visible");
    cy.get('[data-testid="wishlist-purchase-price-paid"]').should(
      "have.value",
      "25"
    );
    cy.get('[data-testid="wishlist-purchase-date"]').should(
      "have.value",
      today
    );
    cy.get('[data-testid="wishlist-purchase-quantity"]').should(
      "have.value",
      "1"
    );
    cy.get('[data-testid="wishlist-purchase-condition"]').should(
      "have.value",
      ""
    );
    cy.get('[data-testid="wishlist-purchase-price-paid"]')
      .click()
      .should("have.value", "")
      .type("18.25");
    cy.get('[data-testid="wishlist-purchase-quantity"]').type(
      "{selectall}2"
    );
    cy.get('[data-testid="wishlist-purchase-condition"]').type("Used - boxed");
    cy.get('[data-testid="wishlist-purchase-url"]').type(
      "https://example.test/purchase"
    );
    cy.get('[data-testid="wishlist-purchase-save"]').click();
    cy.wait("@updateWishlistEntry")
      .its("request.body")
      .should("include", {
        id: "wish-price-1",
        owned: true,
        price_paid: 18.25,
        purchase_url: "https://example.test/purchase",
        purchase_date: today,
        purchase_condition: "Used - boxed",
        quantity: 2,
      });
    cy.get('[data-testid="wishlist-price-paid-value-item-price-1"]').should(
      "contain.text",
      "$18.25"
    );

    cy.get('[data-testid="wishlist-qty-input-item-price-1"]').type(
      "{selectall}3{enter}",
      { force: true }
    );
    cy.wait("@updateWishlistEntry")
      .its("request.body")
      .should("include", { id: "wish-price-1", quantity: 3 });
    cy.get('[data-testid="wishlist-needs-input-item-price-1"]').type(
      "{selectall}5{enter}",
      { force: true }
    );
    cy.wait("@updateWishlistEntry")
      .its("request.body")
      .should("include", { id: "wish-price-1", needed_quantity: 5 });

    cy.reload();
    cy.wait("@wishlistItems");
    cy.wait("@catalogItems");
    cy.get('[data-testid="wishlist-owned-checkbox-item-price-1"]').should(
      "not.exist"
    );
    cy.get('[data-testid="wishlist-purchase-open-item-price-1"]')
      .scrollIntoView()
      .should("be.visible");
    cy.get('[data-testid="wishlist-price-paid-value-item-price-1"]').should(
      "contain.text",
      "$18.25"
    );
    cy.get('[data-testid="wishlist-qty-input-item-price-1"]').should(
      "have.value",
      "3"
    );
    cy.get('[data-testid="wishlist-needs-input-item-price-1"]').should(
      "have.value",
      "5"
    );
    cy.get('[data-testid="wishlist-purchase-open-item-price-1"]').click({
      force: true,
    });
    cy.get('[data-testid="wishlist-purchase-condition"]').should(
      "have.value",
      "Used - boxed"
    );
  });

  it("saves default purchase values when fields are untouched", () => {
    const today = new Date().toISOString().slice(0, 10);
    let wishlistEntries = [
      {
        id: "wish-default-1",
        item_id: "item-default-1",
        priority: "medium",
        below_target_now: false,
        target_price: 25,
        owned: false,
        price_paid: 0,
        quantity: 0,
        needed_quantity: 1,
      },
    ];

    cy.intercept("GET", "/api/wishlist", (req) => {
      req.reply({ statusCode: 200, body: { items: wishlistEntries } });
    }).as("wishlistItems");
    cy.intercept("GET", "/api/items?status=wishlist", {
      statusCode: 200,
      body: {
        items: [
          {
            id: "item-default-1",
            title: "Wishlist Default Purchase Candidate",
            status: "wishlist",
            category: "Cards",
            priority: "medium",
          },
        ],
      },
    }).as("catalogItems");
    cy.intercept("GET", "/api/pricing/stats?item_id=item-default-1", {
      statusCode: 200,
      body: { latest: 22.5 },
    }).as("priceStats");
    cy.intercept("GET", "/api/pricing/trend?item_id=item-default-1", {
      statusCode: 200,
      body: { points: [] },
    }).as("priceTrend");
    cy.intercept("PUT", "/api/wishlist", (req) => {
      expect(req.body.id).to.eq("wish-default-1");
      wishlistEntries = wishlistEntries.map((entry) =>
        entry.id === req.body.id ? { ...entry, ...req.body } : entry
      );
      req.reply({ statusCode: 204, body: "" });
    }).as("updateWishlistEntry");

    openWishlist();

    cy.wait("@wishlistItems");
    cy.wait("@catalogItems");
    cy.wait("@priceStats");
    cy.wait("@priceTrend");

    cy.get('button[aria-label="Switch to rows view"]').click();
    cy.get('[data-testid="wishlist-purchase-open-item-default-1"]').click({
      force: true,
    });
    cy.get('[data-testid="wishlist-purchase-price-paid"]').should(
      "have.value",
      "25"
    );
    cy.get('[data-testid="wishlist-purchase-quantity"]').should(
      "have.value",
      "1"
    );
    cy.get('[data-testid="wishlist-purchase-date"]').should(
      "have.value",
      today
    );
    cy.get('[data-testid="wishlist-purchase-save"]').click();

    cy.wait("@updateWishlistEntry")
      .its("request.body")
      .should("include", {
        id: "wish-default-1",
        owned: true,
        price_paid: 25,
        purchase_date: today,
        quantity: 1,
      });
    cy.get('[data-testid="wishlist-price-paid-value-item-default-1"]').should(
      "contain.text",
      "$25.00"
    );
  });
});
