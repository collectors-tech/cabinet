import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { ScannerQuerySetForm, type ScannerQuerySetValues } from "./scanner-query-set-form";

describe("ScannerQuerySetForm", () => {
  afterEach(() => {
    cleanup();
  });

  it("shows validated fields with schedule and rate-limit hints", () => {
    const onSubmit = vi.fn(async (_values: ScannerQuerySetValues) => {});
    render(<ScannerQuerySetForm onSubmit={onSubmit} />);

    expect(screen.getByLabelText(/query set name/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/keywords/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/schedule cron/i)).toBeInTheDocument();
    expect(screen.getByText(/cron format/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/rate limit/i)).toBeInTheDocument();
  });

  it("prevents invalid values and supports cancel for dirty edit", async () => {
    const onSubmit = vi.fn(async (_values: ScannerQuerySetValues) => {});
    const onCancel = vi.fn();
    render(
      <ScannerQuerySetForm
        initialValues={{ name: "AFX", keywords: "afx", rate_limit_rps: 2 }}
        onSubmit={onSubmit}
        onCancel={onCancel}
      />,
    );

    fireEvent.change(screen.getByLabelText(/rate limit/i), { target: { value: "0" } });
    fireEvent.click(screen.getByRole("button", { name: /save query set/i }));
    expect(await screen.findByText(/rate limit must be between/i)).toBeInTheDocument();

    fireEvent.change(screen.getByLabelText(/query set name/i), { target: { value: "AFX Updated" } });
    expect(await screen.findByText(/unsaved changes/i)).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: /cancel edit/i }));
    expect(onCancel).toHaveBeenCalled();
    await waitFor(() => {
      expect(screen.getByLabelText(/query set name/i)).toHaveValue("AFX");
    });
  });
});
