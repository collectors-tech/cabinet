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
  const [profileStorage, setProfileStorage] = useState<{ db_path?: string; media_dir?: string } | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [authStatus, setAuthStatus] = useState("");
  const [authSessionID, setAuthSessionID] = useState("");
  const [requiresRegistration, setRequiresRegistration] = useState<boolean | null>(null);
  const [credentialJSON, setCredentialJSON] = useState("{}");
  const [sessionToken, setSessionToken] = useState("");
  const [recoveryPassphrase, setRecoveryPassphrase] = useState("");

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

  async function loadProfileStorage(profileID: string) {
    const storageResp = await fetch(`/api/profiles/${profileID}/storage`);
    if (!storageResp.ok) {
      throw new Error("failed_to_load_profile_storage");
    }
    const storage = (await storageResp.json()) as { db_path?: string; media_dir?: string };
    setProfileStorage(storage);
  }

  useEffect(() => {
    let disposed = false;
    async function loadRequirements() {
      if (!activeProfile?.id) {
        setRequiresRegistration(null);
        return;
      }
      try {
        const reqResp = await fetch(`/api/auth/requirements?profile_id=${encodeURIComponent(activeProfile.id)}`);
        if (!reqResp.ok) {
          throw new Error("failed_to_get_auth_requirements");
        }
        const reqs = (await reqResp.json()) as { requires_registration?: boolean };
        if (!disposed) {
          setRequiresRegistration(Boolean(reqs.requires_registration));
        }
      } catch {
        if (!disposed) {
          setRequiresRegistration(null);
        }
      }
    }
    loadRequirements();
    return () => {
      disposed = true;
    };
  }, [activeProfile?.id]);

  async function postJSON(path: string, payload: unknown) {
    const resp = await fetch(path, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    if (!resp.ok) {
      throw new Error(`request_failed:${path}`);
    }
    return (await resp.json()) as Record<string, unknown>;
  }

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
      await loadProfileStorage(active.id);
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
      await loadProfileStorage(active.id);
    } catch (e) {
      setError(e instanceof Error ? e.message : "failed_to_activate_profile");
    }
  }

  async function beginWebAuthnRegistration() {
    if (!activeProfile?.id) {
      return;
    }
    setError("");
    try {
      const data = await postJSON("/api/auth/webauthn/register/begin", { profile_id: activeProfile.id });
      const sid = String(data.session_id || "");
      setAuthSessionID(sid);
      setAuthStatus("registration_begin_ready");
    } catch (e) {
      setError(e instanceof Error ? e.message : "failed_to_begin_registration");
    }
  }

  async function finishWebAuthnRegistration() {
    if (!authSessionID) {
      return;
    }
    setError("");
    try {
      const parsed = JSON.parse(credentialJSON || "{}");
      await postJSON("/api/auth/webauthn/register/finish", { session_id: authSessionID, credential: parsed });
      setAuthStatus("registration_finished");
    } catch (e) {
      setError(e instanceof Error ? e.message : "failed_to_finish_registration");
    }
  }

  async function beginWebAuthnLogin() {
    if (!activeProfile?.id) {
      return;
    }
    setError("");
    try {
      const data = await postJSON("/api/auth/webauthn/login/begin", { profile_id: activeProfile.id });
      const sid = String(data.session_id || "");
      setAuthSessionID(sid);
      setAuthStatus("login_begin_ready");
    } catch (e) {
      setError(e instanceof Error ? e.message : "failed_to_begin_login");
    }
  }

  async function finishWebAuthnLogin() {
    if (!authSessionID) {
      return;
    }
    setError("");
    try {
      const parsed = JSON.parse(credentialJSON || "{}");
      const data = await postJSON("/api/auth/webauthn/login/finish", { session_id: authSessionID, credential: parsed });
      const token = String(data.session_token || "");
      if (token) {
        setSessionToken(token);
      }
      setAuthStatus("login_finished");
    } catch (e) {
      setError(e instanceof Error ? e.message : "failed_to_finish_login");
    }
  }

  async function saveRecoveryPassphrase() {
    if (!activeProfile?.id || !recoveryPassphrase) {
      return;
    }
    setError("");
    try {
      await postJSON("/api/auth/recovery/passphrase", { profile_id: activeProfile.id, passphrase: recoveryPassphrase });
      setAuthStatus("recovery_passphrase_set");
    } catch (e) {
      setError(e instanceof Error ? e.message : "failed_to_set_recovery_passphrase");
    }
  }

  async function beginRecoveryReset() {
    if (!activeProfile?.id || !recoveryPassphrase) {
      return;
    }
    setError("");
    try {
      const data = await postJSON("/api/auth/recovery/reset/begin", {
        profile_id: activeProfile.id,
        passphrase: recoveryPassphrase,
      });
      setAuthSessionID(String(data.session_id || ""));
      setAuthStatus("recovery_begin_ready");
    } catch (e) {
      setError(e instanceof Error ? e.message : "failed_to_begin_recovery_reset");
    }
  }

  async function validateSession() {
    if (!sessionToken) {
      return;
    }
    setError("");
    try {
      await postJSON("/api/auth/session/validate", { session_token: sessionToken });
      setAuthStatus("session_valid");
    } catch (e) {
      setError(e instanceof Error ? e.message : "session_invalid_or_locked");
    }
  }

  async function lockSession() {
    if (!sessionToken) {
      return;
    }
    setError("");
    try {
      await postJSON("/api/auth/session/lock", { session_token: sessionToken });
      setAuthStatus("session_locked");
    } catch (e) {
      setError(e instanceof Error ? e.message : "failed_to_lock_session");
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
          {profileStorage ? (
            <div>
              <p>Database: {profileStorage.db_path || "not set"}</p>
              <p>Media: {profileStorage.media_dir || "not set"}</p>
            </div>
          ) : null}
          {activeProfile ? (
            <div>
              <h3>Authentication</h3>
              <p>Requires registration: {requiresRegistration === null ? "unknown" : String(requiresRegistration)}</p>
              <div>
                <button type="button" onClick={beginWebAuthnRegistration}>
                  Begin WebAuthn Registration
                </button>{" "}
                <button type="button" onClick={finishWebAuthnRegistration}>
                  Finish Registration
                </button>{" "}
                <button type="button" onClick={beginWebAuthnLogin}>
                  Begin WebAuthn Login
                </button>{" "}
                <button type="button" onClick={finishWebAuthnLogin}>
                  Finish Login
                </button>
              </div>
              <p>Auth session: {authSessionID || "none"}</p>
              <textarea
                value={credentialJSON}
                onChange={(e) => setCredentialJSON(e.target.value)}
                rows={4}
                cols={60}
                aria-label="Credential JSON"
              />
              <div>
                <input
                  value={recoveryPassphrase}
                  onChange={(e) => setRecoveryPassphrase(e.target.value)}
                  placeholder="Recovery passphrase"
                  aria-label="Recovery passphrase"
                />{" "}
                <button type="button" onClick={saveRecoveryPassphrase}>
                  Save Recovery Passphrase
                </button>{" "}
                <button type="button" onClick={beginRecoveryReset}>
                  Begin Recovery Reset
                </button>
              </div>
              <div>
                <input
                  value={sessionToken}
                  onChange={(e) => setSessionToken(e.target.value)}
                  placeholder="Session token"
                  aria-label="Session token"
                />{" "}
                <button type="button" onClick={validateSession}>
                  Validate Session
                </button>{" "}
                <button type="button" onClick={lockSession}>
                  Lock Session
                </button>
              </div>
              <p>Auth status: {authStatus || "idle"}</p>
            </div>
          ) : null}
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
