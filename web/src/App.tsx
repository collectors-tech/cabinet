import { useEffect, useState } from "react";

type Theme = "light" | "dark";

function detectInitialTheme(): Theme {
  const saved = localStorage.getItem("cabinet.theme");
  if (saved === "dark" || saved === "light") {
    return saved;
  }
  return "light";
}

export function App() {
  const [theme, setTheme] = useState<Theme>(detectInitialTheme);
  const [profiles, setProfiles] = useState<Array<{ id: string; name: string }>>([]);
  const [activeProfile, setActiveProfile] = useState<{ id: string; name: string } | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    document.body.setAttribute("data-theme", theme);
    localStorage.setItem("cabinet.theme", theme);
  }, [theme]);

  useEffect(() => {
    let disposed = false;
    async function loadProfiles() {
      setLoading(true);
      setError("");
      try {
        const resp = await fetch("/api/profiles");
        if (!resp.ok) {
          throw new Error("failed_to_list_profiles");
        }
        const data = (await resp.json()) as { profiles?: Array<{ id: string; name: string }> };
        if (disposed) {
          return;
        }
        setProfiles(data.profiles ?? []);
      } catch (e) {
        if (disposed) {
          return;
        }
        setError(e instanceof Error ? e.message : "failed_to_load_profiles");
      } finally {
        if (!disposed) {
          setLoading(false);
        }
      }
    }
    loadProfiles();
    return () => {
      disposed = true;
    };
  }, []);

  async function createFirstProfile() {
    setError("");
    try {
      const createResp = await fetch("/api/profiles", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name: "Default" }),
      });
      if (!createResp.ok) {
        throw new Error("failed_to_create_profile");
      }
      const created = (await createResp.json()) as { id: string; name: string };
      setProfiles([created]);

      const activateResp = await fetch("/api/profiles/active", {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ profile_id: created.id }),
      });
      if (!activateResp.ok) {
        throw new Error("failed_to_activate_profile");
      }
      const active = (await activateResp.json()) as { id: string; name: string };
      setActiveProfile(active);
    } catch (e) {
      setError(e instanceof Error ? e.message : "failed_to_setup_profile");
    }
  }

  async function activateProfile(profileID: string) {
    setError("");
    try {
      const activateResp = await fetch("/api/profiles/active", {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ profile_id: profileID }),
      });
      if (!activateResp.ok) {
        throw new Error("failed_to_activate_profile");
      }
      const active = (await activateResp.json()) as { id: string; name: string };
      setActiveProfile(active);
    } catch (e) {
      setError(e instanceof Error ? e.message : "failed_to_activate_profile");
    }
  }

  return (
    <main data-testid="app-shell" className="cabinet-shell">
      <aside className="cabinet-sidebar">
        <h1>Cabinet</h1>
        <nav>
          <a href="#dashboard">Dashboard</a>
          <a href="#collection">Collection</a>
          <a href="#scanner">Scanner</a>
          <a href="#pricing">Pricing</a>
          <a href="#settings">Settings</a>
        </nav>
      </aside>
      <section className="cabinet-content">
        <header className="cabinet-topbar">
          <strong>Runtime connected. UI foundation active.</strong>
          <button
            id="theme-toggle"
            type="button"
            onClick={() => setTheme((current) => (current === "dark" ? "light" : "dark"))}
          >
            Toggle Theme
          </button>
        </header>
        <section className="cabinet-card">
          <h2>Cabinet Frontend Foundation</h2>
          <p>Next: onboarding, WebAuthn flows, collection workspace, and Cypress E2E.</p>
          {loading ? <p>Loading profiles...</p> : null}
          {!loading && profiles.length === 0 ? (
            <div>
              <p>No local profiles found. Create your first profile to continue.</p>
              <button type="button" onClick={createFirstProfile}>
                Create First Profile
              </button>
            </div>
          ) : null}
          {!loading && profiles.length > 0 ? (
            <div>
              <p>Select a profile to continue:</p>
              <ul>
                {profiles.map((p) => (
                  <li key={p.id}>
                    <span>{p.name}</span>{" "}
                    <button type="button" onClick={() => activateProfile(p.id)}>
                      Use {p.name}
                    </button>
                  </li>
                ))}
              </ul>
            </div>
          ) : null}
          {activeProfile ? <p>Active profile: {activeProfile.name}</p> : null}
          {error ? <p>Profile error: {error}</p> : null}
          <ul>
            <li>
              <a href="/healthz">Health Check</a>
            </li>
            <li>
              <a href="/api/runtime">Runtime</a>
            </li>
            <li>
              <a href="/api/runtime/recovery">Recovery State</a>
            </li>
          </ul>
        </section>
      </section>
    </main>
  );
}
