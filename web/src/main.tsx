import React from "react";
import ReactDOM from "react-dom/client";
import { ClerkProvider, SignInButton, SignedIn, SignedOut, useAuth } from "@clerk/clerk-react";
import { App } from "./App";
import { CloudAuthGate, type BootstrapState } from "./cloud-auth";
import "./styles.css";

const publishableKey =
  import.meta.env.VITEST ? undefined : ((import.meta.env.VITE_CLERK_PUBLISHABLE_KEY as string | undefined) || undefined);

function CloudBootstrapShell() {
  const { getToken, isSignedIn } = useAuth();
  const [bootstrapState, setBootstrapState] = React.useState<BootstrapState>({ status: "idle" });

  React.useEffect(() => {
    let alive = true;
    if (!isSignedIn) {
      setBootstrapState({ status: "idle" });
      return () => {
        alive = false;
      };
    }
    (async () => {
      try {
        setBootstrapState({ status: "loading" });
        const token = await getToken();
        if (!token) {
          throw new Error("missing_session_token");
        }
        const resp = await fetch("/api/auth/cloud/session/bootstrap", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ provider: "clerk", token }),
        });
        if (!resp.ok) {
          throw new Error("failed_to_bootstrap_cloud_session");
        }
        const entitlement = (await resp.json()) as { plan?: string; features?: string[] };
        if (alive) {
          const plan = String(entitlement.plan || "free").toLowerCase();
          const features = Array.isArray(entitlement.features)
            ? entitlement.features.map((feature) => String(feature))
            : [];
          localStorage.setItem("cabinet.cloud.plan", plan);
          localStorage.setItem("cabinet.cloud.features", JSON.stringify(features));
          window.dispatchEvent(new Event("cabinet-cloud-entitlement-updated"));
          setBootstrapState({ status: "ready" });
        }
      } catch (error) {
        if (!alive) {
          return;
        }
        const message = error instanceof Error ? error.message : "failed_to_bootstrap_cloud_session";
        setBootstrapState({ status: "error", message });
      }
    })();
    return () => {
      alive = false;
    };
  }, [getToken, isSignedIn]);

  return (
    <>
      <SignedOut>
        <CloudAuthGate
          publishableKey={publishableKey}
          isSignedIn={false}
          bootstrapState={{ status: "idle" }}
          signedOutActions={
            <SignInButton mode="modal">
              <button type="button">Sign In</button>
            </SignInButton>
          }
        >
          <App />
        </CloudAuthGate>
      </SignedOut>
      <SignedIn>
        <CloudAuthGate publishableKey={publishableKey} isSignedIn bootstrapState={bootstrapState}>
          <App />
        </CloudAuthGate>
      </SignedIn>
    </>
  );
}

function Root() {
  if (!publishableKey) {
    return <App />;
  }
  return (
    <ClerkProvider publishableKey={publishableKey}>
      <CloudBootstrapShell />
    </ClerkProvider>
  );
}

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <Root />
  </React.StrictMode>,
);
