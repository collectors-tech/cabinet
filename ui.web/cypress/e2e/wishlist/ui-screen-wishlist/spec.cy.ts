describe("ui-screen-wishlist", () => {
  function stubWishlistData() {
    cy.intercept("GET", "/api/wishlist", {
      statusCode: 200,
      body: {
        items: [
          {
            id: "wish-1",
            item_id: "item-collector-1",
            priority: "medium",
            below_target_now: false,
          },
          {
            id: "wish-2",
            item_id: "item-collector-2",
            priority: "high",
            below_target_now: true,
          },
        ],
      },
    }).as("wishlistItems");
    cy.intercept("GET", "/api/items?status=wishlist", {
      statusCode: 200,
      body: {
        items: [
          {
            id: "item-collector-1",
            title: "AFX Mega-G+ Camaro Wildfire",
            part_number: "22073",
            status: "wishlist",
            category: "Slot Cars",
            priority: "medium",
          },
          {
            id: "item-collector-2",
            title: "F1 Silverline",
            part_number: "F1002",
            status: "wishlist",
            category: "Formula",
            priority: "high",
          },
        ],
      },
    }).as("catalogItems");
    cy.intercept("GET", "/api/pricing/stats?item_id=item-collector-1", {
      statusCode: 200,
      body: { min: 35, median: 40, latest: 42.5 },
    }).as("priceStatsCollector1");
    cy.intercept("GET", "/api/pricing/trend?item_id=item-collector-1", {
      statusCode: 200,
      body: {
        points: [
          { date: "2026-04-01", latest: 36 },
          { date: "2026-04-08", latest: 39 },
          { date: "2026-04-15", latest: 38 },
          { date: "2026-04-22", latest: 42.5 },
        ],
      },
    }).as("priceTrendCollector1");
    cy.intercept("GET", "/api/pricing/history?item_id=item-collector-1", {
      statusCode: 200,
      body: {
        history: [
          {
            snapshot_date: "2026-04-01",
            source: "ebay",
            min_price: 34,
            median_price: 36,
            latest_price: 36,
            stock_count: 3,
          },
          {
            snapshot_date: "2026-04-22",
            source: "ebay",
            min_price: 40,
            median_price: 41,
            latest_price: 42.5,
            stock_count: 5,
          },
        ],
      },
    }).as("priceHistoryCollector1");
    cy.intercept("GET", "/api/pricing/stats?item_id=item-collector-2", {
      statusCode: 200,
      body: { min: 0, median: 0, latest: 0 },
    }).as("priceStatsCollector2");
    cy.intercept("GET", "/api/pricing/trend?item_id=item-collector-2", {
      statusCode: 200,
      body: { points: [] },
    }).as("priceTrendCollector2");
    cy.intercept("GET", "/api/pricing/history?item_id=item-collector-2", {
      statusCode: 200,
      body: { history: [] },
    }).as("priceHistoryCollector2");
  }

  function signInToWishlist(options?: {
    skipStub?: boolean;
    useExistingIntercepts?: boolean;
  }) {
    if (!options?.skipStub) {
      stubWishlistData();
    } else if (!options.useExistingIntercepts) {
      cy.intercept("GET", "/api/wishlist").as("wishlistItems");
      cy.intercept("GET", "/api/items?status=wishlist").as("catalogItems");
    }
    cy.intercept("GET", "/api/profiles/*/settings").as("profileSettings");
    cy.e2eReset();
    cy.window().then((win) => {
      win.localStorage.removeItem("cabinet.wishlist.statusFilters");
    });
    cy.e2eSetSetupState("present");
    cy.e2eBootstrap().then(({ profile_id, profile_name }) => {
      cy.useBootstrappedProfile(profile_id, profile_name, {
        path: "/wishlist/",
      });
    });
    cy.wait("@wishlistItems");
    cy.wait("@catalogItems");
    cy.wait("@profileSettings");
    cy.wait("@profileSettings");
  }

  function openWishlistRowActions(rowText: string) {
    cy.contains("tr", rowText)
      .find('[data-testid="task-row-actions-trigger"]')
      .scrollIntoView()
      .should("be.visible")
      .click({ force: true });
    cy.get('[role="menu"]').should("be.visible");
  }

  function collectionFilterOptionKey(value: string) {
    return value
      .trim()
      .toLowerCase()
      .replace(/[^a-z0-9]+/g, "-")
      .replace(/^-+|-+$/g, "");
  }

  function selectWishlistCollection(collectionName: string) {
    cy.get('[data-testid="wishlist-table-collection-trigger"]').click();
    cy.get(
      `[data-testid="wishlist-table-collection-option-${collectionFilterOptionKey(collectionName)}"]`
    ).click();
  }

  function clearWishlistCollectionFilter() {
    cy.get('[data-testid="wishlist-table-collection-trigger"]').click();
    cy.get('[data-testid="wishlist-table-collection-clear"]').click();
  }

  it("UI-SCREEN-WISHLIST-001 filters list and persists row/card view mode", () => {
    signInToWishlist();

    cy.contains("Wishlist").should("be.visible");
    cy.get('[data-testid="wishlist-global-header-actions"]').within(() => {
      cy.get('[data-testid="wishlist-new-action"]')
        .should("be.visible")
        .and("have.attr", "aria-label", "New wishlist item")
        .and("not.contain.text", "New");
      cy.get('[data-testid="wishlist-create-collection-action"]')
        .should("be.visible")
        .and("have.attr", "aria-label", "Create collection")
        .and("not.contain.text", "Create");
      cy.get('[data-testid="wishlist-import-action"]')
        .should("be.visible")
        .and("have.attr", "aria-label", "Import wishlist entries")
        .and("not.contain.text", "Import");
      cy.get('[data-testid="wishlist-create-menu-trigger"]').should("not.exist");
    });
    cy.contains(
      "Track wanted items, target prices, and planning decisions before they become owned inventory."
    ).should("not.exist");
    cy.contains("wishlist.description").should("not.exist");
    cy.get('[data-testid="wishlist-planning-summary"]').should("not.exist");
    cy.get("table").should("be.visible");
    cy.contains("button", "Cards").click();
    cy.window().its("localStorage").invoke("getItem", "cabinet.viewMode.wishlist").should("eq", "cards");
    cy.contains("Status:").should("be.visible");
    cy.reload();
    cy.contains("Status:").should("be.visible");

    cy.contains("button", "Rows").click();
    cy.get('input[placeholder="Filter by title or part number..."]').type("no-match-wishlist");
    cy.contains("No results.").should("be.visible");
  });

  it("UI-SCREEN-WISHLIST-013 renders representative seeded wishlist rows without stubs", () => {
    signInToWishlist({ skipStub: true });

    cy.get('button[aria-label="Switch to rows view"]').click();
    cy.contains("Wishlist Sample Grail Chase").should("be.visible");
    cy.contains("Wishlist Sample Price Drop Watch").should("be.visible");
    cy.contains("Wishlist Sample Steady Watch").should("be.visible");
    cy.get('[data-testid="wishlist-planning-summary"]').should("not.exist");
    cy.get('select[data-testid^="wishlist-priority-select-"]').then(
      ($selects) => {
        const values = [...$selects].map(
          (select) => (select as HTMLSelectElement).value
        );
        expect(values).to.include("high");
        expect(values).to.include("low");
      }
    );
  });

  it("UI-SCREEN-WISHLIST-018 renders compact deterministic row thumbnails", () => {
    signInToWishlist();

    cy.get('button[aria-label="Switch to rows view"]').click();
    cy.get('[data-testid="wishlist-thumbnail-item-collector-1"]')
      .should("be.visible")
      .and("have.attr", "aria-hidden", "true")
      .and("have.attr", "data-thumbnail-key", "item-collector-1");
    cy.get('[data-testid="wishlist-thumbnail-item-collector-2"]')
      .should("be.visible")
      .and("have.attr", "aria-hidden", "true")
      .and("have.attr", "data-thumbnail-key", "item-collector-2");

    cy.get('[data-testid="wishlist-thumbnail-item-collector-1"]')
      .invoke("attr", "style")
      .then((firstStyle) => {
        cy.get('[data-testid="wishlist-thumbnail-item-collector-2"]')
          .invoke("attr", "style")
          .should("not.eq", firstStyle);
      });

    cy.contains("tr", "AFX Mega-G+ Camaro Wildfire").within(() => {
      cy.contains("AFX Mega-G+ Camaro Wildfire").should("be.visible");
      cy.contains("22073").should("not.exist");
    });
  });

  it("UI-SCREEN-WISHLIST-003 supports multi-select with bulk action toolbar", () => {
    signInToWishlist();

    cy.get('button[aria-label="Switch to rows view"]').click();
    cy.get('button[aria-label="Select all"]').click();

    cy.contains(/selected/i).should("be.visible");
    cy.get('button[aria-label="Update status"]').should("be.visible");
    cy.get('button[aria-label="Update priority"]').should("be.visible");
    cy.get('button[aria-label="Export wishlist entries"]').should("be.visible");
    cy.get('button[aria-label="Delete selected wishlist entries"]').should("be.visible");
  });

  it("UI-SCREEN-WISHLIST-002 exposes direct compact header actions", () => {
    signInToWishlist();

    cy.get('[data-testid="wishlist-global-header-actions"]').within(() => {
      cy.get('[data-testid="wishlist-new-action"]')
        .should("be.visible")
        .and("have.attr", "title", "New wishlist item");
      cy.get('[data-testid="wishlist-create-collection-action"]')
        .should("be.visible")
        .and("have.attr", "title", "Create collection");
      cy.get('[data-testid="wishlist-import-action"]')
        .should("be.visible")
        .and("have.attr", "title", "Import wishlist entries");
      cy.get('[data-testid="wishlist-create-menu-trigger"]').should("not.exist");
    });
  });

  it("UI-SCREEN-WISHLIST-004 shows icon-only header actions", () => {
    signInToWishlist();

    cy.get('[data-testid="wishlist-new-action"]')
      .should("be.visible")
      .and("not.contain", "New");
    cy.get('[data-testid="wishlist-create-collection-action"]')
      .should("be.visible")
      .and("not.contain", "Create");
    cy.get('[data-testid="wishlist-import-action"]')
      .should("be.visible")
      .and("not.contain", "Import");
  });

  it("UI-SCREEN-WISHLIST-005 creates collections from the header action", () => {
    signInToWishlist();

    cy.get('[data-testid="wishlist-create-collection-action"]').click();
    cy.get('[data-testid="wishlist-create-collection-dialog"]').should(
      "be.visible"
    );
    cy.get('[data-testid="wishlist-create-collection-name"]').type(
      "Wishlist Inline Alpha"
    );
    cy.get('[data-testid="wishlist-create-collection-save"]').click();
    cy.get('[data-testid="wishlist-create-collection-dialog"]').should(
      "not.exist"
    );
    cy.get('[data-testid="wishlist-table-collection-trigger"]').click();
    cy.get(
      `[data-testid="wishlist-table-collection-option-${collectionFilterOptionKey(
        "Wishlist Inline Alpha"
      )}"]`
    ).should("be.visible");
    cy.get("body").type("{esc}");
  });

  it("UI-SCREEN-WISHLIST-005 keeps collection create dialog open on blank save", () => {
    signInToWishlist();

    cy.get('[data-testid="wishlist-create-collection-action"]').click();
    cy.get('[data-testid="wishlist-create-collection-save"]').click();
    cy.get('[data-testid="wishlist-create-collection-name"]').should(
      "have.attr",
      "aria-invalid",
      "true"
    );
    cy.get('[data-testid="wishlist-create-collection-validation"]')
      .should("be.visible")
      .and("contain", "Collection name is required.");
    cy.get('[data-testid="wishlist-create-collection-name"]').should("be.visible");
  });

  it("UI-SCREEN-WISHLIST-011 filters table rows by selected wishlist collection", () => {
    signInToWishlist();

    cy.contains("AFX Mega-G+ Camaro Wildfire").should("be.visible");
    cy.contains("F1 Silverline").should("be.visible");

    selectWishlistCollection("Overflow");
    cy.get('[data-testid="wishlist-table-collection-trigger"]').should(
      "contain.text",
      "Overflow"
    );

    cy.contains("AFX Mega-G+ Camaro Wildfire").should("not.exist");
    cy.contains("F1 Silverline").should("not.exist");
    cy.contains("No results.").should("be.visible");

    clearWishlistCollectionFilter();
    cy.contains("AFX Mega-G+ Camaro Wildfire").should("be.visible");
    cy.contains("F1 Silverline").should("be.visible");
  });

  it("UI-SCREEN-WISHLIST-014 does not inherit the Collections screen selection", () => {
    stubWishlistData();
    cy.intercept("GET", "/api/profiles/*/settings").as("profileSettings");
    cy.intercept("PUT", "/api/profiles/e2e-profile-001/settings").as(
      "saveCollectionSettings"
    );
    cy.e2eReset();
    cy.e2eSetSetupState("present");
    cy.e2eBootstrap().then(({ profile_id, profile_name }) => {
      cy.useBootstrappedProfile(profile_id, profile_name, {
        path: "/collections/",
      });
    });
    cy.wait("@profileSettings");

    cy.get('[data-testid="collections-row-store-1"]').click();
    cy.wait("@saveCollectionSettings");
    cy.get('[data-testid="collections-active-context"]').should(
      "contain.text",
      "Store 1"
    );

    cy.visit("/wishlist/");
    cy.wait("@wishlistItems");
    cy.wait("@catalogItems");
    cy.wait("@profileSettings");
    cy.get('[data-testid="wishlist-table-collection-trigger"]').should(
      "contain.text",
      "Collection"
    );
    cy.contains("AFX Mega-G+ Camaro Wildfire").should("be.visible");
    cy.contains("F1 Silverline").should("be.visible");
    cy.contains("No results.").should("not.exist");
  });

  it("UI-SCREEN-WISHLIST-012 uses shared table collection filter instead of old planning and picker sections", () => {
    signInToWishlist();

    cy.get('[data-testid="wishlist-inline-picker"]').should("not.exist");
    cy.get('[data-testid="wishlist-planning-summary"]').should("not.exist");
    cy.get('[data-testid="wishlist-table-add-collection"]').should("not.exist");
    cy.get('[data-testid="wishlist-table-new-collection-form"]').should(
      "not.exist"
    );
    cy.get('[data-testid="wishlist-table-collection-select"]').should(
      "not.exist"
    );
    cy.get('[data-testid="wishlist-table-collection-trigger"]')
      .should("be.visible")
      .and("contain.text", "Collection");

    selectWishlistCollection("Overflow");
    cy.get('[data-testid="wishlist-table-collection-trigger"]').should(
      "contain.text",
      "Overflow"
    );
    cy.contains("AFX Mega-G+ Camaro Wildfire").should("not.exist");
    cy.contains("F1 Silverline").should("not.exist");
    cy.contains("No results.").should("be.visible");
  });

  it("UI-SCREEN-WISHLIST-007 renders wishlist collection semantics instead of task seed rows", () => {
    signInToWishlist();

    cy.get('button[aria-label="Switch to rows view"]').click();
    cy.contains("th", "Item ID").should("not.exist");
    cy.contains("th", "Title").should("exist");
    cy.contains("th", "Watch Status").should("not.exist");
    cy.contains("th", "Market Price").should("exist");
    cy.contains("th", "Price Graph").should("exist");
    cy.contains("th", "Cost").should("exist");
    cy.contains("th", "Priority").should("exist");
    cy.contains("th", "Target Priority").should("not.exist");
    cy.contains("th", "Task").should("not.exist");
    cy.contains("AFX Mega-G+ Camaro Wildfire").should("be.visible");
    cy.contains("22073").should("not.exist");
    cy.contains(/TASK-\d+/).should("not.exist");
    cy.contains("Backlog").should("not.exist");
  });

  it("UI-SCREEN-WISHLIST-015 renders item pricing trajectory from pricing APIs", () => {
    signInToWishlist();

    cy.wait("@priceStatsCollector1");
    cy.wait("@priceTrendCollector1");
    cy.wait("@priceHistoryCollector1");
    cy.get('button[aria-label="Switch to rows view"]').click();

    cy.contains("tr", "AFX Mega-G+ Camaro Wildfire").within(() => {
      cy.get('[data-testid="wishlist-owned-checkbox-item-collector-1"]').should(
        "not.exist"
      );
      cy.get('[data-testid="wishlist-owned-tick-item-collector-1"]').should(
        "not.exist"
      );
      cy.get('[data-testid="wishlist-purchase-open-item-collector-1"]').should(
        "exist"
      );
      cy.get('[data-testid="wishlist-price-trend-item-collector-1"]')
        .find('[data-testid="wishlist-price-sparkline-item-collector-1"]')
        .should("exist");
      cy.get('[data-testid="wishlist-market-price-item-collector-1"]').should(
        "contain.text",
        "$42.50"
      );
      cy.get('[data-testid="wishlist-price-sparkline-item-collector-1"]')
        .should("exist")
        .and("have.attr", "aria-label")
        .and("contain", "4 price points");
      cy.get('[data-testid="wishlist-price-graph-meta-item-collector-1"]')
        .should("contain.text", "4 points")
        .and("contain.text", "2026-04-01")
        .and("contain.text", "2026-04-22")
        .and("contain.text", "ebay");
    });

    cy.contains("tr", "F1 Silverline").within(() => {
      cy.get('[data-testid="wishlist-owned-checkbox-item-collector-2"]').should(
        "not.exist"
      );
      cy.get('[data-testid="wishlist-owned-tick-item-collector-2"]').should(
        "not.exist"
      );
      cy.get('[data-testid="wishlist-purchase-open-item-collector-2"]').should(
        "be.visible"
      );
    });
  });

  it("UI-SCREEN-WISHLIST-006 does not expose Mark owned from the row action menu", () => {
    let wishlistEntries = [
      {
        id: "wish-1",
        item_id: "item-collector-1",
        priority: "high",
        below_target_now: false,
      },
    ];
    let wishlistItems = [
      {
        id: "item-collector-1",
        title: "AFX Mega-G+ Camaro Wildfire",
        part_number: "22073",
        status: "wishlist",
        category: "Slot Cars",
        priority: "high",
      },
    ];

    cy.intercept("GET", "/api/wishlist", (req) => {
      req.reply({ statusCode: 200, body: { items: wishlistEntries } });
    }).as("wishlistItems");
    cy.intercept("GET", "/api/items?status=wishlist", (req) => {
      req.reply({ statusCode: 200, body: { items: wishlistItems } });
    }).as("catalogItems");

    signInToWishlist({ skipStub: true, useExistingIntercepts: true });

    openWishlistRowActions("AFX Mega-G+ Camaro Wildfire");
    cy.get('[data-testid="wishlist-mark-owned-action"]').should("not.exist");
    cy.contains('[role="menuitem"]', "Edit").should("be.visible");
    cy.contains('[role="menuitem"]', "Delete").should("be.visible");
  });

  it("UI-SCREEN-WISHLIST-008 keeps planning filters in the shared table toolbar", () => {
    cy.intercept("GET", "/api/wishlist", {
      statusCode: 200,
      body: {
        items: [
          {
            id: "wish-1",
            item_id: "item-collector-1",
            priority: "medium",
            below_target_now: false,
            notes: "Wait for convention restock",
          },
          {
            id: "wish-2",
            item_id: "item-collector-2",
            priority: "high",
            below_target_now: true,
            notes: "Buy if price drops again",
          },
          {
            id: "wish-3",
            item_id: "item-collector-3",
            priority: "critical",
            below_target_now: false,
            notes: "Need before championship season",
          },
        ],
      },
    }).as("wishlistItems");
    cy.intercept("GET", "/api/items?status=wishlist", {
      statusCode: 200,
      body: {
        items: [
          {
            id: "item-collector-1",
            title: "AFX Mega-G+ Camaro Wildfire",
            part_number: "22073",
            status: "wishlist",
            category: "Slot Cars",
            priority: "medium",
          },
          {
            id: "item-collector-2",
            title: "F1 Silverline",
            part_number: "F1002",
            status: "wishlist",
            category: "Formula",
            priority: "high",
          },
          {
            id: "item-collector-3",
            title: "Team Transport Hauler",
            part_number: "TT-88",
            status: "wishlist",
            category: "Haulers",
            priority: "critical",
          },
        ],
      },
    }).as("catalogItems");

    signInToWishlist({ skipStub: true, useExistingIntercepts: true });

    cy.get('[data-testid="wishlist-planning-summary"]').should("not.exist");
    cy.window()
      .its("localStorage")
      .invoke("getItem", "cabinet.wishlistPlanningFocus")
      .should("not.exist");
    cy.get('[data-testid="wishlist-table-toolbar"]').within(() => {
      cy.contains("button", "Status").should("be.visible");
      cy.contains("button", "Priority").should("be.visible");
      cy.get('[data-testid="wishlist-table-collection-trigger"]').should(
        "be.visible"
      );
    });
  });

  it("UI-SCREEN-WISHLIST-009 wires direct header actions into real wishlist dialogs", () => {
    signInToWishlist();

    cy.get('[data-testid="wishlist-new-action"]').click();
    cy.contains("Create Wishlist Entry").should("be.visible");

    cy.contains("button", "Close").click();

    cy.get('[data-testid="wishlist-create-collection-action"]').click();
    cy.get('[data-testid="wishlist-create-collection-name"]').should(
      "be.visible"
    );
    cy.contains("button", "Cancel").click();

    cy.get('[data-testid="wishlist-import-action"]').click();
    cy.contains("Import Wishlist Entries").should("be.visible");
  });

  it("UI-SCREEN-WISHLIST-010 persists wishlist edit and delete actions", () => {
    let wishlistEntries = [
      {
        id: "wish-edit-1",
        item_id: "item-collector-1",
        priority: "medium",
        below_target_now: false,
        notes: "Original watch note",
        target_price: 20,
      },
    ];
    let wishlistItems = [
      {
        id: "item-collector-1",
        title: "AFX Mega-G+ Camaro Wildfire",
        part_number: "22073",
        status: "wishlist",
        category: "Slot Cars",
        priority: "medium",
      },
    ];

    cy.intercept("GET", "/api/wishlist", (req) => {
      req.reply({ statusCode: 200, body: { items: wishlistEntries } });
    }).as("wishlistItems");
    cy.intercept("GET", "/api/items?status=wishlist", (req) => {
      req.reply({ statusCode: 200, body: { items: wishlistItems } });
    }).as("catalogItems");
    cy.intercept("PUT", "/api/wishlist", (req) => {
      wishlistEntries = wishlistEntries.map((entry) =>
        entry.id === req.body.id
          ? {
              ...entry,
              priority: req.body.priority,
              notes: req.body.notes,
              target_price: req.body.target_price,
            }
          : entry
      );
      req.reply({ statusCode: 204, body: "" });
    }).as("updateWishlistEntry");
    cy.intercept("PUT", "/api/items/item-collector-1", (req) => {
      wishlistItems = wishlistItems.map((item) =>
        item.id === "item-collector-1"
          ? {
              ...item,
              title: req.body.title,
              category: req.body.category,
            }
          : item
      );
      req.reply({
        statusCode: 200,
        body: {
          ...wishlistItems[0],
          title: req.body.title,
          category: req.body.category,
        },
      });
    }).as("updateWishlistItem");
    cy.intercept("DELETE", "/api/wishlist?id=wish-edit-1", (req) => {
      wishlistEntries = [];
      wishlistItems = [];
      req.reply({ statusCode: 204, body: "" });
    }).as("deleteWishlistEntry");

    signInToWishlist({ skipStub: true });

    openWishlistRowActions("AFX Mega-G+ Camaro Wildfire");
    cy.contains('[role="menuitem"]', "Edit").click({ force: true });

    cy.contains("Edit Wishlist Entry").should("be.visible");
    cy.get('input[name="title"]').clear().type("AFX Mega-G+ Camaro Updated");
    cy.get('textarea[name="notes"]').clear().type("Updated watch note");
    cy.contains("button", "Save changes").click();

    cy.wait("@updateWishlistItem");
    cy.wait("@updateWishlistEntry");
    cy.wait("@wishlistItems");
    cy.wait("@catalogItems");
    cy.contains("AFX Mega-G+ Camaro Updated").should("exist");
    cy.contains("Updated watch note").should("exist");

    openWishlistRowActions("AFX Mega-G+ Camaro Updated");
    cy.get('[role="menu"]')
      .should("be.visible")
      .within(() => {
        cy.contains('[role="menuitem"]', /^Delete$/)
          .should("be.visible")
          .click({ force: true });
      });
    cy.contains("Delete this wishlist entry").should("be.visible");
    cy.contains("button", "Delete").click();

    cy.wait("@deleteWishlistEntry");
    cy.wait("@wishlistItems");
    cy.wait("@catalogItems");
    cy.contains("AFX Mega-G+ Camaro Updated").should("not.exist");
  });
});
