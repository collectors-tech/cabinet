describe("wishlist-row-side-panel", () => {
  function signInToWishlistWithRows() {
    let wishlistEntries = [
      {
        id: "wish-panel-1",
        item_id: "item-panel-1",
        priority: "medium",
        below_target_now: false,
        notes: "First panel note",
        target_price: 40,
        owned: true,
        price_paid: 32.5,
        purchase_url: "https://example.test/alpha",
        purchase_date: "2026-04-20",
        purchase_condition: "Boxed",
        quantity: 2,
        needed_quantity: 3,
      },
      {
        id: "wish-panel-2",
        item_id: "item-panel-2",
        priority: "high",
        below_target_now: false,
        notes: "Second panel note",
        target_price: 80,
        owned: false,
        price_paid: 0,
        purchase_url: "",
        purchase_date: "",
        purchase_condition: "",
        quantity: 1,
        needed_quantity: 2,
      },
    ];
    let wishlistItems = [
      {
        id: "item-panel-1",
        title: "Panel Wishlist Alpha",
        part_number: "PANEL-001",
        status: "wishlist",
        category: "Cards",
        priority: "medium",
      },
      {
        id: "item-panel-2",
        title: "Panel Wishlist Beta",
        part_number: "PANEL-002",
        status: "wishlist",
        category: "Comics",
        priority: "high",
      },
    ];

    cy.intercept("GET", "/api/wishlist", (req) => {
      req.reply({ statusCode: 200, body: { items: wishlistEntries } });
    }).as("wishlistItems");
    cy.intercept("GET", "/api/items?status=wishlist", (req) => {
      req.reply({ statusCode: 200, body: { items: wishlistItems } });
    }).as("catalogItems");
    cy.intercept("GET", "/api/pricing/stats?item_id=item-panel-1", {
      statusCode: 200,
      body: { min: 30, median: 35, latest: 39.5 },
    }).as("priceStatsPanel1");
    cy.intercept("GET", "/api/pricing/trend?item_id=item-panel-1", {
      statusCode: 200,
      body: {
        points: [
          { date: "2026-04-01", latest: 34 },
          { date: "2026-04-15", latest: 39.5 },
        ],
      },
    }).as("priceTrendPanel1");
    cy.intercept("GET", "/api/pricing/history?item_id=item-panel-1", {
      statusCode: 200,
      body: {
        history: [
          {
            snapshot_date: "2026-04-01",
            source: "ebay",
            latest_price: 34,
            stock_count: 2,
          },
          {
            snapshot_date: "2026-04-15",
            source: "ebay",
            latest_price: 39.5,
            stock_count: 4,
          },
        ],
      },
    }).as("priceHistoryPanel1");
    cy.intercept("GET", "/api/pricing/stats?item_id=item-panel-2", {
      statusCode: 200,
      body: { min: 70, median: 75, latest: 78 },
    }).as("priceStatsPanel2");
    cy.intercept("GET", "/api/pricing/trend?item_id=item-panel-2", {
      statusCode: 200,
      body: {
        points: [
          { date: "2026-04-01", latest: 80 },
          { date: "2026-04-15", latest: 78 },
        ],
      },
    }).as("priceTrendPanel2");
    cy.intercept("GET", "/api/pricing/history?item_id=item-panel-2", {
      statusCode: 200,
      body: {
        history: [
          {
            snapshot_date: "2026-04-01",
            source: "cardmarket",
            latest_price: 80,
            stock_count: 1,
          },
          {
            snapshot_date: "2026-04-15",
            source: "cardmarket",
            latest_price: 78,
            stock_count: 3,
          },
        ],
      },
    }).as("priceHistoryPanel2");
    cy.intercept("PUT", "/api/items/item-panel-2", (req) => {
      wishlistItems = wishlistItems.map((item) =>
        item.id === "item-panel-2"
          ? {
              ...item,
              title: req.body.title,
              part_number: req.body.part_number,
              category: req.body.category,
            }
          : item
      );
      req.reply({
        statusCode: 200,
        body: wishlistItems.find((item) => item.id === "item-panel-2"),
      });
    }).as("updateWishlistItem");
    cy.intercept("PUT", "/api/wishlist", (req) => {
      wishlistEntries = wishlistEntries.map((entry) =>
        entry.id === req.body.id
          ? {
              ...entry,
              priority: req.body.priority,
              notes: req.body.notes,
              target_price: req.body.target_price,
              owned: req.body.owned ?? entry.owned,
              price_paid: req.body.price_paid ?? entry.price_paid,
              purchase_url: req.body.purchase_url ?? entry.purchase_url,
              purchase_date: req.body.purchase_date ?? entry.purchase_date,
              purchase_condition:
                req.body.purchase_condition ?? entry.purchase_condition,
              quantity: req.body.quantity ?? entry.quantity,
              needed_quantity: req.body.needed_quantity ?? entry.needed_quantity,
            }
          : entry
      );
      req.reply({ statusCode: 204, body: "" });
    }).as("updateWishlistEntry");
    cy.intercept("GET", "/api/profiles/*/settings").as("profileSettings");

    cy.e2eReset();
    cy.e2eSetSetupState("present");
    cy.e2eBootstrap().then(({ profile_id, profile_name }) => {
      cy.useBootstrappedProfile(profile_id, profile_name, {
        path: "/wishlist/",
      });
    });
    cy.wait("@wishlistItems");
    cy.wait("@catalogItems");
    cy.wait("@priceStatsPanel1");
    cy.wait("@priceTrendPanel1");
    cy.wait("@priceHistoryPanel1");
    cy.wait("@profileSettings");
    cy.wait("@profileSettings");
  }

  it("opens a right-side edit panel on double click and navigates visible records", () => {
    signInToWishlistWithRows();

    cy.contains("tr", "Panel Wishlist Alpha")
      .scrollIntoView()
      .click({ force: true });
    cy.get('[data-testid="wishlist-edit-panel"]').should("not.exist");
    cy.get('[data-testid="row-edit-modal"]').should("not.exist");

    cy.contains("tr", "Panel Wishlist Alpha")
      .scrollIntoView()
      .dblclick({ force: true });
    cy.get('[data-testid="wishlist-edit-panel"]')
      .should("be.visible")
      .should("have.attr", "data-side", "right");
    cy.get('input[name="title"]').should("have.value", "Panel Wishlist Alpha");
    cy.get('[data-testid="wishlist-edit-item-id"]').should(
      "contain.text",
      "item-panel-1"
    );
    cy.get('[data-testid="wishlist-edit-entry-id"]').should(
      "contain.text",
      "wish-panel-1"
    );
    cy.get('[data-testid="inventory-item-row-item-panel-1"]').should(
      "have.class",
      "bg-primary/5"
    );
    cy.get('[data-testid="inventory-item-row-item-panel-2"]').should(
      "not.have.class",
      "bg-primary/5"
    );
    cy.get('input[name="partNumber"]').should("have.value", "PANEL-001");
    cy.get('input[name="category"]').should("have.value", "Cards");
    cy.get('[data-testid="wishlist-edit-owned"]').should(
      "have.attr",
      "aria-checked",
      "true"
    );
    cy.get('input[name="pricePaid"]').should("have.value", "32.50");
    cy.get('[data-testid="wishlist-edit-market-price"]').should(
      "contain.text",
      "$39.50"
    );
    cy.get('[data-testid="wishlist-edit-price-graph"]').should(
      "contain.text",
      "2 points"
    );
    cy.get('[data-testid="wishlist-edit-price-graph"]').should(
      "contain.text",
      "ebay"
    );
    cy.get('input[name="targetPrice"]').should("have.value", "40");
    cy.get('input[name="quantity"]').should("have.value", "2");
    cy.get('input[name="neededQuantity"]').should("have.value", "3");
    cy.get('input[name="purchaseUrl"]').should(
      "have.value",
      "https://example.test/alpha"
    );
    cy.get('input[name="purchaseDate"]').should("have.value", "2026-04-20");
    cy.get('input[name="purchaseCondition"]').should("have.value", "Boxed");

    cy.get('[data-testid="wishlist-edit-panel"]').then(($panel) => {
      cy.get('[data-testid="wishlist-edit-next"]').click();
      cy.get('[data-testid="wishlist-edit-panel"]').should(($nextPanel) => {
        expect($nextPanel[0]).to.equal($panel[0]);
      });
    });
    cy.get('input[name="title"]').should("have.value", "Panel Wishlist Beta");
    cy.get('[data-testid="inventory-item-row-item-panel-1"]').should(
      "not.have.class",
      "bg-primary/5"
    );
    cy.get('[data-testid="inventory-item-row-item-panel-2"]').should(
      "have.class",
      "bg-primary/5"
    );
    cy.get('[data-testid="wishlist-edit-panel"]').then(($panel) => {
      cy.get('[data-testid="wishlist-edit-previous"]').click();
      cy.get('[data-testid="wishlist-edit-panel"]').should(($nextPanel) => {
        expect($nextPanel[0]).to.equal($panel[0]);
      });
    });
    cy.get('input[name="title"]').should("have.value", "Panel Wishlist Alpha");
    cy.get('[data-testid="inventory-item-row-item-panel-1"]').should(
      "have.class",
      "bg-primary/5"
    );
    cy.get('[data-testid="inventory-item-row-item-panel-2"]').should(
      "not.have.class",
      "bg-primary/5"
    );
    cy.get('[data-testid="wishlist-edit-next"]').click();

    cy.get('input[name="title"]').clear().type("Panel Wishlist Beta Updated");
    cy.get('textarea[name="notes"]').clear().type("Updated from side panel");
    cy.get('input[name="quantity"]').clear().type("4");
    cy.get('input[name="neededQuantity"]').clear().type("5");
    cy.get('input[name="pricePaid"]').clear().type("77.25");
    cy.get('input[name="purchaseUrl"]').clear().type("https://example.test/beta");
    cy.get('input[name="purchaseDate"]').type("2026-04-29");
    cy.get('input[name="purchaseCondition"]').clear().type("Loose");
    cy.contains("button", "Save changes").click();

    cy.wait("@updateWishlistItem");
    cy.wait("@updateWishlistEntry").then(({ request }) => {
      expect(request.body).to.include({
        quantity: 4,
        needed_quantity: 5,
        price_paid: 77.25,
        purchase_url: "https://example.test/beta",
        purchase_date: "2026-04-29",
        purchase_condition: "Loose",
      });
    });
    cy.wait("@wishlistItems");
    cy.wait("@catalogItems");
    cy.contains("Panel Wishlist Beta Updated").should("be.visible");
    cy.contains("Updated from side panel").should("be.visible");
  });
});
