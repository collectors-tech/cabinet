describe("ui-screen-collections", () => {
  function signInToCollections() {
    cy.visit("/sign-in?redirect=%2Fcollections%2F");
    cy.get('input[name="email"]').clear().type("e2e-collections@example.com");
    cy.get('input[name="password"]').clear().type("password123");
    cy.contains("button", "Sign in").click();
    cy.location("pathname", { timeout: 15000 }).should(
      "match",
      /^\/collections\/?$/
    );
  }

  it("UI-SCREEN-COLLECTIONS-001 renders top-level collections section with list and New action", () => {
    signInToCollections();

    cy.get('[data-testid="collections-section"]').should("be.visible");
    cy.get('[data-testid="collections-item-all-items"]').should("be.visible");
    cy.get('[data-testid="collections-new-action"]').should("be.visible");
  });

  it("UI-SCREEN-COLLECTIONS-002 exposes New flow and adjacent Create menu quick actions", () => {
    signInToCollections();

    cy.get('[data-testid="collections-new-action"]').click();
    cy.get('[data-testid="collections-create-panel"]').should("be.visible");

    cy.get('[data-testid="collections-create-menu-trigger"]').click();
    cy.get('[data-testid="collections-create-menu-new"]').should("be.visible");
    cy.get('[data-testid="collections-create-menu-starter"]').should("be.visible");
  });

  it("UI-SCREEN-COLLECTIONS-003 supports inline-style quick create and auto-activates new collection", () => {
    signInToCollections();

    cy.get('[data-testid="collections-new-action"]').click();
    cy.get('[data-testid="collections-new-input"]').type("Collections Inline Alpha");
    cy.get('[data-testid="collections-new-save"]').click();

    cy.get('[data-testid="collections-item-collections-inline-alpha"]')
      .should("be.visible")
      .and("have.attr", "data-state", "active");
  });

  it("UI-SCREEN-COLLECTIONS-008 keeps create panel open and shows validation on blank save", () => {
    signInToCollections();

    cy.get('[data-testid="collections-new-action"]').click();
    cy.get('[data-testid="collections-create-panel"]').should("be.visible");
    cy.get('[data-testid="collections-new-save"]').click();
    cy.get('[data-testid="collections-create-panel"]').should("be.visible");
    cy.get('[data-testid="collections-new-validation"]')
      .should("be.visible")
      .and("contain", "Collection name is required.");
    cy.get('[data-testid="collections-new-input"]').should(
      "have.attr",
      "aria-invalid",
      "true"
    );
    cy.get('[data-testid="collections-new-cancel"]').click();
    cy.get('[data-testid="collections-create-panel"]').should("not.exist");
  });
});
