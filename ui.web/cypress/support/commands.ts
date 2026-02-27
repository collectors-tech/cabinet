type E2EBootstrapOptions = {
  minimalProfile?: boolean;
};

type E2EBootstrapState = {
  profile_id: string;
  profile_name: string;
  item_ids: string[];
  query_set_id: string;
  thread_id: string;
};

declare global {
  namespace Cypress {
    interface Chainable {
      e2eReset(): Chainable<void>;
      e2eBootstrap(options?: E2EBootstrapOptions): Chainable<E2EBootstrapState>;
      useBootstrappedProfile(profileId: string, profileName: string, options?: { workspace?: boolean; path?: string }): Chainable<void>;
    }
  }
}

Cypress.Commands.add("e2eReset", () => {
  cy.request("POST", "/api/test/reset", {}).its("status").should("eq", 200);
});

Cypress.Commands.add("e2eBootstrap", (options: E2EBootstrapOptions = {}) => {
  return cy
    .request<E2EBootstrapState>("POST", "/api/test/bootstrap", {
      minimal_profile: Boolean(options.minimalProfile),
    })
    .then((resp) => {
      expect(resp.status).to.eq(200);
      expect(resp.body.profile_id).to.eq("e2e-profile-001");
      expect(resp.body.profile_name).to.eq("E2E Local");
      return resp.body;
    });
});

Cypress.Commands.add("useBootstrappedProfile", (profileId: string, profileName: string, options?: { workspace?: boolean; path?: string }) => {
  const withWorkspace = options?.workspace ?? true;
  const targetPath = options?.path ?? "/";
  cy.visit(targetPath, {
    onBeforeLoad(win) {
      if (withWorkspace) {
        win.localStorage.setItem(`cabinet.workspace.${profileId}`, "1");
      }
    },
  });
  cy.contains(`Use ${profileName}`).click();
});

export {};
