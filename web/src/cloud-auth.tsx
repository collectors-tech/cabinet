import { ReactNode } from "react";

export type BootstrapState =
  | { status: "idle" }
  | { status: "loading" }
  | { status: "ready" }
  | { status: "error"; message: string };

type EntitlementPayload = { plan: string; features: string[] };
type EntitlementState = {
  plan: string;
  features: string[];
  canUse: {
    aiAssist: boolean;
    priceTracking: boolean;
    scannerAutomation: boolean;
  };
};

export function resolveEntitlementState(payload: EntitlementPayload): EntitlementState {
  const featureSet = new Set((payload.features || []).map((feature) => feature.toLowerCase()));
  return {
    plan: (payload.plan || "free").toLowerCase(),
    features: [...featureSet],
    canUse: {
      aiAssist: featureSet.has("ai_assist"),
      priceTracking: featureSet.has("price_tracking"),
      scannerAutomation: featureSet.has("scanner_automation"),
    },
  };
}

export function CloudAuthGate(props: {
  publishableKey?: string;
  isSignedIn: boolean;
  bootstrapState: BootstrapState;
  signedOutActions?: ReactNode;
  children: ReactNode;
}) {
  const cloudAuthEnabled = Boolean(props.publishableKey && props.publishableKey.trim());
  if (!cloudAuthEnabled) {
    return <>{props.children}</>;
  }
  if (!props.isSignedIn) {
    return (
      <main className="cabinet-cloud-gate">
        <section className="cabinet-card">
          <h1>Sign in to Cabinet</h1>
          <p>Cloud account ownership is required before opening your collection workspace.</p>
          {props.signedOutActions}
        </section>
      </main>
    );
  }
  if (props.bootstrapState.status === "loading" || props.bootstrapState.status === "idle") {
    return (
      <main className="cabinet-cloud-gate">
        <section className="cabinet-card">
          <h1>Validating account</h1>
          <p>Checking your Clerk account and plan entitlements.</p>
        </section>
      </main>
    );
  }
  if (props.bootstrapState.status === "error") {
    return (
      <main className="cabinet-cloud-gate">
        <section className="cabinet-card">
          <h1>Cloud validation failed</h1>
          <p>{props.bootstrapState.message}</p>
        </section>
      </main>
    );
  }
  return <>{props.children}</>;
}
