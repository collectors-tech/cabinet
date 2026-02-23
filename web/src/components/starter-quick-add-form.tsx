import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { BoundTextInput, Form, SubmitButton } from "./forms";

const starterQuickAddSchema = z.object({
  part_number: z.string().trim().min(1, "Part number is required"),
  title: z.string().trim().min(1, "Title is required"),
  brand: z.string().trim().optional(),
  category: z.string().trim().optional(),
  series: z.string().trim().optional(),
  description: z.string().trim().optional(),
});

export type StarterQuickAddValues = z.infer<typeof starterQuickAddSchema>;

export function StarterQuickAddForm({
  onSubmit,
  isSubmitting,
}: {
  onSubmit: (values: StarterQuickAddValues) => Promise<void> | void;
  isSubmitting?: boolean;
}) {
  const form = useForm<StarterQuickAddValues>({
    resolver: zodResolver(starterQuickAddSchema),
    defaultValues: {
      part_number: "",
      title: "",
      brand: "",
      category: "",
      series: "",
      description: "",
    },
  });

  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(async (values) => {
          await onSubmit(values);
          form.reset({
            part_number: "",
            title: "",
            brand: "",
            category: "",
            series: "",
            description: "",
          });
        })}
      >
        <BoundTextInput<StarterQuickAddValues, "part_number">
          name="part_number"
          label="Part Number"
          placeholder="Part Number"
        />
        <BoundTextInput<StarterQuickAddValues, "title">
          name="title"
          label="Item Title"
          placeholder="Item Title"
        />
        <BoundTextInput<StarterQuickAddValues, "brand">
          name="brand"
          label="Brand"
          placeholder="Brand"
        />
        <details>
          <summary>Advanced Fields (Optional)</summary>
          <BoundTextInput<StarterQuickAddValues, "category">
            name="category"
            label="Category"
            placeholder="Category"
          />
          <BoundTextInput<StarterQuickAddValues, "series">
            name="series"
            label="Series"
            placeholder="Series"
          />
          <div className="cabinet-form-item">
            <label className="cabinet-form-label" htmlFor="description">
              Description
            </label>
            <textarea id="description" className="cabinet-textarea" placeholder="Description" {...form.register("description")} />
          </div>
        </details>
        <SubmitButton isSubmitting={Boolean(isSubmitting)}>Add First Item</SubmitButton>
      </form>
    </Form>
  );
}
