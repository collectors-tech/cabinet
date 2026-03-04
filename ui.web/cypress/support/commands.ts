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

type E2ESetupState = "missing" | "present";

declare global {
  namespace Cypress {
    interface Chainable {
      e2eReset(): Chainable<void>;
      e2eBootstrap(options?: E2EBootstrapOptions): Chainable<E2EBootstrapState>;
      e2eSetSetupState(state: E2ESetupState): Chainable<void>;
      e2eCompleteSetupHelper(overrides?: {
        instance_name?: string;
        profile_key?: string;
        runtime?: { mode?: "auto" | "fixed"; fixed_port?: number };
      }): Chainable<void>;
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

Cypress.Commands.add("e2eSetSetupState", (state: E2ESetupState) => {
  cy.request("POST", "/api/test/runtime/setup-status", { state }).its("status").should("eq", 200);
});

Cypress.Commands.add("e2eCompleteSetupHelper", (overrides = {}) => {
  const instanceName = overrides.instance_name ?? "E2E Setup Helper";
  const profileKey = overrides.profile_key ?? "e2e-setup-helper";
  const runtimeMode = overrides.runtime?.mode ?? "auto";
  const fixedPort = overrides.runtime?.fixed_port;
  return cy
    .request("POST", "/api/runtime/setup-complete", {
      instance_name: instanceName,
      profile_key: profileKey,
      auth: { mode: "local", clerk_publishable_key: "", clerk_sign_in_url: "" },
      storage: { mode: "exe_local", data_dir: "", media_dir: "", portable_mode: false },
      runtime: {
        mode: runtimeMode,
        fixed_port: runtimeMode === "fixed" ? fixedPort ?? 17880 : 0,
      },
      features: { scanner: true, providers: true, chat: true },
    })
    .then((resp) => {
      expect(resp.status).to.eq(200);
      expect(resp.body.setup_required).to.eq(false);
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
