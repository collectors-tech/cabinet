describe("UI test matrix coverage", () => {
  function bootWorkspace() {
    const profileName = `Matrix-${Date.now()}`;
    return cy
      .request("POST", "/api/profiles", { name: profileName })
      .then((createProfile) => {
        expect(createProfile.status).to.eq(201);
        const profileId = String(createProfile.body.id || "");
        expect(profileId).to.not.eq("");
        cy.request("PUT", "/api/profiles/active", { profile_id: profileId }).its("status").should("eq", 200);
        return profileId;
      });
  }

  it("covers IA destinations with navigation mount checks", () => {
    bootWorkspace().then((profileId) => {
      cy.visit("/", {
        onBeforeLoad(win) {
          win.localStorage.setItem(`cabinet.workspace.${profileId}`, "1");
        },
      });
      cy.contains("button", "Use").click();

      cy.contains("button", "Dashboard").click();
      cy.contains("h3", "Dashboard").should("exist");

      cy.contains("button", "Collection").click();
      cy.contains("h3", "Collection").should("exist");

      cy.contains("button", "Scanner").click();
      cy.contains("h3", "Discovery Scanner").should("exist");

      cy.contains("button", "Discoveries").click();
      cy.contains("h3", "Not In My Collection").should("exist");

      cy.contains("button", "AI Assist").click();
      cy.contains("h3", "AI Assist").should("exist");

      cy.contains("button", "Barcodes").click();
      cy.contains("h3", "Barcodes").should("exist");

      cy.contains("button", "Photos").click();
      cy.contains("h3", "Photos").should("exist");

      cy.contains("button", "Pricing").click();
      cy.contains("h3", "Pricing").should("exist");

      cy.contains("button", "Reports").click();
      cy.contains("h3", "Reports").should("exist");

      cy.contains("button", "Settings").click();
      cy.contains("h3", "Settings and Diagnostics").should("exist");
    });
  });

  it("covers dashboard and scanner failure paths", () => {
    bootWorkspace().then((profileId) => {
      cy.intercept("GET", "/api/dashboard", {
        statusCode: 500,
        body: { error: "dashboard_failed" },
      }).as("dashboardFail");
      cy.intercept("POST", "/api/scanner/run", {
        statusCode: 500,
        body: { error: "scanner_failed" },
      }).as("scannerFail");

      cy.visit("/", {
        onBeforeLoad(win) {
          win.localStorage.setItem(`cabinet.workspace.${profileId}`, "1");
        },
      });
      cy.contains("button", "Use").click();

      cy.contains("button", "Dashboard").click();
      cy.contains("button", "Refresh Dashboard").click();
      cy.wait("@dashboardFail");
      cy.contains(/insight error/i).should("exist");

      cy.contains("button", "Scanner").click();
      cy.contains("button", "Run Scanner Now").click();
      cy.wait("@scannerFail");
      cy.contains(/scanner error/i).should("exist");
    });
  });
});
