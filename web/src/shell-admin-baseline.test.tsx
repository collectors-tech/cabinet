import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { App } from "./App";

describe("admin baseline shell", () => {
  afterEach(() => {
    cleanup();
    localStorage.clear();
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it("renders baseline shell landmarks (brand meta, grouped nav, top tabs)", () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(JSON.stringify({ profiles: [] }), { status: 200 })));
    render(<App />);

    expect(screen.getByText(/cabinet admin/i)).toBeInTheDocument();
    expect(screen.getByText(/desktop workspace/i)).toBeInTheDocument();
    expect(screen.getByText(/^general$/i)).toBeInTheDocument();
    expect(screen.getByText(/^pages$/i)).toBeInTheDocument();
    expect(screen.getByText(/^other$/i)).toBeInTheDocument();
    expect(screen.getByRole("navigation", { name: /top tabs/i })).toBeInTheDocument();
  });
});

