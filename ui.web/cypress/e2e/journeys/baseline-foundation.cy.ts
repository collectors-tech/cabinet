type BootstrapResponse = {
  profile_id: string;
  profile_name: string;
  item_ids: string[];
  query_set_id: string;
  thread_id: string;
};

function resetAndBootstrap() {
  cy.e2eReset();
  return cy.e2eBootstrap().then((resp) => {
    expect(resp.item_ids.length).to.be.greaterThan(0);
    return resp;
  });
}

describe("E2E baseline journeys", () => {
  it("onboarding/auth: happy path and failure path", () => {
    cy.e2eReset();
    cy.e2eBootstrap({ minimalProfile: true });

    cy.visit("/");
    cy.contains("Use E2E Local").click();
    cy.contains("Starter Onboarding Wizard").should("exist");
    cy.contains("Complete Identity").click();
    cy.contains(/Auth status: registration_finished/i).should("exist");

    cy.intercept("GET", "/api/auth/requirements*", {
      statusCode: 500,
      body: { error: "failed_to_get_auth_requirements" },
    }).as("authRequirementsFail");
    cy.reload();
    cy.wait("@authRequirementsFail");
    cy.contains("Retry Auth Requirements").should("exist");
  });

  it("shell/navigation: route shell and persisted workspace state", () => {
    resetAndBootstrap().then((boot) => {
      cy.useBootstrappedProfile(boot.profile_id, boot.profile_name, { workspace: true, path: "/" });
      cy.contains("button", "Collection").click();
      cy.contains("h3", "Collection").should("exist");
      cy.reload();
      cy.contains("Onboarding complete.").should("exist");
    });
  });

  it("chat/copilot: open, send, and recover from failure", () => {
    resetAndBootstrap().then((boot) => {
      cy.useBootstrappedProfile(boot.profile_id, boot.profile_name);
      cy.contains("button", "Open Chat").click();
      cy.get('textarea[aria-label="Chat message"]').type("Summarize my collection.");
      cy.contains("button", "Send").click();
      cy.contains("You: Summarize my collection.").should("exist");
      cy.contains("Assistant:").should("exist");

      cy.intercept("GET", "/api/chat/messages*", {
        statusCode: 500,
        body: { error: "failed_to_list_chat_messages" },
      }).as("chatFail");
      cy.reload();
      cy.contains(`Use ${boot.profile_name}`).click();
      cy.contains("button", "Open Chat").click();
      cy.wait("@chatFail");
      cy.contains("Retry Chat").should("exist");
    });
  });

  it("inventory/photos: upload + reorder + fullscreen navigation", () => {
    resetAndBootstrap().then((boot) => {
      cy.useBootstrappedProfile(boot.profile_id, boot.profile_name);
      cy.contains("button", "Photos").click();

      cy.get('input[aria-label="Photo item id"]').clear().type(boot.item_ids[0]);
      cy.contains("button", "Load Photos").click();

      cy.get('input[aria-label="Photo files"]').selectFile(["cypress/fixtures/photo-1.jpg", "cypress/fixtures/photo-2.jpg"], {
        force: true,
      });
      cy.contains("button", "Upload Staged Photos").click();
      cy.contains("photo-1.jpg").should("exist");
      cy.contains("photo-2.jpg").should("exist");

      cy.contains("li", "photo-1.jpg").within(() => {
        cy.contains("button", "Move Down").click();
      });
      cy.contains("li", "photo-2.jpg").within(() => {
        cy.contains("button", "Open Fullscreen Preview").click();
      });
      cy.get('[role="dialog"][aria-label="Fullscreen photo preview"]').should("exist");
      cy.contains("button", "Next Photo").click();
      cy.contains("button", "Previous Photo").click();
      cy.contains("button", "Close Fullscreen").click();
    });
  });

  it("settings: save and reload persisted profile settings", () => {
    resetAndBootstrap().then((boot) => {
      cy.useBootstrappedProfile(boot.profile_id, boot.profile_name);
      cy.contains("button", "Settings").click();
      cy.contains("button", "Load Profile Settings").click();
      cy.get('select[aria-label="Backup frequency"]').select("weekly");
      cy.contains("button", "Save Profile Settings").click();
      cy.contains(/settings_saved/i).should("exist");

      cy.reload();
      cy.contains(`Use ${boot.profile_name}`).click();
      cy.contains("button", "Settings").click();
      cy.contains("button", "Load Profile Settings").click();
      cy.get('select[aria-label="Backup frequency"]').should("have.value", "weekly");
    });
  });

  it("scanner: run and render status visibility", () => {
    resetAndBootstrap().then((boot) => {
      cy.useBootstrappedProfile(boot.profile_id, boot.profile_name);
      cy.contains("button", "Scanner").click();
      cy.get('input[aria-label="Selected query set id"]').clear().type(boot.query_set_id);
      cy.contains("button", "Run Now").click();
      cy.contains(/Scanner run status:/i).should("exist");
    });
  });

  it("discoveries/search: filters update visible records", () => {
    resetAndBootstrap().then((boot) => {
      cy.useBootstrappedProfile(boot.profile_id, boot.profile_name);
      cy.contains("button", "Discoveries").click();
      cy.contains("button", "Load Not In Collection").click();
      cy.contains("E2E Discovery Candidate").should("exist");
      cy.get('input[aria-label="Not in collection query"]').type("no-match-value");
      cy.contains("button", "Load Not In Collection").click();
      cy.contains("E2E Discovery Candidate").should("not.exist");
    });
  });

  it("error envelope: deterministic scanner error surfaced in UI flow", () => {
    resetAndBootstrap().then((boot) => {
      cy.useBootstrappedProfile(boot.profile_id, boot.profile_name);
      cy.contains("button", "Scanner").click();
      cy.intercept("POST", "/api/scanner/run", {
        statusCode: 400,
        body: { error: "failed_to_run_scanner" },
      }).as("scannerError");
      cy.get('input[aria-label="Selected query set id"]').clear().type(boot.query_set_id);
      cy.contains("button", "Run Now").click();
      cy.wait("@scannerError");
      cy.contains(/Scanner error: failed_to_run_scanner/i).should("exist");
    });
  });

  it("license/entitlement: gated pricing actions remain restricted", () => {
    resetAndBootstrap().then((boot) => {
      cy.useBootstrappedProfile(boot.profile_id, boot.profile_name);
      cy.contains("button", "Pricing").click();
      cy.contains("Price tracking is locked. Upgrade your billing plan to Pro.").should("exist");
      cy.contains("button", "Track Pricing").should("be.disabled");
    });
  });

  it("admin operation: run backup from settings and surface completion", () => {
    resetAndBootstrap().then((boot) => {
      cy.useBootstrappedProfile(boot.profile_id, boot.profile_name);
      cy.contains("button", "Settings").click();
      cy.contains("button", "Run Backup").click();
      cy.contains(/backup_ok/i).should("exist");
    });
  });
});
