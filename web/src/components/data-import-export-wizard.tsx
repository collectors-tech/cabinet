import { useMemo, useState } from "react";

type DryRunResult = {
  total_items?: number;
  new_items?: number;
  conflicts?: number;
  conflict_details?: Array<{ part_number: string; existing_id?: string }>;
};

type ApplyOptions = {
  default_action: "merge" | "create" | "skip";
  overrides: Record<string, "merge" | "create" | "skip">;
};

export function DataImportExportWizard({
  onDryRun,
  onApply,
  onExport,
}: {
  onDryRun: (args: { format: "json" | "csv"; payload: string; mapping: Record<string, string> }) => Promise<DryRunResult>;
  onApply: (args: {
    format: "json" | "csv";
    payload: string;
    mapping: Record<string, string>;
    options: ApplyOptions;
  }) => Promise<unknown>;
  onExport: (args: { format: "json" | "csv"; scope: "full" | "items" }) => Promise<void>;
}) {
  const [format, setFormat] = useState<"json" | "csv">("json");
  const [step, setStep] = useState<1 | 2 | 3>(1);
  const [jsonPayload, setJSONPayload] = useState("");
  const [csvPayload, setCSVPayload] = useState("");
  const [mapping, setMapping] = useState<Record<string, string>>({
    part_number: "",
    title: "",
    brand: "",
    category: "",
  });
  const [dryRun, setDryRun] = useState<DryRunResult | null>(null);
  const [defaultAction, setDefaultAction] = useState<"merge" | "create" | "skip">("create");
  const [overrides, setOverrides] = useState<Record<string, "merge" | "create" | "skip">>({});
  const [status, setStatus] = useState("");
  const [error, setError] = useState("");
  const [exportFormat, setExportFormat] = useState<"json" | "csv">("json");
  const [exportScope, setExportScope] = useState<"full" | "items">("full");

  const payload = format === "json" ? jsonPayload : csvPayload;
  const requiredMappingMissing = useMemo(() => {
    if (format !== "csv") {
      return false;
    }
    return ["part_number", "title", "brand", "category"].some((key) => !mapping[key]?.trim());
  }, [format, mapping]);

  async function runDryRun() {
    if (!payload.trim()) {
      setError("import_payload_required");
      return;
    }
    if (requiredMappingMissing) {
      setError("required_csv_mapping_missing");
      return;
    }
    setError("");
    const result = await onDryRun({ format, payload, mapping });
    setDryRun(result);
    setStep(3);
  }

  async function applyImport() {
    if (!dryRun) {
      return;
    }
    setError("");
    await onApply({
      format,
      payload,
      mapping,
      options: {
        default_action: defaultAction,
        overrides,
      },
    });
    setStatus("import_applied");
  }

  async function runExport() {
    setError("");
    await onExport({ format: exportFormat, scope: exportScope });
    setStatus("export_requested");
  }

  return (
    <div>
      <h4>Import / Export Wizard</h4>
      <div>
        <label htmlFor="import-format">Import format</label>
        <select id="import-format" aria-label="Import format" value={format} onChange={(e) => setFormat(e.target.value as "json" | "csv")}>
          <option value="json">json</option>
          <option value="csv">csv</option>
        </select>
      </div>
      {step === 1 ? (
        <div>
          {format === "json" ? (
            <div>
              <label htmlFor="json-payload">JSON payload</label>
              <textarea id="json-payload" aria-label="JSON payload" value={jsonPayload} onChange={(e) => setJSONPayload(e.target.value)} />
            </div>
          ) : (
            <div>
              <label htmlFor="csv-payload">CSV payload</label>
              <textarea id="csv-payload" aria-label="CSV payload" value={csvPayload} onChange={(e) => setCSVPayload(e.target.value)} />
            </div>
          )}
          <button type="button" onClick={() => setStep(2)}>
            Next
          </button>
        </div>
      ) : null}
      {step === 2 ? (
        <div>
          {format === "csv" ? (
            <div>
              <label htmlFor="map-part">Map part_number</label>
              <input id="map-part" aria-label="Map part_number" value={mapping.part_number} onChange={(e) => setMapping((c) => ({ ...c, part_number: e.target.value }))} />
              <label htmlFor="map-title">Map title</label>
              <input id="map-title" aria-label="Map title" value={mapping.title} onChange={(e) => setMapping((c) => ({ ...c, title: e.target.value }))} />
              <label htmlFor="map-brand">Map brand</label>
              <input id="map-brand" aria-label="Map brand" value={mapping.brand} onChange={(e) => setMapping((c) => ({ ...c, brand: e.target.value }))} />
              <label htmlFor="map-category">Map category</label>
              <input id="map-category" aria-label="Map category" value={mapping.category} onChange={(e) => setMapping((c) => ({ ...c, category: e.target.value }))} />
            </div>
          ) : null}
          <button type="button" onClick={runDryRun}>
            Run Dry-Run
          </button>
        </div>
      ) : null}
      {step === 3 && dryRun ? (
        <div>
          <p>Total items: {String(dryRun.total_items || 0)}</p>
          <p>New items: {String(dryRun.new_items || 0)}</p>
          <p>Conflicts: {String(dryRun.conflicts || 0)}</p>
          <div>
            <label htmlFor="default-conflict-action">Default conflict action</label>
            <select
              id="default-conflict-action"
              aria-label="Default conflict action"
              value={defaultAction}
              onChange={(e) => setDefaultAction(e.target.value as "merge" | "create" | "skip")}
            >
              <option value="merge">merge</option>
              <option value="create">create</option>
              <option value="skip">skip</option>
            </select>
          </div>
          <ul>
            {(dryRun.conflict_details || []).map((conflict) => (
              <li key={conflict.part_number}>
                {conflict.part_number}
                <label htmlFor={`conflict-${conflict.part_number}`}>Conflict {conflict.part_number}</label>
                <select
                  id={`conflict-${conflict.part_number}`}
                  aria-label={`Conflict ${conflict.part_number}`}
                  value={overrides[conflict.part_number] || defaultAction}
                  onChange={(e) => setOverrides((current) => ({ ...current, [conflict.part_number]: e.target.value as "merge" | "create" | "skip" }))}
                >
                  <option value="merge">merge</option>
                  <option value="create">create</option>
                  <option value="skip">skip</option>
                </select>
              </li>
            ))}
          </ul>
          <button type="button" onClick={applyImport}>
            Apply Import
          </button>
        </div>
      ) : null}

      <div>
        <h5>Export Options</h5>
        <label htmlFor="export-format">Export format</label>
        <select id="export-format" aria-label="Export format" value={exportFormat} onChange={(e) => setExportFormat(e.target.value as "json" | "csv")}>
          <option value="json">json</option>
          <option value="csv">csv</option>
        </select>
        <label htmlFor="export-scope">Export scope</label>
        <select id="export-scope" aria-label="Export scope" value={exportScope} onChange={(e) => setExportScope(e.target.value as "full" | "items")}>
          <option value="full">full</option>
          <option value="items">items</option>
        </select>
        <button type="button" onClick={runExport}>
          Run Export
        </button>
      </div>

      {error ? <p>{error}</p> : null}
      {status ? <p>{status}</p> : null}
    </div>
  );
}
