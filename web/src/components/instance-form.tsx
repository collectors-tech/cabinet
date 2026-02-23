import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { BoundTextInput, DateInput, Form, SubmitButton, TextAreaInput } from "./forms";

const instanceStatuses = ["sealed", "blister", "loose", "custom", "on_track"] as const;

const instanceSchema = z.object({
  item_id: z.string().trim().min(1, "Item id is required"),
  condition: z.string().trim().min(1, "Condition is required"),
  status: z
    .string()
    .trim()
    .min(1, "Status is required")
    .refine((value) => instanceStatuses.includes(value as (typeof instanceStatuses)[number]), "Invalid status"),
  quantity: z.preprocess((value) => Number(value), z.number().int().min(1, "Quantity must be at least 1")),
  storage_location: z.string().trim().optional(),
  acquisition_price: z.preprocess(
    (value) => {
      if (value === "" || value === null || typeof value === "undefined") {
        return undefined;
      }
      const parsed = Number(value);
      return Number.isNaN(parsed) ? undefined : parsed;
    },
    z.number().min(0, "Acquisition price must be non-negative").optional(),
  ),
  acquisition_date: z.string().trim().optional(),
  notes: z.string().trim().optional(),
});

export type InstanceFormValues = z.infer<typeof instanceSchema>;

export function InstanceForm({
  itemID,
  onSubmit,
  isSubmitting,
}: {
  itemID: string;
  onSubmit: (values: InstanceFormValues) => Promise<void> | void;
  isSubmitting?: boolean;
}) {
  const form = useForm<InstanceFormValues>({
    resolver: zodResolver(instanceSchema),
    defaultValues: {
      item_id: itemID,
      condition: "",
      status: "sealed",
      quantity: 1,
      storage_location: "",
      acquisition_price: undefined,
      acquisition_date: "",
      notes: "",
    },
  });

  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(async (values) => {
          await onSubmit(values);
          form.reset({
            item_id: itemID,
            condition: "",
            status: "sealed",
            quantity: 1,
            storage_location: "",
            acquisition_price: undefined,
            acquisition_date: "",
            notes: "",
          });
        })}
      >
        <BoundTextInput<InstanceFormValues, "item_id"> name="item_id" label="Item ID" placeholder="Item ID" />
        <BoundTextInput<InstanceFormValues, "condition">
          name="condition"
          label="Instance Condition"
          placeholder="Condition"
        />
        <div className="cabinet-form-item">
          <label className="cabinet-form-label" htmlFor="instance-status">
            Instance Status
          </label>
          <select id="instance-status" aria-label="Instance status" className="cabinet-select" {...form.register("status")}>
            {instanceStatuses.map((status) => (
              <option key={status} value={status}>
                {status}
              </option>
            ))}
          </select>
          {form.formState.errors.status ? <p role="alert">{String(form.formState.errors.status.message || "")}</p> : null}
        </div>
        <BoundTextInput<InstanceFormValues, "quantity">
          name="quantity"
          label="Quantity"
          placeholder="1"
          type="number"
        />
        <BoundTextInput<InstanceFormValues, "storage_location">
          name="storage_location"
          label="Storage Location"
          placeholder="Shelf A1"
        />

        <details>
          <summary>Advanced Instance Fields</summary>
          <div className="cabinet-form-item">
            <label className="cabinet-form-label" htmlFor="instance-acquisition-date">
              Acquisition Date
            </label>
            <DateInput
              id="instance-acquisition-date"
              aria-label="Acquisition date"
              value={form.watch("acquisition_date")}
              onChange={(event) => form.setValue("acquisition_date", event.target.value)}
            />
          </div>
          <BoundTextInput<InstanceFormValues, "acquisition_price">
            name="acquisition_price"
            label="Acquisition Price"
            placeholder="0.00"
          />
          <div className="cabinet-form-item">
            <label className="cabinet-form-label" htmlFor="instance-notes">
              Notes
            </label>
            <TextAreaInput id="instance-notes" aria-label="Instance notes" {...form.register("notes")} />
          </div>
        </details>

        <SubmitButton isSubmitting={Boolean(isSubmitting)}>Add Instance</SubmitButton>
      </form>
    </Form>
  );
}
