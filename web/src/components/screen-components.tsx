import { useMemo, useState } from "react";
import { Button, SelectField, TextField } from "./ui-primitives";

type QuickAddValues = {
  part_number?: string;
  title?: string;
  brand?: string;
};

function QuickAddItemForm({
  initialValues,
  onSubmit,
  isSubmitting,
}: {
  initialValues: QuickAddValues;
  onSubmit: (values: Required<Pick<QuickAddValues, "part_number" | "title">> & QuickAddValues) => void;
  isSubmitting?: boolean;
}) {
  const [values, setValues] = useState<QuickAddValues>(initialValues);
  const [error, setError] = useState("");

  return (
    <form
      onSubmit={(event) => {
        event.preventDefault();
        const part_number = (values.part_number || "").trim();
        const title = (values.title || "").trim();
        if (!part_number || !title) {
          setError("Part number and title are required.");
          return;
        }
        setError("");
        onSubmit({ ...values, part_number, title });
      }}
    >
      <TextField value={values.part_number || ""} onChange={(next) => setValues((current) => ({ ...current, part_number: next }))} aria-label="Quick add part number" />
      <TextField value={values.title || ""} onChange={(next) => setValues((current) => ({ ...current, title: next }))} aria-label="Quick add title" />
      <TextField value={values.brand || ""} onChange={(next) => setValues((current) => ({ ...current, brand: next }))} aria-label="Quick add brand" />
      {error ? <p role="alert">{error}</p> : null}
      <Button variant="primary" size="md" loading={Boolean(isSubmitting)} type="submit">
        Quick Add Item
      </Button>
    </form>
  );
}

function ItemListTable({
  rows,
  selectedRowId,
  onSelectRow,
  onSort,
  onPage,
}: {
  rows: Array<{ id: string; part_number: string; title: string; brand?: string }>;
  selectedRowId: string;
  onSelectRow: (id: string) => void;
  onSort: (field: string) => void;
  onPage: (page: number) => void;
}) {
  return (
    <div className="cabinet-item-list-table">
      <div>
        <Button variant="ghost" size="sm" onClick={() => onSort("title")}>
          Sort by title
        </Button>
        <Button variant="ghost" size="sm" onClick={() => onPage(1)}>
          Page 1
        </Button>
      </div>
      <ul>
        {rows.map((row) => (
          <li key={row.id}>
            <button type="button" aria-current={selectedRowId === row.id ? "true" : undefined} onClick={() => onSelectRow(row.id)}>
              {row.title}
            </button>
          </li>
        ))}
      </ul>
    </div>
  );
}

function PhotoUploadPanel({
  selectedItemId,
  onUpload,
}: {
  selectedItemId: string;
  onUpload: (file: File) => void;
}) {
  const [staged, setStaged] = useState<File | null>(null);
  const [error, setError] = useState("");

  function stageFromFileList(files: FileList | null | undefined) {
    if (!files || files.length === 0) {
      return;
    }
    setError("");
    setStaged(files[0]);
  }

  return (
    <section>
      <input
        type="file"
        aria-label="Photo file"
        onChange={(event) => {
          stageFromFileList(event.currentTarget.files);
        }}
      />
      <div
        role="button"
        tabIndex={0}
        aria-label="Photo dropzone"
        onDrop={(event) => {
          event.preventDefault();
          stageFromFileList(event.dataTransfer?.files);
        }}
        onDragOver={(event) => event.preventDefault()}
      >
        Drop photo here
      </div>
      {staged ? <p>{staged.name}</p> : null}
      {error ? <p role="alert">{error}</p> : null}
      <Button
        variant="primary"
        size="sm"
        onClick={() => {
          if (!selectedItemId) {
            setError("Select an item before upload.");
            return;
          }
          if (!staged) {
            setError("Select a photo first.");
            return;
          }
          setError("");
          onUpload(staged);
        }}
      >
        Upload Photo
      </Button>
    </section>
  );
}

function BarcodeLookupResult({
  matches,
  status,
}: {
  matches: Array<{ item_id?: string }>;
  status: "matched" | "none" | "error";
}) {
  if (status === "error") {
    return <p>Barcode lookup failed.</p>;
  }
  if (status === "none") {
    return <p>No local match. Try external search.</p>;
  }
  return <p>{matches.length} matches found.</p>;
}

