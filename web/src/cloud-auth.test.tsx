import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { CloudAuthGate, resolveEntitlementState } from "./cloud-auth";

describe("cloud auth gate", () => {
  it("shows sign-in gate when cloud auth is enabled and user is signed out", () => {
    render(
      <CloudAuthGate
        publishableKey="pk_test_123"
        isSignedIn={false}
        bootstrapState={{ status: "idle" }}
      >
        <div>App Shell</div>
      </CloudAuthGate>,
    );

    expect(screen.getByRole("heading", { name: /sign in to cabinet/i })).toBeInTheDocument();
    expect(screen.queryByText("App Shell")).not.toBeInTheDocument();
  });

  it("renders app shell when signed in", () => {
    render(
      <CloudAuthGate
        publishableKey="pk_test_123"
        isSignedIn
        bootstrapState={{ status: "ready" }}
      >
        <div>App Shell</div>
      </CloudAuthGate>,
    );

    expect(screen.getByText("App Shell")).toBeInTheDocument();
  });
});

describe("entitlement state", () => {
  it("marks pro-only features as blocked for free plan", () => {
    const state = resolveEntitlementState({
      plan: "free",
      features: ["collection_core"],
    });

    expect(state.canUse.aiAssist).toBe(false);
    expect(state.canUse.priceTracking).toBe(false);
    expect(state.canUse.scannerAutomation).toBe(false);
  });

  it("marks pro-only features as allowed for pro plan", () => {
    const state = resolveEntitlementState({
      plan: "pro",
      features: ["collection_core", "ai_assist", "price_tracking", "scanner_automation"],
    });

    expect(state.canUse.aiAssist).toBe(true);
    expect(state.canUse.priceTracking).toBe(true);
    expect(state.canUse.scannerAutomation).toBe(true);
  });
});
