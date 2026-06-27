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
            created_at: "2026-03-05T09:15:00Z",
            updated_at: "2026-03-29T11:30:00Z",
          },
          {
            id: "wish-2",
            item_id: "item-collector-2",
            priority: "high",
            below_target_now: true,
            created_at: "2026-03-10T13:45:00Z",
            updated_at: "2026-03-11T08:00:00Z",
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

  it("UI-SCREEN-WISHLIST-001 keeps view controls inline and omits Delivered from rows", () => {
    cy.viewport(1280, 720);
    signInToWishlist();

    cy.get('[data-testid="wishlist-table-toolbar"]').should("be.visible");
    cy.get('[data-testid="data-table-view-options-trigger"]').should(
      "be.visible"
    );
    cy.get('button[aria-label="Switch to rows view"]').should("be.visible");
    cy.get('button[aria-label="Switch to cards view"]').should("be.visible");

    cy.get('[data-testid="data-table-view-options-trigger"]').then(($view) => {
      const viewTop = Math.round($view[0].getBoundingClientRect().top);
      cy.get('button[aria-label="Switch to rows view"]').should(($rows) => {
        expect(
          Math.round($rows[0].getBoundingClientRect().top),
          "Rows control top aligns with View"
        ).to.eq(viewTop);
      });
      cy.get('button[aria-label="Switch to cards view"]').should(($cards) => {
        expect(
          Math.round($cards[0].getBoundingClientRect().top),
          "Cards control top aligns with View"
        ).to.eq(viewTop);
      });
    });

    cy.contains("th", "Purchased").should("exist");
    cy.contains("th", "Delivered").should("not.exist");
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

  it("UI-SCREEN-WISHLIST-016 renders date added and price update context", () => {
    signInToWishlist();

    cy.get('button[aria-label="Switch to rows view"]').click();
    cy.contains("th", "Date added").should("exist");
    cy.contains("th", "Updated").should("exist");
    cy.get('[data-testid="wishlist-date-added-item-collector-1"]').should(
      "contain.text",
      "Mar 5, 2026"
    );
    cy.get('[data-testid="wishlist-date-updated-item-collector-1"]')
      .should("contain.text", "Apr 22, 2026")
      .and("have.attr", "title", "Latest pricing refresh date");
    cy.get('[data-testid="wishlist-date-added-item-collector-2"]').should(
      "contain.text",
      "Mar 10, 2026"
    );
    cy.get('[data-testid="wishlist-date-updated-item-collector-2"]').should(
      "contain.text",
      "-"
    );

    cy.contains("button", "Cards").click();
    cy.get('[data-testid="wishlist-card-date-added-item-collector-1"]').should(
      "contain.text",
      "Date added: Mar 5, 2026"
    );
    cy.get('[data-testid="wishlist-card-date-updated-item-collector-1"]').should(
      "contain.text",
      "Updated: Apr 22, 2026"
    );
    cy.get('[data-testid="wishlist-card-date-updated-item-collector-2"]').should(
      "contain.text",
      "Updated: -"
    );
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

  it("UI-SCREEN-WISHLIST-017 edits cost and quantity with stable stepper controls", () => {
    let wishlistEntries = [
      {
        id: "wish-stepper-1",
        item_id: "item-collector-1",
        priority: "medium",
        below_target_now: false,
        target_price: 0,
        quantity: 0,
        needed_quantity: 1,
      },
    ];
    const wishlistItems = [
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
    cy.intercept("GET", "/api/items?status=wishlist", {
      statusCode: 200,
      body: { items: wishlistItems },
    }).as("catalogItems");
    cy.intercept("PUT", "/api/wishlist", (req) => {
      wishlistEntries = wishlistEntries.map((entry) =>
        entry.id === req.body.id
          ? {
              ...entry,
              target_price: req.body.target_price ?? entry.target_price,
              quantity: req.body.quantity ?? entry.quantity,
              needed_quantity:
                req.body.needed_quantity ?? entry.needed_quantity,
            }
          : entry
      );
      req.reply({ statusCode: 204, body: "" });
    }).as("updateWishlistEntry");

    signInToWishlist({ skipStub: true, useExistingIntercepts: true });
    cy.get('button[aria-label="Switch to rows view"]').click();
    cy.contains("AFX Mega-G+ Camaro Wildfire").should("exist");
    cy.contains("th", "Cost").scrollIntoView();

    cy.get('[data-testid="wishlist-cost-stepper-item-collector-1"]')
      .scrollIntoView()
      .should("be.visible")
      .and("have.class", "w-[8.75rem]");
    cy.get('[data-testid="wishlist-cost-input-item-collector-1"]')
      .should("have.value", "0")
      .and("have.attr", "min", "0")
      .and("have.attr", "step", "0.01")
      .and("have.class", "[appearance:textfield]");
    cy.get('[data-testid="wishlist-cost-decrease-item-collector-1"]').click();
    cy.get('[data-testid="wishlist-cost-input-item-collector-1"]').should(
      "have.value",
      "0"
    );
    cy.get('[data-testid="wishlist-cost-input-item-collector-1"]').type(
      "{selectall}15.5{enter}"
    );
    cy.wait("@updateWishlistEntry")
      .its("request.body")
      .should("include", { target_price: 15.5 });
    cy.get('[data-testid="wishlist-cost-input-item-collector-1"]').should(
      "not.be.disabled"
    );
    cy.get('[data-testid="wishlist-cost-increase-item-collector-1"]').click();
    cy.wait("@updateWishlistEntry")
      .its("request.body")
      .should("include", { target_price: 15.51 });

    cy.contains("th", "Qty").scrollIntoView();
    cy.get('[data-testid="wishlist-qty-stepper-item-collector-1"]')
      .scrollIntoView()
      .should("be.visible")
      .and("have.class", "w-[7rem]");
    cy.get('[data-testid="wishlist-qty-input-item-collector-1"]')
      .should("have.value", "0")
      .and("have.attr", "min", "0")
      .and("have.attr", "step", "1")
      .and("have.class", "[appearance:textfield]");
    cy.get('[data-testid="wishlist-qty-decrease-item-collector-1"]').click();
    cy.get('[data-testid="wishlist-qty-input-item-collector-1"]').should(
      "have.value",
      "0"
    );
    cy.get('[data-testid="wishlist-qty-input-item-collector-1"]').type(
      "{selectall}4{enter}"
    );
    cy.wait("@updateWishlistEntry")
      .its("request.body")
      .should("include", { quantity: 4 });
    cy.get('[data-testid="wishlist-qty-input-item-collector-1"]').should(
      "not.be.disabled"
    );
    cy.get('[data-testid="wishlist-qty-decrease-item-collector-1"]').click();
    cy.wait("@updateWishlistEntry")
      .its("request.body")
      .should("include", { quantity: 3 });

    cy.get('[data-testid="wishlist-qty-decrease-item-collector-1"]').should(
      "have.attr",
      "aria-label",
      "Decrease quantity for AFX Mega-G+ Camaro Wildfire"
    );
    cy.get('[data-testid="wishlist-qty-increase-item-collector-1"]').should(
      "have.attr",
      "aria-label",
      "Increase quantity for AFX Mega-G+ Camaro Wildfire"
    );
  });

  it("UI-SCREEN-WISHLIST-015 renders item pricing trajectory from pricing APIs", () => {
    signInToWishlist();

    cy.wait("@priceStatsCollector1");
    cy.wait("@priceTrendCollector1");
    cy.wait("@priceHistoryCollector1");
    cy.get('button[aria-label="Switch to rows view"]').click();
    cy.get('[data-testid="wishlist-table-surface"]').scrollTo("left", {
      ensureScrollable: false,
    });

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
      cy.get('[data-testid="wishlist-purchase-open-item-collector-1"]')
        .scrollIntoView()
        .should("be.visible");
    });

    cy.get('[data-testid="wishlist-table-surface"]').scrollTo("right", {
      ensureScrollable: false,
    });
    cy.contains("tr", "AFX Mega-G+ Camaro Wildfire").within(() => {
      cy.get('[data-testid="wishlist-price-trend-item-collector-1"]')
        .scrollIntoView()
        .find('[data-testid="wishlist-price-sparkline-item-collector-1"]')
        .should("be.visible");
      cy.get('[data-testid="wishlist-market-price-item-collector-1"]')
        .scrollIntoView()
        .should("contain.text", "$42.50");
      cy.get('[data-testid="wishlist-price-sparkline-item-collector-1"]')
        .scrollIntoView()
        .should("be.visible")
        .and("have.attr", "aria-label")
        .and("contain", "4 price points");
      cy.get('[data-testid="wishlist-price-graph-meta-item-collector-1"]')
        .should("contain.text", "4 points")
        .and("contain.text", "2026-04-01")
        .and("contain.text", "2026-04-22")
        .and("contain.text", "ebay");
    });

    cy.get('[data-testid="wishlist-table-surface"]').scrollTo("left", {
      ensureScrollable: false,
    });
    cy.contains("tr", "F1 Silverline").within(() => {
      cy.get('[data-testid="wishlist-owned-checkbox-item-collector-2"]').should(
        "not.exist"
      );
      cy.get('[data-testid="wishlist-owned-tick-item-collector-2"]').should(
        "not.exist"
      );
      cy.get('[data-testid="wishlist-purchase-open-item-collector-2"]').should(
        "exist"
      );
    });
  });

  it("UI-SCREEN-WISHLIST-020 edits Purchased, Delivered, and Category workflow fields", () => {
    let wishlistEntries = [
      {
        id: "wish-workflow-1",
        item_id: "item-collector-1",
        priority: "medium",
        below_target_now: false,
        owned: false,
        delivered: false,
        target_price: 20,
        quantity: 0,
        needed_quantity: 1,
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
    cy.intercept("PUT", "/api/items/item-collector-1", (req) => {
      expect(req.body.category).to.eq("Race Cars");
      wishlistItems = wishlistItems.map((item) =>
        item.id === "item-collector-1"
          ? { ...item, category: req.body.category }
          : item
      );
      req.reply({ statusCode: 200, body: wishlistItems[0] });
    }).as("updateWishlistItemCategory");
    cy.intercept("PUT", "/api/wishlist", (req) => {
      expect(req.body).to.include({
        id: "wish-workflow-1",
        item_id: "item-collector-1",
        owned: true,
        delivered: true,
      });
      wishlistEntries = wishlistEntries.map((entry) =>
        entry.id === "wish-workflow-1"
          ? { ...entry, owned: req.body.owned, delivered: req.body.delivered }
          : entry
      );
      req.reply({ statusCode: 204, body: "" });
    }).as("updateWishlistWorkflow");

    signInToWishlist({ skipStub: true, useExistingIntercepts: true });
    cy.get('button[aria-label="Switch to rows view"]').click();
    cy.get('[data-testid="wishlist-table-surface"]').scrollTo("left", {
      ensureScrollable: false,
    });
    cy.contains("th", "Purchased").should("exist");
    cy.contains("th", "Delivered").should("exist");
    cy.contains("th", "Category").should("exist");
    cy.contains("th", "Owned").should("not.exist");
    cy.get('[data-testid="wishlist-category-item-collector-1"]').should(
      "contain.text",
      "Slot Cars"
    );
    cy.get('[data-testid="wishlist-delivered-checkbox-item-collector-1"]')
      .scrollIntoView()
      .should("be.visible")
      .and("have.attr", "aria-checked", "false");

    openWishlistRowActions("AFX Mega-G+ Camaro Wildfire");
    cy.contains('[role="menuitem"]', "Edit").click({ force: true });
    cy.contains("Edit Wishlist Entry").should("be.visible");
    cy.contains("Owned").should("not.exist");
    cy.contains("Purchased").should("exist");
    cy.contains("Delivered").should("exist");
    cy.get('input[name="category"]').clear().type("Race Cars");
    cy.get('[data-testid="wishlist-edit-owned"]')
      .scrollIntoView()
      .should("be.visible");
    cy.get('[data-testid="wishlist-edit-delivered"]')
      .scrollIntoView()
      .should("be.visible")
      .click();
    cy.get('[data-testid="wishlist-edit-owned"]').should(
      "have.attr",
      "aria-checked",
      "true"
    );
    cy.contains("button", "Save changes").click();

    cy.wait("@updateWishlistItemCategory");
    cy.wait("@updateWishlistWorkflow");
    cy.wait("@wishlistItems");
    cy.wait("@catalogItems");
    cy.get('[data-testid="wishlist-edit-panel"]').should("not.exist");
    cy.get('[data-testid="wishlist-category-item-collector-1"]').should(
      "contain.text",
      "Race Cars"
    );
    cy.get('[data-testid="wishlist-delivered-checkbox-item-collector-1"]').should(
      "have.attr",
      "aria-checked",
      "true"
    );

    cy.contains("button", "Cards").click();
    cy.get('[data-testid="wishlist-card-purchased-item-collector-1"]').should(
      "contain.text",
      "Purchased: Yes"
    );
    cy.get('[data-testid="wishlist-card-delivered-item-collector-1"]').should(
      "contain.text",
      "Delivered: Yes"
    );
    cy.contains("Category: Race Cars").should("be.visible");
  });

  it("UI-SCREEN-WISHLIST-006 does not expose Mark owned from the row action menu", () => {
    const wishlistEntries = [
      {
        id: "wish-1",
        item_id: "item-collector-1",
        priority: "high",
        below_target_now: false,
      },
    ];
    const wishlistItems = [
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

  it("UI-SCREEN-WISHLIST-019 creates a title-only wishlist entry", () => {
    let wishlistEntries: Array<Record<string, unknown>> = [];
    let wishlistItems: Array<Record<string, unknown>> = [];

    cy.intercept("GET", "/api/wishlist", (req) => {
      req.reply({ statusCode: 200, body: { items: wishlistEntries } });
    }).as("wishlistItems");
    cy.intercept("GET", "/api/items?status=wishlist", (req) => {
      req.reply({ statusCode: 200, body: { items: wishlistItems } });
    }).as("catalogItems");
    cy.intercept("POST", "/api/items", (req) => {
      expect(req.body.title).to.eq("Title Only Wishlist Save");
      expect(req.body.status).to.eq("wishlist");
      expect(req.body.priority).to.eq("medium");
      expect(req.body.part_number).to.match(/^WISH-TITLE-ONLY-WISHLIST-SAVE-/);
      wishlistItems = [
        ...wishlistItems,
        {
          id: "item-title-only-1",
          title: req.body.title,
          part_number: req.body.part_number,
          status: "wishlist",
          category: req.body.category || "General",
          priority: req.body.priority,
        },
      ];
      req.reply({
        statusCode: 201,
        body: wishlistItems[wishlistItems.length - 1],
      });
    }).as("createWishlistItem");
    cy.intercept("POST", "/api/wishlist", (req) => {
      expect(req.body.item_id).to.eq("item-title-only-1");
      expect(req.body.priority).to.eq("medium");
      expect(req.body.target_price).to.eq(0);
      wishlistEntries = [
        ...wishlistEntries,
        {
          id: "wish-title-only-1",
          item_id: req.body.item_id,
          priority: req.body.priority,
          target_price: req.body.target_price,
          notes: req.body.notes,
          quantity: req.body.quantity,
          needed_quantity: req.body.needed_quantity,
        },
      ];
      req.reply({
        statusCode: 201,
        body: wishlistEntries[wishlistEntries.length - 1],
      });
    }).as("createWishlistEntry");

    signInToWishlist({ skipStub: true, useExistingIntercepts: true });

    cy.get('[data-testid="wishlist-new-action"]').click();
    cy.contains("Create Wishlist Entry").should("be.visible");
    cy.get('input[name="title"]').type("Title Only Wishlist Save");
    cy.contains("button", "Save changes").click();

    cy.wait("@createWishlistItem");
    cy.wait("@createWishlistEntry");
    cy.wait("@wishlistItems");
    cy.wait("@catalogItems");
    cy.contains("Wishlist save failed").should("not.exist");
    cy.get('[data-testid="wishlist-create-panel"]').should("not.exist");
    cy.contains("Title Only Wishlist Save").should("be.visible");
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
