import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { BoundTextInput, Form, SubmitButton, TextAreaInput } from "./forms";

const collectionItemSchema = z.object({
  part_number: z.string().trim().min(1, "Part number is required"),
  title: z.string().trim().min(1, "Title is required"),
  brand: z.string().trim().optional(),
  category: z.string().trim().optional(),
  make: z.string().trim().optional(),
  model: z.string().trim().optional(),
  year: z.string().trim().optional(),
  scale: z.string().trim().optional(),
  series: z.string().trim().optional(),
  description: z.string().trim().optional(),
  tags: z.string().trim().optional(),
});

export type CollectionItemValues = z.infer<typeof collectionItemSchema>;

export function CollectionItemForm({
  onSubmit,
  isSubmitting,
  submitLabel = "Add Item",
  initialValues,
}: {
  onSubmit: (values: CollectionItemValues) => Promise<void> | void;
  isSubmitting?: boolean;
  submitLabel?: string;
  initialValues?: Partial<CollectionItemValues>;
}) {
  const form = useForm<CollectionItemValues>({
    resolver: zodResolver(collectionItemSchema),
    defaultValues: {
      part_number: initialValues?.part_number || "",
      title: initialValues?.title || "",
      brand: initialValues?.brand || "",
      category: initialValues?.category || "General",
      make: initialValues?.make || "",
      model: initialValues?.model || "",
      year: initialValues?.year || "",
      scale: initialValues?.scale || "",
      series: initialValues?.series || "",
      description: initialValues?.description || "",
      tags: initialValues?.tags || "",
    },
  });

  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(
          async (values) => {
            await onSubmit(values);
          },
          (errors) => {
            const first = Object.keys(errors)[0] as keyof CollectionItemValues | undefined;
            if (first) {
              form.setFocus(first);
            }
          },
        )}
      >
        <BoundTextInput<CollectionItemValues, "part_number">
          name="part_number"
          label="Part Number"
          placeholder="Part Number"
        />
        <BoundTextInput<CollectionItemValues, "title">
          name="title"
          label="Item Title"
          placeholder="Item Title"
        />
        <BoundTextInput<CollectionItemValues, "brand">
          name="brand"
          label="Brand"
          placeholder="Brand"
        />
        <BoundTextInput<CollectionItemValues, "category">
          name="category"
          label="Category"
          placeholder="Category"
        />

        <details>
          <summary>Advanced Metadata</summary>
          <BoundTextInput<CollectionItemValues, "make">
            name="make"
            label="Item Make"
            placeholder="Make"
          />
          <BoundTextInput<CollectionItemValues, "model">
            name="model"
            label="Item Model"
            placeholder="Model"
          />
          <BoundTextInput<CollectionItemValues, "year">
            name="year"
            label="Item Year"
            placeholder="Year"
          />
          <BoundTextInput<CollectionItemValues, "scale">
            name="scale"
            label="Scale"
            placeholder="Scale"
          />
          <BoundTextInput<CollectionItemValues, "series">
            name="series"
            label="Series"
            placeholder="Series"
          />
          <div className="cabinet-form-item">
            <label className="cabinet-form-label" htmlFor="item-description">
              Description
            </label>
            <TextAreaInput id="item-description" aria-label="Item description" {...form.register("description")} />
          </div>
          <BoundTextInput<CollectionItemValues, "tags">
            name="tags"
            label="Tags"
            placeholder="tag1, tag2"
          />
        </details>
        <SubmitButton isSubmitting={Boolean(isSubmitting)}>{submitLabel}</SubmitButton>
      </form>
    </Form>
  );
}
