import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { RecoveryPassphraseForm, SessionTokenForm } from "./auth-forms";

describe("auth forms", () => {
  it("validates and submits recovery passphrase", async () => {
    const onSubmit = vi.fn(async () => undefined);
    render(<RecoveryPassphraseForm onSubmit={onSubmit} isSubmitting={false} />);

    fireEvent.click(screen.getByRole("button", { name: /save recovery passphrase/i }));
    expect(await screen.findByText(/recovery passphrase is required/i)).toBeInTheDocument();

    fireEvent.change(screen.getByLabelText(/recovery passphrase/i), {
      target: { value: "strong-passphrase" },
    });
    fireEvent.click(screen.getByRole("button", { name: /save recovery passphrase/i }));

    await waitFor(() => {
      expect(onSubmit).toHaveBeenCalledWith({ passphrase: "strong-passphrase" });
    });
  });

  it("submits session token for validate and lock actions", async () => {
    const onValidate = vi.fn(async () => undefined);
    const onLock = vi.fn(async () => undefined);
    render(<SessionTokenForm onValidate={onValidate} onLock={onLock} />);

    fireEvent.change(screen.getByLabelText(/session token/i), {
      target: { value: "token-123" },
    });
    fireEvent.click(screen.getByRole("button", { name: /validate session/i }));
    fireEvent.click(screen.getByRole("button", { name: /lock session/i }));

    await waitFor(() => {
      expect(onValidate).toHaveBeenCalledWith({ token: "token-123" });
      expect(onLock).toHaveBeenCalledWith({ token: "token-123" });
    });
  });
});
