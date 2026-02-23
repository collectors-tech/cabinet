describe("Starter onboarding", () => {
  it("completes identity, loads sample data, and keeps advanced workspace gated", () => {
    cy.intercept("GET", "/api/profiles", {
      statusCode: 200,
      body: { profiles: [{ id: "p1", name: "Default" }] },
    });
    cy.intercept("PUT", "/api/profiles/active", {
      statusCode: 200,
      body: { id: "p1", name: "Default" },
    });
    cy.intercept("GET", "/api/profiles/p1/storage", {
      statusCode: 200,
      body: { db_path: "/tmp/p1.db", media_dir: "/tmp/p1/media" },
    });
    cy.intercept("GET", "/api/auth/requirements?profile_id=p1", {
      statusCode: 200,
      body: { requires_registration: true },
    });
    cy.intercept("POST", "/api/auth/webauthn/register/begin", {
      statusCode: 200,
      body: { session_id: "sess-reg-1", options: {} },
    });
    cy.intercept("POST", "/api/auth/webauthn/register/finish", {
      statusCode: 200,
      body: { ok: true },
    });
    cy.intercept("POST", "/api/onboarding/sample-data", {
      statusCode: 200,
      body: {
        created_items: 1,
        total_items: 1,
      },
    });
    cy.intercept("GET", "/api/items", {
      statusCode: 200,
      body: {
        items: [{ id: "i1", part_number: "CAB-DEMO-001", title: "Sample Item", brand: "Demo", category: "General" }],
      },
    });

    cy.visit("/");
    cy.contains("Use Default").click();

    cy.contains("Starter Onboarding").should("exist");
    cy.contains("Collection").should("not.exist");

    cy.contains("Complete Identity").click();
    cy.contains("Auth status: registration_finished").should("exist");

    cy.contains("Load Sample Data").click();
    cy.contains("Onboarding sample data loaded").should("exist");
    cy.contains("Current items: 1").should("exist");
  });
});

