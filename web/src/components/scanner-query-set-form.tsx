import { zodResolver } from "@hookform/resolvers/zod";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { BoundTextInput, CheckboxInput, Form, SubmitButton } from "./forms";

const cronRegex = /^(\S+\s+){4}\S+$/;

const scannerQuerySetSchema = z.object({
  name: z.string().trim().min(1, "Query set name is required"),
  keywords: z.string().trim().min(1, "At least one keyword is required"),
  exclusions: z.string().trim().optional(),
  max_price: z.preprocess(
    (value) => {
      if (value === "" || value === null || typeof value === "undefined") {
        return undefined;
      }
      const parsed = Number(value);
      return Number.isNaN(parsed) ? undefined : parsed;
    },
    z.number().min(0, "Max price must be non-negative").optional(),
  ),
  region: z.string().trim().optional(),
  condition: z.string().trim().optional(),
  schedule_cron: z
    .string()
    .trim()
    .optional()
    .refine((value) => !value || cronRegex.test(value), "Cron format must have 5 fields"),
  enabled: z.boolean(),
  rate_limit_rps: z.preprocess(
    (value) => Number(value),
    z.number().int().min(1, "Rate limit must be between 1 and 20").max(20, "Rate limit must be between 1 and 20"),
  ),
  max_retry_count: z.preprocess(
    (value) => Number(value),
    z.number().int().min(0, "Retry count must be between 0 and 10").max(10, "Retry count must be between 0 and 10"),
  ),
});

export type ScannerQuerySetValues = z.infer<typeof scannerQuerySetSchema>;

export function ScannerQuerySetForm({
  initialValues,
  onSubmit,
  onCancel,
  isSubmitting,
}: {
  initialValues?: Partial<ScannerQuerySetValues>;
  onSubmit: (values: ScannerQuerySetValues) => Promise<void> | void;
  onCancel?: () => void;
  isSubmitting?: boolean;
}) {
  const initial = {
    name: initialValues?.name || "",
    keywords: initialValues?.keywords || "",
    exclusions: initialValues?.exclusions || "",
    max_price: initialValues?.max_price,
    region: initialValues?.region || "US",
    condition: initialValues?.condition || "",
    schedule_cron: initialValues?.schedule_cron || "",
    enabled: initialValues?.enabled ?? true,
    rate_limit_rps: initialValues?.rate_limit_rps ?? 2,
    max_retry_count: initialValues?.max_retry_count ?? 1,
  };

  const form = useForm<ScannerQuerySetValues>({
    resolver: zodResolver(scannerQuerySetSchema),
    defaultValues: initial,
  });

  useEffect(() => {
    form.reset(initial);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [
    initialValues?.name,
    initialValues?.keywords,
    initialValues?.exclusions,
    initialValues?.max_price,
    initialValues?.region,
    initialValues?.condition,
    initialValues?.schedule_cron,
    initialValues?.enabled,
    initialValues?.rate_limit_rps,
    initialValues?.max_retry_count,
  ]);

  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(async (values) => {
          await onSubmit(values);
          form.reset(values);
        })}
      >
        <BoundTextInput<ScannerQuerySetValues, "name"> name="name" label="Query Set Name" placeholder="AFX Search" />
        <BoundTextInput<ScannerQuerySetValues, "keywords">
          name="keywords"
          label="Keywords"
          placeholder="afx, porsche"
        />
        <BoundTextInput<ScannerQuerySetValues, "exclusions">
          name="exclusions"
          label="Exclusions"
          placeholder="broken, damaged"
        />
        <BoundTextInput<ScannerQuerySetValues, "max_price">
          name="max_price"
          label="Max Price"
          placeholder="50"
        />
        <BoundTextInput<ScannerQuerySetValues, "region"> name="region" label="Region" placeholder="US" />
        <BoundTextInput<ScannerQuerySetValues, "condition"> name="condition" label="Condition Filter" placeholder="used" />
        <BoundTextInput<ScannerQuerySetValues, "schedule_cron">
          name="schedule_cron"
          label="Schedule Cron"
          placeholder="*/15 * * * *"
        />
        <p>Cron format hint: use 5 fields, for example `*/15 * * * *`.</p>
        <BoundTextInput<ScannerQuerySetValues, "rate_limit_rps">
          name="rate_limit_rps"
          label="Rate Limit (RPS)"
          placeholder="2"
          type="number"
        />
        <p>Rate limit hint: 1-20 requests per second.</p>
        <BoundTextInput<ScannerQuerySetValues, "max_retry_count">
          name="max_retry_count"
          label="Max Retry Count"
          placeholder="1"
          type="number"
        />
        <div className="cabinet-form-item">
          <label className="cabinet-form-label" htmlFor="query-enabled">
            Enabled
          </label>
          <CheckboxInput
            id="query-enabled"
            aria-label="Enabled"
            checked={Boolean(form.watch("enabled"))}
            onChange={(event) => form.setValue("enabled", event.currentTarget.checked, { shouldDirty: true })}
          />
        </div>

        {form.formState.isDirty ? <p>Unsaved changes</p> : null}
        {onCancel ? (
          <button
            type="button"
            onClick={() => {
              form.reset(initial);
              onCancel();
            }}
          >
            Cancel Edit
          </button>
        ) : null}
        <SubmitButton isSubmitting={Boolean(isSubmitting)}>{initialValues?.name ? "Save Query Set" : "Create Query Set"}</SubmitButton>
      </form>
    </Form>
  );
}
