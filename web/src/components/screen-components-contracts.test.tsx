import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import {
  AISuggestionPreview,
  BackupRestorePanel,
  BarcodeLookupResult,
  DiscoveryCandidateRowActions,
  ExportPanel,
  ItemListTable,
  PhotoUploadPanel,
  QuickAddItemForm,
  ScannerFailureList,
} from "./screen-components";

describe("strict screen component contracts", () => {
  afterEach(() => {
    cleanup();
  });

  it("INV-QF-001 required fields enforced", () => {
    const onSubmit = vi.fn();
    render(<QuickAddItemForm initialValues={{}} onSubmit={onSubmit} isSubmitting={false} />);
    fireEvent.click(screen.getByRole("button", { name: /quick add item/i }));
    expect(screen.getByRole("alert")).toHaveTextContent(/part number and title are required/i);
    expect(onSubmit).not.toHaveBeenCalled();
  });

  it("INV-QF-002 submit disabled while submitting", () => {
    render(<QuickAddItemForm initialValues={{}} onSubmit={() => undefined} isSubmitting />);
    expect(screen.getByRole("button", { name: /saving/i })).toBeDisabled();
  });

  it("INV-TBL-001 row select opens details", () => {
    const onSelectRow = vi.fn();
    render(
      <ItemListTable
        rows={[{ id: "i1", part_number: "P1", title: "Car", brand: "AFX" }]}
        selectedRowId=""
        onSelectRow={onSelectRow}
        onSort={() => undefined}
        onPage={() => undefined}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: /car/i }));
    expect(onSelectRow).toHaveBeenCalledWith("i1");
  });

  it("PHO-UP-001 upload without item id blocked", () => {
    const onUpload = vi.fn();
    render(<PhotoUploadPanel selectedItemId="" onUpload={onUpload} />);
    fireEvent.change(screen.getByLabelText(/photo file/i), { target: { files: [new File(["img"], "a.jpg", { type: "image/jpeg" })] } });
    fireEvent.click(screen.getByRole("button", { name: /upload photo/i }));
    expect(screen.getByRole("alert")).toHaveTextContent(/select an item before upload/i);
    expect(onUpload).not.toHaveBeenCalled();
  });

  it("PHO-UP-002 drag/drop stages file", () => {
    const onUpload = vi.fn();
    render(<PhotoUploadPanel selectedItemId="i1" onUpload={onUpload} />);
    const dropzone = screen.getByLabelText(/photo dropzone/i);
    fireEvent.drop(dropzone, { dataTransfer: { files: [new File(["img"], "drop.jpg", { type: "image/jpeg" })] } });
    expect(screen.getByText(/drop\.jpg/i)).toBeInTheDocument();
  });

  it("BAR-RES-001 matched state shows count", () => {
    render(<BarcodeLookupResult matches={[{ item_id: "i1" }, { item_id: "i2" }]} status="matched" />);
    expect(screen.getByText(/2 matches found/i)).toBeInTheDocument();
  });

  it("BAR-RES-002 no-match state suggests external search", () => {
    render(<BarcodeLookupResult matches={[]} status="none" />);
    expect(screen.getByText(/try external search/i)).toBeInTheDocument();
  });

  it("AI-PRV-001 apply hidden when no suggestion", () => {
    render(<AISuggestionPreview suggestion={null} onApply={() => undefined} onRetry={() => undefined} />);
    expect(screen.queryByRole("button", { name: /apply suggestion/i })).not.toBeInTheDocument();
  });

  it("AI-PRV-002 confidence displayed when provided", () => {
    render(<AISuggestionPreview suggestion={{ title: "AFX", confidence: 0.92 }} onApply={() => undefined} onRetry={() => undefined} />);
    expect(screen.getByText(/confidence: 92%/i)).toBeInTheDocument();
  });

  it("DIS-ACT-001 each action sends correct payload", () => {
    const onIgnore = vi.fn();
    const onWishlist = vi.fn();
    const onTrack = vi.fn();
    const onCreateItem = vi.fn();
    render(
      <DiscoveryCandidateRowActions
        candidateId="c1"
        onIgnore={onIgnore}
        onWishlist={onWishlist}
        onTrack={onTrack}
        onCreateItem={onCreateItem}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: /ignore/i }));
    fireEvent.click(screen.getByRole("button", { name: /wishlist/i }));
    fireEvent.click(screen.getByRole("button", { name: /track/i }));
    fireEvent.click(screen.getByRole("button", { name: /create item/i }));
    expect(onIgnore).toHaveBeenCalledWith("c1");
    expect(onWishlist).toHaveBeenCalledWith("c1");
    expect(onTrack).toHaveBeenCalledWith("c1");
    expect(onCreateItem).toHaveBeenCalledWith("c1");
  });

  it("SCN-F-001 retry only enabled with query_set_id", () => {
    const onRetry = vi.fn();
    render(
      <ScannerFailureList
        failures={[
          { id: "f1", reason: "rate limit" },
          { id: "f2", reason: "timeout", query_set_id: "q2" },
        ]}
        onRetry={onRetry}
      />,
    );
    const retryButtons = screen.getAllByRole("button", { name: /retry/i });
    expect(retryButtons[0]).toBeDisabled();
    expect(retryButtons[1]).toBeEnabled();
    fireEvent.click(retryButtons[1]);
    expect(onRetry).toHaveBeenCalledWith("q2");
  });

  it("REP-EXP-001 export disabled with invalid scope", () => {
    render(<ExportPanel scope="item" selectedItemId="" onExport={() => undefined} />);
    expect(screen.getByRole("button", { name: /export/i })).toBeDisabled();
  });

  it("SET-BK-001 restore blocked until confirmation checked", () => {
    const onRestore = vi.fn();
    render(
      <BackupRestorePanel
        backups={[{ path: "a.db", name: "Backup A" }]}
        selectedBackupPath="a.db"
        confirmRestore={false}
        onBackupPathChange={() => undefined}
        onConfirmRestoreChange={() => undefined}
        onRestore={onRestore}
      />,
    );
    const restore = screen.getByRole("button", { name: /restore backup/i });
    expect(restore).toBeDisabled();
    fireEvent.click(restore);
    expect(onRestore).not.toHaveBeenCalled();
  });
});