function AISuggestionPreview({
  suggestion,
  onApply,
  onRetry,
}: {
  suggestion: { title?: string; confidence?: number } | null;
  onApply: () => void;
  onRetry: () => void;
}) {
  if (!suggestion) {
    return (
      <section>
        <p>No suggestion yet.</p>
        <Button variant="secondary" size="sm" onClick={onRetry}>
          Retry
        </Button>
      </section>
    );
  }
  return (
    <section>
      <p>Suggested title: {suggestion.title || "Untitled"}</p>
      {typeof suggestion.confidence === "number" ? <p>Confidence: {Math.round(suggestion.confidence * 100)}%</p> : null}
      <Button variant="primary" size="sm" onClick={onApply}>
        Apply Suggestion
      </Button>
      <Button variant="secondary" size="sm" onClick={onRetry}>
        Retry
      </Button>
    </section>
  );
}

function DiscoveryCandidateRowActions({
  candidateId,
  onIgnore,
  onWishlist,
  onTrack,
  onCreateItem,
}: {
  candidateId: string;
  onIgnore: (candidateId: string) => void;
  onWishlist: (candidateId: string) => void;
  onTrack: (candidateId: string) => void;
  onCreateItem: (candidateId: string) => void;
}) {
  return (
    <div>
      <Button variant="ghost" size="sm" onClick={() => onIgnore(candidateId)}>
        Ignore
      </Button>
      <Button variant="ghost" size="sm" onClick={() => onWishlist(candidateId)}>
        Wishlist
      </Button>
      <Button variant="ghost" size="sm" onClick={() => onTrack(candidateId)}>
        Track
      </Button>
      <Button variant="primary" size="sm" onClick={() => onCreateItem(candidateId)}>
        Create Item
      </Button>
    </div>
  );
}

function ScannerFailureList({
  failures,
  onRetry,
}: {
  failures: Array<{ id?: string; reason?: string; query_set_id?: string }>;
  onRetry: (querySetId: string) => void;
}) {
  if (failures.length === 0) {
    return <p>No scanner failures.</p>;
  }
  return (
    <ul>
      {failures.map((failure, index) => {
        const querySetID = (failure.query_set_id || "").trim();
        return (
          <li key={failure.id || String(index)}>
            {failure.reason || "Unknown failure"}{" "}
            <Button variant="secondary" size="sm" disabled={!querySetID} onClick={() => onRetry(querySetID)}>
              Retry
            </Button>
          </li>
        );
      })}
    </ul>
  );
}

function ExportPanel({
  scope,
  selectedItemId,
  onExport,
}: {
  scope: "item" | "date-range" | "source";
  selectedItemId?: string;
  onExport: (type: "csv" | "json", filters: { scope: string; selectedItemId?: string }) => void;
}) {
  const [format, setFormat] = useState<"csv" | "json">("csv");
  const disabled = scope === "item" && !selectedItemId;

  return (
    <section>
      <SelectField
        aria-label="Export format"
        value={format}
        onChange={(value) => setFormat((value === "json" ? "json" : "csv"))}
        options={[
          { value: "csv", label: "CSV" },
          { value: "json", label: "JSON" },
        ]}
      />
      <Button
        variant="primary"
        size="sm"
        disabled={disabled}
        onClick={() => onExport(format, { scope, selectedItemId })}
      >
        Export
      </Button>
    </section>
  );
}

function BackupRestorePanel({
  backups,
  selectedBackupPath,
  confirmRestore,
  onBackupPathChange,
  onConfirmRestoreChange,
  onRestore,
}: {
  backups: Array<{ path: string; name: string }>;
  selectedBackupPath: string;
  confirmRestore: boolean;
  onBackupPathChange: (path: string) => void;
  onConfirmRestoreChange: (checked: boolean) => void;
  onRestore: (path: string) => void;
}) {
  const canRestore = Boolean(selectedBackupPath && confirmRestore);
  const options = useMemo(
    () => [
      { value: "", label: "Select backup" },
      ...backups.map((backup) => ({ value: backup.path, label: backup.name })),
    ],
    [backups],
  );

  return (
    <section>
      <SelectField
        aria-label="Backup selection"
        value={selectedBackupPath}
        options={options}
        onChange={(next) => onBackupPathChange(next)}
      />
      <label>
        <input type="checkbox" aria-label="Confirm restore" checked={confirmRestore} onChange={(event) => onConfirmRestoreChange(event.currentTarget.checked)} /> Confirm
        restore
      </label>
      <Button variant="danger" size="sm" disabled={!canRestore} onClick={() => onRestore(selectedBackupPath)}>
        Restore Backup
      </Button>
    </section>
  );
}

export {
  AISuggestionPreview,
  BackupRestorePanel,
  BarcodeLookupResult,
  DiscoveryCandidateRowActions,
  ExportPanel,
  ItemListTable,
  PhotoUploadPanel,
  QuickAddItemForm,
  ScannerFailureList,
};
