import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { DataImportExportWizard } from "./data-import-export-wizard";

describe("DataImportExportWizard", () => {
  afterEach(() => {
    cleanup();
  });

  it("runs csv mapping dry-run then apply with conflict choices", async () => {
    const onDryRun = vi.fn(async () => ({
      total_items: 2,
      new_items: 1,
      conflicts: 1,
      conflict_details: [{ part_number: "PN-1", existing_id: "item-1" }],
    }));
    const onApply = vi.fn(async () => ({ imported: 2 }));
    const onExport = vi.fn(async () => undefined);

    render(<DataImportExportWizard onDryRun={onDryRun} onApply={onApply} onExport={onExport} />);

    fireEvent.change(screen.getByLabelText(/import format/i), { target: { value: "csv" } });
    fireEvent.change(screen.getByLabelText(/csv payload/i), {
      target: { value: "brand,category,part_number,title\nAFX,Cars,PN-1,Item One" },
    });
    fireEvent.click(screen.getByRole("button", { name: /next/i }));

    fireEvent.change(screen.getByLabelText(/map part_number/i), { target: { value: "part_number" } });
    fireEvent.change(screen.getByLabelText(/map title/i), { target: { value: "title" } });
    fireEvent.change(screen.getByLabelText(/map brand/i), { target: { value: "brand" } });
    fireEvent.change(screen.getByLabelText(/map category/i), { target: { value: "category" } });
    fireEvent.click(screen.getByRole("button", { name: /run dry-run/i }));

    await waitFor(() => expect(onDryRun).toHaveBeenCalled());
    expect(await screen.findByText(/conflicts: 1/i)).toBeInTheDocument();

    fireEvent.change(screen.getByLabelText(/conflict pn-1/i), { target: { value: "merge" } });
    fireEvent.click(screen.getByRole("button", { name: /apply import/i }));

    await waitFor(() => expect(onApply).toHaveBeenCalled());
  });
});
