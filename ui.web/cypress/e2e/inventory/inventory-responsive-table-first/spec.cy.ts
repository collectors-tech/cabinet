type InventoryFixtureItem = {
  id: string;
  part_number: string;
  title: string;
  status: string;
  category: string;
  brand: string;
  priority: string;
  description: string;
};

describe("inventory responsive table-first redesign", () => {
  const items: InventoryFixtureItem[] = [
    {
      id: "item-responsive-alpha",
      part_number: "PN-RESP-A",
      title: "Responsive Alpha",
      status: "active",
      category: "Cars",
      brand: "AFX",
      priority: "medium",
      description: "Alpha responsive coverage item",
    },
    {
      id: "item-responsive-bravo",
      part_number: "PN-RESP-B",
      title: "Responsive Bravo",
      status: "used",
      category: "Trains",
      brand: "Tyco",
      priority: "medium",
      description: "Bravo responsive coverage item",
    },
  ];

  function signIn() {
    cy.e2eReset();
    cy.e2eSetSetupState("present");
    cy.intercept("GET", "/api/items", {
      statusCode: 200,
      body: { items },
    }).as("items");
    cy.intercept("GET", "/api/items/item-responsive-alpha/photos", {
      statusCode: 200,
      body: {
        photos: [
          {
            id: "responsive-alpha-photo",
            filename: "responsive-alpha.jpg",
            is_primary: true,
          },
        ],
      },
    }).as("alphaPhotos");
    cy.intercept("GET", "/api/items/item-responsive-bravo/photos", {
      statusCode: 200,
      body: {
        photos: [
          {
            id: "responsive-bravo-photo",
            filename: "responsive-bravo.jpg",
            is_primary: true,
          },
        ],
      },
    }).as("bravoPhotos");
    cy.e2eBootstrap().then(({ profile_id, profile_name }) => {
      cy.useBootstrappedProfile(profile_id, profile_name, { path: "/inventory/" });
    });
    cy.wait("@items");
  }

  function assertNoDocumentHorizontalOverflow() {
    cy.window().then((win) => {
      const scrollWidth = win.document.documentElement.scrollWidth;
      expect(scrollWidth, "document horizontal overflow").to.be.lte(win.innerWidth + 1);
    });
  }

  function assertElementFitsViewport(testId: string) {
    cy.get(`[data-testid="${testId}"]`)
      .should("be.visible")
      .then(($element) => {
        const rect = $element[0].getBoundingClientRect();
        expect(rect.left, `${testId} left edge`).to.be.gte(0);
        expect(rect.right, `${testId} right edge`).to.be.lte(Cypress.config("viewportWidth"));
      });
  }

  it("keeps desktop Inventory table-first with compact filters and page modals", () => {
    cy.viewport(1280, 800);
    signIn();
    cy.wait("@alphaPhotos");

    cy.contains("Collection Browser").should("be.visible");
    cy.get('[data-testid="inventory-folder-tree"]').should("not.exist");
    cy.get('[data-testid="inventory-folder-tree-legacy"]').should("not.be.visible");
    cy.get('[data-testid="inventory-collection-filter"]').should("be.visible");
    cy.get('[data-testid="inventory-collection-filter-select"]').should("be.visible");
    cy.get("table").should("be.visible");
    cy.contains("Responsive Alpha").should("be.visible");
    cy.contains("Responsive Bravo").should("be.visible");
    cy.get('[data-testid="inventory-photos-section"]').should("not.exist");
    cy.get('[data-testid="inventory-barcodes-section"]').should("not.exist");

    cy.get('[data-testid="inventory-collection-filter-select"]').select("Store 1");
    cy.get('[data-testid="inventory-collection-filter-selected"]').should("contain", "Store 1");
    cy.get('[data-testid="collection-active-context"]').should("contain", "Store 1");

    cy.get('[data-testid="inventory-photos-action"]').click();
    cy.get('[data-testid="inventory-photos-dialog"]')
      .should("be.visible")
      .and("contain", "Responsive Alpha");
    cy.contains('[data-testid="inventory-photo-row"]', "responsive-alpha.jpg").should(
      "be.visible"
    );
    cy.get('[data-testid="inventory-photos-dialog-close"]').click();
    cy.get('[data-testid="inventory-photos-dialog"]').should("not.exist");

    cy.get('[data-testid="inventory-barcodes-action"]').click();
    cy.get('[data-testid="inventory-barcodes-dialog"]')
      .should("be.visible")
      .and("contain", "Responsive Alpha");
    cy.get('[data-testid="inventory-barcodes-dialog-close"]').click();
    cy.get('[data-testid="inventory-barcodes-dialog"]').should("not.exist");

    assertNoDocumentHorizontalOverflow();
  });

  it("keeps mobile controls reachable and row quick actions item-scoped", () => {
    cy.viewport(390, 844);
    signIn();
    cy.wait("@alphaPhotos");

    cy.get('[data-testid="inventory-collection-filter"]').should("be.visible");
    cy.get("table").should("exist");
    cy.contains("Responsive Alpha").should("be.visible");
    assertElementFitsViewport("inventory-collection-filter");
    assertElementFitsViewport("inventory-photos-action");
    assertElementFitsViewport("inventory-barcodes-action");
    assertElementFitsViewport("inventory-new-action");
    assertElementFitsViewport("inventory-create-menu-trigger");
    assertNoDocumentHorizontalOverflow();

    cy.get(
      '[data-testid="inventory-item-row-item-responsive-bravo"] [data-testid="inventory-row-photos-action"]'
    ).click();
    cy.wait("@bravoPhotos");
    cy.get('[data-testid="inventory-photos-dialog"]')
      .should("be.visible")
      .and("contain", "Responsive Bravo");
    cy.contains('[data-testid="inventory-photo-row"]', "responsive-bravo.jpg").should(
      "be.visible"
    );
    cy.get('[data-testid="inventory-photos-dialog-close"]').click();

    cy.get(
      '[data-testid="inventory-item-row-item-responsive-bravo"] [data-testid="inventory-row-barcodes-action"]'
    ).click();
    cy.get('[data-testid="inventory-barcodes-dialog"]')
      .should("be.visible")
      .and("contain", "Responsive Bravo");
    cy.get('[data-testid="collection-selected-item"]').should("contain", "PN-RESP-B");
  });

  it("keeps tablet editor panel navigation available from table rows", () => {
    cy.viewport(768, 1024);
    signIn();

    cy.get(
      '[data-testid="inventory-item-row-item-responsive-alpha"] [data-testid="task-row-actions-trigger"]'
    ).trigger("pointerdown", { button: 0, pointerType: "mouse" });
    cy.contains('[role="menuitem"]', "Edit").click({ force: true });
    cy.get('[data-testid="inventory-item-editor-panel"]')
      .should("be.visible")
      .and("contain", "Edit Item");
    cy.get('[data-testid="inventory-item-title"]').should("have.value", "Responsive Alpha");

    cy.get('[data-testid="inventory-item-editor-next"]').click();
    cy.get('[data-testid="inventory-item-title"]').should("have.value", "Responsive Bravo");
    cy.get('[data-testid="inventory-item-editor-previous"]').click();
    cy.get('[data-testid="inventory-item-title"]').should("have.value", "Responsive Alpha");
    cy.get('[data-testid="inventory-item-editor-cancel"]').click();
    cy.get('[data-testid="inventory-item-editor-panel"]').should("not.exist");

    assertNoDocumentHorizontalOverflow();
  });
});
