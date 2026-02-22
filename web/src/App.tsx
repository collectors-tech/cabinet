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

  useEffect(() => {
    document.body.setAttribute("data-theme", theme);
    localStorage.setItem("cabinet.theme", theme);
  }, [theme]);

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
