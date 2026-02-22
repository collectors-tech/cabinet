import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { App } from "./App";

describe("App shell", () => {
  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it("renders shell and theme control", () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(new Response(JSON.stringify({ profiles: [] }), { status: 200 })),
    );
    render(<App />);
    expect(screen.getByTestId("app-shell")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /toggle theme/i })).toBeInTheDocument();
  });

  it("shows onboarding create flow when no profiles exist", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(new Response(JSON.stringify({ profiles: [] }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ id: "p1", name: "Default" }), { status: 201 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ id: "p1", name: "Default" }), { status: 200 }))
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ db_path: "C:/Cabinet/profiles/p1/cabinet.db", media_dir: "C:/Cabinet/profiles/p1/media" }), {
          status: 200,
        }),
      );
    vi.stubGlobal("fetch", fetchMock);

    render(<App />);
    const create = await screen.findByRole("button", { name: /create first profile/i });
    create.click();
    expect(await screen.findByText(/active profile: default/i)).toBeInTheDocument();
  });

  it("allows activating an existing profile", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ profiles: [{ id: "p1", name: "Alpha" }, { id: "p2", name: "Beta" }] }), {
          status: 200,
        }),
      )
      .mockResolvedValueOnce(new Response(JSON.stringify({ id: "p2", name: "Beta" }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ db_path: "/tmp/p2.db", media_dir: "/tmp/p2/media" }), { status: 200 }));
    vi.stubGlobal("fetch", fetchMock);

    render(<App />);
    const activate = await screen.findByRole("button", { name: /use beta/i });
    activate.click();
    expect(await screen.findByText(/active profile: beta/i)).toBeInTheDocument();
    expect(await screen.findByText(/database: \/tmp\/p2.db/i)).toBeInTheDocument();
    expect(await screen.findByText(/media: \/tmp\/p2\/media/i)).toBeInTheDocument();
  });

  it("starts WebAuthn registration for active profile", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ profiles: [{ id: "p2", name: "Beta" }] }), {
          status: 200,
        }),
      )
      .mockResolvedValueOnce(new Response(JSON.stringify({ id: "p2", name: "Beta" }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ db_path: "/tmp/p2.db", media_dir: "/tmp/p2/media" }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ requires_registration: true }), { status: 200 }))
      .mockResolvedValueOnce(new Response(JSON.stringify({ session_id: "sess-reg-1", options: {} }), { status: 200 }));
    vi.stubGlobal("fetch", fetchMock);

    render(<App />);
    const activate = await screen.findByRole("button", { name: /use beta/i });
    activate.click();

    const begin = await screen.findByRole("button", { name: /begin webauthn registration/i });
    begin.click();
    expect(await screen.findByText(/auth session: sess-reg-1/i)).toBeInTheDocument();
  });
});
