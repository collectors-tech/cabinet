import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { ProfileSettingsForm, SecretsForm, type ProfileSettingsValues, type SecretsValues } from "./settings-secrets-forms";

describe("settings and secrets forms", () => {
  afterEach(() => {
    cleanup();
  });

  it("validates critical profile settings fields", async () => {
    const onSubmit = vi.fn(async (_values: ProfileSettingsValues) => {});
    render(<ProfileSettingsForm onSubmit={onSubmit} initialValues={{ scanner_schedule: "", backup_frequency: "daily", db_path: "" }} />);

    fireEvent.click(screen.getByRole("button", { name: /save profile settings/i }));
    expect(await screen.findByText(/scanner schedule is required/i)).toBeInTheDocument();
    expect(await screen.findByText(/database path is required/i)).toBeInTheDocument();
    expect(onSubmit).not.toHaveBeenCalled();
  });

  it("supports masked secret inputs with reveal toggle and submit", async () => {
    const onSubmit = vi.fn(async (_values: SecretsValues) => {});
    render(<SecretsForm onSubmit={onSubmit} />);

    const openAIInput = screen.getByLabelText(/openai api key/i) as HTMLInputElement;
    expect(openAIInput.type).toBe("password");
    fireEvent.click(screen.getByRole("button", { name: /show secrets/i }));
    expect(openAIInput.type).toBe("text");

    fireEvent.change(openAIInput, { target: { value: "sk-test" } });
    fireEvent.change(screen.getByLabelText(/ebay app id/i), { target: { value: "app-id" } });
    fireEvent.change(screen.getByLabelText(/ebay auth token/i), { target: { value: "token" } });
    fireEvent.click(screen.getByRole("button", { name: /save secrets/i }));

    await waitFor(() => {
      expect(onSubmit).toHaveBeenCalledWith(
        expect.objectContaining({
          openai_api_key: "sk-test",
          ebay_app_id: "app-id",
          ebay_auth_token: "token",
        }),
      );
    });
  });
});
