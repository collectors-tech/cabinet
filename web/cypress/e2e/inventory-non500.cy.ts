describe("Inventory non-500 regression", () => {
  it("renders inventory workspace without global 500 error", () => {
    cy.request("/api/items").then((resp) => {
      expect(resp.status).to.eq(200);
      expect(resp.body).to.have.property("items");
      expect(Array.isArray(resp.body.items)).to.eq(true);
    });

    cy.visit("/inventory");

    cy.contains("500").should("not.exist");
    cy.contains("Oops! Something went wrong").should("not.exist");
    cy.contains("Inventory").should("exist");
    cy.get("main").should("exist");
  });
});

